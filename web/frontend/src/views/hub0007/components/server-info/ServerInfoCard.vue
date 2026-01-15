<template>
  <n-card title="服务器信息" embedded class="server-info-card">
    <div class="overview-grid">
      <!-- 主机名 -->
      <div class="overview-item">
        <div class="overview-icon hostname">
          <n-icon size="24">
            <DatabaseOutlined />
          </n-icon>
        </div>
        <div class="overview-content">
          <div class="overview-label">主机名</div>
          <n-tooltip :show-arrow="false">
            <template #trigger>
              <div class="overview-value text-truncate">{{ serverInfo.hostname }}</div>
            </template>
            {{ serverInfo.hostname }}
          </n-tooltip>
        </div>
      </div>

      <!-- 操作系统 -->
      <div class="overview-item">
        <div class="overview-icon os">
          <n-icon size="24">
            <component :is="getOSIcon(serverInfo.osType)" />
          </n-icon>
        </div>
        <div class="overview-content">
          <div class="overview-label">操作系统</div>
          <n-tooltip :show-arrow="false">
            <template #trigger>
              <div class="overview-value text-truncate">{{ serverInfo.osType }}</div>
            </template>
            {{ serverInfo.osType }}
          </n-tooltip>
        </div>
      </div>

      <!-- 系统版本 -->
      <div class="overview-item">
        <div class="overview-icon version">
          <n-icon size="24">
            <AndroidOutlined />
          </n-icon>
        </div>
        <div class="overview-content">
          <div class="overview-label">系统版本</div>
          <n-tooltip :show-arrow="false">
            <template #trigger>
              <div class="overview-value text-truncate">{{ getShortVersion(serverInfo.osVersion) }}</div>
            </template>
            {{ serverInfo.osVersion }}
          </n-tooltip>
        </div>
      </div>

      <!-- 系统架构 -->
      <div class="overview-item">
        <div class="overview-icon architecture">
          <n-icon size="24">
            <DesktopOutlined />
          </n-icon>
        </div>
        <div class="overview-content">
          <div class="overview-label">系统架构</div>
          <n-tooltip :show-arrow="false">
            <template #trigger>
              <div class="overview-value text-truncate">{{ serverInfo.architecture }}</div>
            </template>
            {{ serverInfo.architecture }}
          </n-tooltip>
        </div>
      </div>

      <!-- 服务器类型 -->
      <div class="overview-item">
        <div class="overview-icon server-type">
          <n-icon size="24">
            <CloudServerOutlined />
          </n-icon>
        </div>
        <div class="overview-content">
          <div class="overview-label">服务器类型</div>
          <n-tooltip :show-arrow="false">
            <template #trigger>
              <div class="overview-value text-truncate">{{ getServerTypeLabel(serverInfo.serverType) }}</div>
            </template>
            {{ getServerTypeLabel(serverInfo.serverType) }}
          </n-tooltip>
        </div>
      </div>

      <!-- IP地址 -->
      <div class="overview-item">
        <div class="overview-icon ip">
          <n-icon size="24">
            <GlobalOutlined />
          </n-icon>
        </div>
        <div class="overview-content">
          <div class="overview-label">IP地址</div>
          <n-tooltip :show-arrow="false">
            <template #trigger>
              <div class="overview-value text-truncate">{{ serverInfo.ipAddress || 'N/A' }}</div>
            </template>
            {{ serverInfo.ipAddress || 'N/A' }}
          </n-tooltip>
        </div>
      </div>
    </div>
  </n-card>
</template>

<script setup lang="ts">
import {
    AndroidOutlined,
    AppleOutlined,
    CloudServerOutlined,
    DatabaseOutlined,
    DesktopOutlined,
    GlobalOutlined,
    WindowsOutlined
} from '@vicons/antd'
import { NCard, NIcon, NTooltip } from 'naive-ui'
import type { ServerInfo } from '../../types'

defineOptions({
  name: 'ServerInfoCard'
})

interface Props {
  serverInfo: ServerInfo
}

const props = defineProps<Props>()

/**
 * 服务器类型标签转换
 */
const getServerTypeLabel = (serverType?: string): string => {
  const typeMap: Record<string, string> = {
    physical: '物理机',
    virtual: '虚拟机',
    unknown: '未知'
  }
  return typeMap[serverType || 'unknown'] || '未知'
}

/**
 * 根据操作系统类型获取图标
 */
const getOSIcon = (osType: string) => {
  const osLower = osType.toLowerCase()
  if (osLower.includes('windows')) {
    return WindowsOutlined
  } else if (osLower.includes('linux')) {
    return AndroidOutlined // 使用Android图标代表Linux
  } else if (osLower.includes('mac') || osLower.includes('darwin')) {
    return AppleOutlined
  } else {
    return DesktopOutlined
  }
}

/**
 * 获取简化的系统版本信息
 */
const getShortVersion = (version: string): string => {
  if (!version) return 'N/A'

  // 针对Windows系统版本的特殊处理
  if (version.toLowerCase().includes('windows')) {
    // 提取关键信息：Windows 版本号
    const match = version.match(/Windows (\d+(?:\.\d+)?)/i)
    if (match) {
      const windowsVersion = match[1]
      // 如果有额外信息（如 Home, Pro等），也提取出来
      const editionMatch = version.match(/Windows \d+(?:\.\d+)?\s+(\w+)/i)
      if (editionMatch) {
        return `Windows ${windowsVersion} ${editionMatch[1]}`
      }
      return `Windows ${windowsVersion}`
    }
  }

  // 对于其他系统，如果版本信息太长，进行截断
  if (version.length > 20) {
    return version.substring(0, 17) + '...'
  }

  return version
}
</script>

<style lang="scss" scoped>
.server-info-card {
  margin-bottom: 16px;

  .overview-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
    gap: 16px;

    .overview-item {
      display: flex;
      align-items: center;
      gap: 12px;

      .overview-icon {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 40px;
        height: 40px;
        border-radius: 50%;
        color: #fff;
        flex-shrink: 0;

        &.hostname {
          background-color: #1890ff;
        }

        &.os {
          background-color: #52c41a;
        }

        &.version {
          background-color: #fa8c16;
        }

        &.architecture {
          background-color: #722ed1;
        }

        &.server-type {
          background-color: #eb2f96;
        }

        &.ip {
          background-color: #faad14;
        }
      }

      .overview-content {
        flex: 1;
        min-width: 0;

        .overview-label {
          font-size: 12px;
          color: #999;
          margin-bottom: 4px;
        }

        .overview-value {
          font-size: 14px;
          font-weight: 500;
        }

        .text-truncate {
          max-width: 180px;
          white-space: nowrap;
          overflow: hidden;
          text-overflow: ellipsis;
        }
      }
    }
  }
}
</style>

