package cache

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"gateway/internal/registry/core"
	"gateway/pkg/logger"
	"gateway/pkg/plugin/nacos"
)

// NacosRegistryClient Nacos注册中心客户端实现
type NacosRegistryClient struct {
	tool        *nacos.NacosTool
	config      *nacos.NacosConfig
	configHash  string // 配置哈希值，用于检测配置变化
	initialized bool
	mutex       sync.RWMutex
}

// NewNacosRegistryClient 创建Nacos注册中心客户端（无参构造，延迟初始化）
func NewNacosRegistryClient() *NacosRegistryClient {
	return &NacosRegistryClient{
		initialized: false,
	}
}

// initializeFromService 从服务配置初始化客户端，支持配置变化检测和资源释放
func (c *NacosRegistryClient) initializeFromService(service *core.Service) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 检查服务的外部注册中心配置
	if service.ExternalRegistryConfig == "" {
		return fmt.Errorf("服务 %s 的外部注册中心配置为空，无法初始化Nacos客户端", service.ServiceName)
	}

	// 计算新配置的哈希值
	newConfigHash := calculateConfigHash(service.ExternalRegistryConfig)

	// 如果已经初始化且配置没有变化，直接返回
	if c.initialized && c.configHash == newConfigHash {
		logger.DebugWithTrace(context.Background(), "Nacos客户端配置未变化，复用现有连接",
			"serviceName", service.ServiceName,
			"configHash", newConfigHash)
		return nil
	}

	// 如果已经初始化但配置发生变化，需要释放旧资源
	if c.initialized {
		logger.InfoWithTrace(context.Background(), "检测到Nacos配置变化，释放旧资源并重新初始化",
			"serviceName", service.ServiceName,
			"oldConfigHash", c.configHash,
			"newConfigHash", newConfigHash)

		// 释放旧的工具资源
		if c.tool != nil {
			if c.tool.IsConnected() {
				if err := c.tool.Disconnect(); err != nil {
					logger.WarnWithTrace(context.Background(), "断开旧Nacos连接失败",
						"serviceName", service.ServiceName,
						"error", err)
				}
			}
			// 清空旧的工具引用，让GC回收
			c.tool = nil
		}
		c.config = nil
		c.initialized = false
		c.configHash = ""
	}

	// 解析服务的外部注册中心配置
	nacosConfig, err := parseNacosConfigFromService(service.ExternalRegistryConfig)
	if err != nil {
		return fmt.Errorf("解析服务 %s 的Nacos配置失败: %w", service.ServiceName, err)
	}

	// 初始化工具和配置
	c.tool = nacos.NewNacosTool(nacosConfig)
	c.config = nacosConfig
	c.configHash = newConfigHash
	c.initialized = true

	logger.Info("Nacos客户端初始化完成",
		"serviceName", service.ServiceName,
		"servers", len(nacosConfig.Servers),
		"namespace", nacosConfig.Namespace,
		"group", nacosConfig.Group,
		"configHash", newConfigHash)

	return nil
}

// ensureInitialized 确保客户端已初始化（需要服务配置）
func (c *NacosRegistryClient) ensureInitialized() error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if !c.initialized {
		return fmt.Errorf("Nacos客户端未初始化，请先调用RegisterService方法")
	}
	return nil
}

// parseNacosConfigFromService 从服务配置JSON中解析Nacos配置
func parseNacosConfigFromService(configJson string) (*nacos.NacosConfig, error) {
	if configJson == "" {
		return nil, fmt.Errorf("外部注册中心配置为空")
	}

	// 直接解析为NacosConfig
	var config nacos.NacosConfig
	if err := json.Unmarshal([]byte(configJson), &config); err != nil {
		return nil, fmt.Errorf("解析Nacos配置JSON失败: %w", err)
	}

	// 使用nacos包的验证方法
	if err := nacos.Validate(&config); err != nil {
		return nil, fmt.Errorf("Nacos配置验证失败: %w", err)
	}

	return &config, nil
}

// calculateConfigHash 计算配置的哈希值，用于检测配置变化
func calculateConfigHash(configJson string) string {
	hash := md5.Sum([]byte(configJson))
	return fmt.Sprintf("%x", hash)
}

