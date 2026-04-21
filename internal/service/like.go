package service

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
	"github.com/yourname/go-blog/internal/model"
	"github.com/yourname/go-blog/internal/pkg/cache"
	"github.com/yourname/go-blog/internal/repository"
)

var ErrInteractionTargetNotFound = errors.New("interaction target not found")

type LikeService struct {
	likes         *repository.LikeRepository
	articles      *repository.ArticleRepository
	notifications *NotificationService
	redis         *redis.Client
}

func NewLikeService(likes *repository.LikeRepository, articles *repository.ArticleRepository, notifications *NotificationService, redis *redis.Client) *LikeService {
	return &LikeService{likes: likes, articles: articles, notifications: notifications, redis: redis}
}

func (s *LikeService) ToggleLike(ctx context.Context, userID, articleID int64) (bool, int64, error) {
	article, err := s.articles.FindByID(articleID)
	if err != nil {
		return false, 0, err
	}
	if article == nil {
		return false, 0, ErrInteractionTargetNotFound
	}

	liked, count, err := s.likes.Toggle(userID, articleID)
	if err != nil {
		return false, 0, err
	}
	s.syncLikeCache(ctx, userID, articleID, liked)
	if liked && s.notifications != nil && article.UserID != userID {
		s.notifications.SendNotification(ctx, article.UserID, model.NotificationTypeLike, map[string]interface{}{
			"article_id":   articleID,
			"from_user_id": userID,
		})
	}
	return liked, count, nil
}

func (s *LikeService) syncLikeCache(ctx context.Context, userID, articleID int64, liked bool) {
	if s.redis == nil {
		return
	}
	key := cache.ArticleLikeSetKey(articleID)
	if liked {
		_ = s.redis.SAdd(ctx, key, userID).Err()
		return
	}
	_ = s.redis.SRem(ctx, key, userID).Err()
}
