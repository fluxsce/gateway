package handler

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"gateway/internal/servicecenter/cache"
	"gateway/internal/servicecenter/server/connection"
	pb "gateway/internal/servicecenter/server/proto"
	"gateway/pkg/logger"

	"google.golang.org/grpc/peer"
)

// StreamHandler 统一双向流处理器
// 负责处理服务注册发现和配置中心的所有操作
type StreamHandler struct {
	pb.UnimplementedServiceCenterStreamServer

	// 连接管理
	connectionManager *connection.ConnectionManager

	// 业务处理器
	registryHandler *RegistryHandler
	configHandler   *ConfigHandler
}

// StreamHandlerDeps 统一双向流处理器依赖
type StreamHandlerDeps struct {
	RegistryHandler *RegistryHandler // 服务注册发现处理器（共享实例）
	ConfigHandler   *ConfigHandler   // 配置中心处理器（共享实例）
}

// NewStreamHandler 创建统一双向流处理器
// 注意：使用外部传入的 handler 实例，确保订阅者和事件通知使用同一个管理器
func NewStreamHandler(deps *StreamHandlerDeps) *StreamHandler {
	if deps.RegistryHandler == nil {
		panic("RegistryHandler 不能为空")
	}
	if deps.ConfigHandler == nil {
		panic("ConfigHandler 不能为空")
	}

	return &StreamHandler{
		connectionManager: connection.NewConnectionManager(),
		registryHandler:   deps.RegistryHandler,
		configHandler:     deps.ConfigHandler,
	}
}

// Connect 实现统一双向流 RPC
func (h *StreamHandler) Connect(stream pb.ServiceCenterStream_ConnectServer) error {
	ctx := stream.Context()

	// 从 gRPC 上下文获取客户端 IP
	clientIP := ""
	if p, ok := peer.FromContext(ctx); ok {
		if host, _, err := net.SplitHostPort(p.Addr.String()); err == nil {
			clientIP = host
		}
	}

	logger.Debug("客户端尝试建立双向流连接", "clientIP", clientIP)

	// 创建连接上下文
	conn := h.connectionManager.CreateConnection(ctx, stream, clientIP)

	// 启动连接监控（检测断开）
	go h.monitorConnection(conn)

	// 处理客户端消息
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			logger.Info("客户端正常关闭连接",
				"connectionId", conn.ConnectionID,
				"clientId", conn.ClientID)
			h.connectionManager.RemoveConnection(conn.ConnectionID)
			return nil
		}
		if err != nil {
			logger.Error("接收客户端消息失败", err,
				"connectionId", conn.ConnectionID,
				"clientId", conn.ClientID)
			h.connectionManager.RemoveConnection(conn.ConnectionID)
			return err
		}

		// 路由消息
		if err := h.routeClientMessage(conn, msg); err != nil {
			logger.Error("处理客户端消息失败", err,
				"connectionId", conn.ConnectionID,
				"messageType", msg.GetMessageType().String())
			// 发送错误响应
			h.sendErrorResponse(conn, msg.GetRequestId(), err)
		}
	}
}

