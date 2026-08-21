// Package otlp 使用 OpenTelemetry SDK 将 Span 通过 OTLP 导出到 Collector。
// 本包是 tracing 目录下唯一允许 import go.opentelemetry.io 的实现；业务代码不要依赖本包类型。
package otlp

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"

	"gateway/pkg/tracing"
)

// otelExporter 是 tracing.Tracer 的 OTLP 实现。
// 持有 OpenTelemetry TracerProvider，通过 OTLP 把 Span 导出到 Collector。
// 不导出本类型：进程内只通过 tracing.Open(..., otlp.New) 安装，请求热路径不要 New。
type otelExporter struct {
	// provider 管理采样、批量导出与资源属性，进程退出时 Shutdown。
	provider *sdktrace.TracerProvider
	// otel 是 OpenTelemetry 的 span 创建器，不要与 tracing.Tracer 接口混淆。
	otel oteltrace.Tracer
	// prop 仅处理 W3C TraceContext 与 Baggage，从入站头提取、向出站头注入。
	prop propagation.TextMapPropagator
}

// requestState 存放在 context 中的请求根 Span。
// 供后续 StartSpan / StartClientSpan 把子 Span 挂到同一根上，而不依赖调用方是否改写了当前 Span。
type requestState struct {
	// span 本次请求的服务端根 Span。
	span oteltrace.Span
	// ctx 根 Span 对应的 OTel context，用于传播提取。
	ctx context.Context
}

// New 创建 OTLP 追踪器。cfg.Endpoint 必填；采样、Resource、批量导出在此一次性配好。
// 返回 tracing.Tracer 接口，调用方不要断言为本包内部类型。
func New(cfg tracing.Config) (tracing.Tracer, error) {
	cfg = tracing.Normalize(cfg)
	if strings.TrimSpace(cfg.Endpoint) == "" {
		return nil, fmt.Errorf("OTLP endpoint 不能为空")
	}

	exporter, err := newExporter(cfg)
	if err != nil {
		return nil, err
	}

	res, err := newResource(cfg)
	if err != nil {
		return nil, err
	}

	sampler := sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.SamplingRate))
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sampler),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	return &otelExporter{
		provider: provider,
		otel:     provider.Tracer("gateway/pkg/tracing"),
		prop:     propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}),
	}, nil
}

// newResource 组装 APM Resource：服务名/版本/环境，并合并 OTEL_* 环境变量与主机信息。
func newResource(cfg tracing.Config) (*resource.Resource, error) {
	attrs := []attribute.KeyValue{
		attribute.String("service.name", cfg.ServiceName),
	}
	if strings.TrimSpace(cfg.ServiceVersion) != "" {
		attrs = append(attrs, attribute.String("service.version", cfg.ServiceVersion))
	}
	if strings.TrimSpace(cfg.Environment) != "" {
		attrs = append(attrs, attribute.String("deployment.environment", cfg.Environment))
	}
	res, err := resource.New(context.Background(),
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithAttributes(attrs...),
	)
	if err != nil {
		return nil, fmt.Errorf("创建 OTLP resource 失败: %w", err)
	}
	return res, nil
}

// newExporter 按 protocol 创建 OTLP gRPC 或 HTTP 导出器。默认 gRPC。
func newExporter(cfg tracing.Config) (*otlptrace.Exporter, error) {
	ctx := context.Background()
	switch cfg.Protocol {
	case "http", "http/protobuf", "http/protobuf+json":
		opts := []otlptracehttp.Option{
			otlptracehttp.WithEndpoint(cfg.Endpoint),
		}
		if cfg.Insecure {
			opts = append(opts, otlptracehttp.WithInsecure())
		}
		if len(cfg.Headers) > 0 {
			opts = append(opts, otlptracehttp.WithHeaders(cfg.Headers))
		}
		return otlptracehttp.New(ctx, opts...)
	default:
		opts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(cfg.Endpoint),
		}
		if cfg.Insecure {
			opts = append(opts, otlptracegrpc.WithInsecure())
		}
		if len(cfg.Headers) > 0 {
			opts = append(opts, otlptracegrpc.WithHeaders(cfg.Headers))
		}
		return otlptracegrpc.New(ctx, opts...)
	}
}

