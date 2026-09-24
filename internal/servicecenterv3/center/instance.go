package center

import (
	"context"
	"sync/atomic"
	"time"

	"gateway/internal/servicecenterv3/access/stream"
	"gateway/internal/servicecenterv3/config"
	"gateway/internal/servicecenterv3/contract"
	"gateway/internal/servicecenterv3/infra/alert"
	"gateway/internal/servicecenterv3/infra/cache"
	"gateway/internal/servicecenterv3/infra/live"
	"gateway/internal/servicecenterv3/infra/store"
	"gateway/internal/servicecenterv3/model"
	"gateway/internal/servicecenterv3/naming"
	pkgcache "gateway/pkg/cache"
	appconfig "gateway/pkg/config"
	"gateway/pkg/logger"
)

// Instance 是一个中心实例的隔离运行时（cache / naming / config / stream）。
type Instance struct {
	cfg     *model.CenterInstance
	store   *store.Store
	cache   *cache.Cache
	naming  *naming.App
	config  *config.App
	evictor *naming.Evictor
	session *SessionHub
	server  *stream.Server
	running atomic.Bool
}

// NewInstance 装配运行时，不启动监听。
func NewInstance(cfg *model.CenterInstance, st *store.Store) *Instance {
	c := cache.New(cfg.InstanceName)
	n := naming.New(cfg.InstanceName, st, c)
	n.SetCenter(cfg)
	attachLiveView(n)
	cf := config.New(cfg.InstanceName, st, c)
	cf.SetCenter(cfg)
	interval := time.Duration(cfg.HealthCheckInterval) * time.Second
	timeout := time.Duration(cfg.HealthCheckTimeout) * time.Second
	inst := &Instance{
		cfg:     cfg,
		store:   st,
		cache:   c,
		naming:  n,
		config:  cf,
		evictor: naming.NewEvictor(n, interval, timeout),
		session: NewSessionHub(),
	}
	inst.server = stream.NewServer(cfg, n, cf, inst.session, st)
	return inst
}

// Name 返回中心实例名。
func (i *Instance) Name() string { return i.cfg.InstanceName }

// Config 返回实例定义（可被 Reload 替换指针）。
func (i *Instance) Config() *model.CenterInstance { return i.cfg }

// Naming 返回本实例的 Naming 实现。
func (i *Instance) Naming() contract.Naming { return i.naming }

// ConfigService 返回本实例的 Config 实现。
func (i *Instance) ConfigService() contract.Config { return i.config }

// CachedNamespace 读本实例 L1 中的命名空间。
func (i *Instance) CachedNamespace(namespaceID string) (*model.Namespace, bool) {
	if i == nil || i.cache == nil || namespaceID == "" {
		return nil, false
	}
	return i.cache.GetNamespace(namespaceID)
}

// SyncNamespace 把管理面刚写入的命名空间覆盖进本实例 L1。
func (i *Instance) SyncNamespace(ns *model.Namespace) {
	if i == nil || i.cache == nil || ns == nil || ns.NamespaceID == "" {
		return
	}
	i.cache.SetNamespace(ns)
}

// HasCachedServices 报告本实例 L1 是否还有该命名空间下的服务（含自动创建）。
func (i *Instance) HasCachedServices(namespaceID string) bool {
	if i == nil || i.cache == nil || namespaceID == "" {
		return false
	}
	return len(i.cache.ListServices(namespaceID, "")) > 0
}

// DropNamespace 从本实例 L1 去掉命名空间及其服务、节点。
func (i *Instance) DropNamespace(namespaceID string) {
	if i == nil || i.cache == nil {
		return
	}
	i.cache.DeleteNamespace(namespaceID)
}

// CountPublishedInNamespace 统计某一命名空间已发布配置条数。
func (i *Instance) CountPublishedInNamespace(ctx context.Context, tenantID, namespaceID string) int {
	if i == nil || i.config == nil {
		return 0
	}
	return i.config.CountPublishedInNamespace(ctx, tenantID, namespaceID)
}

// IsRunning 报告数据面是否在监听。
func (i *Instance) IsRunning() bool { return i.running.Load() }

// Start 加载持久数据并打开流监听。已运行时为无操作。
func (i *Instance) Start(ctx context.Context) error {
	if i.running.Load() {
		return nil
	}
	if err := i.naming.LoadPersistent(ctx, i.cfg.TenantID); err != nil {
		logger.Warn("加载持久服务失败", "instance", i.cfg.InstanceName, "error", err)
		alert.SyncFailure(i.cfg, err)
	}
	i.naming.MarkViewWarm()
	if err := i.server.Start(ctx); err != nil {
		if !appconfig.IsInstanceStopping() {
			_ = i.store.Center.UpdateStatus(ctx, i.cfg.TenantID, i.cfg.InstanceName, i.cfg.Environment,
				model.CenterStatusError, err.Error())
		}
		alert.StartFailure(i.cfg, err)
		return err
	}
	i.evictor.Start()
	i.naming.StartLiveSync(2 * time.Second)
	i.running.Store(true)
	i.cfg.Status = model.CenterStatusRunning
	if !appconfig.IsInstanceStopping() {
		_ = i.store.Center.UpdateStatus(ctx, i.cfg.TenantID, i.cfg.InstanceName, i.cfg.Environment,
			model.CenterStatusRunning, "running")
	}
	logger.Info("servicecenterv3 中心实例已启动", "instance", i.cfg.InstanceName, "listen", i.cfg.ListenEndpoint())
	return nil
}

