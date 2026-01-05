/**
 * 系统相关API的Mock数据
 * 模拟菜单、权限等系统级接口
 */
import type { MockMethod } from 'vite-plugin-mock'
import type { JsonDataObj, MenuItem, PermissionItem } from '@/types/api'
import Mock from 'mockjs'

/**
 * 请求处理函数参数接口
 */
interface RequestParams {
  url: string
  body: Record<string, any>
  query: Record<string, string>
  headers: Record<string, string>
  method: string
}

/**
 * 创建JsonDataObj格式的响应数据
 */
function createJsonDataResponse<T>(data: T, success = true, message = ''): JsonDataObj {
  if (success) {
    return {
      oK: true,
      state: true,
      bizData: JSON.stringify(data),
      extObj: null,
      pageQueryData: '',
      messageId: '',
      errMsg: '',
      popMsg: message || '操作成功',
      extMsg: '',
      pkey1: '',
      pkey2: '',
      pkey3: '',
      pkey4: '',
      pkey5: '',
      pkey6: '',
    }
  } else {
    return {
      oK: false,
      state: false,
      bizData: '',
      extObj: null,
      pageQueryData: '',
      messageId: '',
      errMsg: message || '操作失败',
      popMsg: message || '操作失败',
      extMsg: '',
      pkey1: '',
      pkey2: '',
      pkey3: '',
      pkey4: '',
      pkey5: '',
      pkey6: '',
    }
  }
}

