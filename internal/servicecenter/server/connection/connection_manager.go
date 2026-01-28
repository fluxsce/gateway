package connection

import (
	"context"
	"sync"
	"time"

	pb "gateway/internal/servicecenter/server/proto"
	"gateway/pkg/logger"
	"gateway/pkg/utils/random"
)

// ConnectionManager 连接管理器
// 负责管理所有的双向流连接
type ConnectionManager struct {
	connections sync.Map // connectionId -> *StreamConnection
	mu          sync.RWMutex
}

// NewConnectionManager 创建连接管理器
func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{}
}

// ========== 连接管理 ==========

// CreateConnection 创建新连接
func (m *ConnectionManager) CreateConnection(
	ctx context.Context,
	stream pb.ServiceCenterStream_ConnectServer,
	clientIP string) *StreamConnection {

	connCtx, cancel := context.WithCancel(ctx)

	conn := &StreamConnection{
		ConnectionID:         random.GenerateUniqueStringWithPrefix("conn-", 32),
		ClientIP:             clientIP,
		Stream:               stream,
		Context:              connCtx,
		Cancel:               cancel,
		RegisteredNodes:      []string{},
		ServiceSubscriptions: []ServiceSubscription{},
		ConfigWatches:        []ConfigWatch{},
		LastPingTime:         time.Now(),
		LastActiveTime:       time.Now(),
	}

	m.connections.Store(conn.ConnectionID, conn)

	logger.Debug("创建新连接",
		"connectionId", conn.ConnectionID,
		"clientIP", clientIP)

	return conn
}

// GetConnection 获取连接
func (m *ConnectionManager) GetConnection(connectionId string) (*StreamConnection, bool) {
	if conn, ok := m.connections.Load(connectionId); ok {
		return conn.(*StreamConnection), true
	}
	return nil, false
}

// RemoveConnection 移除连接
func (m *ConnectionManager) RemoveConnection(connectionId string) {
	if conn, ok := m.connections.LoadAndDelete(connectionId); ok {
		c := conn.(*StreamConnection)
		c.Close() // 取消上下文
		logger.Debug("移除连接", "connectionId", connectionId)
	}
}

// GetAllConnections 获取所有连接（副本）
func (m *ConnectionManager) GetAllConnections() []*StreamConnection {
	var connections []*StreamConnection
	m.connections.Range(func(key, value interface{}) bool {
		connections = append(connections, value.(*StreamConnection))
		return true
	})
	return connections
}

// ========== 连接查询 ==========

// GetConnectionCount 获取连接数量
func (m *ConnectionManager) GetConnectionCount() int {
	count := 0
	m.connections.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

// GetConnectionsByClient 获取指定客户端的所有连接
func (m *ConnectionManager) GetConnectionsByClient(clientId string) []*StreamConnection {
	var connections []*StreamConnection
	m.connections.Range(func(key, value interface{}) bool {
		conn := value.(*StreamConnection)
		if conn.ClientID == clientId {
			connections = append(connections, conn)
		}
		return true
	})
	return connections
}

// ========== 消息广播 ==========

// BroadcastToAll 广播给所有连接
func (m *ConnectionManager) BroadcastToAll(message *pb.ServerMessage) int {
	count := 0
	m.connections.Range(func(key, value interface{}) bool {
		conn := value.(*StreamConnection)
		if err := conn.Send(message); err != nil {
			logger.Error("广播消息失败", err,
				"connectionId", conn.ConnectionID)
		} else {
			count++
		}
		return true
	})
	return count
}

// BroadcastToSubscribers 广播给订阅者（带过滤器）
func (m *ConnectionManager) BroadcastToSubscribers(
	filter func(*StreamConnection) bool,
	message *pb.ServerMessage) int {

	count := 0
	m.connections.Range(func(key, value interface{}) bool {
		conn := value.(*StreamConnection)

		// 应用过滤器
		if filter != nil && !filter(conn) {
			return true
		}

		// 发送消息
		if err := conn.Send(message); err != nil {
			logger.Error("推送消息失败", err,
				"connectionId", conn.ConnectionID)
		} else {
			count++
		}

		return true
	})
	return count
}

// ========== 连接清理 ==========

// CleanupTimeoutConnections 清理心跳超时的连接
// timeout: 心跳超时时间
// 返回清理的连接数
func (m *ConnectionManager) CleanupTimeoutConnections(timeout time.Duration) int {
	var toRemove []*StreamConnection

	m.connections.Range(func(key, value interface{}) bool {
		conn := value.(*StreamConnection)
		if conn.IsPingTimeout(timeout) {
			toRemove = append(toRemove, conn)
		}
		return true
	})

	for _, conn := range toRemove {
		logger.Warn("连接心跳超时",
			"connectionId", conn.ConnectionID,
			"lastPingTime", conn.GetLastPingTime())

		// 发送关闭通知
		closeNotification := &pb.ServerMessage{
			MessageType: pb.ServerMessageType_SERVER_CLOSE,
			Message: &pb.ServerMessage_Close{
				Close: &pb.ServerCloseNotification{
					Reason:      "client_timeout",
					Message:     "客户端心跳超时",
					GracePeriod: 5,
				},
			},
		}
		conn.Send(closeNotification)

		// 移除连接
		m.RemoveConnection(conn.ConnectionID)
	}

	if len(toRemove) > 0 {
		logger.Info("清理心跳超时连接完成", "count", len(toRemove))
	}

	return len(toRemove)
}

// ========== 统计信息 ==========

// GetStats 获取连接统计信息
func (m *ConnectionManager) GetStats() map[string]interface{} {
	totalCount := 0
	languageCounts := make(map[string]int)
	activeCount := 0

	m.connections.Range(func(key, value interface{}) bool {
		conn := value.(*StreamConnection)
		totalCount++

		// 统计语言
		if lang := conn.GetLanguage(); lang != "" {
			languageCounts[lang]++
		}

		// 统计活跃连接（5分钟内）
		if conn.IsActive(5 * time.Minute) {
			activeCount++
		}

		return true
	})

	return map[string]interface{}{
		"totalConnections":  totalCount,
		"activeConnections": activeCount,
		"languageCounts":    languageCounts,
	}
}

// ========== 生命周期管理 ==========

// Close 关闭连接管理器
// 发送关闭通知并清理所有连接
func (m *ConnectionManager) Close() {
	logger.Info("正在关闭连接管理器...")

	// 发送关闭通知给所有客户端
	closeNotification := &pb.ServerMessage{
		MessageType: pb.ServerMessageType_SERVER_CLOSE,
		Message: &pb.ServerMessage_Close{
			Close: &pb.ServerCloseNotification{
				Reason:      "server_shutdown",
				Message:     "服务端正在关闭",
				GracePeriod: 30,
			},
		},
	}

	count := m.BroadcastToAll(closeNotification)
	logger.Info("发送关闭通知完成", "count", count)

	// 等待短暂时间让客户端处理关闭通知
	time.Sleep(1 * time.Second)

	// 关闭所有连接
	var connIds []string
	m.connections.Range(func(key, value interface{}) bool {
		connIds = append(connIds, key.(string))
		return true
	})

	for _, connId := range connIds {
		m.RemoveConnection(connId)
	}

	logger.Info("连接管理器已关闭", "closedConnections", len(connIds))
}
