package bootstrap

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
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
	"gateway/internal/gateway/loader/dbloader"
	"gateway/internal/gateway/logwrite"
	"gateway/pkg/logger"
)

// Gateway 网关核心结构
// 基于处理器链模式的现代化网关实现
type Gateway struct {
	// 配置
	gatewayConfig *config.GatewayConfig

	// 配置文件路径
	configFile string

	// HTTP服务器
	server *http.Server

	// 核心引擎 - 管理处理器链
	engine *core.Engine

	// 处理器实例 - 各功能模块的处理器（使用接口类型以支持多种实现，支持nil表示未启用）
	router   router.RouterHandler     // 路由处理器接口：必需，负责路由匹配和路由级别的处理器链执行
	proxy    proxy.ProxyHandler       // 代理处理器接口：可选，负责请求转发、负载均衡、服务发现等
	auth     auth.Authenticator       // 认证处理器接口：可选，负责身份验证、权限检查、用户上下文设置
	cors     cors.CORSHandler         // CORS处理器接口：可选，负责跨域资源共享的请求处理
	security security.SecurityHandler // 安全处理器接口：可选，负责IP过滤、DDoS防护、恶意请求检测
	limiter  limiter.LimiterHandler   // 限流处理器接口：可选，负责请求频率控制和流量管理
	// 注意：熔断器不在全局级别处理，而是在路由级别或服务级别处理，由路由处理器或代理处理器负责

	// 运行状态
	running bool

	// 互斥锁
	mu sync.RWMutex

	// 停止信号
	stopCh chan struct{}

	// 等待组 - 用于优雅关闭和并发控制
	// WaitGroup的完整作用说明：
	// 1. 服务启动同步：确保HTTP服务器完全启动后再返回Start()方法
	// 2. 并发处理器管理：等待所有后台处理器（如健康检查、指标收集）完成初始化
	// 3. 优雅关闭协调：在Stop()时等待所有正在处理的请求和后台任务完成
	//    - 等待所有正在执行的HTTP请求处理完成
	//    - 等待后台健康检查goroutine停止
	//    - 等待统计指标收集器停止
	//    - 等待配置热重载监听器停止
	//    - 等待日志刷新器完成最后的日志写入
	// 4. 资源清理同步：确保所有资源（连接池、缓存、文件句柄）正确释放
	// 5. 防止数据丢失：确保关键数据（访问日志、统计数据）完整写入存储
	// 6. 防止zombie进程：确保所有子goroutine在主进程退出前正确结束
	// 7. 信号处理配合：与系统信号（SIGTERM、SIGINT）配合实现平滑重启
	wg sync.WaitGroup
}

