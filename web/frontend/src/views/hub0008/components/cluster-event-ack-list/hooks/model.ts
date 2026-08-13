/**
 * 集群事件确认管理模块 Model
 * 统一管理搜索表单、表格配置和数据状态（RsSearchForm / RsGrid）
 */

import type { RsSearchFormProps } from '@/components/form/rs-search'
import type {
  RsGridColumn,
  RsGridMenuConfig,
  RsGridPaginationConfig,
} from '@/components/rs-grid'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import type { PageInfoObj } from '@/types/api'
import { RsTag, type RsTagVariant } from '@/ui'
import { formatDate } from '@/utils/format'
import { h, reactive, ref, watch } from 'vue'
import type { ClusterEventAck } from '../../../types'

/**
 * 集群事件确认表格配置（对齐 RsGrid Props 子集）。
 */
export interface ClusterEventAckGridConfig {
  columns: RsGridColumn<ClusterEventAck>[]
  selectable: boolean
  rowKey: string
  height: string
  paginationConfig: RsGridPaginationConfig
  menuConfig: RsGridMenuConfig
}

/** 确认状态 → RsTag variant */
function getAckStatusVariant(status?: string): RsTagVariant {
  switch (status) {
    case 'SUCCESS':
      return 'success'
    case 'FAILED':
      return 'danger'
    case 'PENDING':
      return 'warning'
    default:
      return 'default'
  }
}

/**
 * 集群事件确认 Model
 */
export function useClusterEventAckModel() {
  const { t, locale } = useModuleI18n('hub0008')

  const moduleId = 'hub0008:event-ack'
  const loading = ref(false)
  const ackList = ref<ClusterEventAck[]>([])
  const pageInfo = ref<PageInfoObj | undefined>()

  const searchFormConfig = reactive<Omit<RsSearchFormProps, 'moduleId'>>({
    fields: [],
    moreFields: [],
    toolbarButtons: [],
    showSearchButton: true,
    showResetButton: true,
  })

  const gridConfig = reactive<ClusterEventAckGridConfig>({
    columns: [],
    selectable: false,
    rowKey: 'ackId',
    height: '100%',
    paginationConfig: {
      show: true,
      pageInfo: pageInfo as any,
      align: 'right',
    },
    menuConfig: {
      enabled: true,
      items: [],
    },
  })

  /** 确认状态文案 */
  function getAckStatusLabel(status?: string): string {
    switch (status) {
      case 'PENDING':
        return t('ack.status.pending')
      case 'SUCCESS':
        return t('ack.status.success')
      case 'FAILED':
        return t('ack.status.failed')
      case 'SKIPPED':
        return t('ack.status.skipped')
      default:
        return status || '-'
    }
  }

  /** 按当前语言刷新表单 / 表格文案 */
  function applyI18n() {
    searchFormConfig.fields = [
      {
        field: 'nodeId',
        label: t('ack.search.nodeId'),
        type: 'input',
        placeholder: t('ack.search.nodeIdPlaceholder'),
        span: 12,
        clearable: true,
      },
      {
        field: 'nodeIp',
        label: t('ack.search.nodeIp'),
        type: 'input',
        placeholder: t('ack.search.nodeIpPlaceholder'),
        span: 12,
        clearable: true,
      },
    ]
    searchFormConfig.moreFields = [
      {
        field: 'ackStatus',
        label: t('ack.search.ackStatus'),
        type: 'select',
        placeholder: t('ack.search.ackStatusPlaceholder'),
        span: 12,
        clearable: true,
        options: [
          { label: t('common.all'), value: '' },
          { label: t('ack.status.pending'), value: 'PENDING' },
          { label: t('ack.status.success'), value: 'SUCCESS' },
          { label: t('ack.status.failed'), value: 'FAILED' },
          { label: t('ack.status.skipped'), value: 'SKIPPED' },
        ],
      },
      {
        field: 'activeFlag',
        label: t('ack.search.activeFlag'),
        type: 'select',
        placeholder: t('ack.search.activeFlagPlaceholder'),
        span: 12,
        clearable: true,
        options: [
          { label: t('common.all'), value: '' },
          { label: t('common.active'), value: 'Y' },
          { label: t('common.inactive'), value: 'N' },
        ],
      },
    ]

    gridConfig.columns = [
      {
        key: 'ackId',
        title: t('ack.columns.ackId'),
        width: 200,
        ellipsis: true,
      },
      {
        key: 'nodeId',
        title: t('ack.columns.nodeId'),
        width: 180,
        ellipsis: true,
        render: (row) =>
          h(
            'span',
            { style: { color: 'var(--g-primary)', fontWeight: 500 } },
            row.nodeId || '-',
          ),
      },
      {
        key: 'nodeIp',
        title: t('ack.columns.nodeIp'),
        width: 140,
        ellipsis: true,
        render: (row) =>
          h(
            'span',
            { style: { color: 'var(--g-success)', fontWeight: 500 } },
            row.nodeIp || '-',
          ),
      },
      {
        key: 'ackStatus',
        title: t('ack.columns.ackStatus'),
        align: 'center',
        width: 120,
        render: (row) =>
          h(
            RsTag,
            { variant: getAckStatusVariant(row.ackStatus), size: 'sm' },
            () => getAckStatusLabel(row.ackStatus),
          ),
      },
      {
        key: 'processTime',
        title: t('ack.columns.processTime'),
        sortable: true,
        width: 160,
        ellipsis: true,
        formatter: (value) =>
          value ? formatDate(value as string, 'YYYY-MM-DD HH:mm:ss') : '-',
      },
      {
        key: 'retryCount',
        title: t('ack.columns.retryCount'),
        align: 'center',
        width: 100,
      },
      {
        key: 'resultMessage',
        title: t('ack.columns.resultMessage'),
        minWidth: 200,
        ellipsis: true,
      },
      {
        key: 'activeFlag',
        title: t('ack.columns.activeFlag'),
        align: 'center',
        width: 100,
        render: (row) =>
          h(
            RsTag,
            {
              variant: row.activeFlag === 'Y' ? 'success' : 'danger',
              size: 'sm',
            },
            () => (row.activeFlag === 'Y' ? t('common.active') : t('common.inactive')),
          ),
      },
    ]
    gridConfig.menuConfig = {
      enabled: true,
      items: [{ key: 'view', label: t('common.viewDetail'), icon: 'EyeOutline' }],
    }
  }

  watch(locale, applyI18n, { immediate: true })

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

  const setAckList = (list: ClusterEventAck[]) => {
    ackList.value = list
  }

  const clearAckList = () => {
    ackList.value = []
  }

  return {
    moduleId,
    loading,
    ackList,
    pageInfo,
    searchFormConfig,
    gridConfig,
    resetPagination,
    updatePagination,
    setAckList,
    clearAckList,
  }
}

export type ClusterEventAckModel = ReturnType<typeof useClusterEventAckModel>
