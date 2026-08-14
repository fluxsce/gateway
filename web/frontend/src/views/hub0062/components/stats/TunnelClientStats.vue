<template>
  <div class="tunnel-client-stats">
    <div class="tunnel-client-stats__grid">
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
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { RsStatCard } from '@/ui'
import { computed } from 'vue'
import type { TunnelClientStats } from '../../types'

defineOptions({
  name: 'TunnelClientStats',
})

interface Props {
  statistics: TunnelClientStats
}

const props = defineProps<Props>()
const { t } = useModuleI18n('hub0062')

const statsCards = computed(() => [
  {
    key: 'total',
    label: t('stats.total'),
    value: props.statistics.totalClients || 0,
    accent: 'primary' as const,
  },
  {
    key: 'connected',
    label: t('stats.connected'),
    value: props.statistics.connectedClients || 0,
    accent: 'success' as const,
  },
  {
    key: 'disconnected',
    label: t('stats.disconnected'),
    value: props.statistics.disconnectedClients || 0,
    accent: 'warning' as const,
  },
  {
    key: 'connecting',
    label: t('stats.connecting'),
    value: props.statistics.connectingClients || 0,
    accent: 'info' as const,
  },
  {
    key: 'error',
    label: t('stats.error'),
    value: props.statistics.errorClients || 0,
    accent: 'danger' as const,
  },
  {
    key: 'services',
    label: t('stats.services'),
    value: props.statistics.totalServices || 0,
    accent: 'info' as const,
  },
])
</script>

<style scoped>
.tunnel-client-stats {
  width: 100%;
  padding: var(--g-space-sm) 0;
}

.tunnel-client-stats__grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  gap: 8px;
}
</style>
