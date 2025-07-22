package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"gateway/pkg/metric/types"
	"gateway/pkg/metric/utils"
)

// TestFormatDiskUsage 测试格式化磁盘使用信息函数
func TestFormatDiskUsage(t *testing.T) {
	tests := []struct {
		name     string
		info     *types.DiskUsageInfo
		expected string
	}{
		{
			name: "正常磁盘信息",
			info: &types.DiskUsageInfo{
				Total:        100 * 1024 * 1024 * 1024, // 100GB
				Used:         50 * 1024 * 1024 * 1024,  // 50GB
				Free:         50 * 1024 * 1024 * 1024,  // 50GB
				UsagePercent: 50.0,
			},
			expected: "总空间: 100.0 GB, 已使用: 50.0 GB (50.0%), 可用: 50.0 GB",
		},
		{
			name:     "空指针",
			info:     nil,
			expected: "无效的磁盘信息",
		},
		{
			name: "小容量磁盘",
			info: &types.DiskUsageInfo{
				Total:        1024, // 1KB
				Used:         512,  // 512B
				Free:         512,  // 512B
				UsagePercent: 50.0,
			},
			expected: "总空间: 1.0 KB, 已使用: 512 B (50.0%), 可用: 512 B",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.FormatDiskUsage(tt.info)
			if result != tt.expected {
				t.Errorf("FormatDiskUsage() = %s, expected %s", result, tt.expected)
			}
		})
	}
}

// TestFormatDiskUsageDetailed 测试格式化详细磁盘使用信息函数
func TestFormatDiskUsageDetailed(t *testing.T) {
	tests := []struct {
		name     string
		info     *types.DiskUsageInfo
		expected string
	}{
		{
			name: "带 inode 信息",
			info: &types.DiskUsageInfo{
				Total:              100 * 1024 * 1024 * 1024, // 100GB
				Used:               50 * 1024 * 1024 * 1024,  // 50GB
				Free:               50 * 1024 * 1024 * 1024,  // 50GB
				UsagePercent:       50.0,
				InodesTotal:        1000,
				InodesUsed:         300,
				InodesFree:         700,
				InodesUsagePercent: 30.0,
			},
			expected: "总空间: 100.0 GB, 已使用: 50.0 GB (50.0%), 可用: 50.0 GB\ninode: 总数: 1000, 已使用: 300 (30.0%), 可用: 700",
		},
		{
			name: "无 inode 信息",
			info: &types.DiskUsageInfo{
				Total:        100 * 1024 * 1024 * 1024, // 100GB
				Used:         50 * 1024 * 1024 * 1024,  // 50GB
				Free:         50 * 1024 * 1024 * 1024,  // 50GB
				UsagePercent: 50.0,
				InodesTotal:  0,
			},
			expected: "总空间: 100.0 GB, 已使用: 50.0 GB (50.0%), 可用: 50.0 GB",
		},
		{
			name:     "空指针",
			info:     nil,
			expected: "无效的磁盘信息",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.FormatDiskUsageDetailed(tt.info)
			if result != tt.expected {
				t.Errorf("FormatDiskUsageDetailed() = %s, expected %s", result, tt.expected)
			}
		})
	}
}

// TestGetDiskUsage 测试获取磁盘使用信息函数
func TestGetDiskUsage(t *testing.T) {
	// 测试当前目录
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("获取当前目录失败: %v", err)
	}

	info, err := utils.GetDiskUsage(currentDir)
	if err != nil {
		t.Fatalf("GetDiskUsage() error = %v", err)
	}
	result := utils.FormatDiskUsageDetailed(info)
	t.Logf("FormatDiskUsage() result = %s", result)
	// 验证返回的信息是否合理
	if info == nil {
		t.Error("GetDiskUsage() returned nil info")
	}

	if info.Total == 0 {
		t.Error("GetDiskUsage() returned zero total space")
	}

	if info.Used+info.Free > info.Total {
		t.Error("GetDiskUsage() returned inconsistent used+free > total")
	}

	if info.UsagePercent < 0 || info.UsagePercent > 100 {
		t.Errorf("GetDiskUsage() returned invalid usage percent: %f", info.UsagePercent)
	}

	// 测试 Windows 根目录
	if runtime.GOOS == "windows" {
		rootInfo, err := utils.GetDiskUsage("C:")
		if err != nil {
			t.Logf("GetDiskUsage('C:') error (this might be expected): %v", err)
		} else {
			if rootInfo.Total == 0 {
				t.Error("GetDiskUsage('C:') returned zero total space")
			}
		}
	}
}

