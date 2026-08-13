<template>
  <RsCard :title="t('serverInfo.title')" variant="outlined" class="server-info-card">
    <div class="overview-grid">
      <div class="overview-item">
        <div class="overview-icon hostname">
          <GIcon size="24">
            <DatabaseOutlined />
          </GIcon>
        </div>
        <div class="overview-content">
          <div class="overview-label">{{ t('serverInfo.hostname') }}</div>
          <RsTooltip :content="serverInfo.hostname">
            <div class="overview-value text-truncate">{{ serverInfo.hostname }}</div>
          </RsTooltip>
        </div>
      </div>

      <div class="overview-item">
        <div class="overview-icon os">
          <GIcon size="24">
            <component :is="getOSIcon(serverInfo.osType)" />
          </GIcon>
        </div>
        <div class="overview-content">
          <div class="overview-label">{{ t('serverInfo.osType') }}</div>
          <RsTooltip :content="serverInfo.osType">
            <div class="overview-value text-truncate">{{ serverInfo.osType }}</div>
          </RsTooltip>
        </div>
      </div>

      <div class="overview-item">
        <div class="overview-icon version">
          <GIcon size="24">
            <AndroidOutlined />
          </GIcon>
        </div>
        <div class="overview-content">
          <div class="overview-label">{{ t('serverInfo.osVersion') }}</div>
          <RsTooltip :content="serverInfo.osVersion">
            <div class="overview-value text-truncate">{{ getShortVersion(serverInfo.osVersion) }}</div>
          </RsTooltip>
        </div>
      </div>

      <div class="overview-item">
        <div class="overview-icon architecture">
          <GIcon size="24">
            <DesktopOutlined />
          </GIcon>
        </div>
        <div class="overview-content">
          <div class="overview-label">{{ t('serverInfo.architecture') }}</div>
          <RsTooltip :content="serverInfo.architecture">
            <div class="overview-value text-truncate">{{ serverInfo.architecture }}</div>
          </RsTooltip>
        </div>
      </div>

      <div class="overview-item">
        <div class="overview-icon server-type">
          <GIcon size="24">
            <CloudServerOutlined />
          </GIcon>
        </div>
        <div class="overview-content">
          <div class="overview-label">{{ t('serverInfo.serverType') }}</div>
          <RsTooltip :content="getServerTypeLabel(serverInfo.serverType)">
            <div class="overview-value text-truncate">{{ getServerTypeLabel(serverInfo.serverType) }}</div>
          </RsTooltip>
        </div>
      </div>

      <div class="overview-item">
        <div class="overview-icon ip">
          <GIcon size="24">
            <GlobalOutlined />
          </GIcon>
        </div>
        <div class="overview-content">
          <div class="overview-label">{{ t('serverInfo.ipAddress') }}</div>
          <RsTooltip :content="serverInfo.ipAddress || t('serverInfo.na')">
            <div class="overview-value text-truncate">{{ serverInfo.ipAddress || t('serverInfo.na') }}</div>
          </RsTooltip>
        </div>
      </div>
    </div>
  </RsCard>
</template>

<script setup lang="ts">
import { useModuleI18n } from '@/hooks/useModuleI18n'
import {
  AndroidOutlined,
  AppleOutlined,
  CloudServerOutlined,
  DatabaseOutlined,
  DesktopOutlined,
  GlobalOutlined,
  WindowsOutlined,
} from '@vicons/antd'
import { GIcon } from '../../../../components/gicon'
import { RsCard, RsTooltip } from '../../../../ui'
import type { ServerInfo } from '../../types'

defineOptions({
  name: 'ServerInfoCard',
})

interface Props {
  serverInfo: ServerInfo
}

defineProps<Props>()

const { t } = useModuleI18n('hub0007')

/** 服务器类型标签转换 */
const getServerTypeLabel = (serverType?: string): string => {
  const typeMap: Record<string, string> = {
    physical: t('serverType.physicalShort'),
    virtual: t('serverType.virtualShort'),
    unknown: t('serverType.unknown'),
  }
  return typeMap[serverType || 'unknown'] || t('serverType.unknown')
}

/** 根据操作系统类型获取图标 */
const getOSIcon = (osType: string) => {
  const osLower = osType.toLowerCase()
  if (osLower.includes('windows')) {
    return WindowsOutlined
  }
  if (osLower.includes('linux')) {
    return AndroidOutlined
  }
  if (osLower.includes('mac') || osLower.includes('darwin')) {
    return AppleOutlined
  }
  return DesktopOutlined
}

/** 获取简化的系统版本信息 */
const getShortVersion = (version: string): string => {
  if (!version) return t('serverInfo.na')

  if (version.toLowerCase().includes('windows')) {
    const match = version.match(/Windows (\d+(?:\.\d+)?)/i)
    if (match) {
      const windowsVersion = match[1]
      const editionMatch = version.match(/Windows \d+(?:\.\d+)?\s+(\w+)/i)
      if (editionMatch) {
        return `Windows ${windowsVersion} ${editionMatch[1]}`
      }
      return `Windows ${windowsVersion}`
    }
  }

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
