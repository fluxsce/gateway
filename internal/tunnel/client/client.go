// Package client 提供隧道客户端的完整实现
// 基于FRP架构，实现客户端连接、服务注册和数据转发功能
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"gateway/internal/tunnel/types"
	"gateway/pkg/logger"
)

// tunnelClient 隧道客户端实现
// 实现 TunnelClient 接口，协调各个子组件的工作
type tunnelClient struct {
	config      *types.TunnelClient
	status      *ClientStatus
	statusMutex sync.RWMutex

	// 子组件
	controlConn      ControlConnection
	serviceManager   ServiceManager
	proxyManager     ProxyManager
	heartbeatManager HeartbeatManager
	reconnectManager ReconnectManager

	// 服务列表
	registeredServices map[string]*types.TunnelService
	servicesMutex      sync.RWMutex

	// 控制状态
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	running      bool
	runningMutex sync.RWMutex
}

// NewTunnelClient 创建隧道客户端实例
//
// 参数:
//   - config: 客户端配置对象
//
// 返回:
//   - TunnelClient: 隧道客户端接口实例
//
// 功能:
//   - 创建客户端实例并初始化各个子组件
//   - 设置初始状态和配置参数
//   - 建立组件间的协调关系
func NewTunnelClient(config *types.TunnelClient) TunnelClient {
	ctx, cancel := context.WithCancel(context.Background())

	client := &tunnelClient{
		config:             config,
		registeredServices: make(map[string]*types.TunnelService),
		ctx:                ctx,
		cancel:             cancel,
		running:            false,
		status: &ClientStatus{
			Status:             StatusDisconnected,
			ServerAddress:      config.ServerAddress,
			ServerPort:         config.ServerPort,
			Connected:          false,
			ReconnectCount:     0,
			RegisteredServices: 0,
			ActiveProxies:      0,
			TotalTraffic:       0,
			Errors:             []string{},
		},
	}

	// 初始化子组件
	client.controlConn = NewControlConnection(client)
	client.serviceManager = NewServiceManager(client)
	client.proxyManager = NewProxyManager(client)
	client.heartbeatManager = NewHeartbeatManager(client.controlConn)
	client.reconnectManager = NewReconnectManager(client)

	logger.Info("Tunnel client created", map[string]interface{}{
		"clientId":      config.TunnelClientId,
		"clientName":    config.ClientName,
		"serverAddress": config.ServerAddress,
		"serverPort":    config.ServerPort,
	})

	return client
}

// Start 启动客户端
func (c *tunnelClient) Start(ctx context.Context) error {
	c.runningMutex.Lock()
	if c.running {
		c.runningMutex.Unlock()
		return fmt.Errorf("client is already running")
	}
	c.running = true
	c.runningMutex.Unlock()

	c.updateStatus(StatusConnecting, false)

	logger.Info("Starting tunnel client", map[string]interface{}{
		"clientId":   c.config.TunnelClientId,
		"clientName": c.config.ClientName,
	})

	// 启动重连管理器
	if err := c.reconnectManager.Start(c.ctx); err != nil {
		c.updateStatus(StatusError, false)
		c.addError(fmt.Sprintf("Failed to start reconnect manager: %v", err))
		return fmt.Errorf("failed to start reconnect manager: %w", err)
	}

	// 建立控制连接
	if err := c.controlConn.Connect(c.ctx, c.config.ServerAddress, c.config.ServerPort); err != nil {
		c.updateStatus(StatusError, false)
		c.addError(fmt.Sprintf("Failed to connect to server: %v", err))
		return fmt.Errorf("failed to connect to server: %w", err)
	}

	// 启动心跳管理器
	heartbeatInterval := time.Duration(c.config.HeartbeatInterval) * time.Second
	if err := c.heartbeatManager.Start(c.ctx, heartbeatInterval); err != nil {
		c.updateStatus(StatusError, false)
		c.addError(fmt.Sprintf("Failed to start heartbeat manager: %v", err))
		return fmt.Errorf("failed to start heartbeat manager: %w", err)
	}

	// 启动消息处理循环
	c.wg.Add(1)
	go c.messageLoop()

	c.updateStatus(StatusConnected, true)
	c.updateConnectTime()

	logger.Info("Tunnel client started successfully", map[string]interface{}{
		"clientId":   c.config.TunnelClientId,
		"clientName": c.config.ClientName,
	})

	return nil
}

