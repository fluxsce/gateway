package manager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gohub/pkg/metric/collector/cpu"
	"gohub/pkg/metric/collector/disk"
	"gohub/pkg/metric/collector/memory"
	"gohub/pkg/metric/collector/network"
	"gohub/pkg/metric/collector/process"
	"gohub/pkg/metric/collector/system"
	"gohub/pkg/metric/types"
)

// DefaultMetricManager 默认指标管理器
type DefaultMetricManager struct {
	// 采集器映射
	collectors map[string]types.Collector
	// 读写锁
	mu sync.RWMutex
	// 是否正在运行
	running bool
	// 取消函数
	cancel context.CancelFunc
	// 采集间隔
	interval time.Duration
	// 采集回调函数
	onCollect func(name string, data interface{}, err error)
}

// NewDefaultMetricManager 创建默认指标管理器
func NewDefaultMetricManager() *DefaultMetricManager {
	manager := &DefaultMetricManager{
		collectors: make(map[string]types.Collector),
		interval:   time.Duration(types.DefaultCollectInterval) * time.Second,
	}

	// 注册默认采集器
	manager.registerDefaultCollectors()

	return manager
}

// registerDefaultCollectors 注册默认采集器
func (m *DefaultMetricManager) registerDefaultCollectors() {
	// 注册CPU采集器
	cpuCollector := cpu.NewCPUCollector()
	m.collectors[cpuCollector.GetName()] = cpuCollector

	// 注册内存采集器
	memoryCollector := memory.NewMemoryCollector()
	m.collectors[memoryCollector.GetName()] = memoryCollector

	// 注册磁盘采集器
	diskCollector := disk.NewDiskCollector()
	m.collectors[diskCollector.GetName()] = diskCollector

	// 注册网络采集器
	networkCollector := network.NewNetworkCollector()
	m.collectors[networkCollector.GetName()] = networkCollector

	// 注册系统采集器
	systemCollector := system.NewSystemCollector()
	m.collectors[systemCollector.GetName()] = systemCollector

	// 注册进程采集器
	processCollector := process.NewProcessCollector()
	m.collectors[processCollector.GetName()] = processCollector
}

// RegisterCollector 注册采集器
func (m *DefaultMetricManager) RegisterCollector(collector types.Collector) error {
	if collector == nil {
		return fmt.Errorf("collector cannot be nil")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	name := collector.GetName()
	if name == "" {
		return fmt.Errorf("collector name cannot be empty")
	}

	if _, exists := m.collectors[name]; exists {
		return fmt.Errorf("%w: %s", types.ErrCollectorAlreadyExists, name)
	}

	m.collectors[name] = collector
	return nil
}

// UnregisterCollector 注销采集器
func (m *DefaultMetricManager) UnregisterCollector(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.collectors[name]; !exists {
		return fmt.Errorf("%w: %s", types.ErrCollectorNotFound, name)
	}

	delete(m.collectors, name)
	return nil
}

// GetCollector 获取指定采集器
func (m *DefaultMetricManager) GetCollector(name string) (types.Collector, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	collector, exists := m.collectors[name]
	if !exists {
		return nil, fmt.Errorf("%w: %s", types.ErrCollectorNotFound, name)
	}

	return collector, nil
}

// GetAllCollectors 获取所有采集器
func (m *DefaultMetricManager) GetAllCollectors() []types.Collector {
	m.mu.RLock()
	defer m.mu.RUnlock()

	collectors := make([]types.Collector, 0, len(m.collectors))
	for _, collector := range m.collectors {
		collectors = append(collectors, collector)
	}

	return collectors
}

// CollectAll 执行所有采集器
func (m *DefaultMetricManager) CollectAll() (*types.AllMetrics, error) {
	m.mu.RLock()
	collectors := make(map[string]types.Collector)
	for name, collector := range m.collectors {
		collectors[name] = collector
	}
	m.mu.RUnlock()

	allMetrics := &types.AllMetrics{
		CollectTime: time.Now(),
	}

	var collectErrors []error

	// 并发采集所有指标
	var wg sync.WaitGroup
	var mu sync.Mutex

	for name, collector := range collectors {
		if !collector.IsEnabled() {
			continue
		}

		wg.Add(1)
		go func(name string, collector types.Collector) {
			defer wg.Done()

			data, err := collector.Collect()
			if err != nil {
				mu.Lock()
				collectErrors = append(collectErrors, fmt.Errorf("采集器 %s 执行失败: %w", name, err))
				mu.Unlock()
				return
			}

			// 根据采集器类型设置对应的指标
			mu.Lock()
			switch name {
			case types.CollectorNameCPU:
				if cpuMetrics, ok := data.(*types.CPUMetrics); ok {
					allMetrics.CPU = cpuMetrics
				}
			case types.CollectorNameMemory:
				if memoryMetrics, ok := data.(*types.MemoryMetrics); ok {
					allMetrics.Memory = memoryMetrics
				}
			case types.CollectorNameDisk:
				if diskMetrics, ok := data.(*types.DiskMetrics); ok {
					allMetrics.Disk = diskMetrics
				}
			case types.CollectorNameNetwork:
				if networkMetrics, ok := data.(*types.NetworkMetrics); ok {
					allMetrics.Network = networkMetrics
				}
			case types.CollectorNameSystem:
				if systemMetrics, ok := data.(*types.SystemMetrics); ok {
					allMetrics.System = systemMetrics
				}
			case types.CollectorNameProcess:
				if processMetrics, ok := data.(*types.ProcessMetrics); ok {
					allMetrics.Process = processMetrics
				}
			}
			mu.Unlock()

			// 调用回调函数
			if m.onCollect != nil {
				m.onCollect(name, data, nil)
			}
		}(name, collector)
	}

	wg.Wait()

	// 如果有采集错误，返回第一个错误
	if len(collectErrors) > 0 {
		return allMetrics, collectErrors[0]
	}

	return allMetrics, nil
}

// CollectByName 按名称执行指定采集器
func (m *DefaultMetricManager) CollectByName(name string) (interface{}, error) {
	collector, err := m.GetCollector(name)
	if err != nil {
		return nil, err
	}

	if !collector.IsEnabled() {
		return nil, fmt.Errorf("%w: %s", types.ErrCollectorDisabled, name)
	}

	data, err := collector.Collect()
	if err != nil {
		return nil, fmt.Errorf("采集器 %s 执行失败: %w", name, err)
	}

	// 调用回调函数
	if m.onCollect != nil {
		m.onCollect(name, data, err)
	}

	return data, nil
}

// Start 启动定时采集
func (m *DefaultMetricManager) Start(interval time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return types.ErrManagerAlreadyRunning
	}

	if interval <= 0 {
		return types.ErrInvalidInterval
	}

	m.interval = interval
	m.running = true

	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel

	// 启动定时采集协程
	go m.collectLoop(ctx)

	return nil
}

