package subscriber

import (
	"context"
	"encoding/json"
	"sync"

	pb "gateway/internal/servicecenter/server/proto"
	"gateway/internal/servicecenter/types"
	"gateway/pkg/logger"
)

// ServiceSubscriber 服务订阅管理器
//
// 功能说明：
//
//	管理所有订阅服务变更的客户端连接，支持两种订阅模式：
//	1. 批量服务订阅：一个客户端订阅多个服务的变更（减少连接数）
//	   单个服务订阅可以看作是批量订阅的特例（serviceNames 数组只有一个元素）
//	2. 命名空间订阅：订阅整个命名空间/分组下的所有服务变更
//
// 事件发送机制详解：
//
//  1. 客户端订阅阶段（SubscribeServices gRPC 方法）：
//     a. 客户端建立 gRPC Server-Side Streaming 连接
//     b. Handler 为每个客户端生成唯一的 subscriberID（如 "SUB_xxx"）
//     c. Handler 调用 SubscribeMultipleServices() 创建共享 channel，并注册到映射中：
//     batchSubscribers[subscriberID][serviceKey] = ch
//     d. Handler 启动 goroutine，持续从 channel 读取事件，通过 gRPC stream 发送给客户端
//
//  2. 事件产生阶段（服务注册/注销/心跳等操作）：
//     a. Handler 执行服务变更操作（如 RegisterNode、UnregisterNode）
//     b. Handler 构建 ServiceChangeEvent 事件对象
//     c. Handler 调用 NotifyServiceChange() 通知订阅管理器
//
//  3. 事件分发阶段（NotifyServiceChange 方法）：
//     a. 根据服务标识（serviceKey）查找所有订阅了该服务的客户端
//     b. 遍历订阅者映射，找到所有相关的 channel：
//     - 批量订阅：batchSubscribers[subscriberID][serviceKey] -> channel
//     - 命名空间订阅：namespaceSubscribers[namespaceKey] -> 所有 subscriberID 的 channel
//     c. 对每个 channel 调用 sendEvent() 非阻塞发送事件
//
//  4. 事件推送阶段（Handler 的 goroutine）：
//     a. Handler 的 goroutine 从 channel 读取事件（阻塞等待）
//     b. 通过 gRPC stream.Send() 将事件推送给对应的客户端
//     c. 如果客户端断开连接，channel 被关闭，goroutine 退出
//
// 如何知道发送给哪个客户端？
//
//   - 每个客户端订阅时，会创建一个共享的 channel，存储在映射中：
//     batchSubscribers[subscriberID][serviceKey] = ch
//   - 当事件发生时，通过 serviceKey 查找所有订阅者，遍历所有 channel
//   - 每个 channel 对应一个客户端的 gRPC stream（在 Handler 的 goroutine 中）
//   - 事件通过 channel 发送到对应的 Handler goroutine，再通过 stream 发送给客户端
//
// 数据结构说明：
//
//	batchSubscribers: map[subscriberID]map[serviceKey]channel
//	  - 第一层 key: subscriberID = 客户端唯一标识（如 "SUB_xxx"）
//	  - 第二层 key: serviceKey = "tenantId:namespaceId:groupName:serviceName"
//	  - value: channel = 该客户端订阅的所有服务共用的共享通道
//
//	namespaceSubscribers: map[namespaceKey]map[subscriberID]channel
//	  - 第一层 key: namespaceKey = "tenantId:namespaceId:groupName"
//	  - 第二层 key: subscriberID = 客户端唯一标识（如 "SUB_xxx"）
//	  - value: channel = 该客户端专属的事件通道
//
// 并发安全：
//
//	使用 RWMutex 保护订阅者映射的并发访问
//	事件发送使用带缓冲的 channel（100 容量），避免阻塞
//
// 资源管理：
//
//	客户端断开连接时，Handler 调用 UnsubscribeMultipleServices 清理资源
//	订阅管理器关闭 channel 并删除订阅记录
type ServiceSubscriber struct {
	mu sync.RWMutex
	// 批量订阅：一个 subscriberID 可以订阅多个服务，所有服务共用同一个 channel
	batchSubscribers map[string]map[string]chan *pb.ServiceChangeEvent // key: subscriberID -> serviceKey -> channel (同一个 channel)

	// 命名空间订阅：订阅整个命名空间/分组
	namespaceSubscribers map[string]map[string]chan *pb.ServiceChangeEvent // key: namespaceKey -> subscriberID -> channel
}

// NewServiceSubscriber 创建服务订阅管理器
func NewServiceSubscriber() *ServiceSubscriber {
	return &ServiceSubscriber{
		batchSubscribers:     make(map[string]map[string]chan *pb.ServiceChangeEvent),
		namespaceSubscribers: make(map[string]map[string]chan *pb.ServiceChangeEvent),
	}
}

