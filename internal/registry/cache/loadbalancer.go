package cache

import (
	"context"
	"fmt"
	"hash/crc32"
	"math/rand"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"gateway/internal/registry/core"
	"gateway/pkg/logger"
)

// filterHealthyInstances 过滤健康的实例
// 基于实例的 IsAvailable 方法进行实时健康检查
func filterHealthyInstances(instances []*core.ServiceInstance) []*core.ServiceInstance {
	if len(instances) == 0 {
		return instances
	}

	healthyInstances := make([]*core.ServiceInstance, 0, len(instances))
	for _, instance := range instances {
		// 使用 ServiceInstance 的 IsAvailable 方法
		// IsAvailable 会检查：健康状态 + 实例状态 + 激活标志
		if instance.IsAvailable() {
			healthyInstances = append(healthyInstances, instance)
		}
	}

	return healthyInstances
}

// LoadBalancer 负载均衡器接口
// 定义不同负载均衡策略的统一接口
type LoadBalancer interface {
	// Select 从可用实例列表中选择一个实例
	// 实现应该：
	// 1. 实时过滤健康实例
	// 2. 懒清理已删除实例的状态
	// 3. 基于健康实例执行负载均衡算法
	Select(ctx context.Context, instances []*core.ServiceInstance) (*core.ServiceInstance, error)

	// GetStrategy 获取负载均衡策略类型
	GetStrategy() string

	// CleanupStaleState 清理已删除实例的陈旧状态（可选实现）
	// 有状态的负载均衡器可以实现此方法进行懒清理
	CleanupStaleState(currentInstanceIds []string)
}

// RandomLoadBalancer 随机负载均衡器
// 随机选择一个可用实例
type RandomLoadBalancer struct{}

// Select 随机选择一个实例
func (lb *RandomLoadBalancer) Select(ctx context.Context, instances []*core.ServiceInstance) (*core.ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, fmt.Errorf("没有可用的服务实例")
	}

	// 过滤健康的实例
	healthyInstances := filterHealthyInstances(instances)
	if len(healthyInstances) == 0 {
		return nil, fmt.Errorf("没有健康的服务实例")
	}

	return healthyInstances[rand.Intn(len(healthyInstances))], nil
}

// GetStrategy 获取负载均衡策略类型
func (lb *RandomLoadBalancer) GetStrategy() string {
	return core.LoadBalanceRandom
}

// CleanupStaleState 清理陈旧状态（随机负载均衡器无状态，无需清理）
func (lb *RandomLoadBalancer) CleanupStaleState(currentInstanceIds []string) {
	// 随机负载均衡器无状态，无需清理
}

// RoundRobinLoadBalancer 轮询负载均衡器
// 按顺序循环选择可用实例
type RoundRobinLoadBalancer struct {
	counter uint64 // 原子计数器
}

// Select 轮询选择一个实例
func (lb *RoundRobinLoadBalancer) Select(ctx context.Context, instances []*core.ServiceInstance) (*core.ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, fmt.Errorf("没有可用的服务实例")
	}

	// 过滤健康的实例
	healthyInstances := filterHealthyInstances(instances)
	if len(healthyInstances) == 0 {
		return nil, fmt.Errorf("没有健康的服务实例")
	}

	// 原子递增计数器并取模
	count := atomic.AddUint64(&lb.counter, 1) - 1
	index := int(count % uint64(len(healthyInstances)))

	return healthyInstances[index], nil
}

// GetStrategy 获取负载均衡策略类型
func (lb *RoundRobinLoadBalancer) GetStrategy() string {
	return core.LoadBalanceRoundRobin
}

// CleanupStaleState 清理陈旧状态（轮询负载均衡器可选择重置计数器）
func (lb *RoundRobinLoadBalancer) CleanupStaleState(currentInstanceIds []string) {
	// 轮询负载均衡器主要基于计数器，通常不需要清理特定实例状态
	// 但可以选择性重置计数器以提高分布均匀性
	// 这里保持简单，不做强制重置，让自然轮询处理
}

// WeightedRoundRobinLoadBalancer 加权轮询负载均衡器
// 根据实例权重进行选择，使用Nginx平滑加权轮询算法
type WeightedRoundRobinLoadBalancer struct {
	mutex            sync.Mutex
	currentWeights   map[string]int // 实例ID -> 当前权重
	effectiveWeights map[string]int // 实例ID -> 有效权重
	totalWeight      int            // 总权重
}

