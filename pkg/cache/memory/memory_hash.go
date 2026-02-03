package memory

import (
	"context"
	"fmt"
	"time"
)

// =============================================================================
// 哈希操作
// =============================================================================

// HSet 设置哈希字段。
//
// 在指定的哈希键中设置字段和值。
// 如果哈希键不存在，会自动创建。
//
// 参数：
//   - ctx: 上下文
//   - key: 哈希键名
//   - field: 字段名
//   - value: 字段值
//
// 返回值：
//   - error: 操作失败时返回错误，如值类型不是哈希
//
// 使用示例：
//
//	// 设置用户信息
//	err := cache.HSet(ctx, "user:1", "name", "Alice")
//	err = cache.HSet(ctx, "user:1", "age", "25")
func (m *MemoryCache) HSet(ctx context.Context, key, field, value string) error {
	fullKey := m.buildKey(key)

	m.mu.Lock()
	defer m.mu.Unlock()

	item, exists := m.items[fullKey]
	if !exists || m.isExpired(item) {
		// 创建新的哈希
		hash := make(hashValue)
		hash[field] = value

		now := time.Now().UnixNano()
		item = &cacheItem{
			value:       hash,
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
		return nil
	}

	// 更新现有哈希
	if hash, ok := item.value.(hashValue); ok {
		hash[field] = value
		item.accessTime = time.Now().UnixNano()
		item.accessCount++

		if item.lruNode != nil {
			m.lruList.moveToHead(item.lruNode)
		}
		return nil
	}

	return fmt.Errorf("value is not a hash")
}

// HGet 获取哈希字段值。
//
// 从指定的哈希键中获取字段的值。
//
// 参数：
//   - ctx: 上下文
//   - key: 哈希键名
//   - field: 字段名
//
// 返回值：
//   - string: 字段值，字段不存在时返回空字符串
//   - error: 操作失败时返回错误
//
// 使用示例：
//
//	name, err := cache.HGet(ctx, "user:1", "name")
func (m *MemoryCache) HGet(ctx context.Context, key, field string) (string, error) {
	fullKey := m.buildKey(key)

	m.mu.RLock()
	item, exists := m.items[fullKey]
	m.mu.RUnlock()

	if !exists || m.isExpired(item) {
		return "", nil
	}

	if hash, ok := item.value.(hashValue); ok {
		return hash[field], nil
	}

	return "", fmt.Errorf("value is not a hash")
}

// HGetAll 获取哈希的所有字段和值。
//
// 获取指定哈希键的所有字段和值。
//
// 参数：
//   - ctx: 上下文
//   - key: 哈希键名
//
// 返回值：
//   - map[string]string: 字段值映射，哈希不存在时返回空映射
//   - error: 操作失败时返回错误
//
// 使用示例：
//
//	userInfo, err := cache.HGetAll(ctx, "user:1")
//	for field, value := range userInfo {
//	    log.Info("字段:", field, "值:", value)
//	}
func (m *MemoryCache) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	fullKey := m.buildKey(key)

	m.mu.RLock()
	item, exists := m.items[fullKey]
	m.mu.RUnlock()

	if !exists || m.isExpired(item) {
		return make(map[string]string), nil
	}

	if hash, ok := item.value.(hashValue); ok {
		result := make(map[string]string)
		for k, v := range hash {
			result[k] = v
		}
		return result, nil
	}

	return nil, fmt.Errorf("value is not a hash")
}

// HDel 删除哈希字段。
//
// 从指定的哈希键中删除一个或多个字段。
// 如果删除后哈希为空，会自动删除整个键。
//
// 参数：
//   - ctx: 上下文
//   - key: 哈希键名
//   - fields: 要删除的字段名列表
//
// 返回值：
//   - int64: 成功删除的字段数量
//   - error: 操作失败时返回错误
//
// 使用示例：
//
//	// 删除单个字段
//	deleted, err := cache.HDel(ctx, "user:1", "age")
//
//	// 删除多个字段
//	deleted, err := cache.HDel(ctx, "user:1", "age", "email")
func (m *MemoryCache) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	fullKey := m.buildKey(key)

	m.mu.Lock()
	defer m.mu.Unlock()

	item, exists := m.items[fullKey]
	if !exists || m.isExpired(item) {
		return 0, nil
	}

	if hash, ok := item.value.(hashValue); ok {
		deleted := int64(0)
		for _, field := range fields {
			if _, exists := hash[field]; exists {
				delete(hash, field)
				deleted++
			}
		}

		item.accessTime = time.Now().UnixNano()
		item.accessCount++

		if item.lruNode != nil {
			m.lruList.moveToHead(item.lruNode)
		}

		// 如果哈希为空，删除键
		if len(hash) == 0 {
			delete(m.items, fullKey)
			if item.lruNode != nil {
				m.lruList.removeNode(item.lruNode)
			}
		}

		return deleted, nil
	}

	return 0, fmt.Errorf("value is not a hash")
}
