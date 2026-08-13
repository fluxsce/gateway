/**
 * 遗留页面使用的过滤器类型兼容导出。
 * 新代码请优先使用 components/filter-config/hooks/types。
 */

import type { FilterConfig, FilterConfigForm } from './index'

export type { FilterConfig }

/** 过滤器编辑表单数据（含可选主键） */
export interface FilterFormData extends FilterConfigForm {
  filterConfigId?: string
  config?: Record<string, unknown>
}
