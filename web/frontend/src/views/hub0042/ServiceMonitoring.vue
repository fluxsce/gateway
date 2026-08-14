<template>
  <div class="service-monitoring" :id="service.model.moduleId">
    <RsSplitPane
      class="service-monitoring__outer"
      orientation="vertical"
      :panes="outerPanes"
    >
      <template #namespace>
        <div class="service-monitoring__pane">
          <div class="service-monitoring__header">
            <h3>命名空间列表</h3>
          </div>
          <div class="service-monitoring__body">
            <NamespaceList
              ref="namespaceListRef"
              moduleId="hub0042-namespace"
              :show-dialog="true"
              :auto-load="true"
              @row-click="handleNamespaceRowClick"
              @namespace-select="handleNamespaceSelect"
            />
          </div>
        </div>
      </template>

      <template #services>
        <div class="service-monitoring__pane">
          <div class="service-monitoring__header">
            <h3>服务列表</h3>
            <span v-if="selectedNamespace" class="service-monitoring__ns">
              当前命名空间: {{ selectedNamespace.namespaceName }} ({{ selectedNamespace.namespaceId }})
            </span>
          </div>
          <RsSplitPane
            class="service-monitoring__inner"
            orientation="vertical"
            :panes="innerPanes"
            disabled
          >
            <template #search>
              <div class="service-monitoring__search">
                <RsSearchForm
                  ref="serviceSearchFormRef"
                  :module-id="service.model.moduleId"
                  v-bind="service.model.searchFormConfig"
                  @search="handleServiceSearch"
                  @toolbar-click="handleServiceToolbarClick"
                />
              </div>
            </template>

            <template #grid>
              <div class="service-monitoring__grid">
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
            </template>
          </RsSplitPane>
        </div>
      </template>
    </RsSplitPane>

    <RsDataFormModal
      v-model:visible="serviceFormDialogVisible"
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
import { RsGrid, type RsGridExpose } from '@/components/rs-grid'
import { RsSplitPane, type RsSplitPaneItem } from '@/ui'
import { ref } from 'vue'
import { NamespaceList } from '../hub0041/components'
import type { Namespace } from '../hub0041/types'
import { useServicePage } from './hooks'

defineOptions({
  name: 'ServiceMonitoring',
})

const outerPanes: RsSplitPaneItem[] = [
  { key: 'namespace', size: 35, min: 20 },
  { key: 'services' },
]

const innerPanes: RsSplitPaneItem[] = [
  { key: 'search', size: 'auto' },
  { key: 'grid' },
]

const namespaceListRef = ref()
const serviceSearchFormRef = ref()
const serviceGridRef = ref<RsGridExpose | null>(null)
const selectedNamespace = ref<Namespace | null>(null)

const {
  service,
  formDialogVisible: serviceFormDialogVisible,
  formDialogMode: serviceFormDialogMode,
  currentEditService,
  submitting: serviceSubmitting,
  handleFormSubmit: handleServiceFormSubmitBase,
  handleToolbarClick: handleServiceToolbarClickBase,
  handleMenuClick: handleServiceMenuClick,
  handleSearch: handleServiceSearch,
} = useServicePage(serviceGridRef, serviceSearchFormRef)

/**
 * 工具栏点击：把当前选中的命名空间传给 page hook，避免与 RsSearchForm 的 formData 签名冲突。
 */
const handleServiceToolbarClick = (key: string) => {
  handleServiceToolbarClickBase(key, selectedNamespace.value)
}

/**
 * 命名空间行点击 - 选择命名空间并加载服务列表
 */
const handleNamespaceRowClick = async (row: Namespace) => {
  if (!row) return
  selectedNamespace.value = row
  if (serviceSearchFormRef.value?.setFormData) {
    const current = serviceSearchFormRef.value.getFormData?.() || {}
    serviceSearchFormRef.value.setFormData({ ...current, namespaceId: row.namespaceId })
  }
  await service.loadServices({ namespaceId: row.namespaceId })
}

/**
 * 命名空间选择变化
 */
const handleNamespaceSelect = (namespace: Namespace | null) => {
  selectedNamespace.value = namespace
}

/**
 * 服务分页变化处理
 */
const handleServicePageChange = (params: { currentPage: number; pageSize: number }) => {
  service.handlePageChange(params.currentPage, params.pageSize)
}

/**
 * 服务表单提交（自动填充命名空间ID）
 */
const handleServiceFormSubmit = (formData?: Record<string, any>) => {
  if (!formData) return
  if (selectedNamespace.value && !formData.namespaceId) {
    formData.namespaceId = selectedNamespace.value.namespaceId
  }
  handleServiceFormSubmitBase(formData)
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

.service-monitoring__outer,
.service-monitoring__inner {
  width: 100%;
  height: 100%;
  min-height: 0;
}

.service-monitoring__pane {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.service-monitoring__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--g-space-sm);
  border-bottom: 1px solid var(--g-border-color, var(--rs-border));
  flex-shrink: 0;

  h3 {
    margin: 0;
    font-size: 16px;
    font-weight: 500;
  }
}

.service-monitoring__ns {
  font-size: 14px;
  color: var(--g-text-color-secondary, var(--rs-text-secondary));
}

.service-monitoring__body {
  flex: 1;
  min-height: 0;
  overflow: hidden;
}

.service-monitoring__search {
  width: 100%;
  box-sizing: border-box;
}

.service-monitoring__grid {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
</style>