// setupHandlers 设置处理器链 - 网关处理的核心思想
func (g *Gateway) setupHandlers(engine *core.Engine) {
	// 处理器执行顺序说明：
	// 详细处理流程说明：
	// 1. 请求接收：HTTP服务器接收客户端请求
	// 2. 上下文构建：为请求创建上下文，包含请求信息和处理状态
	//    - 生成唯一请求ID
	//    - 记录请求开始时间
	//    - 提取客户端信息（IP、User-Agent等）
	//    - 初始化请求上下文对象
	// 3. 全局安全管理控制
	//    - IP白名单/黑名单检查
	//    - 域名验证和过滤
	//    - 基础安全头检查
	//    - DDoS攻击检测
	//    - 恶意请求识别和拦截
	// 4. 全局CORS处理：处理跨域请求，添加必要的跨域响应头
	//    - 验证Origin是否在允许列表中
	//    - 添加Access-Control-Allow-Origin响应头
	//    - 添加Access-Control-Allow-Methods响应头
	//    - 添加Access-Control-Allow-Headers响应头
	//    - 处理OPTIONS预检请求并直接返回
	//    - 设置Access-Control-Max-Age缓存时间
	// 5. 全局认证鉴权：应用基础认证规则（认证在限流前，避免消耗资源）
	//    - API密钥验证：检查X-API-Key头
	//    - 基础Token验证：检查Authorization头
	//    - 签名验证：验证请求签名有效性
	//    - 客户端身份识别和权限检查
	//    - 设置用户上下文信息
	// 6. 全局限流控制：控制整个网关的总体流量
	//    - 基于IP的请求频率限制
	//    - 基于用户的请求频率限制
	//    - 基于API密钥的请求频率限制
	//    - 检查请求频率是否超过网关总阈值
	//    - 超过阈值则返回429 Too Many Requests状态码
	//    - 在响应头中返回限流信息（X-RateLimit-*）
	// 7. 路由匹配：根据请求路径和方法匹配路由规则
	//    - 支持精确匹配：/api/v1/users
	//    - 支持前缀匹配：/api/v1/*
	//    - 支持正则匹配：/api/v\d+/users/\d+
	//    - 支持参数提取：/users/{id}/posts/{postId}
	//    - 匹配成功后设置目标服务信息
	//    - 提取路由级别的配置信息，作为后续处理器的依据
	//    - 路由内部执行路由级别的处理器链：
	//      * 路由级安全控制：特定路由的安全策略
	//      * 路由级CORS处理：特定路由的跨域策略
	//      * 路由级认证鉴权：JWT/OAuth2/特定API Key验证
	//      * 路由级限流控制：特定API的独立限流阈值
	//      * 路由级熔断处理：特定路由或服务的熔断策略
	//      * 前置过滤器：请求预处理和转换
	// 8. 代理转发：将请求转发到目标服务
	//    - 服务发现：从注册中心查找可用的服务实例
	//    - 健康检查：过滤掉不健康的服务实例
	//    - 服务级熔断处理：特定服务的独立熔断策略
	//      * 跟踪服务调用的成功率和失败率
	//      * 监控响应时间和超时情况
	//      * 计算错误率是否超过阈值
	//      * 在服务故障时激活熔断保护
	//      * 熔断状态下快速失败，返回503 Service Unavailable
	//      * 定期尝试恢复，检测服务是否恢复健康
	//    - 负载均衡：使用轮询/权重/最少连接等算法选择目标实例
	//    - 请求转换：根据配置转换请求
	//      * 路径重写：/api/v1/users -> /users
	//      * 头部修改：添加/删除/修改HTTP头部
	//      * 参数转换：查询参数和路径参数转换
	//    - 发送请求：向上游服务发送HTTP请求
	//    - 超时控制：设置连接超时和读取超时
	//    - 重试机制：在失败时进行智能重试
	//    - 响应处理：接收上游响应并进行处理
	//      * 状态码处理：根据状态码进行相应处理
	//      * 响应转换：修改响应头部和内容
	//      * 错误处理：将上游错误转换为标准格式
	//    - 后置过滤器：响应后处理和转换
	//    - 返回响应：将处理后的响应返回给客户端
	// 9. 请求完成处理：
	//     - 记录访问日志
	//     - 统计请求耗时和状态码
	//     - 更新监控指标
	//     - 清理请求上下文

	// === 第一层：全局安全和基础控制 ===

	// 添加全局安全处理器（仅当启用时）
	if g.security != nil && g.gatewayConfig.Security.Enabled {
		engine.UseFunc(func(ctx *core.Context) bool {
			if !g.security.Handle(ctx) {
				logger.Warn("全局安全检查失败", "path", ctx.Request.URL.Path)
				return false
			}
			return true
		})
	}

	// 添加全局CORS处理器（仅当启用时）
	if g.cors != nil && g.gatewayConfig.CORS.Enabled {
		engine.UseFunc(func(ctx *core.Context) bool {
			if !g.cors.Handle(ctx) {
				logger.Debug("全局CORS检查失败", "path", ctx.Request.URL.Path)
				return false
			}
			return true
		})
	}

	// 添加全局认证处理器（仅当启用时）- 认证在限流前，避免无效请求消耗资源
	if g.auth != nil && g.gatewayConfig.Auth.Enabled {
		engine.UseFunc(func(ctx *core.Context) bool {
			if !g.auth.Handle(ctx) {
				logger.Warn("全局认证失败", "path", ctx.Request.URL.Path)
				return false
			}
			return true
		})
	}

	// 添加全局限流处理器（仅当启用且设置了速率时）
	if g.limiter != nil && g.gatewayConfig.RateLimit.Enabled && g.gatewayConfig.RateLimit.Rate > 0 {
		engine.UseFunc(func(ctx *core.Context) bool {
			if !g.limiter.Handle(ctx) {
				logger.Warn("全局限流触发", "path", ctx.Request.URL.Path)
				return false
			}
			return true
		})
	}

	// === 第二层：路由匹配和路由级别控制 ===

	// 添加路由处理器 - 路由匹配和路由级别的处理器链执行
	// 路由处理器内部会执行路由级别的安全、CORS、限流、熔断、认证处理
	engine.UseFunc(func(ctx *core.Context) bool {
		if !g.router.Handle(ctx) {
			logger.Debug("路由处理失败", "path", ctx.Request.URL.Path)
			return false
		}
		return true
	})

	// === 第三层：代理转发 ===

	// 添加代理处理器
	engine.UseFunc(func(ctx *core.Context) bool {
		if !g.proxy.Handle(ctx) {
			logger.Error("代理转发失败", "path", ctx.Request.URL.Path)
			return false
		}
		return true
	})
}

