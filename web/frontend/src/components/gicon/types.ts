/**
 * GIcon 组件类型定义
 */

import type { Component } from 'vue'

export type GIconSize =
  | 'tiny'
  | 'small'
  | 'medium'
  | 'large'
  | 'huge'
  | 'ssm'
  | 'sm'
  | 'md'
  | 'lg'
  | number
  | `${number}`

export const G_ICON_SIZE_MAP: Record<string, number> = {
  tiny: 14,
  small: 16,
  medium: 20,
  large: 24,
  huge: 32,
  ssm: 12,
  sm: 16,
  md: 20,
  lg: 24,
}

export type GIconColor = 'primary' | 'success' | 'warning' | 'error' | 'info' | string

export interface GIconProps {
  /** 图标组件或图标名称（字符串时按需加载，使用 getIconSync 缓存） */
  icon?: Component | string
  size?: GIconSize
  color?: GIconColor
  disabled?: boolean
  class?: string
  style?: string | Record<string, unknown>
  spin?: boolean
  spinSpeed?: number
  /** 图标库（仅当 icon 为字符串时有效），与 @/utils/icon IconLibrary 一致 */
  library?: 'ionicons5' | 'antd'
}

export interface GIconEmits {
  click: [event: MouseEvent]
  mouseenter: [event: MouseEvent]
  mouseleave: [event: MouseEvent]
}

export interface GIconInstance {
  $el: HTMLElement | null
}
