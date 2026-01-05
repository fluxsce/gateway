<template>
  <div class="app-monitor">
    <!-- 组件类型标签页 -->
    <n-tabs v-model:value="activeTab" type="card" animated>
      <!-- 线程池 -->
      <n-tab-pane name="threadpool" :tab="t('threadPool')">
        <thread-pool-monitor :jvm-resource-id="jvmResourceId" />
      </n-tab-pane>

      <!-- HTTP连接池 -->
      <n-tab-pane name="httppool" :tab="t('httpConnectionPool')">
        <http-connection-pool-monitor :jvm-resource-id="jvmResourceId" />
      </n-tab-pane>

      <!-- Druid数据源 -->
      <n-tab-pane name="druid" :tab="t('druidDatasource')">
        <druid-datasource-monitor :jvm-resource-id="jvmResourceId" />
      </n-tab-pane>
    </n-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import ThreadPoolMonitor from './ThreadPoolMonitor.vue'
import HttpConnectionPoolMonitor from './HttpConnectionPoolMonitor.vue'
import DruidDatasourceMonitor from './DruidDatasourceMonitor.vue'

defineProps<{
  jvmResourceId: string
}>()

const { t } = useModuleI18n('hub0042')

const activeTab = ref('threadpool')
</script>

<style scoped lang="scss">
.app-monitor {
  width: 100%;
}
</style>

