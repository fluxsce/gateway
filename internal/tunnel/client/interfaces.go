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
	"net"
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

	// GetStatus 获取客户端状态
	//
	// 返回:
	//   - *ClientStatus: 客户端当前状态信息
	//
	// 功能:
	//   - 返回客户端的连接状态
	//   - 提供服务注册数量和代理连接统计
	//   - 包含流量统计和错误信息
	//   - 显示最后心跳时间和重连次数
	//
	// 返回的状态信息包括:
	//   - 连接状态 (connected/disconnected/reconnecting)
	//   - 服务器地址和端口
	//   - 连接持续时间和重连次数
	//   - 已注册服务数量和活跃代理数量
	//   - 总流量统计和最近错误列表
	GetStatus() *ClientStatus

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
	//   - ctx: 上下文，用于控制发送超时
	//   - message: 要发送的控制消息
	//
	// 返回:
	//   - error: 发送失败时返回错误
	//
	// 功能:
	//   - 将消息序列化为二进制格式
	//   - 通过控制连接发送到服务器
	//   - 处理消息确认和重传机制
	//   - 更新发送统计信息
	//
	// 支持的消息类型:
	//   - 认证消息 (auth)
	//   - 心跳消息 (heartbeat)
	//   - 服务注册消息 (register_service)
	//   - 服务注销消息 (unregister_service)
	//   - 代理控制消息 (new_proxy, close_proxy)
	//
	// 注意事项:
	//   - 消息发送是线程安全的
	//   - 大消息会自动分片发送
	//   - 发送失败会触发连接检查
	SendMessage(ctx context.Context, message *types.ControlMessage) error

	// SendMessageAndWaitResponse 发送控制消息并等待响应
	//
	// 参数:
	//   - ctx: 上下文，用于控制发送和等待过程
	//   - message: 要发送的控制消息
	//   - timeout: 等待响应的超时时间
	//
	// 返回:
	//   - *types.ControlMessage: 服务器的响应消息
	//   - error: 发送失败或等待超时时返回错误
	//
	// 功能:
	//   - 发送控制消息到服务器
	//   - 阻塞等待服务器响应
	//   - 通过SessionID匹配请求和响应
	//   - 支持超时控制
	//
	// 使用场景:
	//   - 服务注册/注销需要确认结果
	//   - 需要同步等待服务器响应的操作
	//   - 需要获取服务器返回的数据（如分配的端口）
	//
	// 注意事项:
	//   - 此方法会阻塞直到收到响应或超时
	//   - SessionID必须唯一以正确匹配响应
	//   - 超时后会自动清理等待状态
	SendMessageAndWaitResponse(ctx context.Context, message *types.ControlMessage, timeout time.Duration) (*types.ControlMessage, error)

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

	// GetConnectionInfo 获取连接信息
	//
	// 返回:
	//   - *ConnectionInfo: 连接的详细信息
	//
	// 功能:
	//   - 返回连接的本地和远程地址端口
	//   - 提供连接建立时间和活动时间
	//   - 包含发送接收的字节数统计
	//   - 显示连接质量相关指标
	//
	// 信息内容:
	//   - 本地地址端口和远程地址端口
	//   - 连接建立时间和最后活动时间
	//   - 累计发送和接收字节数
	//   - 平均延迟和连接质量评分
	//
	// 注意事项:
	//   - 返回的信息为当前快照
	//   - 统计数据会持续更新
	//   - 连接断开后信息仍然可用
	GetConnectionInfo() *ConnectionInfo
}

