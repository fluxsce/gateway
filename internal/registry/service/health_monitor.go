package service

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"gateway/internal/registry/core"
	"gateway/pkg/logger"
	"gateway/pkg/utils/random"
)

// HealthMonitor 健康监控组件
// 负责监控服务实例的健康状态
//
// 主要功能：
//   - 定期检查服务实例的健康状态（支持HTTP和TCP两种检查方式）
//   - 支持主动模式和被动模式的健康检查
//   - 自动删除不健康的临时实例（非持久化实例）
//   - 发布健康状态变更和实例删除事件
//   - 提供健康检查统计信息
type HealthMonitor struct {
	cache          core.CacheStorage
	eventPublisher core.EventPublisher

	// 默认配置（仅在服务未指定配置时使用）
	defaultCheckInterval time.Duration
	defaultTimeout       time.Duration

	// 监控状态
	isRunning bool
	stopChan  chan struct{}
	wg        sync.WaitGroup

	// 监控统计
	stats      core.HealthCheckStats
	statsMutex sync.RWMutex

	// 服务跟踪
	tenantsAndServices map[string]bool
	trackMutex         sync.RWMutex
}

// NewHealthMonitor 创建健康监控组件
func NewHealthMonitor(cache core.CacheStorage, eventPublisher core.EventPublisher) *HealthMonitor {
	return &HealthMonitor{
		cache:                cache,
		eventPublisher:       eventPublisher,
		defaultCheckInterval: 30 * time.Second, // 默认30秒检查一次（仅当服务未配置时使用）
		defaultTimeout:       5 * time.Second,  // 默认5秒超时（仅当服务未配置时使用）
		stopChan:             make(chan struct{}),
		tenantsAndServices:   make(map[string]bool),
	}
}

// Start 启动健康监控
// 实现 HealthChecker 接口
func (m *HealthMonitor) Start(ctx context.Context) error {
	if m.isRunning {
		return fmt.Errorf("健康监控已经启动")
	}

	m.isRunning = true
	m.stopChan = make(chan struct{})

	// 启动监控协程
	m.wg.Add(1)
	go m.monitorLoop(ctx)

	logger.InfoWithTrace(ctx, "健康监控已启动",
		"checkInterval", m.defaultCheckInterval.String(),
		"timeout", m.defaultTimeout.String())

	return nil
}

// Stop 停止健康监控
// 实现 HealthChecker 接口
func (m *HealthMonitor) Stop(ctx context.Context) error {
	if !m.isRunning {
		return nil
	}

	// 发送停止信号
	close(m.stopChan)

	// 等待监控协程退出
	waitChan := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(waitChan)
	}()

	// 等待超时
	select {
	case <-waitChan:
		// 正常退出
	case <-time.After(10 * time.Second):
		// 超时
		logger.WarnWithTrace(ctx, "健康监控停止超时")
	}

	m.isRunning = false
	logger.InfoWithTrace(ctx, "健康监控已停止")

	return nil
}

// SetCheckInterval 设置默认健康检查间隔
// 实现 HealthChecker 接口
// 注意：此设置仅影响未指定检查间隔的服务
func (m *HealthMonitor) SetCheckInterval(interval time.Duration) {
	if interval > 0 {
		m.defaultCheckInterval = interval
		logger.Info("设置默认健康检查间隔", "interval", interval.String())
	}
}

// GetStats 获取健康检查统计信息
// 实现 HealthChecker 接口
func (m *HealthMonitor) GetStats() core.HealthCheckStats {
	m.statsMutex.RLock()
	defer m.statsMutex.RUnlock()

	// 返回副本
	return core.HealthCheckStats{
		TotalChecks:     m.stats.TotalChecks,
		SuccessChecks:   m.stats.SuccessChecks,
		FailedChecks:    m.stats.FailedChecks,
		SuccessRate:     m.stats.SuccessRate,
		AvgResponseTime: m.stats.AvgResponseTime,
		ActiveInstances: m.stats.ActiveInstances,
	}
}

// monitorLoop 监控循环
func (m *HealthMonitor) monitorLoop(ctx context.Context) {
	defer m.wg.Done()

	// 使用最小间隔作为基础时钟周期，确保及时响应
	// 实际检查由各服务自己的配置控制
	baseInterval := m.defaultCheckInterval
	if baseInterval > 10*time.Second {
		baseInterval = 10 * time.Second // 最短10秒一次检查循环，避免过于频繁
	}

	ticker := time.NewTicker(baseInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 执行健康检查
			m.checkAllServices(ctx)
		case <-m.stopChan:
			// 收到停止信号
			return
		}
	}
}

