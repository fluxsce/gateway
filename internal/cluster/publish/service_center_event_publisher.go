package publish

import (
	"context"
	"fmt"
	"time"

	"gateway/internal/cluster/types"
	"gateway/internal/servicecenterv3/model"
	"gateway/pkg/logger"
)

// ServiceCenterEventType 服务中心集群事件。
// 发布节点本地已执行，查询会排除 sourceNodeId。
// 生命周期：其它网关据此 Start/Stop/Reload/Unload。
// 命名/配置：其它网关 ApplyRemote（心跳不发）。
const ServiceCenterEventType = "SERVICE_CENTER_INSTANCE"

const (
	ServiceCenterActionStart     = "START"
	ServiceCenterActionStop      = "STOP"
	ServiceCenterActionReload    = "RELOAD"
	ServiceCenterActionUnload    = "UNLOAD"
	ServiceCenterActionNaming    = "NAMING"
	ServiceCenterActionConfig    = "CONFIG"
	ServiceCenterActionNamespace = "NAMESPACE"
)

// ServiceCenterEventPublisher 把中台对中心实例的变更广播到集群其它节点。
type ServiceCenterEventPublisher struct{}

// NewServiceCenterEventPublisher 构造发布器。集群未就绪时发布为无操作。
func NewServiceCenterEventPublisher() *ServiceCenterEventPublisher {
	return &ServiceCenterEventPublisher{}
}

// PublishStart 其它节点按库中的定义打开数据面监听。
func (p *ServiceCenterEventPublisher) PublishStart(ctx context.Context, tenantID, instanceName, environment, operator string) error {
	return p.publish(ctx, ServiceCenterActionStart, tenantID, instanceName, environment, operator)
}

// PublishStop 其它节点停止监听，保留定义。
func (p *ServiceCenterEventPublisher) PublishStop(ctx context.Context, tenantID, instanceName, environment, operator string) error {
	return p.publish(ctx, ServiceCenterActionStop, tenantID, instanceName, environment, operator)
}

// PublishReload 其它节点从库刷新定义；监听/TLS/鉴权变化时本机 Stop+Start。
func (p *ServiceCenterEventPublisher) PublishReload(ctx context.Context, tenantID, instanceName, environment, operator string) error {
	return p.publish(ctx, ServiceCenterActionReload, tenantID, instanceName, environment, operator)
}

// PublishUnload 删除定义前通知其它节点停监听并移出进程池。
func (p *ServiceCenterEventPublisher) PublishUnload(ctx context.Context, tenantID, instanceName, environment, operator string) error {
	return p.publish(ctx, ServiceCenterActionUnload, tenantID, instanceName, environment, operator)
}

// PublishNaming 其它节点按事件改本机 Cache 并推订阅，不写库。
func (p *ServiceCenterEventPublisher) PublishNaming(ctx context.Context, ev model.NamingEvent) error {
	if ev.CenterInstanceName == "" {
		return nil
	}
	return p.publishEvent(ctx, ServiceCenterActionNaming, ServiceCenterEventPayload{
		InstanceName: ev.CenterInstanceName,
		Environment:  ev.Environment,
		Naming:       &ev,
	})
}

// PublishNamespace 其它节点按事件改本机 L1 命名空间，不写库。
func (p *ServiceCenterEventPublisher) PublishNamespace(ctx context.Context, ev model.NamespaceEvent) error {
	if ev.CenterInstanceName == "" {
		return nil
	}
	return p.publishEvent(ctx, ServiceCenterActionNamespace, ServiceCenterEventPayload{
		InstanceName: ev.CenterInstanceName,
		Environment:  ev.Environment,
		Namespace:    &ev,
	})
}

// PublishConfig 其它节点读库后推本机 Watch。正文不进事件。
func (p *ServiceCenterEventPublisher) PublishConfig(ctx context.Context, ev model.ConfigEvent) error {
	if ev.CenterInstanceName == "" {
		return nil
	}
	if ev.Release != nil && ev.Release.Content != "" {
		cp := *ev.Release
		cp.Content = ""
		ev.Release = &cp
	}
	return p.publishEvent(ctx, ServiceCenterActionConfig, ServiceCenterEventPayload{
		InstanceName: ev.CenterInstanceName,
		Environment:  ev.Environment,
		Config:       &ev,
	})
}

func (p *ServiceCenterEventPublisher) publish(ctx context.Context, action, tenantID, instanceName, environment, operator string) error {
	now := time.Now()
	return p.publishEvent(ctx, action, ServiceCenterEventPayload{
		TenantId:      tenantID,
		InstanceName:  instanceName,
		Environment:   environment,
		Operator:      operator,
		RequestTimeMs: now.UnixMilli(),
	})
}

func (p *ServiceCenterEventPublisher) publishEvent(ctx context.Context, action string, payload ServiceCenterEventPayload) error {
	if payload.InstanceName == "" {
		return fmt.Errorf("instanceName is required")
	}
	if !types.IsClusterInitialized() || !types.IsClusterReady() {
		logger.Debug("集群服务未就绪，跳过服务中心事件",
			"action", action, "instanceName", payload.InstanceName, "environment", payload.Environment)
		return nil
	}
	clusterService := types.GetClusterService()
	if clusterService == nil {
		logger.Warn("无法获取集群服务，跳过服务中心事件", "action", action, "instanceName", payload.InstanceName)
		return nil
	}

	expire := time.Now().Add(10 * time.Minute)
	event := &types.ClusterEvent{
		EventType:   ServiceCenterEventType,
		EventAction: action,
		ExpireTime:  &expire,
	}
	if payload.RequestTimeMs == 0 {
		payload.RequestTimeMs = time.Now().UnixMilli()
	}
	if err := event.SetPayload(payload); err != nil {
		return fmt.Errorf("设置事件数据失败: %w", err)
	}
	if err := clusterService.PublishEvent(ctx, event); err != nil {
		return fmt.Errorf("发布集群事件失败: %w", err)
	}
	logger.Info("服务中心集群事件已发布",
		"action", action,
		"instanceName", payload.InstanceName,
		"environment", payload.Environment,
		"eventId", event.EventId)
	return nil
}

// ServiceCenterEventPayload 服务中心集群事件载荷。Environment 参与运行时键，必须带上。
type ServiceCenterEventPayload struct {
	TenantId      string                `json:"tenantId"`
	InstanceName  string                `json:"instanceName"`
	Environment   string                `json:"environment"`
	Operator      string                `json:"operator"`
	RequestTimeMs int64                 `json:"requestTimeMs"`
	Naming        *model.NamingEvent    `json:"naming,omitempty"`
	Config        *model.ConfigEvent    `json:"config,omitempty"`
	Namespace     *model.NamespaceEvent `json:"namespace,omitempty"`
}
