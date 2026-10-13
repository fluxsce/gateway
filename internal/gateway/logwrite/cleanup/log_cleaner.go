package cleanup

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gateway/internal/gateway/logwrite/types"
	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
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
	cutoffTime := time.Now().AddDate(0, 0, -cleanupConfig.RetentionDays)

	// 从表按本配置的 tenantId + requestStartTime 直接删。
	backendTraceCleaned, err := c.deleteExpiredBackendTraces(ctx, cutoffTime, batchSize)
	if err != nil {
		logger.Error("Failed to delete backend trace logs", "error", err)
	}
	totalCleaned += backendTraceCleaned

	// 主表按实例 + gatewayStartProcessingTime 直接删。
	accessLogCleaned, err := c.deleteExpiredAccessLogs(ctx, cutoffTime, batchSize)
	if err != nil {
		logger.Error("Failed to delete access logs", "error", err)
	}
	totalCleaned += accessLogCleaned

	duration := time.Since(startTime)
	logger.Info("Log cleanup completed",
		"totalCleaned", totalCleaned,
		"duration", duration)
}

// deleteExpiredAccessLogs 按 gatewayInstanceId + gatewayStartProcessingTime 分批删除访问日志。
// 条件对齐 (gatewayInstanceId, gatewayStartProcessingTime)。
func (c *LogCleaner) deleteExpiredAccessLogs(ctx context.Context, cutoffTime time.Time, batchSize int) (int64, error) {
	if c.instanceID == "" {
		return 0, fmt.Errorf("gatewayInstanceId is required to delete access logs")
	}
	return c.limitedDelete(
		ctx,
		"HUB_GW_ACCESS_LOG",
		"gatewayInstanceId = ? AND gatewayStartProcessingTime < ?",
		[]interface{}{c.instanceID, cutoffTime},
		batchSize,
	)
}

// deleteExpiredBackendTraces 按 tenantId + requestStartTime 分批删除后端追踪日志。
// 条件对齐 IDX_GW_BTRACE_TENANT_TIME。
func (c *LogCleaner) deleteExpiredBackendTraces(ctx context.Context, cutoffTime time.Time, batchSize int) (int64, error) {
	tenantID := ""
	if c.config != nil {
		tenantID = c.config.TenantID
	}
	if tenantID == "" {
		return 0, fmt.Errorf("tenantId is required to delete backend trace logs")
	}
	return c.limitedDelete(
		ctx,
		"HUB_GW_BACKEND_TRACE_LOG",
		"tenantId = ? AND requestStartTime < ?",
		[]interface{}{tenantID, cutoffTime},
		batchSize,
	)
}

// limitedDelete 按条件分批直接删除，直到不足一批。
func (c *LogCleaner) limitedDelete(ctx context.Context, table, where string, args []interface{}, batchSize int) (int64, error) {
	dbType := sqlutils.GetDatabaseType(c.db)
	var total int64
	for {
		query, qargs, err := sqlutils.BuildLimitedDeleteQuery(dbType, table, where, args, batchSize)
		if err != nil {
			return total, fmt.Errorf("build %s delete failed: %w", table, err)
		}
		n, err := c.db.Exec(ctx, query, qargs, true)
		if err != nil {
			return total, fmt.Errorf("delete %s failed: %w", table, err)
		}
		total += n
		if n == 0 || n < int64(batchSize) {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	return total, nil
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
	cutoffTime := time.Now().AddDate(0, 0, -cleanupConfig.RetentionDays)
	var totalCleaned int64

	backendTraceCleaned, err := c.deleteExpiredBackendTraces(ctx, cutoffTime, batchSize)
	if err != nil {
		return totalCleaned, err
	}
	totalCleaned += backendTraceCleaned

	accessLogCleaned, err := c.deleteExpiredAccessLogs(ctx, cutoffTime, batchSize)
	if err != nil {
		return totalCleaned, err
	}
	totalCleaned += accessLogCleaned

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
