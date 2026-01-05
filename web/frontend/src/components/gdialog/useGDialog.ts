/**
 * GDialog 程序化调用 Hook
 * 允许在 TypeScript 中直接调用对话框
 */

import { darkThemeOverrides, lightThemeOverrides } from '@/config/theme'
import { getCurrentLocale } from '@/locales'
import { useUserStore } from '@/stores/user'
import hljs from '@/utils/highlight'
import { darkTheme, dateEnUS, dateZhCN, enUS, NConfigProvider, zhCN } from 'naive-ui'
import type { Hljs } from 'naive-ui/es/_mixins'
import { h, render, type Component, type VNode } from 'vue'
import GDialog from './GDialog.vue'

/**
 * 获取当前主题配置
 */
function getThemeConfig() {
  const userStore = useUserStore()
  
  // 是否为深色模式（获取当前值，不使用 computed）
  const theme = userStore.theme
  const isDark = theme === 'system' 
    ? window.matchMedia('(prefers-color-scheme: dark)').matches
    : theme === 'dark'
  
  // Naive UI 主题对象
  const naiveTheme = isDark ? darkTheme : null
  
  // 多语言配置
  const currentLang = getCurrentLocale()
  const naiveLocale = currentLang === 'en' ? enUS : zhCN
  const naiveDateLocale = currentLang === 'en' ? dateEnUS : dateZhCN
  
  // 将 hljs 实例转换为 naive-ui 期望的类型
  const hljsInstance: Hljs = {
    highlight: hljs.highlight.bind(hljs),
    getLanguage: hljs.getLanguage.bind(hljs)
  }
  
  return {
    isDark,
    naiveTheme,
    naiveLocale,
    naiveDateLocale,
    themeOverrides: isDark ? darkThemeOverrides : lightThemeOverrides,
    hljsInstance
  }
}

/**
 * 对话框选项
 */
export interface GDialogOptions {
  /** 对话框标题 */
  title?: string
  /** 对话框副标题 */
  subtitle?: string
  /** 头部图标组件 */
  icon?: Component
  /** 图标大小 */
  iconSize?: number
  /** 头部样式类型 */
  headerStyle?: 'default' | 'gradient'
  /** 对话框宽度 */
  width?: number | string
  /** 内容（可以是字符串、VNode 或组件） */
  content?: string | VNode | Component
  /** 是否可点击遮罩关闭 */
  maskClosable?: boolean
  /** 是否可按下 ESC 关闭 */
  closeOnEsc?: boolean
  /** 确认按钮文案 */
  positiveText?: string
  /** 取消按钮文案 */
  negativeText?: string
  /** 是否显示取消按钮 */
  showCancel?: boolean
  /** 是否显示确认按钮 */
  showConfirm?: boolean
  /** 确认按钮加载状态 */
  confirmLoading?: boolean
  /** 副标题显示位置：'header' 显示在头部 | 'footer' 显示在底部 */
  subtitlePosition?: 'header' | 'footer'
  /** 底部操作按钮对齐方式：'start' 左对齐 | 'end' 右对齐 | 'center' 居中 | 'space-between' 两端对齐 | 'space-around' 环绕分布 */
  footerButtonAlign?: 'start' | 'end' | 'center' | 'space-between' | 'space-around'
}

/**
 * 对话框实例
 */
export interface GDialogReactive {
  /** 关闭对话框 */
  destroy: () => void
  /** 更新确认按钮加载状态 */
  setConfirmLoading: (loading: boolean) => void
}

/**
 * 创建对话框
 */
