package dao

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gateway/internal/registry/core"    // 用于核心数据类型转换
	"gateway/internal/registry/manager" // 用于调用注册中心管理器
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/pkg/utils/huberrors"
	"gateway/pkg/utils/random"
	"gateway/web/views/hub0041/models"
)

// ServiceInstanceDAO 服务实例数据访问对象
// 负责服务实例信息的数据库操作（查询、创建、编辑、删除）
type ServiceInstanceDAO struct {
	db database.Database
}

// NewServiceInstanceDAO 创建服务实例DAO实例
func NewServiceInstanceDAO(db database.Database) *ServiceInstanceDAO {
	return &ServiceInstanceDAO{
		db: db,
	}
}

// QueryServiceInstances 分页查询服务实例列表
//
// 优先从注册中心管理器缓存获取，包括外部注册中心的实例
//
// 参数：
//   - ctx: 上下文对象
//   - tenantId: 租户ID
//   - activeFlag: 活动状态标记(Y活动,N非活动,空为全部)
//   - serviceName: 服务名称过滤
//   - groupName: 分组名称过滤
//   - instanceStatus: 实例状态过滤
//   - healthStatus: 健康状态过滤
//   - hostAddress: 主机地址过滤（模糊查询）
//   - page: 页码
//   - pageSize: 每页数量
//
// 返回值：
//   - []*models.ServiceInstance: 服务实例列表
//   - int: 总数量
//   - error: 错误信息
func (dao *ServiceInstanceDAO) QueryServiceInstances(ctx context.Context, tenantId, activeFlag, serviceName, groupName, instanceStatus, healthStatus, hostAddress string, page, pageSize int) ([]*models.ServiceInstance, int, error) {
	// 设置默认值
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100 // 限制最大分页大小
	}

	// 获取注册中心管理器实例
	registryManager := manager.GetInstance()

	var allInstances []*models.ServiceInstance

	// 如果指定了服务名称，优先从缓存获取
	if serviceName != "" {
		// 直接从数据库获取服务基本信息，避免与service_dao.go形成死循环
		var service struct {
			ServiceGroupId string `db:"serviceGroupId"`
			RegistryType   string `db:"registryType"`
		}

		serviceQuery := `SELECT serviceGroupId, registryType FROM HUB_REGISTRY_SERVICE WHERE tenantId = ? AND serviceName = ?`
		serviceArgs := []interface{}{tenantId, serviceName}

		if activeFlag != "" {
			serviceQuery += " AND activeFlag = ?"
			serviceArgs = append(serviceArgs, activeFlag)
		}

		err := dao.db.QueryOne(ctx, &service, serviceQuery, serviceArgs, true)
		if err != nil {
			if strings.Contains(err.Error(), "no rows") || strings.Contains(err.Error(), "not found") {
				return []*models.ServiceInstance{}, 0, nil
			}
			return nil, 0, huberrors.WrapError(err, "获取服务信息失败")
		}

		// 从注册中心管理器获取实例列表（包括外部注册中心的实例）
		coreInstances, err := registryManager.ListInstances(ctx, tenantId, service.ServiceGroupId, serviceName)
		if err != nil {
			// 判断服务的注册类型，决定错误处理策略
			if service.RegistryType != "" && service.RegistryType != "INTERNAL" {
				// 外部注册中心的服务，抛出错误，不回退到数据库
				logger.ErrorWithTrace(ctx, "从外部注册中心获取实例失败",
					"serviceName", serviceName,
					"registryType", service.RegistryType,
					"error", err)
				return nil, 0, huberrors.WrapError(err, "从外部注册中心获取服务实例失败: %s", service.RegistryType)
			} else {
				// 内部注册中心的服务，回退到数据库查询
				logger.WarnWithTrace(ctx, "从注册中心管理器获取实例失败，回退到数据库查询",
					"serviceName", serviceName,
					"registryType", service.RegistryType,
					"error", err)
				return dao.queryInstancesFromDatabase(ctx, tenantId, activeFlag, serviceName, groupName, instanceStatus, healthStatus, hostAddress, page, pageSize)
			}
		}

		// 转换 core.ServiceInstance 到 models.ServiceInstance
		for _, coreInstance := range coreInstances {
			modelInstance := &models.ServiceInstance{
				ServiceInstanceId:   coreInstance.ServiceInstanceId,
				TenantId:            coreInstance.TenantId,
				ServiceGroupId:      coreInstance.ServiceGroupId,
				ServiceName:         coreInstance.ServiceName,
				GroupName:           coreInstance.GroupName,
				HostAddress:         coreInstance.HostAddress,
				PortNumber:          coreInstance.PortNumber,
				ContextPath:         coreInstance.ContextPath,
				InstanceStatus:      coreInstance.InstanceStatus,
				HealthStatus:        coreInstance.HealthStatus,
				WeightValue:         coreInstance.WeightValue,
				ClientId:            coreInstance.ClientId,
				ClientVersion:       coreInstance.ClientVersion,
				ClientType:          coreInstance.ClientType,
				TempInstanceFlag:    coreInstance.TempInstanceFlag,
				HeartbeatFailCount:  coreInstance.HeartbeatFailCount,
				MetadataJson:        coreInstance.MetadataJson,
				TagsJson:            coreInstance.TagsJson,
				RegisterTime:        coreInstance.RegisterTime,
				LastHeartbeatTime:   coreInstance.LastHeartbeatTime,
				LastHealthCheckTime: coreInstance.LastHealthCheckTime,
				AddTime:             coreInstance.AddTime,
				AddWho:              coreInstance.AddWho,
				EditTime:            coreInstance.EditTime,
				EditWho:             coreInstance.EditWho,
				OprSeqFlag:          coreInstance.OprSeqFlag,
				CurrentVersion:      coreInstance.CurrentVersion,
				ActiveFlag:          coreInstance.ActiveFlag,
				NoteText:            coreInstance.NoteText,
				ExtProperty:         coreInstance.ExtProperty,
				Reserved1:           coreInstance.Reserved1,
				Reserved2:           coreInstance.Reserved2,
				Reserved3:           coreInstance.Reserved3,
				Reserved4:           coreInstance.Reserved4,
				Reserved5:           coreInstance.Reserved5,
				Reserved6:           coreInstance.Reserved6,
				Reserved7:           coreInstance.Reserved7,
				Reserved8:           coreInstance.Reserved8,
				Reserved9:           coreInstance.Reserved9,
				Reserved10:          coreInstance.Reserved10,
			}
			allInstances = append(allInstances, modelInstance)
		}
	} else {
		// 如果没有指定服务名称，回退到数据库查询
		return dao.queryInstancesFromDatabase(ctx, tenantId, activeFlag, serviceName, groupName, instanceStatus, healthStatus, hostAddress, page, pageSize)
	}

	// 应用过滤条件
	var filteredInstances []*models.ServiceInstance
	for _, instance := range allInstances {
		// 活动状态过滤
		if activeFlag != "" && instance.ActiveFlag != activeFlag {
			continue
		}

		// 分组名称过滤
		if groupName != "" && instance.GroupName != groupName {
			continue
		}

		// 实例状态过滤
		if instanceStatus != "" && instance.InstanceStatus != instanceStatus {
			continue
		}

		// 健康状态过滤
		if healthStatus != "" && instance.HealthStatus != healthStatus {
			continue
		}

		// 主机地址过滤（模糊查询）
		if hostAddress != "" && !strings.Contains(instance.HostAddress, hostAddress) {
			continue
		}

		filteredInstances = append(filteredInstances, instance)
	}

	// 手动分页
	total := len(filteredInstances)
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= total {
		// 超出范围，返回空结果
		return []*models.ServiceInstance{}, total, nil
	}

	if end > total {
		end = total
	}

	instances := filteredInstances[start:end]
	return instances, total, nil
}

