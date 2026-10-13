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
	cutoffTime := time.Now().AddDate(0, 0, -cleanupConfig.RetentionDays)
	var totalCleaned int64

	// 从表按本配置的 tenantId + requestStartTime 直接删。
	backendTraceCleaned, err := c.deleteExpiredBackendTraces(ctx, cutoffTime, batchSize)
	if err != nil {
		logger.Error("Failed to delete MongoDB backend trace logs", "error", err)
	}
	totalCleaned += backendTraceCleaned

	// 主表按实例 + gatewayStartProcessingTime 直接删。
	accessLogCleaned, err := c.deleteExpiredAccessLogs(ctx, cutoffTime, batchSize)
	if err != nil {
		logger.Error("Failed to delete MongoDB access logs", "error", err)
	}
	totalCleaned += accessLogCleaned

	duration := time.Since(startTime)
	logger.Info("MongoDB log cleanup completed",
		"totalCleaned", totalCleaned,
		"duration", duration)
}

// deleteExpiredAccessLogs 按 gatewayInstanceId + gatewayStartProcessingTime 分批删除访问日志。
// 条件对齐 idx_instance_time。deleteMany 没有条数上限，所以先限定取出 _id 再删。
func (c *MongoLogCleaner) deleteExpiredAccessLogs(ctx context.Context, cutoffTime time.Time, batchSize int) (int64, error) {
	if c.instanceID == "" {
		return 0, fmt.Errorf("gatewayInstanceId is required to delete access logs")
	}
	return c.deleteByFilter(ctx, "HUB_GW_ACCESS_LOG", map[string]interface{}{
		"gatewayInstanceId": c.instanceID,
		"gatewayStartProcessingTime": map[string]interface{}{
			"$lt": cutoffTime,
		},
	}, batchSize)
}

// deleteExpiredBackendTraces 按 tenantId + requestStartTime 分批删除后端追踪日志。
// 条件对齐 idx_tenant_time。
func (c *MongoLogCleaner) deleteExpiredBackendTraces(ctx context.Context, cutoffTime time.Time, batchSize int) (int64, error) {
	tenantID := ""
	if c.config != nil {
		tenantID = c.config.TenantID
	}
	if tenantID == "" {
		return 0, fmt.Errorf("tenantId is required to delete backend trace logs")
	}
	return c.deleteByFilter(ctx, "HUB_GW_BACKEND_TRACE_LOG", map[string]interface{}{
		"tenantId": tenantID,
		"requestStartTime": map[string]interface{}{
			"$lt": cutoffTime,
		},
	}, batchSize)
}

// deleteByFilter 按条件分批删除。每批只取 _id，再按主键删除。
func (c *MongoLogCleaner) deleteByFilter(ctx context.Context, collectionName string, filter map[string]interface{}, batchSize int) (int64, error) {
	database, err := c.mongoClient.DefaultDatabase()
	if err != nil {
		return 0, fmt.Errorf("failed to get default database: %w", err)
	}
	collection := database.Collection(collectionName)
	if batchSize < 1 {
		batchSize = 1
	}
	limit := int64(batchSize)
	var total int64
	for {
		cursor, err := collection.Find(ctx, filter, &mongotypes.FindOptions{
			Projection: map[string]interface{}{"_id": 1},
			Limit:      &limit,
		})
		if err != nil {
			return total, fmt.Errorf("query expired %s failed: %w", collectionName, err)
		}
		ids := make([]interface{}, 0, batchSize)
		for cursor.Next(ctx) {
			var doc struct {
				ID interface{} `bson:"_id"`
			}
			if err := cursor.Decode(&doc); err != nil || doc.ID == nil {
				continue
			}
			ids = append(ids, doc.ID)
		}
		closeErr := cursor.Close(ctx)
		if closeErr != nil {
			return total, fmt.Errorf("close %s cursor failed: %w", collectionName, closeErr)
		}
		if len(ids) == 0 {
			break
		}
		result, err := collection.DeleteMany(ctx, map[string]interface{}{
			"_id": map[string]interface{}{"$in": ids},
		}, nil)
		if err != nil {
			return total, fmt.Errorf("delete MongoDB %s failed: %w", collectionName, err)
		}
		total += result.DeletedCount
		if len(ids) < batchSize {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	return total, nil
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
	cutoffTime := time.Now().AddDate(0, 0, -cleanupConfig.RetentionDays)

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
