/**
 * 全局自定义渲染插件：挂载 Provider，暴露 $gRender，TS 内可直接调用 $gRender.show(Component, props, options)
 */

import type { App } from 'vue'
import { h, render } from 'vue'
import { $gRender } from './api'
import GlobalCustomRenderProvider from './GlobalCustomRenderProvider.vue'

const gcustomRenderPlugin = {
  install(app: App, options?: { global?: boolean; globalName?: string }) {
    app.config.globalProperties.$gRender = $gRender

    const shouldMountToWindow = options?.global !== false
    const globalName = options?.globalName ?? '$gRender'
    if (shouldMountToWindow && typeof window !== 'undefined') {
      ;(window as unknown as Record<string, unknown>)[globalName] = $gRender
    }

    const container = document.createElement('div')
    container.id = 'g-custom-render-provider'
    document.body.appendChild(container)

    const providerVNode = h(GlobalCustomRenderProvider)
    const appContext = (app as unknown as { _context: App['_context'] })._context
    if (appContext) {
      (providerVNode as unknown as { appContext: typeof appContext }).appContext = appContext
    }
    render(providerVNode, container)
  },
}

export default gcustomRenderPlugin