// queryInstancesFromDatabase 从数据库查询服务实例列表（回退方法）
func (dao *ServiceInstanceDAO) queryInstancesFromDatabase(ctx context.Context, tenantId, activeFlag, serviceName, groupName, instanceStatus, healthStatus, hostAddress string, page, pageSize int) ([]*models.ServiceInstance, int, error) {
	// 构建基础查询条件
	whereConditions := []string{"tenantId = ?"}
	args := []interface{}{tenantId}

	// 添加过滤条件
	if activeFlag != "" {
		whereConditions = append(whereConditions, "activeFlag = ?")
		args = append(args, activeFlag)
	}

	if serviceName != "" {
		whereConditions = append(whereConditions, "serviceName = ?")
		args = append(args, serviceName)
	}

	if groupName != "" {
		whereConditions = append(whereConditions, "groupName = ?")
		args = append(args, groupName)
	}

	if instanceStatus != "" {
		whereConditions = append(whereConditions, "instanceStatus = ?")
		args = append(args, instanceStatus)
	}

	if healthStatus != "" {
		whereConditions = append(whereConditions, "healthStatus = ?")
		args = append(args, healthStatus)
	}

	if hostAddress != "" {
		whereConditions = append(whereConditions, "hostAddress LIKE ?")
		args = append(args, "%"+hostAddress+"%")
	}

	whereClause := strings.Join(whereConditions, " AND ")

	// 构建查询SQL
	query := `SELECT serviceInstanceId, tenantId, serviceGroupId, serviceName, groupName,
		hostAddress, portNumber, contextPath,
		instanceStatus, healthStatus, weightValue,
		clientId, clientVersion, clientType, tempInstanceFlag, heartbeatFailCount,
		metadataJson, tagsJson,
		registerTime, lastHeartbeatTime, lastHealthCheckTime,
		addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty,
		reserved1, reserved2, reserved3, reserved4, reserved5, reserved6, reserved7, reserved8, reserved9, reserved10
	FROM HUB_REGISTRY_SERVICE_INSTANCE WHERE ` + whereClause + ` ORDER BY registerTime DESC`

	// 查询所有符合条件的记录
	var allInstances []*models.ServiceInstance
	err := dao.db.Query(ctx, &allInstances, query, args, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询服务实例列表失败")
	}

	// 手动分页
	total := len(allInstances)
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= total {
		// 超出范围，返回空结果
		return []*models.ServiceInstance{}, total, nil
	}

	if end > total {
		end = total
	}

	instances := allInstances[start:end]
	return instances, total, nil
}

