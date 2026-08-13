/**
 * 静态路由入口
 *
 * - 登录、404 等常量路由仍在此声明。
 * - 主布局下的业务子路由由 `layoutRouteRegistry.buildMainLayoutChildRoutes()` 生成（注册表逻辑分组，路由表拍平为 MainLayout 子级）。
 * - 侧边栏菜单由同注册表 `buildSidebarMenuFromRegistry()` 派生，勿再双写。
 */
import { buildMainLayoutChildRoutes } from '@/router/layoutRouteRegistry'
import MainLayout from '@/views/layout/MainLayout.vue'
import type { RouteRecordRaw } from 'vue-router'

/**
 * 仅开发环境注册，生产打包不包含测试路由及对应 chunk。
 * 侧栏菜单由 {@link buildSidebarMenuFromRegistry} 生成，本段不在 GATEWAY_LAYOUT_ROUTE_TREE 中，
 * 故在 layoutRouteRegistry 内对 DEV 单独追加「组件测试」入口；此处 `menuHide` 仅作 meta 说明（不参与侧栏合并）。
 */
const testRoutes: RouteRecordRaw[] =
  import.meta.env.DEV
    ? [
        {
          path: 'test',
          name: 'test',
          redirect: '/test/index',
          component: () => import('@/views/test/TestLayout.vue'),
          meta: {
            title: '组件测试',
            requiresAuth: true,
            icon: 'flask-conical',
            menuHide: true,
            keepAliveOutletName: 'TestLayout',
          },
          children: [
            {
              path: 'index',
              name: 'testIndex',
              component: () => import('@/views/test/TestIndex.vue'),
              meta: { title: '组件测试中心', requiresAuth: true },
            },
            {
              path: 'gtoolbar',
              name: 'testGToolbar',
              component: () => import('@/views/test/components/GToolbarTest.vue'),
              meta: { title: 'GToolbar 测试', requiresAuth: true },
            },
            {
              path: 'rs-search-form',
              name: 'testRsSearchForm',
              component: () => import('@/views/test/components/RsSearchFormTest.vue'),
              meta: { title: 'RsSearchForm 测试', requiresAuth: true },
            },
            {
              path: 'rs-data-form',
              name: 'testRsDataForm',
              component: () => import('@/views/test/components/RsDataFormTest.vue'),
              meta: { title: 'RsDataForm 测试', requiresAuth: true },
            },
            {
              path: 'rs-grid',
              name: 'testRsGrid',
              component: () => import('@/views/test/components/RsGridTest.vue'),
              meta: { title: 'RsGrid 测试', requiresAuth: true },
            },
            {
              path: 'rs-button',
              name: 'testRsButton',
              component: () => import('@/views/test/components/RsButtonTest.vue'),
              meta: { title: 'RsButton 测试', requiresAuth: true },
            },
            {
              path: 'gcard',
              name: 'testGCard',
              component: () => import('@/views/test/components/GCardTest.vue'),
              meta: { title: 'GCard 测试', requiresAuth: true },
            },
          ],
        },
      ]
    : []

export class StaticRoutes {
  static readonly constantRoutes: RouteRecordRaw[] = [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/hub0001/LoginView.vue'),
      meta: {
        title: '用户登录',
        requiresAuth: false,
        moduleName: 'hub0001',
      },
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'not-found',
      component: () => import('@/views/layout/NotFound.vue'),
      meta: {
        title: '页面未找到',
        requiresAuth: false,
      },
    },
  ]

  static readonly layoutRoute: RouteRecordRaw = {
    path: '/',
    name: 'mainLayout',
    component: MainLayout,
    meta: {
      requiresAuth: true,
    },
    children: buildMainLayoutChildRoutes(),
  }

  static getLayoutRoute(): RouteRecordRaw {
    return {
      ...this.layoutRoute,
      children: [...(this.layoutRoute.children as RouteRecordRaw[]), ...testRoutes],
    }
  }

  static getRoutes(): RouteRecordRaw[] {
    return [this.getLayoutRoute(), ...this.constantRoutes]
  }
}

export default StaticRoutes
