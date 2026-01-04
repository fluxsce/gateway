// Package server 提供控制连接处理的实现
// 控制连接负责处理客户端的认证、心跳、服务注册等
// 本文件包含 DefaultTunnelServer 的控制连接相关方法
package server

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"gateway/internal/tunnel/types"
	"gateway/pkg/logger"
)

// acceptConnections 接受新连接的循环
func (s *DefaultTunnelServer) acceptConnections() {
	defer s.wg.Done()

	for {
		// Accept 连接
		conn, err := s.controlListener.Accept()

		if err != nil {
			select {
			case <-s.ctx.Done():
				return
			default:
				logger.Error("Failed to accept connection", map[string]interface{}{
					"error": err.Error(),
				})
				continue
			}
		}

		// 立即启动 goroutine 处理连接，避免阻塞 Accept 循环
		s.wg.Add(1)
		go func(conn net.Conn) {
			defer s.wg.Done()
			// 使用独立的上下文，避免单个连接处理阻塞影响其他连接
			connCtx, cancel := context.WithCancel(s.ctx)
			defer cancel()
			if err := s.handleConnection(connCtx, conn); err != nil {
				logger.Error("Connection handling failed", map[string]interface{}{
					"error": err.Error(),
				})
			}
		}(conn)
	}
}

// heartbeatChecker 心跳检查器
func (s *DefaultTunnelServer) heartbeatChecker() {
	defer s.wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.checkHeartbeats()
		}
	}
}

// checkHeartbeats 检查所有连接的心跳状态
func (s *DefaultTunnelServer) checkHeartbeats() {
	s.mutex.RLock()
	var expiredClients []*types.TunnelClient
	now := time.Now()

	for _, client := range s.connectedClients {
		if now.Sub(client.LastActivityTime) > s.heartbeatTimeout {
			expiredClients = append(expiredClients, client)
		}
	}
	s.mutex.RUnlock()

	// 关闭过期连接
	for _, client := range expiredClients {
		logger.Warn("Connection timeout, closing", map[string]interface{}{
			"clientId": client.TunnelClientId,
		})
		// 从连接管理中获取连接并关闭
		s.mutex.RLock()
		if clientConn, exists := s.clientConnections[client.TunnelClientId]; exists && clientConn.conn != nil {
			clientConn.conn.Close()
		}
		s.mutex.RUnlock()
	}
}

// handleConnection 处理单个客户端连接
// 这是所有连接的入口点，负责：
// 1. 读取并验证第一条消息
// 2. 根据消息类型路由到相应的处理函数
// 3. 严格验证只允许 auth 或 data_connection 作为第一条消息
func (s *DefaultTunnelServer) handleConnection(ctx context.Context, conn net.Conn) error {
	// 使用心跳超时的一半作为初始读取超时，快速检测无效连接
	initialTimeout := s.heartbeatTimeout / 2
	if initialTimeout < 10*time.Second {
		initialTimeout = 10 * time.Second
	}
	conn.SetReadDeadline(time.Now().Add(initialTimeout))

	// 读取消息长度（4字节大端序）
	lengthBuf := make([]byte, 4)
	if _, err := io.ReadFull(conn, lengthBuf); err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			logger.Warn("Connection timeout reading handshake", map[string]interface{}{
				"remoteAddr": conn.RemoteAddr().String(),
				"localAddr":  conn.LocalAddr().String(),
				"timeout":    initialTimeout.String(),
			})
		}
		return fmt.Errorf("failed to read message length: %w", err)
	}

	// 延长超时时间用于读取消息内容（使用完整的心跳超时）
	conn.SetReadDeadline(time.Now().Add(s.heartbeatTimeout))

	// 解析消息长度
	msgLen := int(lengthBuf[0])<<24 | int(lengthBuf[1])<<16 | int(lengthBuf[2])<<8 | int(lengthBuf[3])

	// 消息长度验证
	if err := s.validateMessageLength(msgLen, lengthBuf, conn); err != nil {
		return err
	}

	// 读取消息内容
	msgBuf := make([]byte, msgLen)
	if _, err := io.ReadFull(conn, msgBuf); err != nil {
		return fmt.Errorf("failed to read message data: %w", err)
	}

	// 尝试解析消息以判断连接类型
	var firstMsg map[string]interface{}
	if err := json.Unmarshal(msgBuf, &firstMsg); err != nil {
		logger.Error("Invalid first message format", map[string]interface{}{
			"error":      err.Error(),
			"remoteAddr": conn.RemoteAddr().String(),
		})
		return fmt.Errorf("failed to parse first message: %w", err)
	}

	// 获取消息类型
	msgType, ok := firstMsg["type"].(string)
	if !ok || msgType == "" {
		logger.Warn("First message missing type field", map[string]interface{}{
			"remoteAddr": conn.RemoteAddr().String(),
			"message":    firstMsg,
		})
		return fmt.Errorf("first message missing type field")
	}

	// 严格验证第一条消息类型
	// 安全策略：只允许以下类型作为第一条消息
	// 1. MessageTypeAuth ("auth") - 控制连接的认证消息 [Client → Server]
	// 2. MessageTypeDataConnection ("data_connection") - 数据连接的握手消息 [Client → Server]
	// 任何其他类型的消息都会被拒绝并关闭连接
	switch msgType {
	case types.MessageTypeDataConnection:
		// 处理数据连接握手 [Client → Server]
		// 数据连接用于实际的流量转发，不需要认证（通过 connectionId 关联）
		return s.handleDataConnection(ctx, conn, firstMsg)

	case types.MessageTypeAuth:
		// 处理控制连接认证 [Client → Server]
		// 控制连接用于客户端注册、服务管理、心跳等控制消息
		return s.handleControlConnection(ctx, conn, msgBuf)

	default:
		// 拒绝其他任何类型的第一条消息
		logger.Warn("Invalid first message type, connection rejected", map[string]interface{}{
			"messageType": msgType,
			"remoteAddr":  conn.RemoteAddr().String(),
			"expected":    "auth or data_connection",
		})
		return fmt.Errorf("invalid first message type: %s, expected 'auth' or 'data_connection'", msgType)
	}
}