// Stop 停止客户端
func (c *tunnelClient) Stop(ctx context.Context) error {
	c.runningMutex.Lock()
	if !c.running {
		c.runningMutex.Unlock()
		return nil
	}
	c.running = false
	c.runningMutex.Unlock()

	logger.Info("Stopping tunnel client", map[string]interface{}{
		"clientId":   c.config.TunnelClientId,
		"clientName": c.config.ClientName,
	})

	// 注销所有服务
	c.servicesMutex.RLock()
	serviceIDs := make([]string, 0, len(c.registeredServices))
	for serviceID := range c.registeredServices {
		serviceIDs = append(serviceIDs, serviceID)
	}
	c.servicesMutex.RUnlock()

	for _, serviceID := range serviceIDs {
		if err := c.UnregisterService(ctx, serviceID); err != nil {
			logger.Error("Failed to unregister service during shutdown", map[string]interface{}{
				"serviceId": serviceID,
				"error":     err.Error(),
			})
		}
	}

	// 停止心跳管理器
	if err := c.heartbeatManager.Stop(ctx); err != nil {
		logger.Error("Failed to stop heartbeat manager", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// 停止重连管理器
	if err := c.reconnectManager.Stop(ctx); err != nil {
		logger.Error("Failed to stop reconnect manager", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// 断开控制连接
	if err := c.controlConn.Disconnect(ctx); err != nil {
		logger.Error("Failed to disconnect control connection", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// 取消上下文并等待协程退出
	c.cancel()

	// 等待消息循环退出
	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		c.updateStatus(StatusStopped, false)
		logger.Info("Tunnel client stopped successfully", map[string]interface{}{
			"clientId":   c.config.TunnelClientId,
			"clientName": c.config.ClientName,
		})
		return nil
	case <-ctx.Done():
		c.updateStatus(StatusError, false)
		c.addError("Client stop timeout")
		return fmt.Errorf("client stop timeout")
	}
}

// GetStatus 获取客户端状态
func (c *tunnelClient) GetStatus() *ClientStatus {
	c.statusMutex.RLock()
	defer c.statusMutex.RUnlock()

	// 创建状态副本
	status := &ClientStatus{
		Status:             c.status.Status,
		ServerAddress:      c.status.ServerAddress,
		ServerPort:         c.status.ServerPort,
		Connected:          c.status.Connected,
		LastConnectTime:    c.status.LastConnectTime,
		ConnectionDuration: c.status.ConnectionDuration,
		ReconnectCount:     c.status.ReconnectCount,
		RegisteredServices: c.status.RegisteredServices,
		ActiveProxies:      c.status.ActiveProxies,
		TotalTraffic:       c.status.TotalTraffic,
		LastHeartbeat:      c.status.LastHeartbeat,
		Errors:             make([]string, len(c.status.Errors)),
	}
	copy(status.Errors, c.status.Errors)

	// 更新连接持续时间
	if status.Connected && !status.LastConnectTime.IsZero() {
		status.ConnectionDuration = time.Since(status.LastConnectTime).Milliseconds()
	}

	// 更新心跳时间
	status.LastHeartbeat = c.heartbeatManager.GetLastHeartbeatTime()

	// 更新代理数量
	status.ActiveProxies = len(c.proxyManager.GetActiveProxies())

	return status
}

// GetConfig 获取客户端配置
func (c *tunnelClient) GetConfig() *types.TunnelClient {
	return c.config
}

// RegisterService 注册服务
func (c *tunnelClient) RegisterService(ctx context.Context, service *types.TunnelService) error {
	if !c.isConnected() {
		return fmt.Errorf("client is not connected")
	}

	// 验证服务配置
	if err := c.serviceManager.ValidateService(ctx, service); err != nil {
		return fmt.Errorf("service validation failed: %w", err)
	}

	// 设置服务基本信息
	service.TunnelClientId = c.config.TunnelClientId
	service.TunnelServiceId = c.generateServiceID(service.ServiceName)

	// 注册服务
	if err := c.serviceManager.RegisterService(ctx, service); err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	// 添加到本地服务列表
	c.servicesMutex.Lock()
	c.registeredServices[service.TunnelServiceId] = service
	c.servicesMutex.Unlock()

	// 更新状态
	c.statusMutex.Lock()
	c.status.RegisteredServices++
	c.statusMutex.Unlock()

	logger.Info("Service registered", map[string]interface{}{
		"serviceId":   service.TunnelServiceId,
		"serviceName": service.ServiceName,
		"serviceType": service.ServiceType,
		"localPort":   service.LocalPort,
	})

	return nil
}

// UnregisterService 注销服务
func (c *tunnelClient) UnregisterService(ctx context.Context, serviceID string) error {
	// 从本地服务列表中查找
	c.servicesMutex.RLock()
	service, exists := c.registeredServices[serviceID]
	c.servicesMutex.RUnlock()

	if !exists {
		return fmt.Errorf("service %s not found", serviceID)
	}

	// 注销服务
	if err := c.serviceManager.UnregisterService(ctx, serviceID); err != nil {
		return fmt.Errorf("failed to unregister service: %w", err)
	}

	// 从本地服务列表中移除
	c.servicesMutex.Lock()
	delete(c.registeredServices, serviceID)
	c.servicesMutex.Unlock()

	// 更新状态
	c.statusMutex.Lock()
	if c.status.RegisteredServices > 0 {
		c.status.RegisteredServices--
	}
	c.statusMutex.Unlock()

	logger.Info("Service unregistered", map[string]interface{}{
		"serviceId":   serviceID,
		"serviceName": service.ServiceName,
	})

	return nil
}

// GetRegisteredServices 获取已注册的服务
func (c *tunnelClient) GetRegisteredServices() []*types.TunnelService {
	c.servicesMutex.RLock()
	defer c.servicesMutex.RUnlock()

	services := make([]*types.TunnelService, 0, len(c.registeredServices))
	for _, service := range c.registeredServices {
		services = append(services, service)
	}

	return services
}

// messageLoop 消息处理循环
func (c *tunnelClient) messageLoop() {
	defer c.wg.Done()

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			// 接收控制消息
			msg, err := c.controlConn.ReceiveMessage(c.ctx)
			if err != nil {
				if c.ctx.Err() != nil {
					return // 正常关闭
				}

				logger.Error("Failed to receive message", map[string]interface{}{
					"error": err.Error(),
				})

				// 触发重连
				c.reconnectManager.TriggerReconnect(c.ctx, "message_receive_error")
				continue
			}

			// 处理消息
			if err := c.handleMessage(msg); err != nil {
				logger.Error("Failed to handle message", map[string]interface{}{
					"messageType": msg.Type,
					"error":       err.Error(),
				})
			}
		}
	}
}

// handleMessage 处理控制消息
func (c *tunnelClient) handleMessage(msg *ControlMessage) error {
	switch msg.Type {
	case MessageTypeNewProxy:
		return c.handleNewProxyMessage(msg)
	case MessageTypeCloseProxy:
		return c.handleCloseProxyMessage(msg)
	case MessageTypeProxyRequest:
		return c.handleProxyRequestMessage(msg)
	case MessageTypePreConnectRequest:
		return c.handlePreConnectRequestMessage(msg)
	case MessageTypeNotification:
		return c.handleNotificationMessage(msg)
	case MessageTypeError:
		return c.handleErrorMessage(msg)
	default:
		logger.Warn("Unknown message type", map[string]interface{}{
			"messageType": msg.Type,
		})
	}

	return nil
}

// handleNewProxyMessage 处理新代理消息
func (c *tunnelClient) handleNewProxyMessage(msg *ControlMessage) error {
	serviceID, ok := msg.Data["serviceId"].(string)
	if !ok {
		return fmt.Errorf("missing serviceId in new proxy message")
	}

	remotePort, ok := msg.Data["remotePort"].(float64)
	if !ok {
		return fmt.Errorf("missing remotePort in new proxy message")
	}

	// 查找服务
	c.servicesMutex.RLock()
	service, exists := c.registeredServices[serviceID]
	c.servicesMutex.RUnlock()

	if !exists {
		return fmt.Errorf("service %s not found for proxy", serviceID)
	}

	// 启动代理
	return c.proxyManager.StartProxy(c.ctx, service, int(remotePort))
}

// handleCloseProxyMessage 处理关闭代理消息
func (c *tunnelClient) handleCloseProxyMessage(msg *ControlMessage) error {
	serviceID, ok := msg.Data["serviceId"].(string)
	if !ok {
		return fmt.Errorf("missing serviceId in close proxy message")
	}

	return c.proxyManager.StopProxy(c.ctx, serviceID)
}

// handleProxyRequestMessage 处理代理请求消息
//
// 当服务器接收到外网连接请求且连接池中没有可用连接时，
// 会发送此消息通知客户端建立新的数据连接。
//
// 参数:
//   - msg: 代理请求控制消息
//
// 返回:
//   - error: 处理失败时返回错误
//
// 工作流程:
//  1. 解析消息中的服务ID和连接ID
//  2. 异步启动数据连接建立过程
//  3. 避免阻塞消息处理循环
//
// 注意:
//   - 此方法立即返回，实际连接建立在后台进行
//   - 连接建立失败会记录错误日志
func (c *tunnelClient) handleProxyRequestMessage(msg *ControlMessage) error {
	serviceID, ok := msg.Data["serviceId"].(string)
	if !ok {
		return fmt.Errorf("missing serviceId in proxy request message")
	}

	connectionID, ok := msg.Data["connectionId"].(string)
	if !ok {
		return fmt.Errorf("missing connectionId in proxy request message")
	}

	logger.Info("Received proxy request", map[string]interface{}{
		"serviceId":    serviceID,
		"connectionId": connectionID,
	})

	// 异步建立数据连接，避免阻塞消息处理循环
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		if err := c.establishDataConnection(serviceID, connectionID); err != nil {
			logger.Error("Failed to establish data connection", map[string]interface{}{
				"serviceId":    serviceID,
				"connectionId": connectionID,
				"error":        err.Error(),
			})
		}
	}()

	return nil
}

// handlePreConnectRequestMessage 处理预连接请求消息
//
// 当服务器的连接池需要补充连接时，会发送此消息请求客户端
// 预先建立数据连接并加入连接池，以提高后续请求的响应速度。
//
// 参数:
//   - msg: 预连接请求控制消息
//
// 返回:
//   - error: 处理失败时返回错误
//
// 工作流程:
//  1. 解析消息中的服务ID和池化标识
//  2. 异步建立池化数据连接
//  3. 连接建立后保持活跃等待服务器使用
//
// 注意:
//   - 池化连接与普通连接的生命周期不同
//   - 连接建立后会长期保持活跃状态
func (c *tunnelClient) handlePreConnectRequestMessage(msg *ControlMessage) error {
	serviceID, ok := msg.Data["serviceId"].(string)
	if !ok {
		return fmt.Errorf("missing serviceId in pre-connect request message")
	}

	isPooled, _ := msg.Data["pooled"].(bool)

	logger.Info("Received pre-connect request", map[string]interface{}{
		"serviceId": serviceID,
		"pooled":    isPooled,
	})

	// 异步建立池化数据连接
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		if err := c.establishPooledDataConnection(serviceID); err != nil {
			logger.Error("Failed to establish pooled data connection", map[string]interface{}{
				"serviceId": serviceID,
				"error":     err.Error(),
			})
		}
	}()

	return nil
}

// establishPooledDataConnection 建立池化数据连接
//
// 为连接池建立数据连接。与普通数据连接不同，池化连接在握手完成后
// 需要保持活跃状态，等待服务器的使用。连接会被服务器加入连接池，
// 当有外网请求时可以被复用。
//
// 参数:
//   - serviceID: 服务唯一标识符
//
// 返回:
//   - error: 建立连接失败时返回错误
//
// 注意:
//   - 池化连接建立后会阻塞等待服务器使用
//   - 连接断开时会自动清理资源
//   - 支持上下文取消和优雅关闭
func (c *tunnelClient) establishPooledDataConnection(serviceID string) error {
	// 查找服务配置
	c.servicesMutex.RLock()
	service, exists := c.registeredServices[serviceID]
	c.servicesMutex.RUnlock()

	if !exists {
		return fmt.Errorf("service %s not found", serviceID)
	}

	// 建立到服务器的数据连接
	serverAddr := net.JoinHostPort(c.config.ServerAddress, fmt.Sprintf("%d", c.config.ServerPort))
	dataConn, err := net.DialTimeout("tcp", serverAddr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to server for pooled data connection: %w", err)
	}

	// 为长连接（如SSE）启用TCP KeepAlive
	if tcpConn, ok := dataConn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
		// 禁用Nagle算法以减少延迟（对流式数据重要）
		tcpConn.SetNoDelay(true)
	}

	logger.Info("Pooled data connection established", map[string]interface{}{
		"serviceId":  serviceID,
		"serverAddr": serverAddr,
	})

	// 发送池化连接握手消息（使用服务ID作为连接ID）
	if err := c.sendPooledDataConnectionHandshake(dataConn, serviceID); err != nil {
		dataConn.Close()
		return fmt.Errorf("failed to send pooled data connection handshake: %w", err)
	}

	// 关键修复：池化连接需要保持活跃并等待服务器使用
	// 当服务器从连接池中取出此连接用于数据传输时，
	// 连接会被传递给代理管理器进行实际的数据转发
	logger.Debug("Pooled connection ready, waiting for server to use", map[string]interface{}{
		"serviceId": serviceID,
	})

	// 使用代理管理器处理这个池化连接
	// 这样连接会保持活跃，等待服务器的数据传输请求
	return c.proxyManager.HandlePooledConnection(c.ctx, dataConn, service)
}

