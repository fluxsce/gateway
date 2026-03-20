import type { GContextMenuItem, GContextProps } from '@/components/gcontext'
import type { TreeOption } from 'naive-ui'

/** 右键菜单配置：对象形式（固定选项 + 可选复制）或函数形式（按节点/空白动态返回） */
export type GTreeMenuConfig =
  | (Partial<GContextProps> & { options?: GContextMenuItem[] | GContextMenuItem[][] })
  | ((node: TreeOption | null) => GContextMenuItem[] | GContextMenuItem[][])

/**
 * GTree 组件属性
 */
export interface GTreeProps {
  /** 树形数据 */
  data?: TreeOption[]
  /** 默认展开的节点 key */
  defaultExpandedKeys?: string[]
  /** 是否默认展开所有节点（参考 ） */
  defaultExpandAll?: boolean
  /** 默认选中的节点 key */
  defaultCheckedKeys?: string[]
  /** 受控的选中节点 key（优先级高于 defaultCheckedKeys） */
  checkedKeys?: string[]
  /** 是否显示复选框 */
  checkable?: boolean
  /** 是否启用级联选择 */
  cascade?: boolean
  /** 选中策略：'all' | 'parent' | 'child' */
  checkStrategy?: 'all' | 'parent' | 'child'
  /** 是否可拖拽 */
  draggable?: boolean
  /** 拖拽时是否允许放置（参考 ） */
  allowDrop?: (
    dragNode: TreeOption,
    dropNode: TreeOption,
    position: 'before' | 'after' | 'inner'
  ) => boolean
  /** 是否显示连接线 */
  showLine?: boolean
  /** 是否显示图标 */
  showIcon?: boolean
  /** 默认节点整行撑开（block-line） */
  blockLine?: boolean
  /** 节点 key 字段名 */
  keyField?: string
  /** 节点 label 字段名 */
  labelField?: string
  /** 子节点字段名 */
  childrenField?: string
  /** 节点高度 */
  nodeHeight?: number
  /** 虚拟滚动 */
  virtualScroll?: boolean
  /** 自定义节点渲染 */
  renderLabel?: (info: any) => any
  /** 自定义节点前缀（支持函数或图标名称字符串） */
  renderPrefix?: ((info: any) => any) | string
  /** 自定义节点后缀（支持函数或图标名称字符串） */
  renderSuffix?: ((info: any) => any) | string
  /** 是否启用文本省略显示 */
  ellipsis?: boolean
  /** 省略显示的最大行数（当 ellipsis 为 true 时有效） */
  ellipsisLineClamp?: number
  /** 省略显示的 tooltip 配置（当 ellipsis 为 true 时有效） */
  ellipsisTooltip?: boolean | { width?: number | 'trigger'; [key: string]: any }
  /** 是否可搜索（参考 ） */
  filterable?: boolean
  /** 自定义搜索方法 */
  filterMethod?: (keyword: string, node: TreeOption) => boolean
  /** 展开/折叠图标名（参考 ） */
  iconOpen?: string
  iconClose?: string
  /** 懒加载子节点（参考 ） */
  loadMethod?: (node: TreeOption, resolve: (children: TreeOption[]) => void) => Promise<void> | void
  /** 模块ID（用于权限控制） */
  moduleId?: string
  /** 右键菜单配置（对象或 (node) => items，空白时 node 为 null） */
  menuConfig?: GTreeMenuConfig
}

/**
 * GTree 组件事件（对齐 ）
 */
export interface GTreeEmits {
  (e: 'update:checkedKeys', keys: string[]): void
  (e: 'update:expandedKeys', keys: string[]): void
  /** 节点点击 */
  (e: 'select', keys: string[], option: TreeOption): void
  /** 节点双击 */
  (e: 'dblclick', option: TreeOption): void
  /** 节点展开/折叠（参考 ） */
  (e: 'node-expand', node: TreeOption, expanded: boolean): void
  /** 复选框变化（参考 ：当前节点、是否选中、当前所有选中节点） */
  (e: 'check-change', node: TreeOption, checked: boolean, checkedNodes: TreeOption[]): void
  /** 拖拽放置（参考 ） */
  (e: 'node-drop', dragNode: TreeOption, dropNode: TreeOption, position: 'before' | 'after' | 'inner', event: DragEvent): void
  (e: 'menu-click', params: { code: string; node?: TreeOption }): void
  (e: 'blank-contextmenu', event: MouseEvent): void
}

/** GTree 暴露的实例方法（参考 ） */
export interface GTreeInstance {
  getTreeRef: () => any
  getNode: (key: string | number) => TreeOption | null
  getCheckedNodes: () => TreeOption[]
  getCheckedKeys: () => (string | number)[]
  setCheckedKeys: (keys: (string | number)[]) => void
  setChecked: (key: string | number, checked: boolean) => void
  getCurrentNode: () => TreeOption | null
  getCurrentKey: () => string | number | null
  setCurrentKey: (key: string | number) => void
  expandNode: (key: string | number) => void
  collapseNode: (key: string | number) => void
  expandAll: () => void
  collapseAll: () => void
  scrollTo: (key: string | number) => void
  refresh: () => void
}

