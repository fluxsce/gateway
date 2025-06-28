package common

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	
	"gohub/pkg/plugin/tools/types"
)

// FileExists 检查文件是否存在
// 参数:
//   path: 文件路径
// 返回:
//   bool: 文件是否存在
//   error: 检查过程中的错误
func FileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// IsDirectory 检查路径是否为目录
// 参数:
//   path: 文件路径
// 返回:
//   bool: 是否为目录
//   error: 检查过程中的错误
func IsDirectory(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

// EnsureDirectoryExists 确保目录存在，如果不存在则创建
// 参数:
//   path: 目录路径
//   recursive: 是否递归创建父目录
// 返回:
//   error: 创建过程中的错误
func EnsureDirectoryExists(path string, recursive bool) error {
	exists, err := FileExists(path)
	if err != nil {
		return err
	}
	
	if exists {
		isDir, err := IsDirectory(path)
		if err != nil {
			return err
		}
		if !isDir {
			return fmt.Errorf("path exists but is not a directory: %s", path)
		}
		return nil
	}
	
	if recursive {
		return os.MkdirAll(path, 0755)
	}
	return os.Mkdir(path, 0755)
}

// GetFileInfo 获取文件信息
// 参数:
//   path: 文件路径
// 返回:
//   *types.FileInfo: 文件信息
//   error: 获取过程中的错误
func GetFileInfo(path string) (*types.FileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	
	fileInfo := &types.FileInfo{
		Name:    info.Name(),
		Path:    path,
		Size:    info.Size(),
		IsDir:   info.IsDir(),
		ModTime: info.ModTime(),
		Mode:    info.Mode(),
	}
	
	return fileInfo, nil
}

// CalculateFileChecksum 计算文件校验和
// 参数:
//   path: 文件路径
//   algorithm: 校验和算法 (md5, sha1, sha256)
// 返回:
//   string: 校验和字符串
//   error: 计算过程中的错误
func CalculateFileChecksum(path string, algorithm string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	
	var hasher hash.Hash
	
	switch strings.ToLower(algorithm) {
	case "md5":
		hasher = md5.New()
	case "sha1":
		hasher = sha1.New()
	case "sha256":
		hasher = sha256.New()
	default:
		return "", fmt.Errorf("unsupported checksum algorithm: %s", algorithm)
	}
	
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}
	
	result := hasher.Sum(nil)
	
	return hex.EncodeToString(result), nil
}

// NormalizePath 标准化路径，处理路径分隔符和相对路径
// 参数:
//   path: 原始路径
//   isUnix: 是否使用Unix风格路径（/）
// 返回:
//   string: 标准化后的路径
func NormalizePath(path string, isUnix bool) string {
	path = filepath.Clean(path)
	
	if isUnix {
		return strings.ReplaceAll(path, "\\", "/")
	}
	return path
}

// JoinPaths 连接路径，处理不同操作系统的路径分隔符
// 参数:
//   isUnix: 是否使用Unix风格路径（/）
//   elem: 路径元素
// 返回:
//   string: 连接后的路径
func JoinPaths(isUnix bool, elem ...string) string {
	result := filepath.Join(elem...)
	
	if isUnix {
		return strings.ReplaceAll(result, "\\", "/")
	}
	return result
}

// MatchPattern 检查路径是否匹配给定的模式
// 参数:
//   path: 文件路径
//   pattern: 匹配模式（支持glob和正则表达式）
//   isRegex: 是否为正则表达式模式
// 返回:
//   bool: 是否匹配
//   error: 匹配过程中的错误
func MatchPattern(path string, pattern string, isRegex bool) (bool, error) {
	if isRegex {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return false, err
		}
		return re.MatchString(path), nil
	}
	
	return filepath.Match(pattern, path)
}

// FormatBytes 格式化字节数为人类可读的形式
// 参数:
//   bytes: 字节数
// 返回:
//   string: 格式化后的字符串
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FormatDuration 格式化持续时间为人类可读的形式
// 参数:
//   d: 持续时间
// 返回:
//   string: 格式化后的字符串
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%d ms", d.Milliseconds())
	}
	
	if d < time.Minute {
		return fmt.Sprintf("%.1f sec", d.Seconds())
	}
	
	if d < time.Hour {
		minutes := d / time.Minute
		seconds := (d % time.Minute) / time.Second
		return fmt.Sprintf("%d min %d sec", minutes, seconds)
	}
	
	hours := d / time.Hour
	minutes := (d % time.Hour) / time.Minute
	return fmt.Sprintf("%d hr %d min", hours, minutes)
}

// GenerateUniqueID 生成唯一标识符
// 参数:
//   prefix: 标识符前缀
// 返回:
//   string: 唯一标识符
func GenerateUniqueID(prefix string) string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%s_%d", prefix, timestamp)
}

// FilterFiles 根据模式过滤文件列表
// 参数:
//   files: 文件列表
//   includePatterns: 包含模式列表
//   excludePatterns: 排除模式列表
// 返回:
//   []*types.FileInfo: 过滤后的文件列表
//   error: 过滤过程中的错误
func FilterFiles(files []*types.FileInfo, includePatterns, excludePatterns []string) ([]*types.FileInfo, error) {
	if len(includePatterns) == 0 && len(excludePatterns) == 0 {
		return files, nil
	}
	
	var result []*types.FileInfo
	
	for _, file := range files {
		// 如果有排除模式，检查文件是否被排除
		excluded := false
		for _, pattern := range excludePatterns {
			match, err := filepath.Match(pattern, file.Name)
			if err != nil {
				return nil, err
			}
			if match {
				excluded = true
				break
			}
		}
		
		if excluded {
			continue
		}
		
		// 如果有包含模式，检查文件是否被包含
		included := len(includePatterns) == 0 // 如果没有包含模式，默认包含所有文件
		for _, pattern := range includePatterns {
			match, err := filepath.Match(pattern, file.Name)
			if err != nil {
				return nil, err
			}
			if match {
				included = true
				break
			}
		}
		
		if included {
			result = append(result, file)
		}
	}
	
	return result, nil
}

// CopyFile 复制文件
// 参数:
//   src: 源文件路径
//   dst: 目标文件路径
//   bufferSize: 缓冲区大小
// 返回:
//   int64: 复制的字节数
//   error: 复制过程中的错误
func CopyFile(src, dst string, bufferSize int) (int64, error) {
	sourceFile, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer sourceFile.Close()
	
	// 获取源文件信息
	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		return 0, err
	}
	
	// 创建目标文件
	destFile, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destFile.Close()
	
	// 设置目标文件权限与源文件相同
	if err := os.Chmod(dst, sourceInfo.Mode()); err != nil {
		return 0, err
	}
	
	// 使用指定缓冲区大小复制文件内容
	buffer := make([]byte, bufferSize)
	return io.CopyBuffer(destFile, sourceFile, buffer)
} 