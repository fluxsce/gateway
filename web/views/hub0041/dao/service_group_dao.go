package dao

import (
	"context"
	"strings"

	"gateway/pkg/database"
	"gateway/pkg/utils/huberrors"
	"gateway/web/views/hub0041/models"
)

// ServiceGroupDAO 服务分组数据访问对象
// 负责服务分组（命名空间）信息的数据库操作
type ServiceGroupDAO struct {
	db database.Database
}

// NewServiceGroupDAO 创建服务分组DAO实例
func NewServiceGroupDAO(db database.Database) *ServiceGroupDAO {
	return &ServiceGroupDAO{
		db: db,
	}
}

// GetServiceGroups 获取租户下的所有服务分组列表
//
// 参数：
//   - ctx: 上下文对象
//   - tenantId: 租户ID
//   - activeFlag: 活动状态标记(Y活动,N非活动,空为全部)
//
// 返回值：
//   - []*models.ServiceGroup: 服务分组列表
//   - error: 错误信息
func (dao *ServiceGroupDAO) GetServiceGroups(ctx context.Context, tenantId, activeFlag string) ([]*models.ServiceGroup, error) {
	whereConditions := []string{"tenantId = ?"}
	args := []interface{}{tenantId}

	if activeFlag != "" {
		whereConditions = append(whereConditions, "activeFlag = ?")
		args = append(args, activeFlag)
	}

	whereClause := strings.Join(whereConditions, " AND ")

	// 构建查询SQL - 只查询必要字段用于列表显示
	query := `SELECT serviceGroupId, tenantId, groupName, groupDescription, groupType,
		ownerUserId, accessControlEnabled,
		defaultProtocolType, defaultLoadBalanceStrategy,
		addTime, addWho, editTime, editWho, activeFlag, noteText
	FROM HUB_REGISTRY_SERVICE_GROUP WHERE ` + whereClause + ` ORDER BY addTime DESC`

	var serviceGroups []*models.ServiceGroup
	err := dao.db.Query(ctx, &serviceGroups, query, args, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询服务分组列表失败")
	}

	return serviceGroups, nil
}

// GetServiceGroup 根据服务分组ID获取详情
//
// 参数：
//   - ctx: 上下文对象
//   - tenantId: 租户ID
//   - serviceGroupId: 服务分组ID
//   - activeFlag: 活动状态标记(为空则不过滤)
//
// 返回值：
//   - *models.ServiceGroup: 服务分组信息
//   - error: 错误信息
func (dao *ServiceGroupDAO) GetServiceGroup(ctx context.Context, tenantId, serviceGroupId, activeFlag string) (*models.ServiceGroup, error) {
	whereConditions := []string{"tenantId = ?", "serviceGroupId = ?"}
	args := []interface{}{tenantId, serviceGroupId}

	if activeFlag != "" {
		whereConditions = append(whereConditions, "activeFlag = ?")
		args = append(args, activeFlag)
	}

	whereClause := strings.Join(whereConditions, " AND ")

	query := `SELECT serviceGroupId, tenantId, groupName, groupDescription, groupType,
		ownerUserId, adminUserIds, readUserIds, accessControlEnabled,
		defaultProtocolType, defaultLoadBalanceStrategy, defaultHealthCheckUrl, defaultHealthCheckIntervalSeconds,
		addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty,
		reserved1, reserved2, reserved3, reserved4, reserved5, reserved6, reserved7, reserved8, reserved9, reserved10
	FROM HUB_REGISTRY_SERVICE_GROUP WHERE ` + whereClause

	serviceGroup := &models.ServiceGroup{}
	err := dao.db.QueryOne(ctx, serviceGroup, query, args, true)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") || strings.Contains(err.Error(), "not found") {
			return nil, huberrors.WrapError(err, "服务分组不存在")
		}
		return nil, huberrors.WrapError(err, "获取服务分组信息失败")
	}

	return serviceGroup, nil
}

// GetServiceGroupTypes 获取支持的服务分组类型列表
func (dao *ServiceGroupDAO) GetServiceGroupTypes() []string {
	return []string{
		"BUSINESS", // 业务服务
		"SYSTEM",   // 系统服务
		"TEST",     // 测试服务
	}
}
