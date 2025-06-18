package core

import (
	"net/http"
	"sync"
	"time"

	"gohub/internal/gateway/constants"
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

// Handle 处理HTTP请求
func (e *Engine) Handle(w http.ResponseWriter, r *http.Request) {
	// 记录请求开始处理时间
	requestStartTime := time.Now()

	// 尝试获取真实的连接建立时间
	var realStartTime time.Time
	if r.Context() != nil {
		if connTime := r.Context().Value(constants.ContextKeyConnectionStartTime); connTime != nil {
			if t, ok := connTime.(time.Time); ok {
				realStartTime = t
			}
		}
	}

	// 如果没有获取到连接时间，使用请求处理开始时间
	if realStartTime.IsZero() {
		realStartTime = requestStartTime
	}

	ctx := NewContext(w, r)

	// 设置时间信息到上下文
	ctx.startTime = realStartTime
	ctx.Set(constants.ContextKeyRequestProcessingStart, requestStartTime)

	// 执行处理器链
	e.chain.Execute(ctx)

	// 如果请求还没有被响应，返回404错误
	if !ctx.IsResponded() {
		ctx.JSON(http.StatusNotFound, map[string]interface{}{
			"error": "Not found",
		})
	}
}

// ServeHTTP 实现http.Handler接口
func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e.Handle(w, r)
}
