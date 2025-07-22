// Package metric 提供系统指标采集和监控功能
// 该包封装了指标管理器，提供便捷的全局API来采集各种系统指标
// 包括CPU、内存、磁盘、网络、系统信息和进程指标等
package metric

import (
	"time"

	"gateway/pkg/metric/manager"
	"gateway/pkg/metric/types"
	"gateway/pkg/metric/utils"
)

// DefaultManager 默认的全局指标管理器
// 提供单例模式的指标管理，简化使用方式
var DefaultManager types.MetricManager

// init 初始化默认管理器
// 在包加载时自动创建默认的指标管理器实例
func init() {
	DefaultManager = manager.NewDefaultMetricManager()
}

// =============================================================================
// 指标采集相关函数
// =============================================================================

// CollectAll 使用默认管理器采集所有指标
// 一次性采集所有启用的采集器的指标数据
// 返回包含所有指标的结构体或错误信息
func CollectAll() (*types.AllMetrics, error) {
	return DefaultManager.CollectAll()
}

// CollectByName 使用默认管理器采集指定名称的指标
// 根据采集器名称采集单个指标，支持的名称包括：
// - cpu: CPU使用率和负载信息
// - memory: 内存使用情况
// - disk: 磁盘使用情况和IO统计
// - network: 网络接口统计信息
// - system: 系统基本信息
// - process: 进程资源使用情况
func CollectByName(name string) (interface{}, error) {
	return DefaultManager.CollectByName(name)
}

// CollectCPU 采集CPU指标
// 获取CPU使用率、负载平均值、核心数等信息
// 返回CPU指标结构体，包含用户态、内核态、空闲时间等详细信息
func CollectCPU() (*types.CPUMetrics, error) {
	data, err := DefaultManager.CollectByName(types.CollectorNameCPU)
	if err != nil {
		return nil, err
	}
	if cpuMetrics, ok := data.(*types.CPUMetrics); ok {
		return cpuMetrics, nil
	}
	return nil, types.ErrCollectFailed
}

// CollectMemory 采集内存指标
// 获取系统内存使用情况，包括总内存、可用内存、已用内存等
// 返回内存指标结构体，包含物理内存和虚拟内存的详细信息
func CollectMemory() (*types.MemoryMetrics, error) {
	data, err := DefaultManager.CollectByName(types.CollectorNameMemory)
	if err != nil {
		return nil, err
	}
	if memoryMetrics, ok := data.(*types.MemoryMetrics); ok {
		return memoryMetrics, nil
	}
	return nil, types.ErrCollectFailed
}

// CollectDisk 采集磁盘指标
// 获取磁盘使用情况和IO统计信息
// 返回磁盘指标结构体，包含各个磁盘分区的使用情况和IO性能数据
func CollectDisk() (*types.DiskMetrics, error) {
	data, err := DefaultManager.CollectByName(types.CollectorNameDisk)
	if err != nil {
		return nil, err
	}
	if diskMetrics, ok := data.(*types.DiskMetrics); ok {
		return diskMetrics, nil
	}
	return nil, types.ErrCollectFailed
}

// CollectNetwork 采集网络指标
// 获取网络接口统计信息，包括接收/发送字节数、包数、错误数等
// 返回网络指标结构体，包含所有网络接口的详细统计信息
func CollectNetwork() (*types.NetworkMetrics, error) {
	data, err := DefaultManager.CollectByName(types.CollectorNameNetwork)
	if err != nil {
		return nil, err
	}
	if networkMetrics, ok := data.(*types.NetworkMetrics); ok {
		return networkMetrics, nil
	}
	return nil, types.ErrCollectFailed
}

// CollectSystem 采集系统指标
// 获取系统基本信息，包括主机名、操作系统、内核版本、运行时间等
// 返回系统指标结构体，包含系统的基本配置和运行状态信息
func CollectSystem() (*types.SystemMetrics, error) {
	data, err := DefaultManager.CollectByName(types.CollectorNameSystem)
	if err != nil {
		return nil, err
	}
	if systemMetrics, ok := data.(*types.SystemMetrics); ok {
		return systemMetrics, nil
	}
	return nil, types.ErrCollectFailed
}

// CollectProcess 采集进程指标
// 获取当前进程的资源使用情况，包括CPU、内存、文件描述符等
// 返回进程指标结构体，包含进程的详细资源使用信息
func CollectProcess() (*types.ProcessMetrics, error) {
	data, err := DefaultManager.CollectByName(types.CollectorNameProcess)
	if err != nil {
		return nil, err
	}
	if processMetrics, ok := data.(*types.ProcessMetrics); ok {
		return processMetrics, nil
	}
	return nil, types.ErrCollectFailed
}

// =============================================================================
// 监控控制相关函数
// =============================================================================

// StartMonitoring 开始定时监控
// 启动后台定时任务，按指定间隔自动采集所有指标
// interval: 采集间隔时间，建议不小于1秒
// 返回错误信息，如果管理器已经在运行则返回相应错误
func StartMonitoring(interval time.Duration) error {
	return DefaultManager.Start(interval)
}

// StopMonitoring 停止定时监控
// 停止后台定时任务，不再自动采集指标
// 返回错误信息，如果管理器未运行则返回相应错误
func StopMonitoring() error {
	return DefaultManager.Stop()
}

// IsMonitoring 检查是否正在监控
// 返回true表示监控正在运行，false表示已停止
func IsMonitoring() bool {
	return DefaultManager.IsRunning()
}

