package service

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/yourname/go-blog/internal/model"
	"github.com/yourname/go-blog/internal/pkg/cache"
	"github.com/yourname/go-blog/internal/pkg/cachemanager"
	"github.com/yourname/go-blog/internal/pkg/markdown"
	"github.com/yourname/go-blog/internal/pkg/slug"
	"github.com/yourname/go-blog/internal/pkg/upload"
	"github.com/yourname/go-blog/internal/repository"
	"gorm.io/gorm"
)

const (
	articleDetailCacheTTL = 10 * time.Minute
	articleListCacheTTL   = 5 * time.Minute
)

var (
	ErrArticleNotFound  = errors.New("article not found")
	ErrArticleForbidden = errors.New("article permission denied")
	ErrInvalidStatus    = errors.New("invalid article status")
	ErrInvalidTags      = errors.New("invalid article tags")
	ErrInvalidCategory  = errors.New("invalid category")
)

type ArticleService struct {
	articles   *repository.ArticleRepository
	categories *repository.CategoryRepository
	tags       *repository.TagRepository
	renderer   *markdown.Renderer
	// articleCache 文章详情缓存旁路(回源 DB 时带 User/Category/Tags 预加载)
	articleCache *cachemanager.ServiceManager[model.Article]
	// listCache 列表结果缓存(纯缓存读写,不回源)
	listCache *cachemanager.ServiceManager[model.PageResult]
}

type CreateArticleReq struct {
	Title      string  `json:"title" binding:"required,min=1,max=255"`
	Content    string  `json:"content" binding:"required"`
	CategoryID *int64  `json:"category_id"`
	TagIDs     []int64 `json:"tag_ids"`
	CoverImage string  `json:"cover_image" binding:"omitempty,max=500"`
	Status     string  `json:"status" binding:"omitempty,oneof=draft published archived"`
}

type UpdateArticleReq struct {
	Title      *string `json:"title" binding:"omitempty,min=1,max=255"`
	Content    *string `json:"content"`
	CategoryID *int64  `json:"category_id"`
	TagIDs     []int64 `json:"tag_ids"`
	CoverImage *string `json:"cover_image" binding:"omitempty,max=500"`
	Status     *string `json:"status" binding:"omitempty,oneof=draft published archived"`
}

type ListArticleReq struct {
	Page       int
	PageSize   int
	Status     string
	CategoryID int64
	TagID      int64
	Keyword    string
	UserID     int64
	OwnerOnly  bool
}

type ChangeArticleStatusReq struct {
	Status string `json:"status" binding:"required,oneof=draft published archived"`
}

func NewArticleService(
	articles *repository.ArticleRepository,
	categories *repository.CategoryRepository,
	tags *repository.TagRepository,
	renderer *markdown.Renderer,
	articleCache *cachemanager.ServiceManager[model.Article],
	listCache *cachemanager.ServiceManager[model.PageResult],
) *ArticleService {
	return &ArticleService{
		articles:     articles,
		categories:   categories,
		tags:         tags,
		renderer:     renderer,
		articleCache: articleCache,
		listCache:    listCache,
	}
}

func (s *ArticleService) CreateArticle(ctx context.Context, userID int64, req CreateArticleReq) (*model.Article, error) {
	if req.Status == "" {
		req.Status = model.ArticleStatusDraft
	}
	if !validArticleStatus(req.Status) {
		return nil, ErrInvalidStatus
	}
	req.Status = normalizeAuthorStatus(req.Status)

	if err := s.ensureCategory(req.CategoryID); err != nil {
		return nil, err
	}

	tags, err := s.loadTags(req.TagIDs)
	if err != nil {
		return nil, err
	}

	html, err := s.renderer.Render(req.Content)
	if err != nil {
		return nil, err
	}

	article := &model.Article{
		Title:       strings.TrimSpace(req.Title),
		Slug:        s.uniqueSlug(req.Title, 0),
		Content:     req.Content,
		ContentHTML: html,
		Summary:     summarize(req.Content, 200),
		CoverImage:  strings.TrimSpace(req.CoverImage),
		Status:      req.Status,
		UserID:      userID,
		CategoryID:  req.CategoryID,
	}

	if err := s.articles.Create(article, tags); err != nil {
		return nil, err
	}
	s.invalidateArticleListCache(ctx)
	return s.articles.FindByID(article.ID)
}

