/**
 * GRestfulApi 静态下拉选项与无状态工具函数，便于单测与主组件瘦身。
 *
 * @remarks
 * 与 `GRestfulApi.vue` 解耦；不含 Vue 响应式状态。
 */

import type { CodeMirrorLanguage } from '@/components/gcodemirror/types'
import type { RestBodyProcessType, RestHttpMethod, RestKeyValueRow } from './types'
import { createKeyValueRow } from './types'

/**
 * Body 处理类型下拉数据（顺序与 Postman 等工具相近）。
 *
 * @see {@link RestBodyProcessType}
 */
export const BODY_PROCESS_OPTIONS: { value: RestBodyProcessType; label: string }[] = [
  { value: 'none', label: 'none' },
  { value: 'form-data', label: 'form-data' },
  { value: 'x-www-form-urlencoded', label: 'x-www-form-urlencoded' },
  { value: 'json', label: 'json' },
  { value: 'xml', label: 'xml' },
  { value: 'raw', label: 'raw' },
  { value: 'binary', label: 'binary' },
  { value: 'graphql', label: 'GraphQL' },
  { value: 'msgpack', label: 'msgpack' }
]

/**
 * HTTP 方法下拉数据。
 *
 * @see {@link RestHttpMethod}
 */
export const METHOD_OPTIONS: { label: string; value: RestHttpMethod }[] = [
  { label: 'GET', value: 'GET' },
  { label: 'POST', value: 'POST' },
  { label: 'PUT', value: 'PUT' },
  { label: 'PATCH', value: 'PATCH' },
  { label: 'DELETE', value: 'DELETE' },
  { label: 'HEAD', value: 'HEAD' },
  { label: 'OPTIONS', value: 'OPTIONS' }
]

/**
 * Raw 模式下 Content-Type 下拉候选项。
 */
export const RAW_CONTENT_TYPE_OPTIONS: { label: string; value: string }[] = [
  { label: 'application/json', value: 'application/json' },
  { label: 'text/plain', value: 'text/plain' },
  { label: 'application/xml', value: 'application/xml' },
  { label: 'text/html', value: 'text/html' },
  { label: 'application/octet-stream', value: 'application/octet-stream' },
  { label: 'application/msgpack', value: 'application/msgpack' }
]

/**
 * 方法选择器左侧色条对应的 BEM 修饰类名（如 `g-restful-api__method--get`）。
 *
 * @param method - 当前 HTTP 方法
 * @returns 用于绑定到组件 `class` 的字符串
 */
export function restfulMethodModifierClass(method: RestHttpMethod): string {
  return `g-restful-api__method--${method.toLowerCase()}`
}

/**
 * 由 Body 处理类型推导默认请求体 `Content-Type`（不含用户在 Headers 中覆盖）。
 *
 * @param processType - Body 处理类型
 * @param rawContentType - `processType === 'raw'` 时选中的 Content-Type
 * @returns 推导出的 MIME 描述字符串；类型为 `none` 时返回 `null`
 */
export function deriveBodyContentType(
  processType: RestBodyProcessType,
  rawContentType: string
): string | null {
  switch (processType) {
    case 'none':
      return null
    case 'form-data':
      return 'multipart/form-data'
    case 'x-www-form-urlencoded':
      return 'application/x-www-form-urlencoded;charset=UTF-8'
    case 'json':
      return 'application/json'
    case 'xml':
      return 'application/xml'
    case 'raw':
      return rawContentType
    case 'binary':
      return 'application/octet-stream'
    case 'graphql':
      return 'application/json'
    case 'msgpack':
      return 'application/msgpack'
    default:
      return null
  }
}

/**
 * 推断 Raw / GraphQL 请求体在编辑器中使用的语法高亮语言。
 *
 * @param bodyProcessType - Body 处理类型
 * @param rawContentType - Raw 模式下的 Content-Type 字符串
 * @returns CodeMirror 语言标识
 */
export function inferRawEditorLanguage(
  bodyProcessType: RestBodyProcessType,
  rawContentType: string
): CodeMirrorLanguage {
  if (bodyProcessType === 'graphql') {
    return 'json'
  }
  const ct = rawContentType.toLowerCase()
  if (ct.includes('json')) {
    return 'json'
  }
  if (ct.includes('html')) {
    return 'html'
  }
  if (ct.includes('xml')) {
    return 'xml'
  }
  return 'plaintext'
}

