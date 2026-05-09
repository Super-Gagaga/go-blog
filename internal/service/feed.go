package service

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/yourname/go-blog/internal/model"
	"github.com/yourname/go-blog/internal/repository"
)

type FeedService struct {
	feeds *repository.FeedRepository
	redis *redis.Client
}

func NewFeedService(feeds *repository.FeedRepository, redis *redis.Client) *FeedService {
	return &FeedService{feeds: feeds, redis: redis}
}

func (s *FeedService) GetUserFeed(ctx context.Context, userID int64, page, pageSize int) (*model.PageResult, error) {
	articles, total, err := s.feeds.UserFeed(userID, page, pageSize)
	if err != nil {
		return nil, err
	}
	return &model.PageResult{
		List:       articles,
		Pagination: model.NewPagination(page, pageSize, total),
	}, nil
}
