package health

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"gateway/internal/registry/core"
)

// HealthCheckConfig 健康检查配置
type HealthCheckConfig struct {
	Enabled          bool          `yaml:"enabled" json:"enabled"`
	Interval         time.Duration `yaml:"interval" json:"interval"`
	Timeout          time.Duration `yaml:"timeout" json:"timeout"`
	MaxRetries       int           `yaml:"maxRetries" json:"maxRetries"`
	RetryInterval    time.Duration `yaml:"retryInterval" json:"retryInterval"`
	ConcurrentChecks int           `yaml:"concurrentChecks" json:"concurrentChecks"`
	FailureThreshold int           `yaml:"failureThreshold" json:"failureThreshold"`
	SuccessThreshold int           `yaml:"successThreshold" json:"successThreshold"`
	DefaultPath      string        `yaml:"defaultPath" json:"defaultPath"`
	EnableTCPCheck   bool          `yaml:"enableTCPCheck" json:"enableTCPCheck"`
	EnableHTTPCheck  bool          `yaml:"enableHTTPCheck" json:"enableHTTPCheck"`
}

// DefaultHealthCheckConfig 默认健康检查配置
func DefaultHealthCheckConfig() *HealthCheckConfig {
	return &HealthCheckConfig{
		Enabled:          true,
		Interval:         30 * time.Second,
		Timeout:          5 * time.Second,
		MaxRetries:       3,
		RetryInterval:    5 * time.Second,
		ConcurrentChecks: 10,
		FailureThreshold: 3,
		SuccessThreshold: 1,
		DefaultPath:      "/health",
		EnableTCPCheck:   true,
		EnableHTTPCheck:  true,
	}
}

// Checker 健康检查器实现
type Checker struct {
	config         *HealthCheckConfig
	storage        core.Storage
	eventPublisher core.EventPublisher
	instances      map[string]*instanceState
	instanceMutex  sync.RWMutex
	checkChan      chan *core.ServiceInstance
	resultChan     chan *core.HealthCheckResult
	workers        []*worker
	running        bool
	mutex          sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	stats          *core.HealthCheckStats
	statsMutex     sync.RWMutex
}

// instanceState 实例状态
type instanceState struct {
	instance           *core.ServiceInstance
	consecutiveFails   int
	consecutiveSuccess int
	lastCheckTime      time.Time
	lastStatus         string
}

// worker 工作协程
type worker struct {
	id      int
	checker *Checker
	client  *http.Client
}

// NewChecker 创建健康检查器
func NewChecker(config *HealthCheckConfig, storage core.Storage, eventPublisher core.EventPublisher) *Checker {
	if config == nil {
		config = DefaultHealthCheckConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	c := &Checker{
		config:         config,
		storage:        storage,
		eventPublisher: eventPublisher,
		instances:      make(map[string]*instanceState),
		checkChan:      make(chan *core.ServiceInstance, config.ConcurrentChecks*2),
		resultChan:     make(chan *core.HealthCheckResult, config.ConcurrentChecks*2),
		ctx:            ctx,
		cancel:         cancel,
		stats: &core.HealthCheckStats{
			LastCheckTime: time.Now(),
		},
	}

	// 创建工作协程
	c.workers = make([]*worker, config.ConcurrentChecks)
	for i := 0; i < config.ConcurrentChecks; i++ {
		c.workers[i] = &worker{
			id:      i,
			checker: c,
			client: &http.Client{
				Timeout: config.Timeout,
			},
		}
	}

	return c
}

// Start 启动健康检查
func (c *Checker) Start(ctx context.Context) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.running {
		return nil
	}

	if !c.config.Enabled {
		return nil
	}

	c.running = true

	// 启动工作协程
	for _, worker := range c.workers {
		go worker.start()
	}

	// 启动结果处理协程
	go c.processResults()

	// 启动定时检查协程
	go c.scheduleChecks()

	return nil
}

// Stop 停止健康检查
func (c *Checker) Stop() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.running {
		return nil
	}

	c.running = false

	// 取消上下文
	c.cancel()

	// 关闭通道
	close(c.checkChan)
	close(c.resultChan)

	return nil
}

