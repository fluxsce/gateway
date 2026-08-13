/**
 * 集群事件管理模块 Model
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
import type { ClusterEvent } from '../../../types'

/**
 * 集群事件表格配置（对齐 RsGrid Props 子集）。
 */
export interface ClusterEventGridConfig {
  columns: RsGridColumn<ClusterEvent>[]
  selectable: boolean
  rowKey: string
  height: string
  paginationConfig: RsGridPaginationConfig
  menuConfig: RsGridMenuConfig
}

/** 事件动作 → RsTag variant */
function getEventActionVariant(action?: string): RsTagVariant {
  switch (action) {
    case 'START':
    case 'CREATE':
      return 'success'
    case 'STOP':
    case 'DELETE':
      return 'danger'
    case 'RELOAD':
    case 'REFRESH':
    case 'INVALIDATE':
      return 'warning'
    case 'RESTART':
    case 'UPDATE':
      return 'info'
    default:
      return 'default'
  }
}

/**
 * 集群事件 Model
 */
export function useClusterEventModel() {
  const { t, locale } = useModuleI18n('hub0008')

  const moduleId = 'hub0008:event-list'
  const loading = ref(false)
  const eventList = ref<ClusterEvent[]>([])
  const pageInfo = ref<PageInfoObj | undefined>()

  const searchFormConfig = reactive<Omit<RsSearchFormProps, 'moduleId'>>({
    fields: [],
    moreFields: [],
    toolbarButtons: [],
    showSearchButton: true,
    showResetButton: true,
  })

  const gridConfig = reactive<ClusterEventGridConfig>({
    columns: [],
    selectable: false,
    rowKey: 'eventId',
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

  /** 按当前语言刷新表单 / 表格文案 */
  function applyI18n() {
    searchFormConfig.fields = [
      {
        field: 'eventType',
        label: t('event.search.eventType'),
        type: 'input',
        placeholder: t('event.search.eventTypePlaceholder'),
        span: 12,
        clearable: true,
      },
      {
        field: 'eventAction',
        label: t('event.search.eventAction'),
        type: 'select',
        placeholder: t('event.search.eventActionPlaceholder'),
        span: 12,
        clearable: true,
        options: [
          { label: t('common.all'), value: '' },
          { label: 'CREATE', value: 'CREATE' },
          { label: 'UPDATE', value: 'UPDATE' },
          { label: 'DELETE', value: 'DELETE' },
          { label: 'REFRESH', value: 'REFRESH' },
          { label: 'INVALIDATE', value: 'INVALIDATE' },
        ],
      },
    ]
    searchFormConfig.moreFields = [
      {
        field: 'activeFlag',
        label: t('event.search.activeFlag'),
        type: 'select',
        placeholder: t('event.search.activeFlagPlaceholder'),
        span: 12,
        clearable: true,
        options: [
          { label: t('common.all'), value: '' },
          { label: t('common.active'), value: 'Y' },
          { label: t('common.inactive'), value: 'N' },
        ],
      },
      {
        field: 'sourceNodeId',
        label: t('event.search.sourceNodeId'),
        type: 'input',
        placeholder: t('event.search.sourceNodeIdPlaceholder'),
        span: 12,
        clearable: true,
      },
      {
        field: 'sourceNodeIp',
        label: t('event.search.sourceNodeIp'),
        type: 'input',
        placeholder: t('event.search.sourceNodeIpPlaceholder'),
        span: 12,
        clearable: true,
      },
    ]
    searchFormConfig.toolbarButtons = [
      {
        key: 'toggleAckList',
        label: t('event.toolbar.collapseAckList'),
        type: 'primary',
        icon: 'ChevronForwardOutline',
        tooltip: t('event.toolbar.toggleAckListTooltip'),
        atEnd: true,
      },
    ]

    gridConfig.columns = [
      {
        key: 'eventId',
        title: t('event.columns.eventId'),
        width: 200,
        ellipsis: true,
      },
      {
        key: 'eventType',
        title: t('event.columns.eventType'),
        width: 150,
        align: 'center',
        render: (row) =>
          h(RsTag, { variant: 'primary', size: 'sm' }, () => row.eventType || '-'),
      },
      {
        key: 'eventAction',
        title: t('event.columns.eventAction'),
        align: 'center',
        width: 120,
        render: (row) =>
          h(
            RsTag,
            { variant: getEventActionVariant(row.eventAction), size: 'sm' },
            () => row.eventAction || '-',
          ),
      },
      {
        key: 'sourceNodeId',
        title: t('event.columns.sourceNodeId'),
        width: 180,
        ellipsis: true,
      },
      {
        key: 'sourceNodeIp',
        title: t('event.columns.sourceNodeIp'),
        width: 140,
        ellipsis: true,
      },
      {
        key: 'eventTime',
        title: t('event.columns.eventTime'),
        sortable: true,
        width: 160,
        ellipsis: true,
        formatter: (value) =>
          value ? formatDate(value as string, 'YYYY-MM-DD HH:mm:ss') : '',
      },
      {
        key: 'activeFlag',
        title: t('event.columns.activeFlag'),
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

  const setEventList = (list: ClusterEvent[]) => {
    eventList.value = list
  }

  const clearEventList = () => {
    eventList.value = []
  }

  return {
    moduleId,
    loading,
    eventList,
    pageInfo,
    searchFormConfig,
    gridConfig,
    resetPagination,
    updatePagination,
    setEventList,
    clearEventList,
  }
}

export type ClusterEventModel = ReturnType<typeof useClusterEventModel>
