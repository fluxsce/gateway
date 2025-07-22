package dao

import (
	"context"
	"errors"
	"fmt"
	"gohub/pkg/database"
	"gohub/pkg/database/sqlutils"
	"gohub/pkg/utils/huberrors"
	"gohub/pkg/utils/random"
	"gohub/web/views/hub0021/models"
	"strings"
	"time"
)

// RouteConfigQueryParams 路由配置查询参数
type RouteConfigQueryParams struct {
	TenantId          string // 租户ID
	GatewayInstanceId string // 网关实例ID
	RouteName         string // 路由名称(支持模糊匹配)
	RoutePath         string // 路由路径(支持模糊匹配)
	MatchType         int    // 匹配类型(0:精确匹配,1:前缀匹配,2:正则匹配)
	ActiveFlag        string // 激活状态(Y:激活,N:未激活)
	Page              int    // 页码
	PageSize          int    // 每页数量
}

// RouteConfigDAO 路由配置数据访问对象
type RouteConfigDAO struct {
	db database.Database
}

// NewRouteConfigDAO 创建路由配置DAO
func NewRouteConfigDAO(db database.Database) *RouteConfigDAO {
	return &RouteConfigDAO{
		db: db,
	}
}

// generateRouteConfigId 生成路由配置ID
// 格式：RT + YYYYMMDD + HHMMSS + 4位随机数
// 示例：RT20240615143022A1B2
func (dao *RouteConfigDAO) generateRouteConfigId() string {
	now := time.Now()
	// 生成时间部分：YYYYMMDDHHMMSS
	timeStr := now.Format("20060102150405")
	
	// 生成4位随机字符（大写字母和数字）
	randomStr := random.GenerateRandomString(4)
	
	return fmt.Sprintf("RT%s%s", timeStr, randomStr)
}

// isRouteConfigIdExists 检查路由配置ID是否已存在
func (dao *RouteConfigDAO) isRouteConfigIdExists(ctx context.Context, routeConfigId string) (bool, error) {
	query := `SELECT COUNT(*) as count FROM HUB_GW_ROUTE_CONFIG WHERE routeConfigId = ?`
	
	var result struct {
		Count int `db:"count"`
	}
	
	err := dao.db.QueryOne(ctx, &result, query, []interface{}{routeConfigId}, true)
	if err != nil {
		return false, err
	}
	
	return result.Count > 0, nil
}

// generateUniqueRouteConfigId 生成唯一的路由配置ID
// 如果生成的ID已存在，会重新生成直到找到唯一的ID（最多尝试10次）
func (dao *RouteConfigDAO) generateUniqueRouteConfigId(ctx context.Context) (string, error) {
	const maxAttempts = 10
	
	for attempt := 0; attempt < maxAttempts; attempt++ {
		routeConfigId := dao.generateRouteConfigId()
		
		exists, err := dao.isRouteConfigIdExists(ctx, routeConfigId)
		if err != nil {
			return "", huberrors.WrapError(err, "检查路由配置ID是否存在失败")
		}
		
		if !exists {
			return routeConfigId, nil
		}
		
		// 如果ID已存在，等待1毫秒后重试（确保时间戳不同）
		time.Sleep(time.Millisecond)
	}
	
	return "", errors.New("生成唯一路由配置ID失败，已达到最大尝试次数")
}

