package dao

import (
	"context"
	"gohub/pkg/database"
	"gohub/pkg/utils/huberrors"
	"gohub/web/views/hub0022/models"
	"errors"
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
	// 确保分页参数有效
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	// 计算偏移量
	offset := (page - 1) * pageSize

	// 查询总数
	countQuery := `SELECT COUNT(*) FROM HUB_GATEWAY_INSTANCE WHERE activeFlag = 'Y'`
	var result struct {
		Count int `db:"COUNT(*)"`
	}
	err := dao.db.QueryOne(ctx, &result, countQuery, []interface{}{}, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询网关实例总数失败")
	}
	total := result.Count

	// 如果没有记录，直接返回空列表
	if total == 0 {
		return []*models.GatewayInstance{}, 0, nil
	}

	// 查询数据
	dataQuery := `
		SELECT * FROM HUB_GATEWAY_INSTANCE 
		WHERE activeFlag = 'Y'
		ORDER BY addTime DESC
		LIMIT ? OFFSET ?
	`

	var instances []*models.GatewayInstance
	err = dao.db.Query(ctx, &instances, dataQuery, []interface{}{pageSize, offset}, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询网关实例列表失败")
	}

	return instances, total, nil
}

// GetGatewayInstanceById 根据ID获取网关实例
func (dao *GatewayInstanceDAO) GetGatewayInstanceById(ctx context.Context, gatewayInstanceId, tenantId string) (*models.GatewayInstance, error) {
	if gatewayInstanceId == "" || tenantId == "" {
		return nil, errors.New("gatewayInstanceId和tenantId不能为空")
	}

	query := `
		SELECT * FROM HUB_GATEWAY_INSTANCE 
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
	if gatewayInstanceId == "" || tenantId == "" {
		return errors.New("gatewayInstanceId和tenantId不能为空")
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
	sql := `DELETE FROM HUB_GATEWAY_INSTANCE WHERE gatewayInstanceId = ? AND tenantId = ?`
	
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
	if tenantId == "" {
		return nil, 0, errors.New("tenantId不能为空")
	}

	// 确保分页参数有效
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	// 计算偏移量
	offset := (page - 1) * pageSize

	// 构建基础查询条件
	whereClause := "WHERE tenantId = ?"
	params := []interface{}{tenantId}

	// 添加筛选条件
	if filters != nil {
		if instanceName, ok := filters["instanceName"].(string); ok && instanceName != "" {
			whereClause += " AND instanceName LIKE ?"
			params = append(params, "%"+instanceName+"%")
		}
		if healthStatus, ok := filters["healthStatus"].(string); ok && healthStatus != "" {
			whereClause += " AND healthStatus = ?"
			params = append(params, healthStatus)
		}
		if tlsEnabled, ok := filters["tlsEnabled"].(string); ok && tlsEnabled != "" {
			whereClause += " AND tlsEnabled = ?"
			params = append(params, tlsEnabled)
		}
	}

	// 查询总数
	countQuery := `SELECT COUNT(*) FROM HUB_GATEWAY_INSTANCE ` + whereClause
	var result struct {
		Count int `db:"COUNT(*)"`
	}
	err := dao.db.QueryOne(ctx, &result, countQuery, params, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询网关实例总数失败")
	}
	total := result.Count

	// 如果没有记录，直接返回空列表
	if total == 0 {
		return []*models.GatewayInstance{}, 0, nil
	}

	// 查询数据
	dataQuery := `
		SELECT * FROM HUB_GATEWAY_INSTANCE 
		` + whereClause + `
		ORDER BY addTime DESC
		LIMIT ? OFFSET ?
	`

	// 添加分页参数
	params = append(params, pageSize, offset)

	var instances []*models.GatewayInstance
	err = dao.db.Query(ctx, &instances, dataQuery, params, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询网关实例列表失败")
	}

	return instances, total, nil
} 