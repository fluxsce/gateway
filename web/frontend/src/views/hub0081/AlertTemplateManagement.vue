<template>
  <div class="alert-template-management" :id="htmlId">
    <RsSplitPane
      class="alert-template-management__split"
      orientation="vertical"
      :panes="splitPanes"
      disabled
    >
      <template #search>
        <div class="alert-template-management__search">
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
        <div class="alert-template-management__grid">
          <RsGrid
            ref="gridRef"
            :module-id="service.model.moduleId"
            :data="service.model.templateList"
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
      :mode="formDialogMode"
      :title="formDialogMode === 'create' ? '新增预警模板' : formDialogMode === 'edit' ? '编辑预警模板' : '查看预警模板'"
      :to="`#${htmlId}`"
      :form-fields="service.model.formFields"
      :form-tabs="service.model.formTabs"
      :initial-data="currentEditTemplate || undefined"
      :auto-close-on-confirm="false"
      :confirm-loading="service.model.loading.value"
      @submit="handleFormSubmit"
    />
  </div>
</template>

<script lang="ts" setup>
import { RsDataFormModal } from '@/components/form/rs-data'
import { RsSearchForm, type RsSearchFormExpose } from '@/components/form/rs-search'
import { RsGrid, type RsGridColumn, type RsGridExpose } from '@/components/rs-grid'
import { RsSplitPane, RsSwitch, type RsSplitPaneItem } from '@/ui'
import { computed, h, ref } from 'vue'
import { useAlertTemplatePage } from './hooks'
import type { AlertTemplate } from './types'

defineOptions({ name: 'AlertTemplateManagement' })

/** 上方搜索区随内容自适应，下方表格占满剩余高度 */
const splitPanes: RsSplitPaneItem[] = [
  { key: 'search', size: 'auto' },
  { key: 'grid' },
]

const searchFormRef = ref<RsSearchFormExpose | null>(null)
const gridRef = ref<RsGridExpose | null>(null)

const {
  service,
  formDialogVisible,
  formDialogMode,
  currentEditTemplate,
  handleFormSubmit,
  handleToolbarClick,
  handleMenuClick,
  handleSearch,
  handlePageChange,
} = useAlertTemplatePage(gridRef, searchFormRef)

/** 固定 HTML id（moduleId 含冒号，不能直接用作 DOM id） */
const htmlId = 'hub0081-alert-template'

/** 启停开关需要页面级回调，在此覆盖 model 列渲染 */
const gridColumns = computed<RsGridColumn<AlertTemplate>[]>(() =>
  service.model.gridConfig.columns.map((col) => {
    if (col.key === 'activeFlag') {
      return {
        ...col,
        render: (row: AlertTemplate) =>
          h(RsSwitch, {
            modelValue: row.activeFlag === 'Y',
            size: 'sm',
            'onUpdate:modelValue': () => {
              void handleToggleActive(row)
            },
          }),
      }
    }
    return col
  }),
)

const handleToggleActive = async (row: AlertTemplate) => {
  const newFlag = row.activeFlag === 'Y' ? 'N' : 'Y'
  await service.editTemplate(row.templateName, { ...row, activeFlag: newFlag })
}
</script>

<style lang="scss" scoped>
.alert-template-management {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.alert-template-management__split {
  flex: 1 1 auto;
  width: 100%;
  height: 100%;
  min-height: 0;
}

.alert-template-management__search {
  width: 100%;
  box-sizing: border-box;
}

.alert-template-management__grid {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
</style>
