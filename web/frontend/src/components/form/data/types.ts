import type { GModalEmits, GModalProps } from '@/components/gmodal/types'
import type { FormItemRule } from 'naive-ui'
import type { Component, VNode } from 'vue'

/**
 * 数据表单字段类型
 * 参考 SearchFieldType，支持常见输入控件
 */
export type DataFormFieldType =
  | 'input'
  | 'textarea'
  | 'select'
  | 'date'
  | 'daterange'
  | 'datetime'
  | 'datetimerange'
  | 'number'
  | 'switch'
  | 'fieldset'
  | 'file'
  | 'custom'

/**
 * 数据表单字段配置
 * 用于在 GDataFormModal 中描述每一列表单项
 */
export interface DataFormField {
  /** 字段名称（对应表单数据的 key） */
  field: string

  /**
   * 字段标签
   * 支持字符串或函数（根据表单数据动态生成）
   */
  label: string | ((formData: Record<string, any>) => string)

  /**
   * 字段提示信息（在标签旁显示问号图标，悬浮时显示提示）
   * 支持字符串、VNode/Component（用于复杂内容）或函数（根据表单数据动态生成）
   */
  tips?: string | Component | VNode | ((formData: Record<string, any>) => string | Component | VNode)

  /**
   * 字段类型
   * @default 'input'
   */
  type?: DataFormFieldType

  /** 占位符文本 */
  placeholder?: string

  /** 默认值 */
  defaultValue?: any

  /**
   * 是否显示该字段
   * 可以是布尔值，也可以是函数（基于表单数据动态计算）
   * @default true
   */
  show?: boolean | ((formData: Record<string, any>) => boolean)

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
   * 是否为主键字段
   * 在编辑模式下，主键字段会自动禁用，防止修改
   * @default false
   */
  primary?: boolean

  /**
   * 是否可清空（针对 input / select 等控件）
   * @default true
   */
  clearable?: boolean

  /**
   * 选项列表（用于 select 等类型）
   */
  options?: Array<{
    label: string
    value: string | number
    disabled?: boolean
  }>

  /**
   * 字段栅格占位（1-24）
   * @default 12
   */
  span?: number

  /**
   * 所属页签 key（用于多页签表单，将字段分组到不同 Tab 中）
   * - 未设置时会归入第一个页签
   */
  tabKey?: string

  /**
   * 子字段列表（用于 fieldset 类型，支持嵌套分组）
   * - 当 type 为 'fieldset' 时，children 中的字段会被包裹在 GFieldset 中
   * - 支持多级嵌套
   */
  children?: DataFormField[]

  /**
   * 自定义渲染函数（用于 custom 类型）
   */
  render?: (formData: Record<string, any>) => Component | VNode

  /**
   * 字段验证规则（Naive UI FormItemRule）
   */
  rules?: FormItemRule | FormItemRule[]

  /**
   * 传递给表单项组件的额外属性
   */
  props?: Record<string, any>
}

/**
 * 数据编辑模态框模式
 */
export type DataModalMode = 'create' | 'edit' | 'view'

/**
 * 数据编辑模态框页签配置
 */
export interface DataFormTab {
  /** 页签唯一标识（与 DataFormField.tabKey 对应） */
  key: string
  /** 页签显示名称 */
  label: string
  /**
   * 是否显示该页签
   * 可以是布尔值，也可以是函数（基于表单数据动态计算）
   * @default true
   */
  show?: boolean | ((formData: Record<string, any>) => boolean)
}

/**
 * 数据编辑模态框 Props（GDataFormModal）
 * 在 GModalProps 的基础上增加常用编辑场景配置
 */
export interface DataModalProps extends GModalProps {
  /**
   * 当前业务模式：
   * - 'create'：新增
   * - 'edit'：编辑
   * - 'view'：查看详情（通常会禁用表单，仅展示）
   * @default 'create'
   */
  mode?: DataModalMode

  /**
   * 表单字段配置列表
   * 参考 SearchFormProps.fields
   * 支持嵌套结构，type 为 'fieldset' 的字段可以包含 children
   */
  formFields?: DataFormField[]

  /**
   * 表单页签配置
   * - 如果不配置，则不显示页签，所有字段在同一页
   * - 如果配置，则会根据 DataFormField.tabKey 将字段分配到对应页签
   */
  formTabs?: DataFormTab[]

  /**
   * 点击确认后是否自动关闭弹窗
   * @default true
   */
  autoCloseOnConfirm?: boolean

  /**
   * 弹窗关闭时是否自动触发 reset 事件（通常用于重置内部表单）
   * @default false
   */
  autoResetOnClose?: boolean

  /**
   * 初始表单数据（用于编辑模式，传入要编辑的数据）
   * 当 visible 变为 true 时，会将此数据填充到表单中
   */
  initialData?: Record<string, any>
}

/**
 * 数据编辑模态框事件
 */
export interface DataModalEmits extends GModalEmits {
  /**
   * 表单提交事件（通常在点击默认“保存”按钮时触发）
   */
  (event: 'submit', formData?: Record<string, any>): void

  /**
   * 需要重置内部表单时触发（例如重置字段）
   */
  (event: 'reset'): void
}