// routeClientMessage 路由客户端消息到对应的处理器
func (h *StreamHandler) routeClientMessage(conn *connection.StreamConnection, msg *pb.ClientMessage) error {
	// 记录消息
	logger.Debug("收到客户端消息",
		"connectionId", conn.ConnectionID,
		"messageType", msg.GetMessageType().String(),
		"requestId", msg.GetRequestId())

	// 更新最后活跃时间
	conn.UpdateLastActiveTime()

	// 根据消息类型路由
	switch msg.GetMessageType() {
	// ========== 连接管理 ==========
	case pb.ClientMessageType_CLIENT_HANDSHAKE:
		return h.handleHandshake(conn, msg)

	case pb.ClientMessageType_CLIENT_PING:
		return h.handlePing(conn, msg)

	// ========== 服务注册发现 ==========
	case pb.ClientMessageType_CLIENT_HEARTBEAT:
		return h.handleHeartbeat(conn, msg)

	case pb.ClientMessageType_CLIENT_REGISTER_SERVICE:
		return h.handleRegisterService(conn, msg)

	case pb.ClientMessageType_CLIENT_UNREGISTER_SERVICE:
		return h.handleUnregisterService(conn, msg)

	case pb.ClientMessageType_CLIENT_REGISTER_NODE:
		return h.handleRegisterNode(conn, msg)

	case pb.ClientMessageType_CLIENT_UNREGISTER_NODE:
		return h.handleUnregisterNode(conn, msg)

	case pb.ClientMessageType_CLIENT_DISCOVER_NODES:
		return h.handleDiscoverNodes(conn, msg)

	case pb.ClientMessageType_CLIENT_SUBSCRIBE_SERVICES:
		return h.handleSubscribeServices(conn, msg)

	case pb.ClientMessageType_CLIENT_SUBSCRIBE_NAMESPACE:
		return h.handleSubscribeNamespace(conn, msg)

	// ========== 配置中心 ==========
	case pb.ClientMessageType_CLIENT_GET_CONFIG:
		return h.handleGetConfig(conn, msg)

	case pb.ClientMessageType_CLIENT_SAVE_CONFIG:
		return h.handleSaveConfig(conn, msg)

	case pb.ClientMessageType_CLIENT_DELETE_CONFIG:
		return h.handleDeleteConfig(conn, msg)

	case pb.ClientMessageType_CLIENT_LIST_CONFIGS:
		return h.handleListConfigs(conn, msg)

	case pb.ClientMessageType_CLIENT_WATCH_CONFIG:
		return h.handleWatchConfig(conn, msg)

	case pb.ClientMessageType_CLIENT_GET_CONFIG_HISTORY:
		return h.handleGetConfigHistory(conn, msg)

	case pb.ClientMessageType_CLIENT_ROLLBACK_CONFIG:
		return h.handleRollbackConfig(conn, msg)

	default:
		return fmt.Errorf("不支持的消息类型: %s", msg.GetMessageType().String())
	}
}

// ========== 连接管理处理 ==========

// handleHandshake 处理握手请求
func (h *StreamHandler) handleHandshake(conn *connection.StreamConnection, msg *pb.ClientMessage) error {
	handshake := msg.GetHandshake()
	if handshake == nil {
		return fmt.Errorf("握手消息为空")
	}

	metadata := handshake.GetMetadata()
	if metadata == nil {
		return fmt.Errorf("客户端元数据为空")
	}

	// 更新连接信息
	conn.ClientID = metadata.GetClientId()
	conn.Metadata = metadata
	conn.NamespaceID = handshake.GetNamespaceId()
	conn.SubscribeTypes = handshake.GetSubscribeTypes()

	// 从认证上下文获取租户ID（由 Auth Interceptor 设置）
	tenantId := GetTenantIdFromContext(conn.Context)
	if tenantId == "" {
		tenantId = "default" // 默认租户
	}
	conn.TenantID = tenantId

	logger.Info("客户端握手成功",
		"connectionId", conn.ConnectionID,
		"clientId", conn.ClientID,
		"tenantId", tenantId,
		"clientIP", conn.ClientIP,
		"language", metadata.GetLanguage(),
		"version", metadata.GetClientVersion())

	// 返回握手响应
	response := &pb.ServerMessage{
		RequestId:   msg.GetRequestId(),
		MessageType: pb.ServerMessageType_SERVER_HANDSHAKE,
		Message: &pb.ServerMessage_Handshake{
			Handshake: &pb.ServerHandshake{
				Success:           true,
				Message:           "连接建立成功",
				ConnectionId:      conn.ConnectionID,
				ServerTime:        time.Now().UnixMilli(),
				HeartbeatInterval: 30, // 建议30秒业务心跳
				ServerInfo: map[string]string{
					"version": "1.0.0",
					"region":  "default",
				},
				TenantId: tenantId,
			},
		},
	}

	return conn.Stream.Send(response)
}

// handlePing 处理 Ping 请求
func (h *StreamHandler) handlePing(conn *connection.StreamConnection, msg *pb.ClientMessage) error {
	ping := msg.GetPing()
	if ping == nil {
		return fmt.Errorf("Ping 消息为空")
	}

	// 更新最后 Ping 时间
	conn.UpdateLastPingTime()

	// 返回 Pong
	response := &pb.ServerMessage{
		RequestId:   msg.GetRequestId(),
		MessageType: pb.ServerMessageType_SERVER_PONG,
		Message: &pb.ServerMessage_Pong{
			Pong: &pb.ServerPong{
				Timestamp:       time.Now().UnixMilli(),
				ClientTimestamp: ping.GetTimestamp(),
			},
		},
	}

	return conn.Stream.Send(response)
}

