package dao

import (
	"context"
	"fmt"
	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
	"gateway/pkg/utils/huberrors"
	"gateway/web/views/hub0021/models"
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
func (dao *GatewayInstanceDAO) ListAllGatewayInstances(ctx context.Context, activeFlag string, page, pageSize int) ([]*models.GatewayInstance, int, error) {
	// 构建查询条件
	whereClause := ""
	args := []interface{}{}

	// 添加activeFlag条件（如果指定了activeFlag参数）
	if activeFlag != "" {
		whereClause = "WHERE activeFlag = ?"
		args = append(args, activeFlag)
	}

	// 构建基础查询语句
	baseQuery := fmt.Sprintf("SELECT * FROM HUB_GW_INSTANCE %s ORDER BY addTime DESC", whereClause)

	// 构建统计查询
	countQuery, err := sqlutils.BuildCountQuery(baseQuery)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建统计查询失败")
	}

	// 执行统计查询
	var result struct {
		Count int `db:"COUNT(*)"`
	}
	err = dao.db.QueryOne(ctx, &result, countQuery, args, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询网关实例总数失败")
	}

	// 如果没有记录，直接返回空列表
	if result.Count == 0 {
		return []*models.GatewayInstance{}, 0, nil
	}

	// 创建分页信息
	paginationInfo := sqlutils.NewPaginationInfo(page, pageSize)

	// 获取数据库类型
	dbType := sqlutils.GetDatabaseType(dao.db)

	// 构建分页查询
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, paginationInfo)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建分页查询失败")
	}

	// 合并查询参数
	allArgs := append(args, paginationArgs...)

	// 执行分页查询
	var instances []*models.GatewayInstance
	err = dao.db.Query(ctx, &instances, paginatedQuery, allArgs, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询网关实例列表失败")
	}

	return instances, result.Count, nil
}
