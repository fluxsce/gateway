<template>
  <div class="tunnel-client-management" :id="service.model.moduleId">
    <GPane direction="vertical" :default-size="0.12" :min="0.1" :max="0.5">
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

      <!-- 下部：统计面板 + 数据表格 -->
      <template #2>
        <div class="bottom-section">
          <!-- 统计面板 -->
          <div class="stats-section" v-if="showStats">
            <tunnel-client-stats :statistics="statistics" />
          </div>

          <!-- 数据表格 -->
          <div class="grid-section">
            <g-grid
              ref="gridRef"
              :module-id="service.model.moduleId"
              :data="service.model.clientList"
              :loading="service.model.loading"
              v-bind="service.model.gridConfig"
              @page-change="handlePageChange"
              @menu-click="({ code, row }) => handleMenuClick({ menu: { code }, row })"
            >
              <!-- 连接状态自定义渲染 -->
              <template #connectionStatus="{ row }">
                <n-tag :type="service.model.getConnectionStatusTagType(row.connectionStatus)" size="small">
                  {{ service.model.getConnectionStatusLabel(row.connectionStatus) }}
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

    <!-- 客户端对话框（新增/编辑/查看共用） -->
    <GdataFormModal
      v-model:visible="formDialogVisible"
      :mode="formDialogMode"
      :title="formDialogMode === 'create' ? '新增隧道客户端' : formDialogMode === 'edit' ? '编辑隧道客户端' : '查看隧道客户端详情'"
      :to="`#${service.model.moduleId}`"
      :form-fields="service.model.formFields"
      :form-tabs="service.model.formTabs"
      :initial-data="currentEditClient || undefined"
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
import * as tunnelClientApi from '../../api'
import type { TunnelClientStats as TunnelClientStatsType } from '../../types'
import TunnelClientStats from './TunnelClientStats.vue'
import { useTunnelClientPage } from './hooks'

// 定义组件名称
defineOptions({
  name: 'TunnelClientList'
})

// ============= Refs =============

const searchFormRef = ref()
const gridRef = ref()

// ============= 统计面板 =============

const showStats = ref(true)
const statistics = ref<TunnelClientStatsType>({
  totalClients: 0,
  connectedClients: 0,
  disconnectedClients: 0,
  connectingClients: 0,
  errorClients: 0,
  totalServices: 0
})

// 加载统计数据
const loadStatistics = async () => {
  try {
    const res = await tunnelClientApi.getClientStats()
    if (isApiSuccess(res)) {
      const stats = parseJsonData<TunnelClientStatsType>(res)
      if (stats) {
        statistics.value = stats
      }
    }
  } catch (error) {
    console.error('加载统计数据失败:', error)
  }
}

// ============= 页面级 Hook（包含服务与对话框、事件处理） =============

const {
  service,
  formDialogVisible,
  formDialogMode,
  currentEditClient,
  handleFormSubmit,
  handleToolbarClick,
  handleMenuClick,
  handleSearch: originalHandleSearch,
  handlePageChange
} = useTunnelClientPage(gridRef, searchFormRef)

// ============= 事件处理 =============

// 包装搜索方法，搜索后刷新统计
const handleSearch = async (searchParams?: Record<string, any>) => {
  await originalHandleSearch(searchParams)
  await loadStatistics()
}

// ============= 生命周期 =============

// 组件挂载时加载数据
onMounted(async () => {
  await Promise.all([
    service.loadClientList(),
    loadStatistics()
  ])
})
</script>

<style lang="scss" scoped>
.tunnel-client-management {
  width: 100%;
  height: 100%;
  overflow: hidden;
  background-color: var(--n-color-target);
}

:deep(.n-split) {
  height: 100%;
}

/* 上半区：搜索表单，内容较少，允许自身滚动 */
:deep(.n-split-pane:first-child) {
  overflow: auto;
  padding: var(--g-space-sm);
}

/* 下半区：统计面板 + 表格区域 */
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

