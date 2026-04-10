package model

import "time"

type Category struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	Slug      string    `gorm:"size:100;not null;uniqueIndex" json:"slug"`
	CreatedAt time.Time `json:"created_at"`
}
