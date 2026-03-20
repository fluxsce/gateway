/**
 * GMessage 全局 Provider 注册与消息/对话框 API
 *
 * 提供：
 * - setMessageProvider / getMessageApi：供 GMessageProvider 注册与按位置获取 API
 * - createMessage：内部统一创建消息条（info/success/error/warning/loading）
 * - $gMessage：对外暴露的 API，包含消息条与对话框（confirm/alert/prompt，基于 GDialog）
 *
 * @see GMessageProvider.vue 各位置 Provider 挂载后调用 setMessageProvider
 * @see plugin.ts 将 $gMessage 挂到 app.config.globalProperties 与 window
 */

import { createDialog } from '@/components/gdialog/useGDialog'
import type {
  GMessageApi,
  GMessageDialogOptions,
  GMessageOptions,
  GMessageProviderApi,
  GMessageType,
} from './types'

/** 按位置存储的 Provider 实例（message 仅包含消息条 API） */
const globalProviders = new Map<string, { message: GMessageProviderApi }>()

/**
 * 注册或移除指定位置的 Message Provider
 *
 * @param position - 位置键，如 'top' | 'top-left' | 'bottom' 等
 * @param provider - Provider 实例（含 message API）或 null（卸载时传 null）
 */
export function setMessageProvider(position: string, provider: { message: GMessageProviderApi } | null) {
  if (provider) {
    globalProviders.set(position, provider)
  } else {
    globalProviders.delete(position)
  }
}

/**
 * 按位置获取消息 API
 *
 * @param position - 位置，不传则用 'top'；若该位置无 Provider 则回退到 'top'
 * @returns 当前可用的 GMessageApi 或 null（未安装插件时）
 */
export function getMessageApi(position?: string): GMessageProviderApi | null {
  const target = position || 'top'
  const provider = globalProviders.get(target)
  if (provider) return provider.message
  const fallback = globalProviders.get('top')
  if (fallback) return fallback.message
  return null
}

/**
 * 创建并显示一条消息（内部使用）
 *
 * @param content - 消息内容或完整选项
 * @param type - 类型，默认 'info'
 * @param options - 额外选项，与 content 合并
 */
function createMessage(
  content: string | GMessageOptions,
  type: GMessageType = 'info',
  options?: GMessageOptions
) {
  const opts = typeof content === 'string' ? { content, ...options } : { ...content, ...options }
  opts.type = opts.type || type
  const position = opts.position || 'top'
  const api = getMessageApi(position)
  if (!api) return
  switch (opts.type) {
    case 'success':
      api.success(opts)
      break
    case 'error':
      api.error(opts)
      break
    case 'warning':
      api.warning(opts)
      break
    case 'loading':
      api.loading(opts)
      break
    default:
      api.info(opts)
  }
}

/**
 * 确认对话框（确定 + 取消）
 *
 * @param options - 对话框选项
 * @returns Promise<boolean> - 用户点击确定为 true，取消/关闭为 false
 */
function confirmDialog(options: GMessageDialogOptions): Promise<boolean> {
  return createDialog({
    title: options.title ?? '确认',
    content: options.content,
    width: options.width ?? 500,
    showCancel: options.showCancel !== false,
    showConfirm: options.showConfirm !== false,
    positiveText: options.confirmText ?? '确定',
    negativeText: options.cancelText ?? '取消',
    headerStyle: options.headerStyle ?? 'default',
  }).then((confirmed) => {
    if (confirmed && options.onConfirm) {
      return Promise.resolve(options.onConfirm!()).then(() => true)
    }
    if (!confirmed && options.onCancel) options.onCancel()
    return confirmed
  })
}

/**
 * 警告/提示对话框（仅确定）
 *
 * @param options - 对话框选项
 * @returns Promise<void>
 */
function alertDialog(options: GMessageDialogOptions): Promise<void> {
  return createDialog({
    title: options.title ?? '提示',
    content: options.content,
    width: options.width ?? 500,
    showCancel: false,
    showConfirm: true,
    positiveText: options.confirmText ?? '确定',
    headerStyle: options.headerStyle ?? 'default',
  }).then((confirmed) => {
    if (confirmed && options.onConfirm) return Promise.resolve(options.onConfirm!())
  })
}

/**
 * 输入提示对话框（待实现）
 *
 * @param _options - 选项，含 defaultValue 等
 * @returns Promise<string | null> - 当前返回 null
 */
function promptDialog(_options: GMessageDialogOptions & { defaultValue?: string }): Promise<string | null> {
  console.warn('[GMessage] prompt 功能待实现')
  return Promise.resolve(null)
}

const $gMessage = Object.assign(
  (content: string | GMessageOptions, options?: GMessageOptions) => {
    createMessage(content, 'info', options)
  },
  {
    success: (c: string | GMessageOptions, o?: GMessageOptions) => createMessage(c, 'success', o),
    error: (c: string | GMessageOptions, o?: GMessageOptions) => createMessage(c, 'error', o),
    warning: (c: string | GMessageOptions, o?: GMessageOptions) => createMessage(c, 'warning', o),
    info: (c: string | GMessageOptions, o?: GMessageOptions) => createMessage(c, 'info', o),
    loading: (c: string | GMessageOptions, o?: GMessageOptions) => createMessage(c, 'loading', o),
    destroyAll: () => getMessageApi()?.destroyAll(),
    confirm: confirmDialog,
    alert: alertDialog,
    prompt: promptDialog,
  }
) as GMessageApi

export { $gMessage }
