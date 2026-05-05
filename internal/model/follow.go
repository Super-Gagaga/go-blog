package model

import "time"

type Follow struct {
	FollowerID int64     `gorm:"primaryKey" json:"follower_id"`
	FolloweeID int64     `gorm:"primaryKey" json:"followee_id"`
	Follower   User      `gorm:"foreignKey:FollowerID" json:"follower,omitempty"`
	Followee   User      `gorm:"foreignKey:FolloweeID" json:"followee,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}
