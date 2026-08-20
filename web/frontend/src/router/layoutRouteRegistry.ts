/**
 * 主布局业务路由「单一数据源」
 *
 * 参考 XiRang `route-mapper` + `buildHomeChildRoutes`：在此维护 path / name / component / meta，
 * 由 `buildMainLayoutChildRoutes()` 生成**拍平后的** `RouteRecordRaw[]`（MainLayout 下一层子路由，无分组嵌套）；
 * 注册表里 `kind: 'group'` 仅表示侧栏逻辑分组，对应路由为 `groupPath` + `redirect` + 多条 `group/child` 叶子。
 * 侧边栏由 `buildSidebarMenuFromRegistry()` 从同一份数据派生，避免与 staticRoutes 双写。
 */
import type { RouteMeta, RouteRecordRaw } from 'vue-router'

/**
 * 网关侧路由 meta（extends vue-router `RouteMeta`，与 `staticRoutes` / 守卫 / 布局约定一致）
 *
 * | 字段 | 说明 |
 * |------|------|
 * | `title` | 页标题、页签与菜单文案 |
 * | `icon` | 菜单/页签图标名（Lucide kebab-case，供 RsMenu / RsIcon） |
 * | `requiresAuth` | 是否需要登录；主布局下一般为 `true` |
 * | `moduleName` | 业务模块标识（如 hub0002），侧栏与路由守卫按此做模块权限判断 |
 * | `permissionExempt` | 已登录即可访问，不按 moduleName 鉴权（如个人设置） |
 * | `keepAliveIncludeName` | 可选；主布局已用「页签 fullPath + wrapWithCacheKey」对齐缓存时一般不必填 |
 * | `keepAliveOutletName` | 主布局下仍有子层 router-view 时（如 dev `TestLayout`）外层 KeepAlive 的组件名 |
 * | `menuHide` | 整段路由不在侧栏显示（一级或分组下子项均可） |
 * | `hideInMenu` | 子级不在侧栏显示（常用于重定向壳等） |
 *
 * 另可携带 vue-router 官方 `RouteMeta` 扩展字段（如 `transition` 等，视项目而定）。
 */
export interface GatewayAppRouteMeta extends RouteMeta {
  title?: string
  icon?: string
  requiresAuth?: boolean
  moduleName?: string
  /** 已登录即可访问，不检查模块权限 */
  permissionExempt?: boolean
  keepAliveIncludeName?: string
  keepAliveOutletName?: string
  menuHide?: boolean
  hideInMenu?: boolean
}

declare module 'vue-router' {
  interface RouteMeta {
    permissionExempt?: boolean
    moduleName?: string
    requiresAuth?: boolean
  }
}

/**
 * 主布局路由注册表节点（叶子与分组共用同一接口）
 *
 * - `kind: 'leaf'`：须 `component`；一般无 `children`
 * - `kind: 'group'`：须 `children`（子项为 `kind: 'leaf'`）；生成路由时为拍平路径 + 分组 path 上的 `redirect`
 */
export interface GatewayAppRoute {
  kind: 'leaf' | 'group'
  path: string
  name: string
  meta?: GatewayAppRouteMeta
  component?: () => Promise<unknown>
  /** 仅 `kind === 'group'`：子路由 path 相对当前分组 */
  children?: GatewayAppRoute[]
}

/**
 * 主布局业务路由树（顺序即菜单顺序来源）
 */
