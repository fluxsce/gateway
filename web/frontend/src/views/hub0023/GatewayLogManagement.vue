<template>
  <div class="gateway-log-management" :id="service.model.moduleId">
    <n-tabs type="line" placement="left" class="management-tabs">
      <!-- 日志查询标签页 -->
      <n-tab-pane name="logs" tab="日志查询">
        <GatewayLogQuery />
      </n-tab-pane>

      <!-- 监控图表标签页 -->
      <n-tab-pane name="monitor" tab="监控图表">
        <MonitoringPanel />
      </n-tab-pane>
    </n-tabs>
  </div>
</template>

<script lang="ts" setup>
import { NTabs } from 'naive-ui'
import { GatewayLogQuery, MonitoringPanel } from './components'
import { useGatewayLogModel } from './components/gateway-log/hooks/model'

// 定义组件名称
defineOptions({
  name: 'GatewayLogManagement'
})

// ============= 获取模块ID（用于样式作用域） =============

const model = useGatewayLogModel()
const service = { model }
</script>

<style lang="scss" scoped>
.gateway-log-management {
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
