/**
 * 格式化工具类
 * 提供各种数据格式化函数，用于将数据转换为适合显示的格式
 */

import type { JsonDataObj, PageInfoObj } from '@/types/api'
import { createLogger } from '@/utils/logger'

// 创建格式化工具专用的日志记录器
const logger = createLogger('FormatUtils')

/**
 * 将 moduleId 转换为符合 HTML id 规范的格式
 * 将冒号（:）替换为连字符（-），确保 id 符合 HTML 规范
 * 注意：此函数仅用于生成 HTML id，权限校验仍应使用原始 moduleId
 * 
 * @param moduleId 模块ID（可能包含冒号，如 "hub0061:static-nodes"）
 * @returns 符合 HTML id 规范的字符串（如 "hub0061-static-nodes"）
 *
 * @example
 * const moduleId = 'hub0061:static-nodes'
 * const htmlId = toHtmlId(moduleId) // 'hub0061-static-nodes'
 * // 在模板中使用：
 * // <div :id="toHtmlId(service.model.moduleId)">
 * // 权限校验仍使用原始 moduleId：
 * // :module-id="service.model.moduleId"
 */
export const toHtmlId = (moduleId: string): string => {
  if (!moduleId) {
    return ''
  }
  // 将冒号替换为连字符，确保符合 HTML id 规范
  return moduleId.replace(/:/g, '-')
}

/**
 * 解析API返回的JsonDataObj数据
 * @param jsonData API返回的JsonDataObj数据
 * @param defaultValue 解析失败时的默认值
 * @returns 解析后的数据
 *
 * @example
 * const users = parseJsonData<User[]>(result, [])
 * const user = parseJsonData<User>(result, {})
 * const config = parseJsonData<Config | null>(result, null) // 不存在时返回 null
 */
export const parseJsonData = <T>(jsonData: JsonDataObj, defaultValue?: T): T => {
  if (jsonData.bizData) {
    try {
      return JSON.parse(jsonData.bizData) as T
    } catch (error) {
      logger.error('解析JsonDataObj失败', error)
      throw new Error('解析数据失败')
    }
  }
  // 如果 bizData 不存在，返回 defaultValue 或 null
  return (defaultValue ?? null) as T
}

/**
 * 解析API返回的分页信息
 * @param jsonData API返回的JsonDataObj数据
 * @returns 解析后的分页信息
 *
 * @example
 * const pageInfo = parsePageInfo(result)
 */
export const parsePageInfo = (jsonData: JsonDataObj): PageInfoObj => {
  if (jsonData.pageQueryData) {
    try {
      return JSON.parse(jsonData.pageQueryData) as PageInfoObj
    } catch (error) {
      logger.error('解析分页信息失败', error)
      throw new Error('解析分页信息失败')
    }
  }
  return {} as PageInfoObj
}

/**
 * 安全地解析API返回的JsonDataObj数据（不抛出异常）
 * @param jsonData API返回的JsonDataObj数据
 * @param defaultValue 解析失败时的默认值
 * @returns 解析后的数据或默认值
 *
 * @example
 * const users = safeParseJsonData<User[]>(result, [])
 * const user = safeParseJsonData<User>(result, null)
 */
export const safeParseJsonData = <T>(jsonData: JsonDataObj, defaultValue: T): T => {
  try {
    return parseJsonData<T>(jsonData, defaultValue)
  } catch (error) {
    logger.warn('安全解析JsonDataObj失败，返回默认值', error)
    return defaultValue
  }
}

/**
 * 检查API操作是否成功
 * @param jsonData API返回的JsonDataObj数据
 * @returns 是否成功
 *
 * @example
 * if (isApiSuccess(result)) {
 *   // 处理成功逻辑
 * }
 */
export const isApiSuccess = (jsonData: JsonDataObj): boolean => {
  return jsonData.oK
}

/**
 * 获取API返回的消息
 * @param jsonData API返回的JsonDataObj数据
 * @param defaultMessage 默认消息
 * @returns 消息文本
 *
 * @example
 * const message = getApiMessage(result, '操作完成')
 */
export const getApiMessage = (jsonData: JsonDataObj, defaultMessage = ''): string => {
  return jsonData.popMsg || jsonData.errMsg || defaultMessage
}