// validateMessageLength 验证消息长度
func (s *DefaultTunnelServer) validateMessageLength(msgLen int, lengthBuf []byte, conn net.Conn) error {
	// 检查长度是否全为0
	allZeros := lengthBuf[0] == 0 && lengthBuf[1] == 0 && lengthBuf[2] == 0 && lengthBuf[3] == 0
	if allZeros {
		logger.Warn("Message length is all zeros", map[string]interface{}{
			"remoteAddr":  conn.RemoteAddr().String(),
			"lengthBytes": lengthBuf,
		})
		return fmt.Errorf("message length is all zeros")
	}

	// 检查长度是否太小
	if msgLen < 10 {
		logger.Error("Message length too small", map[string]interface{}{
			"messageLength": msgLen,
			"lengthBytes":   lengthBuf,
			"remoteAddr":    conn.RemoteAddr().String(),
		})
		return fmt.Errorf("message length too small: %d", msgLen)
	}

	// 检查长度是否过大
	if msgLen > 1024*1024 {
		logger.Error("Invalid message length", map[string]interface{}{
			"messageLength": msgLen,
			"lengthBytes":   lengthBuf,
			"remoteAddr":    conn.RemoteAddr().String(),
		})
		return fmt.Errorf("invalid message length: %d", msgLen)
	}

	return nil
}

// handleDataConnection 处理数据连接
func (s *DefaultTunnelServer) handleDataConnection(ctx context.Context, conn net.Conn, handshake map[string]interface{}) error {
	connectionID, ok := handshake["connectionId"].(string)
	if !ok {
		conn.Close()
		return fmt.Errorf("missing connectionId in data connection handshake")
	}

	clientID, ok := handshake["clientId"].(string)
	if !ok || clientID == "" {
		conn.Close()
		return fmt.Errorf("missing or empty clientId in data connection handshake")
	}

	logger.Info("Data connection received", map[string]interface{}{
		"connectionId": connectionID,
		"clientId":     clientID,
	})

	// 清除读取超时设置
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetReadDeadline(time.Time{})
		tcpConn.SetWriteDeadline(time.Time{})
	}

	// 将数据连接转发给代理服务器处理
	if s.proxyServer == nil {
		conn.Close()
		return fmt.Errorf("proxy server not configured")
	}

	return s.proxyServer.HandleClientDataConnection(ctx, conn, connectionID, clientID)
}

