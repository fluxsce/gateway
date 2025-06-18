package bootstrap

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"gohub/internal/gateway/config"
	"gohub/internal/gateway/constants"
	"gohub/internal/gateway/core"
	"gohub/internal/gateway/handler/auth"
	"gohub/internal/gateway/handler/cors"
	"gohub/internal/gateway/handler/limiter"
	"gohub/internal/gateway/handler/proxy"
	"gohub/internal/gateway/handler/router"
	"gohub/internal/gateway/handler/security"
	"gohub/internal/gateway/handler/service"
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

	// 初始化引擎
	gateway.engine = core.NewEngine()

	// 初始化HTTP服务器，支持连接时间跟踪
	gateway.server = &http.Server{
		Addr:         cfg.Base.Listen,
		Handler:      gateway,
		ReadTimeout:  cfg.Base.ReadTimeout,
		WriteTimeout: cfg.Base.WriteTimeout,
		IdleTimeout:  cfg.Base.IdleTimeout,
		// 连接上下文回调 - 为每个连接添加建立时间
		ConnContext: f.createConnContext,
	}

	// 初始化和设置所有处理器
	if err := f.initializeAndSetHandlers(gateway, cfg); err != nil {
		return nil, fmt.Errorf("初始化处理器失败: %w", err)
	}

	return gateway, nil
}

// createConnContext 创建连接上下文
// 为每个新连接添加连接建立时间到上下文中
func (f *GatewayFactory) createConnContext(ctx context.Context, conn net.Conn) context.Context {
	// 直接在连接建立时记录时间并添加到上下文
	return context.WithValue(ctx, constants.ContextKeyConnectionStartTime, time.Now())
}

// initializeAndSetHandlers 初始化所有处理器并设置到网关
func (f *GatewayFactory) initializeAndSetHandlers(gateway *Gateway, cfg *config.GatewayConfig) error {
	// 1. 路由处理器 - 必需的处理器，总是创建
	routerFactory := router.NewRouterHandlerFactory()
	routerHandler, err := routerFactory.CreateRouter(cfg.Router)
	if err != nil {
		return fmt.Errorf("创建路由处理器失败: %w", err)
	}

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
				return fmt.Errorf("创建负载均衡器失败: %w", err)
			}
		} else {
			serviceManager = serviceFactory.CreateServiceManager()
		}

		// 创建代理工厂
		proxyFactory := proxy.NewProxyFactory(serviceManager)
		proxyHandler, err = proxyFactory.CreateProxy(cfg.Proxy)
		if err != nil {
			return fmt.Errorf("创建代理处理器失败: %w", err)
		}
	}
	// 如果未启用代理，proxyHandler 保持为 nil

	// 3. 认证处理器 - 只在启用时创建
	var authHandler auth.Authenticator
	if cfg.Auth.Enabled {
		authFactory := auth.NewAuthenticatorFactory()
		authHandler, err = authFactory.CreateAuthenticator(cfg.Auth)
		if err != nil {
			return fmt.Errorf("创建认证处理器失败: %w", err)
		}
	}
	// 如果未启用认证，authHandler 保持为 nil

	// 4. CORS处理器 - 只在启用时创建
	var corsHandler cors.CORSHandler
	if cfg.CORS.Enabled {
		corsFactory := cors.NewCORSHandlerFactory()
		corsHandler, err = corsFactory.CreateCORSHandler(cfg.CORS)
		if err != nil {
			return fmt.Errorf("创建CORS处理器失败: %w", err)
		}
	}
	// 如果未启用CORS，corsHandler 保持为 nil

	// 5. 安全处理器 - 只在启用时创建
	var securityHandler security.SecurityHandler
	if cfg.Security.Enabled {
		securityFactory := security.NewSecurityHandlerFactory()
		securityHandler, err = securityFactory.CreateSecurityHandler(cfg.Security)
		if err != nil {
			return fmt.Errorf("创建安全处理器失败: %w", err)
		}
	}
	// 如果未启用安全检查，securityHandler 保持为 nil

	// 6. 限流处理器 - 只在启用且配置有效时创建
	var limiterHandler limiter.LimiterHandler
	if cfg.RateLimit.Enabled && cfg.RateLimit.Rate > 0 {
		limiterFactory := limiter.NewLimiterFactory()
		limiterHandler, err = limiterFactory.CreateLimiter(&cfg.RateLimit)
		if err != nil {
			return fmt.Errorf("创建限流处理器失败: %w", err)
		}
	}
	// 如果未启用限流，limiterHandler 保持为 nil

	// 设置所有处理器到网关（包括 nil 值）
	gateway.router = routerHandler
	gateway.proxy = proxyHandler
	gateway.auth = authHandler
	gateway.cors = corsHandler
	gateway.security = securityHandler
	gateway.limiter = limiterHandler
	return nil
}
