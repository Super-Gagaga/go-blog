package model

import (
	"time"

	"gorm.io/gorm"
)

type Comment struct {
	ID        int64          `gorm:"primaryKey" json:"id"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	ArticleID int64          `gorm:"not null;index" json:"article_id"`
	UserID    int64          `gorm:"not null;index" json:"user_id"`
	ParentID  *int64         `gorm:"index" json:"parent_id"`
	User      User           `json:"user,omitempty"`
	Article   Article        `json:"article,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type CommentTreeNode struct {
	Comment
	Replies []CommentTreeNode `json:"replies"`
}
