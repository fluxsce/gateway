<template>
  <div class="tunnel-service-management" :id="htmlId">
    <GPane direction="vertical" :default-size="0.12" :min="0.1" :max="0.5">
      <!-- 上部：搜索表单 -->
      <template #1>
        <search-form
          ref="searchFormRef"
          :module-id="service.model.moduleId"
          v-bind="service.model.searchFormConfig"
          @search="handleSearchWithStats"
          @toolbar-click="handleToolbarClick"
        />
      </template>

      <!-- 下部：统计面板 + 数据表格 -->
      <template #2>
        <div class="bottom-section">
          <!-- 统计面板 -->
          <div class="stats-section" v-if="showStats">
            <tunnel-service-stats :statistics="statistics" />
          </div>

          <!-- 数据表格 -->
          <div class="grid-section">
            <g-grid
              ref="gridRef"
              :module-id="service.model.moduleId"
              :data="service.model.serviceList"
              :loading="service.model.loading"
              v-bind="service.model.gridConfig"
              @page-change="service.handlePageChange"
              @menu-click="handleMenuClick"
            >
              <!-- 服务类型自定义渲染 -->
              <template #serviceType="{ row }">
                <n-tag :type="getServiceTypeTagType(row.serviceType)" size="small">
                  {{ service.model.getServiceTypeLabel(row.serviceType) }}
                </n-tag>
              </template>

              <!-- 服务状态自定义渲染 -->
              <template #serviceStatus="{ row }">
                <n-tag :type="service.model.getServiceStatusTagType(row.serviceStatus)" size="small">
                  {{ service.model.getServiceStatusLabel(row.serviceStatus) }}
                </n-tag>
              </template>

              <!-- 状态自定义渲染 -->
              <template #activeFlag="{ row }">
                <n-tag :type="row.activeFlag === 'Y' ? 'success' : 'default'" size="small">
                  {{ row.activeFlag === 'Y' ? '启用' : '禁用' }}
                </n-tag>
              </template>
            </g-grid>
          </div>
        </div>
      </template>
    </GPane>

    <!-- 服务对话框（新增/编辑/查看共用） -->
    <GdataFormModal
      v-model:visible="formDialogVisible"
      :mode="formDialogMode"
      :title="formDialogMode === 'create' ? '新增隧道服务' : formDialogMode === 'edit' ? '编辑隧道服务' : '查看隧道服务详情'"
      :to="`#${htmlId}`"
      :form-fields="service.model.formFields"
      :form-tabs="service.model.formTabs"
      :initial-data="currentEditService || undefined"
      :auto-close-on-confirm="false"
      :confirm-loading="service.model.loading.value"
      @submit="handleFormSubmit"
    />
  </div>
</template>

<script lang="ts" setup>
import GdataFormModal from '@/components/form/data/GDataFormModal.vue'
import SearchForm from '@/components/form/search/SearchForm.vue'
import { GPane } from '@/components/gpane'
import { GGrid } from '@/components/grid'
import { isApiSuccess, parseJsonData } from '@/utils/format'
import { NTag } from 'naive-ui'
import { onMounted, ref } from 'vue'
import * as tunnelServiceApi from '../../api'
import type { TunnelServiceStats as TunnelServiceStatsType } from '../../types'
import TunnelServiceStats from './TunnelServiceStats.vue'
import { useTunnelServicePage } from './hooks'

// 定义组件名称
defineOptions({
  name: 'TunnelServiceManagement'
})

// ============= Refs =============

const searchFormRef = ref()
const gridRef = ref()

// ============= 统计信息（可选） =============
const showStats = ref(true) // 显示统计信息
const statistics = ref<TunnelServiceStatsType>({
  totalServices: 0,
  activeServices: 0,
  inactiveServices: 0,
  errorServices: 0,
  offlineServices: 0,
  totalConnections: 0,
  totalTraffic: 0
})

// 获取统计信息
const getStatistics = async () => {
  try {
    const response = await tunnelServiceApi.getServiceStats()
    if (isApiSuccess(response)) {
      const data = parseJsonData<TunnelServiceStatsType>(response, {
        totalServices: 0,
        activeServices: 0,
        inactiveServices: 0,
        errorServices: 0,
        offlineServices: 0,
        totalConnections: 0,
        totalTraffic: 0
      })
      statistics.value = data
    }
  } catch (error) {
    console.error('获取统计信息失败:', error)
  }
}

// ============= 页面级 Hook（包含服务与对话框、事件处理） =============

const {
  service,
  formDialogVisible,
  formDialogMode,
  currentEditService,
  handleFormSubmit,
  handleToolbarClick,
  handleMenuClick,
  handleSearch
} = useTunnelServicePage(gridRef, searchFormRef)

// ============= HTML ID（用于 DOM，符合 HTML 规范） =============

// 固定的 HTML id（符合 HTML 规范，无特殊字符）
// 注意：权限校验仍使用原始 moduleId（service.model.moduleId）
const htmlId = 'hub0062-service'

// 包装搜索方法，搜索后刷新统计
const handleSearchWithStats = async (searchParams?: Record<string, any>) => {
  await handleSearch(searchParams)
  await getStatistics()
}

const getServiceTypeTagType = (type: string): 'primary' | 'info' | 'success' | 'warning' => {
  switch (type) {
    case 'tcp':
    case 'http':
      return 'primary'
    case 'udp':
    case 'https':
      return 'success'
    case 'stcp':
    case 'xtcp':
      return 'info'
    default:
      return 'warning'
  }
}

// 初始化
onMounted(() => {
  // 获取统计信息（如果需要）
  if (showStats.value) {
    getStatistics()
  }
  // 数据由搜索表单的"查询"按钮触发加载
})
</script>

<style lang="scss" scoped>
.tunnel-service-management {
  width: 100%;
  height: 100%;
  overflow: hidden;
  background-color: var(--n-color-target);
}

:deep(.n-split) {
  height: 100%;
}

:deep(.n-split-pane:first-child) {
  overflow: auto;
  padding: var(--g-space-sm);
}

:deep(.n-split-pane:last-child) {
  overflow: hidden;
  padding: var(--g-space-sm);
  display: flex;
  flex-direction: column;
}

.bottom-section {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.stats-section {
  flex-shrink: 0;
}

.grid-section {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

</style>

