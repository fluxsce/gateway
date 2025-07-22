package dao

import (
	"context"
	"errors"
	"fmt"
	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
	"gateway/pkg/utils/huberrors"
	"gateway/pkg/utils/random"
	"gateway/web/views/hubcommon002/models"
	"strings"
	"time"
)

// ApiAccessConfigDAO API访问控制配置数据访问对象
type ApiAccessConfigDAO struct {
	db database.Database
}

// NewApiAccessConfigDAO 创建API访问控制配置DAO
func NewApiAccessConfigDAO(db database.Database) *ApiAccessConfigDAO {
	return &ApiAccessConfigDAO{
		db: db,
	}
}

// generateApiAccessConfigId 生成API访问配置ID
// 格式：API + YYYYMMDD + HHMMSS + 4位随机数
// 示例：API20240615143022A1B2
func (dao *ApiAccessConfigDAO) generateApiAccessConfigId() string {
	now := time.Now()
	timeStr := now.Format("20060102150405")
	randomStr := random.GenerateRandomString(4)
	return fmt.Sprintf("API%s%s", timeStr, randomStr)
}

// isApiAccessConfigIdExists 检查API访问配置ID是否已存在
func (dao *ApiAccessConfigDAO) isApiAccessConfigIdExists(ctx context.Context, apiAccessConfigId string) (bool, error) {
	query := `SELECT COUNT(*) as count FROM HUB_GW_API_ACCESS_CONFIG WHERE apiAccessConfigId = ?`

	var result struct {
		Count int `db:"count"`
	}

	err := dao.db.QueryOne(ctx, &result, query, []interface{}{apiAccessConfigId}, true)
	if err != nil {
		return false, err
	}

	return result.Count > 0, nil
}

// generateUniqueApiAccessConfigId 生成唯一的API访问配置ID
func (dao *ApiAccessConfigDAO) generateUniqueApiAccessConfigId(ctx context.Context) (string, error) {
	const maxAttempts = 10

	for attempt := 0; attempt < maxAttempts; attempt++ {
		apiAccessConfigId := dao.generateApiAccessConfigId()

		exists, err := dao.isApiAccessConfigIdExists(ctx, apiAccessConfigId)
		if err != nil {
			return "", huberrors.WrapError(err, "检查API访问配置ID是否存在失败")
		}

		if !exists {
			return apiAccessConfigId, nil
		}

		time.Sleep(time.Millisecond)
	}

	return "", errors.New("生成唯一API访问配置ID失败，已达到最大尝试次数")
}

// AddApiAccessConfig 添加API访问控制配置
func (dao *ApiAccessConfigDAO) AddApiAccessConfig(ctx context.Context, config *models.ApiAccessConfig, operatorId string) error {
	if config.TenantId == "" || config.SecurityConfigId == "" {
		return errors.New("tenantId和securityConfigId不能为空")
	}

	// 自动生成API访问配置ID
	if config.ApiAccessConfigId == "" {
		generatedId, err := dao.generateUniqueApiAccessConfigId(ctx)
		if err != nil {
			return huberrors.WrapError(err, "生成API访问配置ID失败")
		}
		config.ApiAccessConfigId = generatedId
	}

	// 设置自动填充字段
	now := time.Now()
	config.AddTime = now
	config.AddWho = operatorId
	config.EditTime = now
	config.EditWho = operatorId
	config.OprSeqFlag = config.ApiAccessConfigId + "_" + strings.ReplaceAll(time.Now().String(), ".", "")[:8]
	config.CurrentVersion = 1
	config.ActiveFlag = "Y"

	// 设置默认值
	if config.DefaultPolicy == "" {
		config.DefaultPolicy = "allow"
	}

	_, err := dao.db.Insert(ctx, "HUB_GW_API_ACCESS_CONFIG", config, true)
	if err != nil {
		return huberrors.WrapError(err, "添加API访问控制配置失败")
	}

	return nil
}

// GetApiAccessConfigBySecurityConfigId 根据安全配置ID获取API访问控制配置
func (dao *ApiAccessConfigDAO) GetApiAccessConfigBySecurityConfigId(ctx context.Context, securityConfigId, tenantId string) (*models.ApiAccessConfig, error) {
	if securityConfigId == "" || tenantId == "" {
		return nil, errors.New("securityConfigId和tenantId不能为空")
	}

	query := `
		SELECT * FROM HUB_GW_API_ACCESS_CONFIG 
		WHERE securityConfigId = ? AND tenantId = ?
	`

	var config models.ApiAccessConfig
	err := dao.db.QueryOne(ctx, &config, query, []interface{}{securityConfigId, tenantId}, true)

	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "查询API访问控制配置失败")
	}

	return &config, nil
}

