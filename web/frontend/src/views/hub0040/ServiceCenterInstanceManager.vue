<template>
  <div class="service-center-instance-manager" :id="service.model.moduleId">
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

      <!-- 下部：数据表格 -->
      <template #2>
        <g-grid
          ref="gridRef"
          :module-id="service.model.moduleId"
          :data="service.model.instanceList"
          :loading="service.model.loading"
          v-bind="service.model.gridConfig"
          @page-change="service.handlePageChange"
          @menu-click="handleMenuClick"
        >
          <!-- 实例状态自定义渲染 -->
          <template #instanceStatus="{ row }">
            <n-tag
              :type="getInstanceStatusType(row.instanceStatus)"
              size="small"
            >
              <template #icon>
                <n-icon>
                  <CheckmarkCircleOutline v-if="row.instanceStatus === 'RUNNING'" />
                  <AlertCircleOutline v-else-if="row.instanceStatus === 'ERROR'" />
                  <HourglassOutline v-else-if="row.instanceStatus === 'STARTING' || row.instanceStatus === 'STOPPING'" />
                  <StopCircleOutline v-else />
                </n-icon>
              </template>
              {{ getInstanceStatusText(row.instanceStatus) }}
            </n-tag>
          </template>

          <!-- 运行状态自定义渲染（基于 instanceStatus） -->
          <template #isRunning="{ row }">
            <n-tag :type="row.instanceStatus === 'RUNNING' ? 'success' : 'default'" size="small">
              {{ row.instanceStatus === 'RUNNING' ? '运行中' : '已停止' }}
            </n-tag>
          </template>

          <!-- TLS 状态自定义渲染 -->
          <template #enableTLS="{ row }">
            <n-tag :type="row.enableTLS === 'Y' ? 'success' : 'default'" size="small">
              {{ row.enableTLS === 'Y' ? '启用' : '禁用' }}
            </n-tag>
          </template>

          <!-- 认证状态自定义渲染 -->
          <template #enableAuth="{ row }">
            <n-tag :type="row.enableAuth === 'Y' ? 'warning' : 'default'" size="small">
              {{ row.enableAuth === 'Y' ? '启用' : '禁用' }}
            </n-tag>
          </template>

          <!-- 活动状态自定义渲染 -->
          <template #activeFlag="{ row }">
            <n-tag :type="row.activeFlag === 'Y' ? 'success' : 'default'" size="small">
              {{ row.activeFlag === 'Y' ? '活动' : '非活动' }}
            </n-tag>
          </template>
        </g-grid>
      </template>
    </GPane>

    <!-- 实例对话框（新增/编辑/查看共用） -->
    <GdataFormModal
      v-model:visible="formDialogVisible"
      :mode="formDialogMode"
      :title="formDialogMode === 'create' ? '新增实例' : formDialogMode === 'edit' ? '编辑实例' : '查看实例详情'"
      :to="`#${service.model.moduleId}`"
      :form-fields="service.model.instanceFormConfig.fields"
      :form-tabs="service.model.instanceFormConfig.tabs"
      :initial-data="currentEditInstance || undefined"
      :auto-close-on-confirm="false"
      :confirm-loading="submitting"
      @submit="handleFormSubmit"
    />
  </div>
</template>

<script lang="ts" setup>
import GdataFormModal from '@/components/form/data/GDataFormModal.vue'
import SearchForm from '@/components/form/search/SearchForm.vue'
import { GPane } from '@/components/gpane'
import { GGrid } from '@/components/grid'
import {
    AlertCircleOutline,
    CheckmarkCircleOutline,
    HourglassOutline,
    StopCircleOutline
} from '@vicons/ionicons5'
import { NIcon, NTag } from 'naive-ui'
import { ref } from 'vue'
import { useServiceCenterInstancePage } from './hooks'

// 定义组件名称
defineOptions({
  name: 'ServiceCenterInstanceManager'
})

// ============= Refs =============

const searchFormRef = ref()
const gridRef = ref()

// ============= 页面级 Hook（包含服务与对话框、事件处理） =============

const {
  service,
  formDialogVisible,
  formDialogMode,
  currentEditInstance,
  submitting,
  handleFormSubmit,
  handleToolbarClick,
  handleMenuClick,
  handleSearch,
} = useServiceCenterInstancePage(gridRef, searchFormRef)

// ============= 辅助函数 =============

/**
 * 获取实例状态类型
 */
function getInstanceStatusType(status: string): 'default' | 'success' | 'error' | 'warning' | 'info' {
  const statusMap: Record<string, 'default' | 'success' | 'error' | 'warning' | 'info'> = {
    'RUNNING': 'success',
    'STOPPED': 'default',
    'STARTING': 'info',
    'STOPPING': 'warning',
    'ERROR': 'error',
  }
  return statusMap[status] || 'default'
}

/**
 * 获取实例状态文本
 */
function getInstanceStatusText(status: string): string {
  const statusMap: Record<string, string> = {
    'RUNNING': '运行中',
    'STOPPED': '停止',
    'STARTING': '启动中',
    'STOPPING': '停止中',
    'ERROR': '异常',
  }
  return statusMap[status] || status
}

// 数据由搜索表单的"查询"按钮触发加载
</script>

<style lang="scss" scoped>
.service-center-instance-manager {
  width: 100%;
  height: 100%;
  overflow: hidden;

  :deep(.n-split) {
    height: 100%;
  }

  /* 上半区：搜索表单，内容较少，允许自身滚动 */
  :deep(.n-split-pane:first-child) {
    overflow: auto;
    padding: var(--g-space-sm);
  }

  /* 下半区：表格区域，高度由 GGrid 占满，滚动全部交给 vxe-grid */
  :deep(.n-split-pane:last-child) {
    overflow: hidden;
    padding: var(--g-space-sm);
    display: flex;
    flex-direction: column;
  }
}
</style>