// AddRouteConfig 添加路由配置
// 参数:
//   - ctx: 上下文对象
//   - routeConfig: 路由配置信息
//   - operatorId: 操作人ID
//
// 返回:
//   - routeConfigId: 新创建的路由配置ID
//   - err: 可能的错误
func (dao *RouteConfigDAO) AddRouteConfig(ctx context.Context, routeConfig *models.RouteConfig, operatorId string) (string, error) {
	// 验证租户ID
	if routeConfig.TenantId == "" {
		return "", errors.New("租户ID不能为空")
	}

	// 验证必填字段
	if routeConfig.GatewayInstanceId == "" {
		return "", errors.New("网关实例ID不能为空")
	}
	if routeConfig.RouteName == "" {
		return "", errors.New("路由名称不能为空")
	}
	if routeConfig.RoutePath == "" {
		return "", errors.New("路由路径不能为空")
	}

	// 自动生成路由配置ID（如果为空）
	if routeConfig.RouteConfigId == "" {
		generatedId, err := dao.generateUniqueRouteConfigId(ctx)
		if err != nil {
			return "", huberrors.WrapError(err, "生成路由配置ID失败")
		}
		routeConfig.RouteConfigId = generatedId
	} else {
		// 如果提供了ID，检查是否已存在
		exists, err := dao.isRouteConfigIdExists(ctx, routeConfig.RouteConfigId)
		if err != nil {
			return "", huberrors.WrapError(err, "检查路由配置ID是否存在失败")
		}
		if exists {
			return "", errors.New("路由配置ID已存在")
		}
	}

	// 设置一些自动填充的字段
	now := time.Now()
	routeConfig.AddTime = now
	routeConfig.AddWho = operatorId
	routeConfig.EditTime = now
	routeConfig.EditWho = operatorId
	routeConfig.OprSeqFlag = routeConfig.RouteConfigId + "_" + strings.ReplaceAll(time.Now().String(), ".", "")[:8]
	routeConfig.CurrentVersion = 1
	routeConfig.ActiveFlag = "Y"

	// 设置默认值
	if routeConfig.MatchType < 0 || routeConfig.MatchType > 2 {
		routeConfig.MatchType = 1 // 默认前缀匹配
	}
	if routeConfig.RoutePriority == 0 {
		routeConfig.RoutePriority = 100 // 默认优先级
	}
	if routeConfig.StripPathPrefix == "" {
		routeConfig.StripPathPrefix = "N"
	}
	if routeConfig.EnableWebsocket == "" {
		routeConfig.EnableWebsocket = "N"
	}
	if routeConfig.TimeoutMs == 0 {
		routeConfig.TimeoutMs = 30000 // 默认30秒超时
	}
	if routeConfig.RetryCount == 0 {
		routeConfig.RetryCount = 0 // 默认不重试
	}
	if routeConfig.RetryIntervalMs == 0 {
		routeConfig.RetryIntervalMs = 1000 // 默认1秒重试间隔
	}

	// 使用数据库接口的Insert方法插入记录
	_, err := dao.db.Insert(ctx, "HUB_GW_ROUTE_CONFIG", routeConfig, true)

	if err != nil {
		// 检查是否是路由名重复错误
		if dao.isDuplicateRouteNameError(err) {
			return "", huberrors.WrapError(err, "路由名已存在")
		}
		return "", huberrors.WrapError(err, "添加路由配置失败")
	}

	return routeConfig.RouteConfigId, nil
}

// GetRouteConfigById 根据路由配置ID获取路由配置信息
func (dao *RouteConfigDAO) GetRouteConfigById(ctx context.Context, routeConfigId, tenantId string) (*models.RouteConfig, error) {
	if routeConfigId == "" || tenantId == "" {
		return nil, errors.New("routeConfigId和tenantId不能为空")
	}

	query := `
		SELECT * FROM HUB_GW_ROUTE_CONFIG 
		WHERE routeConfigId = ? AND tenantId = ?
	`

	var routeConfig models.RouteConfig
	err := dao.db.QueryOne(ctx, &routeConfig, query, []interface{}{routeConfigId, tenantId}, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询路由配置失败")
	}

	return &routeConfig, nil
}

// UpdateRouteConfig 更新路由配置
func (dao *RouteConfigDAO) UpdateRouteConfig(ctx context.Context, routeConfig *models.RouteConfig, operatorId string) error {
	if routeConfig.RouteConfigId == "" || routeConfig.TenantId == "" {
		return errors.New("routeConfigId和tenantId不能为空")
	}

	// 验证必填字段
	if routeConfig.GatewayInstanceId == "" {
		return errors.New("网关实例ID不能为空")
	}
	if routeConfig.RouteName == "" {
		return errors.New("路由名称不能为空")
	}
	if routeConfig.RoutePath == "" {
		return errors.New("路由路径不能为空")
	}

	// 检查路由配置是否存在
	existing, err := dao.GetRouteConfigById(ctx, routeConfig.RouteConfigId, routeConfig.TenantId)
	if err != nil {
		return huberrors.WrapError(err, "获取现有路由配置失败")
	}
	if existing == nil {
		return errors.New("路由配置不存在")
	}

	// 保留不可修改的字段
	routeConfig.TenantId = existing.TenantId
	routeConfig.RouteConfigId = existing.RouteConfigId
	routeConfig.AddTime = existing.AddTime
	routeConfig.AddWho = existing.AddWho
	routeConfig.OprSeqFlag = existing.OprSeqFlag
	routeConfig.CurrentVersion = existing.CurrentVersion + 1

	// 更新修改信息
	routeConfig.EditTime = time.Now()
	routeConfig.EditWho = operatorId

	// 构建更新SQL
	sql := `
		UPDATE HUB_GW_ROUTE_CONFIG SET
			gatewayInstanceId = ?, routeName = ?, routePath = ?, allowedMethods = ?, allowedHosts = ?,
			matchType = ?, routePriority = ?, stripPathPrefix = ?, rewritePath = ?,
			enableWebsocket = ?, timeoutMs = ?, retryCount = ?, retryIntervalMs = ?,
			serviceDefinitionId = ?, logConfigId = ?, routeMetadata = ?, reserved1 = ?, reserved2 = ?,
			reserved3 = ?, reserved4 = ?, reserved5 = ?, extProperty = ?, noteText = ?,
			editTime = ?, editWho = ?, currentVersion = ?
		WHERE routeConfigId = ? AND tenantId = ? AND currentVersion = ?
	`

	// 执行更新
	result, err := dao.db.Exec(ctx, sql, []interface{}{
		routeConfig.GatewayInstanceId, routeConfig.RouteName, routeConfig.RoutePath, 
		routeConfig.AllowedMethods, routeConfig.AllowedHosts,
		routeConfig.MatchType, routeConfig.RoutePriority, routeConfig.StripPathPrefix, routeConfig.RewritePath,
		routeConfig.EnableWebsocket, routeConfig.TimeoutMs, routeConfig.RetryCount, routeConfig.RetryIntervalMs,
		routeConfig.ServiceDefinitionId, routeConfig.LogConfigId, routeConfig.RouteMetadata, 
		routeConfig.Reserved1, routeConfig.Reserved2, routeConfig.Reserved3, routeConfig.Reserved4, routeConfig.Reserved5,
		routeConfig.ExtProperty, routeConfig.NoteText,
		routeConfig.EditTime, routeConfig.EditWho, routeConfig.CurrentVersion,
		routeConfig.RouteConfigId, routeConfig.TenantId, existing.CurrentVersion,
	}, true)

	if err != nil {
		// 检查是否是路由名重复错误
		if dao.isDuplicateRouteNameError(err) {
			return huberrors.WrapError(err, "路由名已存在")
		}
		return huberrors.WrapError(err, "更新路由配置失败")
	}

	// 检查是否有记录被更新
	if result == 0 {
		return errors.New("路由配置数据已被其他用户修改，请刷新后重试")
	}

	return nil
}

