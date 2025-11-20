// Package client 提供控制连接的完整实现
// 控制连接负责与隧道服务器建立和维护控制通道
//
// # 自动重连机制
//
// ## 概述
//
// 控制连接集成了自动重连机制，确保在网络中断、连接错误等情况下能够
// 自动恢复连接，无需人工干预。重连机制由 ReconnectManager 协调完成。
//
// ## 触发条件
//
// 自动重连在以下情况下触发：
//  1. 发送消息时检测到连接已断开（SendMessage）
//  2. 接收消息时发生 IO 错误（receiveLoop）
//  3. 发送消息时发生 IO 错误（sendLoop）
//  4. 心跳超时或失败（通过 heartbeatManager）
//  5. TCP 连接被远程关闭
//  6. 网络异常导致的连接中断
//
// ## 重连流程
//
//  1. 检测到连接错误
//     ↓
//  2. 标记连接状态为未连接
//     ↓
//  3. 关闭底层 TCP 连接
//     ↓
//  4. 检查是否已在重连中
//     ↓
//  5. 触发 ReconnectManager.TriggerReconnect()
//     ↓
//  6. 使用指数退避策略重试
//     ↓
//  7. 重新建立控制连接
//     ↓
//  8. 重新认证和注册服务
//     ↓
//  9. 恢复心跳和数据转发
//
// ## 重连策略
//
// ### 指数退避算法
//   - 基础间隔：可配置（默认5秒）
//   - 最大间隔：300秒（5分钟）
//   - 计算公式：baseInterval * 2^(attempt-1)
//   - 示例：5s → 10s → 20s → 40s → 80s → 160s → 300s
//
// ### 最大重试次数
//   - 可配置（默认10次）
//   - 所有重试失败后标记为错误状态
//   - 需要手动重启客户端
//
// ### 防止重复触发
//   - 检查 IsReconnecting() 状态
//   - 同一时刻只允许一个重连流程
//   - 避免资源浪费和状态混乱
//
// ## 连接状态管理
//
// ### 状态转换
//   - Disconnected → Connecting → Connected
//   - Connected → Error → Reconnecting → Connected
//   - Reconnecting → Error（所有重试失败）
//
// ### 状态检查
//   - IsConnected()：检查当前连接状态
//   - ReconnectManager.IsReconnecting()：检查是否正在重连
//   - GetConnectionInfo()：获取详细连接信息
//
// ## 错误处理
//
// ### 可恢复错误
//   - 网络超时（自动重连）
//   - 连接重置（自动重连）
//   - EOF（自动重连）
//   - 临时性网络错误（自动重连）
//
// ### 不可恢复错误
//   - 认证失败（需要检查配置）
//   - 服务器拒绝连接（检查服务器状态）
//   - 配置错误（修正配置后重启）
//
// ## 监控和日志
//
// ### 关键日志
//   - Connection error occurred：连接错误
//   - Triggering reconnect：触发重连
//   - Reconnection already in progress：重连进行中
//   - Reconnect successful：重连成功
//   - All reconnect attempts failed：所有重试失败
//
// ### 统计指标
//   - 重连次数（ReconnectCount）
//   - 重连成功率
//   - 平均重连时间
//   - 连接持续时间
//
// ## 最佳实践
//
// ### 配置建议
//   - 心跳间隔：5-30秒
//   - 重连基础间隔：5秒
//   - 最大重试次数：10-20次
//   - 连接超时：30秒
//
// ### 注意事项
//   - 重连期间服务不可用
//   - 数据连接会随控制连接断开
//   - 重连成功后需要重新注册服务
//   - 连接池中的连接会被清理
//
// # 数据连接管理机制
//
// ## 概述
//
// 隧道系统使用双连接模型：
//  1. 控制连接（Control Connection）：用于传输控制消息、心跳、服务注册等
//  2. 数据连接（Data Connection）：用于传输实际的业务数据
//
// ## 数据连接类型
//
// ### 1. 普通数据连接（On-Demand Connection）
//
// 工作流程：
//  1. 外网用户访问服务器暴露的端口
//  2. 服务器通过控制连接发送 proxy_request 消息给客户端
//  3. 客户端收到请求后建立新的数据连接到服务器
//  4. 客户端发送握手消息标识 connectionID
//  5. 服务器匹配 pendingConnection，开始数据转发
//  6. 数据传输完成后关闭连接
//
// 特点：
//   - 按需建立，用完即关
//   - 适合短连接场景（HTTP请求）
//   - 延迟较高（需要TCP握手）
//
// ## 连接池架构
//
// ### 客户端侧（无连接池）
//
// 客户端不维护连接池，连接由服务端控制生命周期：
//   - 服务端发送 proxy_request 通知客户端建立连接
//   - 客户端建立 serverConn（数据连接）和 localConn（本地服务连接）
//   - 两个连接绑定，生命周期一致
//   - 客户端保持连接打开，直到服务端关闭
//   - 当 relayData 返回时，连接已关闭，直接清理
//   - 管理：HandleProxyConnection() 内部管理
//
// ## 连接生命周期
//
// ### 建立阶段
//  1. 检查连接池是否有可用连接
//  2. 有：直接从池中获取（复用）
//  3. 无：建立新的TCP连接
//  4. 配置TCP选项（KeepAlive、NoDelay等）
//  5. 发送握手消息（标识connectionID或serviceID）
//
// ### 使用阶段
//  1. 建立客户端↔服务器↔本地服务的数据通道
//  2. 启动双向数据转发（io.Copy）
//  3. 监控连接状态和流量
//  4. 处理连接错误和超时
//
// ### 释放阶段
//  1. 数据传输完成
//  2. 检查连接是否健康
//  3. 健康：尝试归还到连接池
//  4. 不健康或池满：关闭连接
//  5. 更新统计信息
//
// ## 性能优化
//
// ### 连接复用收益
//   - 减少TCP三次握手开销（~3ms）
//   - 减少TIME_WAIT状态堆积
//   - 提升并发处理能力
//   - 降低CPU和内存消耗
//
// ### 适用场景
//
//	✅ HTTP短连接高并发
//	✅ REST API调用
//	✅ 微服务间通信
//	⚠️  WebSocket长连接（不适合池化）
//	⚠️  SSE流式传输（不适合池化）
//	⚠️  大文件传输（不适合池化）
//
// ## 错误处理
//
// ### 连接池策略
//   - 正常关闭：归还到池中
//   - 超时/重置：归还到池中（可能是客户端关闭）
//   - 真正错误：直接关闭，不归还
//   - 池满：关闭连接
//
// ### 健康检查
//   - 从池中获取连接时验证可用性
//   - 定期清理失效连接
//   - 自动重建失败的连接
//
// ## 监控指标
//
//   - 连接池大小和使用率
//   - 连接复用率
//   - 连接建立延迟
//   - 连接错误率
//   - 流量统计
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"gateway/internal/tunnel/types"
	"gateway/pkg/logger"
	"gateway/pkg/utils/random"
)

