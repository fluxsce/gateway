package dao

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gateway/internal/registry/core"
	"gateway/internal/registry/manager"
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/pkg/utils/huberrors"
	"gateway/pkg/utils/random"
	"gateway/web/views/hub0040/models"
)

// ServiceGroupDAO 服务分组数据访问对象
// 基于数据库接口实现，提供服务分组的 CRUD 操作
type ServiceGroupDAO struct {
	db database.Database
}

// NewServiceGroupDAO 创建服务分组DAO实例
func NewServiceGroupDAO(db database.Database) *ServiceGroupDAO {
	return &ServiceGroupDAO{
		db: db,
	}
}

// CreateServiceGroup 创建服务分组
//
// 参数：
//   - ctx: 上下文对象
//   - group: 服务分组信息
//   - operatorId: 操作人员ID
//
// 返回值：
//   - *models.ServiceGroup: 创建成功的服务分组信息
//   - error: 错误信息
func (dao *ServiceGroupDAO) CreateServiceGroup(ctx context.Context, group *models.ServiceGroup, operatorId string) (*models.ServiceGroup, error) {
	// 如果没有设置ID，生成唯一的服务分组ID
	if group.ServiceGroupId == "" {
		group.ServiceGroupId = random.Generate32BitRandomString()
	}

	// 设置审计信息
	now := time.Now()
	group.AddTime = now
	group.AddWho = operatorId
	group.EditTime = now
	group.EditWho = operatorId
	group.OprSeqFlag = random.Generate32BitRandomString()
	group.CurrentVersion = 1
	group.ActiveFlag = "Y"

	// 设置默认值
	if group.GroupType == "" {
		group.GroupType = "BUSINESS"
	}
	if group.AccessControlEnabled == "" {
		group.AccessControlEnabled = "N"
	}
	if group.DefaultProtocolType == "" {
		group.DefaultProtocolType = "HTTP"
	}
	if group.DefaultLoadBalanceStrategy == "" {
		group.DefaultLoadBalanceStrategy = "ROUND_ROBIN"
	}
	if group.DefaultHealthCheckUrl == "" {
		group.DefaultHealthCheckUrl = "/health"
	}
	if group.DefaultHealthCheckIntervalSeconds == 0 {
		group.DefaultHealthCheckIntervalSeconds = 30
	}

	// 使用数据库的Insert方法保存
	_, err := dao.db.Insert(ctx, "HUB_REGISTRY_SERVICE_GROUP", group, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "创建服务分组失败")
	}

	// 重新查询获取完整记录（包含数据库自动设置的字段）
	savedGroup, err := dao.GetServiceGroupById(ctx, group.TenantId, group.ServiceGroupId)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询新创建的服务分组失败")
	}

	// 同步到注册中心缓存
	err = dao.syncToRegistryManager(ctx, savedGroup, "CREATE")
	if err != nil {
		logger.WarnWithTrace(ctx, "同步服务分组到注册中心失败",
			"serviceGroupId", savedGroup.ServiceGroupId,
			"groupName", savedGroup.GroupName,
			"error", err)
		// 注册中心同步失败不影响数据库操作的成功
	}

	return savedGroup, nil
}

// UpdateServiceGroup 更新服务分组
//
// 参数：
//   - ctx: 上下文对象
//   - group: 要更新的服务分组信息
//   - operatorId: 操作人员ID
//
// 返回值：
//   - *models.ServiceGroup: 更新后的服务分组信息
//   - error: 错误信息
func (dao *ServiceGroupDAO) UpdateServiceGroup(ctx context.Context, group *models.ServiceGroup, operatorId string) (*models.ServiceGroup, error) {
	// 更新审计信息
	group.EditTime = time.Now()
	group.EditWho = operatorId
	group.CurrentVersion++ // 递增版本号

	// 使用主键更新
	where := "tenantId = ? AND serviceGroupId = ?"
	args := []interface{}{group.TenantId, group.ServiceGroupId}

	affectedRows, err := dao.db.Update(ctx, "HUB_REGISTRY_SERVICE_GROUP", group, where, args, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "更新服务分组失败")
	}

	if affectedRows == 0 {
		return nil, errors.New("未找到要更新的服务分组记录")
	}

	// 重新查询获取完整的更新后记录
	updatedGroup, err := dao.GetServiceGroupById(ctx, group.TenantId, group.ServiceGroupId)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询更新后的服务分组失败")
	}

	// 同步到注册中心缓存
	err = dao.syncToRegistryManager(ctx, updatedGroup, "UPDATE")
	if err != nil {
		logger.WarnWithTrace(ctx, "同步服务分组更新到注册中心失败",
			"serviceGroupId", updatedGroup.ServiceGroupId,
			"groupName", updatedGroup.GroupName,
			"error", err)
		// 注册中心同步失败不影响数据库操作的成功
	}

	return updatedGroup, nil
}

