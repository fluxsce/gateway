<template>
  <RsDialog
    :open="modalVisible"
    :title="props.title || '路由断言配置列表'"
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
      <div class="assert-config-list-modal" id="assert-config-list-modal">
        <RsSplitPane
          class="assert-config-list-modal__split"
          orientation="vertical"
          :panes="splitPanes"
          disabled
        >
          <template #search>
            <div class="assert-config-list-modal__search">
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
            <div class="assert-config-list-modal__grid">
              <RsGrid
                ref="gridRef"
                :module-id="service.model.moduleId"
                :data="service.model.assertList"
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
          :title="formDialogMode === 'create' ? '新增断言配置' : formDialogMode === 'edit' ? '编辑断言配置' : '查看断言配置详情'"
          to="#assert-config-list-modal"
          :form-fields="service.model.formFields"
          :form-tabs="service.model.formTabs"
          :initial-data="currentEditAssert || undefined"
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
import { useAssertConfigPage } from './hooks'
import type { AssertConfigListModalEmits, AssertConfigListModalProps } from './hooks/types'

defineOptions({
  name: 'AssertConfigListModal',
})

const props = withDefaults(defineProps<AssertConfigListModalProps>(), {
  visible: false,
  title: '',
  width: 1200,
  to: undefined,
  routeConfigId: '',
})

const emit = defineEmits<AssertConfigListModalEmits>()

const splitPanes: RsSplitPaneItem[] = [
  { key: 'search', size: 'auto' },
  { key: 'grid' },
]

const searchFormRef = ref()
const gridRef = ref<RsGridExpose | null>(null)
const modalVisible = ref(props.visible)
const routeConfigId = ref<string | undefined>(props.routeConfigId)

const stopVisibleWatch = watch(() => props.visible, (newVal) => {
  modalVisible.value = newVal
})

const stopRouteConfigIdWatch = watch(() => props.routeConfigId, (newVal) => {
  routeConfigId.value = newVal
})

onBeforeUnmount(() => {
  stopVisibleWatch()
  stopRouteConfigIdWatch()
})

const {
  service,
  formDialogVisible,
  formDialogMode,
  currentEditAssert,
  handleFormSubmit,
  handleToolbarClick,
  handleMenuClick,
  handleSearch,
} = useAssertConfigPage(routeConfigId, gridRef, searchFormRef)

const handleUpdateVisible = (value: boolean) => {
  modalVisible.value = value
  emit('update:visible', value)
  if (!value) {
    emit('close')
  } else {
    emit('refresh')
    service.loadAssertList()
  }
}

const handleAfterLeave = () => {
  if (!modalVisible.value) {
    formDialogVisible.value = false
    formDialogMode.value = 'create'
    currentEditAssert.value = null
    service.model.assertList.value = []
    service.model.resetPagination()
  }
}
</script>

<style scoped>
.assert-config-list-modal {
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  width: 100%;
  height: min(70vh, 720px);
  min-height: 0;
  overflow: hidden;
}

.assert-config-list-modal__split {
  flex: 1;
  min-height: 0;
  height: 100%;
}

.assert-config-list-modal__search {
  width: 100%;
}

.assert-config-list-modal__grid {
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
</style>
