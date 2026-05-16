package repository

import (
	"strings"

	"github.com/yourname/go-blog/internal/model"
	"gorm.io/gorm"
)

type AuditLogRepository struct {
	db *gorm.DB
}

type AuditLogFilter struct {
	AdminID    int64
	Action     string
	TargetType string
	DateFrom   string
	DateTo     string
}

func NewAuditLogRepository(db *gorm.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

func (r *AuditLogRepository) Create(log *model.AuditLog) error {
	return r.db.Create(log).Error
}

func (r *AuditLogRepository) List(page, pageSize int, filter AuditLogFilter) ([]model.AuditLog, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	query := applyAuditLogFilters(r.db.Model(&model.AuditLog{}), filter)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var logs []model.AuditLog
	err := applyAuditLogFilters(r.db.Model(&model.AuditLog{}), filter).
		Preload("Admin").
		Order("created_at DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&logs).Error
	return logs, total, err
}

func applyAuditLogFilters(query *gorm.DB, filter AuditLogFilter) *gorm.DB {
	if filter.AdminID > 0 {
		query = query.Where("admin_id = ?", filter.AdminID)
	}
	if strings.TrimSpace(filter.Action) != "" {
		query = query.Where("action = ?", strings.TrimSpace(filter.Action))
	}
	if strings.TrimSpace(filter.TargetType) != "" {
		query = query.Where("target_type = ?", strings.TrimSpace(filter.TargetType))
	}
	if strings.TrimSpace(filter.DateFrom) != "" {
		query = query.Where("created_at >= ?", strings.TrimSpace(filter.DateFrom))
	}
	if strings.TrimSpace(filter.DateTo) != "" {
		query = query.Where("created_at <= ?", strings.TrimSpace(filter.DateTo))
	}
	return query
}
