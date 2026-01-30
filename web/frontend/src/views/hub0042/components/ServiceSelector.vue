<template>
  <div class="service-selector">
    <!-- 已选择时：显示服务卡片 -->
    <div v-if="currentService" class="selected-service-card">
      <div class="service-info">
        <div class="service-header">
          <div class="service-avatar">
            <div class="avatar-circle">
              {{ getServiceInitial(currentService.serviceName) }}
            </div>
          </div>
          <div class="service-main-info">
            <div class="service-name">{{ currentService.serviceName }}</div>
            <div class="service-tags">
              <n-tag type="info" size="small">{{ currentService.namespaceId }}</n-tag>
              <n-tag type="primary" size="small">{{ currentService.groupName }}</n-tag>
            </div>
          </div>
          <div class="service-status">
            <n-tag :type="currentService.activeFlag === 'Y' ? 'success' : 'error'" size="small">
              {{ currentService.activeFlag === 'Y' ? '启用' : '禁用' }}
            </n-tag>
          </div>
        </div>
        <div class="service-details">
          <div class="detail-row">
            <div class="detail-item">
              <span class="detail-label">服务类型</span>
              <span class="detail-value">{{ currentService.serviceType || 'INTERNAL' }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">节点数</span>
              <span class="detail-value">
                {{ currentService.nodeCount ?? 0 }}
                <template v-if="currentService.healthyNodeCount !== undefined || currentService.unhealthyNodeCount !== undefined">
                  (健康: {{ currentService.healthyNodeCount ?? 0 }}, 不健康: {{ currentService.unhealthyNodeCount ?? 0 }})
                </template>
              </span>
            </div>
            <div class="detail-item">
              <span class="detail-label">协议</span>
              <n-tag type="warning" size="small">{{ protocolType }}</n-tag>
            </div>
          </div>
        </div>
        <div class="service-actions">
          <n-button @click="openSelector" size="small" secondary>
            <template #icon>
              <n-icon><RefreshOutline /></n-icon>
            </template>
            重新选择
          </n-button>
          <n-button @click="handleClear" size="small" type="error" secondary>
            <template #icon>
              <n-icon><CloseOutline /></n-icon>
            </template>
            清除
          </n-button>
        </div>
      </div>
    </div>

    <!-- 未选择时：显示选择按钮 -->
    <div v-else class="empty-selector">
      <n-button @click="openSelector" dashed block size="large" class="select-btn">
        <template #icon>
          <n-icon><ServerOutline /></n-icon>
        </template>
        点击选择注册服务
      </n-button>
    </div>

    <!-- 服务选择弹窗 -->
    <ServiceSelectorModal
      v-model:visible="showSelector"
      :to="to"
      @select="handleServiceSelect"
      @close="handleClose"
    />
  </div>
</template>

<script setup lang="ts">
import { getApiMessage, isApiSuccess, parseJsonData } from '@/utils/format'
import {
  CloseOutline,
  RefreshOutline,
  ServerOutline
} from '@vicons/ionicons5'
import { NButton, NIcon, NTag, useMessage } from 'naive-ui'
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { getService } from '../api'
import type { Service } from '../types'
import ServiceSelectorModal from './ServiceSelectorModal.vue'

/**
 * 服务选择元数据 - 与 registry_utils.go 查找信息保持一致
 * 用于服务发现类型的服务定义
 */
export interface ServiceSelectionMetadata {
  /** 租户ID */
  tenantId: string
  /** 命名空间ID */
  namespaceId: string
  /** 分组名称 */
  groupName: string
  /** 服务名称 */
  serviceName: string
  /** 服务发现类型（固定为 INTERNAL，表示从本服务中心注册和发现的内部服务） */
  discoveryType: 'INTERNAL'
  /** 协议类型（http/https） */
  protocolType: string
}

// Props
interface Props {
  /** 当前选中的服务元数据（JSON字符串或对象） */
  modelValue?: string | ServiceSelectionMetadata | null
  /** 弹窗挂载目标 */
  to?: string
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: null,
  to: 'body'
})

// Emits
const emit = defineEmits<{
  'update:modelValue': [value: ServiceSelectionMetadata | null]
  'change': [metadata: ServiceSelectionMetadata | null]
}>()

// 响应式状态
const message = useMessage()
const showSelector = ref(false)
const loading = ref(false)
const currentServiceInfo = ref<Service | null>(null)

// 解析 modelValue
const parsedMetadata = computed<ServiceSelectionMetadata | null>(() => {
  if (!props.modelValue) return null
  if (typeof props.modelValue === 'string') {
    try {
      return JSON.parse(props.modelValue) as ServiceSelectionMetadata
    } catch {
      return null
    }
  }
  return props.modelValue
})

// 当前显示的服务信息
const currentService = computed(() => {
  return currentServiceInfo.value
})

// 协议类型
const protocolType = computed(() => {
  return parsedMetadata.value?.protocolType || 'http'
})

// 获取服务名称首字母
const getServiceInitial = (serviceName: string): string => {
  if (!serviceName) return '?'
  const firstChar = serviceName.charAt(0).toUpperCase()
  return /^[A-Z]$/.test(firstChar) ? firstChar : serviceName.charAt(0)
}

