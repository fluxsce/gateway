<template>
  <RsDialog
    :open="modalVisible"
    :title="props.title || '服务节点管理'"
    layout="window"
    :width="props.width || 1200"
    :teleport-to="props.to"
    :draggable="true"
    :fullscreenable="true"
    :modal="false"
    :show-overlay="false"
    :close-on-overlay-click="false"
    class="hub0022-node-list-dialog"
    @update:open="handleUpdateVisible"
    @after-close="handleAfterLeave"
  >
    <template #body>
      <div class="service-node-list-modal" id="hub0022-service-node-list">
        <RsSplitPane
          class="service-node-list-modal__split"
          orientation="vertical"
          :panes="splitPanes"
          disabled
        >
          <template #search>
            <div class="service-node-list-modal__search">
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
            <div class="service-node-list-modal__grid">
              <RsGrid
                ref="gridRef"
                :module-id="service.model.moduleId"
                :data="service.model.nodeList"
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
          :mode="formDialogMode"
          :title="
            formDialogMode === 'create'
              ? '新增服务节点'
              : formDialogMode === 'edit'
                ? '编辑服务节点'
                : '查看服务节点详情'
          "
          to="#hub0022-service-node-list"
          :form-fields="service.model.nodeFormConfig.fields"
          :form-tabs="service.model.nodeFormConfig.tabs"
          :initial-data="getFormInitialData()"
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
import { useServiceNodePage } from './hooks'
import type { ServiceNodeListModalEmits, ServiceNodeListModalProps } from './types'

defineOptions({
  name: 'ServiceNodeListModal',
})

const splitPanes: RsSplitPaneItem[] = [
  { key: 'search', size: 'auto' },
  { key: 'grid' },
]

const props = withDefaults(defineProps<ServiceNodeListModalProps>(), {
  visible: false,
  title: '服务节点管理',
  width: 1200,
  to: undefined,
  serviceDefinitionId: undefined,
})

const emit = defineEmits<ServiceNodeListModalEmits>()

const searchFormRef = ref()
const gridRef = ref<RsGridExpose | null>(null)
const modalVisible = ref(props.visible)
const serviceDefinitionId = ref<string | undefined>(props.serviceDefinitionId)

const stopVisibleWatch = watch(
  () => props.visible,
  (newVal) => {
    modalVisible.value = newVal
  }
)

const stopServiceDefinitionIdWatch = watch(
  () => props.serviceDefinitionId,
  (newVal) => {
    serviceDefinitionId.value = newVal
    if (modalVisible.value && newVal) {
      service.handleRefresh()
    }
  }
)

onBeforeUnmount(() => {
  stopVisibleWatch()
  stopServiceDefinitionIdWatch()
})

const {
  service,
  formDialogVisible,
  formDialogMode,
  getFormInitialData,
  handleFormSubmit,
  handleToolbarClick,
  handleMenuClick,
  handleSearch,
} = useServiceNodePage(gridRef, serviceDefinitionId, searchFormRef)

/**
 * 处理对话框可见性变化
 */
const handleUpdateVisible = (value: boolean) => {
  modalVisible.value = value
  emit('update:visible', value)
  if (!value) {
    emit('close')
  } else {
    emit('refresh')
    if (serviceDefinitionId.value) {
      service.handleRefresh()
    }
  }
}

/**
 * 对话框关闭后重置业务状态
 */
const handleAfterLeave = () => {
  if (!modalVisible.value) {
    formDialogVisible.value = false
    formDialogMode.value = 'create'
    service.model.nodeList.value = []
    service.model.resetPagination()
  }
}
</script>

<style scoped>
.service-node-list-modal {
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  width: 100%;
  height: min(70vh, 720px);
  min-height: 0;
  overflow: hidden;
}

.service-node-list-modal__split {
  flex: 1;
  min-height: 0;
  height: 100%;
}

.service-node-list-modal__search {
  width: 100%;
}

.service-node-list-modal__grid {
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
</style>
