/**
 * 系统节点监控模块 Model
 * 统一管理搜索表单、表格配置和数据状态
 */

import type { SearchFormProps } from '@/components/form/search/types'
import type { GridProps } from '@/components/grid'
import type { PageInfoObj } from '@/types/api'
import { formatDate } from '@/utils/format'
import { ref } from 'vue'
import type { ServerInfo } from '../types'
import { OsType, ServerType } from '../types'

/**
 * 系统节点 Model
 */
export function useServerNodeModel() {
  // ============= 数据状态 =============
  const moduleId = 'hub0007'
  /** 加载状态 */
  const loading = ref(false)

  /** 系统节点列表数据 */
  const serverList = ref<ServerInfo[]>([])

  /** 后端分页信息对象 */
  const pageInfo = ref<PageInfoObj | undefined>()

  // ============= 搜索表单配置 =============

  /** 搜索表单配置（符合 SearchFormProps 结构） */
  const searchFormConfig: Omit<SearchFormProps, 'moduleId'> = {
    fields: [
      {
        field: 'hostname',
        label: '主机名',
        type: 'input',
        placeholder: '请输入主机名',
        span: 6,
        clearable: true
      },
      {
        field: 'ipAddress',
        label: 'IP地址',
        type: 'input',
        placeholder: '请输入IP地址',
        span: 6,
        clearable: true
      },
      {
        field: 'osType',
        label: '操作系统',
        type: 'select',
        placeholder: '请选择操作系统',
        span: 6,
        clearable: true,
        options: [
          { label: '全部', value: '' },
          { label: 'Linux', value: OsType.LINUX },
          { label: 'Windows', value: OsType.WINDOWS },
          { label: 'MacOS', value: OsType.MACOS },
          { label: 'Unix', value: OsType.UNIX },
          { label: '其他', value: OsType.OTHER }
        ]
      },
      {
        field: 'serverType',
        label: '服务器类型',
        type: 'select',
        placeholder: '请选择服务器类型',
        span: 6,
        clearable: true,
        options: [
          { label: '全部', value: '' },
          { label: '物理服务器', value: ServerType.PHYSICAL },
          { label: '虚拟服务器', value: ServerType.VIRTUAL },
          { label: '未知', value: ServerType.UNKNOWN }
        ]
      }
    ],
    moreFields: [
      {
        field: 'serverLocation',
        label: '服务器位置',
        type: 'input',
        placeholder: '请输入服务器位置',
        span: 6,
        clearable: true
      }
    ],
    toolbarButtons: []
  }

  // ============= 表格配置 =============

  /** 表格配置（符合 GridProps 结构） */
  const gridConfig: Omit<GridProps, 'moduleId' | 'data' | 'loading'> = {
    columns: [
      {
        field: 'metricServerId',
        title: '节点ID',
        showOverflow: true,
        width: 200
      },
      {
        field: 'hostname',
        title: '主机名',
        showOverflow: true
      },
      {
        field: 'ipAddress',
        title: 'IP地址',
        showOverflow: true
      },
      {
        field: 'osType',
        title: '操作系统',
        showOverflow: true,
        cellRender: {
          name: 'VxeTag',
          props: ({ row }: any) => ({
            type: 'info',
            content: row.osType
          })
        }
      },
      {
        field: 'osVersion',
        title: '系统版本',
        showOverflow: true
      },
      {
        field: 'architecture',
        title: '架构',
        showOverflow: true
      },
      {
        field: 'serverType',
        title: '服务器类型',
        cellRender: {
          name: 'VxeTag',
          props: ({ row }: any) => {
            const typeMap: Record<string, { type: string; text: string }> = {
              [ServerType.PHYSICAL]: { type: 'success', text: '物理服务器' },
              [ServerType.VIRTUAL]: { type: 'warning', text: '虚拟服务器' },
              [ServerType.UNKNOWN]: { type: 'default', text: '未知' }
            }
            const config = typeMap[row.serverType as ServerType] || typeMap[ServerType.UNKNOWN]
            return {
              type: config.type,
              content: config.text
            }
          }
        }
      },
      {
        field: 'serverLocation',
        title: '服务器位置',
        showOverflow: true
      },
      {
        field: 'lastUpdateTime',
        title: '最后更新',
        sortable: true,
        showOverflow: true,
        formatter: ({ cellValue }: any) =>
          cellValue ? formatDate(cellValue, 'YYYY-MM-DD HH:mm:ss') : ''
      },
      {
        field: 'addTime',
        title: '创建时间',
        sortable: true,
        showOverflow: true,
        formatter: ({ cellValue }: any) =>
          cellValue ? formatDate(cellValue, 'YYYY-MM-DD HH:mm:ss') : ''
      },
      {
        field: 'addWho',
        title: '创建人',
        showOverflow: true
      },
      {
        field: 'editTime',
        title: '修改时间',
        sortable: true,
        showOverflow: true,
        formatter: ({ cellValue }: any) =>
          cellValue ? formatDate(cellValue, 'YYYY-MM-DD HH:mm:ss') : ''
      },
      {
        field: 'editWho',
        title: '修改人',
        showOverflow: true
      }
    ],
    menuConfig: {
      enabled: true,
      customMenus: [
        {
          code: 'view',
          name: '查看详情',
          prefixIcon: 'vxe-icon-eye-fill'
        }
      ]
    },
    paginationConfig: {
      show: true
    }
  }

  // ============= 数据操作方法 =============

  /**
   * 设置节点列表数据
   */
  const setServerList = (list: ServerInfo[]) => {
    serverList.value = list
  }

  /**
   * 更新分页信息
   */
  const updatePagination = (newPageInfo: PageInfoObj) => {
    pageInfo.value = newPageInfo
  }

  // ============= 导出 =============

  return {
    moduleId,
    loading,
    serverList,
    pageInfo,
    searchFormConfig,
    gridConfig,
    setServerList,
    updatePagination
  }
}

/**
 * ServerNodeModel 类型定义
 */
export type ServerNodeModel = ReturnType<typeof useServerNodeModel>

