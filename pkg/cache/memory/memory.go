package memory

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// MemoryCache 内存缓存实现
// 简化的内存缓存实现，支持基本的键值操作和TTL过期机制
// 特性:
//   - 支持TTL过期机制（已实现）
//   - 支持过期时间淘汰策略（默认且已实现）
//   - 懒惰清理和定时清理
//   - 基础的统计指标
//   - 线程安全的并发访问
//   - 其他淘汰策略（LRU、Random、FIFO）为预留功能，暂未实现
type MemoryCache struct {
	config        *MemoryConfig  // 缓存配置
	items         map[string]*cacheItem // 缓存项映射表
	lruList       *lruList       // LRU双向链表，预留给LRU淘汰策略使用
	mu            sync.RWMutex   // 读写锁，保证并发安全
	keyPrefix     string         // 键前缀，用于区分不同应用
	closed        bool           // 缓存是否已关闭
	closeMu       sync.RWMutex   // 关闭状态的读写锁
	cleanupTicker *time.Ticker   // 定时清理器
	cleanupDone   chan struct{}  // 清理协程停止信号
	metrics       *cacheMetrics  // 缓存统计指标
	lastCleanup   time.Time      // 上次清理时间
}

// cacheItem 缓存条目
// 存储缓存的值和相关元数据，包括过期时间、访问统计等
type cacheItem struct {
	value       interface{} // 缓存值，支持任意类型
	expiration  int64       // 过期时间戳(纳秒)，0表示永不过期，用于TTL策略（已实现）
	accessTime  int64       // 最后访问时间戳(纳秒)，预留给LRU策略使用
	accessCount int64       // 访问次数，预留给LFU策略使用
	createTime  int64       // 创建时间戳(纳秒)，预留给FIFO策略使用
	lruNode     *lruNode    // LRU链表节点指针，预留给LRU策略使用
}

// lruList LRU双向链表
// 实现最近最少使用(LRU)淘汰策略的核心数据结构
// 使用哨兵节点简化链表操作
type lruList struct {
	head *lruNode // 头部哨兵节点
	tail *lruNode // 尾部哨兵节点
	size int      // 链表长度
}

// lruNode LRU链表节点
// 双向链表节点，存储键名和前后指针
type lruNode struct {
	key  string   // 缓存键名
	prev *lruNode // 前驱节点
	next *lruNode // 后继节点
}

// cacheMetrics 缓存指标
// 收集和统计缓存的性能指标，用于监控和调优
type cacheMetrics struct {
	mu           sync.RWMutex // 指标读写锁
	hits         int64        // 命中次数
	misses       int64        // 未命中次数
	evictions    int64        // 淘汰次数
	expirations  int64        // 过期清理次数
	totalOps     int64        // 总操作次数
	totalSize    int64        // 总缓存大小
	totalMemory  int64        // 总内存使用量(估算)
	lastUpdated  time.Time    // 最后更新时间
}

// 数据类型别名定义
// 用于支持Redis兼容的数据结构操作

// dataValue 数据值接口，支持任意类型的缓存值
type dataValue interface{}

// hashValue 哈希值类型，对应Redis的HASH数据结构
type hashValue map[string]string

// listValue 列表值类型，对应Redis的LIST数据结构
type listValue []string

// setValue 集合值类型，对应Redis的SET数据结构
type setValue map[string]struct{}

// zsetValue 有序集合值类型，对应Redis的ZSET数据结构
type zsetValue map[string]float64