/**
 * 格式化日期时间
 * 将Date对象、时间戳或日期字符串转换为指定格式的日期字符串
 *
 * 重要：后端常返回带 `Z` / 偏移的 ISO（如 `2026-08-13T10:20:41Z`）。
 * 若用 `new Date` + `getHours()` 会按浏览器本地时区换算，表格展示与表单回传都会偏几个小时。
 * 对可识别的 ISO 字面量，直接取字符串中的年月日时分秒（墙钟），不做时区换算。
 *
 * @param date 日期对象、时间戳或日期字符串
 * @param format 格式化模板，例如: 'YYYY-MM-DD HH:mm:ss'
 *               - YYYY: 四位年份
 *               - MM: 两位月份
 *               - DD: 两位日期
 *               - HH: 两位小时(24小时制)
 *               - mm: 两位分钟
 *               - ss: 两位秒钟
 *               - SSS: 三位毫秒
 * @returns 格式化后的日期字符串
 *
 * @example
 * formatDate(new Date(), 'YYYY-MM-DD') // 2023-10-18
 * formatDate(1697609463000, 'HH:mm:ss') // 15:24:23
 * formatDate('2026-08-13T10:20:41Z', 'YYYY-MM-DD HH:mm:ss') // 2026-08-13 10:20:41（不转本地）
 */
/** ISO / 类 ISO 字符串的墙钟分量（忽略 Z、±偏移带来的时区换算） */
export type DateWallClock = {
  year: string
  month: string
  day: string
  hour: string
  minute: string
  second: string
  millisecond: string
}

const ISO_WALL_CLOCK_RE =
  /^(\d{4})-(\d{2})-(\d{2})(?:[T ](\d{2}):(\d{2})(?::(\d{2}))?(?:\.(\d{1,9}))?)?(?:Z|[+-]\d{2}:?\d{2})?$/i

/**
 * 从 ISO 字符串解析墙钟分量；无法识别则返回 null。
 */
export function parseIsoWallClock(value: string): DateWallClock | null {
  const matched = value.trim().match(ISO_WALL_CLOCK_RE)
  if (!matched) return null
  return {
    year: matched[1]!,
    month: matched[2]!,
    day: matched[3]!,
    hour: matched[4] ?? '00',
    minute: matched[5] ?? '00',
    second: (matched[6] ?? '00').padStart(2, '0'),
    millisecond: ((matched[7] ?? '0') + '000').slice(0, 3),
  }
}

/** 将墙钟分量套到 format 模板 */
export function formatWallClock(
  parts: DateWallClock,
  format = 'YYYY-MM-DD HH:mm:ss',
): string {
  return format
    .replace(/YYYY/g, parts.year)
    .replace(/MM/g, parts.month)
    .replace(/DD/g, parts.day)
    .replace(/HH/g, parts.hour)
    .replace(/mm/g, parts.minute)
    .replace(/ss/g, parts.second)
    .replace(/SSS/g, parts.millisecond)
}

export const formatDate = (
  date: Date | number | string,
  format = 'YYYY-MM-DD HH:mm:ss',
): string => {
  if (typeof date === 'string') {
    const wall = parseIsoWallClock(date)
    if (wall) {
      return formatWallClock(wall, format)
    }
  }

  const d = new Date(date)
  if (Number.isNaN(d.getTime())) {
    return typeof date === 'string' ? date : ''
  }

  const year = d.getFullYear().toString()
  const month = (d.getMonth() + 1).toString().padStart(2, '0')
  const day = d.getDate().toString().padStart(2, '0')
  const hour = d.getHours().toString().padStart(2, '0')
  const minute = d.getMinutes().toString().padStart(2, '0')
  const second = d.getSeconds().toString().padStart(2, '0')
  const millisecond = d.getMilliseconds().toString().padStart(3, '0')

  return format
    .replace(/YYYY/g, year)
    .replace(/MM/g, month)
    .replace(/DD/g, day)
    .replace(/HH/g, hour)
    .replace(/mm/g, minute)
    .replace(/ss/g, second)
    .replace(/SSS/g, millisecond)
}

/**
 * 格式化金额
 * 将数字转换为千分位格式，并保留指定小数位
 *
 * @param amount 金额数值
 * @param decimals 小数位数，默认为2
 * @param separator 千位分隔符，默认为','
 * @returns 格式化后的金额字符串
 *
 * @example
 * formatAmount(1234567.89) // '1,234,567.89'
 * formatAmount(1234.5, 0) // '1,235'
 */
export const formatAmount = (amount: number, decimals = 2, separator = ','): string => {
  return amount.toFixed(decimals).replace(/\B(?=(\d{3})+(?!\d))/g, separator)
}

/**
 * 文件大小格式化
 * 将字节数转换为更易读的格式(KB, MB, GB等)
 *
 * @param bytes 字节数（可为字符串；null/undefined/空串/非有限数/负数时返回 '-'，避免表格出现 NaN undefined）
 * @returns 格式化后的文件大小，自动选择合适的单位
 *
 * @example
 * formatFileSize(1024) // '1.00 KB'
 * formatFileSize(1048576) // '1.00 MB'
 */
