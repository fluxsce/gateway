package cache

import (
	"fmt"
	"sync"
)

// CacheType 缓存类型
type CacheType string

const (
	// TypeRedis Redis缓存类型
	TypeRedis CacheType = "redis"
	// TypeMemory 内存缓存类型（预留）
	TypeMemory CacheType = "memory"
)

// CacheFactory 缓存工厂函数类型
type CacheFactory func(config interface{}) (Cache, error)

// Manager 缓存管理器
// 负责创建和管理不同类型的缓存实例
type Manager struct {
	// factories 缓存工厂映射
	factories map[CacheType]CacheFactory
	// instances 缓存实例映射
	instances map[string]Cache
	// mutex 互斥锁
	mutex sync.RWMutex
}

// NewManager 创建新的缓存管理器
func NewManager() *Manager {
	manager := &Manager{
		factories: make(map[CacheType]CacheFactory),
		instances: make(map[string]Cache),
	}

	// 注册默认的缓存工厂
	manager.registerDefaultFactories()

	return manager
}

// registerDefaultFactories 注册默认的缓存工厂
func (m *Manager) registerDefaultFactories() {
	// 默认工厂将通过各个实现包的init函数注册
}

// RegisterFactory 注册缓存工厂
// 参数:
//   - cacheType: 缓存类型
//   - factory: 缓存工厂函数
func (m *Manager) RegisterFactory(cacheType CacheType, factory CacheFactory) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.factories[cacheType] = factory
}

// CreateCache 创建缓存实例
// 参数:
//   - name: 缓存实例名称
//   - cacheType: 缓存类型
//   - config: 缓存配置
//
// 返回:
//   - Cache: 缓存接口实例
//   - error: 可能的错误
func (m *Manager) CreateCache(name string, cacheType CacheType, config interface{}) (Cache, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 检查是否已存在同名实例
	if _, exists := m.instances[name]; exists {
		return nil, fmt.Errorf("cache instance '%s' already exists", name)
	}

	// 获取对应的工厂函数
	factory, exists := m.factories[cacheType]
	if !exists {
		return nil, fmt.Errorf("unsupported cache type: %s", cacheType)
	}

	// 创建缓存实例
	cache, err := factory(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache instance '%s': %w", name, err)
	}

	// 存储实例
	m.instances[name] = cache

	return cache, nil
}

// GetCache 获取缓存实例
// 参数:
//   - name: 缓存实例名称
//
// 返回:
//   - Cache: 缓存接口实例，如果不存在则返回nil
func (m *Manager) GetCache(name string) Cache {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.instances[name]
}

// GetOrCreateRedisCache 获取或创建Redis缓存实例
// 如果实例不存在，将使用默认配置创建
// 参数:
//   - name: 缓存实例名称
//
// 返回:
//   - Cache: 缓存接口实例
//   - error: 可能的错误
func (m *Manager) GetOrCreateRedisCache(name string) (Cache, error) {
	// 先尝试获取现有实例
	if cache := m.GetCache(name); cache != nil {
		return cache, nil
	}

	// 创建新的Redis缓存实例
	return m.CreateCache(name, TypeRedis, nil)
}

// RemoveCache 移除缓存实例
// 参数:
//   - name: 缓存实例名称
//
// 返回:
//   - error: 可能的错误
func (m *Manager) RemoveCache(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	cache, exists := m.instances[name]
	if !exists {
		return fmt.Errorf("cache instance '%s' not found", name)
	}

	// 关闭缓存连接
	if err := cache.Close(); err != nil {
		return fmt.Errorf("failed to close cache instance '%s': %w", name, err)
	}

	// 从实例映射中删除
	delete(m.instances, name)

	return nil
}

// ListCaches 列出所有缓存实例名称
// 返回:
//   - []string: 缓存实例名称列表
func (m *Manager) ListCaches() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	names := make([]string, 0, len(m.instances))
	for name := range m.instances {
		names = append(names, name)
	}

	return names
}

// CloseAll 关闭所有缓存实例
// 返回:
//   - error: 第一个遇到的错误，如果没有错误则返回nil
func (m *Manager) CloseAll() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var firstErr error
	for name, cache := range m.instances {
		if err := cache.Close(); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("failed to close cache instance '%s': %w", name, err)
		}
	}

	// 清空实例映射
	m.instances = make(map[string]Cache)

	return firstErr
}

// Stats 获取所有缓存实例的统计信息
// 返回:
//   - map[string]map[string]interface{}: 缓存实例统计信息映射
func (m *Manager) Stats() map[string]map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := make(map[string]map[string]interface{})
	for name, cache := range m.instances {
		stats[name] = cache.Stats()
	}

	return stats
}

// 全局缓存管理器实例
var (
	globalManager *Manager
	once          sync.Once
)

// GetGlobalManager 获取全局缓存管理器实例
// 使用单例模式确保全局只有一个管理器
func GetGlobalManager() *Manager {
	once.Do(func() {
		globalManager = NewManager()
	})
	return globalManager
}

// GetCache 从全局管理器获取缓存实例
// 参数:
//   - name: 缓存实例名称
//
// 返回:
//   - Cache: 缓存接口实例，如果不存在则返回nil
func GetCache(name string) Cache {
	return GetGlobalManager().GetCache(name)
}

// GetOrCreateRedisCache 从全局管理器获取或创建Redis缓存实例
// 参数:
//   - name: 缓存实例名称
//
// 返回:
//   - Cache: 缓存接口实例
//   - error: 可能的错误
func GetOrCreateRedisCache(name string) (Cache, error) {
	return GetGlobalManager().GetOrCreateRedisCache(name)
}

// GetDefaultRedisCache 获取默认的Redis缓存实例
// 如果不存在，将创建一个名为"default"的Redis缓存实例
// 返回:
//   - Cache: 缓存接口实例
//   - error: 可能的错误
func GetDefaultRedisCache() (Cache, error) {
	return GetOrCreateRedisCache("default")
}
