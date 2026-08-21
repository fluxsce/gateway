package core

import (
	"context"
	"fmt"
	"net/http"

	"gateway/internal/gateway/constants"
	"gateway/pkg/tracing"
)

// StartTracing 从入站请求提取或新建 Trace，并把 W3C traceparent 写回客户端响应头。
// 使用进程级 tracing.Global()，网关代际不持有 tracing.Tracer。r 为原始入站请求，spanName 为根 Span 名称。
func (c *Context) StartTracing(r *http.Request, spanName string, attrs ...tracing.Attr) {
	if c == nil {
		return
	}
	c.Ctx, _ = tracing.FromContext(c.Ctx).StartRequest(c.Ctx, r, spanName, attrs...)
	// 只在尚未写出响应时写头。WriteHeader / Hijack 之后再 Header().Set 不会 panic，但也不会发给客户端，WebSocket 升级后更不应再碰 Writer。
	if c.Writer != nil && !c.IsResponded() {
		if tp := tracing.TraceparentValue(c.Ctx); tp != "" {
			c.Writer.Header().Set(tracing.HeaderTraceparent, tp)
		}
	}
}

// FinishTracing 结束当前请求的根 Span。
// 在 End 之前补写路由、网关状态码，以及上下文中最后一个错误，便于和访问日志互查。
// 只读 Context 内存字段，不访问 Request / Writer，响应已写出或连接已 Hijack 时也可调用。
func (c *Context) FinishTracing() {
	if c == nil {
		return
	}
	scope := tracing.RequestScope(c.Ctx)
	// 路由匹配发生在 StartTracing 之后，结束时才能拿到最终 http.route
	if route := c.GetMatchedPath(); route != "" {
		scope.SetAttr(tracing.String("http.route", route))
	}
	if code, ok := c.GetInt(constants.GatewayStatusCode); ok && code > 0 {
		scope.SetAttr(tracing.String("http.response.status_code", fmt.Sprintf("%d", code)))
	}
	if len(c.Errors) > 0 {
		scope.RecordError(c.Errors[len(c.Errors)-1])
	}
	tracing.FinishRequest(c.Ctx)
}

// InjectTracing 将当前请求根 Trace 写入出站请求头。
// WebSocket 握手等没有独立 client Span 的路径使用本方法；HTTP 代理应使用 StartOutboundSpan 返回的 context 再 Inject。
func (c *Context) InjectTracing(header http.Header) {
	if c == nil {
		return
	}
	tracing.Inject(c.Ctx, header)
}

// WithSpan 为一段返回 bool 的处理器创建内部 Span。
// fn 返回 false 时把上下文最后一个错误记到 Span 上。nil 接收者或 nil fn 视为放行。
func (c *Context) WithSpan(name string, fn func() bool) bool {
	if c == nil {
		if fn == nil {
			return true
		}
		return fn()
	}
	if fn == nil {
		return true
	}
	// 阶段结束后恢复根 context，避免后续出站 Span 挂到已结束的认证/限流 Span 上
	parent := c.Ctx
	var span tracing.Scope
	c.Ctx, span = tracing.StartSpan(c.Ctx, name)
	ok := fn()
	if !ok && len(c.Errors) > 0 {
		span.RecordError(c.Errors[len(c.Errors)-1])
	}
	span.End()
	c.Ctx = parent
	return ok
}

// StartOutboundSpan 创建出站调用 Span，返回该 Span 的 context 供 Inject 使用。
// 不改写请求根 c.Ctx，避免多服务并行转发互相覆盖。c 为 nil 时返回空操作 Span。
func (c *Context) StartOutboundSpan(name string, attrs ...tracing.Attr) (context.Context, tracing.Scope) {
	if c == nil {
		return tracing.NewNoop().StartClientSpan(context.Background(), name, attrs...)
	}
	return tracing.StartClientSpan(c.Ctx, name, attrs...)
}
