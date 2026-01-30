package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gateway/internal/servicecenter/cache"
	pb "gateway/internal/servicecenter/server/proto"
	"gateway/internal/servicecenter/server/subscriber"
	"gateway/internal/servicecenter/types"
	"gateway/pkg/logger"
	"gateway/pkg/utils/random"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// 服务注册发现架构说明
//
// 数据写入策略：
//   - 服务注册/发现：直接写缓存（高性能，低延迟）
//   - 配置管理：直接写数据库（保证持久化）
//
// 数据同步策略：
//   - 缓存到数据库：由外部异步同步服务负责（定期批量写入）
//   - 数据库到缓存：服务启动时加载，或者手动触发重新加载
//
// 为什么服务注册不直接写数据库？
//   - 高频操作：服务注册/注销/心跳是高频操作，直接写数据库会成为瓶颈
//   - 低延迟要求：服务发现需要毫秒级响应，缓存可以提供更好的性能
//   - 数据一致性：缓存足以保证服务发现的强一致性需求
//   - 持久化保证：外部异步同步服务确保数据最终持久化
//
// 异常恢复：
//   - 服务重启时从数据库加载数据到缓存
//   - 缓存丢失时可以从数据库恢复

// RegistryHandler gRPC 服务注册发现处理器
type RegistryHandler struct {
	pb.UnimplementedServiceRegistryServer
	serviceSubMgr *subscriber.ServiceSubscriber
}

// NewRegistryHandler 创建服务注册发现处理器
func NewRegistryHandler() *RegistryHandler {
	return &RegistryHandler{
		serviceSubMgr: subscriber.NewServiceSubscriber(),
	}
}

// validateNamespace 验证命名空间是否存在且有效（纯缓存操作）
// 如果命名空间不存在或已被禁用，返回权限错误
// 注意：命名空间应该在服务启动时已加载到缓存，这里只从缓存校验
func (h *RegistryHandler) validateNamespace(ctx context.Context, tenantId, namespaceId string) error {
	if namespaceId == "" {
		return status.Errorf(codes.InvalidArgument, "namespaceId is required")
	}

	// 从缓存获取命名空间
	namespace, found := cache.GetGlobalCache().GetNamespace(ctx, tenantId, namespaceId)
	if !found || namespace == nil {
		return status.Errorf(codes.PermissionDenied, "namespace not found: %s", namespaceId)
	}

	// 检查命名空间是否已禁用
	if namespace.ActiveFlag != "Y" {
		return status.Errorf(codes.PermissionDenied, "namespace is disabled: %s", namespaceId)
	}

	return nil
}

// GetServiceSubscriber 获取服务订阅管理器（供外部手动触发事件使用）
func (h *RegistryHandler) GetServiceSubscriber() *subscriber.ServiceSubscriber {
	return h.serviceSubMgr
}

// 服务注册/注销

