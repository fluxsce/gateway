package service

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// SharedHealthCheckerManager 共享健康检查器管理器
// 用于管理所有服务的健康检查，避免每个服务创建独立的检查器
// 优势：
// 1. 只有一个检查循环 goroutine，而不是每个服务一个
// 2. 使用工作池模式限制并发检查数量，避免 goroutine 爆炸
// 3. 更好的资源利用和可扩展性
// 4. 直接访问 ServiceManager 管理的服务，直接更新节点健康状态，无需回调
// 5. 不维护节点列表，直接从 ServiceManager 获取所有服务的节点，保证数据一致性
// 6. 不需要注册/注销节点，自动检查所有启用了健康检查的节点
// 7. 每个服务的健康检查配置（间隔、超时等）从 ServiceConfig.HealthConfig 获取
type SharedHealthCheckerManager struct {
	mu             sync.RWMutex
	running        bool
	stopCh         chan struct{}
	workers        int                    // 工作池大小，限制并发检查数
	workerPool     chan struct{}          // 工作池信号量
	client         *http.Client           // 共享的 HTTP 客户端
	serviceManager *DefaultServiceManager // ServiceManager 引用，用于直接访问服务和更新节点健康状态
}

// NewSharedHealthCheckerManager 创建共享健康检查器管理器
// serviceManager: ServiceManager 实例，用于直接访问和更新服务节点状态
// workers: 工作池大小，限制并发检查的 goroutine 数量（建议 50-500）
// 注意：健康检查的间隔、超时等配置从每个服务的 ServiceConfig.HealthCheck 获取
func NewSharedHealthCheckerManager(serviceManager *DefaultServiceManager, workers int) *SharedHealthCheckerManager {
	if workers <= 0 {
		workers = 100 // 默认 100 个并发
	}

	// 使用一个合理的默认超时时间作为 HTTP 客户端的超时（实际检查时使用 HealthConfig 中的超时）
	defaultClientTimeout := 30 * time.Second

	return &SharedHealthCheckerManager{
		workers:        workers,
		workerPool:     make(chan struct{}, workers),
		serviceManager: serviceManager,
		client: &http.Client{
			Timeout: defaultClientTimeout,
			Transport: &http.Transport{
				MaxIdleConns:        workers * 2,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

// Start 启动共享健康检查器
func (s *SharedHealthCheckerManager) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("共享健康检查器已经在运行")
	}

	s.running = true
	s.stopCh = make(chan struct{})

	// 启动健康检查循环（只有一个 goroutine）
	go s.healthCheckLoop()

	return nil
}

// Stop 停止共享健康检查器
func (s *SharedHealthCheckerManager) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	s.running = false
	close(s.stopCh)

	return nil
}

// 注意：不再需要 RegisterNode 和 UnregisterNode 方法
// 共享检查器直接从 ServiceManager 获取所有服务和节点，自动检查启用了健康检查的节点

// healthCheckLoop 健康检查循环（只有一个 goroutine）
// 注意：使用一个较短的固定间隔作为 ticker，实际检查间隔由每个服务的 HealthConfig.Interval 控制
func (s *SharedHealthCheckerManager) healthCheckLoop() {
	// 使用 1 秒作为 ticker 间隔，实际检查间隔由每个服务的 HealthConfig.Interval 控制
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.performHealthChecks()
		}
	}
}

// performHealthChecks 执行健康检查（使用工作池模式）
// 注意：此方法从 ServiceManager 获取所有服务和节点，而不是维护自己的节点列表
func (s *SharedHealthCheckerManager) performHealthChecks() {
	if s.serviceManager == nil {
		return
	}

	// 直接从 ServiceManager 获取所有服务（返回 map 副本）
	servicesMap := s.serviceManager.GetServices()
	if len(servicesMap) == 0 {
		return
	}

	now := time.Now()
	checkTasks := make([]*checkTask, 0)

	for serviceID, service := range servicesMap {
		// 获取服务配置，检查是否启用了健康检查
		serviceConfig := service.GetConfig()
		if serviceConfig.HealthCheck == nil || !serviceConfig.HealthCheck.Enabled {
			continue // 该服务未启用健康检查，跳过
		}

		healthConfig := serviceConfig.HealthCheck

		// 遍历服务的所有节点（直接从 serviceConfig.Nodes 获取，这是实际节点的引用）
		for _, node := range serviceConfig.Nodes {
			// 只检查启用的节点
			if !node.Enabled {
				continue
			}

			// 检查是否到了该节点的检查时间（使用服务的 HealthConfig.Interval）
			// 节点的状态信息（连续成功/失败次数、上次检查时间）直接存储在 NodeConfig 中
			interval := healthConfig.Interval
			lastCheck := node.LastHealthCheck

			// 如果还没到检查时间，跳过
			if interval > 0 && now.Sub(lastCheck) < interval {
				continue
			}

			// 添加到检查任务列表
			checkTasks = append(checkTasks, &checkTask{
				node:         node,
				serviceID:    serviceID,
				healthConfig: healthConfig,
			})
		}
	}

	// 使用工作池模式限制并发数
	for _, task := range checkTasks {
		// 获取工作池信号量（如果池满，会阻塞）
		s.workerPool <- struct{}{}

		go func(t *checkTask) {
			defer func() {
				// 释放工作池信号量
				<-s.workerPool
			}()

			s.checkNodeHealth(t.node, t.serviceID, t.healthConfig)
		}(task)
	}
}

