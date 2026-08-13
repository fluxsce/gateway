<template>
  <div class="system-monitoring">
    <!-- 服务器选择器 -->
    <div class="server-selector">
      <RsCard :title="t('systemMonitoring.serverSelect')" variant="outlined">
        <div class="selector-content">
          <RsSelect
            v-model="selectedServerId"
            class="server-select"
            :options="serverOptions"
            :placeholder="t('systemMonitoring.selectServer')"
            :loading="serverListLoading"
            clearable
            block
            @update:model-value="handleServerChange"
          />
          <RsButton class="refresh-btn" @click="refreshAllData" :loading="operationLoading">
            <template #icon>
              <GIcon>
                <ReloadOutlined />
              </GIcon>
            </template>
            {{ t('systemMonitoring.refreshData') }}
          </RsButton>
        </div>
      </RsCard>
    </div>

    <!-- 服务器信息概览卡片 -->
    <div class="overview-cards" v-if="selectedServerInfo">
      <RsCard :title="t('systemMonitoring.serverInfo')" variant="outlined">
        <div class="overview-grid">
          <div class="overview-item">
            <div class="overview-icon hostname">
              <GIcon size="24">
                <DatabaseOutlined />
              </GIcon>
            </div>
            <div class="overview-content">
              <div class="overview-label">{{ t('server.hostname') }}</div>
              <RsTooltip :content="selectedServerInfo.hostname">
                <div class="overview-value text-truncate">{{ selectedServerInfo.hostname }}</div>
              </RsTooltip>
            </div>
          </div>

          <div class="overview-item">
            <div class="overview-icon os">
              <GIcon size="24">
                <component :is="getOSIcon(selectedServerInfo.osType)" />
              </GIcon>
            </div>
            <div class="overview-content">
              <div class="overview-label">{{ t('server.osType') }}</div>
              <RsTooltip :content="selectedServerInfo.osType">
                <div class="overview-value text-truncate">{{ selectedServerInfo.osType }}</div>
              </RsTooltip>
            </div>
          </div>

          <div class="overview-item">
            <div class="overview-icon version">
              <GIcon size="24">
                <AndroidOutlined />
              </GIcon>
            </div>
            <div class="overview-content">
              <div class="overview-label">{{ t('systemMonitoring.osVersion') }}</div>
              <RsTooltip :content="selectedServerInfo.osVersion">
                <div class="overview-value text-truncate">
                  {{ getShortVersion(selectedServerInfo.osVersion) }}
                </div>
              </RsTooltip>
            </div>
          </div>

          <div class="overview-item">
            <div class="overview-icon architecture">
              <GIcon size="24">
                <DesktopOutlined />
              </GIcon>
            </div>
            <div class="overview-content">
              <div class="overview-label">{{ t('systemMonitoring.architecture') }}</div>
              <RsTooltip :content="selectedServerInfo.architecture">
                <div class="overview-value text-truncate">
                  {{ selectedServerInfo.architecture }}
                </div>
              </RsTooltip>
            </div>
          </div>

          <div class="overview-item">
            <div class="overview-icon server-type">
              <GIcon size="24">
                <CloudServerOutlined />
              </GIcon>
            </div>
            <div class="overview-content">
              <div class="overview-label">{{ t('server.serverType') }}</div>
              <RsTooltip :content="getServerTypeLabel(selectedServerInfo.serverType)">
                <div class="overview-value text-truncate">
                  {{ getServerTypeLabel(selectedServerInfo.serverType) }}
                </div>
              </RsTooltip>
            </div>
          </div>

          <div class="overview-item">
            <div class="overview-icon ip">
              <GIcon size="24">
                <GlobalOutlined />
              </GIcon>
            </div>
            <div class="overview-content">
              <div class="overview-label">{{ t('server.ipAddress') }}</div>
              <RsTooltip :content="selectedServerInfo.ipAddress || 'N/A'">
                <div class="overview-value text-truncate">
                  {{ selectedServerInfo.ipAddress || 'N/A' }}
                </div>
              </RsTooltip>
            </div>
          </div>
        </div>
      </RsCard>
    </div>

    <!-- 数据加载中状态 -->
    <div v-if="initialDataLoading" class="loading-container">
      <RsLoading size="lg" />
      <p>{{ t('systemMonitoring.loadingMonitor') }}</p>
    </div>

    <!-- 图表展示区域 - 只有在初始数据加载完成后才显示 -->
    <div v-if="!initialDataLoading && hasConcreteServer" class="charts-container">
      <!-- 第一行：CPU、内存使用率趋势 -->
      <div class="chart-row">
        <div class="chart-item">
          <CpuMonitor
            :data="model.cpuMetrics.value"
            :loading="cpuLoading"
            :warning-threshold="80"
            :danger-threshold="90"
            :cpu-detail-data="model.cpuMetrics.value"
            @refresh="refreshCpuData"
            @time-range-change="handleCpuTimeRangeChange"
          />
        </div>

        <div class="chart-item">
          <MemoryMonitor
            :data="model.memoryMetrics.value"
            :loading="memoryLoading"
            :warning-threshold="80"
            :danger-threshold="90"
            :memory-detail-data="model.memoryMetrics.value"
            @refresh="refreshMemoryData"
            @time-range-change="handleMemoryTimeRangeChange"
          />
        </div>
      </div>

      <!-- 第二行：磁盘使用率、磁盘IO监控 -->
      <div class="chart-row">
        <div class="chart-item">
          <DiskMonitor
            :data="model.diskMetrics.value"
            :loading="diskLoading"
            :warning-threshold="80"
            :danger-threshold="90"
            :disk-detail-data="model.diskMetrics.value"
            @refresh="refreshDiskData"
            @time-range-change="handleDiskTimeRangeChange"
          />
        </div>

        <div class="chart-item">
          <DiskIOMonitor
            :data="model.diskIOMetrics.value"
            :loading="diskIOLoading"
            :disk-io-detail-data="model.diskIOMetrics.value"
            @refresh="refreshDiskIOData"
            @time-range-change="handleDiskIOTimeRangeChange"
          />
        </div>
      </div>

      <!-- 第三行：网络流量监控、进程监控 -->
      <div class="chart-row">
        <div class="chart-item">
          <NetworkMonitor
            :data="model.networkMetrics.value"
            :loading="networkLoading"
            :network-detail-data="model.networkMetrics.value"
            upload-color="#ff4d4f"
            download-color="#52c41a"
            @refresh="refreshNetworkData"
            @time-range-change="handleNetworkTimeRangeChange"
          />
        </div>

        <div class="chart-item">
          <ProcessMonitor
            :data="model.processMetrics.value"
            :loading="processLoading"
            :process-detail-data="model.processMetrics.value"
            @refresh="refreshProcessData"
            @time-range-change="handleProcessTimeRangeChange"
          />
        </div>
      </div>
    </div>

    <!-- 无数据提示 -->
    <div
      v-if="!initialDataLoading && !hasConcreteServer && model.serverList.value.length > 0"
      class="no-data-container"
    >
      <RsEmpty :description="t('systemMonitoring.selectServerHint')" />
    </div>

    <!-- 无服务器提示 -->
    <div
      v-if="!initialDataLoading && model.serverList.value.length === 0"
      class="no-data-container"
    >
      <RsEmpty :description="t('systemMonitoring.noServers')" />
    </div>
  </div>