// handleControlConnection 处理控制连接
func (s *DefaultTunnelServer) handleControlConnection(ctx context.Context, conn net.Conn, firstMsgBuf []byte) error {
	// 创建写入器
	writer := bufio.NewWriter(conn)

	// 临时客户端ID，认证后更新
	tempClientID := ""

	defer func() {
		conn.Close()
		// 注销客户端（如果已认证）
		if tempClientID != "" {
			s.UnregisterClient(context.Background(), tempClientID)
		}
	}()

	// 处理第一条消息（已经读取的）
	var firstMsg types.ControlMessage
	if err := json.Unmarshal(firstMsgBuf, &firstMsg); err != nil {
		return fmt.Errorf("failed to parse first control message: %w", err)
	}

	// 处理第一条消息（认证）
	clientID, err := s.processFirstControlMessage(conn, writer, &firstMsg)
	if err != nil {
		return fmt.Errorf("failed to process first control message: %w", err)
	}
	tempClientID = clientID

	// 设置读取超时
	conn.SetReadDeadline(time.Now().Add(s.heartbeatTimeout))

	// 继续消息处理循环
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := s.handleMessage(conn, writer, clientID); err != nil {
				if err == io.EOF {
					logger.Info("Client disconnected", map[string]interface{}{
						"clientId": clientID,
					})
					return nil
				}
				logger.Error("Error handling message", map[string]interface{}{
					"error":    err.Error(),
					"clientId": clientID,
				})
				return err
			}
		}
	}
}

// processFirstControlMessage 处理第一条控制消息（认证）
// 消息类型: MessageTypeAuth
// 传输方向: Client → Server
// 请求结构: AuthRequest (包含完整的 TunnelClient 对象)
// 响应结构: AuthResponse
// 功能说明: 客户端连接建立后的第一条消息，用于身份验证
//
//	验证成功后，服务端将客户端注册到 connectedClients
//	并设置运行时字段（连接、会话、互斥锁等）
//
// 安全策略: 此函数只能被 handleControlConnection 调用，
//
//	且 handleConnection 已经验证过消息类型必须是 auth
func (s *DefaultTunnelServer) processFirstControlMessage(conn net.Conn, writer *bufio.Writer, msg *types.ControlMessage) (string, error) {
	// 双重验证：确保消息类型是 auth（防御性编程）
	if msg.Type != types.MessageTypeAuth {
		if err := s.sendResponse(conn, writer, "", msg, false, "first message must be auth"); err != nil {
			logger.Error("Failed to send auth error response", map[string]interface{}{
				"error":  err.Error(),
				"reason": "invalid message type",
			})
		}
		return "", fmt.Errorf("first message must be auth, got: %s", msg.Type)
	}

	// 解析认证请求（包含完整的 TunnelClient 对象）
	var req types.AuthRequest
	if err := parseMessageData(msg.Data, &req); err != nil {
		logger.Error("Failed to parse auth request", map[string]interface{}{
			"error": err.Error(),
		})
		if sendErr := s.sendResponse(conn, writer, "", msg, false, fmt.Sprintf("failed to parse auth request: %v", err)); sendErr != nil {
			logger.Error("Failed to send auth error response", map[string]interface{}{
				"error":  sendErr.Error(),
				"reason": "parse error",
			})
		}
		return "", fmt.Errorf("failed to parse auth request: %w", err)
	}

	// 获取客户端对象
	client := req.Client

	// 验证必填字段
	if client.TunnelClientId == "" {
		if err := s.sendResponse(conn, writer, "", msg, false, "missing clientId"); err != nil {
			logger.Error("Failed to send auth error response", map[string]interface{}{
				"error":  err.Error(),
				"reason": "missing clientId",
			})
		}
		return "", fmt.Errorf("missing clientId")
	}

	if client.ClientName == "" {
		if err := s.sendResponse(conn, writer, client.TunnelClientId, msg, false, "missing clientName"); err != nil {
			logger.Error("Failed to send auth error response", map[string]interface{}{
				"error":    err.Error(),
				"clientId": client.TunnelClientId,
				"reason":   "missing clientName",
			})
		}
		return "", fmt.Errorf("missing clientName")
	}

	if client.AuthToken == "" {
		if err := s.sendResponse(conn, writer, client.TunnelClientId, msg, false, "missing token"); err != nil {
			logger.Error("Failed to send auth error response", map[string]interface{}{
				"error":    err.Error(),
				"clientId": client.TunnelClientId,
				"reason":   "missing token",
			})
		}
		return "", fmt.Errorf("missing token")
	}

	// 验证令牌
	if s.config.AuthToken != client.AuthToken {
		// 关键修复：确保认证失败响应被发送，避免客户端等待超时
		if err := s.sendResponse(conn, writer, client.TunnelClientId, msg, false, "invalid token"); err != nil {
			logger.Error("Failed to send auth error response", map[string]interface{}{
				"error":      err.Error(),
				"clientId":   client.TunnelClientId,
				"clientName": client.ClientName,
				"reason":     "invalid token",
			})
		} else {
			logger.Info("Auth error response sent successfully", map[string]interface{}{
				"clientId":   client.TunnelClientId,
				"clientName": client.ClientName,
				"reason":     "invalid token",
			})
		}
		return "", fmt.Errorf("invalid token")
	}

	// 设置运行时状态字段
	client.Authenticated = true
	client.LastActivityTime = time.Now()

	// 初始化服务列表
	if client.Services == nil {
		client.Services = make(map[string]*types.TunnelService)
	}

	// 注册客户端
	if err := s.RegisterClient(context.Background(), &client); err != nil {
		s.sendResponse(conn, writer, client.TunnelClientId, msg, false, err.Error())
		return "", err
	}

	// 注册客户端连接（独立管理连接、写入器、互斥锁）
	s.mutex.Lock()
	s.clientConnections[client.TunnelClientId] = &clientConnection{
		conn:   conn,
		writer: writer,
		mutex:  sync.Mutex{},
	}
	s.mutex.Unlock()

	// 返回认证成功响应 [Server → Client]
	// 响应类型: AuthResponse
	s.sendResponse(conn, writer, client.TunnelClientId, msg, true, "authenticated successfully")
	return client.TunnelClientId, nil
}

