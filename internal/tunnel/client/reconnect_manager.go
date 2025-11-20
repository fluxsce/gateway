// Package client 提供重连管理器的完整实现
// 重连管理器负责在网络连接中断或异常时自动重新建立连接
//
// # 重连管理架构
//
// ## 概述
//
// ReconnectManager 实现了智能重连机制，确保在网络中断、连接错误等情况下
// 能够自动恢复连接并重新注册服务，无需人工干预。
//
// ## 核心功能
//
// ### 1. 自动重连
//   - 使用指数退避算法控制重试间隔
//   - 可配置的最大重试次数
//   - 防止重复触发重连
//   - 异步断开连接避免死锁
//
// ### 2. 服务恢复
//   - 重连成功后自动重新注册所有服务
//   - 从数据库加载服务配置
//   - 跳过非活跃状态的服务
//   - 部分失败不影响连接成功
//
// ### 3. 死锁预防
//   - 断开连接使用带超时的上下文
//   - 异步执行断开操作
//   - 超时后强制继续重连流程
//   - 单个服务注册超时控制（30秒）
//
// ## 重连流程
//
//  1. 检测连接错误触发重连
//     ↓
//  2. 检查是否已在重连中（防止重复）
//     ↓
//  3. 异步断开现有连接（带超时保护）
//     ↓
//  4. 等待资源清理完成
//     ↓
//  5. 重新建立 TCP 连接
//     ↓
//  6. 发送认证消息
//     ↓
//  7. 重新启动心跳管理器
//     ↓
//  8. 从数据库加载服务列表
//     ↓
//  9. 逐个重新注册服务（带超时）
//     ↓
//  10. 更新数据库连接状态
//     ↓
//  11. 记录重连统计信息
//
// ## 错误处理策略
//
// ### 断开连接超时
//   - 超时时间：5秒
//   - 超时后：记录警告，强制继续重连
//   - 不阻塞重连流程
//
// ### 服务注册超时
//   - 单个服务超时：30秒
//   - 超时后：跳过该服务，继续注册下一个
//   - 不影响整体重连成功
//
// ### 部分服务注册失败
//   - 策略：记录错误，继续注册其他服务
//   - 只有全部失败才返回错误
//   - 部分成功视为重连成功
//
// ## 防死锁机制
//
// ### 问题场景
//  1. Disconnect() 可能等待 WaitGroup
//  2. 同时持有多个锁
//  3. 网络 I/O 阻塞
//  4. 心跳协程未退出
//
// ### 解决方案
//  1. 使用带超时的上下文
//  2. 异步执行断开操作
//  3. 超时后强制继续
//  4. 不等待所有资源完全释放
//
// ## 性能优化
//
// ### 并发控制
//   - 同一时刻只允许一个重连流程
//   - 使用互斥锁保护重连状态
//   - 使用 WaitGroup 管理协程生命周期
//
// ### 超时策略
//   - 断开连接：5秒超时
//   - 服务注册：30秒超时
//   - 建立连接：30秒超时（Connect 内部）
//   - 避免无限等待
//
// ## 最佳实践
//
// ### 配置建议
//   - 最大重试次数：10-20次
//   - 基础重连间隔：5秒
//   - 最大重连间隔：300秒（5分钟）
//   - 服务注册超时：30秒
//
// ### 监控指标
//   - 重连次数
//   - 重连成功率
//   - 服务恢复成功率
//   - 平均重连时间
//
// ### 日志关键字
//   - "Reconnect triggered" - 重连开始
//   - "Attempting to reconnect" - 重连尝试
//   - "Reconnect successful" - 重连成功
//   - "Re-registering services" - 服务重新注册
//   - "Service re-registered successfully" - 服务注册成功
//   - "All reconnect attempts failed" - 重连失败
package client

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"gateway/internal/tunnel/types"
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
	// 断开现有连接（使用带超时的上下文避免无限等待）
	if rm.client.controlConn.IsConnected() {
		disconnectCtx, disconnectCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer disconnectCancel()

		// 异步断开连接，避免可能的死锁
		disconnectDone := make(chan error, 1)
		go func() {
			disconnectDone <- rm.client.controlConn.Disconnect(disconnectCtx)
		}()

		select {
		case err := <-disconnectDone:
			if err != nil && err != context.DeadlineExceeded {
				logger.Warn("Failed to disconnect cleanly during reconnect", map[string]interface{}{
					"error": err.Error(),
				})
			}
		case <-time.After(6 * time.Second):
			logger.Warn("Disconnect timeout during reconnect, forcing cleanup", nil)
			// 超时后强制关闭，继续重连流程
		}
	}

	// 等待一小段时间确保资源清理完成
	time.Sleep(1 * time.Second)

	// 重新建立连接
	if err := rm.client.controlConn.Connect(rm.ctx, rm.client.config.ServerAddress, rm.client.config.ServerPort); err != nil {
		return fmt.Errorf("failed to reconnect: %w", err)
	}

	// 等待一小段时间确保连接和认证完成
	// 这可以避免在连接未完全建立时尝试注册服务
	time.Sleep(100 * time.Millisecond)

	// 验证连接是否真正建立（最多等待2秒）
	maxWait := 2 * time.Second
	waitInterval := 100 * time.Millisecond
	waited := time.Duration(0)
	for !rm.client.isConnected() && waited < maxWait {
		time.Sleep(waitInterval)
		waited += waitInterval
	}

	if !rm.client.isConnected() {
		return fmt.Errorf("connection established but not fully ready after %v", waited)
	}

	// 关键修复：先停止旧的心跳管理器，再启动新的
	// 因为旧的心跳管理器可能还在运行（running=true），直接 Start 会被跳过
	logger.Info("Stopping old heartbeat manager before restart", nil)
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 3*time.Second)
	if err := rm.client.heartbeatManager.Stop(stopCtx); err != nil {
		logger.Warn("Failed to stop old heartbeat manager", map[string]interface{}{
			"error": err.Error(),
		})
		// 继续尝试启动，不阻塞重连流程
	}
	stopCancel()

	// 等待一小段时间确保心跳管理器完全停止
	time.Sleep(200 * time.Millisecond)

	// 重新启动心跳
	heartbeatInterval := time.Duration(rm.client.config.HeartbeatInterval) * time.Second
	logger.Info("Starting new heartbeat manager after reconnect", map[string]interface{}{
		"interval": heartbeatInterval.String(),
	})
	if err := rm.client.heartbeatManager.Start(rm.ctx, heartbeatInterval); err != nil {
		return fmt.Errorf("failed to restart heartbeat: %w", err)
	}

	// 更新客户端状态
	rm.client.updateStatus(StatusConnected, true)
	rm.client.updateConnectTime()

	// 更新重连计数
	rm.client.statusMutex.Lock()
	rm.client.status.ReconnectCount++
	reconnectCount := rm.client.status.ReconnectCount
	connectionDuration := rm.client.status.ConnectionDuration
	rm.client.statusMutex.Unlock()

	// 更新数据库连接状态和重连信息
	if rm.client.storageManager != nil {
		connectTime := time.Now()
		// 更新连接状态
		if err := rm.client.storageManager.GetTunnelClientRepository().UpdateConnectionStatus(
			context.Background(),
			rm.client.config.TunnelClientId,
			"connected",
			&connectTime,
		); err != nil {
			logger.Error("Failed to update connection status in database", map[string]interface{}{
				"clientId": rm.client.config.TunnelClientId,
				"error":    err.Error(),
			})
		}

		// 更新重连信息
		if err := rm.client.storageManager.GetTunnelClientRepository().UpdateReconnectInfo(
			context.Background(),
			rm.client.config.TunnelClientId,
			reconnectCount,
			connectionDuration,
		); err != nil {
			logger.Error("Failed to update reconnect info in database", map[string]interface{}{
				"clientId": rm.client.config.TunnelClientId,
				"error":    err.Error(),
			})
		}
	}

	// 关键修复：重连成功后，重新注册所有服务
	// 确保连接完全建立后再注册服务，避免 "client is not connected" 错误
	logger.Info("Reconnection successful, re-registering services", map[string]interface{}{
		"clientId": rm.client.config.TunnelClientId,
	})

	if err := rm.reregisterServices(); err != nil {
		logger.Error("Failed to re-register services after reconnection", map[string]interface{}{
			"clientId": rm.client.config.TunnelClientId,
			"error":    err.Error(),
		})
		// 不返回错误，因为连接本身已成功建立
		// 服务重新注册失败可能是部分失败，不应该触发整个重连失败
	}

	return nil
}

