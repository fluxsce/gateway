import { getByNamePath, hasByNamePath, setByNamePath } from '@/ui'
import { parseIsoWallClock } from '@/utils/format'

export { getByNamePath, hasByNamePath, setByNamePath }

/**
 * 当前时区的 RFC3339 偏移，如 +08:00。
 */
function localRfc3339Offset(date = new Date()): string {
  const offsetMin = -date.getTimezoneOffset()
  const sign = offsetMin >= 0 ? '+' : '-'
  const abs = Math.abs(offsetMin)
  const hh = String(Math.floor(abs / 60)).padStart(2, '0')
  const mm = String(abs % 60).padStart(2, '0')
  return `${sign}${hh}:${mm}`
}

/**
 * 将表单日期值转为 Go time.Time 可解析的 RFC3339。
 * DatePicker valueFormat=iso 已直接产出 RFC3339；string 墙钟或空值仍在此兜底。
 * 空值返回 null，避免 JSON `""` 无法 unmarshal 到 time.Time。
 * @param value - 表单中的日期字符串或空值
 * @param withTime - 是否包含时分秒
 * @returns RFC3339 字符串或 null
 */
export function toRfc3339SubmitValue(value: unknown, withTime: boolean): string | null {
  if (value == null || value === '') return null
  if (typeof value !== 'string') return null
  const trimmed = value.trim()
  if (!trimmed) return null
  if (/^\d{4}-\d{2}-\d{2}T/.test(trimmed) && /(?:Z|[+-]\d{2}:?\d{2})$/i.test(trimmed)) {
    return trimmed
  }
  const wall = parseIsoWallClock(trimmed)
  if (!wall) return trimmed
  const body = withTime
    ? `${wall.year}-${wall.month}-${wall.day}T${wall.hour}:${wall.minute}:${wall.second}`
    : `${wall.year}-${wall.month}-${wall.day}T00:00:00`
  return `${body}${localRfc3339Offset()}`
}

/**
 * 从 initialData 读字段：优先 NamePath 嵌套，兼容历史扁平 `'a.b'` key。
 */
export function readInitialField(
  initial: Record<string, any>,
  name: string,
): { found: boolean; value: unknown } {
  if (hasByNamePath(initial, name)) {
    return { found: true, value: getByNamePath(initial, name) }
  }
  if (Object.prototype.hasOwnProperty.call(initial, name)) {
    return { found: true, value: initial[name] }
  }
  return { found: false, value: undefined }
}

/**
 * 取出 JSON 子树。嵌套对象优先；兼容历史扁平 `prefix.x` key。
 * 提交时业务侧对返回值 JSON.stringify 即可。
 */
export function takeNamedObject(
  formData: Record<string, any>,
  prefix: string,
): Record<string, any> {
  const nested = formData[prefix]
  if (nested && typeof nested === 'object' && !Array.isArray(nested)) {
    return { ...nested }
  }
  const obj: Record<string, any> = {}
  const head = `${prefix}.`
  Object.keys(formData).forEach((key) => {
    if (key.startsWith(head)) obj[key.slice(head.length)] = formData[key]
  })
  return obj
}

/**
 * 把 initial 里的对象子树拷进 model，避免只写 NamePath 叶子时丢掉同级键。
 * 例如 initial.routeMetadata.serviceNameMap 在没有对应 field 时仍应保留。
 */
export function seedNamedObjects(
  model: Record<string, any>,
  initial: Record<string, any>,
): void {
  Object.keys(initial).forEach((key) => {
    const val = initial[key]
    if (val && typeof val === 'object' && !Array.isArray(val) && !(val instanceof File)) {
      model[key] = val
    }
  })
}
