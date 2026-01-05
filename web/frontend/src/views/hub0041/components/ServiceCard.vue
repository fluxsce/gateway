<template>
  <n-card
    class="service-card"
    hoverable
    :bordered="true"
    @click="$emit('viewDetail', service)"
  >
    <div class="card-content">
      <!-- Service Header with Actions -->
      <div class="card-header">
        <div class="service-name-row">
          <div class="name-with-badge">
            <n-text class="service-name" :depth="1">{{ service.serviceName }}</n-text>
            <n-tag size="small" :type="service.activeFlag === 'Y' ? 'success' : 'error'" round>
              {{ getStatusText() }}
            </n-tag>
          </div>
          
          <div class="actions-menu" @click.stop>
            <n-dropdown :options="dropdownOptions" trigger="click" placement="bottom-end" @select="handleDropdownSelect">
              <n-button quaternary circle size="small">
                <template #icon>
                  <n-icon><EllipsisHorizontalOutline /></n-icon>
                </template>
              </n-button>
            </n-dropdown>
          </div>
        </div>
        
        <div class="service-meta">
          <n-tag size="tiny" :type="getProtocolTagType(service.protocolType)">{{ service.protocolType }}</n-tag>
          <n-tooltip trigger="hover" placement="bottom">
            <template #trigger>
              <div class="group-info">
                <n-icon size="14" color="var(--text-color-3)"><FolderOutline /></n-icon>
                <span class="group-name">{{ service.groupName }}</span>
                <span class="group-id" v-if="service.serviceGroupId">({{ formatGroupId(service.serviceGroupId) }})</span>
              </div>
            </template>
            {{ t('serviceGroup') }}: {{ service.groupName }}
          </n-tooltip>
          <div class="tenant-info" v-if="service.tenantId && service.tenantId !== 'default'">
            <n-tooltip trigger="hover" placement="bottom">
              <template #trigger>
                <div class="tenant-tag">
                  <n-icon size="14" color="var(--text-color-3)"><PeopleOutline /></n-icon>
                  <span class="tenant-name">{{ service.tenantId }}</span>
                </div>
              </template>
              {{ t('namespace') }}: {{ service.tenantId }}
            </n-tooltip>
          </div>
        </div>
      </div>

      <!-- Service Description -->
      <div class="service-description" v-if="service.serviceDescription">
        <n-ellipsis :line-clamp="2" :tooltip="{ width: 250 }">
          {{ service.serviceDescription }}
        </n-ellipsis>
      </div>

      <!-- Instance Count -->
      <div class="instance-count">
        <div class="instance-count-item">
          <n-icon size="18" color="var(--info-color)"><ServerOutline /></n-icon>
          <span class="count-text">{{ getInstanceCount() }} {{ t('instanceCount') }}</span>
        </div>
      </div>

      <!-- Footer -->
      <div class="card-footer">
        <n-tooltip trigger="hover">
          <template #trigger>
            <div class="footer-item">
              <n-icon size="14"><CloudOutline /></n-icon>
              <span class="footer-text">{{ getLoadBalanceText(service.loadBalanceStrategy) }}</span>
            </div>
          </template>
          {{ t('loadBalanceStrategy') }}
        </n-tooltip>
        
        <n-tooltip trigger="hover">
          <template #trigger>
            <div class="footer-item">
              <n-icon size="14"><TimeOutline /></n-icon>
              <n-time class="footer-text" :time="new Date(service.editTime)" type="relative" />
            </div>
          </template>
          {{ t('lastUpdated') }}
        </n-tooltip>
      </div>
    </div>
  </n-card>
</template>

<script setup lang="ts">
import { computed, h } from 'vue'
import { 
  NCard, NTag, NButton, NIcon, NEllipsis, NTime, NDropdown, 
  NText, NTooltip
} from 'naive-ui'
import { 
  CloudOutline, ServerOutline, FolderOutline,
  EllipsisHorizontalOutline, TimeOutline, 
  CreateOutline, PeopleOutline, TrashOutline, ListOutline
} from '@vicons/ionicons5'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import type { Service, ProtocolType, LoadBalanceStrategy } from '../types'

interface Props {
  service: Service
}