</template>

<script setup lang="ts">
// @ts-nocheck
import GIcon from '@/components/gicon/GIcon.vue'
import { useAppMessage } from '@/composables/useAppMessage'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { RsButton, RsCard, RsEmpty, RsLoading, RsSelect, RsTooltip } from '@/ui'
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'

import { formatDate } from '@/utils/format'
import {
    AndroidOutlined,
    AppleOutlined,
    CloudServerOutlined,
    DatabaseOutlined,
    DesktopOutlined,
    GlobalOutlined,
    ReloadOutlined,
    WindowsOutlined,
} from '@vicons/antd'
import {
    CpuMonitor,
    DiskIOMonitor,
    DiskMonitor,
    MemoryMonitor,
    NetworkMonitor,
    ProcessMonitor,
} from './components/metrics'
import './components/metrics/echartsTooltip.css'
import { useSystemMonitorManagement } from './hooks/useSystemMonitorManagement'
import { useSystemMonitorModel } from './hooks/useSystemMonitorModel'
import type { ServerInfo } from './types'

const message = useAppMessage()
const { t } = useModuleI18n('hub0000')

// 使用系统监控模型和管理
const model = useSystemMonitorModel()
const management = useSystemMonitorManagement(model)

// 组件状态
const selectedServerId = ref('')
const initialDataLoading = ref(true) // 初始数据加载状态