// ServiceManager 服务管理器接口
//
// 服务管理器负责客户端本地服务的生命周期管理，包括服务注册、配置验证、
// 状态监控等功能。它是连接本地服务和隧道系统的桥梁。
//
// 服务管理流程：
//  1. ValidateService() - 验证服务配置的有效性
//  2. RegisterService() - 向服务器注册服务
//  3. StartService() - 启动服务的本地监听
//  4. 监控服务状态和处理连接
//  5. StopService() - 停止服务
//  6. UnregisterService() - 从服务器注销服务
//
// 支持的服务特性：
//   - 多协议支持 (TCP/UDP/HTTP/HTTPS)
//   - 自定义域名和子域名
//   - HTTP认证和头部重写
//   - 带宽限制和连接数限制
//   - 加密和压缩传输
type ServiceManager interface {
	// RegisterService 注册服务
	//
	// 参数:
	//   - ctx: 上下文
	//   - service: 服务配置对象
	//
	// 返回:
	//   - error: 注册失败时返回错误
	//
	// 功能:
	//   - 验证服务配置的完整性和合法性
	//   - 向服务器发送服务注册请求
	//   - 等待服务器分配远程端口或域名
	//   - 将服务添加到本地服务列表
	//   - 初始化服务状态跟踪
	//
	// 注意事项:
	//   - 服务名称必须在客户端范围内唯一
	//   - 注册前会自动验证本地端口可用性
	//   - 注册成功后服务状态为 inactive，需要调用 StartService 启动
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
	//   - 停止服务的本地监听（如果正在运行）
	//   - 向服务器发送服务注销请求
	//   - 关闭所有相关的代理连接
	//   - 从本地服务列表中移除服务
	//   - 清理服务相关资源
	//
	// 注意事项:
	//   - 注销不存在的服务会返回错误
	//   - 注销过程中会优雅关闭活跃连接
	//   - 注销成功后释放分配的远程资源
	UnregisterService(ctx context.Context, serviceID string) error

	// GetService 获取服务
	//
	// 参数:
	//   - ctx: 上下文
	//   - serviceID: 服务ID
	//
	// 返回:
	//   - *types.TunnelService: 服务配置对象
	//   - error: 获取失败时返回错误
	//
	// 功能:
	//   - 根据服务ID查找本地服务配置
	//   - 返回服务的完整配置信息
	//   - 包含服务的当前状态和统计信息
	//
	// 注意事项:
	//   - 查找不存在的服务会返回错误
	//   - 返回的配置对象包含最新的状态信息
	//   - 配置对象为只读，不应直接修改
	GetService(ctx context.Context, serviceID string) (*types.TunnelService, error)

	// GetAllServices 获取所有服务
	//
	// 参数:
	//   - ctx: 上下文
	//
	// 返回:
	//   - []*types.TunnelService: 所有服务的配置列表
	//   - error: 获取失败时返回错误
	//
	// 功能:
	//   - 返回客户端所有已注册的服务
	//   - 包含每个服务的详细配置和状态
	//   - 提供服务的连接统计和性能指标
	//
	// 注意事项:
	//   - 返回的列表为当前时刻的快照
	//   - 列表按服务注册时间排序
	//   - 包含所有状态的服务（活跃、非活跃、错误）
	GetAllServices(ctx context.Context) ([]*types.TunnelService, error)

	// StartService 启动服务
	//
	// 参数:
	//   - ctx: 上下文
	//   - serviceID: 要启动的服务ID
	//
	// 返回:
	//   - error: 启动失败时返回错误
	//
	// 功能:
	//   - 验证本地服务的可用性
	//   - 启动本地代理监听器
	//   - 建立与服务器的代理通道
	//   - 更新服务状态为活跃
	//   - 开始处理代理连接
	//
	// 注意事项:
	//   - 服务必须先注册才能启动
	//   - 启动前会检查本地端口可用性
	//   - 启动失败会回滚状态变更
	StartService(ctx context.Context, serviceID string) error

	// StopService 停止服务
	//
	// 参数:
	//   - ctx: 上下文
	//   - serviceID: 要停止的服务ID
	//
	// 返回:
	//   - error: 停止失败时返回错误
	//
	// 功能:
	//   - 停止接受新的代理连接
	//   - 优雅关闭现有的代理连接
	//   - 关闭本地代理监听器
	//   - 更新服务状态为非活跃
	//   - 清理服务运行时资源
	//
	// 注意事项:
	//   - 停止不存在或未启动的服务会返回错误
	//   - 停止过程会等待活跃连接处理完成
	//   - 服务停止后仍保持注册状态
	StopService(ctx context.Context, serviceID string) error

	// ValidateService 验证服务配置
	//
	// 参数:
	//   - ctx: 上下文
	//   - service: 要验证的服务配置
	//
	// 返回:
	//   - error: 验证失败时返回错误描述
	//
	// 功能:
	//   - 检查服务配置的完整性和合法性
	//   - 验证服务类型和协议参数
	//   - 检查本地地址和端口的可用性
	//   - 验证自定义域名和认证配置
	//   - 确认带宽和连接数限制的合理性
	//
	// 验证内容:
	//   - 必填字段的完整性
	//   - 服务类型和参数的匹配性
	//   - 网络地址和端口的有效性
	//   - HTTP特有配置的正确性
	//   - 安全和性能参数的合理性
	//
	// 注意事项:
	//   - 验证不会修改服务配置
	//   - 验证失败会返回详细的错误信息
	//   - 建议在注册服务前先进行验证
	ValidateService(ctx context.Context, service *types.TunnelService) error
}

