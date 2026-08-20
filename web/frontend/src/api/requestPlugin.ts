/**
 * HTTP 插件。
 * 请求插件只能改发出去的 config；响应插件只旁路通知，不能改 JsonData 返回值。
 */
import type { AxiosError, AxiosResponse } from 'axios'
import type { InternalRequestConfig } from '@/api/requestTypes'

/**
 * 请求侧插件。只能改发出去的 config（头、meta），不能改 JsonData 契约。
 * 埋点、额外头、实验性能力用 registerHttpPlugin 接入，避免改拦截器内核。
 */
export interface HttpRequestPlugin {
  /** 插件名，重复注册会被忽略 */
  name: string
  /**
   * 在发出请求前改写配置。
   * @param config - 当前请求的内部配置
   */
  apply(config: InternalRequestConfig): void | Promise<void>
}

/**
 * 响应侧插件。在内核处理之后旁路执行，不能改返回值，也不能把 JsonData 失败改成 throw。
 * 用于日志、埋点等扩展。
 */
export interface HttpResponsePlugin {
  /** 插件名，重复注册会被忽略 */
  name: string
  /**
   * HTTP 有响应体时的旁路（含 2xx）。
   * @param response - axios 原始响应
   */
  onFulfilled?(response: AxiosResponse): void
  /**
   * axios 把非 2xx / 网络错误当异常时的旁路。内核仍按 JsonData / 401 处理。
   * @param error - axios 错误
   */
  onRejected?(error: AxiosError): void
}

/**
 * HTTP 插件注册表。
 * 请求插件在发出前执行；响应插件只旁路通知，不参与 JsonData 判定。
 */
export class RequestPluginHelper {
  private readonly requestPlugins: HttpRequestPlugin[] = []
  private readonly responsePlugins: HttpResponsePlugin[] = []

  /**
   * 注册请求插件。同名只保留第一次，保证 boot 可重复调用。
   * @param plugin - 请求侧插件
   */
  registerRequest(plugin: HttpRequestPlugin): void {
    if (this.requestPlugins.some((item) => item.name === plugin.name)) {
      return
    }
    this.requestPlugins.push(plugin)
  }

  /**
   * 注册响应插件。同名只保留第一次。
   * @param plugin - 响应侧插件
   */
  registerResponse(plugin: HttpResponsePlugin): void {
    if (this.responsePlugins.some((item) => item.name === plugin.name)) {
      return
    }
    this.responsePlugins.push(plugin)
  }

  /**
   * 依次执行请求插件。
   * @param config - 当前请求配置
   */
  async applyRequest(config: InternalRequestConfig): Promise<void> {
    for (const plugin of this.requestPlugins) {
      await plugin.apply(config)
    }
  }

  /**
   * 通知响应成功插件。单个插件异常不影响内核。
   * @param response - axios 响应
   */
  notifyFulfilled(response: AxiosResponse): void {
    for (const plugin of this.responsePlugins) {
      try {
        plugin.onFulfilled?.(response)
      } catch (error) {
        if (import.meta.env.DEV) {
          console.warn('[http plugin]', plugin.name, error)
        }
      }
    }
  }

  /**
   * 通知响应失败插件。单个插件异常不影响内核。
   * @param error - axios 错误
   */
  notifyRejected(error: AxiosError): void {
    for (const plugin of this.responsePlugins) {
      try {
        plugin.onRejected?.(error)
      } catch (pluginError) {
        if (import.meta.env.DEV) {
          console.warn('[http plugin]', plugin.name, pluginError)
        }
      }
    }
  }
}

/** 全局单例，供拦截器与 boot 注册共用。 */
export const requestPluginHelper = new RequestPluginHelper()

/**
 * 注册请求插件。同名只保留第一次，保证 boot 可重复调用。
 * @param plugin - 请求侧插件
 */
export function registerHttpPlugin(plugin: HttpRequestPlugin): void {
  requestPluginHelper.registerRequest(plugin)
}

/**
 * 注册响应插件。只做旁路，不能改 JsonData 契约。
 * @param plugin - 响应侧插件
 */
export function registerHttpResponsePlugin(plugin: HttpResponsePlugin): void {
  requestPluginHelper.registerResponse(plugin)
}
