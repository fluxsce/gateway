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

// AuthConfigDAO 认证配置数据访问对象
type AuthConfigDAO struct {
	db database.Database
}

// NewAuthConfigDAO 创建认证配置DAO
func NewAuthConfigDAO(db database.Database) *AuthConfigDAO {
	return &AuthConfigDAO{
		db: db,
	}
}

// generateAuthConfigId 生成认证配置ID
// 格式：AUTH + YYYYMMDD + HHMMSS + 4位随机数
// 示例：AUTH20240615143022A1B2
func (dao *AuthConfigDAO) generateAuthConfigId() string {
	now := time.Now()
	// 生成时间部分：YYYYMMDDHHMMSS
	timeStr := now.Format("20060102150405")
	
	// 生成4位随机字符（大写字母和数字）
	randomStr := random.GenerateRandomString(4)
	
	return fmt.Sprintf("AUTH%s%s", timeStr, randomStr)
}

// isAuthConfigIdExists 检查认证配置ID是否已存在
func (dao *AuthConfigDAO) isAuthConfigIdExists(ctx context.Context, authConfigId string) (bool, error) {
	query := `SELECT COUNT(*) as count FROM HUB_GW_AUTH_CONFIG WHERE authConfigId = ?`
	
	var result struct {
		Count int `db:"count"`
	}
	
	err := dao.db.QueryOne(ctx, &result, query, []interface{}{authConfigId}, true)
	if err != nil {
		return false, err
	}
	
	return result.Count > 0, nil
}

// generateUniqueAuthConfigId 生成唯一的认证配置ID
func (dao *AuthConfigDAO) generateUniqueAuthConfigId(ctx context.Context) (string, error) {
	const maxAttempts = 10
	
	for attempt := 0; attempt < maxAttempts; attempt++ {
		authConfigId := dao.generateAuthConfigId()
		
		exists, err := dao.isAuthConfigIdExists(ctx, authConfigId)
		if err != nil {
			return "", huberrors.WrapError(err, "检查认证配置ID是否存在失败")
		}
		
		if !exists {
			return authConfigId, nil
		}
		
		// 如果ID已存在，等待1毫秒后重试（确保时间戳不同）
		time.Sleep(time.Millisecond)
	}
	
	return "", errors.New("生成唯一认证配置ID失败，已达到最大尝试次数")
}

// AddAuthConfig 添加认证配置
func (dao *AuthConfigDAO) AddAuthConfig(ctx context.Context, config *models.AuthConfig, operatorId string) error {
	if config.TenantId == "" || config.AuthName == "" {
		return errors.New("tenantId和authName不能为空")
	}

	// 自动生成认证配置ID
	if config.AuthConfigId == "" {
		generatedId, err := dao.generateUniqueAuthConfigId(ctx)
		if err != nil {
			return huberrors.WrapError(err, "生成认证配置ID失败")
		}
		config.AuthConfigId = generatedId
	}

	// 设置自动填充字段
	now := time.Now()
	config.AddTime = now
	config.AddWho = operatorId
	config.EditTime = now
	config.EditWho = operatorId
	config.OprSeqFlag = config.AuthConfigId + "_" + strings.ReplaceAll(time.Now().String(), ".", "")[:8]
	config.CurrentVersion = 1
	config.ActiveFlag = "Y"

	// 设置默认值
	if config.AuthStrategy == "" {
		config.AuthStrategy = "REQUIRED"
	}
	if config.FailureStatusCode == 0 {
		config.FailureStatusCode = 401
	}
	if config.FailureMessage == "" {
		config.FailureMessage = "认证失败"
	}
	if config.ConfigPriority == 0 {
		config.ConfigPriority = 100
	}

	_, err := dao.db.Insert(ctx, "HUB_GW_AUTH_CONFIG", config, true)
	if err != nil {
		return huberrors.WrapError(err, "添加认证配置失败")
	}

	return nil
}

// GetAuthConfig 根据租户ID和配置ID获取认证配置
func (dao *AuthConfigDAO) GetAuthConfig(tenantId, authConfigId string) (*models.AuthConfig, error) {
	if authConfigId == "" || tenantId == "" {
		return nil, errors.New("authConfigId和tenantId不能为空")
	}

	query := `
		SELECT * FROM HUB_GW_AUTH_CONFIG 
		WHERE authConfigId = ? AND tenantId = ? AND activeFlag = 'Y'
	`

	var config models.AuthConfig
	err := dao.db.QueryOne(context.Background(), &config, query, []interface{}{authConfigId, tenantId}, true)

	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "查询认证配置失败")
	}

	return &config, nil
}

