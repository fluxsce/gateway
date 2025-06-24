package dao

import (
	"context"
	"gohub/pkg/database"
	"gohub/pkg/utils/huberrors"
	"gohub/web/views/hub0021/models"
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