// Package collector 提供了各种系统资源采集器的基础实现
// 这个包包含了所有采集器的通用基础功能和接口定义
package collector

import (
	"sync"
	"time"

	"gohub/pkg/metric/types"
)

// BaseCollector 基础采集器结构体
// 这是所有具体采集器的基础类型，提供了通用的采集器管理功能
// 包括启用/禁用、采集间隔控制、并发安全等基本功能
//
// 设计思路:
//   - 所有采集器都应该嵌入这个基础结构体
//   - 提供统一的采集器管理接口
//   - 支持线程安全的并发访问
//   - 支持采集间隔和状态管理
type BaseCollector struct {
	// 采集器名称 - 用于标识不同的采集器
	name string
	
	// 采集器描述 - 用于说明采集器的功能
	description string
	
	// 是否启用 - 控制采集器是否执行采集操作
	enabled bool
	
	// 最后采集时间 - 记录上次执行采集的时间戳
	lastCollectTime time.Time
	
	// 采集间隔 - 控制采集频率，避免过于频繁的采集
	collectInterval time.Duration
	
	// 读写锁 - 保证并发安全，支持多个 goroutine 同时读取
	mu sync.RWMutex
}

// NewBaseCollector 创建基础采集器实例
// 这是基础采集器的工厂函数，为所有采集器提供统一的初始化方式
//
// 参数:
//   name: 采集器名称，应该是唯一的标识符
//   description: 采集器描述，用于说明其功能
//
// 返回:
//   *BaseCollector: 初始化完成的基础采集器实例
//
// 默认配置:
//   - 采集器默认启用
//   - 采集间隔使用全局默认值
//   - 最后采集时间为零值（从未采集）
//
// 使用示例:
//   base := NewBaseCollector("cpu", "CPU使用率采集器")
func NewBaseCollector(name, description string) *BaseCollector {
	return &BaseCollector{
		name:            name,
		description:     description,
		enabled:         true, // 默认启用
		collectInterval: time.Duration(types.DefaultCollectInterval) * time.Second,
	}
}

// GetName 获取采集器名称
// 线程安全地获取采集器的名称标识
//
// 返回:
//   string: 采集器名称
//
// 注意: 使用读锁保证并发安全
func (c *BaseCollector) GetName() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.name
}

// GetDescription 获取采集器描述
// 线程安全地获取采集器的功能描述
//
// 返回:
//   string: 采集器描述信息
//
// 注意: 使用读锁保证并发安全
func (c *BaseCollector) GetDescription() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.description
}

// IsEnabled 检查采集器是否启用
// 线程安全地检查采集器当前的启用状态
//
// 返回:
//   bool: true 表示启用，false 表示禁用
//
// 用途:
//   - 在执行采集前检查是否应该执行
//   - 用于管理界面显示采集器状态
//   - 支持动态启用/禁用采集器
func (c *BaseCollector) IsEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.enabled
}

// SetEnabled 设置采集器启用状态
// 线程安全地设置采集器的启用/禁用状态
//
// 参数:
//   enabled: true 启用采集器，false 禁用采集器
//
// 用途:
//   - 动态控制采集器的工作状态
//   - 在系统负载高时可以临时禁用某些采集器
//   - 支持用户配置采集器的启用状态
//
// 注意: 使用写锁保证并发安全
func (c *BaseCollector) SetEnabled(enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.enabled = enabled
}

// GetLastCollectTime 获取最后采集时间
// 线程安全地获取采集器上次执行采集的时间戳
//
// 返回:
//   time.Time: 最后采集时间，如果从未采集则为零值
//
// 用途:
//   - 计算采集间隔
//   - 监控采集器的工作状态
//   - 用于调试和性能分析
func (c *BaseCollector) GetLastCollectTime() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastCollectTime
}

