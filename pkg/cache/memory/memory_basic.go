package memory

import (
	"context"
	"fmt"
	"time"
)

// =============================================================================
// 基本缓存操作
// =============================================================================

// Get 获取缓存值（字节数组）。
//
// 从缓存中获取指定键的值，返回字节数组格式。
// 如果键不存在或已过期，返回 nil 而不是错误。
//
// 参数：
//   - ctx: 上下文，用于控制请求超时和取消
//   - key: 缓存键名
//
// 返回值：
//   - []byte: 缓存值的字节数组，键不存在时返回 nil
//   - error: 操作失败时返回错误（如缓存已关闭）
//
// 特性：
//   - 自动处理过期键：如果键已过期且启用懒惰清理，会自动删除
//   - 更新访问统计：记录访问时间和次数，用于 LRU 策略（预留）
//   - 支持类型转换：自动将 string 类型转换为 []byte
//
// 使用示例：
//
//	value, err := cache.Get(ctx, "user:123")
//	if err != nil {
//	    log.Error("获取缓存失败:", err)
//	    return err
//	}
//	if value == nil {
//	    log.Info("键不存在")
//	    return nil
//	}
//	log.Info("获取到值:", string(value))
func (m *MemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	m.closeMu.RLock()
	if m.closed {
		m.closeMu.RUnlock()
		return nil, fmt.Errorf("cache is closed")
	}
	m.closeMu.RUnlock()

	fullKey := m.buildKey(key)

	m.mu.RLock()
	item, exists := m.items[fullKey]
	m.mu.RUnlock()

	if !exists {
		m.updateMetrics(false, false, false)
		return nil, nil
	}

	if m.isExpired(item) {
		if m.config.EnableLazyCleanup {
			m.mu.Lock()
			delete(m.items, fullKey)
			if item.lruNode != nil {
				m.lruList.removeNode(item.lruNode)
			}
			m.mu.Unlock()
			m.updateMetrics(false, false, true)
		}
		return nil, nil
	}

	now := time.Now().UnixNano()
	m.mu.Lock()
	item.accessTime = now
	item.accessCount++
	if item.lruNode != nil {
		m.lruList.moveToHead(item.lruNode)
	}
	m.mu.Unlock()

	m.updateMetrics(true, false, false)

	switch v := item.value.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	default:
		return []byte(fmt.Sprintf("%v", v)), nil
	}
}

// GetString 获取缓存值（字符串）。
//
// 从缓存中获取指定键的值，返回字符串格式。
// 内部调用 Get 方法，并将结果转换为字符串。
//
// 参数：
//   - ctx: 上下文，用于控制请求超时和取消
//   - key: 缓存键名
//
// 返回值：
//   - string: 缓存值的字符串，键不存在时返回空字符串
//   - error: 操作失败时返回错误（如缓存已关闭）
//
// 使用示例：
//
//	name, err := cache.GetString(ctx, "user:name")
//	if err != nil {
//	    return err
//	}
//	if name == "" {
//	    log.Info("用户名不存在")
//	}
func (m *MemoryCache) GetString(ctx context.Context, key string) (string, error) {
	data, err := m.Get(ctx, key)
	if err != nil || data == nil {
		return "", err
	}
	return string(data), nil
}

// Set 设置缓存值（字节数组）。
//
// 将指定的键值对存入缓存，支持设置过期时间。
//
// 参数：
//   - ctx: 上下文，用于控制请求超时和取消
//   - key: 缓存键名
//   - value: 缓存值（字节数组）
//   - expiration: 过期时间，0 表示使用默认过期时间，负数表示永不过期
//
// 返回值：
//   - error: 操作失败时返回错误（如缓存已关闭或容量已满）
//
// 特性：
//   - 自动淘汰：当缓存满时，根据淘汰策略自动清理过期项
//   - 覆盖写入：如果键已存在，会覆盖旧值
//   - LRU 支持：如果启用 LRU 策略，会更新 LRU 链表（预留功能）
//
// 使用示例：
//
//	// 设置 10 分钟过期
//	err := cache.Set(ctx, "session:abc", []byte("user123"), 10*time.Minute)
//
//	// 使用默认过期时间
//	err := cache.Set(ctx, "key", value, 0)
//
//	// 永不过期
//	err := cache.Set(ctx, "permanent", value, -1)
func (m *MemoryCache) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	m.closeMu.RLock()
	if m.closed {
		m.closeMu.RUnlock()
		return fmt.Errorf("cache is closed")
	}
	m.closeMu.RUnlock()

	fullKey := m.buildKey(key)

	now := time.Now().UnixNano()
	item := &cacheItem{
		value:       value,
		expiration:  m.resolveExpiration(expiration),
		accessTime:  now,
		accessCount: 1,
		createTime:  now,
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.evictIfNeeded()

	// 移除旧项的LRU节点（为预留的LRU策略保留）
	if oldItem, exists := m.items[fullKey]; exists && oldItem.lruNode != nil {
		m.lruList.removeNode(oldItem.lruNode)
	}

	// LRU相关逻辑为预留功能，当前只在配置为LRU时保留结构
	if m.config.EvictionPolicy == EvictionLRU {
		// 为预留的LRU策略保留节点结构
		item.lruNode = &lruNode{key: fullKey}
		m.lruList.addToHead(item.lruNode)
	}

	m.items[fullKey] = item

	return nil
}

