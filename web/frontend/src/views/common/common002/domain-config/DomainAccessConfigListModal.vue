<template>
  <RsDialog
    :open="modalVisible"
    :title="props.title || '域名访问控制配置列表'"
    layout="window"
    :width="props.width || 1200"
    :teleport-to="props.to"
    :draggable="true"
    :fullscreenable="true"
    :modal="false"
    :show-overlay="false"
    :close-on-overlay-click="false"
    class="common002-list-dialog"
    @update:open="handleUpdateVisible"
    @after-close="handleAfterLeave"
  >
    <template #body>
      <div class="domain-access-config-list-modal" id="domain-access-config-list-modal">
        <RsSplitPane
          class="domain-access-config-list-modal__split"
          orientation="vertical"
          :panes="splitPanes"
          disabled
        >
          <template #search>
            <div class="domain-access-config-list-modal__search">
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
            <div class="domain-access-config-list-modal__grid">
              <RsGrid
                ref="gridRef"
                :module-id="service.model.moduleId"
                :data="service.model.configList"
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
          :title="formDialogMode === 'create' ? '新增域名访问控制配置' : formDialogMode === 'edit' ? '编辑域名访问控制配置' : '查看域名访问控制配置详情'"
          to="#domain-access-config-list-modal"
          :form-fields="service.model.formFields"
          :initial-data="currentEditConfig || undefined"
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
import { ref, watch } from 'vue'
import { useDomainAccessConfigPage } from './hooks'
import type { DomainAccessConfigListModalEmits, DomainAccessConfigListModalProps } from './hooks/types'

defineOptions({
  name: 'DomainAccessConfigListModal',
})

const props = withDefaults(defineProps<DomainAccessConfigListModalProps>(), {
  visible: false,
  title: '域名访问控制配置列表',
  width: 1200,
  to: undefined,
  securityConfigId: undefined,
})

const emit = defineEmits<DomainAccessConfigListModalEmits>()

const splitPanes: RsSplitPaneItem[] = [
  { key: 'search', size: 'auto' },
  { key: 'grid' },
]

const searchFormRef = ref()
const gridRef = ref<RsGridExpose | null>(null)
const moduleIdRef = ref<string>(props.moduleId)
const modalVisible = ref(props.visible)

watch(() => props.visible, (newVal) => {
  modalVisible.value = newVal
})

watch(() => props.moduleId, (newVal) => {
  moduleIdRef.value = newVal
})

const securityConfigId = ref<string | undefined>(props.securityConfigId)

watch(() => props.securityConfigId, (newVal) => {
  securityConfigId.value = newVal
})

const {
  service,
  formDialogVisible,
  formDialogMode,
  currentEditConfig,
  handleFormSubmit,
  handleToolbarClick,
  handleMenuClick,
  handleSearch,
} = useDomainAccessConfigPage(moduleIdRef, gridRef, securityConfigId, searchFormRef)

const handleUpdateVisible = (value: boolean) => {
  modalVisible.value = value
  emit('update:visible', value)
  if (!value) {
    emit('close')
  } else {
    emit('refresh')
  }
}

const handleAfterLeave = () => {
  if (!modalVisible.value) {
    formDialogVisible.value = false
    formDialogMode.value = 'create'
    currentEditConfig.value = null
    service.model.configList.value = []
    service.model.resetPagination()
  }
}
</script>

<style scoped>
.domain-access-config-list-modal {
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  width: 100%;
  height: min(70vh, 720px);
  min-height: 0;
  overflow: hidden;
}

.domain-access-config-list-modal__split {
  flex: 1;
  min-height: 0;
  height: 100%;
}

.domain-access-config-list-modal__search {
  width: 100%;
}

.domain-access-config-list-modal__grid {
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
</style>