// checkAllServices 检查所有服务的健康状态
func (m *HealthMonitor) checkAllServices(ctx context.Context) {
	// 从缓存中获取所有服务
	services, err := m.getAllServicesFromCache(ctx)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取服务列表失败", "error", err)
		return
	}

	var successCount, failedCount int64
	var activeInstances int64

	// 遍历所有服务
	for _, service := range services {
		// 判断服务的健康检查模式
		isPassiveMode := service.HealthCheckMode == "PASSIVE"

		if isPassiveMode {
			// 被动模式(PASSIVE)：跳过主动HTTP健康检查，但仍需检查临时实例的心跳超时
			logger.DebugWithTrace(ctx, "被动模式：仅检查临时实例心跳超时",
				"tenantId", service.TenantId,
				"serviceName", service.ServiceName,
				"mode", service.HealthCheckMode)

			// 获取服务实例列表
			instances, err := m.getServiceInstancesFromCache(ctx, service.TenantId, service.ServiceName)
			if err != nil {
				logger.ErrorWithTrace(ctx, "获取服务实例列表失败",
					"tenantId", service.TenantId,
					"serviceName", service.ServiceName,
					"error", err)
				continue
			}

			activeInstances += int64(len(instances))

			// 在被动模式下检查临时实例的心跳超时
			passiveSuccess, passiveFailed := m.checkPassiveModeHeartbeat(ctx, instances, service)
			successCount += passiveSuccess
			failedCount += passiveFailed

			continue
		}

		// 获取服务实例列表
		instances, err := m.getServiceInstancesFromCache(ctx, service.TenantId, service.ServiceName)
		if err != nil {
			logger.ErrorWithTrace(ctx, "获取服务实例列表失败",
				"tenantId", service.TenantId,
				"serviceName", service.ServiceName,
				"error", err)
			continue
		}

		activeInstances += int64(len(instances))

		// 主动模式：进行完整的健康检查（HTTP检查 + 心跳超时检查）
		for _, instance := range instances {
			isHealthy, err := m.checkInstanceHealth(ctx, instance, service)
			if isHealthy {
				successCount++
			} else {
				failedCount++
				logger.WarnWithTrace(ctx, "实例健康检查失败",
					"instanceId", instance.ServiceInstanceId,
					"serviceName", instance.ServiceName,
					"error", err)
			}
		}
	}

	// 更新统计信息
	m.updateStats(successCount, failedCount)
	m.updateActiveInstances(int(activeInstances))
}

// getAllServicesFromCache 从缓存中获取所有服务
func (m *HealthMonitor) getAllServicesFromCache(ctx context.Context) ([]*core.Service, error) {
	// 我们现在无法直接获取所有租户ID，因为接口中没有ListTenants方法
	// 这里使用预设的默认租户ID "default"
	tenantId := "default"

	var allServices []*core.Service

	// 获取默认租户下所有服务组
	serviceGroups, err := m.cache.ListServiceGroups(ctx, tenantId)
	if err != nil {
		logger.WarnWithTrace(ctx, "获取服务组列表失败", "tenantId", tenantId, "error", err)
		return nil, fmt.Errorf("获取服务组列表失败: %w", err)
	}

	// 遍历每个服务组，获取其下的所有服务
	for _, group := range serviceGroups {
		// 获取服务组下所有服务
		services, err := m.cache.ListServices(ctx, tenantId, group.ServiceGroupId)
		if err != nil {
			logger.WarnWithTrace(ctx, "获取服务列表失败",
				"tenantId", tenantId,
				"groupId", group.ServiceGroupId,
				"error", err)
			continue
		}

		// 添加到结果列表
		allServices = append(allServices, services...)
	}

	return allServices, nil
}

// getServiceInstancesFromCache 从缓存中获取服务实例列表
func (m *HealthMonitor) getServiceInstancesFromCache(ctx context.Context, tenantId, serviceName string) ([]*core.ServiceInstance, error) {
	// 首先获取该租户下所有服务组
	serviceGroups, err := m.cache.ListServiceGroups(ctx, tenantId)
	if err != nil {
		return nil, fmt.Errorf("获取服务组列表失败: %w", err)
	}

	// 遍历服务组，查找指定服务名的服务
	for _, group := range serviceGroups {
		// 查找服务
		service, err := m.cache.GetService(ctx, tenantId, group.ServiceGroupId, serviceName)
		if err != nil {
			logger.DebugWithTrace(ctx, "在服务组中查找服务失败",
				"tenantId", tenantId,
				"groupId", group.ServiceGroupId,
				"serviceName", serviceName,
				"error", err)
			continue // 尝试下一个服务组
		}

		if service != nil {
			// 找到服务，获取实例列表
			return m.cache.ListInstances(ctx, tenantId, group.ServiceGroupId, serviceName)
		}
	}

	// 未找到服务
	return nil, fmt.Errorf("未找到服务: %s", serviceName)
}

