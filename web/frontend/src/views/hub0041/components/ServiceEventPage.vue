<template>
  <div class="service-event-page">
    <n-card>
      <template #header>
        <div class="page-header">
          <h3>{{ t('menu.eventList') }}</h3>
          <n-space>
            <n-button type="primary" @click="searchEvents">
              <template #icon>
                <n-icon><SearchOutline /></n-icon>
              </template>
              {{ t('actions.search') }}
            </n-button>
            <n-button @click="resetSearch">
              <template #icon>
                <n-icon><ReloadOutline /></n-icon>
              </template>
              {{ t('actions.reset') }}
            </n-button>
            <n-button @click="refreshEvents">
              <template #icon>
                <n-icon><RefreshOutline /></n-icon>
              </template>
              {{ t('actions.refresh') }}
            </n-button>
          </n-space>
        </div>
      </template>

      <!-- Search Form -->
      <div class="search-section">
        <n-form
          :model="searchForm"
          label-placement="left"
          label-width="auto"
          inline
        >
          <n-form-item :label="t('columns.serviceName')">
            <n-input
              v-model:value="searchForm.serviceName"
              :placeholder="t('search.placeholder.serviceName')"
              clearable
            />
          </n-form-item>
          
          <n-form-item :label="t('columns.eventType')">
            <n-select
              v-model:value="searchForm.eventType"
              :options="eventTypeOptions"
              clearable
              :placeholder="t('selectEventType')"
              style="min-width: 160px"
            />
          </n-form-item>
          
          <n-form-item :label="t('columns.eventSource')">
            <n-input
              v-model:value="searchForm.eventSource"
              :placeholder="t('search.placeholder.eventSource')"
              clearable
            />
          </n-form-item>
          
          <n-form-item :label="t('columns.eventTime')">
            <n-date-picker
              v-model:value="searchForm.timeRange"
              type="daterange"
              clearable
              :placeholder="t('search.placeholder.timeRange')"
              style="width: 240px"
            />
          </n-form-item>
        </n-form>
      </div>

      <!-- Events List -->
      <service-event-list 
        ref="eventListRef"
        @view-event="handleViewEvent"
      />

      <!-- Event Detail Drawer -->
      <service-event-detail 
        v-model:show="showEventDetail"
        :event="selectedEvent"
        :loading="detailLoading"
      />
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { 
  NCard, NButton, NIcon, NSpace, NForm, NFormItem, 
  NInput, NSelect, NDatePicker, useMessage 
} from 'naive-ui'
import { 
  RefreshOutline, ReloadOutline, SearchOutline 
} from '@vicons/ionicons5'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import ServiceEventList from './ServiceEventList.vue'
import ServiceEventDetail from './ServiceEventDetail.vue'
import { useServiceEvents } from '../hooks'
import { EventType, type ServiceEvent } from '../types'

// Hooks（钩子函数）
const message = useMessage()
const { t } = useModuleI18n('hub0041')
const { getServiceEventById } = useServiceEvents()

// Refs（响应式引用）
const eventListRef = ref(null)
const showEventDetail = ref(false) // 是否显示事件详情抽屉
const selectedEvent = ref<ServiceEvent | null>(null) // 当前选中的事件
const detailLoading = ref(false) // 详情加载状态

// Event type options grouped by category（按类别分组的事件类型选项）
const eventTypeOptions = [
  {
    type: 'group',
    label: '分组相关事件',
    key: 'group',
    children: [
      { label: '分组创建', value: EventType.SERVICE_GROUP_CREATED },
      { label: '分组更新', value: EventType.SERVICE_GROUP_UPDATED },
      { label: '分组删除', value: EventType.SERVICE_GROUP_DELETED }
    ]
  },
  {
    type: 'group',
    label: '服务相关事件',
    key: 'service',
    children: [
      { label: '服务注册', value: EventType.SERVICE_REGISTERED },
      { label: '服务更新', value: EventType.SERVICE_UPDATED },
      { label: '服务注销', value: EventType.SERVICE_DEREGISTERED }
    ]
  },
  {
    type: 'group',
    label: '实例相关事件',
    key: 'instance',
    children: [
      { label: '实例注册', value: EventType.INSTANCE_REGISTERED },
      { label: '实例注销', value: EventType.INSTANCE_DEREGISTERED },
      { label: '实例心跳更新', value: EventType.INSTANCE_HEARTBEAT_UPDATED },
      { label: '实例健康变更', value: EventType.INSTANCE_HEALTH_CHANGE },
      { label: '实例状态变更', value: EventType.INSTANCE_STATUS_CHANGE }
    ]
  }
]

