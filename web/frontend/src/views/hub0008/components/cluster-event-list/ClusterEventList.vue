<template>
  <div class="cluster-event-list" id="cluster-event-list">
    <GPane direction="vertical" :no-resize="true">
      <!-- 上部：搜索表单 -->
      <template #1>
        <search-form
          ref="searchFormRef"
          :module-id="service.model.moduleId"
          v-bind="computedSearchFormConfig"
          @search="handleSearch"
          @toolbar-click="handleToolbarClick"
        />
      </template>

      <!-- 下部：数据表格 -->
      <template #2>
        <g-grid
          ref="gridRef"
          :module-id="service.model.moduleId"
          :data="service.model.eventList"
          :loading="service.model.loading"
          v-bind="service.model.gridConfig"
          @page-change="service.handlePageChange"
          @row-click="handleRowClick"
          @menu-click="handleMenuClick"
        >
          <!-- 事件类型自定义渲染 -->
          <template #eventType="{ row }">
            <n-tag type="primary" size="small">
              {{ row.eventType }}
            </n-tag>
          </template>

          <!-- 事件动作自定义渲染 -->
          <template #eventAction="{ row }">
            <n-tag
              :type="
                row.eventAction === 'START'
                  ? 'success'
                  : row.eventAction === 'STOP'
                    ? 'error'
                    : row.eventAction === 'RELOAD'
                      ? 'warning'
                      : row.eventAction === 'RESTART'
                        ? 'info'
                        : row.eventAction === 'CREATE'
                          ? 'success'
                          : row.eventAction === 'UPDATE'
                            ? 'info'
                            : row.eventAction === 'DELETE'
                              ? 'error'
                              : row.eventAction === 'REFRESH' || row.eventAction === 'INVALIDATE'
                                ? 'warning'
                                : 'default'
              "
              size="small"
            >
              {{ row.eventAction }}
            </n-tag>
          </template>
        </g-grid>
      </template>
    </GPane>

    <!-- 事件详情对话框 -->
    <ClusterEventDetailDialog
      v-model:show="detailDialogVisible"
      :event="currentEvent"
    />
  </div>
</template>

<script setup lang="ts">
import SearchForm from '@/components/form/search/SearchForm.vue'
import { GPane } from '@/components/gpane'
import { GGrid } from '@/components/grid'
import { ChevronBackOutline, ChevronForwardOutline } from '@vicons/ionicons5'
import { NTag } from 'naive-ui'
import { computed, onMounted, ref } from 'vue'
import type { ClusterEvent } from '../../types'
import ClusterEventDetailDialog from './ClusterEventDetailDialog.vue'
import { useClusterEventPage } from './hooks'

// 定义组件名称
defineOptions({
  name: 'ClusterEventList'
})

interface Props {
  selectedEventId?: string
  showAckList?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  selectedEventId: undefined,
  showAckList: true
})

const emit = defineEmits<{
  (e: 'select', eventId: string): void
  (e: 'toggle-ack-list'): void
}>()

// ============= Refs =============

const searchFormRef = ref()
const gridRef = ref()

// ============= 详情对话框状态 =============

const detailDialogVisible = ref(false)
const currentEvent = ref<ClusterEvent | null>(null)

// ============= 页面级 Hook（包含服务与事件处理） =============

const {
  service,
  handleSearch
} = useClusterEventPage(searchFormRef)

// 计算搜索表单配置（动态更新按钮图标和文本）
const computedSearchFormConfig = computed(() => {
  const config = { ...service.model.searchFormConfig }
  // 更新 toggleAckList 按钮的图标和文本
  if (config.toolbarButtons) {
    config.toolbarButtons = config.toolbarButtons.map(btn => {
      if (btn.key === 'toggleAckList') {
        return {
          ...btn,
          label: props.showAckList ? '收起处理列表' : '展开处理列表',
          icon: props.showAckList ? ChevronForwardOutline : ChevronBackOutline
        }
      }
      return btn
    })
  }
  return config
})

// 处理行点击
const handleRowClick = ({ row }: { row: any }) => {
  emit('select', row.eventId)
}

// 处理工具栏按钮点击
const handleToolbarClick = (key: string) => {
  if (key === 'toggleAckList') {
    emit('toggle-ack-list')
  }
}

// 处理右键菜单点击
const handleMenuClick = ({ code, row }: { code: string; row?: ClusterEvent }) => {
  if (!row) return

  switch (code) {
    case 'view':
      currentEvent.value = row
      detailDialogVisible.value = true
      break
  }
}

// 组件挂载时加载数据
onMounted(() => {
  service.loadEvents()
})
</script>

<style lang="scss" scoped>
.cluster-event-list {
  width: 100%;
  height: 100%;
  overflow: hidden;

  :deep(.n-split) {
    height: 100%;
  }

  /* 上半区：搜索表单，内容较少，允许自身滚动 */
  :deep(.n-split-pane:first-child) {
    overflow: auto;
    padding: var(--g-space-sm);
  }

  /* 下半区：表格区域，高度由 GGrid 占满，滚动全部交给 vxe-grid */
  :deep(.n-split-pane:last-child) {
    overflow: hidden;
    padding: var(--g-space-sm);
    display: flex;
    flex-direction: column;
  }
}
</style>

