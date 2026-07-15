/**
 * 全局浮层层级。
 * 新增 Teleport、弹窗附属层或提示层时必须在此定义，避免组件间相互遮挡。
 */
export const Z_INDEX = Object.freeze({
  VXE_TABLE: 4000,
  CONTEXT_MENU: 5000,
  MODAL_RESIZE_HANDLE: 200000,
  MODAL_RESIZE_CORNER: 200001,
  TOOLTIP: 300000,
})
