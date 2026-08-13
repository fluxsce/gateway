<template>
  <div class="cluster-event-management" :id="moduleId">
    <RsSplitPane
      ref="splitRef"
      class="cluster-event-management__split"
      orientation="horizontal"
      :panes="splitPanes"
      with-handle
      v-model:sizes="sizes"
    >
      <!-- 左侧：集群事件列表 -->
      <template #events>
        <div class="cluster-event-management__pane">
          <ClusterEventList
            :selected-event-id="selectedEventId"
            :show-ack-list="showAckList"
            @select="handleEventSelect"
            @toggle-ack-list="handleToggleAckList"
          />
        </div>
      </template>

      <!-- 右侧：事件处理节点列表 -->
      <template #acks="{ collapsed }">
        <div v-show="!collapsed" class="cluster-event-management__pane">
          <ClusterEventAckList :event-id="selectedEventId" />
        </div>
      </template>
    </RsSplitPane>
  </div>
</template>

<script setup lang="ts">
import { RsSplitPane, type RsSplitPaneItem } from '@/ui'
import { computed, ref } from 'vue'
import ClusterEventAckList from './components/cluster-event-ack-list/ClusterEventAckList.vue'
import ClusterEventList from './components/cluster-event-list/ClusterEventList.vue'

defineOptions({
  name: 'ClusterEventManagement',
})

const moduleId = 'cluster-event-management'

/** 左右各半；右侧可折叠收起处理列表 */
const splitPanes: RsSplitPaneItem[] = [
  { key: 'events', size: 50, min: 20 },
  { key: 'acks', size: 50, min: 20, collapsible: true, collapsedSize: 0 },
]

type SplitExpose = {
  collapse: (key: string) => void
  expand: (key: string) => void
}

const splitRef = ref<SplitExpose | null>(null)
const selectedEventId = ref('')
/** 百分比尺寸；右侧接近 0 视为已折叠 */
const sizes = ref<number[]>([50, 50])
const showAckList = computed(() => (sizes.value[1] ?? 0) > 0.01)

function handleEventSelect(eventId: string) {
  selectedEventId.value = eventId
}

function handleToggleAckList() {
  if (showAckList.value) {
    splitRef.value?.collapse('acks')
  } else {
    splitRef.value?.expand('acks')
  }
}
</script>

<style lang="scss" scoped>
.cluster-event-management {
  width: 100%;
  height: 100%;
  overflow: hidden;
  display: flex;
  flex-direction: column;

  &__split {
    flex: 1 1 auto;
    width: 100%;
    height: 100%;
    min-height: 0;
  }

  &__pane {
    width: 100%;
    height: 100%;
    min-width: 0;
    min-height: 0;
    overflow: hidden;
  }
}
</style>
