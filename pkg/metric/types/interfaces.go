package types

import "time"

// Collector 资源采集器基础接口
type Collector interface {
	// Collect 执行采集
	Collect() (interface{}, error)
	// GetName 获取采集器名称
	GetName() string
	// GetDescription 获取采集器描述
	GetDescription() string
	// IsEnabled 检查采集器是否启用
	IsEnabled() bool
	// SetEnabled 设置采集器启用状态
	SetEnabled(enabled bool)
}

// CPUCollector CPU采集器接口
type CPUCollector interface {
	Collector
	GetCPUUsage() (*CPUMetrics, error)
}

// MemoryCollector 内存采集器接口
type MemoryCollector interface {
	Collector
	GetMemoryUsage() (*MemoryMetrics, error)
}

// DiskCollector 磁盘采集器接口
type DiskCollector interface {
	Collector
	GetDiskUsage() (*DiskMetrics, error)
}

// NetworkCollector 网络采集器接口
type NetworkCollector interface {
	Collector
	GetNetworkStats() (*NetworkMetrics, error)
}

// SystemCollector 系统信息采集器接口
type SystemCollector interface {
	Collector
	GetSystemInfo() (*SystemMetrics, error)
}

// ProcessCollector 进程信息采集器接口
type ProcessCollector interface {
	Collector
	GetProcessInfo() (*ProcessMetrics, error)
}

// MetricManager 指标管理器接口
type MetricManager interface {
	// RegisterCollector 注册采集器
	RegisterCollector(collector Collector) error
	// UnregisterCollector 注销采集器
	UnregisterCollector(name string) error
	// GetCollector 获取指定采集器
	GetCollector(name string) (Collector, error)
	// GetAllCollectors 获取所有采集器
	GetAllCollectors() []Collector
	// CollectAll 执行所有采集器
	CollectAll() (*AllMetrics, error)
	// CollectByName 按名称执行指定采集器
	CollectByName(name string) (interface{}, error)
	// Start 启动定时采集
	Start(interval time.Duration) error
	// Stop 停止定时采集
	Stop() error
	// IsRunning 检查是否正在运行
	IsRunning() bool
	// GetCollectorNames 获取所有采集器名称
	GetCollectorNames() []string
	// GetEnabledCollectors 获取启用的采集器
	GetEnabledCollectors() []Collector
	// EnableCollector 启用采集器
	EnableCollector(name string) error
	// DisableCollector 禁用采集器
	DisableCollector(name string) error
} 