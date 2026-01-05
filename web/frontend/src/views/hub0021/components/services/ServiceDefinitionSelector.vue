<template>
  <div class="service-definition-selector">
    <!-- 选择器显示区域 -->
    <div v-if="modelValue && (currentServices.length > 0 || currentService)" class="selected-service-card">
      <div class="service-info">
        <!-- 统一的服务显示：多服务和单服务使用相同的结构 -->
        <div v-if="currentServices.length > 0" class="services-list">
          <!-- 多服务：循环显示每个服务 -->
          <div v-for="service in currentServices" :key="service.serviceDefinitionId" class="service-item">
            <div class="service-header">
              <div class="service-avatar">
                <div class="avatar-circle">
                  {{ getServiceInitial(service.serviceName) }}
                </div>
              </div>
              <div class="service-main-info">
                <div class="service-name">{{ service.serviceName }}</div>
                <div class="service-id">{{ service.serviceDefinitionId }}</div>
              </div>
              <div class="service-status">
                <n-tag :type="service.activeFlag === 'Y' ? 'success' : 'error'" size="small">
                  {{ service.activeFlag === 'Y' ? '启用' : '禁用' }}
                </n-tag>
              </div>
            </div>
            <div class="service-details">
              <div class="detail-row">
                <div class="detail-item">
                  <span class="detail-label">服务类型</span>
                  <n-tag :type="service.serviceType === 1 ? 'success' : 'info'" size="small">
                    {{ service.serviceType === 1 ? '服务发现' : '静态配置' }}
                  </n-tag>
                </div>
                <div class="detail-item">
                  <span class="detail-label">负载均衡</span>
                  <span class="detail-value">{{ getLoadBalanceText(service.loadBalanceStrategy) }}</span>
                </div>
              </div>
              <div class="detail-row">
                <div class="detail-item full-width">
                  <span class="detail-label">健康检查</span>
                  <n-tag :type="service.healthCheckEnabled === 'Y' ? 'success' : 'default'" size="small">
                    {{ service.healthCheckEnabled === 'Y' ? '已启用' : '已禁用' }}
                  </n-tag>
                </div>
              </div>
            </div>
          </div>
        </div>
        <!-- 单服务：使用相同的结构 -->
        <div v-else-if="currentService" class="service-item">
          <div class="service-header">
            <div class="service-avatar">
              <div class="avatar-circle">
                {{ getServiceInitial(currentService.serviceName) }}
              </div>
            </div>
            <div class="service-main-info">
              <div class="service-name">{{ currentService.serviceName }}</div>
              <div class="service-id">{{ currentService.serviceDefinitionId }}</div>
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
                <n-tag :type="currentService.serviceType === 1 ? 'success' : 'info'" size="small">
                  {{ currentService.serviceType === 1 ? '服务发现' : '静态配置' }}
                </n-tag>
              </div>
              <div class="detail-item">
                <span class="detail-label">负载均衡</span>
                <span class="detail-value">{{ getLoadBalanceText(currentService.loadBalanceStrategy) }}</span>
              </div>
            </div>
            <div class="detail-row">
              <div class="detail-item full-width">
                <span class="detail-label">健康检查</span>
                <n-tag :type="currentService.healthCheckEnabled === 'Y' ? 'success' : 'default'" size="small">
                  {{ currentService.healthCheckEnabled === 'Y' ? '已启用' : '已禁用' }}
                </n-tag>
              </div>
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

    <!-- 未选择时的选择按钮 -->
    <div v-else class="empty-selector">
      <n-button @click="openSelector" dashed block size="large" class="select-btn">
        <template #icon>
          <n-icon><ServerOutline /></n-icon>
        </template>
        点击选择服务定义
      </n-button>
    </div>

    <!-- 服务定义选择对话框 -->
    <ServiceDefinitionListModal
      v-model:visible="showSelector"
      :title="'选择服务定义'"
      :width="1200"
      :gateway-instance-id="gatewayInstanceId"
      to="#hub0021-route-config-list"
      @select="handleServiceSelect"
      @close="handleClose"
    />
  </div>
</template>

<script setup lang="ts">
import { getApiMessage, isApiSuccess } from '@/utils/format'
import {
  CloseOutline,
  RefreshOutline,
  ServerOutline
} from '@vicons/ionicons5'
import { NButton, NIcon, NTag, useMessage } from 'naive-ui'
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { getServiceDefinitionById } from '../../api'
import ServiceDefinitionListModal from './ServiceDefinitionListModal.vue'
import type { ServiceDefinition } from './types'

interface Props {
  modelValue?: string
  gatewayInstanceId?: string
}

