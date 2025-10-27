package alerttypes

// 渠道类型常量
const (
	ChannelTypeEmail      = "email"       // 邮件渠道
	ChannelTypeQQ         = "qq"          // QQ渠道
	ChannelTypeWeChatWork = "wechat_work" // 企业微信渠道
	ChannelTypeDingTalk   = "dingtalk"    // 钉钉渠道
	ChannelTypeWebhook    = "webhook"     // Webhook渠道
	ChannelTypeSMS        = "sms"         // 短信渠道
)

// 规则类型常量
const (
	RuleTypeThreshold = "threshold" // 阈值告警
	RuleTypeAnomaly   = "anomaly"   // 异常检测
	RuleTypeStatus    = "status"    // 状态变化
	RuleTypeCustom    = "custom"    // 自定义规则
)

// 严重级别常量
const (
	SeverityCritical = "critical" // 严重
	SeverityHigh     = "high"     // 高
	SeverityMedium   = "medium"   // 中
	SeverityLow      = "low"      // 低
	SeverityInfo     = "info"     // 信息
)

// 监控类型常量
const (
	MonitorTypeSystem  = "system"  // 系统监控
	MonitorTypeService = "service" // 服务监控
	MonitorTypeAPI     = "api"     // API监控
	MonitorTypeCustom  = "custom"  // 自定义监控
)

// 告警状态常量
const (
	AlertStatusOpen         = "open"         // 开启
	AlertStatusAcknowledged = "acknowledged" // 已确认
	AlertStatusResolved     = "resolved"     // 已解决
	AlertStatusClosed       = "closed"       // 已关闭
)

// 当前告警状态常量
const (
	CurrentAlertStatusNormal     = "normal"     // 正常
	CurrentAlertStatusAlerting   = "alerting"   // 告警中
	CurrentAlertStatusRecovering = "recovering" // 恢复中
)

// 通知状态常量
const (
	NotifyStatusPending = "pending" // 待发送
	NotifyStatusSent    = "sent"    // 已发送
	NotifyStatusFailed  = "failed"  // 发送失败
	NotifyStatusPartial = "partial" // 部分成功
)

// 发送状态常量
const (
	SendStatusSuccess   = "success"   // 成功
	SendStatusFailed    = "failed"    // 失败
	SendStatusTimeout   = "timeout"   // 超时
	SendStatusCancelled = "cancelled" // 已取消
)

// 消息类型常量
const (
	MessageTypeAlert    = "alert"    // 告警消息
	MessageTypeRecovery = "recovery" // 恢复消息
	MessageTypeReminder = "reminder" // 提醒消息
)

// 触发来源常量
const (
	TriggerSourceAuto     = "auto"     // 自动触发（规则检测）
	TriggerSourceManual   = "manual"   // 手动触发
	TriggerSourceAPI      = "api"      // API调用
	TriggerSourceSchedule = "schedule" // 定时触发
	TriggerSourceEvent    = "event"    // 事件触发
)

// 标志常量
const (
	FlagYes = "Y" // 是
	FlagNo  = "N" // 否
)

// 默认值常量
const (
	DefaultTimeoutSeconds          = 30   // 默认超时时间（秒）
	DefaultRetryCount              = 3    // 默认重试次数
	DefaultRetryIntervalSecs       = 5    // 默认重试间隔（秒）
	DefaultCheckIntervalSecs       = 60   // 默认检查间隔（秒）
	DefaultTriggerDurationSecs     = 60   // 默认触发持续时间（秒）
	DefaultRecoveryDurationSecs    = 60   // 默认恢复持续时间（秒）
	DefaultRepeatIntervalSecs      = 3600 // 默认重复通知间隔（秒）
	DefaultHealthCheckIntervalSecs = 300  // 默认健康检查间隔（秒）
	DefaultPriorityLevel           = 5    // 默认优先级
)
