import { randomUUID } from '@/utils/uuid'
import type { AxiosHeaderValue, AxiosRequestHeaders } from 'axios'

/** 与后端 LoggerMiddleware 对齐的追踪头 */
export const TRACE_ID_HEADER = 'X-Trace-ID'
export const REQUEST_ID_HEADER = 'X-Request-ID'
/**
 * W3C Trace Context。国际链路追踪标准头，APM / 网关可识别。
 * 本服务仍以 X-Trace-ID 为准；本头与其共用同一 trace-id，便于跨系统对齐。
 */
export const TRACEPARENT_HEADER = 'traceparent'

/**
 * 请求头补丁辅助类。
 * 入参使用 axios 的 AxiosRequestHeaders（拦截器里已是 AxiosHeaders 实例）。
 * 当前负责追踪头：X-Trace-ID、X-Request-ID、traceparent。后续要补公共头，加在本类，不要写回 request.ts。
 */
export class RequestHeaderHelper {
  /**
   * 读取单个请求头。没有或非字符串时返回空串。
   * @param headers - axios 请求头
   * @param name - 头名，如 X-Request-ID
   * @returns 头值，没有则为空串
   */
  get(headers: AxiosRequestHeaders, name: string): string {
    return this.toHeaderString(headers.get(name))
  }

  /**
   * 写入单个请求头。
   * @param headers - axios 请求头
   * @param name - 头名
   * @param value - 头值
   */
  set(headers: AxiosRequestHeaders, name: string, value: string): void {
    headers.set(name, value)
  }

  /**
   * 发出请求前补齐追踪头。
   * X-Trace-ID / X-Request-ID 与后端 getOrGenerateTraceID 对齐：已有不覆盖，只带一个则补另一个。
   * traceparent 在未设置时由同一 id 派生，格式 00-{32hex}-{16hex}-01。
   * @param headers - axios 请求头
   * @returns 本次使用的追踪 id（UUID 形式）
   */
  patchTrace(headers: AxiosRequestHeaders): string {
    const traceId = this.get(headers, TRACE_ID_HEADER)
    const requestId = this.get(headers, REQUEST_ID_HEADER)
    const id = traceId || requestId || this.createId()
    if (!traceId) {
      this.set(headers, TRACE_ID_HEADER, id)
    }
    if (!requestId) {
      this.set(headers, REQUEST_ID_HEADER, id)
    }
    if (!this.get(headers, TRACEPARENT_HEADER)) {
      this.set(headers, TRACEPARENT_HEADER, this.toTraceparent(id))
    }
    return id
  }

  /**
   * 从已发出的请求头上读追踪 id，供失败日志使用。
   * @param headers - axios 请求头；请求未发出时可能为空
   * @returns 请求 id，没有则为空串
   */
  readRequestId(headers?: AxiosRequestHeaders): string {
    if (!headers) {
      return ''
    }
    return this.get(headers, REQUEST_ID_HEADER) || this.get(headers, TRACE_ID_HEADER)
  }

  /**
   * AxiosHeaders.get 的返回值是 AxiosHeaderValue。追踪 id 只收字符串。
   * @param value - axios 头值
   * @returns 字符串头值
   */
  private toHeaderString(value: AxiosHeaderValue): string {
    if (typeof value === 'string') {
      return value
    }
    if (Array.isArray(value) && typeof value[0] === 'string') {
      return value[0]
    }
    return ''
  }

  /**
   * 把本项目的请求 id 转成 W3C traceparent。
   * UUID（去掉横线）正好 32 位 hex，可直接当 trace-id；否则现场生成合法 hex。
   * @param requestId - X-Request-ID 值
   * @returns traceparent
   */
  private toTraceparent(requestId: string): string {
    const hex = requestId.replaceAll('-', '').toLowerCase()
    if (hex.length === 32 && this.isHex(hex)) {
      return `00-${hex}-${hex.slice(16)}-01`
    }
    const bytes = new Uint8Array(24)
    if (typeof crypto !== 'undefined' && typeof crypto.getRandomValues === 'function') {
      crypto.getRandomValues(bytes)
    }
    const full = Array.from(bytes, (byte) => byte.toString(16).padStart(2, '0')).join('')
    return `00-${full.slice(0, 32)}-${full.slice(32, 48)}-01`
  }

  /**
   * 是否为小写十六进制字符串。
   * @param value - 待测字符串
   * @returns 是否全为 0-9a-f
   */
  private isHex(value: string): boolean {
    for (const ch of value) {
      const cp = ch.codePointAt(0)
      if (cp === undefined) {
        return false
      }
      const isDigit = cp >= 48 && cp <= 57
      const isAf = cp >= 97 && cp <= 102
      if (!isDigit && !isAf) {
        return false
      }
    }
    return true
  }

  /**
   * 生成追踪 id。
   * 不用 crypto.randomUUID：它要求安全上下文，局域网 HTTP 不可用。
   * getRandomValues 在 HTTP 下可用；都没有时退回项目内 HTTP 兼容的 randomUUID。
   */
  private createId(): string {
    if (typeof crypto !== 'undefined' && typeof crypto.getRandomValues === 'function') {
      const bytes = new Uint8Array(16)
      crypto.getRandomValues(bytes)
      // UUID v4：第 7 字节高 4 位置 0100，第 9 字节高 2 位置 10
      bytes[6] = (bytes[6] & 0x0f) | 0x40
      bytes[8] = (bytes[8] & 0x3f) | 0x80
      const hex = Array.from(bytes, (byte) => byte.toString(16).padStart(2, '0')).join('')
      return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(12, 16)}-${hex.slice(16, 20)}-${hex.slice(20)}`
    }
    return randomUUID()
  }
}

/** 全局单例，供 HTTP 拦截器补追踪头。 */
export const requestHeaderHelper = new RequestHeaderHelper()
