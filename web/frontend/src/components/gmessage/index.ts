/**
 * GMessage 消息与对话框组件
 *
 * 提供消息条（info/success/error/warning/loading）与对话框（confirm/alert/prompt，基于 GDialog）。
 * 插件注册后可通过 $gMessage 或 window.$gMessage 调用。
 */

export { default as GMessage } from './GMessage.vue'
export { default as GMessageProvider } from './GMessageProvider.vue'
export { default as gmessagePlugin } from './plugin'
export type {
  GMessageApi, GMessageDialogOptions, GMessageOptions, GMessagePosition,
  GMessageProviderProps, GMessageType
} from './types'
export { $gMessage, getMessageApi, setMessageProvider } from './utils'

