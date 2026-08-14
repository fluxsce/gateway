<template>
  <div class="service-center-instance-manager" :id="service.model.moduleId">
    <RsSplitPane
      class="service-center-instance-manager__split"
      orientation="vertical"
      :panes="splitPanes"
      disabled
    >
      <template #search>
        <div class="service-center-instance-manager__search">
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
        <div class="service-center-instance-manager__grid">
          <RsGrid
            ref="gridRef"
            :module-id="service.model.moduleId"
            :data="service.model.instanceList"
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
      :title="formDialogMode === 'create' ? '新增实例' : formDialogMode === 'edit' ? '编辑实例' : '查看实例详情'"
      :to="`#${service.model.moduleId}`"
      :form-fields="service.model.instanceFormConfig.fields"
      :form-tabs="service.model.instanceFormConfig.tabs"
      :initial-data="currentEditInstance || undefined"
      :auto-close-on-confirm="false"
      :confirm-loading="submitting"
      @submit="handleFormSubmit"
    />
  </div>
</template>

<script lang="ts" setup>
import { RsDataFormModal } from '@/components/form/rs-data'
import { RsSearchForm } from '@/components/form/rs-search'
import { RsGrid, type RsGridExpose } from '@/components/rs-grid'
import { RsSplitPane, type RsSplitPaneItem } from '@/ui'
import { ref } from 'vue'
import { useServiceCenterInstancePage } from './hooks'

defineOptions({
  name: 'ServiceCenterInstanceManager',
})

/** 上方搜索区随内容自适应，下方表格占满剩余高度 */
const splitPanes: RsSplitPaneItem[] = [
  { key: 'search', size: 'auto' },
  { key: 'grid' },
]

const searchFormRef = ref()
const gridRef = ref<RsGridExpose | null>(null)

const {
  service,
  formDialogVisible,
  formDialogMode,
  currentEditInstance,
  submitting,
  handleFormSubmit,
  handleToolbarClick,
  handleMenuClick,
  handleSearch,
} = useServiceCenterInstancePage(gridRef, searchFormRef)
</script>

<style lang="scss" scoped>
.service-center-instance-manager {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.service-center-instance-manager__split {
  flex: 1 1 auto;
  width: 100%;
  height: 100%;
  min-height: 0;
}

.service-center-instance-manager__search {
  width: 100%;
  box-sizing: border-box;
}

.service-center-instance-manager__grid {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
</style>
