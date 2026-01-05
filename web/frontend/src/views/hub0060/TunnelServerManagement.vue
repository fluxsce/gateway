<template>
  <div class="tunnel-management" :id="moduleId">
    <n-tabs type="line" placement="left" class="management-tabs">
      <!-- 服务器管理标签页 -->
      <n-tab-pane name="server" tab="隧道服务器管理">
        <TunnelServerList />
      </n-tab-pane>

      <!-- 客户端注册列表标签页 -->
      <n-tab-pane name="clients" tab="客户端注册列表">
        <RegistClientList :tunnel-server-id="selectedServerId || ''" />
      </n-tab-pane>

      <!-- 服务注册列表标签页 -->
      <n-tab-pane name="services" tab="服务注册列表">
        <RegistServiceList :tunnel-server-id="selectedServerId || ''" />
      </n-tab-pane>
    </n-tabs>
  </div>
</template>

<script lang="ts" setup>
import { NTabPane, NTabs } from 'naive-ui'
import { ref } from 'vue'
import { RegistClientList } from './components/regist-client'
import { RegistServiceList } from './components/regist-service'
import { TunnelServerList } from './components/tunnel-server'

// 定义组件名称
defineOptions({
  name: 'TunnelServerManagement'
})

// ============= 获取模块ID（用于样式作用域） =============

const moduleId = 'hub0060'

// 选中的服务器ID（可以从服务器管理组件中获取）
const selectedServerId = ref<string>('')
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

