import type { Component, VNode } from 'vue'
import type { RsFormNamePath, RsFormRuleItem } from '@/ui'

/**
 * 自定义字段渲染上下文。
 * value / onUpdate 绑定当前 field 的 NamePath，经 Form 写入；
 * setFieldValue 用于改其它路径（如关联字段）。
 */
export interface RsDataFormRenderContext {
  value: unknown
  onUpdate: (value: unknown) => void
  setFieldValue: (name: RsFormNamePath, value: unknown) => void
}

/**
 * RsDataForm 字段类型（基于 niuma-ui 控件）
 * - input / textarea / number：文本与数字录入
 * - select / switch：选择类
 * - date*：日期与时间（RsDatePicker，valueFormat=iso，展示墙钟、绑定 RFC3339）
 * - fieldset：分组容器（可嵌套 children）
 * - file：文件列表（RsUpload，表单值为 File[]；提交时由业务读内容）
 * - custom：业务自定义渲染
 */
export type RsDataFormFieldType =
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
 * select 选项项
 */
export interface RsDataFormSelectOption {
  /** 显示文案 */
  label: string
  /** 选项值（提交时按原始类型回写，控件内部使用 string） */
  value: string | number
  /** 是否禁用该选项 */
  disabled?: boolean
}

/**
 * RsDataForm 字段配置
 * 描述单个表单项的展示、交互与校验。
 */
export interface RsDataFormField {
  /** 字段名（RsForm NamePath）。`config.headerConfig.x` 对应 model.config.headerConfig.x，不是根上的扁平 key */
  field: string

  /**
   * 字段标签
   * 支持字符串，或根据当前表单数据动态生成
   */
  label: string | ((formData: Record<string, any>) => string)

  /**
   * 字段提示
   * - 字符串：内置 label 控件走 hint；外部 label 走 RsTooltip icon
   * - Component / VNode：直接渲染在标签旁
   * - 函数：按 formData 动态生成上述类型
   */
  tips?:
    | string
    | Component
    | VNode
    | ((formData: Record<string, any>) => string | Component | VNode)

  /**
   * 字段类型
   * @default 'input'
   */
  type?: RsDataFormFieldType

  /** 占位符文本 */
  placeholder?: string

  /** 默认值（无 initialData 对应 key 时使用） */
  defaultValue?: any

  /**
   * 是否显示该字段
   * 函数可读取 formData._mode 区分新增/编辑/查看
   * @default true
   */
  show?: boolean | ((formData: Record<string, any>) => boolean)

  /**
   * 是否必填
   * 未配置 rules 时会自动生成基础 required 规则
   * @default false
   */
  required?: boolean

  /**
   * 是否禁用
   * view 模式、edit+primary 时会强制禁用
   * @default false
   */
  disabled?: boolean

  /**
   * 是否为主键字段
   * 编辑模式下自动禁用，防止修改主键
   * @default false
   */
  primary?: boolean

  /**
   * 是否可清空（input / select 等）
   * @default true
   */
  clearable?: boolean

  /** select 选项列表 */
  options?: RsDataFormSelectOption[]

  /**
   * 字段栅格占位（1-24）
   * @default 12
   */
  span?: number

  /**
   * 所属页签 key
   * 未设置时归入第一个可见页签
   */
  tabKey?: string

  /**
   * 子字段列表
   * 仅 type 为 fieldset 时生效，支持嵌套
   */
  children?: RsDataFormField[]

  /**
   * 自定义渲染（type 为 custom）
   * 值变更走 ctx.onUpdate（经 field-update 写入 Form），不要直接改 formData。
   */
  render?: (
    formData: Record<string, any>,
    ctx?: RsDataFormRenderContext,
  ) => Component | VNode

  /**
   * RsForm 校验规则（单条或数组），优先于 required 自动规则。
   * validator 签名为 `(value, ctx)`，返回 `true` 或错误文案。
   */
  rules?: RsFormRuleItem | RsFormRuleItem[]

  /**
   * 透传给底层控件的额外属性
   * 业务专用键（如 checkedValue、onUpdateValue、file callbacks）会在渲染层剥离
   */
  props?: Record<string, any>
}

/**
 * 数据编辑模式
 * - create：新增
 * - edit：编辑
 * - view：只读查看
 */
export type RsDataModalMode = 'create' | 'edit' | 'view'

/**
 * 数据表单页签配置
 */
