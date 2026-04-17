package model

import "time"

type Like struct {
	UserID    int64     `gorm:"primaryKey" json:"user_id"`
	ArticleID int64     `gorm:"primaryKey" json:"article_id"`
	CreatedAt time.Time `json:"created_at"`
}