// RegisterService 注册服务（可同时注册一个节点）
// 注意：直接写缓存，不写数据库。外部异步同步服务负责持久化。
func (h *RegistryHandler) RegisterService(ctx context.Context, req *pb.Service) (*pb.RegisterServiceResponse, error) {
	if req == nil {
		return &pb.RegisterServiceResponse{
			Success: false,
			Message: "service is required",
		}, nil
	}

	// 验证必填字段
	if req.NamespaceId == "" {
		return &pb.RegisterServiceResponse{
			Success: false,
			Message: "namespaceId is required",
		}, nil
	}
	if req.ServiceName == "" {
		return &pb.RegisterServiceResponse{
			Success: false,
			Message: "serviceName is required",
		}, nil
	}

	// 验证命名空间是否存在
	tenantID := "default" // TODO: 从 context 获取
	if err := h.validateNamespace(ctx, tenantID, req.NamespaceId); err != nil {
		return &pb.RegisterServiceResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// 设置默认值
	groupName := req.GroupName
	if groupName == "" {
		groupName = "DEFAULT_GROUP"
	}

	serviceType := req.ServiceType
	if serviceType == "" {
		serviceType = types.ServiceTypeInternal
	}

	protectThreshold := req.ProtectThreshold
	if protectThreshold < 0 || protectThreshold > 1 {
		protectThreshold = 0.0 // 默认不保护
	}

	// 转换 metadata map 为 JSON 字符串（不校验，失败时使用空字符串）
	metadataJson := ""
	if len(req.Metadata) > 0 {
		if metadataBytes, err := json.Marshal(req.Metadata); err == nil {
			metadataJson = string(metadataBytes)
		}
	}

	// 转换 tags map 为 JSON 字符串（不校验，失败时使用空字符串）
	tagsJson := ""
	if len(req.Tags) > 0 {
		if tagsBytes, err := json.Marshal(req.Tags); err == nil {
			tagsJson = string(tagsBytes)
		}
	}

	// 构建 Service 对象（包含所有字段的默认值）
	now := time.Now()
	service := &types.Service{
		TenantId:           "default", // TODO: 从 context 获取
		NamespaceId:        req.NamespaceId,
		GroupName:          groupName,
		ServiceName:        req.ServiceName,
		ServiceType:        serviceType,
		ServiceVersion:     req.ServiceVersion,
		ServiceDescription: req.ServiceDescription,
		MetadataJson:       metadataJson,
		TagsJson:           tagsJson,
		ProtectThreshold:   protectThreshold,
		SelectorJson:       "", // 不使用外部选择器
		AddTime:            now,
		AddWho:             "", // TODO: 从 context 获取
		EditTime:           now,
		EditWho:            "",                                 // TODO: 从 context 获取
		OprSeqFlag:         random.Generate32BitRandomString(), // 生成32位随机操作序列标识
		CurrentVersion:     1,
		ActiveFlag:         "Y", // 默认激活
		NoteText:           "",
		ExtProperty:        "",
	}

	// 直接写缓存（不写数据库）
	cache.GetGlobalCache().SetService(ctx, service)

	// 如果请求中携带了节点，同时注册节点
	var nodeId string
	if req.Node != nil {
		// 验证节点必填字段
		if req.Node.IpAddress == "" {
			return &pb.RegisterServiceResponse{
				Success: false,
				Message: "node ipAddress is required",
			}, nil
		}
		if req.Node.PortNumber <= 0 {
			return &pb.RegisterServiceResponse{
				Success: false,
				Message: "node portNumber must be greater than 0",
			}, nil
		}

		// 生成节点 ID（使用随机字符串生成器，支持集群环境）
		nodeId = random.Generate32BitRandomString()

		// 转换 metadata map 为 JSON 字符串（不校验，失败时使用空字符串）
		nodeMetadataJson := ""
		if len(req.Node.Metadata) > 0 {
			if metadataBytes, err := json.Marshal(req.Node.Metadata); err == nil {
				nodeMetadataJson = string(metadataBytes)
			}
		}

		// 设置节点默认值
		nodeGroupName := req.Node.GroupName
		if nodeGroupName == "" {
			nodeGroupName = groupName // 使用服务的 groupName
		}

		nodeInstanceStatus := req.Node.InstanceStatus
		if nodeInstanceStatus == "" {
			nodeInstanceStatus = types.NodeStatusUp
		}

		nodeHealthyStatus := req.Node.HealthyStatus
		if nodeHealthyStatus == "" {
			nodeHealthyStatus = types.HealthyStatusHealthy
		}

		nodeEphemeral := req.Node.Ephemeral
		if nodeEphemeral == "" {
			nodeEphemeral = "Y" // 默认临时节点
		}

		nodeWeight := req.Node.Weight
		if nodeWeight <= 0 {
			nodeWeight = 1.0 // 默认权重
		}

		// 构建节点对象（包含所有字段的默认值）
		nodeNow := time.Now()
		node := &types.ServiceNode{
			NodeId:         nodeId,
			TenantId:       "default", // TODO: 从 context 获取
			NamespaceId:    req.NamespaceId,
			GroupName:      nodeGroupName,
			ServiceName:    req.ServiceName,
			IpAddress:      req.Node.IpAddress,
			PortNumber:     int(req.Node.PortNumber),
			InstanceStatus: nodeInstanceStatus,
			HealthyStatus:  nodeHealthyStatus,
			Ephemeral:      nodeEphemeral,
			Weight:         nodeWeight,
			MetadataJson:   nodeMetadataJson,
			RegisterTime:   nodeNow,
			LastBeatTime:   &nodeNow,
			LastCheckTime:  &nodeNow,
			AddTime:        nodeNow,
			AddWho:         "", // TODO: 从 context 获取
			EditTime:       nodeNow,
			EditWho:        "",                                 // TODO: 从 context 获取
			OprSeqFlag:     random.Generate32BitRandomString(), // 生成32位随机操作序列标识
			CurrentVersion: 1,
			ActiveFlag:     "Y", // 默认激活
			NoteText:       "",
			ExtProperty:    "",
		}

		// 直接添加到缓存（不写数据库）
		cache.GetGlobalCache().AddNode(ctx, node)

		// 从缓存获取完整的服务信息（包括所有节点）
		// 注意：AddNode 会自动创建服务（如果不存在），所以服务应该存在
		service, found := cache.GetGlobalCache().GetService(ctx, node.TenantId, node.NamespaceId, node.GroupName, node.ServiceName)
		if !found || service == nil {
			// 理论上不应该发生（AddNode 会自动创建服务），但为了健壮性，记录警告并重新获取
			logger.Warn("节点添加后服务不存在，可能是并发问题，重新获取",
				"nodeId", node.NodeId,
				"namespaceId", node.NamespaceId,
				"groupName", node.GroupName,
				"serviceName", node.ServiceName)
			// 重新获取一次（可能由于并发问题）
			service, found = cache.GetGlobalCache().GetService(ctx, node.TenantId, node.NamespaceId, node.GroupName, node.ServiceName)
		}

		// 构建事件（包含服务信息和所有节点列表）
		// 如果服务仍然不存在，使用空的服务信息（理论上不应该发生）
		var pbService *pb.Service
		var pbNodes []*pb.Node
		if found && service != nil {
			pbService = convertServiceToProto(service)
			pbNodes = make([]*pb.Node, 0, len(service.Nodes))
			for _, n := range service.Nodes {
				pbNodes = append(pbNodes, convertNodeToProto(n))
			}
		} else {
			// 如果服务仍然不存在，创建空的服务信息（理论上不应该发生）
			pbNodes = []*pb.Node{convertNodeToProto(node)}
		}

		event := &pb.ServiceChangeEvent{
			EventType:   "NODE_ADDED",
			Timestamp:   time.Now().Format("2006-01-02 15:04:05"),
			NamespaceId: node.NamespaceId,
			GroupName:   node.GroupName,
			ServiceName: node.ServiceName,
			Service:     pbService,
			Nodes:       pbNodes,
			ChangedNode: convertNodeToProto(node),
		}
		h.serviceSubMgr.NotifyServiceChange(node.TenantId, node.NamespaceId, node.GroupName, node.ServiceName, event)
	}

	return &pb.RegisterServiceResponse{
		Success: true,
		Message: "service registered successfully",
		NodeId:  nodeId, // 如果注册了节点，返回 nodeId；否则为空
	}, nil
}

// UnregisterService 注销服务
// 注意：如果指定了 nodeId，只删除该节点；否则删除整个服务。直接从缓存删除，不操作数据库。
func (h *RegistryHandler) UnregisterService(ctx context.Context, req *pb.ServiceKey) (*pb.RegistryResponse, error) {
	tenantID := "default" // TODO: 从 context 获取

	// 验证命名空间是否存在
	if err := h.validateNamespace(ctx, tenantID, req.NamespaceId); err != nil {
		return &pb.RegistryResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// 如果指定了 nodeId，只删除该节点
	if req.NodeId != "" {
		// 先通过 nodeId 获取节点信息（用于构建事件）
		node, found := cache.GetGlobalCache().GetNode(ctx, tenantID, req.NodeId)
		if !found || node == nil {
			// 如果节点不存在，直接返回成功（幂等性）
			return &pb.RegistryResponse{
				Success: true,
				Message: "node not found or already removed",
			}, nil
		}

		// 保存节点信息（用于构建事件）
		savedNode := node

		// 直接调用 cache 删除节点
		cache.GetGlobalCache().RemoveNode(ctx, tenantID, req.NamespaceId, req.GroupName, req.ServiceName, req.NodeId)

		// 从缓存获取完整的服务信息（包括删除后的所有节点列表）
		service, serviceFound := cache.GetGlobalCache().GetService(ctx, tenantID, req.NamespaceId, req.GroupName, req.ServiceName)
		if !serviceFound || service == nil {
			// 如果服务不存在，说明节点已经被删除（可能是并发删除或之前已经删除）
			// 这是正常的幂等性处理，不需要发送事件（因为节点已经不存在了）
			logger.Info("节点删除后服务不存在，节点可能已被删除，不发送事件",
				"nodeId", req.NodeId,
				"namespaceId", req.NamespaceId,
				"groupName", req.GroupName,
				"serviceName", req.ServiceName)
			return &pb.RegistryResponse{
				Success: true,
				Message: "node unregistered successfully",
			}, nil
		}

		// 构建完整的事件（包含服务信息和删除后的所有节点列表）
		pbNodes := make([]*pb.Node, 0, len(service.Nodes))
		for _, n := range service.Nodes {
			pbNodes = append(pbNodes, convertNodeToProto(n))
		}
		event := &pb.ServiceChangeEvent{
			EventType:   "NODE_REMOVED",
			Timestamp:   time.Now().Format("2006-01-02 15:04:05"),
			NamespaceId: req.NamespaceId,
			GroupName:   req.GroupName,
			ServiceName: req.ServiceName,
			Service:     convertServiceToProto(service),
			Nodes:       pbNodes,
			ChangedNode: convertNodeToProto(savedNode),
		}
		h.serviceSubMgr.NotifyServiceChange(tenantID, req.NamespaceId, req.GroupName, req.ServiceName, event)

		return &pb.RegistryResponse{
			Success: true,
			Message: "node unregistered successfully",
		}, nil
	}

	// 如果没有指定 nodeId，删除整个服务
	cache.GetGlobalCache().DeleteService(ctx, tenantID, req.NamespaceId, req.GroupName, req.ServiceName)

	return &pb.RegistryResponse{
		Success: true,
		Message: "service unregistered successfully",
	}, nil
}

// GetService 获取服务信息（包含节点列表）
func (h *RegistryHandler) GetService(ctx context.Context, req *pb.ServiceKey) (*pb.GetServiceResponse, error) {
	tenantID := "default" // TODO: 从 context 获取

	// 验证命名空间是否存在
	if err := h.validateNamespace(ctx, tenantID, req.NamespaceId); err != nil {
		return &pb.GetServiceResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// 从缓存获取（使用全局单例）
	service, found := cache.GetGlobalCache().GetService(ctx, tenantID, req.NamespaceId, req.GroupName, req.ServiceName)
	if !found || service == nil {
		return &pb.GetServiceResponse{
			Success: false,
			Message: "service not found",
		}, nil
	}

	// 转换为 protobuf 格式
	pbService := convertServiceToProto(service)

	// 转换节点列表
	pbNodes := make([]*pb.Node, 0, len(service.Nodes))
	for _, node := range service.Nodes {
		pbNodes = append(pbNodes, convertNodeToProto(node))
	}

	return &pb.GetServiceResponse{
		Success: true,
		Message: "service found",
		Service: pbService,
		Nodes:   pbNodes,
	}, nil
}

// 节点注册/注销

// RegisterNode 注册服务节点
// 注意：直接写缓存，不写数据库。外部异步同步服务负责持久化。
//
// 支持两种场景：
//  1. 首次注册：不传 nodeId，服务端自动生成新的 nodeId
//  2. 重连注册：传入已有的 nodeId，复用原有 nodeId，更新节点信息而不是创建新节点
//
// 重连场景说明：
//   - 客户端断线重连后，可以传入之前的 nodeId 进行重新注册
//   - 服务端会检查该 nodeId 是否已存在，如果存在则更新节点信息
//   - 这样可以避免重连时注册多个节点，保持服务节点列表的稳定性
func (h *RegistryHandler) RegisterNode(ctx context.Context, req *pb.Node) (*pb.RegisterNodeResponse, error) {
	if req == nil {
		return &pb.RegisterNodeResponse{
			Success: false,
			Message: "node is required",
		}, nil
	}

	// 验证必填字段
	if req.NamespaceId == "" {
		return &pb.RegisterNodeResponse{
			Success: false,
			Message: "namespaceId is required",
		}, nil
	}
	if req.ServiceName == "" {
		return &pb.RegisterNodeResponse{
			Success: false,
			Message: "serviceName is required",
		}, nil
	}
	if req.IpAddress == "" {
		return &pb.RegisterNodeResponse{
			Success: false,
			Message: "ipAddress is required",
		}, nil
	}
	if req.PortNumber <= 0 {
		return &pb.RegisterNodeResponse{
			Success: false,
			Message: "portNumber must be greater than 0",
		}, nil
	}

	tenantID := "default" // TODO: 从 context 获取

	// 验证命名空间是否存在
	if err := h.validateNamespace(ctx, tenantID, req.NamespaceId); err != nil {
		return &pb.RegisterNodeResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// 判断是否为重连注册（客户端传入了 nodeId）
	var nodeID string
	var isReconnect bool
	var existingNode *types.ServiceNode

	if req.NodeId != "" {
		// 客户端传入了 nodeId，检查是否已存在
		nodeID = req.NodeId
		existingNode, _ = cache.GetGlobalCache().GetNode(ctx, tenantID, nodeID)
		if existingNode != nil {
			isReconnect = true
			logger.Info("检测到重连注册，复用已有 nodeId",
				"nodeId", nodeID,
				"namespaceId", req.NamespaceId,
				"serviceName", req.ServiceName)
		} else {
			// nodeId 不存在于缓存，作为新节点注册（使用客户端提供的 nodeId）
			logger.Info("客户端提供的 nodeId 不存在于缓存，作为新节点注册",
				"nodeId", nodeID,
				"namespaceId", req.NamespaceId,
				"serviceName", req.ServiceName)
		}
	} else {
		// 客户端未传入 nodeId，生成新的 nodeId
		nodeID = random.Generate32BitRandomString()
	}

	// 转换 metadata map 为 JSON 字符串（不校验，失败时使用空字符串）
	metadataJson := ""
	if len(req.Metadata) > 0 {
		if metadataBytes, err := json.Marshal(req.Metadata); err == nil {
			metadataJson = string(metadataBytes)
		}
	}

	// 设置默认值
	groupName := req.GroupName
	if groupName == "" {
		groupName = "DEFAULT_GROUP"
	}

	instanceStatus := req.InstanceStatus
	if instanceStatus == "" {
		instanceStatus = types.NodeStatusUp
	}

	healthyStatus := req.HealthyStatus
	if healthyStatus == "" {
		healthyStatus = types.HealthyStatusHealthy
	}

	ephemeral := req.Ephemeral
	if ephemeral == "" {
		ephemeral = "Y" // 默认临时节点
	}

	weight := req.Weight
	if weight <= 0 {
		weight = 1.0 // 默认权重
	}

	nodeNow := time.Now()
	var node *types.ServiceNode

	if isReconnect && existingNode != nil {
		// 重连场景：更新已存在的节点信息
		node = existingNode
		node.NamespaceId = req.NamespaceId
		node.GroupName = groupName
		node.ServiceName = req.ServiceName
		node.IpAddress = req.IpAddress
		node.PortNumber = int(req.PortNumber)
		node.InstanceStatus = instanceStatus
		node.HealthyStatus = healthyStatus
		node.Ephemeral = ephemeral
		node.Weight = weight
		node.MetadataJson = metadataJson
		node.LastBeatTime = &nodeNow
		node.LastCheckTime = &nodeNow
		node.EditTime = nodeNow
		node.ActiveFlag = "Y"

		// 更新缓存中的节点信息
		cache.GetGlobalCache().UpdateNode(ctx, node)

		logger.Info("重连注册：更新已存在的节点信息",
			"nodeId", node.NodeId,
			"namespaceId", node.NamespaceId,
			"groupName", node.GroupName,
			"serviceName", node.ServiceName)
	} else {
		// 首次注册：构建新的节点对象
		node = &types.ServiceNode{
			NodeId:         nodeID,
			TenantId:       tenantID,
			NamespaceId:    req.NamespaceId,
			GroupName:      groupName,
			ServiceName:    req.ServiceName,
			IpAddress:      req.IpAddress,
			PortNumber:     int(req.PortNumber),
			InstanceStatus: instanceStatus,
			HealthyStatus:  healthyStatus,
			Ephemeral:      ephemeral,
			Weight:         weight,
			MetadataJson:   metadataJson,
			RegisterTime:   nodeNow,
			LastBeatTime:   &nodeNow,
			LastCheckTime:  &nodeNow,
			AddTime:        nodeNow,
			AddWho:         "", // TODO: 从 context 获取
			EditTime:       nodeNow,
			EditWho:        "",                                 // TODO: 从 context 获取
			OprSeqFlag:     random.Generate32BitRandomString(), // 生成32位随机操作序列标识
			CurrentVersion: 1,
			ActiveFlag:     "Y", // 默认激活
			NoteText:       "",
			ExtProperty:    "",
		}

		// 直接添加到缓存（不写数据库）
		// 注意：AddNode 会自动创建服务（如果不存在）
		cache.GetGlobalCache().AddNode(ctx, node)
	}

	// 从缓存获取完整的服务信息（包括所有节点）
	// 注意：AddNode 会自动创建服务（如果不存在），所以服务应该存在
	service, found := cache.GetGlobalCache().GetService(ctx, node.TenantId, node.NamespaceId, node.GroupName, node.ServiceName)
	if !found || service == nil {
		// 理论上不应该发生（AddNode 会自动创建服务），但为了健壮性，记录警告并重新获取
		logger.Warn("节点添加后服务不存在，可能是并发问题，重新获取",
			"nodeId", node.NodeId,
			"namespaceId", node.NamespaceId,
			"groupName", node.GroupName,
			"serviceName", node.ServiceName)
		// 重新获取一次（可能由于并发问题）
		service, found = cache.GetGlobalCache().GetService(ctx, node.TenantId, node.NamespaceId, node.GroupName, node.ServiceName)
	}

	// 构建事件（包含服务信息和所有节点列表）
	// 如果服务仍然不存在，使用空的服务信息（理论上不应该发生）
	var pbService *pb.Service
	var pbNodes []*pb.Node
	if found && service != nil {
		pbService = convertServiceToProto(service)
		pbNodes = make([]*pb.Node, 0, len(service.Nodes))
		for _, n := range service.Nodes {
			pbNodes = append(pbNodes, convertNodeToProto(n))
		}
	} else {
		// 如果服务仍然不存在，创建空的服务信息（理论上不应该发生）
		pbNodes = []*pb.Node{convertNodeToProto(node)}
	}

	// 根据是否为重连，使用不同的事件类型
	eventType := "NODE_ADDED"
	if isReconnect {
		eventType = "NODE_UPDATED"
	}

	event := &pb.ServiceChangeEvent{
		EventType:   eventType,
		Timestamp:   time.Now().Format("2006-01-02 15:04:05"),
		NamespaceId: node.NamespaceId,
		GroupName:   node.GroupName,
		ServiceName: node.ServiceName,
		Service:     pbService,
		Nodes:       pbNodes,
		ChangedNode: convertNodeToProto(node),
	}

	logger.Info("节点注册成功，发送服务变更通知",
		"nodeId", node.NodeId,
		"tenantId", node.TenantId,
		"namespaceId", node.NamespaceId,
		"groupName", node.GroupName,
		"serviceName", node.ServiceName,
		"eventType", event.EventType,
		"isReconnect", isReconnect)

	h.serviceSubMgr.NotifyServiceChange(
		node.TenantId,
		node.NamespaceId,
		node.GroupName,
		node.ServiceName,
		event,
	)

	return &pb.RegisterNodeResponse{
		Success: true,
		Message: "node registered successfully",
		NodeId:  nodeID, // 返回 nodeId（新生成的或复用的）
	}, nil
}

// UnregisterNode 注销服务节点
// 注意：直接从缓存删除，不操作数据库。外部异步同步服务负责持久化。
func (h *RegistryHandler) UnregisterNode(ctx context.Context, req *pb.NodeKey) (*pb.RegistryResponse, error) {
	tenantID := "default" // TODO: 从 context 获取

	// 先通过 nodeId 直接获取节点信息（使用 nodeIndex，O(1) 时间复杂度）
	node, found := cache.GetGlobalCache().GetNode(ctx, tenantID, req.NodeId)
	if !found || node == nil {
		// 如果节点不存在，直接返回成功（幂等性）
		return &pb.RegistryResponse{
			Success: true,
			Message: "node not found or already removed",
		}, nil
	}

	// 保存节点信息（用于构建事件）
	savedNode := node

	// 直接从缓存删除（不操作数据库）
	cache.GetGlobalCache().RemoveNode(ctx, node.TenantId, node.NamespaceId, node.GroupName, node.ServiceName, node.NodeId)

	// 从缓存获取完整的服务信息（包括删除后的所有节点列表）
	service, serviceFound := cache.GetGlobalCache().GetService(ctx, node.TenantId, node.NamespaceId, node.GroupName, node.ServiceName)
	if !serviceFound || service == nil {
		// 如果服务不存在，说明节点已经被删除（可能是并发删除或之前已经删除）
		// 这是正常的幂等性处理，不需要发送事件（因为节点已经不存在了）
		logger.Info("节点删除后服务不存在，节点可能已被删除，不发送事件",
			"nodeId", node.NodeId,
			"namespaceId", node.NamespaceId,
			"groupName", node.GroupName,
			"serviceName", node.ServiceName)
		return &pb.RegistryResponse{
			Success: true,
			Message: "node unregistered successfully",
		}, nil
	}

	// 构建完整的事件（包含服务信息和删除后的所有节点列表）
	pbNodes := make([]*pb.Node, 0, len(service.Nodes))
	for _, n := range service.Nodes {
		pbNodes = append(pbNodes, convertNodeToProto(n))
	}
	event := &pb.ServiceChangeEvent{
		EventType:   "NODE_REMOVED",
		Timestamp:   time.Now().Format("2006-01-02 15:04:05"),
		NamespaceId: node.NamespaceId,
		GroupName:   node.GroupName,
		ServiceName: node.ServiceName,
		Service:     convertServiceToProto(service),
		Nodes:       pbNodes,
		ChangedNode: convertNodeToProto(savedNode),
	}
	h.serviceSubMgr.NotifyServiceChange(
		node.TenantId,
		node.NamespaceId,
		node.GroupName,
		node.ServiceName,
		event,
	)

	return &pb.RegistryResponse{
		Success: true,
		Message: "node unregistered successfully",
	}, nil
}

// 服务发现

// DiscoverNodes 发现服务节点
func (h *RegistryHandler) DiscoverNodes(ctx context.Context, req *pb.DiscoverNodesRequest) (*pb.DiscoverNodesResponse, error) {
	tenantID := "default" // TODO: 从 context 获取

	// 验证命名空间是否存在
	if err := h.validateNamespace(ctx, tenantID, req.NamespaceId); err != nil {
		return &pb.DiscoverNodesResponse{
			Success: false,
			Message: err.Error(),
			Nodes:   []*pb.Node{},
		}, nil
	}

	// 从缓存获取（使用全局单例）
	service, found := cache.GetGlobalCache().GetService(ctx, tenantID, req.NamespaceId, req.GroupName, req.ServiceName)
	if !found || service == nil {
		return &pb.DiscoverNodesResponse{
			Success: true,
			Message: "no nodes found",
			Nodes:   []*pb.Node{},
		}, nil
	}

	// 过滤节点
	nodes := service.Nodes
	if req.HealthyOnly {
		healthyNodes := make([]*types.ServiceNode, 0)
		for _, node := range nodes {
			if node.HealthyStatus == "HEALTHY" {
				healthyNodes = append(healthyNodes, node)
			}
		}
		nodes = healthyNodes
	}

	// 转换为 protobuf 格式
	pbNodes := make([]*pb.Node, 0, len(nodes))
	for _, node := range nodes {
		pbNodes = append(pbNodes, convertNodeToProto(node))
	}

	return &pb.DiscoverNodesResponse{
		Success: true,
		Message: fmt.Sprintf("found %d nodes", len(pbNodes)),
		Nodes:   pbNodes,
	}, nil
}

// 服务订阅（实时推送）
//
// ================================================================================
// 订阅机制设计说明
// ================================================================================
//
// 1. 订阅接口统一设计：
//    - SubscribeServices 是统一接口，支持订阅单个或多个服务
//    - 单个服务订阅：serviceNames 数组只包含一个服务名
//    - 多个服务订阅：serviceNames 数组包含多个服务名
//    - 所有订阅的服务共用同一个 channel，统一管理
//
// 2. 每个 gRPC Stream 连接都会创建一个新的订阅：
//    - 每个连接调用时，都会生成新的 subscriberID（如 "SUB_abc123..."）
//    - 每个连接都有独立的 channel
//    - 连接断开时，通过 defer 自动清理订阅
//
// 3. 为什么每个连接都创建新订阅？
//    - 每个 gRPC Stream 连接是独立的，需要独立管理生命周期
//    - 连接断开时需要清理对应的订阅，避免资源泄漏
//    - 同一个客户端应用如果有多个连接，会有多个订阅（这是合理的）
//    - 例如：客户端应用有 3 个连接，就会有 3 个独立的订阅和 channel
//
// 4. 统一订阅接口的优势：
//    - 一个连接可以订阅多个服务，减少连接数
//    - 所有服务的变更事件通过同一个 channel 推送
//    - 适合需要订阅大量服务的场景（如 Java 客户端）
//    - 单个服务订阅也使用相同的机制（统一实现）
//
// 5. 订阅数据结构：
//    - batchSubscribers[subscriberID][serviceKey] = channel
//    - 同一个 subscriberID 可以订阅多个服务，所有服务共用同一个 channel
//    - 例如：
//      batchSubscribers["SUB_abc123"]["default:public:DEFAULT:order-service"] = ch
//      batchSubscribers["SUB_abc123"]["default:public:DEFAULT:user-service"] = ch
//      表示：同一个连接订阅了多个服务，所有服务共用同一个 channel
//
// 6. 事件推送流程：
//    服务变更 -> NotifyServiceChange()
//           -> 查找 batchSubscribers[subscriberID][serviceKey]
//           -> 如果订阅了该服务，发送事件到共享 channel
//           -> Handler goroutine 从 channel 读取
//           -> stream.Send() 推送给客户端
//
// ================================================================================

// SubscribeServices 订阅服务（统一接口，支持订阅单个或多个服务）
//
// 接口设计说明：
//   - 这是统一的订阅接口，支持订阅单个或多个服务
//   - 单个服务订阅：serviceNames = ["user-service"]
//   - 多个服务订阅：serviceNames = ["order-service", "user-service", "pay-service"]
//   - 一个客户端应用可以订阅多个服务，所有服务共用同一个 channel
//   - 减少 gRPC Stream 连接数，提高效率
//
// 处理流程：
//  1. 验证请求参数（namespaceId、serviceNames）
//  2. 生成唯一的 subscriberID（每个连接独立）
//  3. 调用 serviceSubMgr.SubscribeMultipleServices() 订阅服务
//  4. 所有服务共用同一个 channel，减少连接数
//  5. 持续从 channel 读取事件并推送给客户端
//  6. 连接断开时，通过 defer 自动清理订阅
func (h *RegistryHandler) SubscribeServices(req *pb.SubscribeServicesRequest, stream pb.ServiceRegistry_SubscribeServicesServer) error {
	tenantID := "default" // TODO: 从 context 获取
	subscriberID := random.GenerateUniqueStringWithPrefix("SUB", 32)

	// 验证请求参数
	if req.NamespaceId == "" {
		return status.Errorf(codes.InvalidArgument, "namespaceId is required")
	}
	if len(req.ServiceNames) == 0 {
		return status.Errorf(codes.InvalidArgument, "serviceNames is required and cannot be empty")
	}

	// 验证命名空间是否存在
	if err := h.validateNamespace(stream.Context(), tenantID, req.NamespaceId); err != nil {
		return err
	}

	// 设置默认值
	groupName := req.GroupName
	if groupName == "" {
		groupName = "DEFAULT_GROUP"
	}

	// 打印订阅开始日志
	logger.Info("服务订阅注册开始",
		"subscriberID", subscriberID,
		"tenantID", tenantID,
		"namespaceId", req.NamespaceId,
		"groupName", groupName,
		"serviceNames", req.ServiceNames,
		"serviceCount", len(req.ServiceNames))

	// 批量订阅多个服务（使用同一个 channel）
	ch := h.serviceSubMgr.SubscribeMultipleServices(
		stream.Context(),
		tenantID,
		req.NamespaceId,
		groupName,
		req.ServiceNames,
		subscriberID,
	)
	defer func() {
		logger.Info("服务订阅注销",
			"subscriberID", subscriberID,
			"tenantID", tenantID,
			"namespaceId", req.NamespaceId,
			"groupName", groupName,
			"serviceNames", req.ServiceNames)
		h.serviceSubMgr.UnsubscribeMultipleServices(subscriberID)
	}()

	// 订阅成功后，立即推送当前服务信息给客户端（全量推送）
	// 直接发送到当前订阅者的 channel，不影响其他订阅者
	// 这样客户端可以立即获得最新服务信息，而不需要单独调用 GetService
	go func() {
		for _, serviceName := range req.ServiceNames {
			// 从缓存获取当前服务信息
			service, found := cache.GetGlobalCache().GetService(stream.Context(), tenantID, req.NamespaceId, groupName, serviceName)
			if !found || service == nil {
				// 服务不存在，推送服务不存在事件（表示服务未注册）
				notFoundEvent := &pb.ServiceChangeEvent{
					EventType:   "SERVICE_NOT_FOUND",
					Timestamp:   time.Now().Format("2006-01-02 15:04:05"),
					NamespaceId: req.NamespaceId,
					GroupName:   groupName,
					ServiceName: serviceName,
					Service:     nil,
					Nodes:       []*pb.Node{},
					ChangedNode: nil,
				}

				// 直接发送到当前订阅者的 channel（只发送给当前订阅者）
				h.serviceSubMgr.SendToSubscriber(subscriberID, notFoundEvent)

				logger.Debug("已推送服务不存在状态到 channel",
					"subscriberID", subscriberID,
					"namespaceId", req.NamespaceId,
					"groupName", groupName,
					"serviceName", serviceName)
				continue
			}

			// 服务存在，构建初始服务信息事件
			pbService := convertServiceToProto(service)
			pbNodes := make([]*pb.Node, 0, len(service.Nodes))
			for _, node := range service.Nodes {
				pbNodes = append(pbNodes, convertNodeToProto(node))
			}

			initialEvent := &pb.ServiceChangeEvent{
				EventType:   "SERVICE_INITIALIZED", // 使用 SERVICE_INITIALIZED 表示这是初始服务信息
				Timestamp:   time.Now().Format("2006-01-02 15:04:05"),
				NamespaceId: service.NamespaceId,
				GroupName:   service.GroupName,
				ServiceName: service.ServiceName,
				Service:     pbService,
				Nodes:       pbNodes,
				ChangedNode: nil, // 初始推送不包含变更的节点
			}

			// 直接发送到当前订阅者的 channel（只发送给当前订阅者）
			h.serviceSubMgr.SendToSubscriber(subscriberID, initialEvent)

			logger.Debug("已推送初始服务信息到 channel",
				"subscriberID", subscriberID,
				"namespaceId", req.NamespaceId,
				"groupName", groupName,
				"serviceName", serviceName,
				"nodeCount", len(service.Nodes))
		}
	}()

	// 持续监听变更事件并推送给客户端
	// 所有服务的变更事件都会通过同一个 channel 推送
	for {
		select {
		case event, ok := <-ch:
			if !ok {
				return nil // 通道已关闭
			}
			if err := stream.Send(event); err != nil {
				return err
			}
		case <-stream.Context().Done():
			return stream.Context().Err()
		}
	}
}

// SubscribeNamespace 订阅整个命名空间/分组下的所有服务
//
// 处理流程：
//  1. 生成唯一的 subscriberID
//  2. 调用 serviceSubMgr.SubscribeNamespace() 订阅命名空间
//  3. 持续从 channel 读取事件并推送给客户端
//
// 功能说明：
//   - 订阅整个命名空间/分组下的所有服务
//   - 当命名空间下任何服务发生变更时，都会收到事件
//   - 适合需要监控整个命名空间服务变更的场景
func (h *RegistryHandler) SubscribeNamespace(req *pb.SubscribeNamespaceRequest, stream pb.ServiceRegistry_SubscribeNamespaceServer) error {
	tenantID := "default" // TODO: 从 context 获取
	subscriberID := random.GenerateUniqueStringWithPrefix("SUB", 32)

	// 验证请求参数
	if req.NamespaceId == "" {
		return status.Errorf(codes.InvalidArgument, "namespaceId is required")
	}

	// 验证命名空间是否存在
	if err := h.validateNamespace(stream.Context(), tenantID, req.NamespaceId); err != nil {
		return err
	}

	// 设置默认值
	groupName := req.GroupName
	if groupName == "" {
		groupName = "DEFAULT_GROUP"
	}

	// 打印订阅开始日志
	logger.Info("命名空间订阅注册开始",
		"subscriberID", subscriberID,
		"tenantID", tenantID,
		"namespaceId", req.NamespaceId,
		"groupName", groupName)

	// 订阅命名空间（会收到该命名空间下所有服务的变更事件）
	ch := h.serviceSubMgr.SubscribeNamespace(
		stream.Context(),
		tenantID,
		req.NamespaceId,
		groupName,
		subscriberID,
	)
	defer func() {
		logger.Info("命名空间订阅注销",
			"subscriberID", subscriberID,
			"tenantID", tenantID,
			"namespaceId", req.NamespaceId,
			"groupName", groupName)
		h.serviceSubMgr.UnsubscribeNamespace(tenantID, req.NamespaceId, groupName, subscriberID)
	}()

	// 持续监听变更事件并推送给客户端
	// 命名空间下所有服务的变更事件都会通过 channel 推送
	for {
		select {
		case event, ok := <-ch:
			if !ok {
				return nil // 通道已关闭
			}
			if err := stream.Send(event); err != nil {
				return err
			}
		case <-stream.Context().Done():
			return stream.Context().Err()
		}
	}
}

// 心跳

// Heartbeat 心跳上报
// 注意：心跳只更新缓存中的节点健康状态，不写数据库。外部异步同步服务负责持久化。
//
// 重连场景处理：
//   - 客户端可能因为网络问题断开连接后自动重连
//   - 重连后，新的 gRPC 连接会有新的 context
//   - 心跳请求包含完整的 Service 信息（包含节点信息），用于：
//   - 验证服务信息是否一致
//   - 更新服务信息（如服务版本、描述、元数据等可能变化）
//   - 更新节点信息（如 IP、端口、权重、元数据等可能变化）
//   - 连接跟踪器可以基于完整信息建立连接映射
//   - 网络重连后可以完整恢复服务和节点信息
func (h *RegistryHandler) Heartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.RegistryResponse, error) {
	// 验证请求参数
	if req.NodeId == "" {
		return &pb.RegistryResponse{
			Success: false,
			Message: "nodeId is required",
		}, nil
	}

	// 从缓存中快速查找节点（使用 nodeIndex，O(1) 时间复杂度）
	tenantID := "default" // TODO: 从 context 获取
	targetNode, found := cache.GetGlobalCache().GetNode(ctx, tenantID, req.NodeId)

	var targetService *types.Service
	if found && targetNode != nil {
		// 通过节点信息获取服务
		service, serviceFound := cache.GetGlobalCache().GetService(ctx, tenantID, targetNode.NamespaceId, targetNode.GroupName, targetNode.ServiceName)
		if serviceFound {
			targetService = service
		}
	}

	// 如果缓存中找不到节点，但心跳请求包含完整的 Service 信息，尝试自动恢复
	if !found || targetNode == nil || targetService == nil {
		// 如果心跳请求包含完整的 Service 信息，可以自动恢复
		if req.Service != nil && req.Service.Node != nil {
			// 验证 Service 信息是否完整
			if req.Service.NamespaceId == "" || req.Service.ServiceName == "" {
				return &pb.RegistryResponse{
					Success: false,
					Message: fmt.Sprintf("node not found and cannot recover: nodeId=%s (missing namespaceId or serviceName)", req.NodeId),
				}, nil
			}
			if req.Service.Node.IpAddress == "" || req.Service.Node.PortNumber <= 0 {
				return &pb.RegistryResponse{
					Success: false,
					Message: fmt.Sprintf("node not found and cannot recover: nodeId=%s (missing ipAddress or portNumber)", req.NodeId),
				}, nil
			}

			// 验证 nodeId 是否一致
			if req.Service.Node.NodeId != "" && req.Service.Node.NodeId != req.NodeId {
				return &pb.RegistryResponse{
					Success: false,
					Message: fmt.Sprintf("nodeId mismatch: expected=%s, got=%s", req.NodeId, req.Service.Node.NodeId),
				}, nil
			}

			// 自动恢复服务和节点（类似 RegisterService 的逻辑）
			logger.Info("检测到节点不在缓存中，通过心跳自动恢复",
				"nodeId", req.NodeId,
				"namespaceId", req.Service.NamespaceId,
				"serviceName", req.Service.ServiceName)

			// 恢复服务信息
			recoveredService, recoveredNode, err := h.recoverServiceAndNodeFromHeartbeat(ctx, req.Service, req.NodeId)
			if err != nil {
				return &pb.RegistryResponse{
					Success: false,
					Message: fmt.Sprintf("failed to recover service and node: %v", err),
				}, nil
			}

			targetService = recoveredService
			targetNode = recoveredNode
			found = true

			logger.Info("成功通过心跳恢复服务和节点",
				"nodeId", req.NodeId,
				"namespaceId", req.Service.NamespaceId,
				"serviceName", req.Service.ServiceName)
		} else {
			// 没有提供完整的 Service 信息，无法恢复
			return &pb.RegistryResponse{
				Success: false,
				Message: fmt.Sprintf("node not found: nodeId=%s (please provide complete Service information in heartbeat to recover)", req.NodeId),
			}, nil
		}
	}

	// 如果心跳请求包含完整的 Service 信息，进行验证和更新
	if req.Service != nil {
		// 验证服务信息是否一致
		if req.Service.NamespaceId != "" && targetService.NamespaceId != req.Service.NamespaceId {
			return &pb.RegistryResponse{
				Success: false,
				Message: fmt.Sprintf("namespaceId mismatch: expected=%s, got=%s", targetService.NamespaceId, req.Service.NamespaceId),
			}, nil
		}
		if req.Service.GroupName != "" && targetService.GroupName != req.Service.GroupName {
			return &pb.RegistryResponse{
				Success: false,
				Message: fmt.Sprintf("groupName mismatch: expected=%s, got=%s", targetService.GroupName, req.Service.GroupName),
			}, nil
		}
		if req.Service.ServiceName != "" && targetService.ServiceName != req.Service.ServiceName {
			return &pb.RegistryResponse{
				Success: false,
				Message: fmt.Sprintf("serviceName mismatch: expected=%s, got=%s", targetService.ServiceName, req.Service.ServiceName),
			}, nil
		}

		// 更新服务信息（如果提供了新的服务信息）
		now := time.Now()
		serviceUpdated := false

		if req.Service.ServiceVersion != "" && targetService.ServiceVersion != req.Service.ServiceVersion {
			targetService.ServiceVersion = req.Service.ServiceVersion
			serviceUpdated = true
		}
		if req.Service.ServiceDescription != "" && targetService.ServiceDescription != req.Service.ServiceDescription {
			targetService.ServiceDescription = req.Service.ServiceDescription
			serviceUpdated = true
		}
		if req.Service.ProtectThreshold >= 0 && req.Service.ProtectThreshold <= 1 {
			targetService.ProtectThreshold = req.Service.ProtectThreshold
			serviceUpdated = true
		}

		// 更新服务元数据和标签
		if len(req.Service.Metadata) > 0 {
			if metadataBytes, err := json.Marshal(req.Service.Metadata); err == nil {
				targetService.MetadataJson = string(metadataBytes)
				serviceUpdated = true
			}
		}
		if len(req.Service.Tags) > 0 {
			if tagsBytes, err := json.Marshal(req.Service.Tags); err == nil {
				targetService.TagsJson = string(tagsBytes)
				serviceUpdated = true
			}
		}

		if serviceUpdated {
			targetService.EditTime = now
			cache.GetGlobalCache().SetService(ctx, targetService)
		}

		// 如果心跳请求包含节点信息，更新节点信息
		if req.Service.Node != nil {
			// 验证节点信息
			if req.Service.Node.NodeId != "" && req.Service.Node.NodeId != req.NodeId {
				return &pb.RegistryResponse{
					Success: false,
					Message: fmt.Sprintf("nodeId mismatch: expected=%s, got=%s", req.NodeId, req.Service.Node.NodeId),
				}, nil
			}

			// 更新节点信息（允许更新 IP、端口、权重、元数据等）
			nodeUpdated := false

			if req.Service.Node.IpAddress != "" && targetNode.IpAddress != req.Service.Node.IpAddress {
				targetNode.IpAddress = req.Service.Node.IpAddress
				nodeUpdated = true
			}
			if req.Service.Node.PortNumber > 0 && targetNode.PortNumber != int(req.Service.Node.PortNumber) {
				targetNode.PortNumber = int(req.Service.Node.PortNumber)
				nodeUpdated = true
			}
			if req.Service.Node.Weight > 0 && targetNode.Weight != req.Service.Node.Weight {
				targetNode.Weight = req.Service.Node.Weight
				nodeUpdated = true
			}
			if req.Service.Node.InstanceStatus != "" && targetNode.InstanceStatus != req.Service.Node.InstanceStatus {
				targetNode.InstanceStatus = req.Service.Node.InstanceStatus
				nodeUpdated = true
			}

			// 更新节点元数据
			if len(req.Service.Node.Metadata) > 0 {
				if metadataBytes, err := json.Marshal(req.Service.Node.Metadata); err == nil {
					targetNode.MetadataJson = string(metadataBytes)
					nodeUpdated = true
				}
			}

			if nodeUpdated {
				targetNode.EditTime = time.Now()
			}
		}
	}

	// 更新节点的最后心跳时间和健康状态
	now := time.Now()
	targetNode.LastBeatTime = &now
	targetNode.HealthyStatus = types.HealthyStatusHealthy
	targetNode.EditTime = now

	// 更新缓存（使用 UpdateNode 方法）
	cache.GetGlobalCache().UpdateNode(ctx, targetNode)

	return &pb.RegistryResponse{
		Success: true,
		Message: "heartbeat received",
	}, nil
}

// 辅助方法

// recoverServiceAndNodeFromHeartbeat 从心跳请求中恢复服务和节点
// 当缓存中找不到节点时，如果心跳包含完整的 Service 信息，可以自动恢复
func (h *RegistryHandler) recoverServiceAndNodeFromHeartbeat(ctx context.Context, pbService *pb.Service, nodeId string) (*types.Service, *types.ServiceNode, error) {
	if pbService == nil || pbService.Node == nil {
		return nil, nil, fmt.Errorf("service or node information is missing")
	}

	// 设置默认值
	groupName := pbService.GroupName
	if groupName == "" {
		groupName = "DEFAULT_GROUP"
	}

	serviceType := pbService.ServiceType
	if serviceType == "" {
		serviceType = types.ServiceTypeInternal
	}

	protectThreshold := pbService.ProtectThreshold
	if protectThreshold < 0 || protectThreshold > 1 {
		protectThreshold = 0.0
	}

	// 转换 metadata map 为 JSON 字符串
	metadataJson := ""
	if len(pbService.Metadata) > 0 {
		if metadataBytes, err := json.Marshal(pbService.Metadata); err == nil {
			metadataJson = string(metadataBytes)
		}
	}

	// 转换 tags map 为 JSON 字符串
	tagsJson := ""
	if len(pbService.Tags) > 0 {
		if tagsBytes, err := json.Marshal(pbService.Tags); err == nil {
			tagsJson = string(tagsBytes)
		}
	}

	// 构建 Service 对象
	now := time.Now()
	service := &types.Service{
		TenantId:           "default", // TODO: 从 context 获取
		NamespaceId:        pbService.NamespaceId,
		GroupName:          groupName,
		ServiceName:        pbService.ServiceName,
		ServiceType:        serviceType,
		ServiceVersion:     pbService.ServiceVersion,
		ServiceDescription: pbService.ServiceDescription,
		MetadataJson:       metadataJson,
		TagsJson:           tagsJson,
		ProtectThreshold:   protectThreshold,
		SelectorJson:       "",
		AddTime:            now,
		AddWho:             "",
		EditTime:           now,
		EditWho:            "",
		OprSeqFlag:         random.Generate32BitRandomString(), // 生成32位随机操作序列标识
		CurrentVersion:     1,
		ActiveFlag:         "Y",
		NoteText:           "",
		ExtProperty:        "",
	}

	// 直接写缓存（不写数据库）
	cache.GetGlobalCache().SetService(ctx, service)

	// 转换节点 metadata map 为 JSON 字符串
	nodeMetadataJson := ""
	if len(pbService.Node.Metadata) > 0 {
		if metadataBytes, err := json.Marshal(pbService.Node.Metadata); err == nil {
			nodeMetadataJson = string(metadataBytes)
		}
	}

	// 设置节点默认值
	nodeGroupName := pbService.Node.GroupName
	if nodeGroupName == "" {
		nodeGroupName = groupName
	}

	nodeInstanceStatus := pbService.Node.InstanceStatus
	if nodeInstanceStatus == "" {
		nodeInstanceStatus = types.NodeStatusUp
	}

	nodeHealthyStatus := pbService.Node.HealthyStatus
	if nodeHealthyStatus == "" {
		nodeHealthyStatus = types.HealthyStatusHealthy
	}

	nodeEphemeral := pbService.Node.Ephemeral
	if nodeEphemeral == "" {
		nodeEphemeral = "Y" // 默认临时节点
	}

	nodeWeight := pbService.Node.Weight
	if nodeWeight <= 0 {
		nodeWeight = 1.0
	}

	// 构建节点对象（使用心跳中的 nodeId，而不是生成新的）
	node := &types.ServiceNode{
		NodeId:         nodeId, // 使用心跳中的 nodeId
		TenantId:       "default",
		NamespaceId:    pbService.NamespaceId,
		GroupName:      nodeGroupName,
		ServiceName:    pbService.ServiceName,
		IpAddress:      pbService.Node.IpAddress,
		PortNumber:     int(pbService.Node.PortNumber),
		InstanceStatus: nodeInstanceStatus,
		HealthyStatus:  nodeHealthyStatus,
		Ephemeral:      nodeEphemeral,
		Weight:         nodeWeight,
		MetadataJson:   nodeMetadataJson,
		RegisterTime:   now, // 恢复时使用当前时间作为注册时间
		LastBeatTime:   &now,
		LastCheckTime:  &now,
		AddTime:        now,
		AddWho:         "",
		EditTime:       now,
		EditWho:        "",
		OprSeqFlag:     random.Generate32BitRandomString(), // 生成32位随机操作序列标识
		CurrentVersion: 1,
		ActiveFlag:     "Y",
		NoteText:       "",
		ExtProperty:    "",
	}

	// 直接添加到缓存（不写数据库）
	// 注意：AddNode 会自动创建服务（如果不存在）
	cache.GetGlobalCache().AddNode(ctx, node)

	// 从缓存获取完整的服务信息（包括所有节点）
	// 注意：AddNode 会自动创建服务（如果不存在），所以服务应该存在
	service, found := cache.GetGlobalCache().GetService(ctx, node.TenantId, node.NamespaceId, node.GroupName, node.ServiceName)
	if !found || service == nil {
		// 理论上不应该发生（AddNode 会自动创建服务），但为了健壮性，记录警告并重新获取
		logger.Warn("节点恢复后服务不存在，可能是并发问题，重新获取",
			"nodeId", node.NodeId,
			"namespaceId", node.NamespaceId,
			"groupName", node.GroupName,
			"serviceName", node.ServiceName)
		// 重新获取一次（可能由于并发问题）
		service, found = cache.GetGlobalCache().GetService(ctx, node.TenantId, node.NamespaceId, node.GroupName, node.ServiceName)
	}

	// 构建事件（包含服务信息和所有节点列表）
	// 如果服务仍然不存在，使用空的服务信息（理论上不应该发生）
	var eventService *pb.Service
	var pbNodes []*pb.Node
	if found && service != nil {
		eventService = convertServiceToProto(service)
		pbNodes = make([]*pb.Node, 0, len(service.Nodes))
		for _, n := range service.Nodes {
			pbNodes = append(pbNodes, convertNodeToProto(n))
		}
	} else {
		// 如果服务仍然不存在，创建空的服务信息（理论上不应该发生）
		pbNodes = []*pb.Node{convertNodeToProto(node)}
	}

	// 通知订阅者节点已恢复（作为 NODE_ADDED 事件）
	event := &pb.ServiceChangeEvent{
		EventType:   "NODE_ADDED",
		Timestamp:   now.Format("2006-01-02 15:04:05"),
		NamespaceId: node.NamespaceId,
		GroupName:   node.GroupName,
		ServiceName: node.ServiceName,
		Service:     eventService,
		Nodes:       pbNodes,
		ChangedNode: convertNodeToProto(node),
	}
	h.serviceSubMgr.NotifyServiceChange(node.TenantId, node.NamespaceId, node.GroupName, node.ServiceName, event)

	return service, node, nil
}

// convertServiceToProto 转换服务为 protobuf 格式
func convertServiceToProto(service *types.Service) *pb.Service {
	if service == nil {
		return nil
	}
	// 转换 metadata JSON 字符串为 map
	metadata := make(map[string]string)
	if service.MetadataJson != "" {
		if err := json.Unmarshal([]byte(service.MetadataJson), &metadata); err != nil {
			metadata = make(map[string]string)
		}
	}
	// 转换 tags JSON 字符串为 map
	tags := make(map[string]string)
	if service.TagsJson != "" {
		if err := json.Unmarshal([]byte(service.TagsJson), &tags); err != nil {
			tags = make(map[string]string)
		}
	}

	return &pb.Service{
		NamespaceId:        service.NamespaceId,
		GroupName:          service.GroupName,
		ServiceName:        service.ServiceName,
		ServiceType:        service.ServiceType,
		ServiceVersion:     service.ServiceVersion,
		ServiceDescription: service.ServiceDescription,
		ProtectThreshold:   service.ProtectThreshold,
		Metadata:           metadata,
		Tags:               tags,
		// node 字段在响应中不使用
	}
}

// convertNodeToProto 转换节点为 protobuf 格式
func convertNodeToProto(node *types.ServiceNode) *pb.Node {
	if node == nil {
		return nil
	}
	// 转换 JSON 字符串为 map
	metadata := make(map[string]string)
	if node.MetadataJson != "" {
		if err := json.Unmarshal([]byte(node.MetadataJson), &metadata); err != nil {
			// 解析失败时使用空 map
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
