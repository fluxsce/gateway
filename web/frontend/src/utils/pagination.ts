import { PAGINATION_CONFIG } from '@/config'
import type { PageInfoObj } from '@/types/api'

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
  return {
    pageIndex: currentPage ?? PAGINATION_CONFIG.DEFAULT_PAGE_INDEX,
    pageSize: pageSize ?? PAGINATION_CONFIG.DEFAULT_PAGE_SIZE
  }
}
