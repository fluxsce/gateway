// Package memory 提供内存缓存的完整实现。
//
// 该包实现了一个线程安全的内存缓存，支持多种数据类型和淘汰策略。
//
// 主要特性：
//   - 线程安全：所有操作都支持并发调用
//   - TTL 过期机制：支持键值自动过期
//   - 多种淘汰策略：TTL（已实现）、LRU/Random/FIFO（预留）
//   - 懒惰清理：访问时自动清理过期键
//   - 定时清理：后台协程定期清理过期键
//   - 统计指标：命中率、淘汰次数等
//   - 多种数据类型：String、Hash、List、Set、ZSet
//
// 基本使用示例：
//
//	// 创建配置
//	cfg := &memory.MemoryConfig{
//	    MaxSize:           10000,
//	    DefaultExpiration: 5 * time.Minute,
//	    CleanupInterval:   1 * time.Minute,
//	}
//
//	// 创建内存缓存实例
//	cache, err := memory.NewMemoryCache(cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer cache.Close()
//
//	// 设置值
//	err = cache.Set(ctx, "key", []byte("value"), 10*time.Minute)
//
//	// 获取值
//	value, err := cache.Get(ctx, "key")
//
// 高级使用示例：
//
//	// Hash 操作
//	err = cache.HSet(ctx, "user:1", "name", "Alice")
//	name, err := cache.HGet(ctx, "user:1", "name")
//
//	// List 操作
//	length, err := cache.LPush(ctx, "queue", "task1", "task2")
//	task, err := cache.RPop(ctx, "queue")
//
//	// Set 操作
//	added, err := cache.SAdd(ctx, "tags", "go", "redis", "cache")
//	isMember, err := cache.SIsMember(ctx, "tags", "go")
//
//	// ZSet 操作
//	added, err := cache.ZAdd(ctx, "leaderboard", 100.0, "player1")
//	members, err := cache.ZRange(ctx, "leaderboard", 0, 9)
//
// 线程安全示例：
//
//	// 多个 goroutine 可以安全并发使用同一个 MemoryCache 实例
//	go cache.Get(ctx, "key1")
//	go cache.Set(ctx, "key2", value, 0)
//	go cache.Delete(ctx, "key3")
package memory

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"gateway/pkg/config"
	"gateway/pkg/logger"
)

// MemoryCache 内存缓存实现
//
// 简化的内存缓存实现，支持基本的键值操作和TTL过期机制。
// MemoryCache 是线程安全的，可以在多个 goroutine 中并发使用。
//
// 特性：
//   - 支持TTL过期机制（已实现）
//   - 支持过期时间淘汰策略（默认且已实现）
//   - 懒惰清理和定时清理
//   - 基础的统计指标
//   - 线程安全的并发访问
//   - 其他淘汰策略（LRU、Random、FIFO）为预留功能，暂未实现
//
// 并发安全性：
//   - 所有公开方法都是线程安全的
//   - 内部使用 sync.RWMutex 保护共享状态
//   - Close() 方法可以安全地在任何时候调用，且支持重复调用
//
// 错误处理：
//   - 所有方法在遇到错误时返回 error，不会 panic
//   - 如果 MemoryCache 已关闭，所有操作返回 "cache is closed" 错误
//   - 键不存在时，Get 方法返回 nil 而不是错误
//
// 注意事项：
//   - 使用完毕后必须调用 Close() 释放资源
//   - 关闭后的实例不能继续使用
//   - 键前缀会自动添加到所有操作的键名上
type MemoryCache struct {
	// config 缓存配置
	config *MemoryConfig

	// items 缓存项映射表
	items map[string]*cacheItem

	// lruList LRU双向链表，预留给LRU淘汰策略使用
	lruList *lruList

	// mu 读写锁，保证并发安全
	mu sync.RWMutex

	// keyPrefix 键前缀，用于区分不同应用
	keyPrefix string

	// closed 缓存是否已关闭
	closed bool

	// closeMu 关闭状态的读写锁
	closeMu sync.RWMutex

	// cleanupTicker 定时清理器
	cleanupTicker *time.Ticker

	// cleanupDone 清理协程停止信号
	cleanupDone chan struct{}

	// metrics 缓存统计指标
	metrics *cacheMetrics

	// lastCleanup 上次清理时间
	lastCleanup time.Time
}

// cacheItem 缓存条目
//
// 存储缓存的值和相关元数据，包括过期时间、访问统计等。
type cacheItem struct {
	// value 缓存值，支持任意类型
	value interface{}

	// expiration 过期时间戳(纳秒)，0表示永不过期，用于TTL策略（已实现）
	expiration int64

	// accessTime 最后访问时间戳(纳秒)，预留给LRU策略使用
	accessTime int64

	// accessCount 访问次数，预留给LFU策略使用
	accessCount int64

	// createTime 创建时间戳(纳秒)，预留给FIFO策略使用
	createTime int64

	// lruNode LRU链表节点指针，预留给LRU策略使用
	lruNode *lruNode
}

