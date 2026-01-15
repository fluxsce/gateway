<template>
  <div class="cluster-event-management" :id="moduleId">
    <GPane ref="paneRef" direction="horizontal" :default-size="0.5">
      <!-- 左侧：集群事件列表 -->
      <template #1>
        <ClusterEventList
          :selected-event-id="selectedEventId"
          :show-ack-list="showAckList"
          @select="handleEventSelect"
          @toggle-ack-list="handleToggleAckList"
        />
      </template>

      <!-- 右侧：事件处理节点列表 -->
      <template #2>
        <ClusterEventAckList :event-id="selectedEventId" />
      </template>
    </GPane>
  </div>
</template>

<script setup lang="ts">
import { GPane } from '@/components/gpane'
import type { GPaneExpose } from '@/components/gpane/types'
import { computed, ref } from 'vue'
import ClusterEventAckList from './components/cluster-event-ack-list/ClusterEventAckList.vue'
import ClusterEventList from './components/cluster-event-list/ClusterEventList.vue'

// 定义组件名称
defineOptions({
  name: 'ClusterEventManagement'
})

// 模块ID
const moduleId = 'cluster-event-management'

// 状态管理
const selectedEventId = ref<string>('')
const paneRef = ref<GPaneExpose>()

// 计算面板二的可见性（用于传递给子组件）
const showAckList = computed(() => {
  return paneRef.value?.getPane2Visible() ?? true
})

// 处理事件选择
function handleEventSelect(eventId: string) {
  selectedEventId.value = eventId
}

// 处理折叠/展开处理列表
function handleToggleAckList() {
  paneRef.value?.togglePane2Visible()
}
</script>

<style lang="scss" scoped>
.cluster-event-management {
  width: 100%;
  height: 100%;
  overflow: hidden;

  :deep(.n-split) {
    width: 100%;
    height: 100%;
  }
}
</style>

