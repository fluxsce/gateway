package sftp

import (
	"context"
	"fmt"
	"sync"
	
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	
	"gohub/pkg/plugin/tools/configs"
	"gohub/pkg/plugin/tools/interfaces"
	"gohub/pkg/plugin/tools/types"
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
	progressCallback types.ProgressCallback
	
	// 错误回调函数 - 用于处理传输过程中的错误
	errorCallback types.ErrorCallback
}

// NewSFTPClient 创建新的SFTP客户端实例
// 这是客户端的工厂函数，负责初始化客户端配置和状态
// 参数:
//   config: SFTP客户端配置信息
// 返回:
//   interfaces.SFTPTool: SFTP工具接口实现
//   error: 创建失败时返回错误信息
func NewSFTPClient(config *configs.SFTPConfig) (interfaces.SFTPTool, error) {
	if config == nil {
		return nil, fmt.Errorf("配置不能为空")
	}
	
	// 验证必要的配置参数
	if config.Host == "" {
		return nil, fmt.Errorf("主机地址不能为空")
	}
	if config.Port == 0 {
		config.Port = 22 // 默认SFTP端口
	}
	if config.Username == "" {
		return nil, fmt.Errorf("用户名不能为空")
	}
	
	client := &sftpClient{
		config:    config,
		connected: false,
	}
	
	return client, nil
}

// ===== Tool 接口实现 =====

// GetID 获取工具唯一标识符
// 返回当前工具实例的唯一ID，用于在工具池中标识该工具
func (c *sftpClient) GetID() string {
	return fmt.Sprintf("sftp_%s_%s_%d", c.config.Username, c.config.Host, c.config.Port)
}

// GetType 获取工具类型
// 返回工具的类型标识，用于工具池中的分类管理
func (c *sftpClient) GetType() string {
	return "sftp"
}

// Close 关闭工具连接
// 关闭SFTP连接和SSH连接，释放相关资源
func (c *sftpClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	var errs []error
	
	// 关闭SFTP客户端
	if c.sftpClient != nil {
		if err := c.sftpClient.Close(); err != nil {
			errs = append(errs, fmt.Errorf("关闭SFTP客户端失败: %w", err))
		}
		c.sftpClient = nil
	}
	
	// 关闭SSH客户端
	if c.sshClient != nil {
		if err := c.sshClient.Close(); err != nil {
			errs = append(errs, fmt.Errorf("关闭SSH客户端失败: %w", err))
		}
		c.sshClient = nil
	}
	
	c.connected = false
	
	if len(errs) > 0 {
		return fmt.Errorf("关闭连接时发生错误: %v", errs)
	}
	
	return nil
}

// IsActive 检查工具是否处于活跃状态
// 检查SFTP连接是否仍然有效
func (c *sftpClient) IsActive() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	if !c.connected || c.sftpClient == nil {
		return false
	}
	
	// 通过发送一个简单的请求来检查连接是否仍然有效
	_, err := c.sftpClient.Getwd()
	return err == nil
}

// Connect 建立工具连接
// 建立SSH连接并创建SFTP会话
func (c *sftpClient) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// 如果已经连接，先断开
	if c.connected {
		c.disconnect()
	}
	
	// 建立SSH连接
	var authMethods []ssh.AuthMethod
	
	// 添加密码认证
	if c.config.PasswordAuth != nil && c.config.PasswordAuth.Password != "" {
		authMethods = append(authMethods, ssh.Password(c.config.PasswordAuth.Password))
	}
	
	sshConfig := &ssh.ClientConfig{
		User: c.config.Username,
		Auth: authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 注意：生产环境应使用安全的主机密钥验证
	}
	
	addr := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)
	sshClient, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return fmt.Errorf("SSH连接失败: %w", err)
	}
	
	// 创建SFTP会话
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		sshClient.Close()
		return fmt.Errorf("创建SFTP会话失败: %w", err)
	}
	
	c.sshClient = sshClient
	c.sftpClient = sftpClient
	c.connected = true
	
	return nil
}

// ===== ConnectableTool 接口实现 =====

// IsConnected 检查是否已连接
func (c *sftpClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected && c.sftpClient != nil
}

// Reconnect 重新连接
func (c *sftpClient) Reconnect(ctx context.Context) error {
	return c.Connect(ctx)
}

// Disconnect 断开连接
func (c *sftpClient) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.disconnect()
}

// disconnect 内部断开连接方法（不加锁）
func (c *sftpClient) disconnect() error {
	var errs []error
	
	if c.sftpClient != nil {
		if err := c.sftpClient.Close(); err != nil {
			errs = append(errs, err)
		}
		c.sftpClient = nil
	}
	
	if c.sshClient != nil {
		if err := c.sshClient.Close(); err != nil {
			errs = append(errs, err)
		}
		c.sshClient = nil
	}
	
	c.connected = false
	
	if len(errs) > 0 {
		return fmt.Errorf("断开连接时发生错误: %v", errs)
	}
	
	return nil
}

