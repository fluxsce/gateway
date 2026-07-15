package bootstrap

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
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
	"gateway/internal/gateway/helper"
	"gateway/internal/gateway/helper/reqhand"
	"gateway/internal/gateway/loader/dbloader"
	"gateway/internal/gateway/logwrite"
	appconfig "gateway/pkg/config"
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
	running  bool
	stopping bool

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

	// currentGeneration 指向新连接和直接调用ServeHTTP时使用的当前运行时代际。
	currentGeneration atomic.Pointer[gatewayGeneration]
	// dispatcher 持有唯一的底层监听端口，并把新连接分发给当前Server代际。
	dispatcher *listenerDispatcher
	// generationWG 等待后台排空的旧代际完成。
	generationWG sync.WaitGroup
	// requestLimiter 对所有运行时代际实施统一的在途请求上限。
	requestLimiter requestAdmissionLimiter
}

// setCompatibilityHandlers 更新原有处理器字段，供现有管理接口和测试继续访问。
func (g *Gateway) setCompatibilityHandlers(handlers gatewayHandlers) {
	g.router = handlers.router
	g.proxy = handlers.proxy
	g.auth = handlers.auth
	g.cors = handlers.cors
	g.security = handlers.security
	g.limiter = handlers.limiter
}

// installCompatibilityGeneration 更新原有配置、Server、Engine及处理器字段。
func (g *Gateway) installCompatibilityGeneration(generation *gatewayGeneration) {
	g.gatewayConfig = generation.config
	g.server = generation.server
	g.engine = generation.engine
	g.setCompatibilityHandlers(generation.handlers)
	g.requestLimiter.setLimit(generation.config.Base.MaxWorkers)
}

// setupHandlers 设置处理器链 - 网关处理的核心思想
func (g *Gateway) setupHandlers(engine *core.Engine) {
	g.setupHandlersFor(engine, g.gatewayConfig, gatewayHandlers{
		router:   g.router,
		proxy:    g.proxy,
		auth:     g.auth,
		cors:     g.cors,
		security: g.security,
		limiter:  g.limiter,
	})
}

