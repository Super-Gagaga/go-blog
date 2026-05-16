package repository

import (
	"github.com/yourname/go-blog/internal/model"
	"gorm.io/gorm"
)

type StatsRepository struct {
	db *gorm.DB
}

func NewStatsRepository(db *gorm.DB) *StatsRepository {
	return &StatsRepository{db: db}
}

func (r *StatsRepository) CountUsers() (int64, error) {
	var count int64
	err := r.db.Model(&model.User{}).Count(&count).Error
	return count, err
}

func (r *StatsRepository) CountUsersByRole(role string) (int64, error) {
	var count int64
	err := r.db.Model(&model.User{}).Where("role = ?", role).Count(&count).Error
	return count, err
}

func (r *StatsRepository) CountArticles(status string) (int64, error) {
	var count int64
	query := r.db.Model(&model.Article{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Count(&count).Error
	return count, err
}

func (r *StatsRepository) CountComments() (int64, error) {
	var count int64
	err := r.db.Model(&model.Comment{}).Count(&count).Error
	return count, err
}

func (r *StatsRepository) CountLikes() (int64, error) {
	var count int64
	err := r.db.Table("likes").Count(&count).Error
	return count, err
}

func (r *StatsRepository) CountFavorites() (int64, error) {
	var count int64
	err := r.db.Table("favorites").Count(&count).Error
	return count, err
}

func (r *StatsRepository) CountFollows() (int64, error) {
	var count int64
	err := r.db.Table("follows").Count(&count).Error
	return count, err
}

func (r *StatsRepository) RecentUsers(limit int) ([]model.User, error) {
	if limit < 1 {
		limit = 10
	}
	var users []model.User
	err := r.db.Order("created_at DESC").Limit(limit).Find(&users).Error
	return users, err
}

func (r *StatsRepository) RecentArticles(limit int) ([]model.Article, error) {
	if limit < 1 {
		limit = 10
	}
	var articles []model.Article
	err := r.db.Preload("User").Preload("Category").Preload("Tags").
		Order("created_at DESC").Limit(limit).Find(&articles).Error
	return articles, err
}

func (r *StatsRepository) TopAuthors(limit int) ([]model.AuthorStat, error) {
	if limit < 1 {
		limit = 10
	}
	var authors []model.AuthorStat
	err := r.db.Model(&model.User{}).
		Select("id AS user_id, username, avatar, article_count").
		Where("article_count > 0").
		Order("article_count DESC").
		Limit(limit).
		Scan(&authors).Error
	return authors, err
}