// NewMemoryCache 创建新的内存缓存实例
// 根据配置创建并初始化一个新的内存缓存实例
// 
// 参数:
//   cfg: 内存缓存配置，如果为nil则使用默认配置
// 
// 返回:
//   *MemoryCache: 创建的内存缓存实例
//   error: 创建失败时返回错误信息
//
// 特性:
//   - 自动设置配置默认值
//   - 验证配置有效性
//   - 初始化单实例存储
//   - 启动后台清理协程
func NewMemoryCache(cfg *MemoryConfig) (*MemoryCache, error) {
	if cfg == nil {
		cfg = &MemoryConfig{}
		cfg.SetDefaults()
	} else {
		cfg.SetDefaults()
		if err := cfg.Validate(); err != nil {
			return nil, fmt.Errorf("invalid memory config: %w", err)
		}
	}

	cache := &MemoryCache{
		config:      cfg,
		items:       make(map[string]*cacheItem),
		lruList:     newLRUList(),
		keyPrefix:   cfg.KeyPrefix,
		cleanupDone: make(chan struct{}),
		metrics:     &cacheMetrics{lastUpdated: time.Now()},
		lastCleanup: time.Now(),
	}

	// 启动后台清理协程
	if cfg.CleanupInterval > 0 {
		cache.cleanupTicker = time.NewTicker(cfg.CleanupInterval)
		go cache.cleanupRoutine()
	}

	return cache, nil
}

// LoadMemoryConfigFromFile 从配置文件加载内存缓存配置
// 加载配置文件并返回内存缓存配置实例
// 
// 返回:
//   *MemoryConfig: 加载的配置实例
//   error: 加载失败时返回错误信息
//
// 注意: 当前实现返回默认配置，实际项目中应从文件加载
func LoadMemoryConfigFromFile() (*MemoryConfig, error) {
	cfg := &MemoryConfig{}
	cfg.SetDefaults()
	return cfg, nil
}

// newLRUList 创建新的LRU链表
func newLRUList() *lruList {
	head := &lruNode{}
	tail := &lruNode{}
	head.next = tail
	tail.prev = head
	
	return &lruList{
		head: head,
		tail: tail,
		size: 0,
	}
}

// addToHead 添加节点到头部
func (l *lruList) addToHead(node *lruNode) {
	node.prev = l.head
	node.next = l.head.next
	l.head.next.prev = node
	l.head.next = node
	l.size++
}

// removeNode 移除节点
func (l *lruList) removeNode(node *lruNode) {
	node.prev.next = node.next
	node.next.prev = node.prev
	l.size--
}

// moveToHead 移动节点到头部
func (l *lruList) moveToHead(node *lruNode) {
	l.removeNode(node)
	l.addToHead(node)
}

// removeTail 移除尾部节点
func (l *lruList) removeTail() *lruNode {
	if l.size == 0 {
		return nil
	}
	tail := l.tail.prev
	l.removeNode(tail)
	return tail
}

// buildKey 构建完整的键名
func (m *MemoryCache) buildKey(key string) string {
	if m.keyPrefix == "" {
		return key
	}
	return m.keyPrefix + ":" + key
}

// parseKey 解析键名，移除前缀
func (m *MemoryCache) parseKey(key string) string {
	if m.keyPrefix == "" {
		return key
	}
	prefix := m.keyPrefix + ":"
	if strings.HasPrefix(key, prefix) {
		return key[len(prefix):]
	}
	return key
}

// resolveExpiration 解析过期时间
// 处理过期时间逻辑：0表示使用默认过期时间，负数表示永不过期
func (m *MemoryCache) resolveExpiration(expiration time.Duration) int64 {
	if expiration == 0 {
		// 0表示使用配置的默认过期时间
		expiration = m.config.DefaultExpiration
	} else if expiration < 0 {
		// 负数表示永不过期
		return 0
	}
	
	if expiration <= 0 {
		return 0
	}
	
	return time.Now().Add(expiration).UnixNano()
}

// isExpired 检查是否过期
func (m *MemoryCache) isExpired(item *cacheItem) bool {
	if item.expiration == 0 {
		return false
	}
	return time.Now().UnixNano() > item.expiration
}

// updateMetrics 更新指标
func (m *MemoryCache) updateMetrics(hit bool, eviction bool, expiration bool) {
	if !m.config.IsMetricsEnabled() {
		return
	}
	
	m.metrics.mu.Lock()
	defer m.metrics.mu.Unlock()
	
	m.metrics.totalOps++
	if hit {
		m.metrics.hits++
	} else {
		m.metrics.misses++
	}
	if eviction {
		m.metrics.evictions++
	}
	if expiration {
		m.metrics.expirations++
	}
	m.metrics.lastUpdated = time.Now()
}

