<template>
  <n-drawer
    v-model:show="isVisible"
    :width="520"
    :title="t('dialogs.serviceEvent.title')"
    placement="right"
    closable
    resizable
  >
    <n-spin :show="loading || detailLoading">
      <div class="event-detail-wrapper">
        <!-- Event Type and Time -->
        <div class="event-header">
          <n-tag size="medium" :type="getEventTypeColor(displayEvent?.eventType)" round>
            <template #icon>
              <n-icon><component :is="getEventTypeIcon(displayEvent?.eventType)" /></n-icon>
            </template>
            {{ displayEvent?.eventType }}
          </n-tag>
          <n-time
            v-if="displayEvent?.eventTime"
            :time="new Date(displayEvent.eventTime)"
            format="yyyy-MM-dd HH:mm:ss"
            class="event-time"
          />
        </div>

        <!-- Basic Info -->
        <n-card class="detail-card" title="事件基本信息" size="small">
          <div class="info-row">
            <span class="label">{{ t('columns.eventMessage') }}</span>
            <span class="value">{{ displayEvent?.eventMessage || '-' }}</span>
          </div>
          <div class="info-row">
            <span class="label">{{ t('columns.eventSource') }}</span>
            <span class="value">{{ displayEvent?.eventSource || '-' }}</span>
          </div>
          <div class="info-row">
            <span class="label">{{ t('columns.serviceName') }}</span>
            <span class="value">{{ displayEvent?.serviceName || '-' }}</span>
          </div>
          <div class="info-row">
            <span class="label">{{ t('columns.groupName') }}</span>
            <span class="value">{{ displayEvent?.groupName || '-' }}</span>
          </div>
          <div class="info-row">
            <span class="label">{{ t('columns.hostAddress') }}</span>
            <span class="value">{{ displayEvent?.hostAddress ? `${displayEvent.hostAddress}:${displayEvent.portNumber || '-'}` : '-' }}</span>
          </div>
          <div class="info-row">
            <span class="label">服务运行节点IP</span>
            <span class="value">{{ displayEvent?.nodeIpAddress || '-' }}</span>
          </div>
        </n-card>

        <!-- Event Data -->
        <n-card v-if="displayEvent?.eventDataJson" class="detail-card" title="事件数据" size="small">
          <n-code :code="formatJson(displayEvent.eventDataJson)" language="json" show-line-numbers />
        </n-card>

        <!-- Metadata -->
        <n-collapse>
          <n-collapse-item name="audit" title="审计信息">
            <div class="audit-info">
              <div class="info-row">
                <span class="label">事件ID</span>
                <span class="value">{{ displayEvent?.serviceEventId || '-' }}</span>
              </div>
              <div class="info-row">
                <span class="label">创建时间</span>
                <span class="value">
                  <n-time v-if="displayEvent?.addTime" :time="new Date(displayEvent.addTime)" format="yyyy-MM-dd HH:mm:ss" />
                  <span v-else>-</span>
                </span>
              </div>
              <div class="info-row">
                <span class="label">创建人</span>
                <span class="value">{{ displayEvent?.addWho || '-' }}</span>
              </div>
              <div class="info-row">
                <span class="label">租户ID</span>
                <span class="value">{{ displayEvent?.tenantId || '-' }}</span>
              </div>
              <div class="info-row">
                <span class="label">服务实例ID</span>
                <span class="value">{{ displayEvent?.serviceInstanceId || '-' }}</span>
              </div>
            </div>
          </n-collapse-item>
        </n-collapse>
      </div>
    </n-spin>
  </n-drawer>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { 
  NDrawer, NCard, NTag, NIcon, NTime, NSpin, NCollapse, 
  NCollapseItem, NCode, useMessage
} from 'naive-ui'
import { 
  InformationCircleOutline, CloudOutline, 
  ServerOutline, FolderOutline
} from '@vicons/ionicons5'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { fetchServiceEventById } from '../api'
import { parseJsonData } from '@/utils/format'
import type { ServiceEvent } from '../types'

// Props（组件属性）
interface Props {
  show: boolean          // 是否显示详情抽屉
  event?: ServiceEvent | null  // 事件对象数据
  loading?: boolean      // 加载状态
}

/**
 * 组件事件定义
 * update:show - 更新显示状态（用于v-model绑定）
 * close - 关闭抽屉事件
 */
interface Emits {
  (e: 'update:show', value: boolean): void
  (e: 'close'): void
}

const props = withDefaults(defineProps<Props>(), {
  show: false,    // 默认不显示
  event: null,    // 默认无事件数据
  loading: false  // 默认非加载状态
})