// ========== 服务注册发现处理（委托给 RegistryHandler）==========

// handleHeartbeat 处理业务心跳
func (h *StreamHandler) handleHeartbeat(conn *connection.StreamConnection, msg *pb.ClientMessage) error {
	req := msg.GetHeartbeat()
	if req == nil {
		return fmt.Errorf("心跳请求为空")
	}

	// 调用 RegistryHandler
	resp, err := h.registryHandler.Heartbeat(conn.Context, req)
	if err != nil {
		return err
	}

	// 返回响应
	response := &pb.ServerMessage{
		RequestId:   msg.GetRequestId(),
		MessageType: pb.ServerMessageType_SERVER_HEARTBEAT,
		Message: &pb.ServerMessage_Heartbeat{
			Heartbeat: resp,
		},
	}

	return conn.Stream.Send(response)
}

// handleRegisterService 处理注册服务
func (h *StreamHandler) handleRegisterService(conn *connection.StreamConnection, msg *pb.ClientMessage) error {
	req := msg.GetRegisterService()
	if req == nil {
		return fmt.Errorf("注册服务请求为空")
	}

	// 调用 RegistryHandler
	resp, err := h.registryHandler.RegisterService(conn.Context, req)
	if err != nil {
		return err
	}

	// 如果注册成功且有节点ID，记录到连接
	// 注意：变更通知由 RegistryHandler 统一处理，无需手动广播
	if resp.GetSuccess() && resp.GetNodeId() != "" {
		conn.AddRegisteredNode(resp.GetNodeId())
	}

	// 返回响应
	response := &pb.ServerMessage{
		RequestId:   msg.GetRequestId(),
		MessageType: pb.ServerMessageType_SERVER_REGISTER_SERVICE,
		Message: &pb.ServerMessage_RegisterService{
			RegisterService: resp,
		},
	}

	return conn.Stream.Send(response)
}

// handleUnregisterService 处理注销服务
func (h *StreamHandler) handleUnregisterService(conn *connection.StreamConnection, msg *pb.ClientMessage) error {
	req := msg.GetUnregisterService()
	if req == nil {
		return fmt.Errorf("注销服务请求为空")
	}

	// 调用 RegistryHandler
	resp, err := h.registryHandler.UnregisterService(conn.Context, req)
	if err != nil {
		return err
	}

	// 如果注销成功且有节点ID，从连接中移除
	if resp.GetSuccess() && req.GetNodeId() != "" {
		conn.RemoveRegisteredNode(req.GetNodeId())
	}

	// 返回响应
	response := &pb.ServerMessage{
		RequestId:   msg.GetRequestId(),
		MessageType: pb.ServerMessageType_SERVER_UNREGISTER_SERVICE,
		Message: &pb.ServerMessage_UnregisterService{
			UnregisterService: resp,
		},
	}

	return conn.Stream.Send(response)
}

// handleRegisterNode 处理注册节点
func (h *StreamHandler) handleRegisterNode(conn *connection.StreamConnection, msg *pb.ClientMessage) error {
	req := msg.GetRegisterNode()
	if req == nil {
		return fmt.Errorf("注册节点请求为空")
	}

	// 调用 RegistryHandler
	resp, err := h.registryHandler.RegisterNode(conn.Context, req)
	if err != nil {
		return err
	}

	// 如果注册成功，记录节点ID到连接
	// 注意：变更通知由 RegistryHandler 统一处理，无需手动广播
	if resp.GetSuccess() && resp.GetNodeId() != "" {
		conn.AddRegisteredNode(resp.GetNodeId())
		logger.Info("节点注册成功",
			"connectionId", conn.ConnectionID,
			"nodeId", resp.GetNodeId(),
			"serviceName", req.GetServiceName())
	}

	// 返回响应
	response := &pb.ServerMessage{
		RequestId:   msg.GetRequestId(),
		MessageType: pb.ServerMessageType_SERVER_REGISTER_NODE,
		Message: &pb.ServerMessage_RegisterNode{
			RegisterNode: resp,
		},
	}

	return conn.Stream.Send(response)
}