// 模拟菜单数据 - 基于HUB_MENU表结构和MenuItem接口
const menus: MenuItem[] = [
  {
    menuId: 'M001',
    menuName: '仪表盘',
    parentId: undefined,
    menuPath: '/dashboard',
    component: 'views/dashboard/index',
    icon: 'dashboard',
    sortOrder: 1,
    menuType: 2, // 菜单
    permCode: 'system:dashboard:view',
    visibleFlag: 'Y',
    statusFlag: 'Y',
    sysMenuFlag: 'Y',
    addTime: '2023-01-01 00:00:00',
    addWho: 'SYSTEM',
    editTime: '2023-01-01 00:00:00',
    editWho: 'SYSTEM',
    oprSeqFlag: Mock.Random.guid(),
    currentVersion: 1,
    activeFlag: 'Y',
    languageId: 'zh-CN',
    noteText: '系统仪表盘',
    children: [],
  },
  {
    menuId: 'M002',
    menuName: '系统管理',
    parentId: undefined,
    menuPath: '/system',
    component: 'Layout',
    icon: 'setting',
    sortOrder: 2,
    menuType: 1, // 目录
    permCode: undefined,
    visibleFlag: 'Y',
    statusFlag: 'Y',
    sysMenuFlag: 'Y',
    addTime: '2023-01-01 00:00:00',
    addWho: 'SYSTEM',
    editTime: '2023-01-01 00:00:00',
    editWho: 'SYSTEM',
    oprSeqFlag: Mock.Random.guid(),
    currentVersion: 1,
    activeFlag: 'Y',
    languageId: 'zh-CN',
    noteText: '系统管理模块',
    children: [
      {
        menuId: 'M003',
        menuName: '用户管理',
        parentId: 'M002',
        menuPath: '/system/user',
        component: 'views/system/user/index',
        icon: 'user',
        sortOrder: 1,
        menuType: 2, // 菜单
        permCode: 'system:user:view',
        visibleFlag: 'Y',
        statusFlag: 'Y',
        sysMenuFlag: 'Y',
        addTime: '2023-01-01 00:00:00',
        addWho: 'SYSTEM',
        editTime: '2023-01-01 00:00:00',
        editWho: 'SYSTEM',
        oprSeqFlag: Mock.Random.guid(),
        currentVersion: 1,
        activeFlag: 'Y',
        languageId: 'zh-CN',
        noteText: '用户管理页面',
        children: [
          {
            menuId: 'M007',
            menuName: '添加用户',
            parentId: 'M003',
            menuPath: '',
            component: '',
            icon: '',
            sortOrder: 1,
            menuType: 3, // 按钮
            permCode: 'system:user:add',
            visibleFlag: 'Y',
            statusFlag: 'Y',
            sysMenuFlag: 'Y',
            addTime: '2023-01-01 00:00:00',
            addWho: 'SYSTEM',
            editTime: '2023-01-01 00:00:00',
            editWho: 'SYSTEM',
            oprSeqFlag: Mock.Random.guid(),
            currentVersion: 1,
            activeFlag: 'Y',
            languageId: 'zh-CN',
            noteText: '添加用户权限',
            children: [],
          },
          {
            menuId: 'M008',
            menuName: '编辑用户',
            parentId: 'M003',
            menuPath: '',
            component: '',
            icon: '',
            sortOrder: 2,
            menuType: 3, // 按钮
            permCode: 'system:user:edit',
            visibleFlag: 'Y',
            statusFlag: 'Y',
            sysMenuFlag: 'Y',
            addTime: '2023-01-01 00:00:00',
            addWho: 'SYSTEM',
            editTime: '2023-01-01 00:00:00',
            editWho: 'SYSTEM',
            oprSeqFlag: Mock.Random.guid(),
            currentVersion: 1,
            activeFlag: 'Y',
            languageId: 'zh-CN',
            noteText: '编辑用户权限',
            children: [],
          },
        ],
      },
      {
        menuId: 'M004',
        menuName: '角色管理',
        parentId: 'M002',
        menuPath: '/system/role',
        component: 'views/system/role/index',
        icon: 'team',
        sortOrder: 2,
        menuType: 2, // 菜单
        permCode: 'system:role:view',
        visibleFlag: 'Y',
        statusFlag: 'Y',
        sysMenuFlag: 'Y',
        addTime: '2023-01-01 00:00:00',
        addWho: 'SYSTEM',
        editTime: '2023-01-01 00:00:00',
        editWho: 'SYSTEM',
        oprSeqFlag: Mock.Random.guid(),
        currentVersion: 1,
        activeFlag: 'Y',
        languageId: 'zh-CN',
        noteText: '角色管理页面',
        children: [],
      },
      {
        menuId: 'M005',
        menuName: '权限管理',
        parentId: 'M002',
        menuPath: '/system/permission',
        component: 'views/system/permission/index',
        icon: 'safety',
        sortOrder: 3,
        menuType: 2, // 菜单
        permCode: 'system:permission:view',
        visibleFlag: 'Y',
        statusFlag: 'Y',
        sysMenuFlag: 'Y',
        addTime: '2023-01-01 00:00:00',
        addWho: 'SYSTEM',
        editTime: '2023-01-01 00:00:00',
        editWho: 'SYSTEM',
        oprSeqFlag: Mock.Random.guid(),
        currentVersion: 1,
        activeFlag: 'Y',
        languageId: 'zh-CN',
        noteText: '权限管理页面',
        children: [],
      },
    ],
  },
  {
    menuId: 'M006',
    menuName: '用户中心',
    parentId: undefined,
    menuPath: '/account',
    component: 'Layout',
    icon: 'user',
    sortOrder: 3,
    menuType: 1, // 目录
    permCode: undefined,
    visibleFlag: 'Y',
    statusFlag: 'Y',
    sysMenuFlag: 'Y',
    addTime: '2023-01-01 00:00:00',
    addWho: 'SYSTEM',
    editTime: '2023-01-01 00:00:00',
    editWho: 'SYSTEM',
    oprSeqFlag: Mock.Random.guid(),
    currentVersion: 1,
    activeFlag: 'Y',
    languageId: 'zh-CN',
    noteText: '用户个人中心',
    children: [
      {
        menuId: 'M009',
        menuName: '个人信息',
        parentId: 'M006',
        menuPath: '/account/profile',
        component: 'views/account/profile/index',
        icon: 'profile',
        sortOrder: 1,
        menuType: 2, // 菜单
        permCode: 'system:profile:view',
        visibleFlag: 'Y',
        statusFlag: 'Y',
        sysMenuFlag: 'Y',
        addTime: '2023-01-01 00:00:00',
        addWho: 'SYSTEM',
        editTime: '2023-01-01 00:00:00',
        editWho: 'SYSTEM',
        oprSeqFlag: Mock.Random.guid(),
        currentVersion: 1,
        activeFlag: 'Y',
        languageId: 'zh-CN',
        noteText: '个人信息页面',
        children: [],
      },
      {
        menuId: 'M010',
        menuName: '修改密码',
        parentId: 'M006',
        menuPath: '/account/change-password',
        component: 'views/account/changePassword/index',
        icon: 'lock',
        sortOrder: 2,
        menuType: 2, // 菜单
        permCode: 'system:password:change',
        visibleFlag: 'Y',
        statusFlag: 'Y',
        sysMenuFlag: 'Y',
        addTime: '2023-01-01 00:00:00',
        addWho: 'SYSTEM',
        editTime: '2023-01-01 00:00:00',
        editWho: 'SYSTEM',
        oprSeqFlag: Mock.Random.guid(),
        currentVersion: 1,
        activeFlag: 'Y',
        languageId: 'zh-CN',
        noteText: '修改密码页面',
        children: [],
      },
    ],
  },
]

