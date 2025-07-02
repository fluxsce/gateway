package core

import (
	"net/http"
	"sync"
	"time"

	"gohub/internal/gateway/constants"
	"gohub/pkg/utils/random"
)

// Engine 是网关的核心引擎，协调所有处理器处理请求
// 负责接收HTTP请求并将其分发给注册的处理器链
type Engine struct {
	// 处理器链
	chain *HandlerChain

	// 锁，保护引擎的并发访问
	mu sync.RWMutex
}

// NewEngine 创建新的网关引擎
func NewEngine() *Engine {
	return &Engine{
		chain: NewHandlerChain(),
	}
}

// Use 添加处理器到引擎
// 参数:
// - handler: 实现了Handler接口的处理器
// 返回值:
// - 引擎自身，支持链式调用
// 将处理器添加到处理器链的末尾，会按添加顺序执行
func (e *Engine) Use(handler Handler) *Engine {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.chain.Add(handler)
	return e
}

// UseFunc 添加处理器函数到引擎
// 参数:
// - handlerFunc: 处理器函数
// 返回值:
// - 引擎自身，支持链式调用
// 便捷方法，将函数转换为Handler接口后添加到处理器链
func (e *Engine) UseFunc(handlerFunc HandlerFunc) *Engine {
	return e.Use(handlerFunc)
}

// Handle 是网关的核心处理引擎，负责处理所有HTTP请求的生命周期管理
//
// 核心职责：
// 1. 请求生命周期管理：从接收请求到响应完成的全过程控制
// 2. 链路追踪：为每个请求生成唯一的trace_id，便于日志追踪和问题排查
// 3. 时间统计：记录请求处理的各个时间节点，用于性能分析
// 4. 上下文管理：创建和管理请求上下文，在整个处理链中传递状态
// 5. 处理器链执行：按顺序执行所有注册的处理器（路由、认证、限流、代理等）
// 6. 错误处理：统一处理未匹配路由的404错误响应
// 7. 响应保障：确保每个请求都能得到响应，避免连接悬挂
//
// 处理流程：
// 1. 生成唯一trace_id用于链路追踪
// 2. 记录请求开始处理时间和连接建立时间
// 3. 创建请求上下文并设置时间信息和trace_id
// 4. 执行完整的处理器链（认证->限流->路由->代理等）
// 5. 检查响应状态，未响应的请求返回404错误
//
// 参数:
// - w: HTTP响应写入器，用于向客户端发送响应
// - r: HTTP请求对象，包含客户端请求的所有信息
func (e *Engine) HandleWithContext(gatewayCtx *Context,w http.ResponseWriter, r *http.Request) {
	// 生成32位唯一trace_id用于链路追踪和日志关联
	// 使用并发安全的唯一ID生成器，确保在高并发环境下的唯一性
	// 这个trace_id将贯穿整个请求处理过程，便于问题排查和性能分析
	traceID := random.Generate32BitRandomString()

	// 记录请求开始处理时间
	requestStartTime := time.Now()

	// 尝试获取真实的连接建立时间
	// 某些场景下（如负载均衡器、代理服务器）可能会在连接建立时设置此时间
	var realStartTime time.Time
	if r.Context() != nil {
		if connTime := r.Context().Value(constants.ContextKeyConnectionStartTime); connTime != nil {
			if t, ok := connTime.(time.Time); ok {
				realStartTime = t
			}
		}
	}

	// 如果没有获取到连接时间，使用请求处理开始时间作为fallback
	if realStartTime.IsZero() {
		realStartTime = requestStartTime
	}

	// 创建请求上下文，这是整个请求处理过程中的核心数据结构
	// 包含请求、响应、状态信息等，会在整个处理器链中传递
	ctx := gatewayCtx

	// 设置链路追踪ID到上下文，用于日志记录和问题排查
	ctx.Set(constants.ContextKeyTraceID, traceID)

	// 设置时间信息到上下文，用于性能统计和监控
	ctx.startTime = realStartTime
	ctx.Set(constants.ContextKeyRequestProcessingStart, requestStartTime)

	// 执行完整的处理器链
	// 这是网关的核心处理逻辑，会依次执行：
	// - 认证处理器：验证用户身份和权限
	// - 限流处理器：控制请求频率，防止系统过载
	// - 路由处理器：根据请求路径匹配对应的后端服务
	// - 代理处理器：将请求转发到目标服务并返回响应
	// - 日志处理器：记录访问日志和性能指标
	// - 其他自定义处理器
	e.chain.Execute(ctx)

	// 请求处理完成后的保障机制
	// 如果经过完整的处理器链后请求仍未被响应，说明没有匹配的路由
	// 此时返回标准的404错误，确保客户端能收到明确的响应
	if !ctx.IsResponded() {
		ctx.JSON(http.StatusNotFound, map[string]interface{}{
			"error":    "Not found",
			"trace_id": traceID, // 在错误响应中也包含trace_id，便于问题追踪
			"path":     r.URL.Path,
			"method":   r.Method,
		})
	}
}

