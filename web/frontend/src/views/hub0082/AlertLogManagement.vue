<template>
  <div class="alert-log-management" :id="htmlId">
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
          @page-change="handlePageChange"
          @menu-click="({ code, row }) => handleMenuClick({ menu: { code }, row })"
        >
          <!-- 告警级别自定义渲染 -->
          <template #alertLevel="{ row }">
            <n-tag :type="service.model.getAlertLevelTagType(row.alertLevel)" size="small">
              {{ service.model.getAlertLevelLabel(row.alertLevel) }}
            </n-tag>
          </template>

          <!-- 发送状态自定义渲染 -->
          <template #sendStatus="{ row }">
            <n-tag :type="service.model.getSendStatusTagType(row.sendStatus)" size="small">
              {{ service.model.getSendStatusLabel(row.sendStatus) }}
            </n-tag>
          </template>
        </g-grid>
      </template>
    </GPane>

    <!-- 预警日志详情对话框 -->
    <AlertLogDetailDialog
      v-model:visible="viewDialogVisible"
      :alert-log-id="selectedAlertLogId"
    />
  </div>
</template>

<script lang="ts" setup>
import SearchForm from '@/components/form/search/SearchForm.vue'
import { GPane } from '@/components/gpane'
import { GGrid } from '@/components/grid'
import { NTag } from 'naive-ui'
import { ref } from 'vue'
import { AlertLogDetailDialog } from './components'
import { useAlertLogPage } from './hooks'

// 定义组件名称
defineOptions({
  name: 'AlertLogManagement'
})

// ============= Refs =============

const searchFormRef = ref()
const gridRef = ref()

// ============= 页面级 Hook（包含服务与对话框、事件处理） =============

const {
  service,
  viewDialogVisible,
  selectedAlertLogId,
  handleToolbarClick,
  handleMenuClick,
  handleSearch,
  handlePageChange,
} = useAlertLogPage(gridRef, searchFormRef)

// ============= HTML ID（用于 DOM，符合 HTML 规范） =============

// 固定的 HTML id（符合 HTML 规范，无特殊字符）
const htmlId = 'hub0082-alert-log'
</script>

<style scoped>
.alert-log-management {
  height: 100%;
  display: flex;
  flex-direction: column;
}

</style>

