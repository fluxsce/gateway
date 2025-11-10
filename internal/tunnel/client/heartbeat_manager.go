// Package client 提供心跳管理器的完整实现
// 心跳管理器负责维护客户端与服务器之间的连接活性检测
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

	for {
		select {
		case <-hm.ctx.Done():
			return
		case <-ticker.C:
			if err := hm.SendHeartbeat(hm.ctx); err != nil {
				logger.Error("Failed to send heartbeat", map[string]interface{}{
					"error": err.Error(),
				})
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
