package dao

import (
	"context"
	"errors"
	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
	"gateway/pkg/utils/huberrors"
	"gateway/pkg/utils/random"
	"gateway/web/views/hubcommon002/models"
	"strings"
	"time"
)

// CorsConfigDAO CORS配置数据访问对象
type CorsConfigDAO struct {
	db database.Database
}

// NewCorsConfigDAO 创建CORS配置DAO
func NewCorsConfigDAO(db database.Database) *CorsConfigDAO {
	return &CorsConfigDAO{
		db: db,
	}
}

// AddCorsConfig 添加CORS配置
func (dao *CorsConfigDAO) AddCorsConfig(ctx context.Context, config *models.CorsConfig, operatorId string) error {
	if config.TenantId == "" || config.ConfigName == "" {
		return errors.New("tenantId和configName不能为空")
	}

	// 自动生成CORS配置ID
	if config.CorsConfigId == "" {
		// 使用公共方法生成唯一ID，支持集群环境
		// 格式：CORS前缀 + 32位唯一字符串 = 36位总长度
		config.CorsConfigId = random.GenerateUniqueStringWithPrefix("CORS", 32)
	}

	// 设置自动填充字段
	now := time.Now()
	config.AddTime = now
	config.AddWho = operatorId
	config.EditTime = now
	config.EditWho = operatorId
	config.OprSeqFlag = config.CorsConfigId + "_" + strings.ReplaceAll(time.Now().String(), ".", "")[:8]
	config.CurrentVersion = 1
	config.ActiveFlag = "Y"

	// 设置默认值
	if config.AllowCredentials == "" {
		config.AllowCredentials = "N"
	}
	if config.MaxAgeSeconds == 0 {
		config.MaxAgeSeconds = 86400
	}
	if config.ConfigPriority == 0 {
		config.ConfigPriority = 100
	}

	_, err := dao.db.Insert(ctx, "HUB_GW_CORS_CONFIG", config, true)
	if err != nil {
		return huberrors.WrapError(err, "添加CORS配置失败")
	}

	return nil
}

// GetCorsConfig 根据租户ID和配置ID获取CORS配置
func (dao *CorsConfigDAO) GetCorsConfig(tenantId, corsConfigId string) (*models.CorsConfig, error) {
	if corsConfigId == "" || tenantId == "" {
		return nil, errors.New("corsConfigId和tenantId不能为空")
	}

	query := `
		SELECT * FROM HUB_GW_CORS_CONFIG 
		WHERE corsConfigId = ? AND tenantId = ?
	`

	var config models.CorsConfig
	err := dao.db.QueryOne(context.Background(), &config, query, []interface{}{corsConfigId, tenantId}, true)

	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "查询CORS配置失败")
	}

	return &config, nil
}

// UpdateCorsConfig 更新CORS配置
func (dao *CorsConfigDAO) UpdateCorsConfig(ctx context.Context, config *models.CorsConfig, operatorId string) error {
	if config.CorsConfigId == "" || config.TenantId == "" {
		return errors.New("corsConfigId和tenantId不能为空")
	}

	// 首先获取当前配置
	currentConfig, err := dao.GetCorsConfig(config.TenantId, config.CorsConfigId)
	if err != nil {
		return err
	}
	if currentConfig == nil {
		return errors.New("CORS配置不存在")
	}

	// 更新修改信息
	config.EditTime = time.Now()
	config.EditWho = operatorId
	config.CurrentVersion = currentConfig.CurrentVersion + 1
	config.OprSeqFlag = config.CorsConfigId + "_" + strings.ReplaceAll(time.Now().String(), ".", "")[:8]

	// 保留不可修改的字段
	config.AddTime = currentConfig.AddTime
	config.AddWho = currentConfig.AddWho

	// 如果没有设置活动标记，保持原有状态
	if config.ActiveFlag == "" {
		config.ActiveFlag = currentConfig.ActiveFlag
	}

	sql := `
		UPDATE HUB_GW_CORS_CONFIG SET
			gatewayInstanceId = ?, routeConfigId = ?, configName = ?, allowOrigins = ?,
			allowMethods = ?, allowHeaders = ?, exposeHeaders = ?, allowCredentials = ?,
			maxAgeSeconds = ?, configPriority = ?, noteText = ?, editTime = ?, editWho = ?,
			currentVersion = ?, oprSeqFlag = ?, activeFlag = ?
		WHERE corsConfigId = ? AND tenantId = ? AND currentVersion = ?
	`

	result, err := dao.db.Exec(ctx, sql, []interface{}{
		config.GatewayInstanceId, config.RouteConfigId, config.ConfigName, config.AllowOrigins,
		config.AllowMethods, config.AllowHeaders, config.ExposeHeaders, config.AllowCredentials,
		config.MaxAgeSeconds, config.ConfigPriority, config.NoteText, config.EditTime, config.EditWho,
		config.CurrentVersion, config.OprSeqFlag, config.ActiveFlag,
		config.CorsConfigId, config.TenantId, currentConfig.CurrentVersion,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "更新CORS配置失败")
	}

	if result == 0 {
		return errors.New("CORS配置数据已被其他用户修改，请刷新后重试")
	}

	return nil
}

