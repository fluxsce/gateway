/**
 * GTabs 标签页组件类型
 * 顶栏多页签：业务侧仍用 GTabsTabItem；导航 UI 由 RsTabs 承担。
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
  /** 图标：Lucide kebab-case（RsTabs）或历史 Ionicons / 组件 */
  icon?: string | import('vue').Component
  /** 图标颜色（仅历史 GIcon 路径使用） */
  iconColor?: string
  /** 是否可关闭 */
  closable?: boolean
  /** 是否固定（不可关闭、不可拖拽） */
  fixed?: boolean
  /** 自定义数据 */
  meta?: Record<string, unknown>
}

export type GTabsType = 'line' | 'card'
export type GTabsSize = 'sm' | 'md'

export interface GTabsProps {
  /** 标签页数据 */
  tabs?: GTabsTabItem[]
  /** 当前激活的标签页 tabId */
  activeTabId?: string
  /** 标签页类型（映射 RsTabs variant） */
  type?: GTabsType
  /** 尺寸（透传 RsTabs size，默认 md） */
  size?: GTabsSize
  /** 是否可拖拽排序（透传 RsTabs） */
  draggable?: boolean
  /** 是否显示关闭按钮（透传 RsTabs） */
  closable?: boolean
  /** 是否显示右键菜单（透传 RsTabs） */
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