// setupHandlersFor 使用不可变配置和处理器集合构建指定代际的处理链。
func (g *Gateway) setupHandlersFor(engine *core.Engine, cfg *config.GatewayConfig, handlers gatewayHandlers) {
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
	if handlers.security != nil && cfg.Security.Enabled {
		engine.UseFunc(func(ctx *core.Context) bool {
			if !handlers.security.Handle(ctx) {
				logger.Warn("全局安全检查失败", "path", ctx.Request.URL.Path)
				return false
			}
			return true
		})
	}

	// 添加全局CORS处理器（仅当启用时）
	if handlers.cors != nil && cfg.CORS.Enabled {
		engine.UseFunc(func(ctx *core.Context) bool {
			if !handlers.cors.Handle(ctx) {
				logger.Debug("全局CORS检查失败", "path", ctx.Request.URL.Path)
				return false
			}
			return true
		})
	}

	// 添加全局认证处理器（仅当启用时）- 认证在限流前，避免无效请求消耗资源
	if handlers.auth != nil && cfg.Auth.Enabled {
		engine.UseFunc(func(ctx *core.Context) bool {
			if !handlers.auth.Handle(ctx) {
				logger.Warn("全局认证失败", "path", ctx.Request.URL.Path)
				return false
			}
			return true
		})
	}

	// 添加全局限流处理器（仅当启用且设置了速率时）
	if handlers.limiter != nil && cfg.RateLimit.Enabled && cfg.RateLimit.Rate > 0 {
		engine.UseFunc(func(ctx *core.Context) bool {
			if !handlers.limiter.Handle(ctx) {
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
		if !handlers.router.Handle(ctx) {
			logger.Debug("路由处理失败", "path", ctx.Request.URL.Path)
			return false
		}
		return true
	})

	// === 第三层：代理转发 ===

	// 添加代理处理器
	engine.UseFunc(func(ctx *core.Context) bool {
		if !handlers.proxy.Handle(ctx) {
			logger.Debug("代理转发失败", "path", ctx.Request.URL.Path)
			return false
		}
		return true
	})
}

// ServeHTTP 实现http.Handler接口
func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	generation := g.currentGeneration.Load()
	if generation != nil {
		g.serveHTTPGeneration(generation, w, r)
		return
	}
	g.serveHTTPWithRuntime(g.gatewayConfig, g.engine, w, r)
}

// serveHTTPGeneration 使用连接绑定的固定代际处理请求。
func (g *Gateway) serveHTTPGeneration(generation *gatewayGeneration, w http.ResponseWriter, r *http.Request) {
	if !generation.acquire() {
		ctx, traceID := g.prepareRequestContext(generation.config, w, r)
		defer ctx.Cancel()
		ctx.AddError(fmt.Errorf("网关运行时代际正在排空"))
		if r.ProtoMajor == 1 {
			w.Header().Set("Connection", "close")
		}
		ctx.Abort(http.StatusServiceUnavailable, helper.BuildGatewayResponse(
			constants.ErrorCodeServiceUnavailable,
			constants.StatusMessageServiceUnavailable,
			"",
			r.URL.Path,
			traceID,
		))
		g.finishRequest(ctx, generation.config)
		return
	}
	defer generation.release()
	g.serveHTTPWithRuntime(generation.config, generation.engine, w, r)
}

// serveHTTPWithRuntime 执行原有请求处理流程，配置和引擎在请求期间保持不变。
func (g *Gateway) serveHTTPWithRuntime(cfg *config.GatewayConfig, engine *core.Engine, w http.ResponseWriter, r *http.Request) {
	ctx, traceID := g.prepareRequestContext(cfg, w, r)
	defer ctx.Cancel()
	if !g.requestLimiter.tryAcquire() {
		err := fmt.Errorf("网关当前在途请求数已达到上限")
		ctx.AddError(err)
		w.Header().Set("Retry-After", "1")
		ctx.Abort(http.StatusServiceUnavailable, helper.BuildGatewayResponse(
			constants.ErrorCodeGatewayOverloaded,
			constants.StatusMessageServiceUnavailable,
			"",
			r.URL.Path,
			traceID,
		))
		g.finishRequest(ctx, cfg)
		return
	}
	func() {
		defer g.requestLimiter.release()
		// 使用Engine的HandleWithContext方法处理请求
		// 这样可以确保日志记录使用的是同一个上下文
		engine.HandleWithContext(ctx, w, r)
	}()
	g.finishRequest(ctx, cfg)
}

// prepareRequestContext 创建请求上下文并注入日志、实例及trace信息。
func (g *Gateway) prepareRequestContext(cfg *config.GatewayConfig, w http.ResponseWriter, r *http.Request) (*core.Context, string) {
	// 创建网关上下文，这个上下文将贯穿整个请求处理过程
	ctx := core.NewContext(w, r)
	// 直连端口重发：从请求头注入原始 trace 与重发标记（读入后从头中删除，避免透传上游）
	applyGatewayReplayHeaders(r, ctx)
	// 设置实例ID（需要在处理器链执行前设置，供后端追踪日志使用）
	ctx.Set(constants.ContextKeyGatewayInstanceID, cfg.InstanceID)
	// 设置实例名称
	ctx.Set(constants.ContextKeyGatewayInstanceName, cfg.Base.Name)
	//设置日志配置ID
	ctx.Set(constants.ContextKeyLogConfigID, cfg.Log.LogConfigID)
	//设置租户ID
	ctx.Set(constants.ContextKeyTenantID, cfg.Log.TenantID)
	// 直接设置日志配置到上下文，避免重复获取
	ctx.SetLogConfig(&cfg.Log)
	traceID := core.InitializeRequestContext(ctx)
	return ctx, traceID
}

// finishRequest 固化响应时间和HTTP快照，并异步写入访问日志。
func (g *Gateway) finishRequest(ctx *core.Context, cfg *config.GatewayConfig) {
	// 响应时间必须在快照和异步日志之前记录，避免日志准备耗时混入请求处理耗时。
	ctx.SetResponseTime(time.Now())
	if !cfg.Base.EnableAccessLog {
		return
	}
	// 在 Handler 完成后、启动异步写入前，立即缓存 HTTP 对象（Request、Writer）中的必要信息
	// 重要：不能在异步 goroutine 中直接访问 ctx.Request、ctx.Writer
	// 因为这些对象的生命周期与 HTTP 请求绑定，ServeHTTP 返回后可能被回收
	g.snapshotHTTPData(ctx, cfg.InstanceID)
	// 异步写入访问日志
	// Context 对象本身可以安全使用（data、时间字段等），只是不能访问 Request 和 Writer
	go func() {
		// 添加panic恢复机制，防止日志写入错误导致整个服务崩溃
		defer func() {
			if r := recover(); r != nil {
				logger.Error("Panic in access log writer", "error", r, "instanceID", cfg.InstanceID)
			}
		}()

		// 创建独立的context用于日志写入，替换原来的 HTTP 请求 context
		logCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// 替换上下文中的 context.Context，避免使用已取消的 HTTP 请求 context
		originalCtx := ctx.Ctx
		ctx.Ctx = logCtx
		defer func() {
			ctx.Ctx = originalCtx
		}()

		// 写入访问日志
		// 注意：logwrite.WriteLog 内部不能访问 ctx.Request 和 ctx.Writer
		// 所有需要的数据都已经缓存在 ctx.data 中
		if err := logwrite.WriteLog(cfg.InstanceID, ctx); err != nil {
			logger.Error("Failed to write access log", "error", err)
		}
	}()
}

// snapshotHTTPData 在 Handler 完成后立即缓存 HTTP 对象中的必要数据到上下文
// 重要：必须在 ServeHTTP 返回前调用，因为 HTTP 对象（Request、Writer）的生命周期与请求绑定
// 缓存后，异步 goroutine 可以安全使用 Context 对象，但不能直接访问 Request 和 Writer
func (g *Gateway) snapshotHTTPData(ctx *core.Context, instanceID string) {
	// 使用 defer recover 防止快照过程中的任何错误
	defer func() {
		if r := recover(); r != nil {
			logger.Warn("Failed to snapshot HTTP data", "error", r)
		}
	}()

	// 使用 reqhand 包的通用快照方法
	reqhand.SnapshotHTTPData(ctx)

	logger.Debug("HTTP data snapshot created for async logging", "instanceID", instanceID)
}

// startGenerationServer 启动一个绑定虚拟listener的HTTP Server代际。
func (g *Gateway) startGenerationServer(generation *gatewayGeneration) {
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		defer close(generation.serveDone)

		var err error
		if generation.config.Base.EnableHTTPS {
			// 使用TLSConfig启动HTTPS服务器，证书已加载到TLSConfig中。
			err = generation.server.ServeTLS(generation.listener, "", "")
		} else {
			err = generation.server.Serve(generation.listener)
		}
		if err != nil && err != http.ErrServerClosed && err != net.ErrClosed {
			logger.Error("HTTP服务器代际异常退出", "error", err)
		}
	}()
}

