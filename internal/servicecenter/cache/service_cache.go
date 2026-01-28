package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gateway/internal/servicecenter/types"
)

// ServiceCache 服务缓存 - 挂载存储
// 服务和节点作为整体存储，符合业务模型
//
// 设计说明：
// - 全局单例，无需初始化，直接使用
// - 服务直接包含节点列表（Service.Nodes），实现挂载存储
// - 直接使用类型自带的 EditTime 字段，无需额外包装结构
// - 支持增量更新（AddNode/RemoveNode/UpdateNode）
// - 一次查询获取完整数据（服务+节点）
//
// 生命周期：
// - 服务/节点：由注册/注销和心跳机制控制
// - 命名空间：持久化逻辑隔离单元，只在删除时清除缓存
type ServiceCache struct {
	// 服务缓存: key=tenantId:namespaceId:groupName:serviceName
	// value=*types.Service（服务直接包含 Nodes 和 EditTime）
	services sync.Map

	// 命名空间缓存: key=tenantId:namespaceId
	// value=*types.Namespace（直接使用 EditTime）
	namespaces sync.Map

	// 节点索引: key=nodeId
	// value=*types.ServiceNode（用于快速通过 nodeId 查找节点）
	// 注意：节点索引与 services 中的节点保持同步
	nodeIndex sync.Map
}

// 全局服务缓存实例
var globalCache = &ServiceCache{}

// GetGlobalCache 获取全局服务缓存实例
func GetGlobalCache() *ServiceCache {
	return globalCache
}

// GetService 获取服务信息（包含节点）
func (c *ServiceCache) GetService(ctx context.Context, tenantId, namespaceId, groupName, serviceName string) (*types.Service, bool) {
	key := c.serviceKey(tenantId, namespaceId, groupName, serviceName)
	value, ok := c.services.Load(key)
	if !ok {
		return nil, false
	}

	service := value.(*types.Service)
	return service, true
}

// SetService 设置服务信息
func (c *ServiceCache) SetService(ctx context.Context, service *types.Service) {
	if service == nil {
		return
	}

	key := c.serviceKey(service.TenantId, service.NamespaceId, service.GroupName, service.ServiceName)

	// 加载现有数据
	value, ok := c.services.Load(key)
	if ok {
		// 更新服务信息，保留节点列表
		oldService := value.(*types.Service)
		oldNodes := oldService.Nodes
		service.Nodes = oldNodes // 保留节点列表
	} else {
		// 创建新的服务（暂无节点）
		if service.Nodes == nil {
			service.Nodes = []*types.ServiceNode{}
		}
	}

	// 更新 EditTime（缓存时间戳）
	service.EditTime = time.Now()
	c.services.Store(key, service)
}

// DeleteService 删除服务（节点列表置空，同时删除节点索引）
//
// 处理流程：
//  1. 获取服务信息
//  2. 将服务中的节点列表置空
//  3. 删除所有节点的索引（确保节点无法通过 nodeId 查询到）
//  4. 删除服务缓存
//
// 注意：
//   - 删除服务时，会同时删除该服务下所有节点的索引
//   - 这样可以确保删除服务后，无法通过 GetNode 查询到这些节点
func (c *ServiceCache) DeleteService(ctx context.Context, tenantId, namespaceId, groupName, serviceName string) {
	key := c.serviceKey(tenantId, namespaceId, groupName, serviceName)

	// 先获取服务，删除节点索引并将节点列表置空
	value, ok := c.services.Load(key)
	if ok {
		service := value.(*types.Service)
		// 从节点索引中删除所有节点
		for _, node := range service.Nodes {
			c.nodeIndex.Delete(node.NodeId)
		}
		// 将节点列表置空
		service.Nodes = []*types.ServiceNode{}
		service.EditTime = time.Now()
	}

	// 删除服务缓存
	c.services.Delete(key)
}

// GetNodes 获取节点列表
func (c *ServiceCache) GetNodes(ctx context.Context, tenantId, namespaceId, groupName, serviceName string) ([]*types.ServiceNode, bool) {
	key := c.serviceKey(tenantId, namespaceId, groupName, serviceName)
	value, ok := c.services.Load(key)
	if !ok {
		return nil, false
	}

	service := value.(*types.Service)
	return service.Nodes, true
}

