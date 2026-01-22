package types

import "time"

// AlertTemplate 告警模板
// 对应数据库表：HUB_ALERT_TEMPLATE
type AlertTemplate struct {
	// 主键和租户
	TenantId     string `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                 // 租户ID，主键
	TemplateName string `json:"templateName" form:"templateName" query:"templateName" db:"templateName"` // 模板名称，主键

	// 模板基本信息
	TemplateDesc *string `json:"templateDesc" form:"templateDesc" query:"templateDesc" db:"templateDesc"` // 模板描述
	ChannelType  *string `json:"channelType" form:"channelType" query:"channelType" db:"channelType"`     // 适用的渠道类型：email/qq/wechat_work/dingtalk/webhook/sms，为空表示通用模板

	// 模板内容
	TitleTemplate     *string `json:"titleTemplate" form:"titleTemplate" query:"titleTemplate" db:"titleTemplate"`                 // 标题模板，支持变量占位符如{{.Title}}
	ContentTemplate   *string `json:"contentTemplate" form:"contentTemplate" query:"contentTemplate" db:"contentTemplate"`         // 内容模板，支持变量占位符
	DisplayFormat     string  `json:"displayFormat" form:"displayFormat" query:"displayFormat" db:"displayFormat"`                 // 显示格式：table表格格式/text文本格式
	TemplateVariables *string `json:"templateVariables" form:"templateVariables" query:"templateVariables" db:"templateVariables"` // 模板变量定义，JSON格式，描述可用的变量和说明

	// 附件配置
	AttachmentConfig *string `json:"attachmentConfig" form:"attachmentConfig" query:"attachmentConfig" db:"attachmentConfig"` // 附件配置，JSON格式，用于邮件附件等

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