interface Emits {
  (e: 'update:modelValue', value: string | null): void
  (e: 'change', serviceDefinition: ServiceDefinition | ServiceDefinition[] | null): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()
const message = useMessage()

// 响应式状态
const showSelector = ref(false)
const loading = ref(false)
// 当前选中的服务完整信息（从 ServiceDefinitionListModal 返回的信息）
const selectedServiceInfo = ref<ServiceDefinition | null>(null)
// 多服务模式：选中的服务列表
const selectedServicesInfo = ref<ServiceDefinition[]>([])

// 计算属性：解析 modelValue（可能是逗号分割的字符串）
const serviceIds = computed(() => {
  if (!props.modelValue) return []
  // 如果是逗号分割的字符串，分割成数组
  if (typeof props.modelValue === 'string' && props.modelValue.includes(',')) {
    return props.modelValue.split(',').map(id => id.trim()).filter(id => id)
  }
  // 单个服务ID
  return [props.modelValue]
})

// 计算属性：当前显示的服务信息（单服务，向后兼容）
const currentService = computed(() => {
  // 如果是多服务模式，返回 null
  if (serviceIds.value.length > 1) {
    return null
  }
  // 单服务模式
  if (serviceIds.value.length === 1) {
    const serviceId = serviceIds.value[0]
    // 如果已经有保存的选中服务信息，且 ID 匹配，直接返回
    if (selectedServiceInfo.value && selectedServiceInfo.value.serviceDefinitionId === serviceId) {
      return selectedServiceInfo.value
    }
  }
  return null
})

// 计算属性：当前显示的服务列表（多服务）
const currentServices = computed(() => {
  if (serviceIds.value.length <= 1) {
    return []
  }
  // 返回已加载的服务信息
  return selectedServicesInfo.value.filter(service => 
    serviceIds.value.includes(service.serviceDefinitionId)
  )
})

/**
 * 根据服务定义ID加载服务信息（单服务）
 */
const loadServiceById = async (serviceDefinitionId: string) => {
  if (!serviceDefinitionId) {
    selectedServiceInfo.value = null
    return
  }

  try {
    loading.value = true
    const response = await getServiceDefinitionById(serviceDefinitionId)
    
    if (isApiSuccess(response)) {
      const service = JSON.parse(response.bizData) as ServiceDefinition
      selectedServiceInfo.value = service
    } else {
      console.warn('获取服务定义失败:', getApiMessage(response, '获取服务定义失败'))
      selectedServiceInfo.value = null
    }
  } catch (error) {
    console.error('加载服务定义失败:', error)
    selectedServiceInfo.value = null
  } finally {
    loading.value = false
  }
}

/**
 * 根据服务定义ID列表加载服务信息（多服务）
 */
const loadServicesByIds = async (serviceDefinitionIds: string[]) => {
  if (!serviceDefinitionIds || serviceDefinitionIds.length === 0) {
    selectedServicesInfo.value = []
    return
  }

  try {
    loading.value = true
    // 并行加载所有服务信息
    const promises = serviceDefinitionIds.map(id => getServiceDefinitionById(id))
    const responses = await Promise.all(promises)
    
    const services: ServiceDefinition[] = []
    responses.forEach((response, index) => {
      if (isApiSuccess(response)) {
        try {
          const service = JSON.parse(response.bizData) as ServiceDefinition
          services.push(service)
        } catch (error) {
          console.warn(`解析服务定义 ${serviceDefinitionIds[index]} 失败:`, error)
        }
      } else {
        console.warn(`获取服务定义 ${serviceDefinitionIds[index]} 失败:`, getApiMessage(response, '获取服务定义失败'))
      }
    })
    
    selectedServicesInfo.value = services
  } catch (error) {
    console.error('加载服务定义列表失败:', error)
    selectedServicesInfo.value = []
  } finally {
    loading.value = false
  }
}

// 工具函数
const getLoadBalanceText = (algorithm: string): string => {
  const map: Record<string, string> = {
    'round-robin': '轮询',
    'random': '随机',
    'ip-hash': 'IP哈希',
    'least-conn': '最少连接',
    'weighted-round-robin': '加权轮询',
    'consistent-hash': '一致性哈希',
    'ROUND_ROBIN': '轮询',
    'RANDOM': '随机',
    'IP_HASH': 'IP哈希',
    'LEAST_CONN': '最少连接',
    'WEIGHTED_ROUND_ROBIN': '加权轮询',
    'CONSISTENT_HASH': '一致性哈希',
  }
  return map[algorithm] || algorithm
}

const getServiceInitial = (serviceName: string): string => {
  if (!serviceName) return '?'
  const firstChar = serviceName.charAt(0).toUpperCase()
  return /^[A-Z]$/.test(firstChar) ? firstChar : serviceName.charAt(0)
}

// 方法
const openSelector = () => {
  showSelector.value = true
  // 数据加载由 ServiceDefinitionListModal 组件内部处理，不需要在这里调用
}

const handleClose = () => {
  showSelector.value = false
}

const handleClear = () => {
  selectedServiceInfo.value = null
  selectedServicesInfo.value = []
  emit('update:modelValue', null)
  emit('change', null)
  message.info('已清除服务定义选择')
}

const handleServiceSelect = (services: ServiceDefinition[]) => {
  // 始终接收数组（单个服务时也是数组）
  if (services.length === 0) {
    selectedServiceInfo.value = null
    selectedServicesInfo.value = []
    emit('update:modelValue', null)
    emit('change', null)
    handleClose()
    return
  }
  
  if (services.length === 1) {
    // 单服务模式
    const service = services[0]
    selectedServiceInfo.value = service
    selectedServicesInfo.value = []
    emit('update:modelValue', service.serviceDefinitionId)
    emit('change', service)
    message.success(`已选择服务：${service.serviceName}`)
  } else {
    // 多服务模式（逗号分割）
    selectedServiceInfo.value = null
    selectedServicesInfo.value = services
    const serviceIds = services.map(s => s.serviceDefinitionId).join(',')
    emit('update:modelValue', serviceIds)
    emit('change', services)
    message.success(`已选择 ${services.length} 个服务`)
  }
  
  handleClose()
}

// 监听器
const stopGatewayInstanceIdWatch = watch(() => props.gatewayInstanceId, (newId, oldId) => {
  // 如果 gatewayInstanceId 被清空，清空选中的服务信息
  if (!newId && oldId) {
    selectedServiceInfo.value = null
  }
})

// 监听 modelValue 变化
const stopModelValueWatch = watch(() => props.modelValue, (newValue, oldValue) => {
  if (!newValue) {
    // 如果 modelValue 被清空，清空所有服务信息
    selectedServiceInfo.value = null
    selectedServicesInfo.value = []
    return
  }
  
  if (newValue !== oldValue) {
    // 解析服务ID（可能是逗号分割的字符串）
    const ids = typeof newValue === 'string' && newValue.includes(',')
      ? newValue.split(',').map(id => id.trim()).filter(id => id)
      : [newValue]
    
    if (ids.length === 1) {
      // 单服务模式
      const serviceId = ids[0]
      if (!selectedServiceInfo.value || selectedServiceInfo.value.serviceDefinitionId !== serviceId) {
        loadServiceById(serviceId)
      }
      selectedServicesInfo.value = []
    } else if (ids.length > 1) {
      // 多服务模式
      // 检查是否需要加载（比较ID列表）
      const currentIds = selectedServicesInfo.value.map(s => s.serviceDefinitionId).sort()
      const newIds = [...ids].sort()
      if (JSON.stringify(currentIds) !== JSON.stringify(newIds)) {
        loadServicesByIds(ids)
      }
      selectedServiceInfo.value = null
    }
  }
}, { immediate: true })

// ============= 资源清理 =============

// 组件卸载时清理所有监听器
onBeforeUnmount(() => {
  stopGatewayInstanceIdWatch()
  stopModelValueWatch()
})

// 组件挂载时，如果有 modelValue，加载服务信息
onMounted(() => {
  if (props.modelValue) {
    const ids = typeof props.modelValue === 'string' && props.modelValue.includes(',')
      ? props.modelValue.split(',').map(id => id.trim()).filter(id => id)
      : [props.modelValue]
    
    if (ids.length === 1) {
      loadServiceById(ids[0])
    } else if (ids.length > 1) {
      loadServicesByIds(ids)
    }
  }
})
</script>

<style scoped lang="scss">
.service-definition-selector {
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
      }

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