// GetServiceInstance 根据实例ID获取服务实例详情
//
// 优先从注册中心管理器缓存获取，包括外部注册中心的实例
//
// 参数：
//   - ctx: 上下文对象
//   - tenantId: 租户ID
//   - serviceInstanceId: 服务实例ID
//   - activeFlag: 活动状态标记(为空则不过滤)
//
// 返回值：
//   - *models.ServiceInstance: 服务实例信息
//   - error: 错误信息
func (dao *ServiceInstanceDAO) GetServiceInstance(ctx context.Context, tenantId, serviceInstanceId, activeFlag string) (*models.ServiceInstance, error) {
	// 获取注册中心管理器实例
	registryManager := manager.GetInstance()

	// 先尝试从注册中心管理器获取实例（包括外部注册中心的实例）
	coreInstance, err := registryManager.GetInstance(ctx, tenantId, serviceInstanceId)
	if err != nil {
		// 如果从缓存获取失败，回退到数据库查询
		logger.DebugWithTrace(ctx, "从注册中心管理器获取实例失败，回退到数据库查询",
			"instanceId", serviceInstanceId,
			"error", err)
		return dao.getInstanceFromDatabase(ctx, tenantId, serviceInstanceId, activeFlag)
	}

	// 应用活动状态过滤
	if activeFlag != "" && coreInstance.ActiveFlag != activeFlag {
		return nil, huberrors.WrapError(fmt.Errorf("实例状态不匹配"), "服务实例不存在")
	}

	// 转换 core.ServiceInstance 到 models.ServiceInstance
	modelInstance := &models.ServiceInstance{
		ServiceInstanceId:   coreInstance.ServiceInstanceId,
		TenantId:            coreInstance.TenantId,
		ServiceGroupId:      coreInstance.ServiceGroupId,
		ServiceName:         coreInstance.ServiceName,
		GroupName:           coreInstance.GroupName,
		HostAddress:         coreInstance.HostAddress,
		PortNumber:          coreInstance.PortNumber,
		ContextPath:         coreInstance.ContextPath,
		InstanceStatus:      coreInstance.InstanceStatus,
		HealthStatus:        coreInstance.HealthStatus,
		WeightValue:         coreInstance.WeightValue,
		ClientId:            coreInstance.ClientId,
		ClientVersion:       coreInstance.ClientVersion,
		ClientType:          coreInstance.ClientType,
		TempInstanceFlag:    coreInstance.TempInstanceFlag,
		HeartbeatFailCount:  coreInstance.HeartbeatFailCount,
		MetadataJson:        coreInstance.MetadataJson,
		TagsJson:            coreInstance.TagsJson,
		RegisterTime:        coreInstance.RegisterTime,
		LastHeartbeatTime:   coreInstance.LastHeartbeatTime,
		LastHealthCheckTime: coreInstance.LastHealthCheckTime,
		AddTime:             coreInstance.AddTime,
		AddWho:              coreInstance.AddWho,
		EditTime:            coreInstance.EditTime,
		EditWho:             coreInstance.EditWho,
		OprSeqFlag:          coreInstance.OprSeqFlag,
		CurrentVersion:      coreInstance.CurrentVersion,
		ActiveFlag:          coreInstance.ActiveFlag,
		NoteText:            coreInstance.NoteText,
		ExtProperty:         coreInstance.ExtProperty,
		Reserved1:           coreInstance.Reserved1,
		Reserved2:           coreInstance.Reserved2,
		Reserved3:           coreInstance.Reserved3,
		Reserved4:           coreInstance.Reserved4,
		Reserved5:           coreInstance.Reserved5,
		Reserved6:           coreInstance.Reserved6,
		Reserved7:           coreInstance.Reserved7,
		Reserved8:           coreInstance.Reserved8,
		Reserved9:           coreInstance.Reserved9,
		Reserved10:          coreInstance.Reserved10,
	}

	return modelInstance, nil
}

