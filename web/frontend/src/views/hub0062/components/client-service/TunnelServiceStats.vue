<template>
  <div class="tunnel-service-stats">
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
    CheckmarkCircleOutline,
    CloseCircleOutline,
    CloudOfflineOutline,
    CloudOutline,
    GitNetworkOutline,
    PauseCircleOutline
} from '@vicons/ionicons5'
import { NCard, NGi, NGrid, NIcon } from 'naive-ui'
import { computed } from 'vue'
import type { TunnelServiceStats } from '../../types'

interface Props {
  statistics: TunnelServiceStats
}

const props = defineProps<Props>()
const { t } = useModuleI18n('hub0062')

// 统计卡片配置
const statsCards = computed(() => [
  {
    key: 'total',
    label: '总服务数',
    value: props.statistics.totalServices || 0,
    icon: CloudOutline
  },
  {
    key: 'active',
    label: '活动服务',
    value: props.statistics.activeServices || 0,
    icon: CheckmarkCircleOutline
  },
  {
    key: 'inactive',
    label: '不活动服务',
    value: props.statistics.inactiveServices || 0,
    icon: PauseCircleOutline
  },
  {
    key: 'error',
    label: '错误服务',
    value: props.statistics.errorServices || 0,
    icon: CloseCircleOutline
  },
  {
    key: 'offline',
    label: '离线服务',
    value: props.statistics.offlineServices || 0,
    icon: CloudOfflineOutline
  },
  {
    key: 'connections',
    label: '总连接数',
    value: props.statistics.totalConnections || 0,
    icon: GitNetworkOutline
  }
])
</script>

<style scoped>
.tunnel-service-stats {
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

.stat-icon.active {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
  color: white;
}

.stat-icon.inactive {
  background: linear-gradient(135deg, #ffecd2 0%, #fcb69f 100%);
  color: white;
}

.stat-icon.error {
  background: linear-gradient(135deg, #ff9a9e 0%, #fecfef 100%);
  color: white;
}

.stat-icon.offline {
  background: linear-gradient(135deg, #a8edea 0%, #fed6e3 100%);
  color: white;
}

.stat-icon.connections {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
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
