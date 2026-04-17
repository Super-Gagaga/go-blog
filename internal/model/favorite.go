package model

import "time"

type Favorite struct {
	UserID    int64     `gorm:"primaryKey" json:"user_id"`
	ArticleID int64     `gorm:"primaryKey" json:"article_id"`
	Article   Article   `json:"article,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
