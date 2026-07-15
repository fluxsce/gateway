package bootstrap

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	"gateway/internal/gateway/config"
	"gateway/internal/gateway/constants"
	"gateway/internal/gateway/core"
	"gateway/internal/gateway/handler/auth"
	"gateway/internal/gateway/handler/cors"
	"gateway/internal/gateway/handler/limiter"
	"gateway/internal/gateway/handler/proxy"
	"gateway/internal/gateway/handler/router"
	"gateway/internal/gateway/handler/security"
	"gateway/internal/gateway/handler/service"
	"gateway/pkg/logger"
	"gateway/pkg/utils/cert"
)

// GatewayFactory 网关工厂
// 负责创建和初始化网关实例的所有处理器
type GatewayFactory struct{}

// NewGatewayFactory 创建网关工厂
func NewGatewayFactory() *GatewayFactory {
	return &GatewayFactory{}
}

// CreateGateway 创建网关实例
func (f *GatewayFactory) CreateGateway(cfg *config.GatewayConfig, configFile string) (*Gateway, error) {
	// 初始化网关结构
	if cfg == nil {
		cfg = &config.DefaultGatewayConfig
	}

	// 创建 Gateway 实例
	gateway := &Gateway{
		gatewayConfig: cfg,
		configFile:    configFile,
		running:       false,
		stopCh:        make(chan struct{}),
	}

	generation, err := f.buildGeneration(gateway, cfg)
	if err != nil {
		return nil, fmt.Errorf("初始化网关运行时代际失败: %w", err)
	}
	gateway.installCompatibilityGeneration(generation)
	gateway.currentGeneration.Store(generation)

	return gateway, nil
}

// buildGeneration 完整构建尚未发布的运行时代际。
func (f *GatewayFactory) buildGeneration(gateway *Gateway, cfg *config.GatewayConfig) (*gatewayGeneration, error) {
	handlers, err := f.buildHandlers(cfg)
	if err != nil {
		return nil, err
	}
	engine := core.NewEngine()
	gateway.setupHandlersFor(engine, cfg, handlers)
	generation := &gatewayGeneration{
		config:    cfg,
		engine:    engine,
		handlers:  handlers,
		serveDone: make(chan struct{}),
	}
	server, err := f.createGenerationServer(gateway, generation, cfg)
	if err != nil {
		generation.closeHandlers()
		return nil, err
	}
	generation.server = server
	return generation, nil
}

// createGenerationServer 为单个代际创建不可变的HTTP Server配置。
func (f *GatewayFactory) createGenerationServer(gateway *Gateway, generation *gatewayGeneration, cfg *config.GatewayConfig) (*http.Server, error) {
	// 初始化HTTP服务器，支持连接时间跟踪
	server := &http.Server{
		// 监听地址：服务器绑定的网络地址（如 ":8080"）
		Addr: cfg.Base.Listen,
		// 请求处理器：连接固定使用创建该Server时的运行时代际
		Handler: &generationHTTPHandler{gateway: gateway, generation: generation},
		// 读取超时：从客户端读取请求头的最大时间，超时则关闭连接
		ReadTimeout: cfg.Base.ReadTimeout,
		// 写入超时：向客户端写入响应的最大时间，超时则关闭连接
		WriteTimeout: cfg.Base.WriteTimeout,
		// 空闲超时：保持连接空闲的最大时间，超时则关闭连接（用于HTTP Keep-Alive）
		// 只有在KeepAliveEnabled=true时，此值才生效
		IdleTimeout: cfg.Base.IdleTimeout,
		// 最大请求头字节数：限制单个请求头的最大字节数，防止恶意请求头过大
		MaxHeaderBytes: cfg.Base.MaxHeaderBytes,
		// DisableKeepAlives: !cfg.Base.KeepAliveEnabled,
		// 连接上下文回调 - 为每个连接添加建立时间
		ConnContext: f.createConnContext,
	}

	// HTTP Keep-Alive配置：根据KeepAliveEnabled参数直接控制
	// 使用SetKeepAlivesEnabled方法明确启用或禁用HTTP Keep-Alive
	// - KeepAliveEnabled=true: 启用HTTP Keep-Alive，使用配置的IdleTimeout值
	// - KeepAliveEnabled=false: 禁用HTTP Keep-Alive，每个请求后立即关闭连接
	server.SetKeepAlivesEnabled(cfg.Base.KeepAliveEnabled)

	// 如果启用HTTPS，配置TLS
	if cfg.Base.EnableHTTPS {
		tlsConfig, err := f.createTLSConfig(cfg)
		if err != nil {
			return nil, fmt.Errorf("创建TLS配置失败: %w", err)
		}
		server.TLSConfig = tlsConfig
		logger.Info("TLS配置已加载", "certFile", cfg.Base.CertFile, "keyFile", cfg.Base.KeyFile)
	}
	return server, nil
}

