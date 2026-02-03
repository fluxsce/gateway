// Package redis ZSet 操作实现
// 提供 Redis Sorted Set（有序集合）数据类型的操作，适用于排行榜、优先级队列等场景
package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// ZAdd 向 Sorted Set 中添加成员（带分数）。
//
// 参数：
//   - ctx: 上下文，用于控制超时和取消操作
//   - key: Sorted Set 键名（不包含前缀）
//   - score: 成员的分数（用于排序）
//   - member: 成员值
//
// 返回值：
//   - int64: 实际添加的成员数量（0 表示成员已存在且仅更新了分数）
//   - error: 操作失败时返回错误
//
// 特性：
//   - 成员会按分数从小到大自动排序
//   - 如果成员已存在，会更新其分数
//   - Sorted Set 不存在时会自动创建
//
// 使用示例：
//
//	// 游戏排行榜（分数越高排名越前）
//	count, err := cache.ZAdd(ctx, "leaderboard", 9500, "player1")
//	_, err = cache.ZAdd(ctx, "leaderboard", 8800, "player2")
//	_, err = cache.ZAdd(ctx, "leaderboard", 10200, "player3")
//
//	// 优先级队列（分数表示优先级）
//	_, err := cache.ZAdd(ctx, "task_queue", 1.0, "low_priority_task")
//	_, err = cache.ZAdd(ctx, "task_queue", 10.0, "high_priority_task")
//
//	// 时间序列（分数为时间戳）
//	timestamp := float64(time.Now().Unix())
//	_, err := cache.ZAdd(ctx, "events", timestamp, "event_data")
func (r *RedisCache) ZAdd(ctx context.Context, key string, score float64, member string) (int64, error) {
	if key == "" {
		return 0, fmt.Errorf("缓存键不能为空")
	}

	client, err := r.getUniversalClient()
	if err != nil {
		return 0, err
	}

	fullKey := r.buildKey(key)
	z := redis.Z{Score: score, Member: member}
	result, err := client.ZAdd(ctx, fullKey, z).Result()
	if err != nil {
		return 0, fmt.Errorf("redis zadd error: %w", err)
	}
	return result, nil
}

// ZRem 从有序集合移除元素
func (r *RedisCache) ZRem(ctx context.Context, key string, members ...string) (int64, error) {
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

	result, err := client.ZRem(ctx, fullKey, vals...).Result()
	if err != nil {
		return 0, fmt.Errorf("redis zrem error: %w", err)
	}
	return result, nil
}

// ZScore 获取有序集合成员的分数
func (r *RedisCache) ZScore(ctx context.Context, key string, member string) (float64, error) {
	if key == "" {
		return 0, fmt.Errorf("缓存键不能为空")
	}

	client, err := r.getUniversalClient()
	if err != nil {
		return 0, err
	}

	fullKey := r.buildKey(key)
	result, err := client.ZScore(ctx, fullKey, member).Result()
	if err != nil {
		if err.Error() == "redis: nil" {
			return 0, nil // 成员不存在时返回0
		}
		return 0, fmt.Errorf("redis zscore error: %w", err)
	}
	return result, nil
}

// ZRange 按索引范围获取 Sorted Set 中的成员（按分数从小到大排序）。
//
// 参数：
//   - ctx: 上下文，用于控制超时和取消操作
//   - key: Sorted Set 键名（不包含前缀）
//   - start: 起始索引（0 表示第一个，-1 表示最后一个）
//   - stop: 结束索引（-1 表示最后一个）
//
// 返回值：
//   - []string: 指定范围内的成员切片（按分数升序）
//   - error: 操作失败时返回错误
//
// 索引说明：
//   - 正数索引：0 表示第一个元素，1 表示第二个元素
//   - 负数索引：-1 表示最后一个元素，-2 表示倒数第二个
//   - 范围是闭区间，包含 start 和 stop
//
// 使用示例：
//
//	// 获取排行榜前 10 名
//	top10, err := cache.ZRange(ctx, "leaderboard", 0, 9)
//
//	// 获取最后 5 名
//	bottom5, err := cache.ZRange(ctx, "leaderboard", -5, -1)
//
//	// 获取所有成员
//	all, err := cache.ZRange(ctx, "leaderboard", 0, -1)
//
// 注意事项：
//   - 返回的是成员值，不包含分数
//   - 如需获取分数，使用 ZRangeWithScores
func (r *RedisCache) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	if key == "" {
		return nil, fmt.Errorf("缓存键不能为空")
	}

	client, err := r.getUniversalClient()
	if err != nil {
		return nil, err
	}

	fullKey := r.buildKey(key)
	result, err := client.ZRange(ctx, fullKey, start, stop).Result()
	if err != nil {
		return nil, fmt.Errorf("redis zrange error: %w", err)
	}
	return result, nil
}
