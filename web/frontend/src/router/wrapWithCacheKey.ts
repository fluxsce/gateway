/**
 * 与 XiRang `xirang-cached-view/use-cached-view` 一致：
 * KeepAlive 的 `include` 按「打开页签的路径」匹配，而不是页面组件的 PascalCase 名。
 *
 * 通过包装组件将 `name` 设为 `cacheKey`（一般为 `route.fullPath` / `tabId`），
 * 关页签从 include 移除该 path 即可卸载对应缓存实例。
 *
 * 包装内使用 `Suspense` 包裹页面：Vue 文档要求 `KeepAlive` 在 `Suspense` 外侧；
 * 若反过来包，路由切换时异步页与缓存配合异常，表现为页签切回后状态丢失。
 * 内层用 `shallowRef` 保存页面组件；仅在「可渲染实体」变化时更新 inner。
 *
 * 注意：RouterView 插槽里的 `Component` 常为**新对象引用**，与引用相等比较会在
 * 父组件 render 阶段反复写 shallowRef，触发子更新并造成 RouterView「依赖自身」的
 * 递归更新（Maximum recursive updates exceeded）。因此用稳定身份比较，并把写入推迟
 * 到 post-flush，避免在 render 中同步改依赖。
 */
import RouteViewLoadingMask from '@/components/RouteViewLoadingMask.vue'
import {
  Suspense,
  defineComponent,
  h,
  queuePostFlushCb,
  shallowRef,
  type Component,
} from 'vue'

type WrappedEntry = {
  innerRef: ReturnType<typeof shallowRef<Component>>
  Wrapper: Component
  /** 与 `innerRef` 当前内容对应的稳定身份，用于忽略仅引用变化的重复渲染 */
  lastInnerIdentity: unknown
}

const wrappedComponentCache = new Map<string, WrappedEntry>()

/**
 * 从 RouterView 传入的 `Component` 上取出用于「是否同一页面实现」比较的稳定值。
 * 对 VNode / 异步包装对象取 `type` 或 `__asyncLoader`，避免被外层新对象引用误导。
 */
function stableInnerRenderableIdentity(c: Component): unknown {
  if (c == null) return c
  if (typeof c === 'string' || typeof c === 'number' || typeof c === 'boolean') return c
  if (typeof c === 'function') return c
  if (typeof c === 'object') {
    const o = c as Record<string, unknown>
    if (o.__asyncLoader != null) return o.__asyncLoader
    if (o.type != null) return o.type
  }
  return c
}

export function wrapWithCacheKey(component: Component, cacheKey: string): Component {
  let entry = wrappedComponentCache.get(cacheKey)
  if (!entry) {
    const innerRef = shallowRef(component)
    const Wrapper = defineComponent({
      name: cacheKey,
      inheritAttrs: false,
      setup(_, { attrs }) {
        return () =>
          h(Suspense, null, {
            default: () => h(innerRef.value as never, attrs as Record<string, unknown>),
            fallback: () => h(RouteViewLoadingMask),
          })
      },
    })
    entry = {
      innerRef,
      Wrapper: Wrapper as Component,
      lastInnerIdentity: stableInnerRenderableIdentity(component),
    }
    wrappedComponentCache.set(cacheKey, entry)
    return entry.Wrapper
  }

  const nextId = stableInnerRenderableIdentity(component)
  if (nextId !== entry.lastInnerIdentity) {
    entry.lastInnerIdentity = nextId
    queuePostFlushCb(() => {
      entry!.innerRef.value = component
    })
  }
  return entry.Wrapper
}

/** 关闭页签后删除包装器引用，避免 Map 泄漏 */
export function cleanupWrappedCache(activeKeys: Set<string>): void {
  for (const key of wrappedComponentCache.keys()) {
    if (!activeKeys.has(key)) {
      wrappedComponentCache.delete(key)
    }
  }
}
