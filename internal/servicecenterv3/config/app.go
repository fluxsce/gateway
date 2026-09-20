package config

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"time"

	"gateway/internal/servicecenterv3/contract"
	"gateway/internal/servicecenterv3/infra/alert"
	"gateway/internal/servicecenterv3/infra/cache"
	"gateway/internal/servicecenterv3/infra/store"
	"gateway/internal/servicecenterv3/model"
)

// App 实现 contract.Config，绑定单一 CenterInstanceName。
type App struct {
	centerName string
	cfg        *model.CenterInstance
	store      *store.Store
	cache      *cache.Cache
	subs       *watcher
}

// New 构造配置应用。配置正文不进 Cache，草稿和已发布都读库。
func New(centerName string, st *store.Store, c *cache.Cache) *App {
	return &App{
		centerName: centerName,
		store:      st,
		cache:      c,
		subs:       newWatcher(),
	}
}

// SetCenter 绑定中心实例定义，供告警与命名空间归属校验。Reload 后再次调用。
func (a *App) SetCenter(cfg *model.CenterInstance) {
	a.cfg = cfg
	if cfg != nil && cfg.InstanceName != "" {
		a.centerName = cfg.InstanceName
	}
}

// ensureNamespace 与 Naming 同一套规则：存在、属于本实例、未停用、令牌授权。
func (a *App) ensureNamespace(ctx context.Context, cc contract.CallContext) error {
	if err := cc.RequireNamespace(); err != nil {
		return err
	}
	if !cc.AllowsNamespace(cc.NamespaceID) {
		return contract.ErrNamespaceForbidden
	}
	ns, ok := a.cache.GetNamespace(cc.NamespaceID)
	if !ok {
		if a.store == nil {
			return contract.ErrNamespaceNotFound
		}
		loaded, err := a.store.Namespace.Get(ctx, cc.TenantID, cc.NamespaceID)
		if err != nil {
			return err
		}
		if loaded == nil {
			return contract.ErrNamespaceNotFound
		}
		a.cache.SetNamespace(loaded)
		ns = loaded
	}
	if ns.CenterInstanceName != "" && ns.CenterInstanceName != cc.CenterInstanceName {
		return contract.ErrNamespaceNotFound
	}
	if !ns.Active {
		return contract.ErrNamespaceDisabled
	}
	return nil
}

func (a *App) checkConfigQuota(ctx context.Context, cc contract.CallContext) error {
	ns, ok := a.cache.GetNamespace(cc.NamespaceID)
	if !ok || ns.ConfigQuota <= 0 {
		return nil
	}
	n, err := a.countPublished(ctx, cc.TenantID, cc.NamespaceID)
	if err != nil {
		return err
	}
	if n >= ns.ConfigQuota {
		return contract.ErrQuotaExceeded
	}
	return nil
}

func (a *App) countPublished(ctx context.Context, tenantID, namespaceID string) (int, error) {
	if a.store == nil || a.store.Config == nil {
		return 0, nil
	}
	return a.store.Config.Count(ctx, tenantID, namespaceID)
}

// CountPublishedInNamespace 统计某一命名空间已发布配置条数。读库，不走缓存。
func (a *App) CountPublishedInNamespace(ctx context.Context, tenantID, namespaceID string) int {
	if tenantID == "" || namespaceID == "" {
		return 0
	}
	n, err := a.countPublished(ctx, tenantID, namespaceID)
	if err != nil {
		return 0
	}
	return n
}

// CountPublished 统计本中心已发布配置条数，供概览。读库，不走缓存。
func (a *App) CountPublished(ctx context.Context, tenantID string) int {
	if a.store == nil || a.store.Config == nil {
		return 0
	}
	n, err := a.store.Config.CountByCenter(ctx, tenantID, a.centerName)
	if err != nil {
		return 0
	}
	return n
}

