// Package redis List 操作实现
// 提供 Redis List 数据类型的操作，适用于实现队列、栈等数据结构
package redis

import (
	"context"
	"fmt"
)

// LPush 从 List 的左侧（头部）插入一个或多个值。
//
// 参数：
//   - ctx: 上下文，用于控制超时和取消操作
//   - key: List 键名（不包含前缀）
//   - values: 要插入的值（可变参数，可以插入多个值）
//
// 返回值：
//   - int64: 插入后 List 的长度
//   - error: 操作失败时返回错误
//
// 特性：
//   - 多个值按从左到右的顺序依次插入
//   - List 不存在时会自动创建
//   - 适合实现栈（LIFO）结构
//
// 使用示例：
//
//	// 消息队列（最新消息在前）
//	length, err := cache.LPush(ctx, "messages", "msg1", "msg2", "msg3")
//	// List 内容: ["msg3", "msg2", "msg1"]，length = 3
//
//	// 操作日志
//	_, err := cache.LPush(ctx, "logs", "user login")
func (r *RedisCache) LPush(ctx context.Context, key string, values ...string) (int64, error) {
	if key == "" {
		return 0, fmt.Errorf("缓存键不能为空")
	}
	if len(values) == 0 {
		return 0, nil
	}

	client, err := r.getUniversalClient()
	if err != nil {
		return 0, err
	}

	fullKey := r.buildKey(key)
	// 转换为interface{}切片
	vals := make([]interface{}, len(values))
	for i, v := range values {
		vals[i] = v
	}

	result, err := client.LPush(ctx, fullKey, vals...).Result()
	if err != nil {
		return 0, fmt.Errorf("redis lpush error: %w", err)
	}
	return result, nil
}

// RPush 从 List 的右侧（尾部）插入一个或多个值。
//
// 参数：
//   - ctx: 上下文，用于控制超时和取消操作
//   - key: List 键名（不包含前缀）
//   - values: 要插入的值（可变参数）
//
// 返回值：
//   - int64: 插入后 List 的长度
//   - error: 操作失败时返回错误
//
// 特性：
//   - 多个值按从左到右的顺序依次插入到尾部
//   - List 不存在时会自动创建
//   - 适合实现队列（FIFO）结构
//
// 使用示例：
//
//	// 任务队列
//	length, err := cache.RPush(ctx, "tasks", "task1", "task2", "task3")
//	// List 内容: ["task1", "task2", "task3"]，length = 3
//
//	task, _ := cache.LPop(ctx, "tasks") // 取出 "task1"（先进先出）
func (r *RedisCache) RPush(ctx context.Context, key string, values ...string) (int64, error) {
	if key == "" {
		return 0, fmt.Errorf("缓存键不能为空")
	}
	if len(values) == 0 {
		return 0, nil
	}

	client, err := r.getUniversalClient()
	if err != nil {
		return 0, err
	}

	fullKey := r.buildKey(key)
	// 转换为interface{}切片
	vals := make([]interface{}, len(values))
	for i, v := range values {
		vals[i] = v
	}

	result, err := client.RPush(ctx, fullKey, vals...).Result()
	if err != nil {
		return 0, fmt.Errorf("redis rpush error: %w", err)
	}
	return result, nil
}

// LPop 从列表左侧弹出元素
func (r *RedisCache) LPop(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("缓存键不能为空")
	}

	client, err := r.getUniversalClient()
	if err != nil {
		return "", err
	}

	fullKey := r.buildKey(key)
	result, err := client.LPop(ctx, fullKey).Result()
	if err != nil {
		if err.Error() == "redis: nil" {
			return "", nil // 列表为空时返回空字符串
		}
		return "", fmt.Errorf("redis lpop error: %w", err)
	}
	return result, nil
}

// RPop 从列表右侧弹出元素
func (r *RedisCache) RPop(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("缓存键不能为空")
	}

	client, err := r.getUniversalClient()
	if err != nil {
		return "", err
	}

	fullKey := r.buildKey(key)
	result, err := client.RPop(ctx, fullKey).Result()
	if err != nil {
		if err.Error() == "redis: nil" {
			return "", nil // 列表为空时返回空字符串
		}
		return "", fmt.Errorf("redis rpop error: %w", err)
	}
	return result, nil
}

// LLen 获取列表长度
func (r *RedisCache) LLen(ctx context.Context, key string) (int64, error) {
	if key == "" {
		return 0, fmt.Errorf("缓存键不能为空")
	}

	client, err := r.getUniversalClient()
	if err != nil {
		return 0, err
	}

	fullKey := r.buildKey(key)
	result, err := client.LLen(ctx, fullKey).Result()
	if err != nil {
		return 0, fmt.Errorf("redis llen error: %w", err)
	}
	return result, nil
}