// SubscribeMultipleServices 批量订阅多个服务（使用同一个 channel）
//
// 处理流程：
//  1. 创建一个共享的 channel（所有服务共用）
//  2. 为每个服务注册订阅，都使用同一个 channel
//  3. 注册到 batchSubscribers[subscriberID][serviceKey] = ch
//  4. 返回共享 channel 供 Handler 读取
//
// 数据结构：
//
//	batchSubscribers[subscriberID][serviceKey] = channel
//	例如：
//	  batchSubscribers["SUB_abc123"]["default:public:DEFAULT:order-service"] = ch
//	  batchSubscribers["SUB_abc123"]["default:public:DEFAULT:user-service"] = ch
//	  batchSubscribers["SUB_abc123"]["default:public:DEFAULT:pay-service"] = ch
//	表示：同一个客户端订阅了多个服务，所有服务共用同一个 channel
//
// 参数：
//   - ctx: 上下文（用于取消订阅）
//   - tenantId, namespaceId, groupName: 服务标识
//   - serviceNames: 要订阅的服务名称列表
//   - subscriberID: 订阅者唯一标识（由 Handler 生成）
//
// 返回：
//   - channel: 用于接收所有订阅服务变更事件的共享通道
func (s *ServiceSubscriber) SubscribeMultipleServices(
	ctx context.Context,
	tenantId, namespaceId, groupName string,
	serviceNames []string,
	subscriberID string,
) <-chan *pb.ServiceChangeEvent {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 创建共享通道（所有服务共用）
	ch := make(chan *pb.ServiceChangeEvent, 100)

	// 初始化批量订阅记录
	if s.batchSubscribers[subscriberID] == nil {
		s.batchSubscribers[subscriberID] = make(map[string]chan *pb.ServiceChangeEvent)
	}

	// 为每个服务注册订阅（使用同一个 channel）
	serviceKeys := make([]string, 0, len(serviceNames))
	for _, serviceName := range serviceNames {
		serviceKey := s.makeServiceKey(tenantId, namespaceId, groupName, serviceName)
		s.batchSubscribers[subscriberID][serviceKey] = ch
		serviceKeys = append(serviceKeys, serviceKey)
	}

	logger.Info("注册批量服务订阅",
		"subscriberID", subscriberID,
		"tenantId", tenantId,
		"namespaceId", namespaceId,
		"groupName", groupName,
		"serviceNames", serviceNames,
		"serviceKeys", serviceKeys)

	return ch
}

// UnsubscribeMultipleServices 取消批量订阅
//
// 处理流程：
//  1. 查找批量订阅记录
//  2. 关闭共享 channel（所有服务共用）
//  3. 删除订阅记录
func (s *ServiceSubscriber) UnsubscribeMultipleServices(subscriberID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if services, ok := s.batchSubscribers[subscriberID]; ok {
		// 获取通道并关闭（所有服务共用同一个通道）
		var ch chan *pb.ServiceChangeEvent
		for _, c := range services {
			ch = c
			break // 只需要关闭一次
		}
		if ch != nil {
			close(ch)
		}

		// 删除订阅记录
		delete(s.batchSubscribers, subscriberID)
	}
}

// SendToSubscriber 向特定订阅者的 channel 发送事件（用于初始推送）
//
// 处理流程：
//  1. 查找订阅者的 channel
//  2. 非阻塞发送事件到该 channel
//
// 用途：
//   - 用于订阅成功后的初始服务信息推送
//   - 只发送给当前订阅者，不影响其他订阅者
func (s *ServiceSubscriber) SendToSubscriber(subscriberID string, event *pb.ServiceChangeEvent) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if services, ok := s.batchSubscribers[subscriberID]; ok {
		// 获取共享 channel（所有服务共用同一个 channel）
		var ch chan *pb.ServiceChangeEvent
		for _, c := range services {
			ch = c
			break // 只需要获取一次
		}
		if ch != nil {
			// 非阻塞发送
			select {
			case ch <- event:
				// 发送成功
			default:
				// 通道已满，丢弃事件（避免阻塞）
			}
		}
	}
}

