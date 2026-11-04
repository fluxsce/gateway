package cpu

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gateway/pkg/metric/collector"
	"gateway/pkg/metric/types"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/load"
)

const (
	// DefaultCollectTimeout 默认采集超时时间
	DefaultCollectTimeout = 5 * time.Second
	// 首次没有基线时，用这段时间做一次差值。之后每次采集覆盖距上次的间隔。
	firstSampleInterval = time.Second
)

// CPUCollector CPU采集器
// 使用 gopsutil 实现跨平台CPU信息采集
type CPUCollector struct {
	*collector.BaseCollector
	// 采集超时时间
	timeout time.Duration
	// 保护 lastTimes
	mu sync.Mutex
	// 上次聚合 CPU 时间，用于计算间隔内的占比
	lastTimes *cpu.TimesStat
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

// getCPUStats 获取CPU使用率统计。
// 分项和总使用率都来自相邻两次 cpu.Times 的差值，不再把开机至今的累计时间当成当前占比。
func (c *CPUCollector) getCPUStats(ctx context.Context, metrics *types.CPUMetrics) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	current, err := readAggregateTimes(ctx)
	if err != nil {
		return err
	}

	c.mu.Lock()
	previous := c.lastTimes
	c.mu.Unlock()

	if previous == nil {
		if err := waitInterval(ctx, firstSampleInterval); err != nil {
			c.storeTimes(current)
			return fmt.Errorf("CPU采样被取消: %w", err)
		}
		second, err := readAggregateTimes(ctx)
		if err != nil {
			c.storeTimes(current)
			return err
		}
		applyCPUPercents(metrics, current, second)
		c.storeTimes(second)
		return nil
	}

	applyCPUPercents(metrics, *previous, current)
	c.storeTimes(current)
	return nil
}

func (c *CPUCollector) storeTimes(sample cpu.TimesStat) {
	c.mu.Lock()
	c.lastTimes = &sample
	c.mu.Unlock()
}

func readAggregateTimes(ctx context.Context) (cpu.TimesStat, error) {
	times, err := cpu.TimesWithContext(ctx, false)
	if err != nil {
		return cpu.TimesStat{}, fmt.Errorf("获取CPU时间统计失败: %w", err)
	}
	if len(times) == 0 {
		return cpu.TimesStat{}, fmt.Errorf("获取CPU时间统计失败: 空结果")
	}
	return times[0], nil
}

func waitInterval(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// applyCPUPercents 用两次累计 CPU 时间的差值填写占比。
// Linux 的 user 已包含 guest，这里不再把 Guest 加进分母。nice 并入用户态。steal 留在分母里，和 gopsutil 的忙占比一致。
func applyCPUPercents(metrics *types.CPUMetrics, prev, cur cpu.TimesStat) {
	user := nonNeg(cur.User - prev.User)
	nice := nonNeg(cur.Nice - prev.Nice)
	system := nonNeg(cur.System - prev.System)
	idle := nonNeg(cur.Idle - prev.Idle)
	iowait := nonNeg(cur.Iowait - prev.Iowait)
	irq := nonNeg(cur.Irq - prev.Irq)
	softirq := nonNeg(cur.Softirq - prev.Softirq)
	steal := nonNeg(cur.Steal - prev.Steal)

	total := user + nice + system + idle + iowait + irq + softirq + steal
	if total <= 0 {
		return
	}

	metrics.UserPercent = portion(user+nice, total)
	metrics.SystemPercent = portion(system, total)
	metrics.IdlePercent = portion(idle, total)
	metrics.IOWaitPercent = portion(iowait, total)
	metrics.IrqPercent = portion(irq, total)
	metrics.SoftIrqPercent = portion(softirq, total)
	metrics.UsagePercent = portion(user+nice+system+irq+softirq+steal, total)
}

func nonNeg(v float64) float64 {
	if v < 0 {
		return 0
	}
	return v
}

func portion(part, total float64) float64 {
	if part <= 0 || total <= 0 {
		return 0
	}
	value := part / total * 100
	if value > 100 {
		return 100
	}
	return value
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
