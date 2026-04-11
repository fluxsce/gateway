import type { ButtonProps } from 'naive-ui'
import type { Component, StyleValue } from 'vue'

/**
 * Footer toolbar 按钮定义
 */
export interface GModalToolbarButton {
  /** 按钮唯一标识，用于 click 事件回调区分 */
  key: string
  /** 显示文本 */
  label: string
  /** Naive UI ButtonProps 透传（type / size / loading / disabled 等） */
  buttonProps?: Partial<ButtonProps>
}

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

  /**
   * 高度，支持数字（px）或任意 CSS 高度值。
   * 未设置时由 maxHeight（默认 80vh）限制外壳，内部用 flex 仅让内容区滚动。
   */
  height?: number | string

  /**
   * 是否允许通过边框（东、南、东南）拖拽改变宽高；全屏时关闭。
   * 使用像素宽高覆盖当前 width/height 展示值。
   */
  resizable?: boolean

  /** 可缩放时的最小宽度（px） */
  resizeMinWidth?: number

  /** 可缩放时的最小高度（px） */
  resizeMinHeight?: number

  /** 可缩放时的最大宽度，数字为 px，字符串可为任意 CSS（默认贴近视口） */
  resizeMaxWidth?: number | string

  /** 可缩放时的最大高度 */
  resizeMaxHeight?: number | string

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

  /**
   * Footer 自定义 toolbar 按钮列表。
   * 设置后将在内置取消/确认按钮左侧渲染自定义按钮组，
   * 点击时触发 toolbar-click 事件并携带对应 key。
   */
  footerToolbar?: GModalToolbarButton[]

  /** 是否自动聚焦到弹窗内部 */
  autoFocus?: boolean

  /**
   * 是否将焦点限制在弹窗内（透传 NModal `trap-focus`，使用 vueuc VFocusTrap）。
   * 为 true 时两侧占位 div 带 aria-hidden 且可获焦，部分 Chrome 在 Tab 时会报控制台警告。
   * 需要严格键盘陷阱时设为 true；一般业务可设为 false 消除警告。
   */
  trapFocus?: boolean

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

  /** 点击 footerToolbar 中的自定义按钮时触发，携带该按钮的 key */
  (event: 'toolbar-click', key: string): void

  /** 弹窗被关闭时触发（包括遮罩关闭、关闭按钮等） */
  (event: 'close'): void

  /** 动画进入完成 */
  (event: 'after-enter'): void

  /** 动画离开完成 */
  (event: 'after-leave'): void

  /** 边框缩放结束（mouseup），像素宽高 */
  (event: 'resize', payload: { width: number; height: number }): void
}


