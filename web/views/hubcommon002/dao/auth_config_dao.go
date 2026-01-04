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

// AddAuthConfig 添加认证配置
func (dao *AuthConfigDAO) AddAuthConfig(ctx context.Context, config *models.AuthConfig, operatorId string) error {
	if config.TenantId == "" || config.AuthName == "" {
		return errors.New("tenantId和authName不能为空")
	}

	// 自动生成认证配置ID
	if config.AuthConfigId == "" {
		// 使用公共方法生成唯一ID，支持集群环境
		// 格式：AUTH前缀 + 32位唯一字符串 = 35位总长度
		config.AuthConfigId = random.GenerateUniqueStringWithPrefix("AUTH", 32)
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
		WHERE authConfigId = ? AND tenantId = ?
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

	// 保留不可修改的字段
	config.AddTime = currentConfig.AddTime
	config.AddWho = currentConfig.AddWho

	// 如果没有设置活动标记，保持原有状态
	if config.ActiveFlag == "" {
		config.ActiveFlag = currentConfig.ActiveFlag
	}

	sql := `
		UPDATE HUB_GW_AUTH_CONFIG SET
			gatewayInstanceId = ?, routeConfigId = ?, authName = ?, authType = ?,
			authStrategy = ?, authConfig = ?, exemptPaths = ?, exemptHeaders = ?,
			failureStatusCode = ?, failureMessage = ?, configPriority = ?, reserved1 = ?,
			reserved2 = ?, reserved3 = ?, reserved4 = ?, reserved5 = ?, extProperty = ?,
			noteText = ?, editTime = ?, editWho = ?, currentVersion = ?, oprSeqFlag = ?, activeFlag = ?
		WHERE authConfigId = ? AND tenantId = ? AND currentVersion = ?
	`

	result, err := dao.db.Exec(ctx, sql, []interface{}{
		config.GatewayInstanceId, config.RouteConfigId, config.AuthName, config.AuthType,
		config.AuthStrategy, config.AuthConfig, config.ExemptPaths, config.ExemptHeaders,
		config.FailureStatusCode, config.FailureMessage, config.ConfigPriority, config.Reserved1,
		config.Reserved2, config.Reserved3, config.Reserved4, config.Reserved5, config.ExtProperty,
		config.NoteText, config.EditTime, config.EditWho, config.CurrentVersion, config.OprSeqFlag, config.ActiveFlag,
		config.AuthConfigId, config.TenantId, currentConfig.CurrentVersion,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "更新认证配置失败")
	}

	if result == 0 {
		return errors.New("认证配置数据已被其他用户修改，请刷新后重试")
	}

	return nil
}

// DeleteAuthConfig 物理删除认证配置
func (dao *AuthConfigDAO) DeleteAuthConfig(tenantId, authConfigId string) error {
	// 校验主键
	if authConfigId == "" {
		return errors.New("authConfigId不能为空")
	}

	// 执行物理删除
	sql := `DELETE FROM HUB_GW_AUTH_CONFIG WHERE authConfigId = ? AND tenantId = ?`

	result, err := dao.db.Exec(context.Background(), sql, []interface{}{authConfigId, tenantId}, true)
	if err != nil {
		return huberrors.WrapError(err, "删除认证配置失败")
	}

	if result == 0 {
		return errors.New("未找到要删除的认证配置")
	}

	return nil
}

// ListAuthConfigs 查询认证配置列表
func (dao *AuthConfigDAO) ListAuthConfigs(ctx context.Context, tenantId string, page, pageSize int) ([]*models.AuthConfig, int, error) {
	if tenantId == "" {
		return nil, 0, errors.New("tenantId不能为空")
	}

	// 构建基础查询语句
	baseQuery := "SELECT * FROM HUB_GW_AUTH_CONFIG WHERE tenantId = ? ORDER BY configPriority ASC, addTime DESC"

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

	// 构建查询语句（使用 LIMIT 1 只取第一条记录）
	query := `
		SELECT * FROM HUB_GW_AUTH_CONFIG 
		WHERE tenantId = ? AND gatewayInstanceId = ?
		ORDER BY configPriority ASC, addTime DESC
		LIMIT 1
	`

	var config models.AuthConfig
	err := dao.db.QueryOne(context.Background(), &config, query, []interface{}{tenantId, gatewayInstanceId}, true)

	if err != nil {
		if err == database.ErrRecordNotFound {
			// 没有数据返回空，不报错
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "查询网关实例认证配置失败")
	}

	return &config, nil
}

// GetAuthConfigByRouteConfig 根据路由配置ID查询单个认证配置
func (dao *AuthConfigDAO) GetAuthConfigByRouteConfig(tenantId, routeConfigId string) (*models.AuthConfig, error) {
	if tenantId == "" || routeConfigId == "" {
		return nil, errors.New("tenantId和routeConfigId不能为空")
	}

	// 构建查询语句（使用 LIMIT 1 只取第一条记录）
	query := `
		SELECT * FROM HUB_GW_AUTH_CONFIG 
		WHERE tenantId = ? AND routeConfigId = ?
		ORDER BY configPriority ASC, addTime DESC
		LIMIT 1
	`

	var config models.AuthConfig
	err := dao.db.QueryOne(context.Background(), &config, query, []interface{}{tenantId, routeConfigId}, true)

	if err != nil {
		if err == database.ErrRecordNotFound {
			// 没有数据返回空，不报错
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "查询路由配置认证配置失败")
	}

	return &config, nil
}