/** 「全部」用非空 value；'' 表示未选中（与 clearable 清空同语义） */
const ALL_SERVERS = 'all'

const selectedServerInfo = computed<ServerInfo | null>(() => {
  if (!selectedServerId.value || selectedServerId.value === ALL_SERVERS) return null
  return (
    model.serverList.value.find((server) => server.metricServerId === selectedServerId.value) ||
    null
  )
})

const hasConcreteServer = computed(
  () => Boolean(selectedServerId.value) && selectedServerId.value !== ALL_SERVERS,
)

// 服务器类型标签转换
const getServerTypeLabel = (serverType?: string): string => {
  const key = serverType === 'physical' || serverType === 'virtual' ? serverType : 'unknown'
  return t(`server.type.${key}`)
}

// 根据操作系统类型获取图标
const getOSIcon = (osType: string) => {
  const osLower = osType.toLowerCase()
  if (osLower.includes('windows')) {
    return WindowsOutlined
  } else if (osLower.includes('linux')) {
    return AndroidOutlined // 使用Android图标代表Linux
  } else if (osLower.includes('mac') || osLower.includes('darwin')) {
    return AppleOutlined
  } else {
    return DesktopOutlined
  }
}

// 获取简化的系统版本信息
const getShortVersion = (version: string): string => {
  if (!version) return 'N/A'

  // 针对Windows系统版本的特殊处理
  if (version.toLowerCase().includes('windows')) {
    // 提取关键信息：Windows 版本号
    const match = version.match(/Windows (\d+(?:\.\d+)?)/i)
    if (match) {
      const windowsVersion = match[1]
      // 如果有额外信息（如 Home, Pro等），也提取出来
      const editionMatch = version.match(/Windows \d+(?:\.\d+)?\s+(\w+)/i)
      if (editionMatch) {
        return `Windows ${windowsVersion} ${editionMatch[1]}`
      }
      return `Windows ${windowsVersion}`
    }
  }

  // 对于其他系统，如果版本信息太长，进行截断
  if (version.length > 20) {
    return version.substring(0, 17) + '...'
  }

  return version
}

const serverOptions = computed(() => [
  { label: t('systemMonitoring.allServers'), value: ALL_SERVERS },
  ...model.serverList.value.map((server) => ({
    label: `${server.hostname} (${server.ipAddress || 'N/A'})`,
    value: server.metricServerId,
  })),
])

// 时间范围变化处理
const handleCpuTimeRangeChange = async (timeRange: [number, number] | null) => {
  if (timeRange) {
    const [startTime, endTime] = timeRange
    model.updateQueryParams({
      startTime: formatDate(startTime, 'YYYY-MM-DD HH:mm:ss'),
      endTime: formatDate(endTime, 'YYYY-MM-DD HH:mm:ss'),
    })
    await management.loadCPUMetrics(selectedServerId.value)
  }
}

const handleMemoryTimeRangeChange = async (timeRange: [number, number] | null) => {
  if (timeRange) {
    const [startTime, endTime] = timeRange
    model.updateQueryParams({
      startTime: formatDate(startTime, 'YYYY-MM-DD HH:mm:ss'),
      endTime: formatDate(endTime, 'YYYY-MM-DD HH:mm:ss'),
    })
    await management.loadMemoryMetrics(selectedServerId.value)
  }
}

const handleDiskTimeRangeChange = async (timeRange: [number, number] | null) => {
  if (timeRange) {
    const [startTime, endTime] = timeRange
    model.updateQueryParams({
      startTime: formatDate(startTime, 'YYYY-MM-DD HH:mm:ss'),
      endTime: formatDate(endTime, 'YYYY-MM-DD HH:mm:ss'),
    })
    await management.loadDiskMetrics(selectedServerId.value)
  }
}

