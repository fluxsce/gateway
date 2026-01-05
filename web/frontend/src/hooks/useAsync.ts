/**
 * 通用异步操作Hook
 * 为Vue组件提供异步操作能力，简化异步状态管理
 * 适用于各种异步操作，比请求hook更通用
 */
import { ref, computed, shallowRef } from 'vue'
import type { Ref, ComputedRef } from 'vue'

/**
 * 异步操作的状态接口
 */
export interface AsyncState<T> {
  /**
   * 加载状态 - 表示异步操作是否正在进行中
   */
  loading: Ref<boolean>

  /**
   * 异步操作结果数据
   */
  data: Ref<T | undefined>

  /**
   * 异步操作错误
   */
  error: Ref<Error | null>

  /**
   * 是否成功 - 计算属性，表示异步操作是否已成功完成
   */
  success: ComputedRef<boolean>

  /**
   * 是否已执行过 - 计算属性，表示异步操作是否已经执行过（无论成功或失败）
   */
  executed: ComputedRef<boolean>
}

/**
 * 异步操作的执行方法接口
 */
export interface AsyncExecutor<TResult, TParams extends any[]> {
  /**
   * 执行异步操作
   * @param args 传递给异步函数的参数
   */
  execute: (...args: TParams) => Promise<TResult>

  /**
   * 重置异步操作的状态
   */
  reset: () => void
}

/**
 * 异步操作组合接口
 * 包含状态和执行方法
 */
export type AsyncComposed<TResult, TParams extends any[]> = AsyncState<TResult> &
  AsyncExecutor<TResult, TParams>

/**
 * 通用异步操作Hook
 *
 * 为异步函数提供状态管理，适用于任何异步操作
 * 比useRequest更轻量和通用，但功能相对简单
 *
 * @param asyncFn 异步函数，可以是任何返回Promise的函数
 * @param initialValue 初始值，可选
 * @returns 包含状态和控制方法的对象
 *
 * @example
 * // 基本用法
 * const { execute, loading, data, error } = useAsync(fetchUserProfile)
 *
 * // 带参数调用
 * const { execute } = useAsync((id: string) => fetchUserById(id))
 * execute('user-123')
 *
 * // 带初始值
 * const { data } = useAsync(fetchUserList, [])
 */
export default function useAsync<TResult = any, TParams extends any[] = any[]>(
  asyncFn: (...args: TParams) => Promise<TResult>,
  initialValue?: TResult,
): AsyncComposed<TResult, TParams> {
  // 创建响应式状态
  const loading = ref(false)
  const error = ref<Error | null>(null)

  // 使用shallowRef优化性能，特别是当结果是大型对象或数组时
  const data = shallowRef<TResult | undefined>(initialValue)

  // 计算属性
  const success = computed(() => !loading.value && !error.value && data.value !== undefined)
  const executed = computed(() => error.value !== null || data.value !== undefined)

  /**
   * 执行异步函数
   */
  const execute = async (...args: TParams): Promise<TResult> => {
    loading.value = true
    error.value = null

    try {
      const result = await asyncFn(...args)
      data.value = result
      return result
    } catch (e) {
      error.value = e as Error
      return Promise.reject(e)
    } finally {
      loading.value = false
    }
  }

  /**
   * 重置状态
   */
  const reset = () => {
    loading.value = false
    error.value = null
    data.value = initialValue
  }

  return {
    loading,
    data,
    error,
    success,
    executed,
    execute,
    reset,
  }
}

/**
 * 使用异步操作并立即执行
 * useAsync的变体，用于创建时立即执行的异步操作
 *
 * @param asyncFn 异步函数，可以是任何返回Promise的函数
 * @param args 传递给异步函数的参数
 * @param initialValue 初始值，可选
 * @returns 包含状态和控制方法的对象
 *
 * @example
 * // 创建时立即执行
 * const { data, loading, error } = useAsyncImmediate(fetchUserProfile)
 *
 * // 带参数立即执行
 * const { data } = useAsyncImmediate(fetchUserById, ['user-123'])
 */
export function useAsyncImmediate<TResult = any, TParams extends any[] = any[]>(
  asyncFn: (...args: TParams) => Promise<TResult>,
  args: TParams = [] as unknown as TParams,
  initialValue?: TResult,
): AsyncComposed<TResult, TParams> {
  const async = useAsync<TResult, TParams>(asyncFn, initialValue)

  // 立即执行异步函数
  async.execute(...args)

  return async
}