// handleUnregisterNode 处理注销节点
func (h *StreamHandler) handleUnregisterNode(conn *connection.StreamConnection, msg *pb.ClientMessage) error {
	req := msg.GetUnregisterNode()
	if req == nil {
		return fmt.Errorf("注销节点请求为空")
	}

	// 调用 RegistryHandler
	// 注意：变更通知由 RegistryHandler 统一处理，无需手动广播
	resp, err := h.registryHandler.UnregisterNode(conn.Context, req)
	if err != nil {
		return err
	}

	// 如果注销成功，从连接中移除节点ID
	if resp.GetSuccess() {
		conn.RemoveRegisteredNode(req.GetNodeId())
		logger.Info("节点注销成功",
			"connectionId", conn.ConnectionID,
			"nodeId", req.GetNodeId())
	}

	// 返回响应
	response := &pb.ServerMessage{
		RequestId:   msg.GetRequestId(),
		MessageType: pb.ServerMessageType_SERVER_UNREGISTER_NODE,
		Message: &pb.ServerMessage_UnregisterNode{
			UnregisterNode: resp,
		},
	}

	return conn.Stream.Send(response)
}

// handleDiscoverNodes 处理发现节点
func (h *StreamHandler) handleDiscoverNodes(conn *connection.StreamConnection, msg *pb.ClientMessage) error {
	req := msg.GetDiscoverNodes()
	if req == nil {
		return fmt.Errorf("发现节点请求为空")
	}

	// 调用 RegistryHandler
	resp, err := h.registryHandler.DiscoverNodes(conn.Context, req)
	if err != nil {
		return err
	}

	// 返回响应
	response := &pb.ServerMessage{
		RequestId:   msg.GetRequestId(),
		MessageType: pb.ServerMessageType_SERVER_DISCOVER_NODES,
		Message: &pb.ServerMessage_DiscoverNodes{
			DiscoverNodes: resp,
		},
	}

	return conn.Stream.Send(response)
}

