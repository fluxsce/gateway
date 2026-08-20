<template>
  <RsDialog
    :open="modalVisible"
    :title="props.title || `静态节点管理 - ${props.serverName || ''}`"
    layout="window"
    :width="props.width || 1400"
    :teleport-to="props.to"
    :draggable="true"
    :fullscreenable="true"
    :modal="false"
    :show-overlay="false"
    :close-on-overlay-click="false"
    class="hub0061-static-node-list-dialog"
    @update:open="handleUpdateVisible"
    @after-close="handleAfterLeave"
  >
    <template #body>
      <div class="static-node-list-modal" :id="htmlId">
        <RsSplitPane
          class="static-node-list-modal__split"
          orientation="vertical"
          :panes="splitPanes"
          disabled
        >
          <template #search>
            <div class="static-node-list-modal__search">
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
            <div class="static-node-list-modal__grid">
              <RsGrid
                ref="gridRef"
                :module-id="service.model.moduleId"
                :data="service.model.nodeList"
                :loading="service.model.loading"
                :columns="gridColumns"
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
          :module-id="service.model.moduleId"
          :mode="formDialogMode"
          :title="formDialogMode === 'create' ? '新增静态节点' : formDialogMode === 'edit' ? '编辑静态节点' : '查看静态节点详情'"
          :to="`#${htmlId}`"
          :form-fields="service.model.formFields"
          :form-tabs="service.model.formTabs"
          :initial-data="currentEditNode || undefined"
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
import { RsSearchForm, type RsSearchFormExpose } from '@/components/form/rs-search'
import { RsGrid, type RsGridColumn, type RsGridExpose } from '@/components/rs-grid'
import { RsDialog, RsSplitPane, RsSwitch, type RsSplitPaneItem } from '@/ui'
import { computed, h, onBeforeUnmount, ref, watch } from 'vue'
import { useStaticNodePage } from './hooks'
import type { StaticNodeListModalEmits, StaticNodeListModalProps, TunnelStaticNode } from './hooks/types'

defineOptions({
  name: 'StaticNodeListModal',
})

const splitPanes: RsSplitPaneItem[] = [
  { key: 'search', size: 'auto' },
  { key: 'grid' },
]

const props = withDefaults(defineProps<StaticNodeListModalProps>(), {
  visible: false,
  title: '',
  width: 1400,
  to: undefined,
  tunnelStaticServerId: '',
  serverName: '',
})

const emit = defineEmits<StaticNodeListModalEmits>()

const searchFormRef = ref<RsSearchFormExpose | null>(null)
const gridRef = ref<RsGridExpose | null>(null)
const modalVisible = ref(props.visible)

const stopVisibleWatch = watch(
  () => props.visible,
  (newVal) => {
    modalVisible.value = newVal
  },
)

const tunnelStaticServerId = ref<string>(props.tunnelStaticServerId)

const stopServerIdWatch = watch(
  () => props.tunnelStaticServerId,
  (newVal) => {
    tunnelStaticServerId.value = newVal
  },
)

onBeforeUnmount(() => {
  stopVisibleWatch()
  stopServerIdWatch()
})

const {
  service,
  formDialogVisible,
  formDialogMode,
  currentEditNode,
  handleFormSubmit,
  handleToolbarClick,
  handleMenuClick,
  handleSearch,
  handlePageChange,
  handleToggleStatus,
} = useStaticNodePage(gridRef, tunnelStaticServerId, searchFormRef)

const htmlId = 'hub0061-static-nodes'

/** 启停开关需要页面级回调，在此覆盖 model 列渲染 */
const gridColumns = computed<RsGridColumn<TunnelStaticNode>[]>(() =>
  service.model.gridConfig.columns.map((col) => {
    if (col.key === 'activeFlag') {
      return {
        ...col,
        render: (row: TunnelStaticNode) =>
          h(RsSwitch, {
            modelValue: row.activeFlag === 'Y',
            size: 'sm',
            'onUpdate:modelValue': () => {
              void handleToggleStatus(row)
            },
          }),
      }
    }
    return col
  }),
)

/**
 * 处理模态框可见性变化
 */
const handleUpdateVisible = (value: boolean) => {
  modalVisible.value = value
  emit('update:visible', value)
  if (!value) {
    emit('close')
  } else {
    emit('refresh')
    service.loadNodeList()
  }
}

/**
 * 处理模态框关闭后重置业务状态
 */
const handleAfterLeave = () => {
  if (!modalVisible.value) {
    formDialogVisible.value = false
    formDialogMode.value = 'create'
    currentEditNode.value = null
    service.model.nodeList.value = []
    service.model.resetPagination()
  }
}
</script>

<style scoped>
.static-node-list-modal {
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  width: 100%;
  height: min(70vh, 720px);
  min-height: 0;
  overflow: hidden;
}

.static-node-list-modal__split {
  flex: 1;
  min-height: 0;
  height: 100%;
}

.static-node-list-modal__search {
  width: 100%;
}

.static-node-list-modal__grid {
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
</style>
