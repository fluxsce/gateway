// Package client 提供控制连接的完整实现
// 控制连接负责与隧道服务器建立和维护控制通道
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"gateway/pkg/logger"
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
	sendChan    chan *ControlMessage
	receiveChan chan *ControlMessage

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
		client:      client,
		connected:   false,
		sendChan:    make(chan *ControlMessage, 100),
		receiveChan: make(chan *ControlMessage, 100),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Connect 连接到服务器控制端口
func (cc *controlConnection) Connect(ctx context.Context, serverAddress string, serverPort int) error {
	cc.connMutex.Lock()
	defer cc.connMutex.Unlock()

	if cc.connected {
		return fmt.Errorf("already connected")
	}

	// 建立TCP连接
	addr := net.JoinHostPort(serverAddress, fmt.Sprintf("%d", serverPort))
	conn, err := net.DialTimeout("tcp", addr, 30*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", addr, err)
	}

	cc.conn = conn
	cc.connected = true

	// 记录连接信息
	localAddr := conn.LocalAddr().(*net.TCPAddr)
	remoteAddr := conn.RemoteAddr().(*net.TCPAddr)

	cc.connInfo = &ConnectionInfo{
		LocalAddress:  localAddr.IP.String(),
		LocalPort:     localAddr.Port,
		RemoteAddress: remoteAddr.IP.String(),
		RemotePort:    remoteAddr.Port,
		ConnectedAt:   time.Now(),
		LastActivity:  time.Now(),
		BytesSent:     0,
		BytesReceived: 0,
	}

	// 启动消息处理协程
	cc.wg.Add(2)
	go cc.sendLoop()
	go cc.receiveLoop()

	// 发送认证消息
	authMsg := &ControlMessage{
		Type:      MessageTypeAuth,
		RequestID: cc.generateRequestID(),
		Data: map[string]interface{}{
			"clientId": cc.client.config.TunnelClientId,
			"token":    cc.client.config.AuthToken,
		},
		Timestamp: time.Now(),
	}

	if err := cc.SendMessage(ctx, authMsg); err != nil {
		cc.disconnect()
		return fmt.Errorf("failed to send auth message: %w", err)
	}

	logger.Info("Control connection established", map[string]interface{}{
		"serverAddress": serverAddress,
		"serverPort":    serverPort,
		"localAddress":  cc.connInfo.LocalAddress,
		"localPort":     cc.connInfo.LocalPort,
	})

	return nil
}

// Disconnect 断开连接
func (cc *controlConnection) Disconnect(ctx context.Context) error {
	cc.connMutex.Lock()
	defer cc.connMutex.Unlock()

	if !cc.connected {
		return nil
	}

	cc.disconnect()

	// 等待协程退出
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

// SendMessage 发送控制消息
func (cc *controlConnection) SendMessage(ctx context.Context, message *ControlMessage) error {
	if !cc.IsConnected() {
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

// ReceiveMessage 接收控制消息
func (cc *controlConnection) ReceiveMessage(ctx context.Context) (*ControlMessage, error) {
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
func (cc *controlConnection) sendMessageDirect(message *ControlMessage) error {
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
func (cc *controlConnection) receiveMessageDirect() (*ControlMessage, error) {
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
	if msgLen <= 0 || msgLen > 1024*1024 { // 限制消息大小为1MB
		return nil, fmt.Errorf("invalid message length: %d", msgLen)
	}

	// 读取消息内容
	msgBuf := make([]byte, msgLen)
	if _, err := io.ReadFull(conn, msgBuf); err != nil {
		return nil, fmt.Errorf("failed to read message data: %w", err)
	}

	// 反序列化消息
	var message ControlMessage
	if err := json.Unmarshal(msgBuf, &message); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	// 更新连接统计
	cc.updateConnectionStats(0, int64(4+msgLen))

	return &message, nil
}

// disconnect 内部断开连接方法
func (cc *controlConnection) disconnect() {
	if cc.conn != nil {
		cc.conn.Close()
		cc.conn = nil
	}

	cc.connected = false
	cc.cancel()
}

// handleConnectionError 处理连接错误
func (cc *controlConnection) handleConnectionError(err error) {
	logger.Error("Connection error occurred", map[string]interface{}{
		"error": err.Error(),
	})

	cc.connMutex.Lock()
	cc.connected = false
	cc.connMutex.Unlock()

	// 通知客户端连接出现问题
	if cc.client.reconnectManager != nil {
		cc.client.reconnectManager.TriggerReconnect(context.Background(), "connection_error")
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
func (cc *controlConnection) generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

// Close 关闭控制连接
func (cc *controlConnection) Close() error {
	cc.cancel()
	cc.wg.Wait()

	cc.connMutex.Lock()
	defer cc.connMutex.Unlock()

	if cc.conn != nil {
		cc.conn.Close()
		cc.conn = nil
	}

	cc.connected = false

	return nil
}
