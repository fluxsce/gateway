package path

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ReadFileContent 读取文件内容并去除首尾空白
func ReadFileContent(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// WriteFileContent 写入文件内容，自动创建目录
func WriteFileContent(filePath, content string) error {
	dir := filepath.Dir(filePath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建目录失败: %w", err)
		}
	}
	return os.WriteFile(filePath, []byte(content), 0644)
}

// FileExists 检查文件是否存在
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

// DirExists 检查目录是否存在
func DirExists(dirPath string) bool {
	info, err := os.Stat(dirPath)
	return err == nil && info.IsDir()
}

// EnsureDir 确保目录存在，不存在则创建
func EnsureDir(dirPath string) error {
	if !DirExists(dirPath) {
		return os.MkdirAll(dirPath, 0755)
	}
	return nil
}

// GetExecutableDir 获取可执行文件所在目录
func GetExecutableDir() string {
	if execPath, err := os.Executable(); err == nil {
		return filepath.Dir(execPath)
	}
	return "."
}

// GetWorkingDir 获取当前工作目录
func GetWorkingDir() string {
	if wd, err := os.Getwd(); err == nil {
		return wd
	}
	return "."
}

// ResolvePath 解析相对路径为绝对路径
// 如果已经是绝对路径则直接返回
func ResolvePath(basePath, relativePath string) string {
	if filepath.IsAbs(relativePath) {
		return relativePath
	}
	return filepath.Join(basePath, relativePath)
}

// CleanPath 清理路径（去除多余的分隔符等）
func CleanPath(path string) string {
	return filepath.Clean(path)
}

// JoinPath 连接路径
func JoinPath(elem ...string) string {
	return filepath.Join(elem...)
}
