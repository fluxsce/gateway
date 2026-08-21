package tracing

import (
	"context"
	"net/http"
)

// contextKey 是本包私有的 context.Value 键类型。
// 用 iota 生成互不相同的整型键，避免与业务或其它库的字符串键冲突。
// 键只在进程内存中使用，不序列化；新增键追加在常量块末尾即可。
type contextKey int

const (
	// keyTracer 当前请求绑定的 Tracer。
	keyTracer contextKey = iota
	// keyRootScope 当前请求的根 Span。
	keyRootScope
	// keyTraceparent 待写入出站/响应头的 W3C traceparent。
	keyTraceparent
	// keyTracestate 待写入出站头的 W3C tracestate。
	keyTracestate
)

// WithTracer 将 Tracer 放入 context，供后续 Inject / StartSpan 从同一请求取出。
// ctx 或 tracer 为 nil 时分别使用 Background 与空操作实现。
func WithTracer(ctx context.Context, tracer Tracer) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if tracer == nil {
		tracer = NewNoop()
	}
	return context.WithValue(ctx, keyTracer, tracer)
}

// FromContext 取出当前 context 上的 Tracer。缺失时返回 Global，再缺失则空操作，调用方不必判空。
func FromContext(ctx context.Context) Tracer {
	if ctx != nil {
		if tracer, ok := ctx.Value(keyTracer).(Tracer); ok && tracer != nil {
			return tracer
		}
	}
	return Global()
}

// Inject 将当前 Trace 写入出站请求头。关闭追踪或 header 为 nil 时为空操作。
func Inject(ctx context.Context, header http.Header) {
	FromContext(ctx).Inject(ctx, header)
}

// StartSpan 创建内部阶段 Span，等价于 FromContext(ctx).StartSpan。
func StartSpan(ctx context.Context, name string, attrs ...Attr) (context.Context, Scope) {
	return FromContext(ctx).StartSpan(ctx, name, attrs...)
}

// StartClientSpan 创建出站调用 Span，等价于 FromContext(ctx).StartClientSpan。
// 返回的 context 仅用于本次出站 Inject，不要写回请求根 context。
func StartClientSpan(ctx context.Context, name string, attrs ...Attr) (context.Context, Scope) {
	return FromContext(ctx).StartClientSpan(ctx, name, attrs...)
}

// FinishRequest 结束当前请求的根 Span。缺失根 Span 时为空操作。
func FinishRequest(ctx context.Context) {
	RequestScope(ctx).End()
}

// RequestScope 返回当前请求的根 Span。缺失时返回空操作 Scope，可安全调用 End / SetAttr。
func RequestScope(ctx context.Context) Scope {
	if ctx == nil {
		return globalNoopScope
	}
	scope, ok := ctx.Value(keyRootScope).(Scope)
	if !ok || scope == nil {
		return globalNoopScope
	}
	return scope
}

// TraceparentValue 返回当前 context 中待传播的 W3C traceparent，缺失时返回空字符串。
func TraceparentValue(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	value, _ := ctx.Value(keyTraceparent).(string)
	return value
}

// TracestateValue 返回当前 context 中的 W3C tracestate，缺失时返回空字符串。
func TracestateValue(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	value, _ := ctx.Value(keyTracestate).(string)
	return value
}

func withRootScope(ctx context.Context, scope Scope) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, keyRootScope, scope)
}

func withTraceparent(ctx context.Context, traceparent, tracestate string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx = context.WithValue(ctx, keyTraceparent, traceparent)
	if tracestate != "" {
		ctx = context.WithValue(ctx, keyTracestate, tracestate)
	}
	return ctx
}

// WithRootScope 将根 Span 放入 context。仅供 Tracer 实现在 StartRequest 中调用，业务代码不要使用。
func WithRootScope(ctx context.Context, scope Scope) context.Context {
	return withRootScope(ctx, scope)
}

// WithTraceContext 写入待传播的 W3C 头值。仅供 Tracer 实现调用，业务代码不要使用。
func WithTraceContext(ctx context.Context, traceparent, tracestate string) context.Context {
	return withTraceparent(ctx, traceparent, tracestate)
}
