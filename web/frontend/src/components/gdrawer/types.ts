import type { StyleValue } from 'vue'

export type GDrawerPlacement = 'top' | 'right' | 'bottom' | 'left'

/**
 * GDrawer 组件 Props
 * 封装 Naive UI NDrawer，统一项目内抽屉用法
 */
export interface GDrawerProps {
  /** 是否显示抽屉（v-model:show） */
  show: boolean

  /** 标题文本（也可以通过 header 插槽自定义） */
  title?: string

  /** 抽屉宽度（placement 为 left/right 时）或高度（placement 为 top/bottom 时） */
  width?: number | string

  /** 抽屉位置 */
  placement?: GDrawerPlacement

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

  /** 是否自动聚焦到抽屉内部 */
  autoFocus?: boolean

  /**
   * 是否显示遮罩层
   * 透传给 NDrawer 的 show-mask 属性
   * @default true
   */
  mask?: boolean

  /**
   * 是否阻止背景滚动
   * 透传给 NDrawer 的 block-scroll 属性
   * @default true
   */
  blockScroll?: boolean

  /** 抽屉挂载容器（透传给 NDrawer 的 to 属性） */
  to?: string | HTMLElement | false

  /** 是否可调整大小（透传给 NDrawer 的 resizable） */
  resizable?: boolean

  /** 额外样式（可选，直接作用在 NDrawer 上） */
  style?: StyleValue

  /** 主体 header 的类名（透传给 NDrawerContent 的 header-class） */
  headerClass?: string

  /** 主体 header 的样式（透传给 NDrawerContent 的 header-style） */
  headerStyle?: string | Record<string, any>

  /** 主体 footer 的类名（透传给 NDrawerContent 的 footer-class） */
  footerClass?: string

  /** 主体 footer 的样式（透传给 NDrawerContent 的 footer-style） */
  footerStyle?: string | Record<string, any>

  /** 主体 body 的类名（透传给 NDrawerContent 的 body-class） */
  bodyClass?: string

  /** 主体 body 的样式（透传给 NDrawerContent 的 body-style） */
  bodyStyle?: string | Record<string, any>

  /** 主体可滚动内容节点的类名（透传给 NDrawerContent 的 body-content-class） */
  bodyContentClass?: string

  /** 主体可滚动内容节点的样式（透传给 NDrawerContent 的 body-content-style） */
  bodyContentStyle?: string | Record<string, any>
}

/**
 * GDrawer 组件事件
 */
export interface GDrawerEmits {
  /** v-model:show 更新 */
  (event: 'update:show', value: boolean): void

  /** 点击取消按钮或关闭时触发 */
  (event: 'cancel'): void

  /** 点击确认按钮时触发 */
  (event: 'confirm'): void

  /** 抽屉被关闭时触发（包括遮罩关闭、关闭按钮等） */
  (event: 'close'): void

  /** 动画进入完成 */
  (event: 'after-enter'): void

  /** 动画离开完成 */
  (event: 'after-leave'): void
}