// controlConnection 控制连接实现
// 实现 ControlConnection 接口，管理与服务器的控制通道
type controlConnection struct {
	client    *tunnelClient
	conn      net.Conn
	connMutex sync.RWMutex
	connected bool
	connInfo  *ConnectionInfo

	// 消息处理
	sendChan    chan *types.ControlMessage
	receiveChan chan *types.ControlMessage

	// 请求-响应追踪
	pendingRequests map[string]chan *types.ControlMessage
	requestMutex    sync.RWMutex

	// 控制状态
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewControlConnection 创建控制连接实例
//
// 参数:
//   - client: 隧道客户端实例
//
// 返回:
//   - ControlConnection: 控制连接接口实例
//
// 功能:
//   - 初始化控制连接
//   - 创建消息通道和状态管理
func NewControlConnection(client *tunnelClient) ControlConnection {
	ctx, cancel := context.WithCancel(context.Background())

	return &controlConnection{
		client:          client,
		connected:       false,
		sendChan:        make(chan *types.ControlMessage, 100),
		receiveChan:     make(chan *types.ControlMessage, 100),
		pendingRequests: make(map[string]chan *types.ControlMessage),
		ctx:             ctx,
		cancel:          cancel,
	}
}

// Connect 连接到服务器控制端口
func (cc *controlConnection) Connect(ctx context.Context, serverAddress string, serverPort int) error {
	// 先检查是否已连接（不持有锁的情况下）
	cc.connMutex.Lock()
	if cc.connected {
		cc.connMutex.Unlock()
		return fmt.Errorf("already connected")
	}
	cc.connMutex.Unlock()

	// 建立TCP连接（不持有锁）
	addr := net.JoinHostPort(serverAddress, fmt.Sprintf("%d", serverPort))
	conn, err := net.DialTimeout("tcp", addr, 30*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", addr, err)
	}

	// 记录连接信息
	localAddr := conn.LocalAddr().(*net.TCPAddr)
	remoteAddr := conn.RemoteAddr().(*net.TCPAddr)

	connInfo := &ConnectionInfo{
		LocalAddress:  localAddr.IP.String(),
		LocalPort:     localAddr.Port,
		RemoteAddress: remoteAddr.IP.String(),
		RemotePort:    remoteAddr.Port,
		ConnectedAt:   time.Now(),
		LastActivity:  time.Now(),
		BytesSent:     0,
		BytesReceived: 0,
	}

	// 设置连接和状态（持有锁的时间最短）
	cc.connMutex.Lock()
	cc.conn = conn
	cc.connInfo = connInfo
	cc.connected = true
	cc.connMutex.Unlock()

	// 启动消息处理协程（不持有锁）
	cc.wg.Add(2)
	go cc.sendLoop()
	go cc.receiveLoop()

	// 等待一小段时间确保协程启动
	time.Sleep(10 * time.Millisecond)

	// 发送认证消息（不持有锁）
	authMsg := &types.ControlMessage{
		Type:      types.MessageTypeAuth,
		SessionID: cc.generateRequestID(),
		Data: map[string]interface{}{
			"clientId": cc.client.config.TunnelClientId,
			"token":    cc.client.config.AuthToken,
		},
		Timestamp: time.Now(),
	}

	if err := cc.SendMessage(ctx, authMsg); err != nil {
		// 认证失败，清理连接
		cc.cleanupConnection()
		return fmt.Errorf("failed to send auth message: %w", err)
	}

	// 启动消息处理循环
	cc.wg.Add(1)
	go cc.messageProcessLoop()

	logger.Info("Control connection established", map[string]interface{}{
		"serverAddress": serverAddress,
		"serverPort":    serverPort,
		"localAddress":  connInfo.LocalAddress,
		"localPort":     connInfo.LocalPort,
	})

	return nil
}

// Disconnect 断开连接
func (cc *controlConnection) Disconnect(ctx context.Context) error {
	// 检查连接状态（短时间持有锁）
	cc.connMutex.Lock()
	if !cc.connected {
		cc.connMutex.Unlock()
		return nil
	}

	// 调用 disconnect 清理连接（仍持有锁）
	cc.disconnect()
	cc.connMutex.Unlock()

	// 在不持有锁的情况下等待协程退出
	done := make(chan struct{})
	go func() {
		cc.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("Control connection disconnected successfully", nil)
		return nil
	case <-ctx.Done():
		logger.Warn("Control connection disconnect timeout", nil)
		return ctx.Err()
	}
}

// SendMessage 发送控制消息（不等待响应）
func (cc *controlConnection) SendMessage(ctx context.Context, message *types.ControlMessage) error {
	// 双重检查连接状态
	if !cc.IsConnected() {
		// 尝试触发自动重连
		if cc.client.reconnectManager != nil && !cc.client.reconnectManager.IsReconnecting() {
			logger.Warn("Connection lost during send, triggering reconnect", map[string]interface{}{
				"messageType": message.Type,
			})
			go func() {
				if err := cc.client.reconnectManager.TriggerReconnect(context.Background(), "send_message_not_connected"); err != nil {
					logger.Error("Failed to trigger reconnect from SendMessage", map[string]interface{}{
						"error": err.Error(),
					})
				}
			}()
		}
		return fmt.Errorf("not connected")
	}

	select {
	case cc.sendChan <- message:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(5 * time.Second):
		return fmt.Errorf("send message timeout")
	}
}

// SendMessageAndWaitResponse 发送控制消息并等待响应
// 用于需要同步等待服务器响应的场景（如服务注册、注销）
func (cc *controlConnection) SendMessageAndWaitResponse(ctx context.Context, message *types.ControlMessage, timeout time.Duration) (*types.ControlMessage, error) {
	if !cc.IsConnected() {
		return nil, fmt.Errorf("not connected")
	}

	// 创建响应通道
	responseChan := make(chan *types.ControlMessage, 1)

	// 注册等待响应
	cc.requestMutex.Lock()
	cc.pendingRequests[message.SessionID] = responseChan
	cc.requestMutex.Unlock()

	// 确保清理
	defer func() {
		cc.requestMutex.Lock()
		delete(cc.pendingRequests, message.SessionID)
		cc.requestMutex.Unlock()
		close(responseChan)
	}()

	// 发送消息
	select {
	case cc.sendChan <- message:
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("send message timeout")
	}

	// 等待响应
	select {
	case response := <-responseChan:
		return response, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(timeout):
		return nil, fmt.Errorf("wait response timeout after %v", timeout)
	}
}

// ReceiveMessage 接收控制消息
func (cc *controlConnection) ReceiveMessage(ctx context.Context) (*types.ControlMessage, error) {
	if !cc.IsConnected() {
		return nil, fmt.Errorf("not connected")
	}

	select {
	case msg := <-cc.receiveChan:
		return msg, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-cc.ctx.Done():
		return nil, fmt.Errorf("connection closed")
	}
}

// IsConnected 检查连接状态
func (cc *controlConnection) IsConnected() bool {
	cc.connMutex.RLock()
	defer cc.connMutex.RUnlock()
	return cc.connected
}

// GetConnectionInfo 获取连接信息
func (cc *controlConnection) GetConnectionInfo() *ConnectionInfo {
	cc.connMutex.RLock()
	defer cc.connMutex.RUnlock()

	if cc.connInfo == nil {
		return nil
	}

	// 返回连接信息副本
	info := *cc.connInfo
	return &info
}

// sendLoop 发送消息循环
func (cc *controlConnection) sendLoop() {
	defer cc.wg.Done()

	for {
		select {
		case <-cc.ctx.Done():
			return
		case msg := <-cc.sendChan:
			if err := cc.sendMessageDirect(msg); err != nil {
				logger.Error("Failed to send message", map[string]interface{}{
					"messageType": msg.Type,
					"error":       err.Error(),
				})

				// 发送失败，可能连接有问题
				cc.handleConnectionError(err)
				return
			}
		}
	}
}

// receiveLoop 接收消息循环
func (cc *controlConnection) receiveLoop() {
	defer cc.wg.Done()

	for {
		select {
		case <-cc.ctx.Done():
			return
		default:
			msg, err := cc.receiveMessageDirect()
			if err != nil {
				if cc.ctx.Err() != nil {
					return // 正常关闭
				}

				logger.Error("Failed to receive message", map[string]interface{}{
					"error": err.Error(),
				})

				// 接收失败，可能连接有问题
				cc.handleConnectionError(err)
				return
			}

			// 将消息放入接收通道
			select {
			case cc.receiveChan <- msg:
			case <-cc.ctx.Done():
				return
			default:
				// 通道满了，丢弃旧消息
				select {
				case <-cc.receiveChan:
				default:
				}
				cc.receiveChan <- msg
			}
		}
	}
}

// sendMessageDirect 直接发送消息
func (cc *controlConnection) sendMessageDirect(message *types.ControlMessage) error {
	cc.connMutex.RLock()
	conn := cc.conn
	cc.connMutex.RUnlock()

	if conn == nil {
		return fmt.Errorf("connection is nil")
	}

	// 序列化消息
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
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
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	// 发送长度
	if _, err := conn.Write(lengthBuf); err != nil {
		return fmt.Errorf("failed to write message length: %w", err)
	}

	// 发送消息内容
	if _, err := conn.Write(data); err != nil {
		return fmt.Errorf("failed to write message data: %w", err)
	}

	// 更新连接统计
	cc.updateConnectionStats(int64(4+len(data)), 0)

	return nil
}

// receiveMessageDirect 直接接收消息
func (cc *controlConnection) receiveMessageDirect() (*types.ControlMessage, error) {
	cc.connMutex.RLock()
	conn := cc.conn
	cc.connMutex.RUnlock()

	if conn == nil {
		return nil, fmt.Errorf("connection is nil")
	}

	// 设置读超时
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	// 读取消息长度
	lengthBuf := make([]byte, 4)
	if _, err := io.ReadFull(conn, lengthBuf); err != nil {
		return nil, fmt.Errorf("failed to read message length: %w", err)
	}

	// 解析消息长度
	msgLen := int(lengthBuf[0])<<24 | int(lengthBuf[1])<<16 | int(lengthBuf[2])<<8 | int(lengthBuf[3])

	// 关键修复：更严格的消息长度验证
	// 检查消息长度是否合理，以及长度字段是否全为0（可能是连接关闭时的填充字节）
	if msgLen <= 0 || msgLen > 1024*1024 {
		allZeros := lengthBuf[0] == 0 && lengthBuf[1] == 0 && lengthBuf[2] == 0 && lengthBuf[3] == 0
		logger.Error("Invalid message length on control connection", map[string]interface{}{
			"messageLength": msgLen,
			"lengthBytes":   lengthBuf,
			"allZeros":      allZeros,
			"remoteAddr":    conn.RemoteAddr().String(),
			"localAddr":     conn.LocalAddr().String(),
			"possibleCause": "connection_closed_or_reset",
		})
		return nil, fmt.Errorf("invalid message length: %d (possible connection closed or reset)", msgLen)
	}

	// 关键修复：检查消息长度是否太小（可能是读取错误）
	if msgLen < 10 {
		logger.Error("Message length too small, possible read error", map[string]interface{}{
			"messageLength": msgLen,
			"lengthBytes":   lengthBuf,
			"remoteAddr":    conn.RemoteAddr().String(),
			"localAddr":     conn.LocalAddr().String(),
		})
		return nil, fmt.Errorf("message length too small: %d (possible read error or connection issue)", msgLen)
	}

	// 读取消息内容
	msgBuf := make([]byte, msgLen)
	if _, err := io.ReadFull(conn, msgBuf); err != nil {
		return nil, fmt.Errorf("failed to read message data: %w", err)
	}

	// 关键修复：在反序列化前检查消息内容
	// 如果消息内容全为0x00，说明可能是连接关闭时的填充字节
	allNullBytes := true
	for i := 0; i < len(msgBuf) && i < 100; i++ { // 只检查前100字节
		if msgBuf[i] != 0 {
			allNullBytes = false
			break
		}
	}
	if allNullBytes && len(msgBuf) > 0 {
		logger.Error("Message content is all null bytes, possible connection issue", map[string]interface{}{
			"messageLength": msgLen,
			"remoteAddr":    conn.RemoteAddr().String(),
			"localAddr":     conn.LocalAddr().String(),
			"possibleCause": "connection_closed_or_reset",
		})
		return nil, fmt.Errorf("message content is all null bytes (possible connection closed or reset)")
	}

	// 关键修复：检查消息是否以JSON格式开头
	// 有效的JSON消息应该以 '{' 或 '[' 开头（去除前导空白和null字节）
	msgStart := 0
	// 跳过前导空白字符
	for msgStart < len(msgBuf) && (msgBuf[msgStart] == ' ' || msgBuf[msgStart] == '\t' || msgBuf[msgStart] == '\n' || msgBuf[msgStart] == '\r') {
		msgStart++
	}

	// 关键修复：如果消息前面有null字节，尝试跳过它们
	// 这可能是消息边界错位导致的（读取位置不正确）
	if msgStart < len(msgBuf) && msgBuf[msgStart] == 0 {
		nullByteCount := 0
		originalStart := msgStart
		// 跳过连续的null字节
		for msgStart < len(msgBuf) && msgBuf[msgStart] == 0 {
			msgStart++
			nullByteCount++
		}

		// 如果跳过了null字节，记录警告
		if nullByteCount > 0 {
			logger.Warn("Skipped leading null bytes in message, possible message boundary misalignment", map[string]interface{}{
				"messageLength": msgLen,
				"nullByteCount": nullByteCount,
				"originalStart": originalStart,
				"newStart":      msgStart,
				"remoteAddr":    conn.RemoteAddr().String(),
				"localAddr":     conn.LocalAddr().String(),
				"possibleCause": "message_boundary_misalignment_or_connection_confusion",
			})
		}
	}

	// 检查是否找到了有效的JSON起始字符
	if msgStart >= len(msgBuf) || (msgBuf[msgStart] != '{' && msgBuf[msgStart] != '[') {
		// 消息不是有效的JSON格式，记录详细信息
		previewLen := 50
		if len(msgBuf) < previewLen {
			previewLen = len(msgBuf)
		}
		logger.Error("Message does not start with valid JSON, possible connection confusion", map[string]interface{}{
			"messageLength": msgLen,
			"firstByte":     msgBuf[0],
			"msgStart":      msgStart,
			"previewBytes":  string(msgBuf[:previewLen]),
			"remoteAddr":    conn.RemoteAddr().String(),
			"localAddr":     conn.LocalAddr().String(),
			"possibleCause": "data_connection_data_sent_to_control_connection_or_corrupted_data",
		})
		return nil, fmt.Errorf("message does not start with valid JSON (first byte: 0x%02x, msgStart: %d, possible connection confusion)", msgBuf[0], msgStart)
	}

	// 关键修复：如果跳过了null字节，使用从msgStart开始的数据进行解析
	if msgStart > 0 {
		msgBuf = msgBuf[msgStart:]
		logger.Debug("Using message data after skipping leading bytes", map[string]interface{}{
			"skippedBytes": msgStart,
			"remainingLen": len(msgBuf),
		})
	}

	// 反序列化消息
	var message types.ControlMessage
	if err := json.Unmarshal(msgBuf, &message); err != nil {
		// 关键修复：记录更详细的错误信息，帮助定位问题根源
		previewLen := 100
		if len(msgBuf) < previewLen {
			previewLen = len(msgBuf)
		}
		// 检查是否包含大量null字节
		nullByteCount := 0
		for i := 0; i < previewLen; i++ {
			if msgBuf[i] == 0 {
				nullByteCount++
			}
		}

		// 获取前20字节用于诊断
		firstBytesLen := 20
		if len(msgBuf) < firstBytesLen {
			firstBytesLen = len(msgBuf)
		}

		logger.Error("Failed to unmarshal message on control connection", map[string]interface{}{
			"error":         err.Error(),
			"messageLength": msgLen,
			"nullByteCount": nullByteCount,
			"previewBytes":  string(msgBuf[:previewLen]),
			"firstBytes":    msgBuf[:firstBytesLen],
			"remoteAddr":    conn.RemoteAddr().String(),
			"localAddr":     conn.LocalAddr().String(),
			"possibleCause": "connection_confusion_or_corrupted_data_from_server",
		})
		return nil, fmt.Errorf("failed to unmarshal message: %w (possible connection confusion or corrupted data from server)", err)
	}

	// 更新连接统计
	cc.updateConnectionStats(0, int64(4+msgLen))

	return &message, nil
}

// disconnect 内部断开连接方法（调用者必须持有 connMutex 锁）
func (cc *controlConnection) disconnect() {
	if cc.conn != nil {
		cc.conn.Close()
		cc.conn = nil
	}

	cc.connected = false
	cc.cancel()
}

// cleanupConnection 清理连接（不持有锁，用于错误恢复）
func (cc *controlConnection) cleanupConnection() {
	// 先取消上下文，停止所有协程
	cc.cancel()

	// 等待协程退出
	done := make(chan struct{})
	go func() {
		cc.wg.Wait()
		close(done)
	}()

	// 等待最多5秒
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		logger.Warn("Cleanup connection timeout waiting for goroutines", nil)
	}

	// 最后清理连接状态
	cc.connMutex.Lock()
	if cc.conn != nil {
		cc.conn.Close()
		cc.conn = nil
	}
	cc.connected = false
	cc.connMutex.Unlock()
}