// getInstanceFromDatabase 从数据库获取服务实例详情（回退方法）
func (dao *ServiceInstanceDAO) getInstanceFromDatabase(ctx context.Context, tenantId, serviceInstanceId, activeFlag string) (*models.ServiceInstance, error) {
	whereConditions := []string{"tenantId = ?", "serviceInstanceId = ?"}
	args := []interface{}{tenantId, serviceInstanceId}

	if activeFlag != "" {
		whereConditions = append(whereConditions, "activeFlag = ?")
		args = append(args, activeFlag)
	}

	whereClause := strings.Join(whereConditions, " AND ")

	query := `SELECT serviceInstanceId, tenantId, serviceGroupId, serviceName, groupName,
		hostAddress, portNumber, contextPath,
		instanceStatus, healthStatus, weightValue,
		clientId, clientVersion, clientType, tempInstanceFlag, heartbeatFailCount,
		metadataJson, tagsJson,
		registerTime, lastHeartbeatTime, lastHealthCheckTime,
		addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty,
		reserved1, reserved2, reserved3, reserved4, reserved5, reserved6, reserved7, reserved8, reserved9, reserved10
	FROM HUB_REGISTRY_SERVICE_INSTANCE WHERE ` + whereClause

	instance := &models.ServiceInstance{}
	err := dao.db.QueryOne(ctx, instance, query, args, true)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") || strings.Contains(err.Error(), "not found") {
			return nil, huberrors.WrapError(err, "服务实例不存在")
		}
		return nil, huberrors.WrapError(err, "获取服务实例信息失败")
	}

	return instance, nil
}

// UpdateServiceInstance 更新服务实例信息
//
// 直接调用注册中心管理器，确保：
// 1. 缓存更新
// 2. 事件发布（EventDBWriter会监听事件并执行数据库操作）
// 3. 完整的生命周期管理
//
// 参数：
//   - ctx: 上下文对象
//   - instance: 服务实例信息对象
//   - operatorId: 操作人员ID
//
// 返回值：
//   - *models.ServiceInstance: 更新后的服务实例信息
//   - error: 错误信息
func (dao *ServiceInstanceDAO) UpdateServiceInstance(ctx context.Context, instance *models.ServiceInstance, operatorId string) (*models.ServiceInstance, error) {
	// 获取注册中心管理器实例
	registryManager := manager.GetInstance()

	// 更新审计信息
	instance.EditTime = time.Now()
	instance.EditWho = operatorId
	instance.CurrentVersion++
	instance.OprSeqFlag = random.Generate32BitRandomString()

	// 转换 models.ServiceInstance 到 core.ServiceInstance
	coreInstance := &core.ServiceInstance{
		ServiceInstanceId:   instance.ServiceInstanceId,
		TenantId:            instance.TenantId,
		ServiceGroupId:      instance.ServiceGroupId,
		ServiceName:         instance.ServiceName,
		GroupName:           instance.GroupName,
		HostAddress:         instance.HostAddress,
		PortNumber:          instance.PortNumber,
		ContextPath:         instance.ContextPath,
		InstanceStatus:      instance.InstanceStatus,
		HealthStatus:        instance.HealthStatus,
		WeightValue:         instance.WeightValue,
		ClientId:            instance.ClientId,
		ClientVersion:       instance.ClientVersion,
		ClientType:          instance.ClientType,
		TempInstanceFlag:    instance.TempInstanceFlag,
		HeartbeatFailCount:  instance.HeartbeatFailCount,
		MetadataJson:        instance.MetadataJson,
		TagsJson:            instance.TagsJson,
		RegisterTime:        instance.RegisterTime,
		LastHeartbeatTime:   instance.LastHeartbeatTime,
		LastHealthCheckTime: instance.LastHealthCheckTime,
		AddTime:             instance.AddTime,
		AddWho:              instance.AddWho,
		EditTime:            instance.EditTime,
		EditWho:             instance.EditWho,
		OprSeqFlag:          instance.OprSeqFlag,
		CurrentVersion:      instance.CurrentVersion,
		ActiveFlag:          instance.ActiveFlag,
		NoteText:            instance.NoteText,
		ExtProperty:         instance.ExtProperty,
		Reserved1:           instance.Reserved1,
		Reserved2:           instance.Reserved2,
		Reserved3:           instance.Reserved3,
		Reserved4:           instance.Reserved4,
		Reserved5:           instance.Reserved5,
		Reserved6:           instance.Reserved6,
		Reserved7:           instance.Reserved7,
		Reserved8:           instance.Reserved8,
		Reserved9:           instance.Reserved9,
		Reserved10:          instance.Reserved10,
	}

	// 直接通过注册中心管理器更新实例（Manager会发布事件，EventDBWriter会处理数据库操作）
	updatedCoreInstance, err := registryManager.UpdateInstance(ctx, coreInstance)
	if err != nil {
		return nil, huberrors.WrapError(err, "更新服务实例失败")
	}

	// 转换回 models.ServiceInstance 并返回
	resultInstance := &models.ServiceInstance{
		ServiceInstanceId:   updatedCoreInstance.ServiceInstanceId,
		TenantId:            updatedCoreInstance.TenantId,
		ServiceGroupId:      updatedCoreInstance.ServiceGroupId,
		ServiceName:         updatedCoreInstance.ServiceName,
		GroupName:           updatedCoreInstance.GroupName,
		HostAddress:         updatedCoreInstance.HostAddress,
		PortNumber:          updatedCoreInstance.PortNumber,
		ContextPath:         updatedCoreInstance.ContextPath,
		InstanceStatus:      updatedCoreInstance.InstanceStatus,
		HealthStatus:        updatedCoreInstance.HealthStatus,
		WeightValue:         updatedCoreInstance.WeightValue,
		ClientId:            updatedCoreInstance.ClientId,
		ClientVersion:       updatedCoreInstance.ClientVersion,
		ClientType:          updatedCoreInstance.ClientType,
		TempInstanceFlag:    updatedCoreInstance.TempInstanceFlag,
		HeartbeatFailCount:  updatedCoreInstance.HeartbeatFailCount,
		MetadataJson:        updatedCoreInstance.MetadataJson,
		TagsJson:            updatedCoreInstance.TagsJson,
		RegisterTime:        updatedCoreInstance.RegisterTime,
		LastHeartbeatTime:   updatedCoreInstance.LastHeartbeatTime,
		LastHealthCheckTime: updatedCoreInstance.LastHealthCheckTime,
		AddTime:             updatedCoreInstance.AddTime,
		AddWho:              updatedCoreInstance.AddWho,
		EditTime:            updatedCoreInstance.EditTime,
		EditWho:             updatedCoreInstance.EditWho,
		OprSeqFlag:          updatedCoreInstance.OprSeqFlag,
		CurrentVersion:      updatedCoreInstance.CurrentVersion,
		ActiveFlag:          updatedCoreInstance.ActiveFlag,
		NoteText:            updatedCoreInstance.NoteText,
		ExtProperty:         updatedCoreInstance.ExtProperty,
		Reserved1:           updatedCoreInstance.Reserved1,
		Reserved2:           updatedCoreInstance.Reserved2,
		Reserved3:           updatedCoreInstance.Reserved3,
		Reserved4:           updatedCoreInstance.Reserved4,
		Reserved5:           updatedCoreInstance.Reserved5,
		Reserved6:           updatedCoreInstance.Reserved6,
		Reserved7:           updatedCoreInstance.Reserved7,
		Reserved8:           updatedCoreInstance.Reserved8,
		Reserved9:           updatedCoreInstance.Reserved9,
		Reserved10:          updatedCoreInstance.Reserved10,
	}

	logger.InfoWithTrace(ctx, "服务实例更新成功",
		"instanceId", instance.ServiceInstanceId,
		"serviceName", instance.ServiceName)

	return resultInstance, nil
}

