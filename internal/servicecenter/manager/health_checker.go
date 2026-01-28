package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"gateway/internal/servicecenter/cache"
	pb "gateway/internal/servicecenter/server/proto"
	"gateway/internal/servicecenter/types"
	"gateway/pkg/logger"
)

// HealthChecker 健康检查器
// 负责定期执行健康检查和缓存同步
type HealthChecker struct {
	instanceName string
	tenantId     string
	manager      *ServiceCenterManager
	interval     time.Duration // 健康检查间隔
	timeout      time.Duration // 健康检查超时
	stopCh       chan struct{}
	running      atomic.Bool
	wg           sync.WaitGroup
}

// NewHealthChecker 创建健康检查器
// 从实例配置中获取健康检查间隔和超时时间
func NewHealthChecker(instanceName, tenantId string, manager *ServiceCenterManager) *HealthChecker {
	hc := &HealthChecker{
		instanceName: instanceName,
		tenantId:     tenantId,
		manager:      manager,
		stopCh:       make(chan struct{}),
	}

	// 从实例配置获取健康检查间隔和超时时间
	srv := manager.GetInstance(instanceName)
	if srv != nil {
		config := srv.GetConfig()
		if config != nil {
			// 健康检查间隔（秒），默认30秒
			interval := config.HealthCheckInterval
			if interval <= 0 {
				interval = 30
			}
			hc.interval = time.Duration(interval) * time.Second

			// 健康检查超时（秒），默认10秒
			timeout := config.HealthCheckTimeout
			if timeout <= 0 {
				timeout = 10
			}
			hc.timeout = time.Duration(timeout) * time.Second
		} else {
			// 配置不存在，使用默认值
			hc.interval = 30 * time.Second
			hc.timeout = 10 * time.Second
		}
	} else {
		// 实例不存在，使用默认值
		hc.interval = 30 * time.Second
		hc.timeout = 10 * time.Second
	}

	return hc
}

// Start 启动健康检查器
func (hc *HealthChecker) Start() {
	if hc.running.Swap(true) {
		// 已在运行
		return
	}

	hc.wg.Add(1)
	go hc.run()

	logger.Info("健康检查器已启动",
		"instanceName", hc.instanceName,
		"interval", hc.interval)
}

// Stop 停止健康检查器
func (hc *HealthChecker) Stop() {
	if !hc.running.Swap(false) {
		// 未在运行
		return
	}

	close(hc.stopCh)
	hc.wg.Wait()

	// 重新创建 stopCh，以便下次启动
	hc.stopCh = make(chan struct{})

	logger.Info("健康检查器已停止", "instanceName", hc.instanceName)
}

// GetInterval 获取健康检查间隔
func (hc *HealthChecker) GetInterval() time.Duration {
	return hc.interval
}

// GetTimeout 获取健康检查超时时间
func (hc *HealthChecker) GetTimeout() time.Duration {
	return hc.timeout
}

// IsRunning 检查健康检查器是否正在运行
func (hc *HealthChecker) IsRunning() bool {
	return hc.running.Load()
}

// run 健康检查器主循环
func (hc *HealthChecker) run() {
	defer hc.wg.Done()

	ticker := time.NewTicker(hc.interval)
	defer ticker.Stop()

	// 立即执行一次
	hc.performHealthCheck()

	for {
		select {
		case <-ticker.C:
			hc.performHealthCheck()
		case <-hc.stopCh:
			return
		}
	}
}

// performHealthCheck 执行健康检查和缓存同步
func (hc *HealthChecker) performHealthCheck() {
	ctx := context.Background()

	// 检查实例是否还在运行
	srv := hc.manager.GetInstance(hc.instanceName)
	if srv == nil || !srv.IsRunning() {
		// 实例不存在或未运行，跳过同步
		return
	}

	// 1. 先执行节点心跳超时检查和驱逐（快速发现宕机节点）
	// 类似 Nacos 的实现：定期检查所有节点的心跳时间，超时则驱逐
	evictedCount := hc.checkAndEvictTimeoutNodes(ctx)

	// 2. 执行缓存同步
	if err := hc.syncCacheToDB(ctx); err != nil {
		logger.Warn("缓存同步失败",
			"instanceName", hc.instanceName,
			"error", err)
	} else {
		logger.Debug("缓存同步成功", "instanceName", hc.instanceName)
	}

	if evictedCount > 0 {
		logger.Info("节点心跳超时检查和驱逐完成",
			"instanceName", hc.instanceName,
			"evictedCount", evictedCount)
	}
}