// CreateGatewayWithPool 创建网关实例并添加到连接池
func (f *GatewayFactory) CreateGatewayWithPool(cfg *config.GatewayConfig, configFile string) (*Gateway, error) {
	// 创建网关实例
	gateway, err := f.CreateGateway(cfg, configFile)
	if err != nil {
		return nil, err
	}

	// 获取全局连接池
	pool := GetGlobalPool()

	// 确定实例ID
	instanceID := cfg.InstanceID
	if instanceID == "" {
		instanceID = cfg.Base.Listen // 使用监听地址作为默认ID
	}

	// 添加到连接池
	if err := pool.Add(instanceID, gateway); err != nil {
		return nil, fmt.Errorf("添加网关实例到连接池失败: %w", err)
	}

	return gateway, nil
}

// createConnContext 创建连接上下文
// 为每个新连接添加连接建立时间到上下文中
func (f *GatewayFactory) createConnContext(ctx context.Context, conn net.Conn) context.Context {
	// 直接在连接建立时记录时间并添加到上下文
	return context.WithValue(ctx, constants.ContextKeyConnectionStartTime, time.Now())
}

// createTLSConfig 创建TLS配置
func (f *GatewayFactory) createTLSConfig(cfg *config.GatewayConfig) (*tls.Config, error) {
	// 创建证书加载器配置
	// TLSVersions和CipherSuites留空，使用默认安全配置
	certConfig := &cert.CertConfig{
		CertFile:     cfg.Base.CertFile,
		KeyFile:      cfg.Base.KeyFile,
		KeyPassword:  cfg.Base.KeyPassword, // 私钥密码（如果私钥加密）
		TLSVersions:  []string{},           // 使用默认：TLS 1.2+
		CipherSuites: []string{},           // 使用默认安全加密套件
	}

	// 创建证书加载器
	certLoader := cert.NewCertLoader(certConfig)

	// 创建TLS配置
	tlsConfig, err := certLoader.CreateTLSConfig()
	if err != nil {
		return nil, fmt.Errorf("创建TLS配置失败: %w", err)
	}

	// 注意：ServerName 是客户端配置，服务器端不需要设置
	// 服务器端TLS配置已完成

	return tlsConfig, nil
}

