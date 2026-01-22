package types

import "time"

// AlertConfig 告警渠道配置
// 对应数据库表：HUB_ALERT_CONFIG
type AlertConfig struct {
	// 主键和租户
	TenantId    string `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`             // 租户ID，主键
	ChannelName string `json:"channelName" form:"channelName" query:"channelName" db:"channelName"` // 渠道名称，主键

	// 渠道基本信息
	ChannelType         string  `json:"channelType" form:"channelType" query:"channelType" db:"channelType"`                                 // 渠道类型：email/qq/wechat_work/dingtalk/webhook/sms
	ChannelDesc         *string `json:"channelDesc" form:"channelDesc" query:"channelDesc" db:"channelDesc"`                                 // 渠道描述
	ActiveFlag          string  `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag"`                                     // 启用状态：Y-启用，N-禁用
	DefaultFlag         string  `json:"defaultFlag" form:"defaultFlag" query:"defaultFlag" db:"defaultFlag"`                                 // 是否默认渠道：Y-是，N-否
	PriorityLevel       int     `json:"priorityLevel" form:"priorityLevel" query:"priorityLevel" db:"priorityLevel"`                         // 优先级：1-10，数字越小优先级越高
	DefaultTemplateName *string `json:"defaultTemplateName" form:"defaultTemplateName" query:"defaultTemplateName" db:"defaultTemplateName"` // 默认关联的模板名称

	// 服务器配置（JSON格式）
	ServerConfig *string `json:"serverConfig" form:"serverConfig" query:"serverConfig" db:"serverConfig"` // 服务器配置，JSON格式，如SMTP配置、Webhook URL等
	SendConfig   *string `json:"sendConfig" form:"sendConfig" query:"sendConfig" db:"sendConfig"`         // 发送配置，JSON格式，如默认收件人、超时设置等

	// 消息格式配置
	MessageContentFormat *string `json:"messageContentFormat" form:"messageContentFormat" query:"messageContentFormat" db:"messageContentFormat"` // 消息内容格式：text/html/markdown

	// 重试和超时配置
	TimeoutSeconds    int    `json:"timeoutSeconds" form:"timeoutSeconds" query:"timeoutSeconds" db:"timeoutSeconds"`             // 超时时间（秒）
	RetryCount        int    `json:"retryCount" form:"retryCount" query:"retryCount" db:"retryCount"`                             // 重试次数
	RetryIntervalSecs int    `json:"retryIntervalSecs" form:"retryIntervalSecs" query:"retryIntervalSecs" db:"retryIntervalSecs"` // 重试间隔（秒）
	AsyncSendFlag     string `json:"asyncSendFlag" form:"asyncSendFlag" query:"asyncSendFlag" db:"asyncSendFlag"`                 // 是否异步发送：Y-是，N-否

	// 统计信息
	TotalSentCount   int64      `json:"totalSentCount" form:"totalSentCount" query:"totalSentCount" db:"totalSentCount"`         // 总发送次数
	SuccessCount     int64      `json:"successCount" form:"successCount" query:"successCount" db:"successCount"`                 // 成功次数
	FailureCount     int64      `json:"failureCount" form:"failureCount" query:"failureCount" db:"failureCount"`                 // 失败次数
	LastSendTime     *time.Time `json:"lastSendTime" form:"lastSendTime" query:"lastSendTime" db:"lastSendTime"`                 // 最后发送时间
	LastSuccessTime  *time.Time `json:"lastSuccessTime" form:"lastSuccessTime" query:"lastSuccessTime" db:"lastSuccessTime"`     // 最后成功时间
	LastFailureTime  *time.Time `json:"lastFailureTime" form:"lastFailureTime" query:"lastFailureTime" db:"lastFailureTime"`     // 最后失败时间
	LastErrorMessage *string    `json:"lastErrorMessage" form:"lastErrorMessage" query:"lastErrorMessage" db:"lastErrorMessage"` // 最后错误信息

	// 通用字段
	AddTime        time.Time `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`                             // 创建时间
	AddWho         string    `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`                                 // 创建人ID
	EditTime       time.Time `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`                         // 最后修改时间
	EditWho        string    `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`                             // 最后修改人ID
	OprSeqFlag     string    `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`                 // 操作序列标识
	CurrentVersion int       `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"` // 当前版本号
	NoteText       *string   `json:"noteText" form:"noteText" query:"noteText" db:"noteText"`                         // 备注信息
	ExtProperty    *string   `json:"extProperty" form:"extProperty" query:"extProperty" db:"extProperty"`             // 扩展属性，JSON格式

	// 预留字段
	Reserved1  *string `json:"reserved1" form:"reserved1" query:"reserved1" db:"reserved1"`     // 预留字段1
	Reserved2  *string `json:"reserved2" form:"reserved2" query:"reserved2" db:"reserved2"`     // 预留字段2
	Reserved3  *string `json:"reserved3" form:"reserved3" query:"reserved3" db:"reserved3"`     // 预留字段3
	Reserved4  *string `json:"reserved4" form:"reserved4" query:"reserved4" db:"reserved4"`     // 预留字段4
	Reserved5  *string `json:"reserved5" form:"reserved5" query:"reserved5" db:"reserved5"`     // 预留字段5
	Reserved6  *string `json:"reserved6" form:"reserved6" query:"reserved6" db:"reserved6"`     // 预留字段6
	Reserved7  *string `json:"reserved7" form:"reserved7" query:"reserved7" db:"reserved7"`     // 预留字段7
	Reserved8  *string `json:"reserved8" form:"reserved8" query:"reserved8" db:"reserved8"`     // 预留字段8
	Reserved9  *string `json:"reserved9" form:"reserved9" query:"reserved9" db:"reserved9"`     // 预留字段9
	Reserved10 *string `json:"reserved10" form:"reserved10" query:"reserved10" db:"reserved10"` // 预留字段10
}