/**
 * 根据 HTTP 状态码选择 Naive UI `n-tag` 的 `type`（不含 `default`，无结果时由调用方处理）。
 *
 * @param status - HTTP 状态码
 * @returns `success`（2xx）、`error`（4xx/5xx 等 ≥400）、`warning`（其余）
 */
export function httpStatusTagType(
  status: number
): 'default' | 'success' | 'error' | 'warning' {
  if (status >= 200 && status < 300) {
    return 'success'
  }
  if (status >= 400) {
    return 'error'
  }
  return 'warning'
}

/**
 * 将 Params 表格行序列化为 query 字符串（编码规则与 `URLSearchParams` 一致）。
 *
 * @param rows - Query 键值行
 * @returns 无 `?` 前缀的查询串，可能为空
 */
export function queryRowsToSearchString(rows: RestKeyValueRow[]): string {
  const sp = new URLSearchParams()
  for (const r of rows) {
    if (!r.enabled) {
      continue
    }
    const k = r.key.trim()
    if (!k) {
      continue
    }
    sp.append(k, r.value)
  }
  return sp.toString()
}

/**
 * 将 `URLSearchParams` 转为 Params 表格行（末尾追加一行空行供编辑）。
 *
 * @param sp - 来自 `URL` 的查询参数
 * @returns 新行对象列表
 */
export function searchParamsToQueryRows(sp: URLSearchParams): RestKeyValueRow[] {
  const rows: RestKeyValueRow[] = []
  for (const [key, value] of sp) {
    rows.push({
      ...createKeyValueRow(),
      enabled: true,
      key,
      value
    })
  }
  rows.push(createKeyValueRow())
  return rows
}

/**
 * 从请求头表格读取指定头的值（忽略大小写）；同一头既有 `autoFromBody` 又有手写行时优先返回手写行。
 *
 * @param rows - Header 行列表
 * @param headerName - 头名称（如 `content-type`）
 * @returns 第一个匹配且启用的值；无则空字符串
 */
export function findHeaderValueCI(rows: RestKeyValueRow[], headerName: string): string {
  const n = headerName.toLowerCase()
  let autoVal = ''
  for (const r of rows) {
    if (!r.enabled || r.key.trim().toLowerCase() !== n) {
      continue
    }
    if (!r.autoFromBody) {
      return r.value
    }
    autoVal = r.value
  }
  return autoVal
}

/**
 * 浅拷贝键值行数组（合并 Cookie / Auth 发往前避免改写字段引用同一对象）。
 *
 * @param rows - 原始行
 * @returns 新数组，元素为浅拷贝（`file` 引用保留）
 */
export function cloneKeyValueRows(rows: RestKeyValueRow[]): RestKeyValueRow[] {
  return rows.map((r) => ({ ...r, file: r.file }))
}

/**
 * 按名称合并或追加一行请求头（忽略大小写匹配已有启用行则覆盖 `value`）。
 *
 * @param rows - 被原地修改的行列表
 * @param key - 头名
 * @param value - 头值
 */
export function upsertHeaderRow(rows: RestKeyValueRow[], key: string, value: string): void {
  const lower = key.toLowerCase()
  const idx = rows.findIndex((r) => r.enabled && r.key.trim().toLowerCase() === lower)
  if (idx >= 0) {
    rows[idx].value = value
    return
  }
  rows.push({
    ...createKeyValueRow(),
    enabled: true,
    key,
    value
  })
}

/**
 * 将 Cookies 表格行拼接为 `Cookie` 请求头值（`k=v` 以 `; ` 连接）。
 *
 * @param rows - Cookie 行
 * @returns 可直接用于 `Cookie` 头的字符串
 */
export function cookieStringFromRows(rows: RestKeyValueRow[]): string {
  const parts: string[] = []
  for (const r of rows) {
    if (!r.enabled) {
      continue
    }
    const k = r.key.trim()
    if (!k) {
      continue
    }
    parts.push(`${k}=${r.value}`)
  }
  return parts.join('; ')
}

/**
 * 判断两组请求头行是否逐行完全一致（含 `id`、启用状态与 `autoFromBody`）。
 *
 * @param a - 行列表甲
 * @param b - 行列表乙
 */
