package tracing

import (
	"context"
	"net/http"
	"strings"
)

// propagateTracer 仅识别、生成并注入 W3C Trace Context 的 Tracer 实现。
// 用于 enabled 为 true 但未配置 OTLP endpoint 的场景：上下游可串联，不上报 APM。
type propagateTracer struct {
	// samplingRate 新建根 Trace 时的头采样率，范围 [0, 1]。
	samplingRate float64
}

func (t *propagateTracer) StartRequest(ctx context.Context, r *http.Request, _ string, _ ...Attr) (context.Context, Scope) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx = WithTracer(ctx, t)
	if r == nil {
		return withRootScope(ctx, globalNoopScope), globalNoopScope
	}
	incoming := ParseTraceparent(r.Header.Get(HeaderTraceparent))
	var tp Traceparent
	tracestate := ""
	// 有合法入站头则续同一 TraceID；否则按采样率新建根 Trace
	if incoming.Valid() {
		tp = ContinueTraceparent(incoming)
		tracestate = strings.TrimSpace(r.Header.Get(HeaderTracestate))
	} else {
		tp = NewTraceparent(shouldSample(t.samplingRate))
	}
	ctx = withTraceparent(ctx, tp.String(), tracestate)
	ctx = withRootScope(ctx, globalNoopScope)
	return ctx, globalNoopScope
}

func (t *propagateTracer) Inject(ctx context.Context, header http.Header) {
	if header == nil {
		return
	}
	tp := TraceparentValue(ctx)
	if tp == "" {
		return
	}
	header.Set(HeaderTraceparent, tp)
	if ts := TracestateValue(ctx); ts != "" {
		header.Set(HeaderTracestate, ts)
	}
}

func (t *propagateTracer) StartSpan(ctx context.Context, _ string, _ ...Attr) (context.Context, Scope) {
	if ctx == nil {
		ctx = context.Background()
	}
	return ctx, globalNoopScope
}

func (t *propagateTracer) StartClientSpan(ctx context.Context, name string, attrs ...Attr) (context.Context, Scope) {
	return t.StartSpan(ctx, name, attrs...)
}

func (t *propagateTracer) Shutdown(context.Context) error { return nil }
