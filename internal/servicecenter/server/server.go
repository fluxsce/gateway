package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"gateway/internal/servicecenter/dao"
	"gateway/internal/servicecenter/server/handler"
	"gateway/internal/servicecenter/server/interceptor"
	pb "gateway/internal/servicecenter/server/proto"
	"gateway/internal/servicecenter/types"
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/pkg/utils/cert"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

// gRPC 服务器架构说明
//
// 请求处理流程（拦截器链）：
//   客户端请求 -> Unary/Stream 拦截器链 -> Handler -> 响应
//
// 拦截器执行顺序（按注册顺序，实际执行是逆序）：
//   0. Panic 恢复拦截器（interceptor.RecoveryInterceptor）
//      - 捕获所有 panic，记录详细堆栈信息
//      - 返回内部错误给客户端，避免服务崩溃
//      - 最外层拦截器，最先执行
//
//   1. IP 访问控制拦截器（interceptor.IPAccessInterceptor）
//      - 检查客户端 IP 是否在白名单/黑名单中
//      - 拒绝不合规的连接
//
//   2. 认证拦截器（interceptor.AuthInterceptor）
//      - 从 metadata 中提取认证信息
//      - 验证认证令牌的有效性
//      - 将认证信息添加到 context 中
//
//   3. 日志拦截器（interceptor.LoggingInterceptor）
//      - 记录请求开始时间
//      - 记录请求方法、客户端 IP、认证信息
//      - 记录请求处理时间和结果
//
// Handler 处理：
//   - RegistryHandler: 服务注册、发现、心跳
//   - ConfigHandler: 配置管理、订阅、历史
//
// 响应返回：
//   - 逆序执行拦截器的后置处理
//   - 返回给客户端

// Server gRPC 服务器
// 负责处理服务注册、发现、配置管理等核心功能
type Server struct {
	grpcServer      *grpc.Server             // gRPC 服务器实例
	db              database.Database        // 数据库连接（避免循环依赖）
	instanceDAO     *dao.InstanceDAO         // 实例 DAO（用于更新状态）
	config          *types.InstanceConfig    // 实例配置（从数据库加载）
	running         atomic.Bool              // 服务器运行状态（true=运行中，false=已停止）
	mu              sync.RWMutex             // 保护 config 的并发访问
	registryHandler *handler.RegistryHandler // 服务注册发现处理器（用于访问订阅管理器）
	configHandler   *handler.ConfigHandler   // 配置中心处理器（用于访问配置监听器）

	// 停止信号
	stopCh chan struct{}

	// 等待组 - 用于优雅关闭和并发控制（参考网关模式）
	wg sync.WaitGroup

	// 监听器（在 Start 时创建并持有，防止端口被其他实例占用）
	listener   net.Listener
	listenerMu sync.Mutex // 保护 listener 的并发访问
}

// NewServer 创建 gRPC 服务器（根据实例配置）
// 参数：
//   - db: 数据库连接（避免与 manager 循环依赖）
//   - config: 实例配置，包含监听地址、端口、Keep-alive 参数等
//
// 返回：
//   - *Server: gRPC 服务器实例
func NewServer(db database.Database, config *types.InstanceConfig) *Server {
	if config == nil {
		panic("实例配置不能为空")
	}
	if db == nil {
		panic("数据库连接不能为空")
	}

	server := &Server{
		db:          db,
		instanceDAO: dao.NewInstanceDAO(db), // 用于更新实例状态
		config:      config,
		stopCh:      make(chan struct{}), // 初始化停止信号
		// grpcServer、registryHandler、configHandler 将在 Start 方法中创建
		// 这样可以应用最新的配置
	}

	return server
}

