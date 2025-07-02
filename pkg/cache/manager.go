package cache

import (
	"fmt"
	"sync"
)

// Manager 缓存管理器
// 负责管理不同类型的缓存实例，使用简单的实例缓存模式
type Manager struct {
	// instances 缓存实例映射，key为实例名称，value为缓存接口
	instances map[string]Cache
	// mutex 读写互斥锁，保证并发安全
	mutex sync.RWMutex
}

// NewManager 创建新的缓存管理器
// 返回一个初始化完成的缓存管理器实例
func NewManager() *Manager {
	return &Manager{
		instances: make(map[string]Cache),
	}
}

// AddCache 添加缓存实例
// 将创建好的缓存实例添加到管理器中进行管理
// 参数:
//   name: 缓存实例名称，用于后续获取
//   cache: 实现了Cache接口的缓存实例
// 返回:
//   error: 如果实例名称已存在则返回错误
func (m *Manager) AddCache(name string, cache Cache) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 防止添加nil缓存实例
	if cache == nil {
		return fmt.Errorf("不能添加nil缓存实例 '%s'", name)
	}

	// 检查是否已存在同名实例
	if _, exists := m.instances[name]; exists {
		return fmt.Errorf("缓存实例 '%s' 已存在", name)
	}

	// 添加缓存实例
	m.instances[name] = cache
	return nil
}

// GetCache 获取缓存实例
// 根据实例名称获取对应的缓存实例
// 参数:
//   name: 缓存实例名称
// 返回:
//   Cache: 缓存接口实例，如果不存在则返回nil
func (m *Manager) GetCache(name string) Cache {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.instances[name]
}

// RemoveCache 移除缓存实例
// 从管理器中移除指定的缓存实例，并关闭连接
// 参数:
//   name: 缓存实例名称
// 返回:
//   error: 移除过程中的错误信息
func (m *Manager) RemoveCache(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 检查实例是否存在
	cache, exists := m.instances[name]
	if !exists {
		return fmt.Errorf("缓存实例 '%s' 不存在", name)
	}

	// 关闭缓存连接
	if err := cache.Close(); err != nil {
		return fmt.Errorf("关闭缓存实例 '%s' 失败: %w", name, err)
	}

	// 从实例映射中删除
	delete(m.instances, name)
	return nil
}

// ListCaches 列出所有缓存实例名称
// 返回当前管理器中所有缓存实例的名称列表
// 返回:
//   []string: 缓存实例名称列表
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
// 关闭管理器中的所有缓存连接，并清空实例映射
// 返回:
//   error: 第一个遇到的错误，如果没有错误则返回nil
func (m *Manager) CloseAll() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var firstErr error
	// 遍历所有实例并关闭连接
	for name, cache := range m.instances {
		if err := cache.Close(); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("关闭缓存实例 '%s' 失败: %w", name, err)
		}
	}

	// 清空实例映射
	m.instances = make(map[string]Cache)
	return firstErr
}

// Stats 获取所有缓存实例的统计信息
// 收集所有缓存实例的运行统计信息
// 返回:
//   map[string]map[string]interface{}: 缓存实例统计信息映射
func (m *Manager) Stats() map[string]map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := make(map[string]map[string]interface{})
	for name, cache := range m.instances {
		stats[name] = cache.Stats()
	}
	return stats
}

// HasCache 检查缓存实例是否存在
// 检查指定名称的缓存实例是否已在管理器中
// 参数:
//   name: 缓存实例名称
// 返回:
//   bool: 如果存在返回true，否则返回false
func (m *Manager) HasCache(name string) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	_, exists := m.instances[name]
	return exists
}

// 全局缓存管理器实例
var (
	globalManager *Manager
	once          sync.Once
)

// GetGlobalManager 获取全局缓存管理器实例
// 使用单例模式确保全局只有一个缓存管理器
// 返回:
//   *Manager: 全局缓存管理器实例
func GetGlobalManager() *Manager {
	once.Do(func() {
		globalManager = NewManager()
	})
	return globalManager
}

// GetCache 从全局管理器获取缓存实例
// 这是一个便捷函数，直接从全局管理器获取缓存实例
// 参数:
//   name: 缓存实例名称
// 返回:
//   Cache: 缓存接口实例，如果不存在则返回nil
func GetCache(name string) Cache {
	return GetGlobalManager().GetCache(name)
}

// GetDefaultCache 获取默认的缓存实例
// 获取名为"default"的缓存实例，通常在应用启动时设置
// 返回:
//   Cache: 默认缓存实例，如果不存在则返回nil
func GetDefaultCache() Cache {
	return GetCache("default")
}

// AddCache 向全局管理器添加缓存实例
// 这是一个便捷函数，直接向全局管理器添加缓存实例
// 参数:
//   name: 缓存实例名称
//   cache: 缓存实例
// 返回:
//   error: 添加失败时返回错误信息
func AddCache(name string, cache Cache) error {
	return GetGlobalManager().AddCache(name, cache)
}

// CloseAllCaches 关闭全局管理器中的所有缓存
// 这是一个便捷函数，用于应用关闭时清理所有缓存连接
// 返回:
//   error: 关闭过程中的错误信息
func CloseAllCaches() error {
	return GetGlobalManager().CloseAll()
}
