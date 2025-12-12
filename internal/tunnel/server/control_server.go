// Package server 提供控制服务器的完整实现
// 控制服务器负责处理客户端的控制连接，包括认证、心跳、服务注册等
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
	"gateway/pkg/utils/random"
)

// controlServer 控制服务器实现
// 实现 ControlServer 接口，处理客户端控制连接
type controlServer struct {
	tunnelServer     TunnelServer
	proxyServer      ProxyServer
	listener         net.Listener
	activeConns      map[string]*controlConnection
	connMutex        sync.RWMutex
	ctx              context.Context
	cancel           context.CancelFunc
	wg               sync.WaitGroup
	heartbeatTimeout time.Duration
}

// controlConnection 控制连接状态
type controlConnection struct {
	conn          net.Conn
	clientID      string
	sessionID     string
	lastActivity  time.Time
	authenticated bool
	services      map[string]*types.TunnelService
	serviceMutex  sync.RWMutex

	// 关键修复：使用 bufio.Writer 确保消息发送的原子性，避免TCP粘包/拆包问题
	writer      *bufio.Writer
	writerMutex sync.Mutex // 保护 writer 的写入操作
}

// NewControlServerImpl 创建新的控制服务器实例
//
// 参数:
//   - tunnelServer: 隧道服务器实例，用于获取配置和管理客户端
//
// 返回:
//   - ControlServer: 控制服务器接口实例
//
// 功能:
//   - 初始化控制服务器
//   - 设置默认心跳超时时间为 30 秒
//   - 创建活跃连接映射表
func NewControlServerImpl(tunnelServer TunnelServer) ControlServer {
	return &controlServer{
		tunnelServer:     tunnelServer,
		activeConns:      make(map[string]*controlConnection),
		heartbeatTimeout: 180 * time.Second,
	}
}

// SetProxyServer 设置代理服务器引用
// 用于处理客户端数据连接
func (s *controlServer) SetProxyServer(proxyServer ProxyServer) {
	s.proxyServer = proxyServer
}

// Start 启动控制服务器
//
// 参数:
//   - ctx: 上下文，用于控制服务器生命周期
//   - address: 监听地址，如 "0.0.0.0"
//   - port: 监听端口，如 7000
//
// 返回:
//   - error: 启动失败时返回错误
//
// 功能:
//   - 在指定地址和端口启动 TCP 监听器
//   - 启动连接接受循环
//   - 启动心跳检查定时器
func (s *controlServer) Start(ctx context.Context, address string, port int) error {
	s.ctx, s.cancel = context.WithCancel(ctx)

	listenAddr := fmt.Sprintf("%s:%d", address, port)
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return fmt.Errorf("failed to start control server on %s: %w", listenAddr, err)
	}

	s.listener = listener

	logger.Info("Control server started", map[string]interface{}{
		"address": address,
		"port":    port,
	})

	// 启动接受连接的协程
	s.wg.Add(1)
	go s.acceptConnections()

	// 启动心跳检查协程
	s.wg.Add(1)
	go s.heartbeatChecker()

	return nil
}

// Stop 停止控制服务器
//
// 参数:
//   - ctx: 上下文，用于控制停止超时
//
// 返回:
//   - error: 停止失败时返回错误
//
// 功能:
//   - 关闭监听器
//   - 关闭所有活跃连接
//   - 等待所有协程退出
func (s *controlServer) Stop(ctx context.Context) error {
	if s.cancel != nil {
		s.cancel()
	}

	if s.listener != nil {
		s.listener.Close()
	}

	// 收集所有需要关闭的连接（不持有锁）
	s.connMutex.RLock()
	connsToClose := make([]*controlConnection, 0, len(s.activeConns))
	for _, conn := range s.activeConns {
		connsToClose = append(connsToClose, conn)
	}
	s.connMutex.RUnlock()

	// 在不持有锁的情况下关闭所有连接
	for _, conn := range connsToClose {
		conn.conn.Close()
	}

	// 等待所有协程退出
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("Control server stopped gracefully", nil)
		return nil
	case <-ctx.Done():
		logger.Warn("Control server stop timeout", nil)
		return ctx.Err()
	}
}