// ServeHTTP 实现http.Handler接口
func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 创建网关上下文，这个上下文将贯穿整个请求处理过程
	ctx := core.NewContext(w, r)
	// 使用Engine的HandleWithContext方法处理请求
	// 这样可以确保日志记录使用的是同一个上下文
	g.engine.HandleWithContext(ctx, w, r)
	// 设置响应时间
	ctx.SetResponseTime(time.Now())
	// 设置实例名称
	ctx.Set(constants.ContextKeyGatewayInstanceName, g.gatewayConfig.Base.Name)
	//设置日志配置ID
	ctx.Set(constants.ContextKeyLogConfigID, g.gatewayConfig.Log.LogConfigID)
	//设置租户ID
	ctx.Set(constants.ContextKeyTenantID, g.gatewayConfig.Log.TenantID)
	// 链路处理完成后，异步写入访问日志
	// 创建独立的context用于日志写入，避免HTTP请求context取消导致的问题
	go func() {
		// 创建独立的context，设置合理的超时时间
		logCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// 将原始的HTTP请求context替换为独立的context
		// 这样日志写入就不会因为HTTP请求结束而失败
		originalCtx := ctx.Ctx
		ctx.Ctx = logCtx

		// 确保在函数结束时恢复原始context（虽然这里不是必需的，但是好的实践）
		defer func() {
			ctx.Ctx = originalCtx
		}()

		if err := logwrite.WriteLog(g.gatewayConfig.InstanceID, ctx); err != nil {
			logger.Error("Failed to write access log", "error", err)
		}
	}()

}

// Start 启动网关
func (g *Gateway) Start() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.running {
		return fmt.Errorf("网关已经在运行")
	}

	// 设置处理器链
	g.setupHandlers(g.engine)

	// 创建一个通道用于接收启动错误
	errCh := make(chan error, 1)

	// 在启动前检查端口是否已被占用
	listener, err := net.Listen("tcp", g.server.Addr)
	if err != nil {
		// 端口占用或绑定失败，更新数据库状态
		g.updateHealthStatus("N", fmt.Sprintf("端口绑定失败: %v", err))
		return fmt.Errorf("端口 %s 已被占用或无法绑定: %w", g.server.Addr, err)
	}
	// 关闭测试用的监听器
	listener.Close()

	logger.Info("启动网关服务", "listen", g.gatewayConfig.Base.Listen)
	// 初始化日志处理器
	logwrite.InitLogManager(g.gatewayConfig.InstanceID, &g.gatewayConfig.Log)
	// 启动HTTP服务器
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()

		var err error
		if g.gatewayConfig.Base.EnableHTTPS {
			err = g.server.ListenAndServeTLS(g.gatewayConfig.Base.CertFile, g.gatewayConfig.Base.KeyFile)
		} else {
			err = g.server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP服务器启动失败", err)
			// 发送错误到通道
			select {
			case errCh <- err:
			default:
			}

			// 如果网关已标记为运行中，则停止它
			g.mu.Lock()
			if g.running {
				g.mu.Unlock()
				g.Stop()
			} else {
				g.mu.Unlock()
			}
		}
	}()

	// 等待短暂时间检查是否有立即出现的错误
	select {
	case err := <-errCh:
		// 启动失败，更新数据库状态
		g.updateHealthStatus("N", fmt.Sprintf("启动失败: %v", err))
		return fmt.Errorf("启动HTTP服务器失败: %w", err)
	case <-time.After(100 * time.Millisecond):
		// 没有立即出现错误，认为启动成功
		g.running = true
		g.stopCh = make(chan struct{})
		// 启动成功，更新数据库状态
		g.updateHealthStatus("Y", "")
		logger.Info("网关服务启动成功")
	}

	return nil
}

