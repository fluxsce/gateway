package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gateway/internal/servicecenter/types"
	pkgcache "gateway/pkg/cache"
	"gateway/pkg/logger"
)

// RedisServiceCache 基于纯 Redis 的服务缓存实现
// 作为 ServiceCache 的完全替代实现，所有数据存储在 Redis
//
// 特性：
// - 纯 Redis 存储：所有数据存储在 Redis，不使用本地缓存
// - 分布式共享：多节点自动共享服务注册信息
// - 数据持久化：Redis 支持持久化，支持重启恢复
// - 集群友好：适合服务中心集群部署
// - 键列表索引：使用 Set 维护服务键列表，便于快速遍历和健康检查
//
// 适用场景：
// - 服务中心集群部署
// - 需要跨节点共享服务注册信息
// - 需要数据持久化和故障恢复
// - 需要高效的健康检查和批量操作
//
// 过期策略：
// - 命名空间：永不过期（配置数据，由前端管理）
// - 服务：24小时过期（通过心跳刷新，无节点的服务由健康检查清理）
// - 节点：24小时过期（通过心跳刷新，失效节点由健康检查清理）
type RedisServiceCache struct {
	// Redis 缓存后端
	redisCache pkgcache.Cache

	// 缓存配置
	servicePrefix   string        // Redis 键前缀：service_center:service:
	namespacePrefix string        // Redis 键前缀：service_center:namespace:
	nodePrefix      string        // Redis 键前缀：service_center:node:
	serviceSetKey   string        // 服务键集合：service_center:services:all
	namespaceSetKey string        // 命名空间键集合：service_center:namespaces:all
	nodeSetKey      string        // 节点键集合：service_center:nodes:all
	cacheTTL        time.Duration // 缓存过期时间（仅用于服务和节点，命名空间永不过期）
}

// NewRedisServiceCache 创建基于 Redis 的服务缓存
//
// 参数：
//   - redisCache: Redis 缓存实例
//   - ttl: 缓存过期时间（0 表示永不过期）
//
// 返回：
//   - *RedisServiceCache: Redis 服务缓存实例
//   - error: 初始化错误
func NewRedisServiceCache(redisCache pkgcache.Cache, ttl time.Duration) (*RedisServiceCache, error) {
	if redisCache == nil {
		return nil, fmt.Errorf("redis cache is nil")
	}

	// 测试 Redis 连接
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := redisCache.Ping(ctx); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	rsc := &RedisServiceCache{
		redisCache:      redisCache,
		servicePrefix:   "service_center:service:",
		namespacePrefix: "service_center:namespace:",
		nodePrefix:      "service_center:node:",
		serviceSetKey:   "service_center:services:all",
		namespaceSetKey: "service_center:namespaces:all",
		nodeSetKey:      "service_center:nodes:all",
		cacheTTL:        ttl,
	}

	logger.Info("Redis 服务缓存初始化成功",
		"cacheType", redisCache.GetCacheType(),
		"ttl", ttl,
		"enableKeyIndex", true)

	return rsc, nil
}

// ========== 服务操作 ==========

// GetService 获取服务信息（包含节点）
// 直接从 Redis 读取
func (r *RedisServiceCache) GetService(ctx context.Context, tenantId, namespaceId, groupName, serviceName string) (*types.Service, bool) {
	key := r.serviceKey(tenantId, namespaceId, groupName, serviceName)
	redisKey := r.servicePrefix + key

	data, err := r.redisCache.Get(ctx, redisKey)
	if err != nil || data == nil {
		return nil, false
	}

	// 反序列化
	var service types.Service
	if err := r.unmarshalData(data, &service); err != nil {
		logger.Warn("反序列化服务数据失败", "error", err, "key", key)
		return nil, false
	}

	return &service, true
}

