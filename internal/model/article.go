package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	ArticleStatusDraft         = "draft"
	ArticleStatusPendingReview = "pending_review"
	ArticleStatusPublished     = "published"
	ArticleStatusArchived      = "archived"
)

type Article struct {
	ID            int64          `gorm:"primaryKey" json:"id"`
	Title         string         `gorm:"size:255;not null" json:"title"`
	Slug          string         `gorm:"size:300;not null;uniqueIndex" json:"slug"`
	Content       string         `gorm:"type:text;not null" json:"content,omitempty"`
	ContentHTML   string         `gorm:"type:text;not null" json:"content_html,omitempty"`
	Summary       string         `gorm:"size:500" json:"summary"`
	CoverImage    string         `gorm:"size:500" json:"cover_image"`
	Status        string         `gorm:"size:20;not null;default:draft;index" json:"status"`
	ReviewComment *string        `gorm:"size:500" json:"review_comment,omitempty"`
	ViewCount     int64          `gorm:"not null;default:0" json:"view_count"`
	CommentCount  int64          `gorm:"not null;default:0" json:"comment_count"`
	LikeCount     int64          `gorm:"not null;default:0" json:"like_count"`
	FavoriteCount int64          `gorm:"not null;default:0" json:"favorite_count"`
	UserID        int64          `gorm:"not null;index" json:"user_id"`
	CategoryID    *int64         `gorm:"index" json:"category_id"`
	User          User           `json:"user,omitempty"`
	Category      *Category      `json:"category,omitempty"`
	Tags          []Tag          `gorm:"many2many:article_tags;" json:"tags,omitempty"`
	CreatedAt     time.Time      `gorm:"index" json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

type ArticleDetail struct {
	Article
}
