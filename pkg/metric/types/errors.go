package types

import "errors"

// 错误定义
var (
	// ErrCollectorNotFound 采集器未找到
	ErrCollectorNotFound = errors.New("collector not found")
	// ErrCollectorAlreadyExists 采集器已存在
	ErrCollectorAlreadyExists = errors.New("collector already exists")
	// ErrCollectorDisabled 采集器已禁用
	ErrCollectorDisabled = errors.New("collector is disabled")
	// ErrManagerNotRunning 管理器未运行
	ErrManagerNotRunning = errors.New("manager is not running")
	// ErrManagerAlreadyRunning 管理器已在运行
	ErrManagerAlreadyRunning = errors.New("manager is already running")
	// ErrInvalidInterval 无效的间隔时间
	ErrInvalidInterval = errors.New("invalid interval")
	// ErrCollectFailed 采集失败
	ErrCollectFailed = errors.New("collect failed")
	// ErrSystemNotSupported 系统不支持
	ErrSystemNotSupported = errors.New("system not supported")
	// ErrPermissionDenied 权限不足
	ErrPermissionDenied = errors.New("permission denied")
	// ErrDataNotAvailable 数据不可用
	ErrDataNotAvailable = errors.New("data not available")
)

// 常量定义
const (
	// 采集器名称
	CollectorNameCPU     = "cpu"
	CollectorNameMemory  = "memory"
	CollectorNameDisk    = "disk"
	CollectorNameNetwork = "network"
	CollectorNameSystem  = "system"
	CollectorNameProcess = "process"

	// 默认采集间隔 (秒)
	DefaultCollectInterval = 30

	// 数据格式化常量
	ByteToKB = 1024
	ByteToMB = 1024 * 1024
	ByteToGB = 1024 * 1024 * 1024
	ByteToTB = 1024 * 1024 * 1024 * 1024

	// 系统类型
	SystemTypeLinux   = "linux"
	SystemTypeWindows = "windows"
	SystemTypeDarwin  = "darwin"
	SystemTypeFreeBSD = "freebsd"

	// 进程状态
	ProcessStatusRunning = "running"
	ProcessStatusSleeping = "sleeping"
	ProcessStatusStopped = "stopped"
	ProcessStatusZombie = "zombie"
	ProcessStatusUnknown = "unknown"

	// 网络接口状态
	NetworkStatusUp    = "up"
	NetworkStatusDown  = "down"
	NetworkStatusUnknown = "unknown"

	// 网络接口类型
	NetworkTypeEthernet = "ethernet"
	NetworkTypeWifi     = "wifi"
	NetworkTypeLoopback = "loopback"
	NetworkTypeUnknown  = "unknown"
) 