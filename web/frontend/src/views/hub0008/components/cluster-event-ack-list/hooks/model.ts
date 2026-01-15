/**
 * 集群事件确认管理模块 Model
 * 统一管理搜索表单、表格配置和数据状态
 */

import type { SearchFormProps } from '@/components/form/search/types'
import type { GridProps } from '@/components/grid'
import type { PageInfoObj } from '@/types/api'
import { formatDate } from '@/utils/format'
import { ref } from 'vue'
import type { ClusterEventAck } from '../../../types'

/**
 * 集群事件确认 Model
 */
export function useClusterEventAckModel() {
  // ============= 数据状态 =============
  const moduleId = 'hub0008:event-ack'
  /** 加载状态 */
  const loading = ref(false)

  /** 集群事件确认列表数据 */
  const ackList = ref<ClusterEventAck[]>([])

  /** 后端分页信息对象 */
  const pageInfo = ref<PageInfoObj | undefined>()

  // ============= 搜索表单配置 =============

  /** 搜索表单配置（符合 SearchFormProps 结构） */
  const searchFormConfig: Omit<SearchFormProps, 'moduleId'> = {
    fields: [
      {
        field: 'nodeId',
        label: '处理节点ID',
        type: 'input',
        placeholder: '请输入处理节点ID',
        span: 6,
        clearable: true,
      },
      {
        field: 'nodeIp',
        label: '处理节点IP',
        type: 'input',
        placeholder: '请输入处理节点IP',
        span: 6,
        clearable: true,
      },
      {
        field: 'ackStatus',
        label: '确认状态',
        type: 'select',
        placeholder: '请选择确认状态',
        span: 6,
        clearable: true,
        options: [
          { label: '全部', value: '' },
          { label: '待处理', value: 'PENDING' },
          { label: '成功', value: 'SUCCESS' },
          { label: '失败', value: 'FAILED' },
          { label: '跳过', value: 'SKIPPED' },
        ],
      },
      {
        field: 'activeFlag',
        label: '活动状态',
        type: 'select',
        placeholder: '请选择活动状态',
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

  /** 表格配置（符合 GridProps 结构，排除响应式数据） */
  const gridConfig: Omit<GridProps, 'moduleId' | 'data' | 'loading'> = {
    columns: [
      {
        field: 'ackId',
        title: '确认ID',
        showOverflow: true,
        width: 200,
      },
      {
        field: 'nodeId',
        title: '处理节点ID',
        showOverflow: true,
        width: 180,
        slots: { default: 'nodeId' },
      },
      {
        field: 'nodeIp',
        title: '处理节点IP',
        showOverflow: true,
        width: 140,
        slots: { default: 'nodeIp' },
      },
      {
        field: 'ackStatus',
        title: '确认状态',
        align: 'center',
        width: 120,
        slots: { default: 'ackStatus' },
      },
      {
        field: 'processTime',
        title: '处理时间',
        sortable: true,
        showOverflow: true,
        width: 160,
        formatter: ({ cellValue }) =>
          cellValue ? formatDate(cellValue, 'YYYY-MM-DD HH:mm:ss') : '-',
      },
      {
        field: 'retryCount',
        title: '重试次数',
        align: 'center',
        width: 100,
      },
      {
        field: 'resultMessage',
        title: '结果信息',
        showOverflow: 'tooltip',
        minWidth: 200,
      },
      {
        field: 'activeFlag',
        title: '活动状态',
        align: 'center',
        width: 100,
        cellRender: {
          name: 'VxeTag',
          props: ({ row }: any) => ({
            type: row.activeFlag === 'Y' ? 'success' : 'error',
            content: row.activeFlag === 'Y' ? '活动' : '非活动',
          }),
        },
      },
    ],
    showCheckbox: false,
    paginationConfig: {
      show: true,
      pageInfo: pageInfo as any,
      align: 'right',
    },
    menuConfig: {
      enabled: true,
      showCopyRow: true,
      showCopyCell: true,
      customMenus: [
        {
          code: 'view',
          name: '查看详情',
          prefixIcon: 'vxe-icon-eye-fill',
        },
      ],
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
   * 设置确认列表
   */
  const setAckList = (list: ClusterEventAck[]) => {
    ackList.value = list
  }

  /**
   * 清空确认列表
   */
  const clearAckList = () => {
    ackList.value = []
  }

  return {
    // 基本信息
    moduleId,

    // 数据状态
    loading,
    ackList,
    pageInfo,

    // 配置
    searchFormConfig,
    gridConfig,

    // 方法
    resetPagination,
    updatePagination,
    setAckList,
    clearAckList,
  }
}

/**
 * Model 返回类型
 */
export type ClusterEventAckModel = ReturnType<typeof useClusterEventAckModel>