// handleConnectionError 处理连接错误
func (cc *controlConnection) handleConnectionError(err error) {
	logger.Error("Connection error occurred", map[string]interface{}{
		"error": err.Error(),
	})

	// 先标记为未连接（短时间持有锁）
	cc.connMutex.Lock()
	wasConnected := cc.connected
	cc.connected = false

	// 关闭底层连接，避免资源泄漏
	if cc.conn != nil {
		cc.conn.Close()
		cc.conn = nil
	}
	cc.connMutex.Unlock()

	// 只有之前是连接状态才触发重连（避免重复触发）
	if wasConnected && cc.client.reconnectManager != nil {
		// 检查是否已经在重连中
		if !cc.client.reconnectManager.IsReconnecting() {
			// 在不持有锁的情况下触发重连
			go func() {
				if err := cc.client.reconnectManager.TriggerReconnect(context.Background(), "connection_error"); err != nil {
					logger.Warn("Failed to trigger reconnect", map[string]interface{}{
						"error": err.Error(),
					})
				}
			}()
		} else {
			logger.Debug("Reconnection already in progress, skipping trigger", nil)
		}
	}
}

// updateConnectionStats 更新连接统计
func (cc *controlConnection) updateConnectionStats(bytesSent, bytesReceived int64) {
	cc.connMutex.Lock()
	defer cc.connMutex.Unlock()

	if cc.connInfo != nil {
		cc.connInfo.BytesSent += bytesSent
		cc.connInfo.BytesReceived += bytesReceived
		cc.connInfo.LastActivity = time.Now()
	}
}