const emit = defineEmits<Emits>()

// Hooks（钩子函数）
const { t } = useModuleI18n('hub0041')
const message = useMessage()

// 完整的事件数据
const fullEvent = ref<ServiceEvent | null>(null)
const detailLoading = ref(false)

/**
 * 控制抽屉显示状态的计算属性
 * 支持v-model:show双向绑定
 */
const isVisible = computed({
  // 获取值时返回props中的show
  get: () => props.show,
  // 设置值时同时发送update:show和close事件
  set: (value) => {
    emit('update:show', value)
    if (!value) {
      emit('close')
      // 关闭时清空完整数据
      fullEvent.value = null
    }
  }
})

/**
 * 获取完整的事件详情数据
 */
const fetchFullEventDetail = async (serviceEventId: string) => {
  try {
    detailLoading.value = true
    const response = await fetchServiceEventById(serviceEventId)
    
    if (response.oK) {
      fullEvent.value = parseJsonData<ServiceEvent>(response)
    } else {
      message.error('获取事件详情失败')
      fullEvent.value = null
    }
  } catch (error) {
    console.error('Failed to fetch event detail:', error)
    message.error('获取事件详情失败')
    fullEvent.value = null
  } finally {
    detailLoading.value = false
  }
}

// 监听事件变化，自动获取完整数据
watch(
  [() => props.event, () => props.show],
  ([newEvent, show]) => {
    if (newEvent && show && newEvent.serviceEventId) {
      fetchFullEventDetail(newEvent.serviceEventId)
    }
  },
  { immediate: true }
)

// 计算当前显示的事件数据（优先使用完整数据）
const displayEvent = computed(() => fullEvent.value || props.event)

// Event type icons and colors（事件类型图标和颜色）
const getEventTypeIcon = (eventType?: string) => {
  if (!eventType) return InformationCircleOutline

  // 分组相关事件
  if (eventType.includes('GROUP_')) return FolderOutline
  
  // 服务相关事件
  if (eventType.includes('SERVICE_')) return CloudOutline
  
  // 实例相关事件
  if (eventType.includes('INSTANCE_')) return ServerOutline
  
  // 默认图标
  return InformationCircleOutline
}

const getEventTypeColor = (eventType?: string) => {
  if (!eventType) return 'default'

  // 分组相关事件
  if (eventType === 'GROUP_CREATE') return 'success'
  if (eventType === 'GROUP_UPDATE') return 'info'
  if (eventType === 'GROUP_DELETE') return 'error'
  
  // 服务相关事件
  if (eventType === 'SERVICE_CREATE') return 'success'
  if (eventType === 'SERVICE_UPDATE') return 'info'
  if (eventType === 'SERVICE_DELETE') return 'error'
  
  // 实例相关事件
  if (eventType === 'INSTANCE_REGISTER') return 'success'
  if (eventType === 'INSTANCE_DEREGISTER') return 'error'
  if (eventType === 'INSTANCE_HEARTBEAT') return 'default'
  if (eventType === 'INSTANCE_HEALTH_CHANGE') return 'warning'
  if (eventType === 'INSTANCE_STATUS_CHANGE') return 'info'
  
  // 默认颜色
  return 'default'
}

/**
 * 格式化JSON字符串以便显示
 * 将JSON字符串转换为格式化的、带缩进的字符串
 * @param jsonString 原始JSON字符串
 * @returns 格式化后的JSON字符串
 */
const formatJson = (jsonString: string) => {
  try {
    // 尝试解析JSON字符串
    const parsed = JSON.parse(jsonString)
    // 格式化为带缩进的字符串
    return JSON.stringify(parsed, null, 2)
  } catch (e) {
    // 如果解析失败，返回原始字符串
    return jsonString
  }
}
</script>

<style scoped lang="scss">
.event-detail-wrapper {
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  height: 100%;
  overflow-y: auto;

  .event-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 8px;

    .event-time {
      font-size: 14px;
      color: var(--text-color-3);
    }
  }

  .detail-card {
    margin-bottom: 16px;
    
    :deep(.n-card__content) {
      padding: 12px;
    }
  }

  .info-row {
    display: flex;
    margin-bottom: 8px;
    font-size: 14px;
    
    .label {
      width: 120px;
      color: var(--text-color-3);
      flex-shrink: 0;
    }
    
    .value {
      flex: 1;
      word-break: break-word;
    }
  }
  
  .audit-info {
    padding: 12px;
  }
}
</style>
