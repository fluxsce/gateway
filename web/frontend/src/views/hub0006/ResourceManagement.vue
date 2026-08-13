<template>
  <div class="resource-management" :id="service.model.moduleId">
    <RsSplitPane
      class="resource-management__split"
      orientation="vertical"
      :panes="splitPanes"
      disabled
    >
      <!-- 上部：搜索表单（auto 随内容） -->
      <template #search>
        <div class="resource-management__search">
          <RsSearchForm
            ref="searchFormRef"
            :module-id="service.model.moduleId"
            v-bind="service.model.searchFormConfig"
            @search="handleSearch"
            @toolbar-click="handleToolbarClick"
          />
        </div>
      </template>

      <!-- 下部：表格占满剩余高度 -->
      <template #grid>
        <div class="resource-management__grid">
          <RsGrid
            ref="gridRef"
            :module-id="service.model.moduleId"
            :data="service.model.resourceList"
            :loading="service.model.loading"
            :columns="service.model.gridConfig.columns"
            :selectable="service.model.gridConfig.selectable"
            :row-key="service.model.gridConfig.rowKey"
            height="100%"
            :pagination-config="service.model.gridConfig.paginationConfig"
            :menu-config="service.model.gridConfig.menuConfig"
            :tree-config="service.model.gridConfig.treeConfig"
            @page-change="service.handlePageChange"
            @menu-click="handleMenuClick"
          />
        </div>
      </template>
    </RsSplitPane>

    <!-- 资源对话框（新增/编辑/查看共用） -->
    <RsDataFormModal
      v-model:visible="formDialogVisible"
      :mode="formDialogMode"
      :title="formDialogTitle"
      :to="`#${service.model.moduleId}`"
      :form-fields="service.model.formFields"
      :form-tabs="service.model.formTabs"
      :initial-data="currentEditResource || undefined"
      :auto-close-on-confirm="false"
      :confirm-loading="service.model.loading.value"
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
import { useResourcePage } from './hooks'

defineOptions({
  name: 'ResourceManagement',
})

/** 上方面板内容自适应，下方吃满剩余空间；disabled 禁止拖拽 */
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
  formDialogTitle,
  currentEditResource,
  handleFormSubmit,
  handleToolbarClick,
  handleMenuClick,
  handleSearch,
} = useResourcePage(gridRef, searchFormRef)
</script>

<style scoped>
.resource-management {
  box-sizing: border-box;
  position: relative;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.resource-management__split {
  flex: 1 1 auto;
  width: 100%;
  height: 100%;
  min-height: 0;
}

.resource-management__search {
  width: 100%;
}

.resource-management__grid {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
</style>
