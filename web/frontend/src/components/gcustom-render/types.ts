/**
 * GCustomRender 自定义渲染组件类型
 * 用于表格单元格、列表项等根据 scope 做自定义渲染
 */

import type { VNode } from 'vue'

/** 渲染上下文（如表格的 row、column、$rowIndex 等） */
export type CustomRenderScope = Record<string, any>

export interface GCustomRenderProps {
  /** 渲染上下文，会传给 render 或默认插槽 */
  scope?: CustomRenderScope
  /** 渲染函数，优先级高于默认插槽 */
  render?: (scope: CustomRenderScope) => VNode | string | null
}