// buildGRPCOptions 根据实例配置构建 gRPC 服务器选项
// 配置项包括：
//   - 消息大小限制（MaxRecvMsgSize、MaxSendMsgSize）
//   - Keep-alive 参数（Time、Timeout、MinTime、PermitWithoutStream）
//   - 连接管理（MaxConnectionIdle、MaxConnectionAge、MaxConnectionAgeGrace）
//   - 性能调优（MaxConcurrentStreams、ReadBufferSize、WriteBufferSize）
//   - TLS 配置（EnableTLS、证书路径等）
//   - 拦截器链（认证、日志、访问控制）
func (s *Server) buildGRPCOptions() []grpc.ServerOption {
	s.mu.RLock()
	config := s.config
	s.mu.RUnlock()

	opts := []grpc.ServerOption{
		// ========== 消息大小限制 ==========
		// 防止超大消息导致内存溢出
		grpc.MaxRecvMsgSize(config.MaxRecvMsgSize), // 最大接收消息大小（默认16MB）
		grpc.MaxSendMsgSize(config.MaxSendMsgSize), // 最大发送消息大小（默认16MB）

		// ========== Keep-alive 强制策略 ==========
		// 控制客户端 keep-alive 行为，防止恶意或配置不当的客户端
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             time.Duration(config.KeepAliveMinTime) * time.Second, // 客户端最小 keep-alive 间隔（防止过于频繁的 ping）
			PermitWithoutStream: config.PermitWithoutStream == "Y",                    // 允许无活跃流时发送 keep-alive（服务发现场景需要）
		}),

		// ========== Keep-alive 参数 ==========
		// 服务端主动保活和清理连接
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:                  time.Duration(config.KeepAliveTime) * time.Second,         // 服务端主动发送 keep-alive ping 的间隔（默认30秒）
			Timeout:               time.Duration(config.KeepAliveTimeout) * time.Second,      // Keep-alive ping 超时时间（默认10秒），超时则关闭连接
			MaxConnectionIdle:     time.Duration(config.MaxConnectionIdle) * time.Second,     // 最大连接空闲时间（0表示无限制，服务中心建议0）
			MaxConnectionAge:      time.Duration(config.MaxConnectionAge) * time.Second,      // 最大连接存活时间（0表示无限制，服务中心建议0）
			MaxConnectionAgeGrace: time.Duration(config.MaxConnectionAgeGrace) * time.Second, // 连接关闭宽限期（默认20秒，允许正在进行的 RPC 完成）
		}),

		// ========== 拦截器链 ==========
		// 创建拦截器实例（所有拦截器共享同一个 ConfigProvider）
		// 注意：拦截器执行顺序与注册顺序相反（最外层最先执行）
		// 实际执行顺序：Recovery -> IPAccess -> Auth -> Logging -> Handler
		grpc.ChainUnaryInterceptor(
			interceptor.NewRecoveryInterceptor().UnaryServerInterceptor(),    // 0. Panic 恢复（最外层，最先执行）
			interceptor.NewIPAccessInterceptor(s).UnaryServerInterceptor(),   // 1. IP 访问控制
			interceptor.NewAuthInterceptor(s, s.db).UnaryServerInterceptor(), // 2. 认证（支持用户名密码验证）
			interceptor.NewLoggingInterceptor().UnaryServerInterceptor(),     // 3. 日志记录
		),
		grpc.ChainStreamInterceptor(
			interceptor.NewRecoveryInterceptor().StreamServerInterceptor(),    // 0. Panic 恢复（最外层，最先执行）
			interceptor.NewIPAccessInterceptor(s).StreamServerInterceptor(),   // 1. IP 访问控制
			interceptor.NewAuthInterceptor(s, s.db).StreamServerInterceptor(), // 2. 认证（支持用户名密码验证）
			interceptor.NewLoggingInterceptor().StreamServerInterceptor(),     // 3. 日志记录
		),
	}

	// ========== 并发流数限制 ==========
	// 限制单个连接的最大并发流数，防止单个客户端占用过多资源
	if config.MaxConcurrentStreams > 0 {
		opts = append(opts, grpc.MaxConcurrentStreams(uint32(config.MaxConcurrentStreams))) // 默认250个
	}

	// ========== 读写缓冲区大小 ==========
	// 调整缓冲区大小以优化网络 I/O 性能
	if config.ReadBufferSize > 0 {
		opts = append(opts, grpc.ReadBufferSize(config.ReadBufferSize)) // 读缓冲区（默认32KB）
	}
	if config.WriteBufferSize > 0 {
		opts = append(opts, grpc.WriteBufferSize(config.WriteBufferSize)) // 写缓冲区（默认32KB）
	}

	// ========== TLS 配置 ==========
	// 注意：TLS 配置的构建已移到 Start 方法中，以便在构建失败时更新状态并阻止启动
	// 这里只返回基础选项，TLS 选项将在 Start 方法中添加

	return opts
}