// NewWeightedRoundRobinLoadBalancer 创建加权轮询负载均衡器
func NewWeightedRoundRobinLoadBalancer() *WeightedRoundRobinLoadBalancer {
	return &WeightedRoundRobinLoadBalancer{
		currentWeights:   make(map[string]int),
		effectiveWeights: make(map[string]int),
	}
}

// Select 加权轮询选择一个实例
// 使用Nginx平滑加权轮询算法
func (lb *WeightedRoundRobinLoadBalancer) Select(ctx context.Context, instances []*core.ServiceInstance) (*core.ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, fmt.Errorf("没有可用的服务实例")
	}

	// 过滤健康的实例
	healthyInstances := filterHealthyInstances(instances)
	if len(healthyInstances) == 0 {
		return nil, fmt.Errorf("没有健康的服务实例")
	}

	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	// 懒清理：每次Select时清理已删除实例的权重状态
	currentInstanceIds := make([]string, len(instances))
	for i, instance := range instances {
		currentInstanceIds[i] = instance.ServiceInstanceId
	}
	lb.cleanupStaleWeights(currentInstanceIds)

	// 如果只有一个健康实例，直接返回
	if len(healthyInstances) == 1 {
		return healthyInstances[0], nil
	}

	// 初始化或更新权重映射（使用健康实例）
	lb.initWeights(healthyInstances)

	// 选择权重最高的实例
	var selectedInstance *core.ServiceInstance
	var maxWeight int = -1

	// 第一轮：增加currentWeight并找出最大值
	for _, instance := range healthyInstances {
		id := instance.ServiceInstanceId
		weight := instance.WeightValue
		if weight <= 0 {
			weight = 1 // 确保至少为1
		}

		// 当前权重增加有效权重
		lb.currentWeights[id] += lb.effectiveWeights[id]

		// 找出当前权重最大的实例
		if lb.currentWeights[id] > maxWeight {
			maxWeight = lb.currentWeights[id]
			selectedInstance = instance
		}
	}

	// 选中实例的当前权重减去总有效权重
	if selectedInstance != nil {
		lb.currentWeights[selectedInstance.ServiceInstanceId] -= lb.totalWeight
	}

	return selectedInstance, nil
}

// cleanupStaleWeights 清理已删除实例的权重状态（内部方法）
func (lb *WeightedRoundRobinLoadBalancer) cleanupStaleWeights(currentInstanceIds []string) {
	// 构建当前实例ID映射
	instanceIdMap := make(map[string]bool)
	for _, id := range currentInstanceIds {
		instanceIdMap[id] = true
	}

	// 清理已不存在实例的权重状态
	for id := range lb.currentWeights {
		if !instanceIdMap[id] {
			delete(lb.currentWeights, id)
			delete(lb.effectiveWeights, id)
		}
	}

	// 重置总权重，让initWeights重新计算
	lb.totalWeight = 0
}

// initWeights 初始化或更新权重映射
func (lb *WeightedRoundRobinLoadBalancer) initWeights(instances []*core.ServiceInstance) {
	// 重置总权重
	lb.totalWeight = 0

	// 准备实例ID列表
	instanceIds := make(map[string]bool)
	for _, instance := range instances {
		instanceIds[instance.ServiceInstanceId] = true

		// 初始化有效权重
		weight := instance.WeightValue
		if weight <= 0 {
			weight = 1
		}

		if _, exists := lb.effectiveWeights[instance.ServiceInstanceId]; !exists {
			lb.effectiveWeights[instance.ServiceInstanceId] = weight
		}

		// 累加总权重
		lb.totalWeight += weight
	}

	// 清理已不存在的实例
	for id := range lb.currentWeights {
		if !instanceIds[id] {
			delete(lb.currentWeights, id)
			delete(lb.effectiveWeights, id)
		}
	}
}

// GetStrategy 获取负载均衡策略类型
func (lb *WeightedRoundRobinLoadBalancer) GetStrategy() string {
	return core.LoadBalanceWeightedRoundRobin
}

