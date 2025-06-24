package dao

import (
	"context"
	"errors"
	"fmt"
	"gohub/pkg/database"
	"gohub/pkg/utils/huberrors"
	"gohub/pkg/utils/random"
	"gohub/web/views/hubcommon002/models"
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

// generateRateLimitConfigId 生成限流配置ID
// 格式：RATE + YYYYMMDD + HHMMSS + 4位随机数
// 示例：RATE20240615143022A1B2
func (dao *RateLimitConfigDAO) generateRateLimitConfigId() string {
	now := time.Now()
	// 生成时间部分：YYYYMMDDHHMMSS
	timeStr := now.Format("20060102150405")
	
	// 生成4位随机字符（大写字母和数字）
	randomStr := random.GenerateRandomString(4)
	
	return fmt.Sprintf("RATE%s%s", timeStr, randomStr)
}

// isRateLimitConfigIdExists 检查限流配置ID是否已存在
func (dao *RateLimitConfigDAO) isRateLimitConfigIdExists(ctx context.Context, rateLimitConfigId string) (bool, error) {
	query := `SELECT COUNT(*) as count FROM HUB_GATEWAY_RATE_LIMIT_CONFIG WHERE rateLimitConfigId = ?`
	
	var result struct {
		Count int `db:"count"`
	}
	
	err := dao.db.QueryOne(ctx, &result, query, []interface{}{rateLimitConfigId}, true)
	if err != nil {
		return false, err
	}
	
	return result.Count > 0, nil
}

// generateUniqueRateLimitConfigId 生成唯一的限流配置ID
func (dao *RateLimitConfigDAO) generateUniqueRateLimitConfigId(ctx context.Context) (string, error) {
	const maxAttempts = 10
	
	for attempt := 0; attempt < maxAttempts; attempt++ {
		rateLimitConfigId := dao.generateRateLimitConfigId()
		
		exists, err := dao.isRateLimitConfigIdExists(ctx, rateLimitConfigId)
		if err != nil {
			return "", huberrors.WrapError(err, "检查限流配置ID是否存在失败")
		}
		
		if !exists {
			return rateLimitConfigId, nil
		}
		
		// 如果ID已存在，等待1毫秒后重试（确保时间戳不同）
		time.Sleep(time.Millisecond)
	}
	
	return "", errors.New("生成唯一限流配置ID失败，已达到最大尝试次数")
}

// AddRateLimitConfig 添加限流配置
func (dao *RateLimitConfigDAO) AddRateLimitConfig(ctx context.Context, config *models.RateLimitConfig, operatorId string) error {
	if config.TenantId == "" || config.LimitName == "" {
		return errors.New("tenantId和limitName不能为空")
	}

	// 自动生成限流配置ID
	if config.RateLimitConfigId == "" {
		generatedId, err := dao.generateUniqueRateLimitConfigId(ctx)
		if err != nil {
			return huberrors.WrapError(err, "生成限流配置ID失败")
		}
		config.RateLimitConfigId = generatedId
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

	_, err := dao.db.Insert(ctx, "HUB_GATEWAY_RATE_LIMIT_CONFIG", config, true)
	if err != nil {
		return huberrors.WrapError(err, "添加限流配置失败")
	}

	return nil
}

// GetRateLimitConfig 根据租户ID和配置ID获取限流配置
func (dao *RateLimitConfigDAO) GetRateLimitConfig(tenantId, rateLimitConfigId string) (*models.RateLimitConfig, error) {
	if rateLimitConfigId == "" || tenantId == "" {
		return nil, errors.New("rateLimitConfigId和tenantId不能为空")
	}

	query := `
		SELECT * FROM HUB_GATEWAY_RATE_LIMIT_CONFIG 
		WHERE rateLimitConfigId = ? AND tenantId = ? AND activeFlag = 'Y'
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
	if config.RateLimitConfigId == "" || config.TenantId == "" {
		return errors.New("rateLimitConfigId和tenantId不能为空")
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

	sql := `
		UPDATE HUB_GATEWAY_RATE_LIMIT_CONFIG SET
			gatewayInstanceId = ?, routeConfigId = ?, limitName = ?, algorithm = ?, keyStrategy = ?,
			limitRate = ?, burstCapacity = ?, timeWindowSeconds = ?, rejectionStatusCode = ?, 
			rejectionMessage = ?, configPriority = ?, customConfig = ?,
			reserved1 = ?, reserved2 = ?, reserved3 = ?, reserved4 = ?, reserved5 = ?,
			extProperty = ?, noteText = ?, editTime = ?, editWho = ?, currentVersion = ?, oprSeqFlag = ?
		WHERE rateLimitConfigId = ? AND tenantId = ? AND activeFlag = 'Y'
	`

	_, err = dao.db.Exec(ctx, sql, []interface{}{
		config.GatewayInstanceId, config.RouteConfigId, config.LimitName, config.Algorithm, config.KeyStrategy,
		config.LimitRate, config.BurstCapacity, config.TimeWindowSeconds, config.RejectionStatusCode, 
		config.RejectionMessage, config.ConfigPriority, config.CustomConfig,
		config.Reserved1, config.Reserved2, config.Reserved3, config.Reserved4, config.Reserved5,
		config.ExtProperty, config.NoteText, config.EditTime, config.EditWho, config.CurrentVersion, config.OprSeqFlag,
		config.RateLimitConfigId, config.TenantId,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "更新限流配置失败")
	}

	return nil
}

// DeleteRateLimitConfig 软删除限流配置
func (dao *RateLimitConfigDAO) DeleteRateLimitConfig(tenantId, rateLimitConfigId, operatorId string) error {
	if rateLimitConfigId == "" || tenantId == "" {
		return errors.New("rateLimitConfigId和tenantId不能为空")
	}

	// 首先检查配置是否存在
	config, err := dao.GetRateLimitConfig(tenantId, rateLimitConfigId)
	if err != nil {
		return err
	}
	if config == nil {
		return errors.New("限流配置不存在")
	}

	now := time.Now()
	sql := `
		UPDATE HUB_GATEWAY_RATE_LIMIT_CONFIG SET
			activeFlag = 'N', editTime = ?, editWho = ?
		WHERE rateLimitConfigId = ? AND tenantId = ? AND activeFlag = 'Y'
	`

	_, err = dao.db.Exec(context.Background(), sql, []interface{}{
		now, operatorId, rateLimitConfigId, tenantId,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "删除限流配置失败")
	}

	return nil
}

// ListRateLimitConfigs 查询限流配置列表
func (dao *RateLimitConfigDAO) ListRateLimitConfigs(ctx context.Context, tenantId string, page, pageSize int) ([]*models.RateLimitConfig, int, error) {
	if tenantId == "" {
		return nil, 0, errors.New("tenantId不能为空")
	}

	// 设置默认分页参数
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	// 计算偏移量
	offset := (page - 1) * pageSize

	// 查询总数
	countQuery := `SELECT COUNT(*) as total FROM HUB_GATEWAY_RATE_LIMIT_CONFIG WHERE tenantId = ? AND activeFlag = 'Y'`
	var totalResult struct {
		Total int `db:"total"`
	}
	err := dao.db.QueryOne(ctx, &totalResult, countQuery, []interface{}{tenantId}, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询限流配置总数失败")
	}

	// 查询配置列表
	listQuery := `
		SELECT * FROM HUB_GATEWAY_RATE_LIMIT_CONFIG 
		WHERE tenantId = ? AND activeFlag = 'Y'
		ORDER BY configPriority ASC, addTime DESC
		LIMIT ? OFFSET ?
	`

	var configs []*models.RateLimitConfig
	err = dao.db.Query(ctx, &configs, listQuery, []interface{}{tenantId, pageSize, offset}, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询限流配置列表失败")
	}

	return configs, totalResult.Total, nil
}

// GetRateLimitConfigByGatewayInstance 根据网关实例ID查询单个限流配置
func (dao *RateLimitConfigDAO) GetRateLimitConfigByGatewayInstance(tenantId, gatewayInstanceId string) (*models.RateLimitConfig, error) {
	if tenantId == "" || gatewayInstanceId == "" {
		return nil, errors.New("tenantId和gatewayInstanceId不能为空")
	}

	sql := `
		SELECT * FROM HUB_GATEWAY_RATE_LIMIT_CONFIG 
		WHERE tenantId = ? AND gatewayInstanceId = ? AND activeFlag = 'Y'
		ORDER BY configPriority ASC, addTime DESC
		LIMIT 1
	`

	var config models.RateLimitConfig
	err := dao.db.QueryOne(context.Background(), &config, sql, []interface{}{tenantId, gatewayInstanceId}, true)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil // 未找到配置，返回nil而不是错误
		}
		return nil, huberrors.WrapError(err, "查询网关实例限流配置失败")
	}

	return &config, nil
}

// GetRateLimitConfigByRouteConfig 根据路由配置ID查询单个限流配置
func (dao *RateLimitConfigDAO) GetRateLimitConfigByRouteConfig(tenantId, routeConfigId string) (*models.RateLimitConfig, error) {
	if tenantId == "" || routeConfigId == "" {
		return nil, errors.New("tenantId和routeConfigId不能为空")
	}

	sql := `
		SELECT * FROM HUB_GATEWAY_RATE_LIMIT_CONFIG 
		WHERE tenantId = ? AND routeConfigId = ? AND activeFlag = 'Y'
		ORDER BY configPriority ASC, addTime DESC
		LIMIT 1
	`

	var config models.RateLimitConfig
	err := dao.db.QueryOne(context.Background(), &config, sql, []interface{}{tenantId, routeConfigId}, true)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil // 未找到配置，返回nil而不是错误
		}
		return nil, huberrors.WrapError(err, "查询路由配置限流配置失败")
	}

	return &config, nil
}

// GetRateLimitConfigsByKeyStrategy 根据键策略查询限流配置列表
func (dao *RateLimitConfigDAO) GetRateLimitConfigsByKeyStrategy(ctx context.Context, tenantId, keyStrategy string) ([]*models.RateLimitConfig, error) {
	if tenantId == "" || keyStrategy == "" {
		return nil, errors.New("tenantId和keyStrategy不能为空")
	}

	sql := `
		SELECT * FROM HUB_GATEWAY_RATE_LIMIT_CONFIG 
		WHERE tenantId = ? AND keyStrategy = ? AND activeFlag = 'Y'
		ORDER BY configPriority ASC, addTime DESC
	`

	var configs []*models.RateLimitConfig
	err := dao.db.Query(ctx, &configs, sql, []interface{}{tenantId, keyStrategy}, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "根据键策略查询限流配置失败")
	}

	return configs, nil
}

// GetRateLimitConfigsByAlgorithm 根据算法类型查询限流配置列表
func (dao *RateLimitConfigDAO) GetRateLimitConfigsByAlgorithm(ctx context.Context, tenantId, algorithm string) ([]*models.RateLimitConfig, error) {
	if tenantId == "" || algorithm == "" {
		return nil, errors.New("tenantId和algorithm不能为空")
	}

	sql := `
		SELECT * FROM HUB_GATEWAY_RATE_LIMIT_CONFIG 
		WHERE tenantId = ? AND algorithm = ? AND activeFlag = 'Y'
		ORDER BY configPriority ASC, addTime DESC
	`

	var configs []*models.RateLimitConfig
	err := dao.db.Query(ctx, &configs, sql, []interface{}{tenantId, algorithm}, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "根据算法类型查询限流配置失败")
	}

	return configs, nil
} 