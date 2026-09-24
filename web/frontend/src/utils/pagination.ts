import { PAGINATION_CONFIG } from '@/config'
import type { PageInfoObj } from '@/types/api'

/** 代码允许的每页条数上限，与后端 PageSizeCeiling 一致。设置页不能超过它。 */
export const PAGE_SIZE_CEILING = 200

const PAGE_SIZE_STEPS = [10, 20, 50, 100, 200] as const

let pageSizePolicy = {
  defaultPageSize: PAGINATION_CONFIG.DEFAULT_PAGE_SIZE,
  maxPageSize: PAGE_SIZE_CEILING,
}

/** 用登录或环境设置里的租户分页策略覆盖前端档位。 */
export function applyPageSizePolicy(defaultPageSize?: number, maxPageSize?: number) {
  const max = clampInt(maxPageSize, 1, PAGE_SIZE_CEILING, PAGE_SIZE_CEILING)
  const fallback = PAGINATION_CONFIG.DEFAULT_PAGE_SIZE
  let def = clampInt(defaultPageSize, 1, max, fallback)
  if (def > max) {
    def = max
  }
  pageSizePolicy = { defaultPageSize: def, maxPageSize: max }
}

export function getDefaultPageSize() {
  return pageSizePolicy.defaultPageSize
}

export function getMaxPageSize() {
  return pageSizePolicy.maxPageSize
}

/** 下拉档位不超过租户上限。默认值不在标准档里时补进去。 */
export function pageSizeOptions(): number[] {
  const max = pageSizePolicy.maxPageSize
  const options: number[] = PAGE_SIZE_STEPS.filter((size) => size <= max)
  const def = pageSizePolicy.defaultPageSize
  if (def <= max && !options.includes(def)) {
    options.push(def)
    options.sort((a, b) => a - b)
  }
  return options.length > 0 ? options : [def]
}

function clampInt(value: number | undefined, min: number, max: number, fallback: number) {
  if (value == null || Number.isNaN(value)) {
    return fallback
  }
  const n = Math.floor(value)
  if (n < min) {
    return min
  }
  if (n > max) {
    return max
  }
  return n
}

/**
 * 分页配置（业务层 / Grid 配置用）。
 * UI 请使用 `@/ui` 的 RsPagination。
 */
export interface PaginationProps {
  /** 后端分页对象（优先于 currentPage / pageSize / total） */
  pageInfo?: PageInfoObj | import('vue').Ref<PageInfoObj | undefined>
  /** 当前页码（从 1 开始），对应 PageInfoObj.pageIndex */
  currentPage?: number
  /** 每页条数，对应 PageInfoObj.pageSize */
  pageSize?: number
  /** 总条数，对应 PageInfoObj.totalCount */
  total?: number
  /** 每页大小选项 */
  pageSizes?: number[]
  /** 对齐 */
  align?: 'left' | 'center' | 'right'
  /** 是否显示页码跳转（RsPagination showQuickJumper） */
  showJumper?: boolean
  /** 是否显示总数（RsPagination showSummary） */
  showTotal?: boolean
}

/**
 * 创建发送给后端的分页参数。
 */
export function createBackendPaginationParams(
  currentPage?: number,
  pageSize?: number
): {
  pageIndex: number
  pageSize: number
} {
  const max = pageSizePolicy.maxPageSize
  let size = pageSize ?? pageSizePolicy.defaultPageSize
  if (size < 1) {
    size = pageSizePolicy.defaultPageSize
  }
  if (size > max) {
    size = max
  }
  return {
    pageIndex: currentPage ?? PAGINATION_CONFIG.DEFAULT_PAGE_INDEX,
    pageSize: size,
  }
}
