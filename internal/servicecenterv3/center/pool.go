package center

import (
	"context"
	"sync"

	"gateway/internal/servicecenterv3/contract"
	"gateway/internal/servicecenterv3/infra/store"
	"gateway/internal/servicecenterv3/model"
	"gateway/pkg/logger"
)

// Pool 是进程内中心实例池，实现 contract.Admin。
type Pool struct {
	store *store.Store
	mu    sync.RWMutex
	items map[string]*Instance // RuntimeKey = instanceName:environment
}

// NewPool 构造空池。st 由调用方保证生命周期覆盖池。
func NewPool(st *store.Store) *Pool {
	return &Pool{
		store: st,
		items: map[string]*Instance{},
	}
}

// Store 返回持久化聚合，供引导加载使用。
func (p *Pool) Store() *store.Store { return p.store }

// GetInstance 按实例名取运行时。同名多环境时必须用 FindInstance 带 environment。
func (p *Pool) GetInstance(instanceName string) (*Instance, bool) {
	return p.FindInstance(instanceName, "")
}

// FindInstance 按实例名与环境取运行时。environment 空且同名仅一条时返回该条。
func (p *Pool) FindInstance(instanceName, environment string) (*Instance, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if environment != "" {
		inst, ok := p.items[runtimeKey(instanceName, environment)]
		return inst, ok
	}
	if inst, ok := p.items[instanceName]; ok {
		return inst, true
	}
	var found *Instance
	n := 0
	for _, inst := range p.items {
		if inst.cfg != nil && inst.cfg.InstanceName == instanceName {
			found = inst
			n++
		}
	}
	if n == 1 {
		return found, true
	}
	return nil, false
}

// ResolveByNamespace 用命名空间反查中心实例名和环境。历史路由元数据往往没有这两项。
func (p *Pool) ResolveByNamespace(ctx context.Context, tenantID, namespaceID string) (instanceName, environment string, err error) {
	if p == nil || namespaceID == "" {
		return "", "", contract.ErrNamespaceNotFound
	}
	p.mu.RLock()
	items := make([]*Instance, 0, len(p.items))
	for _, inst := range p.items {
		items = append(items, inst)
	}
	p.mu.RUnlock()
	for _, inst := range items {
		ns, ok := inst.CachedNamespace(namespaceID)
		if !ok || ns == nil {
			continue
		}
		if tenantID != "" && ns.TenantID != "" && ns.TenantID != tenantID {
			continue
		}
		name := ns.CenterInstanceName
		env := ns.Environment
		if cfg := inst.Config(); cfg != nil {
			if name == "" {
				name = cfg.InstanceName
			}
			if env == "" {
				env = cfg.Environment
			}
		}
		if name != "" {
			return name, env, nil
		}
	}
	if p.store != nil && p.store.Namespace != nil && tenantID != "" {
		ns, getErr := p.store.Namespace.Get(ctx, tenantID, namespaceID)
		if getErr != nil {
			return "", "", getErr
		}
		if ns != nil && ns.CenterInstanceName != "" {
			return ns.CenterInstanceName, ns.Environment, nil
		}
	}
	return "", "", contract.ErrNamespaceNotFound
}

func runtimeKey(instanceName, environment string) string {
	if environment == "" {
		return instanceName
	}
	return instanceName + ":" + environment
}

func putInstance(p *Pool, inst *Instance) {
	key := inst.Name()
	if inst.cfg != nil {
		key = inst.cfg.RuntimeKey()
	}
	p.items[key] = inst
}

func deleteInstance(p *Pool, instanceName, environment string) {
	delete(p.items, runtimeKey(instanceName, environment))
	if environment != "" {
		delete(p.items, instanceName)
	}
}

// mustInstance 按 CallContext 取已加载运行时，缺失时返回哨兵错误。
func (p *Pool) mustInstance(cc contract.CallContext) (*Instance, error) {
	if err := cc.RequireCenter(); err != nil {
		return nil, err
	}
	inst, ok := p.FindInstance(cc.CenterInstanceName, cc.Environment)
	if !ok {
		return nil, contract.ErrCenterNotFound
	}
	return inst, nil
}

