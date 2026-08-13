<template>
  <div class="role-management" :id="service.model.moduleId">
    <RsSplitPane
      class="role-management__split"
      orientation="vertical"
      :panes="splitPanes"
      disabled
    >
      <!-- 上部：搜索表单（auto 随内容） -->
      <template #search>
        <div class="role-management__search">
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
        <div class="role-management__grid">
          <RsGrid
            ref="gridRef"
            :module-id="service.model.moduleId"
            :data="service.model.roleList"
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

    <!-- 角色对话框（新增/编辑/查看共用） -->
    <RsDataFormModal
      v-model:visible="formDialogVisible"
      :mode="formDialogMode"
      :title="formDialogTitle"
      :to="`#${service.model.moduleId}`"
      :form-fields="service.model.formFields"
      :form-tabs="service.model.formTabs"
      :initial-data="currentEditRole || undefined"
      :auto-close-on-confirm="false"
      :confirm-loading="service.model.loading.value"
      @submit="handleFormSubmit"
    />

    <!-- 角色授权抽屉 -->
    <RoleResourceDrawer
      v-model:show="roleAuthDrawerVisible"
      :role-id="roleAuthRoleId"
      :role-name="roleAuthRoleName"
      @close="closeRoleAuthDrawer"
    />
  </div>
</template>

<script lang="ts" setup>
import { RsDataFormModal } from '@/components/form/rs-data'
import { RsSearchForm } from '@/components/form/rs-search'
import { RsGrid, type RsGridExpose } from '@/components/rs-grid'
import { RsSplitPane, type RsSplitPaneItem } from '@/ui'
import { ref } from 'vue'
import RoleResourceDrawer from './compoents/RoleResourceDrawer.vue'
import { useRolePage } from './hooks'

defineOptions({
  name: 'RoleManagement',
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
  formDialogTitle,
  currentEditRole,
  handleFormSubmit,
  handleToolbarClick,
  handleMenuClick,
  handleSearch,
  roleAuthDrawerVisible,
  roleAuthRoleId,
  roleAuthRoleName,
  closeRoleAuthDrawer,
} = useRolePage(gridRef, searchFormRef)
</script>

<style scoped>
.role-management {
  box-sizing: border-box;
  position: relative;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.role-management__split {
  flex: 1 1 auto;
  width: 100%;
  height: 100%;
  min-height: 0;
}

.role-management__search {
  width: 100%;
}

.role-management__grid {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
</style>
