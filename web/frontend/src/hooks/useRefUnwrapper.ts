/**
 * Ref解包工具Hook
 * 用于将带有ref属性的对象转换为直接包含值的对象，简化模板中的使用
 */
import { unref } from 'vue'
import type { Ref, ComputedRef } from 'vue'

type RefValue<T> = T extends Ref<infer R> ? R : T extends ComputedRef<infer R> ? R : T

/**
 * 解包所有ref属性
 * 将一个包含多个ref的对象转换为包含普通值的对象
 *
 * @param obj 包含多个ref的对象
 * @returns 一个新对象，其中所有的ref属性都被解包为普通值
 */
export function unwrapRefs<T extends Record<string, any>>(
  obj: T,
): { [K in keyof T]: RefValue<T[K]> } {
  const result = {} as { [K in keyof T]: RefValue<T[K]> }

  for (const key in obj) {
    Object.defineProperty(result, key, {
      enumerable: true,
      get: () => unref(obj[key]),
    })
  }

  return result
}

/**
 * 使用解包后的refs
 * 将一个异步状态对象(如useAsync返回值)转换为更易在模板中使用的形式
 * 这个函数返回一个普通对象，访问属性时会自动解包ref
 *
 * @param obj 原始对象，包含多个ref属性
 * @returns 一个新对象，所有属性都会自动解包ref值
 *
 * @example
 * const asyncState = useAsync(fetchData)
 * const { loading, data, error } = useUnwrappedRefs(asyncState)
 *
 * // 在模板中直接使用，无需.value
 * <div v-if="loading">Loading...</div>
 * <div v-else-if="error">{{ error.message }}</div>
 * <div v-else>{{ data }}</div>
 */
export default function useUnwrappedRefs<T extends Record<string, any>>(obj: T) {
  return unwrapRefs(obj)
}