const handleNetworkTimeRangeChange = async (timeRange: [number, number] | null) => {
  if (timeRange) {
    const [startTime, endTime] = timeRange
    model.updateQueryParams({
      startTime: formatDate(startTime, 'YYYY-MM-DD HH:mm:ss'),
      endTime: formatDate(endTime, 'YYYY-MM-DD HH:mm:ss'),
    })
    await management.loadNetworkMetrics(selectedServerId.value)
  }
}

const handleDiskIOTimeRangeChange = async (timeRange: [number, number] | null) => {
  if (timeRange) {
    const [startTime, endTime] = timeRange
    model.updateQueryParams({
      startTime: formatDate(startTime, 'YYYY-MM-DD HH:mm:ss'),
      endTime: formatDate(endTime, 'YYYY-MM-DD HH:mm:ss'),
    })
    await management.loadDiskIOMetrics(selectedServerId.value)
  }
}

const handleProcessTimeRangeChange = async (timeRange: [number, number] | null) => {
  if (timeRange) {
    const [startTime, endTime] = timeRange
    model.updateQueryParams({
      startTime: formatDate(startTime, 'YYYY-MM-DD HH:mm:ss'),
      endTime: formatDate(endTime, 'YYYY-MM-DD HH:mm:ss'),
    })
    await management.loadProcessMetrics(selectedServerId.value)
  }
}

// 监听分页变化
watch(
  () => model.pagination.pagination.page,
  (newPage) => {
    model.queryParams.pageNum = newPage
    loadServerList()
  },
)

watch(
  () => model.pagination.pagination.pageSize,
  (newPageSize) => {
    model.queryParams.pageSize = newPageSize
    loadServerList()
  },
)

// 监听查询参数变化
watch(
  () => ({
    tenantId: model.queryParams.tenantId,
    hostname: model.queryParams.hostname,
    osType: model.queryParams.osType,
    serverType: model.queryParams.serverType,
    activeFlag: model.queryParams.activeFlag,
  }),
  () => {
    if (!initialDataLoading.value) {
      loadServerList()
    }
  },
  { deep: true },
)

// 加载服务器列表
const loadServerList = async () => {
  try {
    const servers = await management.loadServerList()

    // 如果当前没有选中的服务器，且有可用服务器，则选择第一个
    if (!selectedServerId.value && servers.length > 0) {
      selectedServerId.value = servers[0].metricServerId
      await loadServerMonitorData(servers[0].metricServerId)
    }
  } catch {
    message.error(t('systemMonitoring.loadServerListFailed'))
  }
}

// 加载服务器监控数据
const loadServerMonitorData = async (serverId: string) => {
  try {
    // 更新查询参数
    model.updateQueryParams({
      metricServerId: serverId,
      startTime: model.queryParams.startTime,
      endTime: model.queryParams.endTime,
    })

    // 加载监控数据
    await management.loadAllMetrics(serverId)
  } catch {
    message.error(t('systemMonitoring.loadMonitorFailed'))
  }
}

// 初始化数据
const initializeData = async () => {
  try {
    initialDataLoading.value = true

    // 设置默认的时间范围查询参数
    const now = new Date()
    const oneHourAgo = new Date(now.getTime() - 3600 * 1000)
    model.updateQueryParams({
      startTime: formatDate(oneHourAgo, 'YYYY-MM-DD HH:mm:ss'),
      endTime: formatDate(now, 'YYYY-MM-DD HH:mm:ss'),
    })

    // 加载服务器列表
    await loadServerList()
  } catch {
    message.error(t('systemMonitoring.initFailed'))
  } finally {
    initialDataLoading.value = false
  }
}

// 处理服务器选择变更
const handleServerChange = async (serverId: string) => {
  if (!serverId) return

  try {
    operationLoading.value = true
    // 清空旧的图表数据，防止新服务器无数据时显示旧数据
    model.clearAllMetrics()
    await loadServerMonitorData(serverId)
    message.success('服务器监控数据加载成功')
  } catch {
    message.error('加载服务器监控数据失败')
  } finally {
    operationLoading.value = false
  }
}

