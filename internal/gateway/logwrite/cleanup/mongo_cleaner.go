package cleanup

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gateway/internal/gateway/logwrite/types"
	"gateway/pkg/logger"
	"gateway/pkg/mongo/client"
	"gateway/pkg/mongo/factory"
	mongotypes "gateway/pkg/mongo/types"
)

// MongoLogCleaner MongoDB日志清理器
// 专门用于清理MongoDB中的访问日志和后端追踪日志
type MongoLogCleaner struct {
	// 清理配置
	config *types.LogConfig

	// 网关实例ID
	instanceID string

	// MongoDB连接
	mongoClient *client.Client

	// 控制通道
	stopChan chan struct{}
	wg       sync.WaitGroup

	// 状态标识
	closed bool
	mutex  sync.Mutex
}

// NewMongoLogCleaner 创建MongoDB日志清理器
func NewMongoLogCleaner(instanceID string, logConfig *types.LogConfig) (*MongoLogCleaner, error) {
	// 获取默认MongoDB连接
	mongoClient, err := factory.GetDefaultConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to get default MongoDB connection: %w", err)
	}

	// 从 ExtProperty 解析清理配置
	cleanupConfig := logConfig.GetCleanupConfig()
	if !cleanupConfig.CleanupEnabled {
		return nil, fmt.Errorf("cleanup is not enabled in config")
	}

	cleaner := &MongoLogCleaner{
		instanceID:  instanceID,
		config:      logConfig,
		mongoClient: mongoClient,
		stopChan:    make(chan struct{}),
	}

	// 启动清理任务
	cleanupInterval := time.Duration(cleanupConfig.CleanupIntervalHour) * time.Hour
	cleaner.start(cleanupConfig.ScheduledTime, cleanupInterval)

	return cleaner, nil
}

// start 启动清理任务
func (c *MongoLogCleaner) start(scheduledTime string, cleanupInterval time.Duration) {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		defer logger.Info("MongoDB log cleaner stopped")

		cleanupConfig := c.config.GetCleanupConfig()
		logger.Info("MongoDB log cleaner started",
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
func (c *MongoLogCleaner) waitUntilScheduledTime(scheduledTime string) {
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
	logger.Info("Waiting until scheduled MongoDB cleanup time", "targetTime", targetTime, "waitDuration", waitDuration)

	select {
	case <-time.After(waitDuration):
		return
	case <-c.stopChan:
		return
	}
}

// cleanup 执行清理操作
func (c *MongoLogCleaner) cleanup() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.closed {
		return
	}

	startTime := time.Now()
	cleanupConfig := c.config.GetCleanupConfig()
	logger.Info("Starting MongoDB log cleanup",
		"instanceId", c.instanceID,
		"retentionDays", cleanupConfig.RetentionDays)

	ctx := context.Background()
	batchSize := cleanupConfig.BatchDeleteSize
	var totalCleaned int64

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
			logger.Error("Failed to delete MongoDB backend trace logs", "error", err)
			break
		}
		totalCleaned += backendTraceCleaned

		// 再删除访问日志（主表）
		accessLogCleaned, err := c.deleteAccessLogsByTraceIDs(ctx, traceIDs)
		if err != nil {
			logger.Error("Failed to delete MongoDB access logs", "error", err)
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
	logger.Info("MongoDB log cleanup completed",
		"totalCleaned", totalCleaned,
		"duration", duration)
}

// queryExpiredTraceIDBatch 查询一批过期的 traceId（只返回 traceId 字段）
func (c *MongoLogCleaner) queryExpiredTraceIDBatch(ctx context.Context, batchSize int) ([]string, error) {
	cleanupConfig := c.config.GetCleanupConfig()
	cutoffTime := time.Now().AddDate(0, 0, -cleanupConfig.RetentionDays)

	// 获取默认数据库
	database, err := c.mongoClient.DefaultDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to get default database: %w", err)
	}

	// 查询主表中过期的 traceId（只返回 traceId 字段，限制数量）
	accessLogCollection := database.Collection("HUB_GW_ACCESS_LOG")

	// 使用 FindOptions 限制返回字段和数量
	limit := int64(batchSize)
	findOptions := &mongotypes.FindOptions{
		Projection: map[string]interface{}{
			"traceId": 1,
			"_id":     0,
		},
		Limit: &limit,
	}

	cursor, err := accessLogCollection.Find(ctx, map[string]interface{}{
		"gatewayInstanceId": c.instanceID,
		"gatewayStartProcessingTime": map[string]interface{}{
			"$lt": cutoffTime,
		},
	}, findOptions)
	if err != nil {
		return nil, fmt.Errorf("query expired trace IDs failed: %w", err)
	}
	defer cursor.Close(ctx)

	var traceIDs []string
	for cursor.Next(ctx) {
		var doc struct {
			TraceID string `bson:"traceId"`
		}
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		if doc.TraceID != "" {
			traceIDs = append(traceIDs, doc.TraceID)
		}
	}

	return traceIDs, nil
}

// deleteBackendTraceLogsByTraceIDs 根据 traceId 列表删除后端追踪日志
func (c *MongoLogCleaner) deleteBackendTraceLogsByTraceIDs(ctx context.Context, traceIDs []string) (int64, error) {
	if len(traceIDs) == 0 {
		return 0, nil
	}

	// 获取默认数据库
	database, err := c.mongoClient.DefaultDatabase()
	if err != nil {
		return 0, fmt.Errorf("failed to get default database: %w", err)
	}

	// 删除后端追踪日志
	backendTraceCollection := database.Collection("HUB_GW_BACKEND_TRACE_LOG")
	filter := map[string]interface{}{
		"traceId": map[string]interface{}{
			"$in": traceIDs,
		},
	}

	result, err := backendTraceCollection.DeleteMany(ctx, filter, nil)
	if err != nil {
		return 0, fmt.Errorf("delete MongoDB backend trace logs failed: %w", err)
	}

	return result.DeletedCount, nil
}

// deleteAccessLogsByTraceIDs 根据 traceId 列表删除访问日志
func (c *MongoLogCleaner) deleteAccessLogsByTraceIDs(ctx context.Context, traceIDs []string) (int64, error) {
	if len(traceIDs) == 0 {
		return 0, nil
	}

	// 获取默认数据库和集合
	database, err := c.mongoClient.DefaultDatabase()
	if err != nil {
		return 0, fmt.Errorf("failed to get default database: %w", err)
	}
	collection := database.Collection("HUB_GW_ACCESS_LOG")

	// 删除条件：根据 traceId 列表删除
	filter := map[string]interface{}{
		"traceId": map[string]interface{}{
			"$in": traceIDs,
		},
	}

	// 执行删除
	result, err := collection.DeleteMany(ctx, filter, nil)
	if err != nil {
		return 0, fmt.Errorf("delete MongoDB access logs failed: %w", err)
	}

	return result.DeletedCount, nil
}

// CleanupNow 立即执行一次清理
func (c *MongoLogCleaner) CleanupNow() (int64, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.closed {
		return 0, fmt.Errorf("MongoDB log cleaner is closed")
	}

	ctx := context.Background()
	var totalCleaned int64
	cleanupConfig := c.config.GetCleanupConfig()
	batchSize := cleanupConfig.BatchDeleteSize

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
func (c *MongoLogCleaner) Close() error {
	c.mutex.Lock()
	if c.closed {
		c.mutex.Unlock()
		return nil
	}
	c.closed = true
	c.mutex.Unlock()

	close(c.stopChan)
	c.wg.Wait()

	logger.Info("MongoDB log cleaner closed successfully")
	return nil
}