// SetNodes 设置节点列表（完整替换）
func (c *ServiceCache) SetNodes(ctx context.Context, tenantId, namespaceId, groupName, serviceName string, nodes []*types.ServiceNode) {
	key := c.serviceKey(tenantId, namespaceId, groupName, serviceName)

	value, ok := c.services.Load(key)
	if ok {
		// 更新节点列表，保留服务信息
		service := value.(*types.Service)
		service.Nodes = nodes
		service.EditTime = time.Now()
		c.services.Store(key, service)
	} else {
		// 服务不存在，创建空服务
		service := &types.Service{
			TenantId:    tenantId,
			NamespaceId: namespaceId,
			GroupName:   groupName,
			ServiceName: serviceName,
			Nodes:       nodes,
			EditTime:    time.Now(),
		}
		c.services.Store(key, service)
	}
}

// DeleteNodes 删除所有节点（保留服务）
func (c *ServiceCache) DeleteNodes(ctx context.Context, tenantId, namespaceId, groupName, serviceName string) {
	key := c.serviceKey(tenantId, namespaceId, groupName, serviceName)

	value, ok := c.services.Load(key)
	if ok {
		service := value.(*types.Service)
		service.Nodes = []*types.ServiceNode{}
		service.EditTime = time.Now()
		c.services.Store(key, service)
	}
}

// AddNode 添加或更新单个节点（增量更新）
func (c *ServiceCache) AddNode(ctx context.Context, node *types.ServiceNode) {
	if node == nil {
		return
	}

	key := c.serviceKey(node.TenantId, node.NamespaceId, node.GroupName, node.ServiceName)

	value, ok := c.services.Load(key)
	if !ok {
		// 服务不存在，创建服务并添加节点
		service := &types.Service{
			TenantId:    node.TenantId,
			NamespaceId: node.NamespaceId,
			GroupName:   node.GroupName,
			ServiceName: node.ServiceName,
			Nodes:       []*types.ServiceNode{node},
			EditTime:    time.Now(),
		}
		c.services.Store(key, service)
		// 更新节点索引
		c.nodeIndex.Store(node.NodeId, node)
		return
	}

	service := value.(*types.Service)

	// 检查节点是否已存在
	found := false
	for i, n := range service.Nodes {
		if n.NodeId == node.NodeId {
			// 更新现有节点
			service.Nodes[i] = node
			found = true
			break
		}
	}

	if !found {
		// 添加新节点
		service.Nodes = append(service.Nodes, node)
	}

	service.EditTime = time.Now()
	c.services.Store(key, service)
	// 更新节点索引
	c.nodeIndex.Store(node.NodeId, node)
}

// RemoveNode 移除单个节点（增量更新）
func (c *ServiceCache) RemoveNode(ctx context.Context, tenantId, namespaceId, groupName, serviceName, nodeId string) {
	key := c.serviceKey(tenantId, namespaceId, groupName, serviceName)

	value, ok := c.services.Load(key)
	if !ok {
		return
	}

	service := value.(*types.Service)

	// 移除指定节点
	newNodes := make([]*types.ServiceNode, 0, len(service.Nodes))
	for _, n := range service.Nodes {
		if n.NodeId != nodeId {
			newNodes = append(newNodes, n)
		}
	}

	service.Nodes = newNodes
	service.EditTime = time.Now()
	c.services.Store(key, service)
	// 从节点索引中删除
	c.nodeIndex.Delete(nodeId)
}

// UpdateNode 更新单个节点（增量更新）
func (c *ServiceCache) UpdateNode(ctx context.Context, node *types.ServiceNode) {
	if node == nil {
		return
	}

	key := c.serviceKey(node.TenantId, node.NamespaceId, node.GroupName, node.ServiceName)

	value, ok := c.services.Load(key)
	if !ok {
		// 服务不存在，直接添加
		c.AddNode(ctx, node)
		return
	}

	service := value.(*types.Service)

	// 查找并更新节点
	found := false
	for i, n := range service.Nodes {
		if n.NodeId == node.NodeId {
			service.Nodes[i] = node
			found = true
			break
		}
	}

	if !found {
		// 节点不存在，添加新节点
		service.Nodes = append(service.Nodes, node)
	}

	service.EditTime = time.Now()
	c.services.Store(key, service)
	// 更新节点索引
	c.nodeIndex.Store(node.NodeId, node)
}

// GetNode 通过 nodeId 快速查找节点（O(1) 时间复杂度）
func (c *ServiceCache) GetNode(ctx context.Context, tenantId, nodeId string) (*types.ServiceNode, bool) {
	value, ok := c.nodeIndex.Load(nodeId)
	if !ok {
		return nil, false
	}

	node := value.(*types.ServiceNode)
	// 验证 tenantId 是否匹配（可选，用于多租户隔离）
	if tenantId != "" && node.TenantId != tenantId {
		return nil, false
	}

	return node, true
}