export const formatFileSize = (bytes: number | string | null | undefined): string => {
  if (bytes === null || bytes === undefined || bytes === '') {
    return '-'
  }
  const n = typeof bytes === 'number' ? bytes : Number(String(bytes).trim())
  if (!Number.isFinite(n) || n < 0) {
    return '-'
  }
  if (n === 0) {
    return '0 B'
  }

  const units = ['B', 'KB', 'MB', 'GB', 'TB', 'PB']
  const k = 1024
  const i = Math.min(Math.floor(Math.log(n) / Math.log(k)), units.length - 1)

  return `${(n / Math.pow(k, i)).toFixed(2)} ${units[i]}`
}

/**
 * 格式化字节数
 * 将字节数转换为更易读的格式，formatFileSize的别名
 *
 * @param bytes 字节数
 * @returns 格式化后的字节数，自动选择合适的单位
 *
 * @example
 * formatBytes(1024) // '1.00 KB'
 * formatBytes(1048576) // '1.00 MB'
 */
export const formatBytes = formatFileSize

/**
 * 格式化手机号码
 * 将手机号中间4位替换为星号
 *
 * @param phone 手机号码，需要是11位数字
 * @returns 格式化后的手机号码，形如：188****8888
 *
 * @example
 * formatPhone('18812345678') // '188****5678'
 */
export const formatPhone = (phone: string): string => {
  return phone.replace(/^(\d{3})\d{4}(\d{4})$/, '$1****$2')
}

/**
 * 格式化身份证号
 * 保留前3位和后4位，中间用星号代替
 *
 * @param idCard 身份证号
 * @returns 格式化后的身份证号，形如：110***********1234
 *
 * @example
 * formatIdCard('110101199001011234') // '110***********1234'
 */
export const formatIdCard = (idCard: string): string => {
  return idCard.replace(/^(\d{3})\d*(\d{4})$/, '$1***********$2')
}

/**
 * 解析数组字段 - 兼容JSON字符串和逗号分隔字符串
 * 用于处理后端返回的数组字段，支持多种格式
 *
 * @param field 待解析的字段值
 * @returns 解析后的字符串数组
 *
 * @example
 * parseArrayField('["a","b","c"]') // ['a', 'b', 'c']
 * parseArrayField('a,b,c') // ['a', 'b', 'c']
 * parseArrayField(['a', 'b', 'c']) // ['a', 'b', 'c']
 * parseArrayField(null) // []
 */
export const parseArrayField = (field: any): string[] => {
  if (!field) return []
  if (Array.isArray(field)) return field
  if (typeof field === 'string') {
    try {
      // 尝试解析JSON数组
      const parsed = JSON.parse(field)
      return Array.isArray(parsed) ? parsed : []
    } catch {
      // 如果不是JSON，按逗号分隔处理
      return field.split(',').filter(Boolean)
    }
  }
  return []
}

/**
 * 安全解析JSON数组，失败时返回空数组
 * @param jsonString JSON字符串
 * @returns 解析后的数组或空数组
 *
 * @example
 * const array = safeParseJsonArray('[{"id":1,"name":"test"}]') // [{id:1,name:"test"}]
 * const emptyArray = safeParseJsonArray('invalid json') // []
 */
export const safeParseJsonArray = <T = any>(jsonString: string): T[] => {
  if (!jsonString) return []

  try {
    const parsed = JSON.parse(jsonString)
    return Array.isArray(parsed) ? parsed : []
  } catch (error) {
    logger.warn('JSON数组解析失败，返回空数组', error)
    return []
  }
}

/**
 * 格式化持续时间
 * @param duration 持续时间（毫秒）
 * @returns 格式化后的持续时间，如 "2小时30分钟"、"1天5小时" 等
 */
export const formatDuration = (duration: number): string => {
  if (duration < 0) return '0秒'

  const seconds = Math.floor(duration / 1000) % 60
  const minutes = Math.floor(duration / (1000 * 60)) % 60
  const hours = Math.floor(duration / (1000 * 60 * 60)) % 24
  const days = Math.floor(duration / (1000 * 60 * 60 * 24))

  const parts = []
  if (days > 0) parts.push(`${days}天`)
  if (hours > 0) parts.push(`${hours}小时`)
  if (minutes > 0) parts.push(`${minutes}分钟`)
  if (seconds > 0 && days === 0 && hours === 0) parts.push(`${seconds}秒`)

  return parts.length > 0 ? parts.join('') : '0秒'
}

/**
 * 格式化分页数据为Naive UI分页格式
 * 将后端返回的分页数据转换为Naive UI分页组件可识别的格式
 *
 * @param data 后端返回的JsonDataObj数据或pageQueryData字符串
 * @returns Naive UI分页对象
 * {
 *   page: number; // 当前页码
 *   pageSize: number; // 每页条数
 *   itemCount: number; // 总条数
 *   pageCount: number; // 总页数
 * }
 *
 * @example
 * // 使用JsonDataObj
 * const pageData = formatNaivePagination(response)
 * // 使用pageQueryData字符串
 * const pageData = formatNaivePagination(response.pageQueryData)
 * // 用于n-pagination组件
 * <RsPagination v-bind="pageData" />
 */