// UpdateApiAccessConfig 更新API访问控制配置
func (dao *ApiAccessConfigDAO) UpdateApiAccessConfig(ctx context.Context, config *models.ApiAccessConfig, operatorId string) error {
	if config.SecurityConfigId == "" || config.TenantId == "" {
		return errors.New("securityConfigId和tenantId不能为空")
	}

	// 首先获取当前配置
	currentConfig, err := dao.GetApiAccessConfigBySecurityConfigId(ctx, config.SecurityConfigId, config.TenantId)
	if err != nil {
		return err
	}
	if currentConfig == nil {
		return errors.New("API访问控制配置不存在")
	}

	// 更新修改信息
	config.ApiAccessConfigId = currentConfig.ApiAccessConfigId
	config.EditTime = time.Now()
	config.EditWho = operatorId
	config.CurrentVersion = currentConfig.CurrentVersion + 1
	config.OprSeqFlag = config.ApiAccessConfigId + "_" + strings.ReplaceAll(time.Now().String(), ".", "")[:8]

	// 保留不可修改的字段
	config.AddTime = currentConfig.AddTime
	config.AddWho = currentConfig.AddWho

	// 如果没有设置活动标记，保持原有状态
	if config.ActiveFlag == "" {
		config.ActiveFlag = currentConfig.ActiveFlag
	}

	sql := `
		UPDATE HUB_GW_API_ACCESS_CONFIG SET
			configName = ?, defaultPolicy = ?, whitelistPaths = ?, blacklistPaths = ?,
			allowedMethods = ?, blockedMethods = ?, reserved1 = ?, reserved2 = ?, reserved3 = ?, reserved4 = ?, reserved5 = ?,
			extProperty = ?, noteText = ?, editTime = ?, editWho = ?, oprSeqFlag = ?, currentVersion = ?, activeFlag = ?
		WHERE apiAccessConfigId = ? AND tenantId = ? AND currentVersion = ?
	`

	result, err := dao.db.Exec(ctx, sql, []interface{}{
		config.ConfigName, config.DefaultPolicy, config.WhitelistPaths, config.BlacklistPaths,
		config.AllowedMethods, config.BlockedMethods, config.Reserved1, config.Reserved2, config.Reserved3, config.Reserved4, config.Reserved5,
		config.ExtProperty, config.NoteText, config.EditTime, config.EditWho, config.OprSeqFlag, config.CurrentVersion, config.ActiveFlag,
		config.ApiAccessConfigId, config.TenantId, currentConfig.CurrentVersion,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "更新API访问控制配置失败")
	}

	if result == 0 {
		return errors.New("API访问控制配置数据已被其他用户修改，请刷新后重试")
	}

	return nil
}

// DeleteApiAccessConfig 删除API访问控制配置
func (dao *ApiAccessConfigDAO) DeleteApiAccessConfig(ctx context.Context, securityConfigId, tenantId, operatorId string) error {
	if securityConfigId == "" || tenantId == "" {
		return errors.New("securityConfigId和tenantId不能为空")
	}

	sql := `
		DELETE FROM HUB_GW_API_ACCESS_CONFIG
		WHERE securityConfigId = ? AND tenantId = ?
	`

	result, err := dao.db.Exec(ctx, sql, []interface{}{securityConfigId, tenantId}, true)
	if err != nil {
		return huberrors.WrapError(err, "删除API访问控制配置失败")
	}

	if result == 0 {
		return errors.New("未找到要删除的API访问控制配置")
	}

	return nil
}

// ListApiAccessConfigs 获取API访问控制配置列表
func (dao *ApiAccessConfigDAO) ListApiAccessConfigs(ctx context.Context, tenantId string, page, pageSize int) ([]*models.ApiAccessConfig, int, error) {
	if tenantId == "" {
		return nil, 0, errors.New("tenantId不能为空")
	}

	// 构建基础查询语句
	baseQuery := "SELECT * FROM HUB_GW_API_ACCESS_CONFIG WHERE tenantId = ? ORDER BY addTime DESC"

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
		return nil, 0, huberrors.WrapError(err, "查询API访问控制配置总数失败")
	}
	total := result.Count

	// 如果没有记录，直接返回空列表
	if total == 0 {
		return []*models.ApiAccessConfig{}, 0, nil
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
	var configs []*models.ApiAccessConfig
	err = dao.db.Query(ctx, &configs, paginatedQuery, allArgs, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询API访问控制配置列表失败")
	}

	return configs, total, nil
}
