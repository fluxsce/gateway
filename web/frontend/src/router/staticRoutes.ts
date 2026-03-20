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

/** 仅开发环境注册，生产打包不包含测试路由及对应 chunk */
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
            icon: 'FlaskOutline',
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
              path: 'message',
              name: 'testMessage',
              component: () => import('@/views/test/components/MessageTest.vue'),
              meta: { title: 'Message 测试', requiresAuth: true },
            },
            {
              path: 'custom-render',
              name: 'testCustomRender',
              component: () => import('@/views/test/components/CustomRenderTest.vue'),
              meta: { title: '自定义渲染测试', requiresAuth: true },
            },
            {
              path: 'gtabs',
              name: 'testGTabs',
              component: () => import('@/views/test/components/GTabsTest.vue'),
              meta: { title: 'GTabs 测试', requiresAuth: true },
            },
            {
              path: 'gtext-show',
              name: 'testGTextShow',
              component: () => import('@/views/test/components/GTextShowTest.vue'),
              meta: { title: 'GTextShow 测试', requiresAuth: true },
            },
            {
              path: 'gdropdown',
              name: 'testGDropdown',
              component: () => import('@/views/test/components/GDropdownTest.vue'),
              meta: { title: 'GDropdown 测试', requiresAuth: true },
            },
            {
              path: 'gcard',
              name: 'testGCard',
              component: () => import('@/views/test/components/GCardTest.vue'),
              meta: { title: 'GCard 测试', requiresAuth: true },
            },
            {
              path: 'gselect',
              name: 'testGSelect',
              component: () => import('@/views/test/components/GSelectTest.vue'),
              meta: { title: 'GSelect 测试', requiresAuth: true },
            },
            {
              path: 'gdialog',
              name: 'testGDialog',
              component: () => import('@/views/test/components/GDialogTest.vue'),
              meta: { title: 'GDialog 测试', requiresAuth: true },
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
