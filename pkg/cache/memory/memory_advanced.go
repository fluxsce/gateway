package memory

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// =============================================================================
// 高级操作
// =============================================================================

// Increment 原子递增。
//
// 将指定键的值递增指定的增量。
// 注意：基础内存缓存暂未实现此功能。
//
// 参数：
//   - ctx: 上下文
//   - key: 缓存键
//   - delta: 递增量
//
// 返回值：
//   - int64: 递增后的值
//   - error: 未实现错误
func (m *MemoryCache) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	return 0, fmt.Errorf("Increment not implemented for basic memory cache")
}

// Decrement 原子递减。
//
// 将指定键的值递减指定的递减量。
// 注意：基础内存缓存暂未实现此功能。
//
// 参数：
//   - ctx: 上下文
//   - key: 缓存键
//   - delta: 递减量
//
// 返回值：
//   - int64: 递减后的值
//   - error: 未实现错误
func (m *MemoryCache) Decrement(ctx context.Context, key string, delta int64) (int64, error) {
	return 0, fmt.Errorf("Decrement not implemented for basic memory cache")
}

// SetNX 仅当键不存在时设置值（SET if Not eXists）。
//
// 原子性地检查键是否存在，如果不存在则设置值。
// 常用于实现分布式锁等场景。
//
// 参数：
//   - ctx: 上下文
//   - key: 缓存键
//   - value: 缓存值
//   - expiration: 过期时间
//
// 返回值：
//   - bool: true 表示键之前不存在并成功设置，false 表示键已存在
//   - error: 操作失败时返回错误
//
// 使用示例：
//
//	// 实现简单的分布式锁
//	locked, err := cache.SetNX(ctx, "lock:resource", []byte("locked"), 10*time.Second)
//	if err != nil {
//	    return err
//	}
//	if locked {
//	    defer cache.Delete(ctx, "lock:resource")
//	    // 执行需要加锁的操作
//	}
func (m *MemoryCache) SetNX(ctx context.Context, key string, value []byte, expiration time.Duration) (bool, error) {
	fullKey := m.buildKey(key)

	m.mu.Lock()
	defer m.mu.Unlock()

	if item, exists := m.items[fullKey]; exists && !m.isExpired(item) {
		return false, nil
	}

	now := time.Now().UnixNano()
	item := &cacheItem{
		value:       value,
		expiration:  m.resolveExpiration(expiration),
		accessTime:  now,
		accessCount: 1,
		createTime:  now,
	}

	if m.config.EvictionPolicy == EvictionLRU {
		item.lruNode = &lruNode{key: fullKey}
		m.lruList.addToHead(item.lruNode)
	}

	m.items[fullKey] = item
	return true, nil
}

// SetNXString 仅当键不存在时设置字符串值。
//
// SetNX 的字符串版本，内部调用 SetNX 并自动转换类型。
//
// 参数：
//   - ctx: 上下文
//   - key: 缓存键
//   - value: 字符串值
//   - expiration: 过期时间
//
// 返回值：
//   - bool: 是否成功设置
//   - error: 操作失败时返回错误
func (m *MemoryCache) SetNXString(ctx context.Context, key string, value string, expiration time.Duration) (bool, error) {
	return m.SetNX(ctx, key, []byte(value), expiration)
}

// TTL 获取键的剩余生存时间。
//
// 返回指定键的剩余过期时间。
//
// 参数：
//   - ctx: 上下文
//   - key: 缓存键
//
// 返回值：
//   - time.Duration: 剩余生存时间
//   - 正数表示剩余时间
//   - -1 表示键存在但永不过期
//   - -2 表示键不存在
//   - error: 操作失败时返回错误
//
// 使用示例：
//
//	ttl, err := cache.TTL(ctx, "session:abc")
//	if err != nil {
//	    return err
//	}
//	if ttl == -2 {
//	    log.Info("键不存在")
//	} else if ttl == -1 {
//	    log.Info("键永不过期")
//	} else {
//	    log.Info("剩余时间:", ttl)
//	}
func (m *MemoryCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	fullKey := m.buildKey(key)

	m.mu.RLock()
	item, exists := m.items[fullKey]
	m.mu.RUnlock()

	if !exists {
		return -2, nil
	}

	if item.expiration == 0 {
		return -1, nil
	}

	remaining := time.Duration(item.expiration - time.Now().UnixNano())
	if remaining <= 0 {
		return -2, nil
	}

	return remaining, nil
}