// Stop 停止网关
func (g *Gateway) Stop() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if !g.running {
		return nil
	}

	logger.Info("正在停止网关服务...")

	// 发送停止信号
	close(g.stopCh)

	// 清理处理器资源
	// 注意：必须先关闭处理器资源，再关闭HTTP服务器
	// 原因：
	// 1. 处理器可能包含后台goroutine（如健康检查器），需要先停止它们
	// 2. 避免处理器资源泄漏和zombie goroutine
	// 3. 确保所有资源被正确释放，防止内存泄漏

	// 优先关闭代理处理器，因为它通常包含健康检查器和服务发现组件
	// 这些组件会启动后台goroutine，如果不正确关闭会导致资源泄漏
	// 注意：这里使用类型断言(interface{ Close() error })而不是直接定义Close方法的接口
	// 这种设计的优势：
	// 1. 松耦合：处理器接口(RouterHandler, ProxyHandler等)不需要包含Close方法
	// 2. 可选实现：只有需要清理资源的处理器才需要实现Close方法
	// 3. 接口隔离：符合接口隔离原则，保持接口精简
	// 4. 向后兼容：添加新处理器时不强制实现Close方法
	// 5. 动态发现：运行时动态检测处理器是否需要清理资源
	if g.proxy != nil {
		if closer, ok := g.proxy.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				logger.Warn("关闭代理处理器失败", "error", err)
			} else {
				logger.Debug("代理处理器已关闭")
			}
		}
	}

	// 关闭其他处理器资源
	// 使用类型断言检查处理器是否实现了Close方法
	// 这种设计允许处理器自行决定是否需要清理资源
	// 符合接口隔离原则，不强制所有处理器实现Close方法
	if g.router != nil {
		if closer, ok := g.router.(interface{ Close() error }); ok {
			_ = closer.Close()
		}
	}

	if g.auth != nil {
		if closer, ok := g.auth.(interface{ Close() error }); ok {
			_ = closer.Close()
		}
	}

	if g.cors != nil {
		if closer, ok := g.cors.(interface{ Close() error }); ok {
			_ = closer.Close()
		}
	}

	if g.security != nil {
		if closer, ok := g.security.(interface{ Close() error }); ok {
			_ = closer.Close()
		}
	}

	if g.limiter != nil {
		if closer, ok := g.limiter.(interface{ Close() error }); ok {
			_ = closer.Close()
		}
	}
	// 关闭日志处理器
	logwrite.CloseLogWriter(g.gatewayConfig.InstanceID)

	// 关闭HTTP服务器
	// 设置30秒超时确保正在处理的请求有足够时间完成
	// 超时后会强制关闭，避免无限等待
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := g.server.Shutdown(ctx); err != nil {
		logger.Error("关闭HTTP服务器失败", err)
		return err
	}

	// 等待所有goroutine结束
	// 这确保了所有后台任务（包括请求处理）都已完成
	// 防止主进程退出时留下zombie goroutine
	g.wg.Wait()

	g.running = false
	logger.Info("网关服务已停止")

	return nil
}

// IsRunning 检查网关是否在运行
func (g *Gateway) IsRunning() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.running
}

// GetConfig 获取配置
func (g *Gateway) GetConfig() *config.GatewayConfig {
	return g.gatewayConfig
}

// updateHealthStatus 更新网关实例健康状态
func (g *Gateway) updateHealthStatus(healthStatus string, errorMsg string) {
	// 检查是否有实例ID和租户ID
	instanceId := g.gatewayConfig.InstanceID
	tenantId := g.gatewayConfig.Log.TenantID
	
	if instanceId == "" || tenantId == "" {
		logger.Debug("缺少instanceId或tenantId，跳过健康状态更新", "instanceId", instanceId, "tenantId", tenantId)
		return
	}
	
	// 调用静态方法更新健康状态
	dbloader.UpdateGatewayHealthStatus(tenantId, instanceId, healthStatus, errorMsg)
}

// Reload 重新加载网关配置
// 允许在不重启服务的情况下更新网关的配置
func (g *Gateway) Reload(newCfg *config.GatewayConfig) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if !g.running {
		return fmt.Errorf("网关未运行，无法重载配置")
	}

	// 使用工厂方法重载配置
	// 注意：ReloadGateway方法内部已经处理了engine的重建和处理器链设置
	factory := NewGatewayFactory()
	if err := factory.ReloadGateway(g, newCfg); err != nil {
		return fmt.Errorf("重载网关配置失败: %w", err)
	}
	// 重新初始化日志处理器
	err := logwrite.UpdateLogWriter(g.gatewayConfig.InstanceID, &g.gatewayConfig.Log)
	if err != nil {
		return fmt.Errorf("重载日志处理器失败: %w", err)
	}
	logger.Info("网关配置重载成功",
		"instanceId", g.gatewayConfig.InstanceID,
		"listen", g.gatewayConfig.Base.Listen)

	return nil
}
