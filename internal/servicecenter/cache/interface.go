// Package cache 提供服务注册中心的缓存实现
//
// 本包提供了两种缓存实现：
//   - ServiceCache: 纯本地缓存（使用 sync.Map），仅适用于单节点部署
//   - RedisServiceCache: 纯 Redis 缓存，集群部署必须使用此实现
//
// 重要提示：
//   - 单节点部署：可使用本地缓存或 Redis
//   - 集群部署：必须使用 Redis，否则各节点数据不一致
//
// GetGlobalCache() 方法自动根据配置选择实现：
//   - 如果 default 缓存是 Redis → 使用 RedisServiceCache（集群共享）
//   - 否则 → 使用 ServiceCache（单节点本地）
//
// 示例：
//
//	// 自动选择（推荐）
//	cache := cache.GetGlobalCache()
//	service, ok := cache.GetService(ctx, tenantId, namespaceId, groupName, serviceName)
//
//	// 显式使用本地缓存（仅单节点）
//	localCache := cache.GetLocalCache()
//	localCache.SetService(ctx, service)
package cache

import (
	"context"
	"sync"
	"time"

	"gateway/internal/servicecenter/types"
	pkgcache "gateway/pkg/cache"
	"gateway/pkg/logger"
)

// IServiceCache 服务缓存接口
//
// 定义服务注册中心的缓存操作接口，支持服务、节点、命名空间的完整生命周期管理。
//
// 设计特点：
//   - 服务和节点作为整体存储（Service.Nodes），符合业务模型
//   - 支持增量更新（AddNode/RemoveNode/UpdateNode）
//   - 一次查询获取完整数据（服务+节点）
//   - 提供快速节点查找（通过 nodeId，O(1) 时间复杂度）
//
// 实现：
//   - ServiceCache: 本地缓存实现（使用 sync.Map）
//   - RedisServiceCache: Redis 缓存实现（使用 Redis）
//
// 线程安全：
//   - 所有实现都保证线程安全
//   - 支持高并发读写
type IServiceCache interface {
	// ==================== 服务操作 ====================

	// GetService 获取服务信息（包含节点列表）
	//
	// 参数：
	//   - ctx: 上下文
	//   - tenantId: 租户 ID
	//   - namespaceId: 命名空间 ID
	//   - groupName: 服务组名
	//   - serviceName: 服务名
	//
	// 返回：
	//   - *types.Service: 服务信息（包含节点列表）
	//   - bool: 是否找到（true=找到，false=不存在）
	//
	// 示例：
	//
	//	service, ok := cache.GetService(ctx, "default", "public", "group1", "userService")
	//	if ok {
	//	    fmt.Printf("服务节点数: %d\n", len(service.Nodes))
	//	}
	GetService(ctx context.Context, tenantId, namespaceId, groupName, serviceName string) (*types.Service, bool)

	// SetService 设置或更新服务信息
	//
	// 注意：
	//   - 如果服务已存在，会保留现有节点列表（只更新服务元数据）
	//   - 如果服务不存在，会创建新服务（节点列表为空）
	//   - 会自动更新 EditTime 时间戳
	//
	// 参数：
	//   - ctx: 上下文
	//   - service: 服务信息
	//
	// 示例：
	//
	//	service := &types.Service{
	//	    TenantId: "default",
	//	    NamespaceId: "public",
	//	    GroupName: "group1",
	//	    ServiceName: "userService",
	//	    ProtectThreshold: 0.0,
	//	}
	//	cache.SetService(ctx, service)
	SetService(ctx context.Context, service *types.Service)

	// DeleteService 删除服务
	//
	// 注意：
	//   - 会同时删除该服务下的所有节点
	//   - 会同时删除节点索引（确保无法通过 GetNode 查询到这些节点）
	//   - 删除操作是幂等的（重复删除不会报错）
	//
	// 参数：
	//   - ctx: 上下文
	//   - tenantId: 租户 ID
	//   - namespaceId: 命名空间 ID
	//   - groupName: 服务组名
	//   - serviceName: 服务名
	//
	// 示例：
	//
	//	cache.DeleteService(ctx, "default", "public", "group1", "userService")
	DeleteService(ctx context.Context, tenantId, namespaceId, groupName, serviceName string)

	// GetServiceCount 获取服务总数
	//
	// 返回：
	//   - int: 服务总数
	//
	// 示例：
	//
	//	count := cache.GetServiceCount()
	//	fmt.Printf("当前服务数: %d\n", count)
	GetServiceCount() int

	// GetAllServices 遍历所有服务
	//
	// 性能说明：
	//   - sync.Map.Range() 使用快照机制，不会长时间持有锁
	//   - 遍历过程中会创建数据的快照，然后释放锁
	//   - 回调函数中的操作不会阻塞其他 goroutine 的读写操作
	//
	// 注意事项：
	//   - 回调函数中应避免长时间阻塞操作（如网络 I/O）
	//   - 如果回调函数执行时间较长，建议在回调中只收集数据，在回调外处理
	//
	// 参数：
	//   - fn: 回调函数，参数为服务信息
	//
	// 示例：
	//
	//	cache.GetAllServices(func(service *types.Service) {
	//	    fmt.Printf("服务: %s, 节点数: %d\n", service.ServiceName, len(service.Nodes))
	//	})
	GetAllServices(fn func(*types.Service))

	// ==================== 节点操作 ====================

	// GetNodes 获取服务的所有节点
	//
	// 参数：
	//   - ctx: 上下文
	//   - tenantId: 租户 ID
	//   - namespaceId: 命名空间 ID
	//   - groupName: 服务组名
	//   - serviceName: 服务名
	//
	// 返回：
	//   - []*types.ServiceNode: 节点列表
	//   - bool: 是否找到服务（true=找到，false=服务不存在）
	//
	// 示例：
	//
	//	nodes, ok := cache.GetNodes(ctx, "default", "public", "group1", "userService")
	//	if ok {
	//	    for _, node := range nodes {
	//	        fmt.Printf("节点 IP: %s\n", node.IP)
	//	    }
	//	}
	GetNodes(ctx context.Context, tenantId, namespaceId, groupName, serviceName string) ([]*types.ServiceNode, bool)

	// SetNodes 设置节点列表（完整替换）
	//
	// 注意：
	//   - 会完整替换节点列表（不是增量更新）
	//   - 如果服务不存在，会自动创建服务
	//   - 会自动更新 EditTime 时间戳
	//
	// 参数：
	//   - ctx: 上下文
	//   - tenantId: 租户 ID
	//   - namespaceId: 命名空间 ID
	//   - groupName: 服务组名
	//   - serviceName: 服务名
	//   - nodes: 节点列表
	//
	// 示例：
	//
	//	nodes := []*types.ServiceNode{
	//	    {NodeId: "node1", IP: "192.168.1.1", Port: 8080},
	//	    {NodeId: "node2", IP: "192.168.1.2", Port: 8080},
	//	}
	//	cache.SetNodes(ctx, "default", "public", "group1", "userService", nodes)
	SetNodes(ctx context.Context, tenantId, namespaceId, groupName, serviceName string, nodes []*types.ServiceNode)

	// DeleteNodes 删除所有节点（保留服务）
	//
	// 注意：
	//   - 只删除节点列表，不删除服务本身
	//   - 删除后服务的节点列表为空
	//
	// 参数：
	//   - ctx: 上下文
	//   - tenantId: 租户 ID
	//   - namespaceId: 命名空间 ID
	//   - groupName: 服务组名
	//   - serviceName: 服务名
	//
	// 示例：
	//
	//	cache.DeleteNodes(ctx, "default", "public", "group1", "userService")
	DeleteNodes(ctx context.Context, tenantId, namespaceId, groupName, serviceName string)

	// AddNode 添加或更新单个节点（增量更新）
	//
	// 注意：
	//   - 如果节点已存在（相同 NodeId），会更新节点信息
	//   - 如果节点不存在，会添加新节点
	//   - 如果服务不存在，会自动创建服务
	//   - 会自动更新节点索引（支持通过 nodeId 快速查找）
	//
	// 参数：
	//   - ctx: 上下文
	//   - node: 节点信息
	//
	// 示例：
	//
	//	node := &types.ServiceNode{
	//	    TenantId: "default",
	//	    NamespaceId: "public",
	//	    GroupName: "group1",
	//	    ServiceName: "userService",
	//	    NodeId: "node1",
	//	    IP: "192.168.1.1",
	//	    Port: 8080,
	//	}
	//	cache.AddNode(ctx, node)
	AddNode(ctx context.Context, node *types.ServiceNode)

	// RemoveNode 移除单个节点（增量更新）
	//
	// 注意：
	//   - 只移除指定节点，不影响其他节点
	//   - 会同时删除节点索引
	//   - 移除操作是幂等的（重复移除不会报错）
	//
	// 参数：
	//   - ctx: 上下文
	//   - tenantId: 租户 ID
	//   - namespaceId: 命名空间 ID
	//   - groupName: 服务组名
	//   - serviceName: 服务名
	//   - nodeId: 节点 ID
	//
	// 示例：
	//
	//	cache.RemoveNode(ctx, "default", "public", "group1", "userService", "node1")
	RemoveNode(ctx context.Context, tenantId, namespaceId, groupName, serviceName, nodeId string)

	// UpdateNode 更新单个节点（增量更新）
	//
	// 注意：
	//   - 如果节点已存在（相同 NodeId），会更新节点信息
	//   - 如果节点不存在，会添加新节点
	//   - 如果服务不存在，会自动创建服务
	//
	// 参数：
	//   - ctx: 上下文
	//   - node: 节点信息
	//
	// 示例：
	//
	//	node := &types.ServiceNode{
	//	    NodeId: "node1",
	//	    IP: "192.168.1.1",
	//	    Port: 9090,  // 更新端口
	//	    Weight: 200,  // 更新权重
	//	}
	//	cache.UpdateNode(ctx, node)
	UpdateNode(ctx context.Context, node *types.ServiceNode)

	// GetNode 通过 nodeId 快速查找节点（O(1) 时间复杂度）
	//
	// 注意：
	//   - 使用节点索引进行查找，性能最优
	//   - tenantId 用于多租户隔离（可选，空字符串表示不校验）
	//
	// 参数：
	//   - ctx: 上下文
	//   - tenantId: 租户 ID（可选，空字符串表示不校验）
	//   - nodeId: 节点 ID
	//
	// 返回：
	//   - *types.ServiceNode: 节点信息
	//   - bool: 是否找到（true=找到，false=不存在或租户不匹配）
	//
	// 示例：
	//
	//	node, ok := cache.GetNode(ctx, "default", "node1")
	//	if ok {
	//	    fmt.Printf("节点 IP: %s, Port: %d\n", node.IP, node.Port)
	//	}
	GetNode(ctx context.Context, tenantId, nodeId string) (*types.ServiceNode, bool)

	// GetNodeCount 获取某个服务的节点数量
	//
	// 参数：
	//   - ctx: 上下文
	//   - tenantId: 租户 ID
	//   - namespaceId: 命名空间 ID
	//   - groupName: 服务组名
	//   - serviceName: 服务名
	//
	// 返回：
	//   - int: 节点数量（如果服务不存在，返回 0）
	//
	// 示例：
	//
	//	count := cache.GetNodeCount(ctx, "default", "public", "group1", "userService")
	//	fmt.Printf("节点数量: %d\n", count)
	GetNodeCount(ctx context.Context, tenantId, namespaceId, groupName, serviceName string) int

	// ==================== 命名空间操作 ====================

	// GetNamespace 获取命名空间信息
	//
	// 参数：
	//   - ctx: 上下文
	//   - tenantId: 租户 ID
	//   - namespaceId: 命名空间 ID
	//
	// 返回：
	//   - *types.Namespace: 命名空间信息
	//   - bool: 是否找到（true=找到，false=不存在）
	//
	// 示例：
	//
	//	namespace, ok := cache.GetNamespace(ctx, "default", "public")
	//	if ok {
	//	    fmt.Printf("命名空间名称: %s\n", namespace.NamespaceName)
	//	}
	GetNamespace(ctx context.Context, tenantId, namespaceId string) (*types.Namespace, bool)

	// SetNamespace 设置或更新命名空间
	//
	// 注意：
	//   - 会自动更新 EditTime 时间戳
	//
	// 参数：
	//   - ctx: 上下文
	//   - namespace: 命名空间信息
	//
	// 示例：
	//
	//	namespace := &types.Namespace{
	//	    TenantId: "default",
	//	    NamespaceId: "public",
	//	    NamespaceName: "公共命名空间",
	//	}
	//	cache.SetNamespace(ctx, namespace)
	SetNamespace(ctx context.Context, namespace *types.Namespace)

	// DeleteNamespace 删除命名空间
	//
	// 注意：
	//   - 会同时删除该命名空间下的所有服务和节点
	//   - 删除操作是幂等的（重复删除不会报错）
	//
	// 参数：
	//   - ctx: 上下文
	//   - tenantId: 租户 ID
	//   - namespaceId: 命名空间 ID
	//
	// 示例：
	//
	//	cache.DeleteNamespace(ctx, "default", "public")
	DeleteNamespace(ctx context.Context, tenantId, namespaceId string)

	// ==================== 其他操作 ====================

	// GetServiceWithNodes 获取服务及其节点（原子操作）
	//
	// 注意：
	//   - 一次操作获取服务和节点，保证数据一致性
	//
	// 参数：
	//   - ctx: 上下文
	//   - tenantId: 租户 ID
	//   - namespaceId: 命名空间 ID
	//   - groupName: 服务组名
	//   - serviceName: 服务名
	//
	// 返回：
	//   - *types.Service: 服务信息
	//   - []*types.ServiceNode: 节点列表
	//   - bool: 是否找到（true=找到，false=不存在）
	//
	// 示例：
	//
	//	service, nodes, ok := cache.GetServiceWithNodes(ctx, "default", "public", "group1", "userService")
	//	if ok {
	//	    fmt.Printf("服务: %s, 节点数: %d\n", service.ServiceName, len(nodes))
	//	}
	GetServiceWithNodes(ctx context.Context, tenantId, namespaceId, groupName, serviceName string) (*types.Service, []*types.ServiceNode, bool)

	// Clear 清空所有缓存
	//
	// 注意：
	//   - 会删除所有服务、节点、命名空间
	//   - 此操作不可逆
	//
	// 参数：
	//   - ctx: 上下文
	//
	// 示例：
	//
	//	cache.Clear(ctx)
	Clear(ctx context.Context)

	// Close 关闭缓存
	//
	// 注意：
	//   - 对于 ServiceCache（本地缓存），此方法无操作
	//   - 对于 RedisServiceCache，会关闭 Redis 连接
	//   - 关闭后不应再使用缓存
	//
	// 返回：
	//   - error: 错误信息（如果有）
	//
	// 示例：
	//
	//	if err := cache.Close(); err != nil {
	//	    log.Printf("关闭缓存失败: %v", err)
	//	}
	Close() error
}

