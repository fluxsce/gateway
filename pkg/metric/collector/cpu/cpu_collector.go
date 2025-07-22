package cpu

import (
	"context"
	"fmt"
	"time"

	"gohub/pkg/metric/collector"
	"gohub/pkg/metric/types"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/load"
)

// 默认采集超时时间
const DefaultCollectTimeout = 5 * time.Second

// CPUCollector CPU采集器
// 使用 gopsutil 实现跨平台CPU信息采集
type CPUCollector struct {
	*collector.BaseCollector
	// 采集超时时间
	timeout time.Duration
}

// NewCPUCollector 创建CPU采集器
// 使用 gopsutil 实现跨平台CPU信息采集
func NewCPUCollector() *CPUCollector {
	return &CPUCollector{
		BaseCollector: collector.NewBaseCollector(
			types.CollectorNameCPU,
			"基于gopsutil的CPU资源采集器，提供跨平台CPU信息采集",
		),
		timeout: DefaultCollectTimeout,
	}
}

// Collect 执行采集
// 统一的采集入口，使用gopsutil简化实现
func (c *CPUCollector) Collect() (interface{}, error) {
	if !c.IsEnabled() {
		return nil, types.ErrCollectorDisabled
	}

	metrics, err := c.GetCPUUsage()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", types.ErrCollectFailed, err)
	}

	c.SetLastCollectTime(time.Now())
	return metrics, nil
}

// GetCPUUsage 获取CPU使用率
// 使用 gopsutil 采集CPU使用率、负载等信息
func (c *CPUCollector) GetCPUUsage() (*types.CPUMetrics, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	metrics := &types.CPUMetrics{
		CollectTime: time.Now(),
	}

	// 获取CPU核心数信息
	if err := c.getCPUInfo(ctx, metrics); err != nil {
		return nil, fmt.Errorf("获取CPU核心信息失败: %w", err)
	}

	// 获取CPU使用率统计
	if err := c.getCPUStats(ctx, metrics); err != nil {
		return nil, fmt.Errorf("获取CPU统计信息失败: %w", err)
	}

	// 获取负载平均值（Linux/Unix系统）
	c.getLoadAverage(ctx, metrics)

	return metrics, nil
}

// getCPUInfo 获取CPU基本信息
// 使用 gopsutil 获取CPU核心数和逻辑CPU数
func (c *CPUCollector) getCPUInfo(ctx context.Context, metrics *types.CPUMetrics) error {
	// 检查上下文是否已取消
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// 获取逻辑CPU数
	logicalCount, err := cpu.CountsWithContext(ctx, true)
	if err != nil {
		return fmt.Errorf("获取逻辑CPU数失败: %w", err)
	}
	metrics.LogicalCount = logicalCount

	// 获取物理CPU数
	physicalCount, err := cpu.CountsWithContext(ctx, false)
	if err != nil {
		return fmt.Errorf("获取物理CPU数失败: %w", err)
	}
	metrics.CoreCount = physicalCount

	return nil
}

// getCPUStats 获取CPU使用率统计
// 使用 gopsutil 获取详细的CPU使用率信息
func (c *CPUCollector) getCPUStats(ctx context.Context, metrics *types.CPUMetrics) error {
	// 检查上下文是否已取消
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// 获取总CPU使用率
	percentages, err := cpu.PercentWithContext(ctx, time.Second, false)
	if err != nil {
		return fmt.Errorf("获取CPU使用率失败: %w", err)
	}
	
	if len(percentages) > 0 {
		metrics.UsagePercent = percentages[0]
	}

	// 获取详细的CPU时间统计
	times, err := cpu.TimesWithContext(ctx, false)
	if err != nil {
		return fmt.Errorf("获取CPU时间统计失败: %w", err)
	}

	if len(times) > 0 {
		cpuTime := times[0]
		total := cpuTime.User + cpuTime.System + cpuTime.Idle + cpuTime.Nice + 
				cpuTime.Iowait + cpuTime.Irq + cpuTime.Softirq + cpuTime.Steal

		if total > 0 {
			metrics.UserPercent = (cpuTime.User / total) * 100
			metrics.SystemPercent = (cpuTime.System / total) * 100
			metrics.IdlePercent = (cpuTime.Idle / total) * 100
			metrics.IOWaitPercent = (cpuTime.Iowait / total) * 100
			metrics.IrqPercent = (cpuTime.Irq / total) * 100
			metrics.SoftIrqPercent = (cpuTime.Softirq / total) * 100
		}
	}

	return nil
	}

// getLoadAverage 获取负载平均值
// 使用 gopsutil 获取系统负载平均值（主要在Linux/Unix系统有效）
func (c *CPUCollector) getLoadAverage(ctx context.Context, metrics *types.CPUMetrics) {
	// 检查上下文是否已取消
	select {
	case <-ctx.Done():
		return
	default:
	}

	// 获取负载平均值，在Windows系统上可能不支持
	loadAvg, err := load.AvgWithContext(ctx)
	if err != nil {
		// 负载平均值获取失败不影响其他数据，设置为0
		metrics.LoadAvg1 = 0
		metrics.LoadAvg5 = 0
		metrics.LoadAvg15 = 0
		return
	}

		metrics.LoadAvg1 = loadAvg.Load1
		metrics.LoadAvg5 = loadAvg.Load5
		metrics.LoadAvg15 = loadAvg.Load15
}

// SetTimeout 设置采集超时时间
func (c *CPUCollector) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// GetTimeout 获取采集超时时间
func (c *CPUCollector) GetTimeout() time.Duration {
	return c.timeout
} 