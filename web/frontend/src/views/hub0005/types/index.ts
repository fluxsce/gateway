/**
 * 角色管理模块类型定义
 */

/**
 * 角色基本信息接口
 */
export interface Role {
  roleId: string // 角色ID，主键
  tenantId: string // 租户ID，用于多租户数据隔离
  roleName: string // 角色名称
  roleDescription?: string // 角色描述
  roleStatus: string // 角色状态(Y:启用,N:禁用)
  builtInFlag: string // 内置角色标记(Y:内置,N:自定义)
  dataScope?: string // 数据权限范围，TEXT类型，可存储复杂的权限配置
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
}

/**
 * 角色状态枚举
 */
export enum RoleStatus {
  ENABLED = 'Y', // 启用
  DISABLED = 'N', // 禁用
}

/**
 * 标志位枚举（Y/N）
 */
export enum FlagEnum {
  YES = 'Y', // 是
  NO = 'N', // 否
}

/**
 * 内置角色标记枚举
 */
export enum BuiltInFlag {
  BUILT_IN = 'Y', // 内置
  CUSTOM = 'N', // 自定义
}