// SubscribeNamespace 订阅整个命名空间/分组下的所有服务
//
// 处理流程：
//  1. 生成命名空间唯一键
//  2. 创建订阅通道
//  3. 注册到 namespaceSubscribers[namespaceKey][subscriberID] = ch
//  4. 返回 channel 供 Handler 读取
//
// 数据结构：
//
//	namespaceSubscribers[namespaceKey][subscriberID] = channel
//	例如：
//	  namespaceSubscribers["default:public:DEFAULT"]["SUB_abc123"] = ch
//	表示：客户端订阅了整个命名空间，会收到该命名空间下所有服务的变更事件
//
// 参数：
//   - ctx: 上下文（用于取消订阅）
//   - tenantId, namespaceId, groupName: 命名空间标识
//   - subscriberID: 订阅者唯一标识（由 Handler 生成）
//
// 返回：
//   - channel: 用于接收命名空间下所有服务变更事件的通道
func (s *ServiceSubscriber) SubscribeNamespace(
	ctx context.Context,
	tenantId, namespaceId, groupName string,
	subscriberID string,
) <-chan *pb.ServiceChangeEvent {
	namespaceKey := s.makeNamespaceKey(tenantId, namespaceId, groupName)

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.namespaceSubscribers[namespaceKey] == nil {
		s.namespaceSubscribers[namespaceKey] = make(map[string]chan *pb.ServiceChangeEvent)
	}

	// 创建订阅通道
	ch := make(chan *pb.ServiceChangeEvent, 100)
	s.namespaceSubscribers[namespaceKey][subscriberID] = ch

	return ch
}

// UnsubscribeNamespace 取消命名空间订阅
//
// 处理流程：
//  1. 生成命名空间唯一键
//  2. 查找订阅者对应的 channel
//  3. 关闭 channel
//  4. 从订阅者映射中删除
//  5. 如果该命名空间没有订阅者了，删除整个命名空间的订阅记录
func (s *ServiceSubscriber) UnsubscribeNamespace(tenantId, namespaceId, groupName, subscriberID string) {
	namespaceKey := s.makeNamespaceKey(tenantId, namespaceId, groupName)

	s.mu.Lock()
	defer s.mu.Unlock()

	if subs, ok := s.namespaceSubscribers[namespaceKey]; ok {
		if ch, exists := subs[subscriberID]; exists {
			close(ch)
			delete(subs, subscriberID)
		}

		// 如果没有订阅者了，删除整个命名空间的订阅记录
		if len(subs) == 0 {
			delete(s.namespaceSubscribers, namespaceKey)
		}
	}
}

// NotifyServiceChange 通知服务变更
//
// 处理流程：
//  1. 设置事件的服务标识（namespaceId, groupName, serviceName）
//  2. 通知批量订阅者（检查是否订阅了该服务）
//  3. 通知命名空间订阅者（匹配命名空间/分组）
//
// 如何确定发送给哪些客户端？
//
//	步骤 1：生成服务唯一键
//	  serviceKey = "tenantId:namespaceId:groupName:serviceName"
//	  例如："default:public:DEFAULT:user-service"
//
//	步骤 2：查找批量订阅者
//	  遍历 batchSubscribers，检查每个 subscriberID 是否订阅了该 serviceKey
//	  如果订阅了，通过对应的 channel 发送事件
//
//	步骤 3：查找命名空间订阅者
//	  生成 namespaceKey = "tenantId:namespaceId:groupName"
//	  从 namespaceSubscribers[namespaceKey] 获取所有订阅了该命名空间的客户端
//	  遍历 map[subscriberID]channel，发送事件
//
// 通知范围：
//   - 批量订阅：检查 subscriberID 是否订阅了该服务
//   - 命名空间订阅：匹配 tenantId:namespaceId:groupName
//
// 事件传递链路：
//
//	NotifyServiceChange() -> sendEvent() -> channel -> Handler goroutine -> gRPC stream -> 客户端
//
// 并发安全：
//
//	使用 RLock 读取订阅者映射（允许多个 goroutine 并发读取）
//	事件发送使用非阻塞方式，避免长时间持有锁
func (s *ServiceSubscriber) NotifyServiceChange(tenantId, namespaceId, groupName, serviceName string, event *pb.ServiceChangeEvent) {
	// 生成服务唯一键，用于查找订阅者
	serviceKey := s.makeServiceKey(tenantId, namespaceId, groupName, serviceName)
	namespaceKey := s.makeNamespaceKey(tenantId, namespaceId, groupName)

	// 设置事件的服务标识（用于批量订阅时区分）
	// 注意：proto 中移除了 tenantId 字段
	event.NamespaceId = namespaceId
	event.GroupName = groupName
	event.ServiceName = serviceName

	s.mu.RLock()
	defer s.mu.RUnlock()

	logger.Debug("通知服务变更",
		"serviceKey", serviceKey,
		"eventType", event.EventType,
		"batchSubscribersCount", len(s.batchSubscribers))

	// 1. 通知批量订阅者（检查是否订阅了该服务）
	//    遍历所有批量订阅者，检查是否订阅了该 serviceKey
	notifyCount := 0
	for subscriberID, services := range s.batchSubscribers {
		logger.Debug("检查批量订阅者",
			"subscriberID", subscriberID,
			"subscribedServicesCount", len(services))

		if ch, ok := services[serviceKey]; ok {
			// 该批量订阅者订阅了此服务，发送事件
			s.sendEvent(ch, event)
			notifyCount++
			logger.Debug("已通知批量订阅者",
				"subscriberID", subscriberID,
				"serviceKey", serviceKey)
		}
	}

	logger.Debug("批量订阅者通知完成",
		"serviceKey", serviceKey,
		"notifyCount", notifyCount)

	// 2. 通知命名空间订阅者（匹配命名空间的所有客户端）
	//    从 namespaceSubscribers[namespaceKey] 获取所有订阅了该命名空间的客户端
	if subs, ok := s.namespaceSubscribers[namespaceKey]; ok {
		for subscriberID, ch := range subs {
			// 发送事件到该客户端的 channel
			s.sendEvent(ch, event)
			_ = subscriberID // 用于调试
		}
	}
}

