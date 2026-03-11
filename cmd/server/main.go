package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourname/go-blog/internal/config"
	"github.com/yourname/go-blog/internal/pkg/cache"
	"github.com/yourname/go-blog/internal/pkg/cachemanager"
	"github.com/yourname/go-blog/internal/pkg/database"
	"github.com/yourname/go-blog/internal/pkg/slogx"
	"github.com/yourname/go-blog/internal/router"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	logger, err := newLogger(cfg.App.Mode)
	if err != nil {
		panic(err)
	}
	defer func() { _ = logger.Sync() }()

	// 缓存/数据访问层经 slog 上报日志,统一桥接进 zap
	slogx.SetLogger(slog.New(slogx.NewZapHandler(logger)))

	if cfg.App.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	ctx := context.Background()
	db, err := database.NewMySQL(cfg.Database)
	if err != nil {
		logger.Fatal("mysql initialization failed", zap.Error(err))
	}
	if err := database.EnsureRuntimeSchema(db); err != nil {
		logger.Fatal("database schema initialization failed", zap.Error(err))
	}

	redisClient, err := cache.NewRedis(ctx, cfg.Redis)
	if err != nil {
		logger.Fatal("redis initialization failed", zap.Error(err))
	}
	defer redisClient.Close()

	// 注入全局句柄,供未做实例级注入的 ServiceManager 回退使用
	cachemanager.SetGlobalDB(db)
	cachemanager.SetGlobalRedis(redisClient)

	engine, shutdownCache := router.New(cfg, db, redisClient, logger)
	server := &http.Server{
		Addr:         cfg.App.Addr(),
		Handler:      engine,
		ReadTimeout:  cfg.App.ReadTimeout,
		WriteTimeout: cfg.App.WriteTimeout,
	}

	go func() {
		logger.Info("server started", zap.String("addr", cfg.App.Addr()))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("server stopped unexpectedly", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logger.Info("server shutting down")
	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("server shutdown failed", zap.Error(err))
	}
	// 排空缓存管理器的异步回填队列,避免丢失在途写入
	shutdownCache()
	logger.Info("server exited")
}

func newLogger(mode string) (*zap.Logger, error) {
	if mode == "release" {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}