// Search form（搜索表单）
const searchForm = reactive({
  serviceName: '', // 服务名称
  eventType: null as string | null, // 事件类型
  eventSource: '', // 事件来源
  timeRange: null as [number, number] | null // 时间范围
})

// Methods（方法）
/**
 * 刷新事件列表
 * 调用子组件的loadEvents方法重新加载事件数据
 */
const refreshEvents = () => {
  if (eventListRef.value && typeof (eventListRef.value as any).loadEvents === 'function') {
    (eventListRef.value as any).loadEvents()
    message.success(t('messages.refreshSuccess'))
  }
}

/**
 * 搜索事件
 * 根据表单条件构建搜索参数并调用子组件的loadEvents方法
 */
const searchEvents = () => {
  // 准备搜索参数
  const params: Record<string, any> = {}
  
  if (searchForm.serviceName) {
    params.serviceName = searchForm.serviceName
  }
  
  if (searchForm.eventType) {
    params.eventType = searchForm.eventType
  }
  
  if (searchForm.eventSource) {
    params.eventSource = searchForm.eventSource
  }
  
  if (searchForm.timeRange) {
    const [startTime, endTime] = searchForm.timeRange
    params.startTime = new Date(startTime).toISOString()
    params.endTime = new Date(endTime).toISOString()
  }
  
  // 应用搜索
  if (eventListRef.value && typeof (eventListRef.value as any).loadEvents === 'function') {
    (eventListRef.value as any).loadEvents(params)
  }
}

/**
 * 重置搜索条件
 * 清空所有搜索表单项并重新加载默认数据
 */
const resetSearch = () => {
  searchForm.serviceName = ''
  searchForm.eventType = null
  searchForm.eventSource = ''
  searchForm.timeRange = null
  
  // 重置为默认搜索
  if (eventListRef.value && typeof (eventListRef.value as any).loadEvents === 'function') {
    (eventListRef.value as any).loadEvents()
  }
}

/**
 * 处理列表组件发出的查看事件详情事件
 * @param event 事件对象
 */
const handleViewEvent = (event: ServiceEvent) => {
  selectedEvent.value = event
  showEventDetail.value = true
}

/**
 * 通过ID查看事件详情（供外部调用）
 * @param eventId 事件ID
 */
const viewEventDetail = async (eventId: string) => {
  detailLoading.value = true
  showEventDetail.value = true
  
  try {
    selectedEvent.value = await getServiceEventById(eventId)
  } catch (err) {
    message.error(t('messages.loadError'))
    console.error('加载事件详情失败:', err)
  } finally {
    detailLoading.value = false
  }
}

// Expose methods for parent component（向父组件暴露的方法）
defineExpose({
  refreshEvents, // 刷新事件列表
  viewEventDetail // 查看事件详情
})
</script>

<style scoped lang="scss">
.service-event-page {
  padding: 16px;
  
  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    
    h3 {
      margin: 0;
      font-size: 18px;
      font-weight: 500;
    }
  }
  
  .search-section {
    margin-bottom: 20px;
    padding: 16px;
    background-color: var(--bg-color-subtle);
    border-radius: 8px;

    :deep(.n-form) {
      display: flex;
      flex-wrap: wrap;
      gap: 16px;

      .n-form-item {
        margin-bottom: 0;
      }
    }
  }
}
</style>