// checkInstanceHealth 检查实例健康状态
// 对于临时实例（TempInstanceFlag == "Y"），如果连续三次健康检查失败或心跳超时，会自动删除该实例
func (m *HealthMonitor) checkInstanceHealth(ctx context.Context, instance *core.ServiceInstance, service *core.Service) (bool, error) {
	if instance == nil {
		return false, fmt.Errorf("实例不能为空")
	}

	// 检查临时实例的心跳超时
	if instance.TempInstanceFlag == "Y" {
		if m.isHeartbeatTimeout(instance, service) {
			logger.InfoWithTrace(ctx, "临时实例心跳超时，准备删除",
				"instanceId", instance.ServiceInstanceId,
				"serviceName", instance.ServiceName,
				"lastHeartbeatTime", instance.LastHeartbeatTime,
				"heartbeatFailCount", instance.HeartbeatFailCount)

			// 心跳超时，删除临时实例
			return false, m.deleteUnhealthyTemporaryInstance(ctx, instance, instance.HealthStatus, "心跳超时")
		}
	}

	// 判断是否到达检查时间
	// 根据服务配置的健康检查间隔来决定是否需要进行检查
	if !m.shouldCheckHealth(instance, service) {
		logger.DebugWithTrace(ctx, "实例尚未到达下一次检查时间",
			"instanceId", instance.ServiceInstanceId,
			"serviceName", instance.ServiceName)
		return true, nil // 不进行检查，返回默认健康状态
	}

	// 构建健康检查URL
	healthCheckUrl := m.buildHealthCheckUrl(instance, service)
	if healthCheckUrl == "" {
		// 无法构建健康检查URL，标记为未知状态
		err := m.updateInstanceHealth(ctx, instance, core.HealthStatusUnknown)
		return false, err
	}

	// 设置特定于该服务的HTTP客户端超时
	timeout := m.getHealthCheckTimeout(service)
	httpClient := &http.Client{
		Timeout: timeout,
	}

	// 执行HTTP健康检查
	startTime := time.Now()
	isHealthy, httpResponse, err := m.doHttpHealthCheckWithResponse(ctx, httpClient, healthCheckUrl, service.HealthCheckType)
	duration := time.Since(startTime)

	// 更新响应时间统计
	m.updateResponseTime(duration)

	// 更新实例健康状态
	var healthStatus string
	if isHealthy {
		healthStatus = core.HealthStatusHealthy
	} else {
		healthStatus = core.HealthStatusUnhealthy
	}

	// 记录健康检查结果
	logger.DebugWithTrace(ctx, "健康检查结果",
		"instanceId", instance.ServiceInstanceId,
		"serviceName", instance.ServiceName,
		"url", healthCheckUrl,
		"type", service.HealthCheckType,
		"timeout", timeout.String(),
		"status", healthStatus,
		"duration", duration.String(),
		"httpResponse", httpResponse,
		"error", err)

	// 更新实例健康状态，并传递HTTP响应信息用于事件记录
	updateErr := m.updateInstanceHealthWithResponse(ctx, instance, healthStatus, httpResponse, err, healthCheckUrl)
	if updateErr != nil {
		return isHealthy, updateErr
	}

	// 发布健康检查执行事件（无论状态是否变更都发布）
	if m.eventPublisher != nil {
		m.publishHealthCheckEvent(ctx, instance, healthStatus, httpResponse, err, duration, healthCheckUrl)
	}

	return isHealthy, err
}

// buildHealthCheckUrl 构建健康检查URL
func (m *HealthMonitor) buildHealthCheckUrl(instance *core.ServiceInstance, service *core.Service) string {
	// 如果服务配置了健康检查URL
	if service != nil && service.HealthCheckUrl != "" {
		url := fmt.Sprintf("http://%s:%d%s",
			instance.HostAddress,
			instance.PortNumber,
			service.HealthCheckUrl)
		return url
	}

	// 默认健康检查URL
	return fmt.Sprintf("http://%s:%d/health",
		instance.HostAddress,
		instance.PortNumber)
}

// doHttpHealthCheckWithResponse 执行HTTP健康检查并返回响应信息
func (m *HealthMonitor) doHttpHealthCheckWithResponse(ctx context.Context, httpClient *http.Client, url string, checkType string) (bool, string, error) {
	// 根据检查类型执行不同的健康检查
	switch checkType {
	case "TCP":
		// TCP模式：直接建立TCP连接到目标地址验证可用性
		isHealthy, err := m.doTcpHealthCheck(url, m.defaultTimeout)
		response := fmt.Sprintf("TCP连接检查: %s", url)
		if err != nil {
			response += fmt.Sprintf(", 错误: %s", err.Error())
		}
		return isHealthy, response, err
	case "HTTP", "":
		// HTTP模式：发送HTTP GET请求并验证返回状态码
		return m.doHttpGetHealthCheckWithResponse(ctx, httpClient, url)
	default:
		// 未知的检查类型，默认使用HTTP
		logger.WarnWithTrace(ctx, "未知的健康检查类型，使用HTTP", "type", checkType)
		return m.doHttpGetHealthCheckWithResponse(ctx, httpClient, url)
	}
}

// doHttpGetHealthCheckWithResponse 执行HTTP GET健康检查并返回响应信息
func (m *HealthMonitor) doHttpGetHealthCheckWithResponse(ctx context.Context, httpClient *http.Client, url string) (bool, string, error) {
	// 创建带上下文的HTTP请求
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		response := fmt.Sprintf("创建HTTP请求失败: %s", err.Error())
		return false, response, err
	}

	// 设置通用请求头
	req.Header.Set("User-Agent", "Gateway-HealthChecker/1.0")
	req.Header.Set("Accept", "*/*")

	// 执行请求
	resp, err := httpClient.Do(req)
	if err != nil {
		response := fmt.Sprintf("HTTP请求执行失败: %s", err.Error())
		return false, response, err
	}
	defer resp.Body.Close()

	// 构建响应信息
	response := fmt.Sprintf("HTTP %d %s", resp.StatusCode, resp.Status)
	if resp.Header.Get("Content-Type") != "" {
		response += fmt.Sprintf(", Content-Type: %s", resp.Header.Get("Content-Type"))
	}
	if resp.Header.Get("Content-Length") != "" {
		response += fmt.Sprintf(", Content-Length: %s", resp.Header.Get("Content-Length"))
	}

	// 检查HTTP状态码
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true, response, nil
	}

	err = fmt.Errorf("健康检查失败，状态码: %d", resp.StatusCode)
	return false, response, err
}

