<template>
  <div class="server-node-management" :id="service.model.moduleId">
    <RsSplitPane
      class="server-node-management__main"
      orientation="vertical"
      :panes="mainPanes"
      with-handle
    >
      <!-- 上部：节点列表（搜索 + 表格） -->
      <template #list>
        <div class="server-node-management__list">
          <RsSplitPane
            class="server-node-management__list-split"
            orientation="vertical"
            :panes="listPanes"
            disabled
          >
            <template #search>
              <div class="server-node-management__search">
                <RsSearchForm
                  ref="searchFormRef"
                  :module-id="model.moduleId"
                  v-bind="model.searchFormConfig"
                  @search="handleSearch"
                  @toolbar-click="handleToolbarClick"
                />
              </div>
            </template>

            <template #grid>
              <div class="server-node-management__grid">
                <GGrid
                  ref="gridRef"
                  :module-id="model.moduleId"
                  :data="model.serverList"
                  :loading="model.loading"
                  v-bind="model.gridConfig"
                  @page-change="service.handlePageChange"
                  @menu-click="handleMenuClick"
                  @row-click="handleRowClick"
                />
              </div>
            </template>
          </RsSplitPane>
        </div>
      </template>

      <!-- 下部：监控面板（无外层边框，避免与内层监控卡片叠边） -->
      <template #monitor>
        <div class="server-node-management__monitor">
          <div v-if="!monitor.selectedServerId.value" class="monitor-empty">
            <RsEmpty :description="t('selectNodeHint')" />
          </div>

          <div v-else class="monitor-container">
            <ServerInfoCard v-if="selectedServerInfo" :server-info="selectedServerInfo" />

            <div class="chart-row">
              <div class="chart-item">
                <CpuMonitor
                  :data="monitor.cpuMetrics.value"
                  :loading="monitor.cpuLoading.value"
                  :warning-threshold="80"
                  :danger-threshold="90"
                  :cpu-detail-data="monitor.cpuMetrics.value"
                  @refresh="monitor.loadCPUMetrics"
                  @time-range-change="handleCpuTimeRangeChange"
                />
              </div>
              <div class="chart-item">
                <MemoryMonitor
                  :data="monitor.memoryMetrics.value"
                  :loading="monitor.memoryLoading.value"
                  :warning-threshold="80"
                  :danger-threshold="90"
                  :memory-detail-data="monitor.memoryMetrics.value"
                  @refresh="monitor.loadMemoryMetrics"
                  @time-range-change="handleMemoryTimeRangeChange"
                />
              </div>
            </div>

            <div class="chart-row">
              <div class="chart-item">
                <DiskMonitor
                  :data="monitor.diskMetrics.value"
                  :loading="monitor.diskLoading.value"
                  :warning-threshold="80"
                  :danger-threshold="90"
                  :disk-detail-data="monitor.diskMetrics.value"
                  @refresh="monitor.loadDiskMetrics"
                  @time-range-change="handleDiskTimeRangeChange"
                />
              </div>
              <div class="chart-item">
                <DiskIOMonitor
                  :data="monitor.diskIOMetrics.value"
                  :loading="monitor.diskIOLoading.value"
                  :disk-io-detail-data="monitor.diskIOMetrics.value"
                  @refresh="monitor.loadDiskIOMetrics"
                  @time-range-change="handleDiskIOTimeRangeChange"
                />
              </div>
            </div>

            <div class="chart-row">
              <div class="chart-item">
                <NetworkMonitor
                  :data="monitor.networkMetrics.value"
                  :loading="monitor.networkLoading.value"
                  :network-detail-data="monitor.networkMetrics.value"
                  upload-color="#ff4d4f"
                  download-color="#52c41a"
                  @refresh="monitor.loadNetworkMetrics"
                  @time-range-change="handleNetworkTimeRangeChange"
                />
              </div>
              <div class="chart-item">
                <ProcessMonitor
                  :data="monitor.processMetrics.value"
                  :loading="monitor.processLoading.value"
                  :process-detail-data="monitor.processMetrics.value"
                  @refresh="monitor.loadProcessMetrics"
                  @time-range-change="handleProcessTimeRangeChange"
                />
              </div>
            </div>
          </div>
        </div>
      </template>
    </RsSplitPane>
  </div>
