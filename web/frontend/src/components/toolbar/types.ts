import type { Component, VNode } from 'vue'

/**
 * 工具栏按钮类型
 */
export type ToolbarButtonType = 'default' | 'primary' | 'success' | 'warning' | 'error' | 'info'

/**
 * 工具栏按钮尺寸
 */
export type ToolbarButtonSize = 'tiny' | 'small' | 'medium' | 'large'

/**
 * 工具栏按钮配置
 */
export interface ToolbarButton {
  /**
   * 按钮唯一标识
   */
  key: string
  
  /**
   * 按钮文本
   */
  label?: string
  
  /**
   * 按钮图标
   * - Component: 预先获取的图标组件
   * - string: 图标名称（如 'AddOutline'），组件内部会自动获取
   * 
   * @example
   * ```typescript
   * // 方式1：传入组件（推荐，性能更好）
   * const icon = await getIcon('AddOutline')
   * 
   * // 方式2：直接传入图标名称（组件内部自动获取）
   * const icon = 'AddOutline'
   * ```
   */
  icon?: Component | string
  
  /**
   * 按钮类型
   */
  type?: ToolbarButtonType
  
  /**
   * 按钮尺寸
   */
  size?: ToolbarButtonSize
  
  /**
   * 是否禁用
   */
  disabled?: boolean
  
  /**
   * 是否显示
   */
  show?: boolean
  
  /**
   * 是否加载中
   */
  loading?: boolean
  
  /**
   * 是否显示为下拉菜单
   */
  dropdown?: boolean
  
  /**
   * 下拉菜单选项（当 dropdown 为 true 时有效）
   */
  dropdownOptions?: ToolbarDropdownOption[]
  
  /**
   * 工具提示
   */
  tooltip?: string
  
  /**
   * 点击事件处理
   */
  onClick?: (key: string) => void | Promise<void>
  
  /**
   * 自定义渲染函数（高级用法，用于完全自定义按钮内容）
   */
  render?: () => VNode
}

/**
 * 工具栏下拉菜单选项
 */
export interface ToolbarDropdownOption {
  /**
   * 选项唯一标识
   */
  key: string
  
  /**
   * 选项文本
   */
  label: string
  
  /**
   * 选项图标
   * - Component: 预先获取的图标组件
   * - string: 图标名称（如 'AddOutline'），组件内部会自动获取
   */
  icon?: Component | string
  
  /**
   * 是否禁用
   */
  disabled?: boolean
  
  /**
   * 是否显示分割线
   */
  divider?: boolean
  
  /**
   * 点击事件处理
   */
  onClick?: (key: string) => void | Promise<void>
}

/**
 * 工具栏分组配置
 */
export interface ToolbarGroup {
  /**
   * 分组唯一标识
   */
  key: string
  
  /**
   * 分组标题
   */
  title?: string
  
  /**
   * 分组按钮列表
   */
  buttons: ToolbarButton[]
  
  /**
   * 是否显示分割线
   */
  divider?: boolean
}

/**
 * 工具栏对齐方式
 */
export type ToolbarAlign = 'left' | 'center' | 'right' | 'space-between'

/**
 * 工具栏主题
 */
export type ToolbarTheme = 'light' | 'dark'

/**
 * 工具栏配置
 */
export interface ToolbarProps {
  /**
   * 模块ID（必填）
   * 用于标识工具栏所属的模块，便于权限控制和日志追踪
   * @example 'hub0001', 'hub0002'
   */
  moduleId: string
  
  /**
   * 工具栏标题
   */
  title?: string
  
  /**
   * 按钮列表（扁平结构）
   */
  buttons?: ToolbarButton[]
  
  /**
   * 按钮分组列表（分组结构）
   */
  groups?: ToolbarGroup[]
  
  /**
   * 对齐方式
   * @default 'left'
   */
  align?: ToolbarAlign
  
  /**
   * 是否显示边框
   * @default true
   */
  bordered?: boolean
  
  /**
   * 是否显示阴影
   * @default false
   */
  shadow?: boolean
  
  /**
   * 自定义高度（覆盖默认高度）
   */
  height?: string
}

/**
 * 工具栏事件
 */
export interface ToolbarEmits {
  /**
   * 按钮点击事件
   */
  (event: 'button-click', key: string): void
  
  /**
   * 下拉菜单选项点击事件
   */
  (event: 'dropdown-select', buttonKey: string, optionKey: string): void
}

