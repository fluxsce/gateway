<template>
  <div class="tunnel-management" :id="moduleId">
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
      <TunnelClientList v-show="activeTab === 'client'" />
      <TunnelServiceManagement v-show="activeTab === 'service'" />
    </div>
  </div>
</template>

<script lang="ts" setup>
import { RsTabs, type RsTabItem } from '@/ui'
import { ref } from 'vue'
import { TunnelServiceManagement } from './components/client-service'
import { TunnelClientList } from './components/tunnel-client'

defineOptions({
  name: 'TunnelClientManagement',
})

const moduleId = 'hub0062'
const activeTab = ref('client')
const tabItems: RsTabItem[] = [
  { value: 'client', label: '客户端管理' },
  { value: 'service', label: '服务管理' },
]
</script>

<style lang="scss" scoped>
.tunnel-management {
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
