<template>
  <div class="tunnel-management" :id="moduleId">
    <n-tabs type="line" placement="left" class="management-tabs">
      <!-- 客户端管理标签页 -->
      <n-tab-pane name="client" tab="客户端管理">
        <TunnelClientList />
      </n-tab-pane>

      <!-- 服务管理标签页 -->
      <n-tab-pane name="service" tab="服务管理">
        <TunnelServiceManagement />
      </n-tab-pane>
    </n-tabs>
  </div>
</template>

<script lang="ts" setup>
import { NTabs } from 'naive-ui'
import { TunnelServiceManagement } from './components/client-service'
import { TunnelClientList } from './components/tunnel-client'

// 定义组件名称
defineOptions({
  name: 'TunnelClientManagement'
})

// ============= 获取模块ID（用于样式作用域） =============

const moduleId = 'hub0062'
</script>

<style lang="scss" scoped>
.tunnel-management {
  width: 100%;
  height: 100%;
  overflow: hidden;

  .management-tabs {
    height: 100%;
    :deep(.n-tabs-content) {
      height: 100%;
    }

    :deep(.n-tab-pane) {
      height: 100%;
      width: 100%;
      padding: 0px !important;
      min-width: 0; /* 允许 flex 子元素收缩 */
    }
    
    /* 确保 tabs 内容区域宽度正确（placement="left" 时） */
    :deep(.n-tabs-pane-wrapper) {
      width: 100%;
      height: 100%;
    }
    
  }
}
</style>

