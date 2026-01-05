// Package client 定义隧道客户端核心接口
// 基于FRP架构，实现客户端连接和服务注册
//
// 本包提供了完整的隧道客户端功能，包括：
// - 与服务器建立控制连接
// - 服务注册和管理
// - 本地代理和数据转发
// - 心跳保活和自动重连
// - 连接状态监控和统计
//
// 主要组件：
// - TunnelClient: 客户端主接口，负责整体协调
// - ControlConnection: 控制连接管理
// - ServiceManager: 服务注册和管理
// - ProxyManager: 代理连接管理
// - HeartbeatManager: 心跳保活管理
// - ReconnectManager: 断线重连管理
package client

import (
	"context"
	"time"

	"gateway/internal/tunnel/types"
)

// TunnelClient 隧道客户端接口
//
// 隧道客户端是整个客户端系统的核心接口，负责协调各个子组件的工作。
// 它管理与隧道服务器的连接，处理服务注册，并提供统一的状态管理。
//
// 典型的使用流程：
//  1. 创建客户端实例并配置连接参数
//  2. 调用 Start() 启动客户端并连接到服务器
//  3. 使用 RegisterService() 注册需要暴露的本地服务
//  4. 通过 GetStatus() 监控连接状态和统计信息
//  5. 使用完毕后调用 Stop() 优雅关闭客户端
//
// 示例代码：
//
//	client := NewTunnelClient(config)
//	if err := client.Start(ctx); err != nil {
//	    log.Fatal("Failed to start client:", err)
//	}
//	defer client.Stop(ctx)
//
//	service := &types.TunnelService{
//	    ServiceName: "web",
//	    ServiceType: "http",
//	    LocalAddress: "127.0.0.1",
//	    LocalPort: 8080,
//	}
//	if err := client.RegisterService(ctx, service); err != nil {
//	    log.Fatal("Failed to register service:", err)
//	}
type TunnelClient interface {
	// Start 启动客户端
	//
	// 参数:
	//   - ctx: 上下文，用于控制启动过程和取消操作
	//
	// 返回:
	//   - error: 启动失败时返回错误
	//
	// 功能:
	//   - 初始化客户端各个子组件
	//   - 建立与服务器的控制连接
	//   - 启动心跳保活机制
	//   - 启动重连管理器
	//   - 验证客户端身份认证
	//
	// 注意事项:
	//   - 此方法会阻塞直到连接建立成功或失败
	//   - 如果连接失败，会根据配置进行自动重连
	//   - 调用此方法前需要确保客户端配置正确
	Start(ctx context.Context) error

	// Stop 停止客户端
	//
	// 参数:
	//   - ctx: 上下文，用于控制停止过程的超时
	//
	// 返回:
	//   - error: 停止失败时返回错误
	//
	// 功能:
	//   - 注销所有已注册的服务
	//   - 关闭所有活跃的代理连接
	//   - 停止心跳保活机制
	//   - 断开与服务器的控制连接
	//   - 清理资源和停止后台协程
	//
	// 注意事项:
	//   - 此方法会优雅地关闭所有连接
	//   - 如果在指定时间内无法完成关闭，会强制终止
	//   - 重复调用此方法是安全的
	Stop(ctx context.Context) error

	// IsConnected 检查是否已连接
	//
	// 返回:
	//   - bool: 连接状态，true表示已连接
	//
	// 功能:
	//   - 返回客户端的当前连接状态
	//   - 用于判断客户端是否可以注册服务
	IsConnected() bool

	// GetConnectTime 获取连接时间
	//
	// 返回:
	//   - time.Time: 最后一次连接建立的时间
	//
	// 功能:
	//   - 返回客户端连接建立的时间
	//   - 用于计算连接持续时间
	GetConnectTime() time.Time

	// GetReconnectCount 获取重连次数
	//
	// 返回:
	//   - int: 重连次数
	//
	// 功能:
	//   - 返回客户端的重连次数
	//   - 用于监控连接稳定性
	GetReconnectCount() int

	// GetConfig 获取客户端配置
	//
	// 返回:
	//   - *types.TunnelClient: 客户端配置对象
	//
	// 功能:
	//   - 返回当前客户端的完整配置信息
	//   - 包含服务器连接参数和认证信息
	//   - 提供心跳间隔和重连策略配置
	//
	// 注意事项:
	//   - 返回的配置对象为只读，不应直接修改
	//   - 配置变更需要重启客户端才能生效
	GetConfig() *types.TunnelClient

	// RegisterService 注册服务
	//
	// 参数:
	//   - ctx: 上下文
	//   - service: 要注册的服务配置
	//
	// 返回:
	//   - error: 注册失败时返回错误
	//
	// 功能:
	//   - 向服务器注册本地服务
	//   - 分配远程端口或域名
	//   - 启动本地代理监听
	//   - 建立服务到代理的映射关系
	//
	// 支持的服务类型:
	//   - TCP: 纯TCP代理
	//   - UDP: UDP数据转发
	//   - HTTP: HTTP协议代理，支持自定义域名
	//   - HTTPS: HTTPS协议代理，支持TLS终结
	//   - STCP: 安全TCP，需要密钥访问
	//   - SUDP: 安全UDP，需要密钥访问
	//
	// 注意事项:
	//   - 服务名称在客户端范围内必须唯一
	//   - 注册前需要确保本地服务已经启动
	//   - 注册成功后会自动处理来自服务器的代理连接
	RegisterService(ctx context.Context, service *types.TunnelService) error

	// UnregisterService 注销服务
	//
	// 参数:
	//   - ctx: 上下文
	//   - serviceID: 要注销的服务ID
	//
	// 返回:
	//   - error: 注销失败时返回错误
	//
	// 功能:
	//   - 从服务器注销指定服务
	//   - 关闭相关的代理连接
	//   - 停止本地代理监听
	//   - 释放分配的远程端口或域名
	//
	// 注意事项:
	//   - 注销不存在的服务会返回错误
	//   - 注销过程中的活跃连接会被优雅关闭
	//   - 注销后该服务将无法再接收新的连接
	UnregisterService(ctx context.Context, serviceID string) error

	// GetRegisteredServices 获取已注册的服务
	//
	// 返回:
	//   - []*types.TunnelService: 已注册的服务列表
	//
	// 功能:
	//   - 返回当前客户端所有已注册的服务
	//   - 包含服务的详细配置和状态信息
	//   - 提供服务的连接统计和流量信息
	//
	// 注意事项:
	//   - 返回的列表为当前时刻的快照
	//   - 服务状态可能在返回后发生变化
	//   - 列表按服务注册时间排序
	GetRegisteredServices() []*types.TunnelService
}

