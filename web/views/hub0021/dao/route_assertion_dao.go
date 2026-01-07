package dao

import (
	"context"
	"errors"
	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
	"gateway/pkg/utils/empty"
	"gateway/pkg/utils/huberrors"
	"gateway/pkg/utils/random"
	"gateway/web/views/hub0021/models"
	"time"
)

// RouteAssertionDAO 路由断言数据访问对象
type RouteAssertionDAO struct {
	db database.Database
}

// NewRouteAssertionDAO 创建路由断言DAO
func NewRouteAssertionDAO(db database.Database) *RouteAssertionDAO {
	return &RouteAssertionDAO{
		db: db,
	}
}

// AddRouteAssertion 添加路由断言
func (dao *RouteAssertionDAO) AddRouteAssertion(ctx context.Context, assertion *models.RouteAssertion, operatorId string) (string, error) {
	// 验证必填字段
	if assertion.RouteConfigId == "" {
		return "", errors.New("路由配置ID不能为空")
	}
	if assertion.AssertionName == "" {
		return "", errors.New("断言名称不能为空")
	}
	if assertion.AssertionType == "" {
		return "", errors.New("断言类型不能为空")
	}

	// 自动生成路由断言ID（如果为空）
	if assertion.RouteAssertionId == "" {
		// 使用公共方法生成32位唯一字符串，前缀为"RA"
		assertion.RouteAssertionId = random.GenerateUniqueStringWithPrefix("RA", 32)
	}

	// 设置一些自动填充的字段
	now := time.Now()
	assertion.AddTime = now
	assertion.AddWho = operatorId
	assertion.EditTime = now
	assertion.EditWho = operatorId
	// 生成 OprSeqFlag，确保长度不超过32
	// RouteAssertionId 已经是32位，直接使用
	assertion.OprSeqFlag = assertion.RouteAssertionId
	assertion.CurrentVersion = 1
	assertion.ActiveFlag = "Y"

	// 设置默认值
	if assertion.AssertionOperator == "" {
		assertion.AssertionOperator = "EQUAL"
	}
	if assertion.CaseSensitive == "" {
		assertion.CaseSensitive = "Y"
	}
	if assertion.IsRequired == "" {
		assertion.IsRequired = "Y"
	}

	// 使用数据库接口的Insert方法插入记录
	_, err := dao.db.Insert(ctx, "HUB_GW_ROUTE_ASSERTION", assertion, true)
	if err != nil {
		return "", huberrors.WrapError(err, "添加路由断言失败")
	}

	return assertion.RouteAssertionId, nil
}

// GetRouteAssertionsByRouteId 根据路由配置ID获取所有断言
func (dao *RouteAssertionDAO) GetRouteAssertionsByRouteId(ctx context.Context, routeConfigId, tenantId string) ([]*models.RouteAssertion, error) {
	if routeConfigId == "" {
		return nil, errors.New("routeConfigId不能为空")
	}

	query := `
		SELECT * FROM HUB_GW_ROUTE_ASSERTION 
		WHERE routeConfigId = ? AND tenantId = ? 
		ORDER BY assertionOrder ASC, addTime ASC
	`

	var assertions []*models.RouteAssertion
	err := dao.db.Query(ctx, &assertions, query, []interface{}{routeConfigId, tenantId}, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询路由断言失败")
	}

	return assertions, nil
}

// GetRouteAssertionById 根据ID获取路由断言
func (dao *RouteAssertionDAO) GetRouteAssertionById(ctx context.Context, routeAssertionId, tenantId string) (*models.RouteAssertion, error) {
	if routeAssertionId == "" {
		return nil, errors.New("routeAssertionId不能为空")
	}

	query := `
		SELECT * FROM HUB_GW_ROUTE_ASSERTION 
		WHERE routeAssertionId = ? AND tenantId = ?
	`

	var assertion models.RouteAssertion
	err := dao.db.QueryOne(ctx, &assertion, query, []interface{}{routeAssertionId, tenantId}, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询路由断言失败")
	}

	return &assertion, nil
}

