package bootstrap

import (
	"context"
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

	// 保存旧配置和旧处理器，以便在失败时回滚
	oldConfig := gateway.gatewayConfig
	oldRouter := gateway.router
	oldProxy := gateway.proxy
	oldAuth := gateway.auth
	oldCors := gateway.cors
	oldSecurity := gateway.security
	oldLimiter := gateway.limiter

	// 如果监听地址发生变化，需要重启服务
	if oldConfig.Base.Listen != newCfg.Base.Listen {
		return fmt.Errorf("监听地址变更需要重启服务")
	}

	// 临时更新网关配置
	gateway.gatewayConfig = newCfg

	// 尝试初始化新的处理器
	if err := f.initializeAndSetHandlers(gateway, newCfg); err != nil {
		// 初始化失败，回滚配置和处理器
		gateway.gatewayConfig = oldConfig
		gateway.router = oldRouter
		gateway.proxy = oldProxy
		gateway.auth = oldAuth
		gateway.cors = oldCors
		gateway.security = oldSecurity
		gateway.limiter = oldLimiter
		return fmt.Errorf("重新初始化处理器失败: %w", err)
	}

	// 新处理器初始化成功，重新创建engine并设置处理器链
	// 先创建新的engine，设置完成后再一次性替换，确保原子性
	newEngine := core.NewEngine()
	gateway.setupHandlers(newEngine)

	// 处理器链设置完成，现在可以安全替换engine
	gateway.engine = newEngine

	// engine设置完成后，现在可以安全关闭旧处理器
	// 关闭旧的处理器资源，防止资源泄漏
	f.closeOldHandlers(oldRouter, oldProxy, oldAuth, oldCors, oldSecurity, oldLimiter)

	// 更新HTTP服务器配置
	gateway.server.ReadTimeout = newCfg.Base.ReadTimeout
	gateway.server.WriteTimeout = newCfg.Base.WriteTimeout
	gateway.server.IdleTimeout = newCfg.Base.IdleTimeout

	// 更新实例ID
	if oldConfig.InstanceID != newCfg.InstanceID && newCfg.InstanceID != "" {
		// 获取全局连接池
		pool := GetGlobalPool()

		// 从连接池中移除旧ID
		if oldConfig.InstanceID != "" {
			pool.Remove(oldConfig.InstanceID)
		} else if oldConfig.Base.Listen != "" {
			// 如果旧配置没有实例ID，尝试使用监听地址作为ID
			pool.Remove(oldConfig.Base.Listen)
		}

		// 使用新ID添加到连接池
		if err := pool.Add(newCfg.InstanceID, gateway); err != nil {
			return fmt.Errorf("更新连接池中的网关实例失败: %w", err)
		}
	}

	return nil
}

// closeOldHandlers 关闭旧的处理器资源
func (f *GatewayFactory) closeOldHandlers(oldRouter, oldProxy, oldAuth, oldCors, oldSecurity, oldLimiter interface{}) {
	// 优先关闭代理处理器，因为它通常包含健康检查器等后台资源
	if oldProxy != nil {
		if closer, ok := oldProxy.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				logger.Warn("重载配置时关闭旧代理处理器失败", "error", err)
			} else {
				logger.Debug("重载配置时旧代理处理器已关闭")
			}
		}
	}

	// 关闭其他处理器资源
	if oldRouter != nil {
		if closer, ok := oldRouter.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				logger.Warn("重载配置时关闭旧路由处理器失败", "error", err)
			}
		}
	}

	if oldAuth != nil {
		if closer, ok := oldAuth.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				logger.Warn("重载配置时关闭旧认证处理器失败", "error", err)
			}
		}
	}

	if oldCors != nil {
		if closer, ok := oldCors.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				logger.Warn("重载配置时关闭旧CORS处理器失败", "error", err)
			}
		}
	}

	if oldSecurity != nil {
		if closer, ok := oldSecurity.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				logger.Warn("重载配置时关闭旧安全处理器失败", "error", err)
			}
		}
	}

	if oldLimiter != nil {
		if closer, ok := oldLimiter.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				logger.Warn("重载配置时关闭旧限流处理器失败", "error", err)
			}
		}
	}
}
