/**
 * API相关类型定义
 * 包括请求参数和响应数据的类型接口
 */

/**
 * API响应数据通用结构
 */
export interface ApiResponse<T = any> {
  code: number // 状态码，0表示成功
  message: string // 响应消息
  data: T // 响应数据
}

/**
 * 后端返回的分页信息对象结构
 *
 * @property baseData - 基础数据，通常用于存储额外信息
 * @property curPageCount - 当前页面记录数量
 * @property dbsId - 数据库服务ID
 * @property mainKey - 主键字段名
 * @property orderByList - 排序规则列表，通常为字符串形式的SQL排序表达式
 * @property otherData - 额外数据，存储非标准信息
 * @property pageIndex - 当前页码，从1开始
 * @property pageSize - 每页记录数
 * @property paramObjectsJson - 查询参数对象JSON字符串
 * @property timeTypeFieldNames - 时间类型字段名列表
 * @property totalCount - 记录总数
 * @property totalPageIndex - 总页数
 */
export interface PageInfoObj {
  /** 基础数据，通常用于存储额外信息 */
  baseData: string
  /** 当前页面记录数量 */
  curPageCount: number
  /** 数据库服务ID */
  dbsId: string
  /** 主键字段名 */
  mainKey: string
  /** 排序规则列表，通常为字符串形式的SQL排序表达式 */
  orderByList: string
  /** 额外数据，存储非标准信息 */
  otherData: string
  /** 当前页码，从1开始 */
  pageIndex: number
  /** 每页记录数 */
  pageSize: number
  /** 查询参数对象JSON字符串 */
  paramObjectsJson: string
  /** 时间类型字段名列表 */
  timeTypeFieldNames: string
  /** 记录总数 */
  totalCount: number
  /** 总页数 */
  totalPageIndex: number
}

/**
 * 定义了与后端交互的标准响应格式
 *
 * @property oK - 操作是否成功（注意首字母o为小写，K为大写，Java命名习惯）
 * @property state - 状态标识，表示操作状态
 * @property bizData - 业务数据，通常是序列化后的JSON字符串
 * @property extObj - 扩展对象，可以是任意类型
 * @property pageQueryData - 分页查询数据，通常包含分页相关信息
 * @property messageId - 消息标识，用于追踪和日志
 * @property errMsg - 错误消息，当操作失败时提供错误详情
 * @property popMsg - 弹出消息，可用于前端展示
 * @property extMsg - 扩展消息，提供额外信息
 * @property pkey1 - 预留主键字段1，用于特定业务场景
 * @property pkey2 - 预留主键字段2，用于特定业务场景
 * @property pkey3 - 预留主键字段3，用于特定业务场景
 * @property pkey4 - 预留主键字段4，用于特定业务场景
 * @property pkey5 - 预留主键字段5，用于特定业务场景
 * @property pkey6 - 预留主键字段6，用于特定业务场景
 */
export interface JsonDataObj {
  /** 操作是否成功（注意首字母o为小写，K为大写，Java命名习惯） */
  oK: boolean
  /** 状态标识，表示操作状态 */
  state: boolean
  /** 业务数据，通常是序列化后的JSON字符串 */
  bizData: string
  /** 扩展对象，可以是任意类型 */
  extObj: any
  /** 分页查询数据，通常包含分页相关信息 */
  pageQueryData: string
  /** 消息标识，用于追踪和日志 */
  messageId: string
  /** 错误消息，当操作失败时提供错误详情 */
  errMsg: string
  /** 弹出消息，可用于前端展示 */
  popMsg: string
  /** 扩展消息，提供额外信息 */
  extMsg: string
  /** 预留主键字段1，用于特定业务场景 */
  pkey1: string
  /** 预留主键字段2，用于特定业务场景 */
  pkey2: string
  /** 预留主键字段3，用于特定业务场景 */
  pkey3: string
  /** 预留主键字段4，用于特定业务场景 */
  pkey4: string
  /** 预留主键字段5，用于特定业务场景 */
  pkey5: string
  /** 预留主键字段6，用于特定业务场景 */
  pkey6: string
}

/**
 * 菜单项接口定义
 * 对应数据库中的HUB_MENU表
 */
export interface MenuItem {
  /** 菜单ID，主键之一 */
  menuId: string
  /** 菜单名称 */
  menuName: string
  /** 父菜单ID */
  parentId?: string
  /** 菜单路径，如：/system/user */
  menuPath: string
  /** 前端组件路径 */
  component?: string
  /** 菜单图标 */
  icon?: string
  /** 排序 */
  sortOrder: number
  /** 菜单类型：1-目录，2-菜单，3-按钮/权限点 */
  menuType: number
  /** 权限标识，如：system:user:add */
  permCode?: string
  /** 是否可见：Y-可见，N-隐藏 */
  visibleFlag: string
  /** 状态：Y-启用，N-禁用 */
  statusFlag: string
  /** 是否系统菜单：Y-是，N-否 */
  sysMenuFlag: string
  /** 创建时间 */
  addTime: string
  /** 创建人 */
  addWho: string
  /** 修改时间 */
  editTime: string
  /** 修改人 */
  editWho: string
  /** 操作序列标识 */
  oprSeqFlag: string
  /** 当前版本号 */
  currentVersion: number
  /** 活动状态标记：Y-活动，N-非活动 */
  activeFlag: string
  /** 语言，主键之一 */
  languageId: string
  /** 备注信息 */
  noteText?: string
  /** 子菜单列表，用于构建菜单树 */
  children?: MenuItem[]
  /** 权限项列表，非目录结构应该存在，用于构建权限树 */
  permissions?: PermissionItem[]
}

/**
 * 权限项接口定义
 * 对应数据库中的HUB_PERMISSION表
 */
export interface PermissionItem {
  /** 权限标识，如：system:user:add，主键之一 */
  permCode: string
  /** 权限名称 */
  permName: string
  /** 所属菜单ID，主键之一 */
  menuId: string
  /** 权限类型：1-菜单，2-按钮，3-数据 */
  permType: number
  /** 状态：Y-启用，N-禁用 */
  statusFlag: string
  /** 创建时间 */
  addTime: string
  /** 创建人 */
  addWho: string
  /** 修改时间 */
  editTime: string
  /** 修改人 */
  editWho: string
  /** 操作序列标识 */
  oprSeqFlag: string
  /** 当前版本号 */
  currentVersion: number
  /** 活动状态标记：Y-活动，N-非活动 */
  activeFlag: string
  /** 备注信息 */
  noteText?: string
}