// SetCollectCallback 设置采集回调函数
// 设置当指标采集完成时的回调函数，可用于自定义处理采集结果
// callback: 回调函数，参数为采集器名称、采集数据和错误信息
// 注意：回调函数应该快速执行，避免阻塞采集流程
func SetCollectCallback(callback func(name string, data interface{}, err error)) {
	if defaultManager, ok := DefaultManager.(*manager.DefaultMetricManager); ok {
		defaultManager.SetOnCollectCallback(callback)
	}
}

// =============================================================================
// 采集器管理相关函数
// =============================================================================

// GetCollectorNames 获取所有采集器名称
// 返回所有已注册的采集器名称列表
func GetCollectorNames() []string {
	return DefaultManager.GetCollectorNames()
}

// GetEnabledCollectors 获取启用的采集器
// 返回所有已启用的采集器实例列表
func GetEnabledCollectors() []types.Collector {
	return DefaultManager.GetEnabledCollectors()
}

// EnableCollector 启用指定采集器
// 启用指定名称的采集器，使其参与指标采集
// name: 采集器名称，如 "cpu", "memory", "disk" 等
// 返回错误信息，如果采集器不存在则返回相应错误
func EnableCollector(name string) error {
	return DefaultManager.EnableCollector(name)
}

// DisableCollector 禁用指定采集器
// 禁用指定名称的采集器，使其不参与指标采集
// name: 采集器名称，如 "cpu", "memory", "disk" 等
// 返回错误信息，如果采集器不存在则返回相应错误
func DisableCollector(name string) error {
	return DefaultManager.DisableCollector(name)
}

// RegisterCollector 注册新的采集器到默认管理器
// 添加自定义采集器到管理器中
// collector: 实现了 types.Collector 接口的采集器实例
// 返回错误信息，如果采集器名称已存在则返回相应错误
func RegisterCollector(collector types.Collector) error {
	return DefaultManager.RegisterCollector(collector)
}

// UnregisterCollector 从默认管理器注销采集器
// 移除指定名称的采集器
// name: 要移除的采集器名称
// 返回错误信息，如果采集器不存在则返回相应错误
func UnregisterCollector(name string) error {
	return DefaultManager.UnregisterCollector(name)
}

// GetCollector 从默认管理器获取指定采集器
// 根据名称获取采集器实例，可用于直接操作采集器
// name: 采集器名称
// 返回采集器实例和错误信息
func GetCollector(name string) (types.Collector, error) {
	return DefaultManager.GetCollector(name)
}

// GetCollectorStatus 获取所有采集器的状态
// 返回所有采集器的启用/禁用状态映射
// 键为采集器名称，值为true表示启用，false表示禁用
func GetCollectorStatus() map[string]bool {
	if defaultManager, ok := DefaultManager.(*manager.DefaultMetricManager); ok {
		return defaultManager.GetCollectorStatus()
	}
	return make(map[string]bool)
}

// =============================================================================
// 格式化工具函数
// =============================================================================

// FormatMetrics 格式化指标为人类可读的字符串
// 将复杂的指标数据结构转换为易于阅读的文本格式
// metrics: 包含所有指标的结构体
// 返回格式化后的字符串，包含所有指标的详细信息
func FormatMetrics(metrics *types.AllMetrics) string {
	return utils.FormatAllMetrics(metrics)
}

// FormatBytes 格式化字节数
// 将字节数转换为人类可读的格式，如 "1.2GB", "512MB" 等
// bytes: 要格式化的字节数
// 返回格式化后的字符串，自动选择合适的单位
func FormatBytes(bytes uint64) string {
	return utils.FormatBytes(bytes)
}

// FormatPercentage 格式化百分比
// 将浮点数转换为百分比格式，如 "85.2%"
// percentage: 要格式化的百分比值（0-100）
// 返回格式化后的百分比字符串
func FormatPercentage(percentage float64) string {
	return utils.FormatPercentage(percentage)
}

// FormatDuration 格式化持续时间
// 将秒数转换为人类可读的时间格式，如 "2天3小时45分钟"
// seconds: 要格式化的秒数
// 返回格式化后的时间字符串
func FormatDuration(seconds uint64) string {
	return utils.FormatDuration(seconds)
}

// =============================================================================
// 高级管理函数
// =============================================================================

// NewManager 创建新的指标管理器
// 创建一个独立的指标管理器实例，不影响默认管理器
// 适用于需要多个独立管理器的场景
// 返回新的管理器实例
func NewManager() types.MetricManager {
	return manager.NewDefaultMetricManager()
}

// SetCollectInterval 设置采集间隔
// 动态调整定时监控的采集间隔
// interval: 新的采集间隔时间
// 返回错误信息，如果管理器未运行则返回相应错误
// 注意：只有在监控运行时才能设置间隔
func SetCollectInterval(interval time.Duration) error {
	if defaultManager, ok := DefaultManager.(*manager.DefaultMetricManager); ok {
		return defaultManager.SetCollectInterval(interval)
	}
	return types.ErrManagerNotRunning
}

// GetCollectInterval 获取采集间隔
// 返回当前设置的采集间隔时间
// 如果管理器未运行，返回默认间隔时间
func GetCollectInterval() time.Duration {
	if defaultManager, ok := DefaultManager.(*manager.DefaultMetricManager); ok {
		return defaultManager.GetCollectInterval()
	}
	return time.Duration(types.DefaultCollectInterval) * time.Second
}
