<template>
  <div class="server-node-management" :id="service.model.moduleId">
    <GPane direction="vertical" default-size="300px">
      <!-- 上部：系统节点列表 -->
      <template #1>
        <GCard>
          <GPane direction="vertical" default-size="80px">
            <!-- 搜索表单 -->
            <template #1>
              <search-form
                ref="searchFormRef"
                :module-id="model.moduleId"
                v-bind="model.searchFormConfig"
                @search="handleSearch"
                @toolbar-click="handleToolbarClick"
              />
            </template>

            <!-- 数据表格 -->
            <template #2>
              <g-grid
                ref="gridRef"
                :module-id="model.moduleId"
                :data="model.serverList"
                :loading="model.loading"
                v-bind="model.gridConfig"
                @page-change="service.handlePageChange"
                @menu-click="handleMenuClick"
                @row-click="handleRowClick"
              />
            </template>
          </GPane>
        </GCard>
      </template>

      <!-- 下部：监控面板 -->
      <template #2>
        <GCard>
          <!-- 无选中节点提示 -->
          <div v-if="!monitor.selectedServerId.value" class="monitor-empty">
            <n-empty description="请在上方列表中选择一个节点查看监控数据" />
          </div>

          <!-- 监控图表区域 -->
          <div v-else class="monitor-container">
            <!-- 服务器信息卡片 -->
            <ServerInfoCard v-if="selectedServerInfo" :server-info="selectedServerInfo" />
            <!-- 第一行：CPU、内存使用率趋势 -->
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

            <!-- 第二行：磁盘使用率、磁盘IO监控 -->
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

            <!-- 第三行：网络流量监控、进程监控 -->
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
        </GCard>
      </template>
    </GPane>
  </div>
</template>

<script lang="ts" setup>
import SearchForm from '@/components/form/search/SearchForm.vue'
import { GCard } from '@/components/gcard'
import { GPane } from '@/components/gpane'
import { GGrid } from '@/components/grid'
import { NEmpty } from 'naive-ui'
import { computed, ref } from 'vue'
import {
  CpuMonitor,
  DiskIOMonitor,
  DiskMonitor,
  MemoryMonitor,
  NetworkMonitor,
  ProcessMonitor
} from './components/metrics'
import { ServerInfoCard } from './components/server-info'
import { useServerNodeMonitor, useServerNodePage } from './hooks'
import type { ServerInfo } from './types'

// 定义组件名称
defineOptions({
  name: 'ServerNodeManagement'
})

// ============= Refs =============

const searchFormRef = ref()
const gridRef = ref()

// ============= 页面级 Hook（包含服务与对话框、事件处理） =============

const { model, service, handleToolbarClick, handleMenuClick, handleSearch } = useServerNodePage(
  gridRef,
  searchFormRef
)

// ============= 监控 Hook =============

const monitor = useServerNodeMonitor()

// ============= 计算属性 =============

/**
 * 选中的服务器信息
 */
const selectedServerInfo = computed<ServerInfo | null>(() => {
  if (!monitor.selectedServerId.value) return null
  return (
    model.serverList.value.find(
      (server: ServerInfo) => server.metricServerId === monitor.selectedServerId.value
    ) || null
  )
})

// ============= 事件处理 =============

/**
 * 处理行点击事件 - 选择节点并加载监控数据
 */
const handleRowClick = async ({ row }: { row: any }) => {
  if (row && row.metricServerId) {
    await monitor.setSelectedServer(row.metricServerId)
  }
}

/**
 * 时间范围变化处理
 */
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

// 数据由搜索表单的"查询"按钮触发加载
</script>

<style lang="scss" scoped>
.server-node-management {
  width: 100%;
  height: 100%;
  overflow: hidden;

  :deep(.n-split) {
    height: 100%;
  }

  /* 上半区：节点列表 */
  :deep(.n-split-pane:first-child) {
    overflow: hidden;
    padding: var(--g-space-sm);
  }

  /* 下半区：监控面板 */
  :deep(.n-split-pane:last-child) {
    overflow: hidden;
    padding: var(--g-space-sm);

    .g-card {
      height: 100%;
      overflow: hidden;

      :deep(.n-card__content) {
        height: 100%;
        overflow: auto;
      }
    }
  }

  /* 搜索表单区域 */
  :deep(.n-split-pane:first-child .n-split-pane:first-child) {
    overflow: auto;
    padding: var(--g-space-sm);
  }

  /* 表格区域 */
  :deep(.n-split-pane:first-child .n-split-pane:last-child) {
    overflow: hidden;
    padding: var(--g-space-sm);
    display: flex;
    flex-direction: column;
  }

  /* 监控面板样式 */
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
}


</style>