// handleMessage 处理控制消息
func (s *DefaultTunnelServer) handleMessage(conn net.Conn, writer *bufio.Writer, clientID string) error {
	// 读取消息长度
	lengthBuf := make([]byte, 4)
	if _, err := io.ReadFull(conn, lengthBuf); err != nil {
		return err
	}

	// 解析消息长度
	msgLen := int(lengthBuf[0])<<24 | int(lengthBuf[1])<<16 | int(lengthBuf[2])<<8 | int(lengthBuf[3])

	// 消息长度验证
	if err := s.validateMessageLength(msgLen, lengthBuf, conn); err != nil {
		return err
	}

	// 读取消息内容
	msgBuf := make([]byte, msgLen)
	if _, err := io.ReadFull(conn, msgBuf); err != nil {
		return err
	}

	// 解析控制消息
	var msg types.ControlMessage
	if err := json.Unmarshal(msgBuf, &msg); err != nil {
		return fmt.Errorf("failed to parse control message: %w", err)
	}

	return s.processControlMessage(conn, writer, clientID, &msg)
}

// processControlMessage 处理控制消息
// 根据消息类型路由到相应的处理函数
// 所有消息都必须在认证后才能处理（由 handleControlConnection 保证）
func (s *DefaultTunnelServer) processControlMessage(conn net.Conn, writer *bufio.Writer, clientID string, msg *types.ControlMessage) error {
	// 更新客户端最后活动时间
	s.mutex.Lock()
	if client, exists := s.connectedClients[clientID]; exists {
		client.LastActivityTime = time.Now()
	}
	s.mutex.Unlock()

	// 重置读取超时（使用服务器配置的心跳超时时间）
	conn.SetReadDeadline(time.Now().Add(s.heartbeatTimeout))

	// 根据消息类型路由到对应的处理函数
	// 所有消息类型都是 Client → Server 方向
	switch msg.Type {
	case types.MessageTypeHeartbeat:
		// 心跳消息 [Client → Server]
		// 客户端定期发送以保持连接活跃
		return s.handleHeartbeat(conn, writer, clientID, msg)

	case types.MessageTypeRegisterService:
		// 注册服务消息 [Client → Server]
		// 客户端请求注册一个新的隧道服务
		return s.handleRegisterService(conn, writer, clientID, msg)

	case types.MessageTypeUnregisterService:
		// 注销服务消息 [Client → Server]
		// 客户端请求注销一个已注册的隧道服务
		return s.handleUnregisterService(conn, writer, clientID, msg)

	default:
		// 未知消息类型
		logger.Warn("Unknown message type received", map[string]interface{}{
			"messageType": msg.Type,
			"clientId":    clientID,
		})
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}
}

