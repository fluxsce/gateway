import type { Component, StyleValue } from 'vue'

/**
 * GDialog 组件 Props
 * 封装 Naive UI NModal + NCard，提供统一的对话框样式和布局
 */
export interface GDialogProps {
  /** 是否显示对话框（v-model:show） */
  show: boolean

  /** 对话框宽度 */
  width?: number | string

  /** 对话框标题 */
  title?: string

  /** 对话框副标题 */
  subtitle?: string

  /** 副标题显示位置：'header' 显示在头部 | 'footer' 显示在底部 */
  subtitlePosition?: 'header' | 'footer'

  /** 头部图标组件 */
  icon?: Component

  /** 图标大小 */
  iconSize?: number

  /** 头部样式类型：'default' 默认样式 | 'gradient' 渐变背景 */
  headerStyle?: 'default' | 'gradient'

  /** 自定义头部样式（覆盖默认样式） */
  headerCustomStyle?: StyleValue

  /** 是否可点击遮罩关闭 */
  maskClosable?: boolean

  /** 是否可按下 ESC 关闭 */
  closeOnEsc?: boolean

  /** 是否显示关闭按钮 */
  closable?: boolean

  /** 内容区域最大高度 */
  contentMaxHeight?: string

  /** 是否显示滚动条 */
  showScrollbar?: boolean

  /** 是否显示底部操作区 */
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

  /** 确认后是否自动关闭对话框 */
  autoCloseOnConfirm?: boolean

  /** 底部操作按钮对齐方式：'start' 左对齐 | 'end' 右对齐 | 'center' 居中 | 'space-between' 两端对齐 | 'space-around' 环绕分布 */
  footerButtonAlign?: 'start' | 'end' | 'center' | 'space-between' | 'space-around'

  /** 是否显示全屏切换按钮 */
  showFullscreenToggle?: boolean

  /** 是否可拖拽 */
  draggable?: boolean

  /** 额外的卡片样式类名 */
  cardClass?: string

  /** 额外的样式 */
  style?: StyleValue
}

/**
 * GDialog 组件事件
 */
export interface GDialogEmits {
  /** v-model:show 更新 */
  (event: 'update:show', value: boolean): void

  /** 点击取消按钮时触发 */
  (event: 'cancel'): void

  /** 点击确认按钮时触发 */
  (event: 'confirm'): void

  /** 对话框被关闭时触发 */
  (event: 'close'): void

  /** 动画进入完成 */
  (event: 'after-enter'): void

  /** 动画离开完成 */
  (event: 'after-leave'): void
}

