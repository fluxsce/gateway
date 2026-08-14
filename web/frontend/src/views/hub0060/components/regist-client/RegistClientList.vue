<template>
  <div class="regist-client-list">
    <RsGrid
      ref="gridRef"
      module-id="hub0060:regist-client-list"
      :data="clientList"
      :loading="loading"
      :columns="gridConfig.columns"
      :selectable="gridConfig.selectable"
      :row-key="gridConfig.rowKey"
      height="100%"
      :menu-config="gridConfig.menuConfig"
      @menu-click="handleMenuClick"
    />
  </div>
</template>

<script setup lang="ts">
import { RsGrid, type RsGridColumn, type RsGridExpose, type RsGridMenuConfig } from '@/components/rs-grid'
import type { JsonDataObj } from '@/types/api'
import { RsTag } from '@/ui'
import { formatDate } from '@/utils/format'
import { h, onMounted, ref, watch } from 'vue'
import { getRegisteredClients } from '../../api'

defineOptions({
  name: 'RegistClientList',
})

/**
 * 客户端运行时信息（从服务器获取）。
 */
interface TunnelClient {
  tunnelClientId: string
  clientName: string
  serverAddress: string
  serverPort: number
  clientIpAddress?: string
  serviceCount?: number
  lastHeartbeat?: string
  services?: Record<string, any>
  tunnelServerId?: string
  authenticated?: boolean
  connectionStatus?: string
  clientVersion?: string
  operatingSystem?: string
  clientMacAddress?: string
  tlsEnable?: string
  autoReconnect?: string
}

interface Props {
  tunnelServerId?: string
}

const props = withDefaults(defineProps<Props>(), {
  tunnelServerId: '',
})

const gridRef = ref<RsGridExpose | null>(null)
const loading = ref(false)
const clientList = ref<TunnelClient[]>([])

const gridConfig: {
  columns: RsGridColumn<TunnelClient>[]
  selectable: boolean
  rowKey: string
  menuConfig: RsGridMenuConfig
} = {
  columns: [
    {
      key: 'tunnelServerId',
      title: '服务器ID',
      width: 180,
      ellipsis: true,
      filterable: true,
    },
    {
      key: 'tunnelClientId',
      title: '客户端ID',
      width: 200,
      ellipsis: true,
      filterable: true,
    },
    {
      key: 'clientName',
      title: '客户端名称',
      width: 180,
      ellipsis: true,
      filterable: true,
      render: (row) => h('span', { class: 'font-bold' }, row.clientName || '-'),
    },
    {
      key: 'serverAddress',
      title: '服务器地址',
      width: 200,
      ellipsis: true,
      filterable: true,
      formatter: (_value, row) => `${row.serverAddress}:${row.serverPort}`,
    },
    {
      key: 'clientIpAddress',
      title: '客户端IP',
      width: 150,
      ellipsis: true,
      filterable: true,
    },
    {
      key: 'clientMacAddress',
      title: '客户端MAC',
      width: 150,
      ellipsis: true,
    },
    {
      key: 'clientVersion',
      title: '客户端版本',
      width: 120,
      ellipsis: true,
    },
    {
      key: 'operatingSystem',
      title: '操作系统',
      width: 150,
      ellipsis: true,
    },
    {
      key: 'connectionStatus',
      title: '连接状态',
      width: 100,
      align: 'center',
      formatter: (value) => {
        if (!value) return '-'
        const statusMap: Record<string, string> = {
          connected: '已连接',
          disconnected: '已断开',
          connecting: '连接中',
          error: '错误',
        }
        return statusMap[String(value)] || String(value)
      },
    },
    {
      key: 'serviceCount',
      title: '服务数量',
      width: 100,
      align: 'center',
      render: (row) =>
        h(RsTag, { variant: 'info', size: 'sm' }, () => String(row.serviceCount || 0)),
    },
    {
      key: 'authenticated',
      title: '认证状态',
      width: 100,
      align: 'center',
      formatter: (value) => (value ? '已认证' : '未认证'),
    },
    {
      key: 'lastHeartbeat',
      title: '最后心跳',
      width: 160,
      ellipsis: true,
      formatter: (value) => (value ? formatDate(value as string, 'YYYY-MM-DD HH:mm:ss') : '-'),
    },
  ],
  selectable: false,
  rowKey: 'tunnelClientId',
  menuConfig: {
    enabled: true,
    items: [{ key: 'refresh', label: '刷新', icon: 'refresh-cw', requireRow: false }],
  },
}

const loadClientList = async () => {
  loading.value = true
  try {
    const serverId = props.tunnelServerId || ''
    const response: JsonDataObj = await getRegisteredClients(serverId)
    if (response.oK) {
      if (response.bizData) {
        const bizData = JSON.parse(response.bizData)
        clientList.value = Array.isArray(bizData) ? bizData : []
      } else {
        clientList.value = []
      }
    } else {
      clientList.value = []
    }
  } catch (error) {
    console.error('加载已注册客户端列表失败:', error)
    clientList.value = []
  } finally {
    loading.value = false
  }
}

const handleMenuClick = async ({ key }: { key: string }) => {
  if (key === 'refresh') {
    await loadClientList()
  }
}

onMounted(() => {
  loadClientList()
})

watch(
  () => props.tunnelServerId,
  () => {
    loadClientList()
  },
)

defineExpose({
  loadClientList,
  refresh: loadClientList,
})
</script>

<style scoped lang="scss">
.regist-client-list {
  width: 100%;
  height: 100%;
  overflow: hidden;
}
</style>