// SetService 设置服务信息（保留节点列表）
//
// 注意：该方法会保留现有的节点列表，只更新服务的元数据信息
// 与 ServiceCache 的行为保持一致
//
// 直接写入 Redis
func (r *RedisServiceCache) SetService(ctx context.Context, service *types.Service) {
	if service == nil {
		return
	}

	key := r.serviceKey(service.TenantId, service.NamespaceId, service.GroupName, service.ServiceName)

	// 先获取现有服务，保留节点列表
	existingService, ok := r.GetService(ctx, service.TenantId, service.NamespaceId, service.GroupName, service.ServiceName)
	if ok {
		// 更新服务信息，保留节点列表
		service.Nodes = existingService.Nodes
	} else {
		// 新服务，确保节点列表不为 nil
		if service.Nodes == nil {
			service.Nodes = []*types.ServiceNode{}
		}
	}

	// 更新时间戳
	service.EditTime = time.Now()

	// 写入 Redis
	redisKey := r.servicePrefix + key
	jsonData, err := json.Marshal(service)
	if err != nil {
		logger.Warn("序列化服务数据失败", "error", err, "serviceName", service.ServiceName)
		return
	}

	if err := r.redisCache.Set(ctx, redisKey, jsonData, r.cacheTTL); err != nil {
		logger.Warn("写入 Redis 失败", "error", err, "serviceName", service.ServiceName)
		return
	}

	// 添加服务键到索引集合
	if _, err := r.redisCache.SAdd(ctx, r.serviceSetKey, key); err != nil {
		logger.Warn("添加服务键到索引集合失败", "error", err, "serviceName", service.ServiceName)
	}
}

// DeleteService 删除服务
// 直接从 Redis 删除
func (r *RedisServiceCache) DeleteService(ctx context.Context, tenantId, namespaceId, groupName, serviceName string) {
	key := r.serviceKey(tenantId, namespaceId, groupName, serviceName)

	// 先获取服务，删除相关节点
	if service, ok := r.GetService(ctx, tenantId, namespaceId, groupName, serviceName); ok {
		for _, node := range service.Nodes {
			r.deleteNodeFromRedis(ctx, node.NodeId)
		}
	}

	// 从 Redis 删除
	redisKey := r.servicePrefix + key
	if err := r.redisCache.Delete(ctx, redisKey); err != nil {
		logger.Warn("从 Redis 删除服务失败", "error", err, "serviceName", serviceName)
	}

	// 从索引集合中移除服务键
	if _, err := r.redisCache.SRem(ctx, r.serviceSetKey, key); err != nil {
		logger.Warn("从索引集合移除服务键失败", "error", err, "serviceName", serviceName)
	}
}

// ========== 节点操作 ==========

// AddNode 添加或更新节点
// 同时更新服务对象和节点索引到 Redis
func (r *RedisServiceCache) AddNode(ctx context.Context, node *types.ServiceNode) {
	if node == nil {
		return
	}

	// 1. 保存节点索引到 Redis
	r.saveNodeToRedis(ctx, node)

	// 2. 获取或创建服务
	service, ok := r.GetService(ctx, node.TenantId, node.NamespaceId, node.GroupName, node.ServiceName)
	if !ok {
		// 服务不存在，创建服务
		now := time.Now()
		service = &types.Service{
			TenantId:           node.TenantId,
			NamespaceId:        node.NamespaceId,
			GroupName:          node.GroupName,
			ServiceName:        node.ServiceName,
			ServiceType:        types.ServiceTypeInternal,
			ServiceVersion:     "",
			ServiceDescription: "",
			MetadataJson:       "",
			TagsJson:           "",
			ProtectThreshold:   0.0,
			SelectorJson:       "",
			AddTime:            now,
			AddWho:             "",
			EditTime:           now,
			EditWho:            "",
			OprSeqFlag:         "",
			CurrentVersion:     1,
			ActiveFlag:         "Y",
			NoteText:           "",
			ExtProperty:        "",
			Nodes:              []*types.ServiceNode{node},
		}
	} else {
		// 服务已存在，更新节点列表
		found := false
		for i, n := range service.Nodes {
			if n.NodeId == node.NodeId {
				service.Nodes[i] = node
				found = true
				break
			}
		}
		if !found {
			service.Nodes = append(service.Nodes, node)
		}
		service.EditTime = time.Now()
	}

	// 3. 直接序列化并保存服务到 Redis（不调用 SetService，避免节点列表被覆盖）
	key := r.serviceKey(service.TenantId, service.NamespaceId, service.GroupName, service.ServiceName)
	redisKey := r.servicePrefix + key

	jsonData, err := json.Marshal(service)
	if err != nil {
		logger.Warn("序列化服务数据失败", "error", err, "serviceName", service.ServiceName)
		return
	}

	if err := r.redisCache.Set(ctx, redisKey, jsonData, r.cacheTTL); err != nil {
		logger.Warn("写入 Redis 失败", "error", err, "serviceName", service.ServiceName)
		return
	}

	// 添加服务键到索引集合
	if _, err := r.redisCache.SAdd(ctx, r.serviceSetKey, key); err != nil {
		logger.Warn("添加服务键到索引集合失败", "error", err, "serviceName", service.ServiceName)
	}
}

