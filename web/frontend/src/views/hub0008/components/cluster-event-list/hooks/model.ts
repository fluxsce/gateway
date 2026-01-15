/**
 * 集群事件管理模块 Model
 * 统一管理搜索表单、表格配置和数据状态
 */

import type { SearchFormProps } from '@/components/form/search/types'
import type { GridProps } from '@/components/grid'
import type { PageInfoObj } from '@/types/api'
import { formatDate } from '@/utils/format'
import { ref } from 'vue'
import type { ClusterEvent } from '../../../types'

/**
 * 集群事件 Model
 */
export function useClusterEventModel() {
  // ============= 数据状态 =============
  const moduleId = 'hub0008:event-list'
  /** 加载状态 */
  const loading = ref(false)

  /** 集群事件列表数据 */
  const eventList = ref<ClusterEvent[]>([])

  /** 后端分页信息对象 */
  const pageInfo = ref<PageInfoObj | undefined>()

  // ============= 搜索表单配置 =============

  /** 搜索表单配置（符合 SearchFormProps 结构） */
  const searchFormConfig: Omit<SearchFormProps, 'moduleId'> = {
    fields: [
      {
        field: 'eventType',
        label: '事件类型',
        type: 'input',
        placeholder: '请输入事件类型',
        span: 8,
        clearable: true,
      },
      {
        field: 'eventAction',
        label: '事件动作',
        type: 'select',
        placeholder: '请选择事件动作',
        span: 8,
        clearable: true,
        options: [
          { label: '全部', value: '' },
          { label: 'CREATE', value: 'CREATE' },
          { label: 'UPDATE', value: 'UPDATE' },
          { label: 'DELETE', value: 'DELETE' },
          { label: 'REFRESH', value: 'REFRESH' },
          { label: 'INVALIDATE', value: 'INVALIDATE' },
        ],
      },
      {
        field: 'activeFlag',
        label: '活动状态',
        type: 'select',
        placeholder: '请选择活动状态',
        span: 8,
        clearable: true,
        options: [
          { label: '全部', value: '' },
          { label: '活动', value: 'Y' },
          { label: '非活动', value: 'N' },
        ],
      },
    ],
    moreFields: [
      {
        field: 'sourceNodeId',
        label: '发布节点ID',
        type: 'input',
        placeholder: '请输入发布节点ID',
        span: 8,
        clearable: true,
      },
      {
        field: 'sourceNodeIp',
        label: '发布节点IP',
        type: 'input',
        placeholder: '请输入发布节点IP',
        span: 8,
        clearable: true,
      },
    ],
    toolbarButtons: [
      {
        key: 'toggleAckList',
        label: '收起处理列表',
        type: 'primary',
        icon: 'ChevronForwardOutline',
        tooltip: '收起/展开处理列表',
        atEnd: true,
      },
    ],
    showSearchButton: true,
    showResetButton: true,
  }

  // ============= 表格配置 =============

  /** 表格配置（符合 GridProps 结构，排除响应式数据） */
  const gridConfig: Omit<GridProps, 'moduleId' | 'data' | 'loading'> = {
    columns: [
      {
        field: 'eventId',
        title: '事件ID',
        showOverflow: true,
        width: 200,
      },
      {
        field: 'eventType',
        title: '事件类型',
        width: 150,
        align: 'center',
        slots: { default: 'eventType' },
      },
      {
        field: 'eventAction',
        title: '事件动作',
        align: 'center',
        width: 120,
        slots: { default: 'eventAction' },
      },
      {
        field: 'sourceNodeId',
        title: '发布节点',
        showOverflow: true,
        width: 180,
      },
      {
        field: 'sourceNodeIp',
        title: '发布节点IP',
        showOverflow: true,
        width: 140,
      },
      {
        field: 'eventTime',
        title: '事件时间',
        sortable: true,
        showOverflow: true,
        width: 160,
        formatter: ({ cellValue }) =>
          cellValue ? formatDate(cellValue, 'YYYY-MM-DD HH:mm:ss') : '',
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
   * 设置事件列表
   */
  const setEventList = (list: ClusterEvent[]) => {
    eventList.value = list
  }

  /**
   * 清空事件列表
   */
  const clearEventList = () => {
    eventList.value = []
  }

  return {
    // 基本信息
    moduleId,

    // 数据状态
    loading,
    eventList,
    pageInfo,

    // 配置
    searchFormConfig,
    gridConfig,

    // 方法
    resetPagination,
    updatePagination,
    setEventList,
    clearEventList,
  }
}

/**
 * Model 返回类型
 */
export type ClusterEventModel = ReturnType<typeof useClusterEventModel>

