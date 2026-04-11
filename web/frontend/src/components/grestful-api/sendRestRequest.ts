/**
 * REST 调试请求发送：将 UI 组装的 {@link SendRestRequestInput} 转为网关 `hubplugin/http/execute` 调用，
 * 由服务端使用 `pkg/httpclient` 代发，避免浏览器直连目标站。
 *
 * @module
 */

import { request } from '@/api/request'
import type { JsonDataObj } from '@/types/api'
import axios from 'axios'
import { getApiMessage, isApiSuccess, parseJsonData } from '@/utils/format'
import type {
  HttpExecuteResultDTO,
  RestKeyValueRow,
  RestRequestFailure,
  RestRequestResult,
  SendRestRequestInput,
  SendRestRequestOutput
} from './types'

/**
 * 网关 hubplugin HTTP 代发接口路径（相对 `import.meta.env.VITE_API_BASE_URL`）。
 *
 * @remarks
 * 实际出站请求在服务端执行，见 Go 路由 `web/views/hubplugin/http`。
 */
export const GATEWAY_HTTP_EXECUTE_URL = '/gateway/hubplugin/http/execute'

/**
 * 尝试将用户输入解析为 `URL`，相对路径以当前页面的 `location.origin` 为基（与浏览器地址栏行为一致）。
 * 未完成或非法的字符串返回 `null`，供地址栏与 Params 同步时安全调用。
 *
 * @param input - 地址字符串
 * @returns 解析后的 `URL`，无法解析时返回 `null`
 */
export function tryParseUserUrl(input: string): URL | null {
  const s = input.trim()
  if (!s) {
    return null
  }
  try {
    return new URL(s)
  } catch {
    try {
      const base =
        typeof window !== 'undefined' && window.location?.origin
          ? window.location.origin
          : 'http://localhost'
      return new URL(s, base)
    } catch {
      return null
    }
  }
}

/**
 * 将用户输入解析为绝对 `URL`（相对路径基于当前页 `location.origin`）。
 *
 * @param input - 地址字符串
 * @returns 解析后的 `URL`
 * @throws {Error} 空字符串或无法解析时抛出「请输入 URL」或「无效的 URL」
 */
function parseUserUrl(input: string): URL {
  const u = tryParseUserUrl(input)
  if (!u) {
    throw new Error(!input.trim() ? '请输入 URL' : '无效的 URL')
  }
  return u
}

/**
 * 将键值行转为 HTTP 头映射；忽略未启用或空 key；同名（忽略大小写）保留最后一次出现的键名写法与值。
 *
 * @param rows - Header 行列表
 * @returns 头名到值的映射
 */
function rowsToHeaderRecord(rows: RestKeyValueRow[]): Record<string, string> {
  const lowerToLast: Record<string, { canon: string; value: string }> = {}
  for (const row of rows) {
    if (!row.enabled) {
      continue
    }
    const k = row.key.trim()
    if (!k) {
      continue
    }
    const lk = k.toLowerCase()
    lowerToLast[lk] = { canon: k, value: row.value }
  }
  const out: Record<string, string> = {}
  for (const { canon, value } of Object.values(lowerToLast)) {
    out[canon] = value
  }
  return out
}

/**
 * 将 Params 表格中的启用行写入 `url.searchParams`（对同一 key 多次 `set`，后者覆盖前者）。
 *
 * @param url - 待修改的 URL 对象
 * @param params - Query 键值行
 */
function mergeQueryParams(url: URL, params: RestKeyValueRow[]): void {
  for (const row of params) {
    if (!row.enabled) {
      continue
    }
    const k = row.key.trim()
    if (!k) {
      continue
    }
    url.searchParams.set(k, row.value)
  }
}

/**
 * 将 {@link HttpExecuteResultDTO} 中的 `body`（响应报文）转为界面展示的 UTF-8 文本。
 *
 * @param dto - 网关代发成功返回的业务数据
 * @returns UTF-8 字符串；Base64 解码失败时回退为原始 `body`
 */
function decodeExecuteBody(dto: HttpExecuteResultDTO): string {
  if (!dto.bodyBase64) {
    return dto.body ?? ''
  }
  try {
    const binary = atob(dto.body)
    const bytes = new Uint8Array(binary.length)
    for (let i = 0; i < binary.length; i++) {
      bytes[i] = binary.charCodeAt(i)
    }
    return new TextDecoder('utf-8', { fatal: false }).decode(bytes)
  } catch {
    return dto.body
  }
}

/**
 * 判断头映射中是否已存在 `Content-Type`（忽略大小写）。
 *
 * @param headers - 请求头映射
 */
function hasContentTypeHeader(headers: Record<string, string>): boolean {
  return Object.keys(headers).some((k) => k.toLowerCase() === 'content-type')
}

/**
 * 通过网关 `POST /gateway/hubplugin/http/execute` 由服务端代发 HTTP 请求。
 *
 * @remarks
 * - 使用 `@/api/request`，继承 Cookie 与统一 `JsonData` 错误处理。
 * - 对本接口的调用必须带 `Content-Type: application/json`，否则网关默认 `x-www-form-urlencoded` 无法绑定 `method` 等字段。
 * - `bodyMode === 'multipart'` 时由服务端按 `formData`（含 `type`: text|file）代发。
 * - `bodyMode === 'urlencoded'` 时发送 `formUrlEncoded` map，由服务端编码为 application/x-www-form-urlencoded。
 *
 * @param input - UI 组装的请求参数
 * @returns 成功时 `ok: true` 与 {@link RestRequestResult}；失败时 `ok: false`
 */