// CleanupStaleState 清理陈旧状态（加权轮询负载均衡器清理已删除实例的权重状态）
func (lb *WeightedRoundRobinLoadBalancer) CleanupStaleState(currentInstanceIds []string) {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	// 构建当前实例ID映射
	instanceIdMap := make(map[string]bool)
	for _, id := range currentInstanceIds {
		instanceIdMap[id] = true
	}

	// 清理已不存在实例的权重状态
	for id := range lb.currentWeights {
		if !instanceIdMap[id] {
			delete(lb.currentWeights, id)
			delete(lb.effectiveWeights, id)
		}
	}

	// 重置总权重，让下次Select时重新计算
	lb.totalWeight = 0
}

// LeastConnectionsLoadBalancer 最少连接数负载均衡器
// 选择当前连接数最少的实例
type LeastConnectionsLoadBalancer struct {
	mutex       sync.Mutex
	connections map[string]int // 实例ID -> 连接数
}

// NewLeastConnectionsLoadBalancer 创建最少连接数负载均衡器
func NewLeastConnectionsLoadBalancer() *LeastConnectionsLoadBalancer {
	return &LeastConnectionsLoadBalancer{
		connections: make(map[string]int),
	}
}

// Select 选择连接数最少的实例
func (lb *LeastConnectionsLoadBalancer) Select(ctx context.Context, instances []*core.ServiceInstance) (*core.ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, fmt.Errorf("没有可用的服务实例")
	}

	// 过滤健康的实例
	healthyInstances := filterHealthyInstances(instances)
	if len(healthyInstances) == 0 {
		return nil, fmt.Errorf("没有健康的服务实例")
	}

	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	// 懒清理：每次Select时清理已删除实例的连接计数
	currentInstanceIds := make([]string, len(instances))
	for i, instance := range instances {
		currentInstanceIds[i] = instance.ServiceInstanceId
	}
	lb.cleanupStaleConnections(currentInstanceIds)

	// 如果只有一个健康实例，直接返回
	if len(healthyInstances) == 1 {
		instanceId := healthyInstances[0].ServiceInstanceId
		// 初始化连接计数（如果不存在）
		if _, exists := lb.connections[instanceId]; !exists {
			lb.connections[instanceId] = 0
		}
		lb.connections[instanceId]++ // 增加连接计数
		return healthyInstances[0], nil
	}

	// 找出连接数最少的实例
	var selectedInstance *core.ServiceInstance
	var minAdjustedConns float64 = float64(^uint(0) >> 1) // 最大值

	for _, instance := range healthyInstances {
		id := instance.ServiceInstanceId

		// 初始化连接计数（如果不存在）
		if _, exists := lb.connections[id]; !exists {
			lb.connections[id] = 0
		}

		conns := lb.connections[id]

		// 根据权重调整连接数评估值
		weight := instance.WeightValue
		if weight <= 0 {
			weight = 1
		}

		// 连接数/权重比越小越好
		adjustedConns := float64(conns) / float64(weight)

		if selectedInstance == nil || adjustedConns < minAdjustedConns {
			minAdjustedConns = adjustedConns
			selectedInstance = instance
		} else if adjustedConns == minAdjustedConns {
			// 如果连接数/权重比相同，随机选择一个
			if rand.Intn(2) == 0 {
				minAdjustedConns = adjustedConns
				selectedInstance = instance
			}
		}
	}

	// 增加选中实例的连接计数
	if selectedInstance != nil {
		lb.connections[selectedInstance.ServiceInstanceId]++
	}

	return selectedInstance, nil
}

// cleanupStaleConnections 清理已删除实例的连接计数（内部方法）
func (lb *LeastConnectionsLoadBalancer) cleanupStaleConnections(currentInstanceIds []string) {
	// 构建当前实例ID映射
	instanceIdMap := make(map[string]bool)
	for _, id := range currentInstanceIds {
		instanceIdMap[id] = true
	}

	// 清理已不存在实例的连接计数
	for id := range lb.connections {
		if !instanceIdMap[id] {
			delete(lb.connections, id)
		}
	}
}

// ReleaseConnection 释放实例连接，减少连接计数
func (lb *LeastConnectionsLoadBalancer) ReleaseConnection(instanceId string) {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	if count, exists := lb.connections[instanceId]; exists && count > 0 {
		lb.connections[instanceId]--
	}
}

// GetStrategy 获取负载均衡策略类型
func (lb *LeastConnectionsLoadBalancer) GetStrategy() string {
	return core.LoadBalanceLeastConnections
}

