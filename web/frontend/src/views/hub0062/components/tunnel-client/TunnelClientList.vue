<template>
  <div class="tunnel-client-management" :id="htmlId">
    <RsSplitPane
      class="tunnel-client-management__split"
      orientation="vertical"
      :panes="splitPanes"
      disabled
    >
      <template #search>
        <div class="tunnel-client-management__search">
          <RsSearchForm
            ref="searchFormRef"
            :module-id="service.model.moduleId"
            v-bind="service.model.searchFormConfig"
            @search="handleSearch"
            @toolbar-click="handleToolbarClick"
          />
        </div>
      </template>

      <template #grid>
        <div class="tunnel-client-management__body">
          <div v-if="showStats" class="tunnel-client-management__stats">
            <TunnelClientStats :statistics="statistics" />
          </div>
          <div class="tunnel-client-management__grid">
            <RsGrid
              ref="gridRef"
              :module-id="service.model.moduleId"
              :data="service.model.clientList"
              :loading="service.model.loading"
              :columns="service.model.gridConfig.columns"
              :selectable="service.model.gridConfig.selectable"
              :row-key="service.model.gridConfig.rowKey"
              height="100%"
              :pagination-config="service.model.gridConfig.paginationConfig"
              :menu-config="service.model.gridConfig.menuConfig"
              @page-change="handlePageChange"
              @menu-click="handleMenuClick"
            />
          </div>
        </div>
      </template>
    </RsSplitPane>

    <RsDataFormModal
      v-model:visible="formDialogVisible"
      :module-id="service.model.moduleId"
      :mode="formDialogMode"
      :title="formDialogMode === 'create' ? '新增隧道客户端' : formDialogMode === 'edit' ? '编辑隧道客户端' : '查看隧道客户端详情'"
      :to="`#${htmlId}`"
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
import { RsDataFormModal } from '@/components/form/rs-data'
import { RsSearchForm } from '@/components/form/rs-search'
import { RsGrid, type RsGridExpose } from '@/components/rs-grid'
import { RsSplitPane, type RsSplitPaneItem } from '@/ui'
import { isApiSuccess, parseJsonData } from '@/utils/format'
import { onMounted, ref } from 'vue'
import * as tunnelClientApi from '../../api'
import type { TunnelClientStats as TunnelClientStatsType } from '../../types'
import { useTunnelClientPage } from './hooks'
import TunnelClientStats from './TunnelClientStats.vue'

defineOptions({
  name: 'TunnelClientList',
})

const splitPanes: RsSplitPaneItem[] = [
  { key: 'search', size: 'auto' },
  { key: 'grid' },
]

const searchFormRef = ref()
const gridRef = ref<RsGridExpose | null>(null)

const showStats = ref(true)
const statistics = ref<TunnelClientStatsType>({
  totalClients: 0,
  connectedClients: 0,
  disconnectedClients: 0,
  connectingClients: 0,
  errorClients: 0,
  totalServices: 0,
})

/**
 * 加载统计数据
 */
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

const {
  service,
  formDialogVisible,
  formDialogMode,
  currentEditClient,
  handleFormSubmit,
  handleToolbarClick,
  handleMenuClick,
  handleSearch: originalHandleSearch,
  handlePageChange,
} = useTunnelClientPage(gridRef, searchFormRef)

const htmlId = 'hub0062-tunnel-client'

/**
 * 包装搜索方法，搜索后刷新统计
 */
const handleSearch = async (searchParams?: Record<string, any>) => {
  await originalHandleSearch(searchParams)
  await loadStatistics()
}

onMounted(async () => {
  await Promise.all([service.loadClientList(), loadStatistics()])
})
</script>

<style lang="scss" scoped>
.tunnel-client-management {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.tunnel-client-management__split {
  flex: 1 1 auto;
  width: 100%;
  height: 100%;
  min-height: 0;
}

.tunnel-client-management__search {
  width: 100%;
  box-sizing: border-box;
}

.tunnel-client-management__body {
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
}

.tunnel-client-management__stats {
  flex-shrink: 0;
}

.tunnel-client-management__grid {
  box-sizing: border-box;
  flex: 1;
  width: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
</style>
