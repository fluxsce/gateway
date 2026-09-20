package cache

import (
	"sync"
	"time"

	"gateway/internal/servicecenterv3/model"
)

// Cache 按中心实例隔离的内存索引，并发安全。读路径返回拷贝。
// 只缓存命名与命名空间，不缓存配置正文（配置按次读库）。
type Cache struct {
	centerName string
	mu         sync.RWMutex
	services   map[string]*model.Service
	nodes      map[string]*model.Node // nodeID
	namespaces map[string]*model.Namespace
}

// New 构造空缓存。centerName 写入后不可变。
func New(centerName string) *Cache {
	return &Cache{
		centerName: centerName,
		services:   map[string]*model.Service{},
		nodes:      map[string]*model.Node{},
		namespaces: map[string]*model.Namespace{},
	}
}

// CenterName 返回绑定的中心实例名。
func (c *Cache) CenterName() string { return c.centerName }

// serviceKey 生成缓存内服务索引键，分隔符与活视图 Redis 键一致，用冒号。
func serviceKey(namespaceID, group, name string) string {
	return namespaceID + ":" + group + ":" + name
}

// SetNamespace 写入或覆盖命名空间。
func (c *Cache) SetNamespace(ns *model.Namespace) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.namespaces[ns.NamespaceID] = ns
}

// GetNamespace 按 ID 读取命名空间。
func (c *Cache) GetNamespace(namespaceID string) (*model.Namespace, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	ns, ok := c.namespaces[namespaceID]
	return ns, ok
}

// DeleteNamespace 删除命名空间及其下服务、节点索引。
func (c *Cache) DeleteNamespace(namespaceID string) {
	if namespaceID == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.namespaces, namespaceID)
	for key, svc := range c.services {
		if svc == nil || svc.NamespaceID != namespaceID {
			continue
		}
		for _, node := range svc.Nodes {
			if node != nil {
				delete(c.nodes, node.NodeID)
			}
		}
		delete(c.services, key)
	}
	for id, node := range c.nodes {
		if node != nil && node.NamespaceID == namespaceID {
			delete(c.nodes, id)
		}
	}
}

// SetService 写入服务定义；Nodes 为 nil 时保留原节点列表。
func (c *Cache) SetService(svc *model.Service) {
	c.mu.Lock()
	defer c.mu.Unlock()
	key := serviceKey(svc.NamespaceID, svc.GroupName, svc.ServiceName)
	if old, ok := c.services[key]; ok && svc.Nodes == nil {
		svc.Nodes = old.Nodes
	}
	c.services[key] = svc
}

// GetService 返回服务快照（含节点切片拷贝）。
func (c *Cache) GetService(namespaceID, group, name string) (*model.Service, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	svc, ok := c.services[serviceKey(namespaceID, group, name)]
	return cloneService(svc), ok
}

// DeleteService 删除服务及其节点索引。
func (c *Cache) DeleteService(namespaceID, group, name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	key := serviceKey(namespaceID, group, name)
	if svc, ok := c.services[key]; ok {
		for _, node := range svc.Nodes {
			delete(c.nodes, node.NodeID)
		}
	}
	delete(c.services, key)
}

// ListServices 列出命名空间下服务；group 空表示全部组。
func (c *Cache) ListServices(namespaceID, group string) []*model.Service {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]*model.Service, 0)
	for _, svc := range c.services {
		if svc.NamespaceID != namespaceID {
			continue
		}
		if group != "" && svc.GroupName != group {
			continue
		}
		out = append(out, cloneService(svc))
	}
	return out
}

// AllServices 返回全部服务快照，供空壳清理扫描。
func (c *Cache) AllServices() []*model.Service {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]*model.Service, 0, len(c.services))
	for _, svc := range c.services {
		out = append(out, cloneService(svc))
	}
	return out
}