// TestGetDiskUsageMultiple 测试批量获取磁盘使用信息函数
func TestGetDiskUsageMultiple(t *testing.T) {
	// 创建测试目录
	tempDir := t.TempDir()
	subDir := filepath.Join(tempDir, "subdir")
	err := os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatalf("创建测试目录失败: %v", err)
	}

	paths := []string{tempDir, subDir}
	result, err := utils.GetDiskUsageMultiple(paths)
	if err != nil {
		t.Fatalf("GetDiskUsageMultiple() error = %v", err)
	}

	if len(result) != len(paths) {
		t.Errorf("GetDiskUsageMultiple() returned %d results, expected %d", len(result), len(paths))
	}

	for _, path := range paths {
		info, exists := result[path]
		if !exists {
			t.Errorf("GetDiskUsageMultiple() missing result for path %s", path)
		}
		if info.Total == 0 {
			t.Errorf("GetDiskUsageMultiple() returned zero total space for path %s", path)
		}
	}
}

// TestGetDiskUsageThreshold 测试磁盘使用率阈值检查函数
func TestGetDiskUsageThreshold(t *testing.T) {
	tempDir := t.TempDir()

	// 测试低阈值（应该超过）
	exceeded, info, err := utils.GetDiskUsageThreshold(tempDir, 0.1)
	if err != nil {
		t.Fatalf("GetDiskUsageThreshold() error = %v", err)
	}

	if !exceeded {
		t.Error("GetDiskUsageThreshold() should exceed very low threshold")
	}

	if info == nil {
		t.Error("GetDiskUsageThreshold() returned nil info")
	}

	// 测试高阈值（应该不超过）
	exceeded, info, err = utils.GetDiskUsageThreshold(tempDir, 99.9)
	if err != nil {
		t.Fatalf("GetDiskUsageThreshold() error = %v", err)
	}

	if exceeded {
		t.Error("GetDiskUsageThreshold() should not exceed very high threshold")
	}
}

// TestGetDiskUsageWithInvalidPath 测试使用无效路径获取磁盘使用信息
func TestGetDiskUsageWithInvalidPath(t *testing.T) {
	// 测试不存在的路径
	invalidPath := "/this/path/does/not/exist/12345"
	if runtime.GOOS == "windows" {
		invalidPath = "Z:\\this\\path\\does\\not\\exist\\12345"
	}

	_, err := utils.GetDiskUsage(invalidPath)
	if err == nil {
		t.Error("GetDiskUsage() should return error for invalid path")
	}
}

// TestGetDiskUsageMultipleWithInvalidPath 测试批量获取包含无效路径的磁盘使用信息
func TestGetDiskUsageMultipleWithInvalidPath(t *testing.T) {
	tempDir := t.TempDir()
	invalidPath := "/this/path/does/not/exist/12345"
	if runtime.GOOS == "windows" {
		invalidPath = "Z:\\this\\path\\does\\not\\exist\\12345"
	}

	paths := []string{tempDir, invalidPath}
	_, err := utils.GetDiskUsageMultiple(paths)
	if err == nil {
		t.Error("GetDiskUsageMultiple() should return error when one path is invalid")
	}
}

// TestDiskUsageInfoValidation 测试磁盘使用信息的一致性
func TestDiskUsageInfoValidation(t *testing.T) {
	tempDir := t.TempDir()

	info, err := utils.GetDiskUsage(tempDir)
	if err != nil {
		t.Fatalf("GetDiskUsage() error = %v", err)
	}

	// 验证百分比计算的正确性
	expectedPercent := float64(info.Used) / float64(info.Total) * 100
	if expectedPercent != info.UsagePercent {
		t.Errorf("UsagePercent calculation incorrect: got %f, expected %f", info.UsagePercent, expectedPercent)
	}

	// 验证 inode 百分比计算的正确性（仅在有 inode 信息时）
	if info.InodesTotal > 0 {
		expectedInodesPercent := float64(info.InodesUsed) / float64(info.InodesTotal) * 100
		if expectedInodesPercent != info.InodesUsagePercent {
			t.Errorf("InodesUsagePercent calculation incorrect: got %f, expected %f", info.InodesUsagePercent, expectedInodesPercent)
		}
	}
}

// BenchmarkGetDiskUsage 性能测试
func BenchmarkGetDiskUsage(b *testing.B) {
	tempDir := b.TempDir()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := utils.GetDiskUsage(tempDir)
		if err != nil {
			b.Fatalf("GetDiskUsage() error = %v", err)
		}
	}
}

// BenchmarkFormatDiskUsage 性能测试
func BenchmarkFormatDiskUsage(b *testing.B) {
	info := &types.DiskUsageInfo{
		Total:        100 * 1024 * 1024 * 1024, // 100GB
		Used:         50 * 1024 * 1024 * 1024,  // 50GB
		Free:         50 * 1024 * 1024 * 1024,  // 50GB
		UsagePercent: 50.0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		utils.FormatDiskUsage(info)
	}
}
