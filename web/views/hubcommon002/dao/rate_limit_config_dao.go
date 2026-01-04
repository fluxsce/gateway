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

// RateLimitConfigDAO 限流配置数据访问对象
type RateLimitConfigDAO struct {
	db database.Database
}

// NewRateLimitConfigDAO 创建限流配置DAO
func NewRateLimitConfigDAO(db database.Database) *RateLimitConfigDAO {
	return &RateLimitConfigDAO{
		db: db,
	}
}

// AddRateLimitConfig 添加限流配置
func (dao *RateLimitConfigDAO) AddRateLimitConfig(ctx context.Context, config *models.RateLimitConfig, operatorId string) error {
	if config.LimitName == "" {
		return errors.New("limitName不能为空")
	}

	// 自动生成限流配置ID
	if config.RateLimitConfigId == "" {
		// 使用公共方法生成唯一ID，支持集群环境
		// 格式：RATE前缀 + 32位唯一字符串 = 36位总长度
		config.RateLimitConfigId = random.GenerateUniqueStringWithPrefix("RATE", 32)
	}

	// 设置自动填充字段
	now := time.Now()
	config.AddTime = now
	config.AddWho = operatorId
	config.EditTime = now
	config.EditWho = operatorId
	config.OprSeqFlag = config.RateLimitConfigId + "_" + strings.ReplaceAll(time.Now().String(), ".", "")[:8]
	config.CurrentVersion = 1
	config.ActiveFlag = "Y"

	// 设置默认值
	if config.Algorithm == "" {
		config.Algorithm = "token-bucket"
	}
	if config.KeyStrategy == "" {
		config.KeyStrategy = "ip"
	}
	if config.RejectionStatusCode == 0 {
		config.RejectionStatusCode = 429
	}
	if config.RejectionMessage == "" {
		config.RejectionMessage = "请求过于频繁，请稍后再试"
	}
	if config.TimeWindowSeconds == 0 {
		config.TimeWindowSeconds = 1
	}
	if config.CustomConfig == "" {
		config.CustomConfig = "{}"
	}

	_, err := dao.db.Insert(ctx, "HUB_GW_RATE_LIMIT_CONFIG", config, true)
	if err != nil {
		return huberrors.WrapError(err, "添加限流配置失败")
	}

	return nil
}

// GetRateLimitConfig 根据租户ID和配置ID获取限流配置
func (dao *RateLimitConfigDAO) GetRateLimitConfig(tenantId, rateLimitConfigId string) (*models.RateLimitConfig, error) {
	if rateLimitConfigId == "" {
		return nil, errors.New("rateLimitConfigId不能为空")
	}

	query := `
		SELECT * FROM HUB_GW_RATE_LIMIT_CONFIG 
		WHERE rateLimitConfigId = ? AND tenantId = ?
	`

	var config models.RateLimitConfig
	err := dao.db.QueryOne(context.Background(), &config, query, []interface{}{rateLimitConfigId, tenantId}, true)

	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "查询限流配置失败")
	}

	return &config, nil
}

// UpdateRateLimitConfig 更新限流配置
func (dao *RateLimitConfigDAO) UpdateRateLimitConfig(ctx context.Context, config *models.RateLimitConfig, operatorId string) error {
	if config.RateLimitConfigId == "" {
		return errors.New("rateLimitConfigId不能为空")
	}

	// 首先获取当前配置
	currentConfig, err := dao.GetRateLimitConfig(config.TenantId, config.RateLimitConfigId)
	if err != nil {
		return err
	}
	if currentConfig == nil {
		return errors.New("限流配置不存在")
	}

	// 更新修改信息
	config.EditTime = time.Now()
	config.EditWho = operatorId
	config.CurrentVersion = currentConfig.CurrentVersion + 1
	config.OprSeqFlag = config.RateLimitConfigId + "_" + strings.ReplaceAll(time.Now().String(), ".", "")[:8]

	// 保留不可修改的字段
	config.AddTime = currentConfig.AddTime
	config.AddWho = currentConfig.AddWho

	// 如果没有设置活动标记，保持原有状态
	if config.ActiveFlag == "" {
		config.ActiveFlag = currentConfig.ActiveFlag
	}

	sql := `
		UPDATE HUB_GW_RATE_LIMIT_CONFIG SET
			gatewayInstanceId = ?, routeConfigId = ?, limitName = ?, algorithm = ?, keyStrategy = ?,
			limitRate = ?, burstCapacity = ?, timeWindowSeconds = ?, rejectionStatusCode = ?, 
			rejectionMessage = ?, configPriority = ?, customConfig = ?,
			reserved1 = ?, reserved2 = ?, reserved3 = ?, reserved4 = ?, reserved5 = ?,
			extProperty = ?, noteText = ?, editTime = ?, editWho = ?, currentVersion = ?, oprSeqFlag = ?, activeFlag = ?
		WHERE rateLimitConfigId = ? AND tenantId = ? AND currentVersion = ?
	`

	result, err := dao.db.Exec(ctx, sql, []interface{}{
		config.GatewayInstanceId, config.RouteConfigId, config.LimitName, config.Algorithm, config.KeyStrategy,
		config.LimitRate, config.BurstCapacity, config.TimeWindowSeconds, config.RejectionStatusCode,
		config.RejectionMessage, config.ConfigPriority, config.CustomConfig,
		config.Reserved1, config.Reserved2, config.Reserved3, config.Reserved4, config.Reserved5,
		config.ExtProperty, config.NoteText, config.EditTime, config.EditWho, config.CurrentVersion, config.OprSeqFlag, config.ActiveFlag,
		config.RateLimitConfigId, config.TenantId, currentConfig.CurrentVersion,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "更新限流配置失败")
	}

	if result == 0 {
		return errors.New("限流配置数据已被其他用户修改，请刷新后重试")
	}

	return nil
}

