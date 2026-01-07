<template>
  <div class="tunnel-server-management" :id="service.model.moduleId">
    <GPane direction="vertical" default-size="80px">
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

      <!-- 下部：统计信息 + 数据表格 -->
      <template #2>
        <div class="bottom-section">
          <!-- 统计信息 -->
          <div class="stats-section" v-if="showStats">
            <tunnel-server-stats :statistics="statistics" />
          </div>
          
          <!-- 数据表格 -->
          <div class="grid-section">
            <g-grid
              ref="gridRef"
              :module-id="service.model.moduleId"
              :data="service.model.tunnelServerList"
              :loading="service.model.loading"
              v-bind="service.model.gridConfig"
              @page-change="service.handlePageChange"
              @menu-click="handleMenuClick"
            >
              <!-- 服务器名称自定义渲染 -->
              <template #serverName="{ row }">
                <span :class="row.serverStatus === 'running' ? 'text-success font-bold' : 'text-default'">
                  {{ row.serverName }}
                </span>
              </template>

              <!-- 最大客户端自定义渲染 -->
              <template #maxClients="{ row }">
                <span class="text-primary font-bold">{{ row.maxClients || '-' }}</span>
              </template>

              <!-- Token认证自定义渲染 -->
              <template #tokenAuth="{ row }">
                <n-tag :type="row.tokenAuth === 'Y' ? 'success' : 'default'" size="small">
                  <template #icon>
                    <n-icon>
                      <ShieldCheckmarkOutline />
                    </n-icon>
                  </template>
                  {{ row.tokenAuth === 'Y' ? '启用' : '禁用' }}
                </n-tag>
              </template>

              <!-- TLS加密自定义渲染 -->
              <template #tlsEnable="{ row }">
                <n-tag :type="row.tlsEnable === 'Y' ? 'success' : 'default'" size="small">
                  {{ row.tlsEnable === 'Y' ? '启用' : '禁用' }}
                </n-tag>
              </template>

              <!-- 服务器状态自定义渲染 -->
              <template #serverStatus="{ row }">
                <n-tag
                  :type="
                    row.serverStatus === 'running'
                      ? 'success'
                      : row.serverStatus === 'stopped'
                        ? 'warning'
                        : row.serverStatus === 'error'
                          ? 'error'
                          : 'default'
                  "
                  size="small"
                >
                  {{
                    row.serverStatus === 'running'
                      ? '运行中'
                      : row.serverStatus === 'stopped'
                        ? '已停止'
                        : row.serverStatus === 'error'
                          ? '错误'
                          : '未知'
                  }}
                </n-tag>
              </template>
            </g-grid>
          </div>
        </div>
      </template>
    </GPane>

    <!-- 隧道服务器对话框（新增/编辑/查看共用） -->
    <GdataFormModal
      v-model:visible="formDialogVisible"
      :mode="formDialogMode"
      :title="formDialogMode === 'create' ? '新增隧道服务器' : formDialogMode === 'edit' ? '编辑隧道服务器' : '查看隧道服务器详情'"
      :to="`#${service.model.moduleId}`"
      :form-fields="service.model.formFields"
      :form-tabs="service.model.formTabs"
      :initial-data="currentEditServer || undefined"
      :auto-close-on-confirm="false"
      :confirm-loading="service.model.loading.value"
      @submit="handleFormSubmit"
    />
  </div>
</template>

<script lang="ts" setup>
import GdataFormModal from '@/components/form/data/GDataFormModal.vue'
import SearchForm from '@/components/form/search/SearchForm.vue'
import { GPane } from '@/components/gpane'
import { GGrid } from '@/components/grid'
import { isApiSuccess, parseJsonData } from '@/utils/format'
import { ShieldCheckmarkOutline } from '@vicons/ionicons5'
import { NIcon, NTag } from 'naive-ui'
import { onMounted, ref } from 'vue'
import * as tunnelServerApi from '../../api'
import type { TunnelServerStats as TunnelServerStatsType } from '../../types'
import TunnelServerStats from '../stats/TunnelServerStats.vue'
import { useTunnelServerPage } from './hooks'

// 定义组件名称
defineOptions({
  name: 'TunnelServerList'
})

// ============= Refs =============

const searchFormRef = ref()
const gridRef = ref()

// ============= 统计信息（可选） =============
const showStats = ref(true) // 显示统计信息
const statistics = ref<TunnelServerStatsType>({
  totalServers: 0,
  runningServers: 0,
  stoppedServers: 0,
  errorServers: 0,
  totalClients: 0,
  totalConnections: 0
})

// 获取统计信息
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
        totalConnections: 0
      })
      statistics.value = data
    }
  } catch (error) {
    console.error('获取统计信息失败:', error)
  }
}

// ============= 页面级 Hook（包含服务与对话框、事件处理） =============

const {
  service,
  formDialogVisible,
  formDialogMode,
  currentEditServer,
  handleFormSubmit,
  handleToolbarClick,
  handleMenuClick,
  handleSearch
} = useTunnelServerPage(gridRef, searchFormRef)

// 初始化
onMounted(() => {
  // 获取统计信息（如果需要）
  if (showStats.value) {
    getStatistics()
  }
  // 数据由搜索表单的"查询"按钮触发加载
})
</script>

<style lang="scss" scoped>
.tunnel-server-management {
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

/* 下半区：统计信息 + 表格区域 */
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
  background: var(--n-card-color);
  border-radius: var(--n-border-radius);
}

.grid-section {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
</style>

