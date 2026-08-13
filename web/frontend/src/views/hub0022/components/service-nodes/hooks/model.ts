/**
 * 服务节点管理列表 Model
 * 统一管理搜索表单、表格配置和数据状态
 */

import type { DataFormField, DataFormTab } from '@/components/form/data/types'
import type { RsSearchFormProps } from '@/components/form/rs-search'
import type { RsGridColumn, RsGridMenuConfig, RsGridPaginationConfig } from '@/components/rs-grid'
import type { PageInfoObj } from '@/types/api'
import { RsTag } from '@/ui'
import { formatDate } from '@/utils/format'
import { h, ref } from 'vue'
import type { ServiceNode } from '../types'

/**
 * 服务节点表格配置（对齐 RsGrid Props 子集）。
 */
export interface ServiceNodeGridConfig {
  columns: RsGridColumn<ServiceNode>[]
  selectable: boolean
  rowKey: string
  height: string
  paginationConfig: RsGridPaginationConfig
  menuConfig: RsGridMenuConfig
}

/**
 * 在已有完整 URL 的基础上，按新的协议/主机/端口重建节点地址，
 * 同时保留原 URL 中的路径、查询参数与 hash。
 * 避免用户编辑协议、主机或端口时，把节点 URL 上携带的查询参数（如 ?method=putSKU&sign=...）丢失。
 */
function buildNodeUrl(
  protocol: string,
  host: string,
  port: number | string,
  existingUrl?: string
): string {
  const scheme = String(protocol || 'http').toLowerCase()
  let suffix = ''
  if (existingUrl) {
    try {
      const parsed = new URL(existingUrl)
      suffix = `${parsed.pathname}${parsed.search}${parsed.hash}`
      // 仅有根路径且无查询/hash 时不附加，避免产生多余的结尾斜杠
      if (suffix === '/') {
        suffix = ''
      }
    } catch {
      // existingUrl 不是合法的完整 URL（例如尚未填写），忽略其路径与参数
    }
  }
  return `${scheme}://${host}:${port}${suffix}`
}

/**
 * 服务节点管理列表 Model
 */
