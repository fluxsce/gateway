/**
 * 服务列表查询 Model（仅查询功能）
 * 复用 hub0022 的字段，但移除工具栏按钮和右键菜单
 */

import type { RsSearchFormProps } from '@/components/form/rs-search'
import type { RsGridColumn, RsGridMenuConfig, RsGridPaginationConfig } from '@/components/rs-grid'
import type { PageInfoObj } from '@/types/api'
import { formatDate } from '@/utils/format'
import type { ServiceDefinition } from '@/views/hub0022/components/service/types'
import { LoadBalanceStrategy, ServiceType } from '@/views/hub0022/components/service/types'
import { RsTag } from '@/ui'
import { h, ref } from 'vue'

/**
 * 服务列表表格配置（对齐 RsGrid Props 子集）。
 */
export interface ServiceListGridConfig {
  columns: RsGridColumn<ServiceDefinition>[]
  selectable: boolean
  rowKey: string
  height: string
  paginationConfig: RsGridPaginationConfig
  menuConfig: RsGridMenuConfig
}

/**
 * 获取负载均衡策略标签
 */
function getLoadBalanceStrategyLabel(strategy: string): string {
  const strategyMap: Record<string, string> = {
    [LoadBalanceStrategy.ROUND_ROBIN]: '轮询',
    [LoadBalanceStrategy.RANDOM]: '随机',
    [LoadBalanceStrategy.IP_HASH]: 'IP哈希',
    [LoadBalanceStrategy.LEAST_CONN]: '最少连接',
    [LoadBalanceStrategy.WEIGHTED_ROUND_ROBIN]: '加权轮询',
    [LoadBalanceStrategy.CONSISTENT_HASH]: '一致性哈希',
  }
  return strategyMap[strategy] || strategy
}

/**
 * 服务列表查询 Model（仅查询功能）
 */
export function useServiceListModel() {
  const moduleId = 'hub0023'
  const loading = ref(false)
  const serviceList = ref<ServiceDefinition[]>([])
  const pageInfo = ref<PageInfoObj | undefined>()

  const searchFormConfig: Omit<RsSearchFormProps, 'moduleId'> = {
    fields: [
      {
        field: 'serviceName',
        label: '服务名称',
        type: 'input',
        placeholder: '请输入服务名称',
        span: 6,
        clearable: true,
      },
      {
        field: 'serviceType',
        label: '服务类型',
        type: 'select',
        placeholder: '请选择服务类型',
        span: 6,
        clearable: true,
        options: [
          { label: '静态配置', value: ServiceType.STATIC },
          { label: '服务发现', value: ServiceType.DISCOVERY },
        ],
      },
      {
        field: 'loadBalanceStrategy',
        label: '负载均衡策略',
        type: 'select',
        placeholder: '请选择负载均衡策略',
        span: 6,
        clearable: true,
        options: [
          { label: '轮询算法', value: LoadBalanceStrategy.ROUND_ROBIN },
          { label: '随机算法', value: LoadBalanceStrategy.RANDOM },
          { label: 'IP哈希算法', value: LoadBalanceStrategy.IP_HASH },
          { label: '最少连接算法', value: LoadBalanceStrategy.LEAST_CONN },
          { label: '加权轮询算法', value: LoadBalanceStrategy.WEIGHTED_ROUND_ROBIN },
          { label: '一致性哈希算法', value: LoadBalanceStrategy.CONSISTENT_HASH },
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

  const gridConfig: ServiceListGridConfig = {
    columns: [
      {
        key: 'serviceDefinitionId',
        title: '服务定义ID',
        visible: false,
        width: 0,
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
        width: 200,
      },
      {
        key: 'serviceType',
        title: '服务类型',
        align: 'center',
        width: 120,
        render: (row) =>
          h(
            RsTag,
            { variant: row.serviceType === ServiceType.STATIC ? 'info' : 'success', size: 'sm' },
            () => (row.serviceType === ServiceType.STATIC ? '静态配置' : '服务发现'),
          ),
      },
      {
        key: 'loadBalanceStrategy',
        title: '负载均衡策略',
        align: 'center',
        width: 150,
        render: (row) =>
          h(RsTag, { variant: 'default', size: 'sm' }, () =>
            getLoadBalanceStrategyLabel(row.loadBalanceStrategy),
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
    rowKey: 'serviceDefinitionId',
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

  const setServiceList = (list: ServiceDefinition[]) => {
    serviceList.value = list
  }

  const clearServiceList = () => {
    serviceList.value = []
  }

  return {
    moduleId,
    loading,
    serviceList,
    pageInfo,
    searchFormConfig,
    gridConfig,
    resetPagination,
    updatePagination,
    setServiceList,
    clearServiceList,
    getLoadBalanceStrategyLabel,
  }
}

export type ServiceListModel = ReturnType<typeof useServiceListModel>
