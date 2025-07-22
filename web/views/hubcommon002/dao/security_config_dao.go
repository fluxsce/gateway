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

// SecurityConfigDAO 安全配置数据访问对象
type SecurityConfigDAO struct {
	db database.Database
}

// NewSecurityConfigDAO 创建安全配置DAO
func NewSecurityConfigDAO(db database.Database) *SecurityConfigDAO {
	return &SecurityConfigDAO{
		db: db,
	}
}

// generateSecurityConfigId 生成安全配置ID
// 格式：SC + YYYYMMDD + HHMMSS + 4位随机数
// 示例：SC20240615143022A1B2
func (dao *SecurityConfigDAO) generateSecurityConfigId() string {
	now := time.Now()
	// 生成时间部分：YYYYMMDDHHMMSS
	timeStr := now.Format("20060102150405")
	
	// 生成4位随机字符（大写字母和数字）
	randomStr := random.GenerateRandomString(4)
	
	return fmt.Sprintf("SC%s%s", timeStr, randomStr)
}

// isSecurityConfigIdExists 检查安全配置ID是否已存在
func (dao *SecurityConfigDAO) isSecurityConfigIdExists(ctx context.Context, securityConfigId string) (bool, error) {
	query := `SELECT COUNT(*) as count FROM HUB_GW_SECURITY_CONFIG WHERE securityConfigId = ?`
	
	var result struct {
		Count int `db:"count"`
	}
	
	err := dao.db.QueryOne(ctx, &result, query, []interface{}{securityConfigId}, true)
	if err != nil {
		return false, err
	}
	
	return result.Count > 0, nil
}

// generateUniqueSecurityConfigId 生成唯一的安全配置ID
// 如果生成的ID已存在，会重新生成直到找到唯一的ID（最多尝试10次）
func (dao *SecurityConfigDAO) generateUniqueSecurityConfigId(ctx context.Context) (string, error) {
	const maxAttempts = 10
	
	for attempt := 0; attempt < maxAttempts; attempt++ {
		securityConfigId := dao.generateSecurityConfigId()
		
		exists, err := dao.isSecurityConfigIdExists(ctx, securityConfigId)
		if err != nil {
			return "", huberrors.WrapError(err, "检查安全配置ID是否存在失败")
		}
		
		if !exists {
			return securityConfigId, nil
		}
		
		// 如果ID已存在，等待1毫秒后重试（确保时间戳不同）
		time.Sleep(time.Millisecond)
	}
	
	return "", errors.New("生成唯一安全配置ID失败，已达到最大尝试次数")
}

// AddSecurityConfig 添加安全配置
// 参数:
//   - ctx: 上下文对象
//   - config: 安全配置信息
//   - operatorId: 操作人ID
//
// 返回:
//   - securityConfigId: 新创建的安全配置ID
//   - err: 可能的错误
func (dao *SecurityConfigDAO) AddSecurityConfig(ctx context.Context, config *models.SecurityConfig, operatorId string) (string, error) {
	// 验证租户ID
	if config.TenantId == "" {
		return "", errors.New("租户ID不能为空")
	}

	// 自动生成安全配置ID（如果为空）
	if config.SecurityConfigId == "" {
		generatedId, err := dao.generateUniqueSecurityConfigId(ctx)
		if err != nil {
			return "", huberrors.WrapError(err, "生成安全配置ID失败")
		}
		config.SecurityConfigId = generatedId
	} else {
		// 如果提供了ID，检查是否已存在
		exists, err := dao.isSecurityConfigIdExists(ctx, config.SecurityConfigId)
		if err != nil {
			return "", huberrors.WrapError(err, "检查安全配置ID是否存在失败")
		}
		if exists {
			return "", errors.New("安全配置ID已存在")
		}
	}

	// 设置一些自动填充的字段
	now := time.Now()
	config.AddTime = now
	config.AddWho = operatorId
	config.EditTime = now
	config.EditWho = operatorId
	config.OprSeqFlag = config.SecurityConfigId + "_" + strings.ReplaceAll(time.Now().String(), ".", "")[:8]
	config.CurrentVersion = 1
	config.ActiveFlag = "Y"

	// 设置默认值
	if config.ConfigPriority == 0 {
		config.ConfigPriority = 100 // 默认优先级
	}

	// 使用数据库接口的Insert方法插入记录
	_, err := dao.db.Insert(ctx, "HUB_GW_SECURITY_CONFIG", config, true)

	if err != nil {
		// 检查是否是配置名重复错误
		if dao.isDuplicateConfigNameError(err) {
			return "", huberrors.WrapError(err, "配置名已存在")
		}
		return "", huberrors.WrapError(err, "添加安全配置失败")
	}

	return config.SecurityConfigId, nil
}

