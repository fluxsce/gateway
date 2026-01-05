/**
 * 语言包索引文件
 * 集中管理所有支持的语言
 *
 * 简化版 - 更直接地利用vue-i18n的功能
 */

import { nextTick } from 'vue'
import { createI18n, type I18n, type Composer } from 'vue-i18n'

export type LocaleType = 'en' | 'zh-CN'

// 支持的语言列表
export const availableLocales = [
  {
    locale: 'en',
    name: 'English',
    flag: 'usa',
  },
  {
    locale: 'zh-CN',
    name: '简体中文',
    flag: 'china',
  },
]

// 默认语言
export const defaultLocale: LocaleType = 'zh-CN'

/**
 * 根据浏览器设置自动检测首选语言
 */
export function detectLocale(): LocaleType {
  const browserLang = navigator.language

  if (browserLang.startsWith('zh')) {
    return 'zh-CN'
  }

  return browserLang === 'en' || browserLang.startsWith('en-') ? 'en' : defaultLocale
}

// 语言文件目录结构映射
const LOCALE_PATH_MAP: Record<LocaleType, string> = {
  en: 'en',
  'zh-CN': 'zh-CN',
}

// 创建i18n实例
let i18n: I18n | null = null

// 模块加载缓存，避免重复加载
const moduleLoadCache = new Map<string, Promise<any>>()
const loadedModules = new Set<string>()

/**
 * 生成模块缓存键
 * @param moduleName 模块名称
 * @param locale 语言代码
 */
function getModuleCacheKey(moduleName: string, locale: LocaleType): string {
  return `${moduleName}:${locale}`
}

/**
 * 检查模块是否已加载
 * @param moduleName 模块名称
 * @param locale 语言代码
 */
function isModuleLoaded(moduleName: string, locale: LocaleType): boolean {
  const cacheKey = getModuleCacheKey(moduleName, locale)
  return loadedModules.has(cacheKey)
}

/**
 * 加载模块语言文件
 * @param moduleName 模块名称
 * @param locale 语言代码
 * @param force 是否强制重新加载，忽略缓存
 */
export async function loadModuleMessages(
  moduleName: string, 
  locale: LocaleType = defaultLocale,
  force: boolean = false
) {
  if (!i18n) return {}

  const cacheKey = getModuleCacheKey(moduleName, locale)

  // 如果已经加载过且不强制重新加载，直接返回已缓存的消息
  if (!force && isModuleLoaded(moduleName, locale)) {
    const currentMessages = i18n.global.getLocaleMessage(locale) as Record<string, any> || {}
    return currentMessages[moduleName] || {}
  }

  // 如果正在加载中，返回正在进行的Promise
  if (moduleLoadCache.has(cacheKey)) {
    return moduleLoadCache.get(cacheKey)
  }

  // 创建加载Promise
  const loadPromise = (async () => {
    try {
      // 根据不同语言选择不同路径
      const localePath = LOCALE_PATH_MAP[locale] || locale

      // 动态导入模块语言包
      /* @vite-ignore */
      const moduleLocale = await import(`./${localePath}/${moduleName}.ts`)

      // 如果还没有为这个locale注册任何消息
      if (!i18n.global.availableLocales.includes(locale)) {
        i18n.global.setLocaleMessage(locale, {})
      }

      // 将模块消息合并到全局消息中，使用模块名作为命名空间
      const currentMessages = i18n.global.getLocaleMessage(locale) as Record<string, any> || {}
      i18n.global.setLocaleMessage(locale, {
        ...currentMessages,
        [moduleName]: moduleLocale.default,
      })

      // 标记为已加载
      loadedModules.add(cacheKey)

      console.log(`[i18n] 模块 "${moduleName}" 语言包加载完成: ${locale}`)
      return moduleLocale.default
    } catch (error) {
      console.warn(`[i18n] 加载模块语言包失败: ${moduleName}/${locale}`, error)
      return {}
    } finally {
      // 清除加载中的缓存
      moduleLoadCache.delete(cacheKey)
    }
  })()

  // 缓存加载Promise
  moduleLoadCache.set(cacheKey, loadPromise)

  return loadPromise
}

/**
 * 检查模块是否已加载（导出版本）
 * @param moduleName 模块名称
 * @param locale 语言代码
 */
