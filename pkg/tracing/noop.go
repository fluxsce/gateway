package tracing

import (
	"context"
	"net/http"
)

// noopTracer 关闭追踪时的空操作 Tracer 实现。
// 不写追踪头、不上报、不保存 Span 状态，保证 enabled=false 时热路径行为与接入前一致。
// 由 NewNoop / Open(enabled=false) 使用，调用方只应通过 Tracer 接口持有。
type noopTracer struct{}

// noopScope 空操作 Span 实现。
// 所有方法立即返回，供关闭追踪、仅传播模式以及缺失根 Span 时复用，避免调用方判空。
type noopScope struct{}

var globalNoopScope Scope = noopScope{}

func (noopTracer) StartRequest(ctx context.Context, _ *http.Request, _ string, _ ...Attr) (context.Context, Scope) {
	if ctx == nil {
		ctx = context.Background()
	}
	return ctx, globalNoopScope
}

func (noopTracer) Inject(context.Context, http.Header) {}

func (t noopTracer) StartSpan(ctx context.Context, _ string, _ ...Attr) (context.Context, Scope) {
	if ctx == nil {
		ctx = context.Background()
	}
	return ctx, globalNoopScope
}

func (t noopTracer) StartClientSpan(ctx context.Context, name string, attrs ...Attr) (context.Context, Scope) {
	return t.StartSpan(ctx, name, attrs...)
}

func (noopTracer) Shutdown(context.Context) error { return nil }

func (noopScope) End() {}

func (noopScope) RecordError(error) {}

func (noopScope) SetAttr(Attr) {}
