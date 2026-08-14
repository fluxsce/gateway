<template>
  <div class="tunnel-management" :id="moduleId">
    <RsTabs
      v-model="activeTab"
      :items="tabItems"
      variant="line"
      size="md"
      panelless
      class="management-tabs"
    />
    <div class="management-tabs__content">
      <TunnelServerList v-show="activeTab === 'server'" />
      <RegistClientList v-show="activeTab === 'clients'" :tunnel-server-id="selectedServerId || ''" />
      <RegistServiceList v-show="activeTab === 'services'" :tunnel-server-id="selectedServerId || ''" />
    </div>
  </div>
</template>

<script lang="ts" setup>
import { RsTabs, type RsTabItem } from '@/ui'
import { ref } from 'vue'
import { RegistClientList } from './components/regist-client'
import { RegistServiceList } from './components/regist-service'
import { TunnelServerList } from './components/tunnel-server'

defineOptions({
  name: 'TunnelServerManagement',
})

const moduleId = 'hub0060'
const selectedServerId = ref<string>('')
const activeTab = ref('server')
const tabItems: RsTabItem[] = [
  { value: 'server', label: '隧道服务器管理' },
  { value: 'clients', label: '客户端注册列表' },
  { value: 'services', label: '服务注册列表' },
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
