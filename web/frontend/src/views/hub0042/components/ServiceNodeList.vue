<template>
  <div class="service-node-list" id="hub0042-node">
    <RsGrid
      :module-id="moduleId"
      :data="nodes"
      :loading="loading"
      :columns="nodeGridConfig.columns"
      :selectable="false"
      :show-index="true"
      row-key="nodeId"
      height="100%"
      :pagination-config="nodeGridConfig.paginationConfig"
      :menu-config="nodeGridConfig.menuConfig"
      @menu-click="handleMenuClick"
    />

    <RsDataFormModal
      v-model:visible="editDialogVisible"
      module-id="hub0042:node"
      mode="edit"
      title="编辑节点"
      to="#hub0042-node"
      :form-fields="nodeFormFields"
      :initial-data="currentEditNode || undefined"
      :auto-close-on-confirm="false"
      :confirm-loading="submitting"
      @submit="handleFormSubmit"
    />
  </div>
</template>

<script lang="ts" setup>
import { RsDataFormModal, type RsDataFormField } from '@/components/form/rs-data'
import type { RsGridColumn, RsGridMenuConfig, RsGridPaginationConfig } from '@/components/rs-grid'
import { RsGrid } from '@/components/rs-grid'
import { useAppMessage } from '@/composables/useAppMessage'
import { RsButton, RsTag, rsConfirm } from '@/ui'
import { formatDate } from '@/utils/format'
import { h, ref } from 'vue'
import { editNode, offlineNode, onlineNode } from '../api'
import type { ServiceNode } from '../types'

defineOptions({
  name: 'ServiceNodeList',
})

interface Props {
  nodes?: ServiceNode[]
  loading?: boolean
  moduleId?: string
}

const props = withDefaults(defineProps<Props>(), {
  nodes: () => [],
  loading: false,
  moduleId: 'hub0042:node',
})

interface Emits {
  (e: 'refresh'): void
}

const emit = defineEmits<Emits>()

const message = useAppMessage()
const editDialogVisible = ref(false)
const currentEditNode = ref<ServiceNode | null>(null)
const submitting = ref(false)

const nodeFormFields: RsDataFormField[] = [
  {
    field: 'nodeId',
    label: '节点ID',
    type: 'input',
    disabled: true,
    required: true,
  },
  {
    field: 'ipAddress',
    label: 'IP地址',
    type: 'input',
    disabled: true,
    required: true,
  },
  {
    field: 'portNumber',
    label: '端口号',
    type: 'number',
    disabled: true,
    required: true,
  },
  {
    field: 'weight',
    label: '权重',
    type: 'number',
    required: true,
    defaultValue: 1,
  },
  {
    field: 'instanceStatus',
    label: '实例状态',
    type: 'select',
    required: true,
    options: [
      { label: '运行中', value: 'UP' },
      { label: '已下线', value: 'DOWN' },
      { label: '启动中', value: 'STARTING' },
      { label: '停止服务', value: 'OUT_OF_SERVICE' },
    ],
  },
  {
    field: 'healthyStatus',
    label: '健康状态',
    type: 'select',
    required: true,
    options: [
      { label: '健康', value: 'HEALTHY' },
      { label: '不健康', value: 'UNHEALTHY' },
      { label: '未知', value: 'UNKNOWN' },
    ],
  },
  {
    field: 'ephemeral',
    label: '临时实例',
    type: 'select',
    required: true,
    options: [
      { label: '是', value: 'Y' },
      { label: '否', value: 'N' },
    ],
  },
  {
    field: 'metadataJson',
    label: '元数据',
    type: 'textarea',
    placeholder: 'JSON 格式',
  },
]

/**
 * 格式化节点元数据为可读文本。
 */
const formatMetadata = (metadataJson?: string) => {
  if (!metadataJson) return '-'
  try {
    const metadata = JSON.parse(metadataJson)
    return Object.entries(metadata)
      .map(([key, value]) => `${key}=${value}`)
      .join(', ')
  } catch {
    return metadataJson
  }
}

