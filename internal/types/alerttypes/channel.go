package alerttypes

import "time"

// AlertChannel 告警渠道配置表结构
// 表名: HUB_ALERT_CHANNEL
type AlertChannel struct {
	// 主键和租户ID
	ChannelId string `json:"channelId" db:"channelId"`
	TenantId  string `json:"tenantId" db:"tenantId"`

	// 渠道基本信息
	ChannelName       string  `json:"channelName" db:"channelName"`             // 渠道名称
	ChannelType       string  `json:"channelType" db:"channelType"`             // 渠道类型：email/qq/wechat_work/dingtalk/webhook/sms
	ChannelDesc       *string `json:"channelDesc" db:"channelDesc"`             // 渠道描述
	EnabledFlag       string  `json:"enabledFlag" db:"enabledFlag"`             // 启用状态：Y-启用，N-禁用
	DefaultFlag       string  `json:"defaultFlag" db:"defaultFlag"`             // 是否默认渠道：Y-是，N-否
	PriorityLevel     int     `json:"priorityLevel" db:"priorityLevel"`         // 优先级：1-10，数字越小优先级越高
	CategoryName      *string `json:"categoryName" db:"categoryName"`           // 分类名称：用于分组管理
	DefaultTemplateId *string `json:"defaultTemplateId" db:"defaultTemplateId"` // 默认关联的模板ID（可选）

	// 服务器配置（JSON格式）
	ServerConfig *string `json:"serverConfig" db:"serverConfig"` // 服务器配置：SMTP配置、Webhook URL等
	SendConfig   *string `json:"sendConfig" db:"sendConfig"`     // 发送配置：默认收件人、超时设置等

	// 消息格式配置（可覆盖模板）
	MessageTitlePrefix   *string `json:"messageTitlePrefix" db:"messageTitlePrefix"`     // 消息标题前缀，如：【生产环境】
	MessageTitleSuffix   *string `json:"messageTitleSuffix" db:"messageTitleSuffix"`     // 消息标题后缀
	MessageContentFormat *string `json:"messageContentFormat" db:"messageContentFormat"` // 消息内容格式：text/html/markdown
	CustomStyleConfig    *string `json:"customStyleConfig" db:"customStyleConfig"`       // 自定义样式配置（JSON格式），用于邮件HTML样式等

	// 重试和超时配置
	TimeoutSeconds    int    `json:"timeoutSeconds" db:"timeoutSeconds"`       // 超时时间（秒）
	RetryCount        int    `json:"retryCount" db:"retryCount"`               // 重试次数
	RetryIntervalSecs int    `json:"retryIntervalSecs" db:"retryIntervalSecs"` // 重试间隔（秒）
	AsyncSendFlag     string `json:"asyncSendFlag" db:"asyncSendFlag"`         // 异步发送：Y-是，N-否

	// 限流配置
	RateLimitCount    int `json:"rateLimitCount" db:"rateLimitCount"`       // 限流次数（0表示不限流）
	RateLimitInterval int `json:"rateLimitInterval" db:"rateLimitInterval"` // 限流时间窗口（秒）

	// 统计信息
	TotalSentCount    int64      `json:"totalSentCount" db:"totalSentCount"`       // 总发送次数
	SuccessCount      int64      `json:"successCount" db:"successCount"`           // 成功次数
	FailureCount      int64      `json:"failureCount" db:"failureCount"`           // 失败次数
	LastSendTime      *time.Time `json:"lastSendTime" db:"lastSendTime"`           // 最后发送时间
	LastSuccessTime   *time.Time `json:"lastSuccessTime" db:"lastSuccessTime"`     // 最后成功时间
	LastFailureTime   *time.Time `json:"lastFailureTime" db:"lastFailureTime"`     // 最后失败时间
	LastErrorMessage  *string    `json:"lastErrorMessage" db:"lastErrorMessage"`   // 最后错误信息
	AvgDurationMillis int        `json:"avgDurationMillis" db:"avgDurationMillis"` // 平均耗时（毫秒）

	// 健康检查
	HealthCheckFlag         string     `json:"healthCheckFlag" db:"healthCheckFlag"`                 // 健康检查：Y-健康，N-不健康
	LastHealthCheckTime     *time.Time `json:"lastHealthCheckTime" db:"lastHealthCheckTime"`         // 最后健康检查时间
	HealthCheckIntervalSecs int        `json:"healthCheckIntervalSecs" db:"healthCheckIntervalSecs"` // 健康检查间隔（秒）

	// 通用字段
	AddTime        time.Time `json:"addTime" db:"addTime"`
	AddWho         string    `json:"addWho" db:"addWho"`
	EditTime       time.Time `json:"editTime" db:"editTime"`
	EditWho        string    `json:"editWho" db:"editWho"`
	OprSeqFlag     string    `json:"oprSeqFlag" db:"oprSeqFlag"`
	CurrentVersion int       `json:"currentVersion" db:"currentVersion"`
	ActiveFlag     string    `json:"activeFlag" db:"activeFlag"`
	NoteText       *string   `json:"noteText" db:"noteText"`
	ExtProperty    *string   `json:"extProperty" db:"extProperty"`
	Reserved1      *string   `json:"reserved1" db:"reserved1"`
	Reserved2      *string   `json:"reserved2" db:"reserved2"`
	Reserved3      *string   `json:"reserved3" db:"reserved3"`
	Reserved4      *string   `json:"reserved4" db:"reserved4"`
	Reserved5      *string   `json:"reserved5" db:"reserved5"`
	Reserved6      *string   `json:"reserved6" db:"reserved6"`
	Reserved7      *string   `json:"reserved7" db:"reserved7"`
	Reserved8      *string   `json:"reserved8" db:"reserved8"`
	Reserved9      *string   `json:"reserved9" db:"reserved9"`
	Reserved10     *string   `json:"reserved10" db:"reserved10"`
}

// TableName 指定表名
func (AlertChannel) TableName() string {
	return "HUB_ALERT_CHANNEL"
}