// GetSecurityConfigById 根据安全配置ID获取安全配置信息
func (dao *SecurityConfigDAO) GetSecurityConfigById(ctx context.Context, securityConfigId, tenantId string) (*models.SecurityConfig, error) {
	if securityConfigId == "" || tenantId == "" {
		return nil, errors.New("securityConfigId和tenantId不能为空")
	}

	query := `
		SELECT * FROM HUB_GW_SECURITY_CONFIG 
		WHERE securityConfigId = ? AND tenantId = ?
	`

	var config models.SecurityConfig
	err := dao.db.QueryOne(ctx, &config, query, []interface{}{securityConfigId, tenantId}, true)

	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil // 没有找到记录，返回nil而不是错误
		}
		return nil, huberrors.WrapError(err, "查询安全配置失败")
	}

	return &config, nil
}

// UpdateSecurityConfig 更新安全配置信息
func (dao *SecurityConfigDAO) UpdateSecurityConfig(ctx context.Context, config *models.SecurityConfig, operatorId string) error {
	if config.SecurityConfigId == "" || config.TenantId == "" {
		return errors.New("securityConfigId和tenantId不能为空")
	}

	// 首先获取安全配置当前版本
	currentConfig, err := dao.GetSecurityConfigById(ctx, config.SecurityConfigId, config.TenantId)
	if err != nil {
		return err
	}
	if currentConfig == nil {
		return errors.New("安全配置不存在")
	}

	// 更新版本和修改信息
	config.CurrentVersion = currentConfig.CurrentVersion + 1
	config.EditTime = time.Now()
	config.EditWho = operatorId
	config.OprSeqFlag = config.SecurityConfigId + "_" + strings.ReplaceAll(time.Now().String(), ".", "")[:8]

	// 保留不可修改的字段
	config.AddTime = currentConfig.AddTime
	config.AddWho = currentConfig.AddWho
	
	// 如果没有设置活动标记，保持原有状态
	if config.ActiveFlag == "" {
		config.ActiveFlag = currentConfig.ActiveFlag
	}

	// 构建更新SQL
	sql := `
		UPDATE HUB_GW_SECURITY_CONFIG SET
			gatewayInstanceId = ?, routeConfigId = ?, configName = ?, configDesc = ?,
			configPriority = ?, customConfigJson = ?, reserved1 = ?, reserved2 = ?,
			reserved3 = ?, reserved4 = ?, reserved5 = ?, extProperty = ?, noteText = ?,
			editTime = ?, editWho = ?, oprSeqFlag = ?, currentVersion = ?, activeFlag = ?
		WHERE securityConfigId = ? AND tenantId = ? AND currentVersion = ?
	`

	// 执行更新
	result, err := dao.db.Exec(ctx, sql, []interface{}{
		config.GatewayInstanceId, config.RouteConfigId, config.ConfigName, config.ConfigDesc,
		config.ConfigPriority, config.CustomConfigJson, config.Reserved1, config.Reserved2,
		config.Reserved3, config.Reserved4, config.Reserved5, config.ExtProperty, config.NoteText,
		config.EditTime, config.EditWho, config.OprSeqFlag, config.CurrentVersion, config.ActiveFlag,
		config.SecurityConfigId, config.TenantId, currentConfig.CurrentVersion,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "更新安全配置失败")
	}

	// 检查是否有记录被更新
	if result == 0 {
		return errors.New("安全配置数据已被其他用户修改，请刷新后重试")
	}

	return nil
}

// DeleteSecurityConfig 物理删除安全配置
func (dao *SecurityConfigDAO) DeleteSecurityConfig(ctx context.Context, securityConfigId, tenantId, operatorId string) error {
	if securityConfigId == "" || tenantId == "" {
		return errors.New("securityConfigId和tenantId不能为空")
	}

	// 首先获取安全配置当前信息，确认记录存在
	currentConfig, err := dao.GetSecurityConfigById(ctx, securityConfigId, tenantId)
	if err != nil {
		return err
	}
	if currentConfig == nil {
		return errors.New("安全配置不存在")
	}

	// 执行物理删除
	sql := `DELETE FROM HUB_GW_SECURITY_CONFIG WHERE securityConfigId = ? AND tenantId = ?`

	result, err := dao.db.Exec(ctx, sql, []interface{}{securityConfigId, tenantId}, true)
	if err != nil {
		return huberrors.WrapError(err, "删除安全配置失败")
	}

	// 检查是否有记录被删除
	if result == 0 {
		return errors.New("未找到要删除的安全配置")
	}

	return nil
}

