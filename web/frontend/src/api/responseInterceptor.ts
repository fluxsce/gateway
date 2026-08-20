/**
 * 响应拦截器内核。
 * JsonData（含 HTTP 4xx/5xx）resolve 给页面；401 弹窗且不落到业务；
 * 仅取消、超时、断网、非 JsonData 响应才 reject。
 * 后续要改结果判定改本类，不要写回 request.ts。
 */
import { HttpCancelledError, HttpTransportError } from '@/api/requestError'
import { requestHeaderHelper } from '@/api/requestHeader'
import { requestLoadingHelper } from '@/api/requestLoading'
import { requestPluginHelper } from '@/api/requestPlugin'
import type { InternalRequestConfig } from '@/api/requestTypes'
import { config } from '@/config/config'
import type { JsonDataObj } from '@/types/api'
import { rsConfirm } from '@/ui'
import type { AxiosError, AxiosInstance, AxiosResponse } from 'axios'
import axios from 'axios'

/**
 * 响应拦截器辅助类。JsonData 契约只在本类实现，插件不得改返回值。
 */
export class ResponseInterceptorHelper {
  /** 多个 401 同时返回时只弹一次登录框 */
  private is401DialogShowing = false

  /**
   * 把响应拦截器挂到 axios 实例。应用生命周期内调用一次。
   * @param http - 全局 axios 实例
   */
  attach(http: AxiosInstance): void {
    http.interceptors.response.use(
      (response: AxiosResponse) => {
        requestLoadingHelper.end((response.config as InternalRequestConfig).showLoading)
        requestPluginHelper.notifyFulfilled(response)
        // blob 要带 headers（Content-Disposition），不能只回 data
        if (response.config.responseType === 'blob') {
          return response
        }
        return response.data
      },
      (error: AxiosError) => {
        requestLoadingHelper.end((error.config as InternalRequestConfig | undefined)?.showLoading)
        requestPluginHelper.notifyRejected(error)

        if (axios.isCancel(error)) {
          return Promise.reject(new HttpCancelledError())
        }

        const status = error.response?.status
        const data = error.response?.data
        const reqConfig = error.config as InternalRequestConfig | undefined
        const requestId = requestHeaderHelper.readRequestId(reqConfig?.headers)

        if (status === 401) {
          return this.handle401Error(this.readBizErrMsg(data, '未授权，请重新登录'))
        }

        // axios 把非 2xx 当异常；body 已是 JsonData 则转回业务失败，不进 catch
        if (error.response && this.isJsonDataBody(data)) {
          return this.toBizFailure(data, this.statusFallback(status ?? 0))
        }

        // 无 JsonData：超时、断网、网关 HTML 等，只能 reject
        let message = '网络异常，请稍后重试'
        if (error.response) {
          message = this.statusFallback(status ?? 0)
        } else if (error.request) {
          message = '服务器无响应，请检查网络连接'
        }

        const transportError = new HttpTransportError({
          message,
          code: status ?? -1,
          data,
          requestId,
          module: reqConfig?.meta?.module,
          url: reqConfig?.url,
        })

        if (import.meta.env.DEV) {
          console.warn('[http]', message, {
            requestId: transportError.requestId,
            module: transportError.module,
            url: transportError.url,
            status,
          })
        }

        return Promise.reject(transportError)
      },
    )
  }

  /**
   * 判断响应体是否为后端标准 JsonData（有 oK / errMsg / messageId 之一）。
   * @param data - 响应体
   * @returns 是否为 JsonData 形态
   */
  private isJsonDataBody(data: unknown): data is Partial<JsonDataObj> {
    return (
      !!data &&
      typeof data === 'object' &&
      ('oK' in data || 'errMsg' in data || 'messageId' in data)
    )
  }

  /**
   * 优先 errMsg，其次 popMsg；都空则用 HTTP 状态兜底文案。
   * @param data - 响应体
   * @param fallback - 都空时的兜底
   * @returns 错误文案
   */
  private readBizErrMsg(data: unknown, fallback: string): string {
    if (!data || typeof data !== 'object') {
      return fallback
    }
    const body = data as { errMsg?: unknown; popMsg?: unknown }
    if (typeof body.errMsg === 'string' && body.errMsg.trim()) {
      return body.errMsg
    }
    if (typeof body.popMsg === 'string' && body.popMsg.trim()) {
      return body.popMsg
    }
    return fallback
  }

  /**
   * 把带业务包的失败响应收成 oK=false，页面走 else + getApiMessage。
   * @param data - 响应体
   * @param fallback - 无 errMsg 时的兜底
   * @returns 业务失败包
   */
  private toBizFailure(data: unknown, fallback: string): JsonDataObj {
    const body = this.isJsonDataBody(data) ? data : {}
    return {
      oK: false,
      state: false,
      bizData: typeof body.bizData === 'string' ? body.bizData : '',
      extObj: body.extObj ?? null,
      pageQueryData: typeof body.pageQueryData === 'string' ? body.pageQueryData : '',
      messageId: typeof body.messageId === 'string' ? body.messageId : '',
      errMsg: this.readBizErrMsg(body, fallback),
      popMsg: typeof body.popMsg === 'string' ? body.popMsg : '',
      extMsg: typeof body.extMsg === 'string' ? body.extMsg : '',
      pkey1: typeof body.pkey1 === 'string' ? body.pkey1 : '',
      pkey2: typeof body.pkey2 === 'string' ? body.pkey2 : '',
      pkey3: typeof body.pkey3 === 'string' ? body.pkey3 : '',
      pkey4: typeof body.pkey4 === 'string' ? body.pkey4 : '',
      pkey5: typeof body.pkey5 === 'string' ? body.pkey5 : '',
      pkey6: typeof body.pkey6 === 'string' ? body.pkey6 : '',
    }
  }

  /**
   * HTTP 状态无业务文案时的兜底提示。
   * @param status - HTTP 状态码
   * @returns 提示文案
   */
  private statusFallback(status: number): string {
    switch (status) {
      case 400:
        return '请求参数错误'
      case 403:
        return '没有执行此操作的权限'
      case 404:
        return '请求资源不存在'
      case 500:
        return '服务器内部错误'
      default:
        return `请求失败(${status})`
    }
  }

  /**
   * 401：清会话、弹窗跳登录。返回 pending Promise，避免业务层再弹一次。
   * @param message - 弹窗说明
   * @returns 永不 settle 的 Promise
   */
  private handle401Error(message: string): Promise<never> {
    // 已有弹窗时吞掉后续 401，await 一直挂起，页面即将跳走
    if (this.is401DialogShowing) {
      return new Promise(() => {})
    }
    this.is401DialogShowing = true
    const userStore = useUserStore()
    userStore.clearPersistedSession()
    const loginPath = this.getLoginPath()
    void rsConfirm.error({
      title: '认证失败',
      description: message,
      confirmText: '确定',
      onConfirm: () => {
        this.is401DialogShowing = false
        window.location.href = loginPath
      },
      onCancel: () => {
        this.is401DialogShowing = false
      },
    })
    // 不 resolve、不 reject，拦住业务层 catch / else
    return new Promise(() => {})
  }

  /**
   * 拼登录页地址，与前端路由 base 对齐。
   * @returns 登录页完整路径
   */
  private getLoginPath(): string {
    const baseUrl = config.baseUrl.endsWith('/') ? config.baseUrl.slice(0, -1) : config.baseUrl
    return `${baseUrl}/login`
  }
}

/** 全局单例，应用启动时挂一次。 */
export const responseInterceptorHelper = new ResponseInterceptorHelper()
