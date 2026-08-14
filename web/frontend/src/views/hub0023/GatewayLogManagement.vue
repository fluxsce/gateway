<template>
  <div class="gateway-log-management" :id="service.model.moduleId">
    <RsTabs
      v-model="activeTab"
      :items="tabItems"
      variant="line"
      size="md"
      borderless
      panelless
      class="management-tabs"
    />
    <div class="management-tabs__content">
      <GatewayLogQuery v-show="activeTab === 'logs'" />
      <MonitoringPanel v-if="monitorVisited" v-show="activeTab === 'monitor'" />
    </div>
  </div>
</template>

<script lang="ts" setup>
import { RsTabs, type RsTabItem } from '@/ui'
import { ref, watch } from 'vue'
import { GatewayLogQuery, MonitoringPanel } from './components'
import { useGatewayLogModel } from './components/gateway-log/hooks/model'

defineOptions({
  name: 'GatewayLogManagement',
})

const model = useGatewayLogModel()
const service = { model }
const activeTab = ref('logs')
/** 监控页首次点开再挂载，避免隐藏态初始化 ECharts（宽高为 0） */
const monitorVisited = ref(false)
watch(
  activeTab,
  (tab) => {
    if (tab === 'monitor') monitorVisited.value = true
  },
  { immediate: true },
)
const tabItems: RsTabItem[] = [
  { value: 'logs', label: '日志查询' },
  { value: 'monitor', label: '监控图表' },
]
</script>

<style lang="scss" scoped>
.gateway-log-management {
  width: 100%;
  height: 100%;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.management-tabs {
  flex: 0 0 auto;
}

.management-tabs__content {
  flex: 1 1 auto;
  min-height: 0;
  overflow: hidden;
}
</style>