// ProxyManager 代理管理器接口
//
// 代理管理器负责管理客户端与服务器之间的数据代理连接。
// 它处理从服务器转发来的用户请求，并将这些请求代理到本地服务。
//
// 代理工作流程：
//  1. 服务器接收到用户连接
//  2. 服务器通知客户端建立代理连接
//  3. StartProxy() 创建到服务器的代理连接
//  4. HandleProxyConnection() 处理数据转发
//  5. StopProxy() 关闭代理连接
//
// 支持的代理特性：
//   - 多协议代理 (TCP/UDP/HTTP/HTTPS)
//   - 双向数据转发和流量统计
//   - 连接池和复用机制
//   - 错误处理和自动重连
type ProxyManager interface {
	// StartProxy 启动代理
	//
	// 参数:
	//   - ctx: 上下文
	//   - service: 服务配置对象
	//   - remotePort: 服务器分配的远程端口
	//
	// 返回:
	//   - error: 启动失败时返回错误
	//
	// 功能:
	//   - 建立到服务器的代理连接
	//   - 连接到本地服务端口
	//   - 启动双向数据转发
	//   - 初始化连接状态跟踪
	//   - 注册连接到活跃代理列表
	//
	// 注意事项:
	//   - 每个代理连接处理一个用户会话
	//   - 连接建立失败会自动清理资源
	//   - 支持并发启动多个代理连接
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
	//   - 停止指定服务的所有代理连接
	//   - 优雅关闭数据转发
	//   - 清理代理连接资源
	//   - 更新连接统计信息
	//   - 从活跃代理列表中移除
	//
	// 注意事项:
	//   - 会等待正在传输的数据完成
	//   - 停止不存在的代理会返回错误
	//   - 停止后无法恢复，需要重新启动
	StopProxy(ctx context.Context, serviceID string) error

	// GetActiveProxies 获取活跃代理
	//
	// 返回:
	//   - []*ProxyInfo: 活跃代理信息列表
	//
	// 功能:
	//   - 返回当前所有活跃的代理连接信息
	//   - 包含代理的状态和统计数据
	//   - 提供连接数和流量统计
	//   - 显示代理的启动时间和持续时间
	//
	// 注意事项:
	//   - 返回的信息为当前时刻快照
	//   - 代理状态可能在返回后发生变化
	//   - 列表按代理启动时间排序
	GetActiveProxies() []*ProxyInfo

	// HandleProxyConnection 处理代理连接
	//
	// 参数:
	//   - ctx: 上下文
	//   - conn: 网络连接对象
	//   - serviceID: 关联的服务ID
	//
	// 返回:
	//   - error: 处理失败时返回错误
	//
	// 功能:
	//   - 处理单个代理连接的数据转发
	//   - 建立与本地服务的连接
	//   - 启动双向数据复制
	//   - 监控连接状态和流量
	//   - 处理连接错误和异常
	//
	// 数据转发机制:
	//   - 异步双向数据复制
	//   - 自动处理连接关闭
	//   - 流量统计和性能监控
	//   - 错误日志和诊断信息
	//
	// 注意事项:
	//   - 此方法会阻塞直到连接关闭
	//   - 连接异常会自动清理资源
	//   - 支持并发处理多个连接
	HandleProxyConnection(ctx context.Context, conn net.Conn, serviceID string) error
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
	//   - 返回最后一次成功心跳的时间
	//   - 用于判断连接的活跃程度
	//   - 支持超时检测和重连判断
	//
	// 注意事项:
	//   - 返回的是成功心跳的时间
	//   - 失败的心跳不会更新此时间
	//   - 零值表示从未成功心跳
	GetLastHeartbeatTime() time.Time

	// GetHeartbeatStats 获取心跳统计
	//
	// 返回:
	//   - *HeartbeatStats: 心跳统计信息
	//
	// 功能:
	//   - 返回详细的心跳统计数据
	//   - 包含发送接收次数和延迟统计
	//   - 提供连接质量评估指标
	//   - 显示心跳成功率和异常情况
	//
	// 统计内容:
	//   - 总发送和接收次数
	//   - 最后发送和接收时间
	//   - 平均、最大、最小延迟
	//   - 失败次数和错误率
	//
	// 注意事项:
	//   - 统计数据实时更新
	//   - 重启心跳管理器会重置统计
	//   - 可用于网络质量分析
	GetHeartbeatStats() *HeartbeatStats
}