// Stop 停止定时采集
func (m *DefaultMetricManager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return types.ErrManagerNotRunning
	}

	if m.cancel != nil {
		m.cancel()
	}

	m.running = false
	m.cancel = nil

	return nil
}

// IsRunning 检查是否正在运行
func (m *DefaultMetricManager) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}

// collectLoop 采集循环
func (m *DefaultMetricManager) collectLoop(ctx context.Context) {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// 执行采集
			_, err := m.CollectAll()
			if err != nil {
				// 可以记录日志或处理错误
				fmt.Printf("定时采集失败: %v\n", err)
			}
		}
	}
}

// SetCollectInterval 设置采集间隔
func (m *DefaultMetricManager) SetCollectInterval(interval time.Duration) error {
	if interval <= 0 {
		return types.ErrInvalidInterval
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.interval = interval

	// 如果正在运行，重启定时器
	if m.running {
		if m.cancel != nil {
			m.cancel()
		}

		ctx, cancel := context.WithCancel(context.Background())
		m.cancel = cancel

		go m.collectLoop(ctx)
	}

	return nil
}

// GetCollectInterval 获取采集间隔
func (m *DefaultMetricManager) GetCollectInterval() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.interval
}

// SetOnCollectCallback 设置采集回调函数
func (m *DefaultMetricManager) SetOnCollectCallback(callback func(name string, data interface{}, err error)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onCollect = callback
}

// EnableCollector 启用采集器
func (m *DefaultMetricManager) EnableCollector(name string) error {
	collector, err := m.GetCollector(name)
	if err != nil {
		return err
	}

	collector.SetEnabled(true)
	return nil
}

// DisableCollector 禁用采集器
func (m *DefaultMetricManager) DisableCollector(name string) error {
	collector, err := m.GetCollector(name)
	if err != nil {
		return err
	}

	collector.SetEnabled(false)
	return nil
}

// GetCollectorNames 获取所有采集器名称
func (m *DefaultMetricManager) GetCollectorNames() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.collectors))
	for name := range m.collectors {
		names = append(names, name)
	}

	return names
}

// GetEnabledCollectors 获取启用的采集器
func (m *DefaultMetricManager) GetEnabledCollectors() []types.Collector {
	m.mu.RLock()
	defer m.mu.RUnlock()

	collectors := make([]types.Collector, 0)
	for _, collector := range m.collectors {
		if collector.IsEnabled() {
			collectors = append(collectors, collector)
		}
	}

	return collectors
}

// GetCollectorStatus 获取采集器状态
func (m *DefaultMetricManager) GetCollectorStatus() map[string]bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := make(map[string]bool)
	for name, collector := range m.collectors {
		status[name] = collector.IsEnabled()
	}

	return status
} 