func (s *ArticleService) UpdateArticle(ctx context.Context, userID, articleID int64, req UpdateArticleReq) (*model.Article, error) {
	article, err := s.requireAuthor(userID, articleID)
	if err != nil {
		return nil, err
	}

	if req.CategoryID != nil {
		if err := s.ensureCategory(req.CategoryID); err != nil {
			return nil, err
		}
		article.CategoryID = req.CategoryID
	}

	if req.Title != nil {
		title := strings.TrimSpace(*req.Title)
		if title != "" && title != article.Title {
			article.Title = title
			article.Slug = s.uniqueSlug(title, article.ID)
		}
	}

	if req.Content != nil {
		html, err := s.renderer.Render(*req.Content)
		if err != nil {
			return nil, err
		}
		article.Content = *req.Content
		article.ContentHTML = html
		article.Summary = summarize(*req.Content, 200)
	}

	if req.CoverImage != nil {
		article.CoverImage = strings.TrimSpace(*req.CoverImage)
	}
	if req.Status != nil {
		if !validArticleStatus(*req.Status) {
			return nil, ErrInvalidStatus
		}
		article.Status = normalizeAuthorStatus(*req.Status)
		if article.Status == model.ArticleStatusPendingReview {
			article.ReviewComment = nil
		}
	}

	tags, err := s.loadTags(req.TagIDs)
	if err != nil {
		return nil, err
	}

	if err := s.articles.Update(article, tags); err != nil {
		return nil, err
	}
	s.invalidateArticleCache(ctx, article.ID)
	return s.articles.FindByID(article.ID)
}

func (s *ArticleService) DeleteArticle(ctx context.Context, userID, articleID int64) error {
	if _, err := s.requireAuthor(userID, articleID); err != nil {
		return err
	}
	if err := s.articles.SoftDelete(articleID); err != nil {
		return err
	}
	s.invalidateArticleCache(ctx, articleID)
	return nil
}

func (s *ArticleService) GetArticle(ctx context.Context, articleID int64) (*model.ArticleDetail, error) {
	// 缓存旁路:命中直接返回;未命中回源 DB(带预加载)并异步回填缓存
	article, err := s.articleCache.LookupSingleWithFallback(
		ctx,
		cache.ArticleCacheKey(articleID),
		func(db *gorm.DB) *gorm.DB {
			return db.Preload("User").Preload("Category").Preload("Tags")
		},
		articleDetailCacheTTL,
	)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrArticleNotFound
		}
		return nil, err
	}

	go func() { _ = s.articles.IncrementViewCount(articleID) }()
	return &model.ArticleDetail{Article: *article}, nil
}

func (s *ArticleService) ListArticles(ctx context.Context, req ListArticleReq) (*model.PageResult, error) {
	if !req.OwnerOnly {
		req.Status = model.ArticleStatusPublished
	}
	if cached, ok := s.getArticleListCache(ctx, req); ok {
		return cached, nil
	}
	articles, total, err := s.articles.List(repository.ListArticleFilter{
		Page:       req.Page,
		PageSize:   req.PageSize,
		Status:     req.Status,
		CategoryID: req.CategoryID,
		TagID:      req.TagID,
		Keyword:    strings.TrimSpace(req.Keyword),
		UserID:     req.UserID,
	})
	if err != nil {
		return nil, err
	}
	result := &model.PageResult{
		List:       articles,
		Pagination: model.NewPagination(req.Page, req.PageSize, total),
	}
	s.setArticleListCache(ctx, req, result)
	return result, nil
}

func (s *ArticleService) ChangeStatus(ctx context.Context, userID, articleID int64, status string) error {
	if !validArticleStatus(status) {
		return ErrInvalidStatus
	}
	article, err := s.requireAuthor(userID, articleID)
	if err != nil {
		return err
	}

	nextStatus := status
	switch status {
	case model.ArticleStatusPublished:
		nextStatus = model.ArticleStatusPendingReview
	case model.ArticleStatusDraft:
		if article.Status != model.ArticleStatusPendingReview && article.Status != model.ArticleStatusDraft {
			return ErrInvalidStatus
		}
	case model.ArticleStatusArchived:
		if article.Status != model.ArticleStatusPublished {
			return ErrInvalidStatus
		}
	}

	if err := s.articles.ChangeStatusWithReviewComment(articleID, nextStatus, nil); err != nil {
		return err
	}
	s.invalidateArticleCache(ctx, articleID)
	return nil
}

func (s *ArticleService) ListUserArticles(ctx context.Context, userID int64, req ListArticleReq) (*model.PageResult, error) {
	req.UserID = userID
	req.OwnerOnly = true
	return s.ListArticles(ctx, req)
}

func (s *ArticleService) SubmitForReview(ctx context.Context, userID, articleID int64) error {
	return s.ChangeStatus(ctx, userID, articleID, model.ArticleStatusPublished)
}

