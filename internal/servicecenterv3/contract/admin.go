package contract

import (
	"context"

	"gateway/internal/servicecenterv3/model"
)

// Admin 是中心实例生命周期与只读概览，供管理面（网关 HTTP）进程内调用。
// 实现必须按 CallContext.TenantID 隔离，不得用全局列表串租户。
type Admin interface {
	// List 列出当前租户下的中心实例定义（含未运行）。
	List(ctx context.Context, cc CallContext) ([]*model.CenterInstance, error)
	// Get 读取单个中心实例定义。
	Get(ctx context.Context, cc CallContext) (*model.CenterInstance, error)
	// Create 创建中心实例定义，不自动 Start。
	Create(ctx context.Context, cc CallContext, inst *model.CenterInstance) error
	// Update 更新定义；若实例正在运行，实现可要求随后 Reload。
	Update(ctx context.Context, cc CallContext, inst *model.CenterInstance) error
	// Delete 停止并删除定义。运行中的实例必须先停监听。
	Delete(ctx context.Context, cc CallContext) error
	// Start 加载并监听指定实例。
	Start(ctx context.Context, cc CallContext) error
	// Stop 停止监听，不删除定义。
	Stop(ctx context.Context, cc CallContext) error
	// Reload 热加载配置：先 Stop 再 Start，会话会断开。
	Reload(ctx context.Context, cc CallContext) error
	// ListConnections 列出当前数据面会话。
	ListConnections(ctx context.Context, cc CallContext) ([]model.ConnectionInfo, error)
	// GetListenEndpoint 返回 host:port。
	GetListenEndpoint(ctx context.Context, cc CallContext) (string, error)
	// Overview 返回运行时计数快照。
	Overview(ctx context.Context, cc CallContext) (*model.Overview, error)
}
