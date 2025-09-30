package dao

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gateway/internal/registry/core"    // 用于核心数据类型转换
	"gateway/internal/registry/manager" // 用于调用注册中心管理器
	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
	"gateway/pkg/logger"
	"gateway/pkg/utils/huberrors"
	"gateway/pkg/utils/random"
	"gateway/web/views/hub0041/models"
)

// ServiceDAO 服务数据访问对象
// 负责服务注册信息的数据库操作（查询、编辑、删除，不提供新增）
type ServiceDAO struct {
	db database.Database
}

// NewServiceDAO 创建服务DAO实例
func NewServiceDAO(db database.Database) *ServiceDAO {
	return &ServiceDAO{
		db: db,
	}
}

// QueryServices 分页查询服务列表
// 使用sqlutils进行多数据库兼容的分页查询
func (dao *ServiceDAO) QueryServices(ctx context.Context, req *models.ServiceQueryRequest) ([]*models.Service, int, error) {
	// 构建查询条件
	whereClause := "WHERE 1=1"
	var params []interface{}

	if req.TenantId != "" {
		whereClause += " AND tenantId = ?"
		params = append(params, req.TenantId)
	}

	if req.ActiveFlag != "" {
		whereClause += " AND activeFlag = ?"
		params = append(params, req.ActiveFlag)
	}

	if req.GroupName != "" {
		whereClause += " AND groupName = ?"
		params = append(params, req.GroupName)
	}

	if req.ServiceName != "" {
		whereClause += " AND serviceName LIKE ?"
		params = append(params, "%"+req.ServiceName+"%")
	}

	if req.ProtocolType != "" {
		whereClause += " AND protocolType = ?"
		params = append(params, req.ProtocolType)
	}

	if req.RegistryType != "" {
		whereClause += " AND registryType = ?"
		params = append(params, req.RegistryType)
	}

	if req.Keyword != "" {
		whereClause += " AND (serviceName LIKE ? OR serviceDescription LIKE ?)"
		keyword := "%" + req.Keyword + "%"
		params = append(params, keyword, keyword)
	}

	// 构建基础查询语句 - 查询完整的服务信息
	baseQuery := fmt.Sprintf(`
		SELECT tenantId, serviceName, serviceGroupId, groupName, serviceDescription,
			   registryType, externalRegistryConfig,
			   protocolType, contextPath, loadBalanceStrategy,
			   healthCheckUrl, healthCheckIntervalSeconds, healthCheckTimeoutSeconds, healthCheckType, healthCheckMode,
			   metadataJson, tagsJson,
			   addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty,
			   reserved1, reserved2, reserved3, reserved4, reserved5, reserved6, reserved7, reserved8, reserved9, reserved10
		FROM HUB_REGISTRY_SERVICE %s
		ORDER BY addTime DESC
	`, whereClause)

	// 构建统计查询
	countQuery, err := sqlutils.BuildCountQuery(baseQuery)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建统计查询失败")
	}

	// 执行统计查询
	var countResult struct {
		Count int `db:"COUNT(*)"`
	}
	err = dao.db.QueryOne(ctx, &countResult, countQuery, params, true)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询服务总数失败", "error", err)
		return nil, 0, huberrors.WrapError(err, "查询服务总数失败")
	}

	// 如果没有记录，直接返回空列表
	if countResult.Count == 0 {
		return []*models.Service{}, 0, nil
	}

	// 创建分页信息
	pagination := sqlutils.NewPaginationInfo(req.PageIndex, req.PageSize)

	// 获取数据库类型
	dbType := sqlutils.GetDatabaseType(dao.db)

	// 构建分页查询
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建分页查询失败")
	}

	// 合并查询参数
	allArgs := append(params, paginationArgs...)

	// 执行分页查询
	var services []*models.Service
	err = dao.db.Query(ctx, &services, paginatedQuery, allArgs, true)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询服务数据失败", "error", err)
		return nil, 0, huberrors.WrapError(err, "查询服务数据失败")
	}

	// 为每个服务关联实例信息，避免与service_instance_dao.go形成死循环
	if len(services) > 0 {
		registryManager := manager.GetInstance()
		for i := range services {
			// 从注册中心管理器获取实例列表
			coreInstances, err := registryManager.ListInstances(ctx, req.TenantId, services[i].ServiceGroupId, services[i].ServiceName)
			if err != nil {
				logger.WarnWithTrace(ctx, "从注册中心管理器获取实例失败", "error", err, "serviceName", services[i].ServiceName)
				// 获取实例失败不影响服务查询结果，返回空实例列表
				services[i].Instances = []*models.ServiceInstance{}
			} else {
				// 转换 core.ServiceInstance 到 models.ServiceInstance
				services[i].Instances = make([]*models.ServiceInstance, 0, len(coreInstances))
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
					services[i].Instances = append(services[i].Instances, modelInstance)
				}
			}
		}
	}

	return services, countResult.Count, nil
}

