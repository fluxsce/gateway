/**
 * 对话框组件导出
 */
export { default as GDialog } from './GDialog.vue'
export { default as GDialogProvider } from './GDialogProvider.vue'
export { default as gdialogPlugin } from './plugin'
export type { GDialogPluginOptions } from './plugin'
export * from './types'
export { createDialog, getDialogApi, setDialogProvider, useGDialog, $gDialog } from './useGDialog'
export type { GDialogApi, GDialogOptions, GDialogProviderApi, GDialogReactive } from './useGDialog'