// syncCacheToDB 将缓存同步到数据库
// 同步规则：
//   - 服务：如果服务有节点（非空），则同步服务到数据库；如果服务没有节点，则删除缓存中的服务
//   - 节点：只同步非临时节点（ephemeral != "Y"）
func (hc *HealthChecker) syncCacheToDB(ctx context.Context) error {
	globalCache := cache.GetGlobalCache()
	serviceCount := 0
	nodeCount := 0
	deletedServiceCount := 0
	var errors []string

	// 收集需要删除的服务（在回调中只收集，不执行删除，避免阻塞）
	type serviceToDelete struct {
		tenantId    string
		namespaceId string
		groupName   string
		serviceName string
	}
	var servicesToDelete []serviceToDelete

	// 遍历所有服务
	globalCache.GetAllServices(func(service *types.Service) {
		if service == nil {
			return
		}

		// 如果服务没有节点，标记为需要删除
		if len(service.Nodes) == 0 {
			servicesToDelete = append(servicesToDelete, serviceToDelete{
				tenantId:    service.TenantId,
				namespaceId: service.NamespaceId,
				groupName:   service.GroupName,
				serviceName: service.ServiceName,
			})
			return
		}

		// 同步服务
		if err := hc.syncService(ctx, service); err != nil {
			errors = append(errors, fmt.Sprintf("服务 %s/%s/%s: %v",
				service.NamespaceId, service.GroupName, service.ServiceName, err))
			return
		}
		serviceCount++

		// 同步节点（只同步非临时节点）
		for _, node := range service.Nodes {
			// 跳过临时节点
			if node.Ephemeral == "Y" {
				continue
			}

			if err := hc.syncNode(ctx, node); err != nil {
				errors = append(errors, fmt.Sprintf("节点 %s: %v", node.NodeId, err))
				continue
			}
			nodeCount++
		}
	})

	// 在回调外删除没有节点的服务（避免阻塞缓存遍历）
	// 同时删除缓存和数据库中的服务
	for _, svc := range servicesToDelete {
		// 先删除数据库中的服务
		if err := hc.manager.serviceDAO.DeleteService(ctx, svc.tenantId, svc.namespaceId, svc.groupName, svc.serviceName); err != nil {
			errors = append(errors, fmt.Sprintf("删除数据库服务 %s/%s/%s: %v",
				svc.namespaceId, svc.groupName, svc.serviceName, err))
			// 数据库删除失败时，仍然尝试删除缓存（避免缓存和数据库不一致）
		}

		// 删除缓存中的服务
		globalCache.DeleteService(ctx, svc.tenantId, svc.namespaceId, svc.groupName, svc.serviceName)
		deletedServiceCount++
		logger.Debug("删除没有节点的服务（缓存和数据库）",
			"instanceName", hc.instanceName,
			"namespaceId", svc.namespaceId,
			"groupName", svc.groupName,
			"serviceName", svc.serviceName)
	}

	if len(errors) > 0 {
		return fmt.Errorf("部分同步失败: %v", errors)
	}

	if serviceCount > 0 || nodeCount > 0 || deletedServiceCount > 0 {
		logger.Debug("缓存同步完成",
			"instanceName", hc.instanceName,
			"serviceCount", serviceCount,
			"nodeCount", nodeCount,
			"deletedServiceCount", deletedServiceCount)
	}

	return nil
}

// syncService 同步服务到数据库
func (hc *HealthChecker) syncService(ctx context.Context, service *types.Service) error {
	// 通过 manager 访问 serviceDAO
	// 检查服务是否已存在
	existingService, err := hc.manager.serviceDAO.GetService(ctx,
		service.TenantId, service.NamespaceId, service.GroupName, service.ServiceName)

	if err != nil {
		// GetService 在记录不存在时返回 nil, nil，其他错误才返回错误
		return fmt.Errorf("查询服务失败: %w", err)
	}

	if existingService == nil {
		// 服务不存在，创建
		if err := hc.manager.serviceDAO.CreateService(ctx, service); err != nil {
			return fmt.Errorf("创建服务失败: %w", err)
		}
	} else {
		// 服务已存在，更新
		if err := hc.manager.serviceDAO.UpdateService(ctx, service); err != nil {
			return fmt.Errorf("更新服务失败: %w", err)
		}
	}

	return nil
}