// PutNode 写入或更新节点，并挂到对应服务快照。
func (c *Cache) PutNode(node *model.Node) {
	if node == nil || node.NodeID == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.nodes[node.NodeID] = node
	key := serviceKey(node.NamespaceID, node.GroupName, node.ServiceName)
	svc, ok := c.services[key]
	if !ok {
		svc = &model.Service{
			TenantID:    node.TenantID,
			NamespaceID: node.NamespaceID,
			GroupName:   node.GroupName,
			ServiceName: node.ServiceName,
			AutoCreated: true,
		}
		c.services[key] = svc
	}
	replaced := false
	for i, existing := range svc.Nodes {
		if existing.NodeID == node.NodeID {
			svc.Nodes[i] = node
			replaced = true
			break
		}
	}
	if !replaced {
		svc.Nodes = append(svc.Nodes, node)
	}
}

// GetNode 按 ID 读取节点拷贝。
func (c *Cache) GetNode(nodeID string) (*model.Node, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	node, ok := c.nodes[nodeID]
	if !ok {
		return nil, false
	}
	cp := *node
	return &cp, true
}

// RemoveNode 删除节点并返回被删拷贝。
func (c *Cache) RemoveNode(nodeID string) *model.Node {
	c.mu.Lock()
	defer c.mu.Unlock()
	node, ok := c.nodes[nodeID]
	if !ok {
		return nil
	}
	delete(c.nodes, nodeID)
	key := serviceKey(node.NamespaceID, node.GroupName, node.ServiceName)
	if svc, ok := c.services[key]; ok {
		n := 0
		for _, existing := range svc.Nodes {
			if existing.NodeID != nodeID {
				svc.Nodes[n] = existing
				n++
			}
		}
		for i := n; i < len(svc.Nodes); i++ {
			svc.Nodes[i] = nil
		}
		svc.Nodes = svc.Nodes[:n]
	}
	return node
}

// AllNodes 返回全部节点拷贝，供 Evictor 扫描。
func (c *Cache) AllNodes() []*model.Node {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]*model.Node, 0, len(c.nodes))
	for _, node := range c.nodes {
		cp := *node
		out = append(out, &cp)
	}
	return out
}

// CountServicesInNamespace 返回该命名空间下已缓存的服务数，供配额校验。
func (c *Cache) CountServicesInNamespace(namespaceID string) int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	n := 0
	for _, svc := range c.services {
		if svc.NamespaceID == namespaceID {
			n++
		}
	}
	return n
}

// Counts 返回服务、节点、健康节点计数。配置不进本缓存。
func (c *Cache) Counts() (services, nodes, healthy int) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	services = len(c.services)
	nodes = len(c.nodes)
	for _, node := range c.nodes {
		if node.IsHealthy() {
			healthy++
		}
	}
	return
}

// cloneService 浅拷贝服务及其节点切片，避免调用方改到内部 map。
func cloneService(svc *model.Service) *model.Service {
	if svc == nil {
		return nil
	}
	cp := *svc
	if svc.Nodes != nil {
		cp.Nodes = make([]*model.Node, len(svc.Nodes))
		copy(cp.Nodes, svc.Nodes)
	}
	return &cp
}

// TouchBeat 刷新心跳时间。ownerGatewayID 非空且与当前 Owner 不同时认领所有权。
// 返回 (found, claimed)：节点不存在时 found 为 false；claimed 表示 Owner 刚迁到本网关。
func (c *Cache) TouchBeat(nodeID string, at time.Time, ownerGatewayID string) (found, claimed bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	node, ok := c.nodes[nodeID]
	if !ok {
		return false, false
	}
	node.LastBeatTime = at
	node.HealthyStatus = model.Healthy
	if node.Status == "" {
		node.Status = model.NodeUP
	}
	if ownerGatewayID != "" && node.OwnerGatewayID != ownerGatewayID {
		node.OwnerGatewayID = ownerGatewayID
		return true, true
	}
	return true, false
}
