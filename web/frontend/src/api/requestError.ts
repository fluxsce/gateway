/**
 * HTTP 传输层错误。
 * 仅网络中断、超时、非 JsonData 的 HTTP 失败会抛出本类。
 * 业务失败（含 403）走 JsonDataObj.oK=false，不使用本类。
 */
export class HttpTransportError extends Error {
  readonly name = 'HttpTransportError'
  /** HTTP 状态码；无响应时为 -1 */
  readonly code: number
  /** 原始响应体，可能不是 JsonData */
  readonly data: unknown
  /** 与 X-Request-ID / traceparent 对齐，便于排障 */
  readonly requestId: string
  /** 模块编码，如 hub0002 */
  readonly module?: string
  /** 请求 URL */
  readonly url?: string

  /**
   * @param options - 传输失败的上下文
   */
  constructor(options: {
    message: string
    code: number
    data?: unknown
    requestId?: string
    module?: string
    url?: string
  }) {
    super(options.message)
    this.code = options.code
    this.data = options.data
    this.requestId = options.requestId ?? ''
    this.module = options.module
    this.url = options.url
  }
}

/**
 * 请求被取消。catch 里用 isHttpCancelledError 判断，不要当网络故障提示。
 * cancelled 字段兼容既有按属性探测的调用方。
 */
export class HttpCancelledError extends Error {
  readonly name = 'HttpCancelledError'
  readonly cancelled = true as const

  /**
   * @param message - 取消提示
   */
  constructor(message = '请求已取消') {
    super(message)
  }
}

/**
 * 是否为传输层失败（断网、超时、非 JsonData 的 HTTP 错误）。
 * @param error - catch 到的未知错误
 * @returns 是否为 HttpTransportError
 */
export function isHttpTransportError(error: unknown): error is HttpTransportError {
  return error instanceof HttpTransportError
}

/**
 * 是否为用户或页面取消的请求。
 * @param error - catch 到的未知错误
 * @returns 是否为取消
 */
export function isHttpCancelledError(error: unknown): error is HttpCancelledError {
  if (error instanceof HttpCancelledError) {
    return true
  }
  return (
    typeof error === 'object' &&
    error !== null &&
    'cancelled' in error &&
    (error as { cancelled?: unknown }).cancelled === true
  )
}

/**
 * 从 catch 里取出可展示文案。取消返回空串，调用方应跳过 toast。
 * @param error - catch 到的未知错误
 * @param fallback - 无法识别时的兜底
 * @returns 提示文案
 */
export function getHttpErrorMessage(error: unknown, fallback: string): string {
  if (isHttpCancelledError(error)) {
    return ''
  }
  if (isHttpTransportError(error) && error.message) {
    return error.message
  }
  if (error instanceof Error && error.message) {
    return error.message
  }
  return fallback
}