// RemoveNode 移除节点
func (r *RedisServiceCache) RemoveNode(ctx context.Context, tenantId, namespaceId, groupName, serviceName, nodeId string) {
	// 1. 删除节点索引
	r.deleteNodeFromRedis(ctx, nodeId)

	// 2. 从服务中移除节点
	service, ok := r.GetService(ctx, tenantId, namespaceId, groupName, serviceName)
	if !ok {
		return
	}

	// 移除指定节点
	newNodes := make([]*types.ServiceNode, 0, len(service.Nodes))
	for _, n := range service.Nodes {
		if n.NodeId != nodeId {
			newNodes = append(newNodes, n)
		}
	}

	service.Nodes = newNodes
	service.EditTime = time.Now()

	// 3. 直接序列化并保存服务到 Redis（不调用 SetService，避免节点列表被覆盖）
	key := r.serviceKey(service.TenantId, service.NamespaceId, service.GroupName, service.ServiceName)
	redisKey := r.servicePrefix + key

	jsonData, err := json.Marshal(service)
	if err != nil {
		logger.Warn("序列化服务数据失败", "error", err, "serviceName", serviceName)
		return
	}

	if err := r.redisCache.Set(ctx, redisKey, jsonData, r.cacheTTL); err != nil {
		logger.Warn("写入 Redis 失败", "error", err, "serviceName", serviceName)
	}
}

// GetNode 通过 nodeId 获取节点
func (r *RedisServiceCache) GetNode(ctx context.Context, tenantId, nodeId string) (*types.ServiceNode, bool) {
	nodeKey := r.nodePrefix + nodeId
	data, err := r.redisCache.Get(ctx, nodeKey)
	if err != nil || data == nil {
		return nil, false
	}

	var node types.ServiceNode
	if err := r.unmarshalData(data, &node); err != nil {
		logger.Warn("反序列化节点数据失败", "error", err, "nodeId", nodeId)
		return nil, false
	}

	// 验证租户 ID
	if tenantId != "" && node.TenantId != tenantId {
		return nil, false
	}

	return &node, true
}

// ========== 命名空间操作 ==========

// GetNamespace 获取命名空间
func (r *RedisServiceCache) GetNamespace(ctx context.Context, tenantId, namespaceId string) (*types.Namespace, bool) {
	key := r.namespaceKey(tenantId, namespaceId)
	redisKey := r.namespacePrefix + key

	data, err := r.redisCache.Get(ctx, redisKey)
	if err != nil || data == nil {
		return nil, false
	}

	var namespace types.Namespace
	if err := r.unmarshalData(data, &namespace); err != nil {
		logger.Warn("反序列化命名空间数据失败", "error", err, "key", key)
		return nil, false
	}

	return &namespace, true
}

