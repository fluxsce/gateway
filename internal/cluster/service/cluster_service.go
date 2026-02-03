package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gateway/internal/cluster/dao"
	"gateway/internal/cluster/types"
	"gateway/pkg/config"
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/pkg/utils/random"
)

// ClusterServiceImpl 集群服务实现
// 提供集群节点间的事件发布/订阅机制，支持配置变更、缓存失效等场景的消息传递
type ClusterServiceImpl struct {
	// 配置
	tenantId          string        // 租户ID，用于多租户隔离
	nodeId            string        // 当前节点ID，用于标识节点身份
	nodeIp            string        // 当前节点IP地址
	pollInterval      time.Duration // 事件轮询间隔
	batchSize         int           // 每次轮询获取的事件数量
	expireHours       int           // 事件过期时间（小时）
	cleanupEnabled    bool          // 是否启用自动清理
	cleanupInterval   time.Duration // 清理任务执行间隔
	ackRetentionHours int           // 确认记录保留时间（小时）

	// 组件
	dao      *dao.EventDAO                 // 数据访问层
	handlers map[string]types.EventHandler // 事件处理器映射表（eventType -> handler）

	// 状态
	running       bool      // 服务是否正在运行（包含启动中和就绪状态）
	lastEventTime time.Time // 上次处理的事件时间（使用 >= 查询，依赖 ACK 表去重）
	mu            sync.RWMutex

	// 控制
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewClusterService 创建集群服务实例
//
// 参数:
//   - db: 数据库实例，用于事件的持久化存储
//
// 返回:
//   - *ClusterServiceImpl: 集群服务实例
//
// 配置项:
//   - app.cluster.node_id: 节点ID（优先级最高）
//   - app.node_id: 全局节点ID（次优先级）
//   - app.cluster.event.poll_interval: 事件轮询间隔，默认3s
//   - app.cluster.event.batch_size: 每批处理事件数，默认100
//   - app.cluster.event.expire_hours: 事件过期时间（小时），默认24
//   - app.cluster.cleanup.enabled: 是否启用清理，默认true
//   - app.cluster.cleanup.interval: 清理间隔，默认1h
//   - app.cluster.cleanup.ack_retention_hours: 确认记录保留时间（小时），默认48
func NewClusterService(db database.Database) *ClusterServiceImpl {
	// 读取节点ID配置（优先级：app.cluster.node_id > app.node_id > 自动生成）
	nodeId := getNodeId()
	// 使用 random 包的 IP 获取方法
	nodeIp := random.GetNodeIP()

	// 解析时间配置
	pollInterval := parseDuration(config.GetString("app.cluster.event.poll_interval", "3s"), 3*time.Second)
	batchSize := config.GetInt("app.cluster.event.batch_size", 100)
	expireHours := config.GetInt("app.cluster.event.expire_hours", 24)
	cleanupEnabled := config.GetBool("app.cluster.cleanup.enabled", true)
	cleanupInterval := parseDuration(config.GetString("app.cluster.cleanup.interval", "1h"), 1*time.Hour)
	ackRetentionHours := config.GetInt("app.cluster.cleanup.ack_retention_hours", 48)

	return &ClusterServiceImpl{
		tenantId:          "default", // 使用固定租户ID
		nodeId:            nodeId,
		nodeIp:            nodeIp,
		pollInterval:      pollInterval,
		batchSize:         batchSize,
		expireHours:       expireHours,
		cleanupEnabled:    cleanupEnabled,
		cleanupInterval:   cleanupInterval,
		ackRetentionHours: ackRetentionHours,
		dao:               dao.NewEventDAO(db),
		handlers:          make(map[string]types.EventHandler),
		lastEventTime:     time.Now(), // 从当前时间开始，只处理启动后的新事件
	}
}

// parseDuration 解析时间字符串，解析失败返回默认值
//
// 参数:
//   - s: 时间字符串，如 "3s", "5m", "1h"
//   - defaultValue: 解析失败时的默认值
//
// 返回:
//   - time.Duration: 解析后的时间间隔
func parseDuration(s string, defaultValue time.Duration) time.Duration {
	if s == "" {
		return defaultValue
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return defaultValue
	}
	return d
}

// Start 启动集群服务
//
// 启动事件轮询goroutine和清理任务goroutine（如果启用）
// 该方法是非阻塞的，服务将在后台运行直到调用Stop
//
// 参数:
//   - ctx: 上下文对象，用于控制服务生命周期
//
// 返回:
//   - error: 如果服务已在运行，返回错误；否则返回nil
func (s *ClusterServiceImpl) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("集群服务已在运行")
	}
	s.ctx, s.cancel = context.WithCancel(ctx)
	s.running = true
	s.mu.Unlock()

	logger.Info("集群服务启动", "nodeId", s.nodeId, "nodeIp", s.nodeIp, "tenantId", s.tenantId)

	// 启动事件轮询
	s.wg.Add(1)
	go s.eventPollLoop()

	// 启动清理任务(如果启用)
	if s.cleanupEnabled {
		s.wg.Add(1)
		go s.cleanupLoop()
	}

	logger.Info("集群服务启动完成")
	return nil
}

