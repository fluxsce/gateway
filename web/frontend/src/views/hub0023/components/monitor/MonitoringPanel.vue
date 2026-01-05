<template>
  <div class="monitoring-panel">
    <GPane direction="vertical" :no-resize="true">
      <!-- 上部：搜索表单 -->
      <template #1>
        <search-form
          ref="searchFormRef"
          :module-id="page.service.model.moduleId"
          v-bind="page.service.model.searchFormConfig"
          @search="handleSearch"
          @reset="handleReset"
          @toolbar-click="handleToolbarClick"
        />
      </template>

      <!-- 下部：监控内容 -->
      <template #2>
        <div class="monitoring-content">
          <!-- 监控概览 -->
          <n-card class="overview-card" size="small">
            <template #header>
              <div class="overview-header">
                <n-icon size="18" color="#18a058">
                  <StatsChartOutline />
                </n-icon>
                <span>监控概览</span>
              </div>
            </template>
            <div class="overview-content">
              <!-- 主要指标 -->
              <n-grid :cols="3" :x-gap="24" :y-gap="16">
                <n-gi>
                  <n-statistic
                    label="总请求数"
                    :value="page.service.model.overviewData.totalRequests"
                  >
                    <template #prefix>
                      <n-icon color="#18a058">
                        <BarChartOutline />
                      </n-icon>
                    </template>
                  </n-statistic>
                </n-gi>
                <n-gi>
                  <n-statistic
                    label="成功请求数"
                    :value="page.service.model.overviewData.successRequests"
                  >
                    <template #prefix>
                      <n-icon color="#18a058">
                        <CheckmarkCircleOutline />
                      </n-icon>
                    </template>
                  </n-statistic>
                </n-gi>
                <n-gi>
                  <n-statistic
                    label="失败请求数"
                    :value="page.service.model.overviewData.failedRequests"
                  >
                    <template #prefix>
                      <n-icon color="#d03050">
                        <CloseCircleOutline />
                      </n-icon>
                    </template>
                  </n-statistic>
                </n-gi>
              </n-grid>

              <!-- 响应时间指标 -->
              <n-divider style="margin: 20px 0" />
              <n-grid :cols="3" :x-gap="24">
                <n-gi>
                  <n-statistic
                    label="平均响应时间"
                    :value="page.service.model.overviewData.avgResponseTimeMs"
                    suffix="ms"
                  >
                    <template #prefix>
                      <n-icon color="#2080f0">
                        <TimeOutline />
                      </n-icon>
                    </template>
                  </n-statistic>
                </n-gi>
                <n-gi>
                  <n-statistic
                    label="最小响应时间"
                    :value="page.service.model.overviewData.minResponseTimeMs"
                    suffix="ms"
                  >
                    <template #prefix>
                      <n-icon color="#52c41a">
                        <TimeOutline />
                      </n-icon>
                    </template>
                  </n-statistic>
                </n-gi>
                <n-gi>
                  <n-statistic
                    label="最大响应时间"
                    :value="page.service.model.overviewData.maxResponseTimeMs"
                    suffix="ms"
                  >
                    <template #prefix>
                      <n-icon color="#ff4d4f">
                        <TimeOutline />
                      </n-icon>
                    </template>
                  </n-statistic>
                </n-gi>
              </n-grid>
            </div>
          </n-card>

          <!-- 监控图表 -->
          <n-grid :cols="2" :x-gap="20" :y-gap="20">
            <!-- 1. 请求量趋势图 -->
            <n-gi>
              <n-card class="chart-card" size="small">
                <template #header>
                  <div class="chart-header">
                    <n-icon size="18" color="#18a058">
                      <TrendingUpOutline />
                    </n-icon>
                    <span>请求量趋势({{ page.service.model.getTimeGranularityLabel() }})</span>
                  </div>
                </template>
                <div class="chart-container" ref="requestTrendChartRef">
                  <div v-if="page.service.model.loading.value" class="chart-loading">
                    <n-spin size="large" />
                  </div>
                  <div
                    v-else-if="!page.service.model.chartData.requestTrend.length"
                    class="chart-empty"
                  >
                    <n-empty description="暂无数据" />
                  </div>
                </div>
              </n-card>
            </n-gi>

            <!-- 2. 响应时间趋势图 -->
            <n-gi>
              <n-card class="chart-card" size="small">
                <template #header>
                  <div class="chart-header">
                    <n-icon size="18" color="#2080f0">
                      <TimeOutline />
                    </n-icon>
                    <span>响应时间趋势({{ page.service.model.getTimeGranularityLabel() }})</span>
                  </div>
                </template>
                <div class="chart-container" ref="responseTimeChartRef">
                  <div v-if="page.service.model.loading.value" class="chart-loading">
                    <n-spin size="large" />
                  </div>
                  <div
                    v-else-if="!page.service.model.chartData.responseTimeTrend.length"
                    class="chart-empty"
                  >
                    <n-empty description="暂无数据" />
                  </div>
                </div>
              </n-card>
            </n-gi>

            <!-- 3. 请求指标饼图 -->
            <n-gi>
              <n-card class="chart-card" size="small">
                <template #header>
                  <div class="chart-header">
                    <n-icon size="18" color="#f0a020">
                      <PieChartOutline />
                    </n-icon>
                    <span>请求指标分布</span>
                  </div>
                </template>
                <div class="chart-container" ref="requestMetricsChartRef">
                  <div v-if="page.service.model.loading.value" class="chart-loading">
                    <n-spin size="large" />
                  </div>
                  <div
                    v-else-if="page.service.model.overviewData.totalRequests === 0"
                    class="chart-empty"
                  >
                    <n-empty description="暂无数据" />
                  </div>
                </div>
              </n-card>
            </n-gi>

            <!-- 4. 状态码分布图 -->
            <n-gi>
              <n-card class="chart-card" size="small">
                <template #header>
                  <div class="chart-header">
                    <n-icon size="18" color="#8a2be2">
                      <BarChartOutline />
                    </n-icon>
                    <span>状态码分布</span>
                  </div>
                </template>
                <div class="chart-container" ref="statusCodeChartRef">
                  <div v-if="page.service.model.loading.value" class="chart-loading">
                    <n-spin size="large" />
                  </div>
                  <div
                    v-else-if="!page.service.model.chartData.statusCodeDistribution.length"
                    class="chart-empty"
                  >
                    <n-empty description="暂无数据" />
                  </div>
                </div>
              </n-card>
            </n-gi>

            <!-- 5. 热点路由TOP10 -->
            <n-gi>
              <n-card class="chart-card" size="small">
                <template #header>
                  <div class="chart-header">
                    <n-icon size="18" color="#d03050">
                      <FlameOutline />
                    </n-icon>
                    <span>热点路由TOP10</span>
                  </div>
                </template>
                <div class="chart-container" ref="hotRoutesChartRef">
                  <div v-if="page.service.model.loading.value" class="chart-loading">
                    <n-spin size="large" />
                  </div>
                  <div
                    v-else-if="!page.service.model.chartData.hotRoutes.length"
                    class="chart-empty"
                  >
                    <n-empty description="暂无数据" />
                  </div>
                </div>
              </n-card>
            </n-gi>
          </n-grid>
        </div>
      </template>
    </GPane>
  </div>