// doTcpHealthCheck 执行TCP健康检查
func (m *HealthMonitor) doTcpHealthCheck(address string, timeout time.Duration) (bool, error) {
	// 从URL中提取主机和端口
	// 例如从 "http://example.com:8080/health" 提取 "example.com:8080"
	u, err := url.Parse(address)
	if err != nil {
		return false, fmt.Errorf("解析TCP地址失败: %w", err)
	}

	// 确保主机部分包含端口
	host := u.Host
	if !strings.Contains(host, ":") {
		// 根据协议添加默认端口
		if u.Scheme == "https" {
			host = host + ":443"
		} else {
			host = host + ":80"
		}
	}

	// 尝试建立TCP连接
	conn, err := net.DialTimeout("tcp", host, timeout)
	if err != nil {
		return false, err
	}
	defer conn.Close()

	// 连接成功
	return true, nil
}

// updateInstanceHealthWithResponse 更新实例健康状态并记录HTTP响应信息
// 对于临时实例（非持久化实例），当连续三次健康检查失败时，会自动删除实例
func (m *HealthMonitor) updateInstanceHealthWithResponse(ctx context.Context, instance *core.ServiceInstance, status string, httpResponse string, healthCheckError error, healthCheckUrl string) error {
	// 记录原始健康状态以便判断是否发生变化
	oldHealthStatus := instance.HealthStatus

	// 检查是否为临时实例且状态变为不健康
	isTemporaryInstance := instance.TempInstanceFlag == "Y"
	isUnhealthy := status == core.HealthStatusUnhealthy

	if isTemporaryInstance {
		if isUnhealthy {
			// 增加失败次数
			instance.HeartbeatFailCount++
			logger.DebugWithTrace(ctx, "临时实例健康检查失败，增加失败计数",
				"instanceId", instance.ServiceInstanceId,
				"serviceName", instance.ServiceName,
				"heartbeatFailCount", instance.HeartbeatFailCount)

			// 检查是否连续三次失败
			if instance.HeartbeatFailCount >= 3 {
				logger.InfoWithTrace(ctx, "临时实例连续三次健康检查失败，准备删除",
					"instanceId", instance.ServiceInstanceId,
					"serviceName", instance.ServiceName,
					"heartbeatFailCount", instance.HeartbeatFailCount)

				// 连续三次失败，删除临时实例
				return m.deleteUnhealthyTemporaryInstance(ctx, instance, oldHealthStatus, fmt.Sprintf("连续%d次健康检查失败", instance.HeartbeatFailCount))
			}
		} else {
			// 健康检查成功，重置失败次数
			if instance.HeartbeatFailCount > 0 {
				logger.DebugWithTrace(ctx, "临时实例健康检查成功，重置失败计数",
					"instanceId", instance.ServiceInstanceId,
					"serviceName", instance.ServiceName,
					"oldFailCount", instance.HeartbeatFailCount)
				instance.HeartbeatFailCount = 0
			}
		}
	}

	// 对于非临时实例或健康的实例，按正常流程更新健康状态
	instance.HealthStatus = status
	now := time.Now()
	instance.LastHealthCheckTime = &now

	// 主动模式下始终更新心跳时间
	// 无论健康状态如何，主动健康检查都表明实例仍在运行
	// 这样可以防止临时实例因健康检查暂时失败而被误删
	if instance.LastHeartbeatTime == nil || time.Since(*instance.LastHeartbeatTime) > m.defaultCheckInterval {
		instance.LastHeartbeatTime = &now
		logger.DebugWithTrace(ctx, "主动模式下更新实例心跳时间",
			"instanceId", instance.ServiceInstanceId,
			"serviceName", instance.ServiceName,
			"healthStatus", status)
	}

	// 更新缓存中的实例
	err := m.cache.SetInstance(ctx, instance.TenantId, instance)
	if err != nil {
		return fmt.Errorf("更新实例健康状态失败: %w", err)
	}

	// 如果健康状态发生变化，发布健康状态变更事件
	if oldHealthStatus != status && m.eventPublisher != nil {
		m.publishHealthChangeEventWithResponse(ctx, instance, oldHealthStatus, status, httpResponse, healthCheckError, healthCheckUrl)
	}

	return nil
}

// updateInstanceHealth 更新实例健康状态（保留原方法以兼容其他地方的调用）
// 对于临时实例（非持久化实例），当健康状态变为不健康时，会自动删除实例
func (m *HealthMonitor) updateInstanceHealth(ctx context.Context, instance *core.ServiceInstance, status string) error {
	return m.updateInstanceHealthWithResponse(ctx, instance, status, "", nil, "")
}