// 刷新所有数据
const refreshAllData = async () => {
  if (!selectedServerId.value) {
    message.warning('请先选择一个服务器')
    return
  }

  try {
    operationLoading.value = true
    await loadServerMonitorData(selectedServerId.value)
    message.success('所有监控数据刷新成功')
  } catch {
    message.error('刷新监控数据失败')
  } finally {
    operationLoading.value = false
  }
}

const refreshCpuData = async () => {
  await management.loadCPUMetrics(selectedServerId.value)
  message.success('CPU数据刷新成功')
}

const refreshMemoryData = async () => {
  await management.loadMemoryMetrics(selectedServerId.value)
  message.success('内存数据刷新成功')
}

const refreshDiskData = async () => {
  await management.loadDiskMetrics(selectedServerId.value)
  message.success('磁盘数据刷新成功')
}

const refreshNetworkData = async () => {
  await management.loadNetworkMetrics(selectedServerId.value)
  message.success('网络数据刷新成功')
}

const refreshDiskIOData = async () => {
  await management.loadDiskIOMetrics(selectedServerId.value)
  message.success('磁盘IO数据刷新成功')
}

const refreshProcessData = async () => {
  await management.loadProcessMetrics(selectedServerId.value)
  message.success('进程数据刷新成功')
}

// 监听服务器列表变化，自动选择第一个服务器并加载监控数据
watch(
  () => model.serverList.value,
  (newList) => {
    if (newList.length > 0 && !selectedServerId.value) {
      selectedServerId.value = newList[0].metricServerId
    }
  },
)

// 生命周期钩子
onMounted(() => {
  initializeData()
})

onUnmounted(() => {
  // 清理资源已经在 management 中处理
})

// 解构需要的响应式数据 - 从model中解构
const {
  // 加载状态
  serverListLoading,
  cpuLoading,
  memoryLoading,
  diskLoading,
  diskIOLoading,
  networkLoading,
  processLoading,
} = model

// 解构需要的方法和状态 - 从management中解构
const {
  // 操作状态
  operationLoading,
} = management
</script>

<style lang="scss" scoped>
.system-monitoring {
  padding: 16px;

  .server-selector {
    margin-bottom: 16px;

    .selector-content {
      display: grid;
      grid-template-columns: minmax(0, 1fr) auto;
      gap: 12px;
      align-items: center;
      width: 100%;

      :deep(.rs-select) {
        width: 100%;
        max-width: none;
        min-width: 0;
      }

      :deep(.rs-select__anchor),
      :deep(.rs-select__trigger) {
        width: 100%;
      }

      .refresh-btn {
        white-space: nowrap;
      }
    }
  }
}

.overview-cards {
  margin-bottom: 16px;

  .overview-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
    gap: 16px;

    .overview-item {
      display: flex;
      align-items: center;
      gap: 12px;

      .overview-icon {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 40px;
        height: 40px;
        border-radius: 50%;
        color: #fff;

        &.hostname {
          background-color: #1890ff;
        }

        &.os {
          background-color: #52c41a;
        }

        &.version {
          background-color: #fa8c16;
        }

        &.architecture {
          background-color: #722ed1;
        }

        &.server-type {
          background-color: #eb2f96;
        }

        &.ip {
          background-color: #faad14;
        }
      }

      .overview-content {
        flex: 1;

        .overview-label {
          font-size: 12px;
          color: #999;
          margin-bottom: 4px;
        }

        .overview-value {
          font-size: 14px;
          font-weight: 500;
        }

        .text-truncate {
          max-width: 180px;
          white-space: nowrap;
          overflow: hidden;
          text-overflow: ellipsis;
        }
      }
    }
  }
}

.charts-container {
  .chart-row {
    display: flex;
    gap: 16px;
    margin-bottom: 16px;

    @media (max-width: 1200px) {
      flex-direction: column;
    }

    .chart-item {
      flex: 1;
      min-height: 360px;
    }
  }
}

.loading-container,
.no-data-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 400px;

  p {
    margin-top: 16px;
    color: #666;
  }
}
</style>