function createDialog(options: GDialogOptions): Promise<boolean> {
  return new Promise((resolve) => {
    let container: HTMLElement | null = document.createElement('div')
    let isDestroyed = false

    // 获取主题配置
    const themeConfig = getThemeConfig()

    // 处理内容：字符串转换为 VNode
    const contentNode = (() => {
      if (!options.content) {
        return null
      }
      
      if (typeof options.content === 'string') {
        // 将换行符转换为 <br> 标签
        return h('div', {
          style: {
            whiteSpace: 'pre-line',
            lineHeight: '1.6'
          }
        }, options.content)
      }
      
      if (typeof options.content === 'object' && 'render' in options.content) {
        return h(options.content as Component)
      }
      
      return options.content
    })()

    // 创建对话框 VNode
    const dialogVNode = h(GDialog, {
      show: true,
      title: options.title,
      subtitle: options.subtitle,
      subtitlePosition: options.subtitlePosition || 'footer',
      icon: options.icon,
      iconSize: options.iconSize,
      headerStyle: options.headerStyle || 'default',
      width: options.width || 500,
      maskClosable: options.maskClosable !== false,
      closeOnEsc: options.closeOnEsc !== false,
      showCancel: options.showCancel !== false,
      showConfirm: options.showConfirm !== false,
      cancelText: options.negativeText || '取消',
      confirmText: options.positiveText || '确定',
      confirmLoading: options.confirmLoading || false,
      footerButtonAlign: options.footerButtonAlign || 'end',
      autoCloseOnConfirm: false,
      onUpdateShow: (show: boolean) => {
        if (!show && !isDestroyed) {
          destroy()
        }
      },
      onConfirm: () => {
        if (!isDestroyed) {
          destroy()
          resolve(true)
        }
      },
      onCancel: () => {
        if (!isDestroyed) {
          destroy()
          resolve(false)
        }
      },
      onClose: () => {
        if (!isDestroyed) {
          destroy()
          resolve(false)
        }
      }
    }, {
      default: () => contentNode
    })

    // 用 NConfigProvider 包裹对话框，注入主题配置
    const configProviderVNode = h(NConfigProvider, {
      theme: themeConfig.naiveTheme,
      themeOverrides: themeConfig.themeOverrides,
      locale: themeConfig.naiveLocale,
      dateLocale: themeConfig.naiveDateLocale,
      hljs: themeConfig.hljsInstance
    }, {
      default: () => dialogVNode
    })

    // 渲染对话框
    if (container) {
      render(configProviderVNode, container)
      document.body.appendChild(container)
    }

    // 销毁对话框
    function destroy() {
      if (isDestroyed || !container) {
        return
      }
      
      isDestroyed = true
      
      if (container.parentNode) {
        render(null, container)
        container.parentNode.removeChild(container)
        container = null
      }
    }
  })
}

/**
 * 使用 GDialog Hook
 * 提供程序化调用对话框的方法
 */
export function useGDialog() {
  /**
   * 显示警告对话框
   */
  const warning = (options: GDialogOptions | string): Promise<boolean> => {
    if (typeof options === 'string') {
      return createDialog({
        title: '警告',
        content: options,
        headerStyle: 'gradient',
        width: 500
      })
    }
    return createDialog({
      title: '警告',
      headerStyle: 'gradient',
      width: 500,
      ...options
    })
  }

  /**
   * 显示信息对话框
   */
  const info = (options: GDialogOptions | string): Promise<boolean> => {
    if (typeof options === 'string') {
      return createDialog({
        title: '提示',
        content: options,
        width: 500
      })
    }
    return createDialog({
      title: '提示',
      width: 500,
      ...options
    })
  }

  /**
   * 显示成功对话框
   */
  const success = (options: GDialogOptions | string): Promise<boolean> => {
    if (typeof options === 'string') {
      return createDialog({
        title: '成功',
        content: options,
        headerStyle: 'gradient',
        width: 500,
        showCancel: false
      })
    }
    return createDialog({
      title: '成功',
      headerStyle: 'gradient',
      width: 500,
      showCancel: false,
      ...options
    })
  }

  /**
   * 显示错误对话框
   */
  const error = (options: GDialogOptions | string): Promise<boolean> => {
    if (typeof options === 'string') {
      return createDialog({
        title: '错误',
        content: options,
        headerStyle: 'gradient',
        width: 500,
        showCancel: false
      })
    }
    return createDialog({
      title: '错误',
      headerStyle: 'gradient',
      width: 500,
      showCancel: false,
      ...options
    })
  }

  /**
   * 显示确认对话框
   */
  const confirm = (options: GDialogOptions | string): Promise<boolean> => {
    if (typeof options === 'string') {
      return createDialog({
        title: '确认',
        content: options,
        width: 500
      })
    }
    return createDialog({
      title: '确认',
      width: 500,
      ...options
    })
  }

  /**
   * 显示自定义对话框
   */
  const create = (options: GDialogOptions): Promise<boolean> => {
    return createDialog(options)
  }

  return {
    warning,
    info,
    success,
    error,
    confirm,
    create
  }
}