// SetNamespace 设置命名空间
func (r *RedisServiceCache) SetNamespace(ctx context.Context, namespace *types.Namespace) {
	if namespace == nil {
		return
	}

	namespace.EditTime = time.Now()

	key := r.namespaceKey(namespace.TenantId, namespace.NamespaceId)
	redisKey := r.namespacePrefix + key

	jsonData, err := json.Marshal(namespace)
	if err != nil {
		logger.Warn("序列化命名空间数据失败", "error", err, "namespaceId", namespace.NamespaceId)
		return
	}

	// 命名空间是配置数据，不应该过期（由前端操作管理）
	if err := r.redisCache.Set(ctx, redisKey, jsonData, 0); err != nil {
		logger.Warn("写入命名空间到 Redis 失败", "error", err, "namespaceId", namespace.NamespaceId)
		return
	}

	// 添加命名空间键到索引集合
	if _, err := r.redisCache.SAdd(ctx, r.namespaceSetKey, key); err != nil {
		logger.Warn("添加命名空间键到索引集合失败", "error", err, "namespaceId", namespace.NamespaceId)
	}
}

// DeleteNamespace 删除命名空间及其下的所有服务
func (r *RedisServiceCache) DeleteNamespace(ctx context.Context, tenantId, namespaceId string) {
	key := r.namespaceKey(tenantId, namespaceId)
	redisKey := r.namespacePrefix + key

	// 删除命名空间
	if err := r.redisCache.Delete(ctx, redisKey); err != nil {
		logger.Warn("从 Redis 删除命名空间失败", "error", err, "namespaceId", namespaceId)
	}

	// 从索引集合中移除命名空间键
	if _, err := r.redisCache.SRem(ctx, r.namespaceSetKey, key); err != nil {
		logger.Warn("从索引集合移除命名空间键失败", "error", err, "namespaceId", namespaceId)
	}

	// 删除该命名空间下的所有服务
	serviceKeys, err := r.redisCache.SMembers(ctx, r.serviceSetKey)
	if err == nil {
		nsPrefix := r.namespaceKey(tenantId, namespaceId) + ":"
		for _, serviceKey := range serviceKeys {
			if len(serviceKey) >= len(nsPrefix) && serviceKey[:len(nsPrefix)] == nsPrefix {
				// 删除服务数据
				redisKey := r.servicePrefix + serviceKey
				_ = r.redisCache.Delete(ctx, redisKey)
				// 从索引集合移除
				_, _ = r.redisCache.SRem(ctx, r.serviceSetKey, serviceKey)
			}
		}
	}
}

// ========== 批量操作 ==========

// GetNodes 获取节点列表
func (r *RedisServiceCache) GetNodes(ctx context.Context, tenantId, namespaceId, groupName, serviceName string) ([]*types.ServiceNode, bool) {
	service, ok := r.GetService(ctx, tenantId, namespaceId, groupName, serviceName)
	if !ok {
		return nil, false
	}
	return service.Nodes, true
}