// StartRequest 提取入站 W3C 上下文并创建根 Span。
func (e *otelExporter) StartRequest(ctx context.Context, r *http.Request, spanName string, attrs ...tracing.Attr) (context.Context, tracing.Scope) {
	if e == nil {
		return tracing.NewNoop().StartRequest(ctx, r, spanName, attrs...)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	ctx = tracing.WithTracer(ctx, e)
	parent := ctx
	if r != nil {
		parent = e.prop.Extract(ctx, propagation.HeaderCarrier(r.Header))
		if r.Method != "" {
			attrs = append([]tracing.Attr{tracing.String("http.request.method", r.Method)}, attrs...)
		}
		if r.URL != nil && r.URL.Path != "" {
			attrs = append(attrs, tracing.String("url.path", r.URL.Path))
		}
	}
	if strings.TrimSpace(spanName) == "" {
		spanName = "request"
	}
	spanCtx, span := e.otel.Start(parent, spanName,
		oteltrace.WithSpanKind(oteltrace.SpanKindServer),
		oteltrace.WithAttributes(toAttributes(attrs)...),
	)
	scope := &spanScope{span: span}
	ctx = withRequestState(spanCtx, &requestState{span: span, ctx: spanCtx})
	// 用独立 Header 取出即将传播的 W3C 值，写入 context 供响应头与仅传播回退使用
	header := make(http.Header)
	e.prop.Inject(spanCtx, propagation.HeaderCarrier(header))
	ctx = tracing.WithTraceContext(ctx, header.Get(tracing.HeaderTraceparent), header.Get(tracing.HeaderTracestate))
	ctx = tracing.WithRootScope(ctx, scope)
	return ctx, scope
}

// Inject 将当前 Span 上下文写入出站请求头。
func (e *otelExporter) Inject(ctx context.Context, header http.Header) {
	if e == nil || header == nil || ctx == nil {
		return
	}
	if oteltrace.SpanFromContext(ctx).SpanContext().IsValid() {
		e.prop.Inject(ctx, propagation.HeaderCarrier(header))
		return
	}
	// 无有效 OTel Span 时回退到 StartRequest 写入的 W3C 字符串（仅传播模式不会走到本实现）
	if tp := tracing.TraceparentValue(ctx); tp != "" {
		header.Set(tracing.HeaderTraceparent, tp)
		if ts := tracing.TracestateValue(ctx); ts != "" {
			header.Set(tracing.HeaderTracestate, ts)
		}
	}
}

// StartSpan 创建内部阶段 Span。
func (e *otelExporter) StartSpan(ctx context.Context, name string, attrs ...tracing.Attr) (context.Context, tracing.Scope) {
	return e.start(ctx, name, oteltrace.SpanKindInternal, attrs...)
}

// StartClientSpan 创建出站调用 Span。
func (e *otelExporter) StartClientSpan(ctx context.Context, name string, attrs ...tracing.Attr) (context.Context, tracing.Scope) {
	return e.start(ctx, name, oteltrace.SpanKindClient, attrs...)
}

// start 在请求根 Span 下创建子 Span。用根 Span 作父级，避免挂到已结束的阶段 Span 上。
func (e *otelExporter) start(ctx context.Context, name string, kind oteltrace.SpanKind, attrs ...tracing.Attr) (context.Context, tracing.Scope) {
	if e == nil {
		return tracing.NewNoop().StartSpan(ctx, name, attrs...)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	parent := ctx
	if state, ok := runtimeState(ctx); ok {
		// 父级固定为请求根 Span，多服务并行出站互不嵌套
		parent = oteltrace.ContextWithSpan(ctx, state.span)
	}
	spanCtx, span := e.otel.Start(parent, name,
		oteltrace.WithSpanKind(kind),
		oteltrace.WithAttributes(toAttributes(attrs)...),
	)
	return spanCtx, &spanScope{span: span}
}

// Shutdown 刷新并关闭 TracerProvider。应在网关流量排空之后、进程退出时调用。
func (e *otelExporter) Shutdown(ctx context.Context) error {
	if e == nil || e.provider == nil {
		return nil
	}
	return e.provider.Shutdown(ctx)
}

// spanScope 将 otel Span 适配为 tracing.Scope。
// 业务侧只看到 Scope 接口；once 保证 End 只执行一次，避免重复结束。
type spanScope struct {
	// span 底层 OpenTelemetry Span。
	span oteltrace.Span
	// once 串行化 End，并发或重复调用时后续为空操作。
	once sync.Once
}

func (s *spanScope) End() {
	if s == nil || s.span == nil {
		return
	}
	s.once.Do(func() {
		s.span.End()
	})
}

func (s *spanScope) RecordError(err error) {
	if s == nil || s.span == nil || err == nil {
		return
	}
	s.span.RecordError(err)
	s.span.SetStatus(codes.Error, err.Error())
}

func (s *spanScope) SetAttr(attr tracing.Attr) {
	if s == nil || s.span == nil || attr.Key == "" {
		return
	}
	s.span.SetAttributes(attribute.String(attr.Key, attr.Value))
}

func runtimeState(ctx context.Context) (*requestState, bool) {
	if ctx == nil {
		return nil, false
	}
	state, ok := ctx.Value(runtimeKey{}).(*requestState)
	if !ok || state == nil || state.span == nil {
		return nil, false
	}
	return state, true
}

// runtimeKey 是 OTLP 实现私有的 context.Value 键。
// 不暴露到 tracing 公共 API，避免业务代码依赖本包内部状态。
type runtimeKey struct{}

func withRequestState(ctx context.Context, state *requestState) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, runtimeKey{}, state)
}

// toAttributes 将本包 Attr 转为 OTel 属性，跳过空 Key。
func toAttributes(attrs []tracing.Attr) []attribute.KeyValue {
	if len(attrs) == 0 {
		return nil
	}
	out := make([]attribute.KeyValue, 0, len(attrs))
	for _, attr := range attrs {
		if attr.Key == "" {
			continue
		}
		out = append(out, attribute.String(attr.Key, attr.Value))
	}
	return out
}
