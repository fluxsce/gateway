<template>
  <div class="tunnel-service-stats">
    <div class="tunnel-service-stats__grid">
      <RsStatCard
        v-for="card in statsCards"
        :key="card.key"
        :label="card.label"
        :value="card.value"
        :accent="card.accent"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { RsStatCard } from '@/ui'
import { computed } from 'vue'
import type { TunnelServiceStats } from '../../types'

defineOptions({
  name: 'TunnelServiceStats',
})

interface Props {
  statistics: TunnelServiceStats
}

const props = defineProps<Props>()

const statsCards = computed(() => [
  {
    key: 'total',
    label: '总服务数',
    value: props.statistics.totalServices || 0,
    accent: 'primary' as const,
  },
  {
    key: 'active',
    label: '活动服务',
    value: props.statistics.activeServices || 0,
    accent: 'success' as const,
  },
  {
    key: 'inactive',
    label: '不活动服务',
    value: props.statistics.inactiveServices || 0,
    accent: 'warning' as const,
  },
  {
    key: 'error',
    label: '错误服务',
    value: props.statistics.errorServices || 0,
    accent: 'danger' as const,
  },
  {
    key: 'offline',
    label: '离线服务',
    value: props.statistics.offlineServices || 0,
    accent: 'info' as const,
  },
  {
    key: 'connections',
    label: '总连接数',
    value: props.statistics.totalConnections || 0,
    accent: 'info' as const,
  },
])
</script>

<style scoped>
.tunnel-service-stats {
  width: 100%;
  padding: var(--g-space-sm) 0;
}

.tunnel-service-stats__grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  gap: 8px;
}
</style>
