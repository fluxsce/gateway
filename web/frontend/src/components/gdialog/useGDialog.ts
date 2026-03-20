/**
 * GDialog 程序化调用
 *
 * 与 GMessage 一致：由 GDialogProvider 在应用树内渲染，自动继承 NConfigProvider 主题与语言，随系统切换。
 * 安装：app.use(gdialogPlugin) 时插件会挂载 GDialogProvider。
 */

import type { Component, VNode } from 'vue'

/** Provider 挂载时注册的 API，供 createDialog 委托 */
export interface GDialogProviderApi {
  createDialog: (options: GDialogOptions) => Promise<boolean>
}

let dialogProvider: GDialogProviderApi | null = null

export function setDialogProvider(api: GDialogProviderApi | null) {
  dialogProvider = api
}

export function getDialogApi(): GDialogProviderApi | null {
  return dialogProvider
}

/**
 * 对话框选项
 */
export interface GDialogOptions {
  /** 对话框标题 */
  title?: string
  /** 对话框副标题 */
  subtitle?: string
  /** 头部图标组件 */
  icon?: Component
  /** 图标大小 */
  iconSize?: number
  /** 头部样式类型 */
  headerStyle?: 'default' | 'gradient'
  /** 对话框宽度 */
  width?: number | string
  /** 内容（可以是字符串、VNode 或组件） */
  content?: string | VNode | Component
  /** 是否可点击遮罩关闭 */
  maskClosable?: boolean
  /** 是否可按下 ESC 关闭 */
  closeOnEsc?: boolean
  /** 确认按钮文案 */
  positiveText?: string
  /** 取消按钮文案 */
  negativeText?: string
  /** 是否显示取消按钮 */
  showCancel?: boolean
  /** 是否显示确认按钮 */
  showConfirm?: boolean
  /** 确认按钮加载状态 */
  confirmLoading?: boolean
  /** 副标题显示位置：'header' 显示在头部 | 'footer' 显示在底部 */
  subtitlePosition?: 'header' | 'footer'
  /** 底部操作按钮对齐方式 */
  footerButtonAlign?: 'start' | 'end' | 'center' | 'space-between' | 'space-around'
}

/**
 * 对话框实例
 */
export interface GDialogReactive {
  destroy: () => void
  setConfirmLoading: (loading: boolean) => void
}

/**
 * 创建对话框（程序化调用 GDialog）
 *
 * 委托给 GDialogProvider，在应用树内渲染，自动跟随系统主题与语言。
 *
 * @param options - GDialog 选项
 * @returns Promise<boolean> - true 确认，false 取消/关闭；未挂载 Provider 时返回 false
 */
export function createDialog(options: GDialogOptions): Promise<boolean> {
  if (!dialogProvider) return Promise.resolve(false)
  return dialogProvider.createDialog(options)
}

/** 程序化对话框 API 类型（useGDialog 返回值 / window.$gDialog） */
export interface GDialogApi {
  warning: (options: GDialogOptions | string) => Promise<boolean>
  info: (options: GDialogOptions | string) => Promise<boolean>
  success: (options: GDialogOptions | string) => Promise<boolean>
  error: (options: GDialogOptions | string) => Promise<boolean>
  confirm: (options: GDialogOptions | string) => Promise<boolean>
  create: (options: GDialogOptions) => Promise<boolean>
}

const $gDialog: GDialogApi = {
  warning: (options) =>
    typeof options === 'string'
      ? createDialog({ title: '警告', content: options, headerStyle: 'gradient', width: 500 })
      : createDialog({ title: '警告', headerStyle: 'gradient', width: 500, ...options }),
  info: (options) =>
    typeof options === 'string'
      ? createDialog({ title: '提示', content: options, width: 500 })
      : createDialog({ title: '提示', width: 500, ...options }),
  success: (options) =>
    typeof options === 'string'
      ? createDialog({ title: '成功', content: options, headerStyle: 'gradient', width: 500, showCancel: false })
      : createDialog({ title: '成功', headerStyle: 'gradient', width: 500, showCancel: false, ...options }),
  error: (options) =>
    typeof options === 'string'
      ? createDialog({ title: '错误', content: options, headerStyle: 'gradient', width: 500, showCancel: false })
      : createDialog({ title: '错误', headerStyle: 'gradient', width: 500, showCancel: false, ...options }),
  confirm: (options) =>
    typeof options === 'string'
      ? createDialog({ title: '确认', content: options, width: 500 })
      : createDialog({ title: '确认', width: 500, ...options }),
  create: createDialog,
}

/**
 * 使用 GDialog Hook
 */
export function useGDialog(): GDialogApi {
  return $gDialog
}

export { $gDialog }