// buildHandlers 构建一个运行时代际独占的处理器集合。
func (f *GatewayFactory) buildHandlers(cfg *config.GatewayConfig) (gatewayHandlers, error) {
	built := gatewayHandlers{}
	success := false
	defer func() {
		if !success {
			built.close()
		}
	}()

	// 1. 路由处理器 - 必需的处理器，总是创建
	routerFactory := router.NewRouterHandlerFactory()
	routerHandler, err := routerFactory.CreateRouter(cfg.Router)
	if err != nil {
		return gatewayHandlers{}, fmt.Errorf("创建路由处理器失败: %w", err)
	}
	built.router = routerHandler

	// 2. 代理处理器 - 只在启用时创建
	var proxyHandler proxy.ProxyHandler
	if cfg.Proxy.Enabled {
		// 创建负载均衡器
		serviceFactory := service.NewLoadBalancerFactory()
		var serviceManager service.ServiceManager

		// 如果配置了服务，则创建带有服务的负载均衡器
		if len(cfg.Proxy.Service) > 0 {
			serviceManager, err = serviceFactory.CreateManagerWithServices(cfg.Proxy.Service)
			if err != nil {
				return gatewayHandlers{}, fmt.Errorf("创建负载均衡器失败: %w", err)
			}
		} else {
			serviceManager = serviceFactory.CreateServiceManager()
		}

		// 创建代理工厂
		proxyFactory := proxy.NewProxyFactory(serviceManager)
		proxyHandler, err = proxyFactory.CreateProxy(cfg.Proxy)
		if err != nil {
			closeGenerationHandler("service", serviceManager)
			return gatewayHandlers{}, fmt.Errorf("创建代理处理器失败: %w", err)
		}
		built.proxy = proxyHandler
	}
	// 如果未启用代理，proxyHandler 保持为 nil

	// 3. 认证处理器 - 只在启用时创建
	var authHandler auth.Authenticator
	if cfg.Auth.Enabled {
		authFactory := auth.NewAuthenticatorFactory()
		authHandler, err = authFactory.CreateAuthenticator(cfg.Auth)
		if err != nil {
			return gatewayHandlers{}, fmt.Errorf("创建认证处理器失败: %w", err)
		}
		built.auth = authHandler
	}
	// 如果未启用认证，authHandler 保持为 nil

	// 4. CORS处理器 - 只在启用时创建
	var corsHandler cors.CORSHandler
	if cfg.CORS.Enabled {
		corsFactory := cors.NewCORSHandlerFactory()
		corsHandler, err = corsFactory.CreateCORSHandler(cfg.CORS)
		if err != nil {
			return gatewayHandlers{}, fmt.Errorf("创建CORS处理器失败: %w", err)
		}
		built.cors = corsHandler
	}
	// 如果未启用CORS，corsHandler 保持为 nil

	// 5. 安全处理器 - 只在启用时创建
	var securityHandler security.SecurityHandler
	if cfg.Security.Enabled {
		securityFactory := security.NewSecurityHandlerFactory()
		securityHandler, err = securityFactory.CreateSecurityHandler(cfg.Security)
		if err != nil {
			return gatewayHandlers{}, fmt.Errorf("创建安全处理器失败: %w", err)
		}
		built.security = securityHandler
	}
	// 如果未启用安全检查，securityHandler 保持为 nil

	// 6. 限流处理器 - 只在启用且配置有效时创建
	var limiterHandler limiter.LimiterHandler
	if cfg.RateLimit.Enabled && cfg.RateLimit.Rate > 0 {
		limiterFactory := limiter.NewLimiterFactory()
		limiterHandler, err = limiterFactory.CreateLimiter(&cfg.RateLimit)
		if err != nil {
			return gatewayHandlers{}, fmt.Errorf("创建限流处理器失败: %w", err)
		}
		built.limiter = limiterHandler
	}
	// 如果未启用限流，limiterHandler 保持为 nil

	success = true
	return built, nil
}

// ReloadGateway 重新加载网关配置
// 允许在不重启服务的情况下更新网关的配置
// 采用无缝切换策略：先初始化新资源，成功后再关闭旧资源
func (f *GatewayFactory) ReloadGateway(gateway *Gateway, newCfg *config.GatewayConfig) error {
	if gateway == nil {
		return fmt.Errorf("网关实例不能为空")
	}

	if newCfg == nil {
		return fmt.Errorf("新配置不能为空")
	}

	// 保存旧配置，以便校验监听地址和更新连接池索引。
	oldConfig := gateway.gatewayConfig

	// 如果监听地址发生变化，需要重启服务
	if oldConfig.Base.Listen != newCfg.Base.Listen {
		return fmt.Errorf("监听地址变更需要重启服务")
	}
	if oldConfig.Base.EnableHTTPS != newCfg.Base.EnableHTTPS {
		return fmt.Errorf("HTTP/HTTPS协议切换需要重启服务")
	}

	// 完整构建新代际；处理器、Engine、Server和TLS任一构建失败都不会修改当前代际。
	generation, err := f.buildGeneration(gateway, newCfg)
	if err != nil {
		return fmt.Errorf("构建新网关代际失败: %w", err)
	}

	// 实例ID索引先原子换键；激活失败时再回滚，避免调用Remove导致运行中的网关被停止。
	var pool *gatewayPool
	oldPoolKey := oldConfig.InstanceID
	if oldPoolKey == "" {
		oldPoolKey = oldConfig.Base.Listen
	}
	instanceIDChanged := oldConfig.InstanceID != newCfg.InstanceID && newCfg.InstanceID != ""
	if instanceIDChanged {
		pool = GetGlobalPool().(*gatewayPool)
		if err := pool.rekey(oldPoolKey, newCfg.InstanceID, gateway); err != nil {
			generation.closeHandlers()
			return fmt.Errorf("更新连接池中的网关实例失败: %w", err)
		}
	}

	// 新Server先进入等待状态，再原子切换连接入口；旧代际随后在后台排空。
	if err := gateway.activateGeneration(generation); err != nil {
		generation.closeHandlers()
		if instanceIDChanged {
			_ = pool.rekey(newCfg.InstanceID, oldPoolKey, gateway)
		}
		return fmt.Errorf("激活新网关代际失败: %w", err)
	}

	return nil
}
