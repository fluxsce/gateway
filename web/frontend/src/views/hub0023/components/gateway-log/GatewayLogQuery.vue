<template>
  <div :id="gatewayLogQueryRootId" class="gateway-log-query">
    <RsSplitPane
      class="gateway-log-query__split"
      orientation="vertical"
      :panes="splitPanes"
      disabled
    >
      <template #search>
        <div class="gateway-log-query__search">
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
        <div class="gateway-log-query__grid">
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
            @page-change="service.handlePageChange"
            @menu-click="handleMenuClick"
          />
        </div>
      </template>
    </RsSplitPane>

    <BackendLogsDialog
      v-model:visible="detailDialogVisible"
      :trace-id="selectedTraceId"
      :gateway-instance-id="selectedGatewayInstanceId"
    />

    <ResendRequestDialog
      v-model:visible="resendDialogVisible"
      :logs="resendLogs"
      :mount-container-id="gatewayLogQueryRootId"
    />
  </div>
</template>

<script lang="ts" setup>
import { RsSearchForm } from '@/components/form/rs-search'
import { RsGrid, type RsGridExpose } from '@/components/rs-grid'
import { RsSplitPane, type RsSplitPaneItem } from '@/ui'
import { createDomId } from '@/utils/messageUtil'
import { onMounted, ref } from 'vue'
import BackendLogsDialog from '../backed-logs/BackendLogsDialog.vue'
import ResendRequestDialog from '../resend-request/ResendRequestDialog.vue'
import { useGatewayLogPage } from './hooks'

defineOptions({
  name: 'GatewayLogQuery',
})

const splitPanes: RsSplitPaneItem[] = [
  { key: 'search', size: 'auto' },
  { key: 'grid' },
]

// ============= Refs =============

const searchFormRef = ref()
const gridRef = ref<RsGridExpose | null>(null)

/** 本页根节点 id（可作 # 选择器），供重发弹窗 teleport 挂载，避免挂 body 时在多页签下叠到其它标签 */
const gatewayLogQueryRootId = createDomId('hub0023-gateway-log-query')

const {
  service,
  detailDialogVisible,
  selectedTraceId,
  selectedGatewayInstanceId,
  resendDialogVisible,
  resendLogs,
  handleToolbarClick,
  handleMenuClick,
  handleSearch,
} = useGatewayLogPage(gridRef, searchFormRef)

onMounted(() => {
  void service.model.bootstrapDefaultGatewayInstance(searchFormRef)
})

defineExpose({
  searchFormRef,
  gridRef,
  service,
})
</script>

<style lang="scss" scoped>
.gateway-log-query {
  position: relative;
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.gateway-log-query__split {
  flex: 1 1 auto;
  width: 100%;
  height: 100%;
  min-height: 0;
}

.gateway-log-query__search {
  width: 100%;
  box-sizing: border-box;
}

.gateway-log-query__grid {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

:global(.route-name-text) {
  color: var(--g-primary, #7c3aed);
}
</style>
