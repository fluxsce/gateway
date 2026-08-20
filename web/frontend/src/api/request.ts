/**
 * 全局 HTTP 客户端入口。
 *
 * 契约（新模块只走这一条，不要再包一层）：
 * 1. 业务 JSON 用 createApi(moduleApiPrefix('hubxxxx')) 或 request()，得到 JsonDataObj。
 * 2. 页面用 isApiSuccess / getApiMessage / parseJsonData；不要按 HTTP 403 写 catch。
 * 3. 后端已给出 JsonData 的 HTTP 4xx/5xx 会 resolve 为 oK=false，不当异常抛。
 * 4. 仅网络中断、超时、取消才 reject（HttpTransportError / HttpCancelledError）。401 在拦截器弹窗。
 * 5. blob/multipart 才用 named export http；请求前扩展走 registerHttpPlugin，响应旁路走 registerHttpResponsePlugin。
 * 6. 取消使用 RequestConfig.signal（AbortSignal），与国际客户端 AbortController 对齐。
 */
import { requestInterceptorHelper } from '@/api/requestInterceptor'
import { requestPathHelper } from '@/api/requestPath'
import type { ModuleApi, RequestConfig, RequestMethod } from '@/api/requestTypes'
import { responseInterceptorHelper } from '@/api/responseInterceptor'
import type { JsonDataObj } from '@/types/api'
import axios from 'axios'

export {
  HttpCancelledError,
  HttpTransportError,
  getHttpErrorMessage,
  isHttpCancelledError,
  isHttpTransportError,
} from '@/api/requestError'
export { registerHttpPlugin, registerHttpResponsePlugin } from '@/api/requestPlugin'
export type { HttpRequestPlugin, HttpResponsePlugin } from '@/api/requestPlugin'
export type {
  InternalRequestConfig,
  ModuleApi,
  RequestConfig,
  RequestMeta,
  RequestMethod
} from '@/api/requestTypes'

const http = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  timeout: 15000,
  // 会话在 Cookie 里，跨端口开发也要带上
  withCredentials: true,
  // 与后端 BindSafely 默认表单绑定对齐；JSON/multipart 由单次请求覆盖
  headers: {
    'Content-Type': 'application/x-www-form-urlencoded; charset=UTF-8',
  },
})

requestInterceptorHelper.attach(http)
responseInterceptorHelper.attach(http)

/**
 * 运行时更新超时（毫秒），例如按用户会话时长拉长。
 * @param timeout - 超时毫秒
 */
export function updateTimeout(timeout: number): void {
  if (timeout > 0) {
    http.defaults.timeout = timeout
  }
}

/**
 * 当前超时（毫秒）。
 * @returns 当前超时毫秒
 */
export function getCurrentTimeout(): number {
  return http.defaults.timeout || 15000
}

/**
 * 本项目写操作把参数放 body；GET/DELETE 等放 query。
 * @param method - HTTP 方法
 * @returns 是否为写操作
 */
function isWriteMethod(method?: string): boolean {
  const upper = String(method || 'GET').toUpperCase()
  return upper !== 'GET' && upper !== 'DELETE' && upper !== 'HEAD' && upper !== 'OPTIONS'
}

/**
 * 合并两次 RequestConfig。meta 必须显式浅合并，否则后一次会整段覆盖模块编码。
 * @param base - 默认配置
 * @param extra - 单次覆盖
 * @returns 合并后的配置
 */
function mergeRequestConfig(base: RequestConfig = {}, extra: RequestConfig = {}): RequestConfig {
  return {
    showLoading: true,
    ...base,
    ...extra,
    meta: {
      ...base.meta,
      ...extra.meta,
    },
  }
}

/**
 * 发业务请求，成功或后端业务失败都 resolve 为 JsonDataObj。
 * 传输失败 reject HttpTransportError；取消 reject HttpCancelledError。
 *
 * @example
 * request({ method: 'POST', url: '/gateway/hub0002/queryUsers', data: params })
 * request('POST', '/gateway/hub0002/queryUsers', params)
 */
export function request(options: RequestConfig): Promise<JsonDataObj>
export function request(
  method: RequestMethod,
  url: string,
  data?: unknown,
  params?: unknown,
  config?: RequestConfig,
): Promise<JsonDataObj>
export function request(
  methodOrOptions: RequestMethod | RequestConfig,
  url?: string,
  data?: unknown,
  params?: unknown,
  config?: RequestConfig,
): Promise<JsonDataObj> {
  if (typeof methodOrOptions === 'object') {
    return http(mergeRequestConfig(methodOrOptions)) as Promise<JsonDataObj>
  }

  const method = methodOrOptions
  const requestConfig: RequestConfig = mergeRequestConfig(
    {
      method,
      url,
    },
    config,
  )
  if (isWriteMethod(method)) {
    requestConfig.data = data
    requestConfig.params = params
  } else {
    requestConfig.params = { ...(params as object), ...config?.params }
  }
  return http(requestConfig) as Promise<JsonDataObj>
}