var (
	// globalSmartCache 全局智能缓存实例
	// 通过 GetGlobalSmartCache() 初始化，单例模式
	globalSmartCache IServiceCache

	// smartCacheOnce 确保全局智能缓存只初始化一次
	smartCacheOnce sync.Once
)

// GetGlobalSmartCache 获取全局智能缓存实例
//
// 自动根据 default 缓存配置选择实现：
//   - 如果 default 缓存是 Redis → 返回 RedisServiceCache（集群共享）
//   - 否则 → 返回 ServiceCache（单节点本地）
//
// 重要提示：
//   - 单节点部署：可使用本地缓存或 Redis
//   - 集群部署：必须配置 Redis 作为 default 缓存，否则各节点数据不一致
//
// 特点：
//   - 单例模式：全局只初始化一次，后续调用直接返回已创建的实例
//   - 线程安全：使用 sync.Once 保证并发安全
//   - 自动降级：Redis 连接失败时自动降级为本地缓存（注意：集群环境下降级会导致数据不一致）
//   - 配置驱动：通过 pkgcache.GetDefaultCache() 自动选择
//
// 返回：
//   - IServiceCache: 服务缓存实例（RedisServiceCache 或 ServiceCache）
//
// 启动日志：
//   - 使用 Redis: "服务注册中心使用 Redis 缓存, cacheType=redis"
//   - 使用本地: "服务注册中心使用本地缓存"
func GetGlobalSmartCache() IServiceCache {
	smartCacheOnce.Do(func() {
		globalSmartCache = newServiceCacheFromDefault()
	})
	return globalSmartCache
}

