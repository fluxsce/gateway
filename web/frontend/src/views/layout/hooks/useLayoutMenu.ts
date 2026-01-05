import { useModuleI18n } from '@/hooks/useModuleI18n'
import { CommonIcons, IconLibrary, renderIconVNode } from '@/utils/icon'
import type { MenuOption } from 'naive-ui'
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

export function useLayoutMenu() {
  const { t: tCommon } = useModuleI18n('common')
  const router = useRouter()
  const route = useRoute()

  // 创建图标渲染函数，直接使用图标工具类（默认使用 NIcon 包裹）
  const createIconRender = (iconName: string) => {
    return renderIconVNode(iconName, undefined, IconLibrary.IONICONS5)
  }

  // 直接从router实例获取路由表生成菜单
  const menuOptions = computed(() => {
    // 获取所有路由
    const routes = router.getRoutes()

    // 获取主布局路由
    const mainLayoutRoute = routes.find((r) => r.path === '/' && r.name === 'mainLayout')

    if (!mainLayoutRoute || !mainLayoutRoute.children) {
      return []
    }

    // 过滤掉不需要在菜单中显示的路由
    return mainLayoutRoute.children
      .filter((route) => !route.meta?.hideInMenu && !route.meta?.menuHide)
      .map((route) => {
        // 获取图标名称
        const iconName = (route.meta?.icon as string) || CommonIcons.MENU

        // 基本菜单项结构
        const menuItem: MenuOption = {
          label: route.meta?.title || String(route.name),
          key: route.name as string, // 使用route.name作为key
          routePath: `/${route.path}`, // 添加一个routePath字段存储实际路径
          icon: createIconRender(iconName),
        }

        // 如果有子路由，也添加为子菜单项
        if (route.children && route.children.length > 0) {
          // 先创建子菜单项
          const childItems = route.children
            .filter((childRoute) => !childRoute.meta?.hideInMenu && !childRoute.meta?.menuHide)
            .map((childRoute) => {
              // 获取子路由图标
              const childIconName = (childRoute.meta?.icon as string) || CommonIcons.MENU

              return {
                label: childRoute.meta?.title || String(childRoute.name),
                key: childRoute.name as string, // 使用childRoute.name作为key
                routePath: `/${route.path === '/' ? '' : route.path}/${childRoute.path}`, // 存储完整路径
                icon: createIconRender(childIconName),
              }
            })

          // 然后通过赋值添加到菜单项
          menuItem.children = childItems
        }

        return menuItem
      })
  })

  // 处理菜单选择
  const handleMenuSelect = (key: string, item: MenuOption) => {
    // 如果点击的是当前路由名称，不进行跳转，防止重复导航
    if (key === route.name) {
      return
    }

    // 直接使用传入的item对象中的routePath
    if (item && (item as any).routePath) {
      const routePath = (item as any).routePath

      // 使用routePath字段进行导航
      router.push(routePath).catch((err) => {
        // 忽略重定向导致的导航取消错误
        if (err.name !== 'NavigationDuplicated') {
          console.error('导航错误:', err)
        }
      })
    }
  }

  // 面包屑导航
  const breadcrumbs = computed(() => {
    // 根据当前路由生成面包屑
    const result = [{ title: tCommon('app.first'), path: '/dashboard' }]

    if (route.meta && route.meta.title) {
      result.push({
        title: route.meta.title as string,
        path: route.path,
      })
    }

    return result
  })

  return {
    menuOptions,
    handleMenuSelect,
    breadcrumbs,
  }
}
