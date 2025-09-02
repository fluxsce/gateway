package dao

import (
	"context"
	"fmt"
	"time"

	"gateway/internal/registry/core"
	"gateway/internal/registry/manager"
)

// RegistryDAO 注册服务数据访问对象
// 提供与注册中心交互的数据访问接口
type RegistryDAO struct {
	registryManager *manager.RegistryManager
}

// NewRegistryDAO 创建注册服务DAO实例
func NewRegistryDAO() (*RegistryDAO, error) {
	// 获取注册中心管理器实例
	registryManager := manager.GetInstance()
	if registryManager == nil {
		return nil, fmt.Errorf("注册中心管理器未初始化")
	}

	return &RegistryDAO{
		registryManager: registryManager,
	}, nil
}

// =============================================================================
// 服务管理
// =============================================================================

// RegisterService 注册服务
func (dao *RegistryDAO) RegisterService(ctx context.Context, service *core.Service) error {
	// 设置默认值
	now := time.Now()
	service.AddTime = now
	service.AddWho = "SDK_CLIENT"
	service.EditTime = now
	service.EditWho = "SDK_CLIENT"
	service.ActiveFlag = "Y"
	service.CurrentVersion = 1

	// 注册服务
	_, err := dao.registryManager.RegisterService(ctx, service)
	if err != nil {
		return fmt.Errorf("注册服务失败: %w", err)
	}

	return nil
}

// DeregisterService 注销服务
func (dao *RegistryDAO) DeregisterService(ctx context.Context, tenantId, serviceGroupId, serviceName string) error {
	// 注销服务
	err := dao.registryManager.DeregisterService(ctx, tenantId, serviceGroupId, serviceName)
	if err != nil {
		return fmt.Errorf("注销服务失败: %w", err)
	}

	return nil
}

// DiscoverService 发现服务
func (dao *RegistryDAO) DiscoverService(ctx context.Context, tenantId, serviceGroupId, groupName, serviceName string) (*core.Service, error) {
	// 发现服务
	service, err := dao.registryManager.GetService(ctx, tenantId, serviceGroupId, serviceName)
	if err != nil {
		return nil, fmt.Errorf("发现服务失败: %w", err)
	}

	if service == nil {
		return nil, fmt.Errorf("服务不存在: %s", serviceName)
	}

	return service, nil
}

// =============================================================================
// 服务实例管理
// =============================================================================

// RegisterInstance 注册服务实例
func (dao *RegistryDAO) RegisterInstance(ctx context.Context, instance *core.ServiceInstance) (*core.ServiceInstance, error) {
	// 设置默认值
	now := time.Now()
	instance.RegisterTime = now
	instance.AddTime = now
	instance.AddWho = "SDK_CLIENT"
	instance.EditTime = now
	instance.EditWho = "SDK_CLIENT"
	instance.ActiveFlag = "Y"
	instance.CurrentVersion = 1

	// 注册服务实例
	registeredInstance, err := dao.registryManager.RegisterInstance(ctx, instance)
	if err != nil {
		return nil, fmt.Errorf("注册服务实例失败: %w", err)
	}

	return registeredInstance, nil
}

// UpdateInstance 更新服务实例
func (dao *RegistryDAO) UpdateInstance(ctx context.Context, instance *core.ServiceInstance) error {
	// 设置更新时间
	now := time.Now()
	instance.EditTime = now
	instance.EditWho = "SDK_CLIENT"

	// 更新服务实例
	_, err := dao.registryManager.UpdateInstance(ctx, instance)
	if err != nil {
		return fmt.Errorf("更新服务实例失败: %w", err)
	}

	return nil
}

// DeregisterInstance 注销服务实例
func (dao *RegistryDAO) DeregisterInstance(ctx context.Context, tenantId, serviceInstanceId string) error {
	// 先获取实例信息
	instance, err := dao.registryManager.GetInstance(ctx, tenantId, serviceInstanceId)
	if err != nil {
		return fmt.Errorf("获取服务实例失败: %w", err)
	}

	if instance == nil {
		return fmt.Errorf("服务实例不存在: %s", serviceInstanceId)
	}

	// 注销服务实例
	err = dao.registryManager.DeregisterInstance(ctx, tenantId, serviceInstanceId)
	if err != nil {
		return fmt.Errorf("注销服务实例失败: %w", err)
	}

	return nil
}

// DiscoverInstance 发现服务实例
func (dao *RegistryDAO) DiscoverInstance(ctx context.Context, tenantId, serviceGroupId, serviceName string) (*core.ServiceInstance, error) {
	// 使用负载均衡策略发现服务实例
	instance, err := dao.registryManager.DiscoverInstance(ctx, tenantId, serviceGroupId, serviceName)
	if err != nil {
		return nil, fmt.Errorf("发现服务实例失败: %w", err)
	}

	if instance == nil {
		return nil, fmt.Errorf("未找到可用的服务实例: %s", serviceName)
	}

	return instance, nil
}

// ListInstances 获取服务的所有实例列表
func (dao *RegistryDAO) ListInstances(ctx context.Context, tenantId, serviceGroupId, serviceName string) ([]*core.ServiceInstance, error) {
	// 获取服务的所有实例
	instances, err := dao.registryManager.ListInstances(ctx, tenantId, serviceGroupId, serviceName)
	if err != nil {
		return nil, fmt.Errorf("获取服务实例列表失败: %w", err)
	}

	return instances, nil
}

// SendHeartbeat 发送心跳
func (dao *RegistryDAO) SendHeartbeat(ctx context.Context, tenantId, serviceInstanceId string) error {
	// 更新实例的最后心跳时间
	err := dao.registryManager.UpdateInstanceHeartbeat(ctx, tenantId, serviceInstanceId)
	if err != nil {
		return fmt.Errorf("更新心跳时间失败: %w", err)
	}

	return nil
}

// UpdateInstanceStatus 更新实例状态
func (dao *RegistryDAO) UpdateInstanceStatus(ctx context.Context, tenantId, serviceInstanceId, instanceStatus, healthStatus string, weightValue int) error {
	// 先获取实例信息
	instance, err := dao.registryManager.GetInstance(ctx, tenantId, serviceInstanceId)
	if err != nil {
		return fmt.Errorf("获取服务实例失败: %w", err)
	}

	if instance == nil {
		return fmt.Errorf("服务实例不存在: %s", serviceInstanceId)
	}

	// 更新实例状态
	instance.InstanceStatus = instanceStatus
	if healthStatus != "" {
		instance.HealthStatus = healthStatus
	}
	if weightValue > 0 {
		instance.WeightValue = weightValue
	}
	instance.EditTime = time.Now()
	instance.EditWho = "SDK_CLIENT"

	// 保存更新后的实例
	_, err = dao.registryManager.UpdateInstance(ctx, instance)
	if err != nil {
		return fmt.Errorf("更新实例状态失败: %w", err)
	}

	return nil
}
