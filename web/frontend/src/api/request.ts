/**
 * API请求封装
 * 基于axios封装的HTTP请求工具，统一处理请求和响应拦截、错误处理等
 *
 * 使用方式:
 * 1. 推荐使用统一的request函数:
 *    import { request } from '@/api/request'
 *    request('GET', '/api/users', null, { page: 1 })
 *
 * 2. 或者使用对象方式:
 *    import { request } from '@/api/request'
 *    request({
 *      method: 'POST',
 *      url: '/api/users',
 *      data: { name: '张三' }
 *    })
 *
 * 3. 为了向后兼容，仍可使用方法别名:
 *    import { get, post } from '@/api/request'
 *    get('/api/users', { page: 1 })
 *    post('/api/users', { name: '张三' })
 */
import { config } from '@/config/config'
import type {
  AxiosError,
  AxiosRequestConfig,
  AxiosResponse,
  InternalAxiosRequestConfig,
} from 'axios'
import axios from 'axios'

/**
 * 请求方法类型
 */
export type RequestMethod = 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH' | 'HEAD' | 'OPTIONS'

/**
 * 扩展的请求配置接口
 */
export interface ExtendedRequestConfig extends AxiosRequestConfig {
  showLoading?: boolean // 是否显示加载状态
}

/**
 * 扩展的内部请求配置接口
 */
interface ExtendedInternalAxiosRequestConfig extends InternalAxiosRequestConfig {
  showLoading?: boolean // 是否显示加载状态
}

// 创建axios实例
const service = axios.create({
  // 从环境变量获取API基础URL，如果不存在则使用默认值
  baseURL: import.meta.env.VITE_API_BASE_URL,
  // 请求超时时间（毫秒）
  timeout: 15000,
  // 自动携带cookie，支持跨域请求时的认证
  withCredentials: true,
  // 默认请求头
  headers: {
    'Content-Type': 'application/x-www-form-urlencoded; charset=UTF-8',
  },
})

// 全局loading计数器
let loadingCount = 0
// Naive UI的loadingBar实例
let loadingBar: any = null
// Naive UI的dialog实例
let dialog: any = null
// 401弹窗防重复标志
let is401DialogShowing = false

/**
 * 动态更新axios实例的超时时间
 * @param timeout 新的超时时间（毫秒）
 */
export const updateTimeout = (timeout: number) => {
  if (timeout > 0) {
    service.defaults.timeout = timeout
    if (import.meta.env.DEV) {
      console.log(`API请求超时时间已更新为: ${timeout}ms`)
    }
  }
}

/**
 * 获取当前axios实例的超时时间
 * @returns 当前超时时间（毫秒）
 */
export const getCurrentTimeout = (): number => {
  return service.defaults.timeout || 15000
}

/**
 * 设置Naive UI的loading bar和dialog
 * @param naiveUILoadingBar Naive UI的loading bar实例
 * @param naiveUIDialog Naive UI的dialog实例
 */
export const setNaiveUILoading = (naiveUILoadingBar: any, naiveUIDialog?: any) => {
  loadingBar = naiveUILoadingBar
  if (naiveUIDialog) {
    dialog = naiveUIDialog
  }
}

/**
 * 显示加载状态
 */
const showLoading = () => {
  if (loadingCount === 0 && loadingBar) {
    loadingBar.start()
  }
}

/**
 * 隐藏加载状态
 */
const hideLoading = () => {
  if (loadingCount === 0 && loadingBar) {
    loadingBar.finish()
  }
}

/**
 * 获取登录页路径
 * @returns 完整的登录页路径
 */
const getLoginPath = (): string => {
  const baseUrl = config.baseUrl.endsWith('/') 
    ? config.baseUrl.slice(0, -1) 
    : config.baseUrl
  return `${baseUrl}/login`
}

/**
 * 统一处理401认证失败
 * @param message 错误消息
 * @returns Promise 永远不会resolve，阻止错误传播到业务层
 */