// ConfigProvider 接口实现

// GetConfig 实现 interceptor.ConfigProvider 接口
// 供拦截器获取最新的实例配置
func (s *Server) GetConfig() *types.InstanceConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

// TLS 配置

// buildTLSConfig 根据实例配置构建 TLS 配置
// 支持两种证书存储方式：
//   - FILE: 从文件系统加载证书（CertFilePath、KeyFilePath）
//   - DATABASE: 从数据库加载证书（CertContent、KeyContent）
//
// 支持双向 TLS 认证（mTLS）：
//   - EnableMTLS = "Y" 时，要求客户端提供证书进行验证
//
// 统一使用 cert.CertLoader 处理证书加载，支持加密私钥解密
func (s *Server) buildTLSConfig() (*tls.Config, error) {
	s.mu.RLock()
	config := s.config
	s.mu.RUnlock()

	var certConfig *cert.CertConfig

	switch config.CertStorageType {
	case "FILE":
		// 从文件系统加载证书
		if config.CertFilePath == "" || config.KeyFilePath == "" {
			return nil, fmt.Errorf("FILE 存储类型需要提供 certFilePath 和 keyFilePath")
		}

		certConfig = &cert.CertConfig{
			CertFile:     config.CertFilePath,
			KeyFile:      config.KeyFilePath,
			KeyPassword:  config.CertPassword, // 私钥密码（如果私钥加密）
			TLSVersions:  []string{},          // 使用默认：TLS 1.2+
			CipherSuites: []string{},          // 使用默认安全加密套件
		}

		logger.Debug("使用 FILE 模式加载证书",
			"instanceName", config.InstanceName,
			"certFile", config.CertFilePath)

	case "DATABASE":
		// 从数据库加载证书内容（直接使用 CertLoader 的内容加载功能）
		if config.CertContent == "" || config.KeyContent == "" {
			return nil, fmt.Errorf("DATABASE 存储类型需要提供 certContent 和 keyContent")
		}

		certConfig = &cert.CertConfig{
			CertContent:  config.CertContent,
			KeyContent:   config.KeyContent,
			KeyPassword:  config.CertPassword,
			TLSVersions:  []string{},
			CipherSuites: []string{},
		}

		logger.Debug("使用 DATABASE 模式加载证书（直接从内容）",
			"instanceName", config.InstanceName)

	default:
		return nil, fmt.Errorf("不支持的证书存储类型: %s", config.CertStorageType)
	}

	// 使用统一的 CertLoader 加载证书
	// CertLoader 自动支持：
	// 1. 从文件或内容加载
	// 2. 加密私钥解密（PKCS#8、传统 PEM 格式、DES/3DES/AES）
	// 3. 证书验证
	// 4. 统一的错误处理和日志记录
	certLoader := cert.NewCertLoader(certConfig)
	tlsConfig, err := certLoader.CreateTLSConfig()
	if err != nil {
		return nil, fmt.Errorf("创建 TLS 配置失败: %w", err)
	}

	// 配置双向 TLS（mTLS）
	if config.EnableMTLS == "Y" {
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		logger.Info("已启用双向 TLS 认证（mTLS）",
			"instanceName", config.InstanceName,
			"storageType", config.CertStorageType)
	} else {
		tlsConfig.ClientAuth = tls.NoClientCert
	}

	logger.Info("TLS 配置创建成功",
		"instanceName", config.InstanceName,
		"storageType", config.CertStorageType,
		"mTLS", config.EnableMTLS)

	return tlsConfig, nil
}