/**
 * 根据元数据加载服务详情
 */
const loadServiceByMetadata = async (metadata: ServiceSelectionMetadata) => {
  if (!metadata || !metadata.namespaceId || !metadata.groupName || !metadata.serviceName) {
    currentServiceInfo.value = null
    return
  }

  try {
    loading.value = true
    const response = await getService(metadata.namespaceId, metadata.groupName, metadata.serviceName)
    
    if (isApiSuccess(response)) {
      const service = parseJsonData<Service | undefined>(response, undefined)
      if (service) {
        currentServiceInfo.value = service
      } else {
        console.warn('服务数据解析失败')
        currentServiceInfo.value = null
      }
    } else {
      console.warn('获取服务失败:', getApiMessage(response, '获取服务失败'))
      // 如果获取失败，使用元数据构造基本信息
      currentServiceInfo.value = {
        tenantId: metadata.tenantId,
        namespaceId: metadata.namespaceId,
        groupName: metadata.groupName,
        serviceName: metadata.serviceName,
        activeFlag: 'Y'
      } as Service
    }
  } catch (error) {
    console.error('加载服务失败:', error)
    currentServiceInfo.value = null
  } finally {
    loading.value = false
  }
}

// 监听 modelValue 变化
const stopModelValueWatch = watch(() => props.modelValue, (newValue) => {
  if (!newValue) {
    currentServiceInfo.value = null
    return
  }
  
  const metadata = parsedMetadata.value
  if (metadata) {
    // 检查是否需要重新加载
    const current = currentServiceInfo.value
    if (!current || 
        current.namespaceId !== metadata.namespaceId ||
        current.groupName !== metadata.groupName ||
        current.serviceName !== metadata.serviceName) {
      loadServiceByMetadata(metadata)
    }
  }
}, { immediate: true })

// 组件卸载时清理监听器
onBeforeUnmount(() => {
  stopModelValueWatch()
})

// 打开选择器
const openSelector = () => {
  showSelector.value = true
}

// 关闭选择器
const handleClose = () => {
  showSelector.value = false
}

// 清除选择
const handleClear = () => {
  currentServiceInfo.value = null
  emit('update:modelValue', null)
  emit('change', null)
  message.info('已清除服务选择')
}

// 处理服务选择
const handleServiceSelect = (service: Service) => {
  // 保存服务信息
  currentServiceInfo.value = service
  
  // 构建符合 registry_utils.go 格式的元数据
  // discoveryType 使用 'INTERNAL' 表示从本服务中心注册和发现的内部服务
  const metadata: ServiceSelectionMetadata = {
    tenantId: service.tenantId || 'default',
    namespaceId: service.namespaceId,
    groupName: service.groupName || 'DEFAULT_GROUP',
    serviceName: service.serviceName,
    discoveryType: 'INTERNAL',
    protocolType: 'http'
  }
  
  emit('update:modelValue', metadata)
  emit('change', metadata)
  message.success(`已选择服务：${service.serviceName}`)
  handleClose()
}
</script>

<style scoped lang="scss">
.service-selector {
  width: 100%;
}

.selected-service-card {
  background: var(--n-card-color);
  border: 1px solid var(--n-border-color);
  border-radius: 8px;
  padding: 16px;
  transition: all 0.2s ease;

  &:hover {
    border-color: var(--n-color-primary);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  }

  .service-info {
    .service-header {
      display: flex;
      align-items: center;
      gap: 12px;
      margin-bottom: 16px;

      .service-avatar {
        flex-shrink: 0;

        .avatar-circle {
          width: 48px;
          height: 48px;
          border-radius: 50%;
          background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
          display: flex;
          align-items: center;
          justify-content: center;
          color: white;
          font-weight: 600;
          font-size: 18px;
        }
      }

      .service-main-info {
        flex: 1;
        min-width: 0;

        .service-name {
          font-size: 16px;
          font-weight: 600;
          color: var(--n-text-color-1);
          margin-bottom: 6px;
          line-height: 1.2;
        }

        .service-tags {
          display: flex;
          gap: 6px;
        }
      }

      .service-status {
        flex-shrink: 0;
      }
    }

    .service-details {
      margin-bottom: 16px;
      padding: 12px;
      background: var(--n-color-hover);
      border-radius: 6px;

      .detail-row {
        display: flex;
        gap: 24px;
        flex-wrap: wrap;

        .detail-item {
          display: flex;
          align-items: center;
          gap: 8px;

          .detail-label {
            font-size: 12px;
            color: var(--n-text-color-3);
            font-weight: 500;
          }

          .detail-value {
            font-size: 13px;
            color: var(--n-text-color-2);
            font-weight: 500;
          }
        }
      }
    }

    .service-actions {
      display: flex;
      gap: 8px;
      justify-content: flex-end;
      padding-top: 12px;
      border-top: 1px solid var(--n-border-color);
    }
  }
}

.empty-selector {
  width: 100%;

  .select-btn {
    min-height: 80px;
    border-color: var(--n-border-color);
    transition: all 0.2s ease;

    &:hover {
      border-color: var(--n-color-primary);
      background-color: var(--n-color-primary-suppl);
    }
  }
}
</style>