// Stop 停止集群服务
//
// 优雅地停止所有后台goroutine，等待当前处理的事件完成
//
// 参数:
//   - ctx: 上下文对象，用于控制停止超时
//
// 返回:
//   - error: 总是返回nil
func (s *ClusterServiceImpl) Stop(ctx context.Context) error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = false
	s.mu.Unlock()

	logger.Info("集群服务停止中...", "nodeId", s.nodeId)

	// 取消上下文
	if s.cancel != nil {
		s.cancel()
	}

	// 等待所有goroutine结束
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("集群服务已停止")
	case <-ctx.Done():
		logger.Warn("集群服务停止超时")
	}

	return nil
}

// Close 实现io.Closer接口，用于资源释放
//
// 返回:
//   - error: 总是返回nil
func (s *ClusterServiceImpl) Close() error {
	return s.Stop(context.Background())
}

// PublishEvent 发布集群事件
//
// 将事件发布到集群中，其他节点将通过轮询机制接收并处理该事件
//
// 参数:
//   - ctx: 上下文对象
//   - event: 事件对象，必须包含EventType和EventAction，其他字段将自动填充
//
// 返回:
//   - error: 发布失败时返回错误
func (s *ClusterServiceImpl) PublishEvent(ctx context.Context, event *types.ClusterEvent) error {
	// 填充事件基本信息
	now := time.Now()
	if event.EventId == "" {
		// 生成事件ID，长度32位（evt_ + 28位随机字符）
		// 数据库eventId字段为VARCHAR(64)，32位长度安全
		event.EventId = random.GenerateUniqueStringWithPrefix("evt_", 32)
	}

	// 验证EventId长度，防止超出数据库字段限制
	if len(event.EventId) > 64 {
		return fmt.Errorf("事件ID长度超过限制: %d > 64", len(event.EventId))
	}
	event.TenantId = s.tenantId
	event.SourceNodeId = s.nodeId
	event.SourceNodeIp = s.nodeIp
	event.EventTime = now
	event.AddTime = now
	event.EditTime = now
	event.AddWho = s.nodeId
	event.EditWho = s.nodeId
	// OprSeqFlag: 操作序列标识，使用32位随机字符串确保唯一性
	// 数据库字段长度为VARCHAR(64)，32位长度安全
	event.OprSeqFlag = random.Generate32BitRandomString()
	event.CurrentVersion = 1
	event.ActiveFlag = "Y"

	// 设置过期时间
	if event.ExpireTime == nil && s.expireHours > 0 {
		expireTime := now.Add(time.Duration(s.expireHours) * time.Hour)
		event.ExpireTime = &expireTime
	}

	// 保存事件
	if err := s.dao.SaveEvent(ctx, event); err != nil {
		return fmt.Errorf("发布事件失败: %w", err)
	}

	logger.Debug("集群事件已发布", "eventId", event.EventId, "eventType", event.EventType, "eventAction", event.EventAction)
	return nil
}

// RegisterHandler 注册事件处理器
//
// 为指定类型的事件注册处理器，当接收到该类型事件时将调用处理器的Handle方法
// 每个事件类型只能注册一个处理器，重复注册将覆盖之前的处理器
//
// 参数:
//   - eventType: 事件类型，如 "ROUTE_CONFIG", "CACHE_REFRESH" 等
//   - handler: 事件处理器实例
func (s *ClusterServiceImpl) RegisterHandler(handler types.EventHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	eventType := handler.GetEventType()
	s.handlers[eventType] = handler
	logger.Info("注册集群事件处理器", "eventType", eventType)
}

// UnregisterHandler 取消注册事件处理器
//
// 移除指定类型事件的处理器，之后该类型事件将被跳过不处理
//
// 参数:
//   - eventType: 事件类型
func (s *ClusterServiceImpl) UnregisterHandler(eventType string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.handlers, eventType)
	logger.Info("取消注册集群事件处理器", "eventType", eventType)
}

