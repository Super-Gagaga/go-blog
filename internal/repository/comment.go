package repository

import (
	"errors"
	"strings"

	"github.com/yourname/go-blog/internal/model"
	"gorm.io/gorm"
)

type CommentRepository struct {
	db *gorm.DB
}

type AdminCommentFilter struct {
	Keyword   string
	UserID    int64
	ArticleID int64
	DateFrom  string
	DateTo    string
}

func NewCommentRepository(db *gorm.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) Create(comment *model.Comment) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(comment).Error; err != nil {
			return err
		}
		return tx.Model(&model.Article{}).
			Where("id = ?", comment.ArticleID).
			UpdateColumn("comment_count", gorm.Expr("comment_count + ?", 1)).
			Error
	})
}

func (r *CommentRepository) FindByID(id int64) (*model.Comment, error) {
	var comment model.Comment
	err := r.db.Preload("User").First(&comment, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &comment, err
}

func (r *CommentRepository) FindByArticleID(articleID int64) ([]model.Comment, error) {
	var comments []model.Comment
	err := r.db.
		Unscoped().
		Preload("User").
		Where("article_id = ?", articleID).
		Order("created_at ASC").
		Find(&comments).Error
	return comments, err
}

func (r *CommentRepository) SoftDelete(commentID int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var comment model.Comment
		if err := tx.First(&comment, commentID).Error; err != nil {
			return err
		}
		if err := tx.Delete(&comment).Error; err != nil {
			return err
		}
		return tx.Model(&model.Article{}).
			Where("id = ? AND comment_count > 0", comment.ArticleID).
			UpdateColumn("comment_count", gorm.Expr("comment_count - ?", 1)).
			Error
	})
}

func (r *CommentRepository) ListAll(page, pageSize int, filter AdminCommentFilter) ([]model.Comment, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var total int64
	if err := applyAdminCommentFilters(r.db.Model(&model.Comment{}), filter).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var comments []model.Comment
	err := applyAdminCommentFilters(r.db.Model(&model.Comment{}), filter).
		Preload("User").
		Preload("Article").
		Order("created_at DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&comments).Error
	return comments, total, err
}

func (r *CommentRepository) ForceDelete(id int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var comment model.Comment
		if err := tx.First(&comment, id).Error; err != nil {
			return err
		}

		var childCount int64
		if err := tx.Model(&model.Comment{}).Where("parent_id = ?", id).Count(&childCount).Error; err != nil {
			return err
		}

		if childCount > 0 {
			return tx.Model(&model.Comment{}).Where("id = ?", id).
				Update("content", "[该评论已被管理员删除]").Error
		}

		if err := tx.Unscoped().Delete(&comment).Error; err != nil {
			return err
		}
		return tx.Model(&model.Article{}).
			Where("id = ? AND comment_count > 0", comment.ArticleID).
			UpdateColumn("comment_count", gorm.Expr("comment_count - ?", 1)).
			Error
	})
}

func applyAdminCommentFilters(query *gorm.DB, filter AdminCommentFilter) *gorm.DB {
	if strings.TrimSpace(filter.Keyword) != "" {
		query = query.Where("comments.content LIKE ?", "%"+strings.TrimSpace(filter.Keyword)+"%")
	}
	if filter.UserID > 0 {
		query = query.Where("comments.user_id = ?", filter.UserID)
	}
	if filter.ArticleID > 0 {
		query = query.Where("comments.article_id = ?", filter.ArticleID)
	}
	if strings.TrimSpace(filter.DateFrom) != "" {
		query = query.Where("comments.created_at >= ?", strings.TrimSpace(filter.DateFrom))
	}
	if strings.TrimSpace(filter.DateTo) != "" {
		query = query.Where("comments.created_at <= ?", strings.TrimSpace(filter.DateTo))
	}
	return query
}
