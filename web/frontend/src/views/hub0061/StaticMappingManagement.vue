<template>
  <div class="static-server-management" :id="htmlId">
    <RsSplitPane
      class="static-server-management__split"
      orientation="vertical"
      :panes="splitPanes"
      disabled
    >
      <template #search>
        <div class="static-server-management__search">
          <RsSearchForm
            ref="searchFormRef"
            :module-id="service.model.moduleId"
            v-bind="service.model.searchFormConfig"
            @search="handleSearch"
            @toolbar-click="handleToolbarClick"
          />
        </div>
      </template>

      <template #content>
        <div class="static-server-management__content">
          <div v-if="showStats" class="static-server-management__stats">
            <StaticServerStats :statistics="statistics" />
          </div>

          <div class="static-server-management__grid">
            <RsGrid
              ref="gridRef"
              :module-id="service.model.moduleId"
              :data="service.model.serverList"
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
        </div>
      </template>
    </RsSplitPane>

    <RsDataFormModal
      v-model:visible="formDialogVisible"
      :mode="formDialogMode"
      :title="formDialogMode === 'create' ? '新增静态服务' : formDialogMode === 'edit' ? '编辑静态服务' : '查看静态服务详情'"
      :to="`#${htmlId}`"
      :form-fields="service.model.formFields"
      :form-tabs="service.model.formTabs"
      :initial-data="currentEditServer || undefined"
      :auto-close-on-confirm="false"
      :confirm-loading="service.model.loading.value"
      @submit="handleFormSubmit"
    />

    <StaticNodeListModal
      v-model:visible="nodeDialogVisible"
      :tunnel-static-server-id="currentNodeServer?.tunnelStaticServerId || ''"
      :server-name="currentNodeServer?.serverName || ''"
      :to="`#${htmlId}`"
      @close="closeNodeDialog"
      @refresh="handleRefreshAfterNodeChange"
    />
  </div>
</template>

<script lang="ts" setup>
import { RsDataFormModal } from '@/components/form/rs-data'
import { RsSearchForm, type RsSearchFormExpose } from '@/components/form/rs-search'
import { RsGrid, type RsGridColumn, type RsGridExpose } from '@/components/rs-grid'
import { parseJsonData } from '@/utils/format'
import { RsButton, RsSplitPane, RsSwitch, type RsSplitPaneItem } from '@/ui'
import { computed, h, ref } from 'vue'
import { getStaticServerStats } from './api'
import { StaticNodeListModal } from './components/static-nodes'
import { StaticServerStats } from './components/stats'
import type { StaticServerStats as StaticServerStatsType } from './components/stats/types'
import { useStaticServerPage } from './hooks'
import type { TunnelStaticServer } from './types'

defineOptions({
  name: 'StaticMappingManagement',
})

/** 上方面板随搜索表单高度自适应，下方吃满剩余空间 */
const splitPanes: RsSplitPaneItem[] = [
  { key: 'search', size: 'auto' },
  { key: 'content' },
]

const searchFormRef = ref<RsSearchFormExpose | null>(null)
const gridRef = ref<RsGridExpose | null>(null)

const showStats = ref(true)
const statistics = ref<StaticServerStatsType>({
  totalServers: 0,
  runningServers: 0,
  stoppedServers: 0,
  totalConnections: 0,
  totalBytesReceived: 0,
  totalBytesSent: 0,
})

/**
 * 加载统计数据
 */
const loadStatistics = async () => {
  try {
    const res = await getStaticServerStats()
    if (res.oK) {
      const stats = parseJsonData<StaticServerStatsType>(res)
      if (stats) {
        statistics.value = stats
      }
    }
  } catch (error) {
    console.error('加载统计数据失败:', error)
  }
}

const {
  service,
  formDialogVisible,
  formDialogMode,
  currentEditServer,
  nodeDialogVisible,
  currentNodeServer,
  handleFormSubmit,
  handleToolbarClick,
  handleMenuClick,
  handleSearch: originalHandleSearch,
  handlePageChange,
  handleToggleStatus,
  openNodeDialog,
  closeNodeDialog,
} = useStaticServerPage(gridRef, searchFormRef)

/** 固定 HTML id（moduleId 含冒号，不能直接用作 DOM id） */
const htmlId = 'hub0061-static-server'

/** 交互列（节点数、启停开关）需要页面级回调，在此覆盖 model 列渲染 */
const gridColumns = computed<RsGridColumn<TunnelStaticServer>[]>(() =>
  service.model.gridConfig.columns.map((col) => {
    if (col.key === 'nodeCount') {
      return {
        ...col,
        render: (row: TunnelStaticServer) =>
          h(
            RsButton,
            {
              variant: 'text',
              tone: 'primary',
              size: 'sm',
              onClick: () => openNodeDialog(row),
            },
            () => `${row.nodeCount || 0} 个节点`,
          ),
      }
    }
    if (col.key === 'activeFlag') {
      return {
        ...col,
        render: (row: TunnelStaticServer) =>
          h(RsSwitch, {
            modelValue: row.activeFlag === 'Y',
            size: 'sm',
            'onUpdate:modelValue': () => {
              void handleToggleStatus(row)
            },
          }),
      }
    }
    return col
  }),
)

/**
 * 包装搜索方法，搜索后刷新统计
 */
const handleSearch = async (searchParams?: Record<string, any>) => {
  await originalHandleSearch(searchParams)
  await loadStatistics()
}

/**
 * 节点变化后刷新列表和统计
 */
const handleRefreshAfterNodeChange = async () => {
  await service.loadServerList()
  await loadStatistics()
}
</script>

<style lang="scss" scoped>
.static-server-management {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.static-server-management__split {
  flex: 1 1 auto;
  width: 100%;
  height: 100%;
  min-height: 0;
}

.static-server-management__search {
  width: 100%;
  box-sizing: border-box;
}

.static-server-management__content {
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
}

.static-server-management__stats {
  flex-shrink: 0;
}

.static-server-management__grid {
  flex: 1;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
</style>
