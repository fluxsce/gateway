/**
 * 服务定义选择器 Model
 * 统一管理搜索表单、表格配置和数据状态
 */

import type { SearchFormProps } from '@/components/form/search/types'
import type { GridProps } from '@/components/grid'
import type { PageInfoObj } from '@/types/api'
import { formatDate } from '@/utils/format'
import { ref } from 'vue'
import type { ServiceDefinition } from '../types'

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

  /** 搜索表单配置（符合 SearchFormProps 结构） */
  const searchFormConfig: Omit<SearchFormProps, 'moduleId'> = {
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

  /** 表格配置（符合 GridProps 结构，排除响应式数据） */
  const gridConfig: Omit<GridProps, 'moduleId' | 'data' | 'loading'> = {
    paginationConfig: {
      show: true, // 显示分页
      pageInfo: pageInfo, // 传递分页信息 ref
    },
    columns: [
      {
        field: 'serviceDefinitionId',
        title: '服务ID',
        align: 'center',
        showOverflow: 'tooltip',
        width: 200,
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
        width: 250,
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
        title: '负载均衡',
        align: 'center',
        slots: { default: 'loadBalanceStrategy' },
        width: 150,
      },
      {
        field: 'healthCheckEnabled',
        title: '健康检查',
        align: 'center',
        slots: { default: 'healthCheckEnabled' },
        width: 120,
      },
      {
        field: 'activeFlag',
        title: '状态',
        align: 'center',
        slots: { default: 'activeFlag' },
        width: 100,
      },
      {
        field: 'editTime',
        title: '修改时间',
        sortable: true,
        align: 'center',
        showOverflow: 'tooltip',
        formatter: ({ cellValue }) =>
          cellValue ? formatDate(cellValue, 'YYYY-MM-DD HH:mm:ss') : '',
        width: 180,
      },
      {
        field: 'editWho',
        title: '修改人',
        align: 'center',
        showOverflow: 'tooltip',
        width: 120,
      },
    ],
    showCheckbox: true,
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

