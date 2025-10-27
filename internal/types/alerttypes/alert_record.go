package alerttypes

import "time"

// AlertRecord 告警记录表结构（简化版）
// 表名: HUB_ALERT_RECORD
type AlertRecord struct {
	// 主键和租户ID
	AlertRecordId string `json:"alertRecordId" db:"alertRecordId"`
	TenantId      string `json:"tenantId" db:"tenantId"`

	// 关联信息
	ChannelIds   string  `json:"channelIds" db:"channelIds"`     // 使用的渠道ID列表（逗号分隔，支持多渠道）
	ChannelNames *string `json:"channelNames" db:"channelNames"` // 渠道名称列表（冗余字段，逗号分隔）

	// 告警基本信息
	AlertTitle    string  `json:"alertTitle" db:"alertTitle"`       // 告警标题
	AlertContent  string  `json:"alertContent" db:"alertContent"`   // 告警内容
	AlertType     string  `json:"alertType" db:"alertType"`         // 告警类型
	SeverityLevel string  `json:"severityLevel" db:"severityLevel"` // 严重级别
	CategoryName  *string `json:"categoryName" db:"categoryName"`   // 分类名称
	AlertStatus   string  `json:"alertStatus" db:"alertStatus"`     // 告警状态：open/acknowledged/resolved/closed
	AlertTags     *string `json:"alertTags" db:"alertTags"`         // 告警标签（JSON格式）

	// 触发信息
	TriggerTime      time.Time `json:"triggerTime" db:"triggerTime"`           // 触发时间
	TriggerSource    string    `json:"triggerSource" db:"triggerSource"`       // 触发来源：auto/manual/api/schedule/event
	TriggerValue     *string   `json:"triggerValue" db:"triggerValue"`         // 触发值
	TriggerCondition *string   `json:"triggerCondition" db:"triggerCondition"` // 触发条件（快照）
	MonitorTarget    *string   `json:"monitorTarget" db:"monitorTarget"`       // 监控目标
	MetricName       *string   `json:"metricName" db:"metricName"`             // 指标名称
	MetricValue      *string   `json:"metricValue" db:"metricValue"`           // 指标值
	MetricUnit       *string   `json:"metricUnit" db:"metricUnit"`             // 指标单位
	SourceSystem     *string   `json:"sourceSystem" db:"sourceSystem"`         // 来源系统
	SourceHost       *string   `json:"sourceHost" db:"sourceHost"`             // 来源主机

	// 恢复信息
	RecoveryTime  *time.Time `json:"recoveryTime" db:"recoveryTime"`   // 恢复时间
	RecoveryValue *string    `json:"recoveryValue" db:"recoveryValue"` // 恢复时的值
	DurationSecs  *int       `json:"durationSecs" db:"durationSecs"`   // 持续时间（秒）

	// 通知信息
	NotifyTarget         *string    `json:"notifyTarget" db:"notifyTarget"`                 // 通知目标：收件人、手机号等
	NotifyStatus         string     `json:"notifyStatus" db:"notifyStatus"`                 // 通知状态：pending/sending/sent/failed/partial
	NotifySendTime       *time.Time `json:"notifySendTime" db:"notifySendTime"`             // 通知发送时间
	NotifyCompleteTime   *time.Time `json:"notifyCompleteTime" db:"notifyCompleteTime"`     // 通知完成时间
	NotifyDurationMillis *int       `json:"notifyDurationMillis" db:"notifyDurationMillis"` // 通知耗时（毫秒）
	NotifyRetryCount     int        `json:"notifyRetryCount" db:"notifyRetryCount"`         // 通知重试次数
	NotifyErrorMsg       *string    `json:"notifyErrorMsg" db:"notifyErrorMsg"`             // 通知错误信息
	NotifyErrorCode      *string    `json:"notifyErrorCode" db:"notifyErrorCode"`           // 通知错误码
	MessageId            *string    `json:"messageId" db:"messageId"`                       // 消息ID（渠道返回）
	ResponseData         *string    `json:"responseData" db:"responseData"`                 // 响应数据（JSON格式）

	// 处理信息
	AckFlag             string     `json:"ackFlag" db:"ackFlag"`                         // 确认标记：Y-已确认，N-未确认
	AckTime             *time.Time `json:"ackTime" db:"ackTime"`                         // 确认时间
	AckUserId           *string    `json:"ackUserId" db:"ackUserId"`                     // 确认人用户ID
	AckUserName         *string    `json:"ackUserName" db:"ackUserName"`                 // 确认人姓名
	AckComment          *string    `json:"ackComment" db:"ackComment"`                   // 确认备注
	ResolveFlag         string     `json:"resolveFlag" db:"resolveFlag"`                 // 解决标记：Y-已解决，N-未解决
	ResolveTime         *time.Time `json:"resolveTime" db:"resolveTime"`                 // 解决时间
	ResolveUserId       *string    `json:"resolveUserId" db:"resolveUserId"`             // 解决人用户ID
	ResolveUserName     *string    `json:"resolveUserName" db:"resolveUserName"`         // 解决人姓名
	ResolveComment      *string    `json:"resolveComment" db:"resolveComment"`           // 解决备注
	ResolveDurationSecs *int       `json:"resolveDurationSecs" db:"resolveDurationSecs"` // 解决耗时（秒）
	CloseTime           *time.Time `json:"closeTime" db:"closeTime"`                     // 关闭时间
	CloseUserId         *string    `json:"closeUserId" db:"closeUserId"`                 // 关闭人用户ID
	CloseUserName       *string    `json:"closeUserName" db:"closeUserName"`             // 关闭人姓名

	// 元数据和追踪
	AlertMetadata    *string `json:"alertMetadata" db:"alertMetadata"`       // 告警元数据（JSON格式）
	RelatedRecordIds *string `json:"relatedRecordIds" db:"relatedRecordIds"` // 关联的记录ID（逗号分隔）
	RequestId        *string `json:"requestId" db:"requestId"`               // 请求ID（用于追踪）
	TraceId          *string `json:"traceId" db:"traceId"`                   // 追踪ID
	BatchId          *string `json:"batchId" db:"batchId"`                   // 批次ID（同一批发送）
	SequenceNum      *int    `json:"sequenceNum" db:"sequenceNum"`           // 序列号

	// 统计字段
	ViewCount     int        `json:"viewCount" db:"viewCount"`         // 查看次数
	LastViewTime  *time.Time `json:"lastViewTime" db:"lastViewTime"`   // 最后查看时间
	CommentCount  int        `json:"commentCount" db:"commentCount"`   // 评论数
	EscalateFlag  string     `json:"escalateFlag" db:"escalateFlag"`   // 升级标记：Y-已升级，N-未升级
	EscalateTime  *time.Time `json:"escalateTime" db:"escalateTime"`   // 升级时间
	EscalateLevel int        `json:"escalateLevel" db:"escalateLevel"` // 升级级别

	// 通用字段
	AddTime        time.Time `json:"addTime" db:"addTime"`
	AddWho         string    `json:"addWho" db:"addWho"`
	EditTime       time.Time `json:"editTime" db:"editTime"`
	EditWho        string    `json:"editWho" db:"editWho"`
	OprSeqFlag     string    `json:"oprSeqFlag" db:"oprSeqFlag"`
	CurrentVersion int       `json:"currentVersion" db:"currentVersion"`
	ActiveFlag     string    `json:"activeFlag" db:"activeFlag"`
	NoteText       *string   `json:"noteText" db:"noteText"`
	Reserved1      *string   `json:"reserved1" db:"reserved1"`
	Reserved2      *string   `json:"reserved2" db:"reserved2"`
	Reserved3      *string   `json:"reserved3" db:"reserved3"`
}

// TableName 指定表名
func (AlertRecord) TableName() string {
	return "HUB_ALERT_RECORD"
}
