/**
 * GMessage 消息与对话框类型定义
 *
 * 参考 xirang-message，提供消息条与确认/警告/提示对话框的选项与 API 类型。
 */

import type { Component, VNode } from 'vue'

/** 消息类型：info | success | warning | error | loading */
export type GMessageType = 'info' | 'success' | 'warning' | 'error' | 'loading'

/** 消息显示位置（对应各位置 Provider） */
export type GMessagePosition =
  | 'top'
  | 'top-left'
  | 'top-right'
  | 'bottom'
  | 'bottom-left'
  | 'bottom-right'

/** 单条消息的选项（用于 info/success/error/warning/loading） */
export interface GMessageOptions {
  /** 消息内容 */
  content?: string | VNode | Component
  /** 类型 */
  type?: GMessageType
  /** 持续时间（ms），0 表示不自动关闭 */
  duration?: number
  /** 是否显示关闭按钮 */
  closable?: boolean
  /** 自定义图标 */
  icon?: string | Component
  /** 自定义类名 */
  className?: string
  /** 自定义样式 */
  style?: string | Record<string, string | number>
  /** 位置 */
  position?: GMessagePosition
  /** 关闭回调 */
  onClose?: () => void
  /** 关闭后回调 */
  onAfterClose?: () => void
}

/**
 * 对话框选项（用于 confirm / alert / prompt）
 *
 * 与 GDialog 选项对齐，通过 $gMessage.confirm / alert / prompt 调用时使用。
 */
export interface GMessageDialogOptions {
  /** 标题 */
  title?: string
  /** 内容（字符串或 VNode） */
  content?: string | VNode | Component
  /** 是否显示确认按钮，默认 true */
  showConfirm?: boolean
  /** 是否显示取消按钮，默认 true（confirm 为 true，alert 为 false） */
  showCancel?: boolean
  /** 确认按钮文案 */
  confirmText?: string
  /** 取消按钮文案 */
  cancelText?: string
  /** 确认回调（在用户点击确认后执行，可返回 Promise） */
  onConfirm?: () => void | Promise<void>
  /** 取消回调 */
  onCancel?: () => void
  /** 对话框宽度 */
  width?: number | string
  /** 头部样式，用于区分 info/warning/error */
  headerStyle?: 'default' | 'gradient'
}

/**
 * 仅消息条的 API（由各位置 GMessageProvider 实现并注册到 utils）
 *
 * 供 getMessageApi() 返回；$gMessage 的 toask 方法内部会调用 getMessageApi()?.xxx(opts)。
 */
export interface GMessageProviderApi {
  (content: string | GMessageOptions, options?: GMessageOptions): void
  success: (content: string | GMessageOptions, options?: GMessageOptions) => void
  error: (content: string | GMessageOptions, options?: GMessageOptions) => void
  warning: (content: string | GMessageOptions, options?: GMessageOptions) => void
  info: (content: string | GMessageOptions, options?: GMessageOptions) => void
  loading: (content: string | GMessageOptions, options?: GMessageOptions) => void
  destroyAll: () => void
}

/**
 * 函数式消息 API（挂到 $gMessage）
 *
 * 包含消息条：info / success / error / warning / loading / destroyAll；
 * 以及对话框：confirm / alert / prompt（基于 GDialog，在 utils 中实现）。
 */
export interface GMessageApi extends GMessageProviderApi {
  /** 确认对话框（确定+取消），返回 Promise<boolean> */
  confirm: (options: GMessageDialogOptions) => Promise<boolean>
  /** 警告/提示对话框（仅确定），返回 Promise<void> */
  alert: (options: GMessageDialogOptions) => Promise<void>
  /** 输入提示对话框（待实现），返回 Promise<string | null> */
  prompt: (options: GMessageDialogOptions & { defaultValue?: string }) => Promise<string | null>
}

/** Provider 全局默认配置（duration、closable、max 等） */
export interface GMessageProviderProps {
  position?: GMessagePosition
  duration?: number
  closable?: boolean
  max?: number
  containerClass?: string
  containerStyle?: string | Record<string, string | number>
}