// updateStats 更新健康检查统计信息
func (m *HealthMonitor) updateStats(success, failed int64) {
	m.statsMutex.Lock()
	defer m.statsMutex.Unlock()

	m.stats.TotalChecks += success + failed
	m.stats.SuccessChecks += success
	m.stats.FailedChecks += failed

	// 计算成功率
	if m.stats.TotalChecks > 0 {
		m.stats.SuccessRate = float64(m.stats.SuccessChecks) / float64(m.stats.TotalChecks)
	}
}

// updateResponseTime 更新平均响应时间
func (m *HealthMonitor) updateResponseTime(duration time.Duration) {
	m.statsMutex.Lock()
	defer m.statsMutex.Unlock()

	// 简单移动平均算法
	if m.stats.AvgResponseTime == 0 {
		m.stats.AvgResponseTime = int64(duration / time.Millisecond)
	} else {
		// 权重平均，新值占20%
		m.stats.AvgResponseTime = int64(float64(m.stats.AvgResponseTime)*0.8 + float64(duration/time.Millisecond)*0.2)
	}
}

// updateActiveInstances 更新活跃实例数
func (m *HealthMonitor) updateActiveInstances(count int) {
	m.statsMutex.Lock()
	defer m.statsMutex.Unlock()

	m.stats.ActiveInstances = int64(count)
}

// SetTimeout 设置默认健康检查超时时间
// 内部方法，不是接口的一部分
// 注意：此设置仅影响未指定超时时间的服务
func (m *HealthMonitor) SetTimeout(timeout time.Duration) {
	if timeout > 0 {
		m.defaultTimeout = timeout
		logger.Info("设置默认健康检查超时时间", "timeout", timeout.String())
	}
}

// checkPassiveModeHeartbeat 在被动模式下检查临时实例心跳超时
// 被动模式下不进行HTTP健康检查，但需要监控临时实例的心跳状态
// 支持连续三次心跳超时后才删除临时实例
func (m *HealthMonitor) checkPassiveModeHeartbeat(ctx context.Context, instances []*core.ServiceInstance, service *core.Service) (int64, int64) {
	var successCount, failedCount int64

	for _, instance := range instances {
		// 只检查临时实例
		if instance.TempInstanceFlag != "Y" {
			continue
		}

		// 记录原始健康状态以便判断是否发生变化
		oldHealthStatus := instance.HealthStatus

		// 检查心跳是否超时
		if m.isHeartbeatTimeout(instance, service) {
			// 增加失败次数
			instance.HeartbeatFailCount++

			logger.DebugWithTrace(ctx, "被动模式下检测到临时实例心跳超时，增加失败计数",
				"instanceId", instance.ServiceInstanceId,
				"serviceName", instance.ServiceName,
				"lastHeartbeatTime", instance.LastHeartbeatTime,
				"heartbeatFailCount", instance.HeartbeatFailCount)

			// 更新健康状态为不健康
			newHealthStatus := core.HealthStatusUnhealthy
			instance.HealthStatus = newHealthStatus
			now := time.Now()
			instance.LastHealthCheckTime = &now

			// 检查是否连续三次心跳超时
			if instance.HeartbeatFailCount >= 3 {
				logger.InfoWithTrace(ctx, "被动模式下临时实例连续三次心跳超时，准备删除",
					"instanceId", instance.ServiceInstanceId,
					"serviceName", instance.ServiceName,
					"lastHeartbeatTime", instance.LastHeartbeatTime,
					"heartbeatFailCount", instance.HeartbeatFailCount)

				// 连续三次心跳超时，删除临时实例
				err := m.deleteUnhealthyTemporaryInstance(ctx, instance, oldHealthStatus, fmt.Sprintf("被动模式连续%d次心跳超时", instance.HeartbeatFailCount))
				if err != nil {
					logger.ErrorWithTrace(ctx, "删除连续心跳超时的临时实例失败",
						"instanceId", instance.ServiceInstanceId,
						"error", err)
					failedCount++
				} else {
					successCount++ // 成功删除也算一次成功操作
				}
			} else {
				// 未达到删除阈值，更新实例状态到缓存
				err := m.cache.SetInstance(ctx, instance.TenantId, instance)
				if err != nil {
					logger.WarnWithTrace(ctx, "更新实例心跳失败计数到缓存失败",
						"instanceId", instance.ServiceInstanceId,
						"error", err)
					failedCount++
				} else {
					successCount++ // 成功更新计数
				}

				// 发布健康状态变更事件（如果状态发生变化）
				if oldHealthStatus != newHealthStatus && m.eventPublisher != nil {
					m.publishHealthChangeEventWithResponse(ctx, instance, oldHealthStatus, newHealthStatus, "被动模式心跳超时检查", fmt.Errorf("心跳超时"), "")
				}

				// 发布心跳检查事件
				if m.eventPublisher != nil {
					m.publishPassiveModeHeartbeatEvent(ctx, instance, newHealthStatus, "心跳超时", fmt.Errorf("心跳超时"))
				}
			}
		} else {
			// 心跳正常，更新健康状态为健康
			newHealthStatus := core.HealthStatusHealthy
			instance.HealthStatus = newHealthStatus
			now := time.Now()
			instance.LastHealthCheckTime = &now

			// 重置失败次数
			if instance.HeartbeatFailCount > 0 {
				logger.DebugWithTrace(ctx, "被动模式下临时实例心跳正常，重置失败计数",
					"instanceId", instance.ServiceInstanceId,
					"serviceName", instance.ServiceName,
					"oldFailCount", instance.HeartbeatFailCount)

				instance.HeartbeatFailCount = 0
			}

			// 更新实例到缓存
			err := m.cache.SetInstance(ctx, instance.TenantId, instance)
			if err != nil {
				logger.WarnWithTrace(ctx, "更新实例状态到缓存失败",
					"instanceId", instance.ServiceInstanceId,
					"error", err)
				failedCount++
			} else {
				successCount++
			}

			// 发布健康状态变更事件（如果状态发生变化）
			if oldHealthStatus != newHealthStatus && m.eventPublisher != nil {
				m.publishHealthChangeEventWithResponse(ctx, instance, oldHealthStatus, newHealthStatus, "被动模式心跳正常检查", nil, "")
			}

			// 发布心跳检查事件
			if m.eventPublisher != nil {
				m.publishPassiveModeHeartbeatEvent(ctx, instance, newHealthStatus, "心跳正常", nil)
			}
		}
	}

	return successCount, failedCount
}

