package config

import (
	"context"
	"time"

	"gateway/internal/cluster/publish"
	"gateway/internal/servicecenterv3/contract"
	"gateway/internal/servicecenterv3/model"
	"gateway/pkg/logger"
)

// ConfigHook 仅测试用。生产路径在发布/回滚/删除时直接 PublishConfig。ApplyRemote 不得触发。
var ConfigHook func(model.ConfigEvent)

var clusterPub = publish.NewServiceCenterEventPublisher()

func replicateConfig(ev model.ConfigEvent) {
	if ConfigHook != nil {
		ConfigHook(ev)
		return
	}
	if err := clusterPub.PublishConfig(context.Background(), ev); err != nil {
		logger.Warn("发布服务中心配置集群事件失败", "error", err)
	}
}

func (a *App) environment() string {
	if a.cfg != nil {
		return a.cfg.Environment
	}
	return ""
}

func (a *App) tenantID() string {
	if a.cfg != nil {
		return a.cfg.TenantID
	}
	return ""
}

func (a *App) emit(ev model.ConfigEvent) {
	if ev.CenterInstanceName == "" {
		ev.CenterInstanceName = a.centerName
	}
	if ev.Environment == "" {
		ev.Environment = a.environment()
	}
	if ev.Timestamp.IsZero() {
		ev.Timestamp = time.Now()
	}
	a.subs.publish(ev)
	replicateConfig(stripConfigContent(ev))
}

func stripConfigContent(ev model.ConfigEvent) model.ConfigEvent {
	if ev.Release == nil {
		return ev
	}
	cp := *ev.Release
	cp.Content = ""
	ev.Release = &cp
	return ev
}

// ApplyRemote 应用其它副本的已发布配置变更：读库补正文并推给本机 Watch，不写库、不回复制、不缓存。
// 按库中 Version 丢弃过期包。
func (a *App) ApplyRemote(ev model.ConfigEvent) {
	ns, group, dataID := ev.NamespaceID, ev.GroupName, ev.DataID
	if ev.Release != nil {
		if ns == "" {
			ns = ev.Release.NamespaceID
		}
		if group == "" {
			group = ev.Release.GroupName
		}
		if dataID == "" {
			dataID = ev.Release.DataID
		}
	}
	current := a.currentPublished(context.Background(), ev, ns, group, dataID)
	switch ev.Type {
	case model.EventConfigPublished, model.EventConfigRolledBack:
		if ev.Release == nil {
			return
		}
		if staleRemoteConfig(current, ev) {
			return
		}
		if ev.Release.Content == "" {
			if current == nil || current.Content == "" {
				return
			}
			ev.Release = current
		}
	case model.EventConfigDeleted:
		if staleRemoteConfig(current, ev) {
			return
		}
	}
	a.subs.publish(ev)
}

func (a *App) currentPublished(ctx context.Context, ev model.ConfigEvent, namespaceID, group, dataID string) *model.ConfigRelease {
	tenantID := a.tenantID()
	if ev.Release != nil && ev.Release.TenantID != "" {
		tenantID = ev.Release.TenantID
	}
	rel, err := a.loadPublished(ctx, tenantID, namespaceID, group, dataID)
	if err != nil || rel == nil {
		return nil
	}
	return rel
}

func staleRemoteConfig(current *model.ConfigRelease, ev model.ConfigEvent) bool {
	if current == nil {
		return false
	}
	if ev.Release != nil && current.Version > ev.Release.Version {
		return true
	}
	if ev.Type == model.EventConfigDeleted && !ev.Timestamp.IsZero() && current.PublishedAt.After(ev.Timestamp) {
		return true
	}
	return false
}

func (a *App) snapshotPublished(ctx context.Context, cc contract.CallContext, group string, dataIDs []string, ch chan<- model.ConfigEvent) {
	if ch == nil {
		return
	}
	now := time.Now()
	push := func(rel *model.ConfigRelease) {
		if rel == nil {
			return
		}
		deliverConfig(ch, model.ConfigEvent{
			Type:               model.EventConfigPublished,
			Timestamp:          now,
			CenterInstanceName: a.centerName,
			Environment:        a.environment(),
			NamespaceID:        rel.NamespaceID,
			GroupName:          rel.GroupName,
			DataID:             rel.DataID,
			Release:            rel,
		})
	}
	if len(dataIDs) == 0 {
		list, err := a.ListPublished(ctx, cc, group)
		if err != nil {
			return
		}
		for _, rel := range list {
			push(rel)
		}
		return
	}
	for _, id := range dataIDs {
		if id == "" {
			continue
		}
		rel, err := a.GetPublished(ctx, cc, group, id)
		if err != nil || rel == nil {
			continue
		}
		push(rel)
	}
}
