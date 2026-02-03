// Package redis 基础操作实现
// 提供 Redis 的基础键值操作，包括 Get、Set、Delete 等常用方法
package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Get 从 Redis 中获取指定键的值。
//
// 参数：
//   - ctx: 上下文，用于控制超时和取消操作
//   - key: 缓存键名（不包含前缀，前缀会自动添加）
//
// 返回值：
//   - []byte: 缓存值的字节数组
//   - error: 操作失败时返回错误
//
// 特殊情况：
//   - 如果键不存在，返回 (nil, nil)，即值为 nil 但没有错误
//   - 如果 Redis 缓存已关闭，返回 error
//
// 使用示例：
//
//	value, err := cache.Get(ctx, "user:1")
//	if err != nil {
//	    log.Error("获取缓存失败", err)
//	    return err
//	}
//	if value == nil {
//	    log.Debug("缓存未命中")
//	    // 从数据库加载...
//	}
func (r *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	if key == "" {
		return nil, fmt.Errorf("缓存键不能为空")
	}

	client, err := r.getUniversalClient()
	if err != nil {
		return nil, err
	}

	fullKey := r.buildKey(key)
	result, err := client.Get(ctx, fullKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // 键不存在时返回nil而不是错误
		}
		return nil, fmt.Errorf("redis get error: %w", err)
	}
	return []byte(result), nil
}

// GetString 获取缓存值（字符串）
// 从Redis中获取指定键的字符串值
// 参数:
//   - ctx: 上下文，用于控制请求超时和取消
//   - key: 缓存键名（不包含前缀）
//
// 返回:
//   - string: 缓存值的字符串，键不存在时返回空字符串
//   - error: 获取失败时返回错误，键不存在不算错误
func (r *RedisCache) GetString(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("缓存键不能为空")
	}

	client, err := r.getUniversalClient()
	if err != nil {
		return "", err
	}

	fullKey := r.buildKey(key)
	result, err := client.Get(ctx, fullKey).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil // 键不存在时返回空字符串而不是错误
		}
		return "", fmt.Errorf("redis get error: %w", err)
	}
	return result, nil
}

// Set 在 Redis 中设置指定键的值。
//
// 参数：
//   - ctx: 上下文，用于控制超时和取消操作
//   - key: 缓存键名（不包含前缀，前缀会自动添加）
//   - value: 要缓存的值（字节数组）
//   - expiration: 过期时间
//   - 0: 使用配置文件中的默认过期时间
//   - 正数: 指定的过期时间（如 10*time.Minute）
//   - 负数: 永不过期
//
// 返回值：
//   - error: 设置失败时返回错误
//
// 使用示例：
//
//	// 使用默认过期时间
//	err := cache.Set(ctx, "key", []byte("value"), 0)
//
//	// 设置 10 分钟过期
//	err := cache.Set(ctx, "session:123", sessionData, 10*time.Minute)
//
//	// 永不过期
//	err := cache.Set(ctx, "config", configData, -1)
func (r *RedisCache) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	if key == "" {
		return fmt.Errorf("缓存键不能为空")
	}
	if value == nil {
		return fmt.Errorf("缓存值不能为nil")
	}

	client, err := r.getUniversalClient()
	if err != nil {
		return err
	}

	// 处理过期时间：0表示使用默认过期时间，负数表示永不过期
	finalExpiration := r.resolveExpiration(expiration)

	fullKey := r.buildKey(key)
	err = client.Set(ctx, fullKey, value, finalExpiration).Err()
	if err != nil {
		return fmt.Errorf("redis set error: %w", err)
	}
	return nil
}

// SetString 设置缓存值（字符串）
// 在Redis中设置指定键的字符串值，支持过期时间
// 参数:
//   - ctx: 上下文，用于控制请求超时和取消
//   - key: 缓存键名（不包含前缀）
//   - value: 要缓存的字符串值
//   - expiration: 过期时间，0表示使用配置的默认过期时间，负数表示永不过期
//
// 返回:
//   - error: 设置失败时返回错误信息
func (r *RedisCache) SetString(ctx context.Context, key string, value string, expiration time.Duration) error {
	if key == "" {
		return fmt.Errorf("缓存键不能为空")
	}

	client, err := r.getUniversalClient()
	if err != nil {
		return err
	}

	// 处理过期时间：0表示使用默认过期时间，负数表示永不过期
	finalExpiration := r.resolveExpiration(expiration)

	fullKey := r.buildKey(key)
	err = client.Set(ctx, fullKey, value, finalExpiration).Err()
	if err != nil {
		return fmt.Errorf("redis set error: %w", err)
	}
	return nil
}