export const formatNaivePagination = (data: JsonDataObj | string) => {
  try {
    const defaultPagination = {
      page: 1,
      pageSize: 30,
      itemCount: 0,
      pageCount: 1,
    }

    if (!data) {
      return defaultPagination
    }

    let pageQueryData: string
    if (typeof data === 'string') {
      pageQueryData = data
    } else {
      pageQueryData = data.pageQueryData || ''
    }

    if (!pageQueryData) {
      return defaultPagination
    }

    const pageInfo = JSON.parse(pageQueryData)
    const itemCount = pageInfo.totalCount || 0
    const pageSize = pageInfo.pageSize || 10
    const pageCount = Math.ceil(itemCount / pageSize)

    return {
      page: pageInfo.pageIndex || 1,
      pageSize: pageSize,
      itemCount: itemCount,
      pageCount: pageCount,
    }
  } catch (error) {
    logger.error('格式化分页数据失败', error)
    return {
      page: 1,
      pageSize: 10,
      itemCount: 0,
      pageCount: 1,
    }
  }
}

/**
 * 将 JSON 列解析为嵌套对象，供 NamePath 字段 `column.xxx` 读写。
 * 兼容历史扁平 key `column.xxx`。
 */
export const parseJsonObjectColumn = (data: Record<string, any>, column: string): void => {
  const raw = data[column]
  let obj: Record<string, any> = {}
  if (typeof raw === 'string' && raw) {
    try {
      const parsed = JSON.parse(raw)
      if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
        obj = parsed
      }
    } catch (error) {
      logger.warn(`解析 ${column} 失败`, error)
    }
  } else if (raw && typeof raw === 'object' && !Array.isArray(raw)) {
    obj = { ...raw }
  }

  const head = `${column}.`
  Object.keys(data).forEach((key) => {
    if (key.startsWith(head)) {
      obj[key.slice(head.length)] = data[key]
      delete data[key]
    }
  })
  data[column] = obj
}

/**
 * 将嵌套对象（及历史扁平 key）打包回 JSON 字符串列。
 */
export const stringifyJsonObjectColumn = (data: Record<string, any>, column: string): void => {
  const nested = data[column]
  const obj: Record<string, any> =
    nested && typeof nested === 'object' && !Array.isArray(nested) ? { ...nested } : {}
  const head = `${column}.`
  Object.keys(data).forEach((key) => {
    if (key.startsWith(head)) {
      obj[key.slice(head.length)] = data[key]
      delete data[key]
    }
  })
  delete data[column]
  if (Object.keys(obj).length === 0) {
    data[column] = ''
    return
  }
  try {
    data[column] = JSON.stringify(obj)
  } catch (error) {
    logger.error(`打包 ${column} 失败`, error)
    data[column] = ''
  }
}

/**
 * 将 extProperty JSON 字符串解析为嵌套对象
 * 用于表单回填，NamePath 字段读写 `extProperty.xxx`
 *
 * @param data 包含 extProperty 字段的数据对象
 * @returns 修改后的数据对象（原地修改）
 *
 * @example
 * const data = {
 *   id: '123',
 *   name: 'test',
 *   extProperty: '{"alertEnabled":"Y","channelName":"default"}'
 * }
 * flattenExtProperty(data)
 * // 结果：data.extProperty = { alertEnabled: 'Y', channelName: 'default' }
 */
export const flattenExtProperty = (data: Record<string, any>): void => {
  parseJsonObjectColumn(data, 'extProperty')
  const extObj = data.extProperty
  if (!extObj || typeof extObj !== 'object' || Array.isArray(extObj)) {
    return
  }
  if (extObj.alertStatusCodes === undefined) {
    return
  }
  if (Array.isArray(extObj.alertStatusCodes)) {
    extObj.alertStatusCodes = extObj.alertStatusCodes.map(String)
  } else if (typeof extObj.alertStatusCodes === 'string') {
    extObj.alertStatusCodes = extObj.alertStatusCodes.split(',').map((s: string) => s.trim()).filter(Boolean)
  }
}

/**
 * 将 extProperty 嵌套对象打包回 JSON 字符串
 * 用于表单提交
 *
 * @param data 包含 extProperty 对象的数据对象
 * @returns 修改后的数据对象（原地修改）
 *
 * @example
 * const data = {
 *   id: '123',
 *   name: 'test',
 *   extProperty: { alertEnabled: 'Y', channelName: 'default' }
 * }
 * unflattenExtProperty(data)
 * // 结果：data.extProperty = '{"alertEnabled":"Y","channelName":"default"}'
 */
export const unflattenExtProperty = (data: Record<string, any>): void => {
  stringifyJsonObjectColumn(data, 'extProperty')
}
