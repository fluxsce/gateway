<template>
  <div class="cluster-event-list" id="cluster-event-list">
    <RsSplitPane
      class="cluster-event-list__split"
      orientation="vertical"
      :panes="splitPanes"
      disabled
    >
      <template #search>
        <div class="cluster-event-list__search">
          <RsSearchForm
            ref="searchFormRef"
            :module-id="service.model.moduleId"
            v-bind="computedSearchFormConfig"
            @search="handleSearch"
            @toolbar-click="handleToolbarClick"
          />
        </div>
      </template>

      <template #grid>
        <div class="cluster-event-list__grid">
          <RsGrid
            ref="gridRef"
            :module-id="service.model.moduleId"
            :data="service.model.eventList"
            :loading="service.model.loading"
            :columns="service.model.gridConfig.columns"
            :selectable="service.model.gridConfig.selectable"
            :row-key="service.model.gridConfig.rowKey"
            height="100%"
            :pagination-config="service.model.gridConfig.paginationConfig"
            :menu-config="service.model.gridConfig.menuConfig"
            @page-change="service.handlePageChange"
            @row-click="handleRowClick"
            @menu-click="handleMenuClick"
          />
        </div>
      </template>
    </RsSplitPane>

    <ClusterEventDetailDialog
      v-model:show="detailDialogVisible"
      :event="currentEvent"
    />
  </div>
</template>

<script setup lang="ts">
import { RsSearchForm } from '@/components/form/rs-search'
import { RsGrid, type RsGridExpose } from '@/components/rs-grid'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { RsSplitPane, type RsSplitPaneItem } from '@/ui'
import { ChevronBackOutline, ChevronForwardOutline } from '@vicons/ionicons5'
import { computed, onMounted, ref } from 'vue'
import type { ClusterEvent } from '../../types'
import ClusterEventDetailDialog from './ClusterEventDetailDialog.vue'
import { useClusterEventPage } from './hooks'

defineOptions({
  name: 'ClusterEventList',
})

interface Props {
  selectedEventId?: string
  showAckList?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  selectedEventId: undefined,
  showAckList: true,
})

const emit = defineEmits<{
  (e: 'select', eventId: string): void
  (e: 'toggle-ack-list'): void
}>()

const { t } = useModuleI18n('hub0008')

const splitPanes: RsSplitPaneItem[] = [
  { key: 'search', size: 'auto' },
  { key: 'grid' },
]

const searchFormRef = ref()
const gridRef = ref<RsGridExpose | null>(null)
const detailDialogVisible = ref(false)
const currentEvent = ref<ClusterEvent | null>(null)

const { service, handleSearch } = useClusterEventPage(searchFormRef)

const computedSearchFormConfig = computed(() => {
  const config = { ...service.model.searchFormConfig }
  if (config.toolbarButtons) {
    config.toolbarButtons = config.toolbarButtons.map((btn) => {
      if (btn.key === 'toggleAckList') {
        return {
          ...btn,
          label: props.showAckList
            ? t('event.toolbar.collapseAckList')
            : t('event.toolbar.expandAckList'),
          icon: props.showAckList ? ChevronForwardOutline : ChevronBackOutline,
        }
      }
      return btn
    })
  }
  return config
})

const handleRowClick = ({ row }: { row: ClusterEvent }) => {
  emit('select', row.eventId)
}

const handleToolbarClick = (key: string) => {
  if (key === 'toggleAckList') {
    emit('toggle-ack-list')
  }
}

const handleMenuClick = ({ key, row }: { key: string; row?: ClusterEvent }) => {
  if (!row) return
  if (key === 'view') {
    currentEvent.value = row
    detailDialogVisible.value = true
  }
}

onMounted(() => {
  service.loadEvents()
})
</script>

<style scoped>
.cluster-event-list {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.cluster-event-list__split {
  flex: 1 1 auto;
  width: 100%;
  height: 100%;
  min-height: 0;
}

.cluster-event-list__search {
  width: 100%;
}

.cluster-event-list__grid {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
</style>
