package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"gateway/internal/cluster/types"
	"gateway/internal/servicecenterv3"
	"gateway/internal/servicecenterv3/contract"
	"gateway/internal/servicecenterv3/model"
	"gateway/pkg/database"
	"gateway/pkg/logger"
)

// ServiceCenterEventHandler 在本节点执行服务中心集群事件。
// 发布节点已被 GetPendingEvents 排除（sourceNodeId），本机已在来源处执行过。
type ServiceCenterEventHandler struct {
	db database.Database
}

// NewServiceCenterEventHandler 构造处理器。db 仅占位，与其它 handler 签名一致。
func NewServiceCenterEventHandler(db database.Database) *ServiceCenterEventHandler {
	return &ServiceCenterEventHandler{db: db}
}

// GetEventType 返回 SERVICE_CENTER_INSTANCE。
func (h *ServiceCenterEventHandler) GetEventType() string {
	return "SERVICE_CENTER_INSTANCE"
}

// Handle 按 START/STOP/RELOAD/UNLOAD 驱动本进程 Pool。
func (h *ServiceCenterEventHandler) Handle(ctx context.Context, event *types.ClusterEvent) *types.HandleResult {
	logger.Info("处理服务中心集群事件",
		"eventId", event.EventId,
		"eventAction", event.EventAction,
		"eventType", event.EventType)

	if event.IsExpired() {
		return types.NewSkippedResult("事件已过期，跳过处理")
	}

	var payload serviceCenterEventPayload
	if err := json.Unmarshal([]byte(event.EventPayload), &payload); err != nil {
		return types.NewFailedResult(err, fmt.Sprintf("解析事件数据失败: %v", err))
	}
	if payload.InstanceName == "" {
		return types.NewFailedResult(nil, "instanceName不能为空")
	}

	switch event.EventAction {
	case "START":
		return h.handleStart(ctx, &payload)
	case "STOP":
		return h.handleStop(ctx, &payload)
	case "RELOAD":
		return h.handleReload(ctx, &payload)
	case "UNLOAD":
		return h.handleUnload(ctx, &payload)
	case "NAMING":
		return h.handleNaming(&payload)
	case "CONFIG":
		return h.handleConfig(&payload)
	case "NAMESPACE":
		return h.handleNamespace(&payload)
	default:
		return types.NewSkippedResult(fmt.Sprintf("未知的事件动作: %s", event.EventAction))
	}
}

func (h *ServiceCenterEventHandler) handleStart(ctx context.Context, p *serviceCenterEventPayload) *types.HandleResult {
	if pool := servicecenterv3.GetPool(); pool != nil {
		if inst, ok := pool.FindInstance(p.InstanceName, p.Environment); ok && inst.IsRunning() {
			return types.NewSkippedResult("服务中心实例已在运行中")
		}
		if err := pool.Start(ctx, clusterCC(p)); err != nil {
			return types.NewFailedResult(err, fmt.Sprintf("启动服务中心失败: %v", err))
		}
		return types.NewSuccessResult("服务中心实例启动成功")
	}
	return types.NewSkippedResult("服务中心引擎未初始化")
}

func (h *ServiceCenterEventHandler) handleStop(ctx context.Context, p *serviceCenterEventPayload) *types.HandleResult {
	if pool := servicecenterv3.GetPool(); pool != nil {
		err := pool.Stop(ctx, clusterCC(p))
		if err != nil && !errors.Is(err, contract.ErrCenterNotFound) {
			return types.NewFailedResult(err, fmt.Sprintf("停止服务中心失败: %v", err))
		}
		if errors.Is(err, contract.ErrCenterNotFound) {
			return types.NewSkippedResult("服务中心实例未加载")
		}
		return types.NewSuccessResult("服务中心实例停止成功")
	}
	return types.NewSkippedResult("服务中心引擎未初始化")
}

func (h *ServiceCenterEventHandler) handleReload(ctx context.Context, p *serviceCenterEventPayload) *types.HandleResult {
	if pool := servicecenterv3.GetPool(); pool != nil {
		if _, ok := pool.FindInstance(p.InstanceName, p.Environment); !ok {
			return types.NewSkippedResult("服务中心实例未加载，无法重载")
		}
		if err := pool.Reload(ctx, clusterCC(p)); err != nil {
			return types.NewFailedResult(err, fmt.Sprintf("重载服务中心失败: %v", err))
		}
		return types.NewSuccessResult("服务中心实例重载成功")
	}
	return types.NewSkippedResult("服务中心引擎未初始化")
}

func (h *ServiceCenterEventHandler) handleUnload(ctx context.Context, p *serviceCenterEventPayload) *types.HandleResult {
	if pool := servicecenterv3.GetPool(); pool != nil {
		pool.Unload(ctx, p.InstanceName, p.Environment)
		return types.NewSuccessResult("服务中心实例已卸载")
	}
	return types.NewSkippedResult("服务中心引擎未初始化")
}

func (h *ServiceCenterEventHandler) handleNaming(p *serviceCenterEventPayload) *types.HandleResult {
	if p.Naming == nil {
		return types.NewSkippedResult("空命名事件")
	}
	if p.Naming.Environment == "" {
		p.Naming.Environment = p.Environment
	}
	if p.Naming.CenterInstanceName == "" {
		p.Naming.CenterInstanceName = p.InstanceName
	}
	servicecenterv3.ApplyRemoteNaming(*p.Naming)
	return types.NewSuccessResult("已应用服务中心命名变更")
}

func (h *ServiceCenterEventHandler) handleNamespace(p *serviceCenterEventPayload) *types.HandleResult {
	if p.Namespace == nil {
		return types.NewSkippedResult("空命名空间事件")
	}
	if p.Namespace.Environment == "" {
		p.Namespace.Environment = p.Environment
	}
	if p.Namespace.CenterInstanceName == "" {
		p.Namespace.CenterInstanceName = p.InstanceName
	}
	servicecenterv3.ApplyRemoteNamespace(*p.Namespace)
	return types.NewSuccessResult("已应用服务中心命名空间变更")
}

func (h *ServiceCenterEventHandler) handleConfig(p *serviceCenterEventPayload) *types.HandleResult {
	if p.Config == nil {
		return types.NewSkippedResult("空配置事件")
	}
	if p.Config.Environment == "" {
		p.Config.Environment = p.Environment
	}
	if p.Config.CenterInstanceName == "" {
		p.Config.CenterInstanceName = p.InstanceName
	}
	servicecenterv3.ApplyRemoteConfig(*p.Config)
	return types.NewSuccessResult("已应用服务中心配置变更")
}

func clusterCC(p *serviceCenterEventPayload) contract.CallContext {
	return contract.CallContext{
		TenantID:           p.TenantId,
		CenterInstanceName: p.InstanceName,
		Environment:        p.Environment,
		OperatorID:         p.Operator,
	}
}

type serviceCenterEventPayload struct {
	TenantId      string                `json:"tenantId"`
	InstanceName  string                `json:"instanceName"`
	Environment   string                `json:"environment"`
	Operator      string                `json:"operator"`
	RequestTimeMs int64                 `json:"requestTimeMs"`
	Naming        *model.NamingEvent    `json:"naming,omitempty"`
	Config        *model.ConfigEvent    `json:"config,omitempty"`
	Namespace     *model.NamespaceEvent `json:"namespace,omitempty"`
}
