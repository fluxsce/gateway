<template>
  <div class="alert-config-management" :id="htmlId">
    <GPane direction="vertical" :no-resize="true">
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
          :data="service.model.configList"
          :loading="service.model.loading"
          v-bind="service.model.gridConfig"
          @page-change="handlePageChange"
          @menu-click="({ code, row }) => handleMenuClick({ menu: { code }, row })"
        >
          <!-- 渠道类型自定义渲染 -->
          <template #channelType="{ row }">
            <n-tag :type="service.model.getChannelTypeTagType(row.channelType)" size="small">
              {{ service.model.getChannelTypeLabel(row.channelType) }}
            </n-tag>
          </template>

          <!-- 默认渠道自定义渲染 -->
          <template #defaultFlag="{ row }">
            <n-tag :type="row.defaultFlag === 'Y' ? 'success' : 'default'" size="small">
              {{ row.defaultFlag === 'Y' ? '是' : '否' }}
            </n-tag>
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
      </template>
    </GPane>

    <!-- 配置对话框（新增/编辑/查看共用） -->
    <GdataFormModal
      v-model:visible="formDialogVisible"
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

    <!-- 预警测试弹窗 -->
    <AlertTestModal
      v-model:visible="testModalVisible"
      :config="currentTestConfig"
      :to="`#${htmlId}`"
      @close="closeTestModal"
    />
  </div>
</template>

<script lang="ts" setup>
import GdataFormModal from '@/components/form/data/GDataFormModal.vue'
import SearchForm from '@/components/form/search/SearchForm.vue'
import { GPane } from '@/components/gpane'
import { GGrid } from '@/components/grid'
import { NSwitch, NTag } from 'naive-ui'
import { ref } from 'vue'
import { AlertTestModal } from './components'
import { useAlertConfigPage } from './hooks'

// 定义组件名称
defineOptions({
  name: 'AlertConfigManagement'
})

// ============= Refs =============

const searchFormRef = ref()
const gridRef = ref()

// ============= 页面级 Hook（包含服务与对话框、事件处理） =============

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

// ============= HTML ID（用于 DOM，符合 HTML 规范） =============

// 固定的 HTML id（符合 HTML 规范，无特殊字符）
// 注意：权限校验仍使用原始 moduleId（service.model.moduleId）
const htmlId = 'hub0080-alert-config'
</script>

<style scoped>
.alert-config-management {
  height: 100%;
  display: flex;
  flex-direction: column;
}
</style>

