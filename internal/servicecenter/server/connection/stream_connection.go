package connection

import (
	"context"
	"sync"
	"time"

	pb "gateway/internal/servicecenter/server/proto"
)

// ServiceSubscription 服务订阅信息
type ServiceSubscription struct {
	NamespaceID  string   // 命名空间ID
	GroupName    string   // 服务组名称
	ServiceNames []string // 订阅的服务名称列表（空表示订阅整个组）
}

// ConfigWatch 配置监听信息
type ConfigWatch struct {
	NamespaceID   string   // 命名空间ID
	GroupName     string   // 配置组名称
	ConfigDataIDs []string // 监听的配置ID列表
}

// StreamConnection 双向流连接上下文
// 表示一个客户端与服务端之间的长连接
type StreamConnection struct {
	// ========== 连接标识 ==========
	ConnectionID string // 连接ID（服务端生成，UUID）
	ClientID     string // 客户端ID（客户端生成，UUID）
	ClientIP     string // 客户端IP地址（从 gRPC 上下文获取）
	TenantID     string // 租户ID（从认证信息中获取）
	NamespaceID  string // 默认命名空间ID

	// ========== 客户端信息 ==========
	Metadata       *pb.ClientMetadata // 客户端元数据（版本、语言等）
	SubscribeTypes []string           // 订阅类型（["registry", "config"]）

	// ========== gRPC 流 ==========
	Stream  pb.ServiceCenterStream_ConnectServer // gRPC 双向流
	Context context.Context                      // 连接上下文
	Cancel  context.CancelFunc                   // 取消函数

	// ========== 连接状态 ==========
	RegisteredNodes []string  // 已注册的节点ID列表
	LastPingTime    time.Time // 最后 Ping 时间
	LastActiveTime  time.Time // 最后活跃时间

	// ========== 订阅管理 ==========
	ServiceSubscriptions []ServiceSubscription // 服务订阅列表
	ConfigWatches        []ConfigWatch         // 配置监听列表

	// ========== 并发控制 ==========
	mu sync.RWMutex // 保护连接状态的并发访问
}

// ========== 节点管理 ==========

// AddRegisteredNode 添加已注册节点
func (c *StreamConnection) AddRegisteredNode(nodeId string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.RegisteredNodes = append(c.RegisteredNodes, nodeId)
}

// RemoveRegisteredNode 移除已注册节点
func (c *StreamConnection) RemoveRegisteredNode(nodeId string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, id := range c.RegisteredNodes {
		if id == nodeId {
			c.RegisteredNodes = append(c.RegisteredNodes[:i], c.RegisteredNodes[i+1:]...)
			break
		}
	}
}

// GetRegisteredNodes 获取已注册节点列表（副本）
func (c *StreamConnection) GetRegisteredNodes() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	nodes := make([]string, len(c.RegisteredNodes))
	copy(nodes, c.RegisteredNodes)
	return nodes
}

// HasRegisteredNodes 检查是否有已注册节点
func (c *StreamConnection) HasRegisteredNodes() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.RegisteredNodes) > 0
}

// ========== 服务订阅管理 ==========

// AddServiceSubscription 添加服务订阅
func (c *StreamConnection) AddServiceSubscription(namespaceId, groupName string, serviceNames []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ServiceSubscriptions = append(c.ServiceSubscriptions, ServiceSubscription{
		NamespaceID:  namespaceId,
		GroupName:    groupName,
		ServiceNames: serviceNames,
	})
}

// IsSubscribedToService 检查是否订阅了指定服务
func (c *StreamConnection) IsSubscribedToService(namespaceId, groupName, serviceName string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, sub := range c.ServiceSubscriptions {
		// 检查命名空间和组名
		if sub.NamespaceID != namespaceId || sub.GroupName != groupName {
			continue
		}

		// 如果服务名列表为空，表示订阅整个组
		if len(sub.ServiceNames) == 0 {
			return true
		}

		// 检查是否订阅了指定服务
		for _, name := range sub.ServiceNames {
			if name == serviceName {
				return true
			}
		}
	}

	return false
}

