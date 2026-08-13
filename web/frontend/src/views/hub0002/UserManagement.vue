<template>
  <div class="user-management" :id="service.model.moduleId">
    <RsSplitPane
      class="user-management__split"
      orientation="vertical"
      :panes="splitPanes"
      disabled
    >
      <!-- 上部：搜索表单（auto 随内容） -->
      <template #search>
        <div class="user-management__search">
          <RsSearchForm
            ref="searchFormRef"
            :module-id="service.model.moduleId"
            v-bind="service.model.searchFormConfig"
            @search="handleSearch"
            @toolbar-click="handleToolbarClick"
          />
        </div>
      </template>

      <!-- 下部：表格占满剩余高度 -->
      <template #grid>
        <div class="user-management__grid">
          <RsGrid
            ref="gridRef"
            :module-id="service.model.moduleId"
            :data="service.model.userList"
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

    <!-- 用户对话框（新增/编辑/查看共用） -->
    <RsDataFormModal
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
import { RsDataFormModal } from '@/components/form/rs-data'
import { RsSearchForm } from '@/components/form/rs-search'
import { RsGrid, type RsGridExpose } from '@/components/rs-grid'
import { RsSplitPane, type RsSplitPaneItem } from '@/ui'
import { ref } from 'vue'
import UserRoleAuthDialog from './compoents/UserRoleAuthDialog.vue'
import { useUserPage } from './hooks'

defineOptions({
  name: 'UserManagement',
})

/** 上方面板内容自适应，下方吃满剩余空间；disabled 禁止拖拽 */
const splitPanes: RsSplitPaneItem[] = [
  { key: 'search', size: 'auto' },
  { key: 'grid' },
]

const searchFormRef = ref()
const gridRef = ref<RsGridExpose | null>(null)

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
  handleSearch,
} = useUserPage(gridRef, searchFormRef)
</script>

<style scoped>
.user-management {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.user-management__split {
  flex: 1 1 auto;
  width: 100%;
  height: 100%;
  min-height: 0;
}

.user-management__search {
  width: 100%;
}

.user-management__grid {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
</style>
