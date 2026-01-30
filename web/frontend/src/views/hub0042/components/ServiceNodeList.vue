<template>
  <div class="service-node-list" id="hub0042-node">
    <!-- 节点表格 -->
    <g-grid
      :module-id="moduleId"
      :data="nodes"
      :loading="loading"
      v-bind="nodeGridConfig"
      @menu-click="handleMenuClick"
    >
      <!-- 临时实例自定义渲染 -->
      <template #ephemeral="{ row }">
        <n-tag :type="row.ephemeral === 'Y' ? 'warning' : 'default'" size="small">
          {{ row.ephemeral === 'Y' ? '是' : '否' }}
        </n-tag>
      </template>

      <!-- 实例状态自定义渲染 -->
      <template #instanceStatus="{ row }">
        <n-tag :type="getInstanceStatusType(row.instanceStatus)" size="small">
          {{ getInstanceStatusLabel(row.instanceStatus) }}
        </n-tag>
      </template>

      <!-- 健康状态自定义渲染 -->
      <template #healthyStatus="{ row }">
        <n-tag :type="getHealthStatusType(row.healthyStatus)" size="small">
          {{ getHealthStatusLabel(row.healthyStatus) }}
        </n-tag>
      </template>

      <!-- 元数据自定义渲染 -->
      <template #metadataJson="{ row }">
        <n-ellipsis :line-clamp="1" :tooltip="false">
          {{ formatMetadata(row.metadataJson) }}
        </n-ellipsis>
      </template>

      <!-- 心跳时间自定义渲染 -->
      <template #lastBeatTime="{ row }">
        {{ formatTime(row.lastBeatTime) }}
      </template>

      <!-- 操作列 -->
      <template #action="{ row }">
        <n-space :size="8">
          <n-button size="small" type="primary" @click="handleEditNode(row)">
            编辑
          </n-button>
          <n-button
            v-if="row.instanceStatus === 'UP'"
            size="small"
            type="warning"
            @click="handleOfflineNode(row)"
          >
            下线
          </n-button>
          <n-button
            v-if="row.instanceStatus === 'DOWN'"
            size="small"
            type="success"
            @click="handleOnlineNode(row)"
          >
            上线
          </n-button>
        </n-space>
      </template>
    </g-grid>

    <!-- 编辑节点模态框 -->
    <GdataFormModal
      v-model:visible="editDialogVisible"
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
import GdataFormModal from '@/components/form/data/GDataFormModal.vue'
import type { DataFormField } from '@/components/form/data/types'
import type { GridProps } from '@/components/grid'
import { GGrid } from '@/components/grid'
import { formatDate } from '@/utils/format'
import { NButton, NEllipsis, NSpace, NTag, useDialog, useMessage } from 'naive-ui'
import { ref } from 'vue'
import { editNode, offlineNode, onlineNode } from '../api'
import type { ServiceNode } from '../types'

