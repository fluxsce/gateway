/**
 * 用户状态管理
 * 
 * 管理用户基本信息、偏好设置和权限信息
 * 
 * @module stores/user
 * @example
 * ```typescript
 * import { store } from '@/stores'
 * 
 * // 检查模块权限
 * if (store.user.hasModule('hub0002')) {
 *   // 用户有用户管理模块权限
 * }
 * 
 * // 检查按钮权限
 * if (store.user.hasButton('hub0002:add')) {
 *   // 用户有新增用户按钮权限
 * }
 * ```
 */
import { updateTimeout } from '@/api/request'
import { defineStore } from 'pinia'

// ==================== 类型定义 ====================

/**
 * 模块权限信息
 * 
 * 表示用户拥有的模块级别权限，用于控制用户是否可以访问某个功能模块
 */
export interface ModulePermission {
  /** 资源唯一标识ID */
  resourceId: string
  /** 资源编码，用于权限检查（如：'hub0002'） */
  resourceCode: string
  /** 资源名称（中文） */
  resourceName: string
  /** 显示名称，用于界面展示 */
  displayName: string
  /** 资源路径，通常是路由路径（如：'/hub0002'） */
  resourcePath: string
  /** 图标样式类名，用于菜单显示 */
  iconClass?: string
  /** 资源描述信息 */
  description?: string
  /** 资源层级，用于菜单层级结构（1为顶级，2为二级，以此类推） */
  resourceLevel: number
  /** 排序顺序，数值越小越靠前 */
  sortOrder: number
  /** 父资源ID，用于构建层级关系 */
  parentResourceId?: string
}

/**
 * 按钮权限信息
 * 
 * 表示用户拥有的按钮级别权限，用于控制用户是否可以执行某个操作（如：新增、编辑、删除等）
 */
export interface ButtonPermission {
  /** 资源唯一标识ID */
  resourceId: string
  /** 资源编码，用于权限检查（如：'hub0002:add'） */
  resourceCode: string
  /** 资源名称（中文） */
  resourceName: string
  /** 显示名称，用于界面展示 */
  displayName: string
  /** 资源路径，通常是API路径（如：'/api/hub0002/add'） */
  resourcePath: string
  /** HTTP请求方法（如：'POST', 'GET', 'PUT', 'DELETE'） */
  resourceMethod?: string
  /** 父资源ID，通常是所属模块的资源ID（如：'hub0002'） */
  parentResourceId?: string
  /** 资源描述信息 */
  description?: string
}

/**
 * 用户权限响应
 * 
 * 后端返回的用户权限数据结构，包含用户拥有的所有模块权限和按钮权限
 */
export interface UserPermissionResponse {
  /** 模块权限列表 */
  modules: ModulePermission[]
  /** 按钮权限列表 */
  buttons: ButtonPermission[]
}

// ==================== Store 定义 ====================

