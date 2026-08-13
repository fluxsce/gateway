/**
 * GDialog 插件：将 $gDialog 挂到 app.config.globalProperties 与 window。
 * 命令式确认已切到 niuma-ui rsConfirm / openRsDialog，不再挂载常驻 Provider。
 *
 * @example
 * app.use(gdialogPlugin, { global: true, globalName: '$gDialog' })
 */

import type { App } from 'vue'
import { $gDialog } from './useGDialog'

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
  },
}

export default gdialogPlugin
