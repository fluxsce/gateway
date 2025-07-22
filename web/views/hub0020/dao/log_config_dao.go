package dao

import (
	"context"
	"fmt"
	"time"

	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
	"gateway/pkg/logger"
	"gateway/pkg/utils/random"
	"gateway/web/views/hub0020/models"
)

// LogConfigDAO 日志配置数据访问对象
type LogConfigDAO struct {
	db database.Database
}

// NewLogConfigDAO 创建日志配置数据访问对象
func NewLogConfigDAO(db database.Database) *LogConfigDAO {
	return &LogConfigDAO{db: db}
}

// AddLogConfig 添加日志配置
func (dao *LogConfigDAO) AddLogConfig(ctx context.Context, logConfig *models.LogConfig, operatorId string) (string, error) {
	// 生成配置ID
	if logConfig.LogConfigId == "" {
		logConfig.LogConfigId = "LOG_" + random.GenerateRandomString(24)
	}

	// 设置默认值
	now := time.Now()
	logConfig.AddTime = now
	logConfig.EditTime = now
	logConfig.AddWho = operatorId
	logConfig.EditWho = operatorId
	logConfig.OprSeqFlag = fmt.Sprintf("LOG_CONFIG_%d", now.UnixNano())
	logConfig.CurrentVersion = 1
	logConfig.ActiveFlag = "Y"

	// 设置默认配置值
	if logConfig.LogFormat == "" {
		logConfig.LogFormat = "JSON"
	}
	if logConfig.RecordRequestBody == "" {
		logConfig.RecordRequestBody = "N"
	}
	if logConfig.RecordResponseBody == "" {
		logConfig.RecordResponseBody = "N"
	}
	if logConfig.RecordHeaders == "" {
		logConfig.RecordHeaders = "Y"
	}
	if logConfig.MaxBodySizeBytes == 0 {
		logConfig.MaxBodySizeBytes = 4096
	}
	if logConfig.OutputTargets == "" {
		logConfig.OutputTargets = "CONSOLE"
	}
	if logConfig.EnableAsyncLogging == "" {
		logConfig.EnableAsyncLogging = "Y"
	}
	if logConfig.AsyncQueueSize == 0 {
		logConfig.AsyncQueueSize = 10000
	}
	if logConfig.AsyncFlushIntervalMs == 0 {
		logConfig.AsyncFlushIntervalMs = 1000
	}
	if logConfig.EnableBatchProcessing == "" {
		logConfig.EnableBatchProcessing = "Y"
	}
	if logConfig.BatchSize == 0 {
		logConfig.BatchSize = 100
	}
	if logConfig.BatchTimeoutMs == 0 {
		logConfig.BatchTimeoutMs = 5000
	}
	if logConfig.LogRetentionDays == 0 {
		logConfig.LogRetentionDays = 30
	}
	if logConfig.EnableFileRotation == "" {
		logConfig.EnableFileRotation = "Y"
	}
	if logConfig.RotationPattern == "" {
		logConfig.RotationPattern = "DAILY"
	}
	if logConfig.EnableSensitiveDataMasking == "" {
		logConfig.EnableSensitiveDataMasking = "Y"
	}
	if logConfig.MaskingPattern == "" {
		logConfig.MaskingPattern = "***"
	}
	if logConfig.BufferSize == 0 {
		logConfig.BufferSize = 8192
	}
	if logConfig.FlushThreshold == 0 {
		logConfig.FlushThreshold = 100
	}

	// 插入数据库
	_, err := dao.db.Insert(ctx, logConfig.TableName(), logConfig, true)
	if err != nil {
		logger.ErrorWithTrace(ctx, "添加日志配置失败", err)
		return "", fmt.Errorf("添加日志配置失败: %w", err)
	}

	return logConfig.LogConfigId, nil
}

// UpdateLogConfig 更新日志配置
func (dao *LogConfigDAO) UpdateLogConfig(ctx context.Context, logConfig *models.LogConfig, operatorId string) error {
	// 更新操作信息
	logConfig.EditTime = time.Now()
	logConfig.EditWho = operatorId
	logConfig.OprSeqFlag = fmt.Sprintf("LOG_CONFIG_%d", time.Now().UnixNano())

	// 构建更新条件
	whereClause := "tenantId = ? AND logConfigId = ?"
	whereArgs := []interface{}{logConfig.TenantId, logConfig.LogConfigId}

	// 执行更新
	_, err := dao.db.Update(ctx, logConfig.TableName(), logConfig, whereClause, whereArgs, true)
	if err != nil {
		logger.ErrorWithTrace(ctx, "更新日志配置失败", err)
		return fmt.Errorf("更新日志配置失败: %w", err)
	}

	return nil
}

// GetLogConfigById 根据ID获取日志配置
func (dao *LogConfigDAO) GetLogConfigById(ctx context.Context, logConfigId, tenantId string) (*models.LogConfig, error) {
	logConfig := &models.LogConfig{}
	query := fmt.Sprintf("SELECT * FROM %s WHERE tenantId = ? AND logConfigId = ? AND activeFlag = 'Y'",
		logConfig.TableName())

	err := dao.db.QueryOne(ctx, logConfig, query, []interface{}{tenantId, logConfigId}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		logger.ErrorWithTrace(ctx, "查询日志配置失败", err)
		return nil, fmt.Errorf("查询日志配置失败: %w", err)
	}

	return logConfig, nil
}

