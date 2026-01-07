/**
 * 过滤器配置列表类型定义
 * 统一管理业务类型，便于后续重构和维护
 */

// ============= 基础类型定义 =============

// 过滤器类型枚举 - 根据后端FilterType定义
export type FilterType =
  | 'header'
  | 'query-param'
  | 'body'
  | 'strip'
  | 'rewrite'
  | 'method'
  | 'cookie'
  | 'response'

// 过滤器执行时机枚举
export type FilterAction = 'pre-routing' | 'post-routing' | 'pre-response'

// Header修改类型
export type HeaderModifierType = 'add' | 'set' | 'remove' | 'rename'

// 查询参数修改类型
export type QueryParamModifierType = 'add' | 'set' | 'remove' | 'rename'

// 请求体修改类型
export type BodyModifierType = 'transform' | 'validate' | 'modify' | 'filter'

// Cookie操作类型
export type CookieOperation = 'add' | 'remove' | 'modify' | 'validate' | 'filter'

// 方法过滤模式
export type MethodFilterMode = 'allow' | 'deny'

// 路径重写模式
export type PathRewriteMode = 'simple' | 'regex'

// 响应操作类型
export type ResponseOperation =
  | 'add_headers'
  | 'modify_body'
  | 'set_status'
  | 'filter_headers'
  | 'transform_body'
  | 'validate_response'

// ============= 接口定义 =============

// 过滤器配置接口
export interface FilterConfig {
  tenantId: string
  filterConfigId: string
  gatewayInstanceId?: string
  routeConfigId?: string
  filterName: string
  filterType: FilterType
  filterAction: FilterAction
  filterOrder: number
  filterConfig: string // JSON格式的配置
  filterDesc?: string
  configId?: string
  activeFlag: 'Y' | 'N'
  addTime?: string
  addWho?: string
  editTime?: string
  editWho?: string
  noteText?: string
}

// Cookie属性接口
export interface CookieAttributes {
  domain?: string
  path?: string
  expires?: string
  maxAge?: number
  secure?: boolean
  httpOnly?: boolean
  sameSite?: 'Strict' | 'Lax' | 'None'
  custom?: Record<string, string>
}

// ============= 常量定义 =============

// 过滤器类型选项
export const FILTER_TYPE_OPTIONS = [
  {
    label: '请求头处理',
    value: 'header' as FilterType,
    description: '添加、修改或删除HTTP请求头/响应头',
  },
  { label: '查询参数处理', value: 'query-param' as FilterType, description: '处理URL查询参数' },
  { label: '请求体处理', value: 'body' as FilterType, description: '处理和验证请求体内容' },
  { label: '前缀剥离', value: 'strip' as FilterType, description: '移除请求路径中的指定前缀' },
  { label: '路径重写', value: 'rewrite' as FilterType, description: '重写或转换请求路径' },
  { label: 'HTTP方法控制', value: 'method' as FilterType, description: '控制允许的HTTP方法' },
  { label: 'Cookie处理', value: 'cookie' as FilterType, description: '处理HTTP Cookie' },
  { label: '响应处理', value: 'response' as FilterType, description: '处理后端响应' },
]

// 过滤器执行时机选项
export const FILTER_ACTION_OPTIONS = [
  { label: '前置处理', value: 'pre-routing' as FilterAction, description: '在路由匹配前执行' },
  { label: '后置处理', value: 'post-routing' as FilterAction, description: '在路由匹配后执行' },
  { label: '响应前处理', value: 'pre-response' as FilterAction, description: '在响应返回前执行' },
]

// 全局过滤器只支持前置处理
export const GLOBAL_FILTER_ACTION_OPTIONS = FILTER_ACTION_OPTIONS.filter(
  (option) => option.value === 'pre-routing',
)

// Header修改类型选项
export const HEADER_MODIFIER_OPTIONS = [
  { label: '添加', value: 'add' as HeaderModifierType, description: '添加新的请求头' },
  { label: '设置', value: 'set' as HeaderModifierType, description: '设置请求头（替换现有值）' },
  { label: '移除', value: 'remove' as HeaderModifierType, description: '移除指定请求头' },
  { label: '重命名', value: 'rename' as HeaderModifierType, description: '重命名请求头' },
]