// CleanupStaleState 清理陈旧状态（最少连接数负载均衡器清理已删除实例的连接计数）
func (lb *LeastConnectionsLoadBalancer) CleanupStaleState(currentInstanceIds []string) {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	// 构建当前实例ID映射
	instanceIdMap := make(map[string]bool)
	for _, id := range currentInstanceIds {
		instanceIdMap[id] = true
	}

	// 清理已不存在实例的连接计数
	for id := range lb.connections {
		if !instanceIdMap[id] {
			delete(lb.connections, id)
		}
	}
}

// IPHashLoadBalancer IP哈希负载均衡器
// 基于客户端IP地址的哈希值选择实例，确保同一客户端IP总是访问同一实例
type IPHashLoadBalancer struct{}

// Select 基于IP哈希选择实例
// 从上下文中获取客户端IP地址，如果没有则使用随机选择
func (lb *IPHashLoadBalancer) Select(ctx context.Context, instances []*core.ServiceInstance) (*core.ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, fmt.Errorf("没有可用的服务实例")
	}

	// 过滤健康的实例
	healthyInstances := filterHealthyInstances(instances)
	if len(healthyInstances) == 0 {
		return nil, fmt.Errorf("没有健康的服务实例")
	}

	// 尝试从上下文中获取客户端IP
	var clientIP string
	if value := ctx.Value("client_ip"); value != nil {
		clientIP, _ = value.(string)
	}

	// 如果没有获取到客户端IP，使用随机策略
	if clientIP == "" {
		logger.DebugWithTrace(ctx, "IP哈希负载均衡器未找到客户端IP，使用随机策略")
		return healthyInstances[rand.Intn(len(healthyInstances))], nil
	}

	// 计算IP地址的哈希值
	hash := crc32.ChecksumIEEE([]byte(clientIP))
	index := int(hash % uint32(len(healthyInstances)))

	logger.DebugWithTrace(ctx, "IP哈希负载均衡器",
		"clientIP", clientIP,
		"hash", hash,
		"index", index)

	return healthyInstances[index], nil
}

// GetStrategy 获取负载均衡策略类型
func (lb *IPHashLoadBalancer) GetStrategy() string {
	return core.LoadBalanceIpHash
}

// CleanupStaleState 清理陈旧状态（IP哈希负载均衡器无状态，无需清理）
func (lb *IPHashLoadBalancer) CleanupStaleState(currentInstanceIds []string) {
	// IP哈希负载均衡器无状态，无需清理
}

// ConsistentHashLoadBalancer 一致性哈希负载均衡器
// 使用一致性哈希算法选择实例，减少节点变动时的影响
type ConsistentHashLoadBalancer struct {
	mutex      sync.RWMutex
	hashRing   []int            // 排序后的哈希环
	hashMap    map[int]string   // 哈希值 -> 实例ID
	replicas   int              // 每个实例的虚拟节点数
	hashValues map[string][]int // 实例ID -> 该实例的所有虚拟节点哈希值
}

// NewConsistentHashLoadBalancer 创建一致性哈希负载均衡器
func NewConsistentHashLoadBalancer(replicas int) *ConsistentHashLoadBalancer {
	if replicas <= 0 {
		replicas = 10 // 默认10个虚拟节点
	}

	return &ConsistentHashLoadBalancer{
		hashRing:   make([]int, 0),
		hashMap:    make(map[int]string),
		replicas:   replicas,
		hashValues: make(map[string][]int),
	}
}