// DeleteCorsConfig 物理删除CORS配置
func (dao *CorsConfigDAO) DeleteCorsConfig(tenantId, corsConfigId string) error {
	// 校验主键
	if corsConfigId == "" {
		return errors.New("corsConfigId不能为空")
	}

	// 执行物理删除
	sql := `DELETE FROM HUB_GW_CORS_CONFIG WHERE corsConfigId = ? AND tenantId = ?`

	result, err := dao.db.Exec(context.Background(), sql, []interface{}{corsConfigId, tenantId}, true)
	if err != nil {
		return huberrors.WrapError(err, "删除CORS配置失败")
	}

	if result == 0 {
		return errors.New("未找到要删除的CORS配置")
	}

	return nil
}

// ListCorsConfigs 查询CORS配置列表（支持条件查询）
func (dao *CorsConfigDAO) ListCorsConfigs(ctx context.Context, tenantId string, query *models.CorsConfigQuery, page, pageSize int) ([]*models.CorsConfig, int, error) {
	// 构建查询条件
	whereClause := "WHERE tenantId = ?"
	var params []interface{}
	params = append(params, tenantId)

	// 添加查询条件
	if query != nil {
		if query.GatewayInstanceId != "" {
			whereClause += " AND gatewayInstanceId = ?"
			params = append(params, query.GatewayInstanceId)
		}
		if query.RouteConfigId != "" {
			whereClause += " AND routeConfigId = ?"
			params = append(params, query.RouteConfigId)
		}
		if query.ConfigName != "" {
			whereClause += " AND configName LIKE ?"
			params = append(params, "%"+query.ConfigName+"%")
		}
	}

	// 构建基础查询语句
	baseQuery := "SELECT * FROM HUB_GW_CORS_CONFIG " + whereClause + " ORDER BY configPriority ASC, addTime DESC"

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
		return nil, 0, huberrors.WrapError(err, "查询CORS配置总数失败")
	}
	total := result.Count

	// 如果没有记录，直接返回空列表
	if total == 0 {
		return []*models.CorsConfig{}, 0, nil
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
	var configs []*models.CorsConfig
	err = dao.db.Query(ctx, &configs, paginatedQuery, allArgs, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询CORS配置列表失败")
	}

	return configs, total, nil
}

// ListCorsConfigsByGatewayInstance 根据网关实例ID查询CORS配置列表
func (dao *CorsConfigDAO) ListCorsConfigsByGatewayInstance(ctx context.Context, tenantId, gatewayInstanceId string, page, pageSize int) ([]*models.CorsConfig, int, error) {
	if tenantId == "" || gatewayInstanceId == "" {
		return nil, 0, errors.New("tenantId和gatewayInstanceId不能为空")
	}

	// 构建基础查询语句
	baseQuery := "SELECT * FROM HUB_GW_CORS_CONFIG WHERE tenantId = ? AND gatewayInstanceId = ? ORDER BY configPriority ASC, addTime DESC"

	// 构建统计查询
	countQuery, err := sqlutils.BuildCountQuery(baseQuery)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建统计查询失败")
	}

	// 执行统计查询
	var totalResult struct {
		Total int `db:"COUNT(*)"`
	}
	err = dao.db.QueryOne(ctx, &totalResult, countQuery, []interface{}{tenantId, gatewayInstanceId}, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询网关实例CORS配置总数失败")
	}

	// 如果没有记录，直接返回空列表
	if totalResult.Total == 0 {
		return []*models.CorsConfig{}, 0, nil
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
	allArgs := append([]interface{}{tenantId, gatewayInstanceId}, paginationArgs...)

	// 执行分页查询
	var configs []*models.CorsConfig
	err = dao.db.Query(ctx, &configs, paginatedQuery, allArgs, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询网关实例CORS配置列表失败")
	}

	return configs, totalResult.Total, nil
}

