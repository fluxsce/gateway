import { PAGINATION_CONFIG } from '@/config'
import type { PageInfoObj } from '@/types/api'

/**
 * 分页配置属性
 * 对接后端 PageInfoObj 结构
 */
export interface PaginationProps {
  /**
   * 直接传入后端分页对象（优先级最高）
   * 如果传入此属性，将自动解析并覆盖 currentPage、pageSize、total
   * 支持 ref 或普通对象
   */
  pageInfo?: PageInfoObj | import('vue').Ref<PageInfoObj | undefined>

  /**
   * 当前页码（从1开始）
   * 对应后端 PageInfoObj.pageIndex
   * @default 1
   */
  currentPage?: number

  /**
   * 每页记录数
   * 对应后端 PageInfoObj.pageSize
   * @default 20
   */
  pageSize?: number

  /**
   * 总记录数
   * 对应后端 PageInfoObj.totalCount
   * @default 0
   */
  total?: number

  /**
   * 每页大小选项
   * @default [10, 20, 50, 100, 200]
   */
  pageSizes?: number[]

  /**
   * 布局配置
   * @default ['PrevJump', 'PrevPage', 'Number', 'NextPage', 'NextJump', 'Sizes', 'FullJump', 'Total']
   */
  layouts?: Array<
    | 'PrevJump'
    | 'PrevPage'
    | 'Number'
    | 'NextPage'
    | 'NextJump'
    | 'Sizes'
    | 'FullJump'
    | 'Total'
  >

  /**
   * 对齐方式
   * @default 'right'
   */
  align?: 'left' | 'center' | 'right'

  /**
   * 是否显示背景色
   * @default true
   */
  background?: boolean

  /**
   * 是否完美模式（自动隐藏不必要的按钮）
   * @default true
   */
  perfect?: boolean

  /**
   * 是否显示页码跳转
   * @default true
   */
  showJumper?: boolean

  /**
   * 是否显示总数
   * @default true
   */
  showTotal?: boolean

  /**
   * 其他 vxe-pager 原生配置
   */
  pagerOptions?: Record<string, any>
}

/**
 * 分页组件事件
 */
export interface PaginationEmits {
  /**
   * 分页变化事件
   */
  (event: 'page-change', params: { currentPage: number; pageSize: number }): void

  /**
   * 更新当前页码（用于 v-model:currentPage）
   */
  (event: 'update:currentPage', value: number): void

  /**
   * 更新每页大小（用于 v-model:pageSize）
   */
  (event: 'update:pageSize', value: number): void
}

/**
 * 从后端 PageInfoObj 创建前端分页配置
 * @param pageInfo 后端分页信息对象
 * @returns 前端分页配置
 */
export function createPaginationFromBackend(pageInfo: PageInfoObj): {
  currentPage: number
  pageSize: number
  total: number
  totalPages: number
} {
  return {
    currentPage: pageInfo.pageIndex,
    pageSize: pageInfo.pageSize,
    total: pageInfo.totalCount,
    totalPages: pageInfo.totalPageIndex
  }
}

/**
 * 创建发送给后端的分页参数
 * @param currentPage 当前页码（可选，默认使用配置常量）
 * @param pageSize 每页大小（可选，默认使用配置常量）
 * @returns 后端分页参数
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

/**
 * 分页辅助工具类
 */
export class PaginationHelper {
  /**
   * 计算总页数
   */
  static getTotalPages(total: number, pageSize: number): number {
    return Math.ceil(total / pageSize)
  }

  /**
   * 验证页码是否有效
   */
  static isValidPage(page: number, totalPages: number): boolean {
    return page >= 1 && page <= totalPages
  }

  /**
   * 获取数据切片的起始索引
   */
  static getStartIndex(currentPage: number, pageSize: number): number {
    return (currentPage - 1) * pageSize
  }

  /**
   * 获取数据切片的结束索引
   */
  static getEndIndex(currentPage: number, pageSize: number): number {
    return currentPage * pageSize
  }

  /**
   * 从数组中获取当前页数据
   */
  static getPageData<T>(data: T[], currentPage: number, pageSize: number): T[] {
    const start = this.getStartIndex(currentPage, pageSize)
    const end = this.getEndIndex(currentPage, pageSize)
    return data.slice(start, end)
  }
}

