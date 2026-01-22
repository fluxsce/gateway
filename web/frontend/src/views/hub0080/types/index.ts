/**
 * 预警(告警)配置管理模块类型定义
 * 与后端 internal/alert/types/alert_config.go 保持一致
 */

// ============================================================
// 枚举定义
// ============================================================

/** 渠道类型 */
export type ChannelType = 'email' | 'qq' | 'wechat_work' | 'dingtalk' | 'webhook' | 'sms'

/** 活动标记 */
export type ActiveFlag = 'Y' | 'N'

/** 默认标记 */
export type DefaultFlag = 'Y' | 'N'

/** 异步发送标记 */
export type AsyncSendFlag = 'Y' | 'N'

/** 消息内容格式 */
export type MessageContentFormat = 'text' | 'html' | 'markdown'

// ============================================================
// 告警渠道配置类型定义
// ============================================================

/** 告警渠道配置 - 对应 AlertConfig */
export interface AlertConfig {
  // 主键和租户
  tenantId: string                    // 租户ID，主键
  channelName: string                 // 渠道名称，主键

  // 渠道基本信息
  channelType: ChannelType            // 渠道类型：email/qq/wechat_work/dingtalk/webhook/sms
  channelDesc?: string | null         // 渠道描述
  activeFlag: ActiveFlag              // 启用状态：Y-启用，N-禁用
  defaultFlag: DefaultFlag            // 是否默认渠道：Y-是，N-否
  priorityLevel: number               // 优先级：1-10，数字越小优先级越高
  defaultTemplateName?: string | null  // 默认关联的模板名称

  // 服务器配置（JSON格式）
  serverConfig?: string | null        // 服务器配置，JSON格式，如SMTP配置、Webhook URL等
  sendConfig?: string | null          // 发送配置，JSON格式，如默认收件人、超时设置等

  // 消息格式配置
  messageContentFormat?: MessageContentFormat | null  // 消息内容格式：text/html/markdown

  // 重试和超时配置
  timeoutSeconds: number              // 超时时间（秒）
  retryCount: number                  // 重试次数
  retryIntervalSecs: number           // 重试间隔（秒）
  asyncSendFlag: AsyncSendFlag        // 是否异步发送：Y-是，N-否

  // 统计信息
  totalSentCount: number              // 总发送次数
  successCount: number                // 成功次数
  failureCount: number                // 失败次数
  lastSendTime?: string | null        // 最后发送时间
  lastSuccessTime?: string | null      // 最后成功时间
  lastFailureTime?: string | null      // 最后失败时间
  lastErrorMessage?: string | null    // 最后错误信息

  // 通用字段
  addTime: string                     // 创建时间
  addWho: string                      // 创建人ID
  editTime: string                    // 最后修改时间
  editWho: string                     // 最后修改人ID
  oprSeqFlag: string                  // 操作序列标识
  currentVersion: number               // 当前版本号
  noteText?: string | null            // 备注信息
  extProperty?: string | null          // 扩展属性，JSON格式

  // 预留字段
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

/** 告警渠道配置查询请求参数 */
export interface AlertConfigQueryParams {
  pageIndex: number
  pageSize: number
  channelName?: string                 // 渠道名称（精确匹配）
  channelType?: ChannelType           // 渠道类型过滤
  activeFlag?: ActiveFlag             // 活动标记过滤
  defaultFlag?: DefaultFlag           // 默认标记过滤
  priorityLevel?: number               // 优先级过滤
}

// ============= 常量定义 =============

/** 渠道类型选项 */
export const CHANNEL_TYPE_OPTIONS = [
  { label: '邮件', value: 'email' as ChannelType },
  { label: 'QQ', value: 'qq' as ChannelType },
  { label: '企业微信', value: 'wechat_work' as ChannelType },
  { label: '钉钉', value: 'dingtalk' as ChannelType },
  { label: 'Webhook', value: 'webhook' as ChannelType },
  { label: '短信', value: 'sms' as ChannelType },
]

/** 活动标记选项 */
export const ACTIVE_FLAG_OPTIONS = [
  { label: '启用', value: 'Y' as ActiveFlag },
  { label: '禁用', value: 'N' as ActiveFlag },
]

/** 默认标记选项 */
export const DEFAULT_FLAG_OPTIONS = [
  { label: '是', value: 'Y' as DefaultFlag },
  { label: '否', value: 'N' as DefaultFlag },
]

/** 异步发送标记选项 */
export const ASYNC_SEND_FLAG_OPTIONS = [
  { label: '是', value: 'Y' as AsyncSendFlag },
  { label: '否', value: 'N' as AsyncSendFlag },
]

/** 消息内容格式选项 */
export const MESSAGE_CONTENT_FORMAT_OPTIONS = [
  { label: '文本', value: 'text' as MessageContentFormat },
  { label: 'HTML', value: 'html' as MessageContentFormat },
  { label: 'Markdown', value: 'markdown' as MessageContentFormat },
]

