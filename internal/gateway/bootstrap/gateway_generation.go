package bootstrap

import (
	"context"
	"net/http"
	"sync"
	"time"

	"gateway/internal/gateway/config"
	"gateway/internal/gateway/core"
	"gateway/internal/gateway/handler/auth"
	"gateway/internal/gateway/handler/cors"
	"gateway/internal/gateway/handler/limiter"
	"gateway/internal/gateway/handler/proxy"
	"gateway/internal/gateway/handler/router"
	"gateway/internal/gateway/handler/security"
	"gateway/pkg/logger"
)

// gatewayHandlers 保存一个运行时代际独占的处理器集合。
type gatewayHandlers struct {
	router   router.RouterHandler
	proxy    proxy.ProxyHandler
	auth     auth.Authenticator
	cors     cors.CORSHandler
	security security.SecurityHandler
	limiter  limiter.LimiterHandler
}

// close 按资源依赖顺序关闭已经构建的处理器集合。
func (h gatewayHandlers) close() {
	closeGenerationHandler("proxy", h.proxy)
	closeGenerationHandler("router", h.router)
	closeGenerationHandler("auth", h.auth)
	closeGenerationHandler("cors", h.cors)
	closeGenerationHandler("security", h.security)
	closeGenerationHandler("limiter", h.limiter)
}

// gatewayGeneration 表示一次不可变的网关运行时配置及其服务资源。
// 新请求只能获取当前未排空的代际，旧代际在请求结束后统一释放。
type gatewayGeneration struct {
	config   *config.GatewayConfig
	engine   *core.Engine
	handlers gatewayHandlers
	server   *http.Server

	listener  *virtualListener
	serveDone chan struct{}

	gateMu    sync.Mutex
	draining  bool
	inflight  sync.WaitGroup
	closeOnce sync.Once
}

// acquire 为请求获取当前代际；代际进入排空后拒绝新增请求。
func (g *gatewayGeneration) acquire() bool {
	g.gateMu.Lock()
	defer g.gateMu.Unlock()
	if g.draining {
		return false
	}
	g.inflight.Add(1)
	return true
}

// release 释放请求持有的代际引用。
func (g *gatewayGeneration) release() {
	g.inflight.Done()
}

// beginDrain 阻止旧代际继续接收新请求。
func (g *gatewayGeneration) beginDrain() {
	g.gateMu.Lock()
	g.draining = true
	g.gateMu.Unlock()
}

// waitInflight 等待代际内请求完成，超过截止时间时返回false。
func (g *gatewayGeneration) waitInflight(ctx context.Context) bool {
	done := make(chan struct{})
	go func() {
		g.inflight.Wait()
		close(done)
	}()
	select {
	case <-done:
		return true
	case <-ctx.Done():
		return false
	}
}

// closeHandlers 关闭该代际持有的全部处理器资源。
func (g *gatewayGeneration) closeHandlers() {
	g.closeOnce.Do(func() {
		g.handlers.close()
	})
}

// gracefulTimeout 返回该代际的排空上限。
func (g *gatewayGeneration) gracefulTimeout() time.Duration {
	timeout := g.config.Base.GracefulShutdownTimeout
	if timeout <= 0 {
		return 30 * time.Second
	}
	return timeout
}

func closeGenerationHandler(name string, handler interface{}) {
	// 继续使用可选Close接口保持松耦合和接口隔离：
	// 只有持有后台任务或连接资源的处理器需要实现Close，新处理器不被强制要求。
	if closer, ok := handler.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			logger.Warn("关闭旧代际处理器失败", "handler", name, "error", err)
		}
	}
}

// generationHTTPHandler 将连接固定到创建它的运行时代际。
type generationHTTPHandler struct {
	gateway    *Gateway
	generation *gatewayGeneration
}

// ServeHTTP 使用固定代际处理请求，确保热重载不会改变既有连接的处理器。
func (h *generationHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.gateway.serveHTTPGeneration(h.generation, w, r)
}