// ReconnectManager 重连管理器接口
//
// 重连管理器负责在网络连接中断或异常时自动重新建立连接。
// 它实现了指数退避算法和智能重连策略，确保系统的高可用性。
//
// 重连策略：
//  1. 检测连接中断或异常
//  2. 等待初始重连间隔
//  3. 尝试重新连接
//  4. 失败时按指数退避增加间隔
//  5. 达到最大重试次数后停止
//
// 重连特性：
//   - 指数退避算法
//   - 可配置的最大重试次数
//   - 网络状态检测和智能延迟
//   - 重连成功后的服务恢复
type ReconnectManager interface {
	// Start 启动重连管理
	//
	// 参数:
	//   - ctx: 上下文，用于控制重连管理器生命周期
	//
	// 返回:
	//   - error: 启动失败时返回错误
	//
	// 功能:
	//   - 初始化重连配置和策略
	//   - 启动连接状态监控
	//   - 注册连接中断事件监听器
	//   - 初始化重连统计数据
	//   - 设置重连间隔和最大重试次数
	//
	// 注意事项:
	//   - 需要在客户端启动后调用
	//   - 重复调用会重置重连状态
	//   - 会自动响应连接中断事件
	Start(ctx context.Context) error

	// Stop 停止重连管理
	//
	// 参数:
	//   - ctx: 上下文，用于控制停止过程
	//
	// 返回:
	//   - error: 停止失败时返回错误
	//
	// 功能:
	//   - 停止所有正在进行的重连尝试
	//   - 取消待执行的重连任务
	//   - 清理重连相关资源
	//   - 保存重连统计数据
	//   - 停止连接状态监控
	//
	// 注意事项:
	//   - 停止后重连统计仍然可用
	//   - 重复调用是安全的
	//   - 停止后需要重新Start才能恢复功能
	Stop(ctx context.Context) error

	// TriggerReconnect 触发重连
	//
	// 参数:
	//   - ctx: 上下文，用于控制重连过程
	//   - reason: 重连原因描述
	//
	// 返回:
	//   - error: 触发失败时返回错误
	//
	// 功能:
	//   - 手动触发一次重连操作
	//   - 记录重连原因和时间
	//   - 断开当前连接
	//   - 启动重连尝试流程
	//   - 更新重连统计信息
	//
	// 重连原因示例:
	//   - "heartbeat_timeout" - 心跳超时
	//   - "connection_error" - 连接错误
	//   - "auth_failure" - 认证失败
	//   - "manual_trigger" - 手动触发
	//
	// 注意事项:
	//   - 可以在任意时刻手动调用
	//   - 如果已在重连中，会取消当前重连并重新开始
	//   - 重连成功后会自动恢复服务注册
	TriggerReconnect(ctx context.Context, reason string) error

	// IsReconnecting 检查是否正在重连
	//
	// 返回:
	//   - bool: 重连状态，true表示正在重连
	//
	// 功能:
	//   - 检查当前是否正在进行重连操作
	//   - 用于判断客户端的连接状态
	//   - 支持其他组件的状态判断
	//
	// 注意事项:
	//   - 返回值为当前时刻的重连状态
	//   - 重连状态可能在返回后立即变化
	//   - 建议结合其他状态检查方法使用
	IsReconnecting() bool

	// GetReconnectStats 获取重连统计
	//
	// 返回:
	//   - *ReconnectStats: 重连统计信息
	//
	// 功能:
	//   - 返回详细的重连统计数据
	//   - 包含重连次数和成功率
	//   - 提供重连时间和间隔统计
	//   - 显示当前重连状态和进度
	//
	// 统计内容:
	//   - 总重连尝试次数
	//   - 成功和失败的重连次数
	//   - 最后尝试和成功重连的时间
	//   - 当前重试间隔和计数
	//   - 重连状态和进度信息
	//
	// 注意事项:
	//   - 统计数据实时更新
	//   - 重启重连管理器会重置部分统计
	//   - 可用于网络稳定性分析
	GetReconnectStats() *ReconnectStats
}

