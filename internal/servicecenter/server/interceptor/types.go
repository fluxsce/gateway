package interceptor

import "gateway/internal/servicecenter/types"

// ConfigProvider 配置提供者接口
// 用于拦截器获取最新的实例配置
type ConfigProvider interface {
	// GetConfig 获取当前实例配置
	GetConfig() *types.InstanceConfig
}
