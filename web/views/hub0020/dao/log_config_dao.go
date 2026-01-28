package dao

import (
	"context"
	"fmt"
	"time"

	"gateway/pkg/database"
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
		logConfig.LogConfigId = random.GenerateUniqueStringWithPrefix("LOG_", 32)
	}

	// 设置默认值
	now := time.Now()
	logConfig.AddTime = now
	logConfig.EditTime = now
	logConfig.AddWho = operatorId
	logConfig.EditWho = operatorId
	// 生成 OprSeqFlag，确保长度不超过32
	logConfig.OprSeqFlag = random.GenerateUniqueStringWithPrefix("", 32)
	logConfig.CurrentVersion = 1
	logConfig.ActiveFlag = "Y"

	// 设置默认配置值（与前端默认值保持一致）
	if logConfig.ConfigName == "" {
		logConfig.ConfigName = "网关日志"
	}
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
		logConfig.MaxBodySizeBytes = 1048576 // 1MB，与前端保持一致
	}
	if logConfig.OutputTargets == "" {
		logConfig.OutputTargets = "DATABASE" // 与前端保持一致
	}
	if logConfig.EnableAsyncLogging == "" {
		logConfig.EnableAsyncLogging = "Y"
	}
	if logConfig.AsyncQueueSize == 0 {
		logConfig.AsyncQueueSize = 1000 // 与前端保持一致
	}
	if logConfig.AsyncFlushIntervalMs == 0 {
		logConfig.AsyncFlushIntervalMs = 5000 // 与前端保持一致
	}
	if logConfig.EnableBatchProcessing == "" {
		logConfig.EnableBatchProcessing = "Y"
	}
	if logConfig.BatchSize == 0 {
		logConfig.BatchSize = 100
	}
	if logConfig.BatchTimeoutMs == 0 {
		logConfig.BatchTimeoutMs = 1000 // 与前端保持一致
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
		logConfig.EnableSensitiveDataMasking = "N" // 与前端保持一致
	}
	if logConfig.MaskingPattern == "" {
		logConfig.MaskingPattern = "****" // 与前端保持一致
	}
	if logConfig.BufferSize == 0 {
		logConfig.BufferSize = 65536 // 64KB，与前端保持一致
	}
	if logConfig.FlushThreshold == 0 {
		logConfig.FlushThreshold = 1000 // 与前端保持一致
	}
	if logConfig.ConfigPriority == 0 {
		logConfig.ConfigPriority = 0
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
	// 生成 OprSeqFlag，确保长度不超过32
	logConfig.OprSeqFlag = random.GenerateUniqueStringWithPrefix("", 32)

	// 构建更新条件
	whereClause := "tenantId = ? AND logConfigId = ?"
	whereArgs := []interface{}{logConfig.TenantId, logConfig.LogConfigId}

	// 执行更新
	_, err := dao.db.Update(ctx, logConfig.TableName(), logConfig, whereClause, whereArgs, true, true)
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