// ===== SFTPTool 接口实现 =====

// UploadFile 上传单个文件到远程服务器
// 实现在 transfer.go 中
func (c *sftpClient) UploadFile(ctx context.Context, localPath, remotePath string, options *configs.SFTPTransferOptions) (*types.TransferResult, error) {
	// 直接调用内部实现
	return c.uploadFileImpl(ctx, localPath, remotePath, options)
}

// DownloadFile 从远程服务器下载单个文件
// 实现在 transfer.go 中
func (c *sftpClient) DownloadFile(ctx context.Context, remotePath, localPath string, options *configs.SFTPTransferOptions) (*types.TransferResult, error) {
	// 直接调用内部实现
	return c.downloadFileImpl(ctx, remotePath, localPath, options)
}

// UploadDirectory 上传整个目录到远程服务器
// 实现在 operations.go 中
func (c *sftpClient) UploadDirectory(ctx context.Context, localDir, remoteDir string, options *configs.SFTPTransferOptions) (*types.BatchTransferResult, error) {
	// 直接调用内部实现
	return c.uploadDirectoryImpl(ctx, localDir, remoteDir, options)
}

// DownloadDirectory 从远程服务器下载整个目录
// 实现在 operations.go 中
func (c *sftpClient) DownloadDirectory(ctx context.Context, remoteDir, localDir string, options *configs.SFTPTransferOptions) (*types.BatchTransferResult, error) {
	// 直接调用内部实现
	return c.downloadDirectoryImpl(ctx, remoteDir, localDir, options)
}

// ListDirectory 列出远程目录内容
// 实现在 operations.go 中
func (c *sftpClient) ListDirectory(ctx context.Context, remotePath string) ([]*types.FileInfo, error) {
	// 直接调用内部实现
	return c.listDirectoryImpl(ctx, remotePath)
}

// CreateDirectory 在远程服务器创建目录
// 实现在 operations.go 中
func (c *sftpClient) CreateDirectory(ctx context.Context, remotePath string, recursive bool) error {
	// 直接委托给 operations.go 中的实现
	return c.createDirectoryImpl(ctx, remotePath, recursive)
}

// RemoveFile 删除远程文件
// 实现在 operations.go 中
func (c *sftpClient) RemoveFile(ctx context.Context, remotePath string) error {
	// 直接委托给 operations.go 中的实现
	return c.removeFileImpl(ctx, remotePath)
}

// RemoveDirectory 删除远程目录
// 实现在 operations.go 中
func (c *sftpClient) RemoveDirectory(ctx context.Context, remotePath string, recursive bool) error {
	// 直接委托给 operations.go 中的实现
	return c.removeDirectoryImpl(ctx, remotePath, recursive)
}

// GetFileInfo 获取远程文件信息
// 实现在 operations.go 中
func (c *sftpClient) GetFileInfo(ctx context.Context, remotePath string) (*types.FileInfo, error) {
	// 直接调用内部实现
	return c.getFileInfoImpl(ctx, remotePath)
}

// BatchTransfer 批量文件传输
// 实现在 operations.go 中
func (c *sftpClient) BatchTransfer(ctx context.Context, operations []*types.TransferOperation, options *configs.SFTPTransferOptions) (*types.BatchTransferResult, error) {
	// 直接调用内部实现
	return c.batchTransferImpl(ctx, operations, options)
}

// SyncDirectory 目录同步
func (c *sftpClient) SyncDirectory(ctx context.Context, localDir, remoteDir string, syncMode types.SyncMode, options *configs.SFTPSyncOptions) (*types.SyncResult, error) {
	// TODO: 实现目录同步逻辑，这个功能比较复杂，需要单独实现
	return nil, fmt.Errorf("SyncDirectory 方法尚未实现")
}

// GetConfig 获取客户端配置
func (c *sftpClient) GetConfig() *configs.SFTPConfig {
	return c.config
}

// SetProgressCallback 设置进度回调函数
func (c *sftpClient) SetProgressCallback(callback types.ProgressCallback) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.progressCallback = callback
}

// SetErrorCallback 设置错误回调函数
func (c *sftpClient) SetErrorCallback(callback types.ErrorCallback) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.errorCallback = callback
}

// NewClient 创建新的SFTP客户端
// 这是对外暴露的工厂函数，用于创建SFTP客户端实例
// 参数:
//   config: SFTP客户端配置
// 返回:
//   interfaces.SFTPTool: SFTP客户端接口实现
//   error: 创建失败时返回错误信息
func NewClient(config *configs.SFTPConfig) (interfaces.SFTPTool, error) {
	return NewSFTPClient(config)
}

 