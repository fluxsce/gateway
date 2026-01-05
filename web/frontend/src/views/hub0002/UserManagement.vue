<template>
  <div class="user-management" :id="service.model.moduleId">
    <GPane direction="vertical" :default-size="0.1" :min="0.1" :max="0.5">
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
          :data="service.model.userList"
          :loading="service.model.loading"
          v-bind="service.model.gridConfig"
          @page-change="service.handlePageChange"
          @menu-click="handleMenuClick"
        />
      </template>
    </GPane>

    <!-- 用户对话框（新增/编辑/查看共用） -->
    <GdataFormModal
      v-model:visible="formDialogVisible"
      :mode="formDialogMode"
      :title="formDialogMode === 'create' ? '新增用户' : formDialogMode === 'edit' ? '编辑用户' : '查看用户详情'"
      :to="`#${service.model.moduleId}`"
      :form-fields="service.model.formFields"
      :form-tabs="service.model.formTabs"
      :initial-data="currentEditUser || undefined"
      :auto-close-on-confirm="false"
      :confirm-loading="service.model.loading.value"
      @submit="handleFormSubmit"
    />

    <!-- 用户角色授权对话框 -->
    <UserRoleAuthDialog
      v-model:visible="roleAuthDialogVisible"
      :user-id="currentAuthUser?.userId || ''"
      :user="currentAuthUser || undefined"
      @saved="handleSearch"
    />
  </div>
</template>

<script lang="ts" setup>
import { GPane } from '@/components/gpane'
import GdataFormModal from '@/components/form/data/GDataFormModal.vue'
import SearchForm from '@/components/form/search/SearchForm.vue'
import { GGrid } from '@/components/grid'
import { ref } from 'vue'
import UserRoleAuthDialog from './compoents/UserRoleAuthDialog.vue'
import { useUserPage } from './hooks'

// 定义组件名称
defineOptions({
  name: 'UserManagement'
})

// ============= Refs =============

const searchFormRef = ref()
const gridRef = ref()

// ============= 页面级 Hook（包含服务与对话框、事件处理） =============

const {
  service,
  formDialogVisible,
  formDialogMode,
  currentEditUser,
  roleAuthDialogVisible,
  currentAuthUser,
  handleFormSubmit,
  handleToolbarClick,
  handleMenuClick,
  handleSearch
} = useUserPage(gridRef, searchFormRef)

// 数据由搜索表单的"查询"按钮触发加载
</script>

<style lang="scss" scoped>
.user-management {
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
