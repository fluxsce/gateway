<template>
  <div class="gateway-log-query">
    <GPane direction="vertical" :no-resize="true">
      <!-- 上部：搜索表单 -->
      <template #1>
        <search-form
          ref="searchFormRef"
          :module-id="service.model.moduleId"
          v-bind="service.model.searchFormConfig"
          @search="handleSearch"
          @toolbar-click="handleToolbarClick"
        />
      </template>

      <!-- 下部：数据表格 -->
      <template #2>
        <g-grid
          ref="gridRef"
          :module-id="service.model.moduleId"
          :data="service.model.logList"
          :loading="service.model.loading"
          v-bind="service.model.gridConfig"
          @page-change="service.handlePageChange"
          @menu-click="handleMenuClick"
        >
          <!-- 路由名称 -->
          <template #routeName="{ row }">
            <span class="route-name-text">{{ row.routeName || '-' }}</span>
          </template>

          <!-- 请求方法 -->
          <template #requestMethod="{ row }">
            <n-tag
              :type="service.model.getMethodTagType(row.requestMethod)"
              size="small"
            >
              {{ row.requestMethod }}
            </n-tag>
          </template>

          <!-- 状态码 -->
          <template #gatewayStatusCode="{ row }">
            <n-tag
              :type="service.model.getStatusCodeTagType(row.gatewayStatusCode)"
              size="small"
            >
              {{ row.gatewayStatusCode }}
            </n-tag>
          </template>

          <!-- 后端状态码 -->
          <template #backendStatusCode="{ row }">
            <n-tag
              v-if="row.backendStatusCode != null"
              :type="service.model.getStatusCodeTagType(row.backendStatusCode)"
              size="small"
            >
              {{ row.backendStatusCode }}
            </n-tag>
            <span v-else>-</span>
          </template>

          <!-- 总处理时间 -->
          <template #totalProcessingTimeMs="{ row }">
            <n-tag
              v-if="row.totalProcessingTimeMs != null"
              :type="service.model.getTimeTagType(row.totalProcessingTimeMs, 5000, 1000)"
              size="small"
            >
              {{ row.totalProcessingTimeMs }}ms
            </n-tag>
            <span v-else>-</span>
          </template>

          <!-- 网关耗时 -->
          <template #gatewayProcessingTimeMs="{ row }">
            <n-tag
              v-if="row.gatewayProcessingTimeMs != null"
              :type="service.model.getTimeTagType(row.gatewayProcessingTimeMs, 2000, 500)"
              size="small"
            >
              {{ row.gatewayProcessingTimeMs }}ms
            </n-tag>
            <span v-else>-</span>
          </template>

          <!-- 后端耗时 -->
          <template #backendResponseTimeMs="{ row }">
            <n-tag
              v-if="row.backendResponseTimeMs != null"
              :type="service.model.getTimeTagType(row.backendResponseTimeMs, 3000, 1000)"
              size="small"
            >
              {{ row.backendResponseTimeMs }}ms
            </n-tag>
            <span v-else>-</span>
          </template>

          <!-- 处理状态 -->
          <template #processingStatus="{ row }">
            <n-tag
              :type="service.model.getProcessingStatusTagType(row)"
              size="small"
            >
              {{ service.model.getProcessingStatusText(row) }}
            </n-tag>
          </template>

          <!-- 代理类型 -->
          <template #proxyType="{ row }">
            <n-tag
              v-if="row.proxyType"
              :type="service.model.getProxyTypeTagType(row.proxyType)"
              size="small"
            >
              {{ row.proxyType.toUpperCase() }}
            </n-tag>
            <span v-else>-</span>
          </template>

          <!-- 重置状态 -->
          <template #resetFlag="{ row }">
            <n-tag
              :type="row.resetFlag === 'Y' ? 'warning' : 'default'"
              size="small"
            >
              {{ row.resetFlag === 'Y' ? '已重置' : '未重置' }}
            </n-tag>
          </template>

          <!-- 日志级别 -->
          <template #logLevel="{ row }">
            <n-tag
              v-if="row.logLevel"
              :type="service.model.getLogLevelTagType(row.logLevel)"
              size="small"
            >
              {{ row.logLevel }}
            </n-tag>
            <span v-else>-</span>
          </template>
        </g-grid>
      </template>
    </GPane>

    <!-- 后端日志详情对话框 -->
    <BackendLogsDialog
      v-model:visible="detailDialogVisible"
      :trace-id="selectedTraceId"
    />
  </div>
</template>

<script lang="ts" setup>
import SearchForm from '@/components/form/search/SearchForm.vue'
import { GPane } from '@/components/gpane'
import { GGrid } from '@/components/grid'
import { NTag } from 'naive-ui'
import { ref } from 'vue'
import BackendLogsDialog from '../backed-logs/BackendLogsDialog.vue'
import { useGatewayLogPage } from './hooks'

// 定义组件名称
defineOptions({
  name: 'GatewayLogQuery'
})

// ============= Refs =============

const searchFormRef = ref()
const gridRef = ref()

// ============= 页面级 Hook（包含服务与对话框、事件处理） =============

const {
  service,
  detailDialogVisible,
  selectedTraceId,
  handleToolbarClick,
  handleMenuClick,
  handleSearch,
} = useGatewayLogPage(gridRef, searchFormRef)

// 暴露 refs 给父组件（如果需要）
defineExpose({
  searchFormRef,
  gridRef,
  service
})
</script>

<style lang="scss" scoped>
.gateway-log-query {
  width: 100%;
  height: 100%;
  overflow: hidden;

  :deep(.n-split) {
    height: 100%;
  }

  /* 上半区：搜索表单，内容较少，允许自身滚动 */
  :deep(.n-split-pane:first-child) {
    overflow: auto;
    padding: var(--g-space-sm);
  }

  /* 下半区：表格区域，高度由 GGrid 占满，滚动全部交给 vxe-grid */
  :deep(.n-split-pane:last-child) {
    overflow: hidden;
    padding: var(--g-space-sm);
    display: flex;
    flex-direction: column;
  }
}

/* 路由名称突出显示样式 */
.route-name-text {
  color: var(--g-primary, #7c3aed);
}
</style>

