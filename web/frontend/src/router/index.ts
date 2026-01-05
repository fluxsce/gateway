/**
 * 路由配置模块
 * 定义应用的路由结构和导航守卫
 *
 * 本模块是应用路由系统的核心，整合了静态路由和动态路由，
 * 并提供了路由导航守卫，处理权限验证、页面标题和面包屑导航等功能。
 */
import { useUserStore } from '@/stores/user'
import { createRouter, createWebHistory } from 'vue-router'
import { DynamicRoutes } from './dynamicRoutes'
import { setupRouteGuards } from './routeGuards'
import { StaticRoutes } from './staticRoutes'

/**
 * 创建Vue Router实例
 * 使用HTML5 History模式，初始化时只加载静态路由
 */
const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: StaticRoutes.getRoutes(),
})

/**
 * 动态路由处理器实例
 * 负责从后端获取菜单数据并转换为路由
 */
const dynamicRoutes = new DynamicRoutes(router)

/**
 * 动态路由初始化函数
 * 从后端获取菜单数据，转换为路由，并根据用户权限过滤
 *
 * @returns Promise<boolean> 是否成功加载动态路由
 */
export async function initDynamicRoutes(): Promise<boolean> {
  try {
    const userStore = useUserStore()
    return await dynamicRoutes.initDynamicRoutes(userStore)
  } catch (error) {
    console.error('动态路由初始化失败:', error)
    return false
  }
}

// 设置路由守卫
setupRouteGuards(router)

// 在应用启动时异步初始化动态路由
// 这样可以避免路由守卫中的循环引用问题
setTimeout(async () => {
  try {
    const userStore = useUserStore()
    if (userStore.isAuthenticated) {
      await initDynamicRoutes()
      console.log('动态路由初始化完成')
    }
  } catch (error) {
    console.error('应用启动时初始化动态路由失败:', error)
  }
}, 0)

// 导出静态路由，用于向后兼容
export const constantRoutes = StaticRoutes.constantRoutes
export const layoutRoute = StaticRoutes.layoutRoute

export default router
