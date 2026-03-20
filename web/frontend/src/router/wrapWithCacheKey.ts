/**
 * 与 XiRang `xirang-cached-view/use-cached-view` 一致：
 * KeepAlive 的 `include` 按「打开页签的路径」匹配，而不是页面组件的 PascalCase 名。
 *
 * 通过包装组件将 `name` 设为 `cacheKey`（一般为 `route.fullPath` / `tabId`），
 * 关页签从 include 移除该 path 即可卸载对应缓存实例。
 */
import { defineComponent, h, type Component } from 'vue'

const wrappedComponentCache = new Map<string, Component>()

export function wrapWithCacheKey(component: Component, cacheKey: string): Component {
  if (!wrappedComponentCache.has(cacheKey)) {
    const wrapped = defineComponent({
      name: cacheKey,
      render() {
        return h(component as never, this.$attrs as Record<string, unknown>)
      },
    })
    wrappedComponentCache.set(cacheKey, wrapped)
  }
  return wrappedComponentCache.get(cacheKey)!
}

/** 关闭页签后删除包装器引用，避免 Map 泄漏 */
export function cleanupWrappedCache(activeKeys: Set<string>): void {
  for (const key of wrappedComponentCache.keys()) {
    if (!activeKeys.has(key)) {
      wrappedComponentCache.delete(key)
    }
  }
}
