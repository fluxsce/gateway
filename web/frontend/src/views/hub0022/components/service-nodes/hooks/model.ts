/**
 * 服务节点管理列表 Model
 * 统一管理搜索表单、表格配置和数据状态
 */

import type { DataFormField, DataFormTab } from '@/components/form/data/types'
import type { SearchFormProps } from '@/components/form/search/types'
import type { GridProps } from '@/components/grid'
import type { PageInfoObj } from '@/types/api'
import { formatDate } from '@/utils/format'
import { AddOutline, CreateOutline, TrashOutline } from '@vicons/ionicons5'
import { ref } from 'vue'
import type { ServiceNode } from '../types'
import { NodeStatus } from '../types'

/**
 * 服务节点管理列表 Model
 */
export function useServiceNodeModel() {
  // ============= 数据状态 =============
  const moduleId = 'hub0022'
  
  /** 加载状态 */
  const loading = ref(false)

  /** 服务节点列表数据 */
  const nodeList = ref<ServiceNode[]>([])

  /** 后端分页信息对象 */
  const pageInfo = ref<PageInfoObj | undefined>()

  // ============= 搜索表单配置 =============

  /** 搜索表单配置（符合 SearchFormProps 结构） */
  const searchFormConfig: Omit<SearchFormProps, 'moduleId'> = {
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
        label: '启用状态',
        type: 'select',
        placeholder: '请选择启用状态',
        span: 6,
        clearable: true,
        options: [
          { label: '全部', value: '' },
          { label: '启用', value: 'Y' },
          { label: '禁用', value: 'N' },
        ],
      },
      {
        field: 'nodeStatus',
        label: '运行状态',
        type: 'select',
        placeholder: '请选择运行状态',
        span: 6,
        clearable: true,
        options: [
          { label: '全部', value: '' },
          { label: '在线', value: NodeStatus.ONLINE },
          { label: '下线', value: NodeStatus.OFFLINE },
          { label: '维护', value: NodeStatus.MAINTENANCE },
        ],
      },
    ],
    toolbarButtons: [
      {
        key: 'add',
        label: '新建节点',
        icon: AddOutline,
        type: 'primary',
        tooltip: '新建服务节点',
      },
      {
        key: 'edit',
        label: '编辑',
        icon: CreateOutline,
        type: 'default',
        tooltip: '编辑选中的服务节点',
      },
      {
        key: 'delete',
        label: '删除',
        icon: TrashOutline,
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
      { key: 'status', label: '状态配置' },
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
            // 当协议变化时，自动更新 nodeUrl
            if (formData && formData.nodeHost && formData.nodePort !== undefined && value) {
              const protocol = value.toLowerCase()
              formData.nodeUrl = `${protocol}://${formData.nodeHost}:${formData.nodePort}`
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
            // 当主机地址变化时，自动更新 nodeUrl
            if (formData && value && formData.nodePort !== undefined && formData.nodeProtocol) {
              const protocol = formData.nodeProtocol.toLowerCase()
              formData.nodeUrl = `${protocol}://${value}:${formData.nodePort}`
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
            // 当端口变化时，自动更新 nodeUrl
            if (formData && formData.nodeHost && value !== undefined && formData.nodeProtocol) {
              const protocol = formData.nodeProtocol.toLowerCase()
              formData.nodeUrl = `${protocol}://${formData.nodeHost}:${value}`
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
      {
        field: 'nodeStatus',
        label: '运行状态',
        type: 'select' as const,
        placeholder: '请选择运行状态',
        span: 12,
        tabKey: 'basic',
        required: true,
        defaultValue: NodeStatus.ONLINE,
        options: [
          { label: '在线', value: NodeStatus.ONLINE },
          { label: '下线', value: NodeStatus.OFFLINE },
          { label: '维护', value: NodeStatus.MAINTENANCE },
        ],
        rules: [
          { required: true, type: 'number', message: '请选择运行状态', trigger: ['blur', 'change'] },
        ],
      },
      // ============= 状态配置 Tab =============
      {
        field: 'healthStatus',
        label: '健康状态',
        type: 'switch' as const,
        span: 12,
        tabKey: 'status',
        defaultValue: 'Y',
        props: {
          checkedValue: 'Y',
          uncheckedValue: 'N',
        },
      },
      {
        field: 'activeFlag',
        label: '启用状态',
        type: 'switch' as const,
        span: 12,
        tabKey: 'status',
        defaultValue: 'Y',
        props: {
          checkedValue: 'Y',
          uncheckedValue: 'N',
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
        field: 'noteText',
        label: '备注信息',
        type: 'textarea' as const,
        placeholder: '请输入备注信息',
        span: 24,
        tabKey: 'other',
        props: {
          rows: 4,
        },
      },
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

  /** 表格配置（符合 GridProps 结构，排除响应式数据） */
  const gridConfig: Omit<GridProps, 'moduleId' | 'data' | 'loading'> = {
    columns: [
      // ============= 主键字段（隐藏，但必须存在用于数据操作） =============
      {
        field: 'serviceNodeId',
        title: '服务节点ID',
        visible: false,
      },
      {
        field: 'tenantId',
        title: '租户ID',
        visible: false,
      },
      {
        field: 'serviceDefinitionId',
        title: '服务定义ID',
        visible: false,
      },
      // ============= 业务字段 =============
      {
        field: 'nodeId',
        title: '节点ID',
        sortable: true,
        align: 'center',
        showOverflow: true,
        width: 150,
      },
      {
        field: 'nodeUrl',
        title: '节点地址',
        sortable: true,
        align: 'left',
        showOverflow: true,
        width: 250,
      },
      {
        field: 'nodeHost',
        title: '节点主机',
        sortable: true,
        align: 'center',
        showOverflow: true,
        width: 150,
      },
      {
        field: 'nodePort',
        title: '节点端口',
        sortable: true,
        align: 'center',
        width: 100,
      },
      {
        field: 'nodeProtocol',
        title: '协议',
        align: 'center',
        width: 80,
        slots: { default: 'nodeProtocol' },
      },
      {
        field: 'nodeWeight',
        title: '权重',
        align: 'center',
        width: 80,
      },
      {
        field: 'healthStatus',
        title: '健康状态',
        align: 'center',
        width: 100,
        slots: { default: 'healthStatus' },
      },
      {
        field: 'nodeStatus',
        title: '运行状态',
        align: 'center',
        width: 100,
        slots: { default: 'nodeStatus' },
      },
      {
        field: 'activeFlag',
        title: '启用状态',
        align: 'center',
        width: 100,
        slots: { default: 'activeFlag' },
      },
      {
        field: 'lastHealthCheckTime',
        title: '最后检查时间',
        align: 'center',
        showOverflow: true,
        formatter: ({ cellValue }) =>
          cellValue ? formatDate(cellValue, 'YYYY-MM-DD HH:mm:ss') : '-',
        width: 180,
      },
      {
        field: 'noteText',
        title: '备注',
        align: 'left',
        showOverflow: true,
        width: 150,
      },
      {
        field: 'addTime',
        title: '创建时间',
        sortable: true,
        align: 'center',
        showOverflow: true,
        formatter: ({ cellValue }) =>
          cellValue ? formatDate(cellValue, 'YYYY-MM-DD HH:mm:ss') : '',
        width: 180,
      },
      {
        field: 'addWho',
        title: '创建人',
        align: 'center',
        showOverflow: true,
        width: 120,
      },
      {
        field: 'editTime',
        title: '修改时间',
        sortable: true,
        align: 'center',
        showOverflow: true,
        formatter: ({ cellValue }) =>
          cellValue ? formatDate(cellValue, 'YYYY-MM-DD HH:mm:ss') : '',
        width: 180,
      },
      {
        field: 'editWho',
        title: '修改人',
        align: 'center',
        showOverflow: true,
        width: 120,
      },
    ],
    showCheckbox: true,
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
          code: 'edit',
          name: '编辑',
          prefixIcon: 'vxe-icon-edit',
        },
        {
          code: 'delete',
          name: '删除',
          prefixIcon: 'vxe-icon-delete',
        },
      ],
    },
    height: '100%',
    rowId: 'serviceNodeId',
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

