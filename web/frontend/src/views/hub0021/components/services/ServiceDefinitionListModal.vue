<template>
  <RsDialog
    :open="modalVisible"
    :title="props.title || '选择服务定义'"
    layout="window"
    :width="props.width || 1200"
    :teleport-to="props.to"
    :draggable="true"
    :fullscreenable="true"
    :modal="false"
    :show-overlay="false"
    :close-on-overlay-click="false"
    :show-footer="true"
    :show-confirm="true"
    :show-cancel="true"
    confirm-text="确认选择"
    cancel-text="取消"
    class="hub0021-list-dialog"
    @update:open="handleUpdateVisible"
    @confirm="handleConfirm"
    @cancel="handleCancel"
  >
    <template #body>
      <div class="service-definition-list-modal" :id="model.moduleId">
        <RsSplitPane
          class="service-definition-list-modal__split"
          orientation="vertical"
          :panes="splitPanes"
          disabled
        >
          <template #search>
            <div class="service-definition-list-modal__search">
              <RsSearchForm
                ref="searchFormRef"
                :module-id="model.moduleId"
                v-bind="model.searchFormConfig"
                @search="handleSearch"
                @toolbar-click="handleToolbarClick"
              />
            </div>
          </template>

          <template #grid>
            <div class="service-definition-list-modal__grid">
              <RsGrid
                ref="gridRef"
                :module-id="model.moduleId"
                :data="model.serviceList"
                :loading="model.loading"
                :columns="model.gridConfig.columns"
                :selectable="model.gridConfig.selectable"
                :row-key="model.gridConfig.rowKey"
                height="100%"
                :pagination-config="model.gridConfig.paginationConfig"
                :menu-config="model.gridConfig.menuConfig"
                @page-change="handlePageChange"
                @selection-change="handleCheckboxChange"
              />
            </div>
          </template>
        </RsSplitPane>
      </div>
    </template>
  </RsDialog>
</template>

<script lang="ts" setup>
import { RsSearchForm } from '@/components/form/rs-search'
import { RsGrid, type RsGridExpose } from '@/components/rs-grid'
import { useAppMessage } from '@/composables/useAppMessage'
import { RsDialog, RsSplitPane, type RsSplitPaneItem } from '@/ui'
import { nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { useServiceDefinitionSelectorModel } from './hooks/model'
import { useServiceDefinitionListPage } from './hooks/page'
import type { ServiceDefinitionListModalEmits, ServiceDefinitionListModalProps } from './hooks/types'
import type { ServiceDefinition } from './types'

defineOptions({
  name: 'ServiceDefinitionListModal',
})

const props = withDefaults(defineProps<ServiceDefinitionListModalProps>(), {
  visible: false,
  title: '选择服务定义',
  width: 1200,
  to: undefined,
  gatewayInstanceId: undefined,
  selectedIds: () => [],
  selectedServices: () => [],
})

const emit = defineEmits<ServiceDefinitionListModalEmits>()
const message = useAppMessage()

const splitPanes: RsSplitPaneItem[] = [
  { key: 'search', size: 'auto' },
  { key: 'grid' },
]

const searchFormRef = ref()
const gridRef = ref<RsGridExpose | null>(null)
const selectedServices = ref<ServiceDefinition[]>([])
const isRestoringSelection = ref(false)
const model = useServiceDefinitionSelectorModel()
const gatewayInstanceId = ref<string | undefined>(props.gatewayInstanceId)
const modalVisible = ref(props.visible)

const stopGatewayInstanceIdWatch = watch(() => props.gatewayInstanceId, (newVal) => {
  gatewayInstanceId.value = newVal
})

const {
  handleSearch,
  handlePageChange,
  handleToolbarClick,
} = useServiceDefinitionListPage(model, gatewayInstanceId, gridRef, searchFormRef)

function restoreCheckboxSelection() {
  if (!gridRef.value) return

  isRestoringSelection.value = true
  try {
    const allRows = model.serviceList.value || []
    gridRef.value.clearSelection()

    const ids = new Set(selectedServices.value.map(s => s.serviceDefinitionId).filter(Boolean))
    if (ids.size === 0) return

    const matchedRows = allRows.filter(row => ids.has(row.serviceDefinitionId))
    if (matchedRows.length === 0) return

    gridRef.value.setSelectedRows(matchedRows, true)

    const serviceMap = new Map(selectedServices.value.map(s => [s.serviceDefinitionId, s]))
    matchedRows.forEach(row => serviceMap.set(row.serviceDefinitionId, row))
    selectedServices.value = Array.from(serviceMap.values())
  } finally {
    nextTick(() => {
      isRestoringSelection.value = false
    })
  }
}

function initSelectedFromProps() {
  if (props.selectedServices && props.selectedServices.length > 0) {
    selectedServices.value = [...props.selectedServices]
    return
  }
  if (props.selectedIds && props.selectedIds.length > 0) {
    selectedServices.value = props.selectedIds.map(id => ({
      serviceDefinitionId: id,
    } as ServiceDefinition))
    return
  }
  selectedServices.value = []
}

const stopVisibleWatch = watch(() => props.visible, (newVal) => {
  modalVisible.value = newVal
  if (newVal) {
    initSelectedFromProps()
    nextTick(() => {
      restoreCheckboxSelection()
    })
  }
})

const stopServiceListWatch = watch(
  () => model.serviceList.value,
  () => {
    if (!modalVisible.value) return
    nextTick(() => {
      restoreCheckboxSelection()
    })
  },
)

onBeforeUnmount(() => {
  stopGatewayInstanceIdWatch()
  stopVisibleWatch()
  stopServiceListWatch()
})

const handleCheckboxChange = (selection: ServiceDefinition[]) => {
  if (isRestoringSelection.value) return
  const currentPageIds = new Set(
    (model.serviceList.value || []).map(row => row.serviceDefinitionId),
  )
  const keptOffPage = selectedServices.value.filter(
    s => s.serviceDefinitionId && !currentPageIds.has(s.serviceDefinitionId),
  )
  selectedServices.value = [...keptOffPage, ...selection]
}

const handleUpdateVisible = (value: boolean) => {
  modalVisible.value = value
  emit('update:visible', value)
  if (!value) {
    selectedServices.value = []
    emit('close')
  } else {
    initSelectedFromProps()
    nextTick(() => {
      restoreCheckboxSelection()
    })
    emit('refresh')
  }
}

const handleConfirm = () => {
  if (selectedServices.value.length === 0) {
    message.warning('请至少选择一个服务定义')
    return
  }
  emit('select', selectedServices.value)
  modalVisible.value = false
  emit('update:visible', false)
  emit('close')
}

const handleCancel = () => {
  modalVisible.value = false
  emit('update:visible', false)
  emit('close')
}
</script>

<style scoped>
.service-definition-list-modal {
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  width: 100%;
  height: min(70vh, 720px);
  min-height: 0;
  overflow: hidden;
}

.service-definition-list-modal__split {
  flex: 1;
  min-height: 0;
  height: 100%;
}

.service-definition-list-modal__search {
  width: 100%;
}

.service-definition-list-modal__grid {
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
</style>