// HandleConnection 处理单个客户端连接
//
// 参数:
//   - ctx: 上下文
//   - conn: 网络连接
//
// 返回:
//   - error: 处理失败时返回错误
//
// 功能:
//   - 创建控制连接实例
//   - 启动消息处理循环
//   - 处理连接断开清理
func (s *controlServer) HandleConnection(ctx context.Context, conn net.Conn) error {
	// 关键优化：使用较短的初始读取超时（10秒），快速检测无效连接
	// 原因：
	// 1. 客户端建立连接后应该立即发送握手，正常情况下应该在几毫秒内完成
	// 2. 如果10秒内没有收到任何数据，可能是无效连接或网络问题，可以快速关闭释放资源
	// 3. 这样可以避免大量无效连接占用资源（文件描述符、goroutine、内存），导致 Accept 队列满
	// 4. 对于有效的连接，读取到消息长度后会延长超时时间用于读取消息内容
	// 5. 系统参数 somaxconn=2048 已经很大，但关键是处理速度，不是队列大小
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	// 读取消息长度
	lengthBuf := make([]byte, 4)
	if _, err := io.ReadFull(conn, lengthBuf); err != nil {
		// 如果是超时错误，记录详细信息以便诊断
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			logger.Warn("Connection timeout reading handshake", map[string]interface{}{
				"remoteAddr": conn.RemoteAddr().String(),
				"localAddr":  conn.LocalAddr().String(),
				"error":      err.Error(),
				"note":       "client may have failed to send handshake in time, or connection is invalid",
			})
		}
		return fmt.Errorf("failed to read message length: %w", err)
	}

	// 读取到消息长度后，延长超时时间用于读取消息内容
	// 消息长度已经读取到，说明连接是有效的，可以给更多时间读取消息内容
	conn.SetReadDeadline(time.Now().Add(120 * time.Second))

	// 解析消息长度
	msgLen := int(lengthBuf[0])<<24 | int(lengthBuf[1])<<16 | int(lengthBuf[2])<<8 | int(lengthBuf[3])

	// 关键修复：增强消息长度验证，检测各种异常情况
	// 1. 检查长度是否全为0（可能是连接关闭时的填充字节）
	allZeros := lengthBuf[0] == 0 && lengthBuf[1] == 0 && lengthBuf[2] == 0 && lengthBuf[3] == 0
	if allZeros {
		logger.Warn("Message length is all zeros, possible connection closed or reset", map[string]interface{}{
			"remoteAddr":  conn.RemoteAddr().String(),
			"localAddr":   conn.LocalAddr().String(),
			"lengthBytes": lengthBuf,
		})
		return fmt.Errorf("message length is all zeros (possible connection closed or reset)")
	}

	// 2. 检查长度是否太小（可能是读取错误或消息边界错位）
	if msgLen < 10 {
		logger.Error("Message length too small, possible read error or message boundary misalignment", map[string]interface{}{
			"messageLength": msgLen,
			"lengthBytes":   lengthBuf,
			"remoteAddr":    conn.RemoteAddr().String(),
			"localAddr":     conn.LocalAddr().String(),
			"possibleCause": "read_error_or_message_boundary_misalignment",
		})
		return fmt.Errorf("message length too small: %d (possible read error or message boundary misalignment)", msgLen)
	}

	// 3. 检查长度是否过大（可能是读取位置错位，读取到了错误的数据）
	if msgLen > 1024*1024 {
		logger.Error("Invalid message length on control connection", map[string]interface{}{
			"messageLength": msgLen,
			"lengthBytes":   lengthBuf,
			"lengthHex":     fmt.Sprintf("0x%02x%02x%02x%02x", lengthBuf[0], lengthBuf[1], lengthBuf[2], lengthBuf[3]),
			"remoteAddr":    conn.RemoteAddr().String(),
			"localAddr":     conn.LocalAddr().String(),
			"possibleCause": "message_boundary_misalignment_or_connection_confusion",
		})
		return fmt.Errorf("invalid message length: %d (length bytes: [0x%02x, 0x%02x, 0x%02x, 0x%02x], possible message boundary misalignment or connection confusion)",
			msgLen, lengthBuf[0], lengthBuf[1], lengthBuf[2], lengthBuf[3])
	}

	// 读取消息内容
	msgBuf := make([]byte, msgLen)
	if _, err := io.ReadFull(conn, msgBuf); err != nil {
		return fmt.Errorf("failed to read message data: %w", err)
	}

	// 尝试解析消息以判断连接类型
	var firstMsg map[string]interface{}
	if err := json.Unmarshal(msgBuf, &firstMsg); err != nil {
		return fmt.Errorf("failed to parse first message: %w", err)
	}

	// 检查是否为数据连接
	if msgType, ok := firstMsg["type"].(string); ok && msgType == "data_connection" {
		return s.handleDataConnection(ctx, conn, firstMsg)
	}

	// 否则作为控制连接处理
	return s.handleControlConnection(ctx, conn, msgBuf)
}

