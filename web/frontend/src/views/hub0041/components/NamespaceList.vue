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
          />
        </div>
      </template>

      <template #grid>
        <div class="namespace-list__grid">
          <RsGrid
            ref="gridRef"
            :module-id="effectiveModuleId"
            :data="namespaceService.model.namespaceList"
            :loading="namespaceService.model.loading"
            :columns="readonlyGridConfig.columns"
            :selectable="readonlyGridConfig.selectable"
            :row-key="readonlyGridConfig.rowKey"
            height="100%"
            :pagination-config="readonlyGridConfig.paginationConfig"
            :menu-config="readonlyGridConfig.menuConfig"
            @page-change="namespaceService.handlePageChange"
            @menu-click="handleMenuClick"
            @row-click="handleRowClick"
          />
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
import { RsGrid, type RsGridExpose } from '@/components/rs-grid'
import { RsSplitPane, type RsSplitPaneItem } from '@/ui'
import { computed, onMounted, ref } from 'vue'
import { useNamespacePage } from '../hooks'
import type { NamespaceGridConfig } from '../hooks/model'
import type { Namespace } from '../types'

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
  /** 命名空间行点击事件 */
  (e: 'row-click', row: Namespace): void
  /** 命名空间选择变化事件 */
  (e: 'namespace-select', namespace: Namespace | null): void
}

const emit = defineEmits<Emits>()

const searchFormRef = ref()
const gridRef = ref<RsGridExpose | null>(null)

const {
  service: namespaceService,
  formDialogVisible,
  formDialogMode,
  currentEditNamespace,
  submitting,
  handleFormSubmit,
  handleMenuClick: handleMenuClickBase,
  handleSearch,
} = useNamespacePage(gridRef, searchFormRef, effectiveModuleId.value)

/** 只读表格：右键菜单仅保留查看 */
const readonlyGridConfig = computed<NamespaceGridConfig>(() => {
  const baseConfig = namespaceService.model.gridConfig
  return {
    ...baseConfig,
    menuConfig: {
      enabled: true,
      items: [{ key: 'view', label: '查看详情', icon: 'eye' }],
    },
  }
})

/**
 * 菜单点击处理（只处理查看，编辑和删除已移除）
 */
const handleMenuClick = (params: { key: string; row?: Namespace }) => {
  if (params.key === 'view' && params.row) {
    handleMenuClickBase(params)
  }
}

/**
 * 命名空间行点击
 */
const handleRowClick = ({ row }: { row: Namespace }) => {
  emit('row-click', row)
  emit('namespace-select', row)
}

/**
 * 刷新命名空间列表
 */
const refresh = () => {
  namespaceService.handleRefresh()
}

/**
 * 加载命名空间列表
 */
const load = () => {
  namespaceService.loadNamespaces()
}

/**
 * 获取选中的命名空间
 */
const getSelectedNamespace = (): Namespace | null => {
  const selectedRows = gridRef.value?.getSelectedRows() || []
  return selectedRows.length > 0 ? selectedRows[0] : null
}

/**
 * 获取当前行（点击的行）
 */
const getCurrentNamespace = (): Namespace | null => {
  return gridRef.value?.getCurrentRow() || null
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

.namespace-list__grid {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
</style>
