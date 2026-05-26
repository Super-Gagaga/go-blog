package cachemanager

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisManager 是 Redis 客户端之上的 JSON 序列化便捷封装:
// 值自动 JSON 编解码,GetMultiple/SetMultiple 走 pipeline。
// 全局装配见 global.go(SetGlobalRedis/GetRedis)。
type RedisManager struct {
	Client *redis.Client
}

// NewRedisManager 基于已有客户端构建 JSON 包装器。
func NewRedisManager(client *redis.Client) *RedisManager {
	return &RedisManager{Client: client}
}

// GetRedisManager 返回该实例所用 Redis 客户端的 JSON 包装器。
func (sm *ServiceManager[T]) GetRedisManager() *RedisManager {
	return NewRedisManager(sm.Redis())
}

// Set 将对象 JSON 序列化后写入缓存
func (rm *RedisManager) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	return rm.Client.Set(ctx, key, data, expiration).Err()
}

// Get 读取缓存并 JSON 反序列化到 dest
func (rm *RedisManager) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := rm.Client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("failed to unmarshal cache data for key %s: %w", key, err)
	}
	return nil
}

// Delete 删除缓存键
func (rm *RedisManager) Delete(ctx context.Context, keys ...string) error {
	return rm.Client.Del(ctx, keys...).Err()
}

// Exists 判断键是否存在
func (rm *RedisManager) Exists(ctx context.Context, key string) (bool, error) {
	count, err := rm.Client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// SetMultiple 批量写入(pipeline),values 中的对象自动 JSON 序列化
func (rm *RedisManager) SetMultiple(ctx context.Context, values map[string]interface{}, expiration time.Duration) error {
	pipe := rm.Client.Pipeline()
	for key, value := range values {
		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value for key %s: %w", key, err)
		}
		pipe.Set(ctx, key, data, expiration)
	}
	_, err := pipe.Exec(ctx)
	return err
}

// GetMultiple 批量读取(pipeline)并 JSON 反序列化
func (rm *RedisManager) GetMultiple(ctx context.Context, keys []string) (map[string][]byte, error) {
	pipe := rm.Client.Pipeline()
	cmds := make([]*redis.StringCmd, len(keys))
	for i, key := range keys {
		cmds[i] = pipe.Get(ctx, key)
	}
	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	result := make(map[string][]byte, len(keys))
	for i, cmd := range cmds {
		data, err := cmd.Bytes()
		if err != nil {
			continue // 未命中的键跳过
		}
		result[keys[i]] = data
	}
	return result, nil
}
