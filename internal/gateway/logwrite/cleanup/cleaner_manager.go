package cleanup

import (
	"fmt"
	"sync"

	"gateway/internal/gateway/logwrite/types"
	"gateway/pkg/logger"
)

// CleanerManager 清理器管理器
// 统一管理所有日志清理器（数据库、MongoDB、ClickHouse）
type CleanerManager struct {
	// 各类清理器
	dbCleaner         *LogCleaner
	mongoCleaner      *MongoLogCleaner
	clickhouseCleaner *ClickHouseLogCleaner

	// 互斥锁
	mutex sync.Mutex

	// 状态标识
	closed bool
}

// NewCleanerManager 创建清理器管理器
//
// 参数:
//   - instanceID: 网关实例ID
//   - logConfig: 日志配置
//
// 返回:
//   - *CleanerManager: 清理器管理器实例
//   - error: 创建失败时返回错误信息
func NewCleanerManager(instanceID string, logConfig *types.LogConfig) (*CleanerManager, error) {
	if logConfig == nil {
		return nil, fmt.Errorf("log config cannot be nil")
	}

	// 检查 ExtProperty 中的清理配置
	cleanupConfig := logConfig.GetCleanupConfig()
	if !cleanupConfig.CleanupEnabled {
		logger.Info("Log cleanup is not enabled in ExtProperty")
		return &CleanerManager{}, nil
	}

	// 验证保留天数
	if cleanupConfig.RetentionDays <= 0 {
		logger.Info("Retention days is 0 or negative, log cleanup is disabled")
		return &CleanerManager{}, nil
	}

	manager := &CleanerManager{}

	// 根据输出目标创建对应的清理器
	outputTargets := logConfig.GetOutputTargets()

	for _, target := range outputTargets {
		switch target {
		case types.LogOutputDatabase:
			// 创建数据库清理器（MySQL、Oracle、SQLite）
			dbCleaner, err := NewLogCleaner(instanceID, logConfig)
			if err != nil {
				logger.Warn("Failed to create database log cleaner", "error", err)
			} else {
				manager.dbCleaner = dbCleaner
				logger.Info("Database log cleaner created successfully")
			}

		case types.LogOutputMongoDB:
			// 创建MongoDB清理器
			mongoCleaner, err := NewMongoLogCleaner(instanceID, logConfig)
			if err != nil {
				logger.Warn("Failed to create MongoDB log cleaner", "error", err)
			} else {
				manager.mongoCleaner = mongoCleaner
				logger.Info("MongoDB log cleaner created successfully")
			}

		case types.LogOutputClickHouse:
			// 创建ClickHouse清理器
			clickhouseCleaner, err := NewClickHouseLogCleaner(instanceID, logConfig)
			if err != nil {
				logger.Warn("Failed to create ClickHouse log cleaner", "error", err)
			} else {
				manager.clickhouseCleaner = clickhouseCleaner
				logger.Info("ClickHouse log cleaner created successfully")
			}
		}
	}

	return manager, nil
}

// CleanupNow 立即执行一次清理（所有清理器）
//
// 返回:
//   - map[string]int64: 各清理器的清理数量
//   - error: 清理失败时返回错误信息
func (m *CleanerManager) CleanupNow() (map[string]int64, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.closed {
		return nil, fmt.Errorf("cleaner manager is closed")
	}

	result := make(map[string]int64)
	var lastErr error

	// 清理数据库日志
	if m.dbCleaner != nil {
		count, err := m.dbCleaner.CleanupNow()
		if err != nil {
			logger.Error("Database cleanup failed", "error", err)
			lastErr = err
		}
		result["database"] = count
	}

	// 清理MongoDB日志
	if m.mongoCleaner != nil {
		count, err := m.mongoCleaner.CleanupNow()
		if err != nil {
			logger.Error("MongoDB cleanup failed", "error", err)
			lastErr = err
		}
		result["mongodb"] = count
	}

	// 清理ClickHouse日志
	if m.clickhouseCleaner != nil {
		count, err := m.clickhouseCleaner.CleanupNow()
		if err != nil {
			logger.Error("ClickHouse cleanup failed", "error", err)
			lastErr = err
		}
		result["clickhouse"] = count
	}

	return result, lastErr
}

// Close 关闭所有清理器
func (m *CleanerManager) Close() error {
	m.mutex.Lock()
	if m.closed {
		m.mutex.Unlock()
		return nil
	}
	m.closed = true
	m.mutex.Unlock()

	var lastErr error

	// 关闭数据库清理器
	if m.dbCleaner != nil {
		if err := m.dbCleaner.Close(); err != nil {
			logger.Error("Failed to close database log cleaner", "error", err)
			lastErr = err
		}
	}

	// 关闭MongoDB清理器
	if m.mongoCleaner != nil {
		if err := m.mongoCleaner.Close(); err != nil {
			logger.Error("Failed to close MongoDB log cleaner", "error", err)
			lastErr = err
		}
	}

	// 关闭ClickHouse清理器
	if m.clickhouseCleaner != nil {
		if err := m.clickhouseCleaner.Close(); err != nil {
			logger.Error("Failed to close ClickHouse log cleaner", "error", err)
			lastErr = err
		}
	}

	logger.Info("Cleaner manager closed successfully")
	return lastErr
}
