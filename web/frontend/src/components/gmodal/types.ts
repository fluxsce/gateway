import type { Component, StyleValue } from 'vue'

export type GModalPreset = 'dialog' | 'card'

/**
 * GModal 组件 Props
 * 封装 Naive UI NModal，统一项目内弹窗用法
 */
export interface GModalProps {
  /** 是否显示弹窗（v-model:visible） */
  visible: boolean

  /** 标题文本（也可以通过 header 插槽自定义） */
  title?: string

  /** 头部图标组件（用于替换 dialog 的默认图标） */
  headerIcon?: Component

  /** 宽度，支持数字或任意 CSS 宽度值 */
  width?: number | string

  /** 使用的 NModal 预设类型，默认为 dialog */
  preset?: GModalPreset

  /** 是否点击遮罩关闭 */
  maskClosable?: boolean

  /** 是否显示右上角关闭按钮 */
  closable?: boolean

  /** 是否显示底部操作栏 */
  showFooter?: boolean

  /** 是否显示取消按钮 */
  showCancel?: boolean

  /** 是否显示确认按钮 */
  showConfirm?: boolean

  /** 取消按钮文案 */
  cancelText?: string

  /** 确认按钮文案 */
  confirmText?: string

  /** 确认按钮加载状态 */
  confirmLoading?: boolean

  /** 是否自动聚焦到弹窗内部 */
  autoFocus?: boolean

  /**
   * 是否可拖拽
   * 透传给 NModal 的 draggable 属性
   * @default true
   */
  draggable?: boolean

  /**
   * 是否显示遮罩层
   * 透传给 NModal 的 mask 属性
   * @default false
   */
  mask?: boolean

  /**
   * 是否阻止背景滚动
   * 透传给 NModal 的 block-scroll 属性
   * @default false
   */
  blockScroll?: boolean

  /** 弹窗挂载容器（透传给 NModal 的 to 属性） */
  to?: string | HTMLElement | false

  /** 是否使用分割线风格（透传给 NModal 的 segmented） */
  segmented?: boolean | { content?: boolean; footer?: boolean }

  /** 是否显示边框（透传给 NModal 的 bordered） */
  bordered?: boolean

  /** 是否显示全屏切换按钮（位于右上角关闭按钮旁边） */
  showFullscreenToggle?: boolean

  /** 额外样式（可选，直接作用在 NModal 上） */
  style?: StyleValue
}

/**
 * GModal 组件事件
 */
export interface GModalEmits {
  /** v-model:visible 更新 */
  (event: 'update:visible', value: boolean): void

  /** 点击取消按钮或关闭时触发 */
  (event: 'cancel'): void

  /** 点击确认按钮时触发 */
  (event: 'confirm'): void

  /** 弹窗被关闭时触发（包括遮罩关闭、关闭按钮等） */
  (event: 'close'): void

  /** 动画进入完成 */
  (event: 'after-enter'): void

  /** 动画离开完成 */
  (event: 'after-leave'): void
}


