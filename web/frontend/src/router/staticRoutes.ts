/**
 * 静态路由配置类
 * 管理应用的固定路由结构，包括登录页面、404页面和基础布局路由
 *
 * 静态路由是在应用初始化时直接加载的路由，不依赖于用户权限或后端数据
 *
 * 路由meta配置说明：
 * - title: 页面标题
 * - requiresAuth: 是否需要身份验证
 * - icon: 菜单图标
 * - moduleName: 多语言模块名称，用于路由守卫中预加载对应的语言资源
 */
import MainLayout from '@/views/layout/MainLayout.vue'
import type { RouteRecordRaw } from 'vue-router'

export class StaticRoutes {
  /**
   * 常量路由配置
   * 这些路由对所有用户可见，不受权限控制
   * 包括:
   * - 登录页面
   * - 404页面
   */
  static readonly constantRoutes: RouteRecordRaw[] = [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/hub0001/LoginView.vue'),
      meta: {
        title: '用户登录',
        requiresAuth: false, // 不需要身份验证
        moduleName: 'hub0001', // 多语言模块名称
      },
    },
    {
      path: '/:pathMatch(.*)*', // 通配符路径，匹配所有未定义的路由
      name: 'not-found',
      component: () => import('@/views/layout/NotFound.vue'),
      meta: {
        title: '页面未找到',
        requiresAuth: false, // 不需要身份验证
      },
    },
  ]

  /**
   * 主布局路由配置
   * 作为应用的根路由，包含公共布局组件
   * 所有动态路由将作为此布局的子路由添加
   */
  static readonly layoutRoute: RouteRecordRaw = {
    path: '/',
    name: 'mainLayout',
    component: MainLayout, // 应用主布局组件
    redirect: '/dashboard', // 默认重定向到仪表盘
    meta: {
      requiresAuth: true, // 需要身份验证
    },
    children: [
      {
        path: 'dashboard',
        name: 'dashboard',
        component: () => import('@/views/hub0000/SystemMonitoring.vue'),
        meta: {
          title: '系统监控',
          requiresAuth: true, // 需要身份验证
          icon: 'HomeOutline',
          moduleName: 'hub0000', // 多语言模块名称
        },
      },
      {
        path: 'settings',
        name: 'settings',
        component: () => import('@/views/hub0002/UserSettings.vue'),
        meta: {
          title: '用户设置',
          requiresAuth: true,
          icon: 'SettingsOutline',
          menuHide: true, //菜单隐藏
          moduleName: 'hub0002', // 多语言模块名称
        },
      },
      {
        path: 'system',
        name: 'system',
        component: () => import('@/views/layout/EmptyLayout.vue'),
        meta: {
          title: '系统设置',
          requiresAuth: true,
          icon: 'SettingsOutline',
        },
        children: [
          {
            path: 'userManagement',
            name: 'userManagement',
            component: () => import('@/views/hub0002/UserManagement.vue'),
            meta: {
              title: '用户管理',
              requiresAuth: true,
              icon: 'PeopleOutline',
              moduleName: 'hub0002', // 多语言模块名称
            },
          },
          {
            path: 'roleManagement',
            name: 'roleManagement',
            component: () => import('@/views/hub0005/RoleManagement.vue'),
            meta: {
              title: '角色管理',
              requiresAuth: true,
              icon: 'PeopleCircleOutline',
              moduleName: 'hub0005', // 多语言模块名称
            },
          },
          {
            path: 'resourceManagement',
            name: 'resourceManagement',
            component: () => import('@/views/hub0006/ResourceManagement.vue'),
            meta: {
              title: '权限资源管理',
              requiresAuth: true,
              icon: 'KeyOutline',
              moduleName: 'hub0006', // 多语言模块名称
            },
          },
          // {
          //   path: 'taskManagement',
          //   name: 'taskManagement',
          //   component: () => import('@/views/hub0003/TaskManagement.vue'),
          //   meta: {
          //     title: '定时任务管理',
          //     requiresAuth: true,
          //     icon: 'TimerOutline',
          //     moduleName: 'hub0003', // 多语言模块名称
          //   },
          // },
          // {
          //   path: 'toolManagement',
          //   name: 'toolManagement',
          //   component: () => import('@/views/hub0004/ToolManagement.vue'),
          //   meta: {
          //     title: '工具插件管理',
          //     requiresAuth: true,
          //     icon: 'ExtensionPuzzleOutline',
          //     moduleName: 'hub0004', // 多语言模块名称
          //   },
          // },
          // {
          //   path: 'userRoleAssignment',
          //   name: 'userRoleAssignment',
          //   component: () => import('@/views/hub0005/UserRoleAssignment.vue'),
          //   meta: {
          //     title: '用户角色分配',
          //     requiresAuth: true,
          //     icon: 'PersonAddOutline',
          //     moduleName: 'hub0005', // 多语言模块名称
          //   },
          // },
          // {
          //   path: 'dataPermissionManagement',
          //   name: 'dataPermissionManagement',
          //   component: () => import('@/views/hub0005/DataPermissionManagement.vue'),
          //   meta: {
          //     title: '数据权限管理',
          //     requiresAuth: true,
          //     icon: 'LockClosedOutline',
          //     moduleName: 'hub0005', // 多语言模块名称
          //   },
          // },
          // {
          //   path: 'operationLogManagement',
          //   name: 'operationLogManagement',
          //   component: () => import('@/views/hub0005/OperationLogManagement.vue'),
          //   meta: {
          //     title: '操作日志管理',
          //     requiresAuth: true,
          //     icon: 'DocumentTextOutline',
          //     moduleName: 'hub0005', // 多语言模块名称
          //   },
          // },
        ],
      },
      {
        path: 'gateway',
        name: 'gateway',
        component: () => import('@/views/layout/EmptyLayout.vue'),
        meta: {
          title: '网关管理',
          requiresAuth: true,
          icon: 'CloudOutline',
        },
        children: [
          {
            path: 'gatewayInstanceManager',
            name: 'gatewayInstanceManager',
            component: () => import('@/views/hub0020/GatewayInstanceManager.vue'),
            meta: {
              title: '实例管理',
              requiresAuth: true,
              icon: 'ServerOutline',
              moduleName: 'hub0020', // 多语言模块名称
            },
          },
          {
            path: 'proxyManagement',
            name: 'proxyManagement',
            component: () => import('@/views/hub0022/ProxyManagement.vue'),
            meta: {
              title: '代理管理',
              requiresAuth: true,
              icon: 'FlashOutline',
              moduleName: 'hub0022', // 多语言模块名称
            },
          },
          {
            path: 'routeManagement',
            name: 'routeManagement',
            component: () => import('@/views/hub0021/RouteManagement.vue'),
            meta: {
              title: '路由管理',
              requiresAuth: true,
              icon: 'GitNetworkOutline',
              moduleName: 'hub0021', // 多语言模块名称
            },
          },
          {
            path: 'gatewayLogManagement',
            name: 'gatewayLogManagement',
            component: () => import('@/views/hub0023/GatewayLogManagement.vue'),
            meta: {
              title: '网关日志管理',
              requiresAuth: true,
              icon: 'DocumentTextOutline',
              moduleName: 'hub0023', // 多语言模块名称
            },
          },
        ],
      },
      {
        path: 'tunnel',
        name: 'tunnel',
        component: () => import('@/views/layout/EmptyLayout.vue'),
        meta: {
          title: '隧道管理',
          requiresAuth: true,
          icon: 'SwapHorizontalOutline',
        },
        children: [
          {
            path: 'tunnelServerManagement',
            name: 'tunnelServerManagement',
            component: () => import('@/views/hub0060/TunnelServerManagement.vue'),
            meta: {
              title: '隧道服务器',
              requiresAuth: true,
              icon: 'ServerOutline',
              moduleName: 'hub0060', // 多语言模块名称
            },
          },
          {
            path: 'staticMappingManagement',
            name: 'staticMappingManagement',
            component: () => import('@/views/hub0061/StaticMappingManagement.vue'),
            meta: {
              title: '静态映射',
              requiresAuth: true,
              icon: 'GitNetworkOutline',
              moduleName: 'hub0061', // 多语言模块名称
            },
          },
          {
            path: 'tunnelClientManagement',
            name: 'tunnelClientManagement',
            component: () => import('@/views/hub0062/TunnelClientManagement.vue'),
            meta: {
              title: '隧道客户端',
              requiresAuth: true,
              icon: 'DesktopOutline',
              moduleName: 'hub0062', // 多语言模块名称
            },
          },
        ],
      },
      {
        path: 'serviceGovernance',
        name: 'serviceGovernance',
        component: () => import('@/views/layout/EmptyLayout.vue'),
        meta: {
          title: '服务治理',
          requiresAuth: true,
          icon: 'GitNetworkOutline',
        },
        children: [
          {
            path: 'namespaceManagement',
            name: 'namespaceManagement',
            component: () => import('@/views/hub0040/NamespaceManagement.vue'),
            meta: {
              title: '命名空间管理',
              requiresAuth: true,
              icon: 'LayersOutline',
              moduleName: 'hub0040', // 多语言模块名称
            },
          },
          {
            path: 'serviceRegistryManagement',
            name: 'serviceRegistryManagement',
            component: () => import('@/views/hub0041/ServiceRegistryManagement.vue'),
            meta: {
              title: '服务注册管理',
              requiresAuth: true,
              icon: 'ListOutline',
              moduleName: 'hub0041', // 多语言模块名称
            },
          },
          {
            path: 'serviceMonitoring',
            name: 'serviceMonitoring',
            component: () => import('@/views/hub0042/ServiceMonitoring.vue'),
            meta: {
              title: '服务监控',
              requiresAuth: true,
              icon: 'BarChartOutline',
              moduleName: 'hub0042', // 多语言模块名称
            },
          },
        ],
      },
    ],
  }

  /**
   * 获取所有静态路由配置
   * 用于初始化路由实例
   * @returns 所有静态路由的数组
   */
  static getRoutes(): RouteRecordRaw[] {
    return [this.layoutRoute, ...this.constantRoutes]
  }
}

export default StaticRoutes
