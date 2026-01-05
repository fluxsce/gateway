<template>
  <GModal
    :visible="modalVisible"
    title="选择客户端"
    :width="900"
    :to="to"
    :show-footer="false"
    @update:visible="handleUpdateVisible"
  >
    <g-grid
      ref="gridRef"
      module-id="tunnel-client-selector-grid"
      :data="clientList"
      :loading="loading"
      v-bind="gridConfig"
      @page-change="handlePageChange"
      @row-click="handleRowClick"
    >
      <!-- 连接状态自定义渲染 -->
      <template #connectionStatus="{ row }">
        <n-tag
          :type="getConnectionStatusTagType(row.connectionStatus)"
          size="small"
        >
          {{ getConnectionStatusLabel(row.connectionStatus) }}
        </n-tag>
      </template>

      <!-- 状态自定义渲染 -->
      <template #activeFlag="{ row }">
        <n-tag :type="row.activeFlag === 'Y' ? 'success' : 'default'" size="small">
          {{ row.activeFlag === 'Y' ? '启用' : '禁用' }}
        </n-tag>
      </template>
    </g-grid>
  </GModal>
</template>

<script setup lang="ts">
import { GModal } from '@/components/gmodal'
import { createBackendPaginationParams } from '@/components/gpage'
import type { GridProps } from '@/components/grid'
import { GGrid } from '@/components/grid'
import type { PageInfoObj } from '@/types/api'
import { NTag } from 'naive-ui'
import { onMounted, ref, watch } from 'vue'
import * as tunnelClientApi from '../../api'
import type { ConnectionStatus, TunnelClient } from '../../types'

interface Props {
  visible?: boolean
  to?: string
}

const props = withDefaults(defineProps<Props>(), {
  visible: false,
  to: 'body'
})

const emit = defineEmits<{
  'update:visible': [value: boolean]
  'select': [client: TunnelClient]
}>()

const gridRef = ref()
const loading = ref(false)
const clientList = ref<TunnelClient[]>([])
const pageInfo = ref<PageInfoObj | undefined>()
const modalVisible = ref(props.visible)

const gridConfig: Omit<GridProps, 'moduleId' | 'data' | 'loading'> = {
  columns: [
    {
      field: 'tunnelClientId',
      title: '客户端ID',
      width: 200,
      showOverflow: 'tooltip'
    },
    {
      field: 'clientName',
      title: '客户端名称',
      width: 180,
      showOverflow: 'tooltip'
    },
    {
      field: 'serverAddress',
      title: '服务器地址',
      width: 200,
      showOverflow: 'tooltip',
      formatter: ({ row }) => `${row.serverAddress}:${row.serverPort}`
    },
    {
      field: 'connectionStatus',
      title: '连接状态',
      width: 100,
      align: 'center',
      slots: { default: 'connectionStatus' }
    },
    {
      field: 'activeFlag',
      title: '状态',
      width: 80,
      align: 'center',
      slots: { default: 'activeFlag' }
    }
  ],
  showCheckbox: false,
  paginationConfig: {
    show: true,
    pageInfo: pageInfo as any,
    align: 'right'
  }
}

const loadClientList = async () => {
  loading.value = true
  try {
    const paginationParams = createBackendPaginationParams(
      pageInfo.value?.pageIndex,
      pageInfo.value?.pageSize
    )

    const params = {
      activeFlag: 'Y' as const, // 只查询激活的客户端
      pageIndex: paginationParams.pageIndex,
      pageSize: paginationParams.pageSize
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

const handleRowClick = ({ row }: { row: TunnelClient }) => {
  if (row) {
    emit('select', row)
    // 选择后关闭弹窗
    handleUpdateVisible(false)
  }
}

const handleUpdateVisible = (value: boolean) => {
  modalVisible.value = value
  emit('update:visible', value)
}

const getConnectionStatusLabel = (status?: ConnectionStatus): string => {
  const statusMap: Record<string, string> = {
    connected: '已连接',
    disconnected: '已断开',
    connecting: '连接中',
    error: '错误'
  }
  return statusMap[status || ''] || status || '-'
}

const getConnectionStatusTagType = (status?: ConnectionStatus): 'success' | 'warning' | 'info' | 'error' | 'default' => {
  const typeMap: Record<string, 'success' | 'warning' | 'info' | 'error' | 'default'> = {
    connected: 'success',
    disconnected: 'warning',
    connecting: 'info',
    error: 'error'
  }
  return typeMap[status || ''] || 'default'
}

watch(() => props.visible, (newVal) => {
  modalVisible.value = newVal
  if (newVal) {
    loadClientList()
  }
})

onMounted(() => {
  if (props.visible) {
    loadClientList()
  }
})
</script>