export const useUserStore = defineStore('user', {
  state: (): UserState => {
    const preferences = Storage.load(STORAGE_KEYS.PREFERENCES, { 
      theme: 'light', 
      language: 'zh-CN',
      sidebarCollapsed: false,
    })
    
      return {
      // 用户基本信息
      userId: '',
      userName: '',
      realName: '',
      tenantId: '',
      avatar: '',
      email: '',
      mobile: '',
      deptId: '',
      tenantAdminFlag: 'N', // 是否租户管理员：Y-是，N-否
      
      // 权限相关
      modules: [],
      buttons: [],
      moduleCodes: new Set(),
      buttonCodes: new Set(),
      permissionsLoaded: false,

      // 用户偏好
      theme: preferences.theme,
      language: preferences.language,
      sidebarCollapsed: preferences.sidebarCollapsed,
      
      // 会话状态
      rememberMe: false,
      isAuthenticated: false,
      timeout: 0, // API请求超时时间（毫秒）
    }
  },

  getters: {
    /** 显示名称 */
    displayName: (state) => state.realName || state.userName || '游客',
    
    /**
     * 检查是否有指定模块权限
     * 
     * 如果是租户管理员，则全部放行
     * 
     * @param {string} resourceCode - 资源编码，通常是模块编码（如：'hub0002'）
     * @returns {boolean} 如果用户拥有该模块权限则返回 true，否则返回 false
     */
    hasModule: (state) => (resourceCode: string): boolean => {
      // 如果是租户管理员，全部放行
      if (state.tenantAdminFlag === 'Y') {
        return true
      }
      return state.moduleCodes.has(resourceCode)
    },

    /**
     * 检查是否有指定按钮权限
     * 
     * 如果是租户管理员，则全部放行
     * 
     * @param {string} resourceCode - 资源编码，通常是按钮编码（如：'hub0002:add'）
     * @returns {boolean} 如果用户拥有该按钮权限则返回 true，否则返回 false
     */
    hasButton: (state) => (resourceCode: string): boolean => {
      // 如果是租户管理员，全部放行
      if (state.tenantAdminFlag === 'Y') {
        return true
      }
      return state.buttonCodes.has(resourceCode)
    },

    /**
     * 获取指定模块的所有按钮权限
     * 
     * @param {string} moduleCode - 模块编码（如：'hub0002'）
     * @returns {ButtonPermission[]} 该模块下的所有按钮权限列表
     */
    getModuleButtons: (state) => (moduleCode: string): ButtonPermission[] => {
      return state.buttons.filter(
        (btn) => btn.parentResourceId === moduleCode || btn.resourceCode.startsWith(`${moduleCode}:`)
      )
    },

    /**
     * 获取所有模块权限（按层级和排序）
     * 
     * @returns {ModulePermission[]} 排序后的模块权限列表
     */
    sortedModules: (state): ModulePermission[] => {
      return [...state.modules].sort((a, b) => {
        // 先按层级排序
        if (a.resourceLevel !== b.resourceLevel) {
          return a.resourceLevel - b.resourceLevel
        }
        // 再按排序顺序排序
        return a.sortOrder - b.sortOrder
      })
    },
  },

  actions: {
    /**
     * 登录 - 设置用户信息
     * 
     * @param {string} userId - 用户ID
     * @param {string} userName - 用户名
     * @param {string} realName - 真实姓名
     * @param {string} tenantId - 租户ID
     * @param {object} options - 可选参数
     * @param {string} [options.avatar] - 头像
     * @param {string} [options.email] - 邮箱
     * @param {string} [options.mobile] - 手机号
     * @param {string} [options.deptId] - 部门ID
     * @param {number} [options.timeout] - 超时时间
     * @param {boolean} [options.remember] - 是否记住登录
     */
    async setLoginState(
      userId: string,
      userName: string,
      realName: string,
      tenantId: string,
      options?: {
        avatar?: string
        email?: string
        mobile?: string
        deptId?: string
        tenantAdminFlag?: string
        timeout?: number
        remember?: boolean
      }
    ) {
      // 更新用户信息
      this.userId = userId
      this.userName = userName
      this.realName = realName
      this.tenantId = tenantId
      this.avatar = options?.avatar || ''
      this.email = options?.email || ''
      this.mobile = options?.mobile || ''
      this.deptId = options?.deptId || ''
      this.tenantAdminFlag = options?.tenantAdminFlag || 'N'
      this.rememberMe = options?.remember || false
      this.isAuthenticated = true

      // 设置请求超时并保存
      if (options?.timeout && options.timeout > 0) {
        this.timeout = options.timeout
        updateTimeout(options.timeout)
      }

      // 持久化
      this._persist()
      
      return true
    },

    /**
     * 设置用户权限
     * 
     * 从后端获取的权限数据设置到 store 中，并自动构建权限编码集合用于快速查找
     * 通常在用户登录成功后调用
     * 
     * @param {UserPermissionResponse} permissions - 用户权限数据，包含模块权限和按钮权限
     */
    setPermissions(permissions: UserPermissionResponse) {
      this.modules = permissions.modules || []
      this.buttons = permissions.buttons || []

      // 更新权限编码集合，用于快速查找
      this.moduleCodes = new Set(this.modules.map((m) => m.resourceCode))
      this.buttonCodes = new Set(this.buttons.map((b) => b.resourceCode))

      this.permissionsLoaded = true

      // 持久化权限数据
      this._persist()
    },

    /**
     * 清除权限数据
     * 
     * 清除所有权限信息，通常在用户登出时调用
     */
    clearPermissions() {
      this.modules = []
      this.buttons = []
      this.moduleCodes = new Set()
      this.buttonCodes = new Set()
      this.permissionsLoaded = false
    },

    /**
     * 检查是否有指定权限（模块或按钮）
     * 
     * 如果是租户管理员，则全部放行
     * 
     * @param {string} resourceCode - 资源编码，可以是模块编码（如：'hub0002'）或按钮编码（如：'hub0002:add'）
     * @returns {boolean} 如果用户拥有该权限则返回 true，否则返回 false
     */
    hasPermission(resourceCode: string): boolean {
      // 如果是租户管理员，全部放行
      if (this.tenantAdminFlag === 'Y') {
        return true
      }
      return this.moduleCodes.has(resourceCode) || this.buttonCodes.has(resourceCode)
    },

    /**
     * 检查是否有任意一个权限
     * 
     * @param {string[]} resourceCodes - 资源编码数组
     * @returns {boolean} 如果用户拥有数组中的任意一个权限则返回 true，否则返回 false
     */
    hasAnyPermission(resourceCodes: string[]): boolean {
      return resourceCodes.some((code) => this.hasPermission(code))
    },

    /**
     * 检查是否有所有权限
     * 
     * @param {string[]} resourceCodes - 资源编码数组
     * @returns {boolean} 如果用户拥有数组中的所有权限则返回 true，否则返回 false
     */
    hasAllPermissions(resourceCodes: string[]): boolean {
      return resourceCodes.every((code) => this.hasPermission(code))
    },

    /**
     * 登出 - 清除用户信息
     */
    clearUserInfo() {
      this.$reset()
      Storage.clear()
    },

    /**
     * 统一更新方法 - 更新用户信息和设置
     * 
     * @param data 要更新的数据（支持用户信息、偏好设置等）
     * @param options 更新选项
     */
    update(
      data: {
        // 用户基本信息
        realName?: string
        avatar?: string
        email?: string
        mobile?: string
        // 用户偏好设置
        theme?: string
        language?: string
        sidebarCollapsed?: boolean
      },
      options?: {
        /** 是否持久化用户数据（默认 true） */
        persistUserData?: boolean
        /** 是否持久化偏好设置（默认 true） */
        persistPreferences?: boolean
      }
    ) {
      const { persistUserData = true, persistPreferences = true } = options || {}
      
      // 标记哪些数据被更新了
      let userDataUpdated = false
      let preferencesUpdated = false

      // 更新用户基本信息
      if (data.realName !== undefined) {
        this.realName = data.realName
        userDataUpdated = true
      }
      if (data.avatar !== undefined) {
        this.avatar = data.avatar
        userDataUpdated = true
      }
      if (data.email !== undefined) {
        this.email = data.email
        userDataUpdated = true
      }
      if (data.mobile !== undefined) {
        this.mobile = data.mobile
        userDataUpdated = true
      }

      // 更新用户偏好设置
      if (data.theme !== undefined) {
        this.theme = data.theme
        preferencesUpdated = true
      }
      if (data.language !== undefined) {
        this.language = data.language
        preferencesUpdated = true
      }
      if (data.sidebarCollapsed !== undefined) {
        this.sidebarCollapsed = data.sidebarCollapsed
        preferencesUpdated = true
      }

      // 持久化
      if (userDataUpdated && persistUserData) {
        this._persist()
      }
      if (preferencesUpdated && persistPreferences) {
        Storage.save(STORAGE_KEYS.PREFERENCES, {
          theme: this.theme,
          language: this.language,
          sidebarCollapsed: this.sidebarCollapsed,
        }, true)
      }
    },

    /**
     * 更新用户设置（兼容旧方法，内部调用 update）
     */
    updateSettings(settings: { theme?: string; language?: string; sidebarCollapsed?: boolean }) {
      this.update(settings, { persistUserData: false })
    },

    /**
     * 切换侧边栏折叠状态
     */
    toggleSidebar() {
      this.update({ sidebarCollapsed: !this.sidebarCollapsed }, { persistUserData: false })
    },

    /**
     * 初始化 - 从存储恢复状态
     */
    initialize() {
      const rememberMe = Storage.load(STORAGE_KEYS.REMEMBER_ME, false)
      const userData = Storage.load<any>(STORAGE_KEYS.USER_DATA, null)

      if (userData && userData.userId) {
        // 恢复用户基本信息
        this.userId = userData.userId || ''
        this.userName = userData.userName || ''
        this.realName = userData.realName || ''
        this.tenantId = userData.tenantId || ''
        this.avatar = userData.avatar || ''
        this.email = userData.email || ''
        this.mobile = userData.mobile || ''
        this.deptId = userData.deptId || ''
        this.tenantAdminFlag = userData.tenantAdminFlag || 'N'
        this.rememberMe = rememberMe
        this.isAuthenticated = true

        // 恢复权限数据（数组需要转换回 Set）
        if (userData.modules && Array.isArray(userData.modules)) {
          this.modules = userData.modules
        }
        if (userData.buttons && Array.isArray(userData.buttons)) {
          this.buttons = userData.buttons
        }
        if (userData.moduleCodes && Array.isArray(userData.moduleCodes)) {
          this.moduleCodes = new Set(userData.moduleCodes)
        }
        if (userData.buttonCodes && Array.isArray(userData.buttonCodes)) {
          this.buttonCodes = new Set(userData.buttonCodes)
        }
        this.permissionsLoaded = userData.permissionsLoaded || false

        // 恢复超时设置
        if (userData.timeout && userData.timeout > 0) {
          updateTimeout(userData.timeout)
        }
      }
    },

    /**
     * 持久化状态
     */
    _persist() {
      const userData = {
        userId: this.userId,
        userName: this.userName,
        realName: this.realName,
        tenantId: this.tenantId,
        avatar: this.avatar,
        email: this.email,
        mobile: this.mobile,
        deptId: this.deptId,
        tenantAdminFlag: this.tenantAdminFlag,
        timeout: this.timeout, // 持久化超时设置
        // 权限数据（Set 需要转换为数组才能序列化）
        modules: this.modules,
        buttons: this.buttons,
        moduleCodes: Array.from(this.moduleCodes),
        buttonCodes: Array.from(this.buttonCodes),
        permissionsLoaded: this.permissionsLoaded,
      }

      Storage.save(STORAGE_KEYS.USER_DATA, userData, this.rememberMe)
      Storage.save(STORAGE_KEYS.REMEMBER_ME, this.rememberMe, this.rememberMe)
    },

    /**
     * 重置状态
     */
    $reset() {
      this.userId = ''
      this.userName = ''
      this.realName = ''
      this.tenantId = ''
      this.avatar = ''
      this.email = ''
      this.mobile = ''
      this.deptId = ''
      this.tenantAdminFlag = 'N'
      this.modules = []
      this.buttons = []
      this.moduleCodes = new Set()
      this.buttonCodes = new Set()
      this.permissionsLoaded = false
      this.rememberMe = false
      this.isAuthenticated = false
      this.timeout = 0
    },
  },
})

