// Package redis 高级操作实现
// 提供 Redis 的高级功能，包括原子操作、TTL 管理、分布式锁等
package redis

import (
	"context"
	"fmt"
	"time"
)

// Increment 原子性地递增指定键的值。
//
// 参数：
//   - ctx: 上下文，用于控制超时和取消操作
//   - key: 缓存键名（不包含前缀）
//   - delta: 递增量（可以为负数实现递减）
//
// 返回值：
//   - int64: 递增后的值
//   - error: 操作失败时返回错误
//
// 特性：
//   - 原子操作，线程安全
//   - 如果键不存在，会先设置为 0 再递增
//   - 如果键的值不是整数，会返回错误
//
// 使用示例：
//
//	// 计数器
//	count, err := cache.Increment(ctx, "page:views", 1)
//	fmt.Printf("访问次数: %d\n", count)
//
//	// 库存扣减
//	stock, err := cache.Increment(ctx, "product:stock:123", -1)
//	if stock < 0 {
//	    // 库存不足
//	}
func (r *RedisCache) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	if key == "" {
		return 0, fmt.Errorf("缓存键不能为空")
	}

	client, err := r.getUniversalClient()
	if err != nil {
		return 0, err
	}

	fullKey := r.buildKey(key)
	result, err := client.IncrBy(ctx, fullKey, delta).Result()
	if err != nil {
		return 0, fmt.Errorf("redis increment error: %w", err)
	}
	return result, nil
}

// Decrement 原子递减
func (r *RedisCache) Decrement(ctx context.Context, key string, delta int64) (int64, error) {
	if key == "" {
		return 0, fmt.Errorf("缓存键不能为空")
	}

	client, err := r.getUniversalClient()
	if err != nil {
		return 0, err
	}

	fullKey := r.buildKey(key)
	result, err := client.DecrBy(ctx, fullKey, delta).Result()
	if err != nil {
		return 0, fmt.Errorf("redis decrement error: %w", err)
	}
	return result, nil
}

// SetNX 仅当键不存在时设置值（原子操作）。
//
// 该方法是实现分布式锁的基础，只有当键不存在时才会设置成功。
//
// 参数：
//   - ctx: 上下文，用于控制超时和取消操作
//   - key: 缓存键名（不包含前缀）
//   - value: 要设置的值（字节数组）
//   - expiration: 过期时间
//   - 0: 使用配置的默认过期时间
//   - 正数: 指定的过期时间
//   - 负数: 永不过期
//
// 返回值：
//   - bool: true 表示设置成功（键不存在），false 表示键已存在
//   - error: 操作失败时返回错误
//
// 分布式锁示例：
//
//	lockKey := "lock:resource:123"
//	locked, err := cache.SetNX(ctx, lockKey, []byte("locked"), 10*time.Second)
//	if err != nil {
//	    return err
//	}
//	if !locked {
//	    return fmt.Errorf("资源已被锁定")
//	}
//	defer cache.Delete(ctx, lockKey) // 释放锁
//
//	// 执行需要加锁的操作...
//
// 注意事项：
//   - 必须设置过期时间，防止死锁
//   - 建议使用 UUID 作为锁的值，释放时验证
//   - 对于长时间操作，考虑使用 Redlock 算法
func (r *RedisCache) SetNX(ctx context.Context, key string, value []byte, expiration time.Duration) (bool, error) {
	if key == "" {
		return false, fmt.Errorf("缓存键不能为空")
	}
	if value == nil {
		return false, fmt.Errorf("缓存值不能为nil")
	}

	client, err := r.getUniversalClient()
	if err != nil {
		return false, err
	}

	// 处理过期时间：0表示使用默认过期时间，负数表示永不过期
	finalExpiration := r.resolveExpiration(expiration)

	fullKey := r.buildKey(key)
	result, err := client.SetNX(ctx, fullKey, value, finalExpiration).Result()
	if err != nil {
		return false, fmt.Errorf("redis setnx error: %w", err)
	}
	return result, nil
}

// SetNXString 仅当键不存在时设置字符串值
// 原子操作：只有当键不存在时才设置字符串值，常用于分布式锁
// 参数:
//   - ctx: 上下文，用于控制请求超时和取消
//   - key: 缓存键名（不包含前缀）
//   - value: 要设置的字符串值
//   - expiration: 过期时间，0表示使用配置的默认过期时间，负数表示永不过期
//
// 返回:
//   - bool: true表示设置成功（键不存在），false表示键已存在
//   - error: 操作失败时返回错误信息
func (r *RedisCache) SetNXString(ctx context.Context, key string, value string, expiration time.Duration) (bool, error) {
	if key == "" {
		return false, fmt.Errorf("缓存键不能为空")
	}

	client, err := r.getUniversalClient()
	if err != nil {
		return false, err
	}

	// 处理过期时间：0表示使用默认过期时间，负数表示永不过期
	finalExpiration := r.resolveExpiration(expiration)

	fullKey := r.buildKey(key)
	result, err := client.SetNX(ctx, fullKey, value, finalExpiration).Result()
	if err != nil {
		return false, fmt.Errorf("redis setnx error: %w", err)
	}
	return result, nil
}

// TTL 获取键的剩余生存时间。
//
// 参数：
//   - ctx: 上下文，用于控制超时和取消操作
//   - key: 缓存键名（不包含前缀）
//
// 返回值：
//   - time.Duration: 剩余生存时间
//   - 正数: 剩余的过期时间
//   - -1: 键存在但没有设置过期时间
//   - -2: 键不存在
//   - error: 操作失败时返回错误
//
// 使用示例：
//
//	ttl, err := cache.TTL(ctx, "session:123")
//	if err != nil {
//	    return err
//	}
//	if ttl == -2*time.Second {
//	    fmt.Println("会话不存在")
//	} else if ttl == -1*time.Second {
//	    fmt.Println("会话永不过期")
//	} else {
//	    fmt.Printf("会话将在 %v 后过期\n", ttl)
//	}
func (r *RedisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	if key == "" {
		return 0, fmt.Errorf("缓存键不能为空")
	}

	client, err := r.getUniversalClient()
	if err != nil {
		return 0, err
	}

	fullKey := r.buildKey(key)
	result, err := client.TTL(ctx, fullKey).Result()
	if err != nil {
		return 0, fmt.Errorf("redis ttl error: %w", err)
	}
	return result, nil
}

// Expire 为已存在的键设置过期时间。
//
// 参数：
//   - ctx: 上下文，用于控制超时和取消操作
//   - key: 缓存键名（不包含前缀）
//   - expiration: 过期时间（必须为正数）
//
// 返回值：
//   - bool: true 表示设置成功，false 表示键不存在
//   - error: 操作失败时返回错误
//
// 使用示例：
//
//	// 延长会话时间
//	ok, err := cache.Expire(ctx, "session:123", 30*time.Minute)
//	if !ok {
//	    fmt.Println("会话不存在")
//	}
//
//	// 设置临时缓存
//	cache.Set(ctx, "temp", data, -1) // 先设置永不过期
//	cache.Expire(ctx, "temp", 1*time.Hour) // 稍后设置过期时间
func (r *RedisCache) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	if key == "" {
		return false, fmt.Errorf("缓存键不能为空")
	}

	client, err := r.getUniversalClient()
	if err != nil {
		return false, err
	}

	fullKey := r.buildKey(key)
	result, err := client.Expire(ctx, fullKey, expiration).Result()
	if err != nil {
		return false, fmt.Errorf("redis expire error: %w", err)
	}
	return result, nil
}
