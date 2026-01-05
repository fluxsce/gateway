/**
 * 服务列表查询 Model（仅查询功能）
 * 复用 hub0022 的 model，但移除工具栏按钮和右键菜单
 */

import type { SearchFormProps } from '@/components/form/search/types'
import type { GridProps } from '@/components/grid'
import type { PageInfoObj } from '@/types/api'
import { formatDate } from '@/utils/format'
import { ref } from 'vue'
import type { ServiceDefinition } from '@/views/hub0022/components/service/types'
import { LoadBalanceStrategy, ServiceType } from '@/views/hub0022/components/service/types'

/**
 * 服务列表查询 Model（仅查询功能）
 */
export function useServiceListModel() {
  // ============= 数据状态 =============
  const moduleId = 'hub0023-service-list'
  /** 加载状态 */
  const loading = ref(false)

  /** 服务列表数据 */
  const serviceList = ref<ServiceDefinition[]>([])

  /** 后端分页信息对象 */
  const pageInfo = ref<PageInfoObj | undefined>()

  // ============= 搜索表单配置 =============

  /** 搜索表单配置（符合 SearchFormProps 结构，移除工具栏按钮） */
  const searchFormConfig: Omit<SearchFormProps, 'moduleId'> = {
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
        field: 'serviceDefinitionId',
        title: '服务定义ID',
        visible: false,
        width: 0,
      },
      {
        field: 'serviceName',
        title: '服务名称',
        sortable: true,
        align: 'center',
        showOverflow: 'tooltip',
        width: 200,
      },
      {
        field: 'serviceDesc',
        title: '服务描述',
        align: 'center',
        showOverflow: 'tooltip',
        width: 200,
      },
      {
        field: 'serviceType',
        title: '服务类型',
        align: 'center',
        slots: { default: 'serviceType' },
        width: 120,
      },
      {
        field: 'loadBalanceStrategy',
        title: '负载均衡策略',
        align: 'center',
        slots: { default: 'loadBalanceStrategy' },
        width: 150,
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
   * 设置服务列表
   */
  const setServiceList = (list: ServiceDefinition[]) => {
    serviceList.value = list
  }

  /**
   * 清空服务列表
   */
  const clearServiceList = () => {
    serviceList.value = []
  }

  /**
   * 获取负载均衡策略标签
   */
  const getLoadBalanceStrategyLabel = (strategy: string): string => {
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

  return {
    // 基本信息
    moduleId,

    // 数据状态
    loading,
    serviceList,
    pageInfo,

    // 配置
    searchFormConfig,
    gridConfig,

    // 方法
    resetPagination,
    updatePagination,
    setServiceList,
    clearServiceList,
    getLoadBalanceStrategyLabel,
  }
}

/**
 * Model 返回类型
 */
export type ServiceListModel = ReturnType<typeof useServiceListModel>