// checkTask 检查任务
type checkTask struct {
	node         *NodeConfig
	serviceID    string
	healthConfig *HealthConfig
}

// checkNodeHealth 检查单个节点健康状态
// 注意：节点的健康检查状态信息（连续成功/失败次数、上次检查时间）直接存储在 NodeConfig 中
// 通过 ServiceManager 更新实际节点中的状态
func (s *SharedHealthCheckerManager) checkNodeHealth(node *NodeConfig, serviceID string, config *HealthConfig) {
	if node == nil || s.serviceManager == nil || config == nil {
		return
	}

	previousHealth := node.Health
	previousConsecutiveSuccess := node.ConsecutiveSuccess
	previousConsecutiveFailure := node.ConsecutiveFailure

	// 执行健康检查
	healthy := s.doHealthCheck(node, config)

	now := time.Now()

	// 计算新的连续成功/失败计数
	var consecutiveSuccess, consecutiveFailure int
	var newHealth bool
	var shouldUpdateHealth bool

	if healthy {
		consecutiveSuccess = previousConsecutiveSuccess + 1
		consecutiveFailure = 0

		// 连续成功达到阈值，应该标记为健康
		if consecutiveSuccess >= config.HealthyThreshold {
			newHealth = true
			shouldUpdateHealth = !previousHealth // 只有当前不健康时才需要更新
		} else {
			newHealth = previousHealth // 保持当前状态
			shouldUpdateHealth = false
		}
	} else {
		consecutiveFailure = previousConsecutiveFailure + 1
		consecutiveSuccess = 0

		// 连续失败达到阈值，应该标记为不健康
		if consecutiveFailure >= config.UnhealthyThreshold {
			newHealth = false
			shouldUpdateHealth = previousHealth // 只有当前健康时才需要更新
		} else {
			newHealth = previousHealth // 保持当前状态
			shouldUpdateHealth = false
		}
	}

	// 直接更新节点状态（node 是从 service.config.Nodes 中获取的引用，是指向实际节点的指针）
	// 注意：这里直接更新节点字段，因为健康检查状态字段（ConsecutiveSuccess, ConsecutiveFailure, LastHealthCheck）
	// 仅用于健康检查逻辑，不会影响负载均衡的核心逻辑，直接更新是可以接受的
	// 如果需要更严格的线程安全，可以通过 Service 的方法来更新，但会增加复杂度
	node.ConsecutiveSuccess = consecutiveSuccess
	node.ConsecutiveFailure = consecutiveFailure
	node.LastHealthCheck = now

	// 如果健康状态需要更新，通过 ServiceManager 更新节点健康状态（这个方法内部会加锁）
	if shouldUpdateHealth {
		_ = s.serviceManager.UpdateNodeHealth(serviceID, node.ID, newHealth)
	}
}

// doHealthCheck 执行实际的健康检查
func (s *SharedHealthCheckerManager) doHealthCheck(node *NodeConfig, config *HealthConfig) bool {
	if config == nil || !config.Enabled {
		return true
	}

	// 构建健康检查URL
	url := node.URL + config.Path

	// 创建请求（使用 context 控制超时）
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, config.Method, url, nil)
	if err != nil {
		return false
	}

	// 添加自定义头部
	for key, value := range config.Headers {
		req.Header.Set(key, value)
	}

	// 执行请求
	resp, err := s.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// 检查状态码
	for _, expectedCode := range config.ExpectedStatusCodes {
		if resp.StatusCode == expectedCode {
			return true
		}
	}

	return false
}

// GetStats 获取统计信息
func (s *SharedHealthCheckerManager) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["running"] = s.running
	stats["workers"] = s.workers
	stats["active_workers"] = len(s.workerPool)

	// 如果 ServiceManager 可用，添加服务统计信息
	if s.serviceManager != nil {
		services := s.serviceManager.ListServices()
		stats["total_services"] = len(services)
	}

	return stats
}
