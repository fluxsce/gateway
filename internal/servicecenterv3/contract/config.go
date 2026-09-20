package contract

import (
	"context"

	"gateway/internal/servicecenterv3/model"
)

// Config 是草稿、发布、已发布读、历史与 Watch。
//
// SaveDraft 不落已发布表、不推 SDK。Publish / Delete 写库并推送本机 Watch，再通知其它网关读库。
// GetPublished / ListPublished / Watch 快照按次读库。草稿也在库中，任意网关可发布。
// Watch 只投递已发布变更。实现不得 close 调用方传入的 ch。
type Config interface {
	// GetDraft 读取未发布草稿；无草稿时返回 ErrConfigNotFound。
	GetDraft(ctx context.Context, cc CallContext, group, dataID string) (*model.ConfigDraft, error)
	// SaveDraft 写入或覆盖未发布草稿，只写库，不推 SDK。
	SaveDraft(ctx context.Context, cc CallContext, draft *model.ConfigDraft) error
	// Publish 将当前草稿（或传入的已有内容）发布为新版本并推送。
	Publish(ctx context.Context, cc CallContext, group, dataID, reason string) (*model.ConfigRelease, error)

	// GetPublished 读取当前已发布版本。
	GetPublished(ctx context.Context, cc CallContext, group, dataID string) (*model.ConfigRelease, error)
	// ListPublished 列出组内已发布配置；group 空表示全部。
	ListPublished(ctx context.Context, cc CallContext, group string) ([]*model.ConfigRelease, error)

	// ListReleases 按版本倒序列出历史，limit<=0 时由实现给默认上限。
	ListReleases(ctx context.Context, cc CallContext, group, dataID string, limit int) ([]*model.ConfigRelease, error)
	// Diff 比较两个已发布版本的内容。
	Diff(ctx context.Context, cc CallContext, group, dataID string, fromVersion, toVersion int64) (*model.ConfigDiff, error)
	// Rollback 将 targetVersion 重新发布为最新版本（新版本号，不删除历史）。
	Rollback(ctx context.Context, cc CallContext, group, dataID string, targetVersion int64, reason string) (*model.ConfigRelease, error)
	// Delete 删除已发布配置及对应草稿，并推送删除事件。
	Delete(ctx context.Context, cc CallContext, group, dataID string) error

	// Watch 订阅已发布变更；dataIDs 空表示组内全部。
	Watch(ctx context.Context, cc CallContext, group string, dataIDs []string, ch chan<- model.ConfigEvent) error
	// Unwatch 取消 ch 上的配置订阅，不关闭 ch。
	Unwatch(ctx context.Context, cc CallContext, ch chan<- model.ConfigEvent)
	// UnwatchConfigs 取消 ch 上指定 dataID 的 Watch；dataIDs 空则取消该组下按 ID 订阅的条目。
	UnwatchConfigs(ctx context.Context, cc CallContext, group string, dataIDs []string, ch chan<- model.ConfigEvent)
	// BindClient 把 Watch 通道绑到数据面连接，断连时由 OnClientLost 取消。
	BindClient(ch chan<- model.ConfigEvent, connectionID string)

	// OnClientLost 连接断开时取消该连接绑定的 Watch。
	OnClientLost(ctx context.Context, connectionID string)
}