// 实例状态管理

// updateInstanceStatus 更新实例状态到数据库
func (s *Server) updateInstanceStatus(ctx context.Context, status, message string) error {
	s.mu.RLock()
	config := s.config
	s.mu.RUnlock()

	// 在 Go 中生成时间戳（避免数据库函数不兼容）
	now := time.Now()

	// 构建更新参数
	updates := map[string]interface{}{
		"instanceStatus": status,
		"statusMessage":  message,
		"lastStatusTime": now, // 使用 Go 生成的时间
	}

	// 更新数据库
	if err := s.instanceDAO.UpdateInstanceStatus(ctx, config.TenantID, config.InstanceName, config.Environment, updates); err != nil {
		logger.Error("更新实例状态失败", err,
			"instanceName", config.InstanceName,
			"status", status,
			"message", message)
		return err
	}

	logger.Debug("实例状态已更新",
		"instanceName", config.InstanceName,
		"status", status,
		"message", message)

	return nil
}

// 服务器生命周期管理

// Start 启动服务器（同步方法，内部处理 goroutine）
// 监听指定地址和端口，开始接受 gRPC 请求
//
// 处理流程（参考网关启动模式）：
//  1. 检查是否已经在运行
//  2. 构建完整的 gRPC 服务器选项（包括 TLS 配置）
//  3. 如果 TLS 配置构建失败，更新状态为 ERROR 并返回错误
//  4. 创建 gRPC 服务器、Handler 并注册服务
//  5. 在启动前检查端口是否已被占用
//  6. 创建一个通道用于接收启动错误
//  7. 在 goroutine 中启动 gRPC 服务器（grpcServer.Serve 是阻塞的）
//  8. 等待短暂时间检查是否有立即出现的错误
//  9. 如果启动成功，返回 nil
//
// 注意：
//   - 此方法是同步的，会等待服务器启动完成
//   - grpcServer.Serve() 本身是阻塞的，需要在 goroutine 中调用
//   - 使用 WaitGroup 管理 goroutine 生命周期
//   - grpcServer、registryHandler、configHandler 在启动时创建，可以应用最新配置
func (s *Server) Start(ctx context.Context) error {
	// 检查是否已经在运行
	if s.running.Load() {
		err := fmt.Errorf("服务器已在运行")
		// 更新状态为 ERROR（虽然已在运行，但调用 Start 表示状态异常）
		if updateErr := s.updateInstanceStatus(ctx, types.InstanceStatusError, err.Error()); updateErr != nil {
			logger.Warn("更新错误状态失败", "error", updateErr)
		}
		return err
	}

	s.mu.RLock()
	config := s.config
	s.mu.RUnlock()

	// ========== 构建完整的 gRPC 服务器选项（包括 TLS 配置）==========
	// 在启动时构建，以便在配置失败时更新状态并阻止启动
	opts := s.buildGRPCOptions()

	// 构建 TLS 配置（如果启用）
	if config.EnableTLS == "Y" {
		tlsConfig, err := s.buildTLSConfig()
		if err != nil {
			// TLS 配置构建失败，更新状态为 ERROR 并返回错误
			errMsg := fmt.Sprintf("构建 TLS 配置失败: %v", err)
			if updateErr := s.updateInstanceStatus(ctx, types.InstanceStatusError, errMsg); updateErr != nil {
				logger.Warn("更新错误状态失败", "error", updateErr)
			}
			logger.Error("构建 TLS 配置失败，服务器启动被阻止", err,
				"instanceName", config.InstanceName)
			return fmt.Errorf("构建 TLS 配置失败: %w", err)
		}
		opts = append(opts, grpc.Creds(credentials.NewTLS(tlsConfig)))
		logger.Info("TLS 配置已启用",
			"instanceName", config.InstanceName,
			"storageType", config.CertStorageType,
			"mTLS", config.EnableMTLS)
	}

	// ========== 创建 gRPC 服务器和 Handler（在启动时创建，应用最新配置）==========
	// 创建 gRPC 服务器（使用完整的选项，包括 TLS）
	grpcServer := grpc.NewServer(opts...)

	// Server 内部创建 DAO
	// 注意：服务注册发现不需要 DAO（直接操作缓存），只有配置管理需要 DAO
	configDAO := dao.NewConfigDAO(s.db)
	historyDAO := dao.NewHistoryDAO(s.db)

	// 构建 Handler 依赖
	// RegistryHandler 不需要任何 DAO（直接操作缓存）
	registryHandler := handler.NewRegistryHandler()

	// ConfigHandler 需要 DAO（配置需要持久化到数据库）
	configDeps := &handler.ConfigHandlerDeps{
		ConfigDAO:  configDAO,
		HistoryDAO: historyDAO,
	}

	// 创建配置中心处理器
	configHandler := handler.NewConfigHandler(configDeps)

	// 创建统一双向流处理器（使用共享的 handler 实例）
	// 注意：streamHandler 必须使用与 server 相同的 handler 实例
	// 这样客户端订阅和事件通知才能使用同一个订阅管理器
	streamDeps := &handler.StreamHandlerDeps{
		RegistryHandler: registryHandler,
		ConfigHandler:   configHandler,
	}
	streamHandler := handler.NewStreamHandler(streamDeps)

	// 注册服务注册与发现服务（一元 RPC，保留兼容性）
	pb.RegisterServiceRegistryServer(grpcServer, registryHandler)

	// 注册配置中心服务（一元 RPC，保留兼容性）
	pb.RegisterConfigCenterServer(grpcServer, configHandler)

	// 注册统一双向流服务（推荐使用）
	pb.RegisterServiceCenterStreamServer(grpcServer, streamHandler)

	// 启用 gRPC 反射（用于 grpcurl、grpcui 等调试工具）
	if config.EnableReflection == "Y" {
		reflection.Register(grpcServer)
	}

	// 保存引用（用于访问订阅管理器和配置监听器）
	s.mu.Lock()
	s.grpcServer = grpcServer
	s.registryHandler = registryHandler
	s.configHandler = configHandler
	s.mu.Unlock()

	listenAddr := fmt.Sprintf("%s:%d", config.ListenAddress, config.ListenPort)

	// 在启动前检查并创建监听器（立即持有端口，防止被其他实例占用）
	// 注意：不能先检查后关闭再创建，这会导致端口在检查后立即释放，多个实例可能同时通过检查
	s.listenerMu.Lock()
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		s.listenerMu.Unlock()
		// 端口检查失败，更新状态为 ERROR
		errMsg := fmt.Sprintf("端口 %s 已被占用或无法绑定: %v", listenAddr, err)
		if updateErr := s.updateInstanceStatus(ctx, types.InstanceStatusError, errMsg); updateErr != nil {
			logger.Warn("更新错误状态失败", "error", updateErr)
		}
		return fmt.Errorf("端口 %s 已被占用或无法绑定: %w", listenAddr, err)
	}
	// 立即保存监听器，持有端口（不关闭，直到服务器停止）
	s.listener = listener
	s.listenerMu.Unlock()

	logger.Info("启动 gRPC 服务器", "instanceName", config.InstanceName, "listenAddr", listenAddr)

	// 创建一个通道用于接收启动错误
	errCh := make(chan error, 1)

	// 在 goroutine 中启动 gRPC 服务器（grpcServer.Serve 是阻塞的）
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		// 更新状态为 STARTING
		if err := s.updateInstanceStatus(ctx, types.InstanceStatusStarting, "正在启动 gRPC 服务器"); err != nil {
			logger.Warn("更新启动状态失败（继续启动）", "error", err)
		}

		// 使用已创建的监听器（已经在主线程中创建并持有端口）
		s.listenerMu.Lock()
		listener := s.listener
		s.listenerMu.Unlock()

		if listener == nil {
			// 监听器不存在（不应该发生）
			statusMsg := "监听器不存在"
			if updateErr := s.updateInstanceStatus(ctx, types.InstanceStatusError, statusMsg); updateErr != nil {
				logger.Warn("更新错误状态失败", "error", updateErr)
			}
			select {
			case errCh <- fmt.Errorf("监听器不存在"):
			default:
			}
			return
		}

		// 标记为运行状态
		s.running.Store(true)

		// 更新状态为 RUNNING
		if err := s.updateInstanceStatus(ctx, types.InstanceStatusRunning, fmt.Sprintf("gRPC 服务器运行中，监听地址: %s", listenAddr)); err != nil {
			logger.Warn("更新运行状态失败（服务器已启动）", "error", err)
		}

		logger.Info("gRPC 服务器正在监听",
			"listenAddr", listenAddr,
			"instanceName", config.InstanceName)

		// 阻塞式启动服务器，直到服务器停止或发生错误
		err = s.grpcServer.Serve(listener)

		// 服务器停止后，标记为非运行状态
		s.running.Store(false)

		// 根据停止原因更新状态
		if err != nil && err != context.Canceled {
			// 异常停止
			statusMsg := fmt.Sprintf("服务器异常停止: %v", err)
			if updateErr := s.updateInstanceStatus(ctx, types.InstanceStatusError, statusMsg); updateErr != nil {
				logger.Warn("更新错误状态失败", "error", updateErr)
			}
		} else {
			// 正常停止
			if updateErr := s.updateInstanceStatus(ctx, types.InstanceStatusStopped, "服务器已停止"); updateErr != nil {
				logger.Warn("更新停止状态失败", "error", updateErr)
			}
		}
	}()

	// 等待短暂时间检查是否有立即出现的错误（参考网关模式）
	select {
	case err := <-errCh:
		// 启动失败
		return fmt.Errorf("启动 gRPC 服务器失败: %w", err)
	case <-time.After(100 * time.Millisecond):
		// 没有立即出现错误，检查服务器是否正在运行
		if s.running.Load() {
			logger.Info("gRPC 服务器启动成功", "instanceName", config.InstanceName)
			return nil
		}
		// 如果还没运行，再等待一下
		select {
		case err := <-errCh:
			return fmt.Errorf("启动 gRPC 服务器失败: %w", err)
		case <-time.After(400 * time.Millisecond):
			if s.running.Load() {
				logger.Info("gRPC 服务器启动成功", "instanceName", config.InstanceName)
				return nil
			}
			return fmt.Errorf("gRPC 服务器启动超时")
		}
	}
}