export function useServiceNodeModel() {
  // ============= 数据状态 =============
  const moduleId = 'hub0022:manageNodes'
  
  /** 加载状态 */
  const loading = ref(false)

  /** 服务节点列表数据 */
  const nodeList = ref<ServiceNode[]>([])

  /** 后端分页信息对象 */
  const pageInfo = ref<PageInfoObj | undefined>()

  // ============= 搜索表单配置 =============

  /** 搜索表单配置（符合 RsSearchFormProps 结构） */
  const searchFormConfig: Omit<RsSearchFormProps, 'moduleId'> = {
    fields: [
      {
        field: 'nodeHost',
        label: '节点主机',
        type: 'input',
        placeholder: '请输入节点主机地址',
        span: 6,
        clearable: true,
      },
      {
        field: 'healthStatus',
        label: '健康状态',
        type: 'select',
        placeholder: '请选择健康状态',
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
        label: '在线状态',
        type: 'select',
        placeholder: '请选择在线状态',
        span: 6,
        clearable: true,
        options: [
          { label: '全部', value: '' },
          { label: '在线', value: 'Y' },
          { label: '离线', value: 'N' },
        ],
      },
    ],
    toolbarButtons: [
      {
        key: 'add',
        label: '新建节点',
        icon: 'AddOutline',
        type: 'primary',
        tooltip: '新建服务节点',
      },
      {
        key: 'edit',
        label: '编辑',
        icon: 'CreateOutline',
        type: 'default',
        tooltip: '编辑选中的服务节点',
      },
      {
        key: 'delete',
        label: '删除',
        icon: 'TrashOutline',
        type: 'error',
        tooltip: '批量删除选中的服务节点',
      },
    ],
    showSearchButton: true,
    showResetButton: true,
  }

  // ============= 表单配置 =============

  /** 服务节点表单配置 */
  const nodeFormConfig = {
    tabs: [
      { key: 'basic', label: '基本信息' },
      { key: 'metadata', label: '元数据配置' },
      { key: 'other', label: '其他配置' },
    ] as DataFormTab[],
    fields: [
      // ============= 主键字段（隐藏，但必须存在用于编辑） =============
      {
        field: 'serviceNodeId',
        label: '服务节点ID',
        type: 'input' as const,
        span: 12,
        tabKey: 'basic',
        primary: true,
        show: false,
      },
      {
        field: 'nodeId',
        label: '节点ID',
        type: 'input' as const,
        span: 12,
        tabKey: 'basic',
        show: false,
      },
      {
        field: 'serviceDefinitionId',
        label: '服务定义ID',
        type: 'input' as const,
        span: 12,
        tabKey: 'basic',
        show: false,
      },
      {
        field: 'tenantId',
        label: '租户ID',
        type: 'input' as const,
        span: 12,
        tabKey: 'basic',
        show: false,
      },
      // ============= 基本信息 Tab =============
      {
        field: 'nodeProtocol',
        label: '节点协议',
        type: 'select' as const,
        placeholder: '请选择协议',
        span: 12,
        tabKey: 'basic',
        required: true,
        defaultValue: 'HTTP',
        props: {
          onUpdateValue: (value: string, formData?: Record<string, any>) => {
            // 当协议变化时，自动更新 nodeUrl（保留已有的路径与查询参数）
            if (formData && formData.nodeHost && formData.nodePort !== undefined && value) {
              formData.nodeUrl = buildNodeUrl(value, formData.nodeHost, formData.nodePort, formData.nodeUrl)
            }
          },
        },
        options: [
          { label: 'HTTP', value: 'HTTP' },
          { label: 'HTTPS', value: 'HTTPS' },
        ],
        rules: [
          { required: true, message: '请选择节点协议', trigger: ['blur', 'change'] },
        ],
      },
      {
        field: 'nodeHost',
        label: '节点主机',
        type: 'input' as const,
        placeholder: '请输入主机地址',
        span: 12,
        tabKey: 'basic',
        required: true,
        props: {
          onUpdateValue: (value: string, formData?: Record<string, any>) => {
            // 当主机地址变化时，自动更新 nodeUrl（保留已有的路径与查询参数）
            if (formData && value && formData.nodePort !== undefined && formData.nodeProtocol) {
              formData.nodeUrl = buildNodeUrl(formData.nodeProtocol, value, formData.nodePort, formData.nodeUrl)
            }
          },
        },
        rules: [
          { required: true, message: '请输入节点主机地址', trigger: ['blur', 'input'] },
          {
            pattern: /^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$|^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$/,
            message: '请输入有效的IP地址或域名',
            trigger: ['blur', 'input'],
          },
        ],
      },
      {
        field: 'nodePort',
        label: '节点端口',
        type: 'number' as const,
        placeholder: '请输入端口',
        span: 12,
        tabKey: 'basic',
        required: true,
        props: {
          min: 1,
          max: 65535,
          onUpdateValue: (value: number, formData?: Record<string, any>) => {
            // 当端口变化时，自动更新 nodeUrl（保留已有的路径与查询参数）
            if (formData && formData.nodeHost && value !== undefined && formData.nodeProtocol) {
              formData.nodeUrl = buildNodeUrl(formData.nodeProtocol, formData.nodeHost, value, formData.nodeUrl)
            }
          },
        },
        rules: [
          { required: true, type: 'number', message: '请输入节点端口', trigger: ['blur', 'change'] },
        ],
      },
      {
        field: 'nodeWeight',
        label: '节点权重',
        type: 'number' as const,
        placeholder: '权重值',
        span: 12,
        tabKey: 'basic',
        required: true,
        defaultValue: 100,
        props: {
          min: 1,
          max: 1000,
        },
        rules: [
          { required: true, type: 'number', message: '请输入节点权重', trigger: ['blur', 'change'] },
        ],
      },
      {
        field: 'nodeUrl',
        label: '节点URL',
        type: 'input' as const,
        placeholder: '输入完整URL或由上方字段自动生成',
        span: 24,
        tabKey: 'basic',
        required: true,
        tips: '支持直接输入完整URL(如 https://www.example.com)，将自动解析为协议、主机和端口',
        props: {
          onUpdateValue: (value: string, formData?: Record<string, any>) => {
            // 当输入完整URL时，自动解析为协议、主机和端口
            if (formData && value && (value.startsWith('http://') || value.startsWith('https://'))) {
              try {
                const url = new URL(value)
                formData.nodeProtocol = url.protocol === 'https:' ? 'HTTPS' : 'HTTP'
                formData.nodeHost = url.hostname
                if (url.port) {
                  formData.nodePort = parseInt(url.port, 10)
                } else {
                  formData.nodePort = url.protocol === 'https:' ? 443 : 80
                }
              } catch (error) {
                console.warn('URL解析失败:', error)
              }
            }
          },
        },
        rules: [
          { required: true, message: '请输入节点URL', trigger: ['blur', 'input'] },
          {
            validator: (_rule: any, value: any) => {
              if (value && (value.startsWith('http://') || value.startsWith('https://'))) {
                try {
                  new URL(value)
                  return true
                } catch {
                  return new Error('请输入有效的URL格式')
                }
              }
              return true
            },
            trigger: ['blur', 'input'],
          },
        ],
      },
      // ============= nodeStatus 字段暂时未使用，已隐藏 =============
      // {
      //   field: 'nodeStatus',
      //   label: '运行状态',
      //   type: 'select' as const,
      //   placeholder: '请选择运行状态',
      //   span: 12,
      //   tabKey: 'basic',
      //   required: true,
      //   defaultValue: NodeStatus.ONLINE,
      //   options: [
      //     { label: '在线', value: NodeStatus.ONLINE },
      //     { label: '下线', value: NodeStatus.OFFLINE },
      //     { label: '维护', value: NodeStatus.MAINTENANCE },
      //   ],
      //   rules: [
      //     { required: true, type: 'number', message: '请选择运行状态', trigger: ['blur', 'change'] },
      //   ],
      // },
      {
        field: 'healthStatus',
        label: '健康状态',
        type: 'switch' as const,
        span: 12,
        tabKey: 'basic',
        defaultValue: 'Y',
        props: {
          checkedValue: 'Y',
          uncheckedValue: 'N',
        },
      },
      {
        field: 'activeFlag',
        label: '在线状态',
        type: 'switch' as const,
        span: 12,
        tabKey: 'basic',
        defaultValue: 'Y',
        props: {
          checkedValue: 'Y',
          uncheckedValue: 'N',
        },
      },
      {
        field: 'noteText',
        label: '备注信息',
        type: 'textarea' as const,
        placeholder: '请输入备注信息',
        span: 24,
        tabKey: 'basic',
        props: {
          rows: 4,
        },
      },
      // ============= 元数据配置 Tab =============
      {
        field: 'nodeMetadata',
        label: '节点元数据',
        type: 'textarea' as const,
        placeholder: '{}',
        span: 24,
        tabKey: 'metadata',
        props: {
          rows: 8,
        },
        rules: [
          {
            validator: (_rule: any, value: any) => {
              if (value && typeof value === 'string' && value.trim()) {
                try {
                  JSON.parse(value)
                  return true
                } catch {
                  return new Error('请输入有效的JSON格式')
                }
              }
              return true
            },
            trigger: ['blur'],
          },
        ],
      },
      // ============= 其他配置 Tab =============
      {
        field: 'addTime',
        label: '创建时间',
        type: 'datetime' as const,
        span: 12,
        tabKey: 'other',
        disabled: true,
      },
      {
        field: 'addWho',
        label: '创建人',
        type: 'input' as const,
        span: 12,
        tabKey: 'other',
        disabled: true,
      },
      {
        field: 'editTime',
        label: '修改时间',
        type: 'datetime' as const,
        span: 12,
        tabKey: 'other',
        disabled: true,
      },
      {
        field: 'editWho',
        label: '修改人',
        type: 'input' as const,
        span: 12,
        tabKey: 'other',
        disabled: true,
      },
    ] as DataFormField[],
  }

  // ============= 表格配置 =============

  /** 表格配置（符合 RsGrid 结构） */
  const gridConfig: ServiceNodeGridConfig = {
    columns: [
      {
        key: 'serviceNodeId',
        title: '服务节点ID',
        visible: false,
      },
      {
        key: 'tenantId',
        title: '租户ID',
        visible: false,
      },
      {
        key: 'serviceDefinitionId',
        title: '服务定义ID',
        visible: false,
      },
      {
        key: 'nodeId',
        title: '节点ID',
        sortable: true,
        align: 'center',
        ellipsis: true,
        width: 150,
      },
      {
        key: 'nodeUrl',
        title: '节点地址',
        sortable: true,
        align: 'left',
        ellipsis: true,
        width: 250,
      },
      {
        key: 'nodeHost',
        title: '节点主机',
        sortable: true,
        align: 'center',
        ellipsis: true,
        width: 150,
      },
      {
        key: 'nodePort',
        title: '节点端口',
        sortable: true,
        align: 'center',
        width: 100,
      },
      {
        key: 'nodeProtocol',
        title: '协议',
        align: 'center',
        width: 80,
        render: (row) =>
          h(
            RsTag,
            {
              variant: row.nodeProtocol === 'HTTPS' ? 'success' : 'info',
              size: 'sm',
            },
            () => row.nodeProtocol,
          ),
      },
      {
        key: 'nodeWeight',
        title: '权重',
        align: 'center',
        width: 80,
      },
      {
        key: 'healthStatus',
        title: '健康状态',
        align: 'center',
        width: 100,
        render: (row) =>
          h(
            RsTag,
            {
              variant: row.healthStatus === 'Y' ? 'success' : 'danger',
              size: 'sm',
            },
            () => (row.healthStatus === 'Y' ? '健康' : '不健康'),
          ),
      },
      {
        key: 'activeFlag',
        title: '在线状态',
        align: 'center',
        width: 100,
        render: (row) =>
          h(
            RsTag,
            {
              variant: row.activeFlag === 'Y' ? 'success' : 'default',
              size: 'sm',
            },
            () => (row.activeFlag === 'Y' ? '启用' : '禁用'),
          ),
      },
      {
        key: 'lastHealthCheckTime',
        title: '最后检查时间',
        align: 'center',
        ellipsis: true,
        formatter: (value) =>
          value ? formatDate(value as string, 'YYYY-MM-DD HH:mm:ss') : '-',
        width: 180,
      },
      {
        key: 'noteText',
        title: '备注',
        align: 'left',
        ellipsis: true,
        width: 150,
      },
      {
        key: 'addTime',
        title: '创建时间',
        sortable: true,
        align: 'center',
        ellipsis: true,
        formatter: (value) =>
          value ? formatDate(value as string, 'YYYY-MM-DD HH:mm:ss') : '',
        width: 180,
      },
      {
        key: 'addWho',
        title: '创建人',
        align: 'center',
        ellipsis: true,
        width: 120,
      },
      {
        key: 'editTime',
        title: '修改时间',
        sortable: true,
        align: 'center',
        ellipsis: true,
        formatter: (value) =>
          value ? formatDate(value as string, 'YYYY-MM-DD HH:mm:ss') : '',
        width: 180,
      },
      {
        key: 'editWho',
        title: '修改人',
        align: 'center',
        ellipsis: true,
        width: 120,
      },
    ],
    selectable: true,
    rowKey: 'serviceNodeId',
    paginationConfig: {
      show: true,
      pageInfo: pageInfo as any,
      align: 'right',
    },
    menuConfig: {
      enabled: true,
      items: [
        { key: 'edit', label: '编辑', icon: 'pencil' },
        { key: 'delete', label: '删除', icon: 'trash-2', danger: true },
      ],
    },
    height: '100%',
  }

  // ============= 辅助方法 =============

  /**
   * 根据 nodeHost、nodePort、nodeProtocol 自动更新 nodeUrl
   * 用于表单字段的 onUpdateValue 回调
   */

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
   * 设置服务节点列表
   */
  const setNodeList = (list: ServiceNode[]) => {
    nodeList.value = list
  }

  /**
   * 在列表中添加服务节点
   */
  const addNodeToList = (node: ServiceNode) => {
    nodeList.value.push(node)
  }

  /**
   * 更新列表中的服务节点
   */
  const updateNodeInList = (node: ServiceNode) => {
    const index = nodeList.value.findIndex((item) => item.serviceNodeId === node.serviceNodeId)
    if (index >= 0) {
      nodeList.value[index] = node
    }
  }

  /**
   * 从列表中移除服务节点
   */
  const removeNodeFromList = (serviceNodeId: string) => {
    const index = nodeList.value.findIndex((item) => item.serviceNodeId === serviceNodeId)
    if (index >= 0) {
      nodeList.value.splice(index, 1)
    }
  }

  /**
   * 从列表中批量移除服务节点
   */
  const removeNodesFromList = (serviceNodeIds: string[]) => {
    nodeList.value = nodeList.value.filter((item) => !serviceNodeIds.includes(item.serviceNodeId))
  }

  return {
    // 数据状态
    moduleId,
    loading,
    nodeList,
    pageInfo,

    // 配置
    searchFormConfig,
    nodeFormConfig,
    gridConfig,

    // 方法
    resetPagination,
    updatePagination,
    setNodeList,
    addNodeToList,
    updateNodeInList,
    removeNodeFromList,
    removeNodesFromList,
  }
}

/**
 * 服务节点管理 Model 类型
 */
export type ServiceNodeModel = ReturnType<typeof useServiceNodeModel>