// GetDraft 读取未发布草稿（库表）；无草稿时返回 ErrConfigNotFound。
func (a *App) GetDraft(ctx context.Context, cc contract.CallContext, group, dataID string) (*model.ConfigDraft, error) {
	if err := a.ensureNamespace(ctx, cc); err != nil {
		return nil, err
	}
	if group == "" {
		group = "DEFAULT_GROUP"
	}
	if a.store == nil || a.store.Config == nil {
		return nil, contract.ErrConfigNotFound
	}
	draft, err := a.store.Config.GetDraft(ctx, cc.TenantID, cc.NamespaceID, group, dataID)
	if err != nil {
		return nil, err
	}
	if draft == nil {
		return nil, contract.ErrConfigNotFound
	}
	draft.CenterInstanceName = cc.CenterInstanceName
	return draft, nil
}

// SaveDraft 写入或覆盖未发布草稿：只写库，不推给 SDK。
func (a *App) SaveDraft(ctx context.Context, cc contract.CallContext, draft *model.ConfigDraft) error {
	if err := a.ensureNamespace(ctx, cc); err != nil {
		return err
	}
	if draft == nil || draft.DataID == "" {
		return contract.ErrInvalidArgument
	}
	if a.store == nil || a.store.Config == nil {
		return contract.ErrInvalidArgument
	}
	if draft.GroupName == "" {
		draft.GroupName = "DEFAULT_GROUP"
	}
	draft.TenantID = cc.TenantID
	draft.CenterInstanceName = cc.CenterInstanceName
	draft.NamespaceID = cc.NamespaceID
	return a.store.Config.SaveDraft(ctx, draft, cc.OperatorID)
}

// Publish 将当前草稿发布为新版本：写库并推送，不把正文放入内存缓存。
func (a *App) Publish(ctx context.Context, cc contract.CallContext, group, dataID, reason string) (*model.ConfigRelease, error) {
	if err := a.ensureNamespace(ctx, cc); err != nil {
		return nil, err
	}
	if group == "" {
		group = "DEFAULT_GROUP"
	}
	draft, err := a.GetDraft(ctx, cc, group, dataID)
	if err != nil {
		return nil, err
	}
	old, _ := a.GetPublished(ctx, cc, group, dataID)
	if old == nil {
		if err := a.checkConfigQuota(ctx, cc); err != nil {
			return nil, err
		}
	}
	version := int64(1)
	if old != nil {
		version = old.Version + 1
	}
	sum := md5.Sum([]byte(draft.Content))
	rel := &model.ConfigRelease{
		TenantID:           cc.TenantID,
		CenterInstanceName: cc.CenterInstanceName,
		NamespaceID:        cc.NamespaceID,
		GroupName:          group,
		DataID:             dataID,
		Content:            draft.Content,
		ContentType:        draft.ContentType,
		Description:        draft.Description,
		Version:            version,
		MD5:                hex.EncodeToString(sum[:]),
		Reason:             reason,
		PublishedBy:        cc.OperatorID,
		PublishedAt:        time.Now(),
	}
	changeType := "UPDATE"
	if old == nil {
		changeType = "CREATE"
	}
	if err := a.store.Config.SaveRelease(ctx, rel, old, changeType, cc.OperatorID); err != nil {
		return nil, err
	}
	_ = a.store.Config.DeleteDraft(ctx, cc.TenantID, cc.NamespaceID, group, dataID)
	a.emit(model.ConfigEvent{
		Type:               model.EventConfigPublished,
		Timestamp:          rel.PublishedAt,
		CenterInstanceName: cc.CenterInstanceName,
		NamespaceID:        cc.NamespaceID,
		GroupName:          group,
		DataID:             dataID,
		Release:            rel,
	})
	alert.ConfigChange(a.cfg, alert.ConfigInfo{
		ChangeType:  changeType,
		NamespaceID: cc.NamespaceID,
		GroupName:   group,
		DataID:      dataID,
		Version:     rel.Version,
		ChangedBy:   cc.OperatorID,
	})
	return rel, nil
}