// DeleteServiceInstanceDirect 直接删除服务实例（已废弃，请使用 DeregisterInstanceViaManager）
//
// 注意：此方法已被废弃，仅用于直接数据库操作，不会触发事件发布和缓存更新
// 请使用 DeregisterInstanceViaManager 方法替代
/*
func (dao *ServiceInstanceDAO) DeleteServiceInstanceDirect(ctx context.Context, tenantId, serviceInstanceId string) error {
	where := "tenantId = ? AND serviceInstanceId = ?"
	args := []interface{}{tenantId, serviceInstanceId}

	affectedRows, err := dao.db.Delete(ctx, "HUB_REGISTRY_SERVICE_INSTANCE", where, args, true)
	if err != nil {
		return huberrors.WrapError(err, "删除服务实例失败")
	}

	if affectedRows == 0 {
		return huberrors.WrapError(fmt.Errorf("未找到要删除的服务实例"), "服务实例不存在")
	}

	return nil
}
*/

// UpdateInstanceHeartbeat 更新服务实例心跳时间
//
// 直接调用注册中心管理器，确保：
// 1. 缓存更新
// 2. 事件发布（EventDBWriter会监听事件并执行数据库操作）
// 3. 完整的生命周期管理
//
// 参数：
//   - ctx: 上下文对象
//   - tenantId: 租户ID
//   - serviceInstanceId: 服务实例ID
//   - heartbeatTime: 心跳时间
//
// 返回值：
//   - error: 错误信息
func (dao *ServiceInstanceDAO) UpdateInstanceHeartbeat(ctx context.Context, tenantId, serviceInstanceId string, heartbeatTime time.Time) error {
	// 获取注册中心管理器实例
	registryManager := manager.GetInstance()

	// 直接通过注册中心管理器更新心跳（Manager会发布事件，EventDBWriter会处理数据库操作）
	err := registryManager.UpdateInstanceHeartbeat(ctx, tenantId, serviceInstanceId)
	if err != nil {
		return huberrors.WrapError(err, "更新实例心跳失败")
	}

	logger.DebugWithTrace(ctx, "实例心跳更新成功",
		"instanceId", serviceInstanceId)

	return nil
}

