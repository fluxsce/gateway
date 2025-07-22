//go:build windows
// +build windows

package utils

import (
	"fmt"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"gohub/pkg/metric/types"
)

// getDiskUsageImpl 获取 Windows 系统磁盘使用信息
// 使用 Windows API GetDiskFreeSpaceExW 获取磁盘空间信息
//
// 参数:
//   path: 要查询的路径，可以是文件路径或目录路径
//
// 返回:
//   *types.DiskUsageInfo: 磁盘使用信息结构体
//   error: 如果查询失败则返回错误
//
// 技术细节:
//   - 使用 Windows API GetDiskFreeSpaceExW 获取磁盘空间信息
//   - lpFreeBytesAvailable: 调用者可用的字节数
//   - lpTotalNumberOfBytes: 磁盘总字节数
//   - lpTotalNumberOfFreeBytes: 磁盘可用字节数
//   - Windows 不使用 inode 概念，所以相关字段为 0
func getDiskUsageImpl(path string) (*types.DiskUsageInfo, error) {
	// 加载 kernel32.dll 动态链接库
	kernel32 := syscall.MustLoadDLL("kernel32.dll")
	defer kernel32.Release()
	
	// 获取 GetDiskFreeSpaceExW 函数
	getDiskFreeSpaceEx := kernel32.MustFindProc("GetDiskFreeSpaceExW")
	
	// 准备参数
	var freeBytes, totalBytes, totalFreeBytes int64
	
	// 处理路径：如果是文件路径，取其目录部分
	diskPath := path
	if !strings.HasSuffix(path, "\\") && !strings.HasSuffix(path, "/") {
		// 检查是否是驱动器根目录格式 (如 "C:")
		if len(path) == 2 && path[1] == ':' {
			diskPath = path + "\\"
		} else {
			// 如果不是驱动器根目录，获取其目录部分
			diskPath = filepath.Dir(path)
		}
	}
	
	// 将路径转换为 UTF-16 指针
	pathPtr, err := syscall.UTF16PtrFromString(diskPath)
	if err != nil {
		return nil, fmt.Errorf("路径转换失败: %w", err)
	}
	
	// 调用 Windows API
	ret, _, err := getDiskFreeSpaceEx.Call(
		uintptr(unsafe.Pointer(pathPtr)),                    // 路径
		uintptr(unsafe.Pointer(&freeBytes)),                 // 可用字节数
		uintptr(unsafe.Pointer(&totalBytes)),                // 总字节数
		uintptr(unsafe.Pointer(&totalFreeBytes)),            // 总可用字节数
	)
	
	// 检查调用是否成功
	if ret == 0 {
		return nil, fmt.Errorf("获取 Windows 磁盘信息失败 (路径: %s): %v", diskPath, err)
	}
	
	// 计算磁盘使用信息
	total := uint64(totalBytes)
	free := uint64(freeBytes)
	used := total - free
	
	// 计算使用率百分比
	usagePercent := calculateUsagePercent(used, total)
	
	return &types.DiskUsageInfo{
		Total:        total,
		Used:         used,
		Free:         free,
		UsagePercent: usagePercent,
		// Windows 不使用 inode 概念，所以相关字段为 0
		InodesTotal:        0,
		InodesUsed:         0,
		InodesFree:         0,
		InodesUsagePercent: 0,
	}, nil
}

// isPathValid 检查路径是否有效
// 在 Windows 系统上验证路径格式
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
	
	// Windows 系统路径长度限制
	if len(path) > 260 { // Windows 标准路径长度限制
		return false, fmt.Errorf("路径过长，超过 260 字符")
	}
	
	// 检查 Windows 路径格式
	if len(path) >= 2 && path[1] == ':' {
		// 驱动器路径格式 (如 "C:" 或 "C:\")
		drive := path[0]
		if !((drive >= 'A' && drive <= 'Z') || (drive >= 'a' && drive <= 'z')) {
			return false, fmt.Errorf("无效的驱动器字母: %c", drive)
		}
	}
	
	// 检查是否包含无效字符
	invalidChars := []string{"<", ">", ":", "\"", "|", "?", "*"}
	for _, char := range invalidChars {
		if strings.Contains(path, char) && !(char == ":" && len(path) >= 2 && path[1] == ':') {
			return false, fmt.Errorf("路径包含无效字符: %s", char)
		}
	}
	
	return true, nil
}

// GetMountPoints 获取系统挂载点信息
// 在 Windows 上获取所有驱动器列表
//
// 返回:
//   []string: 驱动器路径列表 (如 ["C:", "D:", "E:"])
//   error: 如果获取失败则返回错误
func GetMountPoints() ([]string, error) {
	// 加载 kernel32.dll 动态链接库
	kernel32 := syscall.MustLoadDLL("kernel32.dll")
	defer kernel32.Release()
	
	// 获取 GetLogicalDrives 函数
	getLogicalDrives := kernel32.MustFindProc("GetLogicalDrives")
	
	// 调用 API 获取逻辑驱动器位掩码
	ret, _, err := getLogicalDrives.Call()
	if ret == 0 {
		return nil, fmt.Errorf("获取逻辑驱动器失败: %v", err)
	}
	
	var drives []string
	driveMask := uint32(ret)
	
	// 遍历所有可能的驱动器字母 (A-Z)
	for i := 0; i < 26; i++ {
		// 检查对应位是否设置
		if driveMask&(1<<uint(i)) != 0 {
			// 构造驱动器路径
			drive := fmt.Sprintf("%c:", 'A'+i)
			drives = append(drives, drive)
		}
	}
	
	return drives, nil
}

// GetDriveType 获取驱动器类型
// 获取指定驱动器的类型信息
//
// 参数:
//   drive: 驱动器路径 (如 "C:")
//
// 返回:
//   string: 驱动器类型描述
//   error: 如果获取失败则返回错误
func GetDriveType(drive string) (string, error) {
	// 加载 kernel32.dll 动态链接库
	kernel32 := syscall.MustLoadDLL("kernel32.dll")
	defer kernel32.Release()
	
	// 获取 GetDriveTypeW 函数
	getDriveType := kernel32.MustFindProc("GetDriveTypeW")
	
	// 确保路径以反斜杠结尾
	if !strings.HasSuffix(drive, "\\") {
		drive += "\\"
	}
	
	// 将路径转换为 UTF-16 指针
	pathPtr, err := syscall.UTF16PtrFromString(drive)
	if err != nil {
		return "", fmt.Errorf("路径转换失败: %w", err)
	}
	
	// 调用 Windows API
	ret, _, err := getDriveType.Call(uintptr(unsafe.Pointer(pathPtr)))
	if ret == 0 {
		return "", fmt.Errorf("获取驱动器类型失败: %v", err)
	}
	
	// 根据返回值确定驱动器类型
	switch ret {
	case 0:
		return "未知", nil
	case 1:
		return "根目录不存在", nil
	case 2:
		return "软盘", nil
	case 3:
		return "本地磁盘", nil
	case 4:
		return "网络驱动器", nil
	case 5:
		return "光盘驱动器", nil
	case 6:
		return "RAM 磁盘", nil
	default:
		return "未知类型", nil
	}
} 