// SetString 设置缓存值（字符串）。
//
// 将指定的键值对（字符串）存入缓存。
// 内部调用 Set 方法，自动将字符串转换为字节数组。
//
// 参数：
//   - ctx: 上下文，用于控制请求超时和取消
//   - key: 缓存键名
//   - value: 缓存值（字符串）
//   - expiration: 过期时间，0 表示使用默认过期时间，负数表示永不过期
//
// 返回值：
//   - error: 操作失败时返回错误
//
// 使用示例：
//
//	err := cache.SetString(ctx, "username", "alice", 5*time.Minute)
func (m *MemoryCache) SetString(ctx context.Context, key string, value string, expiration time.Duration) error {
	return m.Set(ctx, key, []byte(value), expiration)
}

// Delete 删除缓存值。
//
// 从缓存中删除指定的键及其关联的值。
// 如果键不存在，不会返回错误。
//
// 参数：
//   - ctx: 上下文，用于控制请求超时和取消
//   - key: 要删除的缓存键名
//
// 返回值：
//   - error: 操作失败时返回错误（如缓存已关闭）
//
// 特性：
//   - 幂等操作：多次删除同一个键不会报错
//   - LRU 清理：如果键有 LRU 节点，会一并清理
//
// 使用示例：
//
//	err := cache.Delete(ctx, "session:expired")
//	if err != nil {
//	    log.Error("删除缓存失败:", err)
//	}
func (m *MemoryCache) Delete(ctx context.Context, key string) error {
	m.closeMu.RLock()
	if m.closed {
		m.closeMu.RUnlock()
		return fmt.Errorf("cache is closed")
	}
	m.closeMu.RUnlock()

	fullKey := m.buildKey(key)

	m.mu.Lock()
	defer m.mu.Unlock()

	if item, exists := m.items[fullKey]; exists {
		delete(m.items, fullKey)
		if item.lruNode != nil {
			m.lruList.removeNode(item.lruNode)
		}
	}

	return nil
}

// Exists 检查键是否存在。
//
// 检查指定的键是否存在于缓存中，且未过期。
//
// 参数：
//   - ctx: 上下文，用于控制请求超时和取消
//   - key: 要检查的缓存键名
//
// 返回值：
//   - bool: true 表示键存在且未过期，false 表示不存在或已过期
//   - error: 操作失败时返回错误（如缓存已关闭）
//
// 特性：
//   - 自动清理：如果键已过期且启用懒惰清理，会自动删除
//   - 不更新统计：不会更新访问时间和计数
//
// 使用示例：
//
//	exists, err := cache.Exists(ctx, "user:123")
//	if err != nil {
//	    return err
//	}
//	if exists {
//	    log.Info("用户缓存存在")
//	} else {
//	    log.Info("用户缓存不存在或已过期")
//	}
func (m *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	m.closeMu.RLock()
	if m.closed {
		m.closeMu.RUnlock()
		return false, fmt.Errorf("cache is closed")
	}
	m.closeMu.RUnlock()

	fullKey := m.buildKey(key)

	m.mu.RLock()
	item, exists := m.items[fullKey]
	m.mu.RUnlock()

	if !exists {
		return false, nil
	}

	if m.isExpired(item) {
		if m.config.EnableLazyCleanup {
			m.mu.Lock()
			delete(m.items, fullKey)
			if item.lruNode != nil {
				m.lruList.removeNode(item.lruNode)
			}
			m.mu.Unlock()
		}
		return false, nil
	}

	return true, nil
}