// ==================== 类型定义 ====================

/** 
 * Store 状态定义
 * 用户信息管理的核心状态
 */
interface UserState {
  // ========== 用户基本信息 ==========
  
  /** 用户唯一标识 ID */
  userId: string
  
  /** 用户登录名/账号 */
  userName: string
  
  /** 用户真实姓名 */
  realName: string
  
  /** 租户 ID（多租户系统使用） */
  tenantId: string
  
  /** 用户头像 URL */
  avatar: string
  
  /** 用户邮箱地址 */
  email: string
  
  /** 用户手机号码 */
  mobile: string
  
  /** 所属部门 ID */
  deptId: string
  
  /** 是否租户管理员：Y-是，N-否 */
  tenantAdminFlag: string
  
  // ========== 权限相关 ==========
  
  /** 模块权限列表 */
  modules: ModulePermission[]
  
  /** 按钮权限列表 */
  buttons: ButtonPermission[]
  
  /** 模块权限编码集合（用于快速查找） */
  moduleCodes: Set<string>
  
  /** 按钮权限编码集合（用于快速查找） */
  buttonCodes: Set<string>
  
  /** 是否已加载权限 */
  permissionsLoaded: boolean
  
  // ========== 用户偏好设置 ==========
  
  /** 主题模式（'light' | 'dark' 等） */
  theme: string
  