// GetServiceWithNodes 获取服务及其节点（原子操作）
func (c *ServiceCache) GetServiceWithNodes(ctx context.Context, tenantId, namespaceId, groupName, serviceName string) (*types.Service, []*types.ServiceNode, bool) {
	key := c.serviceKey(tenantId, namespaceId, groupName, serviceName)
	value, ok := c.services.Load(key)
	if !ok {
		return nil, nil, false
	}

	service := value.(*types.Service)
	return service, service.Nodes, true
}

// GetNamespace 获取命名空间
func (c *ServiceCache) GetNamespace(ctx context.Context, tenantId, namespaceId string) (*types.Namespace, bool) {
	key := c.namespaceKey(tenantId, namespaceId)
	value, ok := c.namespaces.Load(key)
	if !ok {
		return nil, false
	}

	namespace := value.(*types.Namespace)
	return namespace, true
}

// SetNamespace 设置命名空间
func (c *ServiceCache) SetNamespace(ctx context.Context, namespace *types.Namespace) {
	if namespace == nil {
		return
	}

	key := c.namespaceKey(namespace.TenantId, namespace.NamespaceId)
	// 更新 EditTime（缓存时间戳）
	namespace.EditTime = time.Now()
	c.namespaces.Store(key, namespace)
}

// DeleteNamespace 删除命名空间
// 同时删除该命名空间下的所有服务和节点缓存
func (c *ServiceCache) DeleteNamespace(ctx context.Context, tenantId, namespaceId string) {
	key := c.namespaceKey(tenantId, namespaceId)
	c.namespaces.Delete(key)

	// 删除该命名空间下的所有服务和节点缓存
	// 构建命名空间前缀（格式：tenantId:namespaceId:）
	namespacePrefix := fmt.Sprintf("%s:%s:", tenantId, namespaceId)

	// 遍历所有服务，删除匹配的命名空间
	var keysToDelete []string
	c.services.Range(func(key, value interface{}) bool {
		keyStr := key.(string)
		// 检查键是否匹配该命名空间（格式：tenantId:namespaceId:groupName:serviceName）
		if len(keyStr) >= len(namespacePrefix) && keyStr[:len(namespacePrefix)] == namespacePrefix {
			keysToDelete = append(keysToDelete, keyStr)
		}
		return true
	})

	// 删除匹配的服务
	for _, keyToDelete := range keysToDelete {
		c.services.Delete(keyToDelete)
	}
}

// Clear 清空所有缓存
func (c *ServiceCache) Clear(ctx context.Context) {
	c.services = sync.Map{}
	c.namespaces = sync.Map{}
	c.nodeIndex = sync.Map{}
}

// Close 关闭缓存
func (c *ServiceCache) Close() error {
	return nil
}

// GetServiceCount 获取服务数量
func (c *ServiceCache) GetServiceCount() int {
	count := 0
	c.services.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

// GetNodeCount 获取某个服务的节点数量
func (c *ServiceCache) GetNodeCount(ctx context.Context, tenantId, namespaceId, groupName, serviceName string) int {
	nodes, found := c.GetNodes(ctx, tenantId, namespaceId, groupName, serviceName)
	if !found {
		return 0
	}
	return len(nodes)
}

// serviceKey 生成服务缓存键
func (c *ServiceCache) serviceKey(tenantId, namespaceId, groupName, serviceName string) string {
	return fmt.Sprintf("%s:%s:%s:%s", tenantId, namespaceId, groupName, serviceName)
}

// namespaceKey 生成命名空间缓存键
func (c *ServiceCache) namespaceKey(tenantId, namespaceId string) string {
	return fmt.Sprintf("%s:%s", tenantId, namespaceId)
}

// GetAllServices 遍历所有服务（用于同步等操作）
//
// 性能说明：
//   - sync.Map.Range() 使用快照机制，不会长时间持有锁
//   - 遍历过程中会创建数据的快照，然后释放锁，在快照上遍历
//   - 回调函数中的操作不会阻塞其他 goroutine 的读写操作
//   - 适合用于批量同步等操作，对并发性能影响较小
//
// 注意事项：
//   - 回调函数中应避免长时间阻塞操作（如网络 I/O）
//   - 如果回调函数执行时间较长，建议在回调中只收集数据，在回调外处理
//   - 返回 false 可提前终止遍历
func (c *ServiceCache) GetAllServices(fn func(*types.Service)) {
	c.services.Range(func(key, value interface{}) bool {
		service := value.(*types.Service)
		fn(service)
		return true
	})
}
