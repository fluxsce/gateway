/**
 * HTTP 客户端共享类型。
 * 业务模块一般只使用 RequestConfig；拦截器内部才用 InternalRequestConfig。
 */
import type { JsonDataObj } from '@/types/api'
import type { AxiosRequestConfig, InternalAxiosRequestConfig } from 'axios'

/**
 * HTTP 方法。createApi / request 的 method 参数使用本联合类型。
 */
export type RequestMethod = 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH' | 'HEAD' | 'OPTIONS'

/**
 * 单次请求的业务元数据。不参与后端 BindSafely，只给拦截器、插件、排障用。
 * 新模块一般不必手写：createApi 会从 /gateway/hubxxxx 自动填 module。
 */
export interface RequestMeta {
  /** 模块编码，如 hub0002、hubcommon002 */
  module?: string
}

/**
 * 业务请求配置，在 axios AxiosRequestConfig 上增加本项目字段。
 * 取消请求使用继承来的 signal（AbortSignal），与国际客户端的 AbortController 对齐。
 */
export interface RequestConfig extends AxiosRequestConfig {
  /** 是否计入顶栏进度，默认 true；下载/上传可关 */
  showLoading?: boolean
  /** 模块标签等，不发给后端 */
  meta?: RequestMeta
}

/**
 * 拦截器内部配置。axios 只认 AxiosRequestConfig，showLoading / meta 由本层读写。
 */
export interface InternalRequestConfig extends InternalAxiosRequestConfig {
  /** 是否计入顶栏进度 */
  showLoading?: boolean
  /** 模块标签等 */
  meta?: RequestMeta
}

/**
 * 模块级 API 客户端。由 createApi 返回。
 * 新 hub 模块只通过本接口发 JSON，返回值一律按 JsonDataObj 处理。
 */
export interface ModuleApi {
  /**
   * 按方法发请求。
   * @param method - HTTP 方法
   * @param path - 相对模块前缀的子路径
   * @param data - 写操作请求体
   * @param params - 查询参数
   * @param config - 额外配置
   */
  request(
    method: RequestMethod,
    path?: string,
    data?: unknown,
    params?: unknown,
    config?: RequestConfig,
  ): Promise<JsonDataObj>
  /**
   * GET。
   * @param path - 子路径
   * @param params - 查询参数
   * @param config - 额外配置
   */
  get(path?: string, params?: unknown, config?: RequestConfig): Promise<JsonDataObj>
  /**
   * POST。
   * @param path - 子路径
   * @param data - 请求体
   * @param params - 查询参数
   * @param config - 额外配置
   */
  post(path?: string, data?: unknown, params?: unknown, config?: RequestConfig): Promise<JsonDataObj>
  /**
   * PUT。
   * @param path - 子路径
   * @param data - 请求体
   * @param config - 额外配置
   */
  put(path?: string, data?: unknown, config?: RequestConfig): Promise<JsonDataObj>
  /**
   * DELETE。
   * @param path - 子路径
   * @param params - 查询参数
   * @param config - 额外配置
   */
  delete(path?: string, params?: unknown, config?: RequestConfig): Promise<JsonDataObj>
  /**
   * PATCH。
   * @param path - 子路径
   * @param data - 请求体
   * @param config - 额外配置
   */
  patch(path?: string, data?: unknown, config?: RequestConfig): Promise<JsonDataObj>
  /**
   * 拼出完整路径，用于需要原始 URL 的场景。
   * @param path - 子路径
   */
  getURL(path?: string): string
}