// waitGenerationReady 等待Server完成协议初始化并进入Accept状态。
func (g *Gateway) waitGenerationReady(generation *gatewayGeneration) error {
	select {
	case <-generation.listener.readyCh:
		return nil
	case <-generation.serveDone:
		return fmt.Errorf("HTTP服务器代际在就绪前退出")
	case <-time.After(5 * time.Second):
		return fmt.Errorf("等待HTTP服务器代际就绪超时")
	}
}

// activateGeneration 发布新代际，并在后台排空旧代际。
func (g *Gateway) activateGeneration(generation *gatewayGeneration) error {
	if g.dispatcher == nil {
		return fmt.Errorf("连接分发器未初始化")
	}
	generation.listener = g.dispatcher.newGenerationListener()
	g.startGenerationServer(generation)
	if err := g.waitGenerationReady(generation); err != nil {
		_ = generation.server.Close()
		<-generation.serveDone
		return err
	}
	g.requestLimiter.setLimit(generation.config.Base.MaxWorkers)
	g.dispatcher.setMaxConnections(generation.config.Base.MaxConnections)
	g.dispatcher.switchTo(generation.listener)

	old := g.currentGeneration.Swap(generation)
	g.installCompatibilityGeneration(generation)
	if old != nil {
		g.generationWG.Add(1)
		go func() {
			defer g.generationWG.Done()
			if err := g.drainGeneration(old); err != nil {
				logger.Warn("旧网关代际排空失败", "error", err)
			}
		}()
	}
	return nil
}