// lruList LRU双向链表
//
// 实现最近最少使用(LRU)淘汰策略的核心数据结构。
// 使用哨兵节点简化链表操作。
type lruList struct {
	// head 头部哨兵节点
	head *lruNode

	// tail 尾部哨兵节点
	tail *lruNode

	// size 链表长度
	size int
}

// lruNode LRU链表节点
//
// 双向链表节点，存储键名和前后指针。
type lruNode struct {
	// key 缓存键名
	key string

	// prev 前驱节点
	prev *lruNode

	// next 后继节点
	next *lruNode
}

// cacheMetrics 缓存指标
//
// 收集和统计缓存的性能指标，用于监控和调优。
type cacheMetrics struct {
	mu          sync.RWMutex // 指标读写锁
	hits        int64        // 命中次数
	misses      int64        // 未命中次数
	evictions   int64        // 淘汰次数
	expirations int64        // 过期清理次数
	totalOps    int64        // 总操作次数
	totalSize   int64        // 总缓存大小
	totalMemory int64        // 总内存使用量(估算)
	lastUpdated time.Time    // 最后更新时间
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

// =============================================================================
// 构造函数和工厂方法
// =============================================================================

// NewMemoryCache 创建新的内存缓存实例。
//
// 根据配置创建并初始化一个新的内存缓存实例。
//
// 参数：
//   - cfg: 内存缓存配置，如果为nil则使用默认配置
//
// 返回值：
//   - *MemoryCache: 创建的内存缓存实例
//   - error: 创建失败时返回错误信息
//
// 特性：
//   - 自动设置配置默认值
//   - 验证配置有效性
//   - 初始化单实例存储
//   - 启动后台清理协程
//
// 使用示例：
//
//	cfg := &MemoryConfig{
//	    MaxSize:           10000,
//	    DefaultExpiration: 5 * time.Minute,
//	    CleanupInterval:   1 * time.Minute,
//	}
//	cache, err := NewMemoryCache(cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer cache.Close()
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

// CreateFromConfigPath 从配置路径创建内存缓存实例。
//
// 该函数是内存缓存模块对外提供的工厂方法，使用 config.GetSection 自动映射配置。
//
// 参数：
//   - name: 连接名称
//   - configPath: 配置路径（如 "cache.connections.memory_cache.config"）
//
// 返回值：
//   - *MemoryCache: 内存缓存实例，如果未启用则返回 nil
//   - error: 创建失败时返回错误信息
//
// 使用示例：
//
//	cache, err := memory.CreateFromConfigPath("main", "cache.connections.memory_cache.config")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer cache.Close()
func CreateFromConfigPath(name string, configPath string) (*MemoryCache, error) {
	// 创建内存缓存配置实例
	memoryConfig := &MemoryConfig{}

	// 使用 config.GetSection 自动映射配置（就像 MetricConfig 那样）
	if err := config.GetSection(configPath, memoryConfig); err != nil {
		return nil, fmt.Errorf("从配置路径 '%s' 加载内存缓存配置失败: %w", configPath, err)
	}

	// 检查是否启用，如果未启用则跳过
	if !memoryConfig.Enabled {
		logger.Debug("跳过未启用的内存缓存连接", "name", name)
		return nil, nil // 返回nil表示跳过此连接
	}

	// 设置默认值
	memoryConfig.SetDefaults()

	// 验证配置
	if err := memoryConfig.Validate(); err != nil {
		return nil, fmt.Errorf("验证内存缓存配置失败: %w", err)
	}

	// 记录连接创建信息
	logger.Debug("创建内存缓存连接",
		"name", name,
		"config", memoryConfig.String())

	// 创建内存缓存实例
	memoryCache, err := NewMemoryCache(memoryConfig)
	if err != nil {
		return nil, fmt.Errorf("创建内存缓存实例失败: %w", err)
	}

	return memoryCache, nil
}

// =============================================================================
// 内部辅助方法
// =============================================================================

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
//
// 处理过期时间逻辑：0表示使用默认过期时间，负数表示永不过期。
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
//
// 目前只实现了TTL过期策略，当达到容量限制时会尝试清理过期项。
// 其他策略（LRU、Random、FIFO）为预留功能，暂未实现。
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
//
// 当实现LRU策略时，将使用此方法。
func (m *MemoryCache) evictLRU_Reserved(count int) int {
	// TODO: 实现LRU淘汰策略
	// 当前降级为过期清理
	return m.evictExpired(count)
}

// evictRandom_Reserved 随机淘汰（预留功能，未实现）
//
// 当实现Random策略时，将使用此方法。
func (m *MemoryCache) evictRandom_Reserved(count int) int {
	// TODO: 实现随机淘汰策略
	// 当前降级为过期清理
	return m.evictExpired(count)
}