</template>

<script lang="ts" setup>
import { RsSearchForm } from '@/components/form/rs-search'
import { GGrid } from '@/components/grid'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { RsEmpty, RsSplitPane, type RsSplitPaneItem } from '@/ui'
import '@/views/hub0000/components/metrics/echartsTooltip.css'
import { computed, ref } from 'vue'
import {
  CpuMonitor,
  DiskIOMonitor,
  DiskMonitor,
  MemoryMonitor,
  NetworkMonitor,
  ProcessMonitor,
} from './components/metrics'
import { ServerInfoCard } from './components/server-info'
import { useServerNodeMonitor, useServerNodePage } from './hooks'
import type { ServerInfo } from './types'

defineOptions({
  name: 'ServerNodeManagement',
})

const { t } = useModuleI18n('hub0007')

/** 外层：节点列表 / 监控面板 */
const mainPanes: RsSplitPaneItem[] = [
  { key: 'list', size: 40, min: 20, max: 70 },
  { key: 'monitor', min: 25 },
]

/** 列表区内层：搜索自适应，表格占满剩余 */
const listPanes: RsSplitPaneItem[] = [
  { key: 'search', size: 'auto' },
  { key: 'grid' },
]

const searchFormRef = ref()
const gridRef = ref()

const { model, service, handleToolbarClick, handleMenuClick, handleSearch } = useServerNodePage(
  gridRef,
  searchFormRef,
)

const monitor = useServerNodeMonitor()

const selectedServerInfo = computed<ServerInfo | null>(() => {
  if (!monitor.selectedServerId.value) return null
  return (
    model.serverList.value.find(
      (server: ServerInfo) => server.metricServerId === monitor.selectedServerId.value,
    ) || null
  )
})

const handleRowClick = async ({ row }: { row: any }) => {
  if (row && row.metricServerId) {
    await monitor.setSelectedServer(row.metricServerId)
  }
}

const handleCpuTimeRangeChange = async (timeRange: [number, number] | null) => {
  if (timeRange) {
    monitor.updateTimeRange(timeRange)
    await monitor.loadCPUMetrics()
  }
}

const handleMemoryTimeRangeChange = async (timeRange: [number, number] | null) => {
  if (timeRange) {
    monitor.updateTimeRange(timeRange)
    await monitor.loadMemoryMetrics()
  }
}

const handleDiskTimeRangeChange = async (timeRange: [number, number] | null) => {
  if (timeRange) {
    monitor.updateTimeRange(timeRange)
    await monitor.loadDiskMetrics()
  }
}

const handleDiskIOTimeRangeChange = async (timeRange: [number, number] | null) => {
  if (timeRange) {
    monitor.updateTimeRange(timeRange)
    await monitor.loadDiskIOMetrics()
  }
}

const handleNetworkTimeRangeChange = async (timeRange: [number, number] | null) => {
  if (timeRange) {
    monitor.updateTimeRange(timeRange)
    await monitor.loadNetworkMetrics()
  }
}

const handleProcessTimeRangeChange = async (timeRange: [number, number] | null) => {
  if (timeRange) {
    monitor.updateTimeRange(timeRange)
    await monitor.loadProcessMetrics()
  }
}
</script>

<style scoped>
.server-node-management {
  box-sizing: border-box;
  position: relative;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.server-node-management__main {
  flex: 1 1 auto;
  width: 100%;
  height: 100%;
  min-height: 0;
}

.server-node-management__list {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
}

.server-node-management__list-split {
  width: 100%;
  height: 100%;
  min-height: 0;
}

.server-node-management__search {
  width: 100%;
}

.server-node-management__grid {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.server-node-management__monitor {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: auto;
}

.monitor-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  min-height: 300px;
}

.monitor-container {
  overflow-y: auto;
  overflow-x: visible;
  height: 100%;
}

.chart-row {
  display: flex;
  gap: 16px;
  margin-bottom: 16px;
}

.chart-item {
  flex: 1;
  min-height: 360px;
}

@media (max-width: 1200px) {
  .chart-row {
    flex-direction: column;
  }
}
</style>
