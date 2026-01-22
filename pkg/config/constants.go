package config

// 系统配置键常量
// 用于统一管理配置键，便于全局查找和维护

// =============================================================================
// 告警服务配置 (app.alert.*)
// =============================================================================

const (
	// ALERT_POLL_INTERVAL 告警轮询间隔配置键
	// 默认值: "3s"
	// 说明: 后台 worker 轮询数据库获取待发送告警日志的间隔时间
	ALERT_POLL_INTERVAL = "app.alert.poll_interval"

	// ALERT_BATCH_SIZE 告警批处理大小配置键
	// 默认值: 50
	// 说明: 每次轮询从数据库获取的待发送告警日志数量
	ALERT_BATCH_SIZE = "app.alert.batch_size"

	// ALERT_CLEANUP_INTERVAL 告警清理间隔配置键
	// 默认值: "1h"
	// 说明: 清理 worker 执行清理任务的间隔时间
	ALERT_CLEANUP_INTERVAL = "app.alert.cleanup.interval"

	// ALERT_LOG_RETENTION_HOURS 告警日志保留时间配置键（小时）
	// 默认值: 168 (7天)
	// 说明: 告警日志在数据库中的保留时间，超过此时间的日志将被清理
	ALERT_LOG_RETENTION_HOURS = "app.alert.cleanup.log_retention_hours"

	// ALERT_LOG_QUEUE_SIZE 告警日志队列大小配置键
	// 默认值: 1000
	// 说明: 告警日志异步写入队列的缓冲区大小
	ALERT_LOG_QUEUE_SIZE = "app.alert.log.queue_size"

	// ALERT_LOG_BATCH_SIZE 告警日志批量写入大小配置键
	// 默认值: 100
	// 说明: 告警日志批量写入数据库的批次大小
	ALERT_LOG_BATCH_SIZE = "app.alert.log.batch_size"

	// ALERT_LOG_FLUSH_INTERVAL 告警日志刷新间隔配置键
	// 默认值: "5s"
	// 说明: 告警日志批量缓冲区定时刷新的间隔时间
	ALERT_LOG_FLUSH_INTERVAL = "app.alert.log.flush_interval"
)

// =============================================================================
// 集群服务配置 (app.cluster.*)
// =============================================================================

const (
	// CLUSTER_NODE_ID 集群节点ID配置键
	// 说明: 集群模块专用的节点ID配置（优先级最高）
	CLUSTER_NODE_ID = "app.cluster.node_id"

	// CLUSTER_EVENT_POLL_INTERVAL 集群事件轮询间隔配置键
	// 默认值: "3s"
	CLUSTER_EVENT_POLL_INTERVAL = "app.cluster.event.poll_interval"

	// CLUSTER_EVENT_BATCH_SIZE 集群事件批处理大小配置键
	// 默认值: 100
	CLUSTER_EVENT_BATCH_SIZE = "app.cluster.event.batch_size"

	// CLUSTER_EVENT_EXPIRE_HOURS 集群事件过期时间配置键（小时）
	// 默认值: 24
	CLUSTER_EVENT_EXPIRE_HOURS = "app.cluster.event.expire_hours"

	// CLUSTER_CLEANUP_ENABLED 集群清理是否启用配置键
	// 默认值: true
	CLUSTER_CLEANUP_ENABLED = "app.cluster.cleanup.enabled"

	// CLUSTER_CLEANUP_INTERVAL 集群清理间隔配置键
	// 默认值: "1h"
	CLUSTER_CLEANUP_INTERVAL = "app.cluster.cleanup.interval"

	// CLUSTER_CLEANUP_ACK_RETENTION_HOURS 集群确认记录保留时间配置键（小时）
	// 默认值: 48
	CLUSTER_CLEANUP_ACK_RETENTION_HOURS = "app.cluster.cleanup.ack_retention_hours"
)

// =============================================================================
// 应用基础配置 (app.*)
// =============================================================================

const (
	// APP_NODE_ID 应用节点ID配置键
	// 说明: 全局节点ID配置（次优先级，会被 app.cluster.node_id 覆盖）
	APP_NODE_ID = "app.node_id"
)
