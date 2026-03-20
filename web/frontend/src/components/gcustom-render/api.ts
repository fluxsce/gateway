/**
 * 全局自定义渲染 API：在根部渲染自定义组件（对话框、抽屉等），任意 TS 可直接调用
 */

import { shallowRef } from 'vue'
import type { Component } from 'vue'

export interface GlobalCustomRenderOptions {
  onClose?: () => void
  onSuccess?: (data?: unknown) => void
}

export interface GlobalCustomRenderItem {
  component: Component
  props: Record<string, unknown>
  options?: GlobalCustomRenderOptions
}

export interface GlobalCustomRenderApi {
  show(
    component: Component,
    props?: Record<string, unknown>,
    options?: GlobalCustomRenderOptions
  ): void
  close(): void
  closeWithSuccess(data?: unknown): void
}

const currentRef = shallowRef<GlobalCustomRenderItem | null>(null)

function show(
  component: Component,
  props: Record<string, unknown> = {},
  options?: GlobalCustomRenderOptions
) {
  currentRef.value = { component, props, options }
}

function close() {
  const opts = currentRef.value?.options
  currentRef.value = null
  opts?.onClose?.()
}

function closeWithSuccess(data?: unknown) {
  const opts = currentRef.value?.options
  opts?.onSuccess?.(data)
  close()
}

export const current = currentRef

export const $gRender: GlobalCustomRenderApi = {
  show,
  close,
  closeWithSuccess,
}