const handle401Error = (message: string) => {
  // 如果已经在显示401弹窗，则返回一个永远不会resolve的Promise
  if (is401DialogShowing) {
    return new Promise(() => {}) // 返回永远pending的Promise，阻止业务层处理
  }

  is401DialogShowing = true

  // 清理用户状态
  const userStore = useUserStore()
  userStore.clearUserInfo()

  const loginPath = getLoginPath()

  if (dialog) {
    dialog.warning({
      title: '认证失败',
      content: message,
      positiveText: '确定',
      onPositiveClick: () => {
        is401DialogShowing = false
        // 用户点击确定后跳转到登录页
        window.location.href = loginPath
      },
      onClose: () => {
        is401DialogShowing = false
      },
    })
  } else {
    // 如果没有dialog实例，则使用原生alert
    alert(message)
    is401DialogShowing = false
    window.location.href = loginPath
  }

  // 返回永远不会resolve的Promise，这样业务层永远不会收到响应
  // 避免业务层弹出错误提示
  return new Promise(() => {})
}

/**
 * 请求拦截器
 * 在请求发送前处理请求配置，例如添加统一的请求头
 */
service.interceptors.request.use(
  (config: ExtendedInternalAxiosRequestConfig) => {
    // 使用cookie管理会话，不需要手动添加token到请求头
    // cookie会通过withCredentials: true自动携带

    // 为所有请求类型默认启用loading状态，除非明确设置为false
    if (config.showLoading === undefined) {
      config.showLoading = true
    }

    // 如果配置了显示loading，则显示全局loading
    if (config.showLoading) {
      showLoading()
    }

    return config
  },
  (error) => {
    // 请求错误处理
    const config = error.config as ExtendedInternalAxiosRequestConfig
    if (config?.showLoading) {
      hideLoading()
    }
    return Promise.reject(error)
  },
)

/**
 * 响应拦截器
 * 在接收到响应后统一处理响应数据和错误
 */
service.interceptors.response.use(
  (response: AxiosResponse) => {
    // 如果请求配置了显示loading，则在响应后隐藏loading
    const config = response.config as ExtendedInternalAxiosRequestConfig
    if (config.showLoading) {
      hideLoading()
    }

    const res = response.data

    // 根据自定义错误码判断请求是否成功
    // 处理业务逻辑错误码
    if (response !== undefined && response.status !== 0 && response.status !== 200) {
      // 处理token过期或未授权的情况
      if (response.status === 401) {
        // 统一处理401错误，阻止重复弹窗
        return handle401Error(res.message || '会话已失效，请重新登录')
      }

      // 对于业务错误，直接返回响应数据，让业务层处理
      // 避免抛出异常导致无法在业务层获取到原始错误信息
      return res
    }

    // 直接返回响应数据，简化业务层的数据获取
    return res
  },
  (error: AxiosError) => {
    // 如果请求配置了显示loading，则在错误响应后隐藏loading
    const config = error.config as ExtendedInternalAxiosRequestConfig
    if (config?.showLoading) {
      hideLoading()
    }

    // 处理不同HTTP状态码的错误
    let message = '网络异常，请稍后重试'
    let code = -1

    if (error.response) {
      const status = error.response.status
      code = status

      switch (status) {
        case 400:
          message = '请求参数错误'
          break
        case 401:
          message = '未授权，请重新登录'
          // 统一处理401错误，阻止重复弹窗
          return handle401Error(message)
        case 403:
          message = '拒绝访问'
          break
        case 404:
          message = '请求资源不存在'
          break
        case 500:
          message = '服务器内部错误'
          break
        default:
          message = `请求失败(${status})`
      }
    } else if (error.request) {
      // 请求已发送但没有收到响应
      message = '服务器无响应，请检查网络连接'
    } else {
      // 请求配置错误
      message = '请求配置错误'
    }

    // 如果是取消请求产生的错误，不显示错误信息
    if (axios.isCancel(error)) {
      message = '请求已取消'
      return Promise.reject({ code, message, cancelled: true })
    }

    // 开发环境下输出错误日志
    if (import.meta.env.DEV) {
      console.warn('API请求错误:', message, error)
    }

    // 返回统一的错误对象格式，便于业务层处理
    return Promise.reject({
      code,
      message,
      data: (error as AxiosError).response?.data,
      originalError: error,
    })
  },
)

/**
 * 请求参数接口
 */
export interface RequestOptions<D = any, P = any> extends ExtendedRequestConfig {
  method: RequestMethod
  url: string
  data?: D
  params?: P
}