// drainGeneration 停止旧代际接收请求，并在配置的时间内排空连接和处理器。
func (g *Gateway) drainGeneration(generation *gatewayGeneration) error {
	generation.beginDrain()
	timeout := generation.gracefulTimeout()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// HTTP Server和被劫持的WebSocket会话使用同一个排空期限并行关闭。
	proxyShutdownDone := make(chan error, 1)
	if shutdowner, ok := generation.handlers.proxy.(interface{ Shutdown(context.Context) error }); ok {
		go func() {
			proxyShutdownDone <- shutdowner.Shutdown(ctx)
		}()
	} else {
		proxyShutdownDone <- nil
	}
	shutdownErr := generation.server.Shutdown(ctx)
	proxyShutdownErr := <-proxyShutdownDone
	drained := generation.waitInflight(ctx)
	if !drained {
		// 超时后统一强制关闭HTTP连接和被劫持会话，再给退出协程一个有限回收窗口。
		_ = generation.server.Close()
		if closer, ok := generation.handlers.proxy.(interface{ ForceClose() }); ok {
			closer.ForceClose()
		}
		forceCtx, forceCancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = generation.waitInflight(forceCtx)
		forceCancel()
	}
	generation.closeHandlers()
	<-generation.serveDone
	if shutdownErr == nil && proxyShutdownErr != nil &&
		proxyShutdownErr != context.Canceled && proxyShutdownErr != context.DeadlineExceeded {
		shutdownErr = proxyShutdownErr
	}
	return shutdownErr
}

// Start 启动网关
func (g *Gateway) Start() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.running || g.stopping {
		return fmt.Errorf("网关已经在运行")
	}

	generation := g.currentGeneration.Load()
	if generation == nil {
		// Stop会关闭Server和处理器；再次Start时必须创建全新代际，不能复用已关闭资源。
		var err error
		generation, err = NewGatewayFactory().buildGeneration(g, g.gatewayConfig)
		if err != nil {
			return fmt.Errorf("重新构建网关运行时代际失败: %w", err)
		}
		if err := logwrite.InitLogManager(generation.config.InstanceID, &generation.config.Log); err != nil {
			generation.closeHandlers()
			return fmt.Errorf("重新初始化日志处理器失败: %w", err)
		}
		g.installCompatibilityGeneration(generation)
		g.currentGeneration.Store(generation)
	}

	// 在启动前检查端口是否已被占用
	listener, err := net.Listen("tcp", g.server.Addr)
	if err != nil {
		// 端口占用或绑定失败，更新数据库状态
		g.updateHealthStatus("N", fmt.Sprintf("端口绑定失败: %v", err))
		return fmt.Errorf("端口 %s 已被占用或无法绑定: %w", g.server.Addr, err)
	}

	logger.Info("启动网关服务", "listen", g.gatewayConfig.Base.Listen)
	// 初始化日志处理器
	logwrite.InitLogManager(g.gatewayConfig.InstanceID, &g.gatewayConfig.Log)

	// 绑定成功后直接复用同一个底层listener，避免检查端口后再次监听的竞态窗口。
	dispatcher := newListenerDispatcher(listener)
	generation.listener = dispatcher.newGenerationListener()
	g.startGenerationServer(generation)
	if err := g.waitGenerationReady(generation); err != nil {
		_ = generation.server.Close()
		<-generation.serveDone
		_ = listener.Close()
		generation.closeHandlers()
		return err
	}
	g.requestLimiter.setLimit(generation.config.Base.MaxWorkers)
	dispatcher.setMaxConnections(generation.config.Base.MaxConnections)
	dispatcher.switchTo(generation.listener)
	g.dispatcher = dispatcher
	dispatcher.start()

	g.running = true
	g.stopping = false
	g.stopCh = make(chan struct{})
	// 启动成功，更新数据库状态
	g.updateHealthStatus("Y", "")
	logger.Info("网关服务启动成功")
	return nil
}