// handleSubscribeServices 处理订阅服务
func (h *StreamHandler) handleSubscribeServices(conn *connection.StreamConnection, msg *pb.ClientMessage) error {
	req := msg.GetSubscribeServices()
	if req == nil {
		return fmt.Errorf("订阅服务请求为空")
	}

	// 验证参数
	namespaceId := req.GetNamespaceId()
	groupName := req.GetGroupName()
	serviceNames := req.GetServiceNames()

	if namespaceId == "" {
		return fmt.Errorf("命名空间ID不能为空")
	}
	if groupName == "" {
		return fmt.Errorf("服务组名称不能为空")
	}
	if len(serviceNames) == 0 {
		return fmt.Errorf("服务名称列表不能为空")
	}

	// 设置默认值
	if groupName == "" {
		groupName = "DEFAULT_GROUP"
	}

	// 生成订阅ID（基于连接ID）
	subscriberId := conn.ConnectionID + "_" + msg.GetRequestId()

	// 记录订阅信息到连接（用于查询和清理）
	conn.AddServiceSubscription(namespaceId, groupName, serviceNames)

	logger.Info("客户端订阅服务",
		"connectionId", conn.ConnectionID,
		"clientId", conn.ClientID,
		"subscriberId", subscriberId,
		"namespaceId", namespaceId,
		"groupName", groupName,
		"serviceNames", serviceNames,
		"serviceCount", len(serviceNames))

	// 通过 serviceSubMgr 订阅服务（统一通知机制）
	tenantId := conn.TenantID
	if tenantId == "" {
		tenantId = "default"
	}

	ch := h.registryHandler.GetServiceSubscriber().SubscribeMultipleServices(
		conn.Context,
		tenantId,
		namespaceId,
		groupName,
		serviceNames,
		subscriberId,
	)

	// 立即推送当前缓存中的服务状态（初始快照）
	go func() {
		globalCache := cache.GetGlobalCache()
		for _, serviceName := range serviceNames {
			// 获取服务及其节点
			service, found := globalCache.GetService(conn.Context, tenantId, namespaceId, groupName, serviceName)
			if !found || service == nil {
				logger.Debug("订阅的服务不存在，跳过初始推送",
					"serviceName", serviceName,
					"namespaceId", namespaceId,
					"groupName", groupName)
				continue
			}

			// 构建初始服务状态事件
			event := &pb.ServiceChangeEvent{
				EventType:   "SERVICE_ADDED", // 初始状态视为服务添加
				Timestamp:   time.Now().Format("2006-01-02 15:04:05"),
				NamespaceId: namespaceId,
				GroupName:   groupName,
				ServiceName: serviceName,
				Service: &pb.Service{
					NamespaceId:        service.NamespaceId,
					GroupName:          service.GroupName,
					ServiceName:        service.ServiceName,
					ServiceType:        service.ServiceType,
					ServiceDescription: service.ServiceDescription,
				},
			}

			// 添加节点列表
			if len(service.Nodes) > 0 {
				event.Nodes = make([]*pb.Node, 0, len(service.Nodes))
				for _, node := range service.Nodes {
					event.Nodes = append(event.Nodes, &pb.Node{
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
					})
				}
			}

			// 发送到订阅者的 channel（由主循环统一处理，避免并发调用 conn.Send）
			h.registryHandler.GetServiceSubscriber().SendToSubscriber(subscriberId, event)

			logger.Debug("已推送初始服务快照到 channel",
				"connectionId", conn.ConnectionID,
				"subscriberId", subscriberId,
				"serviceName", serviceName,
				"nodeCount", len(service.Nodes))
		}
	}()

	// 启动 goroutine 接收后续变更事件并转发到连接
	go func() {
		defer func() {
			// 连接断开时取消订阅
			h.registryHandler.GetServiceSubscriber().UnsubscribeMultipleServices(subscriberId)
			logger.Info("取消服务订阅",
				"connectionId", conn.ConnectionID,
				"subscriberId", subscriberId)
		}()

		for {
			select {
			case event, ok := <-ch:
				if !ok {
					// Channel 已关闭
					return
				}

				// 转换为 ServerMessage 格式并发送
				serverMsg := &pb.ServerMessage{
					MessageType: pb.ServerMessageType_SERVER_SERVICE_CHANGE,
					Message: &pb.ServerMessage_ServiceChange{
						ServiceChange: event,
					},
				}

				if err := conn.Send(serverMsg); err != nil {
					logger.Error("推送服务变更事件失败", err,
						"connectionId", conn.ConnectionID,
						"subscriberId", subscriberId)
					return
				}

				logger.Debug("推送服务变更事件",
					"connectionId", conn.ConnectionID,
					"subscriberId", subscriberId,
					"eventType", event.EventType,
					"serviceName", event.ServiceName)

			case <-conn.Context.Done():
				// 连接已断开
				return
			}
		}
	}()

	// 注意：订阅是单向的，客户端不需要等待响应
	// 后续服务变更时，会通过 serviceSubMgr 统一推送
	return nil
}

// handleSubscribeNamespace 处理订阅命名空间
func (h *StreamHandler) handleSubscribeNamespace(conn *connection.StreamConnection, msg *pb.ClientMessage) error {
	req := msg.GetSubscribeNamespace()
	if req == nil {
		return fmt.Errorf("订阅命名空间请求为空")
	}

	// 验证参数
	namespaceId := req.GetNamespaceId()
	groupName := req.GetGroupName()

	if namespaceId == "" {
		return fmt.Errorf("命名空间ID不能为空")
	}

	// 订阅整个命名空间或指定组
	// 如果 groupName 为空，订阅整个命名空间下的所有服务
	// 如果 groupName 不为空，订阅该组下的所有服务
	conn.AddServiceSubscription(namespaceId, groupName, nil) // nil 表示订阅所有服务

	logger.Info("客户端订阅命名空间成功",
		"connectionId", conn.ConnectionID,
		"clientId", conn.ClientID,
		"namespaceId", namespaceId,
		"groupName", groupName)

	// 注意：订阅是单向的，客户端不需要等待响应
	return nil
}

// ========== 配置中心处理（委托给 ConfigHandler）==========

