import type { ToolbarProps } from '@/components/toolbar'
import type { RsFormNamePath, RsFormRuleItem } from '@/ui'
import type { Component, VNode } from 'vue'

/**
 * RsSearchForm 字段类型
 */
export type RsSearchFieldType =
  | 'input'
  | 'select'
  | 'date'
  | 'daterange'
  | 'datetime'
  | 'datetimerange'
  | 'number'
  | 'switch'
  | 'custom'

/** DatePicker range（valueFormat=string）的标准形态 */
export interface RsSearchDateRangeValue {
  start: string
  end: string
}

/**
 * 自定义字段渲染上下文（对齐 RsDataForm / Ant Form.Item）。
 * value / onUpdate 绑定当前 field；改其它路径用 setFieldValue。
 */
export interface RsSearchFormRenderContext {
  value: unknown
  onUpdate: (value: unknown) => void
  setFieldValue: (name: RsFormNamePath, value: unknown) => void
}

/**
 * RsSearchForm 字段配置（基于 niuma-ui 表单控件）
 */
export interface RsSearchField {
  /** 字段名称（对应表单数据的 key） */
  field: string
  /** 字段标签（支持函数动态生成） */
  label: string | ((formData: Record<string, any>) => string)
  /** 字段提示（字符串 tips 走 RsTooltip icon） */
  tips?: string | Component | VNode | ((formData: Record<string, any>) => string | Component | VNode)
  /** 字段类型，默认 input */
  type?: RsSearchFieldType
  placeholder?: string
  defaultValue?: unknown
  /** 是否显示，默认 true */
  show?: boolean
  required?: boolean
  disabled?: boolean
  /** 是否可清空，默认 true */
  clearable?: boolean
  options?: Array<{
    label: string
    value: string | number
    disabled?: boolean
  }>
  /** 栅格占位 1-24，默认 6 */
  span?: number
  /**
   * 自定义渲染。优先用 ctx.onUpdate / setFieldValue 回写，
   * formData 仅作只读快照（兼容旧 render 直接改 formData）。
   */
  render?: (
    formData: Record<string, any>,
    ctx: RsSearchFormRenderContext,
  ) => Component | VNode
  rules?: RsFormRuleItem | RsFormRuleItem[]
  props?: Record<string, unknown>
}

/**
 * RsSearchForm Props
 */
export interface RsSearchFormProps extends Pick<ToolbarProps, 'moduleId'> {
  fields: RsSearchField[]
  moreFields?: RsSearchField[]
  labelWidth?: number | string
  labelPlacement?: 'left' | 'top'
  labelAlign?: 'left' | 'right'
  size?: 'small' | 'medium' | 'large'
  inline?: boolean
  cols?: number
  xGap?: number
  yGap?: number
  moreButtonText?: string
  toolbarButtons?: ToolbarProps['buttons']
  showSearchButton?: boolean
  showResetButton?: boolean
  searchButtonText?: string
  resetButtonText?: string
  toolbarAlign?: ToolbarProps['align']
  showToolbar?: boolean
}

/**
 * RsSearchForm 事件
 */
export interface RsSearchFormEmits {
  (event: 'search', formData: Record<string, any>): void
  (event: 'reset'): void
  (event: 'field-change', field: string, value: unknown): void
  (event: 'toolbar-click', key: string, formData?: Record<string, any>): void
}

/**
 * RsSearchForm 暴露方法
 */
export interface RsSearchFormExpose {
  getFormRef: () => unknown
  getFormData: () => Record<string, any>
  setFormData: (data: Record<string, any>) => void
  resetForm: () => void
  validate: () => Promise<void>
  submit: () => void
  toggleMoreFields: () => void
}

/** @deprecated 兼容旧命名，渐进迁移时可用 */
export type SearchField = RsSearchField
/** @deprecated 兼容旧命名 */
export type SearchFieldType = RsSearchFieldType
/** @deprecated 兼容旧命名 */
export type SearchFormProps = RsSearchFormProps
/** @deprecated 兼容旧命名 */
export type SearchFormEmits = RsSearchFormEmits
/** @deprecated 兼容旧命名 */
export type SearchFormExpose = RsSearchFormExpose