// handleDataConnection 处理数据连接
func (s *controlServer) handleDataConnection(ctx context.Context, conn net.Conn, handshake map[string]interface{}) error {
	connectionID, ok := handshake["connectionId"].(string)
	if !ok {
		conn.Close()
		return fmt.Errorf("missing connectionId in data connection handshake")
	}

	// 提取并验证 clientID（必须存在且不能为空）
	clientID, ok := handshake["clientId"].(string)
	if !ok || clientID == "" {
		conn.Close()
		return fmt.Errorf("missing or empty clientId in data connection handshake")
	}

	logger.Info("Data connection received", map[string]interface{}{
		"connectionId": connectionID,
		"clientId":     clientID,
	})

	// 关键修复：清除读取超时设置，确保数据连接可以持续工作
	// HandleConnection 中设置了 120 秒的读取超时用于读取握手消息
	// 握手消息读取完成后，需要清除超时设置，否则数据连接在 120 秒后会触发 i/o timeout
	// 这对于 SSE 等长连接非常重要
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetReadDeadline(time.Time{})  // 清除读取超时
		tcpConn.SetWriteDeadline(time.Time{}) // 清除写入超时
	}

	// 将数据连接转发给代理服务器处理
	if s.proxyServer == nil {
		conn.Close()
		return fmt.Errorf("proxy server not configured")
	}

	return s.proxyServer.HandleClientDataConnection(ctx, conn, connectionID, clientID)
}

// handleControlConnection 处理控制连接
func (s *controlServer) handleControlConnection(ctx context.Context, conn net.Conn, firstMsgBuf []byte) error {
	controlConn := &controlConnection{
		conn:         conn,
		lastActivity: time.Now(),
		services:     make(map[string]*types.TunnelService),
		// 关键修复：创建 bufio.Writer 确保消息发送的原子性
		writer: bufio.NewWriter(conn),
	}

	defer func() {
		// 关键修复：清理 writer，避免资源泄露
		controlConn.writerMutex.Lock()
		controlConn.writer = nil
		controlConn.writerMutex.Unlock()
		conn.Close()
		//不应该清楚连接由客户端注册控制
		//s.removeConnection(controlConn)
	}()

	// 处理第一条消息（已经读取的）
	var firstMsg types.ControlMessage
	if err := json.Unmarshal(firstMsgBuf, &firstMsg); err != nil {
		return fmt.Errorf("failed to parse first control message: %w", err)
	}

	// 处理第一条消息
	if err := s.processControlMessage(controlConn, &firstMsg); err != nil {
		return fmt.Errorf("failed to process first control message: %w", err)
	}

	// 设置读取超时
	conn.SetReadDeadline(time.Now().Add(s.heartbeatTimeout))

	// 继续消息处理循环
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := s.handleMessage(controlConn); err != nil {
				if err == io.EOF {
					logger.Info("Client disconnected", map[string]interface{}{
						"clientId": controlConn.clientID,
					})
					return nil
				}
				logger.Error("Error handling message", map[string]interface{}{
					"error":    err.Error(),
					"clientId": controlConn.clientID,
				})
				return err
			}
		}
	}
}