// Expire 设置键的过期时间。
//
// 为已存在的键设置新的过期时间。
//
// 参数：
//   - ctx: 上下文
//   - key: 缓存键
//   - expiration: 新的过期时间
//
// 返回值：
//   - bool: true 表示成功设置，false 表示键不存在
//   - error: 操作失败时返回错误
//
// 使用示例：
//
//	// 延长会话过期时间
//	success, err := cache.Expire(ctx, "session:abc", 30*time.Minute)
func (m *MemoryCache) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	fullKey := m.buildKey(key)

	m.mu.Lock()
	defer m.mu.Unlock()

	item, exists := m.items[fullKey]
	if !exists || m.isExpired(item) {
		return false, nil
	}

	item.expiration = m.resolveExpiration(expiration)
	return true, nil
}

// GetSet 设置新值并返回旧值。
//
// 原子性地设置新值并返回旧值。
//
// 参数：
//   - ctx: 上下文
//   - key: 缓存键
//   - value: 新值
//
// 返回值：
//   - []byte: 旧值，键不存在时返回 nil
//   - error: 操作失败时返回错误
//
// 使用示例：
//
//	oldValue, err := cache.GetSet(ctx, "counter", []byte("100"))
func (m *MemoryCache) GetSet(ctx context.Context, key string, value []byte) ([]byte, error) {
	oldValue, _ := m.Get(ctx, key)
	err := m.Set(ctx, key, value, m.config.DefaultExpiration)
	return oldValue, err
}

// GetSetString 设置新字符串值并返回旧字符串值。
//
// GetSet 的字符串版本。
//
// 参数：
//   - ctx: 上下文
//   - key: 缓存键
//   - value: 新字符串值
//
// 返回值：
//   - string: 旧字符串值
//   - error: 操作失败时返回错误
func (m *MemoryCache) GetSetString(ctx context.Context, key string, value string) (string, error) {
	oldValue, _ := m.GetString(ctx, key)
	err := m.SetString(ctx, key, value, m.config.DefaultExpiration)
	return oldValue, err
}

// Append 向字符串值追加内容。
//
// 向指定键的字符串值末尾追加内容。
// 注意：基础内存缓存暂未实现此功能。
//
// 参数：
//   - ctx: 上下文
//   - key: 缓存键
//   - value: 要追加的内容
//
// 返回值：
//   - int: 追加后字符串的长度
//   - error: 未实现错误
func (m *MemoryCache) Append(ctx context.Context, key string, value string) (int, error) {
	return 0, fmt.Errorf("Append not implemented for basic memory cache")
}

// Keys 获取匹配模式的所有键。
//
// 返回所有匹配指定模式的键名。
// 注意：此操作会遍历所有键，在大数据量下性能较低，谨慎使用。
//
// 参数：
//   - ctx: 上下文
//   - pattern: 匹配模式，支持通配符：
//   - "*" 匹配所有键
//   - "user:*" 匹配以 "user:" 开头的键
//
// 返回值：
//   - []string: 匹配的键名列表
//   - error: 操作失败时返回错误
//
// 使用示例：
//
//	// 获取所有用户相关的键
//	keys, err := cache.Keys(ctx, "user:*")
//	for _, key := range keys {
//	    log.Info("找到键:", key)
//	}
func (m *MemoryCache) Keys(ctx context.Context, pattern string) ([]string, error) {
	var keys []string

	m.mu.RLock()
	for key, item := range m.items {
		if !m.isExpired(item) {
			originalKey := m.parseKey(key)
			if matchPattern(originalKey, pattern) {
				keys = append(keys, originalKey)
			}
		}
	}
	m.mu.RUnlock()

	return keys, nil
}

// matchPattern 简单的模式匹配
func matchPattern(key, pattern string) bool {
	if pattern == "*" {
		return true
	}

	// 简单实现，只支持*结尾的模式
	if strings.HasSuffix(pattern, "*") {
		prefix := pattern[:len(pattern)-1]
		return strings.HasPrefix(key, prefix)
	}

	return key == pattern
}

// Size 获取缓存中键的数量。
//
// 返回当前缓存中存储的键值对总数（包括已过期但尚未清理的键）。
//
// 参数：
//   - ctx: 上下文
//
// 返回值：
//   - int64: 键的数量
//   - error: 操作失败时返回错误
//
// 使用示例：
//
//	count, err := cache.Size(ctx)
//	log.Info("缓存中有", count, "个键")
func (m *MemoryCache) Size(ctx context.Context) (int64, error) {
	m.mu.RLock()
	total := int64(len(m.items))
	m.mu.RUnlock()
	return total, nil
}

