package model

import "time"

const (
	NotificationTypeComment        = "comment"
	NotificationTypeReply          = "reply"
	NotificationTypeLike           = "like"
	NotificationTypeFollow         = "follow"
	NotificationTypeSystem         = "system"
	NotificationTypeReviewApproved = "review_approved"
	NotificationTypeReviewRejected = "review_rejected"
)

type Notification struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	UserID    int64     `gorm:"not null;index:idx_notifications_user_read" json:"user_id"`
	Type      string    `gorm:"size:30;not null" json:"type"`
	Content   string    `gorm:"type:json;not null" json:"content"`
	IsRead    bool      `gorm:"not null;default:false;index:idx_notifications_user_read" json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

// NotificationItem is used by the frontend to display enriched notification data.
type NotificationItem struct {
	ID        int64     `json:"id"`
	Type      string    `json:"type"`
	Content   string    `json:"content"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
	FromUser  *User     `json:"from_user,omitempty"`
	Article   *Article  `json:"article,omitempty"`
}