// SetNodes 设置节点列表（完整替换）
func (r *RedisServiceCache) SetNodes(ctx context.Context, tenantId, namespaceId, groupName, serviceName string, nodes []*types.ServiceNode) {
	service, ok := r.GetService(ctx, tenantId, namespaceId, groupName, serviceName)
	if !ok {
		// 服务不存在，创建服务
		now := time.Now()
		service = &types.Service{
			TenantId:           tenantId,
			NamespaceId:        namespaceId,
			GroupName:          groupName,
			ServiceName:        serviceName,
			ServiceType:        types.ServiceTypeInternal,
			ServiceVersion:     "",
			ServiceDescription: "",
			MetadataJson:       "",
			TagsJson:           "",
			ProtectThreshold:   0.0,
			SelectorJson:       "",
			AddTime:            now,
			AddWho:             "",
			EditTime:           now,
			EditWho:            "",
			OprSeqFlag:         "",
			CurrentVersion:     1,
			ActiveFlag:         "Y",
			NoteText:           "",
			ExtProperty:        "",
			Nodes:              nodes,
		}
	} else {
		// 服务已存在，先删除旧节点的索引
		for _, oldNode := range service.Nodes {
			r.deleteNodeFromRedis(ctx, oldNode.NodeId)
		}

		service.Nodes = nodes
		service.EditTime = time.Now()
	}

	// 保存所有新节点到 Redis
	for _, node := range nodes {
		r.saveNodeToRedis(ctx, node)
	}

	// 直接序列化并保存服务到 Redis（不调用 SetService，避免节点列表被覆盖）
	key := r.serviceKey(service.TenantId, service.NamespaceId, service.GroupName, service.ServiceName)
	redisKey := r.servicePrefix + key

	jsonData, err := json.Marshal(service)
	if err != nil {
		logger.Warn("序列化服务数据失败", "error", err, "serviceName", serviceName)
		return
	}

	if err := r.redisCache.Set(ctx, redisKey, jsonData, r.cacheTTL); err != nil {
		logger.Warn("写入 Redis 失败", "error", err, "serviceName", serviceName)
		return
	}

	// 添加服务键到索引集合
	if _, err := r.redisCache.SAdd(ctx, r.serviceSetKey, key); err != nil {
		logger.Warn("添加服务键到索引集合失败", "error", err, "serviceName", serviceName)
	}
}

// DeleteNodes 删除所有节点（保留服务）
func (r *RedisServiceCache) DeleteNodes(ctx context.Context, tenantId, namespaceId, groupName, serviceName string) {
	r.SetNodes(ctx, tenantId, namespaceId, groupName, serviceName, []*types.ServiceNode{})
}

// UpdateNode 更新单个节点（增量更新）
func (r *RedisServiceCache) UpdateNode(ctx context.Context, node *types.ServiceNode) {
	r.AddNode(ctx, node)
}

// GetServiceWithNodes 获取服务及其节点（原子操作）
func (r *RedisServiceCache) GetServiceWithNodes(ctx context.Context, tenantId, namespaceId, groupName, serviceName string) (*types.Service, []*types.ServiceNode, bool) {
	service, ok := r.GetService(ctx, tenantId, namespaceId, groupName, serviceName)
	if !ok {
		return nil, nil, false
	}
	return service, service.Nodes, true
}

// Clear 清空所有缓存
func (r *RedisServiceCache) Clear(ctx context.Context) {
	// 1. 从索引集合获取所有服务键
	serviceKeys, err := r.redisCache.SMembers(ctx, r.serviceSetKey)
	if err == nil {
		for _, key := range serviceKeys {
			redisKey := r.servicePrefix + key
			_ = r.redisCache.Delete(ctx, redisKey)
		}
	}

	// 2. 从索引集合获取所有命名空间键
	namespaceKeys, err := r.redisCache.SMembers(ctx, r.namespaceSetKey)
	if err == nil {
		for _, key := range namespaceKeys {
			redisKey := r.namespacePrefix + key
			_ = r.redisCache.Delete(ctx, redisKey)
		}
	}

	// 3. 从索引集合获取所有节点键
	nodeKeys, err := r.redisCache.SMembers(ctx, r.nodeSetKey)
	if err == nil {
		for _, nodeId := range nodeKeys {
			redisKey := r.nodePrefix + nodeId
			_ = r.redisCache.Delete(ctx, redisKey)
		}
	}

	// 4. 清空索引集合本身
	_ = r.redisCache.Delete(ctx, r.serviceSetKey)
	_ = r.redisCache.Delete(ctx, r.namespaceSetKey)
	_ = r.redisCache.Delete(ctx, r.nodeSetKey)
}

// Close 关闭缓存
func (r *RedisServiceCache) Close() error {
	// Redis 缓存由外部管理，这里不关闭
	return nil
}

