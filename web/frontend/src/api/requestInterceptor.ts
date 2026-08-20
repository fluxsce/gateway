/**
 * 请求拦截器：发出前补齐模块标签、追踪头、执行请求插件、打开 loading。
 * 后续要加请求前逻辑改本类，不要写回 request.ts。
 */
import { requestHeaderHelper } from '@/api/requestHeader'
import { requestLoadingHelper } from '@/api/requestLoading'
import { requestPathHelper } from '@/api/requestPath'
import { requestPluginHelper } from '@/api/requestPlugin'
import type { InternalRequestConfig } from '@/api/requestTypes'
import type { AxiosError, AxiosInstance } from 'axios'

/**
 * 请求拦截器辅助类。挂到 axios 实例后，每条请求都会经过 onFulfilled。
 */
export class RequestInterceptorHelper {
  /**
   * 把请求拦截器挂到 axios 实例。应用生命周期内调用一次。
   * @param http - 全局 axios 实例
   */
  attach(http: AxiosInstance): void {
    http.interceptors.request.use(
      async (cfg: InternalRequestConfig) => {
        cfg.showLoading ??= true
        cfg.meta ??= {}
        // 裸 request() 没有 createApi 前缀时，从 URL 第一段补模块标签
        cfg.meta.module ??= requestPathHelper.readModule(cfg.url)
        requestHeaderHelper.patchTrace(cfg.headers)
        await requestPluginHelper.applyRequest(cfg)
        if (cfg.showLoading) {
          requestLoadingHelper.begin()
        }
        return cfg
      },
      (error: AxiosError) => {
        // 请求还没发出就失败时也要成对关掉 loading
        requestLoadingHelper.end((error.config as InternalRequestConfig | undefined)?.showLoading)
        return Promise.reject(error)
      },
    )
  }
}

/** 全局单例，应用启动时挂一次。 */
export const requestInterceptorHelper = new RequestInterceptorHelper()
