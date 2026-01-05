/**
 * 路由列表查询 Model（仅查询功能）
 * 复用 hub0021 的 model，但移除工具栏按钮和右键菜单
 */

import type { SearchFormProps } from '@/components/form/search/types'
import type { GridProps } from '@/components/grid'
import type { PageInfoObj } from '@/types/api'
import { formatDate } from '@/utils/format'
import { ref } from 'vue'
import type { RouteConfig } from '@/views/hub0021/components/routes/types'
import { MatchType } from '@/views/hub0021/components/routes/types'

/**
 * 路由列表查询 Model（仅查询功能）
 */
export function useRouteListModel() {
  // ============= 数据状态 =============
  const moduleId = 'hub0023-route-list'
  /** 加载状态 */
  const loading = ref(false)

  /** 路由列表数据 */
  const routeList = ref<RouteConfig[]>([])

  /** 后端分页信息对象 */
  const pageInfo = ref<PageInfoObj | undefined>()

  // ============= 搜索表单配置 =============

  /** 搜索表单配置（符合 SearchFormProps 结构，移除工具栏按钮） */
  const searchFormConfig: Omit<SearchFormProps, 'moduleId'> = {
    fields: [
      {
        field: 'routeName',
        label: '路由名称',
        type: 'input',
        placeholder: '请输入路由名称',
        span: 6,
        clearable: true,
      },
      {
        field: 'routePath',
        label: '路由路径',
        type: 'input',
        placeholder: '请输入路由路径',
        span: 6,
        clearable: true,
      },
      {
        field: 'matchType',
        label: '匹配类型',
        type: 'select',
        placeholder: '请选择匹配类型',
        span: 6,
        clearable: true,
        options: [
          { label: '精确匹配', value: MatchType.EXACT },
          { label: '前缀匹配', value: MatchType.PREFIX },
          { label: '正则匹配', value: MatchType.REGEX },
        ],
      },
      {
        field: 'activeFlag',
        label: '状态',
        type: 'select',
        placeholder: '请选择状态',
        span: 6,
        clearable: true,
        options: [
          { label: '全部', value: '' },
          { label: '启用', value: 'Y' },
          { label: '禁用', value: 'N' },
        ],
      },
    ],
    // 移除工具栏按钮，只保留查询和重置
    toolbarButtons: [],
    showSearchButton: true,
    showResetButton: true,
  }

  // ============= 表格配置 =============

  /** 表格配置（符合 GridProps 结构，排除响应式数据，移除右键菜单） */
  const gridConfig: Omit<GridProps, 'moduleId' | 'data' | 'loading'> = {
    columns: [
      {
        field: 'routeConfigId',
        title: '路由配置ID',
        visible: false,
        width: 0,
      },
      {
        field: 'routeName',
        title: '路由名称',
        sortable: true,
        align: 'center',
        showOverflow: 'tooltip',
        width: 200,
      },
      {
        field: 'routePath',
        title: '路由路径',
        align: 'center',
        showOverflow: 'tooltip',
        width: 250,
      },
      {
        field: 'matchType',
        title: '匹配类型',
        align: 'center',
        slots: { default: 'matchType' },
        width: 120,
      },
      {
        field: 'routePriority',
        title: '优先级',
        align: 'center',
        sortable: true,
        width: 100,
      },
      {
        field: 'serviceName',
        title: '关联服务',
        align: 'center',
        showOverflow: 'tooltip',
        width: 180,
      },
      {
        field: 'activeFlag',
        title: '状态',
        align: 'center',
        slots: { default: 'activeFlag' },
        width: 100,
      },
      {
        field: 'addTime',
        title: '创建时间',
        sortable: true,
        showOverflow: true,
        formatter: ({ cellValue }) =>
          cellValue ? formatDate(cellValue, 'YYYY-MM-DD HH:mm:ss') : '',
        width: 180,
      },
    ],
    showCheckbox: false, // 查询模式不需要复选框
    paginationConfig: {
      show: true,
      pageInfo: pageInfo as any,
      align: 'right',
    },
    // 移除右键菜单
    menuConfig: {
      enabled: false,
    },
    height: '100%',
  }

  // ============= 辅助方法 =============

  /**
   * 重置分页
   */
  const resetPagination = () => {
    pageInfo.value = undefined
  }

  /**
   * 更新分页信息（接收后端 PageInfoObj）
   */
  const updatePagination = (newPageInfo: Partial<PageInfoObj>) => {
    if (!pageInfo.value) {
      pageInfo.value = newPageInfo as PageInfoObj
    } else {
      Object.assign(pageInfo.value, newPageInfo)
    }
  }

  /**
   * 设置路由列表
   */
  const setRouteList = (list: RouteConfig[]) => {
    routeList.value = list
  }

  /**
   * 清空路由列表
   */
  const clearRouteList = () => {
    routeList.value = []
  }

  /**
   * 获取匹配类型标签类型
   */
  const getMatchTypeTagType = (matchType: number): 'success' | 'info' | 'warning' | 'default' => {
    const typeMap: Record<number, 'success' | 'info' | 'warning' | 'default'> = {
      [MatchType.EXACT]: 'success',
      [MatchType.PREFIX]: 'info',
      [MatchType.REGEX]: 'warning',
    }
    return typeMap[matchType] || 'default'
  }

  /**
   * 获取匹配类型标签
   */
  const getMatchTypeLabel = (matchType: number): string => {
    const labelMap: Record<number, string> = {
      [MatchType.EXACT]: '精确匹配',
      [MatchType.PREFIX]: '前缀匹配',
      [MatchType.REGEX]: '正则匹配',
    }
    return labelMap[matchType] || '未知'
  }

  return {
    // 基本信息
    moduleId,

    // 数据状态
    loading,
    routeList,
    pageInfo,

    // 配置
    searchFormConfig,
    gridConfig,

    // 方法
    resetPagination,
    updatePagination,
    setRouteList,
    clearRouteList,
    getMatchTypeTagType,
    getMatchTypeLabel,
  }
}

/**
 * Model 返回类型
 */
export type RouteListModel = ReturnType<typeof useRouteListModel>

