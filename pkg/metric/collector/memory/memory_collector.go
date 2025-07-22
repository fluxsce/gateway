package memory

import (
	"context"
	"fmt"
	"time"

	"gohub/pkg/metric/collector"
	"gohub/pkg/metric/types"

	"github.com/shirou/gopsutil/v4/mem"
)

// 默认采集超时时间
const DefaultCollectTimeout = 5 * time.Second

// MemoryCollector 内存采集器
// 使用 gopsutil 实现跨平台内存信息采集
type MemoryCollector struct {
	*collector.BaseCollector
	// 采集超时时间
	timeout time.Duration
}

// NewMemoryCollector 创建内存采集器
// 使用 gopsutil 实现跨平台内存信息采集
func NewMemoryCollector() *MemoryCollector {
	return &MemoryCollector{
		BaseCollector: collector.NewBaseCollector(
			types.CollectorNameMemory,
			"基于gopsutil的内存资源采集器，提供跨平台内存信息采集",
		),
		timeout: DefaultCollectTimeout,
	}
}

// Collect 执行采集
// 统一的采集入口，使用gopsutil简化实现
func (c *MemoryCollector) Collect() (interface{}, error) {
	if !c.IsEnabled() {
		return nil, types.ErrCollectorDisabled
	}

	metrics, err := c.GetMemoryUsage()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", types.ErrCollectFailed, err)
	}

	c.SetLastCollectTime(time.Now())
	return metrics, nil
}

// GetMemoryUsage 获取内存使用率
// 使用 gopsutil 采集内存使用率、交换区等信息
func (c *MemoryCollector) GetMemoryUsage() (*types.MemoryMetrics, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	metrics := &types.MemoryMetrics{
		CollectTime: time.Now(),
	}

	// 获取虚拟内存信息
	if err := c.getVirtualMemoryStats(ctx, metrics); err != nil {
		return nil, fmt.Errorf("获取虚拟内存统计信息失败: %w", err)
	}

	// 获取交换区信息
	if err := c.getSwapMemoryStats(ctx, metrics); err != nil {
		return nil, fmt.Errorf("获取交换区统计信息失败: %w", err)
	}

	return metrics, nil
}

// getVirtualMemoryStats 获取虚拟内存统计信息
// 使用 gopsutil 获取详细的内存使用情况
func (c *MemoryCollector) getVirtualMemoryStats(ctx context.Context, metrics *types.MemoryMetrics) error {
	// 检查上下文是否已取消
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// 使用 gopsutil 获取虚拟内存信息
	vmStat, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return fmt.Errorf("gopsutil获取虚拟内存信息失败: %w", err)
	}

	// 转换为内部数据结构
	metrics.Total = vmStat.Total
	metrics.Available = vmStat.Available
	metrics.Used = vmStat.Used
	metrics.UsagePercent = vmStat.UsedPercent
	metrics.Free = vmStat.Free
	
	// 根据平台可用性设置缓存和缓冲区信息
	if vmStat.Cached > 0 {
		metrics.Cached = vmStat.Cached
	}
	if vmStat.Buffers > 0 {
		metrics.Buffers = vmStat.Buffers
	}
	if vmStat.Shared > 0 {
		metrics.Shared = vmStat.Shared
	}

	return nil
}

// getSwapMemoryStats 获取交换区统计信息
// 使用 gopsutil 获取交换区使用情况
func (c *MemoryCollector) getSwapMemoryStats(ctx context.Context, metrics *types.MemoryMetrics) error {
	// 检查上下文是否已取消
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// 使用 gopsutil 获取交换区信息
	swapStat, err := mem.SwapMemoryWithContext(ctx)
	if err != nil {
		return fmt.Errorf("gopsutil获取交换区信息失败: %w", err)
	}

	// 转换为内部数据结构
	metrics.SwapTotal = swapStat.Total
	metrics.SwapUsed = swapStat.Used
	metrics.SwapFree = swapStat.Free
	metrics.SwapUsagePercent = swapStat.UsedPercent

	return nil
}

// SetTimeout 设置采集超时时间
func (c *MemoryCollector) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// GetTimeout 获取采集超时时间
func (c *MemoryCollector) GetTimeout() time.Duration {
	return c.timeout
}

// GetMemoryUsagePercent 获取内存使用率百分比
// 这是一个便捷方法，直接返回内存使用率
func (c *MemoryCollector) GetMemoryUsagePercent() (float64, error) {
	metrics, err := c.GetMemoryUsage()
	if err != nil {
		return 0, err
	}
	return metrics.UsagePercent, nil
}

// GetAvailableMemory 获取可用内存大小（字节）
// 这是一个便捷方法，直接返回可用内存
func (c *MemoryCollector) GetAvailableMemory() (uint64, error) {
	metrics, err := c.GetMemoryUsage()
	if err != nil {
		return 0, err
	}
	return metrics.Available, nil
}

// GetUsedMemory 获取已使用内存大小（字节）
// 这是一个便捷方法，直接返回已使用内存
func (c *MemoryCollector) GetUsedMemory() (uint64, error) {
	metrics, err := c.GetMemoryUsage()
	if err != nil {
		return 0, err
	}
	return metrics.Used, nil
}

// GetSwapUsagePercent 获取交换区使用率百分比
// 这是一个便捷方法，直接返回交换区使用率
func (c *MemoryCollector) GetSwapUsagePercent() (float64, error) {
	metrics, err := c.GetMemoryUsage()
	if err != nil {
		return 0, err
	}
	return metrics.SwapUsagePercent, nil
} 