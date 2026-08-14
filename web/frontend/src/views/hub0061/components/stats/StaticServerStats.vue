<template>
  <div class="static-server-stats">
    <RsStatCard label="总服务数" :value="statistics.totalServers || 0" accent="primary" />
    <RsStatCard label="运行中" :value="statistics.runningServers || 0" accent="success" />
    <RsStatCard label="已停止" :value="statistics.stoppedServers || 0" accent="warning" />
    <RsStatCard label="总连接数" :value="statistics.totalConnections || 0" accent="info" />
    <RsStatCard label="接收流量" :value="formatTraffic(statistics.totalBytesReceived)" accent="info" />
    <RsStatCard label="发送流量" :value="formatTraffic(statistics.totalBytesSent)" accent="primary" />
  </div>
</template>

<script setup lang="ts">
import { RsStatCard } from '@/ui'
import type { StaticServerStats } from './types'

interface Props {
  /** 静态服务统计数据 */
  statistics: StaticServerStats
}

defineProps<Props>()

/**
 * 将字节数格式化为可读流量文案。
 * @param bytes - 字节数
 * @returns 带单位的流量字符串
 */
const formatTraffic = (bytes: number | undefined): string => {
  if (!bytes || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i]
}
</script>

<style scoped>
.static-server-stats {
  display: grid;
  grid-template-columns: repeat(6, minmax(0, 1fr));
  gap: 8px;
  width: 100%;
  padding: var(--g-space-sm) 0;
}

@media (max-width: 1200px) {
  .static-server-stats {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}
</style>