// DeleteRateLimitConfig 物理删除限流配置
func (dao *RateLimitConfigDAO) DeleteRateLimitConfig(tenantId, rateLimitConfigId string) error {
	// 校验主键
	if rateLimitConfigId == "" {
		return errors.New("rateLimitConfigId不能为空")
	}

	// 执行物理删除
	sql := `DELETE FROM HUB_GW_RATE_LIMIT_CONFIG WHERE rateLimitConfigId = ? AND tenantId = ?`

	result, err := dao.db.Exec(context.Background(), sql, []interface{}{rateLimitConfigId, tenantId}, true)
	if err != nil {
		return huberrors.WrapError(err, "删除限流配置失败")
	}

	if result == 0 {
		return errors.New("未找到要删除的限流配置")
	}

	return nil
}

// ListRateLimitConfigs 查询限流配置列表
func (dao *RateLimitConfigDAO) ListRateLimitConfigs(ctx context.Context, tenantId string, page, pageSize int) ([]*models.RateLimitConfig, int, error) {

	// 构建基础查询语句
	baseQuery := "SELECT * FROM HUB_GW_RATE_LIMIT_CONFIG WHERE tenantId = ? ORDER BY configPriority ASC, addTime DESC"

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
		return nil, 0, huberrors.WrapError(err, "查询限流配置总数失败")
	}

	// 如果没有记录，直接返回空列表
	if totalResult.Total == 0 {
		return []*models.RateLimitConfig{}, 0, nil
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
	var configs []*models.RateLimitConfig
	err = dao.db.Query(ctx, &configs, paginatedQuery, allArgs, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询限流配置列表失败")
	}

	return configs, totalResult.Total, nil
}

// GetRateLimitConfigByGatewayInstance 根据网关实例ID查询单个限流配置
func (dao *RateLimitConfigDAO) GetRateLimitConfigByGatewayInstance(tenantId, gatewayInstanceId string) (*models.RateLimitConfig, error) {
	if gatewayInstanceId == "" {
		return nil, errors.New("gatewayInstanceId不能为空")
	}

	// 构建查询语句（使用 LIMIT 1 只取第一条记录）
	query := `
		SELECT * FROM HUB_GW_RATE_LIMIT_CONFIG 
		WHERE tenantId = ? AND gatewayInstanceId = ?
		ORDER BY configPriority ASC, addTime DESC
		LIMIT 1
	`

	var config models.RateLimitConfig
	err := dao.db.QueryOne(context.Background(), &config, query, []interface{}{tenantId, gatewayInstanceId}, true)

	if err != nil {
		if err == database.ErrRecordNotFound {
			// 没有数据返回空，不报错
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "查询网关实例限流配置失败")
	}

	return &config, nil
}

// GetRateLimitConfigByRouteConfig 根据路由配置ID查询单个限流配置
func (dao *RateLimitConfigDAO) GetRateLimitConfigByRouteConfig(tenantId, routeConfigId string) (*models.RateLimitConfig, error) {
	if routeConfigId == "" {
		return nil, errors.New("routeConfigId不能为空")
	}

	// 构建查询语句（使用 LIMIT 1 只取第一条记录）
	query := `
		SELECT * FROM HUB_GW_RATE_LIMIT_CONFIG 
		WHERE tenantId = ? AND routeConfigId = ?
		ORDER BY configPriority ASC, addTime DESC
		LIMIT 1
	`

	var config models.RateLimitConfig
	err := dao.db.QueryOne(context.Background(), &config, query, []interface{}{tenantId, routeConfigId}, true)

	if err != nil {
		if err == database.ErrRecordNotFound {
			// 没有数据返回空，不报错
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "查询路由配置限流配置失败")
	}

	return &config, nil
}
