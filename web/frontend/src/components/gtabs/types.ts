/**
 * GTabs 标签页组件类型
 * 与 xirang-tabs 一致：仅导航栏，支持拖拽、关闭、右键菜单、溢出下拉
 */

export interface GTabsTabItem {
  /** 唯一标识 */
  tabId: string
  /** 配置类型（用于路由映射等） */
  configType?: string
  /** 节点 ID（关联树节点等） */
  nodeId?: string
  /** 标签标题 */
  title: string
  /** 路由路径 */
  path?: string
  /** 图标（GIcon 支持的字符串或组件） */
  icon?: string | import('vue').Component
  /** 图标颜色 */
  iconColor?: string
  /** 是否可关闭 */
  closable?: boolean
  /** 是否固定（不可关闭、不可拖拽） */
  fixed?: boolean
  /** 自定义数据 */
  meta?: Record<string, unknown>
}

export type GTabsType = 'line' | 'card'

export interface GTabsProps {
  /** 标签页数据 */
  tabs?: GTabsTabItem[]
  /** 当前激活的标签页 tabId */
  activeTabId?: string
  /** 标签页类型 */
  type?: GTabsType
  /** 是否可拖拽排序 */
  draggable?: boolean
  /** 是否显示关闭按钮 */
  closable?: boolean
  /** 是否显示右键菜单 */
  contextMenu?: boolean
  /** 最大标签页数量 */
  maxTabs?: number
}

export interface GTabsEmits {
  (e: 'change', tabId: string): void
  (e: 'close', tabId: string): void
  (e: 'sort', tabs: GTabsTabItem[]): void
  (e: 'context-menu', action: string, tabId: string): void
  (e: 'update:tabs', tabs: GTabsTabItem[]): void
  (e: 'update:activeTabId', tabId: string): void
}

export interface GTabsInstance {
  addTab: (tab: GTabsTabItem) => void
  removeTab: (tabId: string) => void
  closeOthers: (tabId: string) => void
  closeLeft: (tabId: string) => void
  closeRight: (tabId: string) => void
  closeAll: () => void
  activateTab: (tabId: string, shouldScroll?: boolean) => void
}
