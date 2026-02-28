package cleanup

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gateway/internal/gateway/logwrite/types"
	"gateway/pkg/database"
	"gateway/pkg/logger"
)

// ClickHouseLogCleaner ClickHouse日志清理器
// 专门用于清理ClickHouse中的访问日志和后端追踪日志
//
// ClickHouse特性：
//   - 使用 ALTER TABLE DROP PARTITION 删除整个分区（最快）
//   - 按天分区，清理时删除过期日期的分区
type ClickHouseLogCleaner struct {
	// 清理配置
	config *types.LogConfig

	// 网关实例ID
	instanceID string

	// ClickHouse数据库连接
	db database.Database

	// 控制通道
	stopChan chan struct{}
	wg       sync.WaitGroup

	// 状态标识
	closed bool
	mutex  sync.Mutex
}

// NewClickHouseLogCleaner 创建ClickHouse日志清理器
func NewClickHouseLogCleaner(instanceID string, logConfig *types.LogConfig) (*ClickHouseLogCleaner, error) {
	// 获取ClickHouse连接
	db := database.GetConnection("clickhouse_main")
	if db == nil {
		return nil, fmt.Errorf("failed to get clickhouse_main database connection")
	}

	// 从 ExtProperty 解析清理配置
	cleanupConfig := logConfig.GetCleanupConfig()
	if !cleanupConfig.CleanupEnabled {
		return nil, fmt.Errorf("cleanup is not enabled in config")
	}

	cleaner := &ClickHouseLogCleaner{
		instanceID: instanceID,
		config:     logConfig,
		db:         db,
		stopChan:   make(chan struct{}),
	}

	// 启动清理任务
	cleaner.start(cleanupConfig.ScheduledTime, time.Duration(cleanupConfig.CleanupIntervalHour)*time.Hour)

	return cleaner, nil
}

