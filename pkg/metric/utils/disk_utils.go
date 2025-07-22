package utils

import (
	"fmt"

	"gohub/pkg/metric/types"
)

// GetDiskUsage 获取指定路径的磁盘使用信息
// 这是一个跨平台的统一接口，内部会根据操作系统调用相应的实现
//
// 参数:
//   path: 要查询的路径，可以是文件路径或目录路径
//
// 返回:
//   *types.DiskUsageInfo: 磁盘使用信息结构体
//   error: 如果查询失败则返回错误
//
// 使用示例:
//   info, err := GetDiskUsage("/")        // Linux/Mac
//   info, err := GetDiskUsage("C:")       // Windows
//
// 注意: 具体实现在平台特定的文件中：
//   - Windows: disk_utils_windows.go
//   - Unix/Linux: disk_utils_unix.go
func GetDiskUsage(path string) (*types.DiskUsageInfo, error) {
	return getDiskUsageImpl(path)
}

// calculateUsagePercent 计算使用率百分比
// 这是一个工具函数，用于计算百分比并处理除零情况
//
// 参数:
//   used: 已使用的数量
//   total: 总数量
//
// 返回:
//   float64: 使用率百分比 (0-100)
func calculateUsagePercent(used, total uint64) float64 {
	if total == 0 {
		return 0.0
	}
	return (float64(used) / float64(total)) * 100
}

// FormatDiskUsage 格式化磁盘使用信息为可读的字符串
// 提供友好的磁盘使用信息显示格式
//
// 参数:
//   info: 磁盘使用信息结构体
//
// 返回:
//   string: 格式化后的字符串
//
// 示例输出:
//   "总空间: 100.0 GB, 已使用: 45.2 GB (45.2%), 可用: 54.8 GB"
func FormatDiskUsage(info *types.DiskUsageInfo) string {
	if info == nil {
		return "无效的磁盘信息"
	}
	
	return fmt.Sprintf("总空间: %s, 已使用: %s (%.1f%%), 可用: %s",
		FormatBytes(info.Total),
		FormatBytes(info.Used),
		info.UsagePercent,
		FormatBytes(info.Free),
	)
}

// FormatDiskUsageDetailed 格式化磁盘使用信息为详细的字符串
// 包含 inode 信息的详细显示格式
//
// 参数:
//   info: 磁盘使用信息结构体
//
// 返回:
//   string: 详细的格式化字符串
func FormatDiskUsageDetailed(info *types.DiskUsageInfo) string {
	if info == nil {
		return "无效的磁盘信息"
	}
	
	result := FormatDiskUsage(info)
	
	// 如果有 inode 信息，添加到结果中
	if info.InodesTotal > 0 {
		result += fmt.Sprintf("\ninode: 总数: %d, 已使用: %d (%.1f%%), 可用: %d",
			info.InodesTotal,
			info.InodesUsed,
			info.InodesUsagePercent,
			info.InodesFree,
		)
	}
	
	return result
}

// GetDiskUsageMultiple 获取多个路径的磁盘使用信息
// 批量查询多个路径的磁盘使用情况
//
// 参数:
//   paths: 要查询的路径列表
//
// 返回:
//   map[string]*types.DiskUsageInfo: 路径到磁盘使用信息的映射
//   error: 如果查询过程中发生错误
func GetDiskUsageMultiple(paths []string) (map[string]*types.DiskUsageInfo, error) {
	result := make(map[string]*types.DiskUsageInfo)
	
	for _, path := range paths {
		info, err := GetDiskUsage(path)
		if err != nil {
			return nil, fmt.Errorf("获取路径 %s 的磁盘信息失败: %w", path, err)
		}
		result[path] = info
	}
	
	return result, nil
}

// GetDiskUsageThreshold 检查磁盘使用率是否超过阈值
// 用于磁盘空间监控和告警
//
// 参数:
//   path: 要检查的路径
//   threshold: 阈值百分比 (0-100)
//
// 返回:
//   bool: 是否超过阈值
//   *types.DiskUsageInfo: 磁盘使用信息
//   error: 如果查询失败则返回错误
func GetDiskUsageThreshold(path string, threshold float64) (bool, *types.DiskUsageInfo, error) {
	info, err := GetDiskUsage(path)
	if err != nil {
		return false, nil, err
	}
	
	return info.UsagePercent > threshold, info, nil
} 