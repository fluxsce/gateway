/**
 * 预警(告警)模板管理模块类型定义
 * 与后端 internal/alert/types/alert_template.go 保持一致（字段名/含义）
 */

// ============================================================
// 枚举定义
// ============================================================

/** 渠道类型 */
export type ChannelType = 'email' | 'qq' | 'wechat_work' | 'dingtalk' | 'webhook' | 'sms'

/** 显示格式 */
export type DisplayFormat = 'table' | 'text'

/** 活动标记 */
export type ActiveFlag = 'Y' | 'N'

// ============================================================
// 预警模板类型定义
// ============================================================

export interface AlertTemplate {
  tenantId: string
  templateName: string

  templateDesc?: string | null
  channelType?: ChannelType | null

  titleTemplate?: string | null
  contentTemplate?: string | null
  displayFormat: DisplayFormat
  // 后端历史字段：前端已移除“模板变量”配置入口，保留类型以兼容详情返回
  templateVariables?: string | null

  attachmentConfig?: string | null

  addTime: string
  addWho: string
  editTime: string
  editWho: string
  oprSeqFlag: string
  currentVersion: number
  activeFlag: ActiveFlag
  noteText?: string | null
  extProperty?: string | null

  reserved1?: string | null
  reserved2?: string | null
  reserved3?: string | null
  reserved4?: string | null
  reserved5?: string | null
  reserved6?: string | null
  reserved7?: string | null
  reserved8?: string | null
  reserved9?: string | null
  reserved10?: string | null
}

/** 预警模板查询请求参数 */
export interface AlertTemplateQueryParams {
  pageIndex: number
  pageSize: number
  templateName?: string
  channelType?: ChannelType | ''
  displayFormat?: DisplayFormat | ''
  activeFlag?: ActiveFlag | ''
}

// ============= 常量定义 =============

export const CHANNEL_TYPE_OPTIONS = [
  { label: '邮件', value: 'email' as ChannelType },
  { label: 'QQ', value: 'qq' as ChannelType },
  { label: '企业微信', value: 'wechat_work' as ChannelType },
  { label: '钉钉', value: 'dingtalk' as ChannelType },
  { label: 'Webhook', value: 'webhook' as ChannelType },
  { label: '短信', value: 'sms' as ChannelType },
]

export const DISPLAY_FORMAT_OPTIONS = [
  { label: '文本', value: 'text' as DisplayFormat },
  { label: '表格', value: 'table' as DisplayFormat },
]

export const ACTIVE_FLAG_OPTIONS = [
  { label: '启用', value: 'Y' as ActiveFlag },
  { label: '禁用', value: 'N' as ActiveFlag },
]


