package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"gateway/internal/cluster/types"
	"gateway/internal/retention"
	"gateway/pkg/database"
	"gateway/pkg/logger"
)

// EnvSettingEventHandler 环境设置集群事件处理器。
// 收到重载事件后从 HUB_SYS_SETTING 读取该分组写入本进程缓存。
type EnvSettingEventHandler struct {
	db database.Database
}

// NewEnvSettingEventHandler 创建环境设置事件处理器。
func NewEnvSettingEventHandler(db database.Database) *EnvSettingEventHandler {
	return &EnvSettingEventHandler{db: db}
}

// GetEventType 返回处理器订阅的事件类型。
func (h *EnvSettingEventHandler) GetEventType() string {
	return "ENV_SETTING"
}

// Handle 处理环境设置重载。发布节点已被查询排除，本机保存时已写入缓存。
func (h *EnvSettingEventHandler) Handle(ctx context.Context, event *types.ClusterEvent) *types.HandleResult {
	logger.Info("处理环境设置集群事件",
		"eventId", event.EventId,
		"eventAction", event.EventAction,
		"eventType", event.EventType,
	)

	if event.IsExpired() {
		return types.NewSkippedResult("事件已过期，跳过处理")
	}

	var payload envSettingEventPayload
	if err := json.Unmarshal([]byte(event.EventPayload), &payload); err != nil {
		return types.NewFailedResult(err, fmt.Sprintf("解析事件数据失败: %v", err))
	}
	if payload.TenantId == "" || payload.GroupCode == "" {
		return types.NewFailedResult(nil, "tenantId和groupCode不能为空")
	}

	switch event.EventAction {
	case "RELOAD":
		if err := retention.ReloadGroup(ctx, h.db, payload.TenantId, payload.GroupCode); err != nil {
			return types.NewFailedResult(err, fmt.Sprintf("重载环境设置失败: %v", err))
		}
		return types.NewSuccessResult("重载环境设置成功")
	default:
		return types.NewSkippedResult(fmt.Sprintf("未知的事件动作: %s", event.EventAction))
	}
}

type envSettingEventPayload struct {
	TenantId      string `json:"tenantId"`
	GroupCode     string `json:"groupCode"`
	Operator      string `json:"operator"`
	RequestTimeMs int64  `json:"requestTimeMs"`
}
