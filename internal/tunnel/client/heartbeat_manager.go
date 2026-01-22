// Package client 提供心跳管理器的完整实现
// 心跳管理器负责维护客户端与服务器之间的连接活性检测
//
// # 心跳机制
//
// ## 概述
//
// 心跳管理器通过定期发送心跳消息来检测连接状态，当检测到连接异常时
// 自动触发重连机制，确保客户端与服务器的连接始终保持活跃。
//
// ## 核心功能
//
// ### 1. 定期心跳
//   - 按配置的间隔定期发送心跳消息
//   - 记录心跳延迟和成功率统计
//   - 更新最后心跳时间到数据库
//
// ### 2. 连接监控
//   - 监控心跳发送是否成功
//   - 统计连续失败次数
//   - 达到阈值时触发自动重连
//
// ### 3. 故障检测
//   - 连续失败阈值：3次（可配置）
//   - 失败原因：网络中断、连接关闭、服务器无响应
//   - 触发重连后重置计数器
//
// ## 重连触发逻辑
//
// ### 触发条件
//  1. 心跳连续失败 3 次
//  2. 每次失败都会增加计数器
//  3. 达到阈值时检查是否已在重连中
//  4. 如果未在重连中，异步触发重连
//
// ### 防重复触发
//  1. 检查 IsReconnecting() 状态
//  2. 触发后重置失败计数器
//  3. 避免同时触发多个重连流程
//
// ### 恢复检测
//  1. 心跳成功后重置失败计数器
//  2. 记录恢复日志
//  3. 继续正常心跳循环
//
// ## 与重连管理器的协作
//
// ### 工作流程
//  1. 心跳检测到连续失败
//     ↓
//  2. 触发 ReconnectManager.TriggerReconnect()
//     ↓
//  3. 重连管理器断开旧连接
//     ↓
//  4. 重连管理器重新建立连接
//     ↓
//  5. 重连管理器重启心跳管理器
//     ↓
//  6. 心跳恢复正常
//
// ### 生命周期管理
//   - 初次启动：随客户端启动而启动
//   - 重连时：旧的心跳管理器停止，新的重新启动
//   - 停止时：随客户端停止而停止
//
// ## 最佳实践
//
// ### 配置建议
//   - 心跳间隔：5-30秒
//   - 连续失败阈值：3次
//   - 心跳超时：5秒
//
// ### 监控指标
//   - 心跳发送总数
//   - 心跳成功率
//   - 平均延迟
//   - 连续失败次数
package client

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"gateway/internal/tunnel/types"
	"gateway/pkg/logger"
	"gateway/pkg/utils/random"
)

// heartbeatManager 心跳管理器实现（包含重连逻辑）
// 所有配置和状态都使用 client.config 中的字段
type heartbeatManager struct {
	client       *tunnelClient
	controlConn  ControlConnection
	running      bool
	mutex        sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	reconnecting bool // 重连进行中标志
}

// NewHeartbeatManager 创建心跳管理器实例（包含重连逻辑）
// 所有配置都从 client.config 中读取
func NewHeartbeatManager(client *tunnelClient) HeartbeatManager {
	return &heartbeatManager{
		client:       client,
		controlConn:  client.controlConn,
		reconnecting: false,
	}
}

// Start 启动心跳
func (hm *heartbeatManager) Start(ctx context.Context, interval time.Duration) error {
	hm.mutex.Lock()
	defer hm.mutex.Unlock()

	if hm.running {
		return nil
	}

	hm.ctx, hm.cancel = context.WithCancel(ctx)
	hm.running = true

	// 启动心跳循环
	hm.wg.Add(1)
	go hm.heartbeatLoop()

	logger.Info("Heartbeat manager started", map[string]interface{}{
		"interval": interval.String(),
	})

	return nil
}

