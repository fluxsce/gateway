/**
 * 网关实例列表查询 Model（仅查询功能）
 * 复用 hub0020 的 model，但移除工具栏按钮和右键菜单
 */

import type { SearchFormProps } from '@/components/form/search/types'
import type { GridProps } from '@/components/grid'
import type { PageInfoObj } from '@/types/api'
import { formatDate } from '@/utils/format'
import type { GatewayInstance } from '@/views/hub0020/types'
import { ref } from 'vue'

/**
 * 网关实例列表查询 Model（仅查询功能）
 */
export function useGatewayInstanceListModel() {
  // ============= 数据状态 =============
  const moduleId = 'hub0023-instance-list'
  /** 加载状态 */
  const loading = ref(false)

  /** 网关实例列表数据 */
  const instanceList = ref<GatewayInstance[]>([])

  /** 后端分页信息对象 */
  const pageInfo = ref<PageInfoObj | undefined>()

  // ============= 搜索表单配置 =============

  /** 搜索表单配置（符合 SearchFormProps 结构，移除工具栏按钮） */
  const searchFormConfig: Omit<SearchFormProps, 'moduleId'> = {
    fields: [
      {
        field: 'instanceName',
        label: '实例名称',
        type: 'input',
        placeholder: '请输入实例名称',
        span: 6,
        clearable: true,
      },
      {
        field: 'healthStatus',
        label: '健康状态',
        type: 'select',
        placeholder: '请选择状态',
        span: 6,
        clearable: true,
        options: [
          { label: '全部', value: '' },
          { label: '健康', value: 'Y' },
          { label: '不健康', value: 'N' },
        ],
      },
      {
        field: 'activeFlag',
        label: '活动状态',
        type: 'select',
        placeholder: '请选择状态',
        span: 6,
        clearable: true,
        options: [
          { label: '全部', value: '' },
          { label: '活动', value: 'Y' },
          { label: '非活动', value: 'N' },
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
        field: 'gatewayInstanceId',
        title: '实例ID',
        sortable: true,
        align: 'center',
        showOverflow: true,
      },
      {
        field: 'instanceName',
        title: '实例名称',
        sortable: true,
        align: 'center',
        showOverflow: true,
      },
      {
        field: 'instanceDesc',
        title: '实例描述',
        align: 'center',
        showOverflow: 'tooltip',
      },
      {
        field: 'bindAddress',
        title: '绑定地址',
        align: 'center',
        showOverflow: true,
      },
      {
        field: 'httpPort',
        title: 'HTTP端口',
        align: 'center',
      },
      {
        field: 'httpsPort',
        title: 'HTTPS端口',
        align: 'center',
      },
      {
        field: 'tlsEnabled',
        title: 'TLS',
        align: 'center',
        slots: { default: 'tlsEnabled' },
      },
      {
        field: 'maxConnections',
        title: '最大连接数',
        align: 'center',
        formatter: ({ cellValue }) => {
          return cellValue ? cellValue.toLocaleString() : '0'
        },
      },
      {
        field: 'healthStatus',
        title: '健康状态',
        align: 'center',
        slots: { default: 'healthStatus' },
      },
      {
        field: 'activeFlag',
        title: '活动状态',
        align: 'center',
        slots: { default: 'activeFlag' },
      },
      {
        field: 'addTime',
        title: '创建时间',
        sortable: true,
        showOverflow: true,
        formatter: ({ cellValue }) =>
          cellValue ? formatDate(cellValue, 'YYYY-MM-DD HH:mm:ss') : '',
      },
      {
        field: 'addWho',
        title: '创建人',
        showOverflow: true,
      },
      {
        field: 'editTime',
        title: '修改时间',
        sortable: true,
        showOverflow: true,
        formatter: ({ cellValue }) =>
          cellValue ? formatDate(cellValue, 'YYYY-MM-DD HH:mm:ss') : '',
      },
      {
        field: 'editWho',
        title: '修改人',
        showOverflow: true,
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
   * 设置实例列表
   */
  const setInstanceList = (list: GatewayInstance[]) => {
    instanceList.value = list
  }

  /**
   * 清空实例列表
   */
  const clearInstanceList = () => {
    instanceList.value = []
  }

  return {
    // 基本信息
    moduleId,

    // 数据状态
    loading,
    instanceList,
    pageInfo,

    // 配置
    searchFormConfig,
    gridConfig,

    // 方法
    resetPagination,
    updatePagination,
    setInstanceList,
    clearInstanceList,
  }
}

/**
 * Model 返回类型
 */
export type GatewayInstanceListModel = ReturnType<typeof useGatewayInstanceListModel>

