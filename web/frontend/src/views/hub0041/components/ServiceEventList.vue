<template>
  <div class="service-event-list">
    <n-data-table
      :columns="columns"
      :data="events"
      :loading="loading"
      :pagination="naiveConfig"
      :bordered="false"
      :scroll-x="1370"
      :single-line="false"
      :size="'medium'"
      remote
      striped
    />
  </div>
</template>

<script setup lang="ts">
import { onMounted, computed } from 'vue'
import { NDataTable } from 'naive-ui'
import type { ServiceEvent } from '../types'
import { useServiceEvents } from '../hooks'

// Props（组件属性）
interface Props {
  serviceId?: string  // 服务ID，用于筛选特定服务的事件
  instanceId?: string // 实例ID，用于筛选特定实例的事件
}

const props = withDefaults(defineProps<Props>(), {
  serviceId: undefined,
  instanceId: undefined
})

// 使用服务事件Hook
const {
  events,
  loading,
  naiveConfig,
  loadEvents,
  createTableColumns,
  setPaginationCallback
} = useServiceEvents()

// 创建表格列定义
const columns = computed(() => createTableColumns(handleViewDetail))

// Define emits（定义事件发射）
const emit = defineEmits<{
  (e: 'viewEvent', event: ServiceEvent): void // 查看事件详情事件
}>()

/**
 * 处理查看详情操作
 * 点击查看按钮时，将事件对象发送给父组件处理
 * @param event 事件对象
 */
const handleViewDetail = (event: ServiceEvent) => {
  // 发送事件到父组件处理详情显示
  emit('viewEvent', event)
}

/**
 * 加载事件列表数据的包装函数
 */
const loadEventsList = (customParams?: Record<string, any>) => {
  loadEvents(props.serviceId, props.instanceId, customParams)
}

// 设置分页回调
setPaginationCallback(() => {
  loadEventsList()
})

// Lifecycle（生命周期）
onMounted(() => {
  loadEventsList() // 组件挂载时自动加载事件列表
})

// Expose methods（暴露方法给父组件）
defineExpose({
  loadEvents: loadEventsList // 加载事件数据方法，供父组件调用
})
</script>

<style scoped lang="scss">
.service-event-list {
  width: 100%;
  
  :deep(.n-data-table) {
    .n-data-table-th {
      background-color: var(--bg-color-subtle);
      font-weight: 500;
      white-space: nowrap;
    }
    
    .n-data-table-td {
      padding: 10px 12px;
      vertical-align: middle;
    }
    
    // 优化事件类型标签显示
    .n-tag {
      max-width: 160px;
      .n-tag__content {
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }
    }
    
    // 优化表格行高度
    .n-data-table-tbody .n-data-table-tr {
      min-height: 48px;
    }
    
    // 优化滚动条样式
    .n-data-table-base-table-body {
      &::-webkit-scrollbar {
        height: 8px;
      }
      
      &::-webkit-scrollbar-track {
        background: var(--scrollbar-color);
        border-radius: 4px;
      }
      
      &::-webkit-scrollbar-thumb {
        background: var(--scrollbar-color-hover);
        border-radius: 4px;
        
        &:hover {
          background: var(--primary-color);
        }
      }
    }
  }
}
</style>