</template>

<script setup lang="ts">
import { GPane } from '@/components/gpane'
import {
  BarChartOutline,
  CheckmarkCircleOutline,
  CloseCircleOutline,
  FlameOutline,
  PieChartOutline,
  StatsChartOutline,
  TimeOutline,
  TrendingUpOutline,
} from '@vicons/ionicons5'
import { nextTick, onMounted, ref } from 'vue'
import { useMonitoringPage } from './hooks'

// 搜索表单引用
const searchFormRef = ref()

// 图表引用
const requestTrendChartRef = ref<HTMLDivElement>()
const responseTimeChartRef = ref<HTMLDivElement>()
const requestMetricsChartRef = ref<HTMLDivElement>()
const statusCodeChartRef = ref<HTMLDivElement>()
const hotRoutesChartRef = ref<HTMLDivElement>()

// 使用监控页面 Hook
const page = useMonitoringPage(searchFormRef)

// 将图表引用传递给 charts hook（需要在 DOM 渲染后）
onMounted(async () => {
  // 等待 DOM 渲染完成
  await nextTick()
  await nextTick() // 多等待一次确保 GPane 内部 DOM 也渲染完成

  // 将图表 ref 传递给 charts hook
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

  // 初始化页面数据（包括图表初始化和数据加载）
  await page.initPageData()
})

