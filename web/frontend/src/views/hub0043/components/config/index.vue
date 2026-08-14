<template>
  <div class="config-management" :id="service.model.moduleId">
    <!-- 列表视图 -->
    <template v-if="currentView === 'list'">
      <RsSplitPane
        class="config-management__split"
        orientation="vertical"
        :panes="splitPanes"
        disabled
      >
        <template #search>
          <div class="config-management__search">
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
          <div class="config-management__grid">
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
    </template>

    <!-- 表单视图（新增/编辑/查看共用） -->
    <template v-else-if="currentView === 'form'">
      <div class="config-form-view">
        <div class="config-form-header">
          <RsButton size="sm" icon="arrow-left" @click="handleBackToList">
            返回列表
          </RsButton>
        </div>

        <RsDataForm
          ref="formRef"
          :mode="formDialogMode"
          :form-fields="service.model.configFormConfig.fields"
          :initial-data="currentEditConfig || undefined"
          :show-footer="true"
          :show-submit="formDialogMode !== 'view'"
          :submit-text="formDialogMode === 'create' ? '发布' : '保存'"
          :submit-loading="submitting"
          @submit="handleFormSubmit"
        />
      </div>
    </template>
  </div>
</template>

<script lang="ts" setup>
import { RsDataForm, type RsDataFormExpose } from '@/components/form/rs-data'
import { RsSearchForm, type RsSearchFormExpose } from '@/components/form/rs-search'
import { RsGrid, type RsGridExpose } from '@/components/rs-grid'
import { RsButton, RsSplitPane, type RsSplitPaneItem } from '@/ui'
import { ref } from 'vue'
import type { Config } from '../../types'
import { useConfigPage } from './hooks'

defineOptions({
  name: 'ConfigManagement',
})

interface Props {
  /** 是否显示历史按钮（用于跳转到历史页面） */
  showHistoryButton?: boolean
}

interface Emits {
  /** 查看历史事件 */
  (e: 'view-history', config: Config): void
}

const props = withDefaults(defineProps<Props>(), {
  showHistoryButton: true,
})

const emit = defineEmits<Emits>()

/** 上方搜索区随内容自适应，下方表格占满剩余高度 */
const splitPanes: RsSplitPaneItem[] = [
  { key: 'search', size: 'auto' },
  { key: 'grid' },
]

const searchFormRef = ref<RsSearchFormExpose | null>(null)
const gridRef = ref<RsGridExpose | null>(null)
const formRef = ref<RsDataFormExpose | null>(null)

const {
  service,
  currentView,
  formDialogMode,
  currentEditConfig,
  submitting,
  handleFormSubmit,
  handleToolbarClick,
  handleMenuClick,
  handleSearch,
  handleBackToList,
} = useConfigPage(
  gridRef,
  searchFormRef,
  formRef,
  {
    useViewMode: true,
    onHistoryClick: (config) => {
      if (props.showHistoryButton) {
        emit('view-history', config)
      }
    },
  }
)

defineExpose({
  refresh: service.handleRefresh,
  loadConfigs: service.loadConfigs,
})
</script>

<style lang="scss" scoped>
.config-management {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.config-management__split {
  flex: 1 1 auto;
  width: 100%;
  height: 100%;
  min-height: 0;
}

.config-management__search {
  width: 100%;
  box-sizing: border-box;
}

.config-management__grid {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.config-form-view {
  height: 100%;
  width: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.config-form-header {
  display: flex;
  align-items: center;
  padding: var(--g-space-sm) var(--g-space-md);
  border-bottom: 1px solid var(--g-border-primary);
  background-color: var(--g-bg-color);
}

.config-form-view :deep(.rs-data-form) {
  flex: 1;
  overflow: auto;
  padding: var(--g-space-md);
}
</style>
