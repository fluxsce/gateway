// Package tracing 提供进程级、厂商无关的链路追踪。
//
// 目录分层（长期扩展按此加，不要把 go.opentelemetry.io 引回本包）：
//
//	pkg/tracing          接口、配置、W3C、noop、仅传播、Open/Close
//	pkg/tracing/otlp     唯一允许依赖 OpenTelemetry SDK 的 OTLP 导出
//
// 进程启动调用 Open(cfg, otlp.New)；网关、管理端及其它模块只依赖 Tracer。
// 换 APM 改外置 Collector，不要在本目录加 SkyWalking 或云厂商 SDK。
// 若以后增加第二种导出协议，新增子包并实现 Tracer，由 Open 的 Exporter 注入。
package tracing

import (
	"context"
	"net/http"
	"sync/atomic"
)

// Attr 是厂商无关的 Span 属性。
// 调用方只传字符串键值，不得依赖 OpenTelemetry 类型，以便业务代码与导出实现解耦。
type Attr struct {
	// Key 属性名，建议使用语义化名称，例如 http.route、gateway.trace_id。
	Key string
	// Value 属性值，一律按字符串传递。
	Value string
}

// String 构造字符串属性，供 StartRequest / StartSpan / SetAttr 使用。
func String(key, value string) Attr {
	return Attr{Key: key, Value: value}
}

// Scope 表示一个进行中的 Span。
// 由 StartRequest / StartSpan / StartClientSpan 返回，调用方在阶段结束时 End。
// 实现必须保证 End 可重复调用，且空接收者安全。
type Scope interface {
	// End 结束 Span。可重复调用，后续调用为空操作。
	End()
	// RecordError 将错误记录到 Span 上，并将状态标为失败。
	RecordError(err error)
	// SetAttr 追加属性。空 Key 应忽略。
	SetAttr(attr Attr)
}

// Tracer 进程级追踪接口。
// 网关数据面、管理端及其它模块只依赖本接口，不要断言 noopTracer / propagateTracer / otlp 内部类型。
// 进程启动经 Open 安装到 Global，请求路径通过 FromContext 或 Global 获取。
type Tracer interface {
	// StartRequest 从入站 HTTP 请求提取或新建 Trace，返回带追踪状态的 context 与根 Span。
	// r 为 nil 时仍创建根 Span。后续 gRPC/MQ 可走 StartSpan，不必改本方法。
	StartRequest(ctx context.Context, r *http.Request, spanName string, attrs ...Attr) (context.Context, Scope)
	// Inject 将当前 Trace 写入出站 HTTP 头（W3C traceparent / tracestate）。header 为 nil 时忽略。
	Inject(ctx context.Context, header http.Header)
	// StartSpan 创建内部阶段 Span（认证、限流等），父 Span 为当前 context 中的根或进行中 Span。
	StartSpan(ctx context.Context, name string, attrs ...Attr) (context.Context, Scope)
	// StartClientSpan 创建出站调用 Span。调用方应使用返回的 context 做 Inject，不要覆盖请求根 context。
	StartClientSpan(ctx context.Context, name string, attrs ...Attr) (context.Context, Scope)
	// Shutdown 刷新并关闭导出器。进程退出时由 Close 调用；请求路径不得调用。
	Shutdown(ctx context.Context) error
}

var globalTracer atomic.Value

// tracerHolder 包装 Tracer 接口再写入 atomic.Value。
// atomic.Value 要求存储的具体类型始终一致，直接存接口会在 noop 与 OTLP 切换时 panic。
type tracerHolder struct {
	// tracer 当前进程级实现，nil 视为未安装。
	tracer Tracer
}

// SetGlobal 设置进程级 Tracer。t 为 nil 时写入空操作实现。运行期应只由 Open / Close 调用。
func SetGlobal(t Tracer) {
	if t == nil {
		t = NewNoop()
	}
	globalTracer.Store(tracerHolder{tracer: t})
}

// Global 返回进程级 Tracer。尚未 Open 或已 Close 时返回空操作实现，保证调用方不必判空。
func Global() Tracer {
	value := globalTracer.Load()
	if value == nil {
		return NewNoop()
	}
	holder, ok := value.(tracerHolder)
	if !ok || holder.tracer == nil {
		return NewNoop()
	}
	return holder.tracer
}

// NewNoop 返回空操作追踪器，关闭追踪或测试时使用。不写任何追踪头、不上报。
func NewNoop() Tracer {
	return noopTracer{}
}

// NewPropagator 返回仅传播 W3C 上下文、不上报的追踪器。
// 用于 enabled 为 true 但未配置 endpoint 的场景。
func NewPropagator(cfg Config) Tracer {
	return &propagateTracer{samplingRate: Normalize(cfg).SamplingRate}
}
