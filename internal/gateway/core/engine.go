package core

import (
	"net/http"
	"sync"
	"time"

	"gateway/internal/gateway/constants"
	"gateway/internal/gateway/helper"
	"gateway/pkg/utils/random"
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
func (e *Engine) HandleWithContext(gatewayCtx *Context, w http.ResponseWriter, r *http.Request) {
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
	// 重要设计决策：使用requestStartTime而不是realStartTime
	// 原因如下：
	// 1. HTTP Keep-Alive连接复用：一个TCP连接可以处理多个HTTP请求
	//    - realStartTime是TCP连接建立时间，只在连接建立时设置一次
	//    - 同一连接上的多个HTTP请求会共享相同的realStartTime
	//    - 这会导致后续请求的"耗时"包含连接空闲时间，统计结果错误
	// 2. 业务语义准确性：我们要统计的是HTTP请求处理性能，不是连接性能
	//    - 第1个请求：耗时 = 响应时间 - realStartTime (包含连接建立等待)
	//    - 第2个请求：耗时 = 响应时间 - realStartTime (包含5秒空闲时间!)
	//    - 正确计算：耗时 = 响应时间 - requestStartTime (纯HTTP处理时间)
	// 3. 监控和SLA统计的实用性：
	//    - API响应时间监控需要的是单个请求的处理时间
	//    - 性能瓶颈分析需要准确的请求级别统计
	//    - 连接级别的时间统计应该单独实现，不应混合在请求统计中
	// 4. 实际场景示例：
	//    - 10:00:00 TCP连接建立 (realStartTime)
	//    - 10:00:01 第1个HTTP请求 (实际处理100ms)
	//    - 10:00:05 第2个HTTP请求 (实际处理80ms)
	//    - 如果用realStartTime：第2个请求"耗时"=5.08秒 (完全错误!)
	//    - 如果用requestStartTime：第2个请求耗时=80ms (正确!)
	ctx.startTime = requestStartTime

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
		response := helper.BuildGatewayResponse(
			constants.ErrorCodeRouteNotFound,
			constants.StatusMessageNotFound,
			"",
			r.URL.Path,
			traceID,
		)

		ctx.JSON(http.StatusNotFound, response)
		// 设置网关状态码
		ctx.Set(constants.GatewayStatusCode, http.StatusNotFound)
	}
}
