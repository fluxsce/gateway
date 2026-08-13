/**
 * @deprecated GDialog / useGDialog 已弃用。
 * - 确认/提示：`rsConfirm` / `RsConfirmDialog`（@/ui）
 * - 工作窗/表单：`RsDialog`（@/ui）
 */

export { default as GDialog } from './GDialog.vue'
export { default as gdialogPlugin } from './plugin'
export type { GDialogPluginOptions } from './plugin'
/** @deprecated 随 GDialog 淘汰；请改用 RsDialog props / #footer + RsButton */
export * from './types'
export { createDialog, useGDialog, $gDialog } from './useGDialog'
export type { GDialogApi, GDialogOptions, GDialogReactive } from './useGDialog'