// 查询参数修改类型选项
export const QUERY_PARAM_MODIFIER_OPTIONS = [
  { label: '添加', value: 'add' as QueryParamModifierType, description: '添加新的查询参数' },
  {
    label: '设置',
    value: 'set' as QueryParamModifierType,
    description: '设置查询参数（替换现有值）',
  },
  { label: '移除', value: 'remove' as QueryParamModifierType, description: '移除指定查询参数' },
  { label: '重命名', value: 'rename' as QueryParamModifierType, description: '重命名查询参数' },
]

// 请求体修改类型选项
export const BODY_MODIFIER_OPTIONS = [
  { label: '转换', value: 'transform' as BodyModifierType, description: '转换请求体格式' },
  { label: '验证', value: 'validate' as BodyModifierType, description: '验证请求体内容' },
  { label: '修改', value: 'modify' as BodyModifierType, description: '修改请求体字段' },
  { label: '过滤', value: 'filter' as BodyModifierType, description: '过滤请求体字段' },
]

// Cookie操作选项
export const COOKIE_OPERATION_OPTIONS = [
  { label: '添加', value: 'add' as CookieOperation, description: '添加新的Cookie' },
  { label: '移除', value: 'remove' as CookieOperation, description: '移除指定Cookie' },
  { label: '修改', value: 'modify' as CookieOperation, description: '修改Cookie值' },
  { label: '验证', value: 'validate' as CookieOperation, description: '验证Cookie格式' },
  { label: '过滤', value: 'filter' as CookieOperation, description: '过滤Cookie内容' },
]

// 方法过滤模式选项
export const METHOD_FILTER_MODE_OPTIONS = [
  { label: '允许模式', value: 'allow' as MethodFilterMode, description: '只允许指定的HTTP方法' },
  { label: '拒绝模式', value: 'deny' as MethodFilterMode, description: '拒绝指定的HTTP方法' },
]

// 路径重写模式选项
export const PATH_REWRITE_MODE_OPTIONS = [
  {
    label: '简单替换',
    value: 'simple' as PathRewriteMode,
    description: '直接替换路径中的特定部分',
  },
  {
    label: '正则替换',
    value: 'regex' as PathRewriteMode,
    description: '使用正则表达式匹配和替换路径',
  },
]

// 响应操作选项
export const RESPONSE_OPERATION_OPTIONS = [
  { label: '添加响应头', value: 'add_headers' as ResponseOperation, description: '添加HTTP响应头' },
  { label: '修改响应体', value: 'modify_body' as ResponseOperation, description: '修改响应体内容' },
  { label: '设置状态码', value: 'set_status' as ResponseOperation, description: '设置HTTP状态码' },
  {
    label: '过滤响应头',
    value: 'filter_headers' as ResponseOperation,
    description: '过滤响应头内容',
  },
  {
    label: '转换响应体',
    value: 'transform_body' as ResponseOperation,
    description: '转换响应体格式',
  },
  {
    label: '验证响应',
    value: 'validate_response' as ResponseOperation,
    description: '验证响应数据',
  },
]

// HTTP方法选项
export const HTTP_METHODS = ['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'HEAD', 'OPTIONS']

// 内容类型选项
export const CONTENT_TYPES = [
  'application/json',
  'application/xml',
  'text/plain',
  'text/html',
  'application/x-www-form-urlencoded',
  'multipart/form-data',
]

// ============= 组件 Props 和 Emits 类型 =============

/**
 * 过滤器配置列表模态框 Props
 */
export interface FilterConfigListModalProps {
  /** 是否显示模态框 */
  visible: boolean
  /** 模块ID（用于权限控制，必传） */
  moduleId: string
  /** 模态框标题 */
  title?: string
  /** 模态框宽度 */
  width?: number | string
  /** 挂载目标 */
  to?: string | HTMLElement | false
  /** 网关实例ID（用于全局过滤器） */
  gatewayInstanceId?: string
  /** 路由配置ID（用于路由过滤器） */
  routeConfigId?: string
  /** 过滤器类型（全局或路由） */
  filterScope?: 'global' | 'route'
}

/**
 * 过滤器配置列表模态框 Emits
 */
export interface FilterConfigListModalEmits {
  (e: 'update:visible', value: boolean): void
  (e: 'close'): void
  (e: 'refresh'): void
}