// Stop 停止服务器（同步方法，参考网关停止模式）
// 优雅停止：等待正在进行的 RPC 完成，不接受新的请求
// 停止时间受 MaxConnectionAgeGrace 配置影响
//
// 处理流程（参考网关停止模式）：
//  1. 检查是否正在运行
//  2. 发送停止信号（关闭 stopCh）
//  3. 更新状态为 STOPPING
//  4. 调用 GracefulStop 优雅关闭 gRPC 服务器
//  5. 等待所有 goroutine 结束（wg.Wait()）
//  6. 更新状态为 STOPPED
//
// 注意：
//   - 此方法是同步的，会等待服务器停止完成
//   - 使用 WaitGroup 确保所有后台任务完成
func (s *Server) Stop(ctx context.Context) {
	// 检查是否正在运行
	if !s.running.Load() {
		logger.Warn("服务器未在运行，无需停止")
		return
	}

	s.mu.RLock()
	config := s.config
	s.mu.RUnlock()

	logger.Info("正在停止 gRPC 服务器", "instanceName", config.InstanceName)

	// 发送停止信号（参考网关模式）
	s.mu.Lock()
	if s.stopCh != nil {
		close(s.stopCh)
		s.stopCh = make(chan struct{}) // 重新创建，为下次启动做准备
	}
	s.mu.Unlock()

	// 更新状态为 STOPPING
	if err := s.updateInstanceStatus(ctx, types.InstanceStatusStopping, "正在停止 gRPC 服务器"); err != nil {
		logger.Warn("更新停止中状态失败", "error", err)
	}

	// GracefulStop 会：
	// 1. 关闭 listener，不再接受新连接
	// 2. 等待现有 RPC 完成（最多等待 MaxConnectionAgeGrace 时间）
	// 3. 关闭所有连接
	s.mu.RLock()
	grpcServer := s.grpcServer
	s.mu.RUnlock()

	if grpcServer != nil {
		grpcServer.GracefulStop()
	}

	// 等待所有 goroutine 结束（参考网关模式）
	// 这确保了所有后台任务（包括请求处理）都已完成
	// 防止主进程退出时留下zombie goroutine
	s.wg.Wait()

	// 关闭并释放监听器（释放端口）
	s.listenerMu.Lock()
	if s.listener != nil {
		s.listener.Close()
		s.listener = nil
	}
	s.listenerMu.Unlock()

	// 标记为非运行状态（Start() 中的 Serve() 返回后也会设置，这里是双保险）
	s.running.Store(false)

	// 更新状态为 STOPPED
	if err := s.updateInstanceStatus(ctx, types.InstanceStatusStopped, "gRPC 服务器已停止"); err != nil {
		logger.Warn("更新已停止状态失败", "error", err)
	}

	logger.Info("gRPC 服务器已停止", "instanceName", config.InstanceName)
}

