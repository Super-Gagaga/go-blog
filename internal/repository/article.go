package repository

import (
	"errors"
	"strings"

	"github.com/yourname/go-blog/internal/model"
	"gorm.io/gorm"
)

type ArticleRepository struct {
	db *gorm.DB
}

type ListArticleFilter struct {
	Page       int
	PageSize   int
	Status     string
	CategoryID int64
	TagID      int64
	Keyword    string
	UserID     int64
}

type AdminArticleFilter struct {
	Keyword    string
	Status     string
	UserID     int64
	CategoryID int64
	DateFrom   string
	DateTo     string
}

func NewArticleRepository(db *gorm.DB) *ArticleRepository {
	return &ArticleRepository{db: db}
}

func (r *ArticleRepository) Create(article *model.Article, tags []model.Tag) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(article).Error; err != nil {
			return err
		}
		if article.Status == model.ArticleStatusPublished {
			if err := tx.Model(&model.User{}).Where("id = ?", article.UserID).UpdateColumn("article_count", gorm.Expr("article_count + ?", 1)).Error; err != nil {
				return err
			}
		}
		if len(tags) > 0 {
			if err := tx.Model(article).Association("Tags").Replace(tags); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *ArticleRepository) Update(article *model.Article, tags []model.Tag) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var old model.Article
		if err := tx.Select("id", "user_id", "status").First(&old, article.ID).Error; err != nil {
			return err
		}
		if err := tx.Save(article).Error; err != nil {
			return err
		}
		if old.Status != article.Status {
			switch {
			case old.Status != model.ArticleStatusPublished && article.Status == model.ArticleStatusPublished:
				if err := tx.Model(&model.User{}).Where("id = ?", article.UserID).UpdateColumn("article_count", gorm.Expr("article_count + ?", 1)).Error; err != nil {
					return err
				}
			case old.Status == model.ArticleStatusPublished && article.Status != model.ArticleStatusPublished:
				if err := tx.Model(&model.User{}).Where("id = ? AND article_count > 0", article.UserID).UpdateColumn("article_count", gorm.Expr("article_count - ?", 1)).Error; err != nil {
					return err
				}
			}
		}
		if err := tx.Model(article).Association("Tags").Replace(tags); err != nil {
			return err
		}
		return nil
	})
}

func (r *ArticleRepository) SoftDelete(id int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var article model.Article
		if err := tx.Select("id", "user_id", "status").First(&article, id).Error; err != nil {
			return err
		}
		if err := tx.Delete(&article).Error; err != nil {
			return err
		}
		if article.Status == model.ArticleStatusPublished {
			return tx.Model(&model.User{}).
				Where("id = ? AND article_count > 0", article.UserID).
				UpdateColumn("article_count", gorm.Expr("article_count - ?", 1)).
				Error
		}
		return nil
	})
}

func (r *ArticleRepository) FindByID(id int64) (*model.Article, error) {
	var article model.Article
	err := r.db.
		Preload("User").
		Preload("Category").
		Preload("Tags").
		First(&article, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &article, err
}

func (r *ArticleRepository) FindBySlug(slug string) (*model.Article, error) {
	var article model.Article
	err := r.db.Where("slug = ?", slug).First(&article).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &article, err
}

func (r *ArticleRepository) List(filter ListArticleFilter) ([]model.Article, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 10
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}

	var total int64
	countQuery := applyArticleFilters(r.db.Model(&model.Article{}), filter)
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var articles []model.Article
	listQuery := applyArticleFilters(r.db.Model(&model.Article{}), filter)
	err := listQuery.
		Preload("User").
		Preload("Category").
		Preload("Tags").
		Order("created_at DESC").
		Limit(filter.PageSize).
		Offset((filter.Page - 1) * filter.PageSize).
		Find(&articles).Error
	if err != nil {
		return nil, 0, err
	}

	return articles, total, nil
}

func (r *ArticleRepository) IncrementViewCount(id int64) error {
	return r.db.Model(&model.Article{}).
		Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + ?", 1)).
		Error
}

func (r *ArticleRepository) SlugExists(slug string, excludeID int64) (bool, error) {
	query := r.db.Model(&model.Article{}).Where("slug = ?", slug)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *ArticleRepository) ChangeStatus(id int64, status string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var article model.Article
		if err := tx.Select("id", "user_id", "status").First(&article, id).Error; err != nil {
			return err
		}
		if article.Status == status {
			return nil
		}
		updates := map[string]interface{}{"status": status}
		if status == model.ArticleStatusPublished {
			updates["review_comment"] = nil
		}
		if err := tx.Model(&model.Article{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			return err
		}
		switch {
		case article.Status != model.ArticleStatusPublished && status == model.ArticleStatusPublished:
			return tx.Model(&model.User{}).Where("id = ?", article.UserID).UpdateColumn("article_count", gorm.Expr("article_count + ?", 1)).Error
		case article.Status == model.ArticleStatusPublished && status != model.ArticleStatusPublished:
			return tx.Model(&model.User{}).Where("id = ? AND article_count > 0", article.UserID).UpdateColumn("article_count", gorm.Expr("article_count - ?", 1)).Error
		default:
			return nil
		}
	})
}