// cleanupRoutine 清理协程
func (m *MemoryCache) cleanupRoutine() {
	for {
		select {
		case <-m.cleanupTicker.C:
			m.cleanup()
		case <-m.cleanupDone:
			return
		}
	}
}

// cleanup 清理过期数据
func (m *MemoryCache) cleanup() {
	m.closeMu.RLock()
	if m.closed {
		m.closeMu.RUnlock()
		return
	}
	m.closeMu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()
	
	now := time.Now()
	for key, item := range m.items {
		if m.isExpired(item) {
			delete(m.items, key)
			if item.lruNode != nil {
				m.lruList.removeNode(item.lruNode)
			}
			m.updateMetrics(false, false, true)
		}
	}
	m.lastCleanup = now
}

// evictIfNeeded 如果需要则进行淘汰
func (m *MemoryCache) evictIfNeeded() {
	if !m.config.IsEvictionEnabled() {
		return
	}

	needEvict := false
	if m.config.MaxSize > 0 && int64(len(m.items)) >= m.config.MaxSize {
		needEvict = true
	}

	if needEvict {
		m.evictItems(1)
	}
}

// evictItems 淘汰指定数量的条目
// 目前只实现了TTL过期策略，当达到容量限制时会尝试清理过期项
// 其他策略（LRU、Random、FIFO）为预留功能，暂未实现
func (m *MemoryCache) evictItems(count int) {
	evicted := 0
	
	// 检查配置的淘汰策略
	switch m.config.EvictionPolicy {
	case EvictionTTL:
		// TTL策略：优先清理过期项
		evicted = m.evictExpired(count)
	case EvictionLRU, EvictionRandom, EvictionFIFO:
		// 预留策略，暂未实现，降级为TTL策略
		evicted = m.evictExpired(count)
	default:
		// 默认使用TTL策略
		evicted = m.evictExpired(count)
	}
	
	if evicted > 0 {
		m.updateMetrics(false, true, false)
	}
}

// evictExpired 清理过期项，这是当前已实现的淘汰策略
func (m *MemoryCache) evictExpired(count int) int {
	evicted := 0
	keysToDelete := make([]string, 0)
	
	// 遍历所有项，查找过期的条目
	for key, item := range m.items {
		if m.isExpired(item) {
			keysToDelete = append(keysToDelete, key)
			if len(keysToDelete) >= count {
				break
			}
		}
	}
	
	// 删除过期的条目
	for _, key := range keysToDelete {
		if item, exists := m.items[key]; exists {
			delete(m.items, key)
			if item.lruNode != nil {
				m.lruList.removeNode(item.lruNode)
			}
			evicted++
		}
	}
	
	return evicted
}

// evictLRU_Reserved LRU淘汰（预留功能，未实现）
// 当实现LRU策略时，将使用此方法
func (m *MemoryCache) evictLRU_Reserved(count int) int {
	// TODO: 实现LRU淘汰策略
	// 当前降级为过期清理
	return m.evictExpired(count)
}

// evictRandom_Reserved 随机淘汰（预留功能，未实现）
// 当实现Random策略时，将使用此方法
func (m *MemoryCache) evictRandom_Reserved(count int) int {
	// TODO: 实现随机淘汰策略
	// 当前降级为过期清理
	return m.evictExpired(count)
}

// Get 获取缓存值（字节数组）
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

// GetString 获取缓存值（字符串）
func (m *MemoryCache) GetString(ctx context.Context, key string) (string, error) {
	data, err := m.Get(ctx, key)
	if err != nil || data == nil {
		return "", err
	}
	return string(data), nil
}

// Set 设置缓存值（字节数组）
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

// SetString 设置缓存值（字符串）
func (m *MemoryCache) SetString(ctx context.Context, key string, value string, expiration time.Duration) error {
	return m.Set(ctx, key, []byte(value), expiration)
}

// Delete 删除缓存值
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

// Exists 检查键是否存在
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