// GetServiceSubscriptions 获取服务订阅列表（副本）
func (c *StreamConnection) GetServiceSubscriptions() []ServiceSubscription {
	c.mu.RLock()
	defer c.mu.RUnlock()
	subs := make([]ServiceSubscription, len(c.ServiceSubscriptions))
	copy(subs, c.ServiceSubscriptions)
	return subs
}

// ========== 配置监听管理 ==========

// AddConfigWatch 添加配置监听
func (c *StreamConnection) AddConfigWatch(namespaceId, groupName string, configDataIds []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ConfigWatches = append(c.ConfigWatches, ConfigWatch{
		NamespaceID:   namespaceId,
		GroupName:     groupName,
		ConfigDataIDs: configDataIds,
	})
}

// IsWatchingConfig 检查是否监听了指定配置
func (c *StreamConnection) IsWatchingConfig(namespaceId, groupName, configDataId string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, watch := range c.ConfigWatches {
		// 检查命名空间和组名
		if watch.NamespaceID != namespaceId || watch.GroupName != groupName {
			continue
		}

		// 检查是否监听了指定配置
		for _, id := range watch.ConfigDataIDs {
			if id == configDataId {
				return true
			}
		}
	}

	return false
}

// GetConfigWatches 获取配置监听列表（副本）
func (c *StreamConnection) GetConfigWatches() []ConfigWatch {
	c.mu.RLock()
	defer c.mu.RUnlock()
	watches := make([]ConfigWatch, len(c.ConfigWatches))
	copy(watches, c.ConfigWatches)
	return watches
}

// ========== 时间管理 ==========

// UpdateLastPingTime 更新最后 Ping 时间
func (c *StreamConnection) UpdateLastPingTime() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.LastPingTime = time.Now()
}

// GetLastPingTime 获取最后 Ping 时间
func (c *StreamConnection) GetLastPingTime() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.LastPingTime
}

// UpdateLastActiveTime 更新最后活跃时间
func (c *StreamConnection) UpdateLastActiveTime() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.LastActiveTime = time.Now()
}

// GetLastActiveTime 获取最后活跃时间
func (c *StreamConnection) GetLastActiveTime() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.LastActiveTime
}

// ========== 连接管理 ==========

// IsActive 检查连接是否活跃
// 如果最后活跃时间超过指定的超时时间，则认为不活跃
func (c *StreamConnection) IsActive(timeout time.Duration) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return time.Since(c.LastActiveTime) <= timeout
}

// IsPingTimeout 检查 Ping 是否超时
func (c *StreamConnection) IsPingTimeout(timeout time.Duration) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return time.Since(c.LastPingTime) > timeout
}

// Close 关闭连接
func (c *StreamConnection) Close() {
	if c.Cancel != nil {
		c.Cancel()
	}
}

// ========== 消息发送 ==========

// Send 发送消息到客户端（线程安全）
func (c *StreamConnection) Send(msg *pb.ServerMessage) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Stream.Send(msg)
}

// ========== 辅助方法 ==========

// GetConnectionInfo 获取连接信息（用于日志）
func (c *StreamConnection) GetConnectionInfo() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return map[string]interface{}{
		"connectionId":         c.ConnectionID,
		"clientId":             c.ClientID,
		"clientIP":             c.ClientIP,
		"tenantId":             c.TenantID,
		"namespaceId":          c.NamespaceID,
		"registeredNodes":      len(c.RegisteredNodes),
		"serviceSubscriptions": len(c.ServiceSubscriptions),
		"configWatches":        len(c.ConfigWatches),
		"lastPingTime":         c.LastPingTime.Format(time.RFC3339),
		"lastActiveTime":       c.LastActiveTime.Format(time.RFC3339),
		"language":             c.GetLanguage(),
		"clientVersion":        c.GetClientVersion(),
	}
}

// GetLanguage 获取客户端语言
func (c *StreamConnection) GetLanguage() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.Metadata != nil {
		return c.Metadata.GetLanguage()
	}
	return ""
}

// GetClientVersion 获取客户端版本
func (c *StreamConnection) GetClientVersion() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.Metadata != nil {
		return c.Metadata.GetClientVersion()
	}
	return ""
}