func (r *ArticleRepository) ChangeStatusWithReviewComment(id int64, status string, reviewComment *string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var article model.Article
		if err := tx.Select("id", "user_id", "status").First(&article, id).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Article{}).Where("id = ?", id).
			Updates(map[string]interface{}{"status": status, "review_comment": reviewComment}).Error; err != nil {
			return err
		}
		switch {
		case article.Status != model.ArticleStatusPublished && status == model.ArticleStatusPublished:
			return tx.Model(&model.User{}).Where("id = ?", article.UserID).UpdateColumn("article_count", gorm.Expr("article_count + ?", 1)).Error
		case article.Status == model.ArticleStatusPublished && status != model.ArticleStatusPublished:
			return tx.Model(&model.User{}).Where("id = ? AND article_count > 0", article.UserID).UpdateColumn("article_count", gorm.Expr("article_count - ?", 1)).Error
		default:
			return nil
		}
	})
}

func (r *ArticleRepository) ListPending(page, pageSize int, keyword string) ([]model.Article, int64, error) {
	return r.ListAll(page, pageSize, AdminArticleFilter{
		Status:  model.ArticleStatusPendingReview,
		Keyword: keyword,
	}, "created_at ASC")
}

func (r *ArticleRepository) ListAll(page, pageSize int, filter AdminArticleFilter, order ...string) ([]model.Article, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	sortOrder := "created_at DESC"
	if len(order) > 0 && order[0] != "" {
		sortOrder = order[0]
	}

	var total int64
	if err := applyAdminArticleFilters(r.db.Model(&model.Article{}), filter).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var articles []model.Article
	err := applyAdminArticleFilters(r.db.Model(&model.Article{}), filter).
		Preload("User").
		Preload("Category").
		Preload("Tags").
		Order(sortOrder).
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&articles).Error
	return articles, total, err
}

func (r *ArticleRepository) ForceDelete(id int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var article model.Article
		if err := tx.Unscoped().Select("id", "user_id", "status").First(&article, id).Error; err != nil {
			return err
		}
		if err := tx.Model(&article).Association("Tags").Clear(); err != nil {
			return err
		}
		if err := tx.Unscoped().Delete(&article).Error; err != nil {
			return err
		}
		if article.Status == model.ArticleStatusPublished {
			return tx.Model(&model.User{}).
				Where("id = ? AND article_count > 0", article.UserID).
				UpdateColumn("article_count", gorm.Expr("article_count - ?", 1)).
				Error
		}
		return nil
	})
}

func (r *ArticleRepository) Ranking(period string, limit int) ([]model.Article, error) {
	if limit < 1 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	var sinceClause string
	switch period {
	case "day":
		sinceClause = "AND articles.created_at >= DATE_SUB(NOW(), INTERVAL 1 DAY)"
	case "month":
		sinceClause = "AND articles.created_at >= DATE_SUB(NOW(), INTERVAL 1 MONTH)"
	default: // week
		sinceClause = "AND articles.created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)"
	}

	var articles []model.Article
	err := r.db.
		Preload("User").
		Preload("Category").
		Preload("Tags").
		Select(`articles.*,
			(articles.view_count * 0.3 + articles.like_count * 2 +
			 articles.favorite_count * 3 + articles.comment_count * 2.5) /
			POWER(GREATEST(TIMESTAMPDIFF(SECOND, articles.created_at, NOW()) / 3600 + 2, 0.001), 1.5)
			AS hotness_score`).
		Where("articles.status = ? "+sinceClause, model.ArticleStatusPublished).
		Order("hotness_score DESC").
		Limit(limit).
		Find(&articles).Error
	if err != nil {
		return nil, err
	}
	return articles, nil
}

func applyArticleFilters(query *gorm.DB, filter ListArticleFilter) *gorm.DB {
	if filter.Status != "" {
		query = query.Where("articles.status = ?", filter.Status)
	}
	if filter.CategoryID > 0 {
		query = query.Where("articles.category_id = ?", filter.CategoryID)
	}
	if filter.UserID > 0 {
		query = query.Where("articles.user_id = ?", filter.UserID)
	}
	if strings.TrimSpace(filter.Keyword) != "" {
		like := "%" + strings.TrimSpace(filter.Keyword) + "%"
		query = query.Where("articles.title LIKE ? OR articles.content LIKE ? OR articles.summary LIKE ?", like, like, like)
	}
	if filter.TagID > 0 {
		query = query.Joins("JOIN article_tags ON article_tags.article_id = articles.id").
			Where("article_tags.tag_id = ?", filter.TagID)
	}
	return query
}

func applyAdminArticleFilters(query *gorm.DB, filter AdminArticleFilter) *gorm.DB {
	if filter.Status != "" {
		query = query.Where("articles.status = ?", filter.Status)
	}
	if filter.CategoryID > 0 {
		query = query.Where("articles.category_id = ?", filter.CategoryID)
	}
	if filter.UserID > 0 {
		query = query.Where("articles.user_id = ?", filter.UserID)
	}
	if strings.TrimSpace(filter.Keyword) != "" {
		like := "%" + strings.TrimSpace(filter.Keyword) + "%"
		query = query.Where("articles.title LIKE ? OR articles.content LIKE ?", like, like)
	}
	if strings.TrimSpace(filter.DateFrom) != "" {
		query = query.Where("articles.created_at >= ?", strings.TrimSpace(filter.DateFrom))
	}
	if strings.TrimSpace(filter.DateTo) != "" {
		query = query.Where("articles.created_at <= ?", strings.TrimSpace(filter.DateTo))
	}
	return query
}
