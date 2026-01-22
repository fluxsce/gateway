package handler

import (
	"context"
	"encoding/json"
	"fmt"

	alertInit "gateway/internal/alert/init"
	"gateway/internal/cluster/types"
	"gateway/pkg/database"
	"gateway/pkg/logger"
)

// AlertConfigEventHandler 告警渠道配置事件处理器
// 用于在各节点同步执行告警渠道配置的注册/卸载/重载
type AlertConfigEventHandler struct {
	db database.Database
}

func NewAlertConfigEventHandler(db database.Database) *AlertConfigEventHandler {
	return &AlertConfigEventHandler{db: db}
}

func (h *AlertConfigEventHandler) GetEventType() string {
	return "ALERT_CONFIG"
}

func (h *AlertConfigEventHandler) Handle(ctx context.Context, event *types.ClusterEvent) *types.HandleResult {
	logger.Info("处理告警配置集群事件",
		"eventId", event.EventId,
		"eventAction", event.EventAction,
		"eventType", event.EventType,
	)

	// 事件过期则跳过（集群服务可能因轮询延迟导致积压）
	if event.IsExpired() {
		return types.NewSkippedResult("事件已过期，跳过处理")
	}

	var payload alertConfigEventPayload
	if err := json.Unmarshal([]byte(event.EventPayload), &payload); err != nil {
		return types.NewFailedResult(err, fmt.Sprintf("解析事件数据失败: %v", err))
	}

	if payload.ChannelName == "" {
		return types.NewFailedResult(nil, "channelName不能为空")
	}

	helper := alertInit.NewAlertConfigHelper(h.db)

	switch event.EventAction {
	case "REGISTER":
		if payload.TenantId == "" {
			return types.NewFailedResult(nil, "tenantId不能为空")
		}
		if err := helper.RegisterChannel(ctx, payload.TenantId, payload.ChannelName); err != nil {
			return types.NewFailedResult(err, fmt.Sprintf("注册告警渠道失败: %v", err))
		}
		return types.NewSuccessResult("注册告警渠道成功")

	case "UNREGISTER":
		if err := helper.UnregisterChannel(ctx, payload.ChannelName); err != nil {
			return types.NewFailedResult(err, fmt.Sprintf("注销告警渠道失败: %v", err))
		}
		return types.NewSuccessResult("注销告警渠道成功")

	case "RELOAD":
		if payload.TenantId == "" {
			return types.NewFailedResult(nil, "tenantId不能为空")
		}
		if err := helper.ReloadChannel(ctx, payload.TenantId, payload.ChannelName); err != nil {
			return types.NewFailedResult(err, fmt.Sprintf("重载告警渠道失败: %v", err))
		}
		return types.NewSuccessResult("重载告警渠道成功")

	default:
		return types.NewSkippedResult(fmt.Sprintf("未知的事件动作: %s", event.EventAction))
	}
}

type alertConfigEventPayload struct {
	TenantId      string `json:"tenantId"`
	ChannelName   string `json:"channelName"`
	Operator      string `json:"operator"`
	RequestTimeMs int64  `json:"requestTimeMs"`
}
