/**
 * 路由导航守卫模块
 * 处理路由导航过程中的权限检查、页面标题设置、面包屑生成和多语言资源预加载
 */
import { getCurrentLocale, loadModuleMessages } from '@/locales'
import { useGlobalStore } from '@/stores/global'
import { useUserStore } from '@/stores/user'
import type { Router } from 'vue-router'

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
        console.log(`[路由守卫] 已预加载模块 "${to.meta.moduleName}" 的语言包: ${currentLocale}`)
      } catch (error) {
        console.warn(`[路由守卫] 预加载模块 "${to.meta.moduleName}" 的语言包失败:`, error)
        // 预加载失败不阻止路由导航，只是记录警告
      }
    }

    /**
     * 设置页面标题
     * 将路由meta中的title与应用名称结合设置为文档标题
     */
    const appTitle = import.meta.env.VITE_APP_TITLE || 'Web Hub Here'
    document.title = `${to.meta.title || '页面'} - ${appTitle}`

    /**
     * 设置导航面包屑
     * 根据当前路由路径构建面包屑导航数据
     */
    if (to.meta.title) {
      // 构建面包屑数据
      const breadcrumbs = []

      // 添加首页
      breadcrumbs.push({ title: '首页', path: '/' })

      // 如果不是首页，添加当前页面
      if (to.path !== '/dashboard') {
        // 如果有父级路由，添加父级
        if (to.matched.length > 2) {
          const parent = to.matched[1]
          if (parent.meta.title) {
            breadcrumbs.push({
              title: parent.meta.title as string,
              path: parent.path,
            })
          }
        }

        // 添加当前页面
        breadcrumbs.push({
          title: to.meta.title as string,
          path: to.path,
        })
      }

      // 设置面包屑和页面标题
      globalStore.setBreadcrumbs(breadcrumbs)
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
     * 权限检查
     * 检查用户是否拥有访问路由所需的权限
     */
    if (
      to.meta.permissions &&
      Array.isArray(to.meta.permissions) &&
      to.meta.permissions.length > 0
    ) {
      const requiredPermissions = to.meta.permissions as string[]
      const hasPermission = requiredPermissions.some((perm) => userStore.hasPermission(perm))

      // 如果既没有所需权限也不是管理员，则重定向到仪表盘
      if (!hasPermission && userStore.tenantAdminFlag !== 'Y') {
        return next({ name: 'dashboard' })
      }
    }

    /**
     * 已登录用户访问登录页检查
     * 已登录用户尝试访问登录页时重定向到首页
     */
    if (to.name === 'login' && userStore.isAuthenticated) {
      return next({ path: '/' })
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