// Delete 从 Redis 中删除指定的键。
//
// 参数：
//   - ctx: 上下文，用于控制超时和取消操作
//   - key: 要删除的缓存键名（不包含前缀）
//
// 返回值：
//   - error: 删除失败时返回错误
//
// 特性：
//   - 删除不存在的键不会返回错误（幂等操作）
//   - 操作是原子的
//
// 使用示例：
//
//	err := cache.Delete(ctx, "session:123")
//	if err != nil {
//	    log.Error("删除缓存失败", err)
//	}
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	if key == "" {
		return fmt.Errorf("缓存键不能为空")
	}

	client, err := r.getUniversalClient()
	if err != nil {
		return err
	}

	fullKey := r.buildKey(key)
	err = client.Del(ctx, fullKey).Err()
	if err != nil {
		return fmt.Errorf("redis delete error: %w", err)
	}
	return nil
}

// Exists 检查指定的键是否存在。
//
// 参数：
//   - ctx: 上下文，用于控制超时和取消操作
//   - key: 要检查的缓存键名（不包含前缀）
//
// 返回值：
//   - bool: true 表示键存在，false 表示键不存在
//   - error: 操作失败时返回错误
//
// 使用示例：
//
//	exists, err := cache.Exists(ctx, "user:1")
//	if err != nil {
//	    return err
//	}
//	if !exists {
//	    log.Debug("缓存不存在")
//	}
func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	if key == "" {
		return false, fmt.Errorf("缓存键不能为空")
	}

	client, err := r.getUniversalClient()
	if err != nil {
		return false, err
	}

	fullKey := r.buildKey(key)
	result, err := client.Exists(ctx, fullKey).Result()
	if err != nil {
		return false, fmt.Errorf("redis exists error: %w", err)
	}
	return result > 0, nil
}

// Keys 获取匹配模式的所有键（谨慎使用）
func (r *RedisCache) Keys(ctx context.Context, pattern string) ([]string, error) {
	if pattern == "" {
		return nil, fmt.Errorf("匹配模式不能为空")
	}

	client, err := r.getUniversalClient()
	if err != nil {
		return nil, err
	}

	// 如果有前缀，需要在模式前加上前缀
	fullPattern := pattern
	if r.keyPrefix != "" {
		fullPattern = r.buildKey(pattern)
	}

	keys, err := client.Keys(ctx, fullPattern).Result()
	if err != nil {
		return nil, fmt.Errorf("redis keys error: %w", err)
	}

	// 如果有前缀，需要去掉前缀
	if r.keyPrefix != "" {
		for i, key := range keys {
			keys[i] = r.parseKey(key)
		}
	}

	return keys, nil
}

// Size 获取缓存中键的数量
func (r *RedisCache) Size(ctx context.Context) (int64, error) {
	client, err := r.getUniversalClient()
	if err != nil {
		return 0, err
	}

	size, err := client.DBSize(ctx).Result()
	if err != nil {
		return 0, fmt.Errorf("redis dbsize error: %w", err)
	}
	return size, nil
}

// GetSet 设置新值并返回旧值
func (r *RedisCache) GetSet(ctx context.Context, key string, value []byte) ([]byte, error) {
	if key == "" {
		return nil, fmt.Errorf("缓存键不能为空")
	}
	if value == nil {
		return nil, fmt.Errorf("缓存值不能为nil")
	}

	client, err := r.getUniversalClient()
	if err != nil {
		return nil, err
	}

	fullKey := r.buildKey(key)
	result, err := client.GetSet(ctx, fullKey, value).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // 键不存在时返回nil
		}
		return nil, fmt.Errorf("redis getset error: %w", err)
	}
	return []byte(result), nil
}

// GetSetString 设置新字符串值并返回旧字符串值
func (r *RedisCache) GetSetString(ctx context.Context, key string, value string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("缓存键不能为空")
	}

	client, err := r.getUniversalClient()
	if err != nil {
		return "", err
	}

	fullKey := r.buildKey(key)
	result, err := client.GetSet(ctx, fullKey, value).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil // 键不存在时返回空字符串
		}
		return "", fmt.Errorf("redis getset error: %w", err)
	}
	return result, nil
}

// Append 向字符串值追加内容
func (r *RedisCache) Append(ctx context.Context, key string, value string) (int, error) {
	if key == "" {
		return 0, fmt.Errorf("缓存键不能为空")
	}

	client, err := r.getUniversalClient()
	if err != nil {
		return 0, err
	}

	fullKey := r.buildKey(key)
	result, err := client.Append(ctx, fullKey, value).Result()
	if err != nil {
		return 0, fmt.Errorf("redis append error: %w", err)
	}
	return int(result), nil
}
