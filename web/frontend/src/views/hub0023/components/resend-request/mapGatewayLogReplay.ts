/**
 * 将网关访问日志详情映射为 GRestfulApi 的初始请求参数，用于日志重发场景。
 */

import type { RestBodyProcessType, RestHttpMethod } from '@/components/grestful-api/types'
import type { GatewayLogInfo } from '../../types'

const METHODS: RestHttpMethod[] = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'HEAD', 'OPTIONS']

/**
 * 与 internal/gateway/constants 中 HeaderXGatewayReplay、HeaderXGatewayReplayTraceID 大小写一致。
 */
const HEADER_X_GATEWAY_REPLAY = 'X-Gateway-Replay'
const HEADER_X_GATEWAY_REPLAY_TRACE_ID = 'X-Gateway-Replay-Trace-Id'

/** 与 bootstrap.isGatewayReplayMarker 约定一致，仅 Y（大小写不敏感）表示重发 */
const GATEWAY_REPLAY_MARKER_VALUE = 'Y'

const HEADER_CONTENT_LENGTH = 'Content-Length'
const HEADER_TRANSFER_ENCODING = 'Transfer-Encoding'
const HEADER_CONNECTION = 'Connection'

/** 与 GatewayReplayOmitFromLogRequestHeaders 顺序与字面量一致（仅 hop-by-hop，不含重放信令头） */
const GATEWAY_REPLAY_OMIT_FROM_LOG_HEADERS: readonly string[] = [
  HEADER_CONTENT_LENGTH,
  HEADER_TRANSFER_ENCODING,
  HEADER_CONNECTION,
]

const GATEWAY_REPLAY_OMIT_FROM_LOG_HEADERS_LOWER = new Set(
  GATEWAY_REPLAY_OMIT_FROM_LOG_HEADERS.map((h) => h.toLowerCase()),
)

/**
 * 从历史 requestHeaders JSON 中剔除 hop-by-hop 头；不重放信令头（由 buildGatewayReplayHeadersJson 统一写入）。
 * 解析失败时返回 undefined。
 */
export function stripHeadersForGatewayReplay(headersJson: string): string | undefined {
  const raw = headersJson.trim()
  if (!raw) {
    return undefined
  }
  try {
    const parsed = JSON.parse(raw) as unknown
    if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) {
      return undefined
    }
    const out: Record<string, string> = {}
    for (const [k, v] of Object.entries(parsed as Record<string, unknown>)) {
      const key = k.trim()
      if (!key || GATEWAY_REPLAY_OMIT_FROM_LOG_HEADERS_LOWER.has(key.toLowerCase())) {
        continue
      }
      const lk = key.toLowerCase()
      if (lk === HEADER_X_GATEWAY_REPLAY.toLowerCase() || lk === HEADER_X_GATEWAY_REPLAY_TRACE_ID.toLowerCase()) {
        continue
      }
      out[k] = v == null ? '' : String(v)
    }
    if (Object.keys(out).length === 0) {
      return undefined
    }
    return JSON.stringify(out)
  } catch {
    return undefined
  }
}

/**
 * 组装重发请求头 JSON：保留日志中除 hop-by-hop 与旧重放信令外的头，并强制带上 X-Gateway-Replay: Y 与 X-Gateway-Replay-Trace-Id（网关入口消费后移除）。
 */
export function buildGatewayReplayHeadersJson(log: GatewayLogInfo): string {
  const base = stripHeadersForGatewayReplay(log.requestHeaders ?? '')
  const obj: Record<string, string> = base
    ? (JSON.parse(base) as Record<string, string>)
    : {}
  obj[HEADER_X_GATEWAY_REPLAY] = GATEWAY_REPLAY_MARKER_VALUE
  obj[HEADER_X_GATEWAY_REPLAY_TRACE_ID] = (log.traceId ?? '').trim()
  return JSON.stringify(obj)
}

/**
 * 将网关记录的请求方法规范化为 GRestfulApi 支持的枚举值，无法识别时回退为 GET。
 */
export function normalizeRestHttpMethod(m: string | undefined): RestHttpMethod {
  const u = (m || 'GET').toUpperCase()
  return METHODS.includes(u as RestHttpMethod) ? (u as RestHttpMethod) : 'GET'
}

/**
 * 根据网关日志详情组装重发所需的初始 URL、方法、可选头与 Body。
 * 优先使用服务端按实例暴露端口拼装的 resetUrl；其次绝对 URL 的 forwardAddress（直连下游）；
 * 否则回退为 requestPath + requestQuery（相对路径，相对当前站点）。
 * 请求头始终携带 X-Gateway-Replay 与 X-Gateway-Replay-Trace-Id，供网关识别重发并由后端剥离。
 */
export function buildGatewayLogReplayInit(log: GatewayLogInfo): {
  initialUrl: string
  initialMethod: RestHttpMethod
  initialHeadersJson: string
  initialRawBody?: string
  initialBodyProcessType?: RestBodyProcessType
} {
  const method = normalizeRestHttpMethod(log.requestMethod)
  let url = ''
  const reset = log.resetUrl?.trim()
  if (reset && /^https?:\/\//i.test(reset)) {
    url = reset
  } else {
    const fa = log.forwardAddress?.trim()
    if (fa && /^https?:\/\//i.test(fa)) {
      url = fa
    } else {
      const path = log.requestPath?.trim() || '/'
      const q = log.requestQuery?.trim()
      if (q) {
        url = q.startsWith('?') ? `${path}${q}` : `${path}?${q}`
      } else {
        url = path
      }
    }
  }

  const initialHeadersJson = buildGatewayReplayHeadersJson(log)
  const bodyStr =
    log.requestBody != null && String(log.requestBody).trim() !== '' ? String(log.requestBody) : ''

  let initialBodyProcessType: RestBodyProcessType | undefined
  if (bodyStr && method !== 'GET' && method !== 'HEAD') {
    initialBodyProcessType = 'json'
    try {
      JSON.parse(bodyStr)
    } catch {
      initialBodyProcessType = 'raw'
    }
  }

  return {
    initialUrl: url,
    initialMethod: method,
    initialHeadersJson,
    ...(bodyStr && method !== 'GET' && method !== 'HEAD'
      ? { initialRawBody: bodyStr, initialBodyProcessType }
      : {}),
  }
}
