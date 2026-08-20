<template>
  <div class="audit-log-management" :id="htmlId">
    <RsSplitPane
      class="audit-log-management__split"
      orientation="vertical"
      :panes="splitPanes"
      disabled
    >
      <template #search>
        <div class="audit-log-management__search">
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
        <div class="audit-log-management__grid">
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

    <AuditLogDetailDialog
      v-model:visible="viewDialogVisible"
      :audit-id="selectedAuditId"
      :to="`#${htmlId}`"
    />

    <GExport
      v-model:visible="exportVisible"
      module-id="hub0004"
      :url="exportUrl"
      :params="exportParams"
      filename="audit-log"
      :dialog-title="t('common.export')"
      @error="(e) => message.error(e.message || t('message.exportFailed'))"
    />
  </div>
</template>

<script lang="ts" setup>
import { moduleApiPrefix, requestPathHelper } from '@/api/requestPath'
import { RsSearchForm, type RsSearchFormExpose } from '@/components/form/rs-search'
import { GExport } from '@/components/gexport-import'
import { RsGrid, type RsGridExpose } from '@/components/rs-grid'
import { useAppMessage } from '@/composables/useAppMessage'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { RsSplitPane, type RsSplitPaneItem } from '@/ui'
import { ref } from 'vue'
import { AuditLogDetailDialog } from './components'
import { useAuditLogPage } from './hooks'

defineOptions({
  name: 'AuditLogManagement',
})

/** 上方搜索区随内容自适应，下方表格占满剩余高度 */
const splitPanes: RsSplitPaneItem[] = [
  { key: 'search', size: 'auto' },
  { key: 'grid' },
]

const searchFormRef = ref<RsSearchFormExpose | null>(null)
const gridRef = ref<RsGridExpose | null>(null)
const message = useAppMessage()
const { t } = useModuleI18n('hub0004')
const exportUrl = requestPathHelper.join(moduleApiPrefix('hub0004'), 'exportAuditLogs')

const {
  service,
  viewDialogVisible,
  selectedAuditId,
  handleMenuClick,
  handleSearch,
  handlePageChange,
  handleToolbarClick,
  exportVisible,
  exportParams,
} = useAuditLogPage(searchFormRef)

/** 固定 HTML id（moduleId 含冒号时不能直接用作 DOM id） */
const htmlId = 'hub0004-audit-log'
</script>

<style lang="scss" scoped>
.audit-log-management {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.audit-log-management__split {
  flex: 1 1 auto;
  width: 100%;
  height: 100%;
  min-height: 0;
}

.audit-log-management__search {
  width: 100%;
  box-sizing: border-box;
}

.audit-log-management__grid {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
</style>
