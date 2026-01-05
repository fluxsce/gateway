import type { ContextMenuConfig } from '@/components/gmenu'
import type { PaginationProps } from '@/components/gpage'
import type { ToolbarProps } from '@/components/toolbar'
import type { VxeColumnProps, VxeGridProps, VxeTablePropTypes } from 'vxe-table'

/**
 * Grid 列配置
 * 扩展 VxeColumnProps
 */
export interface GridColumn extends Partial<VxeColumnProps> {
  /**
   * 列字段名
   */
  field: string

  /**
   * 列标题
   */
  title?: string

  /**
   * 列宽度
   */
  width?: number | string

  /**
   * 最小宽度
   */
  minWidth?: number | string

  /**
   * 对齐方式
   */
  align?: 'left' | 'center' | 'right'

  /**
   * 是否固定列
   */
  fixed?: 'left' | 'right'

  /**
   * 是否可排序
   */
  sortable?: boolean

  /**
   * 是否可筛选
   */
  filters?: any[]

  /**
   * 是否显示
   */
  visible?: boolean

  /**
   * 单元格渲染器
   */
  cellRender?: any

  /**
   * 编辑渲染器
   */
  editRender?: any

  /**
   * 插槽配置
   */
  slots?: {
    default?: string
    header?: string
    edit?: string
  }

  /**
   * 子列（多级表头）
   */
  children?: GridColumn[]
}

/**
 * Grid 右键菜单配置
 * 
 * 基于公共的 ContextMenuConfig，专门为 Grid 组件优化
 * Grid 组件使用 vxe-table，图标需要使用 vxe-table 的内置图标类名（如 vxe-icon-copy、vxe-icon-edit 等）
 * 
 * 注意：Grid 不支持 showCopyNode（Tree 专用），此属性已被移除
 * 
 * @example
 * ```ts
 * const menuConfig: GridMenuConfig = {
 *   enabled: true,
 *   showCopyRow: true,
 *   showCopyCell: true,
 *   customMenus: [
 *     {
 *       code: 'edit',
 *       name: '编辑',
 *       prefixIcon: 'vxe-icon-edit' // vxe-table 图标
 *     },
 *     {
 *       code: 'delete',
 *       name: '删除',
 *       prefixIcon: 'vxe-icon-delete'
 *     }
 *   ],
 *   onMenuClick: ({ code, row }) => {
 *     console.log('菜单点击', code, row)
 *   }
 * }
 * ```
 */
export type GridMenuConfig = Omit<ContextMenuConfig, 'showCopyNode'>

/**
 * Grid 分页配置
 * 扩展 PaginationProps，添加 show 控制
 */
export interface GridPaginationConfig extends PaginationProps {
  /**
   * 是否显示分页
   * @default false
   */
  show?: boolean
}

/**
 * Grid 工具栏配置
 */
export interface GridToolbarConfig
  extends Pick<ToolbarProps, 'moduleId' | 'buttons' | 'align'> {
  /**
   * 是否显示工具栏
   * @default false
   */
  show?: boolean

  /**
   * 是否显示刷新按钮
   * @default true
   */
  showRefresh?: boolean

  /**
   * 是否显示列设置按钮
   * @default true
   */
  showColumnSetting?: boolean

  /**
   * 是否显示全屏按钮
   * @default false
   */
  showFullscreen?: boolean
}

/**
 * Grid 配置
 */
export interface GridProps {
  /**
   * 模块ID（必填）
   * 用于标识表格所属的模块，便于权限控制和日志追踪
   * @example 'hub0001', 'hub0002'
   */
  moduleId: string

  /**
   * 表格数据
   * 支持 ref 或普通数组，Vue 模板会自动解包 ref
   */
  data?: any[] | import('vue').Ref<any[]>

  /**
   * 列配置
   */
  columns: GridColumn[]

  /**
   * 是否显示加载状态
   * 支持 ref 或普通布尔值，Vue 模板会自动解包 ref
   * @default false
   */
  loading?: boolean | import('vue').Ref<boolean>

  /**
   * 工具栏配置
   */
  toolbarConfig?: GridToolbarConfig

  /**
   * 右键菜单配置
   */
  menuConfig?: GridMenuConfig

  /**
   * 分页配置
   */
  paginationConfig?: GridPaginationConfig

  /**
   * 是否显示边框
   * @default true
   */
  border?: boolean

  /**
   * 是否显示斑马纹
   * @default true
   */
  stripe?: boolean

  /**
   * 表格高度
   * 'auto' | number | 'max-content'
   */
  height?: string | number

  /**
   * 表格最大高度
   */
  maxHeight?: string | number

  /**
   * 是否自动高度
   * @default false
   */
  autoResize?: boolean

  /**
   * 行唯一键字段
   * @default 'id'
   */
  rowId?: string