// 处理搜索
const handleSearch = async (formData?: Record<string, any>) => {
  await page.handleSearch(formData)
}

// 处理重置
const handleReset = async () => {
  await page.handleReset()
}

// 处理工具栏点击
const handleToolbarClick = async (key: string) => {
  await page.handleToolbarClick(key)
}
</script>

<style scoped lang="scss">
.monitoring-panel {
  width: 100%;
  height: 100%;
  overflow: hidden;

  /* noResize 模式下使用 flex 布局，需要针对 flex-pane 设置样式 */
  :deep(.g-pane__flex-container--vertical) {
    height: 100%;
    min-height: 0;
  }

  /* 上半区：搜索表单，内容较少，允许自身滚动 */
  :deep(.g-pane__flex-pane--1) {
    overflow: auto;
    padding: var(--g-space-sm);
    min-height: 0;
  }

  /* 下半区：监控内容区域，允许滚动 */
  :deep(.g-pane__flex-pane--2) {
    overflow-y: auto;
    overflow-x: hidden;
    padding: var(--g-space-sm);
    min-height: 0;
    height: 100%;
    /* 确保 tooltip 不会被裁剪 */
    position: relative;
  }

  /* 兼容 n-split 模式（如果 noResize 为 false） */
  :deep(.n-split) {
    height: 100%;
  }

  :deep(.n-split-pane:first-child) {
    overflow: auto;
    padding: var(--g-space-sm);
  }

  :deep(.n-split-pane:last-child) {
    overflow-y: auto;
    overflow-x: hidden;
    padding: var(--g-space-sm);
  }

  .overview-card {
    margin: 24px 0;
  }

  .chart-card {
    margin-bottom: 0;
  }

  .overview-header,
  .chart-header {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .overview-content {
    padding: 20px 0;
  }

  .chart-card {
    .chart-container {
      height: 380px;
      position: relative;
      display: flex;
      align-items: center;
      justify-content: center;
      /* 确保图表容器有足够的层级，tooltip 可以显示在上层 */
      z-index: 1;

      .chart-loading {
        display: flex;
        justify-content: center;
        align-items: center;
        height: 100%;
        width: 100%;
        position: absolute;
        top: 0;
        left: 0;
        background-color: var(--n-color);
        z-index: 10;
      }

      .chart-empty {
        display: flex;
        justify-content: center;
        align-items: center;
        height: 100%;
        width: 100%;
        position: absolute;
        top: 0;
        left: 0;
        z-index: 5;
        background-color: var(--n-color);
      }
    }
  }
  .monitoring-content{
    overflow: auto;
  }
}

/* 确保 ECharts tooltip 显示在最上层，不被其他元素遮挡 */
/* ECharts 5.x 的 tooltip 会附加到 body，但需要确保 z-index 足够高 */
:global(div[id^="echarts-tooltip"]) {
  z-index: 9999 !important;
}

/* 兼容不同版本的 ECharts tooltip 类名 */
:global(.echarts-tooltip),
:global([class*="echarts-tooltip"]) {
  z-index: 9999 !important;
}
</style>
