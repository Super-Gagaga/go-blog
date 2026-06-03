package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"

	"github.com/yourname/go-blog/internal/model"
	"github.com/yourname/go-blog/internal/pkg/cache"
	"github.com/yourname/go-blog/internal/pkg/cachemanager"
	"github.com/yourname/go-blog/internal/repository"
)

var (
	ErrInvalidUserStatus    = errors.New("invalid user status")
	ErrInvalidUserRole      = errors.New("invalid user role")
	ErrReviewReasonRequired = errors.New("review reject reason is required")
	ErrCannotOperateSelf    = errors.New("admin cannot operate self")
)

type AdminService struct {
	users         *repository.UserRepository
	articles      *repository.ArticleRepository
	comments      *repository.CommentRepository
	stats         *repository.StatsRepository
	auditLogs     *repository.AuditLogRepository
	notifications *NotificationService
	articleCache  *cachemanager.ServiceManager[model.Article]
	userCache     *cachemanager.ServiceManager[model.User]
}

type UpdateUserStatusReq struct {
	Status string `json:"status" binding:"required,oneof=active banned"`
}

type UpdateUserRoleReq struct {
	Role string `json:"role" binding:"required,oneof=user admin"`
}

type RejectArticleReq struct {
	Reason string `json:"reason" binding:"required,min=1,max=500"`
}

type AdminArticleFilterReq struct {
	Keyword    string
	Status     string
	UserID     int64
	CategoryID int64
	DateFrom   string
	DateTo     string
}

type AdminCommentFilterReq struct {
	Keyword   string
	UserID    int64
	ArticleID int64
	DateFrom  string
	DateTo    string
}

type AuditLogFilterReq struct {
	AdminID    int64
	Action     string
	TargetType string
	DateFrom   string
	DateTo     string
}

func NewAdminService(
	users *repository.UserRepository,
	articles *repository.ArticleRepository,
	comments *repository.CommentRepository,
	stats *repository.StatsRepository,
	auditLogs *repository.AuditLogRepository,
	notifications *NotificationService,
	articleCache *cachemanager.ServiceManager[model.Article],
	userCache *cachemanager.ServiceManager[model.User],
) *AdminService {
	return &AdminService{
		users:         users,
		articles:      articles,
		comments:      comments,
		stats:         stats,
		auditLogs:     auditLogs,
		notifications: notifications,
		articleCache:  articleCache,
		userCache:     userCache,
	}
}

func (s *AdminService) ListUsers(ctx context.Context, page, pageSize int, keyword string) (*model.PageResult, error) {
	users, total, err := s.users.List(page, pageSize, strings.TrimSpace(keyword))
	if err != nil {
		return nil, err
	}
	return &model.PageResult{List: users, Pagination: model.NewPagination(page, pageSize, total)}, nil
}

func (s *AdminService) GetDashboard(ctx context.Context) (*model.DashboardStats, error) {
	var (
		result model.DashboardStats
		wg     sync.WaitGroup
		once   sync.Once
		errOut error
	)
	setErr := func(err error) {
		if err != nil {
			once.Do(func() { errOut = err })
		}
	}
	setCount := func(dest *int64, fn func() (int64, error)) {
		defer wg.Done()
		count, err := fn()
		if err != nil {
			setErr(err)
			return
		}
		*dest = count
	}

	wg.Add(12)
	go setCount(&result.UserCount, s.stats.CountUsers)
	go setCount(&result.AdminCount, func() (int64, error) { return s.stats.CountUsersByRole("admin") })
	go setCount(&result.ArticleCount, func() (int64, error) { return s.stats.CountArticles("") })
	go setCount(&result.PublishedCount, func() (int64, error) { return s.stats.CountArticles(model.ArticleStatusPublished) })
	go setCount(&result.PendingReviewCount, func() (int64, error) { return s.stats.CountArticles(model.ArticleStatusPendingReview) })
	go setCount(&result.DraftCount, func() (int64, error) { return s.stats.CountArticles(model.ArticleStatusDraft) })
	go setCount(&result.CommentCount, s.stats.CountComments)
	go setCount(&result.LikeCount, s.stats.CountLikes)
	go setCount(&result.FavoriteCount, s.stats.CountFavorites)
	go setCount(&result.FollowCount, s.stats.CountFollows)
	go func() { defer wg.Done(); var err error; result.RecentUsers, err = s.stats.RecentUsers(10); setErr(err) }()
	go func() {
		defer wg.Done()
		var err error
		result.RecentArticles, err = s.stats.RecentArticles(10)
		setErr(err)
	}()
	wg.Wait()
	if errOut != nil {
		return nil, errOut
	}
	authors, err := s.stats.TopAuthors(10)
	if err != nil {
		return nil, err
	}
	result.TopAuthors = authors
	return &result, nil
}

