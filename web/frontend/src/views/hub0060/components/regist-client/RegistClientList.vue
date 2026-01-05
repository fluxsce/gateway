<template>
  <div class="regist-client-list">
    <g-grid
      ref="gridRef"
      module-id="hub0060:regist-client-list"
      :data="clientList"
      :loading="loading"
      v-bind="gridConfig"
      @menu-click="handleMenuClick"
    >
      <!-- 客户端名称自定义渲染 -->
      <template #clientName="{ row }">
        <span class="font-bold">{{ row.clientName || '-' }}</span>
      </template>

      <!-- 服务数量自定义渲染 -->
      <template #serviceCount="{ row }">
        <n-tag type="info" size="small">
          {{ row.serviceCount || 0 }}
        </n-tag>
      </template>
    </g-grid>
  </div>
</template>

<script setup lang="ts">
import type { GridProps } from '@/components/grid'
import { GGrid } from '@/components/grid'
import type { JsonDataObj } from '@/types/api'
import { formatDate } from '@/utils/format'
import { ReloadOutline } from '@vicons/ionicons5'
import { NIcon, NTag } from 'naive-ui'
import { h, onMounted, ref, watch } from 'vue'
import { getRegisteredClients } from '../../api'

// 客户端信息类型（从服务器获取的运行时信息）
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
  tunnelServerId: ''
})

const gridRef = ref()
const loading = ref(false)
const clientList = ref<TunnelClient[]>([])

const gridConfig: Omit<GridProps, 'moduleId' | 'data' | 'loading'> = {
  columns: [
    {
      field: 'tunnelServerId',
      title: '服务器ID',
      width: 180,
      showOverflow: 'tooltip',
      filters: [{ data: '' }],
      filterRender: { name: 'VxeInput' }
    },
    {
      field: 'tunnelClientId',
      title: '客户端ID',
      width: 200,
      showOverflow: 'tooltip',
      filters: [{ data: '' }],
      filterRender: { name: 'VxeInput' }
    },
    {
      field: 'clientName',
      title: '客户端名称',
      width: 180,
      showOverflow: 'tooltip',
      slots: { default: 'clientName' },
      filters: [{ data: '' }],
      filterRender: { name: 'VxeInput' }
    },
    {
      field: 'serverAddress',
      title: '服务器地址',
      width: 200,
      showOverflow: 'tooltip',
      formatter: ({ row }) => `${row.serverAddress}:${row.serverPort}`,
      filters: [{ data: '' }],
      filterRender: { name: 'VxeInput' }
    },
    {
      field: 'clientIpAddress',
      title: '客户端IP',
      width: 150,
      showOverflow: 'tooltip',
      filters: [{ data: '' }],
      filterRender: { name: 'VxeInput' }
    },
    {
      field: 'clientMacAddress',
      title: '客户端MAC',
      width: 150,
      showOverflow: 'tooltip'
    },
    {
      field: 'clientVersion',
      title: '客户端版本',
      width: 120,
      showOverflow: 'tooltip'
    },
    {
      field: 'operatingSystem',
      title: '操作系统',
      width: 150,
      showOverflow: 'tooltip'
    },
    {
      field: 'connectionStatus',
      title: '连接状态',
      width: 100,
      align: 'center',
      formatter: ({ cellValue }) => {
        if (!cellValue) return '-'
        const statusMap: Record<string, string> = {
          connected: '已连接',
          disconnected: '已断开',
          connecting: '连接中',
          error: '错误'
        }
        return statusMap[cellValue] || cellValue
      }
    },
    {
      field: 'serviceCount',
      title: '服务数量',
      width: 100,
      align: 'center',
      slots: { default: 'serviceCount' }
    },
    {
      field: 'authenticated',
      title: '认证状态',
      width: 100,
      align: 'center',
      formatter: ({ cellValue }) => (cellValue ? '已认证' : '未认证')
    },
    {
      field: 'lastHeartbeat',
      title: '最后心跳',
      width: 160,
      showOverflow: true,
      formatter: ({ cellValue }) => (cellValue ? formatDate(cellValue, 'YYYY-MM-DD HH:mm:ss') : '-')
    }
  ],
  showCheckbox: false,
  menuConfig: {
    enabled: true,
    showCopyRow: true,
    showCopyCell: true,
    customMenus: [
      {
        code: 'refresh',
        name: '刷新',
        prefixIcon: () => h(NIcon, { size: 14 }, { default: () => h(ReloadOutline) })
      }
    ]
  },
  height: '100%'
}

// 加载客户端列表
const loadClientList = async () => {
  loading.value = true
  try {
    // 如果 tunnelServerId 为空，传递空字符串以获取所有服务器的客户端列表
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

// 处理右键菜单点击
const handleMenuClick = async ({ code }: { code: string }) => {
  if (code === 'refresh') {
    await loadClientList()
  }
}

onMounted(() => {
  loadClientList()
})

// 监听 tunnelServerId 变化
watch(
  () => props.tunnelServerId,
  () => {
    loadClientList()
  }
)

defineExpose({
  loadClientList,
  refresh: loadClientList
})
</script>

<style scoped lang="scss">
.regist-client-list {
  width: 100%;
  height: 100%;
  overflow: hidden;
  background-color: var(--n-color-target);
}
</style>