// LocalProxy 本地代理接口
//
// 本地代理负责在客户端本地监听特定端口，接收来自服务器的代理连接，
// 并将这些连接转发到本地的实际服务。它是客户端数据转发的核心组件。
//
// 本地代理工作流程：
//  1. Start() 启动本地监听器
//  2. 接收来自服务器的代理连接
//  3. HandleConnection() 处理单个代理连接
//  4. 连接到本地服务并转发数据
//  5. Stop() 停止监听并清理资源
//
// 支持的代理类型：
//   - TCP代理: 原始 TCP 数据流转发
//   - UDP代理: UDP 数据包转发
//   - HTTP代理: HTTP 请求处理和转发
//   - HTTPS代理: HTTPS 请求的 TLS 终结和转发
type LocalProxy interface {
	// Start 启动本地代理
	//
	// 参数:
	//   - ctx: 上下文，用于控制代理生命周期
	//   - service: 服务配置，包含本地地址和端口
	//
	// 返回:
	//   - error: 启动失败时返回错误
	//
	// 功能:
	//   - 根据服务类型启动相应的本地监听器
	//   - 绑定到指定的本地地址和端口
	//   - 初始化连接处理机制
	//   - 设置网络参数（超时、缓冲区等）
	//   - 启动统计数据收集
	//
	// 注意事项:
	//   - 启动前会检查本地端口可用性
	//   - 不同协议类型会使用不同的监听器
	//   - 启动失败会自动清理已分配的资源
	Start(ctx context.Context, service *types.TunnelService) error

	// Stop 停止本地代理
	//
	// 参数:
	//   - ctx: 上下文，用于控制停止过程的超时
	//
	// 返回:
	//   - error: 停止失败时返回错误
	//
	// 功能:
	//   - 停止接受新的代理连接
	//   - 优雅关闭所有活跃连接
	//   - 关闭本地监听器
	//   - 清理连接相关资源
	//   - 保存统计数据
	//
	// 注意事项:
	//   - 会等待正在处理的连接完成
	//   - 重复调用是安全的
	//   - 停止后需要重新Start才能恢复服务
	Stop(ctx context.Context) error

	// HandleConnection 处理连接
	//
	// 参数:
	//   - ctx: 上下文，用于控制连接处理生命周期
	//   - conn: 来自服务器的代理连接
	//
	// 返回:
	//   - error: 处理失败时返回错误
	//
	// 功能:
	//   - 处理单个代理连接的数据转发
	//   - 建立到本地服务的连接
	//   - 启动双向数据复制
	//   - 监控连接状态和性能
	//   - 处理连接异常和错误
	//
	// 数据转发机制:
	//   - 高效的异步 I/O 复制
	//   - 自动缓冲区管理
	//   - 流量控制和限速支持
	//   - 连接超时和保活检测
	//
	// 注意事项:
	//   - 此方法会阻塞直到连接关闭
	//   - 支持并发处理多个连接
	//   - 连接异常会自动清理资源
	HandleConnection(ctx context.Context, conn net.Conn) error

	// GetProxyType 获取代理类型
	//
	// 返回:
	//   - string: 代理类型名称
	//
	// 功能:
	//   - 返回当前代理实例的类型
	//   - 用于区分不同的代理处理逻辑
	//   - 支持统计和监控系统的分类管理
	//
	// 可能的返回值:
	//   - "tcp" - TCP 代理
	//   - "udp" - UDP 代理
	//   - "http" - HTTP 代理
	//   - "https" - HTTPS 代理
	//
	// 注意事项:
	//   - 返回值与服务配置中的类型一致
	//   - 用于运行时类型检查和验证
	GetProxyType() string

	// GetStats 获取代理统计
	//
	// 返回:
	//   - *ProxyStats: 代理统计信息
	//
	// 功能:
	//   - 返回详细的代理统计数据
	//   - 包含连接数和流量统计
	//   - 提供性能指标和错误统计
	//   - 显示代理的运行状态和健康情况
	//
	// 统计内容:
	//   - 活跃连接数和总连接数
	//   - 发送和接收的字节数
	//   - 平均延迟和连接质量
	//   - 错误计数和类型分布
	//   - 启动时间和最后活动时间
	//
	// 注意事项:
	//   - 统计数据实时更新
	//   - 重启代理会重置部分统计
	//   - 可用于性能监控和问题诊断
	GetStats() *ProxyStats
}

