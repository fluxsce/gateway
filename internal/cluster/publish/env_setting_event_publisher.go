package publish

import (
	"context"
	"fmt"
	"time"

	clusterInit "gateway/internal/cluster/init"
	"gateway/internal/cluster/types"
	"gateway/pkg/logger"
)

const (
	envSettingEventType = "ENV_SETTING"

	// EnvSettingActionReload 通知各节点从库重载指定环境设置分组。
	EnvSettingActionReload = "RELOAD"
)

// EnvSettingEventPublisher 环境设置集群事件发布器。
// 保存 hub0009 分组后通知其它节点从 HUB_SYS_SETTING 重载，不把 JSON 当作权威数据。
type EnvSettingEventPublisher struct{}

// NewEnvSettingEventPublisher 创建环境设置事件发布器。
func NewEnvSettingEventPublisher() *EnvSettingEventPublisher {
	return &EnvSettingEventPublisher{}
}

// PublishReload 发布环境设置重载事件。集群未就绪时跳过，返回 nil。
func (p *EnvSettingEventPublisher) PublishReload(ctx context.Context, tenantId, groupCode, operator string) error {
	if tenantId == "" || groupCode == "" {
		return fmt.Errorf("tenantId和groupCode不能为空")
	}

	if !clusterInit.IsClusterInitialized() || !clusterInit.IsClusterReady() {
		logger.Debug("集群服务未初始化或未就绪，跳过环境设置事件发布",
			"tenantId", tenantId,
			"groupCode", groupCode,
		)
		return nil
	}

	clusterService := clusterInit.GetClusterService()
	if clusterService == nil {
		logger.Warn("无法获取集群服务，跳过环境设置事件发布",
			"tenantId", tenantId,
			"groupCode", groupCode,
		)
		return nil
	}

	now := time.Now()
	expire := now.Add(10 * time.Minute)

	event := &types.ClusterEvent{
		EventType:   envSettingEventType,
		EventAction: EnvSettingActionReload,
		ExpireTime:  &expire,
	}

	payload := EnvSettingEventPayload{
		TenantId:      tenantId,
		GroupCode:     groupCode,
		Operator:      operator,
		RequestTimeMs: now.UnixMilli(),
	}
	if err := event.SetPayload(payload); err != nil {
		return fmt.Errorf("设置事件数据失败: %w", err)
	}

	if err := clusterService.PublishEvent(ctx, event); err != nil {
		return fmt.Errorf("发布集群事件失败: %w", err)
	}

	logger.Info("环境设置集群事件发布成功",
		"tenantId", tenantId,
		"groupCode", groupCode,
		"eventId", event.EventId,
	)
	return nil
}

// EnvSettingEventPayload 环境设置事件数据。各节点按 tenantId+groupCode 从库重载。
type EnvSettingEventPayload struct {
	TenantId      string `json:"tenantId"`
	GroupCode     string `json:"groupCode"`
	Operator      string `json:"operator"`
	RequestTimeMs int64  `json:"requestTimeMs"`
}
