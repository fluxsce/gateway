<template>
  <div class="tunnel-service-management" :id="htmlId">
    <RsSplitPane
      class="tunnel-service-management__split"
      orientation="vertical"
      :panes="splitPanes"
      disabled
    >
      <template #search>
        <div class="tunnel-service-management__search">
          <RsSearchForm
            ref="searchFormRef"
            :module-id="service.model.moduleId"
            v-bind="service.model.searchFormConfig"
            @search="handleSearchWithStats"
            @toolbar-click="handleToolbarClick"
          />
        </div>
      </template>

      <template #grid>
        <div class="tunnel-service-management__body">
          <div v-if="showStats" class="tunnel-service-management__stats">
            <TunnelServiceStats :statistics="statistics" />
          </div>
          <div class="tunnel-service-management__grid">
            <RsGrid
              ref="gridRef"
              :module-id="service.model.moduleId"
              :data="service.model.serviceList"
              :loading="service.model.loading"
              :columns="service.model.gridConfig.columns"
              :selectable="service.model.gridConfig.selectable"
              :row-key="service.model.gridConfig.rowKey"
              height="100%"
              :pagination-config="service.model.gridConfig.paginationConfig"
              :menu-config="service.model.gridConfig.menuConfig"
              @page-change="service.handlePageChange"
              @menu-click="handleMenuClick"
            />
          </div>
        </div>
      </template>
    </RsSplitPane>

    <RsDataFormModal
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
import { RsDataFormModal } from '@/components/form/rs-data'
import { RsSearchForm } from '@/components/form/rs-search'
import { RsGrid, type RsGridExpose } from '@/components/rs-grid'
import { RsSplitPane, type RsSplitPaneItem } from '@/ui'
import { isApiSuccess, parseJsonData } from '@/utils/format'
import { onMounted, ref } from 'vue'
import * as tunnelServiceApi from '../../api'
import type { TunnelServiceStats as TunnelServiceStatsType } from '../../types'
import { useTunnelServicePage } from './hooks'
import TunnelServiceStats from './TunnelServiceStats.vue'

defineOptions({
  name: 'TunnelServiceManagement',
})

const splitPanes: RsSplitPaneItem[] = [
  { key: 'search', size: 'auto' },
  { key: 'grid' },
]

const searchFormRef = ref()
const gridRef = ref<RsGridExpose | null>(null)

const showStats = ref(true)
const statistics = ref<TunnelServiceStatsType>({
  totalServices: 0,
  activeServices: 0,
  inactiveServices: 0,
  errorServices: 0,
  offlineServices: 0,
  totalConnections: 0,
  totalTraffic: 0,
})

/**
 * 获取统计信息
 */
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
        totalTraffic: 0,
      })
      statistics.value = data
    }
  } catch (error) {
    console.error('获取统计信息失败:', error)
  }
}

const {
  service,
  formDialogVisible,
  formDialogMode,
  currentEditService,
  handleFormSubmit,
  handleToolbarClick,
  handleMenuClick,
  handleSearch,
} = useTunnelServicePage(gridRef, searchFormRef)

const htmlId = 'hub0062-service'

/**
 * 包装搜索方法，搜索后刷新统计
 */
const handleSearchWithStats = async (searchParams?: Record<string, any>) => {
  await handleSearch(searchParams)
  await getStatistics()
}

onMounted(() => {
  if (showStats.value) {
    getStatistics()
  }
})
</script>

<style lang="scss" scoped>
.tunnel-service-management {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.tunnel-service-management__split {
  flex: 1 1 auto;
  width: 100%;
  height: 100%;
  min-height: 0;
}

.tunnel-service-management__search {
  width: 100%;
  box-sizing: border-box;
}

.tunnel-service-management__body {
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
}

.tunnel-service-management__stats {
  flex-shrink: 0;
}

.tunnel-service-management__grid {
  box-sizing: border-box;
  flex: 1;
  width: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
</style>