// QueryServicesWithInstances 分页查询服务列表（包含实例信息）
// 保持向后兼容，直接调用QueryServices方法
func (dao *ServiceDAO) QueryServicesWithInstances(ctx context.Context, tenantId, activeFlag, groupName, serviceName, protocolType string, page, pageSize int) ([]*models.Service, int, error) {
	// 构建查询请求
	req := &models.ServiceQueryRequest{
		TenantId:     tenantId,
		ActiveFlag:   activeFlag,
		GroupName:    groupName,
		ServiceName:  serviceName,
		ProtocolType: protocolType,
		PageIndex:    page,
		PageSize:     pageSize,
	}

	// 直接调用QueryServices方法，现在已经返回完整的服务信息和实例
	return dao.QueryServices(ctx, req)
}

// GetService 根据服务名称获取服务详情
//
// 参数：
//   - ctx: 上下文对象
//   - tenantId: 租户ID
//   - serviceName: 服务名称
//   - activeFlag: 活动状态标记(为空则不过滤)
//
// 返回值：
//   - *models.Service: 服务信息（包含关联的服务实例）
//   - error: 错误信息
func (dao *ServiceDAO) GetService(ctx context.Context, tenantId, serviceName, activeFlag string) (*models.Service, error) {
	whereConditions := []string{"tenantId = ?", "serviceName = ?"}
	args := []interface{}{tenantId, serviceName}

	if activeFlag != "" {
		whereConditions = append(whereConditions, "activeFlag = ?")
		args = append(args, activeFlag)
	}

	whereClause := strings.Join(whereConditions, " AND ")

	query := `SELECT tenantId, serviceName, serviceGroupId, groupName, serviceDescription,
		registryType, externalRegistryConfig,
		protocolType, contextPath, loadBalanceStrategy,
		healthCheckUrl, healthCheckIntervalSeconds, healthCheckTimeoutSeconds, healthCheckType, healthCheckMode,
		metadataJson, tagsJson,
		addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty,
		reserved1, reserved2, reserved3, reserved4, reserved5, reserved6, reserved7, reserved8, reserved9, reserved10
	FROM HUB_REGISTRY_SERVICE WHERE ` + whereClause

	service := &models.Service{}
	err := dao.db.QueryOne(ctx, service, query, args, true)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") || strings.Contains(err.Error(), "not found") {
			return nil, huberrors.WrapError(err, "服务不存在")
		}
		return nil, huberrors.WrapError(err, "获取服务信息失败")
	}

	// 从注册中心管理器获取服务实例，避免与service_instance_dao.go形成死循环
	registryManager := manager.GetInstance()
	coreInstances, err := registryManager.ListInstances(ctx, tenantId, service.ServiceGroupId, service.ServiceName)
	if err != nil {
		logger.WarnWithTrace(ctx, "从注册中心管理器获取实例失败", "error", err, "serviceName", service.ServiceName)
		// 获取实例失败不影响服务查询结果，返回空实例列表
		service.Instances = []*models.ServiceInstance{}
	} else {
		// 转换 core.ServiceInstance 到 models.ServiceInstance
		service.Instances = make([]*models.ServiceInstance, 0, len(coreInstances))
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
			service.Instances = append(service.Instances, modelInstance)
		}
	}

	return service, nil
}

