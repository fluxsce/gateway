<template>
  <div class="tunnel-server-list" :id="htmlId">
    <RsSplitPane
      class="tunnel-server-list__split"
      orientation="vertical"
      :panes="splitPanes"
      disabled
    >
      <template #search>
        <div class="tunnel-server-list__search">
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
        <div class="tunnel-server-list__body">
          <div v-if="showStats" class="tunnel-server-list__stats">
            <TunnelServerStats :statistics="statistics" />
          </div>
          <div class="tunnel-server-list__grid">
            <RsGrid
              ref="gridRef"
              :module-id="service.model.moduleId"
              :data="service.model.tunnelServerList"
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
        </div>
      </template>
    </RsSplitPane>

    <RsDataFormModal
      v-model:visible="formDialogVisible"
      :module-id="service.model.moduleId"
      :mode="formDialogMode"
      :title="
        formDialogMode === 'create'
          ? '新增隧道服务器'
          : formDialogMode === 'edit'
            ? '编辑隧道服务器'
            : '查看隧道服务器详情'
      "
      :to="`#${htmlId}`"
      :form-fields="service.model.formFields"
      :form-tabs="service.model.formTabs"
      :initial-data="currentEditServer || undefined"
      :auto-close-on-confirm="false"
      :confirm-loading="submitting"
      @submit="handleFormSubmit"
    />
  </div>
</template>

<script lang="ts" setup>
import { RsDataFormModal } from '@/components/form/rs-data'
import { RsSearchForm, type RsSearchFormExpose } from '@/components/form/rs-search'
import { RsGrid, type RsGridExpose } from '@/components/rs-grid'
import { RsSplitPane, type RsSplitPaneItem } from '@/ui'
import { isApiSuccess, parseJsonData } from '@/utils/format'
import { onMounted, ref } from 'vue'
import * as tunnelServerApi from '../../api'
import type { TunnelServerStats as TunnelServerStatsType } from '../../types'
import TunnelServerStats from '../stats/TunnelServerStats.vue'
import { useTunnelServerPage } from './hooks'

defineOptions({
  name: 'TunnelServerList',
})

const splitPanes: RsSplitPaneItem[] = [{ key: 'search', size: 'auto' }, { key: 'grid' }]

const searchFormRef = ref<RsSearchFormExpose | null>(null)
const gridRef = ref<RsGridExpose | null>(null)

const showStats = ref(true)
const statistics = ref<TunnelServerStatsType>({
  totalServers: 0,
  runningServers: 0,
  stoppedServers: 0,
  errorServers: 0,
  totalClients: 0,
  totalConnections: 0,
})

const getStatistics = async () => {
  try {
    const response = await tunnelServerApi.getTunnelServerStats()
    if (isApiSuccess(response)) {
      const data = parseJsonData<TunnelServerStatsType>(response, {
        totalServers: 0,
        runningServers: 0,
        stoppedServers: 0,
        errorServers: 0,
        totalClients: 0,
        totalConnections: 0,
      })
      statistics.value = data
    }
  } catch {
    // 失败时保持当前状态
  }
}

const {
  service,
  formDialogVisible,
  formDialogMode,
  currentEditServer,
  submitting,
  handleFormSubmit,
  handleToolbarClick,
  handleMenuClick,
  handleSearch,
} = useTunnelServerPage(gridRef, searchFormRef)

const htmlId = 'hub0060'

onMounted(() => {
  if (showStats.value) {
    getStatistics()
  }
})
</script>

<style lang="scss" scoped>
.tunnel-server-list {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.tunnel-server-list__split {
  flex: 1 1 auto;
  width: 100%;
  height: 100%;
  min-height: 0;
}

.tunnel-server-list__search {
  width: 100%;
  box-sizing: border-box;
}

.tunnel-server-list__body {
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
}

.tunnel-server-list__stats {
  flex-shrink: 0;
}

.tunnel-server-list__grid {
  box-sizing: border-box;
  flex: 1;
  width: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
</style>
