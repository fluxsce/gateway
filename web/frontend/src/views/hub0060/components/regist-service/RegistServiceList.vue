<template>
  <div class="regist-service-list">
    <RsGrid
      ref="gridRef"
      module-id="hub0060:regist-service-list"
      :data="serviceList"
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
import { formatDate } from '@/utils/format'
import { h, onMounted, ref, watch } from 'vue'
import { getRegisteredServices } from '../../api'

defineOptions({
  name: 'RegistServiceList',
})

/**
 * 服务运行时信息（从服务器获取）。
 */
interface TunnelService {
  tunnelServiceId: string
  tunnelClientId: string
  serviceName: string
  serviceType: string
  localAddress: string
  localPort: number
  remotePort?: number
  serviceStatus?: string
  connectionCount?: number
  totalConnections?: number
  registeredTime?: string
  lastActiveTime?: string
  tunnelServerId?: string
  serviceDescription?: string
  customDomains?: string
  subDomain?: string
}

interface Props {
  tunnelServerId?: string
}

const props = withDefaults(defineProps<Props>(), {
  tunnelServerId: '',
})

const gridRef = ref<RsGridExpose | null>(null)
const loading = ref(false)
const serviceList = ref<TunnelService[]>([])

const gridConfig: {
  columns: RsGridColumn<TunnelService>[]
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
      key: 'tunnelServiceId',
      title: '服务ID',
      width: 200,
      ellipsis: true,
      filterable: true,
    },
    {
      key: 'serviceName',
      title: '服务名称',
      width: 180,
      ellipsis: true,
      filterable: true,
      render: (row) => h('span', { class: 'font-bold' }, row.serviceName || '-'),
    },
    {
      key: 'serviceType',
      title: '服务类型',
      width: 100,
      align: 'center',
      ellipsis: true,
    },
    {
      key: 'localAddress',
      title: '本地地址',
      width: 200,
      ellipsis: true,
      formatter: (_value, row) => `${row.localAddress}:${row.localPort}`,
    },
    {
      key: 'remotePort',
      title: '远程端口',
      width: 100,
      align: 'center',
      ellipsis: true,
      formatter: (value) => (value ? String(value) : '-'),
    },
    {
      key: 'serviceStatus',
      title: '服务状态',
      width: 100,
      align: 'center',
      formatter: (value) => {
        if (!value) return '-'
        const statusMap: Record<string, string> = {
          active: '活跃',
          inactive: '未活跃',
          error: '错误',
        }
        return statusMap[String(value)] || String(value)
      },
    },
    {
      key: 'connectionCount',
      title: '连接数',
      width: 100,
      align: 'center',
      formatter: (value) => String(value || 0),
    },
    {
      key: 'totalConnections',
      title: '总连接数',
      width: 120,
      align: 'center',
      formatter: (value) => String(value || 0),
    },
    {
      key: 'registeredTime',
      title: '注册时间',
      width: 160,
      ellipsis: true,
      formatter: (value) => (value ? formatDate(value as string, 'YYYY-MM-DD HH:mm:ss') : '-'),
    },
    {
      key: 'lastActiveTime',
      title: '最后活动时间',
      width: 160,
      ellipsis: true,
      formatter: (value) => (value ? formatDate(value as string, 'YYYY-MM-DD HH:mm:ss') : '-'),
    },
  ],
  selectable: false,
  rowKey: 'tunnelServiceId',
  menuConfig: {
    enabled: true,
    items: [{ key: 'refresh', label: '刷新', icon: 'refresh-cw', requireRow: false }],
  },
}

const loadServiceList = async () => {
  loading.value = true
  try {
    const serverId = props.tunnelServerId || ''
    const response: JsonDataObj = await getRegisteredServices(serverId)
    if (response.oK) {
      if (response.bizData) {
        const bizData = JSON.parse(response.bizData)
        serviceList.value = Array.isArray(bizData) ? bizData : []
      } else {
        serviceList.value = []
      }
    } else {
      serviceList.value = []
    }
  } catch (error) {
    serviceList.value = []
  } finally {
    loading.value = false
  }
}

const handleMenuClick = async ({ key }: { key: string }) => {
  if (key === 'refresh') {
    await loadServiceList()
  }
}

onMounted(() => {
  loadServiceList()
})

watch(
  () => props.tunnelServerId,
  () => {
    loadServiceList()
  },
)

defineExpose({
  loadServiceList,
  refresh: loadServiceList,
})
</script>

<style scoped lang="scss">
.regist-service-list {
  width: 100%;
  height: 100%;
  overflow: hidden;
}
</style>