export const GATEWAY_LAYOUT_ROUTE_TREE: GatewayAppRoute[] = [
  {
    kind: 'leaf',
    path: 'dashboard',
    name: 'dashboard',
    component: () => import('@/views/hub0000/SystemMonitoring.vue'),
    meta: {
      title: '系统监控',
      requiresAuth: true,
      icon: 'layout-dashboard',
      moduleName: 'hub0000',
      keepAliveIncludeName: 'SystemMonitoring',
    },
  },
  {
    kind: 'leaf',
    path: 'settings',
    name: 'settings',
    component: () => import('@/views/hub0002/UserSettings.vue'),
    meta: {
      title: '用户设置',
      requiresAuth: true,
      icon: 'settings',
      menuHide: true,
      moduleName: 'hub0002',
      permissionExempt: true,
      keepAliveIncludeName: 'UserSettings',
    },
  },
  {
    kind: 'group',
    path: 'system',
    name: 'system',
    meta: {
      title: '系统设置',
      requiresAuth: true,
      icon: 'settings',
    },
    children: [
      {
        kind: 'leaf',
        path: 'userManagement',
        name: 'userManagement',
        component: () => import('@/views/hub0002/UserManagement.vue'),
        meta: {
          title: '用户管理',
          requiresAuth: true,
          icon: 'users',
          moduleName: 'hub0002',
        },
      },
      {
        kind: 'leaf',
        path: 'roleManagement',
        name: 'roleManagement',
        component: () => import('@/views/hub0005/RoleManagement.vue'),
        meta: {
          title: '角色管理',
          requiresAuth: true,
          icon: 'users-round',
          moduleName: 'hub0005',
        },
      },
      {
        kind: 'leaf',
        path: 'resourceManagement',
        name: 'resourceManagement',
        component: () => import('@/views/hub0006/ResourceManagement.vue'),
        meta: {
          title: '权限资源管理',
          requiresAuth: true,
          icon: 'key-round',
          moduleName: 'hub0006',
        },
      },
      {
        kind: 'leaf',
        path: 'auditLogManagement',
        name: 'auditLogManagement',
        component: () => import('@/views/hub0004/AuditLogManagement.vue'),
        meta: {
          title: '审计日志',
          requiresAuth: true,
          icon: 'clipboard-list',
          moduleName: 'hub0004',
        },
      },
      {
        kind: 'leaf',
        path: 'serverNodeManagement',
        name: 'serverNodeManagement',
        component: () => import('@/views/hub0007/ServerNodeManagement.vue'),
        meta: {
          title: '系统节点监控',
          requiresAuth: true,
          icon: 'cpu',
          moduleName: 'hub0007',
        },
      },
      {
        kind: 'leaf',
        path: 'clusterEventManagement',
        name: 'clusterEventManagement',
        component: () => import('@/views/hub0008/ClusterEventManagement.vue'),
        meta: {
          title: '集群节点事件',
          requiresAuth: true,
          icon: 'radio',
          moduleName: 'hub0008',
        },
      },
    ],
  },
  {
    kind: 'group',
    path: 'gateway',
    name: 'gateway',
    meta: {
      title: '网关管理',
      requiresAuth: true,
      icon: 'cloud',
    },
    children: [
      {
        kind: 'leaf',
        path: 'gatewayInstanceManager',
        name: 'gatewayInstanceManager',
        component: () => import('@/views/hub0020/GatewayInstanceManager.vue'),
        meta: {
          title: '实例管理',
          requiresAuth: true,
          icon: 'server',
          moduleName: 'hub0020',
        },
      },
      {
        kind: 'leaf',
        path: 'proxyManagement',
        name: 'proxyManagement',
        component: () => import('@/views/hub0022/ProxyManagement.vue'),
        meta: {
          title: '代理管理',
          requiresAuth: true,
          icon: 'zap',
          moduleName: 'hub0022',
        },
      },
      {
        kind: 'leaf',
        path: 'routeManagement',
        name: 'routeManagement',
        component: () => import('@/views/hub0021/RouteManagement.vue'),
        meta: {
          title: '路由管理',
          requiresAuth: true,
          icon: 'network',
          moduleName: 'hub0021',
        },
      },
      {
        kind: 'leaf',
        path: 'gatewayLogManagement',
        name: 'gatewayLogManagement',
        component: () => import('@/views/hub0023/GatewayLogManagement.vue'),
        meta: {
          title: '网关日志管理',
          requiresAuth: true,
          icon: 'file-text',
          moduleName: 'hub0023',
        },
      },
    ],
  },
  {
    kind: 'group',
    path: 'tunnel',
    name: 'tunnel',
    meta: {
      title: '隧道管理',
      requiresAuth: true,
      icon: 'arrow-left-right',
    },
    children: [
      {
        kind: 'leaf',
        path: 'tunnelServerManagement',
        name: 'tunnelServerManagement',
        component: () => import('@/views/hub0060/TunnelServerManagement.vue'),
        meta: {
          title: '隧道服务器',
          requiresAuth: true,
          icon: 'server',
          moduleName: 'hub0060',
        },
      },
      {
        kind: 'leaf',
        path: 'staticMappingManagement',
        name: 'staticMappingManagement',
        component: () => import('@/views/hub0061/StaticMappingManagement.vue'),
        meta: {
          title: '静态映射',
          requiresAuth: true,
          icon: 'network',
          moduleName: 'hub0061',
        },
      },
      {
        kind: 'leaf',
        path: 'tunnelClientManagement',
        name: 'tunnelClientManagement',
        component: () => import('@/views/hub0062/TunnelClientManagement.vue'),
        meta: {
          title: '隧道客户端',
          requiresAuth: true,
          icon: 'monitor',
          moduleName: 'hub0062',
        },
      },
    ],
  },
  {
    kind: 'group',
    path: 'serviceGovernance',
    name: 'serviceGovernance',
    meta: {
      title: '服务治理',
      requiresAuth: true,
      icon: 'git-branch',
    },
    children: [
      {
        kind: 'leaf',
        path: 'serviceCenterInstanceManager',
        name: 'serviceCenterInstanceManager',
        component: () => import('@/views/hub0040/ServiceCenterInstanceManager.vue'),
        meta: {
          title: '服务中心实例管理',
          requiresAuth: true,
          icon: 'server',
          moduleName: 'hub0040',
        },
      },
      {
        kind: 'leaf',
        path: 'namespaceManagement',
        name: 'namespaceManagement',
        component: () => import('@/views/hub0041/NamespaceManagement.vue'),
        meta: {
          title: '命名空间管理',
          requiresAuth: true,
          icon: 'folder',
          moduleName: 'hub0041',
        },
      },
      {
        kind: 'leaf',
        path: 'serviceList',
        name: 'serviceList',
        component: () => import('@/views/hub0042/ServiceList.vue'),
        meta: {
          title: '服务列表',
          requiresAuth: true,
          icon: 'chart-column',
          moduleName: 'hub0042',
        },
      },
      {
        kind: 'leaf',
        path: 'configManagement',
        name: 'configManagement',
        component: () => import('@/views/hub0043/ConfigManagement.vue'),
        meta: {
          title: '配置中心',
          requiresAuth: true,
          icon: 'code-xml',
          moduleName: 'hub0043',
        },
      },
    ],
  },
  {
    kind: 'group',
    path: 'alert',
    name: 'alert',
    meta: {
      title: '预警管理',
      requiresAuth: true,
      icon: 'bell',
    },
    children: [
      {
        kind: 'leaf',
        path: 'alertConfigManagement',
        name: 'alertConfigManagement',
        component: () => import('@/views/hub0080/AlertConfigManagement.vue'),
        meta: {
          title: '预警服务配置',
          requiresAuth: true,
          icon: 'mail',
          moduleName: 'hub0080',
        },
      },
      {
        kind: 'leaf',
        path: 'alertTemplateManagement',
        name: 'alertTemplateManagement',
        component: () => import('@/views/hub0081/AlertTemplateManagement.vue'),
        meta: {
          title: '预警模板管理',
          requiresAuth: true,
          icon: 'notebook-text',
          moduleName: 'hub0081',
        },
      },
      {
        kind: 'leaf',
        path: 'alertLogManagement',
        name: 'alertLogManagement',
        component: () => import('@/views/hub0082/AlertLogManagement.vue'),
        meta: {
          title: '预警日志管理',
          requiresAuth: true,
          icon: 'file-text',
          moduleName: 'hub0082',
        },
      },
    ],
  },
]