// handleGetConfig 处理获取配置
func (h *StreamHandler) handleGetConfig(conn *connection.StreamConnection, msg *pb.ClientMessage) error {
	req := msg.GetGetConfig()
	if req == nil {
		return fmt.Errorf("获取配置请求为空")
	}

	// 调用 ConfigHandler
	resp, err := h.configHandler.GetConfig(conn.Context, req)
	if err != nil {
		return err
	}

	// 返回响应
	response := &pb.ServerMessage{
		RequestId:   msg.GetRequestId(),
		MessageType: pb.ServerMessageType_SERVER_GET_CONFIG,
		Message: &pb.ServerMessage_GetConfig{
			GetConfig: resp,
		},
	}

	return conn.Stream.Send(response)
}

// handleSaveConfig 处理保存配置
func (h *StreamHandler) handleSaveConfig(conn *connection.StreamConnection, msg *pb.ClientMessage) error {
	req := msg.GetSaveConfig()
	if req == nil {
		return fmt.Errorf("保存配置请求为空")
	}

	// 调用 ConfigHandler
	// 注意：配置变更通知由 ConfigHandler 内部通过 configWatcher 统一处理
	resp, err := h.configHandler.SaveConfig(conn.Context, req)
	if err != nil {
		return err
	}

	// 返回响应
	response := &pb.ServerMessage{
		RequestId:   msg.GetRequestId(),
		MessageType: pb.ServerMessageType_SERVER_SAVE_CONFIG,
		Message: &pb.ServerMessage_SaveConfig{
			SaveConfig: resp,
		},
	}

	return conn.Stream.Send(response)
}

// handleDeleteConfig 处理删除配置
func (h *StreamHandler) handleDeleteConfig(conn *connection.StreamConnection, msg *pb.ClientMessage) error {
	req := msg.GetDeleteConfig()
	if req == nil {
		return fmt.Errorf("删除配置请求为空")
	}

	// 调用 ConfigHandler
	// 注意：配置变更通知由 ConfigHandler 内部通过 configWatcher 统一处理
	resp, err := h.configHandler.DeleteConfig(conn.Context, req)
	if err != nil {
		return err
	}

	// 返回响应
	response := &pb.ServerMessage{
		RequestId:   msg.GetRequestId(),
		MessageType: pb.ServerMessageType_SERVER_DELETE_CONFIG,
		Message: &pb.ServerMessage_DeleteConfig{
			DeleteConfig: resp,
		},
	}

	return conn.Stream.Send(response)
}

// handleListConfigs 处理列出配置
func (h *StreamHandler) handleListConfigs(conn *connection.StreamConnection, msg *pb.ClientMessage) error {
	req := msg.GetListConfigs()
	if req == nil {
		return fmt.Errorf("列出配置请求为空")
	}

	// 调用 ConfigHandler
	resp, err := h.configHandler.ListConfigs(conn.Context, req)
	if err != nil {
		return err
	}

	// 返回响应
	response := &pb.ServerMessage{
		RequestId:   msg.GetRequestId(),
		MessageType: pb.ServerMessageType_SERVER_LIST_CONFIGS,
		Message: &pb.ServerMessage_ListConfigs{
			ListConfigs: resp,
		},
	}

	return conn.Stream.Send(response)
}

