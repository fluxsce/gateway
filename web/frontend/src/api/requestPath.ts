/**
 * 请求路径辅助。
 * 负责模块前缀拼接，以及从 /gateway/hubxxxx 读取模块编码。
 * 路径约定变化只改本文件，不要写回 request.ts。
 */
/**
 * 管理端接口前缀，模块编码是其后第一段路径。
 * 与后端 constants.APIRoot 对应（此处带尾斜杠便于切分），改根路径只改这一处。
 */
export const GATEWAY_PREFIX = '/gateway/'

/**
 * 请求路径辅助类。
 * 负责模块前缀拼接，以及从 /gateway/hubxxxx 读取模块编码。
 * 后续路径约定变化只改本类，不要写回 request.ts。
 */
export class RequestPathHelper {
  /**
   * 拼接模块前缀与子路径，避免出现 // 或漏 /。
   * @param baseURL - 模块前缀，如 /gateway/hub0002
   * @param path - 子路径，如 /queryUsers
   * @returns 拼接后的路径
   */
  join(baseURL: string, path = ''): string {
    if (!path) {
      return baseURL
    }
    return `${baseURL.replace(/\/$/, '')}/${path.replace(/^\//, '')}`
  }

  /**
   * 返回模块 API 前缀，与后端 ModuleAPIPrefix 一致，例如 /gateway/hub0007。
   * 登录接口传 user，得到 /gateway/user。
   * @param moduleName - 模块名
   * @returns 模块前缀
   */
  modulePrefix(moduleName: string): string {
    return this.join(GATEWAY_PREFIX, moduleName)
  }

  /**
   * 读取网关模块编码：/gateway/ 后的第一段路径（hub0002、hubcommon002）。
   * 模块身份以 createApi 前缀为准；拦截器只给未带 meta.module 的裸 request 做同样切分。
   * @param url - 请求地址或模块前缀
   * @returns 模块编码，无法识别时为 undefined
   */
  readModule(url?: string): string | undefined {
    if (!url) {
      return undefined
    }
    const path = this.takePathname(url)
    const start = path.indexOf(GATEWAY_PREFIX)
    if (start < 0) {
      return undefined
    }
    const moduleSeg = path.slice(start + GATEWAY_PREFIX.length).split('/')[0]
    if (!moduleSeg.startsWith('hub')) {
      return undefined
    }
    return moduleSeg
  }

  /**
   * 取 URL 的 pathname，去掉 query/hash；绝对地址用 URL 解析。
   * @param url - 原始地址
   * @returns pathname
   */
  takePathname(url: string): string {
    const withoutQuery = url.split('?')[0].split('#')[0]
    if (withoutQuery.includes('://')) {
      try {
        return new URL(withoutQuery).pathname
      } catch {
        return withoutQuery
      }
    }
    return withoutQuery.startsWith('/') ? withoutQuery : `/${withoutQuery}`
  }
}

/** 全局单例，供 createApi 与请求拦截器共用。 */
export const requestPathHelper = new RequestPathHelper()

/**
 * 返回模块 API 前缀，见 RequestPathHelper.modulePrefix。
 * @param moduleName - 模块名，如 hub0002；登录为 user
 * @returns 模块前缀，如 /gateway/hub0002
 */
export function moduleApiPrefix(moduleName: string): string {
  return requestPathHelper.modulePrefix(moduleName)
}
