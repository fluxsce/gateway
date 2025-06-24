package loader

import (
	"context"
	"fmt"

	"gohub/internal/gateway/config"
	"gohub/internal/gateway/handler/auth"
	"gohub/internal/gateway/handler/cors"
	"gohub/internal/gateway/handler/limiter"
	"gohub/internal/gateway/handler/proxy"
	"gohub/internal/gateway/handler/router"
	"gohub/internal/gateway/handler/security"
	"gohub/internal/gateway/loader/dbloader"
	"gohub/pkg/database"
	"gohub/pkg/logger"
)

// DatabaseConfigLoader 数据库配置加载器
type DatabaseConfigLoader struct {
	db                   database.Database
	tenantId             string
	baseLoader           *dbloader.BaseConfigLoader
	routerLoader         *dbloader.RouterConfigLoader
	securityLoader       *dbloader.SecurityConfigLoader
	authCORSLoader       *dbloader.AuthCORSConfigLoader
	limiterServiceLoader *dbloader.LimiterServiceLoader
}

// NewDatabaseConfigLoader 创建数据库配置加载器
func NewDatabaseConfigLoader(db database.Database, tenantId string) *DatabaseConfigLoader {
	return &DatabaseConfigLoader{
		db:                   db,
		tenantId:             tenantId,
		baseLoader:           dbloader.NewBaseConfigLoader(db, tenantId),
		routerLoader:         dbloader.NewRouterConfigLoader(db, tenantId),
		securityLoader:       dbloader.NewSecurityConfigLoader(db, tenantId),
		authCORSLoader:       dbloader.NewAuthCORSConfigLoader(db, tenantId),
		limiterServiceLoader: dbloader.NewLimiterServiceLoader(db, tenantId),
	}
}

