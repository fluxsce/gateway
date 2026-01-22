<template>
  <div class="alert-template-management" :id="htmlId">
    <GPane direction="vertical" :no-resize="true">
      <template #1>
        <search-form
          ref="searchFormRef"
          :module-id="service.model.moduleId"
          v-bind="service.model.searchFormConfig"
          @search="handleSearch"
          @toolbar-click="handleToolbarClick"
        />
      </template>

      <template #2>
        <g-grid
          ref="gridRef"
          :module-id="service.model.moduleId"
          :data="service.model.templateList"
          :loading="service.model.loading"
          v-bind="service.model.gridConfig"
          @page-change="handlePageChange"
          @menu-click="({ code, row }) => handleMenuClick({ menu: { code }, row })"
        >
          <template #channelType="{ row }">
            <n-tag size="small" type="info">
              {{ service.model.getChannelTypeLabel(row.channelType) || '通用' }}
            </n-tag>
          </template>

          <template #displayFormat="{ row }">
            <n-tag size="small" :type="row.displayFormat === 'table' ? 'warning' : 'default'">
              {{ service.model.getDisplayFormatLabel(row.displayFormat) }}
            </n-tag>
          </template>

          <template #activeFlag="{ row }">
            <n-switch
              :value="row.activeFlag === 'Y'"
              size="small"
              @update:value="() => handleToggleActive(row)"
            />
          </template>
        </g-grid>
      </template>
    </GPane>

    <GdataFormModal
      v-model:visible="formDialogVisible"
      :mode="formDialogMode"
      :title="formDialogMode === 'create' ? '新增预警模板' : formDialogMode === 'edit' ? '编辑预警模板' : '查看预警模板'"
      :to="`#${htmlId}`"
      :form-fields="service.model.formFields"
      :form-tabs="service.model.formTabs"
      :initial-data="currentEditTemplate || undefined"
      :auto-close-on-confirm="false"
      :confirm-loading="service.model.loading.value"
      @submit="handleFormSubmit"
    />
  </div>
</template>

<script setup lang="ts">
import GdataFormModal from '@/components/form/data/GDataFormModal.vue'
import SearchForm from '@/components/form/search/SearchForm.vue'
import { GPane } from '@/components/gpane'
import { GGrid } from '@/components/grid'
import { NSwitch, NTag, useMessage } from 'naive-ui'
import { ref } from 'vue'
import { useAlertTemplatePage } from './hooks'
import type { AlertTemplate } from './types'

defineOptions({ name: 'AlertTemplateManagement' })

const message = useMessage()
const searchFormRef = ref()
const gridRef = ref()

const {
  service,
  formDialogVisible,
  formDialogMode,
  currentEditTemplate,
  handleFormSubmit,
  handleToolbarClick,
  handleMenuClick,
  handleSearch,
  handlePageChange,
} = useAlertTemplatePage(gridRef, searchFormRef)

const htmlId = 'hub0081-alert-template'

// 启用/禁用：通过 update 接口实现（不再单独提供 setActiveFlag）
const handleToggleActive = async (row: AlertTemplate) => {
  const newFlag = row.activeFlag === 'Y' ? 'N' : 'Y'
  await service.editTemplate(row.templateName, { ...row, activeFlag: newFlag })
}
</script>

<style scoped>
.alert-template-management {
  height: 100%;
  display: flex;
  flex-direction: column;
}
</style>