// isHeartbeatTimeout 检查临时实例是否心跳超时
// 临时实例需要定期发送心跳，如果超过指定时间没有心跳则认为超时
func (m *HealthMonitor) isHeartbeatTimeout(instance *core.ServiceInstance, service *core.Service) bool {
	// 如果从未收到心跳，不算超时（可能是刚注册的实例）
	if instance.LastHeartbeatTime == nil {
		return false
	}

	// 获取心跳超时时间，默认使用健康检查间隔的3倍作为心跳超时时间
	heartbeatTimeout := m.getHealthCheckInterval(service) * 3

	// 检查是否超时
	return time.Since(*instance.LastHeartbeatTime) > heartbeatTimeout
}

// shouldCheckHealth 判断是否应该检查实例健康状态
// 根据上次检查时间和服务配置的检查间隔来决定
func (m *HealthMonitor) shouldCheckHealth(instance *core.ServiceInstance, service *core.Service) bool {
	// 如果实例未设置上次健康检查时间，需要立即检查
	if instance.LastHealthCheckTime == nil {
		return true
	}

	// 获取服务配置的健康检查间隔
	checkInterval := m.getHealthCheckInterval(service)

	// 判断是否已经到达下一次检查时间
	return time.Since(*instance.LastHealthCheckTime) >= checkInterval
}

// getHealthCheckInterval 获取服务的健康检查间隔时间
func (m *HealthMonitor) getHealthCheckInterval(service *core.Service) time.Duration {
	// 如果服务配置了健康检查间隔，使用服务配置
	if service != nil && service.HealthCheckIntervalSeconds > 0 {
		return time.Duration(service.HealthCheckIntervalSeconds) * time.Second
	}

	// 否则使用默认检查间隔
	return m.defaultCheckInterval
}

// getHealthCheckTimeout 获取服务的健康检查超时时间
func (m *HealthMonitor) getHealthCheckTimeout(service *core.Service) time.Duration {
	// 如果服务配置了健康检查超时，使用服务配置
	if service != nil && service.HealthCheckTimeoutSeconds > 0 {
		return time.Duration(service.HealthCheckTimeoutSeconds) * time.Second
	}

	// 否则使用默认超时时间
	return m.defaultTimeout
}

// deleteUnhealthyTemporaryInstance 删除不健康的临时实例
// 当临时实例健康检查失败时，自动删除该实例并发布相应事件
func (m *HealthMonitor) deleteUnhealthyTemporaryInstance(ctx context.Context, instance *core.ServiceInstance, oldHealthStatus string, reason string) error {
	logger.InfoWithTrace(ctx, "检测到不健康的临时实例，准备删除",
		"instanceId", instance.ServiceInstanceId,
		"serviceName", instance.ServiceName,
		"hostAddress", instance.HostAddress,
		"portNumber", instance.PortNumber,
		"oldHealthStatus", oldHealthStatus,
		"tempInstanceFlag", instance.TempInstanceFlag,
		"heartbeatFailCount", instance.HeartbeatFailCount,
		"reason", reason)

	// 从缓存中删除实例
	err := m.cache.DeleteInstance(ctx, instance.TenantId, instance.ServiceInstanceId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "删除不健康临时实例失败",
			"instanceId", instance.ServiceInstanceId,
			"serviceName", instance.ServiceName,
			"error", err)
		return fmt.Errorf("删除不健康临时实例失败: %w", err)
	}

	logger.InfoWithTrace(ctx, "成功删除不健康的临时实例",
		"instanceId", instance.ServiceInstanceId,
		"serviceName", instance.ServiceName,
		"hostAddress", instance.HostAddress,
		"portNumber", instance.PortNumber,
		"heartbeatFailCount", instance.HeartbeatFailCount,
		"reason", reason)

	// 发布实例删除事件
	if m.eventPublisher != nil {
		m.publishInstanceDeregisterEvent(ctx, instance, reason)
	}

	return nil
}

