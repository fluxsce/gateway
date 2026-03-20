/**
 * GMessage 插件
 *
 * 安装时：为 6 个位置挂载 GMessageProvider，将 $gMessage 挂到 app.config.globalProperties
 * 与 window[globalName]（默认 '$gMessage'），便于在任意组件或 TS 中调用消息与对话框 API。
 *
 * @example
 * app.use(gmessagePlugin, { global: true, globalName: '$gMessage' })
 */

import type { App } from 'vue'
import { h, render } from 'vue'
import { $gMessage } from './utils'
import GMessageProvider from './GMessageProvider.vue'
import type { GMessageProviderProps } from './types'

const POSITIONS: Array<NonNullable<GMessageProviderProps['position']>> = [
  'top',
  'top-left',
  'top-right',
  'bottom',
  'bottom-left',
  'bottom-right',
]

const gmessagePlugin = {
  install(app: App, options?: { global?: boolean; globalName?: string; provider?: Partial<GMessageProviderProps> }) {
    ;(app.config.globalProperties as Record<string, unknown>).$gMessage = $gMessage

    const shouldMountToWindow = options?.global !== false
    const globalName = options?.globalName ?? '$gMessage'
    if (shouldMountToWindow && typeof window !== 'undefined') {
      ;(window as unknown as Record<string, unknown>)[globalName] = $gMessage
    }

    const defaultProviderProps = {
      duration: 3000,
      closable: true,
      max: 5,
      ...options?.provider,
    }

    POSITIONS.forEach((position) => {
      const container = document.createElement('div')
      container.id = `g-message-provider-${position}`
      document.body.appendChild(container)

      const vnode = h(GMessageProvider, {
        position,
        ...defaultProviderProps,
      })
      const appContext = (app as unknown as { _context: App['_context'] })._context
      if (appContext) {
        (vnode as unknown as { appContext: typeof appContext }).appContext = appContext
      }
      render(vnode, container)
    })
  },
}

export default gmessagePlugin
