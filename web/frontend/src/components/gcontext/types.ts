/**
 * GContext 右键菜单组件类型定义
 * 基于 vxe-context-menu 的简化 API
 */

import type { Component } from 'vue'

/** 菜单项类型 */
export type GContextMenuItemType = 'default' | 'divider' | 'group'

/** 菜单项 */
export interface GContextMenuItem {
  /** 菜单键值（对应 VXE 的 code） */
  code: string
  /** 菜单名称 */
  name?: string
  type?: GContextMenuItemType
  /** 图标（字符串名称或组件） */
  icon?: string | Component
  iconColor?: string
  prefixIcon?: string | Component
  prefixIconColor?: string
  suffixIcon?: string | Component
  suffixIconColor?: string
  /** 快捷键提示 */
  shortcut?: string
  visible?: boolean
  disabled?: boolean
  children?: GContextMenuItem[]
  data?: unknown
  onClick?: () => void
}

export interface GContextProps {
  /** 是否显示（v-model） */
  show?: boolean
  /** 菜单项列表（一维或二维数组，二维时每子数组为一组）；传入后由组件内部按 moduleId 做权限过滤 */
  options?: GContextMenuItem[] | GContextMenuItem[][]
  /** 横坐标 */
  x?: number
  /** 纵坐标 */
  y?: number
  /** 是否显示图标 */
  showIcon?: boolean
  /** 是否显示快捷键 */
  showShortcut?: boolean
  zIndex?: number
  className?: string
  transfer?: boolean
  /** 是否启用（Tree/Grid 等用，为 false 时不显示菜单） */
  enabled?: boolean
  /** 是否显示“复制节点数据”项（Tree 用，由调用方在 options 前追加该条时使用） */
  showCopyNode?: boolean
  /** 是否显示“复制行数据”项（Grid 用） */
  showCopyRow?: boolean
  /** 是否显示“复制单元格”项（Grid 用） */
  showCopyCell?: boolean
  /** 模块ID，传入后组件内部按 moduleId:code 做按钮权限校验并过滤菜单项 */
  moduleId?: string
  /** 菜单点击回调（Tree/Grid 用） */
  onMenuClick?: (menu: { code: string; node?: any; row?: any; column?: any }) => void
}

export interface GContextEmits {
  'update:show': [show: boolean]
  'select': [item: GContextMenuItem, event: MouseEvent]
  'open': []
  'close': []
}

export interface GContextInstance {
  show: (x?: number, y?: number) => void
  hide: () => void
  toggle: () => void
  open: () => void
  close: () => void
}
