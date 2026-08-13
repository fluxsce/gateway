/**
 * 系统节点监控模块 Model
 * 统一管理搜索表单、表格配置和数据状态
 */

import type { RsSearchFormProps } from '@/components/form/rs-search'
import type { GridProps } from '@/components/grid'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import type { PageInfoObj } from '@/types/api'
import { formatDate } from '@/utils/format'
import { reactive, ref, watch } from 'vue'
import type { ServerInfo } from '../types'
import { OsType, ServerType } from '../types'

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

  const gridConfig = reactive<Omit<GridProps, 'moduleId' | 'data' | 'loading'>>({
    columns: [],
    menuConfig: {
      enabled: true,
      options: [],
    },
    paginationConfig: {
      show: true,
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
        field: 'metricServerId',
        title: t('columns.metricServerId'),
        showOverflow: true,
        width: 200,
      },
      {
        field: 'hostname',
        title: t('columns.hostname'),
        showOverflow: true,
      },
      {
        field: 'ipAddress',
        title: t('columns.ipAddress'),
        showOverflow: true,
      },
      {
        field: 'osType',
        title: t('columns.osType'),
        showOverflow: true,
        cellRender: {
          name: 'VxeTag',
          props: ({ row }: any) => ({
            type: 'info',
            content: row.osType,
          }),
        },
      },
      {
        field: 'osVersion',
        title: t('columns.osVersion'),
        showOverflow: true,
      },
      {
        field: 'architecture',
        title: t('columns.architecture'),
        showOverflow: true,
      },
      {
        field: 'serverType',
        title: t('columns.serverType'),
        cellRender: {
          name: 'VxeTag',
          props: ({ row }: any) => {
            const typeMap: Record<string, { type: string; text: string }> = {
              [ServerType.PHYSICAL]: { type: 'success', text: t('serverType.physical') },
              [ServerType.VIRTUAL]: { type: 'warning', text: t('serverType.virtual') },
              [ServerType.UNKNOWN]: { type: 'default', text: t('serverType.unknown') },
            }
            const config = typeMap[row.serverType as ServerType] || typeMap[ServerType.UNKNOWN]
            return {
              type: config.type,
              content: config.text,
            }
          },
        },
      },
      {
        field: 'serverLocation',
        title: t('columns.serverLocation'),
        showOverflow: true,
      },
      {
        field: 'lastUpdateTime',
        title: t('columns.lastUpdateTime'),
        sortable: true,
        showOverflow: true,
        formatter: ({ cellValue }: any) =>
          cellValue ? formatDate(cellValue, 'YYYY-MM-DD HH:mm:ss') : '',
      },
      {
        field: 'addTime',
        title: t('columns.addTime'),
        sortable: true,
        showOverflow: true,
        formatter: ({ cellValue }: any) =>
          cellValue ? formatDate(cellValue, 'YYYY-MM-DD HH:mm:ss') : '',
      },
      {
        field: 'addWho',
        title: t('columns.addWho'),
        showOverflow: true,
      },
      {
        field: 'editTime',
        title: t('columns.editTime'),
        sortable: true,
        showOverflow: true,
        formatter: ({ cellValue }: any) =>
          cellValue ? formatDate(cellValue, 'YYYY-MM-DD HH:mm:ss') : '',
      },
      {
        field: 'editWho',
        title: t('columns.editWho'),
        showOverflow: true,
      },
    ]

    gridConfig.menuConfig = {
      enabled: true,
      options: [
        {
          code: 'view',
          name: t('contextMenu.view'),
          prefixIcon: 'vxe-icon-eye-fill',
        },
      ],
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
