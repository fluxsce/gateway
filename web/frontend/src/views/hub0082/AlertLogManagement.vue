<template>
  <div class="alert-log-management" :id="htmlId">
    <RsSplitPane
      class="alert-log-management__split"
      orientation="vertical"
      :panes="splitPanes"
      disabled
    >
      <template #search>
        <div class="alert-log-management__search">
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
        <div class="alert-log-management__grid">
          <RsGrid
            ref="gridRef"
            :module-id="service.model.moduleId"
            :data="service.model.logList"
            :loading="service.model.loading"
            :columns="service.model.gridConfig.columns"
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

    <AlertLogDetailDialog
      v-model:visible="viewDialogVisible"
      :alert-log-id="selectedAlertLogId"
      :to="`#${htmlId}`"
    />
  </div>
</template>

<script lang="ts" setup>
import { RsSearchForm, type RsSearchFormExpose } from '@/components/form/rs-search'
import { RsGrid, type RsGridExpose } from '@/components/rs-grid'
import { RsSplitPane, type RsSplitPaneItem } from '@/ui'
import { ref } from 'vue'
import { AlertLogDetailDialog } from './components'
import { useAlertLogPage } from './hooks'

defineOptions({
  name: 'AlertLogManagement',
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
  viewDialogVisible,
  selectedAlertLogId,
  handleToolbarClick,
  handleMenuClick,
  handleSearch,
  handlePageChange,
} = useAlertLogPage(gridRef, searchFormRef)

/** 固定 HTML id（moduleId 含冒号，不能直接用作 DOM id） */
const htmlId = 'hub0082-alert-log'
</script>

<style lang="scss" scoped>
.alert-log-management {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.alert-log-management__split {
  flex: 1 1 auto;
  width: 100%;
  height: 100%;
  min-height: 0;
}

.alert-log-management__search {
  width: 100%;
  box-sizing: border-box;
}

.alert-log-management__grid {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
</style>