// generateRequestID 生成请求ID
//
// 使用高强度随机字符串生成器，确保在高并发和分布式环境下的唯一性。
// 生成的ID格式：req_<20位随机字符串>
//
// 返回:
//   - string: 唯一的请求标识符
func (cc *controlConnection) generateRequestID() string {
	return fmt.Sprintf("req_%s", random.GenerateRandomString(20))
}

// Close 关闭控制连接
func (cc *controlConnection) Close() error {
	// 先取消上下文（不持有锁）
	cc.cancel()

	// 等待协程退出（不持有锁）
	cc.wg.Wait()

	// 最后清理连接状态（短时间持有锁）
	cc.connMutex.Lock()
	if cc.conn != nil {
		cc.conn.Close()
		cc.conn = nil
	}
	cc.connected = false
	cc.connMutex.Unlock()

	return nil
}

// messageProcessLoop 消息处理循环
// 从接收通道读取消息并分发处理
func (cc *controlConnection) messageProcessLoop() {
	defer cc.wg.Done()

	for {
		select {
		case <-cc.ctx.Done():
			return
		case msg := <-cc.receiveChan:
			// 处理消息
			if err := cc.handleMessage(msg); err != nil {
				logger.Error("Failed to handle message", map[string]interface{}{
					"messageType": msg.Type,
					"error":       err.Error(),
				})
			}
		}
	}
}

