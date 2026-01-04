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

// AddUseragentAccessConfig 添加User-Agent访问控制配置
func (dao *UseragentAccessConfigDAO) AddUseragentAccessConfig(ctx context.Context, config *models.UseragentAccessConfig, operatorId string) error {
	// 校验安全配置ID
	if config.SecurityConfigId == "" {
		return errors.New("securityConfigId不能为空")
	}

	// 自动生成User-Agent访问配置ID
	if config.UseragentAccessConfigId == "" {
		// 使用公共方法生成唯一ID，支持集群环境
		// 格式：UA前缀 + 32位唯一字符串 = 34位总长度
		config.UseragentAccessConfigId = random.GenerateUniqueStringWithPrefix("UA", 32)
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

// GetUseragentAccessConfigById 根据User-Agent访问配置ID获取配置
func (dao *UseragentAccessConfigDAO) GetUseragentAccessConfigById(ctx context.Context, useragentAccessConfigId, tenantId string) (*models.UseragentAccessConfig, error) {
	if useragentAccessConfigId == "" {
		return nil, errors.New("useragentAccessConfigId不能为空")
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
	// 校验主键
	if config.UseragentAccessConfigId == "" {
		return errors.New("useragentAccessConfigId不能为空")
	}

	// 首先获取当前配置（使用主键）
	currentConfig, err := dao.GetUseragentAccessConfigById(ctx, config.UseragentAccessConfigId, config.TenantId)
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

	// 保留不可修改的字段
	config.AddTime = currentConfig.AddTime
	config.AddWho = currentConfig.AddWho

	// 如果没有设置活动标记，保持原有状态
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

// DeleteUseragentAccessConfig 物理删除User-Agent访问控制配置
func (dao *UseragentAccessConfigDAO) DeleteUseragentAccessConfig(ctx context.Context, useragentAccessConfigId, tenantId string) error {
	// 校验主键
	if useragentAccessConfigId == "" {
		return errors.New("useragentAccessConfigId不能为空")
	}

	// 首先检查配置是否存在（使用主键）
	config, err := dao.GetUseragentAccessConfigById(ctx, useragentAccessConfigId, tenantId)
	if err != nil {
		return err
	}
	if config == nil {
		return errors.New("User-Agent访问控制配置不存在")
	}

	// 执行物理删除（使用主键）
	sql := `DELETE FROM HUB_GW_UA_ACCESS_CONFIG WHERE useragentAccessConfigId = ? AND tenantId = ?`

	result, err := dao.db.Exec(ctx, sql, []interface{}{useragentAccessConfigId, tenantId}, true)
	if err != nil {
		return huberrors.WrapError(err, "删除User-Agent访问控制配置失败")
	}

	if result == 0 {
		return errors.New("未找到要删除的User-Agent访问控制配置")
	}

	return nil
}

// ListUseragentAccessConfigs 获取User-Agent访问控制配置列表（支持条件查询）
func (dao *UseragentAccessConfigDAO) ListUseragentAccessConfigs(ctx context.Context, tenantId string, query *models.UseragentAccessConfigQuery, page, pageSize int) ([]*models.UseragentAccessConfig, int, error) {
	// 构建查询条件
	whereClause := "WHERE tenantId = ?"
	var params []interface{}
	params = append(params, tenantId)

	if query != nil {
		// securityConfigId 是必填条件（避免关联错误）
		if query.SecurityConfigId != "" {
			whereClause += " AND securityConfigId = ?"
			params = append(params, query.SecurityConfigId)
		}
		if query.ConfigName != "" {
			whereClause += " AND configName LIKE ?"
			params = append(params, "%"+query.ConfigName+"%")
		}
		if query.ActiveFlag != "" {
			whereClause += " AND activeFlag = ?"
			params = append(params, query.ActiveFlag)
		}
	}

	// 构建基础查询语句
	baseQuery := "SELECT * FROM HUB_GW_UA_ACCESS_CONFIG " + whereClause + " ORDER BY addTime DESC"

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
	allArgs := append(params, paginationArgs...)

	// 执行分页查询
	var configs []*models.UseragentAccessConfig
	err = dao.db.Query(ctx, &configs, paginatedQuery, allArgs, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询User-Agent访问控制配置列表失败")
	}

	return configs, total, nil
}
