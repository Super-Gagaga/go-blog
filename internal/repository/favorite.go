package repository

import (
	"errors"

	"github.com/yourname/go-blog/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type FavoriteRepository struct {
	db *gorm.DB
}

func NewFavoriteRepository(db *gorm.DB) *FavoriteRepository {
	return &FavoriteRepository{db: db}
}

func (r *FavoriteRepository) Exists(userID, articleID int64) (bool, error) {
	var favorite model.Favorite
	err := r.db.Where("user_id = ? AND article_id = ?", userID, articleID).First(&favorite).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return err == nil, err
}

func (r *FavoriteRepository) Toggle(userID, articleID int64) (bool, error) {
	var favorited bool
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var favorite model.Favorite
		err := tx.Where("user_id = ? AND article_id = ?", userID, articleID).First(&favorite).Error
		switch {
		case err == nil:
			if err := tx.Delete(&favorite).Error; err != nil {
				return err
			}
			favorited = false
			return tx.Model(&model.Article{}).
				Where("id = ? AND favorite_count > 0", articleID).
				UpdateColumn("favorite_count", gorm.Expr("favorite_count - ?", 1)).
				Error
		case errors.Is(err, gorm.ErrRecordNotFound):
			// 并发双击时复合主键冲突:忽略插入冲突并保持已收藏,避免重复计数
			result := tx.Clauses(clause.OnConflict{DoNothing: true}).
				Create(&model.Favorite{UserID: userID, ArticleID: articleID})
			if result.Error != nil {
				return result.Error
			}
			favorited = true
			if result.RowsAffected == 0 {
				return nil
			}
			return tx.Model(&model.Article{}).
				Where("id = ?", articleID).
				UpdateColumn("favorite_count", gorm.Expr("favorite_count + ?", 1)).
				Error
		default:
			return err
		}
	})
	return favorited, err
}

func (r *FavoriteRepository) ListByUser(userID int64, page, pageSize int) ([]model.Favorite, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	query := r.db.Model(&model.Favorite{}).Where("user_id = ?", userID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var favorites []model.Favorite
	err := r.db.
		Preload("Article.User").
		Preload("Article.Category").
		Preload("Article.Tags").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&favorites).Error
	if err != nil {
		return nil, 0, err
	}
	return favorites, total, nil
}