// GetPublished 读取当前已发布版本，每次查库，不缓存正文。
func (a *App) GetPublished(ctx context.Context, cc contract.CallContext, group, dataID string) (*model.ConfigRelease, error) {
	if err := a.ensureNamespace(ctx, cc); err != nil {
		return nil, err
	}
	if group == "" {
		group = "DEFAULT_GROUP"
	}
	rel, err := a.loadPublished(ctx, cc.TenantID, cc.NamespaceID, group, dataID)
	if err != nil {
		return nil, err
	}
	if rel == nil {
		return nil, contract.ErrConfigNotFound
	}
	rel.CenterInstanceName = cc.CenterInstanceName
	return rel, nil
}

// ListPublished 列出组内已发布配置；group 空表示全部。每次查库。
func (a *App) ListPublished(ctx context.Context, cc contract.CallContext, group string) ([]*model.ConfigRelease, error) {
	if err := a.ensureNamespace(ctx, cc); err != nil {
		return nil, err
	}
	if a.store == nil || a.store.Config == nil {
		return nil, nil
	}
	loaded, err := a.store.Config.List(ctx, cc.TenantID, cc.NamespaceID, group)
	if err != nil {
		return nil, err
	}
	for _, rel := range loaded {
		rel.CenterInstanceName = cc.CenterInstanceName
	}
	return loaded, nil
}

// ListReleases 按版本倒序列出发布历史。
func (a *App) ListReleases(ctx context.Context, cc contract.CallContext, group, dataID string, limit int) ([]*model.ConfigRelease, error) {
	if err := a.ensureNamespace(ctx, cc); err != nil {
		return nil, err
	}
	if group == "" {
		group = "DEFAULT_GROUP"
	}
	return a.store.Config.ListHistory(ctx, cc.TenantID, cc.NamespaceID, group, dataID, limit)
}

// Diff 比较两个已发布版本的内容。
func (a *App) Diff(ctx context.Context, cc contract.CallContext, group, dataID string, fromVersion, toVersion int64) (*model.ConfigDiff, error) {
	history, err := a.ListReleases(ctx, cc, group, dataID, 200)
	if err != nil {
		return nil, err
	}
	diff := &model.ConfigDiff{DataID: dataID, FromVersion: fromVersion, ToVersion: toVersion}
	for _, rel := range history {
		if rel.Version == fromVersion {
			diff.FromContent = rel.Content
		}
		if rel.Version == toVersion {
			diff.ToContent = rel.Content
		}
	}
	if diff.FromContent == "" && diff.ToContent == "" {
		return nil, contract.ErrReleaseNotFound
	}
	return diff, nil
}

// Rollback 将 targetVersion 重新发布为最新版本（新版本号，不删历史）。
func (a *App) Rollback(ctx context.Context, cc contract.CallContext, group, dataID string, targetVersion int64, reason string) (*model.ConfigRelease, error) {
	history, err := a.ListReleases(ctx, cc, group, dataID, 200)
	if err != nil {
		return nil, err
	}
	var target *model.ConfigRelease
	for _, rel := range history {
		if rel.Version == targetVersion {
			target = rel
			break
		}
	}
	if target == nil {
		return nil, contract.ErrReleaseNotFound
	}
	if err := a.SaveDraft(ctx, cc, releaseToDraft(target)); err != nil {
		return nil, err
	}
	rel, err := a.Publish(ctx, cc, group, dataID, reason)
	if err != nil {
		return nil, err
	}
	a.emit(model.ConfigEvent{
		Type:               model.EventConfigRolledBack,
		Timestamp:          time.Now(),
		CenterInstanceName: cc.CenterInstanceName,
		NamespaceID:        cc.NamespaceID,
		GroupName:          group,
		DataID:             dataID,
		Release:            rel,
	})
	alert.ConfigChange(a.cfg, alert.ConfigInfo{
		ChangeType:  "ROLLBACK",
		NamespaceID: cc.NamespaceID,
		GroupName:   group,
		DataID:      dataID,
		Version:     rel.Version,
		ChangedBy:   cc.OperatorID,
	})
	return rel, nil
}

