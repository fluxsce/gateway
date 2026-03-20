/**
 * GDialog 插件：install 时挂载 GDialogProvider，并挂到 app.config.globalProperties 与 window，便于直接调用。
 *
 * @example
 * app.use(gdialogPlugin, { global: true, globalName: '$gDialog' })
 */

import type { App } from 'vue'
import { h, render } from 'vue'
import { $gDialog } from './useGDialog'
import GDialogProvider from './GDialogProvider.vue'

export interface GDialogPluginOptions {
  /** 是否挂到 window，便于控制台/TS 直接调用 */
  global?: boolean
  /** window 上的名称，默认 '$gDialog' */
  globalName?: string
}

const gdialogPlugin = {
  install(app: App, options?: GDialogPluginOptions) {
    ;(app.config.globalProperties as Record<string, unknown>).$gDialog = $gDialog

    const shouldMountToWindow = options?.global !== false
    const globalName = options?.globalName ?? '$gDialog'
    if (shouldMountToWindow && typeof window !== 'undefined') {
      ;(window as unknown as Record<string, unknown>)[globalName] = $gDialog
    }

    const container = document.createElement('div')
    container.id = 'g-dialog-provider'
    document.body.appendChild(container)

    const vnode = h(GDialogProvider)
    const appContext = (app as unknown as { _context: App['_context'] })._context
    if (appContext) {
      (vnode as unknown as { appContext: typeof appContext }).appContext = appContext
    }
    render(vnode, container)
  },
}

export default gdialogPlugin
