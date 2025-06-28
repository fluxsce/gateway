// Package interfaces 定义SSH工具接口
// 提供SSH远程命令执行相关的接口规范
package interfaces

import (
	"context"
	"io"
)

// SSHTool SSH工具接口
// 继承ConnectableTool接口，添加SSH特有的远程命令执行方法
type SSHTool interface {
	ConnectableTool
	
	// ===== 命令执行操作 =====
	
	// ExecuteCommand 执行单个命令
	// 在远程服务器上执行指定的命令
	ExecuteCommand(ctx context.Context, command string) (*SSHCommandResult, error)
	
	// ExecuteCommandWithInput 执行命令并提供输入
	// 执行命令时可以向标准输入提供数据
	ExecuteCommandWithInput(ctx context.Context, command string, input io.Reader) (*SSHCommandResult, error)
	
	// ExecuteScript 执行脚本
	// 在远程服务器上执行脚本内容
	ExecuteScript(ctx context.Context, script string) (*SSHCommandResult, error)
	
	// ExecuteScriptFile 执行脚本文件
	// 上传并执行本地脚本文件
	ExecuteScriptFile(ctx context.Context, scriptPath string) (*SSHCommandResult, error)
	
	// ===== 批量操作 =====
	
	// ExecuteBatch 批量执行命令
	// 按顺序执行多个命令
	ExecuteBatch(ctx context.Context, commands []string) ([]*SSHCommandResult, error)
	
	// ExecuteParallel 并行执行命令
	// 并行执行多个命令
	ExecuteParallel(ctx context.Context, commands []string) ([]*SSHCommandResult, error)
	
	// ===== 交互式操作 =====
	
	// StartInteractiveSession 启动交互式会话
	// 创建一个交互式SSH会话
	StartInteractiveSession(ctx context.Context) (SSHSession, error)
	
	// CreateTunnel 创建SSH隧道
	// 创建本地端口转发或远程端口转发
	CreateTunnel(ctx context.Context, config *SSHTunnelConfig) (SSHTunnel, error)
	
	// ===== 文件操作（通过SSH） =====
	
	// CopyFileToRemote 复制文件到远程（通过SCP）
	// 使用SCP协议复制文件到远程服务器
	CopyFileToRemote(ctx context.Context, localPath, remotePath string) error
	
	// CopyFileFromRemote 从远程复制文件（通过SCP）
	// 使用SCP协议从远程服务器复制文件
	CopyFileFromRemote(ctx context.Context, remotePath, localPath string) error
	
	// ===== 系统信息操作 =====
	
	// GetSystemInfo 获取系统信息
	// 获取远程服务器的系统信息
	GetSystemInfo(ctx context.Context) (*SSHSystemInfo, error)
	
	// GetProcessList 获取进程列表
	// 获取远程服务器的进程列表
	GetProcessList(ctx context.Context) ([]*SSHProcessInfo, error)
	
	// ===== 高级操作 =====
	
	// TestConnection 测试连接
	// 测试SSH连接是否正常
	TestConnection(ctx context.Context) error
	
	// KeepAlive 保持连接活跃
	// 发送保活信号保持连接
	KeepAlive(ctx context.Context) error
}

// SSHCommandResult SSH命令执行结果
type SSHCommandResult struct {
	// 命令
	Command string `json:"command"`
	
	// 退出码
	ExitCode int `json:"exit_code"`
	
	// 标准输出
	Stdout string `json:"stdout"`
	
	// 标准错误
	Stderr string `json:"stderr"`
	
	// 执行时间（毫秒）
	Duration int64 `json:"duration"`
	
	// 是否成功
	Success bool `json:"success"`
	
	// 错误信息
	Error string `json:"error,omitempty"`
}

// SSHSession SSH交互式会话接口
type SSHSession interface {
	// Write 向会话写入数据
	Write(data []byte) (int, error)
	
	// Read 从会话读取数据
	Read(data []byte) (int, error)
	
	// Close 关闭会话
	Close() error
	
	// SetSize 设置终端大小
	SetSize(width, height int) error
	
	// RequestPty 请求伪终端
	RequestPty(term string, width, height int) error
	
	// Shell 启动Shell
	Shell() error
	
	// Wait 等待会话结束
	Wait() error
}

// SSHTunnel SSH隧道接口
type SSHTunnel interface {
	// Start 启动隧道
	Start() error
	
	// Stop 停止隧道
	Stop() error
	
	// IsActive 检查隧道是否活跃
	IsActive() bool
	
	// GetLocalAddr 获取本地地址
	GetLocalAddr() string
	
	// GetRemoteAddr 获取远程地址
	GetRemoteAddr() string
}

// SSHTunnelConfig SSH隧道配置
type SSHTunnelConfig struct {
	// 隧道类型
	Type SSHTunnelType `json:"type"`
	
	// 本地地址
	LocalAddr string `json:"local_addr"`
	
	// 远程地址
	RemoteAddr string `json:"remote_addr"`
	
	// 目标地址（用于远程转发）
	TargetAddr string `json:"target_addr,omitempty"`
}

// SSHTunnelType SSH隧道类型
type SSHTunnelType int

const (
	// SSHTunnelLocal 本地端口转发
	SSHTunnelLocal SSHTunnelType = iota + 1
	
	// SSHTunnelRemote 远程端口转发
	SSHTunnelRemote
	
	// SSHTunnelDynamic 动态端口转发（SOCKS代理）
	SSHTunnelDynamic
)

// String 返回隧道类型的字符串表示
func (t SSHTunnelType) String() string {
	switch t {
	case SSHTunnelLocal:
		return "local"
	case SSHTunnelRemote:
		return "remote"
	case SSHTunnelDynamic:
		return "dynamic"
	default:
		return "unknown"
	}
}

// SSHSystemInfo SSH系统信息
type SSHSystemInfo struct {
	// 主机名
	Hostname string `json:"hostname"`
	
	// 操作系统
	OS string `json:"os"`
	
	// 系统架构
	Arch string `json:"arch"`
	
	// 内核版本
	Kernel string `json:"kernel"`
	
	// 运行时间
	Uptime string `json:"uptime"`
	
	// CPU信息
	CPU string `json:"cpu"`
	
	// 内存信息
	Memory string `json:"memory"`
	
	// 磁盘信息
	Disk string `json:"disk"`
	
	// 网络接口
	Network []string `json:"network"`
}

// SSHProcessInfo SSH进程信息
type SSHProcessInfo struct {
	// 进程ID
	PID int `json:"pid"`
	
	// 父进程ID
	PPID int `json:"ppid"`
	
	// 进程名
	Name string `json:"name"`
	
	// 命令行
	Command string `json:"command"`
	
	// 用户
	User string `json:"user"`
	
	// CPU使用率
	CPU float64 `json:"cpu"`
	
	// 内存使用率
	Memory float64 `json:"memory"`
	
	// 状态
	State string `json:"state"`
	
	// 启动时间
	StartTime string `json:"start_time"`
} 