// Select 使用一致性哈希选择实例
func (lb *ConsistentHashLoadBalancer) Select(ctx context.Context, instances []*core.ServiceInstance) (*core.ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, fmt.Errorf("没有可用的服务实例")
	}

	// 过滤健康的实例
	healthyInstances := filterHealthyInstances(instances)
	if len(healthyInstances) == 0 {
		return nil, fmt.Errorf("没有健康的服务实例")
	}

	// 如果只有一个健康实例，直接返回
	if len(healthyInstances) == 1 {
		return healthyInstances[0], nil
	}

	// 更新哈希环（使用健康实例，内部会进行懒清理）
	lb.updateHashRing(healthyInstances)

	// 尝试从上下文中获取路由键（可以是客户端IP、会话ID等）
	var routingKey string
	if value := ctx.Value("routing_key"); value != nil {
		routingKey, _ = value.(string)
	}

	// 如果没有路由键，尝试获取客户端IP作为路由键
	if routingKey == "" {
		if value := ctx.Value("client_ip"); value != nil {
			routingKey, _ = value.(string)
		}
	}

	// 如果仍然没有路由键，生成一个随机键
	if routingKey == "" {
		routingKey = strconv.FormatInt(time.Now().UnixNano(), 10)
	}

	// 计算路由键的哈希值
	hash := int(crc32.ChecksumIEEE([]byte(routingKey)))

	// 在哈希环上查找实例
	lb.mutex.RLock()
	defer lb.mutex.RUnlock()

	if len(lb.hashRing) == 0 {
		// 如果哈希环为空，随机选择一个健康实例
		return healthyInstances[rand.Intn(len(healthyInstances))], nil
	}

	// 二分查找大于等于哈希值的最小索引
	idx := sort.Search(len(lb.hashRing), func(i int) bool {
		return lb.hashRing[i] >= hash
	})

	// 如果找不到大于等于哈希值的索引，则回环到第一个节点
	if idx == len(lb.hashRing) {
		idx = 0
	}

	// 获取对应的实例ID
	instanceID := lb.hashMap[lb.hashRing[idx]]

	// 查找实例对象（在健康实例中查找）
	for _, instance := range healthyInstances {
		if instance.ServiceInstanceId == instanceID {
			logger.DebugWithTrace(ctx, "一致性哈希负载均衡器",
				"routingKey", routingKey,
				"hash", hash,
				"selectedInstanceId", instanceID)
			return instance, nil
		}
	}

	// 如果找不到实例（可能因为实例列表变动），随机选择一个健康实例
	logger.DebugWithTrace(ctx, "一致性哈希负载均衡器未找到匹配实例，使用随机策略",
		"routingKey", routingKey,
		"hash", hash)
	return healthyInstances[rand.Intn(len(healthyInstances))], nil
}

// updateHashRing 更新哈希环
func (lb *ConsistentHashLoadBalancer) updateHashRing(instances []*core.ServiceInstance) {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	// 准备实例ID列表
	currentInstanceIds := make(map[string]bool)
	for _, instance := range instances {
		currentInstanceIds[instance.ServiceInstanceId] = true
	}

	// 检查是否需要更新哈希环
	needUpdate := false

	// 检查新增的实例
	for _, instance := range instances {
		if _, exists := lb.hashValues[instance.ServiceInstanceId]; !exists {
			needUpdate = true
			break
		}
	}

	// 检查删除的实例
	if !needUpdate {
		for id := range lb.hashValues {
			if !currentInstanceIds[id] {
				needUpdate = true
				break
			}
		}
	}

	// 如果不需要更新，直接返回
	if !needUpdate {
		return
	}

	// 清空哈希环
	lb.hashRing = make([]int, 0)
	lb.hashMap = make(map[int]string)
	lb.hashValues = make(map[string][]int)

	// 为每个实例创建虚拟节点
	for _, instance := range instances {
		id := instance.ServiceInstanceId
		lb.hashValues[id] = make([]int, 0, lb.replicas)

		for i := 0; i < lb.replicas; i++ {
			// 使用实例ID和复制因子生成哈希值
			key := fmt.Sprintf("%s:%d", id, i)
			hash := int(crc32.ChecksumIEEE([]byte(key)))
			lb.hashRing = append(lb.hashRing, hash)
			lb.hashMap[hash] = id
			lb.hashValues[id] = append(lb.hashValues[id], hash)
		}
	}

	// 排序哈希环
	sort.Ints(lb.hashRing)
}

// GetStrategy 获取负载均衡策略类型
func (lb *ConsistentHashLoadBalancer) GetStrategy() string {
	return "CONSISTENT_HASH" // 注意：这里使用了自定义策略名，需要在core/types.go中添加对应的常量
}

// CleanupStaleState 清理陈旧状态（一致性哈希负载均衡器更新哈希环）
func (lb *ConsistentHashLoadBalancer) CleanupStaleState(currentInstanceIds []string) {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	// 构建当前实例ID映射
	instanceIdMap := make(map[string]bool)
	for _, id := range currentInstanceIds {
		instanceIdMap[id] = true
	}

	// 检查是否有实例被删除
	hasRemovedInstance := false
	for id := range lb.hashValues {
		if !instanceIdMap[id] {
			hasRemovedInstance = true
			break
		}
	}

	// 如果有实例被删除，清理对应的哈希值
	if hasRemovedInstance {
		for id := range lb.hashValues {
			if !instanceIdMap[id] {
				// 从哈希环中移除该实例的所有虚拟节点
				for _, hash := range lb.hashValues[id] {
					delete(lb.hashMap, hash)
				}
				delete(lb.hashValues, id)
			}
		}

		// 重建哈希环（移除已删除实例的哈希值）
		newHashRing := make([]int, 0, len(lb.hashRing))
		for _, hash := range lb.hashRing {
			if _, exists := lb.hashMap[hash]; exists {
				newHashRing = append(newHashRing, hash)
			}
		}
		lb.hashRing = newHashRing
	}
}

