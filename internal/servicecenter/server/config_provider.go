package server

import "gateway/internal/servicecenter/types"

// ConfigProvider 配置提供者接口
// 用于 Handler 层访问实例配置（支持告警等功能）
type ConfigProvider interface {
	// GetConfig 获取当前实例配置
	GetConfig() *types.InstanceConfig
}
