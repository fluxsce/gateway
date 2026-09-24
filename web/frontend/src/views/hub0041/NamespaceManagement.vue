<template>
  <div class="namespace-management" :id="service.model.moduleId">
    <RsSplitPane
      class="namespace-management__split"
      orientation="vertical"
      :panes="splitPanes"
      disabled
    >
      <template #search>
        <div class="namespace-management__search">
          <RsSearchForm
            ref="searchFormRef"
            :module-id="service.model.moduleId"
            v-bind="service.model.searchFormConfig"
            @search="handleSearch"
            @reset="handleReset"
            @toolbar-click="handleToolbarClick"
          />
        </div>
      </template>

      <template #grid>
        <div class="namespace-management__list">
          <NamespaceMonitorBoard
            :instance-name="queryScope?.instanceName || ''"
            :environment="queryScope?.environment || ''"
            :namespaces="namespaceList"
            :total-count="totalCount"
            :page-count="namespaceList.length"
            :overview="overview"
            :online="overviewOnline"
            :loading="overviewLoading"
            :selected-namespace="selectedNamespace"
          />
          <RsLoading v-if="loading" block size="lg" />
          <div v-else-if="!namespaceList.length" class="namespace-management__empty">
            <RsEmpty :description="emptyDescription">
              <template #icon>
                <GIcon :icon="LayersOutline" :size="32" color="var(--g-primary)" />
              </template>
            </RsEmpty>
          </div>
          <div v-else class="namespace-management__cards">
            <NamespaceCard
              v-for="namespace in namespaceList"
              :key="cardKey(namespace)"
              :namespace="namespace"
              :selected="isSelected(namespace)"
              menu="manage"
              @select="selectNamespace(namespace)"
              @focus="focusNamespace(namespace)"
              @action="(key) => handleCardAction(key, namespace)"
            />
          </div>
          <div
            v-if="totalCount > 0"
            class="namespace-management__pagination"
          >
            <RsPagination
              :page="currentPage"
              :page-size="pageSize"
              :total="totalCount"
              size="sm"
              :show-summary="true"
              @update:page="(p) => service.handlePageChange({ currentPage: p, pageSize })"
              @update:page-size="(s) => service.handlePageChange({ currentPage: 1, pageSize: s })"
            />
          </div>
        </div>
      </template>
    </RsSplitPane>

    <RsDataFormModal
      v-model:visible="formDialogVisible"
      :module-id="service.model.moduleId"
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
import { RsDataFormModal } from '@/components/form/rs-data'
import { RsSearchForm } from '@/components/form/rs-search'
import { GIcon } from '@/components/gicon'
import { RsEmpty, RsLoading, RsPagination, RsSplitPane, type RsSplitPaneItem } from '@/ui'
import { LayersOutline } from '@vicons/ionicons5'
import { getDefaultPageSize } from '@/utils/pagination'
import { computed, ref } from 'vue'
import NamespaceCard from './components/NamespaceCard.vue'
import NamespaceMonitorBoard from './components/NamespaceMonitorBoard.vue'
import { useNamespacePage } from './hooks'
import type { Namespace } from './types'

defineOptions({
  name: 'NamespaceManagement',
})

const splitPanes: RsSplitPaneItem[] = [
  { key: 'search', size: 'auto' },
  { key: 'grid' },
]

const searchFormRef = ref()

const {
  service,
  formDialogVisible,
  formDialogMode,
  currentEditNamespace,
  submitting,
  selectedNamespace,
  handleFormSubmit,
  handleToolbarClick,
  handleCardAction,
  handleSearch,
  handleReset,
  selectNamespace,
  focusNamespace,
  queryScope,
  overview,
  overviewOnline,
  overviewLoading,
} = useNamespacePage(searchFormRef)

const namespaceList = computed(() => service.model.namespaceList.value)
const loading = computed(() => service.model.loading.value)
const totalCount = computed(() => service.model.pageInfo.value?.totalCount || 0)
const currentPage = computed(() => service.model.pageInfo.value?.pageIndex || 1)
const pageSize = computed(() => service.model.pageInfo.value?.pageSize || getDefaultPageSize())

function cardKey(namespace: Namespace) {
  return namespace.oprSeqFlag || `${namespace.tenantId}:${namespace.namespaceId}`
}

function isSelected(namespace: Namespace) {
  const current = selectedNamespace.value
  return Boolean(
    current
    && current.tenantId === namespace.tenantId
    && current.namespaceId === namespace.namespaceId,
  )
}

const emptyDescription = computed(() => (
  queryScope.value?.instanceName
    ? '该实例下暂无匹配的命名空间'
    : '请先选择服务中心实例'
))
</script>

<style lang="scss" scoped>
.namespace-management {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.namespace-management__split {
  flex: 1 1 auto;
  width: 100%;
  height: 100%;
  min-height: 0;
}

.namespace-management__search {
  width: 100%;
  box-sizing: border-box;
}

.namespace-management__list {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  padding: 0;
  gap: 0;
  background: var(--g-bg-secondary, #f9fafb);
}

.namespace-management__empty {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--g-bg-primary);
}

.namespace-management__cards {
  flex: 1 1 auto;
  min-height: 0;
  overflow: auto;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(312px, 1fr));
  align-content: start;
  gap: 18px;
  padding: 16px 16px 12px;
  background: var(--g-bg-primary);
}

.namespace-management__pagination {
  flex: none;
  display: flex;
  justify-content: flex-end;
  padding: 8px 16px 12px;
  background: var(--g-bg-primary);
}
</style>
