package service

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
	"github.com/yourname/go-blog/internal/model"
	"github.com/yourname/go-blog/internal/pkg/cache"
	"github.com/yourname/go-blog/internal/pkg/cachemanager"
	"github.com/yourname/go-blog/internal/repository"
)

var (
	ErrCannotFollowSelf   = errors.New("cannot follow yourself")
	ErrFollowUserNotFound = errors.New("follow user not found")
)

type FollowService struct {
	follows       *repository.FollowRepository
	users         *repository.UserRepository
	notifications *NotificationService
	redis         *redis.Client
	// userCache 关注操作会变更双方的 follower/following 计数,需失效资料缓存
	userCache *cachemanager.ServiceManager[model.User]
}

func NewFollowService(follows *repository.FollowRepository, users *repository.UserRepository, notifications *NotificationService, redis *redis.Client, userCache *cachemanager.ServiceManager[model.User]) *FollowService {
	return &FollowService{follows: follows, users: users, notifications: notifications, redis: redis, userCache: userCache}
}

func (s *FollowService) ToggleFollow(ctx context.Context, followerID, followeeID int64) (bool, error) {
	if followerID == followeeID {
		return false, ErrCannotFollowSelf
	}
	followee, err := s.users.FindByID(followeeID)
	if err != nil {
		return false, err
	}
	if followee == nil {
		return false, ErrFollowUserNotFound
	}

	following, err := s.follows.Toggle(followerID, followeeID)
	if err != nil {
		return false, err
	}
	s.syncFollowCache(ctx, followerID, followeeID, following)
	if following && s.notifications != nil {
		s.notifications.SendNotification(ctx, followeeID, model.NotificationTypeFollow, map[string]interface{}{
			"follower_id": followerID,
		})
	}
	return following, nil
}

func (s *FollowService) GetFollowing(ctx context.Context, userID int64, page, pageSize int) (*model.PageResult, error) {
	follows, total, err := s.follows.ListFollowing(userID, page, pageSize)
	if err != nil {
		return nil, err
	}
	return &model.PageResult{List: follows, Pagination: model.NewPagination(page, pageSize, total)}, nil
}

func (s *FollowService) GetFollowers(ctx context.Context, userID int64, page, pageSize int) (*model.PageResult, error) {
	follows, total, err := s.follows.ListFollowers(userID, page, pageSize)
	if err != nil {
		return nil, err
	}
	return &model.PageResult{List: follows, Pagination: model.NewPagination(page, pageSize, total)}, nil
}

func (s *FollowService) syncFollowCache(ctx context.Context, followerID, followeeID int64, following bool) {
	// 关注关系变更会更新双方的粉丝/关注计数,失效两者的资料缓存
	_ = s.userCache.InvalidateSingleCache(ctx, cache.UserCacheKey(followerID))
	_ = s.userCache.InvalidateSingleCache(ctx, cache.UserCacheKey(followeeID))
	if s.redis == nil {
		return
	}
	if following {
		_ = s.redis.SAdd(ctx, cache.UserFollowingSetKey(followerID), followeeID).Err()
		_ = s.redis.SAdd(ctx, cache.UserFollowersSetKey(followeeID), followerID).Err()
		return
	}
	_ = s.redis.SRem(ctx, cache.UserFollowingSetKey(followerID), followeeID).Err()
	_ = s.redis.SRem(ctx, cache.UserFollowersSetKey(followeeID), followerID).Err()
}