/**
 * GET。查询参数走 params。插件层兼容旧调用；新模块优先 createApi。
 * @param url - 请求地址
 * @param params - 查询参数
 * @param config - 额外配置
 * @returns JsonDataObj
 */
export function get(url: string, params?: unknown, config?: RequestConfig): Promise<JsonDataObj> {
  return request('GET', url, undefined, params, config)
}

/**
 * POST。body 走 data。插件层兼容旧调用；新模块优先 createApi。
 * @param url - 请求地址
 * @param data - 请求体
 * @param params - 查询参数
 * @param config - 额外配置
 * @returns JsonDataObj
 */
export function post(
  url: string,
  data?: unknown,
  params?: unknown,
  config?: RequestConfig,
): Promise<JsonDataObj> {
  return request('POST', url, data, params, config)
}

/**
 * PUT。插件层兼容旧调用；新模块优先 createApi。
 * @param url - 请求地址
 * @param data - 请求体
 * @param config - 额外配置
 * @returns JsonDataObj
 */
export function put(url: string, data?: unknown, config?: RequestConfig): Promise<JsonDataObj> {
  return request('PUT', url, data, undefined, config)
}

/**
 * DELETE。插件层兼容旧调用；新模块优先 createApi。
 * @param url - 请求地址
 * @param params - 查询参数
 * @param config - 额外配置
 * @returns JsonDataObj
 */
export function del(url: string, params?: unknown, config?: RequestConfig): Promise<JsonDataObj> {
  return request('DELETE', url, undefined, params, config)
}

/**
 * PATCH。插件层兼容旧调用；新模块优先 createApi。
 * @param url - 请求地址
 * @param data - 请求体
 * @param config - 额外配置
 * @returns JsonDataObj
 */
export function patch(url: string, data?: unknown, config?: RequestConfig): Promise<JsonDataObj> {
  return request('PATCH', url, data, undefined, config)
}

/**
 * 模块级 API 前缀，例如 createApi(moduleApiPrefix('hub0002'))。
 * 新模块只在 views/hubxxxx/api 里建这一个实例。
 * 模块编码优先用 defaultConfig.meta.module，否则取前缀 /gateway/ 后第一段路径。
 *
 * @param baseURL - 模块前缀，如 moduleApiPrefix('hub0002')
 * @param defaultConfig - 该模块默认配置（超时、showLoading、meta、signal）
 * @returns 模块 API
 */
export function createApi(baseURL: string, defaultConfig: RequestConfig = {}): ModuleApi {
  const defaults = mergeRequestConfig(
    {
      meta: { module: requestPathHelper.readModule(baseURL) },
    },
    defaultConfig,
  )

  const send = (
    method: RequestMethod,
    path = '',
    data?: unknown,
    params?: unknown,
    config?: RequestConfig,
  ): Promise<JsonDataObj> => {
    return request(
      method,
      requestPathHelper.join(baseURL, path),
      data,
      params,
      mergeRequestConfig(defaults, config),
    )
  }

  return {
    request: (
      method: RequestMethod,
      path?: string,
      data?: unknown,
      params?: unknown,
      config?: RequestConfig,
    ) => send(method, path, data, params, config),
    get: (path?: string, params?: unknown, config?: RequestConfig) =>
      send('GET', path, undefined, params, config),
    post: (path?: string, data?: unknown, params?: unknown, config?: RequestConfig) =>
      send('POST', path, data, params, config),
    put: (path?: string, data?: unknown, config?: RequestConfig) =>
      send('PUT', path, data, undefined, config),
    delete: (path?: string, params?: unknown, config?: RequestConfig) =>
      send('DELETE', path, undefined, params, config),
    patch: (path?: string, data?: unknown, config?: RequestConfig) =>
      send('PATCH', path, data, undefined, config),
    getURL: (path = '') => requestPathHelper.join(baseURL, path),
  }
}

/**
 * 原始 axios 实例。仅给 blob 下载、multipart 上传等需要完整 AxiosResponse 的调用。
 * 业务 JSON 请用 request / createApi。不要对 http 再挂拦截器，以免绕过 JsonData 契约。
 */
export { http }
export default http