// handleHeartbeat 处理心跳消息
// 消息类型: MessageTypeHeartbeat
// 传输方向: Client → Server
// 请求结构: HeartbeatRequest
// 响应结构: CommonResponse
// 功能说明: 客户端定期发送心跳以保持连接活跃，服务端更新客户端的最后活动时间
func (s *DefaultTunnelServer) handleHeartbeat(conn net.Conn, writer *bufio.Writer, clientID string, msg *types.ControlMessage) error {
	// 解析心跳请求
	var req types.HeartbeatRequest
	if err := parseMessageData(msg.Data, &req); err != nil {
		logger.Warn("Failed to parse heartbeat request", map[string]interface{}{
			"error":    err.Error(),
			"clientId": clientID,
		})
		// 心跳消息解析失败不影响功能，继续处理
	}

	logger.Debug("Heartbeat received", map[string]interface{}{
		"clientId":  clientID,
		"timestamp": req.Timestamp,
	})

	// 返回心跳响应 [Server → Client]
	return s.sendResponse(conn, writer, clientID, msg, true, "heartbeat received")
}

// handleRegisterService 处理服务注册消息
// 消息类型: MessageTypeRegisterService
// 传输方向: Client → Server
// 请求结构: RegisterServiceRequest (包含完整的 TunnelService 对象)
// 响应结构: RegisterServiceResponse
// 功能说明: 客户端请求注册一个新的隧道服务（如 SSH、HTTP 等）
//
//	服务端验证后启动对应的代理端口，并返回分配的端口信息
func (s *DefaultTunnelServer) handleRegisterService(conn net.Conn, writer *bufio.Writer, clientID string, msg *types.ControlMessage) error {
	// 解析服务注册请求
	var req types.RegisterServiceRequest
	if err := parseMessageData(msg.Data, &req); err != nil {
		logger.Error("Failed to parse register service request", map[string]interface{}{
			"error":    err.Error(),
			"clientId": clientID,
		})
		return s.sendResponse(conn, writer, clientID, msg, false, fmt.Sprintf("failed to parse service data: %v", err))
	}

	// 获取服务对象（完整的 TunnelService）
	service := req.Service

	// 验证必需字段
	if service.TunnelServiceId == "" {
		return s.sendResponse(conn, writer, clientID, msg, false, "missing serviceId")
	}

	if service.ServiceName == "" {
		return s.sendResponse(conn, writer, clientID, msg, false, "missing serviceName")
	}

	// 强制覆盖客户端ID
	service.TunnelClientId = clientID

	// 服务器端设置的状态和时间字段
	service.ServiceStatus = types.ServiceStatusActive
	now := time.Now()
	service.RegisteredTime = now
	service.LastActiveTime = &now
	service.ConnectionCount = 0
	service.TotalConnections = 0
	service.TotalTraffic = 0
	service.AddTime = now
	service.EditTime = now
	service.ActiveFlag = types.ActiveFlagYes

	// 注册服务
	if err := s.RegisterService(context.Background(), clientID, &service); err != nil {
		logger.Error("Failed to register service", map[string]interface{}{
			"error":       err.Error(),
			"clientId":    clientID,
			"serviceId":   service.TunnelServiceId,
			"serviceName": service.ServiceName,
		})
		return s.sendResponse(conn, writer, clientID, msg, false, fmt.Sprintf("failed to register service: %v", err))
	}

	// 返回注册成功响应 [Server → Client]
	// 响应类型: RegisterServiceResponse
	responseData := map[string]interface{}{
		"success":   true,
		"message":   "service registered successfully",
		"serviceId": service.TunnelServiceId,
	}

	if service.RemotePort != nil {
		responseData["remotePort"] = *service.RemotePort
	}

	response := types.ControlMessage{
		Type:      types.MessageTypeResponse,
		SessionID: msg.SessionID,
		Data:      responseData,
		Timestamp: time.Now(),
	}

	return s.sendControlMessageToClient(conn, writer, &response)
}

