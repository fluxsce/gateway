package cleanup

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"gateway/internal/gateway/logwrite/types"
	"gateway/pkg/database"
	"gateway/pkg/logger"
)

// LogCleaner 日志清理器
// 负责定期清理过期的访问日志和后端追踪日志，防止数据库无限增长
//
// 主要功能:
//   - 根据 LogRetentionDays 配置自动清理过期日志
//   - 支持多租户隔离清理
//   - 支持多种数据库（MySQL、Oracle、SQLite、ClickHouse、MongoDB）
//   - 定时执行，可配置清理间隔
//   - 批量删除，避免长事务
type LogCleaner struct {
	// 清理配置
	config *types.LogConfig

	// 网关实例ID
	instanceID string

	// 数据库连接
	db database.Database

	// 控制通道
	stopChan chan struct{}
	wg       sync.WaitGroup

	// 状态标识
	closed bool
	mutex  sync.Mutex
}

// NewLogCleaner 创建日志清理器
//
// 参数:
//   - instanceID: 网关实例ID
//   - logConfig: 日志配置，包含保留天数等信息
//
// 返回:
//   - *LogCleaner: 日志清理器实例
//   - error: 创建失败时返回错误信息
func NewLogCleaner(instanceID string, logConfig *types.LogConfig) (*LogCleaner, error) {
	// 获取数据库连接
	db := database.GetDefaultConnection()
	if db == nil {
		return nil, fmt.Errorf("failed to get default database connection")
	}

	// 从 ExtProperty 解析清理配置
	cleanupConfig := logConfig.GetCleanupConfig()
	if !cleanupConfig.CleanupEnabled {
		return nil, fmt.Errorf("cleanup is not enabled in config")
	}

	cleaner := &LogCleaner{
		instanceID: instanceID,
		config:     logConfig,
		db:         db,
		stopChan:   make(chan struct{}),
	}

	// 启动清理任务
	cleanupInterval := time.Duration(cleanupConfig.CleanupIntervalHour) * time.Hour
	cleaner.start(cleanupConfig.ScheduledTime, cleanupInterval)

	return cleaner, nil
}

// start 启动清理任务
func (c *LogCleaner) start(scheduledTime string, cleanupInterval time.Duration) {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		defer logger.Info("Log cleaner stopped")

		cleanupConfig := c.config.GetCleanupConfig()
		logger.Info("Log cleaner started",
			"instanceId", c.instanceID,
			"retentionDays", cleanupConfig.RetentionDays,
			"cleanupInterval", cleanupInterval,
			"batchDeleteSize", cleanupConfig.BatchDeleteSize,
			"scheduledTime", scheduledTime)

		// 如果设置了定时执行时间，先等待到指定时间
		if scheduledTime != "" {
			c.waitUntilScheduledTime(scheduledTime)
		}

		// 立即执行一次清理
		c.cleanup()

		// 定时执行
		ticker := time.NewTicker(cleanupInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				c.cleanup()
			case <-c.stopChan:
				return
			}
		}
	}()
}

// waitUntilScheduledTime 等待到指定的执行时间
func (c *LogCleaner) waitUntilScheduledTime(scheduledTime string) {
	now := time.Now()
	targetTime, err := parseScheduledTime(scheduledTime)
	if err != nil {
		logger.Warn("Failed to parse scheduled time, starting immediately", "scheduledTime", scheduledTime, "error", err)
		return
	}

	// 如果目标时间已过，等到明天的这个时间
	if targetTime.Before(now) {
		targetTime = targetTime.Add(24 * time.Hour)
	}

	waitDuration := targetTime.Sub(now)
	logger.Info("Waiting until scheduled cleanup time", "targetTime", targetTime, "waitDuration", waitDuration)

	select {
	case <-time.After(waitDuration):
		return
	case <-c.stopChan:
		return
	}
}

// parseScheduledTime 解析定时执行时间（格式：HH:MM）
func parseScheduledTime(scheduledTime string) (time.Time, error) {
	now := time.Now()
	t, err := time.Parse("15:04", scheduledTime)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid scheduled time format, expected HH:MM: %w", err)
	}

	// 构建今天的目标时间
	return time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, now.Location()), nil
}

// cleanup 执行清理操作
func (c *LogCleaner) cleanup() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.closed {
		return
	}

	startTime := time.Now()
	cleanupConfig := c.config.GetCleanupConfig()
	logger.Info("Starting log cleanup",
		"instanceId", c.instanceID,
		"retentionDays", cleanupConfig.RetentionDays)

	ctx := context.Background()
	var totalCleaned int64
	batchSize := cleanupConfig.BatchDeleteSize

	// 循环批量处理
	for {
		// 查询一批过期的 traceId
		traceIDs, err := c.queryExpiredTraceIDBatch(ctx, batchSize)
		if err != nil {
			logger.Error("Failed to query expired trace IDs", "error", err)
			break
		}

		if len(traceIDs) == 0 {
			logger.Info("No more expired logs to cleanup")
			break
		}

		logger.Debug("Found expired logs batch", "traceIdCount", len(traceIDs))

		// 先删除后端追踪日志（从表）
		backendTraceCleaned, err := c.deleteBackendTraceLogsByTraceIDs(ctx, traceIDs)
		if err != nil {
			logger.Error("Failed to delete backend trace logs", "error", err)
			break
		}
		totalCleaned += backendTraceCleaned

		// 再删除访问日志（主表）
		accessLogCleaned, err := c.deleteAccessLogsByTraceIDs(ctx, traceIDs)
		if err != nil {
			logger.Error("Failed to delete access logs", "error", err)
			break
		}
		totalCleaned += accessLogCleaned

		logger.Debug("Batch cleanup completed",
			"batchSize", len(traceIDs),
			"backendTraceCleaned", backendTraceCleaned,
			"accessLogCleaned", accessLogCleaned)

		// 短暂休眠，避免对数据库造成过大压力
		time.Sleep(100 * time.Millisecond)
	}

	duration := time.Since(startTime)
	logger.Info("Log cleanup completed",
		"totalCleaned", totalCleaned,
		"duration", duration)
}

