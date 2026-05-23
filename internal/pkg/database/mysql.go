package database

import (
	"fmt"
	"time"

	"github.com/yourname/go-blog/internal/config"
	"github.com/yourname/go-blog/internal/pkg/cachemanager"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewMySQL(cfg config.DatabaseConfig) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{
		// slog 实现的 GORM logger,慢查询/错误经 slogx 汇入应用统一的 zap 输出
		Logger: cachemanager.NewSlogGormLogger(200 * time.Millisecond),
	})
	if err != nil {
		return nil, fmt.Errorf("connect mysql: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql db: %w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping mysql: %w", err)
	}

	if cfg.ConnMaxLifetime == 0 {
		sqlDB.SetConnMaxLifetime(time.Hour)
	}

	return db, nil
}