export interface RsDataFormTab {
  /** 页签唯一标识（与 RsDataFormField.tabKey 对应） */
  key: string
  /** 页签显示名称 */
  label: string
  /**
   * 是否显示该页签
   * @default true
   */
  show?: boolean | ((formData: Record<string, any>) => boolean)
}

/**
 * RsDataFormFields（字段栅格渲染器）Props
 * 由 RsDataForm 内部使用，也可在 fieldset 内递归嵌套。
 */
export interface RsDataFormFieldsProps {
  /** 当前层级要渲染的字段列表 */
  fields: RsDataFormField[]
  /**
   * 表单数据模型（只读绑定）
   * 值变更请通过 field-update 事件回传，由父级写入
   */
  formModel: Record<string, any>
  /** 按 NamePath 写入（自定义字段改其它路径时使用） */
  setFieldValue?: (name: RsFormNamePath, value: unknown) => void
  /**
   * 业务模式，影响禁用与主键锁定
   * @default 'create'
   */
  mode?: RsDataModalMode
  /**
   * 标签宽度（CSS 长度或数字像素）
   * @default 'auto'
   */
  labelWidth?: number | string
  /**
   * 标签布局
   * @default 'top'
   */
  labelPlacement?: 'left' | 'top'
  /**
   * 标签对齐（labelPlacement 为 left 时生效）
   * @default 'left'
   */
  labelAlign?: 'left' | 'right'
  /**
   * 控件尺寸（会映射为 niuma-ui sm/md/lg）
   * @default 'small'
   */
  size?: 'small' | 'medium' | 'large'
  /**
   * 栅格列数
   * @default 24
   */
  cols?: number
  /**
   * 列间距（px）
   * @default 16
   */
  xGap?: number
  /**
   * 行间距（px）
   * @default 8
   */
  yGap?: number
}

/**
 * RsDataFormFields 事件
 * 字段值变更：select 已还原为 options 原始类型；switch 已映射 checked/unchecked
 */
export type RsDataFormFieldsEmits = {
  'field-update': [field: RsDataFormField, value: any]
}

/**
 * RsDataForm Props
 */
export interface RsDataFormProps {
  /**
   * 当前业务模式
   * @default 'create'
   */
  mode?: RsDataModalMode
  /** 表单字段配置 */
  formFields?: RsDataFormField[]
  /** 页签配置；不传时按字段 tabKey 推导，无 tabKey 则单页 */
  formTabs?: RsDataFormTab[]
  /** 初始表单数据（编辑/查看时填入） */
  initialData?: Record<string, any>
  /**
   * 是否注入 formData._mode（供 show/label 回调使用）
   * @default true
   */
  injectMode?: boolean
  /**
   * 是否显示底部操作区
   * @default false
   */
  showFooter?: boolean
  /**
   * 是否显示提交按钮（需 showFooter）
   * @default false
   */
  showSubmit?: boolean
  /**
   * 提交按钮文案
   * @default '保存'
   */
  submitText?: string
  /**
   * 提交按钮加载态
   * @default false
   */
  submitLoading?: boolean
  /**
   * 标签宽度
   * @default 'auto'
   */
  labelWidth?: number | string
  /**
   * 标签布局
   * @default 'top'
   */
  labelPlacement?: 'left' | 'top'
  /**
   * 标签对齐
   * @default 'left'
   */
  labelAlign?: 'left' | 'right'
  /**
   * 控件尺寸
   * @default 'small'
   */
  size?: 'small' | 'medium' | 'large'
  /**
   * 栅格列数
   * @default 24
   */
  cols?: number
  /**
   * 列间距（px）
   * @default 16
   */
  xGap?: number
  /**
   * 行间距（px）
   * @default 8
   */
  yGap?: number
}

/**
 * RsDataForm 事件
 */
export interface RsDataFormEmits {
  /** 校验通过后的提交 */
  (event: 'submit', formData?: Record<string, any>): void
  /** 表单模型变化 */
  (event: 'update:modelValue', formData: Record<string, any>): void
  /** 重置完成 */
  (event: 'reset'): void
}

/**
 * RsDataForm 暴露给父组件的方法
 */