/** 由注册表生成 MainLayout 的 children（供 createRouter / StaticRoutes 使用） */
export function buildMainLayoutChildRoutes(): RouteRecordRaw[] {
  const routes: RouteRecordRaw[] = []
  for (const def of GATEWAY_LAYOUT_ROUTE_TREE) {
    if (def.kind === 'leaf') {
      if (!def.component) {
        console.error('[layoutRouteRegistry] leaf route missing component, skipped:', def.name)
        continue
      }
      routes.push({
        path: def.path,
        name: def.name,
        component: def.component,
        meta: def.meta,
      } as RouteRecordRaw)
      continue
    }
    const nested = def.children
    if (!nested?.length) {
      console.error('[layoutRouteRegistry] group route missing children, skipped:', def.name)
      continue
    }
    let firstChildName: string | undefined
    const flatChildRoutes: RouteRecordRaw[] = []
    for (const c of nested) {
      if (!c.component) {
        console.error(
          '[layoutRouteRegistry] group child missing component, skipped:',
          c.name,
          '(group:',
          def.name + ')',
        )
        continue
      }
      if (!firstChildName) firstChildName = c.name
      flatChildRoutes.push({
        path: `${def.path}/${c.path}`,
        name: c.name,
        component: c.component,
        meta: c.meta,
      } as RouteRecordRaw)
    }
    if (flatChildRoutes.length === 0 || !firstChildName) {
      console.error('[layoutRouteRegistry] group has no valid children after validation, skipped:', def.name)
      continue
    }
    routes.push({
      path: def.path,
      name: def.name,
      meta: def.meta,
      redirect: { name: firstChildName },
    } as RouteRecordRaw)
    routes.push(...flatChildRoutes)
  }
  return routes
}

