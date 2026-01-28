<template>
  <div class="namespace-management" :id="service.model.moduleId">
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
          :data="service.model.namespaceList"
          :loading="service.model.loading"
          v-bind="service.model.gridConfig"
          @page-change="service.handlePageChange"
          @menu-click="handleMenuClick"
        >
          <!-- 活动状态自定义渲染 -->
          <template #activeFlag="{ row }">
            <n-tag :type="row.activeFlag === 'Y' ? 'success' : 'default'" size="small">
              {{ row.activeFlag === 'Y' ? '活动' : '非活动' }}
            </n-tag>
          </template>
        </g-grid>
      </template>
    </GPane>

    <!-- 命名空间对话框（新增/编辑/查看共用） -->
    <GdataFormModal
      v-model:visible="formDialogVisible"
      :mode="formDialogMode"
      :title="formDialogMode === 'create' ? '新增命名空间' : formDialogMode === 'edit' ? '编辑命名空间' : '查看命名空间详情'"
      :to="`#${service.model.moduleId}`"
      :form-fields="service.model.namespaceFormConfig.fields"
      :form-tabs="service.model.namespaceFormConfig.tabs"
      :initial-data="currentEditNamespace || undefined"
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
import { NTag } from 'naive-ui'
import { ref } from 'vue'
import { useNamespacePage } from './hooks'

// 定义组件名称
defineOptions({
  name: 'NamespaceManagement'
})

// ============= Refs =============

const searchFormRef = ref()
const gridRef = ref()

// ============= 页面级 Hook（包含服务与对话框、事件处理） =============

const {
  service,
  formDialogVisible,
  formDialogMode,
  currentEditNamespace,
  submitting,
  handleFormSubmit,
  handleToolbarClick,
  handleMenuClick,
  handleSearch,
} = useNamespacePage(gridRef, searchFormRef)
</script>

<style scoped>
.namespace-management {
  height: 100%;
  width: 100%;
}
</style>