// UpdateService 更新服务信息（推荐方式）
//
// 此方法会调用 registry_manager.UpdateService，确保：
// 1. 事件发布
// 2. 缓存更新
// 3. 完整的生命周期管理
//
// 参数：
//   - ctx: 上下文对象
//   - service: 服务信息对象
//   - operatorId: 操作人员ID
//
// 返回值：
//   - *models.Service: 更新后的服务信息
//   - error: 错误信息
func (dao *ServiceDAO) UpdateService(ctx context.Context, service *models.Service, operatorId string) (*models.Service, error) {
	// 获取注册中心管理器实例
	registryManager := manager.GetInstance()

	// 更新审计信息
	service.EditTime = time.Now()
	service.EditWho = operatorId
	service.CurrentVersion++
	service.OprSeqFlag = random.Generate32BitRandomString()

	// 转换 models.Service 到 core.Service
	coreService := &core.Service{
		TenantId:                   service.TenantId,
		ServiceName:                service.ServiceName,
		ServiceGroupId:             service.ServiceGroupId,
		GroupName:                  service.GroupName,
		ServiceDescription:         service.ServiceDescription,
		RegistryType:               service.RegistryType,
		ExternalRegistryConfig:     service.ExternalRegistryConfig,
		ProtocolType:               service.ProtocolType,
		ContextPath:                service.ContextPath,
		LoadBalanceStrategy:        service.LoadBalanceStrategy,
		HealthCheckUrl:             service.HealthCheckUrl,
		HealthCheckIntervalSeconds: service.HealthCheckIntervalSeconds,
		HealthCheckTimeoutSeconds:  service.HealthCheckTimeoutSeconds,
		HealthCheckType:            service.HealthCheckType,
		HealthCheckMode:            service.HealthCheckMode,
		MetadataJson:               service.MetadataJson,
		TagsJson:                   service.TagsJson,
		AddTime:                    service.AddTime,
		AddWho:                     service.AddWho,
		EditTime:                   service.EditTime,
		EditWho:                    service.EditWho,
		OprSeqFlag:                 service.OprSeqFlag,
		CurrentVersion:             service.CurrentVersion,
		ActiveFlag:                 service.ActiveFlag,
		NoteText:                   service.NoteText,
		ExtProperty:                service.ExtProperty,
	}

	// 通过注册中心管理器更新服务（RegisterService方法内部会判断是更新还是新增）
	updatedCoreService, err := registryManager.RegisterService(ctx, coreService)
	if err != nil {
		return nil, huberrors.WrapError(err, "通过注册中心管理器更新服务失败")
	}

	// 转换回 models.Service 并返回
	resultService := &models.Service{
		TenantId:                   updatedCoreService.TenantId,
		ServiceName:                updatedCoreService.ServiceName,
		ServiceGroupId:             updatedCoreService.ServiceGroupId,
		GroupName:                  updatedCoreService.GroupName,
		ServiceDescription:         updatedCoreService.ServiceDescription,
		RegistryType:               updatedCoreService.RegistryType,
		ExternalRegistryConfig:     updatedCoreService.ExternalRegistryConfig,
		ProtocolType:               updatedCoreService.ProtocolType,
		ContextPath:                updatedCoreService.ContextPath,
		LoadBalanceStrategy:        updatedCoreService.LoadBalanceStrategy,
		HealthCheckUrl:             updatedCoreService.HealthCheckUrl,
		HealthCheckIntervalSeconds: updatedCoreService.HealthCheckIntervalSeconds,
		HealthCheckTimeoutSeconds:  updatedCoreService.HealthCheckTimeoutSeconds,
		HealthCheckType:            updatedCoreService.HealthCheckType,
		HealthCheckMode:            updatedCoreService.HealthCheckMode,
		MetadataJson:               updatedCoreService.MetadataJson,
		TagsJson:                   updatedCoreService.TagsJson,
		AddTime:                    updatedCoreService.AddTime,
		AddWho:                     updatedCoreService.AddWho,
		EditTime:                   updatedCoreService.EditTime,
		EditWho:                    updatedCoreService.EditWho,
		OprSeqFlag:                 updatedCoreService.OprSeqFlag,
		CurrentVersion:             updatedCoreService.CurrentVersion,
		ActiveFlag:                 updatedCoreService.ActiveFlag,
		NoteText:                   updatedCoreService.NoteText,
		ExtProperty:                updatedCoreService.ExtProperty,
	}

	return resultService, nil
}

