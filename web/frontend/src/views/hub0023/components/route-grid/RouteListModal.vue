<template>
  <RsDialog
    :open="modalVisible"
    :title="props.title || '选择路由'"
    layout="window"
    :width="props.width || 1200"
    :teleport-to="props.to"
    :draggable="true"
    :fullscreenable="true"
    :modal="false"
    :show-overlay="false"
    :close-on-overlay-click="false"
    class="hub0023-route-list-dialog"
    @update:open="handleUpdateVisible"
    @after-close="handleAfterLeave"
  >
    <template #body>
      <div class="route-list-modal" :id="service.model.moduleId">
        <RsSplitPane
          class="route-list-modal__split"
          orientation="vertical"
          :panes="splitPanes"
          disabled
        >
          <template #search>
            <div class="route-list-modal__search">
              <RsSearchForm
                ref="searchFormRef"
                :module-id="service.model.moduleId"
                v-bind="service.model.searchFormConfig"
                @search="handleSearch"
              />
            </div>
          </template>

          <template #grid>
            <div class="route-list-modal__grid">
              <RsGrid
                ref="gridRef"
                :module-id="service.model.moduleId"
                :data="service.model.routeList"
                :loading="service.model.loading"
                :columns="service.model.gridConfig.columns"
                :selectable="false"
                :row-key="service.model.gridConfig.rowKey"
                height="100%"
                :pagination-config="service.model.gridConfig.paginationConfig"
                :menu-config="service.model.gridConfig.menuConfig"
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
import { RsSearchForm, type RsSearchFormExpose } from '@/components/form/rs-search'
import { RsGrid, type RsGridExpose } from '@/components/rs-grid'
import { RsDialog, RsSplitPane, type RsSplitPaneItem } from '@/ui'
import type { RouteConfig } from '@/views/hub0021/components/routes/types'
import { onBeforeUnmount, ref, watch } from 'vue'
import { useRouteListPage } from './hooks'

defineOptions({
  name: 'RouteListModal',
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
  /** 选中的路由名称（v-model） */
  modelValue?: string
  /** 网关实例ID（可选，用于过滤路由） */
  gatewayInstanceId?: string
}

const props = withDefaults(defineProps<Props>(), {
  visible: false,
  title: '',
  width: 1200,
  to: undefined,
  modelValue: '',
  gatewayInstanceId: undefined,
})

interface Emits {
  (e: 'update:visible', visible: boolean): void
  (e: 'after-leave'): void
  (e: 'select', route: RouteConfig): void
  (e: 'update:modelValue', value: string): void
}

const emit = defineEmits<Emits>()

const searchFormRef = ref<RsSearchFormExpose | null>(null)
const gridRef = ref<RsGridExpose | null>(null)
const modalVisible = ref(props.visible)

const stopVisibleWatch = watch(
  () => props.visible,
  (newVal) => {
    modalVisible.value = newVal
    if (newVal) {
      handleSearch()
    }
  },
)

const { service, handleSearch } = useRouteListPage(props.gatewayInstanceId, gridRef, searchFormRef)

const handleUpdateVisible = (visible: boolean) => {
  modalVisible.value = visible
  emit('update:visible', visible)
}

const handleAfterLeave = () => {
  emit('after-leave')
}

const handleRowClick = ({ row }: { row: RouteConfig }) => {
  if (!row) return
  emit('update:modelValue', row.routeName || '')
  emit('select', row)
  handleUpdateVisible(false)
}

onBeforeUnmount(() => {
  stopVisibleWatch()
})
</script>

<style scoped>
.route-list-modal {
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  width: 100%;
  height: min(70vh, 720px);
  min-height: 0;
  overflow: hidden;
}

.route-list-modal__split {
  flex: 1;
  min-height: 0;
  height: 100%;
}

.route-list-modal__search {
  width: 100%;
}

.route-list-modal__grid {
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
</style>
