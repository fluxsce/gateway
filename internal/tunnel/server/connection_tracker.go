// Package server 提供连接跟踪器的完整实现
// 连接跟踪器负责跟踪和统计隧道连接，提供实时监控和历史分析功能
package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gateway/internal/tunnel/storage"
	"gateway/internal/tunnel/types"
	"gateway/pkg/logger"
)

// connectionTracker 连接跟踪器实现
// 实现 ConnectionTracker 接口，管理连接跟踪和统计
type connectionTracker struct {
	storage     storage.RepositoryManager
	connections map[string]*connectionInfo
	connMutex   sync.RWMutex
	stats       *connectionStatistics
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

// connectionInfo 连接信息
type connectionInfo struct {
	connection   *types.TunnelConnection
	startTime    time.Time
	lastActivity time.Time
	stats        *ConnectionStats
	mutex        sync.RWMutex
}

// connectionStatistics 连接统计信息
type connectionStatistics struct {
	totalConnections     int64
	activeConnections    int64
	totalBytesReceived   int64
	totalBytesSent       int64
	totalPacketsReceived int64
	totalPacketsSent     int64
	errorCount           int64
	averageLatency       float64
	mutex                sync.RWMutex
}

// NewConnectionTrackerImpl 创建新的连接跟踪器实例
//
// 参数:
//   - storage: 存储管理器，用于持久化连接数据
//
// 返回:
//   - ConnectionTracker: 连接跟踪器接口实例
//
// 功能:
//   - 初始化连接跟踪器
//   - 创建连接映射表和统计信息
//   - 启动定期统计更新任务
func NewConnectionTrackerImpl(storage storage.RepositoryManager) ConnectionTracker {
	ctx, cancel := context.WithCancel(context.Background())

	tracker := &connectionTracker{
		storage:     storage,
		connections: make(map[string]*connectionInfo),
		stats: &connectionStatistics{
			totalConnections:     0,
			activeConnections:    0,
			totalBytesReceived:   0,
			totalBytesSent:       0,
			totalPacketsReceived: 0,
			totalPacketsSent:     0,
			errorCount:           0,
			averageLatency:       0.0,
		},
		ctx:    ctx,
		cancel: cancel,
	}

	// 启动定期统计更新任务
	tracker.wg.Add(1)
	go tracker.statisticsWorker()

	return tracker
}

// TrackConnection 跟踪连接
//
// 参数:
//   - ctx: 上下文
//   - connection: 要跟踪的连接对象
//
// 返回:
//   - error: 跟踪失败时返回错误
//
// 功能:
//   - 创建连接跟踪记录
//   - 初始化连接统计信息
//   - 持久化连接到数据库
//   - 添加到内存映射表
func (ct *connectionTracker) TrackConnection(ctx context.Context, connection *types.TunnelConnection) error {
	// 设置连接基本信息
	connection.TunnelConnectionId = ct.generateConnectionID()
	connection.ConnectionStatus = types.ConnectionStatusConnected
	connection.StartTime = time.Now()
	connection.BytesReceived = 0
	connection.BytesSent = 0
	connection.PacketsReceived = 0
	connection.PacketsSent = 0
	connection.ErrorCount = 0
	connection.LastActivity = &[]time.Time{time.Now()}[0]
	connection.AddTime = time.Now()
	connection.EditTime = time.Now()
	connection.AddWho = "system"
	connection.EditWho = "system"
	connection.ActiveFlag = types.ActiveFlagYes

	// 持久化到数据库
	if err := ct.storage.GetTunnelConnectionRepository().Create(ctx, connection); err != nil {
		return fmt.Errorf("failed to create connection in database: %w", err)
	}

	// 创建连接信息
	connInfo := &connectionInfo{
		connection:   connection,
		startTime:    time.Now(),
		lastActivity: time.Now(),
		stats: &ConnectionStats{
			BytesReceived:   0,
			BytesSent:       0,
			PacketsReceived: 0,
			PacketsSent:     0,
			LastActivity:    time.Now(),
			Latency:         0.0,
			ErrorCount:      0,
		},
	}

	// 添加到内存映射
	ct.connMutex.Lock()
	ct.connections[connection.TunnelConnectionId] = connInfo
	ct.connMutex.Unlock()

	// 更新统计信息
	ct.stats.mutex.Lock()
	ct.stats.totalConnections++
	ct.stats.activeConnections++
	ct.stats.mutex.Unlock()

	logger.Info("Connection tracked", map[string]interface{}{
		"connectionId":   connection.TunnelConnectionId,
		"sourceAddress":  connection.SourceIpAddress,
		"targetAddress":  connection.TargetIpAddress,
		"connectionType": connection.ConnectionType,
	})

	return nil
}

// UpdateConnectionStats 更新连接统计
//
// 参数:
//   - ctx: 上下文
//   - connectionID: 连接ID
//   - stats: 新的统计数据
//
// 返回:
//   - error: 更新失败时返回错误
//
// 功能:
//   - 更新内存中的连接统计
//   - 计算增量数据
//   - 定期批量更新到数据库
func (ct *connectionTracker) UpdateConnectionStats(ctx context.Context, connectionID string, stats *ConnectionStats) error {
	ct.connMutex.RLock()
	connInfo, exists := ct.connections[connectionID]
	ct.connMutex.RUnlock()

	if !exists {
		return fmt.Errorf("connection %s not found", connectionID)
	}

	connInfo.mutex.Lock()

	// 计算增量
	deltaReceived := stats.BytesReceived - connInfo.stats.BytesReceived
	deltaSent := stats.BytesSent - connInfo.stats.BytesSent
	deltaPacketsReceived := stats.PacketsReceived - connInfo.stats.PacketsReceived
	deltaPacketsSent := stats.PacketsSent - connInfo.stats.PacketsSent
	deltaErrors := stats.ErrorCount - connInfo.stats.ErrorCount

	// 更新连接统计
	connInfo.stats = stats
	connInfo.lastActivity = stats.LastActivity
	connInfo.connection.LastActivity = &stats.LastActivity
	connInfo.connection.BytesReceived += deltaReceived
	connInfo.connection.BytesSent += deltaSent
	connInfo.connection.PacketsReceived += deltaPacketsReceived
	connInfo.connection.PacketsSent += deltaPacketsSent
	connInfo.connection.ErrorCount += int(deltaErrors)
	connInfo.connection.ConnectionLatency = stats.Latency

	connInfo.mutex.Unlock()

	// 更新全局统计
	ct.stats.mutex.Lock()
	ct.stats.totalBytesReceived += deltaReceived
	ct.stats.totalBytesSent += deltaSent
	ct.stats.totalPacketsReceived += deltaPacketsReceived
	ct.stats.totalPacketsSent += deltaPacketsSent
	ct.stats.errorCount += int64(deltaErrors)

	// 更新平均延迟（简单移动平均）
	if ct.stats.activeConnections > 0 {
		ct.stats.averageLatency = (ct.stats.averageLatency*float64(ct.stats.activeConnections-1) + stats.Latency) / float64(ct.stats.activeConnections)
	}
	ct.stats.mutex.Unlock()

	// 异步更新数据库
	go func() {
		if err := ct.storage.GetTunnelConnectionRepository().UpdateTrafficStats(
			context.Background(), connectionID,
			connInfo.connection.BytesReceived, connInfo.connection.BytesSent,
			connInfo.connection.PacketsReceived, connInfo.connection.PacketsSent,
		); err != nil {
			logger.Error("Failed to update connection stats in database", map[string]interface{}{
				"error":        err.Error(),
				"connectionId": connectionID,
			})
		}
	}()

	return nil
}

// CloseConnection 关闭连接跟踪
//
// 参数:
//   - ctx: 上下文
//   - connectionID: 连接ID
//
// 返回:
//   - error: 关闭失败时返回错误
//
// 功能:
//   - 从内存映射中移除连接
//   - 更新数据库中的连接状态
//   - 更新统计信息
func (ct *connectionTracker) CloseConnection(ctx context.Context, connectionID string) error {
	ct.connMutex.Lock()
	connInfo, exists := ct.connections[connectionID]
	if exists {
		delete(ct.connections, connectionID)
	}
	ct.connMutex.Unlock()

	if !exists {
		return fmt.Errorf("connection %s not found", connectionID)
	}

	// 计算连接持续时间
	endTime := time.Now()
	duration := endTime.Sub(connInfo.startTime).Milliseconds()

	// 更新数据库状态
	if err := ct.storage.GetTunnelConnectionRepository().CloseConnection(ctx, connectionID, endTime, duration); err != nil {
		logger.Error("Failed to update connection close status in database", map[string]interface{}{
			"error":        err.Error(),
			"connectionId": connectionID,
		})
	}

	// 更新统计信息
	ct.stats.mutex.Lock()
	ct.stats.activeConnections--
	ct.stats.mutex.Unlock()

	logger.Info("Connection closed", map[string]interface{}{
		"connectionId":  connectionID,
		"duration":      duration,
		"bytesReceived": connInfo.connection.BytesReceived,
		"bytesSent":     connInfo.connection.BytesSent,
	})

	return nil
}

// GetActiveConnections 获取活跃连接
//
// 参数:
//   - ctx: 上下文
//
// 返回:
//   - []*types.TunnelConnection: 活跃连接列表
//
// 功能:
//   - 返回所有当前活跃的连接
//   - 包含实时统计信息
func (ct *connectionTracker) GetActiveConnections(ctx context.Context) []*types.TunnelConnection {
	ct.connMutex.RLock()
	defer ct.connMutex.RUnlock()

	var connections []*types.TunnelConnection
	for _, connInfo := range ct.connections {
		connInfo.mutex.RLock()
		if connInfo.connection.ConnectionStatus == types.ConnectionStatusConnected {
			connections = append(connections, connInfo.connection)
		}
		connInfo.mutex.RUnlock()
	}

	return connections
}

// GetConnectionStats 获取连接统计
//
// 参数:
//   - ctx: 上下文
//   - timeRange: 时间范围
//
// 返回:
//   - *ConnectionStatsReport: 连接统计报告
//   - error: 获取失败时返回错误
//
// 功能:
//   - 生成指定时间范围内的连接统计报告
//   - 包含流量、延迟、错误率等指标
//   - 提供来源IP统计
func (ct *connectionTracker) GetConnectionStats(ctx context.Context, timeRange TimeRange) (*ConnectionStatsReport, error) {
	// 从数据库获取历史数据
	connections, err := ct.storage.GetTunnelConnectionRepository().GetConnectionsByDateRange(
		ctx, timeRange.StartTime, timeRange.EndTime,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get connections from database: %w", err)
	}

	// 计算统计信息
	report := &ConnectionStatsReport{
		TimeRange:         timeRange,
		TotalConnections:  len(connections),
		ActiveConnections: int(ct.stats.activeConnections),
		TotalTraffic:      0,
		AverageLatency:    0.0,
		ErrorRate:         0.0,
		TopSources:        make([]*SourceStats, 0),
	}

	// 统计数据
	var totalLatency float64
	var latencyCount int
	var totalErrors int64
	sourceMap := make(map[string]*SourceStats)

	for _, conn := range connections {
		// 累计流量
		report.TotalTraffic += conn.BytesReceived + conn.BytesSent

		// 累计延迟
		if conn.ConnectionLatency > 0 {
			totalLatency += conn.ConnectionLatency
			latencyCount++
		}

		// 累计错误
		totalErrors += int64(conn.ErrorCount)

		// 统计来源IP
		sourceIP := conn.SourceIpAddress
		if source, exists := sourceMap[sourceIP]; exists {
			source.ConnectionCount++
			source.TotalTraffic += conn.BytesReceived + conn.BytesSent
		} else {
			sourceMap[sourceIP] = &SourceStats{
				SourceIP:        sourceIP,
				ConnectionCount: 1,
				TotalTraffic:    conn.BytesReceived + conn.BytesSent,
			}
		}
	}

	// 计算平均延迟
	if latencyCount > 0 {
		report.AverageLatency = totalLatency / float64(latencyCount)
	}

	// 计算错误率
	if len(connections) > 0 {
		report.ErrorRate = float64(totalErrors) / float64(len(connections)) * 100
	}

	// 转换来源统计并排序（取前10个）
	for _, source := range sourceMap {
		report.TopSources = append(report.TopSources, source)
	}

	// 简单排序（按连接数降序）
	for i := 0; i < len(report.TopSources)-1; i++ {
		for j := i + 1; j < len(report.TopSources); j++ {
			if report.TopSources[j].ConnectionCount > report.TopSources[i].ConnectionCount {
				report.TopSources[i], report.TopSources[j] = report.TopSources[j], report.TopSources[i]
			}
		}
	}

	// 限制为前10个
	if len(report.TopSources) > 10 {
		report.TopSources = report.TopSources[:10]
	}

	return report, nil
}

// statisticsWorker 统计信息更新工作协程
func (ct *connectionTracker) statisticsWorker() {
	defer ct.wg.Done()

	ticker := time.NewTicker(30 * time.Second) // 每30秒更新一次统计
	defer ticker.Stop()

	for {
		select {
		case <-ct.ctx.Done():
			return
		case <-ticker.C:
			ct.updateStatistics()
		}
	}
}

// updateStatistics 更新统计信息
func (ct *connectionTracker) updateStatistics() {
	ct.connMutex.RLock()
	activeCount := len(ct.connections)
	ct.connMutex.RUnlock()

	ct.stats.mutex.Lock()
	ct.stats.activeConnections = int64(activeCount)
	ct.stats.mutex.Unlock()

	// 清理过期连接
	ct.cleanupExpiredConnections()
}

// cleanupExpiredConnections 清理过期连接
func (ct *connectionTracker) cleanupExpiredConnections() {
	ct.connMutex.RLock()
	var expiredConnections []string
	now := time.Now()

	for connectionID, connInfo := range ct.connections {
		connInfo.mutex.RLock()
		// 如果连接超过1小时没有活动，认为已过期
		if now.Sub(connInfo.lastActivity) > time.Hour {
			expiredConnections = append(expiredConnections, connectionID)
		}
		connInfo.mutex.RUnlock()
	}
	ct.connMutex.RUnlock()

	// 关闭过期连接
	for _, connectionID := range expiredConnections {
		if err := ct.CloseConnection(context.Background(), connectionID); err != nil {
			logger.Error("Failed to close expired connection", map[string]interface{}{
				"error":        err.Error(),
				"connectionId": connectionID,
			})
		}
	}

	if len(expiredConnections) > 0 {
		logger.Info("Cleaned up expired connections", map[string]interface{}{
			"count": len(expiredConnections),
		})
	}
}

// generateConnectionID 生成连接ID
func (ct *connectionTracker) generateConnectionID() string {
	return fmt.Sprintf("conn_%d_%d", time.Now().UnixNano(), ct.stats.totalConnections)
}

// GetStatsSummary 获取统计摘要
func (ct *connectionTracker) GetStatsSummary() map[string]interface{} {
	ct.stats.mutex.RLock()
	defer ct.stats.mutex.RUnlock()

	return map[string]interface{}{
		"totalConnections":     ct.stats.totalConnections,
		"activeConnections":    ct.stats.activeConnections,
		"totalBytesReceived":   ct.stats.totalBytesReceived,
		"totalBytesSent":       ct.stats.totalBytesSent,
		"totalPacketsReceived": ct.stats.totalPacketsReceived,
		"totalPacketsSent":     ct.stats.totalPacketsSent,
		"errorCount":           ct.stats.errorCount,
		"averageLatency":       ct.stats.averageLatency,
	}
}

// RecordError 记录连接错误
func (ct *connectionTracker) RecordError(ctx context.Context, connectionID string, errorMessage string) error {
	ct.connMutex.RLock()
	connInfo, exists := ct.connections[connectionID]
	ct.connMutex.RUnlock()

	if !exists {
		return fmt.Errorf("connection %s not found", connectionID)
	}

	connInfo.mutex.Lock()
	connInfo.stats.ErrorCount++
	connInfo.connection.ErrorCount++
	connInfo.connection.LastErrorMessage = &errorMessage
	connInfo.mutex.Unlock()

	// 更新全局错误计数
	ct.stats.mutex.Lock()
	ct.stats.errorCount++
	ct.stats.mutex.Unlock()

	// 异步记录到数据库
	go func() {
		if err := ct.storage.GetTunnelConnectionRepository().RecordError(ctx, connectionID, errorMessage); err != nil {
			logger.Error("Failed to record error in database", map[string]interface{}{
				"error":        err.Error(),
				"connectionId": connectionID,
			})
		}
	}()

	logger.Warn("Connection error recorded", map[string]interface{}{
		"connectionId": connectionID,
		"error":        errorMessage,
	})

	return nil
}

// Close 关闭连接跟踪器
func (ct *connectionTracker) Close() error {
	ct.cancel()
	ct.wg.Wait()

	// 关闭所有活跃连接
	ct.connMutex.Lock()
	defer ct.connMutex.Unlock()

	for connectionID := range ct.connections {
		if err := ct.CloseConnection(context.Background(), connectionID); err != nil {
			logger.Error("Failed to close connection during shutdown", map[string]interface{}{
				"error":        err.Error(),
				"connectionId": connectionID,
			})
		}
	}

	return nil
}