// SendMessageToClient 向指定客户端发送消息
//
// 参数:
//   - clientID: 目标客户端ID
//   - message: 要发送的控制消息
//
// 返回:
//   - error: 发送失败时返回错误
//
// 功能:
//   - 查找指定客户端的控制连接
//   - 发送控制消息到客户端
func (s *controlServer) SendMessageToClient(clientID string, message *types.ControlMessage) error {
	// 快速查找连接（短时间持有读锁）
	s.connMutex.RLock()
	conn, exists := s.activeConns[clientID]
	authenticated := exists && conn != nil && conn.authenticated
	s.connMutex.RUnlock()

	if !exists || conn == nil {
		return fmt.Errorf("client %s not found or not connected", clientID)
	}

	if !authenticated {
		return fmt.Errorf("client %s not authenticated", clientID)
	}

	// 关键修复：在发送前再次验证连接是否仍然有效
	// 防止在获取连接后、发送消息前连接被关闭（竞态条件）
	s.connMutex.RLock()
	conn2, stillExists := s.activeConns[clientID]
	stillAuthenticated := stillExists && conn2 != nil && conn2.authenticated && conn2 == conn
	s.connMutex.RUnlock()

	if !stillExists || conn2 == nil || !stillAuthenticated {
		return fmt.Errorf("client %s connection closed or invalidated before sending message", clientID)
	}

	// 在不持有锁的情况下发送消息
	return s.sendControlMessage(conn, message)
}

// GetActiveConnections 获取活跃连接数
//
// 返回:
//   - int: 当前活跃的控制连接数量
//
// 功能:
//   - 返回当前已认证的控制连接数量
func (s *controlServer) GetActiveConnections() int {
	s.connMutex.RLock()
	defer s.connMutex.RUnlock()

	count := 0
	for _, conn := range s.activeConns {
		if conn.authenticated {
			count++
		}
	}
	return count
}

// acceptConnections 接受新连接的循环
func (s *controlServer) acceptConnections() {
	defer s.wg.Done()

	for {
		// Accept 连接
		// 注意：Accept 会阻塞直到有新连接到来，这是正常的
		// 如果连接队列满，Accept 会阻塞，但这是正常的，因为连接已经在队列中等待
		conn, err := s.listener.Accept()

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

		// 关键优化：立即启动 goroutine 处理连接，避免阻塞 Accept 循环
		// 这样可以快速处理连接，减少队列堆积
		s.wg.Add(1)
		go func(conn net.Conn) {
			defer s.wg.Done()
			// 使用独立的上下文，避免单个连接处理阻塞影响其他连接
			connCtx, cancel := context.WithCancel(s.ctx)
			defer cancel()
			if err := s.HandleConnection(connCtx, conn); err != nil {
				logger.Error("Connection handling failed", map[string]interface{}{
					"error": err.Error(),
				})
			}
		}(conn)
	}
}

