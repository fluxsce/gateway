package publish

import (
	"context"
	"fmt"

	clusterInit "gateway/internal/cluster/init"
	"gateway/internal/cluster/types"
	"gateway/pkg/logger"
)

// GatewayEventPublisher 网关事件发布器
// 用于在 Controller 中发布网关相关的集群事件
type GatewayEventPublisher struct{}

// NewGatewayEventPublisher 创建网关事件发布器
func NewGatewayEventPublisher() *GatewayEventPublisher {
	return &GatewayEventPublisher{}
}

// PublishStartEvent 发布网关启动事件
func (p *GatewayEventPublisher) PublishStartEvent(ctx context.Context, gatewayInstanceId, tenantId, instanceName, configFilePath, operator string) error {
	return p.publishEvent(ctx, "START", gatewayInstanceId, tenantId, instanceName, configFilePath, operator)
}

// PublishStopEvent 发布网关停止事件
func (p *GatewayEventPublisher) PublishStopEvent(ctx context.Context, gatewayInstanceId, tenantId, instanceName, operator string) error {
	return p.publishEvent(ctx, "STOP", gatewayInstanceId, tenantId, instanceName, "", operator)
}

// PublishReloadEvent 发布网关重载事件
func (p *GatewayEventPublisher) PublishReloadEvent(ctx context.Context, gatewayInstanceId, tenantId, instanceName, operator string) error {
	return p.publishEvent(ctx, "RELOAD", gatewayInstanceId, tenantId, instanceName, "", operator)
}

// PublishRestartEvent 发布网关重启事件
func (p *GatewayEventPublisher) PublishRestartEvent(ctx context.Context, gatewayInstanceId, tenantId, instanceName, configFilePath, operator string) error {
	return p.publishEvent(ctx, "RESTART", gatewayInstanceId, tenantId, instanceName, configFilePath, operator)
}

// publishEvent 发布网关事件的通用方法
func (p *GatewayEventPublisher) publishEvent(ctx context.Context, action, gatewayInstanceId, tenantId, instanceName, configFilePath, operator string) error {
	// 检查集群服务是否已初始化
	if !clusterInit.IsClusterInitialized() {
		logger.Debug("集群服务未初始化，跳过事件发布",
			"action", action,
			"gatewayInstanceId", gatewayInstanceId)
		return nil
	}

	// 检查集群服务是否就绪
	if !clusterInit.IsClusterReady() {
		logger.Debug("集群服务未就绪，跳过事件发布",
			"action", action,
			"gatewayInstanceId", gatewayInstanceId)
		return nil
	}

	// 获取集群服务
	clusterService := clusterInit.GetClusterService()
	if clusterService == nil {
		logger.Warn("无法获取集群服务，跳过事件发布",
			"action", action,
			"gatewayInstanceId", gatewayInstanceId)
		return nil
	}

	// 创建事件
	event := &types.ClusterEvent{
		EventType:   "GATEWAY_INSTANCE",
		EventAction: action,
	}

	// 设置事件数据
	payload := GatewayEventPayload{
		GatewayInstanceId: gatewayInstanceId,
		TenantId:          tenantId,
		InstanceName:      instanceName,
		ConfigFilePath:    configFilePath,
		Operator:          operator,
	}

	if err := event.SetPayload(payload); err != nil {
		return fmt.Errorf("设置事件数据失败: %w", err)
	}

	// 发布事件
	if err := clusterService.PublishEvent(ctx, event); err != nil {
		return fmt.Errorf("发布集群事件失败: %w", err)
	}

	logger.Info("网关集群事件发布成功",
		"action", action,
		"gatewayInstanceId", gatewayInstanceId,
		"tenantId", tenantId,
		"eventId", event.EventId)

	return nil
}

// GatewayEventPayload 网关事件数据
type GatewayEventPayload struct {
	GatewayInstanceId string `json:"gatewayInstanceId"` // 网关实例ID
	TenantId          string `json:"tenantId"`          // 租户ID
	InstanceName      string `json:"instanceName"`      // 实例名称（可选）
	ConfigFilePath    string `json:"configFilePath"`    // 配置文件路径（可选）
	Operator          string `json:"operator"`          // 操作人（可选）
}