/**
 * 统一请求函数
 * 支持两种调用方式:
 * 1. 传入独立参数: request('GET', '/api/users', undefined, { page: 1 })
 * 2. 传入配置对象: request({ method: 'GET', url: '/api/users', params: { page: 1 } })
 *
 * @param methodOrOptions 请求方法或完整的请求配置对象
 * @param url 请求地址 (当第一个参数为请求方法时使用)
 * @param data 请求数据 (当第一个参数为请求方法时使用)
 * @param params 查询参数 (当第一个参数为请求方法时使用)
 * @param config 配置选项 (当第一个参数为请求方法时使用)
 * @returns Promise 包含响应数据的Promise
 *
 * @example
 * // 方式1: 传入独立参数
 * request('GET', '/api/users', undefined, { page: 1, size: 10 })
 *
 * // 方式2: 传入配置对象
 * request({
 *   method: 'POST',
 *   url: '/api/users',
 *   data: { name: '张三', age: 30 },
 *   showLoading: true
 * })
 */
export function request<T = any>(
  methodOrOptions: RequestMethod | RequestOptions,
  url?: string,
  data?: any,
  params?: any,
  config?: ExtendedRequestConfig,
): Promise<T> {
  // 检查第一个参数是否为对象，判断调用方式
  if (typeof methodOrOptions === 'object') {
    // 方式2: 使用配置对象
    const options = methodOrOptions as RequestOptions
    const { method, url, data, params, ...restConfig } = options

    // 构建请求配置
    const requestConfig: ExtendedRequestConfig = {
      method,
      url,
      showLoading: true,
      ...restConfig,
    }

    // 根据请求方法设置不同的数据
    if (method === 'GET' || method === 'DELETE' || method === 'HEAD' || method === 'OPTIONS') {
      requestConfig.params = params
    } else {
      requestConfig.data = data
      requestConfig.params = params
    }

    return service(requestConfig) as Promise<T>
  } else {
    // 方式1: 使用独立参数
    const method = methodOrOptions

    const requestConfig: ExtendedRequestConfig = {
      method,
      url: url!,
      showLoading: true,
      ...config,
    }

    // 根据请求方法设置不同的数据
    if (method === 'GET' || method === 'DELETE' || method === 'HEAD' || method === 'OPTIONS') {
      requestConfig.params = { ...params, ...config?.params }
    } else {
      requestConfig.data = data
      requestConfig.params = params
    }

    return service(requestConfig) as Promise<T>
  }
}

/**
 * 封装GET请求
 * 用于获取数据
 *
 * @param url 请求地址 - 接口URL路径
 * @param params 请求参数 - 将会转换为查询字符串附加到URL
 * @param config 配置选项 - axios请求配置，可以设置showLoading: true来显示加载状态
 * @returns Promise 包含响应数据的Promise
 */
export const get = <T = any>(
  url: string,
  params?: any,
  config?: ExtendedRequestConfig,
): Promise<T> => {
  return request<T>('GET', url, undefined, params, config)
}

/**
 * 封装POST请求
 * 用于创建数据
 *
 * @param url 请求地址 - 接口URL路径
 * @param data 请求数据 - 请求体数据，通常是一个对象
 * @param params 查询参数 - URL参数
 * @param config 配置选项 - axios请求配置，可以设置showLoading: false来隐藏加载状态
 * @returns Promise 包含响应数据的Promise
 */
export const post = <T = any>(
  url: string,
  data?: any,
  params?: any,
  config?: ExtendedRequestConfig,
): Promise<T> => {
  return request<T>('POST', url, data, params, config)
}

/**
 * 封装PUT请求
 * 用于更新数据
 *
 * @param url 请求地址 - 接口URL路径
 * @param data 请求数据 - 更新的数据内容
 * @param config 配置选项 - axios请求配置，可以设置showLoading: false来隐藏加载状态
 * @returns Promise 包含响应数据的Promise
 */
export const put = <T = any>(
  url: string,
  data?: any,
  config?: ExtendedRequestConfig,
): Promise<T> => {
  return request<T>('PUT', url, data, undefined, config)
}

/**
 * 封装DELETE请求
 * 用于删除数据
 *
 * @param url 请求地址 - 接口URL路径
 * @param params 请求参数 - 通常用于条件删除
 * @param config 配置选项 - axios请求配置，可以设置showLoading: false来隐藏加载状态
 * @returns Promise 包含响应数据的Promise
 */
export const del = <T = any>(
  url: string,
  params?: any,
  config?: ExtendedRequestConfig,
): Promise<T> => {
  return request<T>('DELETE', url, undefined, params, config)
}