// GetServiceGroup 获取服务分组
//
// 参数：
//   - ctx: 上下文对象
//   - tenantId: 租户ID
//   - groupName: 分组名称
//   - activeFlag: 活动状态标记(Y活动,N非活动,空为全部)
//
// 返回值：
//   - *models.ServiceGroup: 服务分组信息
//   - error: 错误信息
func (dao *ServiceGroupDAO) GetServiceGroup(ctx context.Context, tenantId, groupName, activeFlag string) (*models.ServiceGroup, error) {
	// 构建查询条件
	where := "tenantId = ? AND groupName = ?"
	args := []interface{}{tenantId, groupName}

	if activeFlag != "" {
		where += " AND activeFlag = ?"
		args = append(args, activeFlag)
	}

	// 构建查询SQL
	query := fmt.Sprintf(`SELECT serviceGroupId, tenantId, groupName, groupDescription, groupType,
		ownerUserId, adminUserIds, readUserIds, accessControlEnabled,
		defaultProtocolType, defaultLoadBalanceStrategy, defaultHealthCheckUrl, defaultHealthCheckIntervalSeconds,
		addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty,
		reserved1, reserved2, reserved3, reserved4, reserved5, reserved6, reserved7, reserved8, reserved9, reserved10
	FROM HUB_REGISTRY_SERVICE_GROUP WHERE %s`, where)

	group := &models.ServiceGroup{}
	err := dao.db.QueryOne(ctx, group, query, args, true)

	if err == sql.ErrNoRows {
		return nil, errors.New("服务分组不存在")
	}
	if err != nil {
		return nil, huberrors.WrapError(err, "查询服务分组失败")
	}

	return group, nil
}

// GetServiceGroupById 根据ID获取服务分组
//
// 参数：
//   - ctx: 上下文对象
//   - tenantId: 租户ID
//   - serviceGroupId: 分组ID
//
// 返回值：
//   - *models.ServiceGroup: 服务分组信息
//   - error: 错误信息
func (dao *ServiceGroupDAO) GetServiceGroupById(ctx context.Context, tenantId, serviceGroupId string) (*models.ServiceGroup, error) {
	query := `SELECT serviceGroupId, tenantId, groupName, groupDescription, groupType,
		ownerUserId, adminUserIds, readUserIds, accessControlEnabled,
		defaultProtocolType, defaultLoadBalanceStrategy, defaultHealthCheckUrl, defaultHealthCheckIntervalSeconds,
		addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty,
		reserved1, reserved2, reserved3, reserved4, reserved5, reserved6, reserved7, reserved8, reserved9, reserved10
	FROM HUB_REGISTRY_SERVICE_GROUP WHERE tenantId = ? AND serviceGroupId = ?`

	group := &models.ServiceGroup{}
	err := dao.db.QueryOne(ctx, group, query, []interface{}{tenantId, serviceGroupId}, true)

	if err == sql.ErrNoRows {
		return nil, errors.New("服务分组不存在")
	}
	if err != nil {
		return nil, huberrors.WrapError(err, "查询服务分组失败")
	}

	return group, nil
}

// DeleteServiceGroup 删除服务分组
//
// 参数：
//   - ctx: 上下文对象
//   - tenantId: 租户ID
//   - groupName: 分组名称
//
// 返回值：
//   - error: 错误信息
func (dao *ServiceGroupDAO) DeleteServiceGroup(ctx context.Context, tenantId, groupName string) error {
	// 先获取分组信息以获得主键ID
	group, err := dao.GetServiceGroup(ctx, tenantId, groupName, "")
	if err != nil {
		return huberrors.WrapError(err, "获取要删除的服务分组失败")
	}

	// 按主键删除
	where := "tenantId = ? AND serviceGroupId = ?"
	args := []interface{}{group.TenantId, group.ServiceGroupId}

	affectedRows, err := dao.db.Delete(ctx, "HUB_REGISTRY_SERVICE_GROUP", where, args, true)
	if err != nil {
		return huberrors.WrapError(err, "删除服务分组失败")
	}

	if affectedRows == 0 {
		return errors.New("未找到要删除的服务分组记录")
	}

	// 从注册中心删除服务组
	err = dao.deleteFromRegistryManager(ctx, group.TenantId, group.ServiceGroupId)
	if err != nil {
		logger.WarnWithTrace(ctx, "从注册中心删除服务分组失败",
			"serviceGroupId", group.ServiceGroupId,
			"groupName", group.GroupName,
			"error", err)
		// 注册中心删除失败不影响数据库操作的成功
	}

	return nil
}