// ListSecurityConfigs 获取安全配置列表
func (dao *SecurityConfigDAO) ListSecurityConfigs(ctx context.Context, tenantId string, page, pageSize int) ([]*models.SecurityConfig, int, error) {
	if tenantId == "" {
		return nil, 0, errors.New("tenantId不能为空")
	}

	// 构建基础查询语句
	baseQuery := "SELECT * FROM HUB_GW_SECURITY_CONFIG WHERE tenantId = ? ORDER BY configPriority ASC, addTime DESC"

	// 构建统计查询
	countQuery, err := sqlutils.BuildCountQuery(baseQuery)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建统计查询失败")
	}

	// 执行统计查询
	var result struct {
		Count int `db:"COUNT(*)"`
	}
	err = dao.db.QueryOne(ctx, &result, countQuery, []interface{}{tenantId}, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询安全配置总数失败")
	}
	total := result.Count

	// 如果没有记录，直接返回空列表
	if total == 0 {
		return []*models.SecurityConfig{}, 0, nil
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
	var configs []*models.SecurityConfig
	err = dao.db.Query(ctx, &configs, paginatedQuery, allArgs, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询安全配置列表失败")
	}

	return configs, total, nil
}

// FindSecurityConfigByName 根据配置名查找安全配置
func (dao *SecurityConfigDAO) FindSecurityConfigByName(ctx context.Context, configName, tenantId string) (*models.SecurityConfig, error) {
	if configName == "" || tenantId == "" {
		return nil, errors.New("configName和tenantId不能为空")
	}

	query := `
		SELECT * FROM HUB_GW_SECURITY_CONFIG 
		WHERE configName = ? AND tenantId = ?
	`

	var config models.SecurityConfig
	err := dao.db.QueryOne(ctx, &config, query, []interface{}{configName, tenantId}, true)

	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "查询安全配置失败")
	}

	return &config, nil
}

// ListSecurityConfigsByGatewayInstance 根据网关实例ID获取安全配置列表
func (dao *SecurityConfigDAO) ListSecurityConfigsByGatewayInstance(ctx context.Context, gatewayInstanceId, tenantId string) ([]*models.SecurityConfig, error) {
	if gatewayInstanceId == "" || tenantId == "" {
		return nil, errors.New("gatewayInstanceId和tenantId不能为空")
	}

	query := `
		SELECT * FROM HUB_GW_SECURITY_CONFIG 
		WHERE gatewayInstanceId = ? AND tenantId = ? 
		ORDER BY configPriority ASC, addTime DESC
	`

	var configs []*models.SecurityConfig
	err := dao.db.Query(ctx, &configs, query, []interface{}{gatewayInstanceId, tenantId}, true)

	if err != nil {
		return nil, huberrors.WrapError(err, "查询网关实例安全配置列表失败")
	}

	return configs, nil
}

// ListSecurityConfigsByRouteConfig 根据路由配置ID获取安全配置列表
func (dao *SecurityConfigDAO) ListSecurityConfigsByRouteConfig(ctx context.Context, routeConfigId, tenantId string) ([]*models.SecurityConfig, error) {
	if routeConfigId == "" || tenantId == "" {
		return nil, errors.New("routeConfigId和tenantId不能为空")
	}

	query := `
		SELECT * FROM HUB_GW_SECURITY_CONFIG 
		WHERE routeConfigId = ? AND tenantId = ?  
		ORDER BY configPriority ASC, addTime DESC
	`

	var configs []*models.SecurityConfig
	err := dao.db.Query(ctx, &configs, query, []interface{}{routeConfigId, tenantId}, true)

	if err != nil {
		return nil, huberrors.WrapError(err, "查询路由配置安全配置列表失败")
	}

	return configs, nil
}

// isDuplicateConfigNameError 检查是否是配置名重复错误
func (dao *SecurityConfigDAO) isDuplicateConfigNameError(err error) bool {
	if err == nil {
		return false
	}
	errorMsg := strings.ToLower(err.Error())
	return strings.Contains(errorMsg, "duplicate") && strings.Contains(errorMsg, "configname")
} 