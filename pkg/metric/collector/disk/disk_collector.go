// Package disk 提供磁盘指标采集功能
// 使用 gopsutil 实现跨平台磁盘信息采集
package disk

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"gateway/pkg/metric/collector"
	"gateway/pkg/metric/types"
	"gateway/pkg/metric/utils"

	"github.com/shirou/gopsutil/v4/disk"
)

// 磁盘采集相关常量
const (
	// 默认采集超时时间
	DefaultCollectTimeout = 5 * time.Second

	// 扇区大小（字节）
	SectorSize = 512
)

// DiskCollector 磁盘采集器
// 使用 gopsutil 实现跨平台磁盘信息采集
type DiskCollector struct {
	*collector.BaseCollector
	// 采集超时时间，防止采集操作阻塞过久
	timeout time.Duration
	// 上次IO统计数据
	lastIOStats map[string]types.DiskIOStats
	// 互斥锁保护lastIOStats
	mu sync.RWMutex
	// 是否是第一次采集
	isFirstCollect bool
}

// NewDiskCollector 创建磁盘采集器
// 使用 gopsutil 实现跨平台磁盘信息采集
func NewDiskCollector() *DiskCollector {
	return &DiskCollector{
		BaseCollector: collector.NewBaseCollector(
			types.CollectorNameDisk,
			"基于gopsutil的磁盘资源采集器，提供跨平台磁盘信息采集",
		),
		timeout:        DefaultCollectTimeout,
		lastIOStats:    make(map[string]types.DiskIOStats),
		isFirstCollect: true,
	}
}

// Collect 执行采集操作
// 这是采集器的主要入口点，负责协调磁盘指标采集
// 返回磁盘指标数据或错误信息
func (c *DiskCollector) Collect() (interface{}, error) {
	// 检查采集器是否已启用
	if !c.IsEnabled() {
		return nil, types.ErrCollectorDisabled
	}

	// 执行磁盘使用率采集
	metrics, err := c.GetDiskUsage()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", types.ErrCollectFailed, err)
	}

	// 记录最后一次采集时间
	c.SetLastCollectTime(time.Now())
	return metrics, nil
}

// GetDiskUsage 获取磁盘使用率
// 使用 gopsutil 采集所有磁盘的使用情况，包括分区信息和IO统计
// 返回完整的磁盘指标数据
func (c *DiskCollector) GetDiskUsage() (*types.DiskMetrics, error) {
	// 创建带超时的上下文，防止采集操作阻塞
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	// 初始化指标数据结构
	metrics := &types.DiskMetrics{
		CollectTime: time.Now(),
	}

	// 获取磁盘分区信息
	partitions, err := c.getDiskPartitions(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取磁盘分区信息失败: %w", err)
	}
	metrics.Partitions = partitions

	// 获取磁盘IO统计信息（非关键信息，失败不影响整体采集）
	ioStats, err := c.getDiskIOStats(ctx)
	if err != nil {
		// IO统计获取失败不影响分区信息
		metrics.IOStats = []types.DiskIOStats{}
		// 可以在这里记录日志: log.Warnf("获取磁盘IO统计失败: %v", err)
	} else {
		metrics.IOStats = ioStats
	}

	return metrics, nil
}

// SetTimeout 设置采集超时时间
// 用于控制采集操作的最大执行时间，防止阻塞
func (c *DiskCollector) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// GetTimeout 获取采集超时时间
// 返回当前设置的超时时间
func (c *DiskCollector) GetTimeout() time.Duration {
	return c.timeout
}

// getDiskPartitions 获取磁盘分区信息
// 使用 gopsutil 简化跨平台分区信息获取
func (c *DiskCollector) getDiskPartitions(ctx context.Context) ([]types.DiskPartition, error) {
	// 检查上下文是否已取消
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// 使用 gopsutil 获取分区信息
	partitions, err := disk.PartitionsWithContext(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("gopsutil获取分区信息失败: %w", err)
	}

	var result []types.DiskPartition
	for _, partition := range partitions {
		// 跳过虚拟文件系统
		if isVirtualFileSystem(partition.Device, partition.Fstype) {
			continue
		}

		// 获取分区使用信息
		usage, err := disk.UsageWithContext(ctx, partition.Mountpoint)
		if err != nil {
			// 单个分区获取失败不影响其他分区
			continue
		}

		// 转换为内部数据结构
		diskPartition := types.DiskPartition{
			Device:             partition.Device,
			MountPoint:         partition.Mountpoint,
			FileSystem:         partition.Fstype,
			Total:              usage.Total,
			Used:               usage.Used,
			Free:               usage.Free,
			UsagePercent:       usage.UsedPercent,
			InodesTotal:        usage.InodesTotal,
			InodesUsed:         usage.InodesUsed,
			InodesFree:         usage.InodesFree,
			InodesUsagePercent: usage.InodesUsedPercent,
		}

		result = append(result, diskPartition)
	}

	return result, nil
}

