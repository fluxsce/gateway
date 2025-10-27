package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"gateway/internal/alert/loader"
	"gateway/internal/types/alerttypes"
	"gateway/pkg/alert"
	"gateway/pkg/database"
	"gateway/pkg/logger"
)

// AlertQueue 告警消息队列
// 负责接收、缓存和处理告警消息
// 使用全局 AlertManager 单例
type AlertQueue struct {
	// 消息通道
	messageChan chan *AlertMessage
	// 数据库连接
	db database.Database
	// 租户ID
	tenantId string
	// 配置加载器
	configLoader *loader.AlertConfigLoader
	// 停止信号
	stopChan chan struct{}
	// 等待组
	wg sync.WaitGroup
	// 队列状态
	running bool
	mu      sync.RWMutex
	// 队列配置
	config QueueConfig
}

// QueueConfig 队列配置
type QueueConfig struct {
	// 队列容量
	Capacity int
	// 工作协程数
	Workers int
	// 批处理大小
	BatchSize int
	// 批处理超时时间
	BatchTimeout time.Duration
	// 是否自动保存记录
	AutoSaveRecord bool
}

// AlertMessage 告警消息
type AlertMessage struct {
	// 消息ID
	MessageId string
	// 渠道ID列表（逗号分隔）
	ChannelIds string
	// 告警标题
	Title string
	// 告警内容
	Content string
	// 告警类型
	AlertType string
	// 严重级别
	SeverityLevel string
	// 触发来源
	TriggerSource string
	// 发送选项
	SendOptions *alert.SendOptions
	// 元数据
	Metadata map[string]interface{}
	// 接收时间
	ReceivedTime time.Time
}

// NewAlertQueue 创建告警消息队列
// 使用全局 AlertManager 单例
// 参数:
//   - ctx: 上下文
//   - db: 数据库连接
//   - tenantId: 租户ID
//
// 返回:
//   - *AlertQueue: 队列实例
func NewAlertQueue(ctx context.Context, db database.Database, tenantId string) *AlertQueue {
	return &AlertQueue{
		messageChan:  make(chan *AlertMessage, 1000), // 默认容量1000
		db:           db,
		tenantId:     tenantId,
		configLoader: loader.NewAlertConfigLoader(db, tenantId),
		stopChan:     make(chan struct{}),
		running:      false,
		config: QueueConfig{
			Capacity:       1000,
			Workers:        5,
			BatchSize:      10,
			BatchTimeout:   time.Second * 5,
			AutoSaveRecord: true,
		},
	}
}

// GetGlobalQueue 获取全局告警队列实例（单例）
var (
	globalQueue   *AlertQueue
	globalQueueMu sync.RWMutex
)

// SetGlobalQueue 设置全局告警队列实例
// 在初始化时调用，设置全局队列
func SetGlobalQueue(queue *AlertQueue) {
	globalQueueMu.Lock()
	defer globalQueueMu.Unlock()
	globalQueue = queue
}

// GetGlobalQueue 获取全局告警队列实例
// 如果队列未初始化，返回 nil
func GetGlobalQueue() *AlertQueue {
	globalQueueMu.RLock()
	defer globalQueueMu.RUnlock()
	return globalQueue
}

// SetConfig 设置队列配置
func (q *AlertQueue) SetConfig(config QueueConfig) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.config = config
}

// Start 启动队列处理
func (q *AlertQueue) Start(ctx context.Context) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.running {
		return fmt.Errorf("队列已经在运行")
	}

	logger.InfoWithTrace(ctx, "启动告警消息队列", "workers", q.config.Workers)

	// 启动工作协程
	for i := 0; i < q.config.Workers; i++ {
		q.wg.Add(1)
		go q.worker(i)
	}

	q.running = true
	logger.InfoWithTrace(ctx, "告警消息队列已启动")
	return nil
}

// Stop 停止队列处理
func (q *AlertQueue) Stop(ctx context.Context) error {
	q.mu.Lock()
	if !q.running {
		q.mu.Unlock()
		return fmt.Errorf("队列未运行")
	}
	q.mu.Unlock()

	logger.InfoWithTrace(ctx, "停止告警消息队列")

	// 发送停止信号
	close(q.stopChan)

	// 等待所有工作协程结束
	q.wg.Wait()

	q.mu.Lock()
	q.running = false
	q.mu.Unlock()

	logger.InfoWithTrace(ctx, "告警消息队列已停止")
	return nil
}