// Reload 重新加载服务器配置
// 由于 gRPC 服务器在创建时绑定了大部分配置（如 TLS、拦截器等），
// 完整的 reload 需要重启服务器。但某些配置可以动态更新。
//
// 策略：
//  1. 如果监听地址/端口变化，返回错误（需要重启）
//  2. 如果 TLS 配置变化，返回错误（需要重启）
//  3. 如果认证/IP 白名单配置变化，动态更新（无需重启）
//
// 参数：
//   - ctx: 上下文
//   - newConfig: 新的实例配置
//
// 返回：
//   - error: 如果需要重启才能生效，返回错误
func (s *Server) Reload(ctx context.Context, newConfig *types.InstanceConfig) error {
	if newConfig == nil {
		return fmt.Errorf("新配置不能为空")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	oldConfig := s.config

	// 检查监听地址是否变化（需要重启）
	if oldConfig.ListenAddress != newConfig.ListenAddress ||
		oldConfig.ListenPort != newConfig.ListenPort {
		return fmt.Errorf("监听地址变更需要重启服务器")
	}

	// 检查 TLS 配置是否变化（需要重启）
	if oldConfig.EnableTLS != newConfig.EnableTLS ||
		oldConfig.CertStorageType != newConfig.CertStorageType ||
		oldConfig.CertFilePath != newConfig.CertFilePath ||
		oldConfig.KeyFilePath != newConfig.KeyFilePath ||
		oldConfig.CertContent != newConfig.CertContent ||
		oldConfig.KeyContent != newConfig.KeyContent {
		return fmt.Errorf("TLS 配置变更需要重启服务器")
	}

	// 可以动态更新的配置
	s.config = newConfig

	logger.Info("服务器配置已重新加载",
		"instanceName", newConfig.InstanceName,
		"enableAuth", newConfig.EnableAuth,
		"ipWhitelist", newConfig.IpWhitelist,
		"ipBlacklist", newConfig.IpBlacklist)

	return nil
}

// Port 获取服务器监听端口
// 返回：服务器监听的端口号
func (s *Server) Port() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config.ListenPort
}

// IsRunning 检查服务器是否正在运行
func (s *Server) IsRunning() bool {
	return s.running.Load()
}

// GetRegistryHandler 获取服务注册发现处理器（供外部访问订阅管理器使用）
func (s *Server) GetRegistryHandler() *handler.RegistryHandler {
	return s.registryHandler
}

// GetConfigHandler 获取配置中心处理器（供外部访问配置监听器使用）
func (s *Server) GetConfigHandler() *handler.ConfigHandler {
	return s.configHandler
}
