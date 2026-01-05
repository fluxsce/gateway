/**
 * GFieldset 组件 Props
 * 用于将相关的表单字段分组显示，类似于 HTML 的 <fieldset> 元素
 * 支持语义化分组和批量禁用表单控件
 */
export interface GFieldsetProps {
  /** 分组标题（使用 legend 元素显示） */
  title?: string

  /** 标题是否加粗 */
  titleStrong?: boolean

  /**
   * 标题清晰度/大小
   * - 数字：直接设置字体粗细（font-weight），如 100, 200, 300, 400, 500, 600, 700, 800, 900
   * - 'small': 小号（12px，normal）
   * - 'normal': 正常（14px，normal）
   * - 'large': 大号（16px，600）
   * @default 'normal'
   */
  titleSize?: number | 'small' | 'normal' | 'large'

  /**
   * 边框样式
   * - 'solid': 实线边框
   * - 'dashed': 虚线边框
   * - 'none': 无边框
   * @default 'solid'
   */
  borderStyle?: 'solid' | 'dashed' | 'none'

  /**
   * 是否选中（高亮显示）
   * @default false
   */
  selected?: boolean

  /**
   * 是否禁用整个分组
   * 当设置为 true 时，会禁用分组内的所有表单控件（input、select、textarea 等）
   * @default false
   */
  disabled?: boolean
}

/**
 * GFieldset 组件事件
 */
export interface GFieldsetEmits {
  // 目前不需要事件，后续可根据需要添加
}

