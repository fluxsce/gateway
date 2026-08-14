<template>
  <div class="monitoring-panel">
    <RsSplitPane
      class="monitoring-panel__split"
      orientation="vertical"
      :panes="splitPanes"
      disabled
    >
      <template #search>
        <div class="monitoring-panel__search">
          <RsSearchForm
            ref="searchFormRef"
            :module-id="page.service.model.moduleId"
            v-bind="page.service.model.searchFormConfig"
            @search="handleSearch"
            @reset="handleReset"
            @toolbar-click="handleToolbarClick"
          />
        </div>
      </template>

      <template #content>
        <div class="monitoring-panel__content">
          <RsCard class="overview-card" size="sm" variant="outlined">
            <template #header>
              <div class="overview-header">
                <RsIcon name="chart-column" :size="18" color="#18a058" />
                <span>监控概览</span>
              </div>
            </template>
            <div class="overview-content">
              <div class="overview-grid">
                <RsStatCard
                  label="总请求数"
                  :value="page.service.model.overviewData.totalRequests"
                  accent="success"
                />
                <RsStatCard
                  label="成功请求数"
                  :value="page.service.model.overviewData.successRequests"
                  accent="success"
                />
                <RsStatCard
                  label="失败请求数"
                  :value="page.service.model.overviewData.failedRequests"
                  accent="danger"
                />
              </div>

              <RsDivider />

              <div class="overview-grid">
                <RsStatCard
                  label="平均响应时间"
                  :value="`${page.service.model.overviewData.avgResponseTimeMs} ms`"
                  accent="info"
                />
                <RsStatCard
                  label="最小响应时间"
                  :value="`${page.service.model.overviewData.minResponseTimeMs} ms`"
                  accent="success"
                />
                <RsStatCard
                  label="最大响应时间"
                  :value="`${page.service.model.overviewData.maxResponseTimeMs} ms`"
                  accent="danger"
                />
              </div>
            </div>
          </RsCard>

          <div class="chart-grid">
            <RsCard class="chart-card" size="sm" variant="outlined">
              <template #header>
                <div class="chart-header">
                  <RsIcon name="trending-up" :size="18" color="#18a058" />
                  <span>请求量趋势({{ page.service.model.getTimeGranularityLabel() }})</span>
                </div>
              </template>
              <div class="chart-container" ref="requestTrendChartRef">
                <div v-if="page.service.model.loading.value" class="chart-loading">
                  <RsLoading size="lg" />
                </div>
                <div
                  v-else-if="!page.service.model.chartData.requestTrend.length"
                  class="chart-empty"
                >
                  <RsEmpty description="暂无数据" />
                </div>
              </div>
            </RsCard>

            <RsCard class="chart-card" size="sm" variant="outlined">
              <template #header>
                <div class="chart-header">
                  <RsIcon name="clock" :size="18" color="#2080f0" />
                  <span>响应时间趋势({{ page.service.model.getTimeGranularityLabel() }})</span>
                </div>
              </template>
              <div class="chart-container" ref="responseTimeChartRef">
                <div v-if="page.service.model.loading.value" class="chart-loading">
                  <RsLoading size="lg" />
                </div>
                <div
                  v-else-if="!page.service.model.chartData.responseTimeTrend.length"
                  class="chart-empty"
                >
                  <RsEmpty description="暂无数据" />
                </div>
              </div>
            </RsCard>

            <RsCard class="chart-card" size="sm" variant="outlined">
              <template #header>
                <div class="chart-header">
                  <RsIcon name="chart-pie" :size="18" color="#f0a020" />
                  <span>请求指标分布</span>
                </div>
              </template>
              <div class="chart-container" ref="requestMetricsChartRef">
                <div v-if="page.service.model.loading.value" class="chart-loading">
                  <RsLoading size="lg" />
                </div>
                <div
                  v-else-if="page.service.model.overviewData.totalRequests === 0"
                  class="chart-empty"
                >
                  <RsEmpty description="暂无数据" />
                </div>
              </div>
            </RsCard>

            <RsCard class="chart-card" size="sm" variant="outlined">
              <template #header>
                <div class="chart-header">
                  <RsIcon name="chart-bar" :size="18" color="#8a2be2" />
                  <span>状态码分布</span>
                </div>
              </template>
              <div class="chart-container" ref="statusCodeChartRef">
                <div v-if="page.service.model.loading.value" class="chart-loading">
                  <RsLoading size="lg" />
                </div>
                <div
                  v-else-if="!page.service.model.chartData.statusCodeDistribution.length"
                  class="chart-empty"
                >
                  <RsEmpty description="暂无数据" />
                </div>
              </div>
            </RsCard>

            <RsCard class="chart-card" size="sm" variant="outlined">
              <template #header>
                <div class="chart-header">
                  <RsIcon name="flame" :size="18" color="#d03050" />
                  <span>热点路由TOP10</span>
                </div>
              </template>
              <div class="chart-container" ref="hotRoutesChartRef">
                <div v-if="page.service.model.loading.value" class="chart-loading">
                  <RsLoading size="lg" />
                </div>
                <div
                  v-else-if="!page.service.model.chartData.hotRoutes.length"
                  class="chart-empty"
                >
                  <RsEmpty description="暂无数据" />
                </div>
              </div>
            </RsCard>
          </div>
        </div>
      </template>
    </RsSplitPane>
  </div>
