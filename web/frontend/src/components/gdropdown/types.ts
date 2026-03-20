/**
 * GDropdown 下拉菜单组件类型
 * 基于 Naive UI NDropdown，与 xirang-dropdown 样式布局一致
 */

import type { DropdownOption } from 'naive-ui'

export type GDropdownTrigger = 'click' | 'hover'

export type GDropdownPlacement =
  | 'bottom'
  | 'top'
  | 'bottom-start'
  | 'bottom-end'
  | 'top-start'
  | 'top-end'

export type GDropdownSize = 'small' | 'medium' | 'large' | 'huge'

export interface GDropdownProps {
  /** 菜单项（与 Naive DropdownOption 兼容） */
  options?: DropdownOption[]
  /** 触发方式 */
  trigger?: GDropdownTrigger
  /** 是否禁用 */
  disabled?: boolean
  /** 弹出位置 */
  placement?: GDropdownPlacement
  /** 是否显示箭头 */
  showArrow?: boolean
  /** 延迟显示/隐藏（hover 时）ms */
  delay?: number
  /** 尺寸，默认 small */
  size?: GDropdownSize
}

export interface GDropdownEmits {
  (e: 'select', key: string | number, option: DropdownOption): void
}

export interface GDropdownInstance {
  close: () => void
  open: () => void
  visible: boolean
}