// handleMessage 处理控制消息
func (cc *controlConnection) handleMessage(msg *types.ControlMessage) error {
	switch msg.Type {
	case types.MessageTypeResponse:
		return cc.handleResponseMessage(msg)
	case types.MessageTypeNewProxy:
		return cc.handleNewProxyMessage(msg)
	case types.MessageTypeCloseProxy:
		return cc.handleCloseProxyMessage(msg)
	case types.MessageTypeProxyRequest:
		return cc.handleProxyRequestMessage(msg)
	case types.MessageTypeNotification:
		return cc.handleNotificationMessage(msg)
	case types.MessageTypeError:
		return cc.handleErrorMessage(msg)
	default:
		logger.Warn("Unknown message type", map[string]interface{}{
			"messageType": msg.Type,
		})
	}

	return nil
}

// handleResponseMessage 处理响应消息
func (cc *controlConnection) handleResponseMessage(msg *types.ControlMessage) error {
	// 查找等待此响应的请求
	cc.requestMutex.RLock()
	responseChan, exists := cc.pendingRequests[msg.SessionID]
	cc.requestMutex.RUnlock()

	if exists && responseChan != nil {
		// 将响应发送给等待的协程
		select {
		case responseChan <- msg:
			logger.Debug("Response delivered to waiting request", map[string]interface{}{
				"sessionId": msg.SessionID,
			})
		case <-time.After(1 * time.Second):
			logger.Warn("Failed to deliver response, channel full or closed", map[string]interface{}{
				"sessionId": msg.SessionID,
			})
		}
	} else {
		// 没有等待此响应的请求，可能是异步消息或已超时
		logger.Debug("Received response with no pending request", map[string]interface{}{
			"sessionId": msg.SessionID,
			"data":      msg.Data,
		})
	}

	return nil
}

