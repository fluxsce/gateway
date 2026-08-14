/**
 * 网关实例列表查询 Model（仅查询功能）
 * 复用 hub0020 的字段，但移除工具栏按钮和右键菜单
 */

import type { RsSearchFormProps } from '@/components/form/rs-search'
import type { RsGridColumn, RsGridMenuConfig, RsGridPaginationConfig } from '@/components/rs-grid'
import type { PageInfoObj } from '@/types/api'
import { RsTag } from '@/ui'
import { formatDate } from '@/utils/format'
import type { GatewayInstance } from '@/views/hub0020/types'
import { h, ref } from 'vue'

/**
 * 网关实例列表表格配置（对齐 RsGrid Props 子集）。
 */
export interface GatewayInstanceListGridConfig {
  columns: RsGridColumn<GatewayInstance>[]
  selectable: boolean
  rowKey: string
  height: string
  paginationConfig: RsGridPaginationConfig
  menuConfig: RsGridMenuConfig
}

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

  /** 搜索表单配置（符合 RsSearchFormProps 结构，移除工具栏按钮） */
  const searchFormConfig: Omit<RsSearchFormProps, 'moduleId'> = {
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
    toolbarButtons: [],
    showSearchButton: true,
    showResetButton: true,
  }

  // ============= 表格配置 =============

  /** 表格配置（符合 RsGrid 结构，排除响应式数据，移除右键菜单） */
  const gridConfig: GatewayInstanceListGridConfig = {
    columns: [
      {
        key: 'gatewayInstanceId',
        title: '实例ID',
        sortable: true,
        align: 'center',
        ellipsis: true,
      },
      {
        key: 'instanceName',
        title: '实例名称',
        sortable: true,
        align: 'center',
        ellipsis: true,
      },
      {
        key: 'instanceDesc',
        title: '实例描述',
        align: 'center',
        ellipsis: true,
      },
      {
        key: 'bindAddress',
        title: '绑定地址',
        align: 'center',
        ellipsis: true,
      },
      {
        key: 'httpPort',
        title: 'HTTP端口',
        align: 'center',
      },
      {
        key: 'httpsPort',
        title: 'HTTPS端口',
        align: 'center',
      },
      {
        key: 'tlsEnabled',
        title: 'TLS',
        align: 'center',
        render: (row) =>
          h(
            RsTag,
            {
              variant: row.tlsEnabled === 'Y' ? 'success' : 'default',
              size: 'sm',
            },
            () => (row.tlsEnabled === 'Y' ? '启用' : '禁用'),
          ),
      },
      {
        key: 'maxConnections',
        title: '最大连接数',
        align: 'center',
        formatter: (value) => (value ? Number(value).toLocaleString() : '0'),
      },
      {
        key: 'healthStatus',
        title: '健康状态',
        align: 'center',
        render: (row) =>
          h(
            RsTag,
            {
              variant: row.healthStatus === 'Y' ? 'success' : 'danger',
              size: 'sm',
            },
            () => (row.healthStatus === 'Y' ? '在线' : '离线'),
          ),
      },
      {
        key: 'activeFlag',
        title: '活动状态',
        align: 'center',
        render: (row) =>
          h(
            RsTag,
            {
              variant: row.activeFlag === 'Y' ? 'success' : 'default',
              size: 'sm',
            },
            () => (row.activeFlag === 'Y' ? '活动' : '非活动'),
          ),
      },
      {
        key: 'addTime',
        title: '创建时间',
        sortable: true,
        ellipsis: true,
        formatter: (value) =>
          value ? formatDate(value as string, 'YYYY-MM-DD HH:mm:ss') : '',
      },
      {
        key: 'addWho',
        title: '创建人',
        ellipsis: true,
      },
      {
        key: 'editTime',
        title: '修改时间',
        sortable: true,
        ellipsis: true,
        formatter: (value) =>
          value ? formatDate(value as string, 'YYYY-MM-DD HH:mm:ss') : '',
      },
      {
        key: 'editWho',
        title: '修改人',
        ellipsis: true,
      },
    ],
    selectable: false,
    rowKey: 'gatewayInstanceId',
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