// DeleteService 删除服务（已废弃，请使用 DeregisterServiceViaManager）
//
// 注意：此方法已被废弃，仅用于直接数据库操作，不会触发事件发布和缓存更新
// 请使用 DeregisterServiceViaManager 方法替代
/*
func (dao *ServiceDAO) DeleteService(ctx context.Context, tenantId, serviceName string) error {
	// 先查询服务是否存在
	_, err := dao.GetService(ctx, tenantId, serviceName, "")
	if err != nil {
		return err // 已经包含了错误信息
	}

	// 级联删除实例
	// 查询所有关联的实例
	instanceDAO := NewServiceInstanceDAO(dao.db)
	instances, _, err := instanceDAO.QueryServiceInstances(ctx, tenantId, "", serviceName, "", "", "", "", 1, 1000)
	if err != nil {
		logger.WarnWithTrace(ctx, "查询服务实例失败，无法级联删除", "error", err, "serviceName", serviceName)
	} else if len(instances) > 0 {
		// 删除所有关联的实例
		logger.InfoWithTrace(ctx, "开始级联删除服务实例", "serviceName", serviceName, "instanceCount", len(instances))

		// 构建删除条件
		instanceWhere := "tenantId = ? AND serviceName = ?"
		instanceArgs := []interface{}{tenantId, serviceName}

		// 执行批量删除
		_, err := dao.db.Delete(ctx, "HUB_REGISTRY_SERVICE_INSTANCE", instanceWhere, instanceArgs, true)
		if err != nil {
			logger.WarnWithTrace(ctx, "级联删除服务实例失败", "error", err, "serviceName", serviceName)
			// 继续执行服务删除，不因实例删除失败而中断
		} else {
			logger.InfoWithTrace(ctx, "级联删除服务实例成功", "serviceName", serviceName, "instanceCount", len(instances))
		}
	}

	// 删除服务
	where := "tenantId = ? AND serviceName = ?"
	args := []interface{}{tenantId, serviceName}

	affectedRows, err := dao.db.Delete(ctx, "HUB_REGISTRY_SERVICE", where, args, true)
	if err != nil {
		return huberrors.WrapError(err, "删除服务失败")
	}

	if affectedRows == 0 {
		return huberrors.WrapError(fmt.Errorf("未找到要删除的服务"), "服务不存在")
	}

	logger.InfoWithTrace(ctx, "删除服务成功", "serviceName", serviceName)
	return nil
}
*/

// GetServiceProtocolTypes 获取支持的协议类型列表
func (dao *ServiceDAO) GetServiceProtocolTypes() []string {
	return []string{
		"HTTP",
		"HTTPS",
		"TCP",
		"UDP",
		"GRPC",
	}
}

