package repository

import (
	"github.com/yourname/go-blog/internal/model"
	"gorm.io/gorm"
)

type FeedRepository struct {
	db *gorm.DB
}

func NewFeedRepository(db *gorm.DB) *FeedRepository {
	return &FeedRepository{db: db}
}

func (r *FeedRepository) UserFeed(userID int64, page, pageSize int) ([]model.Article, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	query := r.db.Model(&model.Article{}).
		Joins("JOIN follows ON follows.followee_id = articles.user_id").
		Where("follows.follower_id = ? AND articles.status = ?", userID, model.ArticleStatusPublished)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var articles []model.Article
	err := r.db.Model(&model.Article{}).
		Joins("JOIN follows ON follows.followee_id = articles.user_id").
		Where("follows.follower_id = ? AND articles.status = ?", userID, model.ArticleStatusPublished).
		Preload("User").
		Preload("Category").
		Preload("Tags").
		Order("articles.created_at DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&articles).Error
	if err != nil {
		return nil, 0, err
	}
	return articles, total, nil
}