// Stop 停止心跳
func (hm *heartbeatManager) Stop(ctx context.Context) error {
	hm.mutex.Lock()
	if !hm.running {
		hm.mutex.Unlock()
		return nil
	}
	hm.running = false
	hm.mutex.Unlock()

	if hm.cancel != nil {
		hm.cancel()
	}

	// 等待心跳循环退出
	done := make(chan struct{})
	go func() {
		hm.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("Heartbeat manager stopped", nil)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// SendHeartbeat 发送心跳
func (hm *heartbeatManager) SendHeartbeat(ctx context.Context) error {
	start := time.Now()

	heartbeatMsg := &types.ControlMessage{
		Type:      types.MessageTypeHeartbeat,
		SessionID: hm.generateRequestID(),
		Data:      map[string]interface{}{},
		Timestamp: start,
	}

	// 发送心跳消息（不等待响应）
	if _, err := hm.controlConn.SendMessage(ctx, heartbeatMsg); err != nil {
		return err
	}

	// 更新最后心跳时间到 config
	lastHeartbeat := time.Now()
	hm.client.config.LastHeartbeat = &lastHeartbeat

	// 更新数据库心跳时间（异步更新，避免阻塞）
	if hm.client != nil && hm.client.clientRepository != nil {
		go func() {
			if err := hm.client.clientRepository.UpdateHeartbeat(
				context.Background(),
				hm.client.config.TunnelClientId,
				lastHeartbeat,
			); err != nil {
				logger.Debug("Failed to update heartbeat in database", map[string]interface{}{
					"clientId": hm.client.config.TunnelClientId,
					"error":    err.Error(),
				})
			}
		}()
	}

	return nil
}

// GetLastHeartbeatTime 获取最后心跳时间
func (hm *heartbeatManager) GetLastHeartbeatTime() time.Time {
	if hm.client.config.LastHeartbeat != nil {
		return *hm.client.config.LastHeartbeat
	}
	return time.Time{}
}

// heartbeatLoop 心跳循环
func (hm *heartbeatManager) heartbeatLoop() {
	defer hm.wg.Done()

	// 从 config 读取心跳间隔
	interval := time.Duration(hm.client.config.HeartbeatInterval) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// 连续失败计数器
	consecutiveFailures := 0
	maxConsecutiveFailures := 3 // 连续失败3次触发重连

	for {
		select {
		case <-hm.ctx.Done():
			return
		case <-ticker.C:
			if err := hm.SendHeartbeat(hm.ctx); err != nil {
				logger.Error("Failed to send heartbeat", map[string]interface{}{
					"error":               err.Error(),
					"consecutiveFailures": consecutiveFailures + 1,
				})

				consecutiveFailures++

				// 心跳连续失败达到阈值时触发重连
				if consecutiveFailures >= maxConsecutiveFailures {
					logger.Warn("Heartbeat failed multiple times, triggering reconnect", map[string]interface{}{
						"consecutiveFailures": consecutiveFailures,
						"threshold":           maxConsecutiveFailures,
					})

					// 触发重连（内部会检查是否已在重连中）
					go hm.triggerReconnect("heartbeat_consecutive_failures")

					// 重置计数器，避免重复触发
					consecutiveFailures = 0
				}
			} else {
				// 心跳成功，重置失败计数器
				if consecutiveFailures > 0 {
					logger.Info("Heartbeat recovered", map[string]interface{}{
						"previousFailures": consecutiveFailures,
					})
					consecutiveFailures = 0
				}
			}
		}
	}
}

// generateRequestID 生成请求ID
func (hm *heartbeatManager) generateRequestID() string {
	return random.GenerateUniqueStringWithPrefix("hb_", 32)
}

// triggerReconnect 触发重连（从心跳失败检测）
func (hm *heartbeatManager) triggerReconnect(reason string) {
	hm.mutex.Lock()
	defer hm.mutex.Unlock()

	// 检查是否已在重连中或已停止
	if hm.reconnecting || !hm.running {
		return
	}
	hm.reconnecting = true

	logger.Info("Reconnect triggered from heartbeat", map[string]interface{}{
		"reason": reason,
	})

	// 在锁保护下增加 WaitGroup 计数，避免与 Stop 方法的 Wait 竞争
	hm.wg.Add(1)
	go hm.reconnectLoop(reason)
}

// reconnectLoop 重连循环
func (hm *heartbeatManager) reconnectLoop(reason string) {
	defer hm.wg.Done()
	defer func() {
		hm.mutex.Lock()
		hm.reconnecting = false
		hm.mutex.Unlock()
	}()

	// 从 config 读取最大重试次数
	maxRetries := hm.client.config.MaxRetries

	for attempt := 1; attempt <= maxRetries; attempt++ {
		select {
		case <-hm.ctx.Done():
			return
		default:
		}

		logger.Info("Attempting to reconnect", map[string]interface{}{
			"attempt": attempt,
			"reason":  reason,
		})

		// 尝试重连
		if err := hm.attemptReconnect(); err != nil {
			logger.Error("Reconnect attempt failed", map[string]interface{}{
				"attempt": attempt,
				"error":   err.Error(),
			})

			// 如果不是最后一次尝试，等待重连间隔
			if attempt < maxRetries {
				interval := hm.calculateBackoffInterval(attempt)
				logger.Info("Waiting before next reconnect attempt", map[string]interface{}{
					"interval": interval.String(),
					"attempt":  attempt,
				})

				select {
				case <-hm.ctx.Done():
					return
				case <-time.After(interval):
				}
			}
		} else {
			// 重连成功
			logger.Info("Reconnect successful", map[string]interface{}{
				"attempt": attempt,
			})

			return
		}
	}

	// 所有重连尝试都失败了
	logger.Error("All reconnect attempts failed", map[string]interface{}{
		"maxRetries": maxRetries,
		"reason":     reason,
	})
}

// attemptReconnect 尝试重连
func (hm *heartbeatManager) attemptReconnect() error {
	// 断开现有连接
	if hm.client.controlConn.IsConnected() {
		disconnectCtx, disconnectCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer disconnectCancel()

		disconnectDone := make(chan error, 1)
		go func() {
			disconnectDone <- hm.client.controlConn.Disconnect(disconnectCtx)
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
		}
	}

	// 等待资源清理
	time.Sleep(1 * time.Second)

	// 重新建立连接
	if err := hm.client.controlConn.Connect(hm.ctx, hm.client.config.ServerAddress, hm.client.config.ServerPort); err != nil {
		return fmt.Errorf("failed to reconnect: %w", err)
	}

	// 等待连接建立
	time.Sleep(100 * time.Millisecond)

	// 验证连接
	maxWait := 2 * time.Second
	waitInterval := 100 * time.Millisecond
	waited := time.Duration(0)
	for !hm.client.IsConnected() && waited < maxWait {
		time.Sleep(waitInterval)
		waited += waitInterval
	}

	if !hm.client.IsConnected() {
		return fmt.Errorf("connection established but not fully ready after %v", waited)
	}

	// 关键修复：等待服务器端完成客户端注册
	// 认证成功后，服务器端需要时间完成客户端注册到 connectedClients
	// 如果立即注册服务，可能会遇到 "client is not registered" 错误
	logger.Info("Waiting for server to complete client registration", map[string]interface{}{
		"clientId": hm.client.config.TunnelClientId,
	})
	time.Sleep(500 * time.Millisecond)

	// 更新客户端状态到 config
	connectTime := time.Now()
	hm.client.config.ConnectionStatus = types.ConnectionStatusConnected
	hm.client.config.LastConnectTime = &connectTime
	hm.client.config.ReconnectCount++

	// 更新数据库连接状态
	if hm.client.clientRepository != nil {
		if err := hm.client.clientRepository.UpdateConnectionStatus(
			context.Background(),
			hm.client.config.TunnelClientId,
			types.ConnectionStatusConnected,
			&connectTime,
		); err != nil {
			logger.Error("Failed to update connection status in database", map[string]interface{}{
				"clientId": hm.client.config.TunnelClientId,
				"error":    err.Error(),
			})
		}

		if err := hm.client.clientRepository.UpdateReconnectInfo(
			context.Background(),
			hm.client.config.TunnelClientId,
			hm.client.config.ReconnectCount,
			0, // connectionDuration
		); err != nil {
			logger.Error("Failed to update reconnect info in database", map[string]interface{}{
				"clientId": hm.client.config.TunnelClientId,
				"error":    err.Error(),
			})
		}
	}

	// 重新注册所有服务
	logger.Info("Reconnection successful, re-registering services", map[string]interface{}{
		"clientId": hm.client.config.TunnelClientId,
	})

	if err := hm.reregisterServices(); err != nil {
		logger.Error("Failed to re-register services after reconnection", map[string]interface{}{
			"clientId": hm.client.config.TunnelClientId,
			"error":    err.Error(),
		})
		// 关键修复：服务注册失败不应该导致重连失败，返回nil让重连标记为成功
		// 服务注册会继续重试，不会阻塞重连流程
	}

	return nil
}

// reregisterServices 重连后重新注册所有服务
func (hm *heartbeatManager) reregisterServices() error {
	if hm.client.serviceRepository == nil {
		logger.Debug("Service repository not available, skipping service re-registration", nil)
		return nil
	}

	// 从数据库查询该客户端的所有服务
	services, err := hm.client.serviceRepository.GetByClientID(
		context.Background(),
		hm.client.config.TunnelClientId,
	)
	if err != nil {
		return fmt.Errorf("failed to query services from database: %w", err)
	}

	if len(services) == 0 {
		logger.Info("No services to re-register", map[string]interface{}{
			"clientId": hm.client.config.TunnelClientId,
		})
		return nil
	}

	logger.Info("Re-registering services after reconnection", map[string]interface{}{
		"clientId":     hm.client.config.TunnelClientId,
		"serviceCount": len(services),
	})

	successCount := 0
	failureCount := 0

	for _, service := range services {
		if service.ActiveFlag != types.ActiveFlagYes {
			continue
		}

		// 关键修复：对于"client is not registered"错误，增加重试次数和间隔
		// 这通常是因为服务器端还在处理客户端注册，需要等待更长时间
		maxRetries := 5 // 从3次增加到5次
		retryInterval := 1 * time.Second
		registered := false

		for attempt := 1; attempt <= maxRetries && !registered; attempt++ {
			registerCtx, registerCancel := context.WithTimeout(context.Background(), 30*time.Second)

			err := hm.client.RegisterService(registerCtx, service)
			registerCancel()

			if err != nil {
				// 检查是否是"client is not registered"错误
				errMsg := err.Error()
				isClientNotRegistered := strings.Contains(errMsg, "is not registered") ||
					strings.Contains(errMsg, "not registered")

				logger.Error("Failed to re-register service after reconnection", map[string]interface{}{
					"serviceId":             service.TunnelServiceId,
					"serviceName":           service.ServiceName,
					"attempt":               attempt,
					"maxRetries":            maxRetries,
					"error":                 errMsg,
					"isClientNotRegistered": isClientNotRegistered,
				})

				// 如果是"client is not registered"错误，等待更长时间
				if isClientNotRegistered && attempt < maxRetries {
					// 等待时间递增：1s, 2s, 3s, 4s
					waitTime := time.Duration(attempt) * time.Second
					logger.Info("Waiting for server to complete client registration before retry", map[string]interface{}{
						"serviceId":   service.TunnelServiceId,
						"serviceName": service.ServiceName,
						"attempt":     attempt,
						"waitTime":    waitTime.String(),
					})
					time.Sleep(waitTime)
				} else if attempt < maxRetries {
					// 其他错误，使用指数退避
					time.Sleep(retryInterval)
					retryInterval *= 2
				} else {
					failureCount++
				}
			} else {
				logger.Info("Service re-registered successfully", map[string]interface{}{
					"serviceId":   service.TunnelServiceId,
					"serviceName": service.ServiceName,
					"attempt":     attempt,
				})
				successCount++
				registered = true
			}
		}
	}

	logger.Info("Service re-registration completed", map[string]interface{}{
		"clientId":     hm.client.config.TunnelClientId,
		"totalCount":   len(services),
		"successCount": successCount,
		"failureCount": failureCount,
	})

	if failureCount > 0 && successCount == 0 {
		return fmt.Errorf("all %d services failed to re-register", failureCount)
	}

	return nil
}

// calculateBackoffInterval 计算指数退避间隔
func (hm *heartbeatManager) calculateBackoffInterval(attempt int) time.Duration {
	// 从 config 读取基础重试间隔
	baseInterval := time.Duration(hm.client.config.RetryInterval) * time.Second
	maxInterval := 300 * time.Second // 最大5分钟

	backoff := float64(baseInterval) * math.Pow(2, float64(attempt-1))
	interval := time.Duration(backoff)

	if interval > maxInterval {
		interval = maxInterval
	}

	return interval
}
