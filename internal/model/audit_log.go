package model

import "time"

type AuditLog struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	AdminID    int64     `gorm:"index;not null" json:"admin_id"`
	Action     string    `gorm:"size:80;index;not null" json:"action"`
	TargetType string    `gorm:"size:40;index" json:"target_type"`
	TargetID   int64     `gorm:"index" json:"target_id"`
	Detail     string    `gorm:"type:text" json:"detail"`
	IP         string    `gorm:"size:64" json:"ip"`
	Admin      User      `gorm:"foreignKey:AdminID" json:"admin,omitempty"`
	CreatedAt  time.Time `gorm:"index" json:"created_at"`
}
