package types

import "time"

// AlertLog 告警日志
// 对应数据库表：HUB_ALERT_LOG
type AlertLog struct {
	// 主键和租户
	TenantId   string `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`         // 租户ID，主键
	AlertLogId string `json:"alertLogId" form:"alertLogId" query:"alertLogId" db:"alertLogId"` // 告警日志ID，主键

	// 告警基本信息
	AlertLevel     string    `json:"alertLevel" form:"alertLevel" query:"alertLevel" db:"alertLevel"`                 // 告警级别：INFO/WARN/ERROR/CRITICAL
	AlertType      *string   `json:"alertType" form:"alertType" query:"alertType" db:"alertType"`                     // 告警类型，业务自定义类型标识
	AlertTitle     string    `json:"alertTitle" form:"alertTitle" query:"alertTitle" db:"alertTitle"`                 // 告警标题
	AlertContent   *string   `json:"alertContent" form:"alertContent" query:"alertContent" db:"alertContent"`         // 告警内容
	AlertTimestamp time.Time `json:"alertTimestamp" form:"alertTimestamp" query:"alertTimestamp" db:"alertTimestamp"` // 告警时间戳

	// 关联信息
	ChannelName *string `json:"channelName" form:"channelName" query:"channelName" db:"channelName"` // 使用的渠道名称

	// 发送信息
	SendStatus       *string    `json:"sendStatus" form:"sendStatus" query:"sendStatus" db:"sendStatus"`                         // 发送状态：PENDING待发送/SENDING发送中/SUCCESS成功/FAILED失败
	SendTime         *time.Time `json:"sendTime" form:"sendTime" query:"sendTime" db:"sendTime"`                                 // 发送时间
	SendResult       *string    `json:"sendResult" form:"sendResult" query:"sendResult" db:"sendResult"`                         // 发送结果详情，JSON格式
	SendErrorMessage *string    `json:"sendErrorMessage" form:"sendErrorMessage" query:"sendErrorMessage" db:"sendErrorMessage"` // 发送错误信息

	// 标签和扩展信息
	AlertTags  *string `json:"alertTags" form:"alertTags" query:"alertTags" db:"alertTags"`     // 告警标签，JSON格式
	AlertExtra *string `json:"alertExtra" form:"alertExtra" query:"alertExtra" db:"alertExtra"` // 告警额外数据，JSON格式
	TableData  *string `json:"tableData" form:"tableData" query:"tableData" db:"tableData"`     // 表格数据，JSON格式

	// 通用字段
	AddTime        time.Time `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`                             // 创建时间
	AddWho         string    `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`                                 // 创建人ID
	EditTime       time.Time `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`                         // 最后修改时间
	EditWho        string    `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`                             // 最后修改人ID
	OprSeqFlag     string    `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`                 // 操作序列标识
	CurrentVersion int       `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"` // 当前版本号
	ActiveFlag     string    `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag"`                 // 活动状态标记
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
