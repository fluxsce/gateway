package dbloader

import (
	"context"
	"fmt"

	"gateway/internal/gateway/logwrite/types"
	"gateway/pkg/database"
)

// LogConfigLoader 日志配置加载器
type LogConfigLoader struct {
	db       database.Database
	tenantId string
}

// NewLogConfigLoader 创建日志配置加载器
func NewLogConfigLoader(db database.Database, tenantId string) *LogConfigLoader {
	return &LogConfigLoader{
		db:       db,
		tenantId: tenantId,
	}
}

// LoadLogConfig 加载指定ID的日志配置
func (loader *LogConfigLoader) LoadLogConfig(ctx context.Context, logConfigId string) (*types.LogConfig, error) {
	if logConfigId == "" {
		return nil, nil
	}

	query := `
		SELECT tenantId, logConfigId, configName, configDesc, logFormat,
		       recordRequestBody, recordResponseBody, recordHeaders, maxBodySizeBytes,
		       outputTargets, fileConfig, databaseConfig, mongoConfig, 
		       elasticsearchConfig, clickhouseConfig, enableAsyncLogging,
		       asyncQueueSize, asyncFlushIntervalMs, enableBatchProcessing,
		       batchSize, batchTimeoutMs, logRetentionDays, enableFileRotation,
		       maxFileSizeMB, maxFileCount, rotationPattern, enableSensitiveDataMasking,
		       sensitiveFields, maskingPattern, bufferSize, flushThreshold,
		       configPriority, activeFlag, extProperty
		FROM HUB_GW_LOG_CONFIG
		WHERE tenantId = ? AND logConfigId = ? AND activeFlag = 'Y'
	`

	var record LogConfigRecord
	err := loader.db.QueryOne(ctx, &record, query, []interface{}{loader.tenantId, logConfigId}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询日志配置失败: %w", err)
	}

	// 转换数据库记录为日志配置对象
	return loader.buildLogConfig(&record), nil
}

// buildLogConfig 构建日志配置对象
func (loader *LogConfigLoader) buildLogConfig(record *LogConfigRecord) *types.LogConfig {
	config := &types.LogConfig{
		TenantID:                   record.TenantId,
		LogConfigID:                record.LogConfigId,
		ConfigName:                 record.ConfigName,
		ConfigDesc:                 record.ConfigDesc,
		LogFormat:                  record.LogFormat,
		RecordRequestBody:          record.RecordRequestBody,
		RecordResponseBody:         record.RecordResponseBody,
		RecordHeaders:              record.RecordHeaders,
		MaxBodySizeBytes:           record.MaxBodySizeBytes,
		OutputTargets:              record.OutputTargets,
		FileConfig:                 record.FileConfig,
		DatabaseConfig:             record.DatabaseConfig,
		MongoConfig:                record.MongoConfig,
		ElasticsearchConfig:        record.ElasticsearchConfig,
		ClickHouseConfig:           record.ClickhouseConfig,
		EnableAsyncLogging:         record.EnableAsyncLogging,
		AsyncQueueSize:             record.AsyncQueueSize,
		AsyncFlushIntervalMs:       record.AsyncFlushIntervalMs,
		EnableBatchProcessing:      record.EnableBatchProcessing,
		BatchSize:                  record.BatchSize,
		BatchTimeoutMs:             record.BatchTimeoutMs,
		LogRetentionDays:           record.LogRetentionDays,
		EnableFileRotation:         record.EnableFileRotation,
		RotationPattern:            record.RotationPattern,
		EnableSensitiveDataMasking: record.EnableSensitiveDataMasking,
		SensitiveFields:            record.SensitiveFields,
		MaskingPattern:             record.MaskingPattern,
		BufferSize:                 record.BufferSize,
		FlushThreshold:             record.FlushThreshold,
		ConfigPriority:             record.ConfigPriority,
		ActiveFlag:                 record.ActiveFlag,
		ExtProperty:                record.ExtProperty,
	}

	// 处理可空字段
	if record.MaxFileSizeMB != nil {
		config.MaxFileSizeMB = *record.MaxFileSizeMB
	}
	if record.MaxFileCount != nil {
		config.MaxFileCount = *record.MaxFileCount
	}

	// 设置默认值
	config.SetDefaults()

	// 预解析 extProperty 中的告警配置（构建时解析一次，避免后续重复解析）
	alertCfg := types.ParseAlertConfigFromExtProperty(config.ExtProperty)
	config.SetAlertConfig(alertCfg)

	// 预解析 extProperty 中的清理配置（构建时解析一次，避免后续重复解析）
	cleanupCfg := types.ParseCleanupConfigFromExtProperty(config.ExtProperty)
	config.SetCleanupConfig(cleanupCfg)

	return config
}

// ValidateLogConfig 验证日志配置
func (loader *LogConfigLoader) ValidateLogConfig(config *types.LogConfig) error {
	if config == nil {
		return fmt.Errorf("日志配置不能为空")
	}

	return config.Validate()
}
