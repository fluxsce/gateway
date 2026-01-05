<template>
  <div class="regist-service-list">
    <g-grid
      ref="gridRef"
      module-id="hub0060:regist-service-list"
      :data="serviceList"
      :loading="loading"
      v-bind="gridConfig"
      @menu-click="handleMenuClick"
    >
      <!-- 服务名称自定义渲染 -->
      <template #serviceName="{ row }">
        <span class="font-bold">{{ row.serviceName || '-' }}</span>
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
import { NIcon } from 'naive-ui'
import { h, onMounted, ref, watch } from 'vue'
import { getRegisteredServices } from '../../api'

// 服务信息类型（从服务器获取的运行时信息）
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
  tunnelServerId: ''
})

const gridRef = ref()
const loading = ref(false)
const serviceList = ref<TunnelService[]>([])

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
      field: 'tunnelServiceId',
      title: '服务ID',
      width: 200,
      showOverflow: 'tooltip',
      filters: [{ data: '' }],
      filterRender: { name: 'VxeInput' }
    },
    {
      field: 'serviceName',
      title: '服务名称',
      width: 180,
      showOverflow: 'tooltip',
      slots: { default: 'serviceName' },
      filters: [{ data: '' }],
      filterRender: { name: 'VxeInput' }
    },
    {
      field: 'serviceType',
      title: '服务类型',
      width: 100,
      align: 'center',
      showOverflow: 'tooltip'
    },
    {
      field: 'localAddress',
      title: '本地地址',
      width: 200,
      showOverflow: 'tooltip',
      formatter: ({ row }) => `${row.localAddress}:${row.localPort}`
    },
    {
      field: 'remotePort',
      title: '远程端口',
      width: 100,
      align: 'center',
      showOverflow: 'tooltip',
      formatter: ({ cellValue }) => cellValue || '-'
    },
    {
      field: 'serviceStatus',
      title: '服务状态',
      width: 100,
      align: 'center',
      formatter: ({ cellValue }) => {
        if (!cellValue) return '-'
        const statusMap: Record<string, string> = {
          active: '活跃',
          inactive: '未活跃',
          error: '错误'
        }
        return statusMap[cellValue] || cellValue
      }
    },
    {
      field: 'connectionCount',
      title: '连接数',
      width: 100,
      align: 'center',
      formatter: ({ cellValue }) => cellValue || 0
    },
    {
      field: 'totalConnections',
      title: '总连接数',
      width: 120,
      align: 'center',
      formatter: ({ cellValue }) => cellValue || 0
    },
    {
      field: 'registeredTime',
      title: '注册时间',
      width: 160,
      showOverflow: true,
      formatter: ({ cellValue }) => (cellValue ? formatDate(cellValue, 'YYYY-MM-DD HH:mm:ss') : '-')
    },
    {
      field: 'lastActiveTime',
      title: '最后活动时间',
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

// 加载服务列表
const loadServiceList = async () => {
  loading.value = true
  try {
    // 如果 tunnelServerId 为空，传递空字符串以获取所有服务器的服务列表
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
    console.error('加载已注册服务列表失败:', error)
    serviceList.value = []
  } finally {
    loading.value = false
  }
}

// 处理右键菜单点击
const handleMenuClick = async ({ code }: { code: string }) => {
  if (code === 'refresh') {
    await loadServiceList()
  }
}

onMounted(() => {
  loadServiceList()
})

// 监听 tunnelServerId 变化
watch(
  () => props.tunnelServerId,
  () => {
    loadServiceList()
  }
)

defineExpose({
  loadServiceList,
  refresh: loadServiceList
})
</script>

<style scoped lang="scss">
.regist-service-list {
  width: 100%;
  height: 100%;
  overflow: hidden;
  background-color: var(--n-color-target);
}
</style>

