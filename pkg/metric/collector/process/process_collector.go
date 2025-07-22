package process

import (
	"context"
	"fmt"
	"os"
	"time"

	"gohub/pkg/metric/collector"
	"gohub/pkg/metric/types"

	"github.com/shirou/gopsutil/v4/process"
)

// 默认采集超时时间
const DefaultCollectTimeout = 5 * time.Second

// ProcessCollector 进程信息采集器
// 使用 gopsutil 实现跨平台进程信息采集，包括当前进程详细信息和系统进程统计
type ProcessCollector struct {
	*collector.BaseCollector
	// 采集超时时间，防止进程信息获取操作阻塞过久
	timeout time.Duration
}

// NewProcessCollector 创建进程信息采集器
// 使用 gopsutil 实现跨平台进程信息采集
// 
// 返回值:
//   - *ProcessCollector: 新创建的进程采集器实例
func NewProcessCollector() *ProcessCollector {
	return &ProcessCollector{
		BaseCollector: collector.NewBaseCollector(
			types.CollectorNameProcess,
			"基于gopsutil的进程信息采集器，提供跨平台进程信息采集",
		),
		timeout: DefaultCollectTimeout,
	}
}

// Collect 执行采集
// 统一的采集入口，使用gopsutil简化实现
// 
// 返回值:
//   - interface{}: 采集到的进程指标数据
//   - error: 采集过程中的错误
func (c *ProcessCollector) Collect() (interface{}, error) {
	if !c.IsEnabled() {
		return nil, types.ErrCollectorDisabled
	}

	metrics, err := c.GetProcessInfo()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", types.ErrCollectFailed, err)
	}

	c.SetLastCollectTime(time.Now())
	return metrics, nil
}

// GetProcessInfo 获取进程信息
// 使用 gopsutil 采集当前进程信息和系统进程统计
// 
// 返回值:
//   - *types.ProcessMetrics: 进程指标数据，包括当前进程信息和系统进程统计
//   - error: 获取过程中的错误
func (c *ProcessCollector) GetProcessInfo() (*types.ProcessMetrics, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	metrics := &types.ProcessMetrics{
		CollectTime: time.Now(),
	}

	// 获取当前进程信息
	currentProcess, err := c.getCurrentProcessInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取当前进程信息失败: %w", err)
	}
	metrics.CurrentProcess = currentProcess

	// 获取系统进程统计
	systemProcesses, err := c.getSystemProcessStats(ctx)
	if err != nil {
		// 系统进程统计获取失败不影响当前进程信息，设置默认值
		metrics.SystemProcesses = &types.ProcessSystemStats{}
	} else {
		metrics.SystemProcesses = systemProcesses
	}

	return metrics, nil
}

// getCurrentProcessInfo 获取当前进程信息
// 使用 gopsutil 获取当前进程的详细信息，包括内存使用、CPU使用率、线程数等
// 
// 参数:
//   - ctx: 上下文，用于控制操作超时
// 
// 返回值:
//   - *types.ProcessInfo: 当前进程的详细信息
//   - error: 获取过程中的错误
func (c *ProcessCollector) getCurrentProcessInfo(ctx context.Context) (*types.ProcessInfo, error) {
	// 检查上下文是否已取消
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// 获取当前进程PID
	pid := int32(os.Getpid())
	
	// 使用 gopsutil 创建进程对象
	proc, err := process.NewProcessWithContext(ctx, pid)
	if err != nil {
		return nil, fmt.Errorf("gopsutil创建进程对象失败: %w", err)
	}
	
	processInfo := &types.ProcessInfo{
		PID: pid,
	}

	// 获取进程名称
	if name, err := proc.NameWithContext(ctx); err == nil {
		processInfo.Name = name
}

	// 获取父进程ID
	if ppid, err := proc.PpidWithContext(ctx); err == nil {
		processInfo.PPID = ppid
	}

	// 获取进程状态
	if statusList, err := proc.StatusWithContext(ctx); err == nil && len(statusList) > 0 {
		processInfo.Status = c.convertProcessStatus(statusList[0])
	} else {
		// 当获取进程状态失败时，设置默认值为 "unknown"
		processInfo.Status = "unknown"
	}

	// 获取进程创建时间
	if createTime, err := proc.CreateTimeWithContext(ctx); err == nil {
		// createTime 是 Unix 时间戳（毫秒）
		processInfo.CreateTime = time.Unix(0, createTime*int64(time.Millisecond))
		processInfo.RunTime = uint64(time.Since(processInfo.CreateTime).Seconds())
	}

	// 获取内存使用信息
	if memInfo, err := proc.MemoryInfoWithContext(ctx); err == nil {
		processInfo.MemoryUsage = memInfo.RSS // 实际物理内存使用量
	}

	// 获取内存使用率
	if memPercent, err := proc.MemoryPercentWithContext(ctx); err == nil {
		processInfo.MemoryPercent = float64(memPercent)
	}

	// 获取CPU使用率
	if cpuPercent, err := proc.CPUPercentWithContext(ctx); err == nil {
		processInfo.CPUPercent = cpuPercent
	}

	// 获取线程数
	if numThreads, err := proc.NumThreadsWithContext(ctx); err == nil {
		processInfo.ThreadCount = numThreads
	}

	// 获取文件描述符数量（Unix系统）
	if numFDs, err := proc.NumFDsWithContext(ctx); err == nil {
		processInfo.FileDescriptorCount = numFDs
	}

	// 获取命令行参数
	if cmdline, err := proc.CmdlineSliceWithContext(ctx); err == nil {
		processInfo.CommandLine = cmdline
	}

	// 获取可执行文件路径
	if exe, err := proc.ExeWithContext(ctx); err == nil {
		processInfo.ExecutablePath = exe
	}

	// 获取工作目录
	if cwd, err := proc.CwdWithContext(ctx); err == nil {
		processInfo.WorkingDirectory = cwd
	}

	return processInfo, nil
}