// UpdateRouteAssertion 更新路由断言
func (dao *RouteAssertionDAO) UpdateRouteAssertion(ctx context.Context, assertion *models.RouteAssertion, operatorId string) error {
	if assertion.RouteAssertionId == "" {
		return errors.New("routeAssertionId不能为空")
	}

	// 验证必填字段
	if assertion.RouteConfigId == "" {
		return errors.New("路由配置ID不能为空")
	}
	if assertion.AssertionName == "" {
		return errors.New("断言名称不能为空")
	}
	if assertion.AssertionType == "" {
		return errors.New("断言类型不能为空")
	}

	// 首先获取当前版本信息
	currentAssertion, err := dao.GetRouteAssertionById(ctx, assertion.RouteAssertionId, assertion.TenantId)
	if err != nil {
		return huberrors.WrapError(err, "获取现有路由断言失败")
	}
	if currentAssertion == nil {
		return errors.New("路由断言不存在")
	}

	// 保留不可修改的字段
	assertion.TenantId = currentAssertion.TenantId
	assertion.RouteAssertionId = currentAssertion.RouteAssertionId
	assertion.AddTime = currentAssertion.AddTime
	assertion.AddWho = currentAssertion.AddWho
	assertion.OprSeqFlag = currentAssertion.OprSeqFlag
	assertion.CurrentVersion = currentAssertion.CurrentVersion + 1

	// 更新修改信息
	assertion.EditTime = time.Now()
	assertion.EditWho = operatorId

	// 设置默认值
	if assertion.AssertionOperator == "" {
		assertion.AssertionOperator = "EQUAL"
	}
	if assertion.CaseSensitive == "" {
		assertion.CaseSensitive = "Y"
	}
	if assertion.IsRequired == "" {
		assertion.IsRequired = "Y"
	}

	// 构建更新SQL
	sql := `
		UPDATE HUB_GW_ROUTE_ASSERTION SET
			routeConfigId = ?, assertionName = ?, assertionType = ?, assertionOperator = ?,
			fieldName = ?, expectedValue = ?, patternValue = ?, caseSensitive = ?,
			assertionOrder = ?, isRequired = ?, assertionDesc = ?, reserved1 = ?,
			reserved2 = ?, reserved3 = ?, reserved4 = ?, reserved5 = ?,
			extProperty = ?, noteText = ?, editTime = ?, editWho = ?, currentVersion = ?,
			activeFlag = ?
		WHERE routeAssertionId = ? AND tenantId = ? AND currentVersion = ?
	`

	// 执行更新
	result, err := dao.db.Exec(ctx, sql, []interface{}{
		assertion.RouteConfigId, assertion.AssertionName, assertion.AssertionType, assertion.AssertionOperator,
		assertion.FieldName, assertion.ExpectedValue, assertion.PatternValue, assertion.CaseSensitive,
		assertion.AssertionOrder, assertion.IsRequired, assertion.AssertionDesc, assertion.Reserved1,
		assertion.Reserved2, assertion.Reserved3, assertion.Reserved4, assertion.Reserved5,
		assertion.ExtProperty, assertion.NoteText, assertion.EditTime, assertion.EditWho, assertion.CurrentVersion,
		assertion.ActiveFlag,
		assertion.RouteAssertionId, assertion.TenantId, currentAssertion.CurrentVersion,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "更新路由断言失败")
	}

	// 检查是否有记录被更新
	if result == 0 {
		return errors.New("路由断言更新失败，可能是版本冲突或记录不存在")
	}

	return nil
}

// DeleteRouteAssertion 删除路由断言
func (dao *RouteAssertionDAO) DeleteRouteAssertion(ctx context.Context, routeAssertionId, tenantId, operatorId string) error {
	if routeAssertionId == "" {
		return errors.New("routeAssertionId不能为空")
	}

	// 执行实际删除
	sql := `DELETE FROM HUB_GW_ROUTE_ASSERTION WHERE routeAssertionId = ? AND tenantId = ?`

	result, err := dao.db.Exec(ctx, sql, []interface{}{routeAssertionId, tenantId}, true)
	if err != nil {
		return huberrors.WrapError(err, "删除路由断言失败")
	}

	// 检查是否有记录被删除
	if result == 0 {
		return errors.New("路由断言不存在或已被删除")
	}

	return nil
}

// QueryRouteAssertions 分页查询路由断言列表（支持多条件筛选）
func (dao *RouteAssertionDAO) QueryRouteAssertions(ctx context.Context, page, pageSize int, filters map[string]interface{}, tenantId string) ([]*models.RouteAssertion, int, error) {
	// 构建基础查询条件
	whereClause := "WHERE tenantId = ?"
	params := []interface{}{tenantId}

	// 添加筛选条件
	if filters != nil {
		if routeConfigId, ok := filters["routeConfigId"].(string); ok && !empty.IsEmpty(routeConfigId) {
			whereClause += " AND routeConfigId = ?"
			params = append(params, routeConfigId)
		}
		if assertionName, ok := filters["assertionName"].(string); ok && !empty.IsEmpty(assertionName) {
			whereClause += " AND assertionName LIKE ?"
			params = append(params, "%"+assertionName+"%")
		}
		if assertionType, ok := filters["assertionType"].(string); ok && !empty.IsEmpty(assertionType) {
			whereClause += " AND assertionType = ?"
			params = append(params, assertionType)
		}
		// 添加activeFlag条件（只有当不为空时才添加）
		if activeFlag, ok := filters["activeFlag"].(string); ok && !empty.IsEmpty(activeFlag) {
			whereClause += " AND activeFlag = ?"
			params = append(params, activeFlag)
		}
	}

	// 构建基础查询语句
	baseQuery := "SELECT * FROM HUB_GW_ROUTE_ASSERTION " + whereClause + " ORDER BY assertionOrder ASC, addTime ASC"

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
		return nil, 0, huberrors.WrapError(err, "查询路由断言总数失败")
	}
	total := result.Count

	// 如果没有记录，直接返回空列表
	if total == 0 {
		return []*models.RouteAssertion{}, 0, nil
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
	var assertions []*models.RouteAssertion
	err = dao.db.Query(ctx, &assertions, paginatedQuery, allArgs, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询路由断言列表失败")
	}

	return assertions, total, nil
}
