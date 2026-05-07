package repository

import (
	"encoding/json"

	"github.com/yourname/go-blog/internal/model"
	"gorm.io/gorm"
)

type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(userID int64, nType string, content map[string]interface{}) error {
	payload, err := json.Marshal(content)
	if err != nil {
		return err
	}
	return r.db.Create(&model.Notification{
		UserID:  userID,
		Type:    nType,
		Content: string(payload),
	}).Error
}

func (r *NotificationRepository) ListByUser(userID int64, page, pageSize int) ([]model.Notification, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	query := r.db.Model(&model.Notification{}).Where("user_id = ?", userID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var notifications []model.Notification
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&notifications).Error
	if err != nil {
		return nil, 0, err
	}
	return notifications, total, nil
}

func (r *NotificationRepository) MarkAsRead(userID int64, ids []int64) error {
	query := r.db.Model(&model.Notification{}).Where("user_id = ?", userID)
	if len(ids) > 0 {
		query = query.Where("id IN ?", ids)
	}
	return query.Update("is_read", true).Error
}

func (r *NotificationRepository) CountUnread(userID int64) (int64, error) {
	var count int64
	err := r.db.Model(&model.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error
	return count, err
}
