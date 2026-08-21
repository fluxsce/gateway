package tracing

import (
	"context"
	"fmt"
	"strings"
)

// Exporter 根据配置创建可上报的 Tracer。
// 进程启动时传入 otlp.New。业务代码不要使用本类型，本包也不要 import otlp，否则循环依赖。
type Exporter func(Config) (Tracer, error)

// Build 按配置选择实现，不改 Global。
// enabled=false 返回 noop；开启但 endpoint 为空返回仅传播实现；配了 endpoint 则调用 exporter。
func Build(cfg Config, exporter Exporter) (Tracer, error) {
	cfg = Normalize(cfg)
	if !cfg.Enabled {
		return NewNoop(), nil
	}
	if strings.TrimSpace(cfg.Endpoint) == "" {
		return NewPropagator(cfg), nil
	}
	if exporter == nil {
		return nil, fmt.Errorf("已配置 OTLP endpoint 但未提供导出器，进程启动应传入 otlp.New")
	}
	return exporter(cfg)
}

// Open 创建进程级 Tracer 并设为 Global。Build 失败时不替换已有 Global。
// 每个进程应只在启动时调用一次，由 cmd/init 负责。
func Open(cfg Config, exporter Exporter) (Tracer, error) {
	tracer, err := Build(cfg, exporter)
	if err != nil {
		return nil, err
	}
	SetGlobal(tracer)
	return tracer, nil
}

// OpenFromApp 读取 app.tracing 后调用 Open，供进程启动使用。
func OpenFromApp(exporter Exporter) (Tracer, error) {
	return Open(LoadFromApp(), exporter)
}

// Close 刷新并关闭当前 Global Tracer，然后还原为空操作，避免退出后仍有模块向已关闭的导出器写 Span。
func Close(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	tracer := Global()
	SetGlobal(NewNoop())
	return tracer.Shutdown(ctx)
}
