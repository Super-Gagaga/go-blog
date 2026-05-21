package repository

import (
	"errors"

	"github.com/yourname/go-blog/internal/model"
	"gorm.io/gorm"
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
		var err error
		liked, err = toggleInteraction(
			tx,
			func() interface{} { return &model.Like{UserID: userID, ArticleID: articleID} },
			"like_count",
			userID,
			articleID,
		)
		return err
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