// --- 侧边栏（与注册表同源）---

/** 侧栏节点（不单独 export 类型，由 `buildSidebarMenuFromRegistry` 返回值推断） */
type SidebarMenuNode = {
  key: string
  label: string
  icon: string
  path?: string
  /** 叶子项对应的模块码，用于侧栏权限过滤 */
  moduleName?: string
  children?: SidebarMenuNode[]
}

export function isLayoutMenuGroup(
  node: SidebarMenuNode,
): node is SidebarMenuNode & { children: SidebarMenuNode[] } {
  return Array.isArray(node.children) && node.children.length > 0
}

/** 从 `GATEWAY_LAYOUT_ROUTE_TREE` 派生侧边栏菜单（尊重 menuHide / hideInMenu） */
export function buildSidebarMenuFromRegistry(): SidebarMenuNode[] {
  const menu: SidebarMenuNode[] = []
  for (const def of GATEWAY_LAYOUT_ROUTE_TREE) {
    const topMeta = def.meta
    if (def.kind === 'leaf') {
      if (topMeta?.menuHide || topMeta?.hideInMenu) continue
      menu.push({
        key: def.name,
        label: topMeta?.title ?? String(def.name),
        path: `/${def.path}`,
        icon: topMeta?.icon ?? 'menu',
        moduleName: topMeta?.moduleName,
      })
    } else {
      if (topMeta?.menuHide) continue
      const nested = def.children
      if (!nested?.length) continue
      const children = nested
        .filter((c) => {
          const m = c.meta
          return !m?.menuHide && !m?.hideInMenu
        })
        .map((c) => {
          const m = c.meta
          return {
            key: c.name,
            label: m?.title ?? String(c.name),
            path: `/${def.path}/${c.path}`,
            icon: m?.icon ?? 'circle',
            moduleName: m?.moduleName,
          }
        })
      if (children.length === 0) continue
      menu.push({
        key: def.name,
        label: topMeta?.title ?? String(def.name),
        icon: topMeta?.icon ?? 'menu',
        children,
      })
    }
  }
  // 测试路由仅在 staticRoutes.getLayoutRoute() 中挂载，不在本注册表内；开发环境在侧栏提供入口
  if (import.meta.env.DEV) {
    menu.push({
      key: 'testIndex',
      label: '组件测试',
      path: '/test/index',
      icon: 'flask-conical',
    })
  }
  return menu
}
