package tracing

import (
	"strings"

	appconfig "gateway/pkg/config"
)

// AppConfigKey 是进程级链路追踪在 app.yaml 中的配置节。
// 网关数据面、管理端请求及其它模块都读这一段。
const AppConfigKey = "app.tracing"

// Config 进程级链路追踪配置，对应 configs/app.yaml 的 app.tracing。
// 只对接 OTLP；SkyWalking、腾讯云 APM 等后端应接在 Collector 之后，不要把厂商地址写入本结构。
// 由 LoadFromApp / Normalize 填充默认值后再交给 Open。
type Config struct {
	// Enabled 总开关。false 时使用空操作实现，不改写任何追踪头。
	Enabled bool `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
	// ServiceName 出现在 APM 中的服务名。
	ServiceName string `json:"service_name" yaml:"service_name" mapstructure:"service_name"`
	// ServiceVersion 出现在 APM 中的服务版本。空则在 LoadFromApp 时回落到 app.version。
	ServiceVersion string `json:"service_version" yaml:"service_version" mapstructure:"service_version"`
	// Environment 出现在 APM 中的 deployment.environment，例如 prod / staging。
	Environment string `json:"environment" yaml:"environment" mapstructure:"environment"`
	// Endpoint Collector 地址，例如 otel-collector:4317 或 127.0.0.1:4318。
	// 为空且 Enabled 为 true 时只做 W3C 传播，不上报。
	Endpoint string `json:"endpoint" yaml:"endpoint" mapstructure:"endpoint"`
	// Protocol 上报协议：grpc 或 http，默认 grpc。
	Protocol string `json:"protocol" yaml:"protocol" mapstructure:"protocol"`
	// SamplingRate 新建 Trace 上报 APM 的比例，范围 (0, 1]。0.1 表示约 10% 的新请求会出 Span。
	// 不影响访问日志和内部 trace_id。客户端已带 sampled 的 traceparent 时跟随上游，不再按本值抽一次。
	// 填 0 或负数会被改成默认 0.1；要关闭追踪请用 Enabled=false。
	SamplingRate float64 `json:"sampling_rate" yaml:"sampling_rate" mapstructure:"sampling_rate"`
	// Insecure 使用明文连接 Collector。Sidecar / 内网 Collector 通常为 true。
	Insecure bool `json:"insecure" yaml:"insecure" mapstructure:"insecure"`
	// Headers 附加到 OTLP 导出请求的头，用于 Collector 接入认证，不要填写某一家 APM 的业务 Token。
	Headers map[string]string `json:"headers" yaml:"headers" mapstructure:"headers"`
}

// DefaultConfig 默认关闭追踪。
var DefaultConfig = Config{
	Enabled:      false,
	ServiceName:  "gateway",
	Protocol:     "grpc",
	SamplingRate: 0.1,
	Insecure:     true,
}

// LoadFromApp 从 app.yaml 的 app.tracing 读取进程级追踪配置。
// 配置节不存在时返回默认关闭配置。ServiceVersion 为空时回落 app.version。
func LoadFromApp() Config {
	cfg := DefaultConfig
	var fileCfg Config
	if err := appconfig.GetSection(AppConfigKey, &fileCfg); err == nil {
		cfg = mergeConfig(cfg, fileCfg)
	}
	if strings.TrimSpace(cfg.ServiceVersion) == "" {
		cfg.ServiceVersion = strings.TrimSpace(appconfig.GetString("app.version", ""))
	}
	return Normalize(cfg)
}

// Normalize 规范化配置：补默认服务名与协议，将采样率钳制到 (0, 1]，去掉 endpoint 上的 scheme。
// grpcs / https / grpc-tls 视为加密并将 Insecure 置为 false，同时把协议归一成 grpc 或 http。
func Normalize(cfg Config) Config {
	if strings.TrimSpace(cfg.ServiceName) == "" {
		cfg.ServiceName = DefaultConfig.ServiceName
	}
	protocol := strings.ToLower(strings.TrimSpace(cfg.Protocol))
	if protocol == "" {
		protocol = DefaultConfig.Protocol
	}
	cfg.Protocol = protocol
	if cfg.SamplingRate <= 0 {
		cfg.SamplingRate = DefaultConfig.SamplingRate
	}
	if cfg.SamplingRate > 1 {
		cfg.SamplingRate = 1
	}
	cfg.Endpoint = stripEndpoint(cfg.Endpoint)
	// 非 TLS 协议一律明文，避免调用方漏配 insecure 导致内网 Collector 握手失败
	if cfg.Protocol != "https" && cfg.Protocol != "grpcs" && cfg.Protocol != "grpc-tls" {
		cfg.Insecure = true
	} else {
		cfg.Insecure = false
		if cfg.Protocol == "https" {
			cfg.Protocol = "http"
		} else {
			cfg.Protocol = "grpc"
		}
	}
	return cfg
}

// mergeConfig 用文件配置覆盖默认值。字符串空则保留 base；Enabled / SamplingRate / Insecure 始终取 overlay。
func mergeConfig(base, overlay Config) Config {
	base.Enabled = overlay.Enabled
	if strings.TrimSpace(overlay.ServiceName) != "" {
		base.ServiceName = overlay.ServiceName
	}
	if strings.TrimSpace(overlay.ServiceVersion) != "" {
		base.ServiceVersion = overlay.ServiceVersion
	}
	if strings.TrimSpace(overlay.Environment) != "" {
		base.Environment = overlay.Environment
	}
	if strings.TrimSpace(overlay.Endpoint) != "" {
		base.Endpoint = overlay.Endpoint
	}
	if strings.TrimSpace(overlay.Protocol) != "" {
		base.Protocol = overlay.Protocol
	}
	base.SamplingRate = overlay.SamplingRate
	base.Insecure = overlay.Insecure
	if overlay.Headers != nil {
		base.Headers = overlay.Headers
	}
	return base
}

// stripEndpoint 去掉常见 scheme 前缀，供 OTLP SDK 的 WithEndpoint 使用。
func stripEndpoint(endpoint string) string {
	endpoint = strings.TrimSpace(endpoint)
	endpoint = strings.TrimPrefix(endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")
	endpoint = strings.TrimPrefix(endpoint, "grpc://")
	return strings.TrimSpace(endpoint)
}
