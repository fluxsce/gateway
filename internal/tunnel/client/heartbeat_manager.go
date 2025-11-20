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
	"sync"
	"time"

	"gateway/internal/tunnel/types"
	"gateway/pkg/logger"
)

// heartbeatManager 心跳管理器实现
type heartbeatManager struct {
	client        *tunnelClient
	controlConn   ControlConnection
	interval      time.Duration
	lastHeartbeat time.Time
	stats         *HeartbeatStats
	running       bool
	mutex         sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
}

// NewHeartbeatManager 创建心跳管理器实例
func NewHeartbeatManager(client *tunnelClient) HeartbeatManager {
	return &heartbeatManager{
		client:      client,
		controlConn: client.controlConn,
		stats: &HeartbeatStats{
			TotalSent:      0,
			TotalReceived:  0,
			AverageLatency: 0,
			MaxLatency:     0,
			MinLatency:     999999,
			FailedCount:    0,
		},
	}
}

// Start 启动心跳
func (hm *heartbeatManager) Start(ctx context.Context, interval time.Duration) error {
	hm.mutex.Lock()
	defer hm.mutex.Unlock()

	if hm.running {
		return nil
	}

	hm.interval = interval
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

	if err := hm.controlConn.SendMessage(ctx, heartbeatMsg); err != nil {
		hm.updateStats(false, 0)
		return err
	}

	// 计算延迟（简化处理）
	latency := time.Since(start).Milliseconds()
	hm.updateStats(true, float64(latency))
	hm.lastHeartbeat = time.Now()

	// 更新数据库心跳时间（定期更新，避免频繁写数据库）
	if hm.client != nil && hm.client.storageManager != nil {
		// 只有在心跳间隔较长（大于10秒）或每5次心跳更新一次数据库
		if hm.interval > 10*time.Second || hm.stats.TotalSent%5 == 0 {
			go func() {
				if err := hm.client.storageManager.GetTunnelClientRepository().UpdateHeartbeat(
					context.Background(),
					hm.client.config.TunnelClientId,
					hm.lastHeartbeat,
				); err != nil {
					logger.Debug("Failed to update heartbeat in database", map[string]interface{}{
						"clientId": hm.client.config.TunnelClientId,
						"error":    err.Error(),
					})
				}
			}()
		}
	}

	return nil
}

// GetLastHeartbeatTime 获取最后心跳时间
func (hm *heartbeatManager) GetLastHeartbeatTime() time.Time {
	hm.mutex.RLock()
	defer hm.mutex.RUnlock()
	return hm.lastHeartbeat
}

// GetHeartbeatStats 获取心跳统计
func (hm *heartbeatManager) GetHeartbeatStats() *HeartbeatStats {
	hm.mutex.RLock()
	defer hm.mutex.RUnlock()

	// 返回统计副本
	stats := *hm.stats
	return &stats
}

// heartbeatLoop 心跳循环
func (hm *heartbeatManager) heartbeatLoop() {
	defer hm.wg.Done()

	ticker := time.NewTicker(hm.interval)
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

				// 关键修复：心跳连续失败达到阈值时触发重连
				// 这是心跳机制检测连接断开的主要方式
				if consecutiveFailures >= maxConsecutiveFailures {
					logger.Warn("Heartbeat failed multiple times, triggering reconnect", map[string]interface{}{
						"consecutiveFailures": consecutiveFailures,
						"threshold":           maxConsecutiveFailures,
					})

					// 检查是否已在重连中
					if hm.client.reconnectManager != nil && !hm.client.reconnectManager.IsReconnecting() {
						// 异步触发重连，避免阻塞心跳循环
						go func() {
							if err := hm.client.reconnectManager.TriggerReconnect(context.Background(), "heartbeat_consecutive_failures"); err != nil {
								logger.Error("Failed to trigger reconnect from heartbeat", map[string]interface{}{
									"error": err.Error(),
								})
							}
						}()
					}

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

// updateStats 更新统计信息
func (hm *heartbeatManager) updateStats(success bool, latency float64) {
	hm.mutex.Lock()
	defer hm.mutex.Unlock()

	hm.stats.TotalSent++

	if success {
		hm.stats.TotalReceived++
		hm.stats.LastSentTime = time.Now()
		hm.stats.LastReceivedTime = time.Now()

		// 更新延迟统计
		if latency > hm.stats.MaxLatency {
			hm.stats.MaxLatency = latency
		}
		if latency < hm.stats.MinLatency {
			hm.stats.MinLatency = latency
		}

		// 更新平均延迟
		if hm.stats.TotalReceived == 1 {
			hm.stats.AverageLatency = latency
		} else {
			hm.stats.AverageLatency = (hm.stats.AverageLatency*float64(hm.stats.TotalReceived-1) + latency) / float64(hm.stats.TotalReceived)
		}
	} else {
		hm.stats.FailedCount++
	}
}

// generateRequestID 生成请求ID
func (hm *heartbeatManager) generateRequestID() string {
	return "heartbeat_" + time.Now().Format("20060102150405")
}
