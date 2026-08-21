package init

import (
	"context"
	"time"

	"gateway/pkg/logger"
	"gateway/pkg/tracing"
	"gateway/pkg/tracing/otlp"
)

// InitTracing 在进程启动时创建全局 Tracer，供网关数据面、管理端及其它模块共用。
// 必须在网关与 Web 启动之前调用。endpoint 非法时返回错误，避免带着半残追踪启动。
func InitTracing() error {
	cfg := tracing.LoadFromApp()
	if _, err := tracing.Open(cfg, otlp.New); err != nil {
		return err
	}
	if !cfg.Enabled {
		logger.Info("链路追踪未启用")
		return nil
	}
	if cfg.Endpoint == "" {
		logger.Info("链路追踪已启用，仅传播 W3C traceparent，不上报 OTLP")
		return nil
	}
	logger.Info("链路追踪已启用",
		"service", cfg.ServiceName,
		"version", cfg.ServiceVersion,
		"environment", cfg.Environment,
		"endpoint", cfg.Endpoint,
		"protocol", cfg.Protocol,
		"sampling_rate", cfg.SamplingRate)
	return nil
}

// StopTracing 刷新并关闭进程级 Tracer。应在网关与 Web 停止之后调用。
func StopTracing() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := tracing.Close(ctx); err != nil {
		logger.Warn("关闭链路追踪失败", "error", err)
		return err
	}
	return nil
}
