package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID             int64          `gorm:"primaryKey" json:"id"`
	Username       string         `gorm:"size:50;not null;uniqueIndex" json:"username"`
	Email          string         `gorm:"size:255;not null;uniqueIndex" json:"email"`
	Password       string         `gorm:"size:255;not null" json:"-"`
	Avatar         string         `gorm:"size:500" json:"avatar"`
	Bio            string         `gorm:"type:text" json:"bio"`
	Role           string         `gorm:"size:20;not null;default:user" json:"role"`
	Status         string         `gorm:"size:20;not null;default:active" json:"status"`
	ArticleCount   int64          `gorm:"not null;default:0" json:"article_count"`
	FollowerCount  int64          `gorm:"not null;default:0" json:"follower_count"`
	FollowingCount int64          `gorm:"not null;default:0" json:"following_count"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

type UserProfile struct {
	ID             int64     `json:"id"`
	Username       string    `json:"username"`
	Email          string    `json:"email,omitempty"`
	Avatar         string    `json:"avatar"`
	Bio            string    `json:"bio"`
	Role           string    `json:"role,omitempty"`
	Status         string    `json:"status,omitempty"`
	ArticleCount   int64     `json:"article_count"`
	FollowerCount  int64     `json:"follower_count"`
	FollowingCount int64     `json:"following_count"`
	CreatedAt      time.Time `json:"created_at"`
}

func NewUserProfile(user *User, private bool) UserProfile {
	profile := UserProfile{
		ID:             user.ID,
		Username:       user.Username,
		Avatar:         user.Avatar,
		Bio:            user.Bio,
		ArticleCount:   user.ArticleCount,
		FollowerCount:  user.FollowerCount,
		FollowingCount: user.FollowingCount,
		CreatedAt:      user.CreatedAt,
	}
	if private {
		profile.Email = user.Email
		profile.Role = user.Role
		profile.Status = user.Status
	}
	return profile
}