// handleWatchConfig 处理监听配置
func (h *StreamHandler) handleWatchConfig(conn *connection.StreamConnection, msg *pb.ClientMessage) error {
	req := msg.GetWatchConfig()
	if req == nil {
		return fmt.Errorf("监听配置请求为空")
	}

	// 验证参数
	namespaceId := req.GetNamespaceId()
	groupName := req.GetGroupName()
	configDataIds := req.GetConfigDataIds()

	if namespaceId == "" {
		return fmt.Errorf("命名空间ID不能为空")
	}
	if groupName == "" {
		return fmt.Errorf("配置组名称不能为空")
	}
	if len(configDataIds) == 0 {
		return fmt.Errorf("配置ID列表不能为空")
	}

	// 生成监听ID（基于连接ID）
	watcherId := conn.ConnectionID + "_" + msg.GetRequestId()

	// 记录监听信息到连接（用于查询和清理）
	conn.AddConfigWatch(namespaceId, groupName, configDataIds)

	logger.Info("客户端监听配置",
		"connectionId", conn.ConnectionID,
		"clientId", conn.ClientID,
		"watcherId", watcherId,
		"namespaceId", namespaceId,
		"groupName", groupName,
		"configDataIds", configDataIds,
		"configCount", len(configDataIds))

	// 通过 configWatcher 监听配置（统一通知机制）
	tenantId := conn.TenantID
	if tenantId == "" {
		tenantId = "default"
	}

	// 监听所有配置（共享通道）
	ch := h.configHandler.GetConfigWatcher().Watch(
		conn.Context,
		tenantId,
		namespaceId,
		groupName,
		configDataIds,
		watcherId,
	)

	// 立即推送当前配置快照
	go func() {
		for _, configDataId := range configDataIds {
			// 获取配置内容
			getReq := &pb.ConfigKey{
				NamespaceId:  namespaceId,
				GroupName:    groupName,
				ConfigDataId: configDataId,
			}

			configResp, err := h.configHandler.GetConfig(conn.Context, getReq)
			if err != nil {
				logger.Debug("获取配置快照失败，跳过初始推送",
					"configDataId", configDataId,
					"error", err)
				continue
			}

			if configResp == nil || !configResp.GetSuccess() || configResp.GetConfig() == nil {
				logger.Debug("配置不存在，跳过初始推送",
					"configDataId", configDataId,
					"namespaceId", namespaceId,
					"groupName", groupName)
				continue
			}

			// 构建初始配置状态事件
			event := &pb.ConfigChangeEvent{
				EventType:    "UPDATED", // 初始状态视为已存在的配置
				Timestamp:    time.Now().Format("2006-01-02 15:04:05"),
				NamespaceId:  namespaceId,
				GroupName:    groupName,
				ConfigDataId: configDataId,
				Config:       configResp.GetConfig(), // 完整的配置数据对象
				ContentMd5:   configResp.GetConfig().GetContentMd5(),
			}

			// 发送到监听者的 channel（由主循环统一处理，避免并发调用 conn.Send）
			h.configHandler.GetConfigWatcher().SendToWatcher(watcherId, event)

			logger.Debug("已推送初始配置快照到 channel",
				"connectionId", conn.ConnectionID,
				"watcherId", watcherId,
				"configDataId", configDataId,
				"contentMd5", configResp.GetConfig().GetContentMd5())
		}
	}()

	// 启动 goroutine 接收后续变更事件并转发到连接
	go func() {
		defer func() {
			// 连接断开时取消监听
			h.configHandler.GetConfigWatcher().Unwatch(watcherId)
			logger.Info("取消配置监听",
				"connectionId", conn.ConnectionID,
				"watcherId", watcherId)
		}()

		for {
			select {
			case event, ok := <-ch:
				if !ok {
					// Channel 已关闭
					return
				}

				// 转换为 ServerMessage 格式并发送
				serverMsg := &pb.ServerMessage{
					MessageType: pb.ServerMessageType_SERVER_CONFIG_CHANGE,
					Message: &pb.ServerMessage_ConfigChange{
						ConfigChange: event,
					},
				}

				if err := conn.Send(serverMsg); err != nil {
					logger.Error("推送配置变更事件失败", err,
						"connectionId", conn.ConnectionID,
						"watcherId", watcherId)
					return
				}

				logger.Debug("推送配置变更事件",
					"connectionId", conn.ConnectionID,
					"watcherId", watcherId,
					"configDataId", event.ConfigDataId,
					"eventType", event.EventType)

			case <-conn.Context.Done():
				// 连接已断开
				return
			}
		}
	}()

	// 注意：监听是单向的，客户端不需要等待响应
	// 后续配置变更时，会通过 configWatcher 统一推送
	return nil
}

// handleGetConfigHistory 处理获取配置历史
func (h *StreamHandler) handleGetConfigHistory(conn *connection.StreamConnection, msg *pb.ClientMessage) error {
	req := msg.GetGetConfigHistory()
	if req == nil {
		return fmt.Errorf("获取配置历史请求为空")
	}

	// 调用 ConfigHandler
	resp, err := h.configHandler.GetConfigHistory(conn.Context, req)
	if err != nil {
		return err
	}

	// 返回响应
	response := &pb.ServerMessage{
		RequestId:   msg.GetRequestId(),
		MessageType: pb.ServerMessageType_SERVER_GET_CONFIG_HISTORY,
		Message: &pb.ServerMessage_GetConfigHistory{
			GetConfigHistory: resp,
		},
	}

	return conn.Stream.Send(response)
}