  /** 界面语言（如：'zh-CN', 'en-US'） */
  language: string
  
  /** 侧边栏是否折叠 */
  sidebarCollapsed: boolean
  
  // ========== 会话状态 ==========
  
  /** 
   * 是否记住登录
   * true: 使用 localStorage（持久化）
   * false: 使用 sessionStorage（关闭浏览器清除）
   */
  rememberMe: boolean
  
  /** 是否已认证登录 */
  isAuthenticated: boolean
  
  /** API请求超时时间（毫秒），0表示使用默认值 */
  timeout: number
}

// ==================== 存储工具 ====================

const STORAGE_KEYS = {
  USER_DATA: 'user:data',
  PREFERENCES: 'user:preferences',
  REMEMBER_ME: 'user:rememberMe',
} as const

/** 存储管理器 */
class Storage {
  /** 保存数据（根据 rememberMe 决定存储位置） */
  static save(key: string, value: any, persistent: boolean): void {
    const storage = persistent ? localStorage : sessionStorage
    storage.setItem(key, JSON.stringify(value))
    if (!persistent) localStorage.removeItem(key)
  }

  /** 读取数据（优先 localStorage，其次 sessionStorage） */
  static load<T>(key: string, defaultValue: T): T {
    const value = localStorage.getItem(key) || sessionStorage.getItem(key)
    if (!value) return defaultValue
    try {
      return JSON.parse(value)
    } catch {
      return defaultValue
    }
  }

  /** 删除数据 */
  static remove(key: string): void {
    localStorage.removeItem(key)
    sessionStorage.removeItem(key)
  }

  /** 清除所有用户数据 */
  static clear(): void {
    Object.values(STORAGE_KEYS).forEach(k => this.remove(k))
    this.remove('token')
  }
}
