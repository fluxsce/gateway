<template>
  <RsDialog
    :open="modalVisible"
    :title="props.title || '选择命名空间'"
    layout="window"
    :width="props.width || 1200"
    :teleport-to="props.to"
    :draggable="true"
    :fullscreenable="true"
    :modal="false"
    :show-overlay="false"
    :close-on-overlay-click="false"
    class="hub0041-namespace-list-dialog"
    @update:open="handleUpdateVisible"
    @after-close="handleAfterLeave"
  >
    <template #body>
      <div class="namespace-list-modal" :id="service.model.moduleId">
        <RsSplitPane
          class="namespace-list-modal__split"
          orientation="vertical"
          :panes="splitPanes"
          disabled
        >
          <template #search>
            <div class="namespace-list-modal__search">
              <RsSearchForm
                ref="searchFormRef"
                :module-id="service.model.moduleId"
                v-bind="readonlySearchFormConfig"
                @search="handleSearch"
                @reset="handleReset"
              />
            </div>
          </template>

          <template #grid>
            <div class="namespace-list-modal__grid">
              <RsGrid
                ref="gridRef"
                :module-id="service.model.moduleId"
                :data="service.model.namespaceList"
                :loading="service.model.loading"
                :columns="service.model.gridConfig.columns"
                :selectable="false"
                :row-key="service.model.gridConfig.rowKey"
                height="100%"
                :pagination-config="service.model.gridConfig.paginationConfig"
                @page-change="service.handlePageChange"
                @row-click="handleRowClick"
              />
            </div>
          </template>
        </RsSplitPane>
      </div>
    </template>
  </RsDialog>
</template>

<script lang="ts" setup>
import { RsSearchForm } from '@/components/form/rs-search'
import { RsGrid, type RsGridExpose } from '@/components/rs-grid'
import { RsDialog, RsSplitPane, type RsSplitPaneItem } from '@/ui'
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useNamespaceService } from '../hooks/useNamespaceService'
import type { Namespace } from '../types'

defineOptions({
  name: 'NamespaceListModal',
})

const splitPanes: RsSplitPaneItem[] = [
  { key: 'search', size: 'auto' },
  { key: 'grid' },
]

interface Props {
  /** 是否显示弹窗 */
  visible?: boolean
  /** 弹窗标题 */
  title?: string
  /** 弹窗宽度 */
  width?: number | string
  /** 弹窗挂载目标 */
  to?: string
  /** 选中的命名空间ID（v-model） */
  modelValue?: string
}

const props = withDefaults(defineProps<Props>(), {
  visible: false,
  title: '',
  width: 1200,
  to: undefined,
  modelValue: '',
})

interface Emits {
  (e: 'update:visible', visible: boolean): void
  (e: 'after-leave'): void
  (e: 'select', namespace: Namespace): void
  (e: 'update:modelValue', value: string): void
}

const emit = defineEmits<Emits>()

const searchFormRef = ref()
const gridRef = ref<RsGridExpose | null>(null)
const modalVisible = ref(props.visible)

const service = useNamespaceService(searchFormRef)

/** 选择弹窗只保留搜索/重置，不展示增删改工具栏 */
const readonlySearchFormConfig = computed(() => {
  const baseConfig = service.model.searchFormConfig
  return {
    ...baseConfig,
    toolbarButtons: [],
    showSearchButton: true,
    showResetButton: true,
  }
})

const stopVisibleWatch = watch(
  () => props.visible,
  (newVal) => {
    modalVisible.value = newVal
    if (newVal) {
      handleSearch()
    }
  },
)

/**
 * 处理弹窗可见性更新
 */
const handleUpdateVisible = (visible: boolean) => {
  modalVisible.value = visible
  emit('update:visible', visible)
}

/**
 * 处理弹窗关闭后事件
 */
const handleAfterLeave = () => {
  emit('after-leave')
}

/**
 * 处理搜索
 */
const handleSearch = () => {
  service.handleSearch()
}

/**
 * 处理重置
 */
const handleReset = () => {
  service.handleReset()
}

/**
 * 处理行点击事件（选择命名空间）
 */
const handleRowClick = ({ row }: { row: Namespace }) => {
  if (!row) return
  emit('update:modelValue', row.namespaceId)
  emit('select', row)
  handleUpdateVisible(false)
}

onBeforeUnmount(() => {
  stopVisibleWatch()
})
</script>

<style scoped>
.namespace-list-modal {
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  width: 100%;
  height: min(70vh, 720px);
  min-height: 0;
  overflow: hidden;
}

.namespace-list-modal__split {
  flex: 1;
  min-height: 0;
  height: 100%;
}

.namespace-list-modal__search {
  width: 100%;
}

.namespace-list-modal__grid {
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
</style>