// publishInstanceDeregisterEvent 发布实例注销事件
func (m *HealthMonitor) publishInstanceDeregisterEvent(ctx context.Context, instance *core.ServiceInstance, reason string) {
	// 生成事件ID
	eventId := random.Generate32BitRandomString()

	// 构建详细的事件消息
	eventMessage := fmt.Sprintf("临时实例自动删除: %s (失败次数: %d, 实例状态: %s, 健康状态: %s)",
		reason,
		instance.HeartbeatFailCount,
		instance.InstanceStatus,
		instance.HealthStatus)

	// 创建实例注销事件
	event := &core.ServiceEvent{
		ServiceEventId:    eventId,
		EventType:         core.EventTypeInstanceDeregistered,
		TenantId:          instance.TenantId,
		ServiceGroupId:    instance.ServiceGroupId,
		ServiceInstanceId: instance.ServiceInstanceId,
		GroupName:         instance.GroupName,
		ServiceName:       instance.ServiceName,
		HostAddress:       instance.HostAddress,
		PortNumber:        instance.PortNumber,
		NodeIpAddress:     random.GetNodeIP(),
		EventTime:         time.Now(),
		EventSource:       core.GetEventSourceFromContext(ctx, core.EventSourceHealthMonitor),
		EventMessage:      eventMessage,
		ActiveFlag:        "Y",
		Instance:          instance, // 添加实例对象用于事件处理时的数据传递
	}

	// 发布事件
	err := m.eventPublisher.Publish(ctx, event)
	if err != nil {
		logger.WarnWithTrace(ctx, "发布实例注销事件失败",
			"instanceId", instance.ServiceInstanceId,
			"serviceName", instance.ServiceName,
			"reason", reason,
			"heartbeatFailCount", instance.HeartbeatFailCount,
			"error", err)
	} else {
		logger.InfoWithTrace(ctx, "发布实例注销事件成功",
			"instanceId", instance.ServiceInstanceId,
			"serviceName", instance.ServiceName,
			"reason", reason,
			"heartbeatFailCount", instance.HeartbeatFailCount,
			"eventId", eventId)
	}
}

// publishHealthChangeEventWithResponse 发布健康状态变更事件（包含HTTP响应信息）
func (m *HealthMonitor) publishHealthChangeEventWithResponse(ctx context.Context, instance *core.ServiceInstance, oldStatus, newStatus string, httpResponse string, healthCheckError error, healthCheckUrl string) {
	// 生成事件ID
	eventId := random.Generate32BitRandomString()

	// 如果没有传递健康检查URL，使用默认URL
	if healthCheckUrl == "" {
		healthCheckUrl = fmt.Sprintf("http://%s:%d/health", instance.HostAddress, instance.PortNumber)
	}

	// 构建事件消息
	var eventMessage string
	if httpResponse != "" {
		eventMessage = fmt.Sprintf("实例健康状态从 %s 变更为 %s, 检查地址: %s, HTTP响应: %s", oldStatus, newStatus, healthCheckUrl, httpResponse)
		if healthCheckError != nil {
			eventMessage += fmt.Sprintf(", 错误: %s", healthCheckError.Error())
		}
	} else {
		eventMessage = fmt.Sprintf("实例健康状态从 %s 变更为 %s, 检查地址: %s", oldStatus, newStatus, healthCheckUrl)
		if healthCheckError != nil {
			eventMessage += fmt.Sprintf(", 错误: %s", healthCheckError.Error())
		}
	}

	// 创建健康状态变更事件
	event := &core.ServiceEvent{
		ServiceEventId:    eventId,
		EventType:         core.EventTypeInstanceHealthChange,
		TenantId:          instance.TenantId,
		ServiceGroupId:    instance.ServiceGroupId,
		ServiceInstanceId: instance.ServiceInstanceId,
		GroupName:         instance.GroupName,
		ServiceName:       instance.ServiceName,
		HostAddress:       instance.HostAddress,
		PortNumber:        instance.PortNumber,
		NodeIpAddress:     random.GetNodeIP(),
		EventTime:         time.Now(),
		EventSource:       core.GetEventSourceFromContext(ctx, core.EventSourceHealthMonitor),
		EventMessage:      eventMessage,
		ActiveFlag:        "Y",
		Instance:          instance, // 添加实例对象用于事件处理时的数据传递
	}

	// 发布事件
	err := m.eventPublisher.Publish(ctx, event)
	if err != nil {
		logger.WarnWithTrace(ctx, "发布健康状态变更事件失败",
			"instanceId", instance.ServiceInstanceId,
			"serviceName", instance.ServiceName,
			"oldStatus", oldStatus,
			"newStatus", newStatus,
			"httpResponse", httpResponse,
			"error", err)
	} else {
		logger.InfoWithTrace(ctx, "发布健康状态变更事件成功",
			"instanceId", instance.ServiceInstanceId,
			"serviceName", instance.ServiceName,
			"oldStatus", oldStatus,
			"newStatus", newStatus,
			"httpResponse", httpResponse,
			"eventId", eventId)
	}
}

