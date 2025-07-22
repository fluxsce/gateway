package utils

import (
	"flag"
	"os"
	"path/filepath"
)

var (
	// configDir 全局配置目录变量
	configDir string
	// serviceMode 服务模式标志
	serviceMode bool
	// 命令行参数是否已解析
	flagsParsed bool
)

// parseFlags 解析命令行参数
func parseFlags() {
	if flagsParsed {
		return
	}

	var configFlag string
	flag.StringVar(&configFlag, "config", "", "指定配置文件目录路径")
	flag.BoolVar(&serviceMode, "service", false, "以服务模式运行")
	flag.Parse()

	// 如果通过命令行参数指定了配置目录，则使用该值
	if configFlag != "" {
		configDir = configFlag
	}

	flagsParsed = true
}

// GetConfigDir 获取配置目录路径
// 优先级：命令行参数 > 环境变量 > 默认值
func GetConfigDir() string {
	// 确保命令行参数已解析
	parseFlags()

	if configDir == "" {
		// 优先使用环境变量
		configDir = os.Getenv("GATEWAY_CONFIG_DIR")
		if configDir == "" {
			// 使用默认值
			configDir = "./configs"
		}
	}
	return configDir
}

// IsServiceMode 检查是否为服务模式
func IsServiceMode() bool {
	parseFlags()
	return serviceMode
}

// GetConfigPath 获取配置文件的完整路径
// 参数: filename 配置文件名（如 "database.yaml"）
// 返回: 完整的配置文件路径
func GetConfigPath(filename string) string {
	return filepath.Join(GetConfigDir(), filename)
}

// SetConfigDir 设置配置目录路径（用于程序内部动态设置）
// 参数: dir 配置目录路径
func SetConfigDir(dir string) {
	configDir = dir
}

// ResetFlags 重置flag解析状态（主要用于测试）
func ResetFlags() {
	flagsParsed = false
	configDir = ""
	serviceMode = false
}
