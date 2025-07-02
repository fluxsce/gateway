package dao

import (
	"context"
	"errors"
	"fmt"
	"gohub/pkg/database"
	"gohub/pkg/database/sqlutils"
	"gohub/pkg/utils/huberrors"
	"gohub/pkg/utils/random"
	"gohub/web/views/hubcommon002/models"
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

// generateCorsConfigId 生成CORS配置ID
// 格式：CORS + YYYYMMDD + HHMMSS + 4位随机数
// 示例：CORS20240615143022A1B2
func (dao *CorsConfigDAO) generateCorsConfigId() string {
	now := time.Now()
	// 生成时间部分：YYYYMMDDHHMMSS
	timeStr := now.Format("20060102150405")
	
	// 生成4位随机字符（大写字母和数字）
	randomStr := random.GenerateRandomString(4)
	
	return fmt.Sprintf("CORS%s%s", timeStr, randomStr)
}

// isCorsConfigIdExists 检查CORS配置ID是否已存在
func (dao *CorsConfigDAO) isCorsConfigIdExists(ctx context.Context, corsConfigId string) (bool, error) {
	query := `SELECT COUNT(*) as count FROM HUB_GW_CORS_CONFIG WHERE corsConfigId = ?`
	
	var result struct {
		Count int `db:"count"`
	}
	
	err := dao.db.QueryOne(ctx, &result, query, []interface{}{corsConfigId}, true)
	if err != nil {
		return false, err
	}
	
	return result.Count > 0, nil
}

// generateUniqueCorsConfigId 生成唯一的CORS配置ID
func (dao *CorsConfigDAO) generateUniqueCorsConfigId(ctx context.Context) (string, error) {
	const maxAttempts = 10
	
	for attempt := 0; attempt < maxAttempts; attempt++ {
		corsConfigId := dao.generateCorsConfigId()
		
		exists, err := dao.isCorsConfigIdExists(ctx, corsConfigId)
		if err != nil {
			return "", huberrors.WrapError(err, "检查CORS配置ID是否存在失败")
		}
		
		if !exists {
			return corsConfigId, nil
		}
		
		// 如果ID已存在，等待1毫秒后重试（确保时间戳不同）
		time.Sleep(time.Millisecond)
	}
	
	return "", errors.New("生成唯一CORS配置ID失败，已达到最大尝试次数")
}

// AddCorsConfig 添加CORS配置
func (dao *CorsConfigDAO) AddCorsConfig(ctx context.Context, config *models.CorsConfig, operatorId string) error {
	if config.TenantId == "" || config.ConfigName == "" {
		return errors.New("tenantId和configName不能为空")
	}

	// 自动生成CORS配置ID
	if config.CorsConfigId == "" {
		generatedId, err := dao.generateUniqueCorsConfigId(ctx)
		if err != nil {
			return huberrors.WrapError(err, "生成CORS配置ID失败")
		}
		config.CorsConfigId = generatedId
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
		WHERE corsConfigId = ? AND tenantId = ? AND activeFlag = 'Y'
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

	sql := `
		UPDATE HUB_GW_CORS_CONFIG SET
			gatewayInstanceId = ?, routeConfigId = ?, configName = ?, allowOrigins = ?,
			allowMethods = ?, allowHeaders = ?, exposeHeaders = ?, allowCredentials = ?,
			maxAgeSeconds = ?, configPriority = ?, noteText = ?, editTime = ?, editWho = ?,
			currentVersion = ?, oprSeqFlag = ?
		WHERE corsConfigId = ? AND tenantId = ? AND activeFlag = 'Y'
	`

	_, err = dao.db.Exec(ctx, sql, []interface{}{
		config.GatewayInstanceId, config.RouteConfigId, config.ConfigName, config.AllowOrigins,
		config.AllowMethods, config.AllowHeaders, config.ExposeHeaders, config.AllowCredentials,
		config.MaxAgeSeconds, config.ConfigPriority, config.NoteText, config.EditTime, config.EditWho,
		config.CurrentVersion, config.OprSeqFlag, config.CorsConfigId, config.TenantId,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "更新CORS配置失败")
	}

	return nil
}

// DeleteCorsConfig 软删除CORS配置
func (dao *CorsConfigDAO) DeleteCorsConfig(tenantId, corsConfigId, operatorId string) error {
	if corsConfigId == "" || tenantId == "" {
		return errors.New("corsConfigId和tenantId不能为空")
	}

	// 首先检查配置是否存在
	config, err := dao.GetCorsConfig(tenantId, corsConfigId)
	if err != nil {
		return err
	}
	if config == nil {
		return errors.New("CORS配置不存在")
	}

	now := time.Now()
	sql := `
		UPDATE HUB_GW_CORS_CONFIG SET
			activeFlag = 'N', editTime = ?, editWho = ?
		WHERE corsConfigId = ? AND tenantId = ? AND activeFlag = 'Y'
	`

	_, err = dao.db.Exec(context.Background(), sql, []interface{}{
		now, operatorId, corsConfigId, tenantId,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "删除CORS配置失败")
	}

	return nil
}

// ListCorsConfigs 查询CORS配置列表
func (dao *CorsConfigDAO) ListCorsConfigs(ctx context.Context, tenantId string, page, pageSize int) ([]*models.CorsConfig, int, error) {
	if tenantId == "" {
		return nil, 0, errors.New("tenantId不能为空")
	}

	// 构建基础查询语句
	baseQuery := "SELECT * FROM HUB_GW_CORS_CONFIG WHERE tenantId = ? AND activeFlag = 'Y' ORDER BY configPriority ASC, addTime DESC"

	// 构建统计查询
	countQuery, err := sqlutils.BuildCountQuery(baseQuery)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建统计查询失败")
	}

	// 执行统计查询
	var totalResult struct {
		Total int `db:"COUNT(*)"`
	}
	err = dao.db.QueryOne(ctx, &totalResult, countQuery, []interface{}{tenantId}, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询CORS配置总数失败")
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
	allArgs := append([]interface{}{tenantId}, paginationArgs...)

	// 执行分页查询
	var configs []*models.CorsConfig
	err = dao.db.Query(ctx, &configs, paginatedQuery, allArgs, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询CORS配置列表失败")
	}

	return configs, totalResult.Total, nil
}

// ListCorsConfigsByGatewayInstance 根据网关实例ID查询CORS配置列表
func (dao *CorsConfigDAO) ListCorsConfigsByGatewayInstance(ctx context.Context, tenantId, gatewayInstanceId string, page, pageSize int) ([]*models.CorsConfig, int, error) {
	if tenantId == "" || gatewayInstanceId == "" {
		return nil, 0, errors.New("tenantId和gatewayInstanceId不能为空")
	}

	// 构建基础查询语句
	baseQuery := "SELECT * FROM HUB_GW_CORS_CONFIG WHERE tenantId = ? AND gatewayInstanceId = ? AND activeFlag = 'Y' ORDER BY configPriority ASC, addTime DESC"

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
	baseQuery := "SELECT * FROM HUB_GW_CORS_CONFIG WHERE tenantId = ? AND routeConfigId = ? AND activeFlag = 'Y' ORDER BY configPriority ASC, addTime DESC"

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

	// 构建基础查询语句
	baseQuery := `
		SELECT * FROM HUB_GW_CORS_CONFIG 
		WHERE tenantId = ? AND gatewayInstanceId = ? AND activeFlag = 'Y'
		ORDER BY configPriority ASC, addTime DESC
	`

	// 创建分页信息（只取第一条记录）
	pagination := sqlutils.NewPaginationInfo(1, 1)

	// 获取数据库类型
	dbType := sqlutils.GetDatabaseType(dao.db)

	// 构建分页查询
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
	if err != nil {
		return nil, huberrors.WrapError(err, "构建分页查询失败")
	}

	// 合并查询参数
	allArgs := append([]interface{}{tenantId, gatewayInstanceId}, paginationArgs...)

	// 执行查询
	var configs []*models.CorsConfig
	err = dao.db.Query(context.Background(), &configs, paginatedQuery, allArgs, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询网关实例CORS配置失败")
	}

	// 返回第一条记录或nil
	if len(configs) > 0 {
		return configs[0], nil
	}
	return nil, nil
}

// GetCorsConfigByRouteConfig 根据路由配置ID查询单个CORS配置
func (dao *CorsConfigDAO) GetCorsConfigByRouteConfig(tenantId, routeConfigId string) (*models.CorsConfig, error) {
	if tenantId == "" || routeConfigId == "" {
		return nil, errors.New("tenantId和routeConfigId不能为空")
	}

	// 构建基础查询语句
	baseQuery := `
		SELECT * FROM HUB_GW_CORS_CONFIG 
		WHERE tenantId = ? AND routeConfigId = ? AND activeFlag = 'Y'
		ORDER BY configPriority ASC, addTime DESC
	`

	// 创建分页信息（只取第一条记录）
	pagination := sqlutils.NewPaginationInfo(1, 1)

	// 获取数据库类型
	dbType := sqlutils.GetDatabaseType(dao.db)

	// 构建分页查询
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
	if err != nil {
		return nil, huberrors.WrapError(err, "构建分页查询失败")
	}

	// 合并查询参数
	allArgs := append([]interface{}{tenantId, routeConfigId}, paginationArgs...)

	// 执行查询
	var configs []*models.CorsConfig
	err = dao.db.Query(context.Background(), &configs, paginatedQuery, allArgs, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询路由配置CORS配置失败")
	}

	// 返回第一条记录或nil
	if len(configs) > 0 {
		return configs[0], nil
	}
	return nil, nil
}

 