// publishPassiveModeHeartbeatEvent 发布被动模式心跳检查事件
// 被动模式下每次心跳检查都会发布此事件，用于完整的监控审计
func (m *HealthMonitor) publishPassiveModeHeartbeatEvent(ctx context.Context, instance *core.ServiceInstance, healthStatus string, checkResult string, heartbeatError error) {
	// 生成事件ID
	eventId := random.Generate32BitRandomString()

	// 构建事件消息
	var eventMessage string
	if heartbeatError != nil {
		eventMessage = fmt.Sprintf("被动模式心跳检查完成，状态: %s, 结果: %s, 失败次数: %d, 错误: %s",
			healthStatus, checkResult, instance.HeartbeatFailCount, heartbeatError.Error())
	} else {
		eventMessage = fmt.Sprintf("被动模式心跳检查完成，状态: %s, 结果: %s, 失败次数: %d",
			healthStatus, checkResult, instance.HeartbeatFailCount)
	}

	// 创建心跳检查事件
	event := &core.ServiceEvent{
		ServiceEventId:    eventId,
		EventType:         core.EventTypeInstanceHeartbeatUpdated,
		TenantId:          instance.TenantId,
		ServiceGroupId:    instance.ServiceGroupId,
		ServiceInstanceId: instance.ServiceInstanceId,
		GroupName:         instance.GroupName,
		ServiceName:       instance.ServiceName,
		HostAddress:       instance.HostAddress,
		PortNumber:        instance.PortNumber,
		NodeIpAddress:     random.GetNodeIP(),
		EventTime:         time.Now(),
		EventSource:       core.GetEventSourceFromContext(ctx, core.EventSourceHealthMonitor),
		EventMessage:      eventMessage,
		ActiveFlag:        "Y",
		Instance:          instance, // 添加实例对象用于事件处理时的数据传递
	}

	// 发布事件
	err := m.eventPublisher.Publish(ctx, event)
	if err != nil {
		logger.WarnWithTrace(ctx, "发布被动模式心跳检查事件失败",
			"instanceId", instance.ServiceInstanceId,
			"serviceName", instance.ServiceName,
			"healthStatus", healthStatus,
			"checkResult", checkResult,
			"error", err)
	} else {
		logger.DebugWithTrace(ctx, "发布被动模式心跳检查事件成功",
			"instanceId", instance.ServiceInstanceId,
			"serviceName", instance.ServiceName,
			"healthStatus", healthStatus,
			"checkResult", checkResult,
			"eventId", eventId)
	}
}

// publishHealthCheckEvent 发布健康检查执行事件
// 无论健康状态是否发生变更，每次健康检查都会发布此事件，用于完整的监控审计
func (m *HealthMonitor) publishHealthCheckEvent(ctx context.Context, instance *core.ServiceInstance, healthStatus string, httpResponse string, healthCheckError error, duration time.Duration, healthCheckUrl string) {
	// 生成事件ID
	eventId := random.Generate32BitRandomString()

	// 构建事件消息
	var eventMessage string
	if httpResponse != "" {
		eventMessage = fmt.Sprintf("健康检查执行完成，检查地址: %s, 状态: %s, 响应时间: %v, HTTP响应: %s",
			healthCheckUrl, healthStatus, duration, httpResponse)
		if healthCheckError != nil {
			eventMessage += fmt.Sprintf(", 错误: %s", healthCheckError.Error())
		}
	} else {
		eventMessage = fmt.Sprintf("健康检查执行完成，检查地址: %s, 状态: %s, 响应时间: %v",
			healthCheckUrl, healthStatus, duration)
		if healthCheckError != nil {
			eventMessage += fmt.Sprintf(", 错误: %s", healthCheckError.Error())
		}
	}

	// 创建健康检查执行事件
	event := &core.ServiceEvent{
		ServiceEventId:    eventId,
		EventType:         core.EventTypeInstanceHeartbeatUpdated,
		TenantId:          instance.TenantId,
		ServiceGroupId:    instance.ServiceGroupId,
		ServiceInstanceId: instance.ServiceInstanceId,
		GroupName:         instance.GroupName,
		ServiceName:       instance.ServiceName,
		HostAddress:       instance.HostAddress,
		PortNumber:        instance.PortNumber,
		NodeIpAddress:     random.GetNodeIP(),
		EventTime:         time.Now(),
		EventSource:       core.GetEventSourceFromContext(ctx, core.EventSourceHealthMonitor),
		EventMessage:      eventMessage,
		ActiveFlag:        "Y",
		Instance:          instance, // 添加实例对象用于事件处理时的数据传递
	}

	// 发布事件
	err := m.eventPublisher.Publish(ctx, event)
	if err != nil {
		logger.WarnWithTrace(ctx, "发布健康检查执行事件失败",
			"instanceId", instance.ServiceInstanceId,
			"serviceName", instance.ServiceName,
			"healthStatus", healthStatus,
			"httpResponse", httpResponse,
			"duration", duration.String(),
			"error", err)
	} else {
		logger.DebugWithTrace(ctx, "发布健康检查执行事件成功",
			"instanceId", instance.ServiceInstanceId,
			"serviceName", instance.ServiceName,
			"healthStatus", healthStatus,
			"httpResponse", httpResponse,
			"duration", duration.String(),
			"eventId", eventId)
	}
}