/** 与网关服务端 multipart 单文件解码上限一致。 */
const MAX_MULTIPART_FILE_BYTES = 5 * 1024 * 1024

/**
 * 将 `File` 读为无 `data:` 前缀的标准 Base64。
 */
function readFileAsBase64(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => {
      const r = reader.result
      if (typeof r !== 'string') {
        reject(new Error('无法读取文件'))
        return
      }
      const i = r.indexOf(',')
      resolve(i >= 0 ? r.slice(i + 1) : r)
    }
    reader.onerror = () => reject(reader.error ?? new Error('读取文件失败'))
    reader.readAsDataURL(file)
  })
}

export async function sendRestRequest(input: SendRestRequestInput): Promise<SendRestRequestOutput> {
  let url: URL
  try {
    url = parseUserUrl(input.url)
    mergeQueryParams(url, input.queryParams)
  } catch (e) {
    const fail: RestRequestFailure = {
      ok: false,
      message: e instanceof Error ? e.message : 'URL 解析失败',
      cause: e,
    }
    return fail
  }

  const headerRecord = rowsToHeaderRecord(input.headers)

  let requestBody: string | undefined
  let formData:
    | Array<{
        key: string
        type: 'text' | 'file'
        value?: string
        fileName?: string
        contentBase64?: string
      }>
    | undefined
  let formUrlEncoded: Record<string, string> | undefined

  if (input.bodyMode === 'multipart') {
    try {
      formData = []
      for (const row of input.formFields) {
        if (!row.enabled) {
          continue
        }
        const k = row.key.trim()
        if (!k) {
          continue
        }
        if (row.fieldKind === 'file' && row.file) {
          if (row.file.size > MAX_MULTIPART_FILE_BYTES) {
            const fail: RestRequestFailure = {
              ok: false,
              message: `文件「${row.file.name}」超过 5MB 限制`,
            }
            return fail
          }
          const b64 = await readFileAsBase64(row.file)
          formData.push({
            key: k,
            type: 'file',
            fileName: row.file.name,
            contentBase64: b64,
          })
        } else {
          formData.push({ key: k, type: 'text', value: row.value })
        }
      }
      if (formData.length === 0) {
        const fail: RestRequestFailure = {
          ok: false,
          message: '请至少添加一个有效的表单字段',
        }
        return fail
      }
    } catch (e) {
      const fail: RestRequestFailure = {
        ok: false,
        message: e instanceof Error ? e.message : '构建 multipart 失败',
        cause: e,
      }
      return fail
    }
  } else if (input.bodyMode === 'raw') {
    if (!hasContentTypeHeader(headerRecord) && input.rawContentType.trim()) {
      headerRecord['Content-Type'] = input.rawContentType.trim()
    }
    requestBody = input.rawBody
  } else if (input.bodyMode === 'urlencoded') {
    formUrlEncoded = {}
    for (const row of input.formFields) {
      if (!row.enabled) {
        continue
      }
      const k = row.key.trim()
      if (!k) {
        continue
      }
      formUrlEncoded[k] = row.value
    }
    if (Object.keys(formUrlEncoded).length === 0) {
      const fail: RestRequestFailure = {
        ok: false,
        message: '请至少添加一个有效的表单字段',
      }
      return fail
    }
  }

  const payload: Record<string, unknown> = {
    method: input.method,
    url: url.toString(),
    headers: headerRecord,
    timeoutSeconds: 120,
  }
  if (formData !== undefined) {
    payload.formData = formData
  }
  if (formUrlEncoded !== undefined) {
    payload.formUrlEncoded = formUrlEncoded
  }
  if (
    formData === undefined &&
    formUrlEncoded === undefined &&
    requestBody !== undefined &&
    requestBody.length > 0
  ) {
    payload.body = requestBody
  }

  try {
    const res = await request<JsonDataObj>({
      method: 'POST',
      url: GATEWAY_HTTP_EXECUTE_URL,
      data: payload,
      timeout: 125000,
      showLoading: false,
      signal: input.signal,
      headers: {
        'Content-Type': 'application/json;charset=UTF-8',
      },
    })

    if (!isApiSuccess(res)) {
      const fail: RestRequestFailure = {
        ok: false,
        message: getApiMessage(res, '代发请求失败'),
      }
      return fail
    }

    const dto = parseJsonData<HttpExecuteResultDTO>(res, {} as HttpExecuteResultDTO)
    const responseBodyText = decodeExecuteBody(dto)

    const data: RestRequestResult = {
      status: dto.statusCode,
      statusText: dto.status,
      durationMs: dto.durationMs,
      responseHeaders: dto.headers,
      responseBodyText,
    }

    return { ok: true, data }
  } catch (e: unknown) {
    if (axios.isCancel(e)) {
      const fail: RestRequestFailure = {
        ok: false,
        message: '请求已取消',
        cause: e,
      }
      return fail
    }
    const ex = e as { cancelled?: boolean; message?: string }
    if (ex?.cancelled || ex?.message === '请求已取消') {
      const fail: RestRequestFailure = {
        ok: false,
        message: '请求已取消',
        cause: e,
      }
      return fail
    }
    const msg =
      e && typeof e === 'object' && 'message' in e
        ? String((e as { message: string }).message)
        : '请求失败'
    const fail: RestRequestFailure = {
      ok: false,
      message: msg,
      cause: e,
    }
    return fail
  }
}