// sendPooledDataConnectionHandshake 发送池化数据连接握手消息
func (c *tunnelClient) sendPooledDataConnectionHandshake(conn net.Conn, serviceID string) error {
	// 创建数据连接标识消息（池化连接使用服务ID）
	handshake := map[string]interface{}{
		"type":         "data_connection",
		"connectionId": serviceID, // 池化连接使用服务ID
		"clientId":     c.config.TunnelClientId,
		"pooled":       true, // 标识这是池化连接
	}

	// 序列化消息
	data, err := json.Marshal(handshake)
	if err != nil {
		return fmt.Errorf("failed to marshal pooled handshake: %w", err)
	}

	// 发送消息长度和内容
	lengthBuf := make([]byte, 4)
	msgLen := len(data)
	lengthBuf[0] = byte(msgLen >> 24)
	lengthBuf[1] = byte(msgLen >> 16)
	lengthBuf[2] = byte(msgLen >> 8)
	lengthBuf[3] = byte(msgLen)

	// 发送长度
	if _, err := conn.Write(lengthBuf); err != nil {
		return fmt.Errorf("failed to write message length: %w", err)
	}

	// 发送数据
	if _, err := conn.Write(data); err != nil {
		return fmt.Errorf("failed to write handshake data: %w", err)
	}

	logger.Debug("Pooled data connection handshake sent", map[string]interface{}{
		"serviceId":  serviceID,
		"messageLen": msgLen,
	})

	return nil
}

