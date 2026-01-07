package dao

import (
	"context"
	"errors"
	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
	"gateway/pkg/utils/empty"
	"gateway/pkg/utils/huberrors"
	"gateway/web/views/hub0022/models"
)

// GatewayInstanceDAO 网关实例数据访问对象
type GatewayInstanceDAO struct {
	db database.Database
}

// NewGatewayInstanceDAO 创建网关实例DAO
func NewGatewayInstanceDAO(db database.Database) *GatewayInstanceDAO {
	return &GatewayInstanceDAO{
		db: db,
	}
}

// ListAllGatewayInstances 获取所有网关实例列表（跨租户查询，仅限管理员使用）
func (dao *GatewayInstanceDAO) ListAllGatewayInstances(ctx context.Context, page, pageSize int) ([]*models.GatewayInstance, int, error) {
	// 构建基础查询语句
	baseQuery := "SELECT * FROM HUB_GW_INSTANCE ORDER BY addTime DESC"

	// 构建统计查询
	countQuery, err := sqlutils.BuildCountQuery(baseQuery)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建统计查询失败")
	}

	// 执行统计查询
	var result struct {
		Count int `db:"COUNT(*)"`
	}
	err = dao.db.QueryOne(ctx, &result, countQuery, []interface{}{}, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询网关实例总数失败")
	}
	total := result.Count

	// 如果没有记录，直接返回空列表
	if total == 0 {
		return []*models.GatewayInstance{}, 0, nil
	}

	// 创建分页信息
	pagination := sqlutils.NewPaginationInfo(page, pageSize)

	// 获取数据库类型
	dbType := sqlutils.GetDatabaseType(dao.db)

	// 构建分页查询
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建分页查询失败")
	}

	// 执行分页查询
	var instances []*models.GatewayInstance
	err = dao.db.Query(ctx, &instances, paginatedQuery, paginationArgs, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询网关实例列表失败")
	}

	return instances, total, nil
}

// GetGatewayInstanceById 根据ID获取网关实例
func (dao *GatewayInstanceDAO) GetGatewayInstanceById(ctx context.Context, gatewayInstanceId, tenantId string) (*models.GatewayInstance, error) {
	if gatewayInstanceId == "" {
		return nil, errors.New("gatewayInstanceId不能为空")
	}

	query := `
		SELECT * FROM HUB_GW_INSTANCE 
		WHERE gatewayInstanceId = ? AND tenantId = ?
	`

	var instance models.GatewayInstance
	err := dao.db.QueryOne(ctx, &instance, query, []interface{}{gatewayInstanceId, tenantId}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil // 没有找到记录，返回nil而不是错误
		}
		return nil, huberrors.WrapError(err, "查询网关实例失败")
	}

	return &instance, nil
}

// DeleteGatewayInstance 物理删除网关实例
func (dao *GatewayInstanceDAO) DeleteGatewayInstance(ctx context.Context, gatewayInstanceId, tenantId, operatorId string) error {
	if gatewayInstanceId == "" {
		return errors.New("gatewayInstanceId不能为空")
	}

	// 检查网关实例是否存在
	existing, err := dao.GetGatewayInstanceById(ctx, gatewayInstanceId, tenantId)
	if err != nil {
		return huberrors.WrapError(err, "获取网关实例失败")
	}
	if existing == nil {
		return errors.New("网关实例不存在")
	}

	// 执行物理删除
	sql := `DELETE FROM HUB_GW_INSTANCE WHERE gatewayInstanceId = ? AND tenantId = ?`

	result, err := dao.db.Exec(ctx, sql, []interface{}{gatewayInstanceId, tenantId}, true)
	if err != nil {
		return huberrors.WrapError(err, "删除网关实例失败")
	}

	// 检查是否有记录被删除
	if result == 0 {
		return errors.New("删除失败，网关实例不存在")
	}

	return nil
}

// QueryGatewayInstances 获取租户下的网关实例列表（支持分页和筛选）
func (dao *GatewayInstanceDAO) QueryGatewayInstances(ctx context.Context, tenantId string, page, pageSize int, filters map[string]interface{}) ([]*models.GatewayInstance, int, error) {
	// 构建基础查询条件
	whereClause := "WHERE tenantId = ?"
	params := []interface{}{tenantId}

	// 添加筛选条件
	if filters != nil {
		if instanceName, ok := filters["instanceName"].(string); ok && !empty.IsEmpty(instanceName) {
			whereClause += " AND instanceName LIKE ?"
			params = append(params, "%"+instanceName+"%")
		}
		if healthStatus, ok := filters["healthStatus"].(string); ok && !empty.IsEmpty(healthStatus) {
			whereClause += " AND healthStatus = ?"
			params = append(params, healthStatus)
		}
		if tlsEnabled, ok := filters["tlsEnabled"].(string); ok && !empty.IsEmpty(tlsEnabled) {
			whereClause += " AND tlsEnabled = ?"
			params = append(params, tlsEnabled)
		}
		// 添加activeFlag条件（只有当不为空时才添加）
		if activeFlag, ok := filters["activeFlag"].(string); ok && !empty.IsEmpty(activeFlag) {
			whereClause += " AND activeFlag = ?"
			params = append(params, activeFlag)
		}
	}

	// 构建基础查询语句
	baseQuery := "SELECT * FROM HUB_GW_INSTANCE " + whereClause + " ORDER BY addTime DESC"

	// 构建统计查询
	countQuery, err := sqlutils.BuildCountQuery(baseQuery)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建统计查询失败")
	}

	// 执行统计查询
	var result struct {
		Count int `db:"COUNT(*)"`
	}
	err = dao.db.QueryOne(ctx, &result, countQuery, params, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询网关实例总数失败")
	}
	total := result.Count

	// 如果没有记录，直接返回空列表
	if total == 0 {
		return []*models.GatewayInstance{}, 0, nil
	}

	// 创建分页信息
	pagination := sqlutils.NewPaginationInfo(page, pageSize)

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
	var instances []*models.GatewayInstance
	err = dao.db.Query(ctx, &instances, paginatedQuery, allArgs, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询网关实例列表失败")
	}

	return instances, total, nil
}
