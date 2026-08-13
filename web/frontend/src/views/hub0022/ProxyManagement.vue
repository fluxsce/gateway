<template>
  <div class="proxy-management" :id="moduleId">
    <RsSplitPane
      class="proxy-management__split"
      orientation="horizontal"
      :panes="splitPanes"
      with-handle
    >
      <!-- 左侧：网关实例选择区域 -->
      <template #tree>
        <div class="proxy-management__tree">
          <GatewayInstanceTree
            :parent-module-id="moduleId"
            @select="handleInstanceSelect"
          />
        </div>
      </template>

      <!-- 右侧：服务定义管理 -->
      <template #services>
        <div class="proxy-management__services">
          <ServiceDefinitionList
            :gateway-instance-id="gatewayInstanceId"
            :key="`service-${gatewayInstanceId}`"
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
import ServiceDefinitionList from './components/service/ServiceDefinitionList.vue'

defineOptions({
  name: 'ProxyManagement'
})

const moduleId = 'proxy-management'

/** 左侧约 25%，最大 40%；右侧占满剩余空间 */
const splitPanes: RsSplitPaneItem[] = [
  { key: 'tree', size: 25, min: 15, max: 40 },
  { key: 'services', min: 40 },
]

const selectedInstanceId = ref<string>('')
const gatewayInstanceId = computed(() => selectedInstanceId.value || '')

function handleInstanceSelect(instanceId: string, _instance: GatewayInstance) {
  selectedInstanceId.value = instanceId
}
</script>

<style lang="scss" scoped>
.proxy-management {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.proxy-management__split {
  flex: 1 1 auto;
  width: 100%;
  height: 100%;
  min-height: 0;
}

.proxy-management__tree,
.proxy-management__services {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
</style>
