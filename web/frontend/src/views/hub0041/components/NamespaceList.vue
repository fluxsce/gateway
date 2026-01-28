<template>
  <div class="namespace-list" :id="effectiveModuleId">
    <GPane direction="vertical" default-size="80px">
      <!-- 上部：搜索表单 -->
      <template #1>
        <search-form
          ref="searchFormRef"
          :module-id="effectiveModuleId"
          :fields="namespaceService.model.searchFormConfig.fields"
          :show-search-button="true"
          :show-reset-button="true"
          @search="handleSearch"
        />
      </template>

      <!-- 下部：数据表格 -->
      <template #2>
        <g-grid
          ref="gridRef"
          :module-id="effectiveModuleId"
          :data="namespaceService.model.namespaceList"
          :loading="namespaceService.model.loading"
          v-bind="readonlyGridConfig"
          @page-change="namespaceService.handlePageChange"
          @menu-click="handleMenuClick"
          @row-click="handleRowClick"
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
      v-if="showDialog"
      v-model:visible="formDialogVisible"
      :mode="formDialogMode"
      :title="formDialogMode === 'create' ? '新增命名空间' : formDialogMode === 'edit' ? '编辑命名空间' : '查看命名空间详情'"
      :to="`#${effectiveModuleId}`"
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
import GdataFormModal from '@/components/form/data/GDataFormModal.vue'
import SearchForm from '@/components/form/search/SearchForm.vue'
import { GPane } from '@/components/gpane'
import type { GridProps } from '@/components/grid'
import { GGrid } from '@/components/grid'
import { NTag } from 'naive-ui'
import { computed, onMounted, ref } from 'vue'
import { useNamespacePage } from '../hooks'
import type { Namespace } from '../types'

// 定义组件名称
defineOptions({
  name: 'NamespaceList'
})

// ============= Props =============

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

// 计算有效的模块ID
const effectiveModuleId = computed(() => props.moduleId)

// ============= Emits =============

interface Emits {
  /** 命名空间行点击事件 */
  (e: 'row-click', row: Namespace): void
  /** 命名空间选择变化事件 */
  (e: 'namespace-select', namespace: Namespace | null): void
}

const emit = defineEmits<Emits>()

// ============= Refs =============

const searchFormRef = ref()
const gridRef = ref()

// ============= 页面级 Hook（包含服务与对话框、事件处理） =============

const {
  service: namespaceService,
  formDialogVisible,
  formDialogMode,
  currentEditNamespace,
  submitting,
  handleFormSubmit,
  handleMenuClick: handleMenuClickBase,
  handleSearch,
} = useNamespacePage(gridRef, searchFormRef)

// ============= 只读表格配置（移除编辑和删除菜单，只保留查看） =============

const readonlyGridConfig = computed<Omit<GridProps, 'moduleId' | 'data' | 'loading'>>(() => {
  const baseConfig = namespaceService.model.gridConfig
  return {
    ...baseConfig,
    menuConfig: {
      enabled: true,
      showCopyRow: baseConfig.menuConfig?.showCopyRow || false,
      showCopyCell: baseConfig.menuConfig?.showCopyCell || false,
      customMenus: [
        {
          code: 'view',
          name: '查看详情',
          prefixIcon: 'vxe-icon-eye-fill',
        },
      ],
    },
  }
})

// ============= 事件处理 =============

/**
 * 菜单点击处理（只处理查看，编辑和删除已移除）
 */
const handleMenuClick = (params: { code: string; row?: any }) => {
  if (params.code === 'view' && params.row) {
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

// ============= 暴露方法 =============

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
  const selectedRows = gridRef.value?.getCheckboxRecords() || []
  return selectedRows.length > 0 ? selectedRows[0] : null
}

/**
 * 获取当前行（点击的行）
 */
const getCurrentNamespace = (): Namespace | null => {
  return gridRef.value?.getCurrentRecord() || null
}

defineExpose({
  refresh,
  load,
  getSelectedNamespace,
  getCurrentNamespace,
  namespaceService,
})

// ============= 生命周期 =============

onMounted(() => {
  if (props.autoLoad) {
    // 初始化加载数据
    namespaceService.loadNamespaces()
  }
})
</script>

<style scoped>
.namespace-list {
  height: 100%;
  width: 100%;
}
</style>

