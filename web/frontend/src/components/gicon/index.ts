/**
 * GIcon 图标组件（参考  Icon）
 * 统一封装 n-icon + getIconSync/getIcon，支持字符串按需加载与组件直传
 */

import { IconLibrary } from '@/utils/icon'
import type { Component, VNode } from 'vue'
import { h } from 'vue'
import GIcon from './GIcon.vue'
import type { GIconColor, GIconProps, GIconSize } from './types'

export { default as GIcon } from './GIcon.vue'
export { G_ICON_SIZE_MAP } from './types'
export type { GIconColor, GIconEmits, GIconInstance, GIconProps, GIconSize } from './types'

/** renderIconVNode 的可选配置（颜色、大小等） */
export interface RenderIconVNodeOptions {
  size?: GIconSize
  color?: GIconColor
}

/**
 * 渲染图标为 VNode（用于渲染函数 / 菜单项等）
 * 支持传入颜色、大小等，统一走 GIcon 组件。
 */
export function renderIconVNode(
  icon: Component | string | undefined,
  library: IconLibrary = IconLibrary.IONICONS5,
  options?: RenderIconVNodeOptions
): () => VNode | null {
  if (!icon) return () => null
  const libraryKey = library === IconLibrary.ANTD ? 'antd' : 'ionicons5'
  const attrs: Partial<GIconProps> = {
    icon: icon as string | Component,
    size: options?.size ?? 16,
    color: options?.color,
    library: libraryKey
  }
  return () => h(GIcon, attrs)
}
