package types

import (
	"context"
)

// ClusterService 集群服务接口
type ClusterService interface {
	// Start 启动集群服务
	Start(ctx context.Context) error

	// Stop 停止集群服务
	Stop(ctx context.Context) error

	// PublishEvent 发布事件
	PublishEvent(ctx context.Context, event *ClusterEvent) error

	// RegisterHandler 注册事件处理器
	// eventType 会自动从 handler.GetEventType() 获取
	RegisterHandler(handler EventHandler)

	// GetNodeId 获取当前节点ID
	GetNodeId() string

	// IsReady 检查服务是否就绪
	IsReady() bool
}

// EventHandler 事件处理器接口
type EventHandler interface {
	// Handle 处理事件
	// 返回 HandleResult 包含处理状态、消息等详细信息
	Handle(ctx context.Context, event *ClusterEvent) *HandleResult

	// GetEventType 获取处理器支持的事件类型
	// 用于注册时自动获取事件类型
	GetEventType() string
}

// HandleResult 事件处理结果
type HandleResult struct {
	// Status 处理状态
	// SUCCESS: 处理成功
	// FAILED: 处理失败（不会重试）
	// RETRY: 需要重试（会在下次轮询时重新处理）
	// SKIPPED: 跳过处理
	Status HandleStatus

	// Message 结果消息
	Message string

	// Error 错误信息（如果有）
	Error error

	// Data 扩展数据（可选，用于记录额外信息）
	Data map[string]interface{}
}

// HandleStatus 处理状态枚举
type HandleStatus string

const (
	// HandleStatusSuccess 处理成功
	HandleStatusSuccess HandleStatus = "SUCCESS"

	// HandleStatusFailed 处理失败（不会重试）
	HandleStatusFailed HandleStatus = "FAILED"

	// HandleStatusRetry 需要重试
	HandleStatusRetry HandleStatus = "RETRY"

	// HandleStatusSkipped 跳过处理
	HandleStatusSkipped HandleStatus = "SKIPPED"
)

// NewSuccessResult 创建成功结果
func NewSuccessResult(message string) *HandleResult {
	return &HandleResult{
		Status:  HandleStatusSuccess,
		Message: message,
	}
}

// NewFailedResult 创建失败结果
func NewFailedResult(err error, message string) *HandleResult {
	return &HandleResult{
		Status:  HandleStatusFailed,
		Message: message,
		Error:   err,
	}
}

// NewRetryResult 创建重试结果
func NewRetryResult(message string) *HandleResult {
	return &HandleResult{
		Status:  HandleStatusRetry,
		Message: message,
	}
}

// NewSkippedResult 创建跳过结果
func NewSkippedResult(message string) *HandleResult {
	return &HandleResult{
		Status:  HandleStatusSkipped,
		Message: message,
	}
}
