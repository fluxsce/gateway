/**
 * GTips 组件 Props
 */
export interface GTipsProps {
  /**
   * 提示内容
   * 如果提供了 slot，则优先使用 slot 内容
   */
  content?: string

  /**
   * 触发方式
   * @default 'hover'
   */
  trigger?: 'hover' | 'click' | 'focus' | 'manual'

  /**
   * 提示位置
   * @default 'top'
   */
  placement?:
    | 'top'
    | 'top-start'
    | 'top-end'
    | 'right'
    | 'right-start'
    | 'right-end'
    | 'bottom'
    | 'bottom-start'
    | 'bottom-end'
    | 'left'
    | 'left-start'
    | 'left-end'

  /**
   * 是否显示箭头
   * @default true
   */
  showArrow?: boolean

  /**
   * 最大宽度（像素）
   * @default 320
   */
  maxWidth?: number

  /**
   * 延迟显示时间（毫秒）
   * @default 0
   */
  delay?: number

  /**
   * 显示持续时间（毫秒）
   * @default 100
   */
  duration?: number

  /**
   * 图标大小
   * @default 16
   */
  iconSize?: 12 | 14 | 16 | 18 | 20

  /**
   * 图标样式
   */
  iconStyle?: Record<string, any>
}

