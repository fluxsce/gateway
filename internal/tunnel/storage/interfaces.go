// Package storage 定义隧道管理系统的数据存储接口
// 遵循Repository模式，抽象数据访问层
package storage

import (
	"context"
	"time"

	"gateway/internal/tunnel/types"
)

// TunnelServerRepository 隧道服务器存储接口
type TunnelServerRepository interface {
	// Create 创建隧道服务器配置
	Create(ctx context.Context, server *types.TunnelServer) error

	// GetByID 根据ID获取隧道服务器配置
	GetByID(ctx context.Context, serverID string) (*types.TunnelServer, error)

	// GetByTenantID 根据租户ID获取隧道服务器列表
	GetByTenantID(ctx context.Context, tenantID string) ([]*types.TunnelServer, error)

	// Update 更新隧道服务器配置
	Update(ctx context.Context, server *types.TunnelServer) error

	// Delete 删除隧道服务器配置
	Delete(ctx context.Context, serverID string) error

	// UpdateStatus 更新服务器状态
	UpdateStatus(ctx context.Context, serverID string, status string, startTime *time.Time) error
}

// TunnelServerNodeRepository 隧道服务器节点存储接口
type TunnelServerNodeRepository interface {
	// Create 创建服务器节点（静态端口映射）
	Create(ctx context.Context, node *types.TunnelServerNode) error

	// GetByID 根据ID获取服务器节点
	GetByID(ctx context.Context, nodeID string) (*types.TunnelServerNode, error)

	// GetByServerID 根据服务器ID获取节点列表
	GetByServerID(ctx context.Context, serverID string) ([]*types.TunnelServerNode, error)

	// GetActiveNodes 获取活跃的服务器节点
	GetActiveNodes(ctx context.Context, serverID string) ([]*types.TunnelServerNode, error)

	// GetByPortAndType 根据端口和类型查找节点（检查端口冲突）
	GetByPortAndType(ctx context.Context, listenAddress string, listenPort int, proxyType string) (*types.TunnelServerNode, error)

	// Update 更新服务器节点
	Update(ctx context.Context, node *types.TunnelServerNode) error

	// Delete 删除服务器节点
	Delete(ctx context.Context, nodeID string) error

	// UpdateConnectionCount 更新连接计数
	UpdateConnectionCount(ctx context.Context, nodeID string, count int) error

	// UpdateHealthCheck 更新健康检查状态
	UpdateHealthCheck(ctx context.Context, nodeID string, lastCheck time.Time, status string) error
}

// TunnelClientRepository 隧道客户端存储接口
type TunnelClientRepository interface {
	// Create 创建客户端注册
	Create(ctx context.Context, client *types.TunnelClient) error

	// GetByID 根据ID获取客户端
	GetByID(ctx context.Context, clientID string) (*types.TunnelClient, error)

	// GetByName 根据名称获取客户端
	GetByName(ctx context.Context, clientName string) (*types.TunnelClient, error)

	// GetByTenantID 根据租户ID获取客户端列表
	GetByTenantID(ctx context.Context, tenantID string) ([]*types.TunnelClient, error)

	// GetActiveClients 获取活跃连接的客户端
	GetActiveClients(ctx context.Context, tenantID string) ([]*types.TunnelClient, error)

	// Update 更新客户端信息
	Update(ctx context.Context, client *types.TunnelClient) error

	// Delete 删除客户端
	Delete(ctx context.Context, clientID string) error

	// UpdateConnectionStatus 更新连接状态
	UpdateConnectionStatus(ctx context.Context, clientID string, status string, connectTime *time.Time) error

	// UpdateHeartbeat 更新心跳时间
	UpdateHeartbeat(ctx context.Context, clientID string, heartbeatTime time.Time) error

	// UpdateReconnectInfo 更新重连信息
	UpdateReconnectInfo(ctx context.Context, clientID string, reconnectCount int, totalConnectTime int64) error
}

// TunnelServiceRepository 隧道服务存储接口
type TunnelServiceRepository interface {
	// Create 创建服务注册
	Create(ctx context.Context, service *types.TunnelService) error

	// GetByID 根据ID获取服务
	GetByID(ctx context.Context, serviceID string) (*types.TunnelService, error)

	// GetByClientID 根据客户端ID获取服务列表
	GetByClientID(ctx context.Context, clientID string) ([]*types.TunnelService, error)

	// GetByName 根据服务名称获取服务
	GetByName(ctx context.Context, serviceName string) (*types.TunnelService, error)

	// GetActiveServices 获取活跃的服务
	GetActiveServices(ctx context.Context, clientID string) ([]*types.TunnelService, error)

	// GetByRemotePort 根据远程端口获取服务（检查端口冲突）
	GetByRemotePort(ctx context.Context, remotePort int) (*types.TunnelService, error)

	// Update 更新服务配置
	Update(ctx context.Context, service *types.TunnelService) error

	// Delete 删除服务
	Delete(ctx context.Context, serviceID string) error

	// UpdateStatus 更新服务状态
	UpdateStatus(ctx context.Context, serviceID string, status string, lastActiveTime *time.Time) error

	// UpdateConnectionCount 更新连接计数
	UpdateConnectionCount(ctx context.Context, serviceID string, count int, totalConnections int64, totalTraffic int64) error

	// AssignRemotePort 分配远程端口
	AssignRemotePort(ctx context.Context, serviceID string, remotePort int) error
}

