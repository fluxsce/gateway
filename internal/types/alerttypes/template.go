package alerttypes

import "time"

// AlertTemplate 告警模板表结构
// 表名: HUB_ALERT_TEMPLATE
type AlertTemplate struct {
	// 主键和租户ID
	TemplateId string `json:"templateId" db:"templateId"`
	TenantId   string `json:"tenantId" db:"tenantId"`

	// 模板基本信息
	TemplateName  string  `json:"templateName" db:"templateName"`   // 模板名称
	TemplateDesc  *string `json:"templateDesc" db:"templateDesc"`   // 模板描述
	TemplateType  string  `json:"templateType" db:"templateType"`   // 模板类型：threshold/anomaly/status/business/custom
	SeverityLevel string  `json:"severityLevel" db:"severityLevel"` // 严重级别：critical/high/medium/low/info
	EnabledFlag   string  `json:"enabledFlag" db:"enabledFlag"`     // 启用状态：Y-启用，N-禁用
	CategoryName  *string `json:"categoryName" db:"categoryName"`   // 分类名称：system/service/api/business等
	OwnerUserId   *string `json:"ownerUserId" db:"ownerUserId"`     // 负责人用户ID

	// 告警内容模板
	TitleTemplate   string  `json:"titleTemplate" db:"titleTemplate"`     // 标题模板，支持变量如：{{service}} CPU使用率告警
	ContentTemplate string  `json:"contentTemplate" db:"contentTemplate"` // 内容模板，支持变量如：服务器 {{host}} CPU使用率达到 {{value}}%
	TagsTemplate    *string `json:"tagsTemplate" db:"tagsTemplate"`       // 标签模板（JSON格式）

	// 默认通知渠道配置（可在发送时覆盖）
	DefaultChannelIds *string `json:"defaultChannelIds" db:"defaultChannelIds"` // 默认告警渠道ID列表（逗号分隔），发送时可覆盖
	NotifyRecipients  *string `json:"notifyRecipients" db:"notifyRecipients"`   // 默认收件人配置（JSON格式），发送时可覆盖
	SendConfig        *string `json:"sendConfig" db:"sendConfig"`               // 默认发送配置（JSON格式：超时、重试等），发送时可覆盖

	// 触发条件（可选，用于自动触发）
	TriggerCondition    *string `json:"triggerCondition" db:"triggerCondition"`       // 触发条件（JSON格式）
	TriggerDurationSecs int     `json:"triggerDurationSecs" db:"triggerDurationSecs"` // 持续时间（秒），0表示立即触发
	CheckIntervalSecs   int     `json:"checkIntervalSecs" db:"checkIntervalSecs"`     // 检查间隔（秒），0表示手动触发
	AutoTriggerFlag     string  `json:"autoTriggerFlag" db:"autoTriggerFlag"`         // 自动触发：Y-自动，N-手动
	MonitorTarget       *string `json:"monitorTarget" db:"monitorTarget"`             // 监控目标（用于自动触发）
	MetricName          *string `json:"metricName" db:"metricName"`                   // 指标名称（用于自动触发）

	// 静默和抑制
	SilenceFlag             string     `json:"silenceFlag" db:"silenceFlag"`                         // 静默标记：Y-静默，N-正常
	SilenceStartTime        *time.Time `json:"silenceStartTime" db:"silenceStartTime"`               // 静默开始时间
	SilenceEndTime          *time.Time `json:"silenceEndTime" db:"silenceEndTime"`                   // 静默结束时间
	SilenceReason           *string    `json:"silenceReason" db:"silenceReason"`                     // 静默原因
	RepeatIntervalSecs      int        `json:"repeatIntervalSecs" db:"repeatIntervalSecs"`           // 重复通知间隔（秒）
	DeduplicationFlag       string     `json:"deduplicationFlag" db:"deduplicationFlag"`             // 去重标记：Y-启用去重，N-不去重
	DeduplicationWindowSecs int        `json:"deduplicationWindowSecs" db:"deduplicationWindowSecs"` // 去重时间窗口（秒）

	// 统计信息
	UsageCount      int64      `json:"usageCount" db:"usageCount"`           // 使用次数
	LastUsedTime    *time.Time `json:"lastUsedTime" db:"lastUsedTime"`       // 最后使用时间
	TotalAlertCount int64      `json:"totalAlertCount" db:"totalAlertCount"` // 总告警次数
	LastAlertTime   *time.Time `json:"lastAlertTime" db:"lastAlertTime"`     // 最后告警时间

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
func (AlertTemplate) TableName() string {
	return "HUB_ALERT_TEMPLATE"
}