// UpdateAuthConfig 更新认证配置
func (dao *AuthConfigDAO) UpdateAuthConfig(ctx context.Context, config *models.AuthConfig, operatorId string) error {
	if config.AuthConfigId == "" || config.TenantId == "" {
		return errors.New("authConfigId和tenantId不能为空")
	}

	// 首先获取当前配置
	currentConfig, err := dao.GetAuthConfig(config.TenantId, config.AuthConfigId)
	if err != nil {
		return err
	}
	if currentConfig == nil {
		return errors.New("认证配置不存在")
	}

	// 更新修改信息
	config.EditTime = time.Now()
	config.EditWho = operatorId
	config.CurrentVersion = currentConfig.CurrentVersion + 1
	config.OprSeqFlag = config.AuthConfigId + "_" + strings.ReplaceAll(time.Now().String(), ".", "")[:8]

	sql := `
		UPDATE HUB_GW_AUTH_CONFIG SET
			gatewayInstanceId = ?, routeConfigId = ?, authName = ?, authType = ?,
			authStrategy = ?, authConfig = ?, exemptPaths = ?, exemptHeaders = ?,
			failureStatusCode = ?, failureMessage = ?, configPriority = ?, reserved1 = ?,
			reserved2 = ?, reserved3 = ?, reserved4 = ?, reserved5 = ?, extProperty = ?,
			noteText = ?, editTime = ?, editWho = ?, currentVersion = ?, oprSeqFlag = ?
		WHERE authConfigId = ? AND tenantId = ? AND activeFlag = 'Y'
	`

	_, err = dao.db.Exec(ctx, sql, []interface{}{
		config.GatewayInstanceId, config.RouteConfigId, config.AuthName, config.AuthType,
		config.AuthStrategy, config.AuthConfig, config.ExemptPaths, config.ExemptHeaders,
		config.FailureStatusCode, config.FailureMessage, config.ConfigPriority, config.Reserved1,
		config.Reserved2, config.Reserved3, config.Reserved4, config.Reserved5, config.ExtProperty,
		config.NoteText, config.EditTime, config.EditWho, config.CurrentVersion, config.OprSeqFlag,
		config.AuthConfigId, config.TenantId,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "更新认证配置失败")
	}

	return nil
}

// DeleteAuthConfig 软删除认证配置
func (dao *AuthConfigDAO) DeleteAuthConfig(tenantId, authConfigId, operatorId string) error {
	if authConfigId == "" || tenantId == "" {
		return errors.New("authConfigId和tenantId不能为空")
	}

	// 首先检查配置是否存在
	config, err := dao.GetAuthConfig(tenantId, authConfigId)
	if err != nil {
		return err
	}
	if config == nil {
		return errors.New("认证配置不存在")
	}

	now := time.Now()
	sql := `
		UPDATE HUB_GW_AUTH_CONFIG SET
			activeFlag = 'N', editTime = ?, editWho = ?
		WHERE authConfigId = ? AND tenantId = ? AND activeFlag = 'Y'
	`

	_, err = dao.db.Exec(context.Background(), sql, []interface{}{
		now, operatorId, authConfigId, tenantId,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "删除认证配置失败")
	}

	return nil
}

// ListAuthConfigs 查询认证配置列表
func (dao *AuthConfigDAO) ListAuthConfigs(ctx context.Context, tenantId string, page, pageSize int) ([]*models.AuthConfig, int, error) {
	if tenantId == "" {
		return nil, 0, errors.New("tenantId不能为空")
	}

	// 构建基础查询语句
	baseQuery := "SELECT * FROM HUB_GW_AUTH_CONFIG WHERE tenantId = ? AND activeFlag = 'Y' ORDER BY configPriority ASC, addTime DESC"

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
		return nil, 0, huberrors.WrapError(err, "查询认证配置总数失败")
	}

	// 如果没有记录，直接返回空列表
	if totalResult.Total == 0 {
		return []*models.AuthConfig{}, 0, nil
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
	var configs []*models.AuthConfig
	err = dao.db.Query(ctx, &configs, paginatedQuery, allArgs, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询认证配置列表失败")
	}

	return configs, totalResult.Total, nil
}

// GetAuthConfigByGatewayInstance 根据网关实例ID查询单个认证配置
func (dao *AuthConfigDAO) GetAuthConfigByGatewayInstance(tenantId, gatewayInstanceId string) (*models.AuthConfig, error) {
	if tenantId == "" || gatewayInstanceId == "" {
		return nil, errors.New("tenantId和gatewayInstanceId不能为空")
	}

	// 构建基础查询语句
	baseQuery := `
		SELECT * FROM HUB_GW_AUTH_CONFIG 
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
	var configs []*models.AuthConfig
	err = dao.db.Query(context.Background(), &configs, paginatedQuery, allArgs, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询网关实例认证配置失败")
	}

	// 返回第一条记录或nil
	if len(configs) > 0 {
		return configs[0], nil
	}
	return nil, nil
}

// GetAuthConfigByRouteConfig 根据路由配置ID查询单个认证配置
func (dao *AuthConfigDAO) GetAuthConfigByRouteConfig(tenantId, routeConfigId string) (*models.AuthConfig, error) {
	if tenantId == "" || routeConfigId == "" {
		return nil, errors.New("tenantId和routeConfigId不能为空")
	}

	// 构建基础查询语句
	baseQuery := `
		SELECT * FROM HUB_GW_AUTH_CONFIG 
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
	var configs []*models.AuthConfig
	err = dao.db.Query(context.Background(), &configs, paginatedQuery, allArgs, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询路由配置认证配置失败")
	}

	// 返回第一条记录或nil
	if len(configs) > 0 {
		return configs[0], nil
	}
	return nil, nil
} 