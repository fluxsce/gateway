// Package types 定义隧道系统的消息类型
// 本文件包含客户端和服务端之间的所有控制消息定义
// 基于FRP架构设计，支持双向通信
package types

import (
	"time"
)

// ============================================================
// 消息类型常量定义 - Message Type Constants
// ============================================================

const (
	// ============================================================
	// 客户端 → 服务端消息类型 (Client to Server)
	// ============================================================

	// MessageTypeAuth 客户端认证请求
	// 方向: Client → Server
	// 说明: 客户端连接后首先发送认证请求，验证身份
	MessageTypeAuth = "auth"

	// MessageTypeHeartbeat 客户端心跳消息
	// 方向: Client → Server
	// 说明: 客户端定期发送心跳，保持连接活跃
	MessageTypeHeartbeat = "heartbeat"

	// MessageTypeRegisterService 客户端注册服务请求
	// 方向: Client → Server
	// 说明: 客户端请求注册一个新的隧道服务
	MessageTypeRegisterService = "register_service"

	// MessageTypeUnregisterService 客户端注销服务请求
	// 方向: Client → Server
	// 说明: 客户端请求注销一个已注册的隧道服务
	MessageTypeUnregisterService = "unregister_service"

	// ============================================================
	// 服务端 → 客户端消息类型 (Server to Client)
	// ============================================================

	// MessageTypeResponse 服务端响应消息（通用响应）
	// 方向: Server → Client
	// 说明: 服务端对客户端请求的通用响应
	MessageTypeResponse = "response"

	// MessageTypeProxyRequest 服务端请求客户端建立数据连接
	// 方向: Server → Client
	// 说明: 当有外网请求到达时，服务端通知客户端建立数据连接进行转发
	MessageTypeProxyRequest = "proxy_request"

	// MessageTypeNotification 服务端通知消息
	// 方向: Server → Client
	// 说明: 服务端向客户端发送的通知消息（非关键操作）
	MessageTypeNotification = "notification"

	// MessageTypeError 服务端错误消息
	// 方向: Server → Client
	// 说明: 服务端向客户端发送的错误消息
	MessageTypeError = "error"

	// ============================================================
	// 特殊消息类型 (Special Message Types)
	// ============================================================

	// MessageTypeDataConnection 数据连接类型标识
	// 方向: Client → Server (握手阶段)
	// 说明: 用于标识这是一个数据连接而非控制连接
	MessageTypeDataConnection = "data_connection"
)

// ============================================================
// 基础消息结构 - Base Message Structure
// ============================================================

// ControlMessage 控制消息结构（基础消息格式）
// 用于客户端和服务端之间的控制通信
// 所有控制消息都使用此结构进行传输
type ControlMessage struct {
	Type      string                 `json:"type"`      // 消息类型（见上述常量定义）
	SessionID string                 `json:"sessionId"` // 会话ID（用于关联请求和响应）
	Data      map[string]interface{} `json:"data"`      // 消息数据（具体格式见各消息类型定义）
	Timestamp time.Time              `json:"timestamp"` // 消息时间戳
}

// ============================================================
// 客户端 → 服务端消息定义 (Client to Server Messages)
// ============================================================

// AuthRequest 客户端认证请求消息
// 消息类型: MessageTypeAuth
// 传输方向: Client → Server
// 使用场景: 客户端连接建立后的第一条消息，用于身份验证
// 响应消息: AuthResponse
// 注意: data 字段包含完整的 TunnelClient 对象（不包含运行时字段）
//
//	服务端认证成功后直接使用该对象，添加运行时字段后注册到 connectedClients
type AuthRequest struct {
	Client TunnelClient `json:"client"` // 完整的客户端信息（TunnelClient对象）
}

// HeartbeatRequest 客户端心跳请求消息
// 消息类型: MessageTypeHeartbeat
// 传输方向: Client → Server
// 使用场景: 客户端定期发送（默认每30秒），保持连接活跃
// 响应消息: CommonResponse
type HeartbeatRequest struct {
	ClientID  string    `json:"clientId"`  // 客户端ID
	Timestamp time.Time `json:"timestamp"` // 心跳时间戳
}

// RegisterServiceRequest 客户端注册服务请求消息
// 消息类型: MessageTypeRegisterService
// 传输方向: Client → Server
// 使用场景: 客户端请求注册一个新的隧道服务（如SSH、HTTP等）
// 响应消息: RegisterServiceResponse
type RegisterServiceRequest struct {
	Service TunnelService `json:"service"` // 要注册的服务配置（完整的TunnelService对象）
}

// UnregisterServiceRequest 客户端注销服务请求消息
// 消息类型: MessageTypeUnregisterService
// 传输方向: Client → Server
// 使用场景: 客户端请求注销一个已注册的隧道服务
// 响应消息: CommonResponse
type UnregisterServiceRequest struct {
	ServiceID   string `json:"serviceId"`   // 服务ID（TunnelServiceId）
	ServiceName string `json:"serviceName"` // 服务名称（用于日志记录）
}

