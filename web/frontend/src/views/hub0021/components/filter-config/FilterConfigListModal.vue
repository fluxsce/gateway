<template>
  <RsDialog
    :open="modalVisible"
    :title="props.title || (props.filterScope === 'global' ? '全局过滤器配置列表' : '路由过滤器配置列表')"
    layout="window"
    :width="props.width || 1200"
    :teleport-to="props.to"
    :draggable="true"
    :fullscreenable="true"
    :modal="false"
    :show-overlay="false"
    :close-on-overlay-click="false"
    class="hub0021-list-dialog"
    @update:open="handleUpdateVisible"
    @after-close="handleAfterLeave"
  >
    <template #body>
      <div class="filter-config-list-modal" id="filter-config-list-modal">
        <RsSplitPane
          class="filter-config-list-modal__split"
          orientation="vertical"
          :panes="splitPanes"
          disabled
        >
          <template #search>
            <div class="filter-config-list-modal__search">
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
            <div class="filter-config-list-modal__grid">
              <RsGrid
                ref="gridRef"
                :module-id="service.model.moduleId"
                :data="service.model.filterList"
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
          </template>
        </RsSplitPane>

        <RsDataFormModal
          v-model:visible="formDialogVisible"
          :mode="formDialogMode"
          :title="formDialogMode === 'create' ? '新增过滤器配置' : formDialogMode === 'edit' ? '编辑过滤器配置' : '查看过滤器配置详情'"
          to="#filter-config-list-modal"
          :form-fields="service.model.formFields"
          :form-tabs="service.model.formTabs"
          :initial-data="currentEditFilter || undefined"
          :auto-close-on-confirm="false"
          :confirm-loading="service.model.loading.value"
          @submit="handleFormSubmit"
        />
      </div>
    </template>
  </RsDialog>
</template>

<script lang="ts" setup>
import { RsDataFormModal } from '@/components/form/rs-data'
import { RsSearchForm } from '@/components/form/rs-search'
import { RsGrid, type RsGridExpose } from '@/components/rs-grid'
import { RsDialog, RsSplitPane, type RsSplitPaneItem } from '@/ui'
import { onBeforeUnmount, ref, watch } from 'vue'
import { useFilterConfigPage } from './hooks'
import type { FilterConfigListModalEmits, FilterConfigListModalProps } from './hooks/types'

defineOptions({
  name: 'FilterConfigListModal',
})

const props = withDefaults(defineProps<FilterConfigListModalProps>(), {
  visible: false,
  title: '',
  width: 1200,
  to: undefined,
  gatewayInstanceId: undefined,
  routeConfigId: undefined,
  filterScope: 'global',
})

const emit = defineEmits<FilterConfigListModalEmits>()

const splitPanes: RsSplitPaneItem[] = [
  { key: 'search', size: 'auto' },
  { key: 'grid' },
]

const searchFormRef = ref()
const gridRef = ref<RsGridExpose | null>(null)
const moduleIdRef = ref<string>(props.moduleId)
const modalVisible = ref(props.visible)
const gatewayInstanceId = ref<string | undefined>(props.gatewayInstanceId)
const routeConfigId = ref<string | undefined>(props.routeConfigId)

const stopVisibleWatch = watch(() => props.visible, (newVal) => {
  modalVisible.value = newVal
})

const stopGatewayInstanceIdWatch = watch(() => props.gatewayInstanceId, (newVal) => {
  gatewayInstanceId.value = newVal
})

const stopRouteConfigIdWatch = watch(() => props.routeConfigId, (newVal) => {
  routeConfigId.value = newVal
})

onBeforeUnmount(() => {
  stopVisibleWatch()
  stopGatewayInstanceIdWatch()
  stopRouteConfigIdWatch()
})

const {
  service,
  formDialogVisible,
  formDialogMode,
  currentEditFilter,
  handleFormSubmit,
  handleToolbarClick,
  handleMenuClick,
  handleSearch,
  handlePageChange,
} = useFilterConfigPage(moduleIdRef, gridRef, gatewayInstanceId, routeConfigId, searchFormRef)

const handleUpdateVisible = (value: boolean) => {
  modalVisible.value = value
  emit('update:visible', value)
  if (!value) {
    emit('close')
  } else {
    emit('refresh')
    service.loadFilterList()
  }
}

const handleAfterLeave = () => {
  if (!modalVisible.value) {
    formDialogVisible.value = false
    formDialogMode.value = 'create'
    currentEditFilter.value = null
    service.model.filterList.value = []
    service.model.resetPagination()
  }
}
</script>

<style scoped>
.filter-config-list-modal {
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  width: 100%;
  height: min(70vh, 720px);
  min-height: 0;
  overflow: hidden;
}

.filter-config-list-modal__split {
  flex: 1;
  min-height: 0;
  height: 100%;
}

.filter-config-list-modal__search {
  width: 100%;
}

.filter-config-list-modal__grid {
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
</style>