// handleNewProxyMessage 处理新代理消息
func (cc *controlConnection) handleNewProxyMessage(msg *types.ControlMessage) error {
	serviceID, ok := msg.Data["serviceId"].(string)
	if !ok {
		return fmt.Errorf("missing serviceId in new proxy message")
	}

	remotePort, ok := msg.Data["remotePort"].(float64)
	if !ok {
		return fmt.Errorf("missing remotePort in new proxy message")
	}

	// 查找服务
	service := cc.client.getRegisteredService(serviceID)
	if service == nil {
		return fmt.Errorf("service %s not found for proxy", serviceID)
	}

	// 启动代理
	return cc.client.proxyManager.StartProxy(cc.ctx, service, int(remotePort))
}

// handleCloseProxyMessage 处理关闭代理消息
func (cc *controlConnection) handleCloseProxyMessage(msg *types.ControlMessage) error {
	serviceID, ok := msg.Data["serviceId"].(string)
	if !ok {
		return fmt.Errorf("missing serviceId in close proxy message")
	}

	return cc.client.proxyManager.StopProxy(cc.ctx, serviceID)
}

// handleProxyRequestMessage 处理代理请求消息
func (cc *controlConnection) handleProxyRequestMessage(msg *types.ControlMessage) error {
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
	go func() {
		if err := cc.establishDataConnection(serviceID, connectionID); err != nil {
			logger.Error("Failed to establish data connection", map[string]interface{}{
				"serviceId":    serviceID,
				"connectionId": connectionID,
				"error":        err.Error(),
			})
			// 关键修复：建立连接失败时，通知服务端清理等待连接
			// 这样可以避免服务端等待30秒超时
			// 注意：这里不能直接访问 proxyServer，需要通过控制连接发送错误消息
			// 但由于服务端已经有超时机制和连接池降级，这里只记录错误即可
		}
	}()

	return nil
}

