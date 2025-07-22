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

// UseragentAccessConfigDAO User-Agent访问控制配置数据访问对象
type UseragentAccessConfigDAO struct {
	db database.Database
}

// NewUseragentAccessConfigDAO 创建User-Agent访问控制配置DAO
func NewUseragentAccessConfigDAO(db database.Database) *UseragentAccessConfigDAO {
	return &UseragentAccessConfigDAO{
		db: db,
	}
}

// generateUseragentAccessConfigId 生成User-Agent访问配置ID
// 格式：UA + YYYYMMDD + HHMMSS + 4位随机数
// 示例：UA20240615143022A1B2
func (dao *UseragentAccessConfigDAO) generateUseragentAccessConfigId() string {
	now := time.Now()
	// 生成时间部分：YYYYMMDDHHMMSS
	timeStr := now.Format("20060102150405")
	
	// 生成4位随机字符（大写字母和数字）
	randomStr := random.GenerateRandomString(4)
	
	return fmt.Sprintf("UA%s%s", timeStr, randomStr)
}



// isUseragentAccessConfigIdExists 检查User-Agent访问配置ID是否已存在
func (dao *UseragentAccessConfigDAO) isUseragentAccessConfigIdExists(ctx context.Context, useragentAccessConfigId string) (bool, error) {
	query := `SELECT COUNT(*) as count FROM HUB_GW_UA_ACCESS_CONFIG WHERE useragentAccessConfigId = ?`
	
	var result struct {
		Count int `db:"count"`
	}
	
	err := dao.db.QueryOne(ctx, &result, query, []interface{}{useragentAccessConfigId}, true)
	if err != nil {
		return false, err
	}
	
	return result.Count > 0, nil
}

// generateUniqueUseragentAccessConfigId 生成唯一的User-Agent访问配置ID
func (dao *UseragentAccessConfigDAO) generateUniqueUseragentAccessConfigId(ctx context.Context) (string, error) {
	const maxAttempts = 10
	
	for attempt := 0; attempt < maxAttempts; attempt++ {
		useragentAccessConfigId := dao.generateUseragentAccessConfigId()
		
		exists, err := dao.isUseragentAccessConfigIdExists(ctx, useragentAccessConfigId)
		if err != nil {
			return "", huberrors.WrapError(err, "检查User-Agent访问配置ID是否存在失败")
		}
		
		if !exists {
			return useragentAccessConfigId, nil
		}
		
		// 如果ID已存在，等待1毫秒后重试（确保时间戳不同）
		time.Sleep(time.Millisecond)
	}
	
	return "", errors.New("生成唯一User-Agent访问配置ID失败，已达到最大尝试次数")
}

// AddUseragentAccessConfig 添加User-Agent访问控制配置
func (dao *UseragentAccessConfigDAO) AddUseragentAccessConfig(ctx context.Context, config *models.UseragentAccessConfig, operatorId string) error {
	if config.TenantId == "" || config.SecurityConfigId == "" {
		return errors.New("tenantId和securityConfigId不能为空")
	}

	// 自动生成User-Agent访问配置ID
	if config.UseragentAccessConfigId == "" {
		generatedId, err := dao.generateUniqueUseragentAccessConfigId(ctx)
		if err != nil {
			return huberrors.WrapError(err, "生成User-Agent访问配置ID失败")
		}
		config.UseragentAccessConfigId = generatedId
	}

	// 设置自动填充字段
	now := time.Now()
	config.AddTime = now
	config.AddWho = operatorId
	config.EditTime = now
	config.EditWho = operatorId
	config.OprSeqFlag = config.UseragentAccessConfigId + "_" + strings.ReplaceAll(time.Now().String(), ".", "")[:8]
	config.CurrentVersion = 1
	config.ActiveFlag = "Y"

	// 设置默认值
	if config.BlockEmptyUserAgent == "" {
		config.BlockEmptyUserAgent = "N"
	}
	if config.DefaultPolicy == "" {
		config.DefaultPolicy = "allow"
	}

	_, err := dao.db.Insert(ctx, "HUB_GW_UA_ACCESS_CONFIG", config, true)
	if err != nil {
		return huberrors.WrapError(err, "添加User-Agent访问控制配置失败")
	}

	return nil
}

// GetUseragentAccessConfigBySecurityConfigId 根据安全配置ID获取User-Agent访问控制配置
func (dao *UseragentAccessConfigDAO) GetUseragentAccessConfigBySecurityConfigId(ctx context.Context, securityConfigId, tenantId string) (*models.UseragentAccessConfig, error) {
	if securityConfigId == "" || tenantId == "" {
		return nil, errors.New("securityConfigId和tenantId不能为空")
	}

	query := `
		SELECT * FROM HUB_GW_UA_ACCESS_CONFIG 
		WHERE securityConfigId = ? AND tenantId = ?
	`

	var config models.UseragentAccessConfig
	err := dao.db.QueryOne(ctx, &config, query, []interface{}{securityConfigId, tenantId}, true)

	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "查询User-Agent访问控制配置失败")
	}

	return &config, nil
}

