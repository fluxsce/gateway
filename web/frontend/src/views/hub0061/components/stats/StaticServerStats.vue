<template>
  <div class="static-server-stats">
    <n-grid :cols="6" :x-gap="8" :y-gap="8" responsive="screen">
      <!-- 总服务数 -->
      <n-gi>
        <n-card class="stat-card" hoverable>
          <div class="stat-content">
            <div class="stat-icon total">
              <n-icon size="18" :component="ServerOutline" />
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ statistics.totalServers || 0 }}</div>
              <div class="stat-label">总服务数</div>
            </div>
          </div>
        </n-card>
      </n-gi>

      <!-- 运行中服务 -->
      <n-gi>
        <n-card class="stat-card" hoverable>
          <div class="stat-content">
            <div class="stat-icon running">
              <n-icon size="18" :component="PlayOutline" />
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ statistics.runningServers || 0 }}</div>
              <div class="stat-label">运行中</div>
            </div>
          </div>
        </n-card>
      </n-gi>

      <!-- 已停止服务 -->
      <n-gi>
        <n-card class="stat-card" hoverable>
          <div class="stat-content">
            <div class="stat-icon stopped">
              <n-icon size="18" :component="StopCircleOutline" />
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ statistics.stoppedServers || 0 }}</div>
              <div class="stat-label">已停止</div>
            </div>
          </div>
        </n-card>
      </n-gi>

      <!-- 总连接数 -->
      <n-gi>
        <n-card class="stat-card" hoverable>
          <div class="stat-content">
            <div class="stat-icon connections">
              <n-icon size="18" :component="LinkOutline" />
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ statistics.totalConnections || 0 }}</div>
              <div class="stat-label">总连接数</div>
            </div>
          </div>
        </n-card>
      </n-gi>

      <!-- 接收流量 -->
      <n-gi>
        <n-card class="stat-card" hoverable>
          <div class="stat-content">
            <div class="stat-icon received">
              <n-icon size="18" :component="ArrowDownOutline" />
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ formatTraffic(statistics.totalBytesReceived) }}</div>
              <div class="stat-label">接收流量</div>
            </div>
          </div>
        </n-card>
      </n-gi>

      <!-- 发送流量 -->
      <n-gi>
        <n-card class="stat-card" hoverable>
          <div class="stat-content">
            <div class="stat-icon sent">
              <n-icon size="18" :component="ArrowUpOutline" />
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ formatTraffic(statistics.totalBytesSent) }}</div>
              <div class="stat-label">发送流量</div>
            </div>
          </div>
        </n-card>
      </n-gi>
    </n-grid>
  </div>
</template>

<script setup lang="ts">
import {
  ArrowDownOutline,
  ArrowUpOutline,
  LinkOutline,
  PlayOutline,
  ServerOutline,
  StopCircleOutline
} from '@vicons/ionicons5';
import {
  NCard,
  NGi,
  NGrid,
  NIcon
} from 'naive-ui';
import type { StaticServerStats } from './types';

interface Props {
  statistics: StaticServerStats
}

defineProps<Props>()

// 格式化流量
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

.stat-icon.running {
  background: linear-gradient(135deg, #52c41a 0%, #73d13d 100%);
  color: white;
}

.stat-icon.stopped {
  background: linear-gradient(135deg, #faad14 0%, #ffc53d 100%);
  color: white;
}

.stat-icon.connections {
  background: linear-gradient(135deg, #13c2c2 0%, #36cfc9 100%);
  color: white;
}

.stat-icon.received {
  background: linear-gradient(135deg, #1890ff 0%, #40a9ff 100%);
  color: white;
}

.stat-icon.sent {
  background: linear-gradient(135deg, #722ed1 0%, #9254de 100%);
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
