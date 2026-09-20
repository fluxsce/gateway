<template>
  <div class="namespace-list" id="namespace-list">
    <RsSplitPane
      class="namespace-list__split"
      orientation="vertical"
      :panes="splitPanes"
      disabled
    >
      <template #search>
        <div class="namespace-list__search">
          <RsSearchForm
            ref="searchFormRef"
            :module-id="effectiveModuleId"
            :fields="namespaceService.model.searchFormConfig.fields"
            :show-search-button="true"
            :show-reset-button="true"
            @search="handleSearch"
            @reset="handleReset"
          />
        </div>
      </template>

      <template #grid>
        <div class="namespace-list__list">
          <RsLoading v-if="loading" block size="lg" />
          <div v-else-if="!namespaceList.length" class="namespace-list__empty">
            <RsEmpty :description="emptyDescription">
              <template #icon>
                <GIcon :icon="LayersOutline" :size="32" color="var(--g-primary)" />
              </template>
            </RsEmpty>
          </div>
          <div v-else class="namespace-list__cards">
            <NamespaceCard
              v-for="namespace in namespaceList"
              :key="cardKey(namespace)"
              :namespace="namespace"
              :selected="isSelected(namespace)"
              menu="view"
              @select="handleCardSelect(namespace)"
              @focus="focusNamespace(namespace)"
              @action="(key) => handleCardAction(key, namespace)"
            />
          </div>
          <div
            v-if="totalCount > 0"
            class="namespace-list__pagination"
          >
            <RsPagination
              :page="currentPage"
              :page-size="pageSize"
              :total="totalCount"
              size="sm"
              :show-summary="true"
              @update:page="(p) => namespaceService.handlePageChange({ currentPage: p, pageSize })"
              @update:page-size="(s) => namespaceService.handlePageChange({ currentPage: 1, pageSize: s })"
            />
          </div>
        </div>
      </template>
    </RsSplitPane>

    <RsDataFormModal
      v-if="showDialog"
      :module-id="effectiveModuleId"
      v-model:visible="formDialogVisible"
      :mode="formDialogMode"
      :title="formDialogMode === 'create' ? '新增命名空间' : formDialogMode === 'edit' ? '编辑命名空间' : '查看命名空间详情'"
      :to="`#namespace-list`"
      :form-fields="namespaceService.model.namespaceFormConfig.fields"
      :form-tabs="namespaceService.model.namespaceFormConfig.tabs"
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
import { computed, onMounted, ref } from 'vue'
import { useNamespacePage } from '../hooks'
import type { Namespace } from '../types'
import NamespaceCard from './NamespaceCard.vue'

defineOptions({
  name: 'NamespaceList',
})

const splitPanes: RsSplitPaneItem[] = [
  { key: 'search', size: 'auto' },
  { key: 'grid' },
]

interface Props {
  /** 是否显示对话框（默认 true） */
  showDialog?: boolean
  /** 是否自动加载数据（默认 true） */
  autoLoad?: boolean
  /** 自定义模块ID（用于区分不同实例） */
  moduleId?: string
}

const props = withDefaults(defineProps<Props>(), {
  showDialog: true,
  autoLoad: true,
  moduleId: 'hub0041',
})

const effectiveModuleId = computed(() => props.moduleId)

interface Emits {
  (e: 'row-click', row: Namespace): void
  (e: 'namespace-select', namespace: Namespace | null): void
}

const emit = defineEmits<Emits>()

const searchFormRef = ref()

const {
  service: namespaceService,
  formDialogVisible,
  formDialogMode,
  currentEditNamespace,
  submitting,
  selectedNamespace,
  handleFormSubmit,
  handleCardAction: handleCardActionBase,
  handleSearch,
  handleReset,
  selectNamespace,
  focusNamespace,
  queryScope,
} = useNamespacePage(searchFormRef, effectiveModuleId.value)

const namespaceList = computed(() => namespaceService.model.namespaceList.value)
const loading = computed(() => namespaceService.model.loading.value)
const totalCount = computed(() => namespaceService.model.pageInfo.value?.totalCount || 0)
const currentPage = computed(() => namespaceService.model.pageInfo.value?.pageIndex || 1)
const pageSize = computed(() => namespaceService.model.pageInfo.value?.pageSize || 10)
const emptyDescription = computed(() => (
  queryScope.value?.instanceName
    ? '该实例下暂无匹配的命名空间'
    : '请先选择服务中心实例'
))

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

const handleCardSelect = (namespace: Namespace) => {
  selectNamespace(namespace)
  emit('row-click', namespace)
  emit('namespace-select', namespace)
}

const handleCardAction = (key: string, namespace: Namespace) => {
  selectNamespace(namespace)
  emit('namespace-select', namespace)
  if (key === 'view') {
    handleCardActionBase(key, namespace)
  }
}

const refresh = () => {
  namespaceService.handleRefresh()
}

const load = () => {
  namespaceService.loadNamespaces()
}

const getSelectedNamespace = (): Namespace | null => {
  return selectedNamespace.value
}

const getCurrentNamespace = (): Namespace | null => {
  return selectedNamespace.value
}

defineExpose({
  refresh,
  load,
  getSelectedNamespace,
  getCurrentNamespace,
  namespaceService,
})

onMounted(() => {
  if (props.autoLoad) {
    namespaceService.loadNamespaces()
  }
})
</script>

<style lang="scss" scoped>
.namespace-list {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.namespace-list__split {
  flex: 1 1 auto;
  width: 100%;
  height: 100%;
  min-height: 0;
}

.namespace-list__search {
  width: 100%;
  box-sizing: border-box;
}

.namespace-list__list {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  padding: 12px;
  gap: 12px;
  background: var(--g-bg-primary);
}

.namespace-list__empty {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.namespace-list__cards {
  flex: 1 1 auto;
  min-height: 0;
  overflow: auto;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  align-content: start;
  gap: 18px;
  padding: 4px 6px 12px;
}

.namespace-list__pagination {
  flex: none;
  display: flex;
  justify-content: flex-end;
}
</style>