interface Emits {
  (e: 'viewDetail', service: Service): void
  (e: 'edit', service: Service): void
  (e: 'delete', service: Service): void
  (e: 'viewEvents', service: Service): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

// i18n
const { t } = useModuleI18n('hub0041')

// Dropdown options
const dropdownOptions = computed(() => [
  {
    label: t('actions.viewEvents'),
    key: 'viewEvents',
    icon: () => h(NIcon, null, { default: () => h(ListOutline) })
  },
  {
    label: t('actions.edit'),
    key: 'edit',
    icon: () => h(NIcon, null, { default: () => h(CreateOutline) })
  },
  {
    label: t('actions.delete'),
    key: 'delete',
    icon: () => h(NIcon, null, { default: () => h(TrashOutline) })
  }
])

// Status text
const getStatusText = () => {
  return props.service.activeFlag === 'Y' ? t('status.Y') : t('status.N')
}

// Protocol tag type
const getProtocolTagType = (protocol: ProtocolType): "error" | "default" | "success" | "warning" | "info" | "primary" => {
  switch (protocol) {
    case 'HTTP': return 'success'
    case 'HTTPS': return 'warning'
    case 'TCP': return 'info'
    case 'UDP': return 'default'
    case 'GRPC': return 'error'
    default: return 'default'
  }
}

// Format group ID to make it shorter if needed
const formatGroupId = (groupId: string): string => {
  // If group ID is too long, truncate it
  if (groupId.length > 8) {
    return groupId.substring(0, 8) + '...'
  }
  return groupId
}

// Load balance text
const getLoadBalanceText = (strategy: LoadBalanceStrategy) => {
  const strategyMap = {
    'ROUND_ROBIN': t('roundRobin'),
    'WEIGHTED_ROUND_ROBIN': t('weightedRoundRobin'),
    'LEAST_CONNECTIONS': t('leastConnections'),
    'RANDOM': t('random'),
    'IP_HASH': t('ipHash')
  }
  return strategyMap[strategy] || strategy
}

// 获取实例数量
const getInstanceCount = () => {
  // 优先使用实例列表长度
  if (props.service.instances && Array.isArray(props.service.instances)) {
    return props.service.instances.length
  }
  // 其次使用实例计数字段
  return props.service.instanceCount || 0
}

// Dropdown select handler
const handleDropdownSelect = (key: string) => {
  if (key === 'edit') {
    emit('edit', props.service)
  } else if (key === 'delete') {
    emit('delete', props.service)
  } else if (key === 'viewEvents') {
    emit('viewEvents', props.service)
  }
}
</script>

<style scoped lang="scss">
.service-card {
  cursor: pointer;
  transition: all 0.2s ease;
  height: 100%;
  border-radius: 12px;
  background: var(--card-color);
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.12);
  border: 1px solid var(--border-color);

  &:hover {
    box-shadow: 0 4px 18px rgba(0, 0, 0, 0.15);
    transform: translateY(-2px);
    border-color: var(--primary-color-hover);
  }

  :deep(.n-card__content) {
    padding: 0;
    height: 100%;
  }

  .card-content {
    padding: 16px;
    height: 100%;
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  // Card Header
  .card-header {
    display: flex;
    flex-direction: column;
    gap: 8px;
    
    .service-name-row {
      display: flex;
      justify-content: space-between;
      align-items: center;
      
      .name-with-badge {
        display: flex;
        flex-direction: column;
        align-items: flex-start;
        gap: 4px;
        max-width: 90%;
        
        .service-name {
          font-size: 16px;
          font-weight: 600;
          line-height: 1.3;
          /* 移除以下属性以允许服务名称完整显示 */
          /* white-space: nowrap; */
          /* overflow: hidden; */
          /* text-overflow: ellipsis; */
          /* 添加以下属性以确保长名称能够换行显示 */
          white-space: normal;
          word-break: break-word;
          color: var(--text-color-1);
        }
      }
      
      .actions-menu {
        margin-left: auto;
      }
    }
    
    .service-meta {
      display: flex;
      align-items: center;
      gap: 8px;
      flex-wrap: wrap;
      
      .group-info {
        display: flex;
        align-items: center;
        gap: 4px;
        background-color: var(--tag-color);
        padding: 1px 6px;
        border-radius: 4px;
        cursor: pointer;
        
        .group-name {
          font-size: 11px;
          color: var(--text-color-2);
          font-weight: 500;
        }
        
        .group-id {
          font-size: 10px;
          color: var(--text-color-3);
          font-style: italic;
        }
      }
      
      .tenant-info {
        .tenant-tag {
          display: flex;
          align-items: center;
          gap: 4px;
          background-color: rgba(var(--primary-color-rgb), 0.1);
          padding: 1px 6px;
          border-radius: 4px;
          cursor: pointer;
          
          .tenant-name {
            font-size: 11px;
            color: var(--text-color-2);
          }
        }
      }
    }
  }

  // Service Description
  .service-description {
    font-size: 12px;
    line-height: 1.5;
    color: var(--text-color-2);
    margin: 2px 0;
    padding: 6px;
    background-color: var(--bg-color-subtle);
    border-radius: 6px;
  }

  // Instance Count
  .instance-count {
    padding: 4px 0;
    margin: 2px 0;
    
    .instance-count-item {
      display: flex;
      align-items: center;
      gap: 8px;
      padding: 6px 10px;
      background-color: var(--card-color);
      border: 1px solid var(--border-color);
      border-radius: 6px;
      
      .count-text {
        font-size: 13px;
        color: var(--text-color-2);
        font-weight: 500;
      }
    }
  }

  // Footer
  .card-footer {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-top: auto;
    padding-top: 8px;
    border-top: 1px solid var(--divider-color);
    font-size: 12px;
    color: var(--text-color-3);
    
    .footer-item {
      display: flex;
      align-items: center;
      gap: 4px;
      
      .footer-text {
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
      }
    }
  }
}

// Responsive design
@media (max-width: 1200px) {
  .service-card {
    .card-content {
      padding: 12px;
      gap: 8px;
    }
    
    .card-header {
      .service-name-row .name-with-badge .service-name {
        font-size: 14px;
      }
    }
    
    .instance-count .instance-count-item .count-text {
      font-size: 12px;
    }
  }
}
</style>
