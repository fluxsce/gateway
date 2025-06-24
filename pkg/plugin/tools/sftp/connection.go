package sftp

import (
	"fmt"
	"time"
	
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// keepAlive 保持SSH连接活跃
// 定期发送保活请求以防止连接因空闲而被断开
// 这个方法在goroutine中运行，会持续监控连接状态
func (c *sftpClient) keepAlive() {
	// 创建定时器，根据配置的保活间隔发送保活请求
	ticker := time.NewTicker(c.config.KeepAliveInterval)
	defer ticker.Stop()
	
	for {
		// 等待下一个保活间隔
		<-ticker.C
		
		// 获取当前连接状态（使用读锁避免阻塞其他操作）
		c.mu.RLock()
		connected := c.connected
		sshClient := c.sshClient
		c.mu.RUnlock()
		
		// 如果连接已断开，退出保活循环
		if !connected || sshClient == nil {
			return
		}
		
		// 发送保活请求
		// 使用golang.org标准的保活请求类型
		_, _, err := sshClient.SendRequest("keepalive@golang.org", true, nil)
		if err != nil {
			// 保活请求失败，可能连接已断开
			// 如果配置了自动重连，尝试重新连接
			if c.config.AutoReconnect {
				go c.reconnect()
			}
			return
		}
	}
}

// reconnect 重新连接到SFTP服务器
// 当检测到连接断开时，尝试重新建立连接
// 支持多次重试和指数退避策略
func (c *sftpClient) reconnect() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// 如果当前仍然连接，不需要重连
	if c.connected {
		return
	}
	
	// 清理现有连接
	c.cleanupConnections()
	
	// 尝试重新连接
	for attempt := 1; attempt <= c.config.MaxReconnectAttempts; attempt++ {
		// 创建SSH配置
		sshConfig, err := c.createSSHConfig()
		if err != nil {
			// SSH配置创建失败，等待后重试
			time.Sleep(c.calculateBackoffDelay(attempt))
			continue
		}
		
		// 尝试连接SSH服务器
		addr := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)
		sshClient, err := ssh.Dial("tcp", addr, sshConfig)
		if err != nil {
			// 连接失败，等待后重试
			time.Sleep(c.calculateBackoffDelay(attempt))
			continue
		}
		
		// 尝试创建SFTP客户端
		sftpClient, err := sftp.NewClient(sshClient)
		if err != nil {
			// SFTP客户端创建失败，关闭SSH连接并重试
			sshClient.Close()
			time.Sleep(c.calculateBackoffDelay(attempt))
			continue
		}
		
		// 重连成功，更新客户端状态
		c.sshClient = sshClient
		c.sftpClient = sftpClient
		c.connected = true
		
		// 重新启动保活机制
		if c.config.KeepAliveInterval > 0 {
			go c.keepAlive()
		}
		
		return
	}
	
	// 所有重连尝试都失败了
	// 这里可以触发错误回调或记录日志
}

// cleanupConnections 清理现有连接
// 安全地关闭SFTP和SSH连接，释放相关资源
func (c *sftpClient) cleanupConnections() {
	// 关闭SFTP客户端
	if c.sftpClient != nil {
		c.sftpClient.Close()
		c.sftpClient = nil
	}
	
	// 关闭SSH客户端
	if c.sshClient != nil {
		c.sshClient.Close()
		c.sshClient = nil
	}
	
	// 更新连接状态
	c.connected = false
}

// calculateBackoffDelay 计算退避延迟时间
// 使用指数退避策略，避免频繁重连对服务器造成压力
// 参数:
//   attempt: 当前重试次数
// 返回:
//   time.Duration: 延迟时间
func (c *sftpClient) calculateBackoffDelay(attempt int) time.Duration {
	// 基础延迟时间
	baseDelay := c.config.ReconnectInterval
	
	// 如果没有配置重连间隔，使用默认值
	if baseDelay <= 0 {
		baseDelay = 5 * time.Second
	}
	
	// 计算指数退避延迟
	// 延迟时间 = 基础延迟 * 2^(重试次数-1)
	// 但最大不超过配置的最大重连间隔
	delay := baseDelay * time.Duration(1<<uint(attempt-1))
	
	// 限制最大延迟时间
	maxDelay := c.config.MaxReconnectInterval
	if maxDelay <= 0 {
		maxDelay = 60 * time.Second // 默认最大延迟60秒
	}
	
	if delay > maxDelay {
		delay = maxDelay
	}
	
	return delay
}

// testConnection 测试连接可用性
// 通过执行一个简单的操作来验证连接是否正常
// 返回:
//   error: 连接不可用时返回错误
func (c *sftpClient) testConnection() error {
	if !c.IsConnected() {
		return fmt.Errorf("SFTP连接未建立")
	}
	
	// 尝试获取工作目录，这是一个轻量级的操作
	_, err := c.sftpClient.Getwd()
	if err != nil {
		return fmt.Errorf("连接测试失败: %w", err)
	}
	
	return nil
}

// getConnectionInfo 获取连接信息
// 返回当前连接的详细信息，用于调试和监控
// 返回:
//   map[string]interface{}: 连接信息映射
func (c *sftpClient) getConnectionInfo() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	info := make(map[string]interface{})
	info["connected"] = c.connected
	info["host"] = c.config.Host
	info["port"] = c.config.Port
	info["username"] = c.config.Username
	
	if c.sshClient != nil {
		// 获取SSH连接的详细信息
		info["remote_addr"] = c.sshClient.RemoteAddr().String()
		info["client_version"] = c.sshClient.ClientVersion()
		info["server_version"] = c.sshClient.ServerVersion()
	}
	
	return info
} 