// CreateService 创建新服务（已废弃，请使用 RegisterServiceViaManager）
//
// 注意：此方法已被废弃，仅用于直接数据库操作，不会触发事件发布和缓存更新
// 请使用 RegisterServiceViaManager 方法替代
/*
func (dao *ServiceDAO) CreateService(ctx context.Context, service *models.Service, operatorId string) (*models.Service, error) {
	// 设置审计信息
	now := time.Now()
	service.AddTime = now
	service.EditTime = now
	service.AddWho = operatorId
	service.EditWho = operatorId
	service.CurrentVersion = 1
	service.OprSeqFlag = random.Generate32BitRandomString()

	// 设置默认值
	if service.ActiveFlag == "" {
		service.ActiveFlag = "Y"
	}
	if service.ProtocolType == "" {
		service.ProtocolType = "HTTP"
	}
	if service.LoadBalanceStrategy == "" {
		service.LoadBalanceStrategy = "ROUND_ROBIN"
	}
	if service.ContextPath == "" {
		service.ContextPath = ""
	}
	if service.HealthCheckUrl == "" {
		service.HealthCheckUrl = "/health"
	}
	if service.HealthCheckIntervalSeconds == 0 {
		service.HealthCheckIntervalSeconds = 30
	}
	if service.HealthCheckTimeoutSeconds == 0 {
		service.HealthCheckTimeoutSeconds = 5
	}
	if service.HealthCheckType == "" {
		service.HealthCheckType = "HTTP" // 默认为HTTP健康检查
	}
	if service.HealthCheckMode == "" {
		service.HealthCheckMode = "ACTIVE" // 默认为主动探测模式
	}

	// 检查服务名称是否已存在
	existingService, err := dao.GetService(ctx, service.TenantId, service.ServiceName, "")
	if err != nil {
		// 如果是"服务不存在"错误，说明可以创建，继续执行
		if !strings.Contains(err.Error(), "服务不存在") {
			return nil, huberrors.WrapError(err, "检查服务名称唯一性失败")
		}
	} else if existingService != nil {
		return nil, huberrors.WrapError(fmt.Errorf("服务名称已存在"), "服务名称重复")
	}

	// 验证服务分组是否存在
	serviceGroupDAO := NewServiceGroupDAO(dao.db)
	_, err = serviceGroupDAO.GetServiceGroup(ctx, service.TenantId, service.ServiceGroupId, "Y")
	if err != nil {
		if strings.Contains(err.Error(), "服务分组不存在") {
			return nil, huberrors.WrapError(err, "指定的服务分组不存在或已禁用")
		}
		return nil, huberrors.WrapError(err, "验证服务分组失败")
	}

	// 插入服务记录
	affectedRows, err := dao.db.Insert(ctx, "HUB_REGISTRY_SERVICE", service, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "创建服务失败")
	}

	if affectedRows == 0 {
		return nil, huberrors.WrapError(fmt.Errorf("插入服务记录失败"), "创建服务失败")
	}

	// 返回创建后的服务信息，包含实例列表（虽然新创建的服务还没有实例）
	createdService, err := dao.GetService(ctx, service.TenantId, service.ServiceName, "")
	if err != nil {
		return nil, huberrors.WrapError(err, "获取创建后的服务信息失败")
	}

	return createdService, nil
}
*/

// GetLoadBalanceStrategies 获取支持的负载均衡策略列表
func (dao *ServiceDAO) GetLoadBalanceStrategies() []string {
	return []string{
		"ROUND_ROBIN",          // 轮询
		"WEIGHTED_ROUND_ROBIN", // 加权轮询
		"LEAST_CONNECTIONS",    // 最少连接数
		"IP_HASH",              // IP哈希
		"RANDOM",               // 随机
	}
}

// GetHealthCheckTypeOptions 获取健康检查类型选项
func (dao *ServiceDAO) GetHealthCheckTypeOptions() []string {
	return []string{
		"HTTP", // HTTP健康检查
		"TCP",  // TCP健康检查
	}
}

// GetHealthCheckModeOptions 获取健康检查模式选项
func (dao *ServiceDAO) GetHealthCheckModeOptions() []string {
	return []string{
		"ACTIVE",  // 主动探测模式
		"PASSIVE", // 客户端上报模式
	}
}

