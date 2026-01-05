<template>
  <div class="proxy-management" :id="moduleId">
    <GPane direction="horizontal" :default-size="0.25"  :max="0.4">
      <!-- 左侧：网关实例选择区域 -->
      <template #1>
        <div class="left-panel">
          <GatewayInstanceTree
            :parent-module-id="moduleId"
            @select="handleInstanceSelect"
          />
        </div>
      </template>

      <!-- 右侧：服务定义管理 -->
      <template #2>
        <ServiceDefinitionList 
          :gateway-instance-id="gatewayInstanceId" 
          :key="`service-${gatewayInstanceId}`"
        />
      </template>
    </GPane>
  </div>
</template>

<script setup lang="ts">
import { GPane } from '@/components/gpane'
import { computed, ref } from 'vue'
import type { GatewayInstance } from './components/instance-tree'

// 组件导入
import { GatewayInstanceTree } from './components/instance-tree'
import ServiceDefinitionList from './components/service/ServiceDefinitionList.vue'

// 定义组件名称
defineOptions({
  name: 'ProxyManagement'
})

// 模块ID
const moduleId = 'proxy-management'

// 状态管理
const selectedInstanceId = ref<string>('')
const gatewayInstanceId = computed(() => selectedInstanceId.value || '')

// 处理实例选择
function handleInstanceSelect(instanceId: string, instance: GatewayInstance) {
  selectedInstanceId.value = instanceId
}


</script>

<style lang="scss" scoped>
.proxy-management {
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
