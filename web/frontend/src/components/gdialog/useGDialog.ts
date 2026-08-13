/**
 * GDialog 程序化调用（兼容层）
 *
 * @deprecated 请直接使用 `@/ui` 的 `rsConfirm` / `rsConfirm.warning` 等。
 * 本模块仅映射旧字段（content/positiveText…）到 RsConfirmOptions。
 */

import { rsConfirm, type RsConfirmOptions, type RsFeedbackTone } from '@/ui'
import type { Component, VNode } from 'vue'
import { h, isVNode } from 'vue'

/**
 * 对话框选项（兼容旧 GDialogOptions；全部映射到 rsConfirm）
 */
export interface GDialogOptions {
  /** 对话框标题 */
  title?: string
  /** 次要说明 → rsConfirm.subtitle */
  subtitle?: string
  /** 头部图标组件 → rsConfirm.icon */
  icon?: Component
  /** @deprecated 确认框使用固定尺寸，忽略 */
  iconSize?: number
  /** @deprecated 确认框无渐变头，忽略 */
  headerStyle?: 'default' | 'gradient'
  /** 对话框宽度 → rsConfirm.width */
  width?: number | string
  /** 正文（字符串 → description；VNode/组件 → #extra） */
  content?: string | VNode | Component
  /** @deprecated 确认框默认不可点遮罩关闭 */
  maskClosable?: boolean
  /** 遮罩不透明度 0–1 */
  overlayOpacity?: number
  /** 遮罩模糊；number 为 px */
  overlayBlur?: number | string
  /** @deprecated 确认框 Esc 行为由组件内部管理 */
  closeOnEsc?: boolean
  /** 确认按钮文案 → confirmText */
  positiveText?: string
  /** 取消按钮文案 → cancelText */
  negativeText?: string
  /** 是否显示取消按钮 */
  showCancel?: boolean
  /** @deprecated 确认框始终显示确认按钮 */
  showConfirm?: boolean
  /** 确认按钮加载状态 */
  confirmLoading?: boolean
  /** @deprecated 确认框副标题固定在标题下 */
  subtitlePosition?: 'header' | 'footer'
  /** @deprecated 忽略 */
  footerButtonAlign?: 'start' | 'end' | 'center' | 'space-between' | 'space-around'
  /** 反馈色调 */
  tone?: RsFeedbackTone
  /** 确认按钮样式 */
  confirmVariant?: 'primary' | 'danger'
  /** 异步确认：完成后关闭；reject 则保持打开 */
  onConfirm?: () => void | Promise<void>
}

/**
 * @deprecated 命令式 rsConfirm 场景为 no-op
 */
export interface GDialogReactive {
  destroy: () => void
  setConfirmLoading: (loading: boolean) => void
}

function hasRichContent(content: GDialogOptions['content']): content is VNode | Component {
  return content != null && typeof content !== 'string'
}

function resolveStringContent(content: GDialogOptions['content']): string | undefined {
  return typeof content === 'string' && content ? content : undefined
}

function toConfirmOptions(options: GDialogOptions): RsConfirmOptions {
  const description = resolveStringContent(options.content)
  const tone = options.tone ?? 'danger'
  const confirmVariant =
    options.confirmVariant ?? (tone === 'danger' || tone === 'warning' ? 'danger' : 'primary')

  const mapped: RsConfirmOptions = {
    title: options.title,
    subtitle: options.subtitle,
    description,
    tone,
    icon: options.icon,
    width: options.width,
    confirmText: options.positiveText,
    cancelText: options.negativeText,
    confirmVariant,
    showCancel: options.showCancel,
    confirmLoading: options.confirmLoading,
    onConfirm: options.onConfirm,
    showOverlay: true,
    overlayOpacity: options.overlayOpacity,
    overlayBlur: options.overlayBlur,
  }

  if (hasRichContent(options.content)) {
    mapped.extra = () =>
      isVNode(options.content!) ? options.content : h(options.content as Component)
  }

  return mapped
}

/**
 * 创建确认框（程序化调用）
 *
 * @deprecated 请改用 `rsConfirm` / `rsConfirm.warning`
 * @returns Promise&lt;boolean&gt; - true 确认，false 取消/关闭
 */
export function createDialog(options: GDialogOptions): Promise<boolean> {
  if (import.meta.env.DEV) {
    console.warn('[useGDialog] 已弃用：请改用 rsConfirm（@/ui）。')
  }
  return rsConfirm(toConfirmOptions(options))
}

/** 程序化对话框 API 类型（useGDialog 返回值 / app.$gDialog） */
export interface GDialogApi {
  warning: (options: GDialogOptions | string) => Promise<boolean>
  info: (options: GDialogOptions | string) => Promise<boolean>
  success: (options: GDialogOptions | string) => Promise<boolean>
  error: (options: GDialogOptions | string) => Promise<boolean>
  confirm: (options: GDialogOptions | string) => Promise<boolean>
  create: (options: GDialogOptions) => Promise<boolean>
}

function withDefaults(
  options: GDialogOptions | string,
  defaults: GDialogOptions,
): GDialogOptions {
  return typeof options === 'string'
    ? { ...defaults, content: options }
    : { ...defaults, ...options }
}

const $gDialog: GDialogApi = {
  warning: (options) =>
    createDialog(
      withDefaults(options, {
        title: '警告',
        tone: 'danger',
        confirmVariant: 'danger',
        width: 500,
      }),
    ),
  info: (options) =>
    createDialog(
      withDefaults(options, {
        title: '提示',
        tone: 'info',
        confirmVariant: 'primary',
        showCancel: false,
        width: 500,
      }),
    ),
  success: (options) =>
    createDialog(
      withDefaults(options, {
        title: '成功',
        tone: 'success',
        confirmVariant: 'primary',
        showCancel: false,
        width: 500,
      }),
    ),
  error: (options) =>
    createDialog(
      withDefaults(options, {
        title: '错误',
        tone: 'danger',
        confirmVariant: 'danger',
        showCancel: false,
        width: 500,
      }),
    ),
  confirm: (options) =>
    createDialog(
      withDefaults(options, {
        title: '确认',
        tone: 'warning',
        confirmVariant: 'primary',
        width: 500,
      }),
    ),
  create: createDialog,
}

/**
 * @deprecated 请改用 `rsConfirm`（@/ui）
 */
export function useGDialog(): GDialogApi {
  return $gDialog
}

export { $gDialog }