// GetRegistryTypeOptions 获取注册类型选项
func (dao *ServiceDAO) GetRegistryTypeOptions() []string {
	return []string{
		"INTERNAL",  // 内部管理（默认）
		"NACOS",     // Nacos注册中心
		"CONSUL",    // Consul注册中心
		"EUREKA",    // Eureka注册中心
		"ETCD",      // ETCD注册中心
		"ZOOKEEPER", // ZooKeeper注册中心
	}
}

// RegisterServiceViaManager 通过注册中心管理器注册服务（推荐方式）
//
// 此方法会调用 registry_manager.RegisterService，确保：
// 1. 事件发布
// 2. 缓存更新
// 3. 完整的生命周期管理
//
// 参数：
//   - ctx: 上下文对象
//   - service: 服务信息对象
//   - operatorId: 操作人员ID
//
// 返回值：
//   - *models.Service: 创建后的服务信息
//   - error: 错误信息
func (dao *ServiceDAO) RegisterServiceViaManager(ctx context.Context, service *models.Service, operatorId string) (*models.Service, error) {
	// 获取注册中心管理器实例
	registryManager := manager.GetInstance()

	// 设置默认值
	if service.ActiveFlag == "" {
		service.ActiveFlag = "Y"
	}
	if service.RegistryType == "" {
		service.RegistryType = "INTERNAL" // 默认为内部管理
	}
	if service.ProtocolType == "" {
		service.ProtocolType = "HTTP"
	}
	if service.LoadBalanceStrategy == "" {
		service.LoadBalanceStrategy = "ROUND_ROBIN"
	}
	if service.ContextPath == "" {
		service.ContextPath = ""
	}
	if service.HealthCheckUrl == "" {
		service.HealthCheckUrl = "/health"
	}
	if service.HealthCheckIntervalSeconds == 0 {
		service.HealthCheckIntervalSeconds = 30
	}
	if service.HealthCheckTimeoutSeconds == 0 {
		service.HealthCheckTimeoutSeconds = 5
	}
	if service.HealthCheckType == "" {
		service.HealthCheckType = "HTTP" // 默认为HTTP健康检查
	}
	if service.HealthCheckMode == "" {
		service.HealthCheckMode = "ACTIVE" // 默认为主动探测模式
	}

	// 转换 models.Service 到 core.Service
	coreService := &core.Service{
		TenantId:                   service.TenantId,
		ServiceName:                service.ServiceName,
		ServiceGroupId:             service.ServiceGroupId,
		GroupName:                  service.GroupName,
		ServiceDescription:         service.ServiceDescription,
		RegistryType:               service.RegistryType,
		ExternalRegistryConfig:     service.ExternalRegistryConfig,
		ProtocolType:               service.ProtocolType,
		ContextPath:                service.ContextPath,
		LoadBalanceStrategy:        service.LoadBalanceStrategy,
		HealthCheckUrl:             service.HealthCheckUrl,
		HealthCheckIntervalSeconds: service.HealthCheckIntervalSeconds,
		HealthCheckTimeoutSeconds:  service.HealthCheckTimeoutSeconds,
		HealthCheckType:            service.HealthCheckType,
		HealthCheckMode:            service.HealthCheckMode,
		MetadataJson:               service.MetadataJson,
		TagsJson:                   service.TagsJson,
		AddTime:                    time.Now(),
		AddWho:                     operatorId,
		EditTime:                   time.Now(),
		EditWho:                    operatorId,
		OprSeqFlag:                 random.Generate32BitRandomString(),
		CurrentVersion:             1,
		ActiveFlag:                 service.ActiveFlag,
		NoteText:                   service.NoteText,
		ExtProperty:                service.ExtProperty,
	}

	// 通过注册中心管理器注册服务
	registeredService, err := registryManager.RegisterService(ctx, coreService)
	if err != nil {
		return nil, huberrors.WrapError(err, "通过注册中心管理器注册服务失败")
	}

	// 转换回 models.Service 并返回
	resultService := &models.Service{
		TenantId:                   registeredService.TenantId,
		ServiceName:                registeredService.ServiceName,
		ServiceGroupId:             registeredService.ServiceGroupId,
		GroupName:                  registeredService.GroupName,
		ServiceDescription:         registeredService.ServiceDescription,
		RegistryType:               registeredService.RegistryType,
		ExternalRegistryConfig:     registeredService.ExternalRegistryConfig,
		ProtocolType:               registeredService.ProtocolType,
		ContextPath:                registeredService.ContextPath,
		LoadBalanceStrategy:        registeredService.LoadBalanceStrategy,
		HealthCheckUrl:             registeredService.HealthCheckUrl,
		HealthCheckIntervalSeconds: registeredService.HealthCheckIntervalSeconds,
		HealthCheckTimeoutSeconds:  registeredService.HealthCheckTimeoutSeconds,
		HealthCheckType:            registeredService.HealthCheckType,
		HealthCheckMode:            registeredService.HealthCheckMode,
		MetadataJson:               registeredService.MetadataJson,
		TagsJson:                   registeredService.TagsJson,
		AddTime:                    registeredService.AddTime,
		AddWho:                     registeredService.AddWho,
		EditTime:                   registeredService.EditTime,
		EditWho:                    registeredService.EditWho,
		OprSeqFlag:                 registeredService.OprSeqFlag,
		CurrentVersion:             registeredService.CurrentVersion,
		ActiveFlag:                 registeredService.ActiveFlag,
		NoteText:                   registeredService.NoteText,
		ExtProperty:                registeredService.ExtProperty,
	}

	return resultService, nil
}

