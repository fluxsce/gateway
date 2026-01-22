/**
 * 预警(告警)日志管理模块类型定义
 * 与后端 internal/alert/types/alert_log.go 保持一致
 */

// ============================================================
// 枚举定义
// ============================================================

/** 告警级别 */
export type AlertLevel = 'INFO' | 'WARN' | 'ERROR' | 'CRITICAL'

/** 发送状态 */
export type SendStatus = 'PENDING' | 'SENDING' | 'SUCCESS' | 'FAILED'

// ============================================================
// 预警日志类型定义
// ============================================================

/** 预警日志 - 对应 AlertLog */
export interface AlertLog {
  // 主键和租户
  tenantId: string                    // 租户ID，主键
  alertLogId: string                  // 告警日志ID，主键

  // 告警基本信息
  alertLevel: AlertLevel              // 告警级别：INFO/WARN/ERROR/CRITICAL
  alertType?: string | null           // 告警类型，业务自定义类型标识
  alertTitle: string                  // 告警标题
  alertContent?: string | null        // 告警内容
  alertTimestamp: string              // 告警时间戳

  // 关联信息
  channelName?: string | null         // 使用的渠道名称

  // 发送信息
  sendStatus?: SendStatus | null      // 发送状态：PENDING待发送/SENDING发送中/SUCCESS成功/FAILED失败
  sendTime?: string | null            // 发送时间
  sendResult?: string | null          // 发送结果详情，JSON格式
  sendErrorMessage?: string | null     // 发送错误信息

  // 标签和扩展信息
  alertTags?: string | null           // 告警标签，JSON格式
  alertExtra?: string | null          // 告警额外数据，JSON格式
  tableData?: string | null           // 表格数据，JSON格式

  // 通用字段
  addTime: string                     // 创建时间
  addWho: string                      // 创建人ID
  editTime: string                    // 最后修改时间
  editWho: string                     // 最后修改人ID
  oprSeqFlag: string                  // 操作序列标识
  currentVersion: number               // 当前版本号
  activeFlag: string                  // 活动状态标记
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

/** 预警日志查询请求参数 */
export interface AlertLogQueryParams {
  pageIndex: number
  pageSize: number
  alertLogId?: string                 // 告警日志ID（精确匹配）
  alertLevel?: AlertLevel              // 告警级别过滤
  alertType?: string                  // 告警类型过滤
  alertTitle?: string                 // 告警标题（模糊匹配）
  channelName?: string                // 渠道名称过滤
  sendStatus?: SendStatus             // 发送状态过滤
  alertTimestamp?: string             // 告警时间戳（精确匹配）
  startTime?: string                  // 开始时间（用于时间范围查询）
  endTime?: string                   // 结束时间（用于时间范围查询）
}

// ============= 常量定义 =============

/** 告警级别选项 */
export const ALERT_LEVEL_OPTIONS = [
  { label: '信息', value: 'INFO' as AlertLevel },
  { label: '警告', value: 'WARN' as AlertLevel },
  { label: '错误', value: 'ERROR' as AlertLevel },
  { label: '严重', value: 'CRITICAL' as AlertLevel },
]

/** 发送状态选项 */
export const SEND_STATUS_OPTIONS = [
  { label: '待发送', value: 'PENDING' as SendStatus },
  { label: '发送中', value: 'SENDING' as SendStatus },
  { label: '成功', value: 'SUCCESS' as SendStatus },
  { label: '失败', value: 'FAILED' as SendStatus },
]