// ControlConnection 控制连接接口
//
// 控制连接负责与隧道服务器建立和维护控制通道，用于传输认证、心跳、
// 服务注册等控制消息。这是客户端与服务器通信的核心组件。
//
// 控制连接的生命周期：
//  1. Connect() - 建立TCP连接到服务器控制端口
//  2. 发送认证消息进行身份验证
//  3. 进入消息循环，处理心跳和控制指令
//  4. Disconnect() - 优雅关闭连接
//
// 消息处理机制：
//   - 支持同步和异步消息发送
//   - 自动处理消息序列化和反序列化
//   - 提供消息重传和确认机制
//   - 支持消息优先级和超时控制
type ControlConnection interface {
	// Connect 连接到服务器控制端口
	//
	// 参数:
	//   - ctx: 上下文，用于控制连接过程和超时
	//   - serverAddress: 服务器地址，如 "tunnel.example.com"
	//   - serverPort: 服务器控制端口，通常为 7000
	//
	// 返回:
	//   - error: 连接失败时返回错误
	//
	// 功能:
	//   - 建立到服务器的TCP连接
	//   - 设置连接参数（超时、缓冲区等）
	//   - 启动消息接收协程
	//   - 初始化连接状态跟踪
	//
	// 注意事项:
	//   - 连接建立后需要进行身份认证
	//   - 支持TLS加密连接（如果配置启用）
	//   - 连接失败会触发重连机制
	Connect(ctx context.Context, serverAddress string, serverPort int) error

	// Disconnect 断开连接
	//
	// 参数:
	//   - ctx: 上下文，用于控制断开过程的超时
	//
	// 返回:
	//   - error: 断开失败时返回错误
	//
	// 功能:
	//   - 发送断开连接通知给服务器
	//   - 关闭底层TCP连接
	//   - 停止消息接收协程
	//   - 清理连接相关资源
	//
	// 注意事项:
	//   - 会等待正在处理的消息完成
	//   - 重复调用是安全的
	//   - 断开后需要重新Connect才能使用
	Disconnect(ctx context.Context) error

	// SendMessage 发送控制消息
	//
	// 参数:
	//   - ctx: 上下文，用于控制发送和等待过程
	//   - message: 要发送的控制消息
	//   - options: 可选参数
	//     - options[0] (bool): 是否等待响应，默认 false
	//     - options[1] (time.Duration): 等待响应的超时时间，默认 30秒
	//
	// 返回:
	//   - *types.ControlMessage: 服务器的响应消息（仅当 waitResponse=true 时返回）
	//   - error: 发送失败或等待超时时返回错误
	//
	// 功能:
	//   - 将消息序列化为二进制格式
	//   - 通过控制连接同步发送到服务器
	//   - 可选：阻塞等待服务器响应
	//   - 通过SessionID匹配请求和响应
	//
	// 支持的消息类型:
	//   - 认证消息 (auth)
	//   - 心跳消息 (heartbeat)
	//   - 服务注册消息 (register_service)
	//   - 服务注销消息 (unregister_service)
	//   - 代理控制消息 (new_proxy, close_proxy)
	//
	// 使用示例:
	//
	//	// 不等待响应（心跳等）
	//	_, err := cc.SendMessage(ctx, heartbeatMsg)
	//
	//	// 等待响应（服务注册等）
	//	response, err := cc.SendMessage(ctx, registerMsg, true, 10*time.Second)
	//
	// 注意事项:
	//   - 消息发送是线程安全的（使用 writerMutex 保护）
	//   - 发送失败会触发连接检查
	//   - 等待响应时会阻塞直到收到响应或超时
	//   - SessionID必须唯一以正确匹配响应
	SendMessage(ctx context.Context, message *types.ControlMessage, options ...interface{}) (*types.ControlMessage, error)

	// ReceiveMessage 接收控制消息
	//
	// 参数:
	//   - ctx: 上下文，用于控制接收超时
	//
	// 返回:
	//   - *types.ControlMessage: 接收到的控制消息
	//   - error: 接收失败时返回错误
	//
	// 功能:
	//   - 从控制连接读取二进制数据
	//   - 反序列化为控制消息对象
	//   - 验证消息完整性和格式
	//   - 更新接收统计信息
	//
	// 消息处理:
	//   - 自动处理心跳响应
	//   - 解析服务器指令和通知
	//   - 处理错误和异常消息
	//   - 触发相应的回调函数
	//
	// 注意事项:
	//   - 此方法会阻塞直到收到消息或超时
	//   - 连接断开时会返回特定错误
	//   - 需要在循环中调用以持续接收消息
	ReceiveMessage(ctx context.Context) (*types.ControlMessage, error)

	// IsConnected 检查连接状态
	//
	// 返回:
	//   - bool: 连接状态，true表示已连接
	//
	// 功能:
	//   - 检查底层TCP连接是否活跃
	//   - 验证最近的心跳响应时间
	//   - 确认消息发送接收是否正常
	//
	// 判断标准:
	//   - TCP连接状态正常
	//   - 最近收到服务器响应
	//   - 没有连接错误标记
	//
	// 注意事项:
	//   - 返回值为当前时刻的连接状态
	//   - 状态可能在返回后立即发生变化
	//   - 建议结合心跳机制使用
	IsConnected() bool

	// StartProxy 启动代理
	//
	// 参数:
	//   - ctx: 上下文
	//   - service: 服务配置
	//   - remotePort: 远程端口
	//
	// 返回:
	//   - error: 启动失败时返回错误
	//
	// 功能:
	//   - 将服务添加到活跃代理列表
	//   - 准备接收来自服务器的代理请求
	//
	// 注意事项:
	//   - 服务注册成功后需要调用此方法
	//   - 重复调用是安全的（已存在则忽略）
	StartProxy(ctx context.Context, service *types.TunnelService, remotePort int) error

	// StopProxy 停止代理
	//
	// 参数:
	//   - ctx: 上下文
	//   - serviceID: 服务ID
	//
	// 返回:
	//   - error: 停止失败时返回错误
	//
	// 功能:
	//   - 从活跃代理列表中移除服务
	//   - 停止接收该服务的代理请求
	//
	// 注意事项:
	//   - 服务注销时需要调用此方法
	//   - 停止不存在的代理会返回错误
	StopProxy(ctx context.Context, serviceID string) error
}