// UpdateInstanceHealthStatus 更新服务实例健康状态
//
// 直接调用注册中心管理器，确保：
// 1. 缓存更新
// 2. 事件发布（EventDBWriter会监听事件并执行数据库操作）
// 3. 完整的生命周期管理
//
// 参数：
//   - ctx: 上下文对象
//   - tenantId: 租户ID
//   - serviceInstanceId: 服务实例ID
//   - healthStatus: 健康状态
//   - healthCheckTime: 健康检查时间
//
// 返回值：
//   - error: 错误信息
func (dao *ServiceInstanceDAO) UpdateInstanceHealthStatus(ctx context.Context, tenantId, serviceInstanceId, healthStatus string, healthCheckTime time.Time) error {
	// 获取注册中心管理器实例
	registryManager := manager.GetInstance()

	// 直接通过注册中心管理器更新健康状态（Manager会发布事件，EventDBWriter会处理数据库操作）
	err := registryManager.UpdateInstanceHealthStatus(ctx, tenantId, serviceInstanceId, healthStatus, healthCheckTime)
	if err != nil {
		return huberrors.WrapError(err, "更新实例健康状态失败")
	}

	logger.DebugWithTrace(ctx, "实例健康状态更新成功",
		"instanceId", serviceInstanceId,
		"healthStatus", healthStatus)

	return nil
}

// GetInstanceStatusOptions 获取实例状态选项
func (dao *ServiceInstanceDAO) GetInstanceStatusOptions() []string {
	return []string{
		"UP",             // 运行中
		"DOWN",           // 停止
		"STARTING",       // 启动中
		"OUT_OF_SERVICE", // 暂停服务
	}
}

// GetHealthStatusOptions 获取健康状态选项
func (dao *ServiceInstanceDAO) GetHealthStatusOptions() []string {
	return []string{
		"HEALTHY",   // 健康
		"UNHEALTHY", // 不健康
		"UNKNOWN",   // 未知
	}
}

// CreateServiceInstance 创建新服务实例（已废弃，请使用 RegisterInstanceViaManager）
//
// 注意：此方法已被废弃，仅用于直接数据库操作，不会触发事件发布和缓存更新
// 请使用 RegisterInstanceViaManager 方法替代
/*
func (dao *ServiceInstanceDAO) CreateServiceInstance(ctx context.Context, instance *models.ServiceInstance, operatorId string) (*models.ServiceInstance, error) {
	// 设置审计信息
	now := time.Now()
	instance.RegisterTime = now
	instance.AddTime = now
	instance.EditTime = now
	instance.AddWho = operatorId
	instance.EditWho = operatorId
	instance.CurrentVersion = 1
	instance.OprSeqFlag = random.Generate32BitRandomString()

	// 设置默认值
	if instance.ActiveFlag == "" {
		instance.ActiveFlag = "Y"
	}
	if instance.InstanceStatus == "" {
		instance.InstanceStatus = "UP"
	}
	if instance.HealthStatus == "" {
		instance.HealthStatus = "UNKNOWN"
	}
	if instance.ClientType == "" {
		instance.ClientType = "SERVICE"
	}
	if instance.TempInstanceFlag == "" {
		instance.TempInstanceFlag = "N" // 默认为非临时实例
	}
	if instance.ContextPath == "" {
		instance.ContextPath = ""
	}
	if instance.WeightValue == 0 {
		instance.WeightValue = 100
	}

	// 如果未提供实例ID，自动生成一个唯一ID
	if strings.TrimSpace(instance.ServiceInstanceId) == "" {
		// 使用32位随机字符串生成唯一实例ID
		randomID := random.Generate32BitRandomString()
		instance.ServiceInstanceId = fmt.Sprintf("INST-%s", randomID)
		logger.DebugWithTrace(ctx, "DAO层自动生成服务实例ID",
			"serviceInstanceId", instance.ServiceInstanceId,
			"serviceName", instance.ServiceName)
	} else {
		// 检查实例ID是否已存在
		existingInstance, checkErr := dao.GetServiceInstance(ctx, instance.TenantId, instance.ServiceInstanceId, "")
		if checkErr != nil {
			// 如果是"实例不存在"错误，说明可以创建，继续执行
			if !strings.Contains(checkErr.Error(), "实例不存在") {
				return nil, huberrors.WrapError(checkErr, "检查实例ID唯一性失败")
			}
		} else if existingInstance != nil {
			return nil, huberrors.WrapError(fmt.Errorf("实例ID已存在"), "实例ID重复")
		}
	}

	// 验证所属服务是否存在
	serviceDAO := NewServiceDAO(dao.db)
	_, svcErr := serviceDAO.GetService(ctx, instance.TenantId, instance.ServiceName, "Y")
	if svcErr != nil {
		if strings.Contains(svcErr.Error(), "不存在") {
			return nil, huberrors.WrapError(svcErr, "指定的服务不存在或已禁用")
		}
		return nil, huberrors.WrapError(svcErr, "验证服务失败")
	}

	// 插入实例记录
	affectedRows, insertErr := dao.db.Insert(ctx, "HUB_REGISTRY_SERVICE_INSTANCE", instance, true)
	if insertErr != nil {
		return nil, huberrors.WrapError(insertErr, "创建服务实例失败")
	}

	if affectedRows == 0 {
		return nil, huberrors.WrapError(fmt.Errorf("插入实例记录失败"), "创建服务实例失败")
	}

	// 返回创建后的服务实例信息
	createdInstance, queryErr := dao.GetServiceInstance(ctx, instance.TenantId, instance.ServiceInstanceId, "")
	if queryErr != nil {
		return nil, huberrors.WrapError(queryErr, "获取创建后的服务实例信息失败")
	}

	return createdInstance, nil
}
*/

