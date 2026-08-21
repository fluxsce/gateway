/**
 * 路由导航守卫模块
 * 处理路由导航过程中的权限检查、页面标题设置和多语言资源预加载
 */
import { getCurrentLocale, loadModuleMessages } from '@/locales'
import { GATEWAY_LAYOUT_ROUTE_TREE } from '@/router/layoutRouteRegistry'
import { useGlobalStore } from '@/stores/global'
import { useUserStore } from '@/stores/user'
import type { Router } from 'vue-router'

/** 第一个有模块权限的业务页，供无权限跳转兜底 */
function findFirstAllowedPath(hasModule: (code: string) => boolean): string | null {
  for (const def of GATEWAY_LAYOUT_ROUTE_TREE) {
    if (def.kind === 'leaf') {
      const code = def.meta?.moduleName
      if (code && !def.meta?.menuHide && !def.meta?.permissionExempt && hasModule(code)) {
        return `/${def.path}`
      }
      continue
    }
    for (const child of def.children ?? []) {
      const code = child.meta?.moduleName
      if (code && !child.meta?.menuHide && !child.meta?.permissionExempt && hasModule(code)) {
        return `/${def.path}/${child.path}`
      }
    }
  }
  return null
}

/**
 * 设置路由导航守卫
 * @param router - Vue Router实例
 */
export function setupRouteGuards(router: Router): void {
  /**
   * 全局前置守卫
   * 在导航被确认前调用
   */
  router.beforeEach(async (to, from, next) => {
    // 获取存储
    const userStore = useUserStore()
    const globalStore = useGlobalStore()

    /**
     * 预加载多语言资源
     * 如果路由meta中配置了moduleName，则在进入路由前预加载对应的多语言资源
     * 使用底层loadModuleMessages函数，因为路由守卫不能使用Composition API Hook
     * 
     * 注意：使用 await 确保语言包加载完成后再进入页面，避免页面显示时出现多语言键名
     */
    if (to.meta.moduleName && typeof to.meta.moduleName === 'string') {
      try {
        const currentLocale = getCurrentLocale()
        await loadModuleMessages(to.meta.moduleName, currentLocale)
      } catch {
        // 预加载失败不阻止路由导航
      }
    }

    /**
     * 设置页面标题
     * 将路由meta中的title与应用名称结合设置为文档标题
     */
    const appTitle = import.meta.env.VITE_APP_TITLE || 'Web Hub Here'
    document.title = `${to.meta.title || '页面'} - ${appTitle}`

    if (to.meta.title) {
      globalStore.setPageTitle(to.meta.title as string)
    }

    /**
     * 初始化用户状态（仅首次访问时）
     * 加载用户信息和权限，初始化动态路由
     */
    // 注意：user store 已简化，没有 initialized 标志
    // 初始化逻辑已在应用启动时完成

    /**
     * 身份验证检查
     * 如果路由需要认证但用户未登录，重定向到登录页
     */
    if (to.meta.requiresAuth && !userStore.isAuthenticated) {
      return next({ name: 'login', query: { redirect: to.fullPath } })
    }

    /**
     * 模块权限检查
     * 业务页用 meta.moduleName 对应资源目录 MODULE 码；个人设置等 permissionExempt 页跳过。
     */
    const moduleName = typeof to.meta.moduleName === 'string' ? to.meta.moduleName : ''
    if (to.meta.requiresAuth && moduleName && !to.meta.permissionExempt) {
      if (!userStore.hasPermission(moduleName)) {
        const fallback = findFirstAllowedPath((code) => userStore.hasPermission(code))
        if (fallback && fallback !== to.path) {
          return next(fallback)
        }
        if (to.name !== 'settings') {
          return next({ name: 'settings' })
        }
      }
    }

    /**
     * 已登录用户访问登录页检查
     * 已登录用户尝试访问登录页时重定向到首页
     */
    if (to.name === 'login' && userStore.isAuthenticated) {
      if (userStore.mustChangePwd === 'Y') {
        return next({ name: 'settings', query: { tab: 'password' } })
      }
      return next({ path: '/' })
    }

    if (
      userStore.isAuthenticated &&
      userStore.mustChangePwd === 'Y' &&
      to.name !== 'settings' &&
      to.name !== 'login'
    ) {
      return next({ name: 'settings', query: { tab: 'password' } })
    }

    // 正常导航
    next()
  })

  /**
   * 路由错误处理
   * 捕获组件加载失败等错误，并提供优雅的降级处理
   */
  if (import.meta.env.DEV) {
    router.onError((error) => {
      const failedMatches = router.currentRoute.value.matched
      const failedRoute = failedMatches[failedMatches.length - 1]

      // 组件解析失败时的处理
      if (error.message.includes('Failed to resolve component')) {
        console.error(`路由组件加载失败: ${error.message}`)
        console.error(`路径: ${router.currentRoute.value.path}`)

        if (failedRoute) {
          console.error(`组件: ${failedRoute.path}`)
        }

        // 自动导航到404页面
        router.push({ name: 'not-found' })
      }
    })
  }
}
