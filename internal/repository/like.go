package repository

import (
	"errors"

	"github.com/yourname/go-blog/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type LikeRepository struct {
	db *gorm.DB
}

func NewLikeRepository(db *gorm.DB) *LikeRepository {
	return &LikeRepository{db: db}
}

func (r *LikeRepository) Exists(userID, articleID int64) (bool, error) {
	var like model.Like
	err := r.db.Where("user_id = ? AND article_id = ?", userID, articleID).First(&like).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return err == nil, err
}

func (r *LikeRepository) Toggle(userID, articleID int64) (bool, int64, error) {
	var liked bool
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var like model.Like
		err := tx.Where("user_id = ? AND article_id = ?", userID, articleID).First(&like).Error
		switch {
		case err == nil:
			if err := tx.Delete(&like).Error; err != nil {
				return err
			}
			liked = false
			return tx.Model(&model.Article{}).
				Where("id = ? AND like_count > 0", articleID).
				UpdateColumn("like_count", gorm.Expr("like_count - ?", 1)).
				Error
		case errors.Is(err, gorm.ErrRecordNotFound):
			// 并发双击时复合主键冲突:忽略插入冲突并保持已赞,避免重复计数
			result := tx.Clauses(clause.OnConflict{DoNothing: true}).
				Create(&model.Like{UserID: userID, ArticleID: articleID})
			if result.Error != nil {
				return result.Error
			}
			liked = true
			if result.RowsAffected == 0 {
				return nil
			}
			return tx.Model(&model.Article{}).
				Where("id = ?", articleID).
				UpdateColumn("like_count", gorm.Expr("like_count + ?", 1)).
				Error
		default:
			return err
		}
	})
	if err != nil {
		return false, 0, err
	}

	count, err := r.Count(articleID)
	return liked, count, err
}

func (r *LikeRepository) Count(articleID int64) (int64, error) {
	var count int64
	err := r.db.Model(&model.Like{}).Where("article_id = ?", articleID).Count(&count).Error
	return count, err
}