defineOptions({
  name: 'ServiceNodeList'
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

// 对话框和表单状态
const dialog = useDialog()
const message = useMessage()
const editDialogVisible = ref(false)
const currentEditNode = ref<ServiceNode | null>(null)
const submitting = ref(false)

// 节点表单配置
const nodeFormFields: DataFormField[] = [
  {
    field: 'nodeId',
    label: '节点ID',
    type: 'input' as const,
    disabled: true,
    required: true,
  },
  {
    field: 'ipAddress',
    label: 'IP地址',
    type: 'input' as const,
    disabled: true,
    required: true,
  },
  {
    field: 'portNumber',
    label: '端口号',
    type: 'input' as const,
    disabled: true,
    required: true,
  },
  {
    field: 'weight',
    label: '权重',
    type: 'input' as const,
    required: true,
    defaultValue: 1,
  },
  {
    field: 'instanceStatus',
    label: '实例状态',
    type: 'select' as const,
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
    type: 'select' as const,
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
    type: 'select' as const,
    required: true,
    options: [
      { label: '是', value: 'Y' },
      { label: '否', value: 'N' },
    ],
  },
  {
    field: 'metadataJson',
    label: '元数据',
    type: 'textarea' as const,
    placeholder: 'JSON 格式',
  },
]

// 节点表格配置
const nodeGridConfig: Omit<GridProps, 'moduleId' | 'data' | 'loading'> = {
  columns: [
    {
      field: 'nodeId',
      title: '节点ID',
      align: 'center',
      width: 200,
      showOverflow: true,
      sortable: true,
      filters: [
        { data: '' }
      ],
      filterRender: { name: 'input' },
    },
    {
      field: 'ipAddress',
      title: 'IP',
      align: 'center',
      showOverflow: true,
      sortable: true,
      filters: [
        { data: '' }
      ],
      filterRender: { name: 'input' },
    },
    {
      field: 'portNumber',
      title: '端口',
      align: 'center',
      width: 100,
      sortable: true,
      filters: [
        { data: '' }
      ],
      filterRender: { name: 'input' },
    },
    {
      field: 'ephemeral',
      title: '临时实例',
      align: 'center',
      width: 100,
      slots: { default: 'ephemeral' },
      filterRender: { 
        name: 'select', 
        options: [
          { label: '全部', value: '' },
          { label: '是', value: 'Y' },
          { label: '否', value: 'N' },
        ]
      },
    },
    {
      field: 'weight',
      title: '权重',
      align: 'center',
      width: 100,
      sortable: true,
      filters: [
        { data: '' }
      ],
      filterRender: { name: 'input' },
    },
    {
      field: 'instanceStatus',
      title: '实例状态',
      align: 'center',
      width: 120,
      slots: { default: 'instanceStatus' },
      filterRender: { 
        name: 'select', 
        options: [
          { label: '全部', value: '' },
          { label: '运行中', value: 'UP' },
          { label: '已下线', value: 'DOWN' },
          { label: '启动中', value: 'STARTING' },
          { label: '停止服务', value: 'OUT_OF_SERVICE' },
        ]
      },
    },
    {
      field: 'healthyStatus',
      title: '健康状态',
      align: 'center',
      width: 120,
      slots: { default: 'healthyStatus' },
      filterRender: { 
        name: 'select', 
        options: [
          { label: '全部', value: '' },
          { label: '健康', value: 'HEALTHY' },
          { label: '不健康', value: 'UNHEALTHY' },
          { label: '未知', value: 'UNKNOWN' },
        ]
      },
    },
    {
      field: 'metadataJson',
      title: '元数据',
      align: 'left',
      showOverflow: true,
      slots: { default: 'metadataJson' },
    },
    {
      field: 'lastBeatTime',
      title: '心跳时间',
      align: 'center',
      width: 180,
      sortable: true,
      slots: { default: 'lastBeatTime' },
    },
    {
      field: 'action',
      title: '操作',
      align: 'center',
      width: 160,
      slots: { default: 'action' },
    },
  ],
  filterConfig: {
    remote: false, // 本地过滤
  },
  showCheckbox: false,
  showSeq: true,
  paginationConfig: {
    show: false, // 不分页
  },
  menuConfig: {
    enabled: true,
    showCopyRow: false,
    showCopyCell: false,
    customMenus: [
      { code: 'edit', name: '编辑' },
      { code: 'online', name: '上线' },
      { code: 'offline', name: '下线' },
      { code: 'refresh', name: '刷新' },
    ],
  },
  height: '100%',
}

// 工具方法
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
  const statusMap: Record<string, 'success' | 'error' | 'warning' | 'info'> = {
    'UP': 'success',
    'DOWN': 'error',
    'STARTING': 'warning',
    'OUT_OF_SERVICE': 'info',
  }
  return statusMap[status] || 'default'
}

const getInstanceStatusLabel = (status: string) => {
  const statusMap: Record<string, string> = {
    'UP': '运行中',
    'DOWN': '已下线',
    'STARTING': '启动中',
    'OUT_OF_SERVICE': '停止服务',
  }
  return statusMap[status] || status
}

const getHealthStatusType = (status: string) => {
  const statusMap: Record<string, 'success' | 'error' | 'warning'> = {
    'HEALTHY': 'success',
    'UNHEALTHY': 'error',
    'UNKNOWN': 'warning',
  }
  return statusMap[status] || 'default'
}

const getHealthStatusLabel = (status: string) => {
  const statusMap: Record<string, string> = {
    'HEALTHY': '健康',
    'UNHEALTHY': '不健康',
    'UNKNOWN': '未知',
  }
  return statusMap[status] || status
}

// 格式化时间
const formatTime = (timeStr?: string) => {
  if (!timeStr) return '-'
  return formatDate(timeStr, 'YYYY-MM-DD HH:mm:ss') || '-'
}

// 菜单点击事件处理
const handleMenuClick = ({ code, row }: { code: string; row?: any }) => {
  // 刷新操作不需要选中行
  if (code === 'refresh') {
    handleRefresh()
    return
  }
  
  if (!row) return
  const node = row as ServiceNode
  switch (code) {
    case 'edit':
      handleEditNode(node)
      break
    case 'online':
      handleOnlineNodeConfirm(node)
      break
    case 'offline':
      handleOfflineNodeConfirm(node)
      break
  }
}

// 事件处理
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

// 编辑表单提交
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

// 上线节点确认
const handleOnlineNodeConfirm = (node: ServiceNode) => {
  dialog.warning({
    title: '确认上线',
    content: `确定要上线节点 ${node.ipAddress}:${node.portNumber} 吗？`,
    positiveText: '确定',
    negativeText: '取消',
    onPositiveClick: async () => {
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
    },
  })
}

// 下线节点确认
const handleOfflineNodeConfirm = (node: ServiceNode) => {
  dialog.warning({
    title: '确认下线',
    content: `确定要下线节点 ${node.ipAddress}:${node.portNumber} 吗？`,
    positiveText: '确定',
    negativeText: '取消',
    onPositiveClick: async () => {
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
    },
  })
}
</script>

<style lang="scss" scoped>
.service-node-list {
  width: 100%;
  height: 100%;
}
</style>