// DataConnectionHandshake 数据连接握手消息
// 消息类型: MessageTypeDataConnection
// 传输方向: Client → Server
// 使用场景: 客户端建立数据连接时发送的握手消息，用于标识这是数据连接而非控制连接
// 响应消息: 无（直接进入数据转发模式）
type DataConnectionHandshake struct {
	Type         string `json:"type"`         // 固定为 "data_connection"
	ConnectionID string `json:"connectionId"` // 连接ID（由服务端在 ProxyRequest 中分配）
	ClientID     string `json:"clientId"`     // 客户端ID
	ServiceID    string `json:"serviceId"`    // 服务ID（可选，用于验证）
}

// ============================================================
// 服务端 → 客户端消息定义 (Server to Client Messages)
// ============================================================

// CommonResponse 通用响应消息
// 消息类型: MessageTypeResponse
// 传输方向: Server → Client
// 使用场景: 服务端对客户端请求的通用响应（如心跳、注销服务等）
type CommonResponse struct {
	Success bool   `json:"success"` // 操作是否成功
	Message string `json:"message"` // 响应消息（成功或失败的描述）
	Code    string `json:"code"`    // 错误码（可选，失败时提供）
}

// AuthResponse 认证响应消息
// 消息类型: MessageTypeResponse
// 传输方向: Server → Client
// 使用场景: 服务端对客户端认证请求的响应
type AuthResponse struct {
	Success   bool   `json:"success"`   // 认证是否成功
	Message   string `json:"message"`   // 响应消息
	ClientID  string `json:"clientId"`  // 客户端ID（认证成功时返回）
	SessionID string `json:"sessionId"` // 会话ID（认证成功时返回，用于后续通信）
}

// RegisterServiceResponse 注册服务响应消息
// 消息类型: MessageTypeResponse
// 传输方向: Server → Client
// 使用场景: 服务端对客户端注册服务请求的响应
type RegisterServiceResponse struct {
	Success    bool   `json:"success"`    // 注册是否成功
	Message    string `json:"message"`    // 响应消息
	ServiceID  string `json:"serviceId"`  // 服务ID（注册成功时返回）
	RemotePort *int   `json:"remotePort"` // 分配的远程端口（如果服务端分配了端口）
}

// ProxyRequestMessage 代理请求消息
// 消息类型: MessageTypeProxyRequest
// 传输方向: Server → Client
// 使用场景: 当有外网请求到达服务端时，服务端通知客户端建立数据连接进行转发
// 响应消息: 客户端建立数据连接（DataConnectionHandshake）
type ProxyRequestMessage struct {
	ConnectionID  string `json:"connectionId"`  // 连接唯一标识（客户端需在数据连接握手时提供）
	ServiceID     string `json:"serviceId"`     // 服务ID
	ServiceName   string `json:"serviceName"`   // 服务名称
	RemoteAddress string `json:"remoteAddress"` // 外网请求来源地址（用于日志和访问控制）
	RemotePort    int    `json:"remotePort"`    // 外网请求来源端口
}

// NotificationMessage 通知消息
// 消息类型: MessageTypeNotification
// 传输方向: Server → Client
// 使用场景: 服务端向客户端发送的通知消息（非关键操作，如配置更新提醒等）
// 响应消息: 无
type NotificationMessage struct {
	Level   string `json:"level"`   // 通知级别: info/warning/error
	Title   string `json:"title"`   // 通知标题
	Message string `json:"message"` // 通知内容
}

// ErrorMessage 错误消息
// 消息类型: MessageTypeError
// 传输方向: Server → Client
// 使用场景: 服务端向客户端发送的错误消息（如服务异常、配置错误等）
// 响应消息: 无（客户端记录错误日志）
type ErrorMessage struct {
	Code      string    `json:"code"`      // 错误码（用于程序化处理）
	Message   string    `json:"message"`   // 错误消息（人类可读）
	Details   string    `json:"details"`   // 错误详情（可选，用于调试）
	Timestamp time.Time `json:"timestamp"` // 错误发生时间
}

// ============================================================
// 消息构建辅助函数 (Message Builder Helpers)
// ============================================================

// NewControlMessage 创建控制消息
// 这是创建所有控制消息的基础函数
func NewControlMessage(msgType string, sessionID string, data map[string]interface{}) *ControlMessage {
	return &ControlMessage{
		Type:      msgType,
		SessionID: sessionID,
		Data:      data,
		Timestamp: time.Now(),
	}
}

// ============================================================
// 客户端消息构建函数 (Client Message Builders)
// ============================================================

// NewAuthRequestMessage 创建认证请求消息
// 方向: Client → Server
// 参数: client - 完整的 TunnelClient 对象（运行时字段会被忽略）
func NewAuthRequestMessage(sessionID string, client *TunnelClient) *ControlMessage {
	return NewControlMessage(MessageTypeAuth, sessionID, map[string]interface{}{
		"client": client,
	})
}