const getInstanceStatusType = (status: string) => {
  const statusMap: Record<string, 'success' | 'danger' | 'warning' | 'info' | 'default'> = {
    UP: 'success',
    DOWN: 'danger',
    STARTING: 'warning',
    OUT_OF_SERVICE: 'info',
  }
  return statusMap[status] || 'default'
}

const getInstanceStatusLabel = (status: string) => {
  const statusMap: Record<string, string> = {
    UP: '运行中',
    DOWN: '已下线',
    STARTING: '启动中',
    OUT_OF_SERVICE: '停止服务',
  }
  return statusMap[status] || status
}

const getHealthStatusType = (status: string) => {
  const statusMap: Record<string, 'success' | 'danger' | 'warning' | 'default'> = {
    HEALTHY: 'success',
    UNHEALTHY: 'danger',
    UNKNOWN: 'warning',
  }
  return statusMap[status] || 'default'
}

const getHealthStatusLabel = (status: string) => {
  const statusMap: Record<string, string> = {
    HEALTHY: '健康',
    UNHEALTHY: '不健康',
    UNKNOWN: '未知',
  }
  return statusMap[status] || status
}

const formatTime = (timeStr?: string) => {
  if (!timeStr) return '-'
  return formatDate(timeStr, 'YYYY-MM-DD HH:mm:ss') || '-'
}

const nodeGridConfig: {
  columns: RsGridColumn<ServiceNode>[]
  paginationConfig: RsGridPaginationConfig
  menuConfig: RsGridMenuConfig
} = {
  columns: [
    {
      key: 'nodeId',
      title: '节点ID',
      align: 'center',
      width: 200,
      ellipsis: true,
      sortable: true,
      filterable: true,
    },
    {
      key: 'ipAddress',
      title: 'IP',
      align: 'center',
      ellipsis: true,
      sortable: true,
      filterable: true,
    },
    {
      key: 'portNumber',
      title: '端口',
      align: 'center',
      width: 100,
      sortable: true,
      filterable: true,
    },
    {
      key: 'ephemeral',
      title: '临时实例',
      align: 'center',
      width: 100,
      filterable: true,
      render: (row) =>
        h(
          RsTag,
          { variant: row.ephemeral === 'Y' ? 'warning' : 'default', size: 'sm' },
          () => (row.ephemeral === 'Y' ? '是' : '否'),
        ),
    },
    {
      key: 'weight',
      title: '权重',
      align: 'center',
      width: 100,
      sortable: true,
      filterable: true,
    },
    {
      key: 'instanceStatus',
      title: '实例状态',
      align: 'center',
      width: 120,
      filterable: true,
      render: (row) =>
        h(
          RsTag,
          { variant: getInstanceStatusType(row.instanceStatus), size: 'sm' },
          () => getInstanceStatusLabel(row.instanceStatus),
        ),
    },
    {
      key: 'healthyStatus',
      title: '健康状态',
      align: 'center',
      width: 120,
      filterable: true,
      render: (row) =>
        h(
          RsTag,
          { variant: getHealthStatusType(row.healthyStatus), size: 'sm' },
          () => getHealthStatusLabel(row.healthyStatus),
        ),
    },
    {
      key: 'metadataJson',
      title: '元数据',
      align: 'left',
      ellipsis: true,
      formatter: (value) => formatMetadata(value as string | undefined),
    },
    {
      key: 'lastBeatTime',
      title: '心跳时间',
      align: 'center',
      width: 180,
      sortable: true,
      formatter: (value) => formatTime(value as string | undefined),
    },
    {
      key: 'action',
      title: '操作',
      align: 'center',
      width: 180,
      render: (row) =>
        h('div', { class: 'node-actions' }, [
          h(RsButton, { size: 'sm', variant: 'primary', onClick: () => handleEditNode(row) }, () => '编辑'),
          row.instanceStatus === 'UP'
            ? h(
                RsButton,
                { size: 'sm', variant: 'ghost', tone: 'warning', onClick: () => handleOfflineNode(row) },
                () => '下线',
              )
            : null,
          row.instanceStatus === 'DOWN'
            ? h(
                RsButton,
                { size: 'sm', variant: 'ghost', tone: 'success', onClick: () => handleOnlineNode(row) },
                () => '上线',
              )
            : null,
        ]),
    },
  ],
  paginationConfig: {
    show: false,
  },
  menuConfig: {
    enabled: true,
    items: [
      { key: 'edit', label: '编辑', icon: 'pencil' },
      { key: 'online', label: '上线', icon: 'play' },
      { key: 'offline', label: '下线', icon: 'square' },
      { key: 'refresh', label: '刷新', icon: 'refresh-cw' },
    ],
  },
}