// ListLogConfigs 获取日志配置列表
func (dao *LogConfigDAO) ListLogConfigs(ctx context.Context, tenantId string, page, pageSize int) ([]*models.LogConfig, int64, error) {
	tableName := (&models.LogConfig{}).TableName()

	// 构建基础查询
	baseQuery := fmt.Sprintf("SELECT * FROM %s WHERE tenantId = ? AND activeFlag = 'Y' ORDER BY configPriority ASC, addTime DESC", tableName)
	args := []interface{}{tenantId}

	// 构建统计查询
	countQuery, err := sqlutils.BuildCountQuery(baseQuery)
	if err != nil {
		logger.ErrorWithTrace(ctx, "构建统计查询失败", err)
		return nil, 0, fmt.Errorf("构建统计查询失败: %w", err)
	}

	// 执行统计查询
	var result struct {
		Count int64 `db:"COUNT(*)"`
	}
	err = dao.db.QueryOne(ctx, &result, countQuery, args, true)
	if err != nil {
		logger.ErrorWithTrace(ctx, "统计日志配置总数失败", err)
		return nil, 0, fmt.Errorf("统计日志配置总数失败: %w", err)
	}

	if result.Count == 0 {
		return []*models.LogConfig{}, 0, nil
	}

	// 创建分页信息
	paginationInfo := sqlutils.NewPaginationInfo(page, pageSize)
	// 获取数据库类型
	dbType := sqlutils.GetDatabaseType(dao.db)
	// 构建分页查询
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, paginationInfo)
	if err != nil {
		logger.ErrorWithTrace(ctx, "构建分页查询失败", err)
		return nil, 0, fmt.Errorf("构建分页查询失败: %w", err)
	}
	allArgs := append(args, paginationArgs...)

	// 执行分页查询
	var logConfigs []*models.LogConfig
	err = dao.db.Query(ctx, &logConfigs, paginatedQuery, allArgs, true)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询日志配置列表失败", err)
		return nil, 0, fmt.Errorf("查询日志配置列表失败: %w", err)
	}

	return logConfigs, result.Count, nil
}

// DeleteLogConfig 删除日志配置（逻辑删除）
func (dao *LogConfigDAO) DeleteLogConfig(ctx context.Context, logConfigId, tenantId, operatorId string) error {
	// 构建更新SQL
	now := time.Now()
	sql := fmt.Sprintf(`
		UPDATE %s SET 
			activeFlag = ?, editTime = ?, editWho = ?, oprSeqFlag = ?
		WHERE tenantId = ? AND logConfigId = ?
	`, (&models.LogConfig{}).TableName())

	// 执行更新
	result, err := dao.db.Exec(ctx, sql, []interface{}{
		"N", now, operatorId, fmt.Sprintf("LOG_CONFIG_%d", now.UnixNano()),
		tenantId, logConfigId,
	}, true)
	if err != nil {
		logger.ErrorWithTrace(ctx, "删除日志配置失败", err)
		return fmt.Errorf("删除日志配置失败: %w", err)
	}

	if result == 0 {
		return fmt.Errorf("未找到要删除的日志配置")
	}

	return nil
}

// GetLogConfigsByTenant 根据租户获取所有激活的日志配置
func (dao *LogConfigDAO) GetLogConfigsByTenant(ctx context.Context, tenantId string) ([]*models.LogConfig, error) {
	var logConfigs []*models.LogConfig
	tableName := (&models.LogConfig{}).TableName()

	query := fmt.Sprintf("SELECT * FROM %s WHERE tenantId = ? AND activeFlag = 'Y' ORDER BY configPriority ASC, addTime DESC",
		tableName)

	err := dao.db.Query(ctx, &logConfigs, query, []interface{}{tenantId}, true)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询租户日志配置失败", err)
		return nil, fmt.Errorf("查询租户日志配置失败: %w", err)
	}

	return logConfigs, nil
}

// CheckLogConfigExists 检查日志配置是否存在
func (dao *LogConfigDAO) CheckLogConfigExists(ctx context.Context, logConfigId, tenantId string) (bool, error) {
	tableName := (&models.LogConfig{}).TableName()
	countQuery := fmt.Sprintf("SELECT COUNT(*) as count FROM %s WHERE tenantId = ? AND logConfigId = ? AND activeFlag = 'Y'",
		tableName)

	var result struct {
		Count int `db:"count"`
	}
	err := dao.db.QueryOne(ctx, &result, countQuery, []interface{}{tenantId, logConfigId}, true)
	if err != nil {
		logger.ErrorWithTrace(ctx, "检查日志配置是否存在失败", err)
		return false, fmt.Errorf("检查日志配置是否存在失败: %w", err)
	}

	return result.Count > 0, nil
}

// GetDefaultLogConfig 获取默认的日志配置
func (dao *LogConfigDAO) GetDefaultLogConfig(ctx context.Context, tenantId string) (*models.LogConfig, error) {
	var logConfigs []*models.LogConfig
	tableName := (&models.LogConfig{}).TableName()

	// 获取优先级最高的配置（数值越小优先级越高）
	query := fmt.Sprintf("SELECT * FROM %s WHERE tenantId = ? AND activeFlag = 'Y' ORDER BY configPriority ASC, addTime DESC LIMIT 1",
		tableName)

	err := dao.db.Query(ctx, &logConfigs, query, []interface{}{tenantId}, true)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询默认日志配置失败", err)
		return nil, fmt.Errorf("查询默认日志配置失败: %w", err)
	}

	if len(logConfigs) == 0 {
		return nil, nil
	}

	return logConfigs[0], nil
}
