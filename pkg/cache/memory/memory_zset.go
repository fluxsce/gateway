package memory

import (
	"context"
	"fmt"
	"time"
)

// =============================================================================
// 有序集合操作
// =============================================================================

// ZAdd 向有序集合添加元素。
//
// 向指定的有序集合中添加一个成员及其分数。
// 如果成员已存在，会更新其分数。
//
// 参数：
//   - ctx: 上下文
//   - key: 有序集合键名
//   - score: 成员的分数（用于排序）
//   - member: 成员名称
//
// 返回值：
//   - int64: 新添加的成员数量（更新已存在成员返回 0）
//   - error: 操作失败时返回错误
//
// 使用示例：
//
//	// 添加排行榜成员
//	added, err := cache.ZAdd(ctx, "leaderboard", 100.5, "player1")
//	added, err = cache.ZAdd(ctx, "leaderboard", 95.0, "player2")
func (m *MemoryCache) ZAdd(ctx context.Context, key string, score float64, member string) (int64, error) {
	fullKey := m.buildKey(key)

	m.mu.Lock()
	defer m.mu.Unlock()

	item, exists := m.items[fullKey]
	if !exists || m.isExpired(item) {
		// 创建新有序集合
		zset := make(zsetValue)
		zset[member] = score

		now := time.Now().UnixNano()
		item = &cacheItem{
			value:       zset,
			expiration:  m.resolveExpiration(m.config.DefaultExpiration),
			accessTime:  now,
			accessCount: 1,
			createTime:  now,
		}

		if m.config.EvictionPolicy == EvictionLRU {
			item.lruNode = &lruNode{key: fullKey}
			m.lruList.addToHead(item.lruNode)
		}

		m.items[fullKey] = item
		return 1, nil
	}

	if zset, ok := item.value.(zsetValue); ok {
		added := int64(0)
		if _, exists := zset[member]; !exists {
			added = 1
		}
		zset[member] = score

		item.accessTime = time.Now().UnixNano()
		item.accessCount++

		if item.lruNode != nil {
			m.lruList.moveToHead(item.lruNode)
		}

		return added, nil
	}

	return 0, fmt.Errorf("value is not a sorted set")
}

// ZRem 从有序集合移除元素。
//
// 从指定的有序集合中移除一个或多个成员。
// 如果移除后有序集合为空，会自动删除整个键。
//
// 参数：
//   - ctx: 上下文
//   - key: 有序集合键名
//   - members: 要移除的成员列表
//
// 返回值：
//   - int64: 成功移除的成员数量
//   - error: 操作失败时返回错误
//
// 使用示例：
//
//	removed, err := cache.ZRem(ctx, "leaderboard", "player1", "player2")
func (m *MemoryCache) ZRem(ctx context.Context, key string, members ...string) (int64, error) {
	fullKey := m.buildKey(key)

	m.mu.Lock()
	defer m.mu.Unlock()

	item, exists := m.items[fullKey]
	if !exists || m.isExpired(item) {
		return 0, nil
	}

	if zset, ok := item.value.(zsetValue); ok {
		removed := int64(0)
		for _, member := range members {
			if _, exists := zset[member]; exists {
				delete(zset, member)
				removed++
			}
		}

		item.accessTime = time.Now().UnixNano()
		item.accessCount++

		if item.lruNode != nil {
			m.lruList.moveToHead(item.lruNode)
		}

		// 如果有序集合为空，删除键
		if len(zset) == 0 {
			delete(m.items, fullKey)
			if item.lruNode != nil {
				m.lruList.removeNode(item.lruNode)
			}
		}

		return removed, nil
	}

	return 0, fmt.Errorf("value is not a sorted set")
}

// ZScore 获取有序集合成员的分数。
//
// 返回指定成员的分数值。
//
// 参数：
//   - ctx: 上下文
//   - key: 有序集合键名
//   - member: 成员名称
//
// 返回值：
//   - float64: 成员的分数，成员不存在时返回 0
//   - error: 操作失败时返回错误
//
// 使用示例：
//
//	score, err := cache.ZScore(ctx, "leaderboard", "player1")
//	log.Info("玩家分数:", score)
func (m *MemoryCache) ZScore(ctx context.Context, key string, member string) (float64, error) {
	fullKey := m.buildKey(key)

	m.mu.RLock()
	item, exists := m.items[fullKey]
	m.mu.RUnlock()

	if !exists || m.isExpired(item) {
		return 0, nil // 键不存在时返回0，不返回错误
	}

	if zset, ok := item.value.(zsetValue); ok {
		if score, exists := zset[member]; exists {
			return score, nil
		}
		return 0, nil // 成员不存在时返回0，不返回错误
	}

	return 0, fmt.Errorf("value is not a sorted set")
}

// ZRange 获取有序集合指定范围的成员。
//
// 按分数从低到高返回指定索引范围内的成员。
// 支持负数索引，-1 表示最后一个成员。
//
// 参数：
//   - ctx: 上下文
//   - key: 有序集合键名
//   - start: 起始索引（包含），支持负数
//   - stop: 结束索引（包含），支持负数
//
// 返回值：
//   - []string: 成员列表（按分数排序），范围无效时返回空切片
//   - error: 操作失败时返回错误
//
// 使用示例：
//
//	// 获取前10名
//	top10, err := cache.ZRange(ctx, "leaderboard", 0, 9)
//
//	// 获取最后3名
//	bottom3, err := cache.ZRange(ctx, "leaderboard", -3, -1)
func (m *MemoryCache) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	fullKey := m.buildKey(key)

	m.mu.RLock()
	item, exists := m.items[fullKey]
	m.mu.RUnlock()

	if !exists || m.isExpired(item) {
		return []string{}, nil
	}

	if zset, ok := item.value.(zsetValue); ok {
		// 将有序集合转换为排序后的切片
		type memberScore struct {
			member string
			score  float64
		}

		memberScores := make([]memberScore, 0, len(zset))
		for member, score := range zset {
			memberScores = append(memberScores, memberScore{member: member, score: score})
		}

		// 按分数排序（简单冒泡排序）
		for i := 0; i < len(memberScores)-1; i++ {
			for j := i + 1; j < len(memberScores); j++ {
				if memberScores[i].score > memberScores[j].score {
					memberScores[i], memberScores[j] = memberScores[j], memberScores[i]
				}
			}
		}

		// 处理范围
		length := int64(len(memberScores))
		if start < 0 {
			start = length + start
		}
		if stop < 0 {
			stop = length + stop
		}

		if start < 0 {
			start = 0
		}
		if stop >= length {
			stop = length - 1
		}

		if start > stop || start >= length {
			return []string{}, nil
		}

		result := make([]string, 0, stop-start+1)
		for i := start; i <= stop; i++ {
			result = append(result, memberScores[i].member)
		}

		return result, nil
	}

	return nil, fmt.Errorf("value is not a sorted set")
}