// getSystemProcessStats 获取系统进程统计
// 使用 gopsutil 获取系统中所有进程的统计信息，包括运行、睡眠、停止、僵尸进程数量
// 
// 参数:
//   - ctx: 上下文，用于控制操作超时
// 
// 返回值:
//   - *types.ProcessSystemStats: 系统进程统计信息
//   - error: 获取过程中的错误
func (c *ProcessCollector) getSystemProcessStats(ctx context.Context) (*types.ProcessSystemStats, error) {
	// 检查上下文是否已取消
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	stats := &types.ProcessSystemStats{}

	// 使用 gopsutil 获取所有进程列表
	processes, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("gopsutil获取进程列表失败: %w", err)
	}

	// 统计各种状态的进程数量
	for _, proc := range processes {
		// 获取进程状态，失败时跳过该进程
		statusList, err := proc.StatusWithContext(ctx)
		if err != nil || len(statusList) == 0 {
			continue
		}

		// 根据状态分类统计（取第一个状态）
		switch c.convertProcessStatus(statusList[0]) {
		case "running":
					stats.Running++
		case "sleeping":
					stats.Sleeping++
		case "stopped":
					stats.Stopped++
		case "zombie":
					stats.Zombie++
				}
				stats.Total++
	}

	return stats, nil
}

// convertProcessStatus 转换进程状态
// 将 gopsutil 返回的进程状态转换为内部使用的标准状态字符串
// 
// 参数:
//   - status: gopsutil 返回的进程状态字符串
// 
// 返回值:
//   - string: 标准化的进程状态字符串
func (c *ProcessCollector) convertProcessStatus(status string) string {
	switch status {
	case "R", "running":
		return "running"
	case "S", "sleeping":
		return "sleeping"
	case "T", "stopped":
		return "stopped"
	case "Z", "zombie":
		return "zombie"
	case "D", "disk-sleep":
		return "sleeping" // 不可中断睡眠归类为睡眠
	case "I", "idle":
		return "sleeping" // 空闲状态归类为睡眠
	default:
		return "unknown"
	}
}

// SetTimeout 设置采集超时时间
// 用于控制进程信息采集操作的最大执行时间
// 
// 参数:
//   - timeout: 超时时间
func (c *ProcessCollector) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// GetTimeout 获取采集超时时间
// 返回当前设置的采集超时时间
// 
// 返回值:
//   - time.Duration: 当前的超时时间设置
func (c *ProcessCollector) GetTimeout() time.Duration {
	return c.timeout
}

// GetCurrentPID 获取当前进程ID
// 这是一个便捷方法，直接返回当前进程的PID
// 
// 返回值:
//   - int32: 当前进程的PID
func (c *ProcessCollector) GetCurrentPID() int32 {
	return int32(os.Getpid())
}

// GetCurrentProcessName 获取当前进程名称
// 这是一个便捷方法，用于快速获取当前进程的名称
// 
// 返回值:
//   - string: 当前进程的名称
//   - error: 获取过程中的错误
func (c *ProcessCollector) GetCurrentProcessName() (string, error) {
	processInfo, err := c.getCurrentProcessInfo(context.Background())
	if err != nil {
		return "", err
	}
	return processInfo.Name, nil
}

// GetProcessCount 获取系统进程总数
// 这是一个便捷方法，用于快速获取系统中运行的进程总数
// 
// 返回值:
//   - uint32: 系统进程总数
//   - error: 获取过程中的错误
func (c *ProcessCollector) GetProcessCount() (uint32, error) {
	stats, err := c.getSystemProcessStats(context.Background())
	if err != nil {
		return 0, err
	}
	return stats.Total, nil
}

// GetCurrentProcessMemoryUsage 获取当前进程内存使用量
// 这是一个便捷方法，用于快速获取当前进程的内存使用情况
// 
// 返回值:
//   - uint64: 内存使用量（字节）
//   - float64: 内存使用率（百分比）
//   - error: 获取过程中的错误
func (c *ProcessCollector) GetCurrentProcessMemoryUsage() (uint64, float64, error) {
	processInfo, err := c.getCurrentProcessInfo(context.Background())
	if err != nil {
		return 0, 0, err
	}
	return processInfo.MemoryUsage, processInfo.MemoryPercent, nil
}

// GetCurrentProcessCPUUsage 获取当前进程CPU使用率
// 这是一个便捷方法，用于快速获取当前进程的CPU使用率
// 
// 返回值:
//   - float64: CPU使用率（百分比）
//   - error: 获取过程中的错误
func (c *ProcessCollector) GetCurrentProcessCPUUsage() (float64, error) {
	processInfo, err := c.getCurrentProcessInfo(context.Background())
	if err != nil {
		return 0, err
	}
	return processInfo.CPUPercent, nil
}

// GetProcessesByStatus 根据状态获取进程数量
// 这是一个便捷方法，用于获取特定状态的进程数量
// 
// 参数:
//   - status: 进程状态（"running", "sleeping", "stopped", "zombie"）
// 
// 返回值:
//   - uint32: 指定状态的进程数量
//   - error: 获取过程中的错误
func (c *ProcessCollector) GetProcessesByStatus(status string) (uint32, error) {
	stats, err := c.getSystemProcessStats(context.Background())
	if err != nil {
		return 0, err
	}

	switch status {
	case "running":
		return stats.Running, nil
	case "sleeping":
		return stats.Sleeping, nil
	case "stopped":
		return stats.Stopped, nil
	case "zombie":
		return stats.Zombie, nil
	default:
		return 0, fmt.Errorf("不支持的进程状态: %s", status)
	}
} 