export interface RsDataFormExpose {
  /** 执行校验，返回是否通过 */
  validate: () => Promise<boolean>
  /** 按 initialData / 默认值重新初始化 */
  reset: () => void
  /** 获取提交用表单数据（已剥离 _mode，select 已还原类型） */
  getFormData: () => Record<string, any>
  /** 直接覆盖表单数据 */
  setFormData: (data: Record<string, any>) => void
  /** 按 NamePath 读取当前 store 中的值 */
  getFieldValue: (name: RsFormNamePath) => unknown
  /** 按 NamePath 写入当前 store，并同步 update:modelValue */
  setFieldValue: (name: RsFormNamePath, value: unknown) => void
  /** 当前 store 快照（不含 _ 前缀内部键） */
  getFieldsValue: () => Record<string, any>
  /** 按路径批量写入 */
  setFieldsValue: (values: Record<string, unknown>) => void
  /** 底层 RsForm 实例引用 */
  formRef: any
}

/**
 * RsDataFormModal Props
 */
export interface RsDataFormModalProps {
  /**
   * 是否显示弹窗
   * @default false
   */
  visible?: boolean
  /** 标题；为空时按 mode 使用 common.formModal.* */
  title?: string
  /**
   * 弹窗宽度：niuma-ui 预设（sm/md/lg…）或 CSS 宽度
   * @default '720px'
   */
  width?: string | number
  /**
   * 业务模式
   * @default 'create'
   */
  mode?: RsDataModalMode
  /** 表单字段 */
  formFields?: RsDataFormField[]
  /** 表单页签 */
  formTabs?: RsDataFormTab[]
  /** 初始数据 */
  initialData?: Record<string, any>
  /**
   * 确认成功后是否自动关闭
   * @default true
   */
  autoCloseOnConfirm?: boolean
  /**
   * 关闭时是否自动 reset 表单
   * @default false
   */
  autoResetOnClose?: boolean
  /**
   * 是否显示底部按钮区
   * @default true
   */
  showFooter?: boolean
  /**
   * 是否显示取消按钮
   * @default true
   */
  showCancel?: boolean
  /**
   * 是否显示确认按钮（view 模式始终隐藏）
   * @default true
   */
  showConfirm?: boolean
  /**
   * 取消按钮文案；未传时用 common.cancel
   */
  cancelText?: string
  /**
   * 确认按钮文案；未传时用 common.save
   */
  confirmText?: string
  /**
   * 确认按钮加载态
   * @default false
   */
  confirmLoading?: boolean
  /**
   * 是否显示遮罩
   * @default false
   */
  mask?: boolean
  /**
   * 是否模态（锁焦点 / 拦截背后点击）。
   * 默认 false，与无遮罩窗口一致，不挡其它模块操作。
   * @default false
   */
  modal?: boolean
  /**
   * 点击遮罩是否可关闭
   * @default false
   */
  maskClosable?: boolean
  /**
   * 是否显示关闭按钮
   * @default true
   */
  closable?: boolean
  /**
   * 是否可拖拽
   * @default true
   */
  draggable?: boolean
  /**
   * 是否显示全屏切换
   * @default true
   */
  showFullscreenToggle?: boolean
  /**
   * 表单标签布局（透传 RsDataForm；对齐旧 GDataFormModal 顶标签栅格）
   * @default 'top'
   */
  labelPlacement?: 'left' | 'top'
  /**
   * 标签对齐（labelPlacement 为 left 时）
   * @default 'left'
   */
  labelAlign?: 'left' | 'right'
  /**
   * 标签宽度
   * @default 'auto'
   */
  labelWidth?: number | string
  /**
   * 控件尺寸
   * @default 'small'
   */
  size?: 'small' | 'medium' | 'large'
  /**
   * 弹窗挂载点（对齐 RsDialog.teleportTo）
   * - string / HTMLElement：Teleport 到指定容器（如 `#hub0002`，避免挂到全局 body）
   * - false：禁用 Teleport，就地渲染在当前组件树内
   * - undefined：使用 Reka/ConfigProvider 默认（通常为 body）
   */
  to?: string | HTMLElement | false
}

/**
 * RsDataFormModal 事件
 */
export interface RsDataFormModalEmits {
  /** 显隐变化（v-model:visible） */
  (event: 'update:visible', visible: boolean): void
  /** 校验通过后的业务提交 */
  (event: 'submit', formData?: Record<string, any>): void
  /** 确认（与 submit 同时触发，便于兼容旧监听） */
  (event: 'confirm', formData?: Record<string, any>): void
  /** 取消 */
  (event: 'cancel'): void
  /** 关闭（含取消、遮罩、确认后自动关闭等） */
  (event: 'close'): void
  /** 关闭且 autoResetOnClose 时触发 */
  (event: 'reset'): void
  /** 打开动画结束 */
  (event: 'after-enter'): void
  /** 关闭动画结束 */
  (event: 'after-leave'): void
}