// syncNode 同步节点到数据库
func (hc *HealthChecker) syncNode(ctx context.Context, node *types.ServiceNode) error {
	// 通过 manager 访问 nodeDAO
	// 检查节点是否已存在
	existingNode, err := hc.manager.nodeDAO.GetNode(ctx, node.TenantId, node.NodeId)

	if err != nil {
		// GetNode 在记录不存在时返回 nil, nil，其他错误才返回错误
		return fmt.Errorf("查询节点失败: %w", err)
	}

	if existingNode == nil {
		// 节点不存在，创建
		if err := hc.manager.nodeDAO.CreateNode(ctx, node); err != nil {
			return fmt.Errorf("创建节点失败: %w", err)
		}
	} else {
		// 节点已存在，更新
		if err := hc.manager.nodeDAO.UpdateNode(ctx, node); err != nil {
			return fmt.Errorf("更新节点失败: %w", err)
		}
	}

	return nil
}

// checkAndEvictTimeoutNodes 检查并驱逐心跳超时的节点
// 类似 Nacos 的实现机制：
//   - 定期检查所有节点的 LastBeatTime
//   - 如果超过阈值（心跳间隔 * 2 或配置的超时时间），则认为节点不可用
//   - 对于临时节点（ephemeral="Y"），直接驱逐（从缓存中移除）
//   - 对于持久节点（ephemeral="N"），标记为不健康，但不驱逐
//
// 返回驱逐的节点数量
func (hc *HealthChecker) checkAndEvictTimeoutNodes(ctx context.Context) int {
	globalCache := cache.GetGlobalCache()
	now := time.Now()
	evictedCount := 0

	// 计算心跳超时阈值
	// 使用健康检查间隔作为超时阈值，确保在每次健康检查时能及时清理不健康的节点
	// 逻辑：
	//   - 健康检查器每 hc.interval 秒执行一次检查
	//   - 如果节点在 hc.interval 秒内没有心跳，说明在本次检查周期内不健康
	//   - 应该在本次检查中被清理，然后同步到数据库
	//   - hc.timeout 是服务端主动检查客户端地址的超时时间，不用于健康判断
	// 例如：如果健康检查间隔是 30 秒，则超时阈值是 30 秒
	// 这意味着如果节点在 30 秒内没有心跳，就会在本次检查中被清理
	timeoutThreshold := hc.interval

	// 收集需要驱逐的节点（在回调中只收集，不执行驱逐，避免阻塞）
	type evictItem struct {
		service *types.Service
		node    *types.ServiceNode
	}
	var evictItems []evictItem

	// 遍历所有服务，检查节点心跳超时
	globalCache.GetAllServices(func(service *types.Service) {
		if service == nil || len(service.Nodes) == 0 {
			return
		}

		for _, node := range service.Nodes {
			// 检查最后心跳时间
			if node.LastBeatTime == nil {
				// 如果最后心跳时间为空，且节点是临时的，标记为需要驱逐
				if node.Ephemeral == "Y" {
					evictItems = append(evictItems, evictItem{
						service: service,
						node:    node,
					})
				}
				continue
			}

			// 计算心跳超时时间
			timeSinceLastBeat := now.Sub(*node.LastBeatTime)
			if timeSinceLastBeat > timeoutThreshold {
				// 心跳超时
				if node.Ephemeral == "Y" {
					// 临时节点：直接驱逐
					evictItems = append(evictItems, evictItem{
						service: service,
						node:    node,
					})
					logger.Warn("检测到心跳超时的临时节点，将驱逐",
						"nodeId", node.NodeId,
						"serviceName", node.ServiceName,
						"lastBeatTime", node.LastBeatTime.Format("2006-01-02 15:04:05"),
						"timeout", timeSinceLastBeat,
						"threshold", timeoutThreshold)
				} else {
					// 持久节点：标记为不健康，但不驱逐
					if node.HealthyStatus != types.HealthyStatusUnhealthy {
						node.HealthyStatus = types.HealthyStatusUnhealthy
						node.EditTime = now
						globalCache.UpdateNode(ctx, node)
						logger.Warn("检测到心跳超时的持久节点，标记为不健康",
							"nodeId", node.NodeId,
							"serviceName", node.ServiceName,
							"lastBeatTime", node.LastBeatTime.Format("2006-01-02 15:04:05"),
							"timeout", timeSinceLastBeat)
					}
				}
			}
		}
	})

	// 在回调外执行驱逐操作（避免阻塞缓存遍历）
	for _, item := range evictItems {
		// 保存节点信息（用于构建事件）
		savedNode := item.node

		// 从缓存中移除节点
		globalCache.RemoveNode(ctx, item.node.TenantId, item.node.NamespaceId,
			item.node.GroupName, item.node.ServiceName, item.node.NodeId)

		// 从缓存获取完整的服务信息（包括删除后的所有节点列表）
		service, serviceFound := globalCache.GetService(ctx, item.node.TenantId, item.node.NamespaceId,
			item.node.GroupName, item.node.ServiceName)
		if !serviceFound || service == nil {
			// 如果服务不存在，说明节点已经被删除（可能是并发删除或之前已经删除）
			// 这是正常的幂等性处理，不需要发送事件（因为节点已经不存在了）
			logger.Info("节点驱逐后服务不存在，节点可能已被删除，不发送事件",
				"nodeId", item.node.NodeId,
				"namespaceId", item.node.NamespaceId,
				"groupName", item.node.GroupName,
				"serviceName", item.node.ServiceName)
			evictedCount++
			continue
		}

		// 构建完整的事件（包含服务信息和删除后的所有节点列表）
		// 需要导入 handler 包来使用 convertServiceToProto 和 convertNodeToProto
		// 或者在这里实现转换逻辑
		pbNodes := make([]*pb.Node, 0, len(service.Nodes))
		for _, n := range service.Nodes {
			pbNodes = append(pbNodes, convertNodeToProto(n))
		}

		event := &pb.ServiceChangeEvent{
			EventType:   "NODE_REMOVED",
			Timestamp:   now.Format("2006-01-02 15:04:05"),
			NamespaceId: item.node.NamespaceId,
			GroupName:   item.node.GroupName,
			ServiceName: item.node.ServiceName,
			Service:     convertServiceToProto(service),
			Nodes:       pbNodes,
			ChangedNode: convertNodeToProto(savedNode),
		}
		// 直接调用 manager 的通知方法
		if err := hc.manager.NotifyServiceChange(ctx, hc.instanceName, item.node.TenantId,
			item.node.NamespaceId, item.node.GroupName, item.node.ServiceName, event); err != nil {
			logger.Warn("通知服务变更失败",
				"nodeId", item.node.NodeId,
				"error", err)
		}

		evictedCount++
		logger.Info("已驱逐心跳超时的临时节点",
			"nodeId", item.node.NodeId,
			"serviceName", item.node.ServiceName,
			"namespaceId", item.node.NamespaceId,
			"groupName", item.node.GroupName)
	}

	return evictedCount
}