// HeartbeatManager 心跳管理器接口
//
// 心跳管理器负责维护客户端与服务器之间的连接活性检测。
// 通过定期发送心跳消息来确保连接的可用性，并检测网络异常。
//
// 心跳机制：
//  1. Start() 启动定时心跳发送
//  2. 按配置间隔发送心跳消息
//  3. 监听服务器心跳响应
//  4. 统计心跳延迟和成功率
//  5. 检测心跳超时并触发重连
//
// 心跳特性：
//   - 可配置的心跳间隔
//   - 自动延迟和丢包检测
//   - 连接质量评估
//   - 异常自动恢复
type HeartbeatManager interface {
	// Start 启动心跳
	//
	// 参数:
	//   - ctx: 上下文，用于控制心跳生命周期
	//   - interval: 心跳间隔时间
	//
	// 返回:
	//   - error: 启动失败时返回错误
	//
	// 功能:
	//   - 启动心跳定时器
	//   - 开始定期发送心跳消息
	//   - 启动心跳响应监听
	//   - 初始化心跳统计
	//   - 设置超时检测机制
	//
	// 注意事项:
	//   - 需要在控制连接建立后调用
	//   - 重复调用会重置心跳间隔
	//   - 心跳失败会触发连接检查
	Start(ctx context.Context, interval time.Duration) error

	// Stop 停止心跳
	//
	// 参数:
	//   - ctx: 上下文，用于控制停止过程
	//
	// 返回:
	//   - error: 停止失败时返回错误
	//
	// 功能:
	//   - 停止心跳定时器
	//   - 取消待发送的心跳
	//   - 停止响应监听
	//   - 保存心跳统计数据
	//   - 清理心跳相关资源
	//
	// 注意事项:
	//   - 停止后心跳统计仍然可用
	//   - 重复调用是安全的
	//   - 停止后需要重新Start才能恢复
	Stop(ctx context.Context) error

	// SendHeartbeat 发送心跳
	//
	// 参数:
	//   - ctx: 上下文，用于控制发送超时
	//
	// 返回:
	//   - error: 发送失败时返回错误
	//
	// 功能:
	//   - 立即发送一个心跳消息
	//   - 记录发送时间戳
	//   - 等待服务器响应
	//   - 计算往返延迟
	//   - 更新心跳统计
	//
	// 注意事项:
	//   - 可以手动调用进行连接检测
	//   - 发送失败可能表示连接异常
	//   - 会更新最后心跳时间
	SendHeartbeat(ctx context.Context) error

	// GetLastHeartbeatTime 获取最后心跳时间
	//
	// 返回:
	//   - time.Time: 最后一次心跳的时间
	//
	// 功能:
	//   - 返回最后一次成功心跳的时间（从 client.config.LastHeartbeat 读取）
	//   - 用于判断连接的活跃程度
	//   - 支持超时检测和重连判断
	//
	// 注意事项:
	//   - 返回的是成功心跳的时间
	//   - 失败的心跳不会更新此时间
	//   - 零值表示从未成功心跳
	GetLastHeartbeatTime() time.Time
}

// 数据结构定义

// ProxyInfo 代理信息
type ProxyInfo struct {
	ServiceID         string    `json:"serviceId"`
	ServiceName       string    `json:"serviceName"`
	ProxyType         string    `json:"proxyType"`
	LocalAddress      string    `json:"localAddress"`
	LocalPort         int       `json:"localPort"`
	RemotePort        int       `json:"remotePort"`
	Status            string    `json:"status"`
	StartTime         time.Time `json:"startTime"`
	ActiveConnections int       `json:"activeConnections"`
	TotalConnections  int64     `json:"totalConnections"`
	TotalTraffic      int64     `json:"totalTraffic"`
}

// ProxyStats 代理统计
type ProxyStats struct {
	ActiveConnections int       `json:"activeConnections"`
	TotalConnections  int64     `json:"totalConnections"`
	BytesSent         int64     `json:"bytesSent"`
	BytesReceived     int64     `json:"bytesReceived"`
	AverageLatency    float64   `json:"averageLatency"`
	ErrorCount        int       `json:"errorCount"`
	StartTime         time.Time `json:"startTime"`
	LastActivityTime  time.Time `json:"lastActivityTime"`
}