// ListCorsConfigsByRouteConfig 根据路由配置ID查询CORS配置列表
func (dao *CorsConfigDAO) ListCorsConfigsByRouteConfig(ctx context.Context, tenantId, routeConfigId string, page, pageSize int) ([]*models.CorsConfig, int, error) {
	if tenantId == "" || routeConfigId == "" {
		return nil, 0, errors.New("tenantId和routeConfigId不能为空")
	}

	// 构建基础查询语句
	baseQuery := "SELECT * FROM HUB_GW_CORS_CONFIG WHERE tenantId = ? AND routeConfigId = ? ORDER BY configPriority ASC, addTime DESC"

	// 构建统计查询
	countQuery, err := sqlutils.BuildCountQuery(baseQuery)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建统计查询失败")
	}

	// 执行统计查询
	var totalResult struct {
		Total int `db:"COUNT(*)"`
	}
	err = dao.db.QueryOne(ctx, &totalResult, countQuery, []interface{}{tenantId, routeConfigId}, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询路由配置CORS配置总数失败")
	}

	// 如果没有记录，直接返回空列表
	if totalResult.Total == 0 {
		return []*models.CorsConfig{}, 0, nil
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
	allArgs := append([]interface{}{tenantId, routeConfigId}, paginationArgs...)

	// 执行分页查询
	var configs []*models.CorsConfig
	err = dao.db.Query(ctx, &configs, paginatedQuery, allArgs, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询路由配置CORS配置列表失败")
	}

	return configs, totalResult.Total, nil
}

// GetCorsConfigByGatewayInstance 根据网关实例ID查询单个CORS配置
func (dao *CorsConfigDAO) GetCorsConfigByGatewayInstance(tenantId, gatewayInstanceId string) (*models.CorsConfig, error) {
	if tenantId == "" || gatewayInstanceId == "" {
		return nil, errors.New("tenantId和gatewayInstanceId不能为空")
	}

	// 构建查询语句（使用 LIMIT 1 只取第一条记录）
	query := `
		SELECT * FROM HUB_GW_CORS_CONFIG 
		WHERE tenantId = ? AND gatewayInstanceId = ?
		ORDER BY configPriority ASC, addTime DESC
		LIMIT 1
	`

	var config models.CorsConfig
	err := dao.db.QueryOne(context.Background(), &config, query, []interface{}{tenantId, gatewayInstanceId}, true)

	if err != nil {
		if err == database.ErrRecordNotFound {
			// 没有数据返回空，不报错
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "查询网关实例CORS配置失败")
	}

	return &config, nil
}

// GetCorsConfigByRouteConfig 根据路由配置ID查询单个CORS配置
func (dao *CorsConfigDAO) GetCorsConfigByRouteConfig(tenantId, routeConfigId string) (*models.CorsConfig, error) {
	if tenantId == "" || routeConfigId == "" {
		return nil, errors.New("tenantId和routeConfigId不能为空")
	}

	// 构建查询语句（使用 LIMIT 1 只取第一条记录）
	query := `
		SELECT * FROM HUB_GW_CORS_CONFIG 
		WHERE tenantId = ? AND routeConfigId = ?
		ORDER BY configPriority ASC, addTime DESC
		LIMIT 1
	`

	var config models.CorsConfig
	err := dao.db.QueryOne(context.Background(), &config, query, []interface{}{tenantId, routeConfigId}, true)

	if err != nil {
		if err == database.ErrRecordNotFound {
			// 没有数据返回空，不报错
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "查询路由配置CORS配置失败")
	}

	return &config, nil
}