// RegisterInstanceViaManager 通过注册中心管理器注册服务实例（推荐方式）
//
// 此方法会调用 registry_manager.RegisterInstance，确保：
// 1. 事件发布
// 2. 缓存更新
// 3. 完整的生命周期管理
//
// 参数：
//   - ctx: 上下文对象
//   - instance: 服务实例信息对象
//   - operatorId: 操作人员ID
//
// 返回值：
//   - *models.ServiceInstance: 创建后的服务实例信息
//   - error: 错误信息
func (dao *ServiceInstanceDAO) RegisterInstanceViaManager(ctx context.Context, instance *models.ServiceInstance, operatorId string) (*models.ServiceInstance, error) {
	// 获取注册中心管理器实例
	registryManager := manager.GetInstance()

	// 设置默认值
	if instance.ActiveFlag == "" {
		instance.ActiveFlag = "Y"
	}
	if instance.InstanceStatus == "" {
		instance.InstanceStatus = "UP"
	}
	if instance.HealthStatus == "" {
		instance.HealthStatus = "UNKNOWN"
	}
	if instance.ClientType == "" {
		instance.ClientType = "SERVICE"
	}
	if instance.TempInstanceFlag == "" {
		instance.TempInstanceFlag = "N" // 默认为非临时实例
	}
	if instance.ContextPath == "" {
		instance.ContextPath = ""
	}
	if instance.WeightValue == 0 {
		instance.WeightValue = 100
	}

	// 如果未提供实例ID，自动生成一个唯一ID
	if strings.TrimSpace(instance.ServiceInstanceId) == "" {
		randomID := random.Generate32BitRandomString()
		instance.ServiceInstanceId = fmt.Sprintf("INST-%s", randomID)
		logger.DebugWithTrace(ctx, "DAO层自动生成服务实例ID",
			"serviceInstanceId", instance.ServiceInstanceId,
			"serviceName", instance.ServiceName)
	}

	// 转换 models.ServiceInstance 到 core.ServiceInstance
	coreInstance := &core.ServiceInstance{
		ServiceInstanceId:  instance.ServiceInstanceId,
		TenantId:           instance.TenantId,
		ServiceGroupId:     instance.ServiceGroupId,
		ServiceName:        instance.ServiceName,
		GroupName:          instance.GroupName,
		HostAddress:        instance.HostAddress,
		PortNumber:         instance.PortNumber,
		ContextPath:        instance.ContextPath,
		InstanceStatus:     instance.InstanceStatus,
		HealthStatus:       instance.HealthStatus,
		WeightValue:        instance.WeightValue,
		ClientId:           instance.ClientId,
		ClientVersion:      instance.ClientVersion,
		ClientType:         instance.ClientType,
		TempInstanceFlag:   instance.TempInstanceFlag,
		HeartbeatFailCount: 0, // 新注册的实例失败次数为0
		MetadataJson:       instance.MetadataJson,
		TagsJson:           instance.TagsJson,
		RegisterTime:       time.Now(),
		LastHeartbeatTime:  func() *time.Time { t := time.Now(); return &t }(),
		AddTime:            time.Now(),
		AddWho:             operatorId,
		EditTime:           time.Now(),
		EditWho:            operatorId,
		OprSeqFlag:         random.Generate32BitRandomString(),
		CurrentVersion:     1,
		ActiveFlag:         instance.ActiveFlag,
		NoteText:           instance.NoteText,
		ExtProperty:        instance.ExtProperty,
	}

	// 通过注册中心管理器注册实例
	registeredInstance, err := registryManager.RegisterInstance(ctx, coreInstance)
	if err != nil {
		return nil, huberrors.WrapError(err, "通过注册中心管理器注册服务实例失败")
	}

	// 转换回 models.ServiceInstance 并返回
	resultInstance := &models.ServiceInstance{
		ServiceInstanceId:   registeredInstance.ServiceInstanceId,
		TenantId:            registeredInstance.TenantId,
		ServiceGroupId:      registeredInstance.ServiceGroupId,
		ServiceName:         registeredInstance.ServiceName,
		GroupName:           registeredInstance.GroupName,
		HostAddress:         registeredInstance.HostAddress,
		PortNumber:          registeredInstance.PortNumber,
		ContextPath:         registeredInstance.ContextPath,
		InstanceStatus:      registeredInstance.InstanceStatus,
		HealthStatus:        registeredInstance.HealthStatus,
		WeightValue:         registeredInstance.WeightValue,
		ClientId:            registeredInstance.ClientId,
		ClientVersion:       registeredInstance.ClientVersion,
		ClientType:          registeredInstance.ClientType,
		TempInstanceFlag:    registeredInstance.TempInstanceFlag,
		HeartbeatFailCount:  registeredInstance.HeartbeatFailCount,
		MetadataJson:        registeredInstance.MetadataJson,
		TagsJson:            registeredInstance.TagsJson,
		RegisterTime:        registeredInstance.RegisterTime,
		LastHeartbeatTime:   registeredInstance.LastHeartbeatTime,
		LastHealthCheckTime: registeredInstance.LastHealthCheckTime,
		AddTime:             registeredInstance.AddTime,
		AddWho:              registeredInstance.AddWho,
		EditTime:            registeredInstance.EditTime,
		EditWho:             registeredInstance.EditWho,
		OprSeqFlag:          registeredInstance.OprSeqFlag,
		CurrentVersion:      registeredInstance.CurrentVersion,
		ActiveFlag:          registeredInstance.ActiveFlag,
		NoteText:            registeredInstance.NoteText,
		ExtProperty:         registeredInstance.ExtProperty,
		Reserved1:           registeredInstance.Reserved1,
		Reserved2:           registeredInstance.Reserved2,
		Reserved3:           registeredInstance.Reserved3,
		Reserved4:           registeredInstance.Reserved4,
		Reserved5:           registeredInstance.Reserved5,
		Reserved6:           registeredInstance.Reserved6,
		Reserved7:           registeredInstance.Reserved7,
		Reserved8:           registeredInstance.Reserved8,
		Reserved9:           registeredInstance.Reserved9,
		Reserved10:          registeredInstance.Reserved10,
	}

	return resultInstance, nil
}