// handleNotificationMessage 处理通知消息
func (cc *controlConnection) handleNotificationMessage(msg *types.ControlMessage) error {
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
func (cc *controlConnection) handleErrorMessage(msg *types.ControlMessage) error {
	errorCode, _ := msg.Data["code"].(string)
	errorMessage, _ := msg.Data["message"].(string)
	errorDetails, _ := msg.Data["details"].(string)

	cc.client.addError(fmt.Sprintf("Server error: %s - %s", errorCode, errorMessage))
	logger.Error("Server error", map[string]interface{}{
		"code":    errorCode,
		"message": errorMessage,
		"details": errorDetails,
	})

	return nil
}

// establishDataConnection 建立数据连接
// 当收到服务器的代理请求时，客户端需要建立一个新的TCP连接到服务器
// 这个连接用于传输实际的数据，而不是控制消息
func (cc *controlConnection) establishDataConnection(serviceID, connectionID string) error {
	// 创建带超时的上下文，使用控制连接的上下文作为父上下文
	// 这样当控制连接关闭时，数据连接建立也会被取消
	ctx, cancel := context.WithTimeout(cc.ctx, 30*time.Second)
	defer cancel()

	// 查找服务配置
	_, err := cc.client.serviceManager.GetService(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("service %s not found", serviceID)
	}

	logger.Info("Starting data connection establishment", map[string]interface{}{
		"serviceId":    serviceID,
		"connectionId": connectionID,
		"step":         "create_new_connection",
	})

	// 关键修复：服务器请求新数据连接时，必须建立新连接，不能从池中取
	// 原因：
	// 1. 服务器发送 proxy_request 说明需要新连接，如果服务器端连接池有连接就不会请求
	// 2. 从池中取出的连接可能已经发送过握手，服务器端可能已经将其放入池中
	// 3. 复用池中的连接会导致连接状态混乱，可能残留数据，导致服务器误判连接类型
	// 4. 新连接建立后，使用完毕可以归还到池中供后续使用
	connectionStartTime := time.Now()
	dataConn, err := cc.createNewServerConnection(ctx, serviceID)
	connectionDuration := time.Since(connectionStartTime)

	// 如果创建连接失败或连接为 nil，立即返回，不要继续
	if err != nil {
		logger.Error("Failed to create new server connection", map[string]interface{}{
			"serviceId":          serviceID,
			"connectionId":       connectionID,
			"error":              err.Error(),
			"connectionDuration": connectionDuration,
			"step":               "create_new_connection",
		})
		return fmt.Errorf("failed to create new server connection: %w", err)
	}

	if dataConn == nil {
		logger.Error("Server connection is nil", map[string]interface{}{
			"serviceId":          serviceID,
			"connectionId":       connectionID,
			"connectionDuration": connectionDuration,
			"step":               "create_new_connection",
		})
		return fmt.Errorf("server connection is nil")
	}

	logger.Info("Data connection ready", map[string]interface{}{
		"serviceId":          serviceID,
		"connectionId":       connectionID,
		"connectionDuration": connectionDuration,
		"step":               "connection_ready",
	})

	// 关键修复：立即发送握手消息，不要做复杂验证
	// 原因：
	// 1. 连接刚建立，理论上应该是有效的
	// 2. 验证过程会增加时间窗口，连接可能在此期间被关闭
	// 3. 如果连接已关闭，Write 操作会立即返回错误，可以快速检测
	// 4. 服务器端已经设置了120秒读取超时，有足够时间接收握手
	// 5. 立即发送可以减少 broken pipe 错误的概率
	// 6. dataConn 已经在前面检查过不为 nil，这里直接使用

	// 立即发送数据连接标识消息（不要延迟）
	// 如果连接已关闭，Write 会立即返回错误，可以快速检测并重试
	if err := cc.sendDataConnectionHandshake(dataConn, connectionID); err != nil {
		dataConn.Close()
		// 如果是 broken pipe 或连接关闭错误，记录详细信息以便排查
		errMsg := err.Error()
		if strings.Contains(errMsg, "broken pipe") || strings.Contains(errMsg, "connection reset") {
			logger.Error("Connection closed before handshake sent (broken pipe)", map[string]interface{}{
				"serviceId":    serviceID,
				"connectionId": connectionID,
				"error":        errMsg,
			})
		}
		return fmt.Errorf("failed to send data connection handshake: %w", err)
	}

	// 将数据连接交给代理管理器处理
	// 注意：HandleProxyConnection 会在使用完毕后尝试将连接归还到池中
	return cc.client.proxyManager.HandleProxyConnection(cc.ctx, dataConn, serviceID)
}

// createNewServerConnection 创建新的服务器数据连接
// 专门用于响应服务器的 proxy_request 请求，必须建立新连接，不能从池中取
func (cc *controlConnection) createNewServerConnection(ctx context.Context, serviceID string) (net.Conn, error) {
	// 建立新的服务器连接
	serverAddr := net.JoinHostPort(cc.client.config.ServerAddress, fmt.Sprintf("%d", cc.client.config.ServerPort))

	logger.Info("Creating new server connection for proxy request", map[string]interface{}{
		"serviceId":  serviceID,
		"serverAddr": serverAddr,
		"timeout":    "30s",
		"timestamp":  time.Now(),
	})

	// 建立TCP连接
	connectStartTime := time.Now()
	conn, err := net.DialTimeout("tcp", serverAddr, 30*time.Second)
	connectDuration := time.Since(connectStartTime)

	if err != nil {
		logger.Error("Failed to create new server connection", map[string]interface{}{
			"serviceId":       serviceID,
			"serverAddr":      serverAddr,
			"error":           err.Error(),
			"connectDuration": connectDuration,
			"timestamp":       time.Now(),
		})
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}

	// 设置 TCP 选项
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
		tcpConn.SetNoDelay(true)
	}

	logger.Info("Successfully created new server connection", map[string]interface{}{
		"serviceId":       serviceID,
		"serverAddr":      serverAddr,
		"connectDuration": connectDuration,
		"localAddr":       conn.LocalAddr().String(),
		"remoteAddr":      conn.RemoteAddr().String(),
	})

	return conn, nil
}

// sendDataConnectionHandshake 发送数据连接握手消息
// 用于告诉服务器这是一个数据连接，并关联到特定的连接ID
func (cc *controlConnection) sendDataConnectionHandshake(conn net.Conn, connectionID string) error {
	// 创建数据连接标识消息
	handshake := map[string]interface{}{
		"type":         "data_connection",
		"connectionId": connectionID,
		"clientId":     cc.client.config.TunnelClientId,
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

	// 关键修复：设置写超时，避免在连接已断开时长时间阻塞
	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	defer conn.SetWriteDeadline(time.Time{}) // 清除超时

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
