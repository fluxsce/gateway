<template>
  <div class="static-server-management" :id="htmlId">
    <GPane direction="vertical" :default-size="0.12" :min="0.1" :max="0.5">
      <!-- 上部：搜索表单 -->
      <template #1>
        <search-form
          ref="searchFormRef"
          :module-id="service.model.moduleId"
          v-bind="service.model.searchFormConfig"
          @search="handleSearch"
          @toolbar-click="handleToolbarClick"
        />
      </template>

      <!-- 下部：统计面板 + 数据表格 -->
      <template #2>
        <div class="bottom-section">
          <!-- 统计面板 -->
          <div class="stats-section" v-if="showStats">
            <StaticServerStats :statistics="statistics" />
          </div>

          <!-- 数据表格 -->
          <div class="grid-section">
            <g-grid
              ref="gridRef"
              :module-id="service.model.moduleId"
              :data="service.model.serverList"
              :loading="service.model.loading"
              v-bind="service.model.gridConfig"
              @page-change="handlePageChange"
              @menu-click="({ code, row }) => handleMenuClick({ menu: { code }, row })"
            >
              <!-- 监听地址自定义渲染 -->
              <template #listenAddress="{ row }">
                <span class="text-primary font-mono">{{ row.listenAddress }}:{{ row.listenPort }}</span>
              </template>

              <!-- 服务类型自定义渲染 -->
              <template #serverType="{ row }">
                <n-tag :type="row.serverType === 'tcp' ? 'primary' : 'info'" size="small">
                  {{ service.model.getServerTypeLabel(row.serverType) }}
                </n-tag>
              </template>

              <!-- 服务状态自定义渲染 -->
              <template #serverStatus="{ row }">
                <n-tag :type="service.model.getServerStatusTagType(row.serverStatus)" size="small">
                  {{ service.model.getServerStatusLabel(row.serverStatus) }}
                </n-tag>
              </template>

              <!-- 节点数自定义渲染 -->
              <template #nodeCount="{ row }">
                <n-button 
                  text 
                  type="primary" 
                  @click="openNodeDialog(row)"
                >
                  {{ row.nodeCount || 0 }} 个节点
                </n-button>
              </template>

              <!-- 状态自定义渲染 -->
              <template #activeFlag="{ row }">
                <n-switch
                  :value="row.activeFlag === 'Y'"
                  @update:value="() => handleToggleStatus(row)"
                  size="small"
                />
              </template>
            </g-grid>
          </div>
        </div>
      </template>
    </GPane>

    <!-- 静态服务对话框（新增/编辑/查看共用） -->
    <GdataFormModal
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

    <!-- 节点管理对话框 -->
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
import GdataFormModal from '@/components/form/data/GDataFormModal.vue'
import SearchForm from '@/components/form/search/SearchForm.vue'
import { GPane } from '@/components/gpane'
import { GGrid } from '@/components/grid'
import { parseJsonData } from '@/utils/format'
import { NButton, NSwitch, NTag } from 'naive-ui'
import { ref } from 'vue'
import { getStaticServerStats } from './api'
import { StaticNodeListModal } from './components/static-nodes'
import { StaticServerStats } from './components/stats'
import type { StaticServerStats as StaticServerStatsType } from './components/stats/types'
import { useStaticServerPage } from './hooks'

// 定义组件名称
defineOptions({
  name: 'StaticMappingManagement'
})

// ============= Refs =============

const searchFormRef = ref()
const gridRef = ref()

// ============= 统计面板 =============

const showStats = ref(true)
const statistics = ref<StaticServerStatsType>({
  totalServers: 0,
  runningServers: 0,
  stoppedServers: 0,
  totalConnections: 0,
  totalBytesReceived: 0,
  totalBytesSent: 0,
})

// 加载统计数据
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

// ============= 页面级 Hook（包含服务与对话框、事件处理） =============

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

// ============= HTML ID（用于 DOM，符合 HTML 规范） =============

// 固定的 HTML id（符合 HTML 规范，无特殊字符）
// 注意：权限校验仍使用原始 moduleId（service.model.moduleId）
const htmlId = 'hub0061-static-server'

// ============= 事件处理 =============

// 包装搜索方法，搜索后刷新统计
const handleSearch = async (searchParams?: Record<string, any>) => {
  await originalHandleSearch(searchParams)
  await loadStatistics()
}

// 节点变化后刷新列表和统计
const handleRefreshAfterNodeChange = async () => {
  await service.loadServerList()
  await loadStatistics()
}
</script>

<style lang="scss" scoped>
.static-server-management {
  width: 100%;
  height: 100%;
  overflow: hidden;
  background-color: var(--n-color-target);
}

:deep(.n-split) {
  height: 100%;
}

/* 上半区：搜索表单，内容较少，允许自身滚动 */
:deep(.n-split-pane:first-child) {
  overflow: auto;
  padding: var(--g-space-sm);
}

/* 下半区：统计面板 + 表格区域 */
:deep(.n-split-pane:last-child) {
  overflow: hidden;
  padding: var(--g-space-sm);
  display: flex;
  flex-direction: column;
}

.bottom-section {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.stats-section {
  flex-shrink: 0;
}

.grid-section {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

/* 等宽字体 */
.font-mono {
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
}

/* 主色文字 */
.text-primary {
  color: var(--n-primary-color);
}
</style>
