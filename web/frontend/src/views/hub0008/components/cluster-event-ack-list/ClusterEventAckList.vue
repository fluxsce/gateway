<template>
  <div class="cluster-event-ack-list" id="cluster-event-ack-list">
    <GPane direction="vertical" :no-resize="true" >
      <!-- 上部：搜索表单 -->
      <template #1>
        <search-form
          ref="searchFormRef"
          :module-id="service.model.moduleId"
          v-bind="service.model.searchFormConfig"
          @search="handleSearch"
        />
      </template>

      <!-- 下部：数据表格 -->
      <template #2>
        <g-grid
          ref="gridRef"
          :module-id="service.model.moduleId"
          :data="service.model.ackList"
          :loading="service.model.loading"
          v-bind="service.model.gridConfig"
          @page-change="service.handlePageChange"
          @menu-click="handleMenuClick"
        >
          <!-- 处理节点ID自定义渲染 -->
          <template #nodeId="{ row }">
            <span style="color: var(--n-color-primary); font-weight: 500;">
              {{ row.nodeId || '-' }}
            </span>
          </template>

          <!-- 处理节点IP自定义渲染 -->
          <template #nodeIp="{ row }">
            <span style="color: var(--n-color-success); font-weight: 500;">
              {{ row.nodeIp || '-' }}
            </span>
          </template>

          <!-- 确认状态自定义渲染 -->
          <template #ackStatus="{ row }">
            <n-tag
              :type="
                row.ackStatus === 'SUCCESS'
                  ? 'success'
                  : row.ackStatus === 'FAILED'
                    ? 'error'
                    : row.ackStatus === 'PENDING'
                      ? 'warning'
                      : 'default'
              "
              size="small"
            >
              {{
                row.ackStatus === 'PENDING'
                  ? '待处理'
                  : row.ackStatus === 'SUCCESS'
                    ? '成功'
                    : row.ackStatus === 'FAILED'
                      ? '失败'
                      : row.ackStatus === 'SKIPPED'
                        ? '跳过'
                        : row.ackStatus
              }}
            </n-tag>
          </template>
        </g-grid>
      </template>
    </GPane>

    <!-- 事件确认详情对话框 -->
    <ClusterEventAckDetailDialog
      v-model:show="detailDialogVisible"
      :ack="currentAck"
    />
  </div>
</template>

<script setup lang="ts">
import SearchForm from '@/components/form/search/SearchForm.vue'
import { GPane } from '@/components/gpane'
import { GGrid } from '@/components/grid'
import { NTag } from 'naive-ui'
import { computed, onMounted, ref, watch } from 'vue'
import type { ClusterEventAck } from '../../types'
import ClusterEventAckDetailDialog from './ClusterEventAckDetailDialog.vue'
import { useClusterEventAckService } from './hooks'

// 定义组件名称
defineOptions({
  name: 'ClusterEventAckList'
})

interface Props {
  eventId?: string
}

const props = withDefaults(defineProps<Props>(), {
  eventId: undefined
})

// ============= Refs =============

const searchFormRef = ref()
const gridRef = ref()
const eventIdRef = computed(() => props.eventId)

// ============= 详情对话框状态 =============

const detailDialogVisible = ref(false)
const currentAck = ref<ClusterEventAck | null>(null)

// ============= 页面级 Hook（包含服务与事件处理） =============

const service = useClusterEventAckService(eventIdRef, searchFormRef)

// 处理搜索
const handleSearch = async (formData?: Record<string, any>) => {
  await service.handleSearch(formData)
}

// 处理右键菜单点击
const handleMenuClick = async ({ code, row }: { code: string; row?: ClusterEventAck }) => {
  if (!row) return

  switch (code) {
    case 'view':
      // 获取最新详情
      const detail = await service.getAckDetail(row.ackId)
      if (detail) {
        currentAck.value = detail
        detailDialogVisible.value = true
      } else {
        // 如果获取详情失败，使用当前行数据
        currentAck.value = row
        detailDialogVisible.value = true
      }
      break
  }
}

// 监听 eventId 变化
watch(
  () => props.eventId,
  (newEventId) => {
    if (newEventId) {
      service.model.resetPagination()
      service.loadAcks()
    } else {
      service.model.clearAckList()
      service.model.resetPagination()
    }
  },
  { immediate: true }
)

// 组件挂载时加载数据
onMounted(() => {
  if (props.eventId) {
    service.loadAcks()
  }
})
</script>

<style lang="scss" scoped>
.cluster-event-ack-list {
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

