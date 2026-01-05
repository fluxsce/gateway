import type { VNode } from 'vue'

/**
 * 自定义菜单项配置
 */
export interface ContextMenuItemConfig {
  /**
   * 菜单项代码（唯一标识）
   */
  code: string

  /**
   * 菜单项名称（显示文本）
   */
  name: string

  /**
   * 前缀图标
   * - 字符串：图标类名或图标名称
   *   - 对于 Tree：支持 naive-ui 内置图标或 FontAwesome 图标类名（如 'fas fa-copy'）
   *   - 对于 Grid：支持 vxe-table 内置图标类名（如 'vxe-icon-copy'、'vxe-icon-lock'、'vxe-icon-user'、'vxe-icon-code' 等）
   * - 函数：返回 VNode，支持自定义图标组件
   */
  prefixIcon?: string | ((params?: {}) => VNode | VNode[])

  /**
   * 后缀图标
   * - 字符串：图标类名或图标名称（同上）
   * - 函数：返回 VNode，支持自定义图标组件
   * 
   * 注意：Grid 组件使用 vxe-table 时，后缀图标可能不被支持，建议使用 prefixIcon
   */
  suffixIcon?: string | ((params?: {}) => VNode | VNode[])

  /**
   * 是否禁用
   * @default false
   */
  disabled?: boolean

  /**
   * 是否显示
   * @default true
   */
  visible?: boolean

  /**
   * 子菜单项（支持多级菜单）
   */
  children?: ContextMenuItemConfig[]
}

/**
 * 公共右键菜单配置
 * 适用于 Tree、Grid 等组件
 * 
 * @example
 * ```ts
 * // Tree 使用示例
 * const menuConfig: ContextMenuConfig = {
 *   enabled: true,
 *   showCopyNode: true,
 *   customMenus: [
 *     {
 *       code: 'edit',
 *       name: '编辑',
 *       prefixIcon: 'fas fa-edit'
 *     },
 *     {
 *       code: 'delete',
 *       name: '删除',
 *       prefixIcon: 'fas fa-trash'
 *     }
 *   ],
 *   onMenuClick: ({ code, node }) => {
 *     console.log('菜单点击', code, node)
 *   }
 * }
 * 
 * // Grid 使用示例
 * const menuConfig: ContextMenuConfig = {
 *   enabled: true,
 *   showCopyRow: true,
 *   showCopyCell: true,
 *   customMenus: [
 *     {
 *       code: 'edit',
 *       name: '编辑',
 *       prefixIcon: 'vxe-icon-edit'
 *     }
 *   ],
 *   onMenuClick: ({ code, row }) => {
 *     console.log('菜单点击', code, row)
 *   }
 * }
 * ```
 */
export interface ContextMenuConfig {
  /**
   * 是否启用右键菜单
   * @default true
   */
  enabled?: boolean

  /**
   * 是否显示复制节点数据菜单项（Tree 组件使用）
   * @default true
   */
  showCopyNode?: boolean

  /**
   * 是否显示复制行数据菜单项（Grid 组件使用）
   * @default true
   */
  showCopyRow?: boolean

  /**
   * 是否显示复制单元格数据菜单项（Grid 组件使用）
   * @default true
   */
  showCopyCell?: boolean

  /**
   * 自定义菜单项
   * 
   * 支持多级菜单（通过 children 属性）
   * 支持权限控制（通过组件内部根据 moduleId 自动检查）
   */
  customMenus?: ContextMenuItemConfig[]

  /**
   * 菜单点击回调
   * 
   * @param menu 菜单信息对象
   * @param menu.code 菜单项代码
   * @param menu.node 节点数据（Tree 组件使用）
   * @param menu.row 行数据（Grid 组件使用）
   * @param menu.column 列数据（Grid 组件使用，仅复制单元格时）
   */
  onMenuClick?: (menu: { 
    code: string
    node?: any
    row?: any
    column?: any
  }) => void
}