// DeregisterInstanceViaManager 通过注册中心管理器注销服务实例（推荐方式）
//
// 此方法会调用 registry_manager.DeregisterInstance，确保：
// 1. 事件发布
// 2. 缓存清理
// 3. 完整的生命周期管理
//
// 参数：
//   - ctx: 上下文对象
//   - tenantId: 租户ID
//   - serviceInstanceId: 服务实例ID
//
// 返回值：
//   - error: 错误信息
func (dao *ServiceInstanceDAO) DeregisterInstanceViaManager(ctx context.Context, tenantId, serviceInstanceId string) error {
	// 获取注册中心管理器实例
	registryManager := manager.GetInstance()

	// 通过注册中心管理器注销实例
	err := registryManager.DeregisterInstance(ctx, tenantId, serviceInstanceId)
	if err != nil {
		return huberrors.WrapError(err, "通过注册中心管理器注销服务实例失败")
	}

	return nil
}

// CreateServiceInstance 创建新服务实例（别名方法，调用 RegisterInstanceViaManager）
//
// 此方法是 RegisterInstanceViaManager 的别名，用于保持向后兼容性
// 新代码建议直接使用 RegisterInstanceViaManager
//
// 参数：
//   - ctx: 上下文对象
//   - instance: 服务实例信息对象
//   - operatorId: 操作人员ID
//
// 返回值：
//   - *models.ServiceInstance: 创建后的服务实例信息
//   - error: 错误信息
func (dao *ServiceInstanceDAO) CreateServiceInstance(ctx context.Context, instance *models.ServiceInstance, operatorId string) (*models.ServiceInstance, error) {
	return dao.RegisterInstanceViaManager(ctx, instance, operatorId)
}

// DeleteServiceInstance 删除服务实例（别名方法，调用 DeregisterInstanceViaManager）
//
// 此方法是 DeregisterInstanceViaManager 的别名，用于保持向后兼容性
// 新代码建议直接使用 DeregisterInstanceViaManager
//
// 参数：
//   - ctx: 上下文对象
//   - tenantId: 租户ID
//   - serviceInstanceId: 服务实例ID
//
// 返回值：
//   - error: 错误信息
func (dao *ServiceInstanceDAO) DeleteServiceInstance(ctx context.Context, tenantId, serviceInstanceId string) error {
	return dao.DeregisterInstanceViaManager(ctx, tenantId, serviceInstanceId)
}
