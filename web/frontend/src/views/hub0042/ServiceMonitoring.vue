<template>
  <div class="service-monitoring" :id="service.model.moduleId">
    <RsSplitPane
      class="service-monitoring__split"
      orientation="vertical"
      :panes="splitPanes"
      disabled
    >
      <template #search>
        <div class="service-monitoring__search">
          <RsSearchForm
            ref="serviceSearchFormRef"
            :module-id="service.model.moduleId"
            v-bind="service.model.searchFormConfig"
            @search="handleServiceSearch"
            @reset="handleServiceReset"
            @toolbar-click="(key) => handleServiceToolbarClick(key)"
          />
        </div>
      </template>

      <template #grid>
        <div class="service-monitoring__list">
          <div v-if="selectedNamespace" class="service-monitoring__summary">
            <div
              v-for="item in summaryItems"
              :key="item.label"
              class="service-monitoring__stat"
            >
              <span>{{ item.label }}</span>
              <strong>{{ item.value }}</strong>
            </div>
          </div>
          <div v-if="!selectedNamespace && !hasServices && !isLoading" class="service-monitoring__empty">
            <RsEmpty :description="emptyDescription">
              <template #icon>
                <GIcon :icon="ServerOutline" :size="32" color="var(--g-primary)" />
              </template>
            </RsEmpty>
          </div>
          <div v-else class="service-monitoring__grid">
            <RsGrid
              ref="serviceGridRef"
              :module-id="service.model.moduleId"
              :data="service.model.serviceList"
              :loading="service.model.loading"
              :columns="service.model.gridConfig.columns"
              :selectable="service.model.gridConfig.selectable"
              :row-key="service.model.gridConfig.rowKey"
              height="100%"
              :pagination-config="service.model.gridConfig.paginationConfig"
              :menu-config="service.model.gridConfig.menuConfig"
              @page-change="handleServicePageChange"
              @menu-click="handleServiceMenuClick"
            />
          </div>
        </div>
      </template>
    </RsSplitPane>

    <RsDataFormModal
      v-model:visible="serviceFormDialogVisible"
      :module-id="service.model.moduleId"
      :mode="serviceFormDialogMode"
      :title="serviceFormDialogMode === 'create' ? '新增服务' : serviceFormDialogMode === 'edit' ? '编辑服务' : '查看服务详情'"
      :to="`#${service.model.moduleId}`"
      :form-fields="service.model.serviceFormConfig.fields"
      :form-tabs="service.model.serviceFormConfig.tabs"
      :initial-data="currentEditService || undefined"
      :auto-close-on-confirm="false"
      :confirm-loading="serviceSubmitting"
      @submit="handleServiceFormSubmit"
    />
  </div>
</template>

<script lang="ts" setup>
import { RsDataFormModal } from '@/components/form/rs-data'
import { RsSearchForm } from '@/components/form/rs-search'
import { GIcon } from '@/components/gicon'
import { RsGrid, type RsGridExpose } from '@/components/rs-grid'
import { RsEmpty, RsSplitPane, type RsSplitPaneItem } from '@/ui'
import { ServerOutline } from '@vicons/ionicons5'
import { computed, ref } from 'vue'
import { useServicePage } from './hooks'

defineOptions({
  name: 'ServiceMonitoring',
})

const splitPanes: RsSplitPaneItem[] = [
  { key: 'search', size: 'auto' },
  { key: 'grid' },
]

const serviceSearchFormRef = ref()
const serviceGridRef = ref<RsGridExpose | null>(null)

const {
  service,
  formDialogVisible: serviceFormDialogVisible,
  formDialogMode: serviceFormDialogMode,
  currentEditService,
  submitting: serviceSubmitting,
  selectedNamespace,
  handleFormSubmit: handleServiceFormSubmit,
  handleToolbarClick: handleServiceToolbarClick,
  handleMenuClick: handleServiceMenuClick,
  handleSearch: handleServiceSearch,
  handleReset: handleServiceReset,
} = useServicePage(serviceGridRef, serviceSearchFormRef)

const summaryItems = computed(() => {
  const ns = selectedNamespace.value
  const quota = ns?.serviceQuotaLimit
  const quotaText = quota === undefined || quota === null || quota === 0 ? '无限制' : String(quota)
  return [
    { label: '命名空间', value: ns?.namespaceName || ns?.namespaceId || '--' },
    { label: '服务', value: String(ns?.serviceCount ?? 0) },
    { label: '节点', value: String(ns?.nodeCount ?? 0) },
    { label: '健康', value: String(ns?.healthyNodeCount ?? 0) },
    { label: '连接', value: String(ns?.connectionCount ?? 0) },
    { label: '服务配额', value: quotaText },
  ]
})

const hasServices = computed(() => service.model.serviceList.value.length > 0)
const isLoading = computed(() => Boolean(service.model.loading.value))
const emptyDescription = computed(() => (
  selectedNamespace.value ? '该命名空间下暂无匹配的服务' : '请先选择命名空间'
))

const handleServicePageChange = (params: { currentPage: number; pageSize: number }) => {
  service.handlePageChange(params.currentPage, params.pageSize)
}
</script>

<style lang="scss" scoped>
.service-monitoring {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.service-monitoring__split {
  flex: 1 1 auto;
  width: 100%;
  height: 100%;
  min-height: 0;
}

.service-monitoring__search {
  width: 100%;
  box-sizing: border-box;
}

.service-monitoring__list {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  background: var(--g-bg-primary);
}

.service-monitoring__empty,
.service-monitoring__grid {
  flex: 1 1 auto;
  min-height: 0;
}

.service-monitoring__empty {
  display: flex;
  align-items: center;
  justify-content: center;
}

.service-monitoring__summary {
  flex: none;
  display: grid;
  grid-template-columns: repeat(6, minmax(0, 1fr));
  gap: 8px;
  padding: 10px 12px;
  border-bottom: 1px solid var(--g-border-primary);
}

.service-monitoring__stat {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
  padding: 8px 10px;
  border: 1px solid var(--g-border-primary);
  border-radius: var(--g-radius-lg);
  background: var(--g-bg-secondary);

  span {
    font-size: 12px;
    color: var(--g-text-secondary);
  }

  strong {
    font-size: 16px;
    font-weight: 600;
    font-variant-numeric: tabular-nums;
    color: var(--g-text-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

@media (max-width: 1280px) {
  .service-monitoring__summary {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}
</style>
