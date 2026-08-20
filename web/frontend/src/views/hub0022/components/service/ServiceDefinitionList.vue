<template>
  <div class="service-definition-list" id="hub0022-service-definition-list">
    <RsSplitPane
      class="service-definition-list__split"
      orientation="vertical"
      :panes="splitPanes"
      disabled
    >
      <template #search>
        <div class="service-definition-list__search">
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
        <div class="service-definition-list__grid">
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
      </template>
    </RsSplitPane>

    <RsDataFormModal
      v-model:visible="formDialogVisible"
      :module-id="service.model.moduleId"
      :mode="formDialogMode"
      :title="formDialogMode === 'create' ? '新增服务定义' : formDialogMode === 'edit' ? '编辑服务定义' : '查看服务定义详情'"
      to="#hub0022-service-definition-list"
      :form-fields="service.model.serviceFormConfig.fields"
      :form-tabs="service.model.serviceFormConfig.tabs"
      :initial-data="currentEditService || undefined"
      :auto-close-on-confirm="false"
      :confirm-loading="service.model.loading.value"
      @submit="handleFormSubmit"
    />

    <ServiceNodeListModal
      v-model:visible="showNodeDialog"
      :service-definition-id="currentServiceId"
      :title="'服务节点管理'"
      :width="1200"
      to="#hub0022-service-definition-list"
    />

    <CircuitBreakerConfigFormModal
      v-model:visible="showCircuitBreakerDialog"
      :target-service-id="currentCircuitBreakerService?.serviceDefinitionId"
      :service-name="currentCircuitBreakerService?.serviceName"
      to="#hub0022-service-definition-list"
    />
  </div>
</template>

<script lang="ts" setup>
import { RsDataFormModal } from '@/components/form/rs-data'
import { RsSearchForm } from '@/components/form/rs-search'
import { RsGrid, type RsGridExpose } from '@/components/rs-grid'
import { RsSplitPane, type RsSplitPaneItem } from '@/ui'
import { onBeforeUnmount, ref, watch } from 'vue'
import { CircuitBreakerConfigFormModal } from '../circuit-breaker'
import { ServiceNodeListModal } from '../service-nodes'
import { useServiceDefinitionPage } from './hooks/page'

defineOptions({
  name: 'ServiceDefinitionList'
})

/** 上方搜索区随内容自适应，下方表格占满剩余高度 */
const splitPanes: RsSplitPaneItem[] = [
  { key: 'search', size: 'auto' },
  { key: 'grid' },
]

interface Props {
  gatewayInstanceId?: string
}

const props = withDefaults(defineProps<Props>(), {
  gatewayInstanceId: undefined,
})

const searchFormRef = ref()
const gridRef = ref<RsGridExpose | null>(null)

const {
  service,
  formDialogVisible,
  formDialogMode,
  currentEditService,
  showNodeDialog,
  currentServiceId,
  showCircuitBreakerDialog,
  currentCircuitBreakerService,
  handleFormSubmit,
  handleToolbarClick,
  handleMenuClick,
} = useServiceDefinitionPage(props.gatewayInstanceId, searchFormRef, gridRef)

const stopGatewayInstanceIdWatch = watch(
  () => props.gatewayInstanceId,
  (newId, oldId) => {
    if (newId && newId !== oldId) {
      service.loadServiceList({ proxyConfigId: newId })
    } else if (!newId && oldId) {
      service.model.serviceList.value = []
    }
  },
  { immediate: false }
)

onBeforeUnmount(() => {
  stopGatewayInstanceIdWatch()
})

/**
 * 处理搜索（确保使用最新的 gatewayInstanceId）
 */
function handleSearch(formData?: Record<string, any>) {
  if (!props.gatewayInstanceId) {
    return
  }
  const searchParams = formData
    ? {
        ...formData,
        ...(props.gatewayInstanceId ? { proxyConfigId: props.gatewayInstanceId } : {}),
      }
    : props.gatewayInstanceId
      ? { proxyConfigId: props.gatewayInstanceId }
      : undefined
  service.handleSearch(searchParams)
}
</script>

<style lang="scss" scoped>
.service-definition-list {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.service-definition-list__split {
  flex: 1 1 auto;
  width: 100%;
  height: 100%;
  min-height: 0;
}

.service-definition-list__search {
  width: 100%;
}

.service-definition-list__grid {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
</style>