// GetNodeId 获取当前节点ID
//
// 返回:
//   - string: 当前节点的唯一标识符
func (s *ClusterServiceImpl) GetNodeId() string {
	return s.nodeId
}

// IsReady 检查服务是否就绪
//
// 返回:
//   - bool: true表示服务已启动并就绪，可以接收和处理事件
func (s *ClusterServiceImpl) IsReady() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// eventPollLoop 事件轮询循环（后台goroutine）
//
// 定期轮询数据库获取待处理事件并分发给相应的处理器
func (s *ClusterServiceImpl) eventPollLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	logger.Info("集群事件轮询启动", "interval", s.pollInterval)

	for {
		select {
		case <-s.ctx.Done():
			logger.Info("集群事件轮询停止")
			return
		case <-ticker.C:
			s.pollAndProcessEvents()
		}
	}
}

// pollAndProcessEvents 轮询并处理事件
//
// 从数据库获取自上次处理后的新事件，并逐个调用处理器处理
// 使用 lastEventTime 追踪，配合 ACK 表自动去重，避免重复处理
func (s *ClusterServiceImpl) pollAndProcessEvents() {
	ctx := context.Background()

	// 获取待处理事件
	// 注意：使用 >= lastEventTime 查询，可能返回已处理的事件
	// 但 NOT EXISTS 子查询会自动过滤掉已确认的事件，实现自然去重
	events, err := s.dao.GetPendingEvents(ctx, s.tenantId, s.nodeId, s.batchSize, s.lastEventTime)
	if err != nil {
		logger.Error("获取待处理事件失败", "error", err)
		return
	}

	if len(events) == 0 {
		return
	}

	logger.Debug("获取到待处理事件", "count", len(events))

	// 处理每个事件
	for _, event := range events {
		s.processEvent(ctx, event)

		// 更新最后处理时间
		// 使用事件中最大的时间作为下次查询的起点
		// 注意：使用 >= 查询，所以 lastEventTime 这一秒的事件会被重复查询
		// 但已处理的会被 ACK 表过滤掉，不会重复处理
		if event.EventTime.After(s.lastEventTime) {
			s.lastEventTime = event.EventTime
		}
	}
}

// processEvent 处理单个事件
//
// 检查事件是否过期，查找并调用对应的处理器，记录处理结果
//
// 参数:
//   - ctx: 上下文对象
//   - event: 待处理的事件
//
// 处理逻辑：
//  1. 检查事件是否过期 -> SKIPPED
//  2. 查找对应的处理器 -> 如果没有，不记录 ACK（等待处理器注册）
//  3. 调用处理器处理 -> 根据返回结果记录 ACK
//  4. RETRY 状态不记录 ACK，下次轮询会重新处理
func (s *ClusterServiceImpl) processEvent(ctx context.Context, event *types.ClusterEvent) {
	startTime := time.Now()

	// 检查事件是否过期
	if event.IsExpired() {
		s.saveAckResult(ctx, event, "SKIPPED", "事件已过期", 0)
		return
	}

	// 获取处理器
	s.mu.RLock()
	handler, exists := s.handlers[event.EventType]
	s.mu.RUnlock()

	if !exists {
		// 没有处理器，跳过
		s.saveAckResult(ctx, event, "SKIPPED", "无对应处理器", 0)
		return
	}

	// 执行处理
	result := handler.Handle(ctx, event)
	duration := time.Since(startTime).Milliseconds()

	// 根据处理结果记录 ACK
	switch result.Status {
	case types.HandleStatusSuccess:
		logger.Debug("处理集群事件成功",
			"eventId", event.EventId,
			"eventType", event.EventType,
			"eventAction", event.EventAction,
			"duration", duration,
			"message", result.Message)
		s.saveAckResult(ctx, event, "SUCCESS", result.Message, duration)

	case types.HandleStatusFailed:
		logger.Error("处理集群事件失败",
			"eventId", event.EventId,
			"eventType", event.EventType,
			"eventAction", event.EventAction,
			"error", result.Error,
			"message", result.Message)
		s.saveAckResult(ctx, event, "FAILED", result.Message, duration)

	case types.HandleStatusRetry:
		// 需要重试，不记录 ACK，下次轮询会重新处理
		logger.Warn("处理集群事件需要重试",
			"eventId", event.EventId,
			"eventType", event.EventType,
			"eventAction", event.EventAction,
			"message", result.Message)
		// 不调用 saveAckResult，让事件在下次轮询时重新处理

	case types.HandleStatusSkipped:
		logger.Debug("跳过处理集群事件",
			"eventId", event.EventId,
			"eventType", event.EventType,
			"eventAction", event.EventAction,
			"message", result.Message)
		s.saveAckResult(ctx, event, "SKIPPED", result.Message, duration)

	default:
		// 未知状态，记录为失败
		logger.Error("未知的处理结果状态",
			"eventId", event.EventId,
			"eventType", event.EventType,
			"status", result.Status)
		s.saveAckResult(ctx, event, "FAILED", "未知的处理结果状态", duration)
	}
}