// newServiceCacheFromDefault 根据 default 缓存创建服务缓存（内部方法）
//
// 选择逻辑：
//  1. 获取 pkgcache.GetDefaultCache()
//  2. 检查缓存类型是否为 "redis"
//  3. 如果是 Redis 且连接成功 → 返回 RedisServiceCache
//  4. 如果 Redis 连接失败或不是 Redis → 返回 ServiceCache
//
// 降级策略：
//   - Redis 连接失败时，记录警告日志，降级为本地缓存
//   - 保证服务启动不受 Redis 故障影响
//   - 注意：集群部署时如果降级为本地缓存，会导致各节点数据不一致
//
// 返回：
//   - IServiceCache: 服务缓存实例
func newServiceCacheFromDefault() IServiceCache {
	// 获取 default 缓存
	defaultCache := pkgcache.GetDefaultCache()

	if defaultCache != nil {
		// 检查缓存类型
		cacheType := defaultCache.GetCacheType()
		if cacheType == "redis" {
			// 使用 Redis 实现
			redisCache, err := NewRedisServiceCache(defaultCache, 24*time.Hour)
			if err != nil {
				logger.Warn("创建 Redis 服务缓存失败，降级使用本地缓存", "error", err)
				return GetLocalCache()
			}
			logger.Info("服务注册中心使用 Redis 缓存", "cacheType", cacheType)
			return redisCache
		}
	}

	// 使用本地缓存
	logger.Info("服务注册中心使用本地缓存")
	return GetLocalCache()
}