// Push 推送告警消息到队列
// 参数:
//   - ctx: 上下文
//   - msg: 告警消息
//
// 返回:
//   - error: 推送错误
func (q *AlertQueue) Push(ctx context.Context, msg *AlertMessage) error {
	q.mu.RLock()
	if !q.running {
		q.mu.RUnlock()
		return fmt.Errorf("队列未运行")
	}
	q.mu.RUnlock()

	// 设置接收时间
	if msg.ReceivedTime.IsZero() {
		msg.ReceivedTime = time.Now()
	}

	// 生成消息ID（如果没有）
	if msg.MessageId == "" {
		msg.MessageId = generateMessageId()
	}

	// 推送到队列
	select {
	case q.messageChan <- msg:
		logger.DebugWithTrace(ctx, "告警消息已推送到队列",
			"messageId", msg.MessageId,
			"title", msg.Title)
		return nil
	case <-time.After(time.Second * 5):
		return fmt.Errorf("推送告警消息超时，队列可能已满")
	}
}

// worker 工作协程
func (q *AlertQueue) worker(workerId int) {
	defer q.wg.Done()

	logger.Info("告警队列工作协程启动", "workerId", workerId)

	for {
		select {
		case <-q.stopChan:
			logger.Info("告警队列工作协程退出", "workerId", workerId)
			return

		case msg := <-q.messageChan:
			q.processMessage(context.Background(), msg, workerId)
		}
	}
}

// processMessage 处理单个告警消息
func (q *AlertQueue) processMessage(ctx context.Context, msg *AlertMessage, workerId int) {
	startTime := time.Now()

	logger.InfoWithTrace(ctx, "开始处理告警消息",
		"workerId", workerId,
		"messageId", msg.MessageId,
		"title", msg.Title,
		"channels", msg.ChannelIds)

	// 1. 创建告警记录
	var recordId string
	if q.config.AutoSaveRecord {
		var err error
		recordId, err = q.saveAlertRecord(ctx, msg, "pending")
		if err != nil {
			logger.ErrorWithTrace(ctx, "保存告警记录失败",
				"messageId", msg.MessageId,
				"error", err)
			// 继续发送，不因为保存失败而中断
		}
	}

	// 2. 解析渠道ID列表
	channelIds := strings.Split(msg.ChannelIds, ",")
	if len(channelIds) == 0 {
		logger.WarnWithTrace(ctx, "告警消息没有指定渠道",
			"messageId", msg.MessageId)
		q.updateRecordStatus(ctx, recordId, "failed", "没有指定渠道", 0)
		return
	}

	// 3. 构建告警消息
	alertMsg := &alert.Message{
		Title:     msg.Title,
		Content:   msg.Content,
		Timestamp: msg.ReceivedTime,
		Tags: map[string]string{
			"type":     msg.AlertType,
			"severity": msg.SeverityLevel,
			"source":   msg.TriggerSource,
		},
		Extra: msg.Metadata,
	}

	// 4. 使用全局管理器发送到所有指定渠道
	mgr := alert.GetGlobalManager()
	results := mgr.SendToMultiple(ctx, channelIds, alertMsg, msg.SendOptions)

	// 5. 处理发送结果
	successCount := 0
	failureCount := 0
	var errorMessages []string

	for channelId, result := range results {
		if result.Success {
			successCount++
			logger.InfoWithTrace(ctx, "告警发送成功",
				"messageId", msg.MessageId,
				"channelId", channelId,
				"duration", result.Duration)

			// 更新渠道统计
			if err := q.configLoader.UpdateChannelStatistics(ctx, channelId, result); err != nil {
				logger.WarnWithTrace(ctx, "更新渠道统计失败",
					"channelId", channelId,
					"error", err)
			}
		} else {
			failureCount++
			errorMsg := "未知错误"
			if result.Error != nil {
				errorMsg = result.Error.Error()
			}
			errorMessages = append(errorMessages, fmt.Sprintf("%s: %s", channelId, errorMsg))

			logger.WarnWithTrace(ctx, "告警发送失败",
				"messageId", msg.MessageId,
				"channelId", channelId,
				"error", errorMsg)

			// 更新渠道统计
			if err := q.configLoader.UpdateChannelStatistics(ctx, channelId, result); err != nil {
				logger.WarnWithTrace(ctx, "更新渠道统计失败",
					"channelId", channelId,
					"error", err)
			}
		}
	}

	// 6. 更新告警记录状态
	duration := time.Since(startTime)
	status := "sent"
	if failureCount > 0 {
		if successCount == 0 {
			status = "failed"
		} else {
			status = "partial"
		}
	}

	errorMsgStr := strings.Join(errorMessages, "; ")
	q.updateRecordStatus(ctx, recordId, status, errorMsgStr, duration.Milliseconds())

	logger.InfoWithTrace(ctx, "告警消息处理完成",
		"workerId", workerId,
		"messageId", msg.MessageId,
		"status", status,
		"success", successCount,
		"failure", failureCount,
		"duration", duration)
}

