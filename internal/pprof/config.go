package pprof

import (
	"time"
)

// Config pprof配置
type Config struct {
	// 是否启用pprof
	Enabled bool `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
	// 监听地址
	Listen string `json:"listen" yaml:"listen" mapstructure:"listen"`
	// 服务名称
	ServiceName string `json:"service_name" yaml:"service_name" mapstructure:"service_name"`
	// 读取超时
	ReadTimeout time.Duration `json:"read_timeout" yaml:"read_timeout" mapstructure:"read_timeout"`
	// 写入超时
	WriteTimeout time.Duration `json:"write_timeout" yaml:"write_timeout" mapstructure:"write_timeout"`
	// 是否启用认证
	EnableAuth bool `json:"enable_auth" yaml:"enable_auth" mapstructure:"enable_auth"`
	// 认证token
	AuthToken string `json:"auth_token" yaml:"auth_token" mapstructure:"auth_token"`
	// 自动分析配置
	AutoAnalysis AutoAnalysisConfig `json:"auto_analysis" yaml:"auto_analysis" mapstructure:"auto_analysis"`
}

// AutoAnalysisConfig 自动分析配置
type AutoAnalysisConfig struct {
	// 是否启用自动分析
	Enabled bool `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
	// 分析间隔
	Interval time.Duration `json:"interval" yaml:"interval" mapstructure:"interval"`
	// CPU采样时间
	CPUSampleDuration time.Duration `json:"cpu_sample_duration" yaml:"cpu_sample_duration" mapstructure:"cpu_sample_duration"`
	// 输出目录
	OutputDir string `json:"output_dir" yaml:"output_dir" mapstructure:"output_dir"`
	// 是否保存历史数据
	SaveHistory bool `json:"save_history" yaml:"save_history" mapstructure:"save_history"`
	// 历史数据保留天数
	HistoryRetentionDays int `json:"history_retention_days" yaml:"history_retention_days" mapstructure:"history_retention_days"`
}

// DefaultConfig 默认配置
var DefaultConfig = Config{
	Enabled:      false,
	Listen:       ":6060",
	ServiceName:  "Gateway-pprof",
	ReadTimeout:  30 * time.Second,
	WriteTimeout: 30 * time.Second,
	EnableAuth:   false,
	AuthToken:    "",
	AutoAnalysis: AutoAnalysisConfig{
		Enabled:              false,
		Interval:             30 * time.Minute,
		CPUSampleDuration:    30 * time.Second,
		OutputDir:            "./pprof_analysis",
		SaveHistory:          true,
		HistoryRetentionDays: 7,
	},
}