// reregisterServices 重连后重新注册所有服务
//
// 在重连成功后调用，确保所有之前注册的服务重新注册到服务器。
// 这是保证服务持续可用的关键步骤。
//
// 返回:
//   - error: 重新注册失败时返回错误（部分失败不算失败）
//
// 工作流程:
//  1. 从数据库查询该客户端的所有活跃服务
//  2. 逐个重新注册到服务器
//  3. 记录成功和失败的服务数量
//  4. 只有全部失败才返回错误
func (rm *reconnectManager) reregisterServices() error {
	if rm.client.storageManager == nil {
		logger.Debug("Storage manager not available, skipping service re-registration", nil)
		return nil
	}

	// 从数据库查询该客户端的所有服务
	services, err := rm.client.storageManager.GetTunnelServiceRepository().GetByClientID(
		context.Background(),
		rm.client.config.TunnelClientId,
	)
	if err != nil {
		return fmt.Errorf("failed to query services from database: %w", err)
	}

	if len(services) == 0 {
		logger.Info("No services to re-register", map[string]interface{}{
			"clientId": rm.client.config.TunnelClientId,
		})
		return nil
	}

	logger.Info("Re-registering services after reconnection", map[string]interface{}{
		"clientId":     rm.client.config.TunnelClientId,
		"serviceCount": len(services),
	})

	// 统计注册结果
	successCount := 0
	failureCount := 0

	// 逐个注册服务
	for _, service := range services {
		// 跳过非活跃状态的服务
		if service.ActiveFlag != types.ActiveFlagYes {
			logger.Debug("Skipping inactive service during re-registration", map[string]interface{}{
				"serviceId":   service.TunnelServiceId,
				"serviceName": service.ServiceName,
			})
			continue
		}

		// 关键修复：对每个服务进行重试
		// 重连后立即注册可能失败（连接还在建立中，消息处理循环还在初始化）
		maxRetries := 3
		retryInterval := 2 * time.Second
		registered := false

		for attempt := 1; attempt <= maxRetries && !registered; attempt++ {
			// 创建带超时的上下文，避免单个服务注册阻塞太久
			registerCtx, registerCancel := context.WithTimeout(context.Background(), 30*time.Second)

			// 注册服务到服务器
			err := rm.client.RegisterService(registerCtx, service)
			registerCancel() // 立即释放资源

			if err != nil {
				logger.Error("Failed to re-register service after reconnection", map[string]interface{}{
					"serviceId":   service.TunnelServiceId,
					"serviceName": service.ServiceName,
					"attempt":     attempt,
					"maxRetries":  maxRetries,
					"error":       err.Error(),
				})

				// 如果不是最后一次尝试，等待后重试
				if attempt < maxRetries {
					logger.Info("Waiting before retry service registration", map[string]interface{}{
						"serviceId":     service.TunnelServiceId,
						"serviceName":   service.ServiceName,
						"attempt":       attempt,
						"retryInterval": retryInterval.String(),
					})
					time.Sleep(retryInterval)
					// 指数退避：每次重试间隔翻倍
					retryInterval *= 2
				} else {
					// 所有重试都失败了
					failureCount++
				}
			} else {
				logger.Info("Service re-registered successfully after reconnection", map[string]interface{}{
					"serviceId":   service.TunnelServiceId,
					"serviceName": service.ServiceName,
					"serviceType": service.ServiceType,
					"attempt":     attempt,
				})
				successCount++
				registered = true
			}
		}
	}

	logger.Info("Service re-registration completed", map[string]interface{}{
		"clientId":     rm.client.config.TunnelClientId,
		"totalCount":   len(services),
		"successCount": successCount,
		"failureCount": failureCount,
	})

	// 只有所有服务都失败了才返回错误
	if failureCount > 0 && successCount == 0 {
		return fmt.Errorf("all %d services failed to re-register", failureCount)
	}

	// 部分成功或全部成功都不算失败
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
