package utils

import (
	"fmt"
	"time"

	"gateway/pkg/metric/types"
)

// FormatBytes 格式化字节数为人类可读的格式
func FormatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}
	if exp >= len(units) {
		exp = len(units) - 1
	}

	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}

// FormatPercentage 格式化百分比
func FormatPercentage(percentage float64) string {
	return fmt.Sprintf("%.2f%%", percentage)
}

// FormatDuration 格式化持续时间
func FormatDuration(seconds uint64) string {
	duration := time.Duration(seconds) * time.Second

	days := int(duration.Hours()) / 24
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60
	secs := int(duration.Seconds()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, secs)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, secs)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, secs)
	}
	return fmt.Sprintf("%ds", secs)
}

// FormatTimestamp 格式化时间戳
func FormatTimestamp(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// FormatCPUMetrics 格式化CPU指标
func FormatCPUMetrics(metrics *types.CPUMetrics) string {
	if metrics == nil {
		return "CPU指标为空"
	}

	return fmt.Sprintf(`CPU指标:
  使用率: %s
  用户态: %s
  系统态: %s
  空闲: %s
  I/O等待: %s
  CPU核心数: %d
  逻辑CPU数: %d
  负载平均值: %.2f, %.2f, %.2f
  采集时间: %s`,
		FormatPercentage(metrics.UsagePercent),
		FormatPercentage(metrics.UserPercent),
		FormatPercentage(metrics.SystemPercent),
		FormatPercentage(metrics.IdlePercent),
		FormatPercentage(metrics.IOWaitPercent),
		metrics.CoreCount,
		metrics.LogicalCount,
		metrics.LoadAvg1,
		metrics.LoadAvg5,
		metrics.LoadAvg15,
		FormatTimestamp(metrics.CollectTime))
}

// FormatMemoryMetrics 格式化内存指标
func FormatMemoryMetrics(metrics *types.MemoryMetrics) string {
	if metrics == nil {
		return "内存指标为空"
	}

	return fmt.Sprintf(`内存指标:
  总内存: %s
  已使用: %s (%s)
  可用: %s
  空闲: %s
  缓存: %s
  缓冲区: %s
  交换区总大小: %s
  交换区已使用: %s (%s)
  采集时间: %s`,
		FormatBytes(metrics.Total),
		FormatBytes(metrics.Used),
		FormatPercentage(metrics.UsagePercent),
		FormatBytes(metrics.Available),
		FormatBytes(metrics.Free),
		FormatBytes(metrics.Cached),
		FormatBytes(metrics.Buffers),
		FormatBytes(metrics.SwapTotal),
		FormatBytes(metrics.SwapUsed),
		FormatPercentage(metrics.SwapUsagePercent),
		FormatTimestamp(metrics.CollectTime))
}

// FormatDiskMetrics 格式化磁盘指标
func FormatDiskMetrics(metrics *types.DiskMetrics) string {
	if metrics == nil {
		return "磁盘指标为空"
	}

	result := fmt.Sprintf("磁盘指标:\n")

	// 格式化分区信息
	result += "  分区信息:\n"
	for _, partition := range metrics.Partitions {
		result += fmt.Sprintf("    %s (%s) - %s:\n",
			partition.Device,
			partition.MountPoint,
			partition.FileSystem)
		result += fmt.Sprintf("      总大小: %s\n", FormatBytes(partition.Total))
		result += fmt.Sprintf("      已使用: %s (%s)\n",
			FormatBytes(partition.Used),
			FormatPercentage(partition.UsagePercent))
		result += fmt.Sprintf("      可用: %s\n", FormatBytes(partition.Free))
		if partition.InodesTotal > 0 {
			result += fmt.Sprintf("      inode使用: %d/%d (%s)\n",
				partition.InodesUsed,
				partition.InodesTotal,
				FormatPercentage(partition.InodesUsagePercent))
		}
		result += "\n"
	}

	// 格式化IO统计
	if len(metrics.IOStats) > 0 {
		result += "  IO统计:\n"
		for _, iostat := range metrics.IOStats {
			result += fmt.Sprintf("    %s:\n", iostat.Device)
			result += fmt.Sprintf("      读取: %d次 %s\n",
				iostat.ReadCount,
				FormatBytes(iostat.ReadBytes))
			result += fmt.Sprintf("      写入: %d次 %s\n",
				iostat.WriteCount,
				FormatBytes(iostat.WriteBytes))
			result += fmt.Sprintf("      IO时间: %dms\n", iostat.IOTime)
			result += "\n"
		}
	}

	result += fmt.Sprintf("  采集时间: %s", FormatTimestamp(metrics.CollectTime))
	return result
}

// FormatNetworkMetrics 格式化网络指标
func FormatNetworkMetrics(metrics *types.NetworkMetrics) string {
	if metrics == nil {
		return "网络指标为空"
	}

	result := fmt.Sprintf("网络指标:\n")

	for _, iface := range metrics.Interfaces {
		result += fmt.Sprintf("  接口: %s (%s)\n", iface.Name, iface.Status)
		result += fmt.Sprintf("    MAC地址: %s\n", iface.HardwareAddr)
		result += fmt.Sprintf("    类型: %s\n", iface.Type)

		if len(iface.IPAddresses) > 0 {
			result += fmt.Sprintf("    IP地址: %v\n", iface.IPAddresses)
		}

		result += fmt.Sprintf("    接收: %s (%d包)\n",
			FormatBytes(iface.BytesReceived),
			iface.PacketsReceived)
		result += fmt.Sprintf("    发送: %s (%d包)\n",
			FormatBytes(iface.BytesSent),
			iface.PacketsSent)

		if iface.ErrorsReceived > 0 || iface.ErrorsSent > 0 {
			result += fmt.Sprintf("    错误: 接收%d 发送%d\n",
				iface.ErrorsReceived,
				iface.ErrorsSent)
		}

		if iface.DroppedReceived > 0 || iface.DroppedSent > 0 {
			result += fmt.Sprintf("    丢包: 接收%d 发送%d\n",
				iface.DroppedReceived,
				iface.DroppedSent)
		}

		result += "\n"
	}

	result += fmt.Sprintf("  采集时间: %s", FormatTimestamp(metrics.CollectTime))
	return result
}

// FormatSystemMetrics 格式化系统指标
func FormatSystemMetrics(metrics *types.SystemMetrics) string {
	if metrics == nil {
		return "系统指标为空"
	}

	result := fmt.Sprintf(`系统指标:
  主机名: %s
  操作系统: %s %s
  架构: %s
  内核版本: %s
  启动时间: %s
  运行时间: %s
  进程数: %d
  用户数: %d`,
		metrics.Hostname,
		metrics.OS,
		metrics.OSVersion,
		metrics.Architecture,
		metrics.KernelVersion,
		FormatTimestamp(metrics.BootTime),
		FormatDuration(metrics.Uptime),
		metrics.ProcessCount,
		metrics.UserCount)

	// 格式化温度信息
	if len(metrics.Temperature) > 0 {
		result += "\n  温度信息:\n"
		for _, temp := range metrics.Temperature {
			result += fmt.Sprintf("    %s: %.1f°C", temp.SensorName, temp.Temperature)
			if temp.High > 0 {
				result += fmt.Sprintf(" (高温阈值: %.1f°C)", temp.High)
			}
			if temp.Critical > 0 {
				result += fmt.Sprintf(" (严重阈值: %.1f°C)", temp.Critical)
			}
			result += "\n"
		}
	}

	result += fmt.Sprintf("\n  采集时间: %s", FormatTimestamp(metrics.CollectTime))
	return result
}

// FormatProcessMetrics 格式化进程指标
func FormatProcessMetrics(metrics *types.ProcessMetrics) string {
	if metrics == nil {
		return "进程指标为空"
	}

	result := fmt.Sprintf("进程指标:\n")

	// 格式化当前进程信息
	if metrics.CurrentProcess != nil {
		proc := metrics.CurrentProcess
		result += fmt.Sprintf("  当前进程:\n")
		result += fmt.Sprintf("    PID: %d\n", proc.PID)
		result += fmt.Sprintf("    名称: %s\n", proc.Name)
		result += fmt.Sprintf("    状态: %s\n", proc.Status)
		result += fmt.Sprintf("    父进程ID: %d\n", proc.PPID)
		result += fmt.Sprintf("    启动时间: %s\n", FormatTimestamp(proc.CreateTime))
		result += fmt.Sprintf("    运行时间: %s\n", FormatDuration(proc.RunTime))
		result += fmt.Sprintf("    内存使用: %s\n", FormatBytes(proc.MemoryUsage))
		result += fmt.Sprintf("    CPU使用率: %s\n", FormatPercentage(proc.CPUPercent))
		result += fmt.Sprintf("    线程数: %d\n", proc.ThreadCount)
		result += fmt.Sprintf("    文件描述符数: %d\n", proc.FileDescriptorCount)

		if proc.ExecutablePath != "" {
			result += fmt.Sprintf("    可执行文件: %s\n", proc.ExecutablePath)
		}

		if proc.WorkingDirectory != "" {
			result += fmt.Sprintf("    工作目录: %s\n", proc.WorkingDirectory)
		}

		result += "\n"
	}

	// 格式化系统进程统计
	if metrics.SystemProcesses != nil {
		stats := metrics.SystemProcesses
		result += fmt.Sprintf("  系统进程统计:\n")
		result += fmt.Sprintf("    总进程数: %d\n", stats.Total)
		result += fmt.Sprintf("    运行中: %d\n", stats.Running)
		result += fmt.Sprintf("    睡眠中: %d\n", stats.Sleeping)
		result += fmt.Sprintf("    已停止: %d\n", stats.Stopped)
		result += fmt.Sprintf("    僵尸进程: %d\n", stats.Zombie)
		result += "\n"
	}

	result += fmt.Sprintf("  采集时间: %s", FormatTimestamp(metrics.CollectTime))
	return result
}

// FormatAllMetrics 格式化所有指标
func FormatAllMetrics(metrics *types.AllMetrics) string {
	if metrics == nil {
		return "指标为空"
	}

	result := fmt.Sprintf("=== 系统资源指标 ===\n")
	result += fmt.Sprintf("采集时间: %s\n\n", FormatTimestamp(metrics.CollectTime))

	if metrics.CPU != nil {
		result += FormatCPUMetrics(metrics.CPU) + "\n\n"
	}

	if metrics.Memory != nil {
		result += FormatMemoryMetrics(metrics.Memory) + "\n\n"
	}

	if metrics.Disk != nil {
		result += FormatDiskMetrics(metrics.Disk) + "\n\n"
	}

	if metrics.Network != nil {
		result += FormatNetworkMetrics(metrics.Network) + "\n\n"
	}

	if metrics.System != nil {
		result += FormatSystemMetrics(metrics.System) + "\n\n"
	}

	if metrics.Process != nil {
		result += FormatProcessMetrics(metrics.Process) + "\n\n"
	}

	return result
}
