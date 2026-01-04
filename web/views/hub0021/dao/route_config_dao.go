package dao

import (
	"context"
	"errors"
	"fmt"
	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
	"gateway/pkg/utils/huberrors"
	"gateway/pkg/utils/random"
	"gateway/web/views/hub0021/models"
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
		// 使用公共方法生成32位唯一字符串，前缀为"RT"
		routeConfig.RouteConfigId = random.GenerateUniqueStringWithPrefix("RT", 32)
	}

	// 设置一些自动填充的字段
	now := time.Now()
	routeConfig.AddTime = now
	routeConfig.AddWho = operatorId
	routeConfig.EditTime = now
	routeConfig.EditWho = operatorId
	routeConfig.OprSeqFlag = routeConfig.RouteConfigId
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
	if routeConfigId == "" {
		return nil, errors.New("routeConfigId不能为空")
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
	if routeConfig.RouteConfigId == "" {
		return errors.New("routeConfigId不能为空")
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
	if routeConfigId == "" {
		return errors.New("routeConfigId不能为空")
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
	if routeName == "" {
		return nil, errors.New("routeName不能为空")
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
func (dao *RouteConfigDAO) GetRouteConfigsByGatewayInstance(ctx context.Context, gatewayInstanceId, tenantId string, activeFlag string) ([]*models.RouteConfigWithService, error) {
	if gatewayInstanceId == "" {
		return nil, errors.New("gatewayInstanceId不能为空")
	}

	// 构建查询条件
	whereConditions := []string{"rc.gatewayInstanceId = ?", "rc.tenantId = ?"}
	args := []interface{}{gatewayInstanceId, tenantId}

	// 添加activeFlag条件（如果指定了activeFlag参数）
	if activeFlag != "" {
		whereConditions = append(whereConditions, "rc.activeFlag = ?")
		args = append(args, activeFlag)
	}

	whereClause := strings.Join(whereConditions, " AND ")

	query := fmt.Sprintf(`
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
		WHERE %s
		ORDER BY rc.routePriority ASC, rc.addTime DESC
	`, whereClause)

	var routeConfigs []*models.RouteConfigWithService
	err := dao.db.Query(ctx, &routeConfigs, query, args, true)
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

// GetRouteStatistics 获取路由统计信息
func (dao *RouteConfigDAO) GetRouteStatistics(ctx context.Context, tenantId string, gatewayInstanceId string) (map[string]int, error) {

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