export function sameHeaderRowsForSync(a: RestKeyValueRow[], b: RestKeyValueRow[]): boolean {
  if (a.length !== b.length) {
    return false
  }
  for (let i = 0; i < a.length; i++) {
    const x = a[i]
    const y = b[i]
    if (
      x.id !== y.id ||
      x.enabled !== y.enabled ||
      x.key !== y.key ||
      x.value !== y.value ||
      Boolean(x.autoFromBody) !== Boolean(y.autoFromBody)
    ) {
      return false
    }
  }
  return true
}

/** {@link demoteAutoContentTypeRows} 的可选行为 */
export interface DemoteAutoContentTypeOptions {
  /**
   * 为 true 时：自动 Content-Type 行的值与当前 Body 推导不一致则降级为手写（表示用户改过头）。
   * 为 false 时：不因值不一致降级（用于 Body 类型切换后由 computeDesired 写回新值）。
   * @default true
   */
  demoteOnValueMismatch?: boolean
}

/**
 * 将不符合「仍随 Body 同步」条件的自动 `Content-Type` 行降级为普通行（清除 `autoFromBody`）。
 *
 * @param rows - 当前 Header 行
 * @param derived - 当前 Body 推导的 Content-Type；为 `null` 时表示无推导
 * @param options - 是否因「值与推导不一致」降级（见 {@link DemoteAutoContentTypeOptions}）
 * @returns 有变更时返回新数组，否则 `null`
 */
export function demoteAutoContentTypeRows(
  rows: RestKeyValueRow[],
  derived: string | null,
  options?: DemoteAutoContentTypeOptions
): RestKeyValueRow[] | null {
  const demoteOnValueMismatch = options?.demoteOnValueMismatch !== false
  let changed = false
  const next = rows.map((r) => {
    if (!r.autoFromBody) {
      return r
    }
    const k = r.key.trim().toLowerCase()
    if (k !== 'content-type') {
      changed = true
      return { ...r, autoFromBody: false }
    }
    if (!r.enabled) {
      changed = true
      return { ...r, autoFromBody: false }
    }
    if (
      demoteOnValueMismatch &&
      derived !== null &&
      r.value.trim() !== derived
    ) {
      changed = true
      return { ...r, autoFromBody: false }
    }
    return r
  })
  return changed ? next : null
}

/**
 * 根据 Body 推导的 Content-Type 计算 Header 表格下一状态（自动行与手写 `Content-Type` 互斥）。
 *
 * @param current - 当前行
 * @param derived - Body 推导的 Content-Type
 * @returns 需要替换整表时返回新行数组；无需变更时 `null`
 */
export function computeDesiredHeaderRows(
  current: RestKeyValueRow[],
  derived: string | null
): RestKeyValueRow[] | null {
  const hasManualCt = current.some(
    (r) => r.key.trim().toLowerCase() === 'content-type' && !r.autoFromBody
  )
  const withoutAuto = current.filter((r) => !r.autoFromBody)

  if (hasManualCt) {
    if (!current.some((r) => r.autoFromBody)) {
      return null
    }
    return sameHeaderRowsForSync(current, withoutAuto) ? null : withoutAuto
  }

  if (derived === null) {
    if (!current.some((r) => r.autoFromBody)) {
      return null
    }
    return sameHeaderRowsForSync(current, withoutAuto) ? null : withoutAuto
  }

  const existingAuto = current.find((r) => r.autoFromBody)
  const autoRow: RestKeyValueRow = {
    ...createKeyValueRow(),
    id: existingAuto?.id ?? crypto.randomUUID(),
    key: 'Content-Type',
    value: derived,
    enabled: true,
    autoFromBody: true
  }
  const next = [autoRow, ...withoutAuto]
  return sameHeaderRowsForSync(current, next) ? null : next
}

/**
 * 将响应头对象序列化为多行文本（键排序后 `key: value`）。
 *
 * @param h - 响应头映射
 */
export function stringifyResponseHeaders(h: Record<string, string>): string {
  const keys = Object.keys(h).sort((a, b) => a.localeCompare(b))
  return keys.map((k) => `${k}: ${h[k]}`).join('\n')
}
