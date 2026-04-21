package service

import (
	"context"
	"errors"
	"strings"

	"github.com/redis/go-redis/v9"
	"github.com/yourname/go-blog/internal/model"
	"github.com/yourname/go-blog/internal/pkg/cache"
	"github.com/yourname/go-blog/internal/repository"
)

var (
	ErrCommentNotFound  = errors.New("comment not found")
	ErrCommentForbidden = errors.New("comment permission denied")
	ErrInvalidComment   = errors.New("invalid comment")
)

type CommentService struct {
	comments      *repository.CommentRepository
	articles      *repository.ArticleRepository
	notifications *NotificationService
	redis         *redis.Client
}

type CreateCommentReq struct {
	Content string `json:"content" binding:"required,min=1,max=2000"`
}

type ReplyCommentReq struct {
	Content string `json:"content" binding:"required,min=1,max=2000"`
}

func NewCommentService(comments *repository.CommentRepository, articles *repository.ArticleRepository, notifications *NotificationService, redis *redis.Client) *CommentService {
	return &CommentService{comments: comments, articles: articles, notifications: notifications, redis: redis}
}

func (s *CommentService) CreateComment(ctx context.Context, userID, articleID int64, req CreateCommentReq) (*model.Comment, error) {
	article, err := s.articles.FindByID(articleID)
	if err != nil {
		return nil, err
	}
	if article == nil {
		return nil, ErrArticleNotFound
	}

	comment := &model.Comment{
		Content:   strings.TrimSpace(req.Content),
		ArticleID: articleID,
		UserID:    userID,
	}
	if comment.Content == "" {
		return nil, ErrInvalidComment
	}
	if err := s.comments.Create(comment); err != nil {
		return nil, err
	}
	s.incrCommentCountCache(ctx, articleID)
	if s.notifications != nil && article.UserID != userID {
		s.notifications.SendNotification(ctx, article.UserID, model.NotificationTypeComment, map[string]interface{}{
			"article_id":   articleID,
			"comment_id":   comment.ID,
			"from_user_id": userID,
		})
	}
	return s.comments.FindByID(comment.ID)
}

func (s *CommentService) ReplyComment(ctx context.Context, userID, commentID int64, req ReplyCommentReq) (*model.Comment, error) {
	parent, err := s.comments.FindByID(commentID)
	if err != nil {
		return nil, err
	}
	if parent == nil {
		return nil, ErrCommentNotFound
	}

	content := strings.TrimSpace(req.Content)
	if content == "" {
		return nil, ErrInvalidComment
	}

	comment := &model.Comment{
		Content:   content,
		ArticleID: parent.ArticleID,
		UserID:    userID,
		ParentID:  &parent.ID,
	}
	if err := s.comments.Create(comment); err != nil {
		return nil, err
	}
	s.incrCommentCountCache(ctx, parent.ArticleID)
	if s.notifications != nil && parent.UserID != userID {
		s.notifications.SendNotification(ctx, parent.UserID, model.NotificationTypeReply, map[string]interface{}{
			"article_id":   parent.ArticleID,
			"comment_id":   comment.ID,
			"parent_id":    parent.ID,
			"from_user_id": userID,
		})
	}
	return s.comments.FindByID(comment.ID)
}

func (s *CommentService) DeleteComment(ctx context.Context, userID, commentID int64) error {
	comment, err := s.comments.FindByID(commentID)
	if err != nil {
		return err
	}
	if comment == nil {
		return ErrCommentNotFound
	}

	article, err := s.articles.FindByID(comment.ArticleID)
	if err != nil {
		return err
	}
	if article == nil {
		return ErrArticleNotFound
	}
	if comment.UserID != userID && article.UserID != userID {
		return ErrCommentForbidden
	}

	if err := s.comments.SoftDelete(commentID); err != nil {
		return err
	}
	s.decrCommentCountCache(ctx, comment.ArticleID)
	return nil
}

func (s *CommentService) GetArticleComments(ctx context.Context, articleID int64) ([]model.CommentTreeNode, error) {
	article, err := s.articles.FindByID(articleID)
	if err != nil {
		return nil, err
	}
	if article == nil {
		return nil, ErrArticleNotFound
	}

	comments, err := s.comments.FindByArticleID(articleID)
	if err != nil {
		return nil, err
	}
	return buildCommentTree(comments), nil
}

func buildCommentTree(comments []model.Comment) []model.CommentTreeNode {
	children := make(map[int64][]model.Comment, len(comments))
	roots := make([]model.Comment, 0)

	for _, comment := range comments {
		if comment.ParentID != nil {
			children[*comment.ParentID] = append(children[*comment.ParentID], comment)
			continue
		}
		roots = append(roots, comment)
	}

	tree := make([]model.CommentTreeNode, 0, len(roots))
	for _, root := range roots {
		tree = append(tree, buildCommentNode(root, children))
	}

	return tree
}

func buildCommentNode(comment model.Comment, children map[int64][]model.Comment) model.CommentTreeNode {
	if comment.DeletedAt.Valid {
		comment.Content = "[deleted]"
		comment.User = model.User{}
	}

	node := model.CommentTreeNode{
		Comment: comment,
		Replies: []model.CommentTreeNode{},
	}
	for _, child := range children[comment.ID] {
		node.Replies = append(node.Replies, buildCommentNode(child, children))
	}
	return node
}

func (s *CommentService) incrCommentCountCache(ctx context.Context, articleID int64) {
	if s.redis == nil {
		return
	}
	_ = s.redis.Incr(ctx, cache.ArticleCommentCountKey(articleID)).Err()
}

func (s *CommentService) decrCommentCountCache(ctx context.Context, articleID int64) {
	if s.redis == nil {
		return
	}
	_ = s.redis.Decr(ctx, cache.ArticleCommentCountKey(articleID)).Err()
}
