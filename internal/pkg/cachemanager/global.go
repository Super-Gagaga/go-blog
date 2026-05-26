package cachemanager

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// 全局 DB/Redis 句柄,由宿主应用在启动时通过 SetGlobalDB / SetGlobalRedis
// 注入(应用侧通常用配置驱动工厂构建实例后注入);ServiceManager 未通过
// WithDB/WithRedis 实例级注入时回退到这里。传 nil 可用于测试清理。
var (
	globalDB    *gorm.DB
	globalRedis *redis.Client
)

// SetGlobalDB 注入全局数据库句柄。
func SetGlobalDB(db *gorm.DB) {
	globalDB = db
}

// GetDB 返回全局数据库句柄,未注入时 panic(属于装配错误,应尽早暴露)。
func GetDB() *gorm.DB {
	if globalDB == nil {
		panic("cachemanager: global DB not initialized, call SetGlobalDB first")
	}
	return globalDB
}

// SetGlobalRedis 注入全局 Redis 客户端。
func SetGlobalRedis(client *redis.Client) {
	globalRedis = client
}

// GetRedis 返回全局 Redis 客户端,未注入时 panic。
func GetRedis() *redis.Client {
	if globalRedis == nil {
		panic("cachemanager: global redis not initialized, call SetGlobalRedis first")
	}
	return globalRedis
}

// ScanKeys 使用游标 SCAN 安全遍历匹配 pattern 的全部键(避免 KEYS 阻塞)。
func ScanKeys(ctx context.Context, client *redis.Client, pattern string, batchSize int64) ([]string, error) {
	var allKeys []string
	var cursor uint64
	for {
		keys, nextCursor, err := client.Scan(ctx, cursor, pattern, batchSize).Result()
		if err != nil {
			return nil, fmt.Errorf("scan keys failed: %w", err)
		}
		allKeys = append(allKeys, keys...)
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return allKeys, nil
}