export function isModuleMessagesLoaded(moduleName: string, locale: LocaleType = defaultLocale): boolean {
  return isModuleLoaded(moduleName, locale)
}

/**
 * 清除模块加载缓存
 * @param moduleName 模块名称，不传则清除所有缓存
 * @param locale 语言代码，不传则清除指定模块的所有语言缓存
 */
export function clearModuleCache(moduleName?: string, locale?: LocaleType) {
  if (!moduleName) {
    // 清除所有缓存
    moduleLoadCache.clear()
    loadedModules.clear()
    console.log('[i18n] 已清除所有模块语言包缓存')
    return
  }

  if (!locale) {
    // 清除指定模块的所有语言缓存
    const keysToDelete = Array.from(loadedModules).filter(key => key.startsWith(`${moduleName}:`))
    keysToDelete.forEach(key => {
      loadedModules.delete(key)
      moduleLoadCache.delete(key)
    })
    console.log(`[i18n] 已清除模块 "${moduleName}" 的所有语言包缓存`)
    return
  }

  // 清除指定模块和语言的缓存
  const cacheKey = getModuleCacheKey(moduleName, locale)
  loadedModules.delete(cacheKey)
  moduleLoadCache.delete(cacheKey)
  console.log(`[i18n] 已清除模块 "${moduleName}" 的 ${locale} 语言包缓存`)
}

/**
 * 预加载多个模块的语言包
 * @param modules 模块名称列表
 * @param locales 要加载的语言列表
 */
export async function preloadModules(
  modules: string[],
  locales: LocaleType[] = [defaultLocale, 'en'],
) {
  const loadPromises = []

  for (const moduleName of modules) {
    for (const locale of locales) {
      loadPromises.push(loadModuleMessages(moduleName, locale))
    }
  }

  return Promise.all(loadPromises)
}

/**
 * 切换语言
 * @param locale 目标语言
 */
export async function setLocale(locale: LocaleType) {
  if (!i18n) return

  // 使用类型断言确保TS理解i18n.global.locale的类型
  const globalComposer = i18n.global as Composer
  const previousLocale = globalComposer.locale.value as LocaleType

  // 如果语言没有变化，则不执行操作
  if (previousLocale === locale) return

  // 更改语言设置
  globalComposer.locale.value = locale

  // 设置HTML lang属性和本地存储
  document.querySelector('html')?.setAttribute('lang', locale)
  localStorage.setItem('locale', locale)

  // 重新加载已加载的模块
  // 这里先仅加载common模块确保基础翻译可用
  await preloadModules(['common'], [locale])

  return nextTick()
}

/**
 * 获取当前使用的语言
 */
export function getCurrentLocale(): LocaleType {
  if (!i18n) return defaultLocale

  const globalComposer = i18n.global as Composer
  return globalComposer.locale.value as LocaleType
}

/**
 * 初始化i18n实例
 * @param additionalMessages 可选的额外初始消息
 */
export function setupI18n(additionalMessages: Record<string, any> = {}) {
  // 从localStorage获取已保存的语言设置，或使用浏览器检测的语言
  const savedLocale = (localStorage.getItem('locale') as LocaleType) || detectLocale()

  // 确保additionalMessages符合预期的类型
  const typedMessages: Record<LocaleType, any> = {
    en: additionalMessages['en'] || {},
    'zh-CN': additionalMessages['zh-CN'] || {},
  }

  // 使用初始消息创建i18n实例
  i18n = createI18n({
    legacy: false, // 使用Composition API模式
    locale: savedLocale,
    fallbackLocale: false, // 禁用回退翻译，不使用任何其他语言作为后备
    messages: typedMessages,
    silentTranslationWarn: true, // 始终关闭翻译缺失警告
    missingWarn: false, // 关闭缺失警告
    fallbackWarn: false, // 关闭回退警告
  })

  // 设置HTML lang属性
  document.querySelector('html')?.setAttribute('lang', savedLocale)

  // 预加载公共模块和系统常用模块
  setTimeout(() => {
    preloadModules(['common'], [savedLocale])
  }, 0)

  return i18n
}

/**
 * 获取i18n实例
 */
export function getI18nInstance() {
  return i18n
}