// List 列出租户下的中心实例定义（含未运行）。
func (p *Pool) List(ctx context.Context, cc contract.CallContext) ([]*model.CenterInstance, error) {
	if cc.TenantID == "" {
		return nil, contract.ErrInvalidContext
	}
	return p.store.Center.ListByTenant(ctx, cc.TenantID)
}

// Get 读取单个中心实例：优先运行时，否则回源库。
func (p *Pool) Get(ctx context.Context, cc contract.CallContext) (*model.CenterInstance, error) {
	if err := cc.RequireCenter(); err != nil {
		return nil, err
	}
	if inst, ok := p.FindInstance(cc.CenterInstanceName, cc.Environment); ok {
		return inst.Config(), nil
	}
	return p.store.Center.Get(ctx, cc.TenantID, cc.CenterInstanceName, cc.Environment)
}

// Create 创建中心实例定义，不自动 Start。
func (p *Pool) Create(ctx context.Context, cc contract.CallContext, inst *model.CenterInstance) error {
	if inst == nil || inst.InstanceName == "" {
		return contract.ErrInvalidArgument
	}
	inst.TenantID = cc.TenantID
	existing, err := p.store.Center.Get(ctx, cc.TenantID, inst.InstanceName, inst.Environment)
	if err != nil {
		return err
	}
	if existing != nil {
		return contract.ErrCenterAlreadyExists
	}
	return p.store.Center.Create(ctx, inst, cc.OperatorID)
}

// Update 更新中心实例定义；监听参数变更需随后 Reload 或 Stop+Start。
func (p *Pool) Update(ctx context.Context, cc contract.CallContext, inst *model.CenterInstance) error {
	if inst == nil {
		return contract.ErrInvalidArgument
	}
	inst.TenantID = cc.TenantID
	if inst.InstanceName == "" {
		inst.InstanceName = cc.CenterInstanceName
	}
	return p.store.Center.Update(ctx, inst, cc.OperatorID)
}

// Delete 停止运行中的实例并删除定义。
func (p *Pool) Delete(ctx context.Context, cc contract.CallContext) error {
	if err := cc.RequireCenter(); err != nil {
		return err
	}
	if inst, ok := p.FindInstance(cc.CenterInstanceName, cc.Environment); ok && inst.IsRunning() {
		_ = inst.Stop(ctx)
	}
	p.mu.Lock()
	deleteInstance(p, cc.CenterInstanceName, cc.Environment)
	p.mu.Unlock()
	cfg, err := p.store.Center.Get(ctx, cc.TenantID, cc.CenterInstanceName, cc.Environment)
	if err != nil {
		return err
	}
	env := ""
	if cfg != nil {
		env = cfg.Environment
	}
	return p.store.Center.Delete(ctx, cc.TenantID, cc.CenterInstanceName, env)
}

// Start 加载定义并打开数据面监听。
func (p *Pool) Start(ctx context.Context, cc contract.CallContext) error {
	if err := cc.RequireCenter(); err != nil {
		return err
	}
	if inst, ok := p.FindInstance(cc.CenterInstanceName, cc.Environment); ok {
		return inst.Start(ctx)
	}
	cfg, err := p.store.Center.Get(ctx, cc.TenantID, cc.CenterInstanceName, cc.Environment)
	if err != nil {
		return err
	}
	if cfg == nil {
		return contract.ErrCenterNotFound
	}
	inst := NewInstance(cfg, p.store)
	p.mu.Lock()
	putInstance(p, inst)
	p.mu.Unlock()
	return inst.Start(ctx)
}

// Stop 停止监听，保留定义。
func (p *Pool) Stop(ctx context.Context, cc contract.CallContext) error {
	inst, err := p.mustInstance(cc)
	if err != nil {
		return err
	}
	return inst.Stop(ctx)
}

// Reload 从库刷新定义指针；端口等监听参数需 Stop+Start 才生效。
func (p *Pool) Reload(ctx context.Context, cc contract.CallContext) error {
	inst, err := p.mustInstance(cc)
	if err != nil {
		return err
	}
	return inst.Reload(ctx)
}

