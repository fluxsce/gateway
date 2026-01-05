<template>
  <div class="tunnel-client-stats">
    <n-grid :cols="6" :x-gap="8" :y-gap="8" responsive="screen">
      <n-gi v-for="card in statsCards" :key="card.key">
        <n-card class="stat-card" hoverable>
          <div class="stat-content">
            <div :class="['stat-icon', card.key]">
              <n-icon size="18" :component="card.icon" />
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ card.value }}</div>
              <div class="stat-label">{{ card.label }}</div>
            </div>
          </div>
        </n-card>
      </n-gi>
    </n-grid>
  </div>
</template>

<script setup lang="ts">
import { useModuleI18n } from '@/hooks/useModuleI18n'
import {
  AlertCircleOutline,
  CheckmarkCircleOutline,
  CloseCircleOutline,
  LinkOutline,
  ServerOutline
} from '@vicons/ionicons5'
import { NCard, NGi, NGrid, NIcon } from 'naive-ui'
import { computed } from 'vue'
import type { TunnelClientStats } from '../../types'

interface Props {
  statistics: TunnelClientStats
}

const props = defineProps<Props>()
const { t } = useModuleI18n('hub0062')

// 统计卡片配置
const statsCards = computed(() => [
  {
    key: 'total',
    label: t('stats.total'),
    value: props.statistics.totalClients || 0,
    icon: ServerOutline
  },
  {
    key: 'connected',
    label: t('stats.connected'),
    value: props.statistics.connectedClients || 0,
    icon: CheckmarkCircleOutline
  },
  {
    key: 'disconnected',
    label: t('stats.disconnected'),
    value: props.statistics.disconnectedClients || 0,
    icon: CloseCircleOutline
  },
  {
    key: 'connecting',
    label: t('stats.connecting'),
    value: props.statistics.connectingClients || 0,
    icon: AlertCircleOutline
  },
  {
    key: 'error',
    label: t('stats.error'),
    value: props.statistics.errorClients || 0,
    icon: CloseCircleOutline
  },
  {
    key: 'services',
    label: t('stats.services'),
    value: props.statistics.totalServices || 0,
    icon: LinkOutline
  }
])
</script>

<style scoped>
.tunnel-client-stats {
  width: 100%;
  padding: var(--g-space-sm) 0px;
}

.stat-card {
  cursor: pointer;
  transition: all 0.3s ease;
}

.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
}

:deep(.stat-card .n-card__content) {
  padding: 12px;
  height: 100%;
  display: flex;
  align-items: center;
}

.stat-content {
  display: flex;
  align-items: center;
  width: 100%;
}

.stat-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: 8px;
  margin-right: 10px;
}

.stat-icon.total {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.stat-icon.connected {
  background: linear-gradient(135deg, #52c41a 0%, #73d13d 100%);
  color: white;
}

.stat-icon.disconnected {
  background: linear-gradient(135deg, #faad14 0%, #ffc53d 100%);
  color: white;
}

.stat-icon.connecting {
  background: linear-gradient(135deg, #13c2c2 0%, #36cfc9 100%);
  color: white;
}

.stat-icon.error {
  background: linear-gradient(135deg, #ff4d4f 0%, #ff7875 100%);
  color: white;
}

.stat-icon.services {
  background: linear-gradient(135deg, #1890ff 0%, #40a9ff 100%);
  color: white;
}

.stat-info {
  flex: 1;
}

.stat-value {
  font-size: 18px;
  font-weight: 700;
  line-height: 1;
  margin-bottom: 2px;
  color: var(--n-text-color);
}

.stat-label {
  font-size: 11px;
  color: var(--n-text-color-2);
  font-weight: 500;
}
</style>

