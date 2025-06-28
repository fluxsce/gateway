// Package interfaces 定义工具包的基础接口
// 提供所有工具必须实现的基础接口规范
package interfaces

import (
	"context"
)

// Tool 工具基础接口
// 定义所有工具必须实现的基础方法
type Tool interface {
	// GetID 获取工具唯一标识
	GetID() string
	
	// GetType 获取工具类型
	GetType() string
	
	// Close 关闭工具并释放资源
	Close() error
	
	// IsActive 检查工具是否处于活跃状态
	IsActive() bool
	
	// Connect 连接到服务器
	Connect(ctx context.Context) error
}

// ConnectableTool 可连接工具接口
// 为需要网络连接的工具提供连接状态管理
type ConnectableTool interface {
	Tool
	
	// IsConnected 检查连接状态
	IsConnected() bool
	
	// Reconnect 重新连接
	Reconnect(ctx context.Context) error
	
	// Disconnect 断开连接
	Disconnect() error
}

// ConfigurableTool 可配置工具接口
// 为需要配置管理的工具提供配置访问
type ConfigurableTool interface {
	Tool
	
	// GetConfig 获取工具配置
	GetConfig() interface{}
	
	// UpdateConfig 更新工具配置
	UpdateConfig(config interface{}) error
	
	// ValidateConfig 验证配置有效性
	ValidateConfig(config interface{}) error
} 