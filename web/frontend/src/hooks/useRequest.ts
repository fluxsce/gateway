/**
 * 请求数据的Hook
 * 为Vue组件提供数据请求能力，封装加载状态、错误处理和数据管理
 * 基于Vue 3的组合式API设计，支持TypeScript类型推导
 */
import { ref, computed, unref } from 'vue'
import type { Ref, ComputedRef } from 'vue'

/**
 * 请求选项接口
 * 定义了数据请求的配置项
 */
interface RequestOptions<T, P extends any[]> {
  /**
   * 请求函数 - 实际执行数据请求的异步函数
   * 例如API调用函数:
   * () => get('/users')
   * (id) => get(`/users/${id}`)
   */
  requestFn: (...args: P) => Promise<T>

  /**
   * 是否立即执行 - 创建hook时是否立即发起请求
   * 默认值: false (不立即执行)
   */
  immediate?: boolean

  /**
   * 默认值 - 请求完成前或请求失败时使用的默认数据
   */
  defaultValue?: T

  /**
   * 初始参数 - 立即执行请求时使用的参数
   */
  initialParams?: P

  /**
   * 请求前钩子 - 请求发起前的回调函数
   * 可以用于设置加载状态或准备请求参数
   */
  onBefore?: (params: P) => void

  /**
   * 请求成功钩子 - 请求成功后的回调函数
   * 可以用于处理响应数据
   */
  onSuccess?: (data: T, params: P) => void

  /**
   * 请求失败钩子 - 请求失败时的回调函数
   * 可以用于错误处理和日志记录
   */
  onError?: (error: Error, params: P) => void

  /**
   * 请求完成钩子 - 无论成功失败都会调用的回调函数
   * 可以用于清理工作或状态重置
   */
  onFinally?: (params: P, data?: T, error?: Error) => void
}

/**
 * 请求结果接口
 * 定义了useRequest返回的对象结构
 */
interface RequestResult<T, P extends any[]> {
  /**
   * 加载状态 - 表示请求是否正在进行中
   * 可用于显示加载指示器
   */
  loading: Ref<boolean>

  /**
   * 响应数据 - 请求成功后的数据
   * 初始值为undefined或defaultValue
   */
  data: Ref<T | undefined>

  /**
   * 请求错误 - 请求失败的错误对象
   * 请求成功时为undefined
   */
  error: Ref<Error | undefined>

  /**
   * 是否成功 - 计算属性，表示请求是否已成功完成
   * 当loading为false且没有error且data有值时为true
   */
  success: ComputedRef<boolean>

  /**
   * 执行请求函数 - 手动触发请求的方法
   * 接受与requestFn相同的参数
   */
  run: (...args: P) => Promise<T>

  /**
   * 刷新请求 - 使用上一次的参数重新发起请求
   */
  refresh: () => Promise<T>

  /**
   * 取消请求 - 取消正在进行的请求
   * 实际上是将loading状态设为false
   */
  cancel: () => void

  /**
   * 重置状态 - 将所有状态重置为初始值
   */
  reset: () => void
}

/**
 * 请求数据Hook
 * 用于在Vue组件中优雅地处理异步数据请求
 *
 * @param options 请求选项配置
 * @returns 包含请求状态和控制方法的对象
 *
 * @example
 * // 基本用法
 * const { data, loading, error, run } = useRequest({
 *   requestFn: () => get('/api/users')
 * })
 *
 * // 带参数的请求
 * const { data, run } = useRequest({
 *   requestFn: (id: string) => get(`/api/users/${id}`)
 * })
 * // 然后在需要时调用run传入参数
 * run('123')
 *
 * // 立即执行带参数的请求
 * const { data } = useRequest({
 *   requestFn: (id: string) => get(`/api/users/${id}`),
 *   immediate: true,
 *   initialParams: ['123']
 * })
 */
export default function useRequest<T = any, P extends any[] = any[]>(
  options: RequestOptions<T, P>,
): RequestResult<T, P> {
  const {
    requestFn,
    immediate = false,
    defaultValue,
    initialParams = [] as unknown as P,
    onBefore,
    onSuccess,
    onError,
    onFinally,
  } = options

  // 记录上一次的参数，用于refresh功能
  let lastParams = initialParams

  // 创建响应式状态变量
  // 响应数据
  const data = ref<T | undefined>(defaultValue) as Ref<T | undefined>
  // 加载状态
  const loading = ref<boolean>(false)
  // 请求错误
  const error = ref<Error | undefined>(undefined)

  // 是否成功的计算属性
  const success = computed(() => !loading.value && !error.value && data.value !== undefined)

  /**
   * 执行请求函数
   * 发起实际的数据请求，更新状态，并调用相应的钩子函数
   *
   * @param args 传递给requestFn的参数
   * @returns Promise 包含请求结果的Promise
   */
  const run = async (...args: P): Promise<T> => {
    // 取消可能正在进行的请求
    cancel()

    // 记录本次请求的参数，为refresh做准备
    lastParams = args as P

    // 调用请求前钩子
    onBefore?.(args)

    // 设置加载状态
    loading.value = true
    error.value = undefined

    try {
      // 发起请求
      const res = await requestFn(...args)

      // 更新数据
      data.value = res

      // 调用请求成功钩子
      onSuccess?.(res, args)

      return res
    } catch (err) {
      // 处理错误并更新错误状态
      const _error = err as Error
      error.value = _error

      // 调用请求失败钩子
      onError?.(_error, args)

      return Promise.reject(_error)
    } finally {
      // 结束加载状态
      loading.value = false

      // 调用请求完成钩子
      onFinally?.(args, data.value, error.value)
    }
  }

  /**
   * 使用上一次参数刷新请求
   * 用于重新获取数据而不需要再次提供参数
   *
   * @returns Promise 包含请求结果的Promise
   */
  const refresh = (): Promise<T> => {
    return run(...lastParams)
  }

  /**
   * 取消请求
   * 将loading状态设置为false
   * 注意：不会取消实际的网络请求，只是更新状态
   */
  const cancel = (): void => {
    loading.value = false
  }

  /**
   * 重置状态
   * 将所有状态重置为初始值
   */
  const reset = (): void => {
    cancel()
    data.value = defaultValue
    error.value = undefined
  }

  // 如果配置了立即执行，则在创建hook时自动执行请求
  if (immediate) {
    run(...initialParams)
  }

  // 返回包含状态和方法的对象
  return {
    loading,
    data,
    error,
    success,
    run,
    refresh,
    cancel,
    reset,
  }
}