// saveAckResult 保存事件处理结果
//
// 将事件处理结果保存到确认记录表，用于追踪各节点的处理状态
//
// 参数:
//   - ctx: 上下文对象
//   - event: 被处理的事件
//   - status: 处理状态，如 "SUCCESS", "FAILED", "SKIPPED"
//   - message: 结果消息
//   - duration: 处理耗时（毫秒）
func (s *ClusterServiceImpl) saveAckResult(ctx context.Context, event *types.ClusterEvent, status, message string, duration int64) {
	now := time.Now()
	ack := &types.ClusterEventAck{
		AckId:         random.GenerateUniqueStringWithPrefix("ack_", 32),
		TenantId:      s.tenantId,
		EventId:       event.EventId,
		NodeId:        s.nodeId,
		NodeIp:        s.nodeIp,
		AckStatus:     status,
		ProcessTime:   &now,
		ResultMessage: message,
		RetryCount:    0,
		AddTime:       now,
		AddWho:        s.nodeId,
		EditTime:      now,
		EditWho:       s.nodeId,
		// OprSeqFlag: 操作序列标识，使用32位随机字符串确保唯一性
		// 数据库字段长度为VARCHAR(64)，32位长度安全
		OprSeqFlag:     random.Generate32BitRandomString(),
		CurrentVersion: 1,
		ActiveFlag:     "Y",
	}

	if err := s.dao.SaveEventAck(ctx, ack); err != nil {
		logger.Error("保存事件确认失败", "eventId", event.EventId, "error", err)
	}
}

// cleanupLoop 清理循环（后台goroutine）
//
// 定期清理过期事件和旧的确认记录，防止数据库无限增长
func (s *ClusterServiceImpl) cleanupLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.cleanupInterval)
	defer ticker.Stop()

	logger.Info("集群事件清理任务启动", "interval", s.cleanupInterval)

	for {
		select {
		case <-s.ctx.Done():
			logger.Info("集群事件清理任务停止")
			return
		case <-ticker.C:
			s.cleanup()
		}
	}
}

// cleanup 执行清理操作
//
// 清理过期事件和超过保留期的确认记录
func (s *ClusterServiceImpl) cleanup() {
	ctx := context.Background()

	// 清理过期事件
	expireTime := time.Now().Add(-time.Duration(s.expireHours) * time.Hour)
	if affected, err := s.dao.CleanupExpiredEvents(ctx, s.tenantId, expireTime); err != nil {
		logger.Error("清理过期事件失败", "error", err)
	} else if affected > 0 {
		logger.Info("清理过期事件", "count", affected)
	}

	// 清理旧确认记录
	ackExpireTime := time.Now().Add(-time.Duration(s.ackRetentionHours) * time.Hour)
	if affected, err := s.dao.CleanupOldAcks(ctx, s.tenantId, ackExpireTime); err != nil {
		logger.Error("清理旧确认记录失败", "error", err)
	} else if affected > 0 {
		logger.Info("清理旧确认记录", "count", affected)
	}
}

// getNodeId 获取节点ID
//
// 优先级：
//  1. app.cluster.node_id 配置（集群模块专用）
//  2. config.GetNodeId() 统一方法（包含完整的fallback逻辑）
//     - app.node_id 配置
//     - 环境变量（GATEWAY_NODE_ID、POD_NAME）
//     - 持久化文件 .node_id
//     - 基于主机名和MAC地址的哈希生成
//
// 返回:
//   - string: 节点唯一标识符
func getNodeId() string {
	// 1. 优先使用 cluster 模块单独配置
	nodeId := config.GetString("app.cluster.node_id", "")
	if nodeId != "" {
		return nodeId
	}

	// 2. 使用 config 包的统一节点ID获取方法
	// 该方法包含完整的fallback逻辑：
	// - app.node_id 配置
	// - 环境变量（GATEWAY_NODE_ID、POD_NAME）
	// - 持久化文件 .node_id
	// - 基于主机名和MAC地址的哈希生成
	return config.GetNodeId()
}