// GetUseragentAccessConfigById 根据User-Agent访问配置ID获取配置
func (dao *UseragentAccessConfigDAO) GetUseragentAccessConfigById(ctx context.Context, useragentAccessConfigId, tenantId string) (*models.UseragentAccessConfig, error) {
	if useragentAccessConfigId == "" || tenantId == "" {
		return nil, errors.New("useragentAccessConfigId和tenantId不能为空")
	}

	query := `
		SELECT * FROM HUB_GW_UA_ACCESS_CONFIG 
		WHERE useragentAccessConfigId = ? AND tenantId = ?
	`

	var config models.UseragentAccessConfig
	err := dao.db.QueryOne(ctx, &config, query, []interface{}{useragentAccessConfigId, tenantId}, true)

	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "查询User-Agent访问控制配置失败")
	}

	return &config, nil
}

// UpdateUseragentAccessConfig 更新User-Agent访问控制配置
func (dao *UseragentAccessConfigDAO) UpdateUseragentAccessConfig(ctx context.Context, config *models.UseragentAccessConfig, operatorId string) error {
	if config.SecurityConfigId == "" || config.TenantId == "" {
		return errors.New("securityConfigId和tenantId不能为空")
	}

	// 首先获取当前配置
	currentConfig, err := dao.GetUseragentAccessConfigBySecurityConfigId(ctx, config.SecurityConfigId, config.TenantId)
	if err != nil {
		return err
	}
	if currentConfig == nil {
		return errors.New("User-Agent访问控制配置不存在")
	}

	// 更新修改信息
	config.UseragentAccessConfigId = currentConfig.UseragentAccessConfigId
	config.EditTime = time.Now()
	config.EditWho = operatorId
	config.CurrentVersion = currentConfig.CurrentVersion + 1
	config.OprSeqFlag = config.UseragentAccessConfigId + "_" + strings.ReplaceAll(time.Now().String(), ".", "")[:8]
	
	// 保持活动标记不变，如果没有设置则使用当前值
	if config.ActiveFlag == "" {
		config.ActiveFlag = currentConfig.ActiveFlag
	}

	sql := `
		UPDATE HUB_GW_UA_ACCESS_CONFIG SET
			configName = ?, defaultPolicy = ?, whitelistPatterns = ?, blacklistPatterns = ?,
			blockEmptyUserAgent = ?, reserved1 = ?, reserved2 = ?, reserved3 = ?, reserved4 = ?, reserved5 = ?,
			extProperty = ?, noteText = ?, editTime = ?, editWho = ?, oprSeqFlag = ?, currentVersion = ?, activeFlag = ?
		WHERE useragentAccessConfigId = ? AND tenantId = ? AND currentVersion = ?
	`

	result, err := dao.db.Exec(ctx, sql, []interface{}{
		config.ConfigName, config.DefaultPolicy, config.WhitelistPatterns, config.BlacklistPatterns,
		config.BlockEmptyUserAgent, config.Reserved1, config.Reserved2, config.Reserved3, config.Reserved4, config.Reserved5,
		config.ExtProperty, config.NoteText, config.EditTime, config.EditWho, config.OprSeqFlag, config.CurrentVersion, config.ActiveFlag,
		config.UseragentAccessConfigId, config.TenantId, currentConfig.CurrentVersion,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "更新User-Agent访问控制配置失败")
	}

	if result == 0 {
		return errors.New("User-Agent访问控制配置数据已被其他用户修改，请刷新后重试")
	}

	return nil
}

// DeleteUseragentAccessConfig 删除User-Agent访问控制配置
func (dao *UseragentAccessConfigDAO) DeleteUseragentAccessConfig(ctx context.Context, securityConfigId, tenantId, operatorId string) error {
	if securityConfigId == "" || tenantId == "" {
		return errors.New("securityConfigId和tenantId不能为空")
	}

	sql := `
		DELETE FROM HUB_GW_UA_ACCESS_CONFIG 
		WHERE securityConfigId = ? AND tenantId = ?
	`

	result, err := dao.db.Exec(ctx, sql, []interface{}{securityConfigId, tenantId}, true)
	if err != nil {
		return huberrors.WrapError(err, "删除User-Agent访问控制配置失败")
	}

	if result == 0 {
		return errors.New("未找到要删除的User-Agent访问控制配置")
	}

	return nil
}

// ListUseragentAccessConfigs 获取User-Agent访问控制配置列表
func (dao *UseragentAccessConfigDAO) ListUseragentAccessConfigs(ctx context.Context, tenantId string, page, pageSize int) ([]*models.UseragentAccessConfig, int, error) {
	if tenantId == "" {
		return nil, 0, errors.New("tenantId不能为空")
	}

	// 构建基础查询语句
	baseQuery := "SELECT * FROM HUB_GW_UA_ACCESS_CONFIG WHERE tenantId = ? ORDER BY addTime DESC"

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
		return nil, 0, huberrors.WrapError(err, "查询User-Agent访问控制配置总数失败")
	}
	total := result.Count

	// 如果没有记录，直接返回空列表
	if total == 0 {
		return []*models.UseragentAccessConfig{}, 0, nil
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
	var configs []*models.UseragentAccessConfig
	err = dao.db.Query(ctx, &configs, paginatedQuery, allArgs, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询User-Agent访问控制配置列表失败")
	}

	return configs, total, nil
} 