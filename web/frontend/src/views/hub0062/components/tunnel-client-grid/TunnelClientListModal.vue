<template>
  <RsDialog
    :open="modalVisible"
    title="选择客户端"
    layout="window"
    :width="900"
    :teleport-to="props.to"
    :draggable="true"
    :fullscreenable="true"
    :modal="false"
    :show-overlay="false"
    :close-on-overlay-click="false"
    class="hub0062-tunnel-client-list-dialog"
    @update:open="handleUpdateVisible"
  >
    <template #body>
      <div class="tunnel-client-list-modal">
        <RsGrid
          ref="gridRef"
          module-id="tunnel-client-selector-grid"
          :data="clientList"
          :loading="loading"
          :columns="gridColumns"
          :selectable="false"
          row-key="tunnelClientId"
          height="100%"
          :pagination-config="paginationConfig"
          @page-change="handlePageChange"
          @row-click="handleRowClick"
        />
      </div>
    </template>
  </RsDialog>
</template>

<script setup lang="ts">
import { RsGrid, type RsGridColumn, type RsGridExpose, type RsGridPaginationConfig } from '@/components/rs-grid'
import type { PageInfoObj } from '@/types/api'
import { RsDialog, RsTag, type RsTagVariant } from '@/ui'
import { createBackendPaginationParams } from '@/utils/pagination'
import { h, onMounted, ref, watch } from 'vue'
import * as tunnelClientApi from '../../api'
import type { ConnectionStatus, TunnelClient } from '../../types'

defineOptions({
  name: 'TunnelClientListModal',
})

interface Props {
  visible?: boolean
  to?: string
}

const props = withDefaults(defineProps<Props>(), {
  visible: false,
  to: 'body',
})

const emit = defineEmits<{
  'update:visible': [value: boolean]
  select: [client: TunnelClient]
}>()

const gridRef = ref<RsGridExpose | null>(null)
const loading = ref(false)
const clientList = ref<TunnelClient[]>([])
const pageInfo = ref<PageInfoObj | undefined>()
const modalVisible = ref(props.visible)

/**
 * 获取连接状态展示文案
 */
const getConnectionStatusLabel = (status?: ConnectionStatus): string => {
  const statusMap: Record<string, string> = {
    connected: '已连接',
    disconnected: '已断开',
    connecting: '连接中',
    error: '错误',
  }
  return statusMap[status || ''] || status || '-'
}

/**
 * 获取连接状态标签变体
 */
const getConnectionStatusVariant = (status?: ConnectionStatus): RsTagVariant => {
  const typeMap: Record<string, RsTagVariant> = {
    connected: 'success',
    disconnected: 'warning',
    connecting: 'info',
    error: 'danger',
  }
  return typeMap[status || ''] || 'default'
}

const gridColumns: RsGridColumn<TunnelClient>[] = [
  {
    key: 'tunnelClientId',
    title: '客户端ID',
    width: 200,
    ellipsis: true,
  },
  {
    key: 'clientName',
    title: '客户端名称',
    width: 180,
    ellipsis: true,
  },
  {
    key: 'serverAddress',
    title: '服务器地址',
    width: 200,
    ellipsis: true,
    formatter: (_value, row) => `${row.serverAddress}:${row.serverPort}`,
  },
  {
    key: 'connectionStatus',
    title: '连接状态',
    width: 100,
    align: 'center',
    render: (row) =>
      h(
        RsTag,
        { variant: getConnectionStatusVariant(row.connectionStatus), size: 'sm' },
        () => getConnectionStatusLabel(row.connectionStatus),
      ),
  },
  {
    key: 'activeFlag',
    title: '状态',
    width: 80,
    align: 'center',
    render: (row) =>
      h(
        RsTag,
        { variant: row.activeFlag === 'Y' ? 'success' : 'default', size: 'sm' },
        () => (row.activeFlag === 'Y' ? '启用' : '禁用'),
      ),
  },
]

const paginationConfig: RsGridPaginationConfig = {
  show: true,
  pageInfo: pageInfo as any,
  align: 'right',
}

/**
 * 加载客户端列表
 */
const loadClientList = async () => {
  loading.value = true
  try {
    const paginationParams = createBackendPaginationParams(
      pageInfo.value?.pageIndex,
      pageInfo.value?.pageSize,
    )

    const params = {
      activeFlag: 'Y' as const,
      pageIndex: paginationParams.pageIndex,
      pageSize: paginationParams.pageSize,
    }

    const response = await tunnelClientApi.queryTunnelClients(params)

    if (response.oK) {
      if (response.bizData) {
        const bizData = JSON.parse(response.bizData)
        clientList.value = Array.isArray(bizData) ? bizData : []
      }

      if (response.pageQueryData) {
        const backendPageInfo = JSON.parse(response.pageQueryData)
        if (!pageInfo.value) {
          pageInfo.value = backendPageInfo as PageInfoObj
        } else {
          Object.assign(pageInfo.value, backendPageInfo)
        }
      }
    }
  } catch (error) {
    console.error('加载客户端列表失败:', error)
  } finally {
    loading.value = false
  }
}

/**
 * 处理分页变化
 */
const handlePageChange = (params: { currentPage: number; pageSize: number }) => {
  if (!pageInfo.value) {
    pageInfo.value = { pageIndex: 1, pageSize: 20 } as PageInfoObj
  }
  if (params.currentPage) {
    pageInfo.value.pageIndex = params.currentPage
  }
  if (params.pageSize) {
    pageInfo.value.pageSize = params.pageSize
  }
  loadClientList()
}

/**
 * 处理行点击事件（选择客户端）
 */
const handleRowClick = ({ row }: { row: TunnelClient }) => {
  if (!row) return
  emit('select', row)
  handleUpdateVisible(false)
}

/**
 * 处理弹窗可见性更新
 */
const handleUpdateVisible = (value: boolean) => {
  modalVisible.value = value
  emit('update:visible', value)
}

watch(
  () => props.visible,
  (newVal) => {
    modalVisible.value = newVal
    if (newVal) {
      loadClientList()
    }
  },
)

onMounted(() => {
  if (props.visible) {
    loadClientList()
  }
})
</script>

<style scoped>
.tunnel-client-list-modal {
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  width: 100%;
  height: min(70vh, 720px);
  min-height: 0;
  overflow: hidden;
}
</style>