// handleRollbackConfig 处理回滚配置
func (h *StreamHandler) handleRollbackConfig(conn *connection.StreamConnection, msg *pb.ClientMessage) error {
	req := msg.GetRollbackConfig()
	if req == nil {
		return fmt.Errorf("回滚配置请求为空")
	}

	// 调用 ConfigHandler
	resp, err := h.configHandler.RollbackConfig(conn.Context, req)
	if err != nil {
		return err
	}

	// 返回响应
	response := &pb.ServerMessage{
		RequestId:   msg.GetRequestId(),
		MessageType: pb.ServerMessageType_SERVER_ROLLBACK_CONFIG,
		Message: &pb.ServerMessage_RollbackConfig{
			RollbackConfig: resp,
		},
	}

	return conn.Stream.Send(response)
}

// ========== 辅助方法 ==========

// sendErrorResponse 发送错误响应
func (h *StreamHandler) sendErrorResponse(conn *connection.StreamConnection, requestId string, err error) {
	response := &pb.ServerMessage{
		RequestId:   requestId,
		MessageType: pb.ServerMessageType_SERVER_ERROR,
		Message: &pb.ServerMessage_Error{
			Error: &pb.ErrorResponse{
				Code:    "INTERNAL_ERROR",
				Message: err.Error(),
			},
		},
	}

	if err := conn.Stream.Send(response); err != nil {
		logger.Error("发送错误响应失败", err,
			"connectionId", conn.ConnectionID,
			"requestId", requestId)
	}
}

// monitorConnection 监控连接状态
func (h *StreamHandler) monitorConnection(conn *connection.StreamConnection) {
	<-conn.Context.Done()

	// 连接断开，清理资源
	logger.Info("检测到连接断开，开始清理资源",
		"connectionId", conn.ConnectionID,
		"clientId", conn.ClientID,
		"clientIP", conn.ClientIP)

	h.cleanupConnection(conn)
}

// cleanupConnection 清理连接资源
func (h *StreamHandler) cleanupConnection(conn *connection.StreamConnection) {
	logger.Info("开始清理连接资源",
		"connectionId", conn.ConnectionID,
		"clientId", conn.ClientID,
		"registeredNodesCount", len(conn.GetRegisteredNodes()))

	// 1. 注销所有已注册的节点
	// 注意：变更通知由 RegistryHandler 统一处理，无需手动广播
	for _, nodeId := range conn.GetRegisteredNodes() {
		// 获取节点信息用于日志记录
		nodeInfo, exists := cache.GetGlobalCache().GetNode(context.Background(), conn.TenantID, nodeId)
		serviceName := ""
		if exists && nodeInfo != nil {
			serviceName = nodeInfo.ServiceName
		}

		req := &pb.NodeKey{NodeId: nodeId}
		resp, err := h.registryHandler.UnregisterNode(context.Background(), req)
		if err != nil {
			logger.Error("自动注销节点失败", err,
				"connectionId", conn.ConnectionID,
				"serviceName", serviceName,
				"nodeId", nodeId)
			continue
		}
		if resp.GetSuccess() {
			logger.Info("清理服务节点",
				"connectionId", conn.ConnectionID,
				"serviceName", serviceName,
				"nodeId", nodeId)
		} else {
			logger.Error("自动注销节点失败", fmt.Errorf("%s", resp.GetMessage()),
				"connectionId", conn.ConnectionID,
				"serviceName", serviceName,
				"nodeId", nodeId)
		}
	}

	// 2. 从连接管理器中移除
	h.connectionManager.RemoveConnection(conn.ConnectionID)

	logger.Info("连接资源清理完成",
		"connectionId", conn.ConnectionID,
		"clientId", conn.ClientID)
}

// GetTenantIdFromContext 从上下文获取租户ID
// （由 Auth Interceptor 设置）
func GetTenantIdFromContext(ctx context.Context) string {
	if tenantId, ok := ctx.Value("tenantId").(string); ok {
		return tenantId
	}
	return ""
}