// start 启动清理任务
func (c *ClickHouseLogCleaner) start(scheduledTime string, cleanupInterval time.Duration) {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		defer logger.Info("ClickHouse log cleaner stopped")

		cleanupConfig := c.config.GetCleanupConfig()
		logger.Info("ClickHouse log cleaner started",
			"instanceId", c.instanceID,
			"retentionDays", cleanupConfig.RetentionDays,
			"cleanupInterval", cleanupInterval,
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
func (c *ClickHouseLogCleaner) waitUntilScheduledTime(scheduledTime string) {
	now := time.Now()
	targetTime, err := parseScheduledTime(scheduledTime)
	if err != nil {
		logger.Warn("Failed to parse scheduled time, starting immediately", "scheduledTime", scheduledTime, "error", err)
		return
	}

	if targetTime.Before(now) {
		targetTime = targetTime.Add(24 * time.Hour)
	}

	waitDuration := targetTime.Sub(now)
	logger.Info("Waiting until scheduled ClickHouse cleanup time", "targetTime", targetTime, "waitDuration", waitDuration)

	select {
	case <-time.After(waitDuration):
		return
	case <-c.stopChan:
		return
	}
}

// cleanup 执行清理操作
func (c *ClickHouseLogCleaner) cleanup() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.closed {
		return
	}

	startTime := time.Now()
	cleanupConfig := c.config.GetCleanupConfig()
	logger.Info("Starting ClickHouse log cleanup",
		"instanceId", c.instanceID,
		"retentionDays", cleanupConfig.RetentionDays)

	ctx := context.Background()
	var totalDropped int

	// 先清理后端追踪日志（按分区删除）
	backendTraceDropped, err := c.cleanupBackendTraceLogs(ctx)
	if err != nil {
		logger.Error("Failed to cleanup ClickHouse backend trace logs", "error", err)
	} else {
		totalDropped += backendTraceDropped
		logger.Info("ClickHouse backend trace logs partitions dropped", "count", backendTraceDropped)
	}

	// 再清理访问日志（按分区删除）
	accessLogDropped, err := c.cleanupAccessLogs(ctx)
	if err != nil {
		logger.Error("Failed to cleanup ClickHouse access logs", "error", err)
	} else {
		totalDropped += accessLogDropped
		logger.Info("ClickHouse access logs partitions dropped", "count", accessLogDropped)
	}

	duration := time.Since(startTime)
	logger.Info("ClickHouse log cleanup completed",
		"totalPartitionsDropped", totalDropped,
		"duration", duration)
}

// cleanupAccessLogs 清理ClickHouse访问日志（按分区删除）
func (c *ClickHouseLogCleaner) cleanupAccessLogs(ctx context.Context) (int, error) {
	cleanupConfig := c.config.GetCleanupConfig()
	cutoffDate := time.Now().AddDate(0, 0, -cleanupConfig.RetentionDays)

	logger.Info("Cleaning up ClickHouse access logs by partition",
		"table", "HUB_GW_ACCESS_LOG",
		"instanceId", c.instanceID,
		"cutoffDate", cutoffDate.Format("2006-01-02"),
		"retentionDays", cleanupConfig.RetentionDays)

	// 查询需要删除的分区列表
	querySQL := `
		SELECT DISTINCT partition
		FROM system.parts
		WHERE database = currentDatabase()
		  AND table = 'HUB_GW_ACCESS_LOG'
		  AND partition < ?
		  AND active = 1
		ORDER BY partition
	`

	var partitions []struct {
		Partition string `db:"partition"`
	}
	err := c.db.Query(ctx, &partitions, querySQL, []interface{}{
		cutoffDate.Format("2006-01-02"),
	}, true)
	if err != nil {
		return 0, fmt.Errorf("query partitions failed: %w", err)
	}

	if len(partitions) == 0 {
		logger.Info("No partitions to drop for access logs")
		return 0, nil
	}

	// 删除每个分区
	droppedCount := 0
	for _, p := range partitions {
		dropSQL := fmt.Sprintf("ALTER TABLE HUB_GW_ACCESS_LOG DROP PARTITION '%s'", p.Partition)
		_, err := c.db.Exec(ctx, dropSQL, nil, true)
		if err != nil {
			logger.Error("Failed to drop partition",
				"table", "HUB_GW_ACCESS_LOG",
				"partition", p.Partition,
				"error", err)
			continue
		}
		droppedCount++
		logger.Info("Dropped partition",
			"table", "HUB_GW_ACCESS_LOG",
			"partition", p.Partition)
	}

	return droppedCount, nil
}

// cleanupBackendTraceLogs 清理ClickHouse后端追踪日志（按分区删除）
func (c *ClickHouseLogCleaner) cleanupBackendTraceLogs(ctx context.Context) (int, error) {
	cleanupConfig := c.config.GetCleanupConfig()
	cutoffDate := time.Now().AddDate(0, 0, -cleanupConfig.RetentionDays)

	logger.Info("Cleaning up ClickHouse backend trace logs by partition",
		"table", "HUB_GW_BACKEND_TRACE_LOG",
		"instanceId", c.instanceID,
		"cutoffDate", cutoffDate.Format("2006-01-02"),
		"retentionDays", cleanupConfig.RetentionDays)

	// 查询需要删除的分区列表
	querySQL := `
		SELECT DISTINCT partition
		FROM system.parts
		WHERE database = currentDatabase()
		  AND table = 'HUB_GW_BACKEND_TRACE_LOG'
		  AND partition < ?
		  AND active = 1
		ORDER BY partition
	`

	var partitions []struct {
		Partition string `db:"partition"`
	}
	err := c.db.Query(ctx, &partitions, querySQL, []interface{}{
		cutoffDate.Format("2006-01-02"),
	}, true)
	if err != nil {
		return 0, fmt.Errorf("query partitions failed: %w", err)
	}

	if len(partitions) == 0 {
		logger.Info("No partitions to drop for backend trace logs")
		return 0, nil
	}

	// 删除每个分区
	droppedCount := 0
	for _, p := range partitions {
		dropSQL := fmt.Sprintf("ALTER TABLE HUB_GW_BACKEND_TRACE_LOG DROP PARTITION '%s'", p.Partition)
		_, err := c.db.Exec(ctx, dropSQL, nil, true)
		if err != nil {
			logger.Error("Failed to drop partition",
				"table", "HUB_GW_BACKEND_TRACE_LOG",
				"partition", p.Partition,
				"error", err)
			continue
		}
		droppedCount++
		logger.Info("Dropped partition",
			"table", "HUB_GW_BACKEND_TRACE_LOG",
			"partition", p.Partition)
	}

	return droppedCount, nil
}

// CleanupNow 立即执行一次清理
func (c *ClickHouseLogCleaner) CleanupNow() (int64, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.closed {
		return 0, fmt.Errorf("ClickHouse log cleaner is closed")
	}

	ctx := context.Background()
	var totalDropped int64

	// 先删除后端追踪日志分区
	backendTraceDropped, err := c.cleanupBackendTraceLogs(ctx)
	if err != nil {
		return totalDropped, err
	}
	totalDropped += int64(backendTraceDropped)

	// 再删除访问日志分区
	accessLogDropped, err := c.cleanupAccessLogs(ctx)
	if err != nil {
		return totalDropped, err
	}
	totalDropped += int64(accessLogDropped)

	return totalDropped, nil
}

// Close 关闭清理器
func (c *ClickHouseLogCleaner) Close() error {
	c.mutex.Lock()
	if c.closed {
		c.mutex.Unlock()
		return nil
	}
	c.closed = true
	c.mutex.Unlock()

	close(c.stopChan)
	c.wg.Wait()

	logger.Info("ClickHouse log cleaner closed successfully")
	return nil
}