// handleNotificationMessage 处理通知消息
func (c *tunnelClient) handleNotificationMessage(msg *ControlMessage) error {
	message, ok := msg.Data["message"].(string)
	if !ok {
		return fmt.Errorf("missing message in notification")
	}

	logger.Info("Server notification", map[string]interface{}{
		"message": message,
	})

	return nil
}

// handleErrorMessage 处理错误消息
func (c *tunnelClient) handleErrorMessage(msg *ControlMessage) error {
	if msg.Error != nil {
		c.addError(fmt.Sprintf("Server error: %s - %s", msg.Error.Code, msg.Error.Message))
		logger.Error("Server error", map[string]interface{}{
			"code":    msg.Error.Code,
			"message": msg.Error.Message,
			"details": msg.Error.Details,
		})
	}

	return nil
}

// updateStatus 更新客户端状态
func (c *tunnelClient) updateStatus(status string, connected bool) {
	c.statusMutex.Lock()
	c.status.Status = status
	c.status.Connected = connected
	c.statusMutex.Unlock()
}

// establishDataConnection 建立数据连接
// 当收到服务器的代理请求时，客户端需要建立一个新的TCP连接到服务器
// 这个连接用于传输实际的数据，而不是控制消息
func (c *tunnelClient) establishDataConnection(serviceID, connectionID string) error {
	// 查找服务配置
	c.servicesMutex.RLock()
	_, exists := c.registeredServices[serviceID]
	c.servicesMutex.RUnlock()

	if !exists {
		return fmt.Errorf("service %s not found", serviceID)
	}

	// 建立到服务器的数据连接
	serverAddr := net.JoinHostPort(c.config.ServerAddress, fmt.Sprintf("%d", c.config.ServerPort))
	dataConn, err := net.DialTimeout("tcp", serverAddr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to server for data connection: %w", err)
	}

	// 为长连接（如SSE）启用TCP KeepAlive
	if tcpConn, ok := dataConn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
		// 禁用Nagle算法以减少延迟（对流式数据重要）
		tcpConn.SetNoDelay(true)
	}

	logger.Info("Data connection established", map[string]interface{}{
		"serviceId":    serviceID,
		"connectionId": connectionID,
		"serverAddr":   serverAddr,
	})

	// 发送数据连接标识消息
	if err := c.sendDataConnectionHandshake(dataConn, connectionID); err != nil {
		dataConn.Close()
		return fmt.Errorf("failed to send data connection handshake: %w", err)
	}

	// 将数据连接交给代理管理器处理
	return c.proxyManager.HandleProxyConnection(c.ctx, dataConn, serviceID)
}