// DeleteServiceGroupById 根据ID删除服务分组
//
// 参数：
//   - ctx: 上下文对象
//   - tenantId: 租户ID
//   - serviceGroupId: 分组ID
//
// 返回值：
//   - error: 错误信息
func (dao *ServiceGroupDAO) DeleteServiceGroupById(ctx context.Context, tenantId, serviceGroupId string) error {
	where := "tenantId = ? AND serviceGroupId = ?"
	args := []interface{}{tenantId, serviceGroupId}

	affectedRows, err := dao.db.Delete(ctx, "HUB_REGISTRY_SERVICE_GROUP", where, args, true)
	if err != nil {
		return huberrors.WrapError(err, "根据ID删除服务分组失败")
	}

	if affectedRows == 0 {
		return errors.New("未找到要删除的服务分组记录")
	}

	// 从注册中心删除服务组
	err = dao.deleteFromRegistryManager(ctx, tenantId, serviceGroupId)
	if err != nil {
		logger.WarnWithTrace(ctx, "从注册中心删除服务分组失败",
			"serviceGroupId", serviceGroupId,
			"tenantId", tenantId,
			"error", err)
		// 注册中心删除失败不影响数据库操作的成功
	}

	return nil
}

// ListServiceGroups 列出服务分组
//
// 参数：
//   - ctx: 上下文对象
//   - tenantId: 租户ID
//   - activeFlag: 活动状态标记(Y活动,N非活动,空为全部)
//
// 返回值：
//   - []*models.ServiceGroup: 服务分组列表
//   - error: 错误信息
func (dao *ServiceGroupDAO) ListServiceGroups(ctx context.Context, tenantId, activeFlag string) ([]*models.ServiceGroup, error) {
	// 构建查询条件
	where := "tenantId = ?"
	args := []interface{}{tenantId}

	if activeFlag != "" {
		where += " AND activeFlag = ?"
		args = append(args, activeFlag)
	}

	// 构建查询SQL，按分组名称排序
	query := fmt.Sprintf(`SELECT serviceGroupId, tenantId, groupName, groupDescription, groupType,
		ownerUserId, adminUserIds, readUserIds, accessControlEnabled,
		defaultProtocolType, defaultLoadBalanceStrategy, defaultHealthCheckUrl, defaultHealthCheckIntervalSeconds,
		addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty,
		reserved1, reserved2, reserved3, reserved4, reserved5, reserved6, reserved7, reserved8, reserved9, reserved10
	FROM HUB_REGISTRY_SERVICE_GROUP WHERE %s ORDER BY groupName`, where)

	var groups []*models.ServiceGroup
	err := dao.db.Query(ctx, &groups, query, args, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询服务分组列表失败")
	}

	return groups, nil
}

// QueryServiceGroups 查询服务分组（支持分页和字段过滤）
//
// 参数：
//   - ctx: 上下文对象
//   - tenantId: 租户ID
//   - activeFlag: 活动状态标记(Y活动,N非活动,空为全部)
//   - groupType: 分组类型过滤
//   - ownerUserId: 拥有者过滤
//   - page: 页码
//   - pageSize: 每页数量
//
// 返回值：
//   - []*models.ServiceGroup: 查询结果
//   - int: 总数量
//   - error: 错误信息
func (dao *ServiceGroupDAO) QueryServiceGroups(ctx context.Context, tenantId, activeFlag, groupType, ownerUserId string, page, pageSize int) ([]*models.ServiceGroup, int, error) {
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

	// 获取所有分组
	allGroups, err := dao.ListServiceGroups(ctx, tenantId, activeFlag)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询服务分组列表失败")
	}

	// 过滤
	var filteredGroups []*models.ServiceGroup
	for _, group := range allGroups {
		// 分组类型过滤
		if groupType != "" && group.GroupType != groupType {
			continue
		}

		// 拥有者过滤
		if ownerUserId != "" && group.OwnerUserId != ownerUserId {
			continue
		}

		filteredGroups = append(filteredGroups, group)
	}

	// 分页
	total := len(filteredGroups)
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= total {
		// 超出范围，返回空结果
		return []*models.ServiceGroup{}, total, nil
	}

	if end > total {
		end = total
	}

	pagedGroups := filteredGroups[start:end]

	return pagedGroups, total, nil
}