// queryExpiredTraceIDBatch 查询一批过期的 traceId
func (c *LogCleaner) queryExpiredTraceIDBatch(ctx context.Context, batchSize int) ([]string, error) {
	cleanupConfig := c.config.GetCleanupConfig()
	cutoffTime := time.Now().AddDate(0, 0, -cleanupConfig.RetentionDays)

	// 查询一批过期的 traceId（只返回 traceId 字段）
	sql := `
		SELECT traceId 
		FROM HUB_GW_ACCESS_LOG 
		WHERE gatewayInstanceId = ? 
		  AND gatewayStartProcessingTime < ?
		LIMIT ?
	`

	var results []struct {
		TraceID string `db:"traceId"`
	}

	err := c.db.Query(ctx, &results, sql, []interface{}{
		c.instanceID,
		cutoffTime,
		batchSize,
	}, true)
	if err != nil {
		return nil, fmt.Errorf("query expired trace IDs failed: %w", err)
	}

	traceIDs := make([]string, 0, len(results))
	for _, r := range results {
		if r.TraceID != "" {
			traceIDs = append(traceIDs, r.TraceID)
		}
	}

	return traceIDs, nil
}

// deleteBackendTraceLogsByTraceIDs 根据 traceId 列表删除后端追踪日志
func (c *LogCleaner) deleteBackendTraceLogsByTraceIDs(ctx context.Context, traceIDs []string) (int64, error) {
	if len(traceIDs) == 0 {
		return 0, nil
	}

	// 构建 IN 子句的占位符
	placeholders := make([]string, len(traceIDs))
	args := make([]interface{}, len(traceIDs))
	for i, traceID := range traceIDs {
		placeholders[i] = "?"
		args[i] = traceID
	}

	sql := fmt.Sprintf(`
		DELETE FROM HUB_GW_BACKEND_TRACE_LOG 
		WHERE traceId IN (%s)
	`, strings.Join(placeholders, ","))

	result, err := c.db.Exec(ctx, sql, args, true)
	if err != nil {
		return 0, fmt.Errorf("delete backend trace logs failed: %w", err)
	}

	return result, nil
}

// deleteAccessLogsByTraceIDs 根据 traceId 列表删除访问日志
func (c *LogCleaner) deleteAccessLogsByTraceIDs(ctx context.Context, traceIDs []string) (int64, error) {
	if len(traceIDs) == 0 {
		return 0, nil
	}

	// 构建 IN 子句的占位符
	placeholders := make([]string, len(traceIDs))
	args := make([]interface{}, len(traceIDs))
	for i, traceID := range traceIDs {
		placeholders[i] = "?"
		args[i] = traceID
	}

	sql := fmt.Sprintf(`
		DELETE FROM HUB_GW_ACCESS_LOG 
		WHERE traceId IN (%s)
	`, strings.Join(placeholders, ","))

	result, err := c.db.Exec(ctx, sql, args, true)
	if err != nil {
		return 0, fmt.Errorf("delete access logs failed: %w", err)
	}

	return result, nil
}

// CleanupNow 立即执行一次清理（手动触发）
//
// 返回:
//   - int64: 清理的记录数
//   - error: 清理失败时返回错误信息
func (c *LogCleaner) CleanupNow() (int64, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.closed {
		return 0, fmt.Errorf("log cleaner is closed")
	}

	ctx := context.Background()
	cleanupConfig := c.config.GetCleanupConfig()
	batchSize := cleanupConfig.BatchDeleteSize
	var totalCleaned int64

	// 循环批量处理
	for {
		// 查询一批过期的 traceId
		traceIDs, err := c.queryExpiredTraceIDBatch(ctx, batchSize)
		if err != nil {
			return totalCleaned, err
		}

		if len(traceIDs) == 0 {
			break
		}

		// 先删除后端追踪日志（从表）
		backendTraceCleaned, err := c.deleteBackendTraceLogsByTraceIDs(ctx, traceIDs)
		if err != nil {
			return totalCleaned, err
		}
		totalCleaned += backendTraceCleaned

		// 再删除访问日志（主表）
		accessLogCleaned, err := c.deleteAccessLogsByTraceIDs(ctx, traceIDs)
		if err != nil {
			return totalCleaned, err
		}
		totalCleaned += accessLogCleaned

		// 短暂休眠
		time.Sleep(100 * time.Millisecond)
	}

	return totalCleaned, nil
}

// Close 关闭清理器
func (c *LogCleaner) Close() error {
	c.mutex.Lock()
	if c.closed {
		c.mutex.Unlock()
		return nil
	}
	c.closed = true
	c.mutex.Unlock()

	// 发送停止信号
	close(c.stopChan)

	// 等待清理goroutine结束
	c.wg.Wait()

	logger.Info("Log cleaner closed successfully")
	return nil
}