// sendDataConnectionHandshake 发送数据连接握手消息
// 用于告诉服务器这是一个数据连接，并关联到特定的连接ID
func (c *tunnelClient) sendDataConnectionHandshake(conn net.Conn, connectionID string) error {
	// 创建数据连接标识消息
	handshake := map[string]interface{}{
		"type":         "data_connection",
		"connectionId": connectionID,
		"clientId":     c.config.TunnelClientId,
	}

	// 序列化消息
	data, err := json.Marshal(handshake)
	if err != nil {
		return fmt.Errorf("failed to marshal handshake: %w", err)
	}

	// 发送消息长度和内容
	lengthBuf := make([]byte, 4)
	msgLen := len(data)
	lengthBuf[0] = byte(msgLen >> 24)
	lengthBuf[1] = byte(msgLen >> 16)
	lengthBuf[2] = byte(msgLen >> 8)
	lengthBuf[3] = byte(msgLen)

	// 发送长度
	if _, err := conn.Write(lengthBuf); err != nil {
		return fmt.Errorf("failed to write message length: %w", err)
	}

	// 发送数据
	if _, err := conn.Write(data); err != nil {
		return fmt.Errorf("failed to write handshake data: %w", err)
	}

	logger.Debug("Data connection handshake sent", map[string]interface{}{
		"connectionId": connectionID,
		"messageLen":   msgLen,
	})

	return nil
}

