<template>
  <div class="service-list" :id="service.model.moduleId">
    <GPane direction="vertical" default-size="300px">
      <!-- 上部：命名空间列表 -->
      <template #1>
        <GCard>
          <div class="namespace-section">
            <div class="section-header">
              <h3>命名空间列表</h3>
            </div>
            <NamespaceList
              ref="namespaceListRef"
              moduleId="hub0042-namespace"
              :show-dialog="true"
              :auto-load="true"
              @row-click="handleNamespaceRowClick"
              @namespace-select="handleNamespaceSelect"
            />
          </div>
        </GCard>
      </template>

      <!-- 下部：服务列表 -->
      <template #2>
        <GCard>
          <div class="service-section">
            <div class="section-header">
              <h3>服务列表</h3>
              <span v-if="selectedNamespace" class="selected-namespace">
                当前命名空间: {{ selectedNamespace.namespaceName }} ({{ selectedNamespace.namespaceId }})
              </span>
            </div>
            <GPane direction="vertical" default-size="80px">
              <!-- 服务搜索表单 -->
              <template #1>
                <search-form
                  ref="serviceSearchFormRef"
                  :module-id="service.model.moduleId"
                  v-bind="service.model.searchFormConfig"
                  @search="handleServiceSearch"
                  @toolbar-click="handleServiceToolbarClick"
                />
              </template>

              <!-- 服务数据表格 -->
              <template #2>
                <g-grid
                  ref="serviceGridRef"
                  :module-id="service.model.moduleId"
                  :data="service.model.serviceList"
                  :loading="service.model.loading"
                  v-bind="service.model.gridConfig"
                  @page-change="service.handlePageChange"
                  @menu-click="handleServiceMenuClick"
                >
                  <!-- 活动状态自定义渲染 -->
                  <template #activeFlag="{ row }">
                    <RsTag :variant="row.activeFlag === 'Y' ? 'success' : 'default'" size="sm">
                      {{ row.activeFlag === 'Y' ? '活动' : '非活动' }}
                    </RsTag>
                  </template>
                </g-grid>
              </template>
            </GPane>
          </div>
        </GCard>
      </template>
    </GPane>


    <!-- 服务对话框（新增/编辑/查看共用） -->
    <GdataFormModal
      v-model:visible="serviceFormDialogVisible"
      :mode="serviceFormDialogMode"
      :title="serviceFormDialogMode === 'create' ? '新增服务' : serviceFormDialogMode === 'edit' ? '编辑服务' : '查看服务详情'"
      :to="`#${service.model.moduleId}`"
      :form-fields="service.model.serviceFormConfig.fields"
      :form-tabs="service.model.serviceFormConfig.tabs"
      :initial-data="currentEditService || undefined"
      :auto-close-on-confirm="false"
      :confirm-loading="serviceSubmitting"
      @submit="handleServiceFormSubmit"
    />
  </div>
</template>

<script lang="ts" setup>
import GdataFormModal from '@/components/form/data/GDataFormModal.vue'
import SearchForm from '@/components/form/search/SearchForm.vue'
import { GCard } from '@/components/gcard'
import { GPane } from '@/components/gpane'
import { GGrid } from '@/components/grid'
import { RsTag } from '@/ui'
import { onMounted, ref } from 'vue'
import { NamespaceList } from '../hub0041/components'
import type { Namespace } from '../hub0041/types'
import { useServicePage } from './hooks'

// 定义组件名称
defineOptions({
  name: 'ServiceList'
})

// ============= Refs =============

const namespaceListRef = ref()
const serviceSearchFormRef = ref()
const serviceGridRef = ref()

// 选中的命名空间
const selectedNamespace = ref<Namespace | null>(null)

// ============= 服务相关 =============

const {
  service,
  formDialogVisible: serviceFormDialogVisible,
  formDialogMode: serviceFormDialogMode,
  currentEditService,
  submitting: serviceSubmitting,
  handleFormSubmit: handleServiceFormSubmitBase,
  handleToolbarClick: handleServiceToolbarClick,
  handleMenuClick: handleServiceMenuClick,
  handleSearch: handleServiceSearch,
} = useServicePage(serviceGridRef, serviceSearchFormRef)

// ============= 事件处理 =============

/**
 * 命名空间行点击 - 选择命名空间并加载服务列表
 */
const handleNamespaceRowClick = async (row: Namespace) => {
  if (row) {
    selectedNamespace.value = row
    // 自动设置服务搜索表单的命名空间ID
    if (serviceSearchFormRef.value?.setFieldValue) {
      serviceSearchFormRef.value.setFieldValue('namespaceId', row.namespaceId)
    }
    // 加载该命名空间下的服务列表
    await service.loadServices({ namespaceId: row.namespaceId })
  }
}

/**
 * 命名空间选择变化
 */
const handleNamespaceSelect = (namespace: Namespace | null) => {
  selectedNamespace.value = namespace
}

/**
 * 服务表单提交（自动填充命名空间ID）
 */
const handleServiceFormSubmit = (formData?: Record<string, any>) => {
  if (formData) {
    // 如果选中了命名空间，自动填充命名空间ID
    if (selectedNamespace.value && !formData.namespaceId) {
      formData.namespaceId = selectedNamespace.value.namespaceId
    }
    handleServiceFormSubmitBase(formData)
  }
}

// ============= 生命周期 =============

onMounted(() => {
  // 命名空间列表组件会自动加载数据
})
</script>

<style lang="scss" scoped>
.service-list {
  width: 100%;
  height: 100%;
  overflow: hidden;

  :deep(.n-split) {
    height: 100%;
  }

  /* 上半区：命名空间列表 */
  :deep(.n-split-pane:first-child) {
    overflow: hidden;
    padding: var(--g-space-sm);

    .g-card {
      height: 100%;
      overflow: hidden;

      :deep(.n-card__content) {
        height: 100%;
        overflow: hidden;
      }
    }
  }

  /* 下半区：服务列表 */
  :deep(.n-split-pane:last-child) {
    overflow: hidden;
    padding: var(--g-space-sm);

    .g-card {
      height: 100%;
      overflow: hidden;

      :deep(.n-card__content) {
        height: 100%;
        overflow: hidden;
      }
    }
  }

  .namespace-section,
  .service-section {
    height: 100%;
    display: flex;
    flex-direction: column;

    .section-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: var(--g-space-sm);
      border-bottom: 1px solid var(--g-border-color);

      h3 {
        margin: 0;
        font-size: 16px;
        font-weight: 500;
      }

      .selected-namespace {
        font-size: 14px;
        color: var(--g-text-color-secondary);
      }
    }
  }

  /* 搜索表单区域 */
  :deep(.n-split-pane .n-split-pane:first-child) {
    overflow: auto;
    padding: var(--g-space-sm);
  }

  /* 表格区域 */
  :deep(.n-split-pane .n-split-pane:last-child) {
    overflow: hidden;
    padding: var(--g-space-sm);
    display: flex;
    flex-direction: column;
  }
}
</style>

