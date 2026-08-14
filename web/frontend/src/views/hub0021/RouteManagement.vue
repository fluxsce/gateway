<template>
  <div class="route-management" :id="moduleId">
    <RsSplitPane
      class="route-management__split"
      orientation="horizontal"
      :panes="splitPanes"
      with-handle
    >
      <template #tree>
        <div class="route-management__tree">
          <GatewayInstanceTree
            :parent-module-id="moduleId"
            @select="handleInstanceSelect"
          />
        </div>
      </template>

      <template #routes>
        <div class="route-management__routes">
          <RouteConfigList
            :gateway-instance-id="selectedGatewayInstanceId"
            :key="`routes-${selectedGatewayInstanceId}`"
          />
        </div>
      </template>
    </RsSplitPane>
  </div>
</template>

<script setup lang="ts">
import { RsSplitPane, type RsSplitPaneItem } from '@/ui'
import { computed, ref } from 'vue'
import type { GatewayInstance } from './components/instance-tree'
import { GatewayInstanceTree } from './components/instance-tree'
import { RouteConfigList } from './components/routes'

defineOptions({
  name: 'RouteManagement',
})

const moduleId = 'route-management'

/** 左侧约 25%，最大 40%；右侧占满剩余空间 */
const splitPanes: RsSplitPaneItem[] = [
  { key: 'tree', size: 25, min: 15, max: 40 },
  { key: 'routes', min: 40 },
]

const selectedInstanceId = ref<string>('')
const selectedGatewayInstanceId = computed(() => selectedInstanceId.value || '')

function handleInstanceSelect(instanceId: string, _instance: GatewayInstance) {
  selectedInstanceId.value = instanceId
}
</script>

<style lang="scss" scoped>
.route-management {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.route-management__split {
  flex: 1 1 auto;
  width: 100%;
  height: 100%;
  min-height: 0;
}

.route-management__tree,
.route-management__routes {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
</style>