/**
 * 封装PATCH请求
 * 用于部分更新数据
 *
 * @param url 请求地址 - 接口URL路径
 * @param data 请求数据 - 需要更新的数据字段
 * @param config 配置选项 - axios请求配置
 * @returns Promise 包含响应数据的Promise
 */
export const patch = <T = any>(
  url: string,
  data?: any,
  config?: ExtendedRequestConfig,
): Promise<T> => {
  return request<T>('PATCH', url, data, undefined, config)
}

/**
 * 封装HEAD请求
 * 用于获取资源头信息
 *
 * @param url 请求地址 - 接口URL路径
 * @param params 请求参数
 * @param config 配置选项 - axios请求配置
 * @returns Promise
 */
export const head = <T = any>(
  url: string,
  params?: any,
  config?: ExtendedRequestConfig,
): Promise<T> => {
  return request<T>('HEAD', url, undefined, params, config)
}

/**
 * 封装OPTIONS请求
 * 用于获取资源支持的通信选项
 *
 * @param url 请求地址 - 接口URL路径
 * @param params 请求参数
 * @param config 配置选项 - axios请求配置
 * @returns Promise
 */
export const options = <T = any>(
  url: string,
  params?: any,
  config?: ExtendedRequestConfig,
): Promise<T> => {
  return request<T>('OPTIONS', url, undefined, params, config)
}

/**
 * 初始化请求工具
 * 设置loading bar和dialog
 *
 * @param loadingBarInstance Naive UI的loading bar实例
 * @param dialogInstance Naive UI的dialog实例
 */
export const initRequestTools = (loadingBarInstance: any, dialogInstance?: any) => {
  loadingBar = loadingBarInstance
  if (dialogInstance) {
    dialog = dialogInstance
  }

  // 输出日志确认加载功能已初始化
  if (import.meta.env.DEV) {
    console.log('API请求工具已初始化，加载状态显示功能已启用')
  }
}

/**
 * 创建API接口
 * 用于创建预配置的API端点，简化重复调用
 *
 * @param baseURL 基础URL路径
 * @param defaultConfig 默认配置选项
 * @returns 预配置的请求函数对象
 *
 * @example
 * // 创建用户API
 * const userApi = createApi('/api/users')
 *
 * // 使用预配置的API
 * userApi.get() // GET /api/users
 * userApi.get('123') // GET /api/users/123
 * userApi.post({ name: '张三' }) // POST /api/users
 * userApi.put('123', { name: '李四' }) // PUT /api/users/123
 * userApi.delete('123') // DELETE /api/users/123
 */
export const createApi = (baseURL: string, defaultConfig: ExtendedRequestConfig = {}) => {
  const apiRequest = <T = any>(
    method: RequestMethod,
    path: string = '',
    data?: any,
    params?: any,
    config?: ExtendedRequestConfig,
  ): Promise<T> => {
    const url = path ? `${baseURL}/${path}`.replace(/\/+/g, '/') : baseURL
    return request<T>(method, url, data, params, { ...defaultConfig, ...config })
  }

  return {
    // 基础请求方法
    request: <T = any>(
      method: RequestMethod,
      path?: string,
      data?: any,
      params?: any,
      config?: ExtendedRequestConfig,
    ) => apiRequest<T>(method, path, data, params, config),

    // 常用HTTP方法
    get: <T = any>(path?: string, params?: any, config?: ExtendedRequestConfig) =>
      apiRequest<T>('GET', path, undefined, params, config),
    post: <T = any>(path?: string, data?: any, params?: any, config?: ExtendedRequestConfig) =>
      apiRequest<T>('POST', path, data, params, config),
    put: <T = any>(path?: string, data?: any, config?: ExtendedRequestConfig) =>
      apiRequest<T>('PUT', path, data, undefined, config),
    delete: <T = any>(path?: string, params?: any, config?: ExtendedRequestConfig) =>
      apiRequest<T>('DELETE', path, undefined, params, config),
    patch: <T = any>(path?: string, data?: any, config?: ExtendedRequestConfig) =>
      apiRequest<T>('PATCH', path, data, undefined, config),

    // 获取构造的URL（不发送请求）
    getURL: (path: string = '') => (path ? `${baseURL}/${path}`.replace(/\/+/g, '/') : baseURL),
  }
}

export default service
