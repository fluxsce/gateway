/**
 * 断言配置列表类型定义
 * 统一管理业务类型，便于后续重构和维护
 * 根据后端 models/route_config.go 中的 RouteAssertion 模型定义
 */

// ============= 枚举定义 =============

// 断言类型枚举（根据后端 RouteAssertion.AssertionType 字段）
export enum AssertionType {
  PATH = 'PATH',
  HEADER = 'HEADER',
  QUERY = 'QUERY',
  COOKIE = 'COOKIE',
  IP = 'IP',
  BODY_CONTENT = 'BODY_CONTENT',
}

// 断言操作符枚举（根据后端 RouteAssertion.AssertionOperator 字段）
export enum AssertionOperator {
  EQUAL = 'EQUAL',
  NOT_EQUAL = 'NOT_EQUAL',
  CONTAINS = 'CONTAINS',
  NOT_CONTAINS = 'NOT_CONTAINS',
  MATCHES = 'MATCHES',
  NOT_MATCHES = 'NOT_MATCHES',
  STARTS_WITH = 'STARTS_WITH',
  ENDS_WITH = 'ENDS_WITH',
  IN = 'IN',
  NOT_IN = 'NOT_IN',
}

// ============= 接口定义 =============

/**
 * 路由断言配置接口
 * 根据后端 models/route_config.go 中的 RouteAssertion 结构定义
 */
export interface RouteAssertion {
  tenantId: string // 租户ID，联合主键
  routeAssertionId: string // 路由断言ID，联合主键
  routeConfigId: string // 关联的路由配置ID
  assertionName: string // 断言名称
  assertionType: AssertionType | string // 断言类型(PATH,HEADER,QUERY,COOKIE,IP,BODY_CONTENT)
  assertionOperator: AssertionOperator | string // 断言操作符(EQUAL,NOT_EQUAL,CONTAINS,MATCHES等)
  fieldName: string // 字段名称(HEADER/QUERY/COOKIE类型时使用，可能为空字符串)
  expectedValue: string // 期望值(EQUAL/NOT_EQUAL等操作符时使用，可能为空字符串)
  patternValue: string // 匹配模式(MATCHES/NOT_MATCHES操作符时使用,支持正则表达式，可能为空字符串)
  caseSensitive: 'Y' | 'N' // 是否区分大小写(N否,Y是)
  assertionOrder: number // 断言执行顺序(数值越小越先执行)
  isRequired: 'Y' | 'N' // 是否必须匹配(N否,Y是)
  assertionDesc: string // 断言描述（可能为空字符串）
  reserved1: string // 预留字段1（可能为空字符串）
  reserved2: string // 预留字段2（可能为空字符串）
  reserved3?: number | null // 预留字段3（指针类型，可为nil）
  reserved4?: number | null // 预留字段4（指针类型，可为nil）
  reserved5?: string | null // 预留字段5（指针类型，可为nil）
  extProperty: string // 扩展属性,JSON格式（可能为空字符串）
  addTime: string // 创建时间
  addWho: string // 创建人ID
  editTime: string // 最后修改时间
  editWho: string // 最后修改人ID
  oprSeqFlag: string // 操作序列标识
  currentVersion: number // 当前版本号
  activeFlag: 'Y' | 'N' // 活动状态标记(N非活动,Y活动)
  noteText: string // 备注信息（可能为空字符串）
}


// ============= 类型别名 =============

// 断言配置接口（别名，用于组件内部）
export type AssertConfig = RouteAssertion


// ============= 常量定义 =============

// 断言类型选项
export const ASSERTION_TYPE_OPTIONS = [
  { label: '路径', value: 'PATH' as AssertionType, description: '路径匹配断言，用于复杂路径匹配规则' },
  { label: '请求头', value: 'HEADER' as AssertionType, description: '请求头断言，检查特定请求头的值' },
  { label: '查询参数', value: 'QUERY' as AssertionType, description: '查询参数断言，检查URL查询参数' },
  { label: 'Cookie', value: 'COOKIE' as AssertionType, description: 'Cookie断言，检查特定Cookie的值' },
  { label: 'IP地址', value: 'IP' as AssertionType, description: 'IP地址断言，基于客户端IP进行匹配' },
  { label: '请求体内容', value: 'BODY_CONTENT' as AssertionType, description: '请求体内容断言，检查HTTP请求体的内容' },
]

// 断言操作符选项
export const ASSERTION_OPERATOR_OPTIONS = [
  { label: '等于', value: 'EQUAL' as AssertionOperator, description: '值必须完全相等' },
  { label: '不等于', value: 'NOT_EQUAL' as AssertionOperator, description: '值必须不相等' },
  { label: '包含', value: 'CONTAINS' as AssertionOperator, description: '值必须包含指定字符串' },
  { label: '不包含', value: 'NOT_CONTAINS' as AssertionOperator, description: '值必须不包含指定字符串' },
  { label: '正则匹配', value: 'MATCHES' as AssertionOperator, description: '值必须匹配正则表达式' },
  { label: '正则不匹配', value: 'NOT_MATCHES' as AssertionOperator, description: '值必须不匹配正则表达式' },
  { label: '开头匹配', value: 'STARTS_WITH' as AssertionOperator, description: '值必须以指定字符串开头' },
  { label: '结尾匹配', value: 'ENDS_WITH' as AssertionOperator, description: '值必须以指定字符串结尾' },
  { label: '在列表中', value: 'IN' as AssertionOperator, description: '值必须在指定列表中' },
  { label: '不在列表中', value: 'NOT_IN' as AssertionOperator, description: '值必须不在指定列表中' },
]

// ============= 组件 Props 和 Emits 类型 =============

/**
 * 断言配置列表模态框 Props
 */
export interface AssertConfigListModalProps {
  /** 是否显示模态框 */
  visible: boolean
  /** 模态框标题 */
  title?: string
  /** 模态框宽度 */
  width?: number | string
  /** 挂载目标 */
  to?: string | HTMLElement | false
  /** 路由配置ID（必填） */
  routeConfigId: string
}

/**
 * 断言配置列表模态框 Emits
 */
export interface AssertConfigListModalEmits {
  (e: 'update:visible', value: boolean): void
  (e: 'close'): void
  (e: 'refresh'): void
}

