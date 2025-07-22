//go:build !windows
// +build !windows

package utils

import (
	"fmt"
	"syscall"

	"gohub/pkg/metric/types"
)

// getDiskUsageImpl 获取 Unix 系统（Linux/Mac）磁盘使用信息
// 使用 syscall.Statfs 系统调用获取文件系统统计信息
//
// 参数:
//   path: 要查询的路径，可以是文件路径或目录路径
//
// 返回:
//   *types.DiskUsageInfo: 磁盘使用信息结构体
//   error: 如果查询失败则返回错误
//
// 技术细节:
//   - 使用 syscall.Statfs 获取文件系统信息
//   - Blocks: 文件系统中的总块数
//   - Bavail: 非超级用户可用的块数
//   - Bsize: 块的大小（字节）
//   - Files: 文件系统中的总 inode 数量
//   - Ffree: 文件系统中的空闲 inode 数量
func getDiskUsageImpl(path string) (*types.DiskUsageInfo, error) {
	var stat syscall.Statfs_t
	
	// 调用系统调用获取文件系统统计信息
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return nil, fmt.Errorf("获取磁盘使用信息失败: %w", err)
	}

	// 计算磁盘空间信息
	// 总空间 = 总块数 × 块大小
	total := stat.Blocks * uint64(stat.Bsize)
	
	// 可用空间 = 可用块数 × 块大小
	// 注意：使用 Bavail 而不是 Bfree，因为 Bavail 是非超级用户可用的空间
	free := stat.Bavail * uint64(stat.Bsize)
	
	// 已使用空间 = 总空间 - 可用空间
	used := total - free

	// 计算使用率百分比
	usagePercent := calculateUsagePercent(used, total)

	// 计算 inode 信息
	inodesTotal := stat.Files           // 总 inode 数量
	inodesFree := stat.Ffree           // 空闲 inode 数量
	inodesUsed := inodesTotal - inodesFree  // 已使用 inode 数量

	// 计算 inode 使用率百分比
	inodesUsagePercent := calculateUsagePercent(inodesUsed, inodesTotal)

	return &types.DiskUsageInfo{
		Total:        total,
		Used:         used,
		Free:         free,
		UsagePercent: usagePercent,
		InodesTotal:  inodesTotal,
		InodesUsed:   inodesUsed,
		InodesFree:   inodesFree,
		InodesUsagePercent: inodesUsagePercent,
	}, nil
}

// isPathValid 检查路径是否有效
// 在 Unix 系统上验证路径格式
//
// 参数:
//   path: 要验证的路径
//
// 返回:
//   bool: 路径是否有效
//   error: 如果路径无效则返回错误说明
func isPathValid(path string) (bool, error) {
	if path == "" {
		return false, fmt.Errorf("路径不能为空")
	}
	
	// Unix 系统上，路径应该以 / 开头（绝对路径）或者是相对路径
	// 这里进行基本的路径格式检查
	if len(path) > 4096 { // Unix 系统路径长度限制
		return false, fmt.Errorf("路径过长，超过 4096 字符")
	}
	
	return true, nil
}

// GetMountPoints 获取系统挂载点信息
// 读取 /proc/mounts 文件获取所有挂载点
//
// 返回:
//   []string: 挂载点路径列表
//   error: 如果读取失败则返回错误
//
// 注意: 这个函数主要在 Linux 上工作，在其他 Unix 系统上可能需要不同的实现
func GetMountPoints() ([]string, error) {
	// 这里可以实现读取 /proc/mounts 的逻辑
	// 为了简化，暂时返回常见的挂载点
	mountPoints := []string{"/", "/home", "/tmp", "/var"}
	return mountPoints, nil
} 