// DeregisterServiceViaManager 通过注册中心管理器注销服务（推荐方式）
//
// 此方法会调用 registry_manager.DeregisterService，确保：
// 1. 事件发布
// 2. 缓存清理
// 3. 完整的生命周期管理
//
// 参数：
//   - ctx: 上下文对象
//   - tenantId: 租户ID
//   - serviceName: 服务名称
//
// 返回值：
//   - error: 错误信息
func (dao *ServiceDAO) DeregisterServiceViaManager(ctx context.Context, tenantId, serviceName string) error {
	// 获取注册中心管理器实例
	registryManager := manager.GetInstance()

	// 首先获取服务信息以获取 serviceGroupId
	service, err := dao.GetService(ctx, tenantId, serviceName, "")
	if err != nil {
		return huberrors.WrapError(err, "获取服务信息失败")
	}

	// 通过注册中心管理器注销服务
	err = registryManager.DeregisterService(ctx, tenantId, service.ServiceGroupId, serviceName)
	if err != nil {
		return huberrors.WrapError(err, "通过注册中心管理器注销服务失败")
	}

	return nil
}

// CreateService 创建新服务（别名方法，调用 RegisterServiceViaManager）
//
// 此方法是 RegisterServiceViaManager 的别名，用于保持向后兼容性
// 新代码建议直接使用 RegisterServiceViaManager
//
// 参数：
//   - ctx: 上下文对象
//   - service: 服务信息对象
//   - operatorId: 操作人员ID
//
// 返回值：
//   - *models.Service: 创建后的服务信息
//   - error: 错误信息
func (dao *ServiceDAO) CreateService(ctx context.Context, service *models.Service, operatorId string) (*models.Service, error) {
	return dao.RegisterServiceViaManager(ctx, service, operatorId)
}

// DeleteService 删除服务（别名方法，调用 DeregisterServiceViaManager）
//
// 此方法是 DeregisterServiceViaManager 的别名，用于保持向后兼容性
// 新代码建议直接使用 DeregisterServiceViaManager
//
// 参数：
//   - ctx: 上下文对象
//   - tenantId: 租户ID
//   - serviceName: 服务名称
//
// 返回值：
//   - error: 错误信息
func (dao *ServiceDAO) DeleteService(ctx context.Context, tenantId, serviceName string) error {
	return dao.DeregisterServiceViaManager(ctx, tenantId, serviceName)
}