// 模拟权限数据 - 基于HUB_PERMISSION表结构和PermissionItem接口
const permissions: PermissionItem[] = [
  {
    permCode: 'system:dashboard:view',
    permName: '查看仪表盘',
    menuId: 'M001',
    permType: 1, // 菜单权限
    statusFlag: 'Y',
    addTime: '2023-01-01 00:00:00',
    addWho: 'SYSTEM',
    editTime: '2023-01-01 00:00:00',
    editWho: 'SYSTEM',
    oprSeqFlag: Mock.Random.guid(),
    currentVersion: 1,
    activeFlag: 'Y',
    noteText: '允许查看系统仪表盘',
  },
  {
    permCode: 'system:user:view',
    permName: '查看用户列表',
    menuId: 'M003',
    permType: 1, // 菜单权限
    statusFlag: 'Y',
    addTime: '2023-01-01 00:00:00',
    addWho: 'SYSTEM',
    editTime: '2023-01-01 00:00:00',
    editWho: 'SYSTEM',
    oprSeqFlag: Mock.Random.guid(),
    currentVersion: 1,
    activeFlag: 'Y',
    noteText: '允许查看用户管理页面',
  },
  {
    permCode: 'system:user:add',
    permName: '新增用户',
    menuId: 'M007',
    permType: 2, // 按钮权限
    statusFlag: 'Y',
    addTime: '2023-01-01 00:00:00',
    addWho: 'SYSTEM',
    editTime: '2023-01-01 00:00:00',
    editWho: 'SYSTEM',
    oprSeqFlag: Mock.Random.guid(),
    currentVersion: 1,
    activeFlag: 'Y',
    noteText: '允许新增用户',
  },
  {
    permCode: 'system:user:edit',
    permName: '编辑用户',
    menuId: 'M008',
    permType: 2, // 按钮权限
    statusFlag: 'Y',
    addTime: '2023-01-01 00:00:00',
    addWho: 'SYSTEM',
    editTime: '2023-01-01 00:00:00',
    editWho: 'SYSTEM',
    oprSeqFlag: Mock.Random.guid(),
    currentVersion: 1,
    activeFlag: 'Y',
    noteText: '允许编辑用户信息',
  },
  {
    permCode: 'system:user:delete',
    permName: '删除用户',
    menuId: 'M003',
    permType: 2, // 按钮权限
    statusFlag: 'Y',
    addTime: '2023-01-01 00:00:00',
    addWho: 'SYSTEM',
    editTime: '2023-01-01 00:00:00',
    editWho: 'SYSTEM',
    oprSeqFlag: Mock.Random.guid(),
    currentVersion: 1,
    activeFlag: 'Y',
    noteText: '允许删除用户',
  },
  {
    permCode: 'system:role:view',
    permName: '查看角色列表',
    menuId: 'M004',
    permType: 1, // 菜单权限
    statusFlag: 'Y',
    addTime: '2023-01-01 00:00:00',
    addWho: 'SYSTEM',
    editTime: '2023-01-01 00:00:00',
    editWho: 'SYSTEM',
    oprSeqFlag: Mock.Random.guid(),
    currentVersion: 1,
    activeFlag: 'Y',
    noteText: '允许查看角色管理页面',
  },
  {
    permCode: 'system:permission:view',
    permName: '查看权限列表',
    menuId: 'M005',
    permType: 1, // 菜单权限
    statusFlag: 'Y',
    addTime: '2023-01-01 00:00:00',
    addWho: 'SYSTEM',
    editTime: '2023-01-01 00:00:00',
    editWho: 'SYSTEM',
    oprSeqFlag: Mock.Random.guid(),
    currentVersion: 1,
    activeFlag: 'Y',
    noteText: '允许查看权限管理页面',
  },
  {
    permCode: 'system:profile:view',
    permName: '查看个人信息',
    menuId: 'M009',
    permType: 1, // 菜单权限
    statusFlag: 'Y',
    addTime: '2023-01-01 00:00:00',
    addWho: 'SYSTEM',
    editTime: '2023-01-01 00:00:00',
    editWho: 'SYSTEM',
    oprSeqFlag: Mock.Random.guid(),
    currentVersion: 1,
    activeFlag: 'Y',
    noteText: '允许查看个人信息页面',
  },
  {
    permCode: 'system:password:change',
    permName: '修改密码',
    menuId: 'M010',
    permType: 1, // 菜单权限
    statusFlag: 'Y',
    addTime: '2023-01-01 00:00:00',
    addWho: 'SYSTEM',
    editTime: '2023-01-01 00:00:00',
    editWho: 'SYSTEM',
    oprSeqFlag: Mock.Random.guid(),
    currentVersion: 1,
    activeFlag: 'Y',
    noteText: '允许修改个人密码',
  },
]