// handleUnregisterService 处理服务注销消息
// 消息类型: MessageTypeUnregisterService
// 传输方向: Client → Server
// 请求结构: UnregisterServiceRequest
// 响应结构: CommonResponse
// 功能说明: 客户端请求注销一个已注册的隧道服务
//
//	服务端停止对应的代理端口，并清理相关资源
func (s *DefaultTunnelServer) handleUnregisterService(conn net.Conn, writer *bufio.Writer, clientID string, msg *types.ControlMessage) error {
	// 解析服务注销请求
	var req types.UnregisterServiceRequest
	if err := parseMessageData(msg.Data, &req); err != nil {
		logger.Error("Failed to parse unregister service request", map[string]interface{}{
			"error":    err.Error(),
			"clientId": clientID,
		})
		return s.sendResponse(conn, writer, clientID, msg, false, fmt.Sprintf("failed to parse request: %v", err))
	}

	// 优先使用 ServiceID，如果没有则使用 ServiceName 查找
	var serviceID string
	if req.ServiceID != "" {
		serviceID = req.ServiceID
	} else if req.ServiceName != "" {
		// 通过 ServiceName 查找 ServiceID
		s.mutex.RLock()
		client, exists := s.connectedClients[clientID]
		if exists && client.Services != nil {
			for _, svc := range client.Services {
				if svc.ServiceName == req.ServiceName {
					serviceID = svc.TunnelServiceId
					break
				}
			}
		}
		s.mutex.RUnlock()

		if serviceID == "" {
			return s.sendResponse(conn, writer, clientID, msg, false, "service not found")
		}
	} else {
		return s.sendResponse(conn, writer, clientID, msg, false, "missing serviceId or serviceName")
	}

	// 注销服务
	if err := s.UnregisterService(context.Background(), clientID, serviceID); err != nil {
		logger.Error("Failed to unregister service", map[string]interface{}{
			"error":     err.Error(),
			"clientId":  clientID,
			"serviceId": serviceID,
		})
		return s.sendResponse(conn, writer, clientID, msg, false, fmt.Sprintf("failed to unregister service: %v", err))
	}

	// 返回注销成功响应 [Server → Client]
	// 响应类型: CommonResponse
	return s.sendResponse(conn, writer, clientID, msg, true, "service unregistered successfully")
}

// sendResponse 发送通用响应消息
// 传输方向: Server → Client
// 响应类型: MessageTypeResponse (CommonResponse)
// 功能说明: 构造并发送一个通用的成功/失败响应消息
func (s *DefaultTunnelServer) sendResponse(conn net.Conn, writer *bufio.Writer, clientID string, originalMsg *types.ControlMessage, success bool, message string) error {
	if conn == nil {
		return fmt.Errorf("connection is nil")
	}

	response := types.ControlMessage{
		Type:      types.MessageTypeResponse,
		SessionID: originalMsg.SessionID,
		Data: map[string]interface{}{
			"success": success,
			"message": message,
		},
		Timestamp: time.Now(),
	}

	return s.sendControlMessageToClient(conn, writer, &response)
}

// sendControlMessageToClient 发送控制消息到客户端
// 传输方向: Server → Client
// 功能说明: 底层消息发送函数，负责：
//  1. 序列化消息为 JSON
//  2. 发送消息长度（4字节大端序）
//  3. 发送消息内容
//  4. Flush 确保数据立即发送
//
// 协议格式: [4字节长度][JSON消息内容]
func (s *DefaultTunnelServer) sendControlMessageToClient(conn net.Conn, writer *bufio.Writer, msg *types.ControlMessage) error {
	if conn == nil {
		return fmt.Errorf("connection is nil")
	}

	if writer == nil {
		return fmt.Errorf("writer is nil")
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal control message: %w", err)
	}

	// 发送消息长度
	length := len(data)
	lengthBuf := []byte{
		byte(length >> 24),
		byte(length >> 16),
		byte(length >> 8),
		byte(length),
	}

	// 设置写超时
	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	defer conn.SetWriteDeadline(time.Time{})

	// 写入长度
	if _, err := writer.Write(lengthBuf); err != nil {
		return fmt.Errorf("failed to write message length: %w", err)
	}

	// 写入消息内容
	if _, err := writer.Write(data); err != nil {
		return fmt.Errorf("failed to write message data: %w", err)
	}

	// Flush 确保数据立即发送
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush message: %w", err)
	}

	return nil
}

// parseMessageData 解析消息数据到指定的结构体
// 功能说明: 统一的消息数据解析函数，用于将 map[string]interface{} 转换为具体的消息结构体
// 实现方式: 使用 JSON 序列化/反序列化来处理类型转换
// 使用场景: 所有消息处理函数都应该使用此函数来解析 msg.Data
func parseMessageData(data map[string]interface{}, target interface{}) error {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal message data: %w", err)
	}

	if err := json.Unmarshal(dataJSON, target); err != nil {
		return fmt.Errorf("failed to unmarshal message data: %w", err)
	}

	return nil
}
