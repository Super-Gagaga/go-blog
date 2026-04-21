package service

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/yourname/go-blog/internal/model"
	"github.com/yourname/go-blog/internal/pkg/cache"
	"github.com/yourname/go-blog/internal/repository"
)

type FavoriteService struct {
	favorites *repository.FavoriteRepository
	articles  *repository.ArticleRepository
	redis     *redis.Client
}

func NewFavoriteService(favorites *repository.FavoriteRepository, articles *repository.ArticleRepository, redis *redis.Client) *FavoriteService {
	return &FavoriteService{favorites: favorites, articles: articles, redis: redis}
}

func (s *FavoriteService) ToggleFavorite(ctx context.Context, userID, articleID int64) (bool, error) {
	article, err := s.articles.FindByID(articleID)
	if err != nil {
		return false, err
	}
	if article == nil {
		return false, ErrInteractionTargetNotFound
	}

	favorited, err := s.favorites.Toggle(userID, articleID)
	if err != nil {
		return false, err
	}
	s.syncFavoriteCache(ctx, userID, articleID, favorited)
	return favorited, nil
}

func (s *FavoriteService) GetMyFavorites(ctx context.Context, userID int64, page, pageSize int) (*model.PageResult, error) {
	favorites, total, err := s.favorites.ListByUser(userID, page, pageSize)
	if err != nil {
		return nil, err
	}
	return &model.PageResult{
		List:       favorites,
		Pagination: model.NewPagination(page, pageSize, total),
	}, nil
}

func (s *FavoriteService) syncFavoriteCache(ctx context.Context, userID, articleID int64, favorited bool) {
	if s.redis == nil {
		return
	}
	key := cache.ArticleFavoriteSetKey(articleID)
	if favorited {
		_ = s.redis.SAdd(ctx, key, userID).Err()
		return
	}
	_ = s.redis.SRem(ctx, key, userID).Err()
}