// heartbeatChecker 心跳检查器
func (s *controlServer) heartbeatChecker() {
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
func (s *controlServer) checkHeartbeats() {
	s.connMutex.RLock()
	var expiredConns []*controlConnection
	now := time.Now()

	for _, conn := range s.activeConns {
		if now.Sub(conn.lastActivity) > s.heartbeatTimeout {
			expiredConns = append(expiredConns, conn)
		}
	}
	s.connMutex.RUnlock()

	// 关闭过期连接
	for _, conn := range expiredConns {
		logger.Warn("Connection timeout, closing", map[string]interface{}{
			"clientId": conn.clientID,
		})
		conn.conn.Close()
	}
}

// handleMessage 处理控制消息
func (s *controlServer) handleMessage(conn *controlConnection) error {
	// 读取消息长度
	lengthBuf := make([]byte, 4)
	if _, err := io.ReadFull(conn.conn, lengthBuf); err != nil {
		return err
	}

	// 解析消息长度
	msgLen := int(lengthBuf[0])<<24 | int(lengthBuf[1])<<16 | int(lengthBuf[2])<<8 | int(lengthBuf[3])

	// 关键修复：增强消息长度验证，检测各种异常情况
	// 1. 检查长度是否全为0（可能是连接关闭时的填充字节）
	allZeros := lengthBuf[0] == 0 && lengthBuf[1] == 0 && lengthBuf[2] == 0 && lengthBuf[3] == 0
	if allZeros {
		logger.Warn("Message length is all zeros in handleMessage, possible connection closed", map[string]interface{}{
			"clientId":    conn.clientID,
			"lengthBytes": lengthBuf,
		})
		return fmt.Errorf("message length is all zeros (possible connection closed)")
	}

	// 2. 检查长度是否太小（可能是读取错误或消息边界错位）
	if msgLen < 10 {
		logger.Error("Message length too small in handleMessage, possible read error", map[string]interface{}{
			"messageLength": msgLen,
			"lengthBytes":   lengthBuf,
			"clientId":      conn.clientID,
			"possibleCause": "read_error_or_message_boundary_misalignment",
		})
		return fmt.Errorf("message length too small: %d (possible read error or message boundary misalignment)", msgLen)
	}

	// 3. 检查长度是否过大（可能是读取位置错位，读取到了错误的数据）
	if msgLen > 1024*1024 { // 限制消息大小为1MB
		logger.Error("Invalid message length in handleMessage", map[string]interface{}{
			"messageLength": msgLen,
			"lengthBytes":   lengthBuf,
			"lengthHex":     fmt.Sprintf("0x%02x%02x%02x%02x", lengthBuf[0], lengthBuf[1], lengthBuf[2], lengthBuf[3]),
			"clientId":      conn.clientID,
			"possibleCause": "message_boundary_misalignment_or_connection_confusion",
		})
		return fmt.Errorf("invalid message length: %d (length bytes: [0x%02x, 0x%02x, 0x%02x, 0x%02x], possible message boundary misalignment or connection confusion)",
			msgLen, lengthBuf[0], lengthBuf[1], lengthBuf[2], lengthBuf[3])
	}

	// 读取消息内容
	msgBuf := make([]byte, msgLen)
	if _, err := io.ReadFull(conn.conn, msgBuf); err != nil {
		return err
	}

	// 解析控制消息
	var msg types.ControlMessage
	if err := json.Unmarshal(msgBuf, &msg); err != nil {
		return fmt.Errorf("failed to parse control message: %w", err)
	}

	return s.processControlMessage(conn, &msg)
}

// processControlMessage 处理控制消息
func (s *controlServer) processControlMessage(conn *controlConnection, msg *types.ControlMessage) error {
	// 更新活动时间
	conn.lastActivity = time.Now()
	conn.conn.SetReadDeadline(time.Now().Add(s.heartbeatTimeout))

	// 处理不同类型的消息
	switch msg.Type {
	case types.MessageTypeAuth:
		return s.handleAuth(conn, msg)
	case types.MessageTypeHeartbeat:
		return s.handleHeartbeat(conn, msg)
	case types.MessageTypeRegisterService:
		return s.handleRegisterService(conn, msg)
	case types.MessageTypeUnregisterService:
		return s.handleUnregisterService(conn, msg)
	default:
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}
}

// handleAuth 处理认证消息
func (s *controlServer) handleAuth(conn *controlConnection, msg *types.ControlMessage) error {
	// 提取认证信息
	clientID, ok := msg.Data["clientId"].(string)
	if !ok {
		return s.sendResponse(conn, msg, false, "missing clientId")
	}

	token, ok := msg.Data["token"].(string)
	if !ok {
		return s.sendResponse(conn, msg, false, "missing token")
	}

	// 验证客户端和令牌（不持有锁）
	config := s.tunnelServer.GetConfig()
	if config.AuthToken != token {
		return s.sendResponse(conn, msg, false, "invalid token")
	}

	// 检查是否已有相同客户端连接（需要处理重复连接）
	s.connMutex.Lock()
	if existingConn, exists := s.activeConns[clientID]; exists {
		// 关闭旧连接
		logger.Warn("Duplicate client connection, closing old connection", map[string]interface{}{
			"clientId": clientID,
		})
		go existingConn.conn.Close()
	}

	// 设置连接状态并添加到活跃连接
	conn.clientID = clientID
	conn.sessionID = msg.SessionID
	conn.authenticated = true
	s.activeConns[clientID] = conn
	s.connMutex.Unlock()

	logger.Info("Client authenticated", map[string]interface{}{
		"clientId":  clientID,
		"sessionId": msg.SessionID,
	})

	return s.sendResponse(conn, msg, true, "authenticated successfully")
}

// handleHeartbeat 处理心跳消息
func (s *controlServer) handleHeartbeat(conn *controlConnection, msg *types.ControlMessage) error {
	if !conn.authenticated {
		return s.sendResponse(conn, msg, false, "not authenticated")
	}

	return s.sendResponse(conn, msg, true, "heartbeat received")
}

// handleRegisterService 处理服务注册消息
func (s *controlServer) handleRegisterService(conn *controlConnection, msg *types.ControlMessage) error {
	if !conn.authenticated {
		return s.sendResponse(conn, msg, false, "not authenticated")
	}

	// 解析服务配置
	// 关键改进：直接从客户端接收完整的 types.TunnelService 对象
	// 避免字段遗漏和类型转换错误
	serviceData, ok := msg.Data["service"]
	if !ok {
		return s.sendResponse(conn, msg, false, "missing service data")
	}

	// 将 service 数据转换为 types.TunnelService
	// 使用 JSON 序列化/反序列化来处理类型转换
	var service types.TunnelService

	// 将 serviceData 重新编码为 JSON，然后解码到 service 结构体
	// 这样可以正确处理 json tag 的映射
	serviceJSON, err := json.Marshal(serviceData)
	if err != nil {
		logger.Error("Failed to marshal service data", map[string]interface{}{
			"error":       err.Error(),
			"clientId":    conn.clientID,
			"serviceData": serviceData,
		})
		return s.sendResponse(conn, msg, false, fmt.Sprintf("failed to parse service data: %v", err))
	}

	if err := json.Unmarshal(serviceJSON, &service); err != nil {
		logger.Error("Failed to unmarshal service data", map[string]interface{}{
			"error":       err.Error(),
			"clientId":    conn.clientID,
			"serviceJSON": string(serviceJSON),
		})
		return s.sendResponse(conn, msg, false, fmt.Sprintf("failed to unmarshal service data: %v", err))
	}

	// 验证必需字段
	if service.TunnelServiceId == "" {
		logger.Warn("Service registration missing serviceId", map[string]interface{}{
			"clientId":    conn.clientID,
			"serviceJSON": string(serviceJSON),
			"service":     service,
		})
		return s.sendResponse(conn, msg, false, "missing serviceId")
	}

	if service.ServiceName == "" {
		return s.sendResponse(conn, msg, false, "missing serviceName")
	}

	// 强制覆盖客户端ID，确保一致性
	service.TunnelClientId = conn.clientID

	// 服务器端设置的状态和时间字段
	// 这些字段由服务器管理，客户端提供的值会被覆盖
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

	// 如果客户端没有指定远程端口，服务器需要分配一个
	// （这部分逻辑由服务注册器处理）

	logger.Info("Registering service with complete data", map[string]interface{}{
		"clientId":      conn.clientID,
		"serviceId":     service.TunnelServiceId,
		"serviceName":   service.ServiceName,
		"serviceType":   service.ServiceType,
		"localAddress":  service.LocalAddress,
		"localPort":     service.LocalPort,
		"remotePort":    service.RemotePort,
		"customDomains": service.CustomDomains,
		"subDomain":     service.SubDomain,
	})

	// 通过隧道服务器的服务注册器注册服务
	if err := s.tunnelServer.GetServiceRegistry().RegisterService(context.Background(), conn.clientID, &service); err != nil {
		logger.Error("Failed to register service with service registry", map[string]interface{}{
			"error":       err.Error(),
			"clientId":    conn.clientID,
			"serviceId":   service.TunnelServiceId,
			"serviceName": service.ServiceName,
		})
		return s.sendResponse(conn, msg, false, fmt.Sprintf("failed to register service: %v", err))
	}

	// 添加到连接的服务列表
	conn.serviceMutex.Lock()
	conn.services[service.ServiceName] = &service
	conn.serviceMutex.Unlock()

	logger.Info("Service registered successfully", map[string]interface{}{
		"clientId":    conn.clientID,
		"serviceId":   service.TunnelServiceId,
		"serviceName": service.ServiceName,
		"serviceType": service.ServiceType,
		"localPort":   service.LocalPort,
		"remotePort":  service.RemotePort,
	})

	// 返回成功响应，包含分配的远程端口
	responseData := map[string]interface{}{
		"success":   true,
		"message":   "service registered successfully",
		"serviceId": service.TunnelServiceId,
	}

	// 如果有远程端口，返回给客户端
	if service.RemotePort != nil {
		responseData["remotePort"] = *service.RemotePort
	}

	response := types.ControlMessage{
		Type:      types.MessageTypeResponse,
		SessionID: msg.SessionID,
		Data:      responseData,
		Timestamp: time.Now(),
	}

	return s.sendControlMessage(conn, &response)
}

// handleUnregisterService 处理服务注销消息
func (s *controlServer) handleUnregisterService(conn *controlConnection, msg *types.ControlMessage) error {
	if !conn.authenticated {
		return s.sendResponse(conn, msg, false, "not authenticated")
	}

	serviceName, ok := msg.Data["serviceName"].(string)
	if !ok {
		return s.sendResponse(conn, msg, false, "missing service name")
	}

	// 查找服务
	conn.serviceMutex.Lock()
	service, exists := conn.services[serviceName]
	if exists {
		delete(conn.services, serviceName)
	}
	conn.serviceMutex.Unlock()

	if !exists {
		return s.sendResponse(conn, msg, false, "service not found")
	}

	// 从服务注册器中注销服务
	if err := s.tunnelServer.GetServiceRegistry().UnregisterService(context.Background(), service.TunnelServiceId); err != nil {
		logger.Error("Failed to unregister service from service registry", map[string]interface{}{
			"error":     err.Error(),
			"serviceId": service.TunnelServiceId,
		})
	}

	logger.Info("Service unregistered", map[string]interface{}{
		"clientId":    conn.clientID,
		"serviceName": serviceName,
		"serviceId":   service.TunnelServiceId,
	})

	return s.sendResponse(conn, msg, true, "service unregistered successfully")
}

// sendResponse 发送响应消息
func (s *controlServer) sendResponse(conn *controlConnection, originalMsg *types.ControlMessage, success bool, message string) error {
	// 关键修复：在发送前检查连接是否有效
	if conn.conn == nil {
		return fmt.Errorf("connection is nil")
	}

	// 检查连接是否已关闭
	localAddr := conn.conn.LocalAddr()
	if localAddr == nil {
		return fmt.Errorf("connection is closed (localAddr is nil)")
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

	data, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	// 发送消息长度
	length := len(data)
	lengthBuf := []byte{
		byte(length >> 24),
		byte(length >> 16),
		byte(length >> 8),
		byte(length),
	}

	// 关键修复：设置写超时，避免在连接已关闭时长时间阻塞
	conn.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	defer conn.conn.SetWriteDeadline(time.Time{}) // 清除超时

	// 关键修复：使用 bufio.Writer 确保原子写入
	conn.writerMutex.Lock()
	defer conn.writerMutex.Unlock()

	if conn.writer == nil {
		return fmt.Errorf("writer is nil")
	}

	// 写入长度
	if _, err := conn.writer.Write(lengthBuf); err != nil {
		return fmt.Errorf("failed to write message length: %w", err)
	}

	// 写入消息内容
	if _, err := conn.writer.Write(data); err != nil {
		return fmt.Errorf("failed to write message data: %w", err)
	}

	// 关键修复：Flush 确保数据立即发送，避免缓冲导致的消息边界问题
	if err := conn.writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush message: %w", err)
	}

	return nil
}

// removeConnection 移除连接
func (s *controlServer) removeConnection(conn *controlConnection) {
	if conn.clientID == "" {
		return
	}

	// 控制连接只负责通信，不管理监听器

	// 收集需要注销的服务（短时间持有锁）
	conn.serviceMutex.Lock()
	servicesToUnregister := make([]*types.TunnelService, 0, len(conn.services))
	for _, service := range conn.services {
		servicesToUnregister = append(servicesToUnregister, service)
	}
	conn.services = make(map[string]*types.TunnelService)
	conn.serviceMutex.Unlock()

	// 在不持有锁的情况下注销所有服务
	for _, service := range servicesToUnregister {
		if err := s.tunnelServer.GetServiceRegistry().UnregisterService(context.Background(), service.TunnelServiceId); err != nil {
			logger.Error("Failed to unregister service during connection cleanup", map[string]interface{}{
				"error":     err.Error(),
				"serviceId": service.TunnelServiceId,
				"clientId":  conn.clientID,
			})
		}
	}

	// 从活跃连接中移除（短时间持有锁）
	s.connMutex.Lock()
	delete(s.activeConns, conn.clientID)
	s.connMutex.Unlock()

	logger.Info("Connection removed and all services cleaned up", map[string]interface{}{
		"clientId":     conn.clientID,
		"serviceCount": len(servicesToUnregister),
	})
}

// 辅助函数
func getStringValue(data map[string]interface{}, key, defaultValue string) string {
	if value, ok := data[key].(string); ok {
		return value
	}
	return defaultValue
}

func getIntValue(data map[string]interface{}, key string, defaultValue int) int {
	if value, ok := data[key].(float64); ok {
		return int(value)
	}
	return defaultValue
}

// generateID 生成唯一标识符
//
// 使用高强度随机字符串生成器，确保在高并发和分布式环境下的唯一性。
//
// 参数:
//   - prefix: ID前缀，用于标识ID类型
//
// 返回:
//   - string: 唯一标识符，格式为 <prefix>_<20位随机字符串>
func generateID(prefix string) string {
	return fmt.Sprintf("%s_%s", prefix, random.GenerateRandomString(20))
}

// sendControlMessage 发送控制消息
func (s *controlServer) sendControlMessage(conn *controlConnection, msg *types.ControlMessage) error {
	// 关键修复：在发送前检查连接是否有效
	// 防止在连接关闭过程中写入数据，导致客户端读取到不完整消息或 \x00 字节
	if conn.conn == nil {
		return fmt.Errorf("connection is nil")
	}

	// 检查连接是否已关闭（通过尝试获取本地地址）
	// 如果连接已关闭，LocalAddr() 可能返回 nil 或 panic
	localAddr := conn.conn.LocalAddr()
	if localAddr == nil {
		return fmt.Errorf("connection is closed (localAddr is nil)")
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

	// 关键修复：设置写超时，避免在连接已关闭时长时间阻塞
	conn.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	defer conn.conn.SetWriteDeadline(time.Time{}) // 清除超时

	// 关键修复：使用 bufio.Writer 确保原子写入
	conn.writerMutex.Lock()
	defer conn.writerMutex.Unlock()

	if conn.writer == nil {
		return fmt.Errorf("writer is nil")
	}

	// 写入长度
	if _, err := conn.writer.Write(lengthBuf); err != nil {
		return fmt.Errorf("failed to write message length: %w", err)
	}

	// 写入消息内容
	if _, err := conn.writer.Write(data); err != nil {
		return fmt.Errorf("failed to write message data: %w", err)
	}

	// 关键修复：Flush 确保数据立即发送，避免缓冲导致的消息边界问题
	if err := conn.writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush message: %w", err)
	}

	return nil
}