// getEffectiveGroupName 获取有效的分组名称
// 优先使用 NacosConfig 中配置的分组，如果没有配置则使用传入的分组名
func (c *NacosRegistryClient) getEffectiveGroupName(requestGroupName string) string {
	// 注意：这里不需要加锁，因为调用方法已经持有了锁
	if c.config != nil && c.config.Group != "" {
		return c.config.Group
	}
	if requestGroupName != "" {
		return requestGroupName
	}
	return "DEFAULT_GROUP"
}

// Connect 连接到Nacos
func (c *NacosRegistryClient) Connect(ctx context.Context) error {
	if err := c.ensureInitialized(); err != nil {
		return err
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.tool.Connect(ctx)
}

// Disconnect 断开Nacos连接
func (c *NacosRegistryClient) Disconnect() error {
	if err := c.ensureInitialized(); err != nil {
		return err
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.tool.Disconnect()
}

// Close 完全关闭客户端并释放所有资源
func (c *NacosRegistryClient) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	var lastErr error

	// 如果已连接，先断开连接
	if c.initialized && c.tool != nil && c.tool.IsConnected() {
		if err := c.tool.Disconnect(); err != nil {
			logger.Warn("关闭Nacos客户端时断开连接失败", "error", err)
			lastErr = err
		}
	}

	// 清理所有资源
	c.tool = nil
	c.config = nil
	c.configHash = ""
	c.initialized = false

	logger.Debug("Nacos客户端资源已完全释放")
	return lastErr
}

// IsConnected 检查连接状态
func (c *NacosRegistryClient) IsConnected() bool {
	if err := c.ensureInitialized(); err != nil {
		return false
	}

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.tool.IsConnected()
}

// RegisterService 注册服务到Nacos
func (c *NacosRegistryClient) RegisterService(ctx context.Context, service *core.Service) error {
	// 确保客户端已初始化
	if err := c.initializeFromService(service); err != nil {
		return fmt.Errorf("初始化Nacos客户端失败: %w", err)
	}

	// 自动连接到Nacos
	if !c.IsConnected() {
		if err := c.Connect(ctx); err != nil {
			return fmt.Errorf("连接Nacos服务器失败: %w", err)
		}
		logger.InfoWithTrace(ctx, "Nacos客户端连接成功",
			"serviceName", service.ServiceName,
			"servers", len(c.config.Servers))
	}

	// Nacos中服务是通过实例注册自动创建的，这里暂时不需要特别处理
	logger.DebugWithTrace(ctx, "Nacos服务注册（通过实例自动创建）",
		"serviceName", service.ServiceName,
		"groupName", service.GroupName)
	return nil
}

// DeregisterService 从Nacos注销服务
func (c *NacosRegistryClient) DeregisterService(ctx context.Context, service *core.Service) error {
	if err := c.ensureInitialized(); err != nil {
		// 如果未初始化，视为注销成功
		logger.DebugWithTrace(ctx, "Nacos客户端未初始化，跳过服务注销",
			"serviceName", service.ServiceName,
			"groupName", service.GroupName)
		return nil
	}

	// 如果未连接，也视为注销成功
	if !c.IsConnected() {
		logger.DebugWithTrace(ctx, "Nacos客户端未连接，跳过服务注销",
			"serviceName", service.ServiceName,
			"groupName", service.GroupName)
		return nil
	}

	// Nacos中服务注销是通过注销所有实例来实现的
	instanceCount := 0
	if service.Instances != nil {
		for _, instance := range service.Instances {
			if err := c.DeregisterInstance(ctx, instance); err != nil {
				logger.WarnWithTrace(ctx, "从Nacos注销实例失败",
					"instanceId", instance.ServiceInstanceId,
					"serviceName", service.ServiceName,
					"groupName", service.GroupName,
					"error", err)
			} else {
				instanceCount++
			}
		}
	}

	logger.InfoWithTrace(ctx, "Nacos服务注销完成",
		"serviceName", service.ServiceName,
		"groupName", service.GroupName,
		"deregisteredInstances", instanceCount,
		"totalInstances", len(service.Instances))

	return nil
}

// RegisterInstance 注册实例到Nacos
func (c *NacosRegistryClient) RegisterInstance(ctx context.Context, instance *core.ServiceInstance) error {
	if err := c.ensureInitialized(); err != nil {
		return err
	}

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if !c.tool.IsConnected() {
		return fmt.Errorf("Nacos客户端未连接")
	}

	// 构建元数据
	metadata := make(map[string]string)
	if instance.MetadataJson != "" {
		var metadataMap map[string]interface{}
		if err := json.Unmarshal([]byte(instance.MetadataJson), &metadataMap); err == nil {
			for k, v := range metadataMap {
				if str, ok := v.(string); ok {
					metadata[k] = str
				}
			}
		}
	}

	// 添加系统元数据
	metadata["instanceId"] = instance.ServiceInstanceId
	metadata["clientType"] = instance.ClientType
	metadata["instanceStatus"] = instance.InstanceStatus
	metadata["healthStatus"] = instance.HealthStatus
	metadata["tempInstanceFlag"] = instance.TempInstanceFlag

	// 注册实例到Nacos，优先使用配置文件中的分组
	groupName := c.getEffectiveGroupName(instance.GroupName)
	err := c.tool.RegisterServiceWithMetadata(
		instance.ServiceName,
		instance.HostAddress,
		uint64(instance.PortNumber),
		groupName,
		metadata,
	)

	if err != nil {
		return fmt.Errorf("注册实例到Nacos失败: %w", err)
	}

	logger.InfoWithTrace(ctx, "实例注册到Nacos成功",
		"instanceId", instance.ServiceInstanceId,
		"serviceName", instance.ServiceName,
		"groupName", groupName,
		"address", fmt.Sprintf("%s:%d", instance.HostAddress, instance.PortNumber))

	return nil
}

// DeregisterInstance 从Nacos注销实例
func (c *NacosRegistryClient) DeregisterInstance(ctx context.Context, instance *core.ServiceInstance) error {
	if err := c.ensureInitialized(); err != nil {
		return err
	}

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if !c.tool.IsConnected() {
		return fmt.Errorf("Nacos客户端未连接")
	}

	// 从Nacos注销实例，优先使用配置文件中的分组
	groupName := c.getEffectiveGroupName(instance.GroupName)
	err := c.tool.DeregisterServiceWithGroup(
		instance.ServiceName,
		instance.HostAddress,
		uint64(instance.PortNumber),
		groupName,
	)

	if err != nil {
		return fmt.Errorf("从Nacos注销实例失败: %w", err)
	}

	logger.InfoWithTrace(ctx, "实例从Nacos注销成功",
		"instanceId", instance.ServiceInstanceId,
		"serviceName", instance.ServiceName,
		"groupName", groupName,
		"address", fmt.Sprintf("%s:%d", instance.HostAddress, instance.PortNumber))

	return nil
}

// UpdateInstance 更新Nacos中的实例
func (c *NacosRegistryClient) UpdateInstance(ctx context.Context, instance *core.ServiceInstance) error {
	// Nacos中更新实例通常是重新注册
	return c.RegisterInstance(ctx, instance)
}

// DiscoverServices 从Nacos发现服务
func (c *NacosRegistryClient) DiscoverServices(ctx context.Context, groupName string) ([]*core.Service, error) {
	if err := c.ensureInitialized(); err != nil {
		return nil, err
	}

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if !c.tool.IsConnected() {
		return nil, fmt.Errorf("Nacos客户端未连接")
	}

	// 优先使用配置文件中的分组
	effectiveGroupName := c.getEffectiveGroupName(groupName)

	// 获取所有服务信息
	serviceList, err := c.tool.GetAllServicesWithGroup(1, 1000, effectiveGroupName)
	if err != nil {
		return nil, fmt.Errorf("从Nacos获取服务列表失败: %w", err)
	}

	var services []*core.Service
	for _, serviceName := range serviceList.Doms {
		// 获取服务实例
		instances, err := c.tool.DiscoverServiceWithGroup(serviceName, effectiveGroupName)
		if err != nil {
			logger.WarnWithTrace(ctx, "获取服务实例失败", "serviceName", serviceName, "error", err)
			continue
		}

		// 转换为core.Service
		service := &core.Service{
			ServiceName:  serviceName,
			GroupName:    effectiveGroupName,
			RegistryType: core.RegistryTypeNacos,
			ProtocolType: core.ProtocolTypeHTTP, // 默认值
			ActiveFlag:   "Y",
			AddTime:      time.Now(),
			EditTime:     time.Now(),
		}

		// 转换实例
		for _, nacosInstance := range instances {
			coreInstance := &core.ServiceInstance{
				ServiceInstanceId: nacosInstance.InstanceId,
				ServiceName:       serviceName,
				GroupName:         effectiveGroupName,
				HostAddress:       nacosInstance.Ip,
				PortNumber:        int(nacosInstance.Port),
				WeightValue:       int(nacosInstance.Weight),
				HealthStatus:      core.HealthStatusHealthy,
				InstanceStatus:    core.InstanceStatusUp,
				ClientType:        core.ClientTypeService,
				TempInstanceFlag:  core.TempInstanceFlagYes,
				ActiveFlag:        "Y",
				AddTime:           time.Now(),
				EditTime:          time.Now(),
			}

			// 设置健康状态
			if !nacosInstance.Healthy {
				coreInstance.HealthStatus = core.HealthStatusUnhealthy
			}

			// 处理元数据
			if len(nacosInstance.Metadata) > 0 {
				metadataBytes, _ := json.Marshal(nacosInstance.Metadata)
				coreInstance.MetadataJson = string(metadataBytes)
			}

			service.Instances = append(service.Instances, coreInstance)
		}

		services = append(services, service)
	}

	logger.DebugWithTrace(ctx, "从Nacos发现服务",
		"groupName", effectiveGroupName,
		"serviceCount", len(services))

	return services, nil
}

// DiscoverInstances 从Nacos发现服务实例
func (c *NacosRegistryClient) DiscoverInstances(ctx context.Context, serviceName, groupName string) ([]*core.ServiceInstance, error) {
	if err := c.ensureInitialized(); err != nil {
		return nil, err
	}

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if !c.tool.IsConnected() {
		return nil, fmt.Errorf("Nacos客户端未连接")
	}

	// 优先使用配置文件中的分组
	effectiveGroupName := c.getEffectiveGroupName(groupName)

	// 获取服务实例
	nacosInstances, err := c.tool.DiscoverServiceWithGroup(serviceName, effectiveGroupName)
	if err != nil {
		return nil, fmt.Errorf("从Nacos获取服务实例失败: %w", err)
	}

	var instances []*core.ServiceInstance
	for _, nacosInstance := range nacosInstances {
		coreInstance := &core.ServiceInstance{
			ServiceInstanceId: nacosInstance.InstanceId,
			ServiceName:       serviceName,
			GroupName:         effectiveGroupName,
			HostAddress:       nacosInstance.Ip,
			PortNumber:        int(nacosInstance.Port),
			WeightValue:       int(nacosInstance.Weight),
			HealthStatus:      core.HealthStatusHealthy,
			InstanceStatus:    core.InstanceStatusUp,
			ClientType:        core.ClientTypeService,
			TempInstanceFlag:  core.TempInstanceFlagYes,
			ActiveFlag:        "Y",
			AddTime:           time.Now(),
			EditTime:          time.Now(),
		}

		// 设置健康状态
		if !nacosInstance.Healthy {
			coreInstance.HealthStatus = core.HealthStatusUnhealthy
		}

		// 处理元数据
		if len(nacosInstance.Metadata) > 0 {
			metadataBytes, _ := json.Marshal(nacosInstance.Metadata)
			coreInstance.MetadataJson = string(metadataBytes)
		}

		instances = append(instances, coreInstance)
	}

	logger.DebugWithTrace(ctx, "从Nacos发现服务实例",
		"serviceName", serviceName,
		"groupName", effectiveGroupName,
		"instanceCount", len(instances))

	return instances, nil
}

// UpdateInstanceHealth 更新实例健康状态
func (c *NacosRegistryClient) UpdateInstanceHealth(ctx context.Context, instanceId string, healthy bool) error {
	// Nacos中实例健康状态由心跳机制维护，这里记录日志即可
	status := "HEALTHY"
	if !healthy {
		status = "UNHEALTHY"
	}

	logger.DebugWithTrace(ctx, "Nacos实例健康状态更新",
		"instanceId", instanceId,
		"healthy", healthy,
		"status", status)

	return nil
}