// CheckInstance 检查单个实例
func (c *Checker) CheckInstance(ctx context.Context, instance *core.ServiceInstance) *core.HealthCheckResult {
	if instance == nil {
		return core.NewHealthCheckResult("", core.HealthStatusUnhealthy, 0, fmt.Errorf("instance is nil"))
	}

	startTime := time.Now()

	// 选择检查方法
	var err error
	if c.config.EnableHTTPCheck {
		err = c.checkHTTP(ctx, instance)
	} else if c.config.EnableTCPCheck {
		err = c.checkTCP(ctx, instance)
	} else {
		err = fmt.Errorf("no health check method enabled")
	}

	responseTime := time.Since(startTime)
	status := core.HealthStatusHealthy
	if err != nil {
		status = core.HealthStatusUnhealthy
	}

	// 更新统计信息
	c.updateStats(status == core.HealthStatusHealthy, responseTime)

	return core.NewHealthCheckResult(instance.ServiceInstanceId, status, responseTime, err)
}

// CheckInstances 批量检查实例
func (c *Checker) CheckInstances(ctx context.Context, instances []*core.ServiceInstance) []*core.HealthCheckResult {
	results := make([]*core.HealthCheckResult, len(instances))

	// 使用goroutine并发检查
	var wg sync.WaitGroup
	for i, instance := range instances {
		wg.Add(1)
		go func(idx int, inst *core.ServiceInstance) {
			defer wg.Done()
			results[idx] = c.CheckInstance(ctx, inst)
		}(i, instance)
	}

	wg.Wait()
	return results
}

// AddInstance 添加实例到检查列表
func (c *Checker) AddInstance(instance *core.ServiceInstance) error {
	if instance == nil {
		return fmt.Errorf("instance is nil")
	}

	c.instanceMutex.Lock()
	defer c.instanceMutex.Unlock()

	c.instances[instance.ServiceInstanceId] = &instanceState{
		instance:      instance,
		lastStatus:    instance.HealthStatus,
		lastCheckTime: time.Now(),
	}

	c.updateInstanceStats()
	return nil
}

// RemoveInstance 从检查列表移除实例
func (c *Checker) RemoveInstance(instanceId string) error {
	c.instanceMutex.Lock()
	defer c.instanceMutex.Unlock()

	delete(c.instances, instanceId)
	c.updateInstanceStats()
	return nil
}

// GetStats 获取健康检查统计
func (c *Checker) GetStats() *core.HealthCheckStats {
	c.statsMutex.RLock()
	defer c.statsMutex.RUnlock()

	// 复制统计信息
	stats := *c.stats
	return &stats
}

// ================== 内部方法 ==================

// scheduleChecks 定时检查
func (c *Checker) scheduleChecks() {
	ticker := time.NewTicker(c.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.performChecks()
		case <-c.ctx.Done():
			return
		}
	}
}

// performChecks 执行检查
func (c *Checker) performChecks() {
	c.instanceMutex.RLock()
	instances := make([]*core.ServiceInstance, 0, len(c.instances))
	for _, state := range c.instances {
		instances = append(instances, state.instance)
	}
	c.instanceMutex.RUnlock()

	// 发送检查任务
	for _, instance := range instances {
		select {
		case c.checkChan <- instance:
		default:
			// 检查通道满了，跳过这次检查
			fmt.Printf("health check channel full, skipping instance: %s\n", instance.ServiceInstanceId)
		}
	}
}

// processResults 处理检查结果
func (c *Checker) processResults() {
	for result := range c.resultChan {
		c.handleResult(result)
	}
}

// handleResult 处理单个结果
func (c *Checker) handleResult(result *core.HealthCheckResult) {
	c.instanceMutex.Lock()
	state, exists := c.instances[result.InstanceId]
	if !exists {
		c.instanceMutex.Unlock()
		return
	}

	oldStatus := state.lastStatus
	state.lastCheckTime = result.CheckTime

	// 更新连续失败/成功计数
	if result.Status == core.HealthStatusHealthy {
		state.consecutiveSuccess++
		state.consecutiveFails = 0
	} else {
		state.consecutiveFails++
		state.consecutiveSuccess = 0
	}

	// 判断是否需要更新状态
	var newStatus string
	statusChanged := false

	if result.Status == core.HealthStatusHealthy && state.consecutiveSuccess >= c.config.SuccessThreshold {
		newStatus = core.HealthStatusHealthy
		if oldStatus != newStatus {
			statusChanged = true
		}
	} else if result.Status == core.HealthStatusUnhealthy && state.consecutiveFails >= c.config.FailureThreshold {
		newStatus = core.HealthStatusUnhealthy
		if oldStatus != newStatus {
			statusChanged = true
		}
	} else {
		newStatus = oldStatus
	}

	state.lastStatus = newStatus
	result.StatusChanged = statusChanged

	c.instanceMutex.Unlock()

	// 如果状态发生变化，更新存储并发布事件
	if statusChanged {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// 更新存储中的健康状态
		if err := c.storage.UpdateInstanceHealth(ctx, state.instance.TenantId, result.InstanceId, newStatus); err != nil {
			fmt.Printf("update instance health failed: %v\n", err)
		}

		// 发布健康状态变更事件
		if c.eventPublisher != nil {
			event := core.NewServiceEvent(
				state.instance.TenantId,
				core.EventTypeInstanceHealthChange,
				state.instance.ServiceName,
				state.instance.GroupName,
				"health-checker",
				fmt.Sprintf("Instance %s health changed from %s to %s", result.InstanceId, oldStatus, newStatus),
			)
			event.HostAddress = state.instance.HostAddress
			event.PortNumber = &state.instance.PortNumber

			if err := c.eventPublisher.Publish(ctx, event); err != nil {
				fmt.Printf("publish health change event failed: %v\n", err)
			}
		}
	}
}