// Delete 删除已发布配置及对应草稿，并推送 CONFIG_DELETED。
func (a *App) Delete(ctx context.Context, cc contract.CallContext, group, dataID string) error {
	if err := a.ensureNamespace(ctx, cc); err != nil {
		return err
	}
	if group == "" {
		group = "DEFAULT_GROUP"
	}
	old, _ := a.loadPublished(ctx, cc.TenantID, cc.NamespaceID, group, dataID)
	if err := a.store.Config.Delete(ctx, cc.TenantID, cc.NamespaceID, group, dataID); err != nil {
		return err
	}
	a.emit(model.ConfigEvent{
		Type:               model.EventConfigDeleted,
		Timestamp:          time.Now(),
		CenterInstanceName: cc.CenterInstanceName,
		NamespaceID:        cc.NamespaceID,
		GroupName:          group,
		DataID:             dataID,
		Release:            old,
	})
	alert.ConfigChange(a.cfg, alert.ConfigInfo{
		ChangeType:  "DELETE",
		NamespaceID: cc.NamespaceID,
		GroupName:   group,
		DataID:      dataID,
		ChangedBy:   cc.OperatorID,
	})
	return nil
}

// Watch 订阅已发布变更；dataIDs 空表示组内全部。
func (a *App) Watch(ctx context.Context, cc contract.CallContext, group string, dataIDs []string, ch chan<- model.ConfigEvent) error {
	if err := a.ensureNamespace(ctx, cc); err != nil {
		return err
	}
	if group == "" {
		group = "DEFAULT_GROUP"
	}
	a.subs.add(cc.NamespaceID, group, dataIDs, ch, cc.ConnectionID)
	a.snapshotPublished(ctx, cc, group, dataIDs, ch)
	return nil
}

// Unwatch 取消 ch 上的配置订阅，不关闭 ch。
func (a *App) Unwatch(_ context.Context, _ contract.CallContext, ch chan<- model.ConfigEvent) {
	a.subs.remove(ch)
}

// UnwatchConfigs 取消 ch 上指定 dataID 的 Watch。
func (a *App) UnwatchConfigs(_ context.Context, cc contract.CallContext, group string, dataIDs []string, ch chan<- model.ConfigEvent) {
	if group == "" {
		group = "DEFAULT_GROUP"
	}
	a.subs.removeDataIDs(ch, cc.NamespaceID, group, dataIDs)
}

// OnClientLost 取消该连接绑定的全部 Watch。
func (a *App) OnClientLost(_ context.Context, connectionID string) {
	a.subs.removeByConnection(connectionID)
}

// BindClient 实现 contract.Config。
func (a *App) BindClient(ch chan<- model.ConfigEvent, connectionID string) {
	a.subs.bindConnection(ch, connectionID)
}

func (a *App) loadPublished(ctx context.Context, tenantID, namespaceID, group, dataID string) (*model.ConfigRelease, error) {
	if a.store == nil || a.store.Config == nil {
		return nil, nil
	}
	return a.store.Config.Get(ctx, tenantID, namespaceID, group, dataID)
}

var _ contract.Config = (*App)(nil)

// releaseToDraft 把已发布快照转成草稿结构，供回滚前编辑。
func releaseToDraft(rel *model.ConfigRelease) *model.ConfigDraft {
	return &model.ConfigDraft{
		TenantID:           rel.TenantID,
		CenterInstanceName: rel.CenterInstanceName,
		NamespaceID:        rel.NamespaceID,
		GroupName:          rel.GroupName,
		DataID:             rel.DataID,
		Content:            rel.Content,
		ContentType:        rel.ContentType,
		Description:        rel.Description,
	}
}
