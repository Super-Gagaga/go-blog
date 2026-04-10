package model

import "time"

type Tag struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:50;not null;uniqueIndex" json:"name"`
	CreatedAt time.Time `json:"created_at"`
}