// updateConnectTime 更新连接时间
func (c *tunnelClient) updateConnectTime() {
	c.statusMutex.Lock()
	c.status.LastConnectTime = time.Now()
	c.statusMutex.Unlock()
}

// addError 添加错误信息
func (c *tunnelClient) addError(errorMsg string) {
	c.statusMutex.Lock()
	c.status.Errors = append(c.status.Errors, errorMsg)
	// 限制错误数量，只保留最近的10个
	if len(c.status.Errors) > 10 {
		c.status.Errors = c.status.Errors[len(c.status.Errors)-10:]
	}
	c.statusMutex.Unlock()
}

// isConnected 检查是否已连接
func (c *tunnelClient) isConnected() bool {
	return c.controlConn.IsConnected()
}

// generateServiceID 生成服务ID
func (c *tunnelClient) generateServiceID(serviceName string) string {
	return fmt.Sprintf("%s_%s_%d", c.config.TunnelClientId, serviceName, time.Now().UnixNano())
}

// GetControlConnection 获取控制连接（供子组件使用）
func (c *tunnelClient) GetControlConnection() ControlConnection {
	return c.controlConn
}

// GetProxyManager 获取代理管理器（供子组件使用）
func (c *tunnelClient) GetProxyManager() ProxyManager {
	return c.proxyManager
}