// getDiskIOStats 获取磁盘IO统计信息
// 使用 gopsutil 简化跨平台IO统计获取
func (c *DiskCollector) getDiskIOStats(ctx context.Context) ([]types.DiskIOStats, error) {
	// 检查上下文是否已取消
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// 使用 gopsutil 获取IO统计信息
	ioCounters, err := disk.IOCountersWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("gopsutil获取IO统计失败: %w", err)
	}

	now := time.Now()
	var result []types.DiskIOStats

	// 获取读写锁
	c.mu.Lock()
	defer c.mu.Unlock()

	for deviceName, counter := range ioCounters {
		// 只保留整盘。分区和 loop/ram 一类虚拟块设备不计入。
		if skipDiskIODevice(deviceName) {
			continue
		}

		// 转换为内部数据结构
		iostat := types.DiskIOStats{
			Device:          deviceName,
			ReadCount:       counter.ReadCount,
			WriteCount:      counter.WriteCount,
			ReadBytes:       counter.ReadBytes,
			WriteBytes:      counter.WriteBytes,
			ReadTime:        counter.ReadTime,
			WriteTime:       counter.WriteTime,
			IOInProgress:    counter.IopsInProgress,
			IOTime:          counter.IoTime,
			LastCollectTime: now,
		}

		// 计算IO速率。第一次没有基线，速率保持 0。
		if !c.isFirstCollect {
			if last, exists := c.lastIOStats[deviceName]; exists {
				duration := now.Sub(last.LastCollectTime).Seconds()
				iostat.ReadRate = utils.CounterRate(counter.ReadBytes, last.ReadBytes, duration)
				iostat.WriteRate = utils.CounterRate(counter.WriteBytes, last.WriteBytes, duration)
			}
		}

		// 保存当前统计数据用于下次计算
		c.lastIOStats[deviceName] = iostat
		result = append(result, iostat)
	}

	c.isFirstCollect = false
	return result, nil
}

// GetDiskPartitions 获取磁盘分区信息
// 这是一个便捷方法，直接返回分区信息
func (c *DiskCollector) GetDiskPartitions() ([]types.DiskPartition, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	return c.getDiskPartitions(ctx)
}

// GetDiskIOStats 获取磁盘IO统计信息
// 这是一个便捷方法，直接返回IO统计信息
func (c *DiskCollector) GetDiskIOStats() ([]types.DiskIOStats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	return c.getDiskIOStats(ctx)
}

// GetDiskUsageForPath 获取指定路径的磁盘使用信息
// 这是一个便捷方法，用于获取特定路径的磁盘使用情况
func (c *DiskCollector) GetDiskUsageForPath(path string) (*types.DiskUsageInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	// 使用 gopsutil 获取路径使用信息
	usage, err := disk.UsageWithContext(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("gopsutil获取路径使用信息失败: %w", err)
	}

	return &types.DiskUsageInfo{
		Total:              usage.Total,
		Used:               usage.Used,
		Free:               usage.Free,
		UsagePercent:       usage.UsedPercent,
		InodesTotal:        usage.InodesTotal,
		InodesUsed:         usage.InodesUsed,
		InodesFree:         usage.InodesFree,
		InodesUsagePercent: usage.InodesUsedPercent,
	}, nil
}

// IsVirtualFileSystem 检查是否为虚拟文件系统
// 用于过滤不需要统计的虚拟文件系统
func IsVirtualFileSystem(device, fileSystem string) bool {
	return isVirtualFileSystem(device, fileSystem)
}

// IsPartition 检查是否为分区
// 用于区分整个磁盘和分区
func IsPartition(device string) bool {
	return isPartition(device)
}

// isVirtualFileSystem 检查是否为虚拟文件系统（内部函数）
func isVirtualFileSystem(device, fileSystem string) bool {
	// 跳过虚拟文件系统
	virtualFS := []string{"proc", "sysfs", "devtmpfs", "devpts", "tmpfs", "securityfs", "cgroup", "pstore", "bpf", "systemd-1", "mqueue", "hugetlbfs", "debugfs", "tracefs", "fusectl", "configfs", "rpc_pipefs", "binfmt_misc"}
	for _, vfs := range virtualFS {
		if fileSystem == vfs {
			return true
		}
	}

	// 跳过特殊设备
	if strings.HasPrefix(device, "/dev/loop") || strings.HasPrefix(device, "/dev/ram") || strings.HasPrefix(device, "udev") || strings.HasPrefix(device, "none") {
		return true
	}

	return false
}

var (
	partitionNVMe = regexp.MustCompile(`^nvme\d+n\d+p\d+$`)
	partitionMMC  = regexp.MustCompile(`^mmcblk\d+p\d+$`)
	partitionSCSI = regexp.MustCompile(`^(?:sd|hd|vd|xvd)[a-z]+\d+$`)
)

// deviceBaseName 去掉路径，只留内核设备名。
func deviceBaseName(device string) string {
	name := device
	if i := strings.LastIndex(name, "/"); i >= 0 {
		name = name[i+1:]
	}
	return strings.ToLower(name)
}

// isPartition 判断名字是不是磁盘分区。
// 整盘 nvme0n1、mmcblk0、sda 返回 false；分区 nvme0n1p1、mmcblk0p1、sda1 返回 true。
// 不能按“末尾是数字”判断，否则 nvme0n1 整盘会被丢掉。
func isPartition(device string) bool {
	name := deviceBaseName(device)
	if name == "" {
		return false
	}
	return partitionNVMe.MatchString(name) || partitionMMC.MatchString(name) || partitionSCSI.MatchString(name)
}

// skipDiskIODevice 跳过分区和没有整盘意义的虚拟块设备。
func skipDiskIODevice(device string) bool {
	if isPartition(device) {
		return true
	}
	name := deviceBaseName(device)
	for _, prefix := range []string{"loop", "ram", "zram", "fd"} {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}
