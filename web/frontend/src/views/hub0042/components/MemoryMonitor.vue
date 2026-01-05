<template>
  <div class="memory-monitor">
    <!-- 内存使用趋势图表 -->
    <n-card :title="t('memoryUsageTrend')" :bordered="false">
      <n-tabs type="line" animated>
        <n-tab-pane :name="'heap'" :tab="t('heapMemory')">
          <heap-memory-chart :jvm-resource-id="jvmResourceId" />
        </n-tab-pane>
        
        <n-tab-pane :name="'nonheap'" :tab="t('nonHeapMemory')">
          <non-heap-memory-chart :jvm-resource-id="jvmResourceId" />
        </n-tab-pane>
      </n-tabs>
    </n-card>

    <!-- 内存池使用趋势 -->
    <n-card :title="t('memoryPoolUsageTrend')" :bordered="false" style="margin-top: 16px">
      <memory-pool-chart :jvm-resource-id="jvmResourceId" />
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { useModuleI18n } from '@/hooks/useModuleI18n'
import HeapMemoryChart from './HeapMemoryChart.vue'
import NonHeapMemoryChart from './NonHeapMemoryChart.vue'
import MemoryPoolChart from './MemoryPoolChart.vue'

defineProps<{
  jvmResourceId: string
}>()

const { t } = useModuleI18n('hub0042')
</script>

<style scoped lang="scss">
// 样式文件
</style>