// 数据结构定义

// ClientStatus 客户端状态
type ClientStatus struct {
	Status             string    `json:"status"`
	ServerAddress      string    `json:"serverAddress"`
	ServerPort         int       `json:"serverPort"`
	Connected          bool      `json:"connected"`
	LastConnectTime    time.Time `json:"lastConnectTime"`
	ConnectionDuration int64     `json:"connectionDuration"`
	ReconnectCount     int       `json:"reconnectCount"`
	RegisteredServices int       `json:"registeredServices"`
	ActiveProxies      int       `json:"activeProxies"`
	TotalTraffic       int64     `json:"totalTraffic"`
	LastHeartbeat      time.Time `json:"lastHeartbeat"`
	Errors             []string  `json:"errors"`
}

// ConnectionInfo 连接信息
type ConnectionInfo struct {
	LocalAddress  string    `json:"localAddress"`
	LocalPort     int       `json:"localPort"`
	RemoteAddress string    `json:"remoteAddress"`
	RemotePort    int       `json:"remotePort"`
	ConnectedAt   time.Time `json:"connectedAt"`
	LastActivity  time.Time `json:"lastActivity"`
	BytesSent     int64     `json:"bytesSent"`
	BytesReceived int64     `json:"bytesReceived"`
}

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

// HeartbeatStats 心跳统计
type HeartbeatStats struct {
	TotalSent        int       `json:"totalSent"`
	TotalReceived    int       `json:"totalReceived"`
	LastSentTime     time.Time `json:"lastSentTime"`
	LastReceivedTime time.Time `json:"lastReceivedTime"`
	AverageLatency   float64   `json:"averageLatency"`
	MaxLatency       float64   `json:"maxLatency"`
	MinLatency       float64   `json:"minLatency"`
	FailedCount      int       `json:"failedCount"`
}

// ReconnectStats 重连统计
type ReconnectStats struct {
	TotalAttempts      int       `json:"totalAttempts"`
	SuccessfulAttempts int       `json:"successfulAttempts"`
	FailedAttempts     int       `json:"failedAttempts"`
	LastAttemptTime    time.Time `json:"lastAttemptTime"`
	LastSuccessTime    time.Time `json:"lastSuccessTime"`
	MaxRetryInterval   int       `json:"maxRetryInterval"`
	CurrentRetryCount  int       `json:"currentRetryCount"`
	IsReconnecting     bool      `json:"isReconnecting"`
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

// ServiceConfig 服务配置
type ServiceConfig struct {
	Name           string            `json:"name"`
	Type           string            `json:"type"`
	LocalAddress   string            `json:"localAddress"`
	LocalPort      int               `json:"localPort"`
	RemotePort     *int              `json:"remotePort,omitempty"`
	CustomDomains  []string          `json:"customDomains,omitempty"`
	SubDomain      string            `json:"subDomain,omitempty"`
	HttpAuth       *HttpAuthConfig   `json:"httpAuth,omitempty"`
	Headers        map[string]string `json:"headers,omitempty"`
	UseEncryption  bool              `json:"useEncryption"`
	UseCompression bool              `json:"useCompression"`
	BandwidthLimit string            `json:"bandwidthLimit,omitempty"`
	MaxConnections int               `json:"maxConnections,omitempty"`
}

// HttpAuthConfig HTTP认证配置
type HttpAuthConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// 状态常量
const (
	// 客户端状态
	StatusDisconnected = "disconnected"
	StatusConnecting   = "connecting"
	StatusConnected    = "connected"
	StatusReconnecting = "reconnecting"
	StatusError        = "error"
	StatusStopped      = "stopped"

	// 服务状态
	ServiceStatusActive   = "active"
	ServiceStatusInactive = "inactive"
	ServiceStatusStarting = "starting"
	ServiceStatusStopping = "stopping"
	ServiceStatusError    = "error"

	// 代理状态
	ProxyStatusRunning = "running"
	ProxyStatusStopped = "stopped"
	ProxyStatusError   = "error"
)
