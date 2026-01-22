package publish

import (
	"context"
	"fmt"
	"time"

	clusterInit "gateway/internal/cluster/init"
	"gateway/internal/cluster/types"
	"gateway/pkg/logger"
)

// AlertConfigEventPublisher 告警渠道配置事件发布器
// 用于在 Controller 中发布告警渠道配置的注册/卸载/重载事件，通知集群其它节点同步更新
type AlertConfigEventPublisher struct{}

func NewAlertConfigEventPublisher() *AlertConfigEventPublisher {
	return &AlertConfigEventPublisher{}
}

const (
	alertConfigEventType = "ALERT_CONFIG"

	AlertConfigActionRegister   = "REGISTER"
	AlertConfigActionUnregister = "UNREGISTER"
	AlertConfigActionReload     = "RELOAD"
)

// PublishRegister 发布告警渠道注册事件
func (p *AlertConfigEventPublisher) PublishRegister(ctx context.Context, tenantId, channelName, operator string) error {
	return p.publish(ctx, AlertConfigActionRegister, tenantId, channelName, operator)
}

// PublishUnregister 发布告警渠道卸载事件
func (p *AlertConfigEventPublisher) PublishUnregister(ctx context.Context, tenantId, channelName, operator string) error {
	return p.publish(ctx, AlertConfigActionUnregister, tenantId, channelName, operator)
}

// PublishReload 发布告警渠道重载事件
func (p *AlertConfigEventPublisher) PublishReload(ctx context.Context, tenantId, channelName, operator string) error {
	return p.publish(ctx, AlertConfigActionReload, tenantId, channelName, operator)
}

func (p *AlertConfigEventPublisher) publish(ctx context.Context, action, tenantId, channelName, operator string) error {
	if channelName == "" {
		return fmt.Errorf("channelName不能为空")
	}

	// 检查集群服务是否已初始化/就绪
	if !clusterInit.IsClusterInitialized() || !clusterInit.IsClusterReady() {
		logger.Debug("集群服务未初始化或未就绪，跳过告警配置事件发布",
			"action", action,
			"tenantId", tenantId,
			"channelName", channelName,
		)
		return nil
	}

	clusterService := clusterInit.GetClusterService()
	if clusterService == nil {
		logger.Warn("无法获取集群服务，跳过告警配置事件发布",
			"action", action,
			"tenantId", tenantId,
			"channelName", channelName,
		)
		return nil
	}

	now := time.Now()
	expire := now.Add(10 * time.Minute) // 告警配置同步事件通常需要快速生效，过期后无需再处理

	event := &types.ClusterEvent{
		EventType:   alertConfigEventType,
		EventAction: action,
		ExpireTime:  &expire,
	}

	payload := AlertConfigEventPayload{
		TenantId:      tenantId,
		ChannelName:   channelName,
		Operator:      operator,
		RequestTimeMs: now.UnixMilli(),
	}
	if err := event.SetPayload(payload); err != nil {
		return fmt.Errorf("设置事件数据失败: %w", err)
	}

	if err := clusterService.PublishEvent(ctx, event); err != nil {
		return fmt.Errorf("发布集群事件失败: %w", err)
	}

	logger.Info("告警配置集群事件发布成功",
		"action", action,
		"tenantId", tenantId,
		"channelName", channelName,
		"eventId", event.EventId,
	)
	return nil
}

// AlertConfigEventPayload 告警配置事件数据
type AlertConfigEventPayload struct {
	TenantId      string `json:"tenantId"`      // 租户ID
	ChannelName   string `json:"channelName"`   // 渠道名称
	Operator      string `json:"operator"`      // 操作人（可选）
	RequestTimeMs int64  `json:"requestTimeMs"` // 触发时间（毫秒），用于节点侧辅助判断过期/排查
}