/**
 * 将权限分配到对应的菜单项中
 * @param menuItems 菜单项数组
 * @param permissions 权限项数组
 */
function attachPermissionsToMenus(menuItems: MenuItem[], permissionItems: PermissionItem[]): void {
  // 为每个菜单找到对应的权限并添加到菜单的permissions属性中
  menuItems.forEach((menu) => {
    // 收集当前菜单的权限
    const menuPermissions = permissionItems.filter((p) => p.menuId === menu.menuId)
    if (menuPermissions.length > 0) {
      menu.permissions = menuPermissions
    }

    // 递归处理子菜单
    if (menu.children && menu.children.length > 0) {
      attachPermissionsToMenus(menu.children, permissionItems)
    }
  })
}

// 将权限附加到对应的菜单
attachPermissionsToMenus(menus, permissions)

// 系统模块相关接口Mock
export default [
  // 获取菜单列表和权限
  {
    url: '/gateway/system/getMenuList',
    method: 'get',
    response: ({ query }: Pick<RequestParams, 'query'>) => {
      // 可选的筛选条件
      const { tenantId, languageId } = query

      let filteredMenus = [...menus]

      // 根据查询条件筛选菜单
      if (languageId) {
        filteredMenus = filteredMenus.filter((menu) => menu.languageId === languageId)
      }

      return createJsonDataResponse(
        {
          menus: filteredMenus,
          permissions: permissions,
        },
        true,
        '获取菜单和权限列表成功',
      )
    },
  },

  // 获取用户权限码列表
  {
    url: '/gateway/system/getUserPermissions',
    method: 'get',
    response: ({ query }: Pick<RequestParams, 'query'>) => {
      const { userId } = query

      // 模拟根据不同用户ID返回不同权限
      let userPermCodes: string[] = []

      if (userId === '1' || !userId) {
        // 管理员或未指定
        // 管理员拥有所有权限
        userPermCodes = permissions.map((p) => p.permCode)
      } else if (userId === '2') {
        // 普通用户
        // 普通用户只有查看权限
        userPermCodes = permissions
          .filter((p) => p.permCode.includes(':view') || p.permCode.includes('profile'))
          .map((p) => p.permCode)
      }

      return createJsonDataResponse(userPermCodes, true, '获取用户权限码列表成功')
    },
  },
] as MockMethod[]
