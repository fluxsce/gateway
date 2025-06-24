package sftp

import (
	"context"
	"fmt"
	"sync"
	
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	
	"gohub/pkg/plugin/tools/common"
	"gohub/pkg/plugin/tools/configs"
)

// sftpClient SFTP客户端实现
// 这是SFTP客户端的核心实现，负责管理SSH连接、SFTP会话以及各种文件操作
type sftpClient struct {
	// 配置信息 - 包含连接、认证、传输等所有配置参数
	config *configs.SFTPConfig
	
	// SSH客户端 - 底层的SSH连接客户端
	sshClient *ssh.Client
	
	// SFTP客户端 - 基于SSH连接的SFTP会话客户端
	sftpClient *sftp.Client
	
	// 连接状态 - 标识当前是否已连接到服务器
	connected bool
	
	// 互斥锁 - 用于保护连接状态和客户端实例的并发安全
	mu sync.RWMutex
	
	// 进度回调函数 - 用于监控文件传输进度
	progressCallback common.ProgressCallback
	
	// 错误回调函数 - 用于处理传输过程中的错误
	errorCallback common.ErrorCallback
}

// NewSFTPClient 创建新的SFTP客户端实例
// 这是客户端的工厂函数，用于创建并初始化SFTP客户端
// 参数:
//   config: SFTP配置信息，如果为nil则使用默认配置
// 返回:
//   Client: SFTP客户端接口实例
//   error: 创建过程中的错误
func NewSFTPClient(config *configs.SFTPConfig) (Client, error) {
	// 如果未提供配置，使用默认配置
	if config == nil {
		config = configs.DefaultSFTPConfig()
	}
	
	// 验证必要的配置项
	if err := validateConfig(config); err != nil {
		return nil, common.NewInvalidArgumentError(fmt.Sprintf("配置验证失败: %v", err))
	}
	
	return &sftpClient{
		config:    config,
		connected: false,
	}, nil
}

// validateConfig 验证SFTP配置的有效性
// 检查必要的配置项是否完整且合理
func validateConfig(config *configs.SFTPConfig) error {
	if config.Host == "" {
		return fmt.Errorf("主机地址不能为空")
	}
	
	if config.Port <= 0 || config.Port > 65535 {
		return fmt.Errorf("端口号必须在1-65535范围内")
	}
	
	if config.Username == "" {
		return fmt.Errorf("用户名不能为空")
	}
	
	// 检查是否至少配置了一种认证方式
	hasAuth := false
	if config.PasswordAuth != nil && (config.PasswordAuth.Password != "" || config.PasswordAuth.AllowEmptyPassword) {
		hasAuth = true
	}
	if config.PublicKeyAuth != nil && (len(config.PublicKeyAuth.PrivateKeyData) > 0 || config.PublicKeyAuth.PrivateKeyPath != "") {
		hasAuth = true
	}
	
	if !hasAuth {
		return fmt.Errorf("必须配置至少一种认证方式（密码或公钥）")
	}
	
	return nil
}

// Connect 连接到SFTP服务器
// 建立SSH连接并创建SFTP会话，支持超时控制和重连机制
// 参数:
//   ctx: 上下文，用于超时控制和取消操作
// 返回:
//   error: 连接失败时返回错误信息
func (c *sftpClient) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// 如果已经连接，直接返回成功
	if c.connected {
		return nil
	}
	
	// 创建SSH配置
	sshConfig, err := c.createSSHConfig()
	if err != nil {
		return common.NewAuthenticationError("创建SSH配置失败", err)
	}
	
	// 设置连接超时上下文
	connectCtx, cancel := context.WithTimeout(ctx, c.config.ConnectTimeout)
	defer cancel()
	
	// 执行连接操作
	sshClient, err := c.dialWithTimeout(connectCtx, sshConfig)
	if err != nil {
		return err
	}
	
	// 创建SFTP客户端
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		sshClient.Close()
		return common.NewConnectionError("创建SFTP客户端失败", err)
	}
	
	// 更新客户端状态
	c.sshClient = sshClient
	c.sftpClient = sftpClient
	c.connected = true
	
	// 启动保活机制（如果配置了保活间隔）
	if c.config.KeepAliveInterval > 0 {
		go c.keepAlive()
	}
	
	return nil
}

// dialWithTimeout 带超时的SSH连接
// 使用goroutine实现带超时的SSH连接，避免阻塞
func (c *sftpClient) dialWithTimeout(ctx context.Context, sshConfig *ssh.ClientConfig) (*ssh.Client, error) {
	addr := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)
	
	// 创建通道用于goroutine通信
	resultChan := make(chan *ssh.Client, 1)
	errorChan := make(chan error, 1)
	
	// 在goroutine中执行连接
	go func() {
		client, err := ssh.Dial("tcp", addr, sshConfig)
		if err != nil {
			errorChan <- err
			return
		}
		resultChan <- client
	}()
	
	// 等待连接结果或超时
	select {
	case client := <-resultChan:
		return client, nil
	case err := <-errorChan:
		return nil, common.NewConnectionError(fmt.Sprintf("连接到SFTP服务器失败: %s", addr), err)
	case <-ctx.Done():
		return nil, common.NewTimeoutError(fmt.Sprintf("连接超时: %s", addr), ctx.Err())
	}
}

// Close 关闭SFTP连接
// 释放所有相关资源，包括SFTP会话和SSH连接
// 返回:
//   error: 关闭失败时返回错误信息
func (c *sftpClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// 如果未连接，直接返回成功
	if !c.connected {
		return nil
	}
	
	var closeErrors []error
	
	// 关闭SFTP客户端
	if c.sftpClient != nil {
		if err := c.sftpClient.Close(); err != nil {
			closeErrors = append(closeErrors, fmt.Errorf("关闭SFTP客户端失败: %w", err))
		}
		c.sftpClient = nil
	}
	
	// 关闭SSH客户端
	if c.sshClient != nil {
		if err := c.sshClient.Close(); err != nil {
			closeErrors = append(closeErrors, fmt.Errorf("关闭SSH客户端失败: %w", err))
		}
		c.sshClient = nil
	}
	
	// 更新连接状态
	c.connected = false
	
	// 如果有关闭错误，返回第一个错误
	if len(closeErrors) > 0 {
		return closeErrors[0]
	}
	
	return nil
}

// IsConnected 检查连接状态
// 返回当前SFTP连接是否处于活跃状态
// 返回:
//   bool: true表示已连接，false表示未连接
func (c *sftpClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// GetConfig 获取客户端配置
// 返回当前客户端使用的配置信息
// 返回:
//   *configs.SFTPConfig: 配置信息
func (c *sftpClient) GetConfig() *configs.SFTPConfig {
	return c.config
}

// SetProgressCallback 设置进度回调函数
// 设置用于监控传输进度的回调函数
// 参数:
//   callback: 进度回调函数
func (c *sftpClient) SetProgressCallback(callback common.ProgressCallback) {
	c.progressCallback = callback
}

// SetErrorCallback 设置错误回调函数
// 设置用于处理传输错误的回调函数
// 参数:
//   callback: 错误回调函数
func (c *sftpClient) SetErrorCallback(callback common.ErrorCallback) {
	c.errorCallback = callback
} 