// FlushAll 清空所有缓存。
//
// 删除缓存中的所有键值对，重置为空状态。
// 注意：此操作不可逆，谨慎使用。
//
// 参数：
//   - ctx: 上下文
//
// 返回值：
//   - error: 操作失败时返回错误
//
// 使用示例：
//
//	err := cache.FlushAll(ctx)
//	if err != nil {
//	    log.Error("清空缓存失败:", err)
//	}
func (m *MemoryCache) FlushAll(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.items = make(map[string]*cacheItem)
	m.lruList = newLRUList()
	return nil
}

// Ping 测试连接。
//
// 检查缓存实例是否可用。
// 对于内存缓存，主要检查是否已关闭。
//
// 参数：
//   - ctx: 上下文
//
// 返回值：
//   - error: 缓存不可用时返回错误，nil 表示正常
func (m *MemoryCache) Ping(ctx context.Context) error {
	if m.closed {
		return fmt.Errorf("cache is closed")
	}
	return nil
}

// Close 关闭缓存连接并释放所有资源。
//
// 停止后台清理协程，清空所有缓存数据。
// 关闭后的实例不能继续使用。
//
// 返回值：
//   - error: 关闭失败时返回错误
//
// 特性：
//   - 幂等性：可以安全地多次调用，后续调用会立即返回 nil
//   - 线程安全：可以在任何时候从任何 goroutine 调用
//
// 使用示例：
//
//	cache, err := memory.NewMemoryCache(cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer cache.Close() // 确保资源释放
func (m *MemoryCache) Close() error {
	m.closeMu.Lock()
	defer m.closeMu.Unlock()

	if m.closed {
		return nil
	}

	m.closed = true

	if m.cleanupTicker != nil {
		m.cleanupTicker.Stop()
		close(m.cleanupDone)
	}

	m.mu.Lock()
	m.items = make(map[string]*cacheItem)
	m.lruList = newLRUList()
	m.mu.Unlock()

	return nil
}

// Stats 获取缓存统计信息。
//
// 返回包含缓存性能指标和状态信息的映射，用于监控和调试。
//
// 返回值：
//   - map[string]interface{}: 统计信息映射，包含以下字段：
//   - type: 缓存类型 "memory"
//   - total_ops: 总操作次数
//   - hits: 命中次数
//   - misses: 未命中次数
//   - hit_rate: 命中率
//   - evictions: 淘汰次数
//   - expirations: 过期清理次数
//   - total_items: 当前键数量
//   - max_size: 最大容量
//   - last_updated: 最后更新时间
//
// 使用示例：
//
//	stats := cache.Stats()
//	fmt.Printf("命中率: %.2f%%\n", stats["hit_rate"].(float64)*100)
//	fmt.Printf("当前键数: %d\n", stats["total_items"])
func (m *MemoryCache) Stats() map[string]interface{} {
	m.metrics.mu.RLock()
	defer m.metrics.mu.RUnlock()

	m.mu.RLock()
	totalItems := int64(len(m.items))
	m.mu.RUnlock()

	hitRate := float64(0)
	if m.metrics.totalOps > 0 {
		hitRate = float64(m.metrics.hits) / float64(m.metrics.totalOps)
	}

	return map[string]interface{}{
		"type":         "memory",
		"total_ops":    m.metrics.totalOps,
		"hits":         m.metrics.hits,
		"misses":       m.metrics.misses,
		"hit_rate":     hitRate,
		"evictions":    m.metrics.evictions,
		"expirations":  m.metrics.expirations,
		"total_items":  totalItems,
		"max_size":     m.config.MaxSize,
		"last_updated": m.metrics.lastUpdated,
	}
}

// GetCacheType 获取缓存类型。
//
// 返回缓存类型标识，用于识别缓存实现。
//
// 返回值：
//   - string: 缓存类型标识 "memory"
func (m *MemoryCache) GetCacheType() string {
	return "memory"
}

// SelectDB 选择数据库。
//
// 内存缓存不支持数据库概念，始终返回错误。
// 此方法仅为实现 Cache 接口，与 Redis 保持接口一致性。
//
// 参数：
//   - ctx: 上下文
//   - db: 数据库编号
//
// 返回值：
//   - error: 总是返回不支持错误
func (m *MemoryCache) SelectDB(ctx context.Context, db int) error {
	return fmt.Errorf("memory cache does not support database selection")
}
