/**
 * Store模块索引
 * 集中导出所有状态存储模块，方便在模板中全局访问
 */

// 导入状态存储模块
import type { LocaleType } from '@/locales'
import { setLocale } from '@/locales'
import { useGlobalStore } from './global'
import { useUserStore } from './user'

// 导出用户信息存储（包含权限管理）
export { useUserStore } from './user'
export type { ButtonPermission, ModulePermission, UserPermissionResponse } from './user'

// 导出全局信息缓存
export { useGlobalStore } from './global'

/**
 * 全局Store助手对象(使用代理对象)
 * 提供一个更简单的方式在任何地方访问Store
 */
export const store = {
  /**
   * 用户相关的store访问
   */
  user: new Proxy({} as ReturnType<typeof useUserStore>, {
    get(_, property) {
      try {
        const storeInstance = useUserStore()
        if (typeof property === 'string') {
          const value = (storeInstance as any)[property]
          // 对于函数，返回一个绑定了正确this上下文的新函数
          if (typeof value === 'function') {
            return value.bind(storeInstance)
          }
          // 直接返回属性值
          return value
        }
      } catch (error) {
        console.warn(`访问user store属性${String(property)}时出错:`, error)
        return undefined
      }
      return undefined
    },
  }),

  /**
   * 全局配置相关的store访问
   */
  global: new Proxy({} as ReturnType<typeof useGlobalStore>, {
    get(_, property) {
      try {
        const storeInstance = useGlobalStore()
        if (typeof property === 'string') {
          const value = (storeInstance as any)[property]
          // 对于函数，返回一个绑定了正确this上下文的新函数
          if (typeof value === 'function') {
            return value.bind(storeInstance)
          }
          return value
        }
      } catch (error) {
        console.warn(`访问global store属性${String(property)}时出错:`, error)
        return undefined
      }
      return undefined
    },
  }),

  /**
   * 语言相关的访问代理
   * 通过用户设置管理语言首选项
   */
  locale: new Proxy(
    {} as {
      currentLocale: LocaleType
      loading: boolean
      setLocale: (locale: LocaleType) => Promise<void>
    },
    {
      get(_, property) {
        try {
          // 避免重复引用userStore
          if (property === 'currentLocale') {
            const userStore = useUserStore()
            // 从用户设置转换为i18n使用的locale格式
            return userStore.language === 'en-US' ? 'en' : 'zh-CN'
          }

          if (property === 'loading') {
            return false
          }

          if (property === 'setLocale') {
            return async (locale: LocaleType) => {
              try {
                const userStore = useUserStore()
                const storeLanguage = locale === 'en' ? 'en-US' : 'zh-CN'

                // 更新用户设置
                userStore.updateSettings({ language: storeLanguage })

                // 更新i18n设置
                await setLocale(locale)
              } catch (error) {
                console.error('设置locale失败:', error)
              }
            }
          }
        } catch (error) {
          console.warn(`访问locale属性${String(property)}时出错:`, error)
        }

        return undefined
      },
    },
  ),
}

/**
 * 初始化所有状态存储
 * 应用启动时调用
 */
export async function initializeStores() {
  const userStore = useUserStore()
  const globalStore = useGlobalStore()

  // 初始化用户状态
  userStore.initialize()

  // 返回初始化后的store实例，以便进一步操作
  return {
    userStore,
    globalStore,
  }
}

// 导出辅助函数
/**
 * 在模板中快速获取用户信息，支持在模板中直接使用
 * 例如：{{ $user.displayName }}
 */
export function setupStoreHelpers(app: any) {
  // 全局属性 - 通过$user访问用户信息
  app.config.globalProperties.$user = {
    get displayName() {
      const store = useUserStore()
      return store.displayName
    },

    get avatar() {
      const store = useUserStore()
      return store.avatar || '/img/default-avatar.png'
    },

    get isLoggedIn() {
      const store = useUserStore()
      return store.isAuthenticated
    },

    hasPermission(permCode: string) {
      const store = useUserStore()
      return store.hasPermission(permCode)
    },

    hasModule(moduleCode: string) {
      const store = useUserStore()
      return store.hasModule(moduleCode)
    },

    hasButton(buttonCode: string) {
      const store = useUserStore()
      return store.hasButton(buttonCode)
    },
  }

  // 全局属性 - 通过$app访问应用信息
  app.config.globalProperties.$app = {
    get name() {
      const store = useGlobalStore()
      return store.appName
    },

    get version() {
      const store = useGlobalStore()
      return store.appVersion
    },
  }

  // 添加全局store对象到应用实例
  app.config.globalProperties.$store = store
}
