<template>
  <div class="tunnel-server-stats">
    <div class="stats-grid stats-grid--summary">
      <RsCard class="stat-card" size="sm" variant="outlined" hoverable>
        <div class="stat-content">
          <div class="stat-icon total">
            <RsIcon name="server" :size="18" />
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ statistics.totalServers || 1 }}</div>
            <div class="stat-label">总服务器数</div>
          </div>
        </div>
      </RsCard>

      <RsCard class="stat-card" size="sm" variant="outlined" hoverable>
        <div class="stat-content">
          <div class="stat-icon running">
            <RsIcon name="play" :size="18" />
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ statistics.runningServers || 0 }}</div>
            <div class="stat-label">运行中</div>
          </div>
        </div>
      </RsCard>

      <RsCard class="stat-card" size="sm" variant="outlined" hoverable>
        <div class="stat-content">
          <div class="stat-icon stopped">
            <RsIcon name="square" :size="18" />
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ statistics.stoppedServers || 0 }}</div>
            <div class="stat-label">已停止</div>
          </div>
        </div>
      </RsCard>

      <RsCard class="stat-card" size="sm" variant="outlined" hoverable>
        <div class="stat-content">
          <div class="stat-icon error">
            <RsIcon name="circle-alert" :size="18" />
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ statistics.errorServers || 0 }}</div>
            <div class="stat-label">错误状态</div>
          </div>
        </div>
      </RsCard>
    </div>

    <div class="stats-grid stats-grid--detail">
      <RsCard class="detail-card" size="sm" variant="outlined" hoverable>
        <template #header>
          <div class="card-header">
            <RsIcon name="users" :size="16" />
            <span>客户端统计</span>
          </div>
        </template>
        <div class="detail-content">
          <div class="detail-item">
            <div class="detail-label">总客户端数</div>
            <div class="detail-value primary">{{ statistics.totalClients || 0 }}</div>
          </div>
          <div class="detail-progress">
            <div class="progress-track">
              <div
                class="progress-bar"
                :style="{ width: `${clientUsagePercentage}%`, background: getProgressColor(clientUsagePercentage) }"
              />
            </div>
            <div class="progress-text">{{ clientUsagePercentage }}% 使用率</div>
          </div>
        </div>
      </RsCard>

      <RsCard class="detail-card" size="sm" variant="outlined" hoverable>
        <template #header>
          <div class="card-header">
            <RsIcon name="link" :size="16" />
            <span>连接统计</span>
          </div>
        </template>
        <div class="detail-content">
          <div class="detail-item">
            <div class="detail-label">总连接数</div>
            <div class="detail-value success">{{ statistics.totalConnections || 0 }}</div>
          </div>
          <div class="detail-progress">
            <div class="progress-track">
              <div
                class="progress-bar"
                :style="{ width: `${connectionUsagePercentage}%`, background: getProgressColor(connectionUsagePercentage) }"
              />
            </div>
            <div class="progress-text">{{ connectionUsagePercentage }}% 负载</div>
          </div>
        </div>
      </RsCard>
    </div>
  </div>
</template>

<script setup lang="ts">
import { RsCard, RsIcon } from '@/ui'
import { computed } from 'vue'
import type { TunnelServerStats } from '../../types/index'

defineOptions({
  name: 'TunnelServerStats',
})

interface Props {
  statistics: TunnelServerStats
}

const props = defineProps<Props>()

const clientUsagePercentage = computed(() => {
  const maxClients = 1000
  const clients = props.statistics.totalClients || 0
  return Math.min(Math.round((clients / maxClients) * 100), 100)
})

const connectionUsagePercentage = computed(() => {
  const maxConnections = 5000
  const connections = props.statistics.totalConnections || 0
  return Math.min(Math.round((connections / maxConnections) * 100), 100)
})

const getProgressColor = (percentage: number) => {
  if (percentage < 50) return '#52c41a'
  if (percentage < 80) return '#faad14'
  return '#ff4d4f'
}
</script>

<style scoped>
.tunnel-server-stats {
  width: 100%;
  padding: var(--g-space-sm) 0;
}

.stats-grid {
  display: grid;
  gap: 8px;
}

.stats-grid--summary {
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.stats-grid--detail {
  grid-template-columns: repeat(2, minmax(0, 1fr));
  margin-top: 8px;
}

.stat-card {
  cursor: pointer;
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
  color: #fff;
}

.stat-icon.total {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.stat-icon.running {
  background: linear-gradient(135deg, #52c41a 0%, #73d13d 100%);
}

.stat-icon.stopped {
  background: linear-gradient(135deg, #faad14 0%, #ffc53d 100%);
}

.stat-icon.error {
  background: linear-gradient(135deg, #ff4d4f 0%, #ff7875 100%);
}

.stat-info {
  flex: 1;
}

.stat-value {
  font-size: 20px;
  font-weight: 700;
  line-height: 1;
  margin-bottom: 2px;
  color: var(--rs-text, var(--n-text-color));
}

.stat-label {
  font-size: 11px;
  color: var(--rs-muted, var(--n-text-color-2));
  font-weight: 500;
}

.detail-card {
  min-height: 90px;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 6px;
  font-weight: 600;
  font-size: 13px;
}

.detail-content {
  padding: 8px 0;
}

.detail-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6px;
}

.detail-label {
  font-size: 12px;
  color: var(--rs-muted, var(--n-text-color-2));
}

.detail-value {
  font-size: 18px;
  font-weight: 700;
}

.detail-value.primary {
  color: #1890ff;
}

.detail-value.success {
  color: #52c41a;
}

.progress-track {
  height: 6px;
  overflow: hidden;
  border-radius: 999px;
  background: var(--rs-border, rgba(0, 0, 0, 0.08));
}

.progress-bar {
  height: 100%;
  border-radius: inherit;
  transition: width 0.2s ease;
}

.progress-text {
  font-size: 11px;
  color: var(--rs-muted, var(--n-text-color-2));
  text-align: center;
  margin-top: 3px;
}
</style>
