<template>
  <div class="route-management" :id="moduleId">
    <GPane direction="horizontal" :default-size="0.25" :max="0.4">
      <!-- 左侧：网关实例树 -->
      <template #1>
        <div class="left-panel">
          <GatewayInstanceTree
            :parent-module-id="moduleId"
            @select="handleInstanceSelect"
          />
        </div>
      </template>

      <!-- 右侧：路由配置管理 -->
      <template #2>
        <RouteConfigList 
          :gateway-instance-id="selectedGatewayInstanceId" 
          :key="`routes-${selectedGatewayInstanceId}`"
        />
      </template>
    </GPane>
  </div>
</template>

<script setup lang="ts">
import { GPane } from '@/components/gpane'
import { computed, ref } from 'vue'
import type { GatewayInstance } from './components/instance-tree'
import { GatewayInstanceTree } from './components/instance-tree'
import { RouteConfigList } from './components/routes'

// 定义组件名称
defineOptions({
  name: 'RouteManagement'
})

// 模块ID
const moduleId = 'route-management'

// 状态管理
const selectedInstanceId = ref<string>('')
const selectedGatewayInstanceId = computed(() => selectedInstanceId.value || '')

// 处理实例选择
function handleInstanceSelect(instanceId: string, instance: GatewayInstance) {
  selectedInstanceId.value = instanceId
}
</script>

<style lang="scss" scoped>
.route-management {
  width: 100%;
  height: 100%;
  overflow: hidden;

  :deep(.n-split) {
    width: 100%;
    height: 100%;
  }
}

.left-panel {
  display: flex;
  flex-direction: column;
  height: 100%;
}
</style>
