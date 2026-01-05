/**
 * 路由异步数据加载Hook
 * 用于根据路由参数自动加载数据，支持参数变化时重新加载
 */
import { ref, watch, onMounted, unref } from 'vue'
import { useRoute, onBeforeRouteUpdate } from 'vue-router'
import useAsync, { type AsyncState } from './useAsync'
import useUnwrappedRefs from './useRefUnwrapper'

/**
 * useRouteAsync返回值类型
 * 包含异步状态和刷新方法
 */
export interface RouteAsyncResult<T> {
  /** 加载状态 */
  loading: boolean
  /** 数据结果 */
  data: T | undefined
  /** 错误对象 */
  error: Error | null
  /** 是否成功 */
  success: boolean
  /** 是否已执行 */
  executed: boolean
  /** 执行异步操作 */
  execute: () => Promise<T>
  /** 重置状态 */
  reset: () => void
  /** 手动刷新数据 */
  refresh: () => Promise<T>
}

/**
 * 基于路由的异步数据加载
 * 在组件挂载或路由参数变化时自动加载数据
 *
 * @param fetchFn 获取数据的函数，接收路由参数作为参数
 * @param options 配置选项
 * @returns 包含数据和状态的对象
 *
 * @example
 * // 基本用法 - 自动根据路由参数加载数据
 * const { data, loading, error } = useRouteAsync(
 *   (params) => fetchUserById(params.id)
 * )
 *
 * // 指定要监听的参数，仅在这些参数变化时重新加载
 * const { data, loading } = useRouteAsync(
 *   (params) => fetchArticleList(params.category, params.page),
 *   { watchParams: ['category', 'page'] }
 * )
 *
 * // 提供初始值
 * const { data } = useRouteAsync(
 *   (params) => fetchComments(params.postId),
 *   { initialValue: [] }
 * )
 */
export default function useRouteAsync<T = any>(
  fetchFn: (params: Record<string, string>) => Promise<T>,
  options: {
    /** 初始值 */
    initialValue?: T
    /** 要监听的参数名称数组，仅在这些参数变化时才重新加载 */
    watchParams?: string[]
    /** 是否在组件挂载时自动加载数据，默认为true */
    immediate?: boolean
    /** 是否监听路由更新，默认为true */
    watchRoute?: boolean
  } = {},
): RouteAsyncResult<T> {
  const { initialValue, watchParams = [], immediate = true, watchRoute = true } = options

  // 获取当前路由
  const route = useRoute()

  // 当前路由参数
  const routeParams = ref<Record<string, string>>({})

  // 更新路由参数
  const updateRouteParams = () => {
    const params: Record<string, string> = {}

    // 合并路由参数、查询参数和动态参数
    Object.assign(params, route.params, route.query)

    // 转换所有值为字符串，以便比较
    Object.keys(params).forEach((key) => {
      if (params[key] !== null && params[key] !== undefined) {
        params[key] = String(params[key])
      }
    })

    routeParams.value = params
  }

  // 创建异步状态
  const asyncState = useAsync<T, []>(async () => fetchFn(routeParams.value), initialValue)

  // 加载数据
  const loadData = async (): Promise<T> => {
    updateRouteParams()
    return asyncState.execute()
  }

  // 在组件挂载时加载数据
  onMounted(() => {
    if (immediate) {
      loadData()
    }
  })

  // 监听指定的路由参数变化
  if (watchRoute && watchParams.length > 0) {
    watch(
      () =>
        watchParams.map((param) => {
          // 同时监听params和query中的参数
          return {
            param,
            value: route.params[param] || route.query[param],
          }
        }),
      () => loadData(),
      { deep: true },
    )
  }
  // 监听整个路由对象变化
  else if (watchRoute) {
    watch(
      () => [route.path, route.query, route.params],
      () => loadData(),
      { deep: true },
    )
  }

  // 在路由更新前重新加载数据
  if (watchRoute) {
    onBeforeRouteUpdate((to, from, next) => {
      // 仅当有指定的参数变化时才重新加载
      if (watchParams.length > 0) {
        const shouldReload = watchParams.some((param) => {
          const oldValue = from.params[param] || from.query[param]
          const newValue = to.params[param] || to.query[param]
          return oldValue !== newValue
        })

        if (shouldReload) {
          // 更新参数并加载数据，但不阻塞路由切换
          const newParams: Record<string, string> = {}
          Object.assign(newParams, to.params, to.query)
          routeParams.value = newParams
          asyncState.execute().catch((error) => {
            console.error('路由数据加载失败:', error)
          })
        }
      } else {
        // 如果没有指定参数，任何路由更新都会触发重新加载
        const newParams: Record<string, string> = {}
        Object.assign(newParams, to.params, to.query)
        routeParams.value = newParams
        asyncState.execute().catch((error) => {
          console.error('路由数据加载失败:', error)
        })
      }

      next()
    })
  }

  // 返回解包后的状态和refresh方法
  const unwrappedState = useUnwrappedRefs(asyncState)

  return {
    ...unwrappedState,
    data: unwrappedState.data,
    refresh: loadData,
  }
}