// ListConnections 列出当前数据面会话；实例未运行返回 ErrCenterNotRunning。
func (p *Pool) ListConnections(_ context.Context, cc contract.CallContext) ([]model.ConnectionInfo, error) {
	inst, err := p.mustInstance(cc)
	if err != nil {
		return nil, err
	}
	if !inst.IsRunning() {
		return nil, contract.ErrCenterNotRunning
	}
	return inst.Connections(), nil
}

// GetListenEndpoint 返回该实例的 host:port。
func (p *Pool) GetListenEndpoint(_ context.Context, cc contract.CallContext) (string, error) {
	inst, err := p.mustInstance(cc)
	if err != nil {
		return "", err
	}
	return inst.Config().ListenEndpoint(), nil
}

// Overview 返回运行时计数快照。
func (p *Pool) Overview(_ context.Context, cc contract.CallContext) (*model.Overview, error) {
	inst, err := p.mustInstance(cc)
	if err != nil {
		return nil, err
	}
	return inst.Overview(), nil
}

// LoadAndStart 加载库中全部已启用中心实例并尝试 Start；单个失败只记日志。
func (p *Pool) LoadAndStart(ctx context.Context) error {
	list, err := p.store.Center.ListActive(ctx)
	if err != nil {
		return err
	}
	for _, cfg := range list {
		inst := NewInstance(cfg, p.store)
		p.mu.Lock()
		putInstance(p, inst)
		p.mu.Unlock()
		if err := inst.Start(ctx); err != nil {
			logger.Error("启动 servicecenterv3 实例失败", err, "instance", cfg.InstanceName)
			continue
		}
	}
	return nil
}

// StopAll 停止池内全部运行时。
func (p *Pool) StopAll(ctx context.Context) error {
	p.mu.RLock()
	items := make([]*Instance, 0, len(p.items))
	for _, inst := range p.items {
		items = append(items, inst)
	}
	p.mu.RUnlock()
	for _, inst := range items {
		_ = inst.Stop(ctx)
	}
	return nil
}

// HasRunning 报告是否至少有一个实例在监听。
func (p *Pool) HasRunning() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, inst := range p.items {
		if inst.IsRunning() {
			return true
		}
	}
	return false
}

// RunningInstances 返回当前正在监听的中心实例快照。
func (p *Pool) RunningInstances() []*Instance {
	p.mu.RLock()
	defer p.mu.RUnlock()
	out := make([]*Instance, 0, len(p.items))
	for _, inst := range p.items {
		if inst != nil && inst.IsRunning() {
			out = append(out, inst)
		}
	}
	return out
}

// NamingOf 返回已运行实例的 Naming；environment 空时同名仅一条才命中。
func (p *Pool) NamingOf(instanceName, environment string) (contract.Naming, error) {
	inst, ok := p.FindInstance(instanceName, environment)
	if !ok {
		return nil, contract.ErrCenterNotFound
	}
	if !inst.IsRunning() {
		return nil, contract.ErrCenterNotRunning
	}
	return inst.Naming(), nil
}

// ConfigOf 返回已运行实例的 Config；environment 空时同名仅一条才命中。
func (p *Pool) ConfigOf(instanceName, environment string) (contract.Config, error) {
	inst, ok := p.FindInstance(instanceName, environment)
	if !ok {
		return nil, contract.ErrCenterNotFound
	}
	if !inst.IsRunning() {
		return nil, contract.ErrCenterNotRunning
	}
	return inst.ConfigService(), nil
}

// FindNode 在已加载运行时中按业务节点 ID 查找，供管理面编辑/下线使用。
func (p *Pool) FindNode(nodeID string) (*model.Node, *Instance, bool) {
	if nodeID == "" {
		return nil, nil, false
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, inst := range p.items {
		if node, ok := inst.FindNode(nodeID); ok {
			return node, inst, true
		}
	}
	return nil, nil, false
}

// Unload 停止并移出运行时，不删除库中的中心实例定义。
func (p *Pool) Unload(ctx context.Context, instanceName, environment string) {
	inst, ok := p.FindInstance(instanceName, environment)
	if ok {
		_ = inst.Stop(ctx)
	}
	p.mu.Lock()
	deleteInstance(p, instanceName, environment)
	p.mu.Unlock()
}

// 保证 Pool 实现 contract.Admin
var _ contract.Admin = (*Pool)(nil)