func (s *AdminService) UpdateUserStatus(ctx context.Context, adminID, userID int64, status, clientIP string) error {
	if adminID == userID {
		return ErrCannotOperateSelf
	}
	if status != "active" && status != "banned" {
		return ErrInvalidUserStatus
	}
	if err := s.users.UpdateStatus(userID, status); err != nil {
		return err
	}
	s.invalidateUserCache(ctx, userID)
	action := "unban_user"
	if status == "banned" {
		action = "ban_user"
	}
	s.LogAudit(ctx, adminID, action, "user", userID, `{"status":"`+status+`"}`, clientIP)
	return nil
}

func (s *AdminService) UpdateUserRole(ctx context.Context, adminID, userID int64, role, clientIP string) error {
	if adminID == userID {
		return ErrCannotOperateSelf
	}
	if role != "user" && role != "admin" {
		return ErrInvalidUserRole
	}
	if err := s.users.UpdateRole(userID, role); err != nil {
		return err
	}
	s.invalidateUserCache(ctx, userID)
	s.LogAudit(ctx, adminID, "change_role", "user", userID, `{"role":"`+role+`"}`, clientIP)
	return nil
}

func (s *AdminService) ListPendingArticles(ctx context.Context, page, pageSize int, keyword string) (*model.PageResult, error) {
	articles, total, err := s.articles.ListPending(page, pageSize, strings.TrimSpace(keyword))
	if err != nil {
		return nil, err
	}
	return &model.PageResult{List: articles, Pagination: model.NewPagination(page, pageSize, total)}, nil
}

func (s *AdminService) ApproveArticle(ctx context.Context, adminID, articleID int64, clientIP string) error {
	article, err := s.articles.FindByID(articleID)
	if err != nil {
		return err
	}
	if article == nil {
		return ErrArticleNotFound
	}
	if article.Status != model.ArticleStatusPendingReview {
		return ErrInvalidStatus
	}
	if err := s.articles.ChangeStatusWithReviewComment(articleID, model.ArticleStatusPublished, nil); err != nil {
		return err
	}
	s.invalidateArticleCache(ctx, articleID)
	s.LogAudit(ctx, adminID, "approve_article", "article", articleID, auditDetail(map[string]interface{}{"title": article.Title}), clientIP)
	if s.notifications != nil {
		s.notifications.NotifyReviewApproved(ctx, article.UserID, article.Title, article.Slug)
	}
	return nil
}

func (s *AdminService) RejectArticle(ctx context.Context, adminID, articleID int64, reason, clientIP string) error {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return ErrReviewReasonRequired
	}
	article, err := s.articles.FindByID(articleID)
	if err != nil {
		return err
	}
	if article == nil {
		return ErrArticleNotFound
	}
	if article.Status != model.ArticleStatusPendingReview {
		return ErrInvalidStatus
	}
	if err := s.articles.ChangeStatusWithReviewComment(articleID, model.ArticleStatusDraft, &reason); err != nil {
		return err
	}
	s.invalidateArticleCache(ctx, articleID)
	s.LogAudit(ctx, adminID, "reject_article", "article", articleID, auditDetail(map[string]interface{}{"title": article.Title, "reason": reason}), clientIP)
	if s.notifications != nil {
		s.notifications.NotifyReviewRejected(ctx, article.UserID, article.Title, reason)
	}
	return nil
}

