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

// AddRouterConfig 添加Router配置
func (dao *RouterConfigDAO) AddRouterConfig(ctx context.Context, routerConfig *models.RouterConfig, operatorId string) (string, error) {
	// 验证必填字段
	if routerConfig.GatewayInstanceId == "" {
		return "", errors.New("网关实例ID不能为空")
	}
	if routerConfig.RouterName == "" {
		return "", errors.New("Router名称不能为空")
	}

	// 自动生成Router配置ID（如果为空）
	if routerConfig.RouterConfigId == "" {
		// 使用公共方法生成32位唯一字符串，前缀为"RC"
		routerConfig.RouterConfigId = random.GenerateUniqueStringWithPrefix("RC", 32)
	}

	// 设置一些自动填充的字段
	now := time.Now()
	routerConfig.AddTime = now
	routerConfig.AddWho = operatorId
	routerConfig.EditTime = now
	routerConfig.EditWho = operatorId
	// 生成 OprSeqFlag，确保长度不超过32
	// RouterConfigId 已经是32位，直接使用
	routerConfig.OprSeqFlag = routerConfig.RouterConfigId
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
	// 生成 OprSeqFlag，确保长度不超过32
	// RouterConfigId 已经是32位，直接使用
	routerConfig.OprSeqFlag = routerConfig.RouterConfigId

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
	whereConditions := []string{"tenantId = ?", "activeFlag = 'Y'"}
	args := []interface{}{tenantId}

	if gatewayInstanceId != "" {
		whereConditions = append(whereConditions, "gatewayInstanceId = ?")
		args = append(args, gatewayInstanceId)
	}

	whereClause := "WHERE " + strings.Join(whereConditions, " AND ")

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

// GetRouterConfigByGatewayInstance 根据网关实例ID获取Router配置（返回单条数据）
func (dao *RouterConfigDAO) GetRouterConfigByGatewayInstance(ctx context.Context, gatewayInstanceId, tenantId string) (*models.RouterConfig, error) {
	if gatewayInstanceId == "" {
		return nil, errors.New("gatewayInstanceId不能为空")
	}

	query := `
		SELECT * FROM HUB_GW_ROUTER_CONFIG 
		WHERE gatewayInstanceId = ? AND tenantId = ? AND activeFlag = 'Y'
		ORDER BY addTime DESC
		LIMIT 1
	`

	var routerConfig models.RouterConfig
	err := dao.db.QueryOne(ctx, &routerConfig, query, []interface{}{gatewayInstanceId, tenantId}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil // 没有找到记录，返回nil而不是错误
		}
		return nil, huberrors.WrapError(err, "查询Router配置失败")
	}

	return &routerConfig, nil
}

// isDuplicateRouterNameError 检查是否是Router名重复错误
func (dao *RouterConfigDAO) isDuplicateRouterNameError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "duplicate") && strings.Contains(errStr, "routername")
}
