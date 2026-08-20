/**
 * 路由列表查询 Model（仅查询功能）
 * 复用 hub0021 的字段，但移除工具栏按钮和右键菜单
 */

import type { RsSearchFormProps } from '@/components/form/rs-search'
import type { RsGridColumn, RsGridMenuConfig, RsGridPaginationConfig } from '@/components/rs-grid'
import type { PageInfoObj } from '@/types/api'
import { formatDate } from '@/utils/format'
import type { RouteConfig } from '@/views/hub0021/components/routes/types'
import { MatchType } from '@/views/hub0021/components/routes/types'
import { RsTag, type RsTagVariant } from '@/ui'
import { h, ref } from 'vue'

/**
 * 路由列表表格配置（对齐 RsGrid Props 子集）。
 */
export interface RouteListGridConfig {
  columns: RsGridColumn<RouteConfig>[]
  selectable: boolean
  rowKey: string
  height: string
  paginationConfig: RsGridPaginationConfig
  menuConfig: RsGridMenuConfig
}

/**
 * 获取匹配类型标签变体
 */
function getMatchTypeTagType(matchType: number): RsTagVariant {
  const typeMap: Record<number, RsTagVariant> = {
    [MatchType.EXACT]: 'success',
    [MatchType.PREFIX]: 'info',
    [MatchType.REGEX]: 'warning',
  }
  return typeMap[matchType] || 'default'
}

/**
 * 获取匹配类型标签
 */
function getMatchTypeLabel(matchType: number): string {
  const labelMap: Record<number, string> = {
    [MatchType.EXACT]: '精确匹配',
    [MatchType.PREFIX]: '前缀匹配',
    [MatchType.REGEX]: '正则匹配',
  }
  return labelMap[matchType] || '未知'
}

/**
 * 路由列表查询 Model（仅查询功能）
 */
export function useRouteListModel() {
  const moduleId = 'hub0023'
  const loading = ref(false)
  const routeList = ref<RouteConfig[]>([])
  const pageInfo = ref<PageInfoObj | undefined>()

  const searchFormConfig: Omit<RsSearchFormProps, 'moduleId'> = {
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
    toolbarButtons: [],
    showSearchButton: true,
    showResetButton: true,
    resetButtonKey: 'resetQuery',
  }

  const gridConfig: RouteListGridConfig = {
    columns: [
      {
        key: 'routeConfigId',
        title: '路由配置ID',
        visible: false,
        width: 0,
      },
      {
        key: 'routeName',
        title: '路由名称',
        sortable: true,
        align: 'center',
        ellipsis: true,
        width: 200,
      },
      {
        key: 'routePath',
        title: '路由路径',
        align: 'center',
        ellipsis: true,
        width: 250,
      },
      {
        key: 'matchType',
        title: '匹配类型',
        align: 'center',
        width: 120,
        render: (row) =>
          h(
            RsTag,
            { variant: getMatchTypeTagType(row.matchType), size: 'sm' },
            () => getMatchTypeLabel(row.matchType),
          ),
      },
      {
        key: 'routePriority',
        title: '优先级',
        align: 'center',
        sortable: true,
        width: 100,
      },
      {
        key: 'serviceName',
        title: '关联服务',
        align: 'center',
        ellipsis: true,
        width: 180,
      },
      {
        key: 'activeFlag',
        title: '状态',
        align: 'center',
        width: 100,
        render: (row) =>
          h(
            RsTag,
            { variant: row.activeFlag === 'Y' ? 'success' : 'danger', size: 'sm' },
            () => (row.activeFlag === 'Y' ? '启用' : '禁用'),
          ),
      },
      {
        key: 'addTime',
        title: '创建时间',
        sortable: true,
        ellipsis: true,
        width: 180,
        formatter: (value) =>
          value ? formatDate(value as string, 'YYYY-MM-DD HH:mm:ss') : '',
      },
    ],
    selectable: false,
    rowKey: 'routeConfigId',
    height: '100%',
    paginationConfig: {
      show: true,
      pageInfo: pageInfo as any,
      align: 'right',
    },
    menuConfig: {
      enabled: false,
    },
  }

  const resetPagination = () => {
    pageInfo.value = undefined
  }

  const updatePagination = (newPageInfo: Partial<PageInfoObj>) => {
    if (!pageInfo.value) {
      pageInfo.value = newPageInfo as PageInfoObj
    } else {
      Object.assign(pageInfo.value, newPageInfo)
    }
  }

  const setRouteList = (list: RouteConfig[]) => {
    routeList.value = list
  }

  const clearRouteList = () => {
    routeList.value = []
  }

  return {
    moduleId,
    loading,
    routeList,
    pageInfo,
    searchFormConfig,
    gridConfig,
    resetPagination,
    updatePagination,
    setRouteList,
    clearRouteList,
    getMatchTypeTagType,
    getMatchTypeLabel,
  }
}

export type RouteListModel = ReturnType<typeof useRouteListModel>