// NewHeartbeatMessage 创建心跳消息
// 方向: Client → Server
func NewHeartbeatMessage(sessionID string, clientID string) *ControlMessage {
	return NewControlMessage(MessageTypeHeartbeat, sessionID, map[string]interface{}{
		"clientId":  clientID,
		"timestamp": time.Now(),
	})
}

// NewRegisterServiceMessage 创建注册服务消息
// 方向: Client → Server
func NewRegisterServiceMessage(sessionID string, service *TunnelService) *ControlMessage {
	return NewControlMessage(MessageTypeRegisterService, sessionID, map[string]interface{}{
		"service": service,
	})
}

// NewUnregisterServiceMessage 创建注销服务消息
// 方向: Client → Server
func NewUnregisterServiceMessage(sessionID string, serviceID string, serviceName string) *ControlMessage {
	return NewControlMessage(MessageTypeUnregisterService, sessionID, map[string]interface{}{
		"serviceId":   serviceID,
		"serviceName": serviceName,
	})
}

// ============================================================
// 服务端消息构建函数 (Server Message Builders)
// ============================================================

// NewResponseMessage 创建响应消息
// 方向: Server → Client
func NewResponseMessage(sessionID string, success bool, message string) *ControlMessage {
	return NewControlMessage(MessageTypeResponse, sessionID, map[string]interface{}{
		"success": success,
		"message": message,
	})
}

// NewAuthResponseMessage 创建认证响应消息
// 方向: Server → Client
func NewAuthResponseMessage(sessionID string, success bool, message string, clientID string) *ControlMessage {
	data := map[string]interface{}{
		"success": success,
		"message": message,
	}
	if success {
		data["clientId"] = clientID
		data["sessionId"] = sessionID
	}
	return NewControlMessage(MessageTypeResponse, sessionID, data)
}

// NewRegisterServiceResponseMessage 创建注册服务响应消息
// 方向: Server → Client
func NewRegisterServiceResponseMessage(sessionID string, success bool, message string, serviceID string, remotePort *int) *ControlMessage {
	data := map[string]interface{}{
		"success": success,
		"message": message,
	}
	if success {
		data["serviceId"] = serviceID
		if remotePort != nil {
			data["remotePort"] = *remotePort
		}
	}
	return NewControlMessage(MessageTypeResponse, sessionID, data)
}

// NewProxyRequestMessage 创建代理请求消息
// 方向: Server → Client
func NewProxyRequestMessage(sessionID string, connectionID string, serviceID string, serviceName string, remoteAddr string, remotePort int) *ControlMessage {
	return NewControlMessage(MessageTypeProxyRequest, sessionID, map[string]interface{}{
		"connectionId":  connectionID,
		"serviceId":     serviceID,
		"serviceName":   serviceName,
		"remoteAddress": remoteAddr,
		"remotePort":    remotePort,
	})
}

// NewNotificationMessage 创建通知消息
// 方向: Server → Client
func NewNotificationMessage(sessionID string, level string, title string, message string) *ControlMessage {
	return NewControlMessage(MessageTypeNotification, sessionID, map[string]interface{}{
		"level":   level,
		"title":   title,
		"message": message,
	})
}

// NewErrorMessage 创建错误消息
// 方向: Server → Client
func NewErrorMessage(sessionID string, code string, message string, details string) *ControlMessage {
	return NewControlMessage(MessageTypeError, sessionID, map[string]interface{}{
		"code":      code,
		"message":   message,
		"details":   details,
		"timestamp": time.Now(),
	})
}

// ============================================================
// 消息流程说明 (Message Flow Documentation)
// ============================================================

/*
消息交互流程示例：

1. 客户端连接和认证流程：
   Client → Server: AuthRequest (MessageTypeAuth)
   Server → Client: AuthResponse (MessageTypeResponse)

2. 心跳保活流程：
   Client → Server: HeartbeatRequest (MessageTypeHeartbeat) [定期发送]
   Server → Client: CommonResponse (MessageTypeResponse)

3. 服务注册流程：
   Client → Server: RegisterServiceRequest (MessageTypeRegisterService)
   Server → Client: RegisterServiceResponse (MessageTypeResponse)

4. 数据转发流程：
   外网请求 → Server
   Server → Client: ProxyRequestMessage (MessageTypeProxyRequest)
   Client → Server: 建立新的TCP连接
   Client → Server: DataConnectionHandshake (MessageTypeDataConnection)
   Server ↔ Client: 数据双向转发

5. 服务注销流程：
   Client → Server: UnregisterServiceRequest (MessageTypeUnregisterService)
   Server → Client: CommonResponse (MessageTypeResponse)

6. 错误处理流程：
   Server → Client: ErrorMessage (MessageTypeError) [任何时候发生错误]
   Client: 记录错误日志，根据错误类型决定是否重连

7. 通知流程：
   Server → Client: NotificationMessage (MessageTypeNotification) [非关键通知]
   Client: 显示通知给用户

注意事项：
- 所有控制消息都通过控制连接传输
- 数据连接仅用于实际的流量转发
- SessionID 用于关联请求和响应
- 客户端必须先认证成功才能发送其他消息
- 心跳消息用于保持连接活跃，防止超时断开
*/
