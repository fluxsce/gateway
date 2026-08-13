/**
 * 服务定义选择器 Model
 * 统一管理搜索表单、表格配置和数据状态
 */

import type { RsSearchFormProps } from '@/components/form/rs-search'
import type { RsGridColumn, RsGridMenuConfig, RsGridPaginationConfig } from '@/components/rs-grid'
import type { PageInfoObj } from '@/types/api'
import { formatDate } from '@/utils/format'
import { RsTag } from '@/ui'
import { h, ref } from 'vue'
import type { ServiceDefinition } from '../types'

/**
 * 服务定义选择器表格配置（对齐 RsGrid Props 子集）。
 */
export interface ServiceDefinitionSelectorGridConfig {
  columns: RsGridColumn<ServiceDefinition>[]
  selectable: boolean
  rowKey: string
  height: string
  paginationConfig: RsGridPaginationConfig
  menuConfig: RsGridMenuConfig
}

const loadBalanceLabelMap: Record<string, string> = {
  'round-robin': '轮询',
  random: '随机',
  'ip-hash': 'IP哈希',
  'least-conn': '最少连接',
  'weighted-round-robin': '加权轮询',
  'consistent-hash': '一致性哈希',
  ROUND_ROBIN: '轮询',
  RANDOM: '随机',
  IP_HASH: 'IP哈希',
  LEAST_CONN: '最少连接',
  WEIGHTED_ROUND_ROBIN: '加权轮询',
  CONSISTENT_HASH: '一致性哈希',
}

/**
 * 服务定义选择器 Model
 */
export function useServiceDefinitionSelectorModel() {
  // ============= 数据状态 =============
  const moduleId = 'hub0021-service-selector'

  /** 加载状态 */
  const loading = ref(false)

  /** 服务定义列表数据 */
  const serviceList = ref<ServiceDefinition[]>([])

  /** 后端分页信息对象 */
  const pageInfo = ref<PageInfoObj | undefined>()

  // ============= 搜索表单配置 =============

  /** 搜索表单配置（符合 RsSearchFormProps 结构） */
  const searchFormConfig: Omit<RsSearchFormProps, 'moduleId'> = {
    fields: [
      {
        field: 'serviceName',
        label: '服务名称',
        type: 'input',
        placeholder: '请输入服务名称',
        span: 8,
        clearable: true,
      },
      {
        field: 'serviceDefinitionId',
        label: '服务ID',
        type: 'input',
        placeholder: '请输入服务ID',
        span: 8,
        clearable: true,
      },
      {
        field: 'serviceType',
        label: '服务类型',
        type: 'select',
        placeholder: '请选择服务类型',
        span: 8,
        clearable: true,
        options: [
          { label: '全部', value: '' },
          { label: '服务发现', value: 1 },
          { label: '静态配置', value: 0 },
        ],
      },
    ],
    showSearchButton: true,
    showResetButton: true,
  }

  // ============= 表格配置 =============

  /** 表格配置（符合 RsGrid 结构） */
  const gridConfig: ServiceDefinitionSelectorGridConfig = {
    columns: [
      {
        key: 'serviceDefinitionId',
        title: '服务ID',
        align: 'center',
        ellipsis: true,
        width: 200,
      },
      {
        key: 'serviceName',
        title: '服务名称',
        sortable: true,
        align: 'center',
        ellipsis: true,
        width: 200,
      },
      {
        key: 'serviceDesc',
        title: '服务描述',
        align: 'center',
        ellipsis: true,
        width: 250,
      },
      {
        key: 'serviceType',
        title: '服务类型',
        align: 'center',
        width: 120,
        render: (row) =>
          h(
            RsTag,
            {
              variant: row.serviceType === 1 ? 'success' : 'info',
              size: 'sm',
            },
            () => (row.serviceType === 1 ? '服务发现' : '静态配置'),
          ),
      },
      {
        key: 'loadBalanceStrategy',
        title: '负载均衡',
        align: 'center',
        width: 150,
        render: (row) =>
          h(RsTag, { variant: 'default', size: 'sm' }, () =>
            loadBalanceLabelMap[row.loadBalanceStrategy] || row.loadBalanceStrategy,
          ),
      },
      {
        key: 'healthCheckEnabled',
        title: '健康检查',
        align: 'center',
        width: 120,
        render: (row) =>
          h(
            RsTag,
            {
              variant: row.healthCheckEnabled === 'Y' ? 'success' : 'default',
              size: 'sm',
            },
            () => (row.healthCheckEnabled === 'Y' ? '已启用' : '未启用'),
          ),
      },
      {
        key: 'activeFlag',
        title: '状态',
        align: 'center',
        width: 100,
        render: (row) =>
          h(
            RsTag,
            {
              variant: row.activeFlag === 'Y' ? 'success' : 'danger',
              size: 'sm',
            },
            () => (row.activeFlag === 'Y' ? '启用' : '禁用'),
          ),
      },
      {
        key: 'editTime',
        title: '修改时间',
        sortable: true,
        align: 'center',
        ellipsis: true,
        formatter: (value) =>
          value ? formatDate(value as string, 'YYYY-MM-DD HH:mm:ss') : '',
        width: 180,
      },
      {
        key: 'editWho',
        title: '修改人',
        align: 'center',
        ellipsis: true,
        width: 120,
      },
    ],
    selectable: true,
    rowKey: 'serviceDefinitionId',
    height: '100%',
    paginationConfig: {
      show: true,
      pageInfo: pageInfo as any,
    },
    menuConfig: {
      enabled: false,
      items: [],
    },
  }

  // ============= 状态更新方法 =============

  /**
   * 设置服务列表
   */
  function setServiceList(list: ServiceDefinition[]) {
    serviceList.value = list
  }

  /**
   * 设置加载状态
   */
  function setLoading(value: boolean) {
    loading.value = value
  }

  /**
   * 更新分页信息（接收后端 PageInfoObj）
   */
  function updatePagination(newPageInfo: Partial<PageInfoObj>) {
    if (!pageInfo.value) {
      pageInfo.value = newPageInfo as PageInfoObj
    } else {
      Object.assign(pageInfo.value, newPageInfo)
    }
  }

  /**
   * 重置分页信息
   */
  function resetPagination() {
    pageInfo.value = undefined
  }

  return {
    // 状态
    moduleId,
    loading,
    serviceList,
    pageInfo,

    // 配置
    searchFormConfig,
    gridConfig,

    // 方法
    setServiceList,
    setLoading,
    updatePagination,
    resetPagination,
  }
}

/**
 * 服务定义选择器 Model 类型
 */
export type ServiceDefinitionSelectorModel = ReturnType<typeof useServiceDefinitionSelectorModel>