// Stop 停止 Evictor 与流服务。未运行时为无操作。
func (i *Instance) Stop(ctx context.Context) error {
	if !i.running.Load() {
		return nil
	}
	i.naming.StopLiveSync()
	i.evictor.Stop()
	i.server.Stop()
	i.running.Store(false)
	i.cfg.Status = model.CenterStatusStopped
	// 这一行按租户、实例名、环境共用，没有 Pod 身份。进程退出时不回写，
	// 避免滚动发布里旧副本把新副本的 RUNNING 盖成 STOPPED。
	// 管理面或集群事件主动停止时进程仍在运行，仍会写成 STOPPED。
	if appconfig.IsInstanceStopping() {
		logger.Info("进程停止流程中关闭服务中心，跳过实例状态落库", "instance", i.cfg.InstanceName)
	} else if err := i.store.Center.UpdateStatus(ctx, i.cfg.TenantID, i.cfg.InstanceName, i.cfg.Environment,
		model.CenterStatusStopped, "stopped"); err != nil {
		logger.Error("更新中心实例停止状态失败", "instance", i.cfg.InstanceName, "error", err)
	}
	logger.Info("servicecenterv3 中心实例已停止", "instance", i.cfg.InstanceName)
	return nil
}

// Reload 从库刷新定义。监听/TLS/鉴权/缓冲/健康检查变化时 Stop+Start（会话会断）。
// 仅 ExtProperty 告警开关等非监听字段变化时只换指针，不打断数据面。
func (i *Instance) Reload(ctx context.Context) error {
	fresh, err := i.store.Center.Get(ctx, i.cfg.TenantID, i.cfg.InstanceName, i.cfg.Environment)
	if err != nil {
		return err
	}
	if fresh == nil {
		return contract.ErrCenterNotFound
	}
	needRestart := i.IsRunning() && i.cfg.ListenFingerprint() != fresh.ListenFingerprint()
	i.cfg = fresh
	i.naming.SetCenter(fresh)
	i.config.SetCenter(fresh)
	if !needRestart {
		return nil
	}
	if err := i.Stop(ctx); err != nil {
		return err
	}
	i.rebuildListen()
	return i.Start(ctx)
}

// rebuildListen 按当前 cfg 重建 Evictor 与 stream.Server（不碰 cache/naming 数据）。
func (i *Instance) rebuildListen() {
	interval := time.Duration(i.cfg.HealthCheckInterval) * time.Second
	timeout := time.Duration(i.cfg.HealthCheckTimeout) * time.Second
	i.evictor = naming.NewEvictor(i.naming, interval, timeout)
	i.server = stream.NewServer(i.cfg, i.naming, i.config, i.session, i.store)
}

// Overview 返回当前缓存与连接计数快照。
func (i *Instance) Overview() *model.Overview {
	s, nodes, healthy := i.cache.Counts()
	return &model.Overview{
		CenterInstanceName: i.cfg.InstanceName,
		ServiceCount:       s,
		NodeCount:          nodes,
		HealthyNodes:       healthy,
		ConfigCount:        i.config.CountPublished(context.Background(), i.cfg.TenantID),
		ConnectionCount:    i.session.Count(),
		OwnerGatewayID:     naming.LocalGatewayID,
		ReplicaGatewayIDs:  i.naming.ReplicaGatewayIDs(),
	}
}

// Connections 返回数据面会话副本。
func (i *Instance) Connections() []model.ConnectionInfo {
	return i.session.List()
}

// ListServiceSubscribers 订了该服务的对端服务，叠会话 ClientID/IP。
func (i *Instance) ListServiceSubscribers(namespaceID, group, serviceName string) []model.ServiceSubscriber {
	if i == nil || i.naming == nil {
		return nil
	}
	return decorateSubscribers(i, i.naming.ListSubscribers(namespaceID, group, serviceName), serviceName)
}

// ListServiceSubscriptions 本服务节点所在连接订了哪些服务。
func (i *Instance) ListServiceSubscriptions(namespaceID, group, serviceName string) []model.ServiceSubscriber {
	if i == nil || i.naming == nil {
		return nil
	}
	return decorateSubscribers(i, i.naming.ListSubscriptions(namespaceID, group, serviceName), "")
}

func decorateSubscribers(i *Instance, out []model.ServiceSubscriber, fallbackService string) []model.ServiceSubscriber {
	if len(out) == 0 {
		return nil
	}
	conns := map[string]model.ConnectionInfo{}
	if i != nil && i.session != nil {
		for _, c := range i.session.List() {
			conns[c.ConnectionID] = c
		}
	}
	for idx := range out {
		if out[idx].ServiceName == "" && fallbackService != "" && out[idx].Scope == "service" {
			out[idx].ServiceName = fallbackService
		}
		c, ok := conns[out[idx].ConnectionID]
		if !ok {
			continue
		}
		out[idx].ClientID = c.ClientID
		out[idx].ClientIP = c.ClientIP
		out[idx].LastActive = c.LastActive
		if out[idx].NamespaceID == "" {
			out[idx].NamespaceID = c.NamespaceID
		}
	}
	return out
}

// FindNode 按 nodeID 读取本中心缓存中的业务节点拷贝。
func (i *Instance) FindNode(nodeID string) (*model.Node, bool) {
	return i.cache.GetNode(nodeID)
}

// attachLiveView 集群默认 Redis 可用时挂上注册发现活视图；单机或未配 Redis 仍走本机 Cache。
func attachLiveView(n *naming.App) {
	c := pkgcache.GetDefaultCache()
	if c == nil || c.GetCacheType() != "redis" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := c.Ping(ctx); err != nil {
		logger.Warn("服务注册发现未挂 Redis 活视图", "error", err)
		return
	}
	n.SetLive(live.New(c))
	logger.Info("服务注册发现已挂 Redis 活视图")
}