// MGet 批量获取缓存值（字节数组）
func (m *MemoryCache) MGet(ctx context.Context, keys []string) (map[string][]byte, error) {
	result := make(map[string][]byte)
	
	for _, key := range keys {
		if value, err := m.Get(ctx, key); err == nil && value != nil {
			result[key] = value
		}
	}
	
	return result, nil
}

// MGetString 批量获取缓存值（字符串）
func (m *MemoryCache) MGetString(ctx context.Context, keys []string) (map[string]string, error) {
	result := make(map[string]string)
	
	for _, key := range keys {
		if value, err := m.GetString(ctx, key); err == nil && value != "" {
			result[key] = value
		}
	}
	
	return result, nil
}

// MSet 批量设置缓存值（字节数组）
func (m *MemoryCache) MSet(ctx context.Context, kvPairs map[string][]byte, expiration time.Duration) error {
	for key, value := range kvPairs {
		if err := m.Set(ctx, key, value, expiration); err != nil {
			return err
		}
	}
	return nil
}

// MSetString 批量设置缓存值（字符串）
func (m *MemoryCache) MSetString(ctx context.Context, kvPairs map[string]string, expiration time.Duration) error {
	for key, value := range kvPairs {
		if err := m.SetString(ctx, key, value, expiration); err != nil {
			return err
		}
	}
	return nil
}

// MDelete 批量删除缓存值
func (m *MemoryCache) MDelete(ctx context.Context, keys []string) error {
	for _, key := range keys {
		if err := m.Delete(ctx, key); err != nil {
			return err
		}
	}
	return nil
}

// Increment 原子递增
func (m *MemoryCache) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	return 0, fmt.Errorf("Increment not implemented for basic memory cache")
}

// Decrement 原子递减
func (m *MemoryCache) Decrement(ctx context.Context, key string, delta int64) (int64, error) {
	return 0, fmt.Errorf("Decrement not implemented for basic memory cache")
}

// SetNX 仅当键不存在时设置值
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

// SetNXString 仅当键不存在时设置字符串值
func (m *MemoryCache) SetNXString(ctx context.Context, key string, value string, expiration time.Duration) (bool, error) {
	return m.SetNX(ctx, key, []byte(value), expiration)
}

// TTL 获取键的剩余生存时间
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

// Expire 设置键的过期时间
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

// Ping 测试连接
func (m *MemoryCache) Ping(ctx context.Context) error {
	if m.closed {
		return fmt.Errorf("cache is closed")
	}
	return nil
}

// Close 关闭缓存连接
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

// Stats 获取缓存统计信息
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

// FlushAll 清空所有缓存
func (m *MemoryCache) FlushAll(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.items = make(map[string]*cacheItem)
	m.lruList = newLRUList()
	return nil
}

// SelectDB 选择数据库（内存缓存不支持）
func (m *MemoryCache) SelectDB(ctx context.Context, db int) error {
	return fmt.Errorf("memory cache does not support database selection")
}

// Keys 获取匹配模式的所有键
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

// Size 获取缓存中键的数量
func (m *MemoryCache) Size(ctx context.Context) (int64, error) {
	m.mu.RLock()
	total := int64(len(m.items))
	m.mu.RUnlock()
	return total, nil
}

// GetSet 设置新值并返回旧值
func (m *MemoryCache) GetSet(ctx context.Context, key string, value []byte) ([]byte, error) {
	oldValue, _ := m.Get(ctx, key)
	err := m.Set(ctx, key, value, m.config.DefaultExpiration)
	return oldValue, err
}

// GetSetString 设置新字符串值并返回旧字符串值
func (m *MemoryCache) GetSetString(ctx context.Context, key string, value string) (string, error) {
	oldValue, _ := m.GetString(ctx, key)
	err := m.SetString(ctx, key, value, m.config.DefaultExpiration)
	return oldValue, err
}

// Append 向字符串值追加内容
func (m *MemoryCache) Append(ctx context.Context, key string, value string) (int, error) {
	return 0, fmt.Errorf("Append not implemented for basic memory cache")
} 