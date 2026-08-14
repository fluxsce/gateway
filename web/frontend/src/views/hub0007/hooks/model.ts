/**
 * 系统节点监控模块 Model
 * 统一管理搜索表单、表格配置和数据状态
 */

import type { RsSearchFormProps } from '@/components/form/rs-search'
import type { RsGridColumn, RsGridMenuConfig, RsGridPaginationConfig } from '@/components/rs-grid'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import type { PageInfoObj } from '@/types/api'
import { RsTag, type RsTagVariant } from '@/ui'
import { formatDate } from '@/utils/format'
import { h, reactive, ref, watch } from 'vue'
import type { ServerInfo } from '../types'
import { OsType, ServerType } from '../types'

/**
 * 系统节点表格配置（对齐 RsGrid Props 子集）。
 */
export interface ServerNodeGridConfig {
  columns: RsGridColumn<ServerInfo>[]
  rowKey: string
  height: string
  paginationConfig: RsGridPaginationConfig
  menuConfig: RsGridMenuConfig
}

function serverTypeTag(serverType: string, t: (key: string) => string): { variant: RsTagVariant; text: string } {
  if (serverType === ServerType.PHYSICAL) {
    return { variant: 'success', text: t('serverType.physical') }
  }
  if (serverType === ServerType.VIRTUAL) {
    return { variant: 'warning', text: t('serverType.virtual') }
  }
  return { variant: 'default', text: t('serverType.unknown') }
}

/**
 * 系统节点 Model
 */
export function useServerNodeModel() {
  const { t, locale } = useModuleI18n('hub0007')

  const moduleId = 'hub0007'
  const loading = ref(false)
  const serverList = ref<ServerInfo[]>([])
  const pageInfo = ref<PageInfoObj | undefined>()

  const searchFormConfig = reactive<Omit<RsSearchFormProps, 'moduleId'>>({
    fields: [],
    moreFields: [],
    toolbarButtons: [],
  })

  const gridConfig = reactive<ServerNodeGridConfig>({
    columns: [],
    rowKey: 'metricServerId',
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

  /** 按当前语言刷新搜索表单 / 表格文案 */
  function applyI18n() {
    searchFormConfig.fields = [
      {
        field: 'hostname',
        label: t('search.hostname'),
        type: 'input',
        placeholder: t('search.hostnamePlaceholder'),
        span: 6,
        clearable: true,
      },
      {
        field: 'ipAddress',
        label: t('search.ipAddress'),
        type: 'input',
        placeholder: t('search.ipAddressPlaceholder'),
        span: 6,
        clearable: true,
      },
      {
        field: 'osType',
        label: t('search.osType'),
        type: 'select',
        placeholder: t('search.osTypePlaceholder'),
        span: 6,
        clearable: true,
        options: [
          { label: t('search.all'), value: '' },
          { label: t('osType.linux'), value: OsType.LINUX },
          { label: t('osType.windows'), value: OsType.WINDOWS },
          { label: t('osType.macos'), value: OsType.MACOS },
          { label: t('osType.unix'), value: OsType.UNIX },
          { label: t('osType.other'), value: OsType.OTHER },
        ],
      },
      {
        field: 'serverType',
        label: t('search.serverType'),
        type: 'select',
        placeholder: t('search.serverTypePlaceholder'),
        span: 6,
        clearable: true,
        options: [
          { label: t('search.all'), value: '' },
          { label: t('serverType.physical'), value: ServerType.PHYSICAL },
          { label: t('serverType.virtual'), value: ServerType.VIRTUAL },
          { label: t('serverType.unknown'), value: ServerType.UNKNOWN },
        ],
      },
    ]

    searchFormConfig.moreFields = [
      {
        field: 'serverLocation',
        label: t('search.serverLocation'),
        type: 'input',
        placeholder: t('search.serverLocationPlaceholder'),
        span: 6,
        clearable: true,
      },
    ]

    gridConfig.columns = [
      {
        key: 'metricServerId',
        title: t('columns.metricServerId'),
        ellipsis: true,
        width: 200,
      },
      {
        key: 'hostname',
        title: t('columns.hostname'),
        ellipsis: true,
      },
      {
        key: 'ipAddress',
        title: t('columns.ipAddress'),
        ellipsis: true,
      },
      {
        key: 'osType',
        title: t('columns.osType'),
        render: (row) =>
          h(RsTag, { variant: 'info', size: 'sm' }, () => row.osType || '-'),
      },
      {
        key: 'osVersion',
        title: t('columns.osVersion'),
        ellipsis: true,
      },
      {
        key: 'architecture',
        title: t('columns.architecture'),
        ellipsis: true,
      },
      {
        key: 'serverType',
        title: t('columns.serverType'),
        render: (row) => {
          const tag = serverTypeTag(row.serverType, t)
          return h(RsTag, { variant: tag.variant, size: 'sm' }, () => tag.text)
        },
      },
      {
        key: 'serverLocation',
        title: t('columns.serverLocation'),
        ellipsis: true,
      },
      {
        key: 'lastUpdateTime',
        title: t('columns.lastUpdateTime'),
        sortable: true,
        ellipsis: true,
        formatter: (value) => (value ? formatDate(value as string, 'YYYY-MM-DD HH:mm:ss') : ''),
      },
      {
        key: 'addTime',
        title: t('columns.addTime'),
        sortable: true,
        ellipsis: true,
        formatter: (value) => (value ? formatDate(value as string, 'YYYY-MM-DD HH:mm:ss') : ''),
      },
      {
        key: 'addWho',
        title: t('columns.addWho'),
        ellipsis: true,
      },
      {
        key: 'editTime',
        title: t('columns.editTime'),
        sortable: true,
        ellipsis: true,
        formatter: (value) => (value ? formatDate(value as string, 'YYYY-MM-DD HH:mm:ss') : ''),
      },
      {
        key: 'editWho',
        title: t('columns.editWho'),
        ellipsis: true,
      },
    ]

    gridConfig.menuConfig = {
      enabled: true,
      items: [{ key: 'view', label: t('contextMenu.view'), icon: 'eye' }],
    }
  }

  applyI18n()
  watch(locale, () => applyI18n())

  const setServerList = (list: ServerInfo[]) => {
    serverList.value = list
  }

  const updatePagination = (newPageInfo: PageInfoObj) => {
    pageInfo.value = newPageInfo
  }

  return {
    moduleId,
    loading,
    serverList,
    pageInfo,
    searchFormConfig,
    gridConfig,
    setServerList,
    updatePagination,
  }
}

/**
 * ServerNodeModel 类型定义
 */
export type ServerNodeModel = ReturnType<typeof useServerNodeModel>
