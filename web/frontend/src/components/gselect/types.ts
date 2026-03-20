import type { SelectGroupOption, SelectOption } from 'naive-ui'

export type GSelectValue = string | number | Array<string | number> | null

export type GSelectOption = SelectOption | SelectGroupOption

export interface GSelectProps {
  /** 当前值（支持单选与多选） */
  value?: GSelectValue
  /** 选项数据 */
  options?: GSelectOption[]
  /** 占位文本 */
  placeholder?: string
  /** 是否禁用 */
  disabled?: boolean
  /** 是否允许清空 */
  clearable?: boolean
  /** 组件尺寸 */
  size?: 'small' | 'medium' | 'large'
  /** 是否多选 */
  multiple?: boolean
  /** 是否可过滤 */
  filterable?: boolean
}

export interface GSelectEmits {
  (e: 'update:value', value: GSelectValue): void
}

export interface GSelectInstance {
  focus: () => void
  blur: () => void
}
