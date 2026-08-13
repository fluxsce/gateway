import type { PaginationProps } from '@/utils/pagination'
import type { ToolbarProps } from '@/components/toolbar'
import type { VNodeChild } from 'vue'

/**
 * RsGrid 列配置（对齐 RsTableColumn，不兼容 vxe / 旧 GGrid）。
 */
export interface RsGridColumn<T = Record<string, any>> {
  /** 列唯一键，默认同时作为 dataIndex */
  key: string
  /** 列标题 */
  title: string
  /** 数据字段路径；缺省等于 key */
  dataIndex?: string
  width?: number | string
  minWidth?: number | string
  align?: 'left' | 'center' | 'right'
  fixed?: 'left' | 'right'
  sortable?: boolean
  filterable?: boolean
  /** 单元格溢出省略 */
  ellipsis?: boolean
  /** 是否显示，默认 true */
  visible?: boolean
  /** 文本格式化：(value, row, index) => string */
  formatter?: (value: unknown, row: T, index: number) => string
  /** 自定义单元格渲染 */
  render?: (row: T, index: number) => VNodeChild
}

/**
 * RsGrid 右键菜单项（对齐 RsContextMenuItem 语义）。
 * icon 建议传 Lucide kebab-case（如 `eye`、`pencil`）；仍兼容历史 ionicons 名（如 `EyeOutline`）。
 */
export interface RsGridMenuItem {
  key: string
  label: string
  icon?: string
  disabled?: boolean
  danger?: boolean
  separator?: boolean
  shortcut?: string
  /** 子菜单（分组）；有 children 时渲染为可展开分组 */
  children?: RsGridMenuItem[]
}

/**
 * RsGrid 右键菜单配置。
 */
export interface RsGridMenuConfig {
  /** 是否启用右键菜单，默认 true */
  enabled?: boolean
  /** 业务菜单项；无行时不展示 */
  items?: RsGridMenuItem[]
  /** 菜单点击回调 */
  onMenuClick?: (payload: { key: string; row?: any }) => void
}

/**
 * RsGrid 分页配置。
 */
export interface RsGridPaginationConfig extends PaginationProps {
  /** 是否显示分页，默认 false */
  show?: boolean
  /**
   * 是否显示跳转「确定」按钮
   * @default false（回车 / 失焦即可跳转）
   */
  showJumpConfirm?: boolean
}

/**
 * RsGrid 工具栏配置。
 * moduleId 可选：未传时使用 RsGrid 的 moduleId。
 */
export interface RsGridToolbarConfig
  extends Partial<Pick<ToolbarProps, 'moduleId'>>,
    Pick<ToolbarProps, 'buttons' | 'align'> {
  show?: boolean
  showRefresh?: boolean
  showColumnSetting?: boolean
  showFullscreen?: boolean
}

/**
 * RsGrid Props（基于 RsTable + RsPagination）。
 */
export interface RsGridProps {
  /** 模块 ID，用于权限与日志 */
  moduleId: string
  /** 表格数据（支持 ref） */
  data?: any[] | import('vue').Ref<any[]>
  /** 列配置（用 any 避免业务行类型与默认 Record 不兼容） */
  columns: RsGridColumn<any>[]
  /** 加载中（支持 ref） */
  loading?: boolean | import('vue').Ref<boolean>
  toolbarConfig?: RsGridToolbarConfig
  menuConfig?: RsGridMenuConfig
  paginationConfig?: RsGridPaginationConfig
  /** 边框，默认 true */
  border?: boolean
  /** 斑马纹，默认 true */
  stripe?: boolean
  /** 高度；`100%` 或未设置时启用 fill */
  height?: string | number
  maxHeight?: string | number
  /** 未设固定高度时是否 fill 父容器，默认 true */
  fill?: boolean
  /** 行唯一键字段，默认 id */
  rowKey?: string
  /** 是否显示复选框，默认 false */
  selectable?: boolean
  /** 是否显示序号列，默认 true */
  showIndex?: boolean
  /**
   * 远程排序：仅更新排序状态并抛出 sort-change，不排序当前页数据。
   * 默认 false（对当前页本地排序）。服务端排序时设为 true 并监听 sort-change。
   */
  remoteSort?: boolean
  /**
   * 树形表格配置（透传 RsTable.treeConfig）。
   * 使用 `any` 行类型，避免 `RsTableTreeConfig<Concrete>` 因 loadData 参数逆变无法赋给默认 `object`。
   */
  treeConfig?: import('niuma-ui').RsTableTreeConfig<any>
  /** 受控展开行 keys（树表 / 明细展开共用） */
  expandedRowKeys?: string[]
  /** 非受控初始展开 keys */
  defaultExpandedRowKeys?: string[]
}

/**
 * RsGrid 事件。
 */
export interface RsGridEmits {
  (event: 'toolbar-button-click', key: string): void
  (event: 'selection-change', selection: any[]): void
  (event: 'cell-click', params: { row: any; column: any; index: number }): void
  (event: 'row-dblclick', params: { row: any }): void
  (event: 'row-click', params: { row: any }): void
  (event: 'sort-change', params: any): void
  (event: 'filter-change', params: any): void
  (event: 'refresh'): void
  (event: 'menu-click', params: { key: string; row?: any }): void
  (event: 'page-change', params: { currentPage: number; pageSize: number }): void
  (event: 'update:expandedRowKeys', keys: string[]): void
  (event: 'expand-change', keys: string[]): void
}

/**
 * RsGrid 暴露方法（新 API，不再对齐 vxe / 旧 GGrid 方法名）。
 */
export interface RsGridExpose {
  refresh: () => void
  /** 当前勾选行 */
  getSelectedRows: () => any[]
  /** 当前高亮行 */
  getCurrentRow: () => any | null
  /** 勾选优先，否则高亮行 */
  getActiveRow: () => any | null
  /**
   * @deprecated 使用 getActiveRow；保留旧名兼容未改完的业务页
   */
  getSelectedOrCurrentRecord: () => any | null
  /** 勾选全部；无勾选时返回高亮行 0～1 条 */
  getActiveRows: () => any[]
  setSelectedRows: (rows: any[], selected: boolean) => void
  clearSelection: () => void
  getData: () => any[]
  reloadData: (data: any[]) => Promise<void>
  clearData: () => Promise<void>
  /**
   * 按 rowKey 合并字段到现有行（不替换 data 数组引用）。
   * 适合编辑回写；与浅监听 props.data 配套使用。
   */
  patchRows: (patches: Array<{ key: string; data: Record<string, any> }>) => void
  /**
   * 替换整表数据（使用传入数组引用），并清理失效选中/高亮。
   */
  replaceRows: (rows: any[]) => void
  toggleFullscreen: () => void
}