// sendEvent 发送事件到通道（非阻塞）
//
// 事件传递流程：
//  1. 事件通过 channel 发送到 Handler 的 goroutine
//  2. Handler 的 goroutine 从 channel 读取事件（在 SubscribeService 方法中）
//  3. Handler 通过 gRPC stream.Send() 将事件推送给客户端
//
// 为什么使用非阻塞发送？
//   - 如果客户端处理慢，channel 可能已满
//   - 阻塞发送会导致 NotifyServiceChange 阻塞，影响其他订阅者
//   - 非阻塞发送保证事件通知流程不被阻塞
//
// 处理策略：
//   - 使用 select 的 default 分支实现非阻塞发送
//   - 如果 channel 已满（客户端处理慢），丢弃事件
//   - 避免阻塞通知流程，保证其他订阅者能及时收到事件
//
// 性能考虑：
//   - channel 容量为 100，正常情况下不会满
//   - 如果频繁丢弃事件，说明客户端处理能力不足
//   - 可以考虑增加 channel 容量或记录警告日志
//
// 示例流程：
//
//	NotifyServiceChange() 调用 sendEvent(ch1, event)
//	-> event 发送到 ch1（非阻塞）
//	-> Handler 的 goroutine 从 ch1 读取 event
//	-> stream.Send(event) 推送给客户端 A
func (s *ServiceSubscriber) sendEvent(ch chan *pb.ServiceChangeEvent, event *pb.ServiceChangeEvent) {
	select {
	case ch <- event:
		// 发送成功：事件已放入 channel，Handler 的 goroutine 会读取并推送给客户端
	default:
		// 通道已满，丢弃事件（避免阻塞）
		// 说明：客户端处理慢，channel 缓冲区（100）已满
		// 此时丢弃事件，避免阻塞其他订阅者的通知流程
	}
}

// GetSubscriberCount 获取订阅者数量（批量订阅）
func (s *ServiceSubscriber) GetSubscriberCount(tenantId, namespaceId, groupName, serviceName string) int {
	serviceKey := s.makeServiceKey(tenantId, namespaceId, groupName, serviceName)

	s.mu.RLock()
	defer s.mu.RUnlock()

	// 统计批量订阅者数量
	count := 0
	for _, services := range s.batchSubscribers {
		if _, ok := services[serviceKey]; ok {
			count++
		}
	}
	return count
}

// makeServiceKey 生成服务唯一键
func (s *ServiceSubscriber) makeServiceKey(tenantId, namespaceId, groupName, serviceName string) string {
	return tenantId + ":" + namespaceId + ":" + groupName + ":" + serviceName
}

// makeNamespaceKey 生成命名空间唯一键
func (s *ServiceSubscriber) makeNamespaceKey(tenantId, namespaceId, groupName string) string {
	if groupName == "" {
		return tenantId + ":" + namespaceId
	}
	return tenantId + ":" + namespaceId + ":" + groupName
}

// 辅助转换方法

// convertNodeToProto 转换节点为 protobuf 格式
func convertNodeToProto(node *types.ServiceNode) *pb.Node {
	if node == nil {
		return nil
	}
	// 转换 JSON 字符串为 map
	metadata := make(map[string]string)
	if node.MetadataJson != "" {
		if err := json.Unmarshal([]byte(node.MetadataJson), &metadata); err != nil {
			metadata = make(map[string]string)
		}
	}

	return &pb.Node{
		NodeId:         node.NodeId,
		NamespaceId:    node.NamespaceId,
		GroupName:      node.GroupName,
		ServiceName:    node.ServiceName,
		IpAddress:      node.IpAddress,
		PortNumber:     int32(node.PortNumber),
		Weight:         node.Weight,
		Ephemeral:      node.Ephemeral,
		InstanceStatus: node.InstanceStatus,
		HealthyStatus:  node.HealthyStatus,
		Metadata:       metadata,
	}
}

// convertNodesToProto 转换节点列表为 protobuf 格式
func convertNodesToProto(nodes []*types.ServiceNode) []*pb.Node {
	result := make([]*pb.Node, 0, len(nodes))
	for _, node := range nodes {
		result = append(result, convertNodeToProto(node))
	}
	return result
}