// LoadGatewayConfig 从数据库加载网关配置
func (loader *DatabaseConfigLoader) LoadGatewayConfig(instanceId string) (*config.GatewayConfig, error) {
	ctx := context.Background()
	
	// 1. 加载网关实例基础配置
	instance, err := loader.baseLoader.LoadGatewayInstance(ctx, instanceId)
	if err != nil {
		return nil, fmt.Errorf("加载网关实例配置失败: %w", err)
	}
	if instance == nil {
		return nil, fmt.Errorf("网关实例不存在: %s", instanceId)
	}

	// 2. 构建网关配置
	gatewayConfig := &config.GatewayConfig{
		InstanceID: instanceId,
		Base:       loader.baseLoader.BuildBaseConfig(instance),
	}

	// 3. 加载Router配置
	routerConfig, err := loader.routerLoader.LoadRouterConfig(ctx, instanceId)
	if err != nil {
		logger.Warn("加载Router配置失败，使用默认配置", "error", err)
		gatewayConfig.Router = router.DefaultRouterConfig
	} else if routerConfig != nil {
		gatewayConfig.Router = *routerConfig
	} else {
		gatewayConfig.Router = router.DefaultRouterConfig
	}

	// 4. 加载代理配置和服务定义
	proxyConfig, err := loader.limiterServiceLoader.LoadProxyConfig(ctx, instanceId)
	if err != nil {
		logger.Warn("加载代理配置失败，使用默认配置", "error", err)
		gatewayConfig.Proxy = proxy.DefaultProxyConfig
	} else if proxyConfig != nil {
		gatewayConfig.Proxy = *proxyConfig
	} else {
		gatewayConfig.Proxy = proxy.DefaultProxyConfig
	}

	// 5. 加载安全配置
	securityConfig, err := loader.securityLoader.LoadSecurityConfig(ctx, instanceId)
	if err != nil {
		logger.Warn("加载安全配置失败，使用默认配置", "error", err)
		gatewayConfig.Security = security.DefaultSecurityConfig
	} else if securityConfig != nil {
		gatewayConfig.Security = *securityConfig
	} else {
		gatewayConfig.Security = security.DefaultSecurityConfig
	}

	// 6. 加载认证配置
	authConfig, err := loader.authCORSLoader.LoadAuthConfig(ctx, instanceId)
	if err != nil {
		logger.Warn("加载认证配置失败，使用默认配置", "error", err)
		gatewayConfig.Auth = auth.AuthConfig{
			Enabled:  false,
			Strategy: auth.StrategyNoAuth,
		}
	} else if authConfig != nil {
		gatewayConfig.Auth = *authConfig
	} else {
		gatewayConfig.Auth = auth.AuthConfig{
			Enabled:  false,
			Strategy: auth.StrategyNoAuth,
		}
	}

	// 7. 加载CORS配置
	corsConfig, err := loader.authCORSLoader.LoadCORSConfig(ctx, instanceId)
	if err != nil {
		logger.Warn("加载CORS配置失败，使用默认配置", "error", err)
		gatewayConfig.CORS = cors.DefaultCORSConfig
	} else if corsConfig != nil {
		gatewayConfig.CORS = *corsConfig
	} else {
		gatewayConfig.CORS = cors.DefaultCORSConfig
	}

	// 8. 加载限流配置
	rateLimitConfig, err := loader.limiterServiceLoader.LoadRateLimitConfig(ctx, instanceId)
	if err != nil {
		logger.Warn("加载限流配置失败，使用默认配置", "error", err)
		gatewayConfig.RateLimit = limiter.RateLimitConfig{
			Enabled:         false,
			Algorithm:       limiter.AlgorithmTokenBucket,
			Rate:            100,
			Burst:           50,
			ErrorStatusCode: 429,
			ErrorMessage:    "Rate limit exceeded",
		}
	} else {
		gatewayConfig.RateLimit = *rateLimitConfig
	}

	// 9. 为每个路由加载路由级别配置
	for i := range gatewayConfig.Router.Routes {
		route := &gatewayConfig.Router.Routes[i]
		
		// 加载路由级别的安全配置
		routeSecurityConfig, err := loader.securityLoader.LoadRouteSecurityConfig(ctx, route.ID)
		if err != nil {
			logger.Warn("加载路由安全配置失败", 
				"routeId", route.ID, 
				"error", err)
		} else if routeSecurityConfig != nil {
			route.SecurityConfig = routeSecurityConfig
		}

		// 加载路由级别的认证配置
		routeAuthConfig, err := loader.authCORSLoader.LoadRouteAuthConfig(ctx, route.ID)
		if err != nil {
			logger.Warn("加载路由认证配置失败", 
				"routeId", route.ID, 
				"error", err)
		} else if routeAuthConfig != nil {
			route.AuthConfig = routeAuthConfig
		}

		// 加载路由级别的CORS配置
		routeCorsConfig, err := loader.authCORSLoader.LoadRouteCORSConfig(ctx, route.ID)
		if err != nil {
			logger.Warn("加载路由CORS配置失败", 
				"routeId", route.ID, 
				"error", err)
		} else if routeCorsConfig != nil {
			route.CorsConfig = routeCorsConfig
		}

		// 加载路由级别的限流配置
		routeRateLimitConfig, err := loader.limiterServiceLoader.LoadRouteRateLimitConfig(ctx, route.ID)
		if err != nil {
			logger.Warn("加载路由限流配置失败", 
				"routeId", route.ID, 
				"error", err)
		} else if routeRateLimitConfig != nil {
			route.LimiterConfig = routeRateLimitConfig
		}

		// 加载路由级别的过滤器配置
		routeFilters, err := loader.routerLoader.LoadRouteFilters(ctx, route.ID)
		if err != nil {
			logger.Warn("加载路由过滤器配置失败", 
				"routeId", route.ID, 
				"error", err)
		} else if routeFilters != nil {
			route.FilterConfig = routeFilters
		}
	}

	return gatewayConfig, nil
}