// TunnelSessionRepository 隧道会话存储接口
type TunnelSessionRepository interface {
	// Create 创建会话
	Create(ctx context.Context, session *types.TunnelSession) error

	// GetByID 根据ID获取会话
	GetByID(ctx context.Context, sessionID string) (*types.TunnelSession, error)

	// GetByToken 根据令牌获取会话
	GetByToken(ctx context.Context, sessionToken string) (*types.TunnelSession, error)

	// GetByClientID 根据客户端ID获取会话列表
	GetByClientID(ctx context.Context, clientID string) ([]*types.TunnelSession, error)

	// GetActiveSessions 获取活跃会话
	GetActiveSessions(ctx context.Context, clientID string) ([]*types.TunnelSession, error)

	// Update 更新会话信息
	Update(ctx context.Context, session *types.TunnelSession) error

	// Delete 删除会话
	Delete(ctx context.Context, sessionID string) error

	// UpdateHeartbeat 更新心跳信息
	UpdateHeartbeat(ctx context.Context, sessionID string, heartbeatTime time.Time, heartbeatCount int) error

	// UpdateActivity 更新活动时间
	UpdateActivity(ctx context.Context, sessionID string, activityTime time.Time) error

	// UpdateProxyCount 更新代理连接数量
	UpdateProxyCount(ctx context.Context, sessionID string, proxyCount int) error

	// CloseSession 关闭会话
	CloseSession(ctx context.Context, sessionID string, endTime time.Time, duration int64) error
}

// TunnelConnectionRepository 隧道连接存储接口
type TunnelConnectionRepository interface {
	// Create 创建连接记录
	Create(ctx context.Context, connection *types.TunnelConnection) error

	// GetByID 根据ID获取连接
	GetByID(ctx context.Context, connectionID string) (*types.TunnelConnection, error)

	// GetBySessionID 根据会话ID获取连接列表
	GetBySessionID(ctx context.Context, sessionID string) ([]*types.TunnelConnection, error)

	// GetActiveConnections 获取活跃连接
	GetActiveConnections(ctx context.Context, sessionID string) ([]*types.TunnelConnection, error)

	// GetConnectionsByDateRange 根据时间范围获取连接
	GetConnectionsByDateRange(ctx context.Context, startTime, endTime time.Time) ([]*types.TunnelConnection, error)

	// Update 更新连接信息
	Update(ctx context.Context, connection *types.TunnelConnection) error

	// Delete 删除连接记录
	Delete(ctx context.Context, connectionID string) error

	// UpdateTrafficStats 更新流量统计
	UpdateTrafficStats(ctx context.Context, connectionID string, bytesReceived, bytesSent int64, packetsReceived, packetsSent int64) error

	// UpdateActivity 更新活动时间
	UpdateActivity(ctx context.Context, connectionID string, activityTime time.Time) error

	// CloseConnection 关闭连接
	CloseConnection(ctx context.Context, connectionID string, endTime time.Time, duration int64) error

	// RecordError 记录错误信息
	RecordError(ctx context.Context, connectionID string, errorMessage string) error

	// GetTrafficStats 获取流量统计
	GetTrafficStats(ctx context.Context, startTime, endTime time.Time, groupBy string) ([]*TrafficStats, error)
}

// TrafficStats 流量统计结果
type TrafficStats struct {
	GroupKey           string    `json:"groupKey" db:"groupKey"`
	ConnectionCount    int       `json:"connectionCount" db:"connectionCount"`
	TotalBytesReceived int64     `json:"totalBytesReceived" db:"totalBytesReceived"`
	TotalBytesSent     int64     `json:"totalBytesSent" db:"totalBytesSent"`
	AverageLatency     float64   `json:"averageLatency" db:"averageLatency"`
	Date               time.Time `json:"date" db:"date"`
}

// RepositoryManager 存储管理器接口
type RepositoryManager interface {
	// GetTunnelServerRepository 获取隧道服务器存储接口
	GetTunnelServerRepository() TunnelServerRepository

	// GetTunnelServerNodeRepository 获取隧道服务器节点存储接口
	GetTunnelServerNodeRepository() TunnelServerNodeRepository

	// GetTunnelClientRepository 获取隧道客户端存储接口
	GetTunnelClientRepository() TunnelClientRepository

	// GetTunnelServiceRepository 获取隧道服务存储接口
	GetTunnelServiceRepository() TunnelServiceRepository

	// GetTunnelSessionRepository 获取隧道会话存储接口
	GetTunnelSessionRepository() TunnelSessionRepository

	// GetTunnelConnectionRepository 获取隧道连接存储接口
	GetTunnelConnectionRepository() TunnelConnectionRepository

	// Close 关闭存储管理器
	Close() error
}