// convertServiceToProto 转换服务为 protobuf 格式
func convertServiceToProto(service *types.Service) *pb.Service {
	if service == nil {
		return nil
	}
	// 转换 metadata JSON 字符串为 map
	metadata := make(map[string]string)
	if service.MetadataJson != "" {
		if err := json.Unmarshal([]byte(service.MetadataJson), &metadata); err != nil {
			metadata = make(map[string]string)
		}
	}
	// 转换 tags JSON 字符串为 map
	tags := make(map[string]string)
	if service.TagsJson != "" {
		if err := json.Unmarshal([]byte(service.TagsJson), &tags); err != nil {
			tags = make(map[string]string)
		}
	}

	return &pb.Service{
		NamespaceId:        service.NamespaceId,
		GroupName:          service.GroupName,
		ServiceName:        service.ServiceName,
		ServiceType:        service.ServiceType,
		ServiceVersion:     service.ServiceVersion,
		ServiceDescription: service.ServiceDescription,
		ProtectThreshold:   service.ProtectThreshold,
		Metadata:           metadata,
		Tags:               tags,
		// node 字段在响应中不使用
	}
}

// convertNodeToProto 转换节点为 protobuf 格式
func convertNodeToProto(node *types.ServiceNode) *pb.Node {
	if node == nil {
		return nil
	}
	// 转换 JSON 字符串为 map
	metadata := make(map[string]string)
	if node.MetadataJson != "" {
		if err := json.Unmarshal([]byte(node.MetadataJson), &metadata); err != nil {
			// 解析失败时使用空 map
			metadata = make(map[string]string)
		}
	}

	return &pb.Node{
		NodeId:         node.NodeId,
		NamespaceId:    node.NamespaceId,
		GroupName:      node.GroupName,
		ServiceName:    node.ServiceName,
		IpAddress:      node.IpAddress,
		PortNumber:     int32(node.PortNumber),
		Weight:         node.Weight,
		Ephemeral:      node.Ephemeral,
		InstanceStatus: node.InstanceStatus,
		HealthyStatus:  node.HealthyStatus,
		Metadata:       metadata,
	}
}
