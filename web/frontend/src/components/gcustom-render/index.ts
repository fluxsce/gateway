/**
 * GCustomRender 自定义渲染
 * 插件注册后可直接使用 $gRender.show(Component, props, options)
 */

export { $gRender, current } from './api'
export type { GlobalCustomRenderApi, GlobalCustomRenderOptions } from './api'
export { default as gcustomRenderPlugin } from './plugin'
export type { GCustomRenderProps, CustomRenderScope } from './types'
