<template>
  <div class="tunnel-server-stats">
    <n-grid :cols="4" :x-gap="8" :y-gap="8" responsive="screen">
      <!-- 总服务器数 -->
      <n-gi>
        <n-card class="stat-card" hoverable>
          <div class="stat-content">
            <div class="stat-icon total">
              <n-icon size="18" :component="ServerOutline" />
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ statistics.totalServers || 1 }}</div>
              <div class="stat-label">总服务器数</div>
            </div>
          </div>
        </n-card>
      </n-gi>

      <!-- 运行中服务器 -->
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

      <!-- 已停止服务器 -->
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

      <!-- 错误服务器 -->
      <n-gi>
        <n-card class="stat-card" hoverable>
          <div class="stat-content">
            <div class="stat-icon error">
              <n-icon size="18" :component="AlertCircleOutline" />
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ statistics.errorServers || 0 }}</div>
              <div class="stat-label">错误状态</div>
            </div>
          </div>
        </n-card>
      </n-gi>
    </n-grid>

    <!-- 详细统计信息 -->
    <n-grid :cols="2" :x-gap="8" :y-gap="8" style="margin-top: 8px" responsive="screen">
      <!-- 客户端统计 -->
      <n-gi>
        <n-card class="detail-card" hoverable>
          <template #header>
            <div class="card-header">
              <n-icon size="16" :component="PeopleOutline" />
              <span>客户端统计</span>
            </div>
          </template>
          
          <div class="detail-content">
            <div class="detail-item">
              <div class="detail-label">总客户端数</div>
              <div class="detail-value primary">{{ statistics.totalClients || 0 }}</div>
            </div>
            <div class="detail-progress">
              <n-progress 
                type="line" 
                :percentage="clientUsagePercentage" 
                :show-indicator="false"
                :height="6"
                :color="getProgressColor(clientUsagePercentage)"
              />
              <div class="progress-text">
                {{ clientUsagePercentage }}% 使用率
              </div>
            </div>
          </div>
        </n-card>
      </n-gi>

      <!-- 连接统计 -->
      <n-gi>
        <n-card class="detail-card" hoverable>
          <template #header>
            <div class="card-header">
              <n-icon size="16" :component="LinkOutline" />
              <span>连接统计</span>
            </div>
          </template>
          
          <div class="detail-content">
            <div class="detail-item">
              <div class="detail-label">总连接数</div>
              <div class="detail-value success">{{ statistics.totalConnections || 0 }}</div>
            </div>
            <div class="detail-progress">
              <n-progress 
                type="line" 
                :percentage="connectionUsagePercentage" 
                :show-indicator="false"
                :height="6"
                :color="getProgressColor(connectionUsagePercentage)"
              />
              <div class="progress-text">
                {{ connectionUsagePercentage }}% 负载
              </div>
            </div>
          </div>
        </n-card>
      </n-gi>
    </n-grid>

  </div>
</template>

<script setup lang="ts">
import {
  AlertCircleOutline,
  LinkOutline,
  PeopleOutline,
  PlayOutline,
  ServerOutline,
  StopCircleOutline
} from '@vicons/ionicons5'
import {
  NCard,
  NGi,
  NGrid,
  NIcon,
  NProgress
} from 'naive-ui'
import { computed } from 'vue'
import type { TunnelServerStats } from '../../types/index'

interface Props {
  statistics: TunnelServerStats
}

const props = defineProps<Props>()

// 客户端使用率（假设最大客户端数为1000）
const clientUsagePercentage = computed(() => {
  const maxClients = 1000 // 可以从配置中获取
  const clients = props.statistics.totalClients || 0
  return Math.min(Math.round((clients / maxClients) * 100), 100)
})

// 连接使用率（假设最大连接数为5000）
const connectionUsagePercentage = computed(() => {
  const maxConnections = 5000 // 可以从配置中获取
  const connections = props.statistics.totalConnections || 0
  return Math.min(Math.round((connections / maxConnections) * 100), 100)
})

// 获取进度条颜色
const getProgressColor = (percentage: number) => {
  if (percentage < 50) return '#52c41a' // 绿色
  if (percentage < 80) return '#faad14' // 橙色
  return '#ff4d4f' // 红色
}
</script>

<style scoped>
.tunnel-server-stats {
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

.stat-icon.error {
  background: linear-gradient(135deg, #ff4d4f 0%, #ff7875 100%);
  color: white;
}

.stat-info {
  flex: 1;
}

.stat-value {
  font-size: 20px;
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
  color: var(--n-text-color-2);
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

.detail-progress {
  margin-top: 6px;
}

.progress-text {
  font-size: 11px;
  color: var(--n-text-color-2);
  text-align: center;
  margin-top: 3px;
}


:deep(.n-card-header) {
  padding-bottom: 8px;
}

:deep(.n-card__content) {
  padding-top: 0;
}
</style>
