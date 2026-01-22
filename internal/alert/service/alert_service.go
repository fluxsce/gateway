package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"gateway/internal/alert/dao"
	"gateway/internal/alert/types"
	"gateway/pkg/alert"
	"gateway/pkg/config"
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/pkg/utils/random"
)

// AlertServiceImpl 告警服务实现
type AlertServiceImpl struct {
	tenantId    string
	configDAO   *dao.ConfigDAO
	logDAO      *dao.LogDAO
	templateDAO *dao.TemplateDAO

	// 队列和状态
	logQueue    chan *types.AlertLog // 日志写入队列（直接使用 AlertLog）
	batchBuffer []*types.AlertLog    // 批量写入缓冲区
	batchMutex  sync.Mutex           // 保护批量缓冲区
	flushTicker *time.Ticker         // 定时刷新ticker
	running     bool                 // 服务是否运行中
	mu          sync.RWMutex         // 保护状态
	ctx         context.Context      // 上下文
	cancel      context.CancelFunc   // 取消函数
	wg          sync.WaitGroup       // 等待组

	// 配置
	pollInterval      time.Duration // 轮询间隔（默认3秒）
	batchSize         int           // 每批处理数量（默认50）
	cleanupInterval   time.Duration // 清理间隔（默认1小时）
	logRetentionHours int           // 日志保留时间（小时，默认7天）
	logQueueSize      int           // 日志队列大小（默认1000）
	logBatchSize      int           // 日志批量写入大小（默认100）
	logFlushInterval  time.Duration // 日志刷新间隔（默认5秒）
}

// NewAlertService 创建告警服务实例
func NewAlertService(db database.Database, tenantId string) *AlertServiceImpl {
	// 读取配置（使用配置常量）
	pollInterval := parseDuration(config.GetString(config.ALERT_POLL_INTERVAL, "3s"), 3*time.Second)
	batchSize := config.GetInt(config.ALERT_BATCH_SIZE, 50)
	cleanupInterval := parseDuration(config.GetString(config.ALERT_CLEANUP_INTERVAL, "1h"), 1*time.Hour)
	logRetentionHours := config.GetInt(config.ALERT_LOG_RETENTION_HOURS, 168) // 默认7天
	logQueueSize := config.GetInt(config.ALERT_LOG_QUEUE_SIZE, 1000)
	logBatchSize := config.GetInt(config.ALERT_LOG_BATCH_SIZE, 100)
	logFlushInterval := parseDuration(config.GetString(config.ALERT_LOG_FLUSH_INTERVAL, "5s"), 5*time.Second)

	return &AlertServiceImpl{
		tenantId:          tenantId,
		configDAO:         dao.NewConfigDAO(db),
		logDAO:            dao.NewLogDAO(db),
		templateDAO:       dao.NewTemplateDAO(db),
		logQueue:          make(chan *types.AlertLog, logQueueSize),
		batchBuffer:       make([]*types.AlertLog, 0, logBatchSize),
		pollInterval:      pollInterval,
		batchSize:         batchSize,
		cleanupInterval:   cleanupInterval,
		logRetentionHours: logRetentionHours,
		logQueueSize:      logQueueSize,
		logBatchSize:      logBatchSize,
		logFlushInterval:  logFlushInterval,
	}
}

// Start 启动告警服务
func (s *AlertServiceImpl) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("告警服务已在运行")
	}
	s.ctx, s.cancel = context.WithCancel(ctx)
	s.running = true
	s.mu.Unlock()

	logger.Info("告警服务启动", "tenantId", s.tenantId)

	// 初始化定时刷新器
	s.flushTicker = time.NewTicker(s.logFlushInterval)

	// 启动日志写入 worker
	s.wg.Add(1)
	go s.logWriteWorker()

	// 启动定时刷新 worker
	s.wg.Add(1)
	go s.flushWorker()

	// 启动发送处理 worker
	s.wg.Add(1)
	go s.sendWorker()

	// 启动清理 worker
	s.wg.Add(1)
	go s.cleanupWorker()

	logger.Info("告警服务启动完成")
	return nil
}

// Stop 停止告警服务
func (s *AlertServiceImpl) Stop(ctx context.Context) error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = false
	s.mu.Unlock()

	logger.Info("告警服务停止中...", "tenantId", s.tenantId)

	// 取消上下文
	if s.cancel != nil {
		s.cancel()
	}

	// 停止定时刷新
	if s.flushTicker != nil {
		s.flushTicker.Stop()
	}

	// 关闭队列
	close(s.logQueue)

	// 等待所有 worker 结束
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// 刷新剩余的缓冲区数据
		if err := s.flushBatch(); err != nil {
			logger.Error("刷新剩余缓冲区失败", "error", err)
		}
		logger.Info("告警服务已停止")
	case <-ctx.Done():
		logger.Warn("告警服务停止超时")
		// 即使超时也尝试刷新缓冲区
		if err := s.flushBatch(); err != nil {
			logger.Error("刷新剩余缓冲区失败", "error", err)
		}
	}

	return nil
}

// SendAlert 发送告警（异步写入队列，不阻塞）
func (s *AlertServiceImpl) SendAlert(ctx context.Context, level, alertType, title, content, channelName string, tags map[string]string, extra map[string]interface{}, tableData map[string]interface{}) (string, error) {
	// 生成日志ID
	alertLogId := random.GenerateUniqueStringWithPrefix("log_", 32)

	// 确定渠道名称
	if channelName == "" {
		// 从 alert manager 获取默认渠道
		alertManager := alert.GetGlobalManager()
		defaultChannel := alertManager.GetDefaultChannel()
		if defaultChannel == nil {
			return "", fmt.Errorf("未找到默认渠道，请先配置告警渠道")
		}
		channelName = defaultChannel.Name()
	}

	// 构建告警日志（直接构造 AlertLog）
	now := time.Now()
	alertLog := &types.AlertLog{
		TenantId:       s.tenantId,
		AlertLogId:     alertLogId,
		AlertLevel:     level,
		AlertType:      &alertType,
		AlertTitle:     title,
		AlertContent:   &content,
		AlertTimestamp: now,
		ChannelName:    &channelName,
		SendStatus:     stringPtr("PENDING"),
		AddTime:        now,
		AddWho:         "system",
		EditTime:       now,
		EditWho:        "system",
		OprSeqFlag:     random.Generate32BitRandomString(),
		CurrentVersion: 1,
		ActiveFlag:     "Y",
	}

	// 设置标签和额外数据（序列化为JSON）
	if len(tags) > 0 {
		tagsJSON, _ := json.Marshal(tags)
		tagsStr := string(tagsJSON)
		alertLog.AlertTags = &tagsStr
	}
	if len(extra) > 0 {
		extraJSON, _ := json.Marshal(extra)
		extraStr := string(extraJSON)
		alertLog.AlertExtra = &extraStr
	}
	if len(tableData) > 0 {
		tableDataJSON, _ := json.Marshal(tableData)
		tableDataStr := string(tableDataJSON)
		alertLog.TableData = &tableDataStr
	}

	// 异步写入队列（非阻塞）
	select {
	case s.logQueue <- alertLog:
		// 成功入队
	case <-ctx.Done():
		return "", ctx.Err()
	default:
		// 队列已满，直接丢弃，避免影响全局效率
		logger.Warn("告警日志队列已满，丢弃告警日志", "alertLogId", alertLogId, "title", title)
		return "", fmt.Errorf("告警日志队列已满，告警已丢弃")
	}

	return alertLogId, nil
}