// checkHTTP HTTP健康检查
func (c *Checker) checkHTTP(ctx context.Context, instance *core.ServiceInstance) error {
	// 构建健康检查URL
	url := instance.GetURL()

	// 从元数据获取健康检查路径
	metadata := instance.GetMetadata()
	healthPath := metadata["healthPath"]
	if healthPath == "" {
		healthPath = c.config.DefaultPath
	}

	if healthPath != "" && healthPath != "/" {
		url += healthPath
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}

	// 设置请求头
	req.Header.Set("User-Agent", "Registry-HealthChecker/1.0")
	req.Header.Set("Accept", "application/json,text/plain")

	// 执行请求
	client := &http.Client{Timeout: c.config.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unhealthy status code: %d", resp.StatusCode)
	}

	return nil
}

// checkTCP TCP健康检查
func (c *Checker) checkTCP(ctx context.Context, instance *core.ServiceInstance) error {
	address := fmt.Sprintf("%s:%d", instance.HostAddress, instance.PortNumber)

	dialer := &net.Dialer{Timeout: c.config.Timeout}
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return fmt.Errorf("tcp connection failed: %w", err)
	}
	defer conn.Close()

	return nil
}

// updateStats 更新统计信息
func (c *Checker) updateStats(success bool, responseTime time.Duration) {
	c.statsMutex.Lock()
	defer c.statsMutex.Unlock()

	c.stats.CheckCount++
	c.stats.LastCheckTime = time.Now()

	if success {
		// 更新平均响应时间
		if c.stats.AverageResponseTime == 0 {
			c.stats.AverageResponseTime = responseTime
		} else {
			c.stats.AverageResponseTime = (c.stats.AverageResponseTime + responseTime) / 2
		}
	} else {
		c.stats.ErrorCount++
	}
}

// updateInstanceStats 更新实例统计
func (c *Checker) updateInstanceStats() {
	c.statsMutex.Lock()
	defer c.statsMutex.Unlock()

	c.stats.TotalInstances = len(c.instances)
	c.stats.HealthyInstances = 0
	c.stats.UnhealthyInstances = 0

	for _, state := range c.instances {
		if state.lastStatus == core.HealthStatusHealthy {
			c.stats.HealthyInstances++
		} else {
			c.stats.UnhealthyInstances++
		}
	}
}

// ================== 工作协程方法 ==================

// start 启动工作协程
func (w *worker) start() {
	for instance := range w.checker.checkChan {
		// 执行健康检查
		result := w.checker.CheckInstance(w.checker.ctx, instance)

		// 发送结果
		select {
		case w.checker.resultChan <- result:
		default:
			// 结果通道满了，跳过
			fmt.Printf("health check result channel full, dropping result for instance: %s\n", result.InstanceId)
		}
	}
}

// IsRunning 检查是否运行中
func (c *Checker) IsRunning() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.running
}

// LoadInstances 从存储加载实例
func (c *Checker) LoadInstances(ctx context.Context, tenantId string) error {
	instances, err := c.storage.ListAllInstances(ctx, tenantId)
	if err != nil {
		return fmt.Errorf("load instances failed: %w", err)
	}

	c.instanceMutex.Lock()
	defer c.instanceMutex.Unlock()

	// 清空现有实例
	c.instances = make(map[string]*instanceState)

	// 添加新实例
	for _, instance := range instances {
		if instance.IsActive() {
			c.instances[instance.ServiceInstanceId] = &instanceState{
				instance:      instance,
				lastStatus:    instance.HealthStatus,
				lastCheckTime: time.Now(),
			}
		}
	}

	c.updateInstanceStats()
	return nil
}

// GetInstanceCount 获取实例数量
func (c *Checker) GetInstanceCount() int {
	c.instanceMutex.RLock()
	defer c.instanceMutex.RUnlock()
	return len(c.instances)
}