// saveAlertRecord 保存告警记录
func (q *AlertQueue) saveAlertRecord(ctx context.Context, msg *AlertMessage, status string) (string, error) {
	recordId := generateRecordId()

	// 序列化元数据
	var metadataJSON string
	if msg.Metadata != nil {
		data, err := json.Marshal(msg.Metadata)
		if err != nil {
			logger.WarnWithTrace(ctx, "序列化告警元数据失败", "error", err)
		} else {
			metadataJSON = string(data)
		}
	}

	insertSQL := `INSERT INTO HUB_ALERT_RECORD 
		(alertRecordId, tenantId, channelIds, alertTitle, alertContent, 
		 alertType, severityLevel, alertStatus, triggerTime, triggerSource, 
		 notifyStatus, notifyRetryCount, alertMetadata, 
		 addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, 
		 ackFlag, resolveFlag, viewCount, commentCount, escalateFlag, escalateLevel)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	oprSeqFlag := fmt.Sprintf("ALERT_%d", time.Now().Unix())

	_, err := q.db.Exec(ctx, insertSQL, []interface{}{
		recordId, q.tenantId, msg.ChannelIds, msg.Title, msg.Content,
		msg.AlertType, msg.SeverityLevel, alerttypes.AlertStatusOpen, msg.ReceivedTime, msg.TriggerSource,
		status, 0, metadataJSON,
		now, "system", now, "system", oprSeqFlag, 1, "Y",
		"N", "N", 0, 0, "N", 0,
	}, false)

	if err != nil {
		return "", fmt.Errorf("插入告警记录失败: %w", err)
	}

	return recordId, nil
}

// updateRecordStatus 更新告警记录状态
func (q *AlertQueue) updateRecordStatus(ctx context.Context, recordId, status, errorMsg string, durationMs int64) {
	if recordId == "" {
		return
	}

	now := time.Now()
	updateSQL := `UPDATE HUB_ALERT_RECORD SET 
		notifyStatus = ?, notifySendTime = ?, notifyCompleteTime = ?,
		notifyDurationMillis = ?, notifyErrorMsg = ?, editTime = ?
		WHERE tenantId = ? AND alertRecordId = ?`

	_, err := q.db.Exec(ctx, updateSQL, []interface{}{
		status, now, now, durationMs, errorMsg, now,
		q.tenantId, recordId,
	}, false)

	if err != nil {
		logger.ErrorWithTrace(ctx, "更新告警记录状态失败",
			"recordId", recordId,
			"error", err)
	}
}

// GetQueueSize 获取队列大小
func (q *AlertQueue) GetQueueSize() int {
	return len(q.messageChan)
}

// IsRunning 检查队列是否运行中
func (q *AlertQueue) IsRunning() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.running
}

// generateMessageId 生成消息ID
func generateMessageId() string {
	now := time.Now()
	return fmt.Sprintf("MSG_%d_%d", now.Unix(), now.Nanosecond()%1000000)
}

// generateRecordId 生成记录ID
func generateRecordId() string {
	now := time.Now()
	return fmt.Sprintf("REC_%d_%d", now.Unix(), now.Nanosecond()%1000000)
}