func (s *AdminService) ListArticles(ctx context.Context, page, pageSize int, filter AdminArticleFilterReq) (*model.PageResult, error) {
	articles, total, err := s.articles.ListAll(page, pageSize, repository.AdminArticleFilter{
		Keyword: strings.TrimSpace(filter.Keyword), Status: filter.Status, UserID: filter.UserID,
		CategoryID: filter.CategoryID, DateFrom: filter.DateFrom, DateTo: filter.DateTo,
	})
	if err != nil {
		return nil, err
	}
	return &model.PageResult{List: articles, Pagination: model.NewPagination(page, pageSize, total)}, nil
}

func (s *AdminService) DeleteArticle(ctx context.Context, adminID, articleID int64, clientIP string) error {
	article, err := s.articles.FindByID(articleID)
	if err != nil {
		return err
	}
	if article == nil {
		return ErrArticleNotFound
	}
	if err := s.articles.ForceDelete(articleID); err != nil {
		return err
	}
	s.invalidateArticleCache(ctx, articleID)
	s.LogAudit(ctx, adminID, "delete_article", "article", articleID, auditDetail(map[string]interface{}{"title": article.Title}), clientIP)
	return nil
}

func (s *AdminService) ListComments(ctx context.Context, page, pageSize int, filter AdminCommentFilterReq) (*model.PageResult, error) {
	comments, total, err := s.comments.ListAll(page, pageSize, repository.AdminCommentFilter{
		Keyword: strings.TrimSpace(filter.Keyword), UserID: filter.UserID, ArticleID: filter.ArticleID,
		DateFrom: filter.DateFrom, DateTo: filter.DateTo,
	})
	if err != nil {
		return nil, err
	}
	return &model.PageResult{List: comments, Pagination: model.NewPagination(page, pageSize, total)}, nil
}

func (s *AdminService) DeleteComment(ctx context.Context, adminID, commentID int64, clientIP string) error {
	comment, err := s.comments.FindByID(commentID)
	if err != nil {
		return err
	}
	if comment == nil {
		return ErrCommentNotFound
	}
	if err := s.comments.ForceDelete(commentID); err != nil {
		return err
	}
	s.invalidateArticleCache(ctx, comment.ArticleID)
	s.LogAudit(ctx, adminID, "delete_comment", "comment", commentID, auditDetail(map[string]interface{}{"content": comment.Content}), clientIP)
	return nil
}

func (s *AdminService) LogAudit(ctx context.Context, adminID int64, action, targetType string, targetID int64, detail, ip string) {
	if s.auditLogs == nil {
		return
	}
	go func() {
		_ = s.auditLogs.Create(&model.AuditLog{
			AdminID: adminID, Action: action, TargetType: targetType,
			TargetID: targetID, Detail: detail, IP: ip,
		})
	}()
}

func (s *AdminService) GetAuditLogs(ctx context.Context, page, pageSize int, filter AuditLogFilterReq) (*model.PageResult, error) {
	logs, total, err := s.auditLogs.List(page, pageSize, repository.AuditLogFilter{
		AdminID: filter.AdminID, Action: filter.Action, TargetType: filter.TargetType,
		DateFrom: filter.DateFrom, DateTo: filter.DateTo,
	})
	if err != nil {
		return nil, err
	}
	return &model.PageResult{List: logs, Pagination: model.NewPagination(page, pageSize, total)}, nil
}

func (s *AdminService) invalidateArticleCache(ctx context.Context, articleID int64) {
	_ = s.articleCache.InvalidateSingleCache(ctx, cache.ArticleCacheKey(articleID))
	_ = s.articleCache.InvalidateCacheByPattern(ctx, cache.ArticleListKeyPattern)
}

func (s *AdminService) invalidateUserCache(ctx context.Context, userID int64) {
	_ = s.userCache.InvalidateSingleCache(ctx, cache.UserCacheKey(userID))
}

func auditDetail(v map[string]interface{}) string {
	payload, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(payload)
}
