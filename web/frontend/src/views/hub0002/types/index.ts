/**
 * 用户模块类型定义
 */

/**
 * 用户基本信息接口
 */
export interface User {
  userId: string // 用户ID，联合主键
  tenantId: string // 租户ID，联合主键
  userName: string // 用户名，登录账号
  password?: string // 密码，加密存储
  realName: string // 真实姓名
  deptId: string // 所属部门ID
  email?: string // 电子邮箱
  mobile?: string // 手机号码
  avatar?: string // 头像URL
  gender?: number // 性别：1-男，2-女，0-未知
  roles?: string[] // 用户角色列表
  statusFlag: string // 状态：Y-启用，N-禁用
  deptAdminFlag: string // 是否部门管理员：Y-是，N-否
  tenantAdminFlag: string // 是否租户管理员：Y-是，N-否
  userExpireDate: string // 用户过期时间，格式：yyyy-MM-dd HH:mm:ss
  lastLoginTime?: string // 最后登录时间，格式：yyyy-MM-dd HH:mm:ss
  lastLoginIp?: string // 最后登录IP
  addTime: string // 创建时间，格式：yyyy-MM-dd HH:mm:ss
  addWho: string // 创建人
  editTime: string // 修改时间，格式：yyyy-MM-dd HH:mm:ss
  editWho: string // 修改人
  oprSeqFlag: string // 操作序列标识
  currentVersion: number // 当前版本号
  activeFlag: string // 活动状态标记：Y-活动，N-非活动
  noteText?: string // 备注信息
}


/**
 * 用户状态枚举
 */
export enum UserStatus {
  ENABLED = 'Y', // 启用
  DISABLED = 'N', // 禁用
}

/**
 * 性别枚举
 */
export enum Gender {
  UNKNOWN = 0, // 未知
  MALE = 1, // 男
  FEMALE = 2, // 女
}

/**
 * 标志位枚举（Y/N）
 */
export enum FlagEnum {
  YES = 'Y', // 是
  NO = 'N', // 否
}