// Stop 停止网关
func (g *Gateway) Stop() error {
	g.mu.Lock()
	if !g.running || g.stopping {
		g.mu.Unlock()
		return nil
	}
	g.stopping = true
	dispatcher := g.dispatcher
	current := g.currentGeneration.Load()
	g.mu.Unlock()

	logger.Info("正在停止网关服务...")

	// 发送停止信号
	select {
	case <-g.stopCh:
	default:
		close(g.stopCh)
	}

	// 先关闭唯一的底层listener，确保不再接收新连接。
	if dispatcher != nil {
		if err := dispatcher.Close(); err != nil && err != net.ErrClosed {
			logger.Warn("关闭连接分发器失败", "error", err)
		}
	}

	// 清理处理器资源。
	// 处理器可能包含后台goroutine（如健康检查器、服务发现组件），必须通过可选Close接口释放；
	// 代际先等待正在处理的请求，再按代理、路由、认证、CORS、安全、限流顺序关闭资源。
	var stopErr error
	if current != nil {
		stopErr = g.drainGeneration(current)
	}
	// 等待Reload期间已经进入后台排空的旧代际结束。
	g.generationWG.Wait()

	// 关闭日志处理器
	instanceID := g.gatewayConfig.InstanceID
	logwrite.CloseLogWriter(instanceID)

	// 等待所有goroutine结束
	// 这确保了所有后台任务（包括请求处理）都已完成
	// 防止主进程退出时留下zombie goroutine
	g.wg.Wait()

	g.mu.Lock()
	g.running = false
	g.stopping = false
	g.dispatcher = nil
	g.currentGeneration.CompareAndSwap(current, nil)
	g.mu.Unlock()
	// 实例随进程退出而停止时，starter 已先置 IsInstanceStopping；此时不再把库中健康状态改为 N，
	// 避免与其它节点或注册中心对实例存活判断不一致（由集群/下线流程统一收敛状态）。
	if !appconfig.IsInstanceStopping() {
		g.updateHealthStatus("N", "")
	} else {
		logger.Info("进程停止流程中关闭网关，跳过实例健康状态落库",
			"instanceId", g.gatewayConfig.InstanceID)
	}
	logger.Info("网关服务已停止")

	return stopErr
}

// IsRunning 检查网关是否在运行
func (g *Gateway) IsRunning() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.running
}

// GetConfig 获取配置
func (g *Gateway) GetConfig() *config.GatewayConfig {
	if generation := g.currentGeneration.Load(); generation != nil {
		return generation.config
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
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

	if !g.running || g.stopping {
		return fmt.Errorf("网关未运行，无法重载配置")
	}

	oldCfg := g.gatewayConfig
	oldInstanceID := oldCfg.InstanceID
	newInstanceID := newCfg.InstanceID
	if newInstanceID == "" {
		newInstanceID = oldInstanceID
	}
	instanceIDChanged := newInstanceID != oldInstanceID

	// 日志写入器先完成可回滚的预更新，避免代际发布后才发现日志配置不可用。
	var logErr error
	if instanceIDChanged {
		logErr = logwrite.InitLogManager(newInstanceID, &newCfg.Log)
	} else {
		logErr = logwrite.UpdateLogWriter(oldInstanceID, &newCfg.Log)
	}
	if logErr != nil {
		return fmt.Errorf("预更新日志处理器失败: %w", logErr)
	}

	// 使用工厂方法重载配置
	// 注意：ReloadGateway方法内部已经处理了engine的重建和处理器链设置
	factory := NewGatewayFactory()
	if err := factory.ReloadGateway(g, newCfg); err != nil {
		if instanceIDChanged {
			_ = logwrite.CloseLogWriter(newInstanceID)
		} else {
			_ = logwrite.UpdateLogWriter(oldInstanceID, &oldCfg.Log)
		}
		return fmt.Errorf("重载网关配置失败: %w", err)
	}
	if instanceIDChanged {
		_ = logwrite.CloseLogWriter(oldInstanceID)
	}
	logger.Info("网关配置重载成功",
		"instanceId", g.gatewayConfig.InstanceID,
		"listen", g.gatewayConfig.Base.Listen)

	return nil
}
