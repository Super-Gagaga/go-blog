package service

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
	"github.com/yourname/go-blog/internal/model"
	"github.com/yourname/go-blog/internal/pkg/cache"
	"github.com/yourname/go-blog/internal/repository"
)

type NotificationService struct {
	notifications *repository.NotificationRepository
	users         *repository.UserRepository
	articles      *repository.ArticleRepository
	redis         *redis.Client
}

type MarkNotificationsReadReq struct {
	IDs []int64 `json:"ids"`
}

func NewNotificationService(notifications *repository.NotificationRepository, users *repository.UserRepository, articles *repository.ArticleRepository, redis *redis.Client) *NotificationService {
	return &NotificationService{notifications: notifications, users: users, articles: articles, redis: redis}
}

func (s *NotificationService) SendNotification(ctx context.Context, userID int64, nType string, content map[string]interface{}) {
	if userID <= 0 {
		return
	}
	go func() {
		if err := s.notifications.Create(userID, nType, content); err == nil && s.redis != nil {
			_ = s.redis.Incr(context.Background(), cache.UserUnreadNotificationsKey(userID)).Err()
		}
	}()
}

func (s *NotificationService) NotifyReviewApproved(ctx context.Context, userID int64, articleTitle, articleSlug string) {
	s.SendNotification(ctx, userID, model.NotificationTypeReviewApproved, map[string]interface{}{
		"type":          model.NotificationTypeReviewApproved,
		"title":         "文章审核通过",
		"article_title": articleTitle,
		"article_slug":  articleSlug,
	})
}

func (s *NotificationService) NotifyReviewRejected(ctx context.Context, userID int64, articleTitle, reason string) {
	s.SendNotification(ctx, userID, model.NotificationTypeReviewRejected, map[string]interface{}{
		"type":          model.NotificationTypeReviewRejected,
		"title":         "文章审核未通过",
		"article_title": articleTitle,
		"reason":        reason,
	})
}

func (s *NotificationService) GetNotifications(ctx context.Context, userID int64, page, pageSize int) (*model.PageResult, error) {
	notifications, total, err := s.notifications.ListByUser(userID, page, pageSize)
	if err != nil {
		return nil, err
	}

	// Enrich notifications with from_user and article data
	items := make([]model.NotificationItem, 0, len(notifications))
	for _, n := range notifications {
		item := model.NotificationItem{
			ID:        n.ID,
			Type:      n.Type,
			Content:   n.Content,
			IsRead:    n.IsRead,
			CreatedAt: n.CreatedAt,
		}
		// Parse content JSON to extract ids
		var content map[string]interface{}
		if json.Unmarshal([]byte(n.Content), &content) == nil {
			if v, ok := content["from_user_id"]; ok {
				if uid, ok := toInt64(v); ok {
					u, _ := s.users.FindByID(uid)
					if u != nil {
						item.FromUser = u
					}
				}
			}
			if v, ok := content["article_id"]; ok {
				if aid, ok := toInt64(v); ok {
					a, _ := s.articles.FindByID(aid)
					if a != nil {
						item.Article = a
					}
				}
			}
		}
		items = append(items, item)
	}

	return &model.PageResult{
		List:       items,
		Pagination: model.NewPagination(page, pageSize, total),
	}, nil
}

func toInt64(v interface{}) (int64, bool) {
	switch n := v.(type) {
	case float64:
		return int64(n), true
	case int:
		return int64(n), true
	case int64:
		return n, true
	case json.Number:
		i, err := n.Int64()
		return i, err == nil
	default:
		return 0, false
	}
}

func (s *NotificationService) MarkAsRead(ctx context.Context, userID int64, ids []int64) error {
	if err := s.notifications.MarkAsRead(userID, ids); err != nil {
		return err
	}
	if s.redis != nil {
		count, err := s.notifications.CountUnread(userID)
		if err == nil {
			_ = s.redis.Set(ctx, cache.UserUnreadNotificationsKey(userID), count, 0).Err()
		}
	}
	return nil
}

func (s *NotificationService) UnreadCount(ctx context.Context, userID int64) (int64, error) {
	if s.redis != nil {
		count, err := s.redis.Get(ctx, cache.UserUnreadNotificationsKey(userID)).Int64()
		if err == nil {
			return count, nil
		}
	}
	count, err := s.notifications.CountUnread(userID)
	if err == nil && s.redis != nil {
		_ = s.redis.Set(ctx, cache.UserUnreadNotificationsKey(userID), count, 0).Err()
	}
	return count, err
}