// ServiceKey 服务标识，用于唯一标识一个服务
type ServiceKey struct {
	TenantId       string
	ServiceGroupId string
	ServiceName    string
}

// String 返回服务标识的字符串表示
func (sk ServiceKey) String() string {
	return fmt.Sprintf("%s:%s:%s", sk.TenantId, sk.ServiceGroupId, sk.ServiceName)
}

// LoadBalancerFactory 负载均衡器工厂
// 为每个服务维护独立的负载均衡器实例，确保状态隔离
type LoadBalancerFactory struct {
	// 缓存已创建的负载均衡器实例
	// key: "strategy:tenantId:serviceGroupId:serviceName"
	// value: LoadBalancer实例
	balancers map[string]LoadBalancer
	mutex     sync.RWMutex
}

// NewLoadBalancerFactory 创建负载均衡器工厂
func NewLoadBalancerFactory() *LoadBalancerFactory {
	return &LoadBalancerFactory{
		balancers: make(map[string]LoadBalancer),
	}
}

// GetLoadBalancer 获取指定服务和策略的负载均衡器
// 为每个服务创建独立的负载均衡器实例，确保状态隔离
func (f *LoadBalancerFactory) GetLoadBalancer(strategy string, serviceKey ServiceKey) LoadBalancer {
	// 构造缓存key: "strategy:tenantId:serviceGroupId:serviceName"
	cacheKey := fmt.Sprintf("%s:%s", strategy, serviceKey.String())

	// 先尝试从缓存获取
	f.mutex.RLock()
	if lb, exists := f.balancers[cacheKey]; exists {
		f.mutex.RUnlock()
		return lb
	}
	f.mutex.RUnlock()

	// 缓存未命中，创建新实例
	f.mutex.Lock()
	defer f.mutex.Unlock()

	// 双重检查
	if lb, exists := f.balancers[cacheKey]; exists {
		return lb
	}

	// 创建新实例
	var lb LoadBalancer

	switch strategy {
	case core.LoadBalanceRandom:
		lb = &RandomLoadBalancer{}
	case core.LoadBalanceRoundRobin:
		lb = &RoundRobinLoadBalancer{}
	case core.LoadBalanceWeightedRoundRobin:
		lb = NewWeightedRoundRobinLoadBalancer()
	case core.LoadBalanceLeastConnections:
		lb = NewLeastConnectionsLoadBalancer()
	case core.LoadBalanceIpHash:
		lb = &IPHashLoadBalancer{}
	case "CONSISTENT_HASH": // 一致性哈希算法
		lb = NewConsistentHashLoadBalancer(10)
	default:
		// 默认使用随机负载均衡
		lb = &RandomLoadBalancer{}
	}

	// 缓存实例
	f.balancers[cacheKey] = lb

	return lb
}

// RemoveLoadBalancer 移除指定策略的负载均衡器实例
func (f *LoadBalancerFactory) RemoveLoadBalancer(strategy string, serviceKey ServiceKey) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	// 删除指定策略的负载均衡器 - O(1)操作
	cacheKey := fmt.Sprintf("%s:%s", strategy, serviceKey.String())
	delete(f.balancers, cacheKey)
}

// RemoveAllLoadBalancers 移除指定服务的所有负载均衡器实例
// 当服务被删除时调用，清理相关的负载均衡器状态
func (f *LoadBalancerFactory) RemoveAllLoadBalancers(serviceKey ServiceKey) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	serviceKeyStr := serviceKey.String()

	// 删除该服务的所有负载均衡器实例
	strategies := core.GetValidLoadBalanceStrategies()
	// 添加扩展策略
	strategies = append(strategies, "CONSISTENT_HASH")

	for _, strategy := range strategies {
		cacheKey := fmt.Sprintf("%s:%s", strategy, serviceKeyStr)
		delete(f.balancers, cacheKey)
	}
}