// syncToRegistryManager 同步服务组到注册中心管理器
// 将数据库中的服务组信息同步到注册中心的缓存中
func (dao *ServiceGroupDAO) syncToRegistryManager(ctx context.Context, group *models.ServiceGroup, operation string) error {
	// 获取注册中心管理器实例
	registryMgr := manager.GetInstance()
	if registryMgr == nil {
		return fmt.Errorf("注册中心管理器未初始化")
	}

	// 检查注册中心是否就绪
	if !registryMgr.IsReady() {
		return fmt.Errorf("注册中心管理器尚未就绪")
	}

	// 转换为注册中心的服务组结构
	coreGroup := dao.convertToCore(group)

	// 根据操作类型调用相应方法
	switch operation {
	case "CREATE", "UPDATE":
		_, err := registryMgr.SetServiceGroup(ctx, coreGroup)
		if err != nil {
			return fmt.Errorf("同步服务组到注册中心失败: %w", err)
		}
		logger.DebugWithTrace(ctx, "成功同步服务组到注册中心",
			"operation", operation,
			"serviceGroupId", group.ServiceGroupId,
			"groupName", group.GroupName)
	default:
		return fmt.Errorf("不支持的操作类型: %s", operation)
	}

	return nil
}

// deleteFromRegistryManager 从注册中心管理器中删除服务组
func (dao *ServiceGroupDAO) deleteFromRegistryManager(ctx context.Context, tenantId, serviceGroupId string) error {
	// 获取注册中心管理器实例
	registryMgr := manager.GetInstance()
	if registryMgr == nil {
		return fmt.Errorf("注册中心管理器未初始化")
	}

	// 检查注册中心是否就绪
	if !registryMgr.IsReady() {
		return fmt.Errorf("注册中心管理器尚未就绪")
	}

	// 调用注册中心管理器删除服务组
	err := registryMgr.DeleteServiceGroup(ctx, tenantId, serviceGroupId)
	if err != nil {
		return fmt.Errorf("从注册中心删除服务组失败: %w", err)
	}

	logger.DebugWithTrace(ctx, "成功从泣册中心删除服务组",
		"tenantId", tenantId,
		"serviceGroupId", serviceGroupId)

	return nil
}

// convertToCore 将models.ServiceGroup转换为core.ServiceGroup
func (dao *ServiceGroupDAO) convertToCore(group *models.ServiceGroup) *core.ServiceGroup {
	return &core.ServiceGroup{
		ServiceGroupId:                    group.ServiceGroupId,
		TenantId:                          group.TenantId,
		GroupName:                         group.GroupName,
		GroupDescription:                  group.GroupDescription,
		GroupType:                         group.GroupType,
		OwnerUserId:                       group.OwnerUserId,
		AdminUserIds:                      group.AdminUserIds,
		ReadUserIds:                       group.ReadUserIds,
		AccessControlEnabled:              group.AccessControlEnabled,
		DefaultProtocolType:               group.DefaultProtocolType,
		DefaultLoadBalanceStrategy:        group.DefaultLoadBalanceStrategy,
		DefaultHealthCheckUrl:             group.DefaultHealthCheckUrl,
		DefaultHealthCheckIntervalSeconds: group.DefaultHealthCheckIntervalSeconds,
		AddTime:                           group.AddTime,
		AddWho:                            group.AddWho,
		EditTime:                          group.EditTime,
		EditWho:                           group.EditWho,
		OprSeqFlag:                        group.OprSeqFlag,
		CurrentVersion:                    group.CurrentVersion,
		ActiveFlag:                        group.ActiveFlag,
		NoteText:                          group.NoteText,
		ExtProperty:                       group.ExtProperty,
		Reserved1:                         group.Reserved1,
		Reserved2:                         group.Reserved2,
		Reserved3:                         group.Reserved3,
		Reserved4:                         group.Reserved4,
		Reserved5:                         group.Reserved5,
		Reserved6:                         group.Reserved6,
		Reserved7:                         group.Reserved7,
		Reserved8:                         group.Reserved8,
		Reserved9:                         group.Reserved9,
		Reserved10:                        group.Reserved10,
	}
}