/**
 * 菜单点击事件处理
 */
const handleMenuClick = ({ key, row }: { key: string; row?: ServiceNode }) => {
  if (key === 'refresh') {
    handleRefresh()
    return
  }
  if (!row) return
  switch (key) {
    case 'edit':
      handleEditNode(row)
      break
    case 'online':
      handleOnlineNodeConfirm(row)
      break
    case 'offline':
      handleOfflineNodeConfirm(row)
      break
  }
}

const handleRefresh = () => {
  message.info('正在刷新节点列表...')
  emit('refresh')
}

const handleEditNode = (node: ServiceNode) => {
  currentEditNode.value = { ...node }
  editDialogVisible.value = true
}

const handleOnlineNode = (node: ServiceNode) => {
  handleOnlineNodeConfirm(node)
}

const handleOfflineNode = (node: ServiceNode) => {
  handleOfflineNodeConfirm(node)
}

const handleFormSubmit = async (formData?: Record<string, any>) => {
  if (!formData || !formData.nodeId) {
    message.error('节点ID不能为空')
    return
  }

  submitting.value = true
  try {
    const res = await editNode(formData as Partial<ServiceNode> & { nodeId: string })
    if (res.oK) {
      message.success('节点编辑成功')
      editDialogVisible.value = false
      currentEditNode.value = null
      emit('refresh')
    } else {
      message.error(res.messageId || '节点编辑失败')
    }
  } catch (error: any) {
    message.error(error.message || '节点编辑失败')
  } finally {
    submitting.value = false
  }
}

const handleOnlineNodeConfirm = async (node: ServiceNode) => {
  const confirmed = await rsConfirm.warning({
    title: '确认上线',
    description: `确定要上线节点 ${node.ipAddress}:${node.portNumber} 吗？`,
    confirmText: '确定',
    cancelText: '取消',
  })
  if (!confirmed) return

  try {
    const res = await onlineNode(node.nodeId)
    if (res.oK) {
      message.success('节点上线成功')
      emit('refresh')
    } else {
      message.error(res.messageId || '节点上线失败')
    }
  } catch (error: any) {
    message.error(error.message || '节点上线失败')
  }
}

const handleOfflineNodeConfirm = async (node: ServiceNode) => {
  const confirmed = await rsConfirm.warning({
    title: '确认下线',
    description: `确定要下线节点 ${node.ipAddress}:${node.portNumber} 吗？`,
    confirmText: '确定',
    cancelText: '取消',
  })
  if (!confirmed) return

  try {
    const res = await offlineNode(node.nodeId)
    if (res.oK) {
      message.success('节点下线成功')
      emit('refresh')
    } else {
      message.error(res.messageId || '节点下线失败')
    }
  } catch (error: any) {
    message.error(error.message || '节点下线失败')
  }
}
</script>

<style lang="scss" scoped>
.service-node-list {
  width: 100%;
  height: 100%;
  min-height: 0;
}

:deep(.node-actions) {
  display: flex;
  gap: 8px;
  justify-content: center;
  align-items: center;
}
</style>
