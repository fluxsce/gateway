<template>
  <div class="alert-config-management" :id="htmlId">
    <RsSplitPane
      class="alert-config-management__split"
      orientation="vertical"
      :panes="splitPanes"
      disabled
    >
      <template #search>
        <div class="alert-config-management__search">
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
        <div class="alert-config-management__grid">
          <RsGrid
            ref="gridRef"
            :module-id="service.model.moduleId"
            :data="service.model.configList"
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
      :title="formDialogMode === 'create' ? '新增告警渠道配置' : formDialogMode === 'edit' ? '编辑告警渠道配置' : '查看告警渠道配置详情'"
      :to="`#${htmlId}`"
      :form-fields="service.model.formFields"
      :form-tabs="service.model.formTabs"
      :initial-data="currentEditConfig || undefined"
      :auto-close-on-confirm="false"
      :confirm-loading="service.model.loading.value"
      @submit="handleFormSubmit"
    />

    <AlertTestModal
      v-model:visible="testModalVisible"
      :config="currentTestConfig"
      :to="`#${htmlId}`"
      @close="closeTestModal"
    />
  </div>
</template>

<script lang="ts" setup>
import { RsDataFormModal } from '@/components/form/rs-data'
import { RsSearchForm, type RsSearchFormExpose } from '@/components/form/rs-search'
import { RsGrid, type RsGridColumn, type RsGridExpose } from '@/components/rs-grid'
import { RsSplitPane, RsSwitch, type RsSplitPaneItem } from '@/ui'
import { computed, h, ref } from 'vue'
import { AlertTestModal } from './components'
import { useAlertConfigPage } from './hooks'
import type { AlertConfig } from './types'

defineOptions({
  name: 'AlertConfigManagement',
})

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
  currentEditConfig,
  testModalVisible,
  currentTestConfig,
  handleFormSubmit,
  handleToolbarClick,
  handleMenuClick,
  handleSearch,
  handlePageChange,
  handleToggleStatus,
  closeTestModal,
} = useAlertConfigPage(gridRef, searchFormRef)

/** 固定 HTML id（moduleId 含冒号，不能直接用作 DOM id） */
const htmlId = 'hub0080-alert-config'

/** 启停开关需要页面级回调，在此覆盖 model 列渲染 */
const gridColumns = computed<RsGridColumn<AlertConfig>[]>(() =>
  service.model.gridConfig.columns.map((col) => {
    if (col.key === 'activeFlag') {
      return {
        ...col,
        render: (row: AlertConfig) =>
          h(RsSwitch, {
            modelValue: row.activeFlag === 'Y',
            size: 'ssm',
            'onUpdate:modelValue': () => {
              void handleToggleStatus(row)
            },
          }),
      }
    }
    return col
  }),
)
</script>

<style lang="scss" scoped>
.alert-config-management {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.alert-config-management__split {
  flex: 1 1 auto;
  width: 100%;
  height: 100%;
  min-height: 0;
}

.alert-config-management__search {
  width: 100%;
  box-sizing: border-box;
}

.alert-config-management__grid {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
</style>
