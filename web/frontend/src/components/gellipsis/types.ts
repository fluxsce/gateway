/**
 * GEllipsis 组件类型定义
 */

import type { EllipsisProps as NaiveEllipsisProps } from 'naive-ui'

/**
 * GEllipsis 组件属性
 * 基于 naive-ui 的 Ellipsis 组件封装
 */
export interface GEllipsisProps extends /* @vue-ignore */ Omit<NaiveEllipsisProps, 'default'> {
  /** 文本内容（可通过 slot 传入，优先级低于 slot） */
  text?: string
}

