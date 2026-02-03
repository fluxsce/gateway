// Package redis Set 操作实现
// 提供 Redis Set 数据类型的操作，适用于存储不重复的元素集合
package redis

import (
	"context"
	"fmt"
)

// SAdd 向 Set 中添加一个或多个成员。
//
// 参数：
//   - ctx: 上下文，用于控制超时和取消操作
//   - key: Set 键名（不包含前缀）
//   - members: 要添加的成员（可变参数）
//
// 返回值：
//   - int64: 实际添加的成员数量（不包括已存在的成员）
//   - error: 操作失败时返回错误
//
// 特性：
//   - Set 自动去重，重复的成员会被忽略
//   - Set 不存在时会自动创建
//   - 适合存储标签、权限等不重复的集合
//
// 使用示例：
//
//	// 用户标签
//	count, err := cache.SAdd(ctx, "user:1:tags", "golang", "redis", "docker")
//	// count = 3（添加了3个标签）
//
//	// 在线用户集合
//	_, err := cache.SAdd(ctx, "online_users", "user123", "user456")
//
//	// 重复添加不会报错
//	count, err := cache.SAdd(ctx, "user:1:tags", "golang") // count = 0（已存在）
func (r *RedisCache) SAdd(ctx context.Context, key string, members ...string) (int64, error) {
	if key == "" {
		return 0, fmt.Errorf("缓存键不能为空")
	}
	if len(members) == 0 {
		return 0, nil
	}

	client, err := r.getUniversalClient()
	if err != nil {
		return 0, err
	}

	fullKey := r.buildKey(key)
	// 转换为interface{}切片
	vals := make([]interface{}, len(members))
	for i, v := range members {
		vals[i] = v
	}

	result, err := client.SAdd(ctx, fullKey, vals...).Result()
	if err != nil {
		return 0, fmt.Errorf("redis sadd error: %w", err)
	}
	return result, nil
}

// SRem 从集合移除元素
func (r *RedisCache) SRem(ctx context.Context, key string, members ...string) (int64, error) {
	if key == "" {
		return 0, fmt.Errorf("缓存键不能为空")
	}
	if len(members) == 0 {
		return 0, nil
	}

	client, err := r.getUniversalClient()
	if err != nil {
		return 0, err
	}

	fullKey := r.buildKey(key)
	// 转换为interface{}切片
	vals := make([]interface{}, len(members))
	for i, v := range members {
		vals[i] = v
	}

	result, err := client.SRem(ctx, fullKey, vals...).Result()
	if err != nil {
		return 0, fmt.Errorf("redis srem error: %w", err)
	}
	return result, nil
}

// SMembers 获取 Set 中的所有成员。
//
// 参数：
//   - ctx: 上下文，用于控制超时和取消操作
//   - key: Set 键名（不包含前缀）
//
// 返回值：
//   - []string: 所有成员的切片（无序）
//   - error: 操作失败时返回错误
//
// 注意事项：
//   - 返回的成员顺序是随机的（Set 无序）
//   - Set 不存在时返回空切片
//   - 对于大型 Set，该操作可能较慢，考虑使用 SSCAN
//
// 使用示例：
//
//	tags, err := cache.SMembers(ctx, "user:1:tags")
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("用户标签: %v\n", tags)
func (r *RedisCache) SMembers(ctx context.Context, key string) ([]string, error) {
	if key == "" {
		return nil, fmt.Errorf("缓存键不能为空")
	}

	client, err := r.getUniversalClient()
	if err != nil {
		return nil, err
	}

	fullKey := r.buildKey(key)
	result, err := client.SMembers(ctx, fullKey).Result()
	if err != nil {
		return nil, fmt.Errorf("redis smembers error: %w", err)
	}
	return result, nil
}

// SIsMember 检查元素是否在集合中
func (r *RedisCache) SIsMember(ctx context.Context, key string, member string) (bool, error) {
	if key == "" {
		return false, fmt.Errorf("缓存键不能为空")
	}

	client, err := r.getUniversalClient()
	if err != nil {
		return false, err
	}

	fullKey := r.buildKey(key)
	result, err := client.SIsMember(ctx, fullKey, member).Result()
	if err != nil {
		return false, fmt.Errorf("redis sismember error: %w", err)
	}
	return result, nil
}