func (s *ArticleService) WithdrawReview(ctx context.Context, userID, articleID int64) error {
	return s.ChangeStatus(ctx, userID, articleID, model.ArticleStatusDraft)
}

func (s *ArticleService) Ranking(ctx context.Context, period string, limit int) ([]model.Article, error) {
	return s.articles.Ranking(period, limit)
}

func (s *ArticleService) UploadArticleImage(ctx context.Context, file io.Reader, filename string, size int64) (string, error) {
	return upload.SaveImage("uploads/articles", file, filename, size)
}

func (s *ArticleService) requireAuthor(userID, articleID int64) (*model.Article, error) {
	article, err := s.articles.FindByID(articleID)
	if err != nil {
		return nil, err
	}
	if article == nil {
		return nil, ErrArticleNotFound
	}
	if article.UserID != userID {
		return nil, ErrArticleForbidden
	}
	return article, nil
}

func (s *ArticleService) ensureCategory(categoryID *int64) error {
	if categoryID == nil || *categoryID == 0 {
		return nil
	}
	category, err := s.categories.FindByID(*categoryID)
	if err != nil {
		return err
	}
	if category == nil {
		return ErrInvalidCategory
	}
	return nil
}

func (s *ArticleService) loadTags(tagIDs []int64) ([]model.Tag, error) {
	if len(tagIDs) == 0 {
		return []model.Tag{}, nil
	}

	unique := make([]int64, 0, len(tagIDs))
	seen := make(map[int64]struct{}, len(tagIDs))
	for _, id := range tagIDs {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}

	tags, err := s.tags.FindByIDs(unique)
	if err != nil {
		return nil, err
	}
	if len(tags) != len(unique) {
		return nil, ErrInvalidTags
	}
	return tags, nil
}

func (s *ArticleService) uniqueSlug(title string, excludeID int64) string {
	base := slug.Make(title)
	candidate := base
	for i := 0; i < 5; i++ {
		exists, err := s.articles.SlugExists(candidate, excludeID)
		if err != nil || !exists {
			return candidate
		}
		candidate = slug.WithRandomSuffix(base)
	}
	return slug.WithRandomSuffix(base)
}

func validArticleStatus(status string) bool {
	return status == model.ArticleStatusDraft ||
		status == model.ArticleStatusPendingReview ||
		status == model.ArticleStatusPublished ||
		status == model.ArticleStatusArchived
}

func normalizeAuthorStatus(status string) string {
	if status == model.ArticleStatusPublished {
		return model.ArticleStatusPendingReview
	}
	return status
}

var markdownSyntaxPattern = regexp.MustCompile(`(?m)[#>*_` + "`" + `\[\]()!-]`)

func summarize(content string, limit int) string {
	text := markdownSyntaxPattern.ReplaceAllString(content, " ")
	text = strings.Join(strings.Fields(text), " ")
	if utf8.RuneCountInString(text) <= limit {
		return text
	}

	runes := []rune(text)
	return string(runes[:limit])
}

func (s *ArticleService) getArticleListCache(ctx context.Context, req ListArticleReq) (*model.PageResult, bool) {
	cached, err := s.listCache.LookupSingle(ctx, cache.ArticleListKey(articleListSignature(req)), nil)
	if err != nil || cached == nil {
		return nil, false
	}
	return cached, true
}

func (s *ArticleService) setArticleListCache(ctx context.Context, req ListArticleReq, result *model.PageResult) {
	_ = s.listCache.WritedownSingle(ctx, cache.ArticleListKey(articleListSignature(req)), result, &cachemanager.WritedownSingleOptions{
		Expiration: articleListCacheTTL,
		Overwrite:  true,
	})
}

func (s *ArticleService) invalidateArticleCache(ctx context.Context, articleID int64) {
	_ = s.articleCache.InvalidateSingleCache(ctx, cache.ArticleCacheKey(articleID))
	s.invalidateArticleListCache(ctx)
}

func (s *ArticleService) invalidateArticleListCache(ctx context.Context) {
	_ = s.listCache.InvalidateCacheByPattern(ctx, cache.ArticleListKeyPattern)
}

func articleListSignature(req ListArticleReq) string {
	source := strings.Join([]string{
		strconv.Itoa(req.Page),
		strconv.Itoa(req.PageSize),
		req.Status,
		strconv.FormatInt(req.CategoryID, 10),
		strconv.FormatInt(req.TagID, 10),
		req.Keyword,
		strconv.FormatInt(req.UserID, 10),
		strconv.FormatBool(req.OwnerOnly),
	}, "|")
	sum := sha1.Sum([]byte(source))
	return hex.EncodeToString(sum[:])
}
