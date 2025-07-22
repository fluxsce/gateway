package logwrite

import (
	"fmt"

	"gohub/internal/gateway/logwrite/clickhouse"
	"gohub/internal/gateway/logwrite/console"
	"gohub/internal/gateway/logwrite/dbwrite"
	"gohub/internal/gateway/logwrite/filewrite"
	"gohub/internal/gateway/logwrite/mongowrite"
	"gohub/internal/gateway/logwrite/types"
)

// CreateWriter 根据配置创建写入器实例（静态方法）
func CreateWriter(config *types.LogConfig) (LogWriter, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	
	// 获取输出目标类型（只有一种）
	targets := config.GetOutputTargets()
	if len(targets) == 0 {
		return nil, fmt.Errorf("no output target specified")
	}
	
	target := targets[0] // 只取第一个，因为只会有一种类型
	
	// 创建写入器实例
	switch target {
	case types.LogOutputConsole:
		return createConsoleWriter(config)
	case types.LogOutputFile:
		return createFileWriter(config)
	case types.LogOutputDatabase:
		return createDatabaseWriter(config)
	case types.LogOutputMongoDB:
		return createMongoWriter(config)
	case types.LogOutputElasticsearch:
		return createElasticsearchWriter(config)
	case types.LogOutputClickHouse:
		return createClickHouseWriter(config)
	default:
		return nil, fmt.Errorf("unsupported output target: %s", string(target))
	}
}

// createConsoleWriter 创建控制台写入器
func createConsoleWriter(config *types.LogConfig) (LogWriter, error) {
	writer, err := console.NewConsoleWriter(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create console writer: %w", err)
	}
	return writer, nil
}

// createFileWriter 创建文件写入器
func createFileWriter(config *types.LogConfig) (LogWriter, error) {
	writer, err := filewrite.NewFileWriter(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create file writer: %w", err)
	}
	return writer, nil
}

// createDatabaseWriter 创建数据库写入器
func createDatabaseWriter(config *types.LogConfig) (LogWriter, error) {
	writer, err := dbwrite.NewDBWriter(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create database writer: %w", err)
	}
	return writer, nil
}

// createMongoWriter 创建MongoDB写入器
func createMongoWriter(config *types.LogConfig) (LogWriter, error) {
	writer, err := mongowrite.NewMongoWriter(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create mongodb writer: %w", err)
	}
	return writer, nil
}

// createElasticsearchWriter 创建Elasticsearch写入器
func createElasticsearchWriter(config *types.LogConfig) (LogWriter, error) {
	return nil, fmt.Errorf("elasticsearch writer not implemented")
}

// createClickHouseWriter 创建ClickHouse写入器
func createClickHouseWriter(config *types.LogConfig) (LogWriter, error) {
	writer, err := clickhouse.NewClickHouseWriter(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clickhouse writer: %w", err)
	}
	return writer, nil
}

// GetSupportedTargets 获取支持的输出目标列表（静态方法）
func GetSupportedTargets() []types.LogOutputTarget {
	return []types.LogOutputTarget{
		types.LogOutputConsole,
		types.LogOutputFile,
		types.LogOutputDatabase,
		types.LogOutputMongoDB,
		types.LogOutputElasticsearch,
		types.LogOutputClickHouse, // 已实现
	}
}

// IsTargetSupported 检查指定的输出目标是否支持（静态方法）
func IsTargetSupported(target types.LogOutputTarget) bool {
	supportedTargets := GetSupportedTargets()
	for _, supportedTarget := range supportedTargets {
		if target == supportedTarget {
			return true
		}
	}
	return false
}

// ValidateConfig 验证配置是否有效（静态方法）
func ValidateConfig(config *types.LogConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	targets := config.GetOutputTargets()
	if len(targets) == 0 {
		return fmt.Errorf("no output target specified")
	}
	
	if len(targets) > 1 {
		return fmt.Errorf("only one output target is allowed, got %d targets", len(targets))
	}

	target := targets[0]
	if !IsTargetSupported(target) {
		return fmt.Errorf("unsupported output target: %s", string(target))
	}

	// 验证特定目标的配置
	switch target {
	case types.LogOutputFile:
		if _, err := config.GetFileConfig(); err != nil {
			return fmt.Errorf("invalid file config: %w", err)
		}
	case types.LogOutputDatabase:
		if _, err := config.GetDatabaseConfig(); err != nil {
			return fmt.Errorf("invalid database config: %w", err)
		}
	case types.LogOutputMongoDB:
		if _, err := config.GetMongoConfig(); err != nil {
			return fmt.Errorf("invalid mongodb config: %w", err)
		}
	case types.LogOutputElasticsearch:
		if _, err := config.GetElasticsearchConfig(); err != nil {
			return fmt.Errorf("invalid elasticsearch config: %w", err)
		}
	case types.LogOutputClickHouse:
		if _, err := config.GetClickHouseConfig(); err != nil {
			return fmt.Errorf("invalid clickhouse config: %w", err)
		}
	}

	return nil
}
