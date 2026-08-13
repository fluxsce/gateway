<template>
  <div class="cluster-event-ack-list" id="cluster-event-ack-list">
    <RsSplitPane
      class="cluster-event-ack-list__split"
      orientation="vertical"
      :panes="splitPanes"
      disabled
    >
      <template #search>
        <div class="cluster-event-ack-list__search">
          <RsSearchForm
            ref="searchFormRef"
            :module-id="service.model.moduleId"
            v-bind="service.model.searchFormConfig"
            @search="handleSearch"
          />
        </div>
      </template>

      <template #grid>
        <div class="cluster-event-ack-list__grid">
          <RsGrid
            ref="gridRef"
            :module-id="service.model.moduleId"
            :data="service.model.ackList"
            :loading="service.model.loading"
            :columns="service.model.gridConfig.columns"
            :selectable="service.model.gridConfig.selectable"
            :row-key="service.model.gridConfig.rowKey"
            height="100%"
            :pagination-config="service.model.gridConfig.paginationConfig"
            :menu-config="service.model.gridConfig.menuConfig"
            @page-change="service.handlePageChange"
            @menu-click="handleMenuClick"
          />
        </div>
      </template>
    </RsSplitPane>

    <ClusterEventAckDetailDialog
      v-model:show="detailDialogVisible"
      :ack="currentAck"
    />
  </div>
</template>

<script setup lang="ts">
import { RsSearchForm } from '@/components/form/rs-search'
import { RsGrid, type RsGridExpose } from '@/components/rs-grid'
import { RsSplitPane, type RsSplitPaneItem } from '@/ui'
import { computed, onMounted, ref, watch } from 'vue'
import type { ClusterEventAck } from '../../types'
import ClusterEventAckDetailDialog from './ClusterEventAckDetailDialog.vue'
import { useClusterEventAckService } from './hooks'

defineOptions({
  name: 'ClusterEventAckList',
})

interface Props {
  eventId?: string
}

const props = withDefaults(defineProps<Props>(), {
  eventId: undefined,
})

const splitPanes: RsSplitPaneItem[] = [
  { key: 'search', size: 'auto' },
  { key: 'grid' },
]

const searchFormRef = ref()
const gridRef = ref<RsGridExpose | null>(null)
const eventIdRef = computed(() => props.eventId)
const detailDialogVisible = ref(false)
const currentAck = ref<ClusterEventAck | null>(null)

const service = useClusterEventAckService(eventIdRef, searchFormRef)

const handleSearch = async (formData?: Record<string, any>) => {
  await service.handleSearch(formData)
}

const handleMenuClick = async ({ key, row }: { key: string; row?: ClusterEventAck }) => {
  if (!row) return
  if (key === 'view') {
    const detail = await service.getAckDetail(row.ackId)
    currentAck.value = detail || row
    detailDialogVisible.value = true
  }
}

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
  { immediate: true },
)

onMounted(() => {
  if (props.eventId) {
    service.loadAcks()
  }
})
</script>

<style scoped>
.cluster-event-ack-list {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.cluster-event-ack-list__split {
  flex: 1 1 auto;
  width: 100%;
  height: 100%;
  min-height: 0;
}

.cluster-event-ack-list__search {
  width: 100%;
}

.cluster-event-ack-list__grid {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
</style>
