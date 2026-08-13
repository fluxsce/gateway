/**
 * GEllipsis 组件类型定义
 */
export interface GEllipsisProps {
  /** 文本内容（可通过 slot 传入，优先级低于 slot） */
  text?: string
  lineClamp?: number
  tooltip?: boolean
  [key: string]: unknown
}