// GetServiceCount 获取服务数量（通过索引集合）
func (r *RedisServiceCache) GetServiceCount() int {
	// 直接获取索引集合的大小
	members, err := r.redisCache.SMembers(context.Background(), r.serviceSetKey)
	if err != nil {
		return 0
	}
	return len(members)
}

// GetNodeCount 获取某个服务的节点数量
func (r *RedisServiceCache) GetNodeCount(ctx context.Context, tenantId, namespaceId, groupName, serviceName string) int {
	nodes, found := r.GetNodes(ctx, tenantId, namespaceId, groupName, serviceName)
	if !found {
		return 0
	}
	return len(nodes)
}

// GetAllServices 遍历所有服务（用于同步等操作）
func (r *RedisServiceCache) GetAllServices(fn func(*types.Service)) {
	ctx := context.Background()

	// 从索引集合获取所有服务键
	serviceKeys, err := r.redisCache.SMembers(ctx, r.serviceSetKey)
	if err != nil {
		logger.Warn("获取服务键集合失败", "error", err)
		return
	}

	// 遍历服务键，获取服务数据
	for _, key := range serviceKeys {
		redisKey := r.servicePrefix + key
		data, err := r.redisCache.Get(ctx, redisKey)
		if err != nil || data == nil {
			continue
		}

		var service types.Service
		if err := r.unmarshalData(data, &service); err != nil {
			logger.Warn("反序列化服务数据失败", "error", err, "key", key)
			continue
		}

		fn(&service)
	}
}

// ========== 辅助方法 ==========

// serviceKey 生成服务缓存键
func (r *RedisServiceCache) serviceKey(tenantId, namespaceId, groupName, serviceName string) string {
	return fmt.Sprintf("%s:%s:%s:%s", tenantId, namespaceId, groupName, serviceName)
}

// namespaceKey 生成命名空间缓存键
func (r *RedisServiceCache) namespaceKey(tenantId, namespaceId string) string {
	return fmt.Sprintf("%s:%s", tenantId, namespaceId)
}

// saveNodeToRedis 保存节点到 Redis
func (r *RedisServiceCache) saveNodeToRedis(ctx context.Context, node *types.ServiceNode) {
	if node == nil {
		return
	}

	nodeKey := r.nodePrefix + node.NodeId
	jsonData, err := json.Marshal(node)
	if err != nil {
		logger.Warn("序列化节点数据失败", "error", err, "nodeId", node.NodeId)
		return
	}

	if err := r.redisCache.Set(ctx, nodeKey, jsonData, r.cacheTTL); err != nil {
		logger.Warn("保存节点到 Redis 失败", "error", err, "nodeId", node.NodeId)
		return
	}

	// 添加节点键到索引集合
	if _, err := r.redisCache.SAdd(ctx, r.nodeSetKey, node.NodeId); err != nil {
		logger.Warn("添加节点键到索引集合失败", "error", err, "nodeId", node.NodeId)
	}
}

// deleteNodeFromRedis 从 Redis 删除节点
func (r *RedisServiceCache) deleteNodeFromRedis(ctx context.Context, nodeId string) {
	nodeKey := r.nodePrefix + nodeId
	if err := r.redisCache.Delete(ctx, nodeKey); err != nil {
		logger.Warn("从 Redis 删除节点失败", "error", err, "nodeId", nodeId)
	}

	// 从索引集合中移除节点键
	if _, err := r.redisCache.SRem(ctx, r.nodeSetKey, nodeId); err != nil {
		logger.Warn("从索引集合移除节点键失败", "error", err, "nodeId", nodeId)
	}
}

// unmarshalData 反序列化数据（支持 string 和 []byte）
func (r *RedisServiceCache) unmarshalData(data interface{}, v interface{}) error {
	var jsonData []byte

	switch d := data.(type) {
	case string:
		jsonData = []byte(d)
	case []byte:
		jsonData = d
	default:
		return fmt.Errorf("unsupported data type: %T", data)
	}

	return json.Unmarshal(jsonData, v)
}
