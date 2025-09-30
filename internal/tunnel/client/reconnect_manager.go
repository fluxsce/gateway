// Package client 提供重连管理器的完整实现
// 重连管理器负责在网络连接中断或异常时自动重新建立连接
package client

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"gateway/pkg/logger"
)

// reconnectManager 重连管理器实现
type reconnectManager struct {
	client       *tunnelClient
	stats        *ReconnectStats
	running      bool
	reconnecting bool
	mutex        sync.RWMutex

	// 重连配置
	maxRetries   int
	baseInterval time.Duration
	maxInterval  time.Duration

	// 控制状态
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewReconnectManager 创建重连管理器实例
func NewReconnectManager(client *tunnelClient) ReconnectManager {
	return &reconnectManager{
		client:       client,
		maxRetries:   client.config.MaxRetries,
		baseInterval: time.Duration(client.config.RetryInterval) * time.Second,
		maxInterval:  300 * time.Second, // 最大5分钟
		stats: &ReconnectStats{
			TotalAttempts:      0,
			SuccessfulAttempts: 0,
			FailedAttempts:     0,
			MaxRetryInterval:   300,
			CurrentRetryCount:  0,
			IsReconnecting:     false,
		},
	}
}

// Start 启动重连管理
func (rm *reconnectManager) Start(ctx context.Context) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	if rm.running {
		return nil
	}

	rm.ctx, rm.cancel = context.WithCancel(ctx)
	rm.running = true

	logger.Info("Reconnect manager started", map[string]interface{}{
		"maxRetries":   rm.maxRetries,
		"baseInterval": rm.baseInterval.String(),
		"maxInterval":  rm.maxInterval.String(),
	})

	return nil
}

// Stop 停止重连管理
func (rm *reconnectManager) Stop(ctx context.Context) error {
	rm.mutex.Lock()
	if !rm.running {
		rm.mutex.Unlock()
		return nil
	}
	rm.running = false
	rm.mutex.Unlock()

	if rm.cancel != nil {
		rm.cancel()
	}

	// 等待重连协程退出
	done := make(chan struct{})
	go func() {
		rm.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("Reconnect manager stopped", nil)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// TriggerReconnect 触发重连
func (rm *reconnectManager) TriggerReconnect(ctx context.Context, reason string) error {
	rm.mutex.Lock()
	if rm.reconnecting {
		rm.mutex.Unlock()
		return fmt.Errorf("already reconnecting")
	}
	rm.reconnecting = true
	rm.stats.IsReconnecting = true
	rm.mutex.Unlock()

	logger.Info("Reconnect triggered", map[string]interface{}{
		"reason": reason,
	})

	// 启动重连协程
	rm.wg.Add(1)
	go rm.reconnectLoop(reason)

	return nil
}

// IsReconnecting 检查是否正在重连
func (rm *reconnectManager) IsReconnecting() bool {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	return rm.reconnecting
}

// GetReconnectStats 获取重连统计
func (rm *reconnectManager) GetReconnectStats() *ReconnectStats {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	// 返回统计副本
	stats := *rm.stats
	return &stats
}

// reconnectLoop 重连循环
func (rm *reconnectManager) reconnectLoop(reason string) {
	defer rm.wg.Done()
	defer func() {
		rm.mutex.Lock()
		rm.reconnecting = false
		rm.stats.IsReconnecting = false
		rm.stats.CurrentRetryCount = 0
		rm.mutex.Unlock()
	}()

	for attempt := 1; attempt <= rm.maxRetries; attempt++ {
		select {
		case <-rm.ctx.Done():
			return
		default:
		}

		rm.updateStats(attempt, false)

		logger.Info("Attempting to reconnect", map[string]interface{}{
			"attempt": attempt,
			"reason":  reason,
		})

		// 尝试重连
		if err := rm.attemptReconnect(); err != nil {
			logger.Error("Reconnect attempt failed", map[string]interface{}{
				"attempt": attempt,
				"error":   err.Error(),
			})

			rm.mutex.Lock()
			rm.stats.FailedAttempts++
			rm.mutex.Unlock()

			// 如果不是最后一次尝试，等待重连间隔
			if attempt < rm.maxRetries {
				interval := rm.calculateBackoffInterval(attempt)
				logger.Info("Waiting before next reconnect attempt", map[string]interface{}{
					"interval": interval.String(),
					"attempt":  attempt,
				})

				select {
				case <-rm.ctx.Done():
					return
				case <-time.After(interval):
				}
			}
		} else {
			// 重连成功
			rm.mutex.Lock()
			rm.stats.SuccessfulAttempts++
			rm.stats.LastSuccessTime = time.Now()
			rm.mutex.Unlock()

			logger.Info("Reconnect successful", map[string]interface{}{
				"attempt": attempt,
			})

			return
		}
	}

	// 所有重连尝试都失败了
	logger.Error("All reconnect attempts failed", map[string]interface{}{
		"maxRetries": rm.maxRetries,
		"reason":     reason,
	})

	rm.client.updateStatus(StatusError, false)
	rm.client.addError(fmt.Sprintf("Reconnect failed after %d attempts: %s", rm.maxRetries, reason))
}

// attemptReconnect 尝试重连
func (rm *reconnectManager) attemptReconnect() error {
	// 断开现有连接
	if rm.client.controlConn.IsConnected() {
		rm.client.controlConn.Disconnect(context.Background())
	}

	// 等待一小段时间
	time.Sleep(1 * time.Second)

	// 重新建立连接
	if err := rm.client.controlConn.Connect(rm.ctx, rm.client.config.ServerAddress, rm.client.config.ServerPort); err != nil {
		return fmt.Errorf("failed to reconnect: %w", err)
	}

	// 重新启动心跳
	heartbeatInterval := time.Duration(rm.client.config.HeartbeatInterval) * time.Second
	if err := rm.client.heartbeatManager.Start(rm.ctx, heartbeatInterval); err != nil {
		return fmt.Errorf("failed to restart heartbeat: %w", err)
	}

	// 更新客户端状态
	rm.client.updateStatus(StatusConnected, true)
	rm.client.updateConnectTime()

	// 更新重连计数
	rm.client.statusMutex.Lock()
	rm.client.status.ReconnectCount++
	rm.client.statusMutex.Unlock()

	return nil
}

// calculateBackoffInterval 计算指数退避间隔
func (rm *reconnectManager) calculateBackoffInterval(attempt int) time.Duration {
	// 指数退避算法: baseInterval * 2^(attempt-1)
	backoff := float64(rm.baseInterval) * math.Pow(2, float64(attempt-1))
	interval := time.Duration(backoff)

	// 限制最大间隔
	if interval > rm.maxInterval {
		interval = rm.maxInterval
	}

	return interval
}

// updateStats 更新统计信息
func (rm *reconnectManager) updateStats(attempt int, success bool) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	rm.stats.TotalAttempts++
	rm.stats.LastAttemptTime = time.Now()
	rm.stats.CurrentRetryCount = attempt

	if success {
		rm.stats.SuccessfulAttempts++
		rm.stats.LastSuccessTime = time.Now()
	}
}

// Close 关闭重连管理器
func (rm *reconnectManager) Close() error {
	if rm.cancel != nil {
		rm.cancel()
	}
	rm.wg.Wait()

	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	rm.running = false
	rm.reconnecting = false
	rm.stats.IsReconnecting = false

	return nil
}
