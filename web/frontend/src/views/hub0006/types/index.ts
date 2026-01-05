/**
 * 权限资源管理模块类型定义
 */

/**
 * 权限资源基本信息接口
 */
export interface Resource {
  resourceId: string // 资源ID，主键
  tenantId: string // 租户ID，用于多租户数据隔离
  resourceName: string // 资源名称
  resourceCode: string // 资源编码，用于程序判断
  resourceType: string // 资源类型(MODULE:模块,MENU:菜单,BUTTON:按钮,API:接口)
  resourcePath?: string // 资源路径(菜单路径或API路径)
  resourceMethod?: string // 请求方法(GET,POST,PUT,DELETE等)
  parentResourceId?: string // 父资源ID
  resourceLevel: number // 资源层级
  sortOrder: number // 排序顺序
  displayName?: string // 显示名称
  iconClass?: string // 图标样式类
  description?: string // 资源描述
  language?: string // 语言标识（如：zh-CN, en-US），用于多语言支持
  resourceStatus: string // 资源状态(Y:启用,N:禁用)
  builtInFlag: string // 内置资源标记(Y:内置,N:自定义)
  addTime: string // 创建时间，格式：yyyy-MM-dd HH:mm:ss
  addWho: string // 创建人ID
  editTime: string // 最后修改时间，格式：yyyy-MM-dd HH:mm:ss
  editWho: string // 最后修改人ID
  oprSeqFlag: string // 操作序列标识
  currentVersion: number // 当前版本号
  activeFlag: string // 活动状态标记(N非活动,Y活动)
  noteText?: string // 备注信息
  extProperty?: string // 扩展属性，JSON格式
  reserved1?: string // 预留字段1
  reserved2?: string // 预留字段2
  reserved3?: string // 预留字段3
  reserved4?: string // 预留字段4
  reserved5?: string // 预留字段5
  reserved6?: string // 预留字段6
  reserved7?: string // 预留字段7
  reserved8?: string // 预留字段8
  reserved9?: string // 预留字段9
  reserved10?: string // 预留字段10
  children?: Resource[] // 子资源列表（用于树形展示）
}

/**
 * 资源状态枚举
 */
export enum ResourceStatus {
  ENABLED = 'Y', // 启用
  DISABLED = 'N', // 禁用
}

/**
 * 资源类型枚举
 */
export enum ResourceType {
  MODULE = 'MODULE', // 模块
  MENU = 'MENU', // 菜单
  BUTTON = 'BUTTON', // 按钮
  API = 'API', // 接口
}

/**
 * 标志位枚举（Y/N）
 */
export enum FlagEnum {
  YES = 'Y', // 是
  NO = 'N', // 否
}

/**
 * 内置资源标记枚举
 */
export enum BuiltInFlag {
  BUILT_IN = 'Y', // 内置
  CUSTOM = 'N', // 自定义
}