  /**
   * 是否显示复选框
   * @default false
   */
  showCheckbox?: boolean

  /**
   * 是否显示序号列
   * @default false
   */
  showSeq?: boolean

  /**
   * 序号列配置
   * 继承 vxe-table 的 SeqConfig 类型
   */
  seqConfig?: VxeTablePropTypes.SeqConfig

  /**
   * 是否显示表尾合计
   * @default false
   */
  showFooter?: boolean

  /**
   * 表尾数据
   */
  footerData?: any[][]

  /**
   * 表尾合计方法
   */
  footerMethod?: (params: any) => any[][]

  /**
   * 排序配置
   * 继承 vxe-table 的 SortConfig 类型
   */
  sortConfig?: VxeTablePropTypes.SortConfig

  /**
   * 筛选配置
   * 继承 vxe-table 的 FilterConfig 类型
   */
  filterConfig?: VxeTablePropTypes.FilterConfig

  /**
   * 编辑配置
   * 继承 vxe-table 的 EditConfig 类型
   */
  editConfig?: VxeTablePropTypes.EditConfig

  /**
   * 树形配置
   * 继承 vxe-table 的 TreeConfig 类型
   */
  treeConfig?: VxeTablePropTypes.TreeConfig

  /**
   * 展开配置
   * 继承 vxe-table 的 ExpandConfig 类型
   */
  expandConfig?: VxeTablePropTypes.ExpandConfig

  /**
   * 导出配置
   * 继承 vxe-table 的 ExportConfig 类型
   */
  exportConfig?: VxeTablePropTypes.ExportConfig

  /**
   * 打印配置
   * 继承 vxe-table 的 PrintConfig 类型
   */
  printConfig?: VxeTablePropTypes.PrintConfig

  /**
   * 其他 vxe-grid 原生配置
   */
  gridOptions?: Partial<VxeGridProps>
}

/**
 * Grid 事件
 */
export interface GridEmits {
  /**
   * 工具栏按钮点击事件
   */
  (event: 'toolbar-button-click', key: string): void

  /**
   * 复选框选中变化事件
   */
  (event: 'checkbox-change', selection: any[]): void

  /**
   * 单元格点击事件
   */
  (event: 'cell-click', params: any): void

  /**
   * 单元格双击事件
   */
  (event: 'cell-dblclick', params: any): void

  /**
   * 行点击事件
   */
  (event: 'row-click', params: any): void

  /**
   * 排序变化事件
   */
  (event: 'sort-change', params: any): void

  /**
   * 筛选变化事件
   */
  (event: 'filter-change', params: any): void

  /**
   * 刷新事件
   */
  (event: 'refresh'): void

  /**
   * 右键菜单点击事件
   */
  (event: 'menu-click', params: { code: string; row?: any; column?: any }): void

  /**
   * 编辑激活事件
   */
  (event: 'edit-actived', params: any): void

  /**
   * 编辑关闭事件
   */
  (event: 'edit-closed', params: any): void

  /**
   * 分页变化事件
   */
  (event: 'page-change', params: { currentPage: number; pageSize: number }): void
}

/**
 * Grid 暴露的方法
 */
export interface GridExpose {
  /**
   * 获取表格实例
   */
  getGridInstance: () => any

  /**
   * 刷新数据
   */
  refresh: () => void

  /**
   * 获取选中行
   */
  getCheckboxRecords: () => any[]

  /**
   * 获取当前高亮的行（点击选中的行）
   */
  getCurrentRecord: () => any

  /**
   * 设置选中行
   */
  setCheckboxRow: (rows: any[], checked: boolean) => void

  /**
   * 清空选中
   */
  clearCheckboxRow: () => void

  /**
   * 获取全部数据
   */
  getTableData: () => any[]

  /**
   * 插入行
   */
  insert: (record: any) => Promise<any>

  /**
   * 插入行到指定位置
   */
  insertAt: (record: any, row: any) => Promise<any>

  /**
   * 删除行
   */
  remove: (row: any) => Promise<any>

  /**
   * 批量删除行
   */
  removeCheckboxRow: () => Promise<any>

  /**
   * 获取表格数据（包含新增、删除、修改）
   */
  getRecordset: () => {
    insertRecords: any[]
    removeRecords: any[]
    updateRecords: any[]
  }

  /**
   * 清空表格数据
   */
  clearData: () => Promise<void>

  /**
   * 重新加载数据
   */
  reloadData: (data: any[]) => Promise<void>

  /**
   * 导出数据
   */
  exportData: (options?: any) => Promise<any>

  /**
   * 打印表格
   */
  print: (options?: any) => void

  /**
   * 全屏切换
   */
  zoom: () => void

  /**
   * 手动验证
   */
  validate: (callback?: (valid: boolean) => void) => Promise<boolean>
}

