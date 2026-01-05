import type { ContextMenuConfig } from '@/components/gmenu'
import type { TreeOption } from 'naive-ui'

/**
 * GTree 组件属性
 */
export interface GTreeProps {
  /** 树形数据 */
  data?: TreeOption[]
  /** 默认展开的节点 key */
  defaultExpandedKeys?: string[]
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
  /** 模块ID（用于权限控制） */
  moduleId?: string
  /** 右键菜单配置 */
  menuConfig?: ContextMenuConfig
}

/**
 * GTree 组件事件
 */
export interface GTreeEmits {
  /** 节点选中变化 */
  (e: 'update:checkedKeys', keys: string[]): void
  /** 节点展开变化 */
  (e: 'update:expandedKeys', keys: string[]): void
  /** 节点点击 */
  (e: 'select', keys: string[], option: TreeOption): void
  /** 节点双击 */
  (e: 'dblclick', option: TreeOption): void
  /** 右键菜单点击 */
  (e: 'menu-click', params: { code: string; node?: TreeOption }): void
}