// SetLastCollectTime 设置最后采集时间
// 线程安全地更新采集器的最后采集时间戳
//
// 参数:
//   t: 采集时间戳，通常使用 time.Now()
//
// 用途:
//   - 在完成采集后更新时间戳
//   - 用于采集间隔控制
//   - 支持采集器状态跟踪
//
// 注意: 
//   - 使用写锁保证并发安全
//   - 应该在每次采集完成后调用
func (c *BaseCollector) SetLastCollectTime(t time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lastCollectTime = t
}

// GetCollectInterval 获取采集间隔
// 线程安全地获取采集器的采集间隔设置
//
// 返回:
//   time.Duration: 采集间隔时长
//
// 用途:
//   - 查看当前采集频率设置
//   - 用于管理界面显示
//   - 支持动态调整采集频率
func (c *BaseCollector) GetCollectInterval() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.collectInterval
}

// SetCollectInterval 设置采集间隔
// 线程安全地设置采集器的采集间隔
//
// 参数:
//   interval: 采集间隔时长，建议不要设置过短以避免系统负载过高
//
// 用途:
//   - 动态调整采集频率
//   - 根据系统负载情况调整采集间隔
//   - 支持用户自定义采集频率
//
// 建议:
//   - CPU/内存采集器: 1-5 秒
//   - 磁盘采集器: 5-30 秒
//   - 网络采集器: 1-10 秒
//   - 进程采集器: 5-60 秒
//
// 注意: 使用写锁保证并发安全
func (c *BaseCollector) SetCollectInterval(interval time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.collectInterval = interval
}

// ShouldCollect 检查是否应该执行采集
// 综合考虑采集器状态和时间间隔，判断是否应该执行采集操作
//
// 返回:
//   bool: true 表示应该执行采集，false 表示不应该执行
//
// 判断逻辑:
//   1. 首先检查采集器是否启用
//   2. 然后检查是否超过了采集间隔
//   3. 两个条件都满足才返回 true
//
// 用途:
//   - 在采集主循环中控制采集频率
//   - 避免过于频繁的采集操作
//   - 支持动态启用/禁用采集器
//
// 使用示例:
//   if collector.ShouldCollect() {
//       // 执行采集操作
//       metrics, err := collector.Collect()
//       if err == nil {
//           collector.SetLastCollectTime(time.Now())
//       }
//   }
func (c *BaseCollector) ShouldCollect() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	// 首先检查采集器是否启用
	if !c.enabled {
		return false
	}
	
	// 检查是否超过采集间隔
	return time.Since(c.lastCollectTime) >= c.collectInterval
}

// UpdateCollectTime 更新采集时间为当前时间
// 这是一个便捷方法，用于在采集完成后更新时间戳
//
// 用途:
//   - 简化采集后的时间戳更新操作
//   - 确保时间戳的一致性
//   - 提供更简洁的 API
//
// 使用示例:
//   metrics, err := collector.Collect()
//   if err == nil {
//       collector.UpdateCollectTime()
//   }
func (c *BaseCollector) UpdateCollectTime() {
	c.SetLastCollectTime(time.Now())
}

// GetCollectStatus 获取采集器状态信息
// 返回采集器的综合状态信息，用于监控和调试
//
// 返回:
//   map[string]interface{}: 包含采集器状态的映射
//
// 返回字段:
//   - "name": 采集器名称
//   - "description": 采集器描述
//   - "enabled": 是否启用
//   - "last_collect_time": 最后采集时间
//   - "collect_interval": 采集间隔（秒）
//   - "should_collect": 是否应该执行采集
//   - "time_since_last_collect": 距离上次采集的时间（秒）
//
// 用途:
//   - 监控采集器运行状态
//   - 调试采集器问题
//   - 管理界面状态显示
func (c *BaseCollector) GetCollectStatus() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	timeSinceLastCollect := time.Since(c.lastCollectTime)
	
	return map[string]interface{}{
		"name":                     c.name,
		"description":              c.description,
		"enabled":                  c.enabled,
		"last_collect_time":        c.lastCollectTime,
		"collect_interval":         c.collectInterval.Seconds(),
		"should_collect":           c.enabled && timeSinceLastCollect >= c.collectInterval,
		"time_since_last_collect": timeSinceLastCollect.Seconds(),
	}
} 