package service

import (
	"context"
	"fmt"
	"time"

	"github.com/yourname/go-blog/internal/pkg/cachemanager"
	"github.com/yourname/go-blog/internal/repository"
)

const recommendationCacheTTL = 10 * time.Minute

type RecommendationService struct {
	repo *repository.RecommendationRepository
	// 两个 manager 只做纯缓存读写(推荐结果是聚合 DTO 列表,回源逻辑由 repo 提供)
	authorsCache *cachemanager.ServiceManager[[]repository.RecommendedAuthor]
	tagsCache    *cachemanager.ServiceManager[[]repository.TrendingTag]
}

func NewRecommendationService(
	repo *repository.RecommendationRepository,
	authorsCache *cachemanager.ServiceManager[[]repository.RecommendedAuthor],
	tagsCache *cachemanager.ServiceManager[[]repository.TrendingTag],
) *RecommendationService {
	return &RecommendationService{repo: repo, authorsCache: authorsCache, tagsCache: tagsCache}
}

func recommendedAuthorsCacheKey(limit int) string {
	return fmt.Sprintf("recommendations:authors:%d", limit)
}

func trendingTagsCacheKey(limit int) string {
	return fmt.Sprintf("topics:trending:%d", limit)
}

func (s *RecommendationService) GetRecommendedAuthors(ctx context.Context, limit int) ([]repository.RecommendedAuthor, error) {
	cached, err := s.authorsCache.LookupSingle(ctx, recommendedAuthorsCacheKey(limit), nil)
	if err == nil && cached != nil {
		return *cached, nil
	}

	authors, err := s.repo.GetRecommendedAuthors(limit)
	if err != nil {
		return nil, err
	}
	s.authorsCache.WritedownSingleAsync(ctx, recommendedAuthorsCacheKey(limit), &authors, recommendationCacheTTL)
	return authors, nil
}

func (s *RecommendationService) GetTrendingTags(ctx context.Context, limit int) ([]repository.TrendingTag, error) {
	cached, err := s.tagsCache.LookupSingle(ctx, trendingTagsCacheKey(limit), nil)
	if err == nil && cached != nil {
		return *cached, nil
	}

	tags, err := s.repo.GetTrendingTags(limit)
	if err != nil {
		return nil, err
	}
	s.tagsCache.WritedownSingleAsync(ctx, trendingTagsCacheKey(limit), &tags, recommendationCacheTTL)
	return tags, nil
}