      .service-main-info {
        flex: 1;
        min-width: 0;

        .service-name {
          font-size: 16px;
          font-weight: 600;
          color: var(--n-text-color-1);
          margin-bottom: 4px;
          line-height: 1.2;
        }

        .service-id {
          font-size: 12px;
          color: var(--n-text-color-3);
          font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
          background: var(--n-color-surface);
          padding: 2px 6px;
          border-radius: 4px;
          display: inline-block;
          line-height: 1.2;
        }
      }

      .service-status {
        flex-shrink: 0;
      }
    }

    .service-details {
      margin-bottom: 16px;

      .detail-row {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 12px;
        margin-bottom: 12px;

        &:last-child {
          margin-bottom: 0;
        }

        .detail-item {
          display: flex;
          align-items: center;
          gap: 8px;

          &.full-width {
            grid-column: 1 / -1;
          }

          .detail-label {
            font-size: 12px;
            color: var(--n-text-color-3);
            font-weight: 500;
            min-width: 60px;
            flex-shrink: 0;
          }

          .detail-value {
            font-size: 13px;
            color: var(--n-text-color-2);
            background: var(--n-color-surface);
            padding: 2px 8px;
            border-radius: 4px;
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

    // 服务项样式（单服务和多服务共用）
    .service-item {
      margin-bottom: 16px;

      &:last-child {
        margin-bottom: 0;
      }
    }

    .services-list {
      margin-bottom: 16px;

      .service-item {
        padding-bottom: 16px;
        border-bottom: 1px solid var(--n-border-color);

        &:last-child {
          padding-bottom: 0;
          border-bottom: none;
        }
      }
    }
  }
}

.empty-selector {
  width: 100%;

  .select-btn {
    min-height: 56px;
    border-color: var(--n-border-color);
    transition: all 0.2s ease;

    &:hover {
      border-color: var(--n-color-primary);
      background-color: var(--n-color-primary-light);
    }
  }
}

</style>

