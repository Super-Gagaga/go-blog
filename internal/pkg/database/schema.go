package database

import (
	"fmt"

	"github.com/yourname/go-blog/internal/model"
	"gorm.io/gorm"
)

func EnsureRuntimeSchema(db *gorm.DB) error {
	if db == nil {
		return nil
	}

	migrator := db.Migrator()
	if migrator.HasTable(&model.Article{}) && !migrator.HasColumn(&model.Article{}, "ReviewComment") {
		if err := migrator.AddColumn(&model.Article{}, "ReviewComment"); err != nil {
			return fmt.Errorf("add articles.review_comment: %w", err)
		}
	}

	if !migrator.HasTable(&model.AuditLog{}) {
		if err := db.AutoMigrate(&model.AuditLog{}); err != nil {
			return fmt.Errorf("create audit_logs: %w", err)
		}
	}

	if migrator.HasTable(&model.User{}) {
		result := db.Exec(`
			UPDATE users
			SET role = 'admin'
			WHERE id = (
				SELECT id
				FROM (
					SELECT id
					FROM users
					ORDER BY created_at ASC, id ASC
					LIMIT 1
				) AS first_user
			)
			AND role <> 'admin'
		`)
		if result.Error != nil {
			return fmt.Errorf("promote first registered user: %w", result.Error)
		}
	}

	return nil
}