</template>

<script setup lang="ts">
import { RsSearchForm, type RsSearchFormExpose } from '@/components/form/rs-search'
import {
  RsCard,
  RsDivider,
  RsEmpty,
  RsIcon,
  RsLoading,
  RsSplitPane,
  RsStatCard,
  type RsSplitPaneItem,
} from '@/ui'
import { nextTick, onMounted, ref } from 'vue'
import { useMonitoringPage } from './hooks'

defineOptions({
  name: 'MonitoringPanel',
})

const splitPanes: RsSplitPaneItem[] = [
  { key: 'search', size: 'auto' },
  { key: 'content' },
]

const searchFormRef = ref<RsSearchFormExpose | null>(null)

const requestTrendChartRef = ref<HTMLDivElement>()
const responseTimeChartRef = ref<HTMLDivElement>()
const requestMetricsChartRef = ref<HTMLDivElement>()
const statusCodeChartRef = ref<HTMLDivElement>()
const hotRoutesChartRef = ref<HTMLDivElement>()

const page = useMonitoringPage(searchFormRef)

onMounted(async () => {
  await nextTick()
  await nextTick()

  if (requestTrendChartRef.value) {
    page.charts.requestTrendChartRef.value = requestTrendChartRef.value
  }
  if (responseTimeChartRef.value) {
    page.charts.responseTimeChartRef.value = responseTimeChartRef.value
  }
  if (requestMetricsChartRef.value) {
    page.charts.requestMetricsChartRef.value = requestMetricsChartRef.value
  }
  if (statusCodeChartRef.value) {
    page.charts.statusCodeChartRef.value = statusCodeChartRef.value
  }
  if (hotRoutesChartRef.value) {
    page.charts.hotRoutesChartRef.value = hotRoutesChartRef.value
  }

  await page.service.model.bootstrapDefaultGatewayInstance(searchFormRef)
  await page.initPageData()
})

const handleSearch = async (formData?: Record<string, any>) => {
  await page.handleSearch(formData)
}

const handleReset = async () => {
  await page.handleReset()
}

const handleToolbarClick = async (key: string) => {
  await page.handleToolbarClick(key)
}
</script>

<style scoped lang="scss">
.monitoring-panel {
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
}

.monitoring-panel__split {
  flex: 1 1 auto;
  width: 100%;
  height: 100%;
  min-height: 0;
}

.monitoring-panel__search {
  width: 100%;
  box-sizing: border-box;
}

.monitoring-panel__content {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: auto;
}

.overview-card {
  margin: 24px 0;
}

.overview-header,
.chart-header {
  display: flex;
  align-items: center;
  gap: 8px;
}

.overview-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
  padding: 8px 0 12px;
}

.overview-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 16px 24px;
}

.chart-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 20px;
}

.chart-card {
  margin-bottom: 0;

  .chart-container {
    height: 380px;
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1;

    .chart-loading,
    .chart-empty {
      display: flex;
      justify-content: center;
      align-items: center;
      height: 100%;
      width: 100%;
      position: absolute;
      top: 0;
      left: 0;
      background-color: var(--rs-surface);
    }

    .chart-loading {
      z-index: 10;
    }

    .chart-empty {
      z-index: 5;
    }
  }
}

:global(.gateway-monitoring-echarts-tooltip),
:global(div[id^='echarts-tooltip']),
:global(.echarts-tooltip),
:global([class*='echarts-tooltip']) {
  z-index: 10000 !important;
}
</style>
