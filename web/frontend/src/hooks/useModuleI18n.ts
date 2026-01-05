/**
 * 模块化国际化Hook
 * 提供一种简单的方式在Vue组件中使用模块化的多语言翻译
 */
import { computed, watch, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { loadModuleMessages, isModuleMessagesLoaded, type LocaleType } from '@/locales'

/**
 * 使用模块特定的国际化资源
 * 此Hook简单包装了vue-i18n，添加了模块加载功能
 *
 * @param moduleName 模块名称，对应于语言文件夹下的文件名
 * @param options 配置选项
 * @returns 国际化工具
 *
 * 使用示例:
 * ```
 * // 在组件中加载并使用login模块的翻译
 * const { t, loading } = useModuleI18n('hub0001')
 *
 * // 在模板中使用，会自动使用命名空间
 * {{ t('title') }} // 实际访问 hub0001.title
 * ```
 * 
 * 注意：
 * 1. 底层已实现缓存机制，同一个模块的同一语言只会加载一次
 * 2. 如果在路由配置中已经设置了 meta.moduleName，
 *    路由守卫会在进入页面前预加载对应的多语言资源
 * 3. Hook会自动检测模块是否已加载，避免重复处理
 * 4. 多个组件同时调用同一模块不会导致重复加载
 * 5. 此Hook只能在Vue组件的setup函数中使用，不能在路由守卫等非组件环境中调用
 */
export function useModuleI18n(
  moduleName: string,
  options: {
    immediate?: boolean // 是否立即加载模块语言资源
    silent?: boolean // 是否静默处理错误，不在控制台输出
  } = {},
) {
  const { immediate = true, silent = false } = options

  // 使用全局i18n
  const i18n = useI18n()

  // 加载状态
  const loading = ref(false)
  const loaded = ref(false)
  const error = ref<Error | null>(null)

  /**
   * 加载模块语言资源
   * @param locale 要加载的语言
   */
  const loadModule = async (locale = i18n.locale.value) => {
    const targetLocale = locale as LocaleType

    // 检查是否已经加载过
    if (isModuleMessagesLoaded(moduleName, targetLocale)) {
      loaded.value = true
      if (!silent) {
        console.log(`[useModuleI18n] 模块 "${moduleName}" 的语言包已缓存: ${locale}`)
      }
      return
    }

    // 如果正在加载，则跳过
    if (loading.value) {
      return
    }

    loading.value = true
    error.value = null

    try {
      await loadModuleMessages(moduleName, targetLocale)
      loaded.value = true

      if (!silent) {
        console.log(`[useModuleI18n] 已加载模块 "${moduleName}" 的语言包: ${locale}`)
      }
    } catch (err) {
      if (!silent) {
        console.error(`[useModuleI18n] 加载模块 "${moduleName}" 的语言包失败: ${locale}`, err)
      }
      error.value = err instanceof Error ? err : new Error(String(err))
    } finally {
      loading.value = false
    }
  }

  /**
   * 创建一个带命名空间的翻译函数
   * 自动将键添加模块前缀，例如:
   * t('login.title') -> i18n.t('hub0001.login.title')
   */
  const t = (key: string, params?: Record<string, any>) => {
    // 完整的翻译键
    const fullKey = key.startsWith(`${moduleName}.`) ? key : `${moduleName}.${key}`

    // 尝试翻译
    return i18n.t(fullKey, params || {})
  }

  // 监听语言变化，加载新语言的模块资源
  watch(
    () => i18n.locale.value,
    (newLocale) => {
      if (loaded.value) {
        loadModule(newLocale)
      }
    },
  )

  // 初始化时检查当前语言是否已加载
  const currentLocale = i18n.locale.value as LocaleType
  if (isModuleMessagesLoaded(moduleName, currentLocale)) {
    loaded.value = true
    if (!silent) {
      console.log(`[useModuleI18n] 模块 "${moduleName}" 的语言包已存在: ${currentLocale}`)
    }
  } else if (immediate) {
    // 使用Promise.resolve().then确保微任务队列执行，避免阻塞UI
    Promise.resolve().then(() => loadModule(currentLocale))
  }

  return {
    t,
    loadModule,
    loading: computed(() => loading.value),
    loaded: computed(() => loaded.value),
    error: computed(() => error.value),
    locale: computed(() => i18n.locale.value),
  }
}
