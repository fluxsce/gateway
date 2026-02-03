package memory

import (
	"context"
	"fmt"
	"time"
)

// =============================================================================
// 集合操作
// =============================================================================

// SAdd 向集合添加元素。
//
// 向指定的集合中添加一个或多个成员。
// 如果集合不存在，会自动创建。已存在的成员会被忽略。
//
// 参数：
//   - ctx: 上下文
//   - key: 集合键名
//   - members: 要添加的成员列表
//
// 返回值：
//   - int64: 成功添加的成员数量（不包括已存在的）
//   - error: 操作失败时返回错误
//
// 使用示例：
//
//	// 添加标签
//	added, err := cache.SAdd(ctx, "tags", "golang", "redis", "cache")
//	log.Info("添加了", added, "个新标签")
func (m *MemoryCache) SAdd(ctx context.Context, key string, members ...string) (int64, error) {
	fullKey := m.buildKey(key)

	m.mu.Lock()
	defer m.mu.Unlock()

	item, exists := m.items[fullKey]
	if !exists || m.isExpired(item) {
		// 创建新集合
		set := make(setValue)
		added := int64(0)
		for _, member := range members {
			set[member] = struct{}{}
			added++
		}

		now := time.Now().UnixNano()
		item = &cacheItem{
			value:       set,
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
		return added, nil
	}

	if set, ok := item.value.(setValue); ok {
		added := int64(0)
		for _, member := range members {
			if _, exists := set[member]; !exists {
				set[member] = struct{}{}
				added++
			}
		}

		item.accessTime = time.Now().UnixNano()
		item.accessCount++

		if item.lruNode != nil {
			m.lruList.moveToHead(item.lruNode)
		}

		return added, nil
	}

	return 0, fmt.Errorf("value is not a set")
}

// SRem 从集合移除元素。
//
// 从指定的集合中移除一个或多个成员。
// 如果移除后集合为空，会自动删除整个键。
//
// 参数：
//   - ctx: 上下文
//   - key: 集合键名
//   - members: 要移除的成员列表
//
// 返回值：
//   - int64: 成功移除的成员数量
//   - error: 操作失败时返回错误
//
// 使用示例：
//
//	removed, err := cache.SRem(ctx, "tags", "deprecated", "old")
func (m *MemoryCache) SRem(ctx context.Context, key string, members ...string) (int64, error) {
	fullKey := m.buildKey(key)

	m.mu.Lock()
	defer m.mu.Unlock()

	item, exists := m.items[fullKey]
	if !exists || m.isExpired(item) {
		return 0, nil
	}

	if set, ok := item.value.(setValue); ok {
		removed := int64(0)
		for _, member := range members {
			if _, exists := set[member]; exists {
				delete(set, member)
				removed++
			}
		}

		item.accessTime = time.Now().UnixNano()
		item.accessCount++

		if item.lruNode != nil {
			m.lruList.moveToHead(item.lruNode)
		}

		// 如果集合为空，删除键
		if len(set) == 0 {
			delete(m.items, fullKey)
			if item.lruNode != nil {
				m.lruList.removeNode(item.lruNode)
			}
		}

		return removed, nil
	}

	return 0, fmt.Errorf("value is not a set")
}

// SMembers 获取集合所有成员。
//
// 返回集合中的所有成员。
//
// 参数：
//   - ctx: 上下文
//   - key: 集合键名
//
// 返回值：
//   - []string: 成员列表，集合不存在时返回空切片
//   - error: 操作失败时返回错误
//
// 使用示例：
//
//	members, err := cache.SMembers(ctx, "tags")
//	for _, tag := range members {
//	    log.Info("标签:", tag)
//	}
func (m *MemoryCache) SMembers(ctx context.Context, key string) ([]string, error) {
	fullKey := m.buildKey(key)

	m.mu.RLock()
	item, exists := m.items[fullKey]
	m.mu.RUnlock()

	if !exists || m.isExpired(item) {
		return []string{}, nil
	}

	if set, ok := item.value.(setValue); ok {
		members := make([]string, 0, len(set))
		for member := range set {
			members = append(members, member)
		}
		return members, nil
	}

	return nil, fmt.Errorf("value is not a set")
}

// SIsMember 检查元素是否在集合中。
//
// 检查指定的成员是否存在于集合中。
//
// 参数：
//   - ctx: 上下文
//   - key: 集合键名
//   - member: 要检查的成员
//
// 返回值：
//   - bool: true 表示成员存在，false 表示不存在
//   - error: 操作失败时返回错误
//
// 使用示例：
//
//	exists, err := cache.SIsMember(ctx, "tags", "golang")
//	if exists {
//	    log.Info("标签存在")
//	}
func (m *MemoryCache) SIsMember(ctx context.Context, key string, member string) (bool, error) {
	fullKey := m.buildKey(key)

	m.mu.RLock()
	item, exists := m.items[fullKey]
	m.mu.RUnlock()

	if !exists || m.isExpired(item) {
		return false, nil
	}

	if set, ok := item.value.(setValue); ok {
		_, exists := set[member]
		return exists, nil
	}

	return false, fmt.Errorf("value is not a set")
}
