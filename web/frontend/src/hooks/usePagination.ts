/**
 * 通用分页 Hook
 * 封装所有分页相关逻辑，提供简单的 API
 */

import { PAGINATION_CONFIG } from '@/config'
import type { PageInfoObj } from '@/types/api'
import { computed, reactive } from 'vue'

export interface PaginationOptions {
  /** 初始页码 */
  initialPage?: number
  /** 初始每页数量 */
  initialPageSize?: number
  /** 每页数量选项 */
  pageSizes?: number[]
  /** 是否显示每页数量选择器 */
  showSizePicker?: boolean
  /** 是否显示快速跳转 */
  showQuickJumper?: boolean
  /** 页码变化回调 */
  onPageChange?: (page: number, pageSize: number) => void
  /** 每页数量变化回调 */
  onPageSizeChange?: (page: number, pageSize: number) => void
}

export const usePagination = (options: PaginationOptions = {}) => {
  const {
    initialPage = PAGINATION_CONFIG.DEFAULT_PAGE_INDEX,
    initialPageSize = PAGINATION_CONFIG.DEFAULT_PAGE_SIZE,
    pageSizes = PAGINATION_CONFIG.PAGE_SIZES,
    showSizePicker = PAGINATION_CONFIG.SHOW_SIZE_PICKER,
    showQuickJumper = PAGINATION_CONFIG.SHOW_QUICK_JUMPER,
    onPageChange,
    onPageSizeChange,
  } = options

  // 分页状态
  const pagination = reactive({
    page: initialPage,
    pageSize: initialPageSize,
    total: 0,
    totalPages: 0,
  })

  // 查询参数（用于后端请求）
  const queryParams = computed(() => ({
    pageIndex: pagination.page,
    pageSize: pagination.pageSize,
  }))

  // Naive UI 分页配置
  const naiveConfig = computed(() => ({
    page: pagination.page,
    pageSize: pagination.pageSize,
    itemCount: pagination.total,
    showSizePicker,
    pageSizes,
    showQuickJumper,
    prefix: () => `共 ${pagination.total} 条`,
    onUpdatePage: handlePageChange,
    onUpdatePageSize: handlePageSizeChange,
  }))

  // 页码变化处理
  const handlePageChange = (page: number) => {
    pagination.page = page
    onPageChange?.(page, pagination.pageSize)
  }

  // 每页数量变化处理
  const handlePageSizeChange = (pageSize: number) => {
    pagination.pageSize = pageSize
    pagination.page = 1 // 重置到第一页
    onPageSizeChange?.(1, pageSize)
  }

  // 更新分页信息（从后端响应更新）
  const updatePagination = (pageInfo: PageInfoObj) => {
    pagination.page = pageInfo.pageIndex || 1
    pagination.pageSize = pageInfo.pageSize || initialPageSize
    pagination.total = pageInfo.totalCount || 0
    pagination.totalPages = pageInfo.totalPageIndex || 0
  }

  // 重置分页
  const resetPagination = () => {
    pagination.page = initialPage
    pagination.pageSize = initialPageSize
    pagination.total = 0
    pagination.totalPages = 0
  }

  // 设置总数
  const setTotal = (total: number) => {
    pagination.total = total
    pagination.totalPages = Math.ceil(total / pagination.pageSize)
  }

  // 跳转到指定页
  const goToPage = (page: number) => {
    if (page >= 1 && page <= pagination.totalPages) {
      handlePageChange(page)
    }
  }

  // 获取当前页面信息
  const getCurrentPageInfo = () => ({
    page: pagination.page,
    pageSize: pagination.pageSize,
    total: pagination.total,
    totalPages: pagination.totalPages,
    hasNext: pagination.page < pagination.totalPages,
    hasPrev: pagination.page > 1,
  })

  // 清理资源，防止内存泄露
  const cleanup = () => {
    // 虽然这些是局部变量，但为了确保完全清理，我们可以重置相关引用
    // 注意：在这个hook中没有需要特别清理的资源，但保留此方法供将来扩展
  }

  return {
    // 状态
    pagination: readonly(pagination),
    queryParams,
    naiveConfig,

    // 方法
    updatePagination,
    resetPagination,
    setTotal,
    goToPage,
    getCurrentPageInfo,
    cleanup,

    // 事件处理
    handlePageChange,
    handlePageSizeChange,
  }
}

// 只读包装器
function readonly<T>(obj: T): Readonly<T> {
  return obj as Readonly<T>
}
