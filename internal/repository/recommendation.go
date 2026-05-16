package repository

import (
	"time"

	"github.com/yourname/go-blog/internal/model"
	"gorm.io/gorm"
)

type RecommendationRepository struct {
	db *gorm.DB
}

func NewRecommendationRepository(db *gorm.DB) *RecommendationRepository {
	return &RecommendationRepository{db: db}
}

// RecommendedAuthor represents a recommended author with metadata.
type RecommendedAuthor struct {
	ID            int64  `json:"id"`
	Username      string `json:"username"`
	Avatar        string `json:"avatar"`
	Bio           string `json:"bio"`
	ArticleCount  int64  `json:"article_count"`
	FollowerCount int64  `json:"follower_count"`
}

// TrendingTag represents a trending topic with hotness score.
type TrendingTag struct {
	TagID        int64   `json:"tag_id"`
	Name         string  `json:"name"`
	ArticleCount int64   `json:"article_count"`
	HotnessScore float64 `json:"hotness_score"`
}

// GetRecommendedAuthors returns authors sorted by follower count, article count, and recent activity.
// Only active users with at least one published article are included.
func (r *RecommendationRepository) GetRecommendedAuthors(limit int) ([]RecommendedAuthor, error) {
	if limit < 1 {
		limit = 5
	}
	if limit > 50 {
		limit = 50
	}

	var authors []RecommendedAuthor
	err := r.db.Model(&model.User{}).
		Select("users.id, users.username, users.avatar, users.bio, users.article_count, users.follower_count").
		Where("users.status = ?", "active").
		Where("users.article_count > 0").
		Order("users.follower_count DESC, users.article_count DESC").
		Limit(limit).
		Find(&authors).Error
	if err != nil {
		return nil, err
	}
	return authors, nil
}

// GetTrendingTags returns tags ordered by recent article activity (last 30 days weighted higher)
// combined with total article count.
func (r *RecommendationRepository) GetTrendingTags(limit int) ([]TrendingTag, error) {
	if limit < 1 {
		limit = 12
	}
	if limit > 50 {
		limit = 50
	}

	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

	var results []TrendingTag
	err := r.db.Table("tags").
		Select(`tags.id as tag_id,
			tags.name,
			COUNT(DISTINCT article_tags.article_id) as article_count,
			CAST(COUNT(DISTINCT CASE WHEN articles.created_at >= ? THEN article_tags.article_id END) AS FLOAT) * 0.7 +
			CAST(COUNT(DISTINCT article_tags.article_id) AS FLOAT) * 0.3 AS hotness_score`, thirtyDaysAgo).
		Joins("JOIN article_tags ON article_tags.tag_id = tags.id").
		Joins("JOIN articles ON articles.id = article_tags.article_id AND articles.status = ?", model.ArticleStatusPublished).
		Group("tags.id, tags.name").
		Order("hotness_score DESC").
		Limit(limit).
		Find(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}
