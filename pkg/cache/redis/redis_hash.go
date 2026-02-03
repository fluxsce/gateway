// Package redis Hash 操作实现
// 提供 Redis Hash 数据类型的操作，适用于存储对象的多个字段
package redis

import (
	"context"
	"fmt"
)

// HSet 设置 Hash 表中指定字段的值。
//
// 参数：
//   - ctx: 上下文，用于控制超时和取消操作
//   - key: Hash 键名（不包含前缀）
//   - field: 字段名
//   - value: 字段值
//
// 返回值：
//   - error: 操作失败时返回错误
//
// 使用示例：
//
//	// 存储用户信息
//	err := cache.HSet(ctx, "user:1", "name", "Alice")
//	err = cache.HSet(ctx, "user:1", "age", "25")
//	err = cache.HSet(ctx, "user:1", "email", "alice@example.com")
//
// 注意事项：
//   - 如果字段已存在，会覆盖旧值
//   - Hash 表不存在时会自动创建
func (r *RedisCache) HSet(ctx context.Context, key, field, value string) error {
	if key == "" {
		return fmt.Errorf("缓存键不能为空")
	}
	if field == "" {
		return fmt.Errorf("字段名不能为空")
	}

	client, err := r.getUniversalClient()
	if err != nil {
		return err
	}

	fullKey := r.buildKey(key)
	err = client.HSet(ctx, fullKey, field, value).Err()
	if err != nil {
		return fmt.Errorf("redis hset error: %w", err)
	}
	return nil
}

// HGet 获取哈希字段值
func (r *RedisCache) HGet(ctx context.Context, key, field string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("缓存键不能为空")
	}
	if field == "" {
		return "", fmt.Errorf("字段名不能为空")
	}

	client, err := r.getUniversalClient()
	if err != nil {
		return "", err
	}

	fullKey := r.buildKey(key)
	result, err := client.HGet(ctx, fullKey, field).Result()
	if err != nil {
		if err.Error() == "redis: nil" {
			return "", nil // 字段不存在时返回空字符串
		}
		return "", fmt.Errorf("redis hget error: %w", err)
	}
	return result, nil
}

// HGetAll 获取 Hash 表中的所有字段和值。
//
// 参数：
//   - ctx: 上下文，用于控制超时和取消操作
//   - key: Hash 键名（不包含前缀）
//
// 返回值：
//   - map[string]string: 字段名到字段值的映射
//   - error: 操作失败时返回错误
//
// 使用示例：
//
//	user, err := cache.HGetAll(ctx, "user:1")
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Name: %s, Age: %s\n", user["name"], user["age"])
//
// 注意事项：
//   - 如果 Hash 表不存在，返回空映射而不是错误
//   - 大型 Hash 表可能影响性能，建议使用 HGet 获取单个字段
func (r *RedisCache) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	if key == "" {
		return nil, fmt.Errorf("缓存键不能为空")
	}

	client, err := r.getUniversalClient()
	if err != nil {
		return nil, err
	}

	fullKey := r.buildKey(key)
	result, err := client.HGetAll(ctx, fullKey).Result()
	if err != nil {
		return nil, fmt.Errorf("redis hgetall error: %w", err)
	}
	return result, nil
}

// HDel 删除哈希字段
func (r *RedisCache) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	if key == "" {
		return 0, fmt.Errorf("缓存键不能为空")
	}
	if len(fields) == 0 {
		return 0, nil
	}

	client, err := r.getUniversalClient()
	if err != nil {
		return 0, err
	}

	fullKey := r.buildKey(key)
	result, err := client.HDel(ctx, fullKey, fields...).Result()
	if err != nil {
		return 0, fmt.Errorf("redis hdel error: %w", err)
	}
	return result, nil
}
