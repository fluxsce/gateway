export type GDropdownTrigger = 'click' | 'hover' | 'manual'
export type GDropdownPlacement =
  | 'bottom'
  | 'top'
  | 'bottom-start'
  | 'bottom-end'
  | 'top-start'
  | 'top-end'
export type GDropdownSize = 'small' | 'medium' | 'large' | 'huge'

export type GDropdownOption = {
  key?: string | number
  label?: string
  value?: string | number
  disabled?: boolean
  type?: string
  [key: string]: unknown
}

export interface GDropdownProps {
  options?: GDropdownOption[]
  placement?: GDropdownPlacement
  disabled?: boolean
  trigger?: GDropdownTrigger
  showArrow?: boolean
  delay?: number
  size?: GDropdownSize
}

export interface GDropdownEmits {
  (e: 'select', key: string | number, option: GDropdownOption): void
}

export interface GDropdownInstance {
  close: () => void
  open: () => void
  readonly visible: boolean
}