// DeleteRouteConfig 删除路由配置
func (dao *RouteConfigDAO) DeleteRouteConfig(ctx context.Context, routeConfigId, tenantId, operatorId string) error {
	if routeConfigId == "" || tenantId == "" {
		return errors.New("routeConfigId和tenantId不能为空")
	}

	// 检查路由配置是否存在
	existing, err := dao.GetRouteConfigById(ctx, routeConfigId, tenantId)
	if err != nil {
		return huberrors.WrapError(err, "获取路由配置失败")
	}
	if existing == nil {
		return errors.New("路由配置不存在")
	}

	// 执行实际删除
	sql := `DELETE FROM HUB_GW_ROUTE_CONFIG WHERE routeConfigId = ? AND tenantId = ?`

	result, err := dao.db.Exec(ctx, sql, []interface{}{routeConfigId, tenantId}, true)
	if err != nil {
		return huberrors.WrapError(err, "删除路由配置失败")
	}

	// 检查是否有记录被删除
	if result == 0 {
		return errors.New("路由配置不存在或已被删除")
	}

	return nil
}

// ListRouteConfigs 获取路由配置列表（分页，关联服务定义）
func (dao *RouteConfigDAO) ListRouteConfigs(ctx context.Context, params *RouteConfigQueryParams) ([]*models.RouteConfigWithService, int, error) {
	if params.TenantId == "" {
		return nil, 0, errors.New("tenantId不能为空")
	}

	// 构建查询条件
	whereClause := "WHERE rc.tenantId = ?"
	args := []interface{}{params.TenantId}

	// 激活状态过滤：如果指定了activeFlag参数则使用，否则默认只显示激活的记录
	if params.ActiveFlag != "" {
		whereClause += " AND rc.activeFlag = ?"
		args = append(args, params.ActiveFlag)
	} else {
		whereClause += " AND rc.activeFlag = 'Y'"
	}

	if params.GatewayInstanceId != "" {
		whereClause += " AND rc.gatewayInstanceId = ?"
		args = append(args, params.GatewayInstanceId)
	}
	if params.RouteName != "" {
		whereClause += " AND rc.routeName LIKE ?"
		args = append(args, "%"+params.RouteName+"%")
	}
	if params.RoutePath != "" {
		whereClause += " AND rc.routePath LIKE ?"
		args = append(args, "%"+params.RoutePath+"%")
	}
	if params.MatchType > 0 {
		whereClause += " AND rc.matchType = ?"
		args = append(args, params.MatchType)
	}

	// 构建基础查询语句（关联服务定义表）
	baseQuery := fmt.Sprintf(`
		SELECT 
			rc.tenantId,
			rc.routeConfigId,
			rc.gatewayInstanceId,
			rc.routeName,
			rc.routePath,
			rc.allowedMethods,
			rc.allowedHosts,
			rc.matchType,
			rc.routePriority,
			rc.stripPathPrefix,
			rc.rewritePath,
			rc.enableWebsocket,
			rc.timeoutMs,
			rc.retryCount,
			rc.retryIntervalMs,
			rc.serviceDefinitionId,
			rc.logConfigId,
			rc.routeMetadata,
			rc.reserved1,
			rc.reserved2,
			rc.reserved3,
			rc.reserved4,
			rc.reserved5,
			rc.extProperty,
			rc.addTime,
			rc.addWho,
			rc.editTime,
			rc.editWho,
			rc.oprSeqFlag,
			rc.currentVersion,
			rc.activeFlag,
			rc.noteText,
			sd.serviceName,
			sd.serviceDesc,
			sd.serviceType,
			sd.loadBalanceStrategy
		FROM HUB_GW_ROUTE_CONFIG rc
		LEFT JOIN HUB_GW_SERVICE_DEFINITION sd ON rc.tenantId = sd.tenantId AND rc.serviceDefinitionId = sd.serviceDefinitionId AND sd.activeFlag = 'Y'
		%s 
		ORDER BY rc.routePriority ASC, rc.addTime DESC
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
	err = dao.db.QueryOne(ctx, &countResult, countQuery, args, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询路由配置总数失败")
	}

	// 如果没有数据，直接返回
	if countResult.Count == 0 {
		return []*models.RouteConfigWithService{}, 0, nil
	}

	// 创建分页信息
	paginationInfo := sqlutils.NewPaginationInfo(params.Page, params.PageSize)

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
	var routeConfigs []*models.RouteConfigWithService
	err = dao.db.Query(ctx, &routeConfigs, paginatedQuery, allArgs, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询路由配置列表失败")
	}

	return routeConfigs, countResult.Count, nil
}

// FindRouteConfigByName 根据路由名称查找路由配置
func (dao *RouteConfigDAO) FindRouteConfigByName(ctx context.Context, routeName, tenantId, gatewayInstanceId string) (*models.RouteConfig, error) {
	if routeName == "" || tenantId == "" {
		return nil, errors.New("routeName和tenantId不能为空")
	}

	whereClause := "WHERE routeName = ? AND tenantId = ? AND activeFlag = 'Y'"
	args := []interface{}{routeName, tenantId}

	if gatewayInstanceId != "" {
		whereClause += " AND gatewayInstanceId = ?"
		args = append(args, gatewayInstanceId)
	}

	query := fmt.Sprintf("SELECT * FROM HUB_GW_ROUTE_CONFIG %s", whereClause)

	var routeConfig models.RouteConfig
	err := dao.db.QueryOne(ctx, &routeConfig, query, args, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询路由配置失败")
	}

	return &routeConfig, nil
}

// GetRouteConfigsByGatewayInstance 根据网关实例ID获取所有路由配置（关联服务定义）
func (dao *RouteConfigDAO) GetRouteConfigsByGatewayInstance(ctx context.Context, gatewayInstanceId, tenantId string) ([]*models.RouteConfigWithService, error) {
	if gatewayInstanceId == "" || tenantId == "" {
		return nil, errors.New("gatewayInstanceId和tenantId不能为空")
	}

	query := `
		SELECT 
			rc.tenantId,
			rc.routeConfigId,
			rc.gatewayInstanceId,
			rc.routeName,
			rc.routePath,
			rc.allowedMethods,
			rc.allowedHosts,
			rc.matchType,
			rc.routePriority,
			rc.stripPathPrefix,
			rc.rewritePath,
			rc.enableWebsocket,
			rc.timeoutMs,
			rc.retryCount,
			rc.retryIntervalMs,
			rc.serviceDefinitionId,
			rc.logConfigId,
			rc.routeMetadata,
			rc.reserved1,
			rc.reserved2,
			rc.reserved3,
			rc.reserved4,
			rc.reserved5,
			rc.extProperty,
			rc.addTime,
			rc.addWho,
			rc.editTime,
			rc.editWho,
			rc.oprSeqFlag,
			rc.currentVersion,
			rc.activeFlag,
			rc.noteText,
			sd.serviceName,
			sd.serviceDesc,
			sd.serviceType,
			sd.loadBalanceStrategy
		FROM HUB_GW_ROUTE_CONFIG rc
		LEFT JOIN HUB_GW_SERVICE_DEFINITION sd ON rc.tenantId = sd.tenantId AND rc.serviceDefinitionId = sd.serviceDefinitionId AND sd.activeFlag = 'Y'
		WHERE rc.gatewayInstanceId = ? AND rc.tenantId = ? AND rc.activeFlag = 'Y'
		ORDER BY rc.routePriority ASC, rc.addTime DESC
	`

	var routeConfigs []*models.RouteConfigWithService
	err := dao.db.Query(ctx, &routeConfigs, query, []interface{}{gatewayInstanceId, tenantId}, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询网关实例路由配置失败")
	}

	return routeConfigs, nil
}

// isDuplicateRouteNameError 检查是否是路由名重复错误
func (dao *RouteConfigDAO) isDuplicateRouteNameError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "duplicate") && strings.Contains(errStr, "routename")
}

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

// generateRouteAssertionId 生成路由断言ID
// 格式：RA + YYYYMMDD + HHMMSS + 4位随机数
// 示例：RA20240615143022A1B2
func (dao *RouteAssertionDAO) generateRouteAssertionId() string {
	now := time.Now()
	// 生成时间部分：YYYYMMDDHHMMSS
	timeStr := now.Format("20060102150405")
	
	// 生成4位随机字符（大写字母和数字）
	randomStr := random.GenerateRandomString(4)
	
	return fmt.Sprintf("RA%s%s", timeStr, randomStr)
}

// AddRouteAssertion 添加路由断言
func (dao *RouteAssertionDAO) AddRouteAssertion(ctx context.Context, assertion *models.RouteAssertion, operatorId string) (string, error) {
	// 验证必填字段
	if assertion.TenantId == "" {
		return "", errors.New("租户ID不能为空")
	}
	if assertion.RouteConfigId == "" {
		return "", errors.New("路由配置ID不能为空")
	}
	if assertion.AssertionName == "" {
		return "", errors.New("断言名称不能为空")
	}
	if assertion.AssertionType == "" {
		return "", errors.New("断言类型不能为空")
	}

	// 自动生成路由断言ID
	if assertion.RouteAssertionId == "" {
		assertion.RouteAssertionId = dao.generateRouteAssertionId()
	}

	// 设置一些自动填充的字段
	now := time.Now()
	assertion.AddTime = now
	assertion.AddWho = operatorId
	assertion.EditTime = now
	assertion.EditWho = operatorId
	assertion.OprSeqFlag = assertion.RouteAssertionId + "_" + strings.ReplaceAll(time.Now().String(), ".", "")[:8]
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
	if routeConfigId == "" || tenantId == "" {
		return nil, errors.New("routeConfigId和tenantId不能为空")
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

// DeleteRouteAssertion 删除路由断言
func (dao *RouteAssertionDAO) DeleteRouteAssertion(ctx context.Context, routeAssertionId, tenantId, operatorId string) error {
	if routeAssertionId == "" || tenantId == "" {
		return errors.New("routeAssertionId和tenantId不能为空")
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

// UpdateRouteAssertion 更新路由断言
func (dao *RouteAssertionDAO) UpdateRouteAssertion(ctx context.Context, assertion *models.RouteAssertion, operatorId string) error {
	if assertion.RouteAssertionId == "" || assertion.TenantId == "" {
		return errors.New("routeAssertionId和tenantId不能为空")
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

// GetRouteAssertionById 根据ID获取路由断言
func (dao *RouteAssertionDAO) GetRouteAssertionById(ctx context.Context, routeAssertionId, tenantId string) (*models.RouteAssertion, error) {
	if routeAssertionId == "" || tenantId == "" {
		return nil, errors.New("routeAssertionId和tenantId不能为空")
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

// RouterConfigDAO Router配置数据访问对象
type RouterConfigDAO struct {
	db database.Database
}

// NewRouterConfigDAO 创建Router配置DAO
func NewRouterConfigDAO(db database.Database) *RouterConfigDAO {
	return &RouterConfigDAO{
		db: db,
	}
}

// generateRouterConfigId 生成Router配置ID
// 格式：RC + YYYYMMDD + HHMMSS + 4位随机数
// 示例：RC20240615143022A1B2
func (dao *RouterConfigDAO) generateRouterConfigId() string {
	now := time.Now()
	// 生成时间部分：YYYYMMDDHHMMSS
	timeStr := now.Format("20060102150405")
	
	// 生成4位随机字符（大写字母和数字）
	randomStr := random.GenerateRandomString(4)
	
	return fmt.Sprintf("RC%s%s", timeStr, randomStr)
}

// isRouterConfigIdExists 检查Router配置ID是否已存在
func (dao *RouterConfigDAO) isRouterConfigIdExists(ctx context.Context, routerConfigId string) (bool, error) {
	query := `SELECT COUNT(*) as count FROM HUB_GW_ROUTER_CONFIG WHERE routerConfigId = ?`
	
	var result struct {
		Count int `db:"count"`
	}
	
	err := dao.db.QueryOne(ctx, &result, query, []interface{}{routerConfigId}, true)
	if err != nil {
		return false, err
	}
	
	return result.Count > 0, nil
}

// generateUniqueRouterConfigId 生成唯一的Router配置ID
func (dao *RouterConfigDAO) generateUniqueRouterConfigId(ctx context.Context) (string, error) {
	const maxAttempts = 10
	
	for attempt := 0; attempt < maxAttempts; attempt++ {
		routerConfigId := dao.generateRouterConfigId()
		
		exists, err := dao.isRouterConfigIdExists(ctx, routerConfigId)
		if err != nil {
			return "", huberrors.WrapError(err, "检查Router配置ID是否存在失败")
		}
		
		if !exists {
			return routerConfigId, nil
		}
		
		// 如果ID已存在，等待1毫秒后重试（确保时间戳不同）
		time.Sleep(time.Millisecond)
	}
	
	return "", errors.New("生成唯一Router配置ID失败，已达到最大尝试次数")
}

// AddRouterConfig 添加Router配置
func (dao *RouterConfigDAO) AddRouterConfig(ctx context.Context, routerConfig *models.RouterConfig, operatorId string) (string, error) {
	// 验证必填字段
	if routerConfig.TenantId == "" {
		return "", errors.New("租户ID不能为空")
	}
	if routerConfig.GatewayInstanceId == "" {
		return "", errors.New("网关实例ID不能为空")
	}
	if routerConfig.RouterName == "" {
		return "", errors.New("Router名称不能为空")
	}

	// 自动生成Router配置ID
	if routerConfig.RouterConfigId == "" {
		generatedId, err := dao.generateUniqueRouterConfigId(ctx)
		if err != nil {
			return "", huberrors.WrapError(err, "生成Router配置ID失败")
		}
		routerConfig.RouterConfigId = generatedId
	}

	// 设置一些自动填充的字段
	now := time.Now()
	routerConfig.AddTime = now
	routerConfig.AddWho = operatorId
	routerConfig.EditTime = now
	routerConfig.EditWho = operatorId
	routerConfig.OprSeqFlag = routerConfig.RouterConfigId + "_" + strings.ReplaceAll(time.Now().String(), ".", "")[:8]
	routerConfig.CurrentVersion = 1
	routerConfig.ActiveFlag = "Y"

	// 设置默认值
	if routerConfig.DefaultPriority == 0 {
		routerConfig.DefaultPriority = 100
	}
	if routerConfig.EnableRouteCache == "" {
		routerConfig.EnableRouteCache = "Y"
	}
	if routerConfig.RouteCacheTtlSeconds == 0 {
		routerConfig.RouteCacheTtlSeconds = 300
	}
	if routerConfig.EnableStrictMode == "" {
		routerConfig.EnableStrictMode = "N"
	}
	if routerConfig.EnableMetrics == "" {
		routerConfig.EnableMetrics = "Y"
	}
	if routerConfig.EnableTracing == "" {
		routerConfig.EnableTracing = "N"
	}
	if routerConfig.CaseSensitive == "" {
		routerConfig.CaseSensitive = "Y"
	}
	if routerConfig.RemoveTrailingSlash == "" {
		routerConfig.RemoveTrailingSlash = "Y"
	}
	if routerConfig.EnableGlobalFilters == "" {
		routerConfig.EnableGlobalFilters = "Y"
	}
	if routerConfig.FilterExecutionMode == "" {
		routerConfig.FilterExecutionMode = "SEQUENTIAL"
	}
	if routerConfig.EnableRoutePooling == "" {
		routerConfig.EnableRoutePooling = "N"
	}
	if routerConfig.EnableAsyncProcessing == "" {
		routerConfig.EnableAsyncProcessing = "N"
	}
	if routerConfig.EnableFallback == "" {
		routerConfig.EnableFallback = "Y"
	}
	if routerConfig.NotFoundStatusCode == 0 {
		routerConfig.NotFoundStatusCode = 404
	}
	if routerConfig.NotFoundMessage == "" {
		routerConfig.NotFoundMessage = "Route not found"
	}
	if routerConfig.CustomConfig == "" {
		routerConfig.CustomConfig = "{}"
	}

	// 使用数据库接口的Insert方法插入记录
	_, err := dao.db.Insert(ctx, "HUB_GW_ROUTER_CONFIG", routerConfig, true)
	if err != nil {
		// 检查是否是Router名重复错误
		if dao.isDuplicateRouterNameError(err) {
			return "", huberrors.WrapError(err, "Router名称已存在")
		}
		return "", huberrors.WrapError(err, "添加Router配置失败")
	}

	return routerConfig.RouterConfigId, nil
}

// GetRouterConfigById 根据Router配置ID获取Router配置信息
func (dao *RouterConfigDAO) GetRouterConfigById(ctx context.Context, routerConfigId, tenantId string) (*models.RouterConfig, error) {
	if routerConfigId == "" || tenantId == "" {
		return nil, errors.New("routerConfigId和tenantId不能为空")
	}

	query := `
		SELECT * FROM HUB_GW_ROUTER_CONFIG 
		WHERE routerConfigId = ? AND tenantId = ?
	`

	var routerConfig models.RouterConfig
	err := dao.db.QueryOne(ctx, &routerConfig, query, []interface{}{routerConfigId, tenantId}, true)

	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil // 没有找到记录，返回nil而不是错误
		}
		return nil, huberrors.WrapError(err, "查询Router配置失败")
	}

	return &routerConfig, nil
}

// UpdateRouterConfig 更新Router配置信息
func (dao *RouterConfigDAO) UpdateRouterConfig(ctx context.Context, routerConfig *models.RouterConfig, operatorId string) error {
	if routerConfig.RouterConfigId == "" || routerConfig.TenantId == "" {
		return errors.New("routerConfigId和tenantId不能为空")
	}

	// 首先获取Router配置当前版本
	currentConfig, err := dao.GetRouterConfigById(ctx, routerConfig.RouterConfigId, routerConfig.TenantId)
	if err != nil {
		return err
	}
	if currentConfig == nil {
		return errors.New("Router配置不存在")
	}

	// 更新版本和修改信息
	routerConfig.CurrentVersion = currentConfig.CurrentVersion + 1
	routerConfig.EditTime = time.Now()
	routerConfig.EditWho = operatorId
	routerConfig.OprSeqFlag = routerConfig.RouterConfigId + "_" + strings.ReplaceAll(time.Now().String(), ".", "")[:8]

	// 构建更新SQL
	sql := `
		UPDATE HUB_GW_ROUTER_CONFIG SET
			gatewayInstanceId = ?, routerName = ?, routerDesc = ?,
			defaultPriority = ?, enableRouteCache = ?, routeCacheTtlSeconds = ?,
			maxRoutes = ?, routeMatchTimeout = ?, enableStrictMode = ?,
			enableMetrics = ?, enableTracing = ?, caseSensitive = ?,
			removeTrailingSlash = ?, enableGlobalFilters = ?, filterExecutionMode = ?,
			maxFilterChainDepth = ?, enableRoutePooling = ?, routePoolSize = ?,
			enableAsyncProcessing = ?, enableFallback = ?, fallbackRoute = ?,
			notFoundStatusCode = ?, notFoundMessage = ?, routerMetadata = ?,
			customConfig = ?, reserved1 = ?, reserved2 = ?, reserved3 = ?,
			reserved4 = ?, reserved5 = ?, extProperty = ?, noteText = ?,
			editTime = ?, editWho = ?, oprSeqFlag = ?, currentVersion = ?
		WHERE routerConfigId = ? AND tenantId = ? AND currentVersion = ?
	`

	// 执行更新
	result, err := dao.db.Exec(ctx, sql, []interface{}{
		routerConfig.GatewayInstanceId, routerConfig.RouterName, routerConfig.RouterDesc,
		routerConfig.DefaultPriority, routerConfig.EnableRouteCache, routerConfig.RouteCacheTtlSeconds,
		routerConfig.MaxRoutes, routerConfig.RouteMatchTimeout, routerConfig.EnableStrictMode,
		routerConfig.EnableMetrics, routerConfig.EnableTracing, routerConfig.CaseSensitive,
		routerConfig.RemoveTrailingSlash, routerConfig.EnableGlobalFilters, routerConfig.FilterExecutionMode,
		routerConfig.MaxFilterChainDepth, routerConfig.EnableRoutePooling, routerConfig.RoutePoolSize,
		routerConfig.EnableAsyncProcessing, routerConfig.EnableFallback, routerConfig.FallbackRoute,
		routerConfig.NotFoundStatusCode, routerConfig.NotFoundMessage, routerConfig.RouterMetadata,
		routerConfig.CustomConfig, routerConfig.Reserved1, routerConfig.Reserved2, routerConfig.Reserved3,
		routerConfig.Reserved4, routerConfig.Reserved5, routerConfig.ExtProperty, routerConfig.NoteText,
		routerConfig.EditTime, routerConfig.EditWho, routerConfig.OprSeqFlag, routerConfig.CurrentVersion,
		routerConfig.RouterConfigId, routerConfig.TenantId, currentConfig.CurrentVersion,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "更新Router配置失败")
	}

	// 检查是否有记录被更新
	if result == 0 {
		return errors.New("Router配置数据已被其他用户修改，请刷新后重试")
	}

	return nil
}

// DeleteRouterConfig 删除Router配置
func (dao *RouterConfigDAO) DeleteRouterConfig(ctx context.Context, routerConfigId, tenantId, operatorId string) error {
	if routerConfigId == "" || tenantId == "" {
		return errors.New("routerConfigId和tenantId不能为空")
	}

	// 首先获取Router配置当前信息
	currentConfig, err := dao.GetRouterConfigById(ctx, routerConfigId, tenantId)
	if err != nil {
		return err
	}
	if currentConfig == nil {
		return errors.New("Router配置不存在")
	}

	// 执行实际删除
	sql := `DELETE FROM HUB_GW_ROUTER_CONFIG WHERE routerConfigId = ? AND tenantId = ?`

	result, err := dao.db.Exec(ctx, sql, []interface{}{routerConfigId, tenantId}, true)
	if err != nil {
		return huberrors.WrapError(err, "删除Router配置失败")
	}

	// 检查是否有记录被删除
	if result == 0 {
		return errors.New("Router配置不存在或已被删除")
	}

	return nil
}

// ListRouterConfigs 获取Router配置列表
func (dao *RouterConfigDAO) ListRouterConfigs(ctx context.Context, tenantId string, gatewayInstanceId string, page, pageSize int) ([]*models.RouterConfig, int, error) {
	if tenantId == "" {
		return nil, 0, errors.New("tenantId不能为空")
	}

	// 构建查询条件
	whereClause := "WHERE tenantId = ? AND activeFlag = 'Y'"
	args := []interface{}{tenantId}

	if gatewayInstanceId != "" {
		whereClause += " AND gatewayInstanceId = ?"
		args = append(args, gatewayInstanceId)
	}

	// 构建基础查询语句
	baseQuery := fmt.Sprintf("SELECT * FROM HUB_GW_ROUTER_CONFIG %s ORDER BY addTime DESC", whereClause)

	// 构建统计查询
	countQuery, err := sqlutils.BuildCountQuery(baseQuery)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建统计查询失败")
	}

	// 执行统计查询
	var countResult struct {
		Count int `db:"COUNT(*)"`
	}
	err = dao.db.QueryOne(ctx, &countResult, countQuery, args, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询Router配置总数失败")
	}

	// 如果没有记录，直接返回空列表
	if countResult.Count == 0 {
		return []*models.RouterConfig{}, 0, nil
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
	var routerConfigs []*models.RouterConfig
	err = dao.db.Query(ctx, &routerConfigs, paginatedQuery, allArgs, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询Router配置列表失败")
	}

	return routerConfigs, countResult.Count, nil
}

// GetRouterConfigsByGatewayInstance 根据网关实例ID获取所有Router配置
func (dao *RouterConfigDAO) GetRouterConfigsByGatewayInstance(ctx context.Context, gatewayInstanceId, tenantId string) ([]*models.RouterConfig, error) {
	if gatewayInstanceId == "" || tenantId == "" {
		return nil, errors.New("gatewayInstanceId和tenantId不能为空")
	}

	query := `
		SELECT * FROM HUB_GW_ROUTER_CONFIG 
		WHERE gatewayInstanceId = ? AND tenantId = ? AND activeFlag = 'Y'
		ORDER BY addTime DESC
	`

	var routerConfigs []*models.RouterConfig
	err := dao.db.Query(ctx, &routerConfigs, query, []interface{}{gatewayInstanceId, tenantId}, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询网关实例Router配置失败")
	}

	return routerConfigs, nil
}

// isDuplicateRouterNameError 检查是否是Router名重复错误
func (dao *RouterConfigDAO) isDuplicateRouterNameError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "duplicate") && strings.Contains(errStr, "routername")
} 

// GetRouteStatistics 获取路由统计信息
func (dao *RouteConfigDAO) GetRouteStatistics(ctx context.Context, tenantId string, gatewayInstanceId string) (map[string]int, error) {
	if tenantId == "" {
		return nil, errors.New("tenantId不能为空")
	}

	// 构建查询条件
	whereClause := "WHERE tenantId = ?"
	args := []interface{}{tenantId}

	if gatewayInstanceId != "" {
		whereClause += " AND gatewayInstanceId = ?"
		args = append(args, gatewayInstanceId)
	}

	// 构建统计查询SQL
	query := fmt.Sprintf(`
		SELECT 
			COUNT(*) as totalRoutes,
			SUM(CASE WHEN activeFlag = 'Y' THEN 1 ELSE 0 END) as activeRoutes,
			SUM(CASE WHEN activeFlag = 'N' THEN 1 ELSE 0 END) as inactiveRoutes,
			SUM(CASE WHEN matchType = 0 THEN 1 ELSE 0 END) as exactMatchRoutes,
			SUM(CASE WHEN matchType = 1 THEN 1 ELSE 0 END) as prefixMatchRoutes,
			SUM(CASE WHEN matchType = 2 THEN 1 ELSE 0 END) as regexMatchRoutes
		FROM HUB_GW_ROUTE_CONFIG 
		%s
	`, whereClause)

	// 执行统计查询
	var result struct {
		TotalRoutes       int `db:"totalRoutes"`
		ActiveRoutes      int `db:"activeRoutes"`
		InactiveRoutes    int `db:"inactiveRoutes"`
		ExactMatchRoutes  int `db:"exactMatchRoutes"`
		PrefixMatchRoutes int `db:"prefixMatchRoutes"`
		RegexMatchRoutes  int `db:"regexMatchRoutes"`
	}

	err := dao.db.QueryOne(ctx, &result, query, args, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询路由统计信息失败")
	}

	// 返回统计结果
	statistics := map[string]int{
		"totalRoutes":       result.TotalRoutes,
		"activeRoutes":      result.ActiveRoutes,
		"inactiveRoutes":    result.InactiveRoutes,
		"exactMatchRoutes":  result.ExactMatchRoutes,
		"prefixMatchRoutes": result.PrefixMatchRoutes,
		"regexMatchRoutes":  result.RegexMatchRoutes,
	}

	return statistics, nil
} 