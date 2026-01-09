import type { ToolbarProps } from '@/components/toolbar'
import type { FormInst, FormItemRule } from 'naive-ui'
import type { Component, VNode } from 'vue'

/**
 * 搜索表单字段类型
 */
export type SearchFieldType =
  | 'input'
  | 'select'
  | 'date'
  | 'daterange'
  | 'datetime'
  | 'datetimerange'
  | 'number'
  | 'switch'
  | 'custom'

/**
 * 搜索表单字段配置
 */
export interface SearchField {
  /**
   * 字段名称（对应表单数据的 key）
   */
  field: string

  /**
   * 字段标签
   */
  label: string

  /**
   * 字段类型
   * @default 'input'
   */
  type?: SearchFieldType

  /**
   * 占位符文本
   */
  placeholder?: string

  /**
   * 默认值
   */
  defaultValue?: any

  /**
   * 是否显示该字段
   * @default true
   */
  show?: boolean

  /**
   * 是否必填
   * 仅用于配置层标记，具体校验请结合 rules 一起使用
   * @default false
   */
  required?: boolean

  /**
   * 是否禁用
   * @default false
   */
  disabled?: boolean

  /**
   * 是否可清空
   * @default true
   */
  clearable?: boolean

  /**
   * 选项列表（用于 select 类型）
   */
  options?: Array<{
    label: string
    value: string | number
    disabled?: boolean
  }>

  /**
   * 字段栅格占位（1-24）
   * @default 6
   */
  span?: number

  /**
   * 自定义渲染函数（用于 custom 类型）
   */
  render?: (formData: Record<string, any>) => Component | VNode

  /**
   * 字段验证规则
   */
  rules?: FormItemRule | FormItemRule[]

  /**
   * 传递给表单项组件的额外属性
   */
  props?: Record<string, any>
}

/**
 * 搜索表单配置
 */
export interface SearchFormProps extends Pick<ToolbarProps, 'moduleId'> {
  /**
   * 表单字段配置列表
   */
  fields: SearchField[]

  /**
   * 更多查询条件字段（可选）
   * 如果存在，工具栏会自动添加"更多条件"按钮
   */
  moreFields?: SearchField[]

  /**
   * 标签宽度
   * @default 80
   */
  labelWidth?: number | string

  /**
   * 标签位置
   * @default 'left'
   */
  labelPlacement?: 'left' | 'top'

  /**
   * 表单尺寸
   * @default 'medium'
   */
  size?: 'small' | 'medium' | 'large'

  /**
   * 是否显示为行内表单
   * @default false
   */
  inline?: boolean

  /**
   * 栅格列数
   * @default 24
   */
  cols?: number

  /**
   * 栅格间距
   * @default 16
   */
  xGap?: number

  /**
   * 栅格行间距
   * @default 16
   */
  yGap?: number

  /**
   * 更多条件按钮文本
   * @default '更多条件'
   */
  moreButtonText?: string

  /**
   * 工具栏按钮配置（继承自 ToolbarProps.buttons）
   * 如果不提供，将使用默认的查询和重置按钮
   */
  toolbarButtons?: ToolbarProps['buttons']

  /**
   * 是否显示默认的查询按钮
   * @default true
   */
  showSearchButton?: boolean

  /**
   * 是否显示默认的重置按钮
   * @default true
   */
  showResetButton?: boolean

  /**
   * 查询按钮文本
   * @default '查询'
   */
  searchButtonText?: string

  /**
   * 重置按钮文本
   * @default '重置'
   */
  resetButtonText?: string

  /**
   * 工具栏对齐方式（继承自 ToolbarProps.align）
   * @default 'right'
   */
  toolbarAlign?: ToolbarProps['align']

  /**
   * 是否显示工具栏
   * @default true
   */
  showToolbar?: boolean
}

/**
 * 搜索表单事件
 */
export interface SearchFormEmits {
  /**
   * 查询事件（传递表单数据）
   */
  (event: 'search', formData: Record<string, any>): void

  /**
   * 字段值变化事件
   */
  (event: 'field-change', field: string, value: any): void

  /**
   * 工具栏按钮点击事件
   * @param key 按钮 key
   * @param formData 表单数据（可选，search 操作时会传递）
   */
  (event: 'toolbar-click', key: string, formData?: Record<string, any>): void
}

/**
 * 搜索表单暴露的方法
 */
export interface SearchFormExpose {
  /**
   * 获取表单实例
   */
  getFormRef: () => FormInst | undefined

  /**
   * 获取表单数据
   */
  getFormData: () => Record<string, any>

  /**
   * 设置表单数据
   */
  setFormData: (data: Record<string, any>) => void

  /**
   * 重置表单
   */
  resetForm: () => void

  /**
   * 验证表单
   */
  validate: () => Promise<void>

  /**
   * 提交查询
   */
  submit: () => void

  /**
   * 切换更多条件显示/隐藏
   */
  toggleMoreFields: () => void
}

