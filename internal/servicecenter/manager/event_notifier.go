package manager

import (
	"context"
	"encoding/json"
	"time"

	"gateway/internal/servicecenter/cache"
	"gateway/internal/servicecenter/server"
	pb "gateway/internal/servicecenter/server/proto"
	"gateway/internal/servicecenter/types"
	"gateway/pkg/logger"
)

// EventNotifier 事件通知器
// 负责在缓存更新时自动通知相关实例的订阅者
type EventNotifier struct {
	manager *ServiceCenterManager
}

// NewEventNotifier 创建事件通知器
func NewEventNotifier(manager *ServiceCenterManager) *EventNotifier {
	return &EventNotifier{
		manager: manager,
	}
}

// NotifyServiceChange 通知服务变更（自动查找相关实例）
//
// 处理流程：
//  1. 从缓存获取完整的服务信息（包含节点列表）
//  2. 构建 ServiceChangeEvent
//  3. 查找属于该租户的所有运行中实例
//  4. 异步通知所有相关实例的订阅者
//
// 参数:
//   - tenantId: 租户ID
//   - namespaceId: 命名空间ID
//   - groupName: 服务组名
//   - serviceName: 服务名
//   - eventType: 事件类型（SERVICE_ADDED, SERVICE_UPDATED, SERVICE_DELETED, NODE_ADDED, NODE_UPDATED, NODE_REMOVED）
func (n *EventNotifier) NotifyServiceChange(ctx context.Context, tenantId, namespaceId, groupName, serviceName, eventType string) {
	// 从缓存获取完整的服务信息
	globalCache := cache.GetGlobalCache()
	service, found := globalCache.GetService(ctx, tenantId, namespaceId, groupName, serviceName)

	// 构建 ServiceChangeEvent
	event := &pb.ServiceChangeEvent{
		EventType:   eventType,
		Timestamp:   time.Now().Format("2006-01-02 15:04:05"),
		NamespaceId: namespaceId,
		GroupName:   groupName,
		ServiceName: serviceName,
	}

	// 如果服务存在，添加完整的服务信息和节点列表
	if found && service != nil {
		event.Service = n.serviceToProto(service)
		event.Nodes = make([]*pb.Node, 0, len(service.Nodes))
		for _, node := range service.Nodes {
			event.Nodes = append(event.Nodes, n.nodeToProto(node))
		}
	}

	// 查找并通知相关实例
	n.notifyInstances(ctx, tenantId, func(srv *server.Server) error {
		registryHandler := srv.GetRegistryHandler()
		if registryHandler == nil {
			return nil // 跳过未初始化的实例
		}

		serviceSubMgr := registryHandler.GetServiceSubscriber()
		if serviceSubMgr == nil {
			return nil // 跳过未初始化的实例
		}

		// 触发事件通知
		serviceSubMgr.NotifyServiceChange(tenantId, namespaceId, groupName, serviceName, event)
		return nil
	})

	logger.Debug("服务变更事件已发送",
		"tenantId", tenantId,
		"namespaceId", namespaceId,
		"groupName", groupName,
		"serviceName", serviceName,
		"eventType", eventType)
}

// NotifyConfigChange 通知配置变更（自动查找相关实例）
//
// 处理流程：
//  1. 构建 ConfigChangeEvent
//  2. 查找属于该租户的所有运行中实例
//  3. 异步通知所有相关实例的监听者
//
// 参数:
//   - tenantId: 租户ID
//   - namespaceId: 命名空间ID
//   - groupName: 配置组名
//   - configDataId: 配置ID
//   - eventType: 事件类型（UPDATED, DELETED）
func (n *EventNotifier) NotifyConfigChange(ctx context.Context, tenantId, namespaceId, groupName, configDataId, eventType string) {
	// 构建 ConfigChangeEvent
	event := &pb.ConfigChangeEvent{
		EventType:    eventType,
		Timestamp:    time.Now().Format("2006-01-02 15:04:05"),
		NamespaceId:  namespaceId,
		GroupName:    groupName,
		ConfigDataId: configDataId,
	}

	// 查找并通知相关实例
	n.notifyInstances(ctx, tenantId, func(srv *server.Server) error {
		configHandler := srv.GetConfigHandler()
		if configHandler == nil {
			return nil // 跳过未初始化的实例
		}

		configWatcher := configHandler.GetConfigWatcher()
		if configWatcher == nil {
			return nil // 跳过未初始化的实例
		}

		// 触发事件通知
		configWatcher.NotifyConfigChange(tenantId, namespaceId, groupName, configDataId, event)
		return nil
	})

	logger.Debug("配置变更事件已发送",
		"tenantId", tenantId,
		"namespaceId", namespaceId,
		"groupName", groupName,
		"configDataId", configDataId,
		"eventType", eventType)
}

// notifyInstances 通知属于指定租户的所有运行中实例
func (n *EventNotifier) notifyInstances(ctx context.Context, tenantId string, notifyFn func(*server.Server) error) {
	// 获取所有运行中的实例
	n.manager.mu.RLock()
	instances := make(map[string]*server.Server)
	for name, srv := range n.manager.instances {
		// 只选择正在运行的实例
		if srv != nil && srv.IsRunning() {
			// 检查实例是否属于该租户
			config := srv.GetConfig()
			if config != nil && config.TenantID == tenantId {
				instances[name] = srv
			}
		}
	}
	n.manager.mu.RUnlock()

	if len(instances) == 0 {
		logger.Debug("未找到属于该租户的运行中实例", "tenantId", tenantId)
		return
	}

	// 异步通知所有相关实例
	for instanceName, srv := range instances {
		go func(instName string, s *server.Server) {
			if err := notifyFn(s); err != nil {
				logger.Warn("通知实例失败",
					"instanceName", instName,
					"tenantId", tenantId,
					"error", err)
			}
		}(instanceName, srv)
	}
}

// serviceToProto 将服务对象转换为 Proto 格式
func (n *EventNotifier) serviceToProto(service *types.Service) *pb.Service {
	if service == nil {
		return nil
	}

	metadata := n.parseMetadataJson(service.MetadataJson)

	return &pb.Service{
		NamespaceId:        service.NamespaceId,
		GroupName:          service.GroupName,
		ServiceName:        service.ServiceName,
		ServiceType:        service.ServiceType,
		ServiceDescription: service.ServiceDescription,
		Metadata:           metadata,
	}
}

// nodeToProto 将节点对象转换为 Proto 格式
func (n *EventNotifier) nodeToProto(node *types.ServiceNode) *pb.Node {
	if node == nil {
		return nil
	}

	return &pb.Node{
		NodeId:         node.NodeId,
		NamespaceId:    node.NamespaceId,
		GroupName:      node.GroupName,
		ServiceName:    node.ServiceName,
		IpAddress:      node.IpAddress,
		PortNumber:     int32(node.PortNumber),
		InstanceStatus: node.InstanceStatus,
		HealthyStatus:  node.HealthyStatus,
		Ephemeral:      node.Ephemeral,
		Weight:         node.Weight,
		Metadata:       n.parseMetadataJson(node.MetadataJson),
	}
}

// parseMetadataJson 解析元数据 JSON 字符串为 map
func (n *EventNotifier) parseMetadataJson(metadataJson string) map[string]string {
	if metadataJson == "" {
		return nil
	}

	var metadata map[string]string
	if err := json.Unmarshal([]byte(metadataJson), &metadata); err != nil {
		logger.Warn("解析元数据 JSON 失败", "error", err)
		return nil
	}
	return metadata
}
