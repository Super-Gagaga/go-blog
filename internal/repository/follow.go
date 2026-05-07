package repository

import (
	"errors"

	"github.com/yourname/go-blog/internal/model"
	"gorm.io/gorm"
)

type FollowRepository struct {
	db *gorm.DB
}

func NewFollowRepository(db *gorm.DB) *FollowRepository {
	return &FollowRepository{db: db}
}

func (r *FollowRepository) Exists(followerID, followeeID int64) (bool, error) {
	var follow model.Follow
	err := r.db.Where("follower_id = ? AND followee_id = ?", followerID, followeeID).First(&follow).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return err == nil, err
}

func (r *FollowRepository) Toggle(followerID, followeeID int64) (bool, error) {
	var following bool
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var follow model.Follow
		err := tx.Where("follower_id = ? AND followee_id = ?", followerID, followeeID).First(&follow).Error
		switch {
		case err == nil:
			if err := tx.Delete(&follow).Error; err != nil {
				return err
			}
			following = false
			if err := tx.Model(&model.User{}).Where("id = ? AND following_count > 0", followerID).UpdateColumn("following_count", gorm.Expr("following_count - ?", 1)).Error; err != nil {
				return err
			}
			return tx.Model(&model.User{}).Where("id = ? AND follower_count > 0", followeeID).UpdateColumn("follower_count", gorm.Expr("follower_count - ?", 1)).Error
		case errors.Is(err, gorm.ErrRecordNotFound):
			if err := tx.Create(&model.Follow{FollowerID: followerID, FolloweeID: followeeID}).Error; err != nil {
				return err
			}
			following = true
			if err := tx.Model(&model.User{}).Where("id = ?", followerID).UpdateColumn("following_count", gorm.Expr("following_count + ?", 1)).Error; err != nil {
				return err
			}
			return tx.Model(&model.User{}).Where("id = ?", followeeID).UpdateColumn("follower_count", gorm.Expr("follower_count + ?", 1)).Error
		default:
			return err
		}
	})
	return following, err
}

func (r *FollowRepository) ListFollowing(userID int64, page, pageSize int) ([]model.Follow, int64, error) {
	return r.list("follower_id = ?", userID, "Followee", page, pageSize)
}

func (r *FollowRepository) ListFollowers(userID int64, page, pageSize int) ([]model.Follow, int64, error) {
	return r.list("followee_id = ?", userID, "Follower", page, pageSize)
}

func (r *FollowRepository) list(where string, userID int64, preload string, page, pageSize int) ([]model.Follow, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var total int64
	if err := r.db.Model(&model.Follow{}).Where(where, userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var follows []model.Follow
	err := r.db.Preload(preload).
		Where(where, userID).
		Order("created_at DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&follows).Error
	if err != nil {
		return nil, 0, err
	}
	return follows, total, nil
}
