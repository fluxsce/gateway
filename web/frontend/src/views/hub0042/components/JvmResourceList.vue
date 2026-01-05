<template>
  <div class="jvm-resource-list">
    <!-- 搜索和筛选区域 -->
    <n-card :bordered="false" class="search-card">
      <n-space vertical :size="16">
        <n-space :size="12" wrap>
          <n-input
            v-model:value="searchParams.applicationName"
            :placeholder="t('searchByApplicationName')"
            clearable
            style="width: 200px"
            @keyup.enter="handleSearch"
          >
            <template #prefix>
              <n-icon :component="SearchOutline" />
            </template>
          </n-input>
          
          <n-input
            v-model:value="searchParams.groupName"
            :placeholder="t('searchByGroupName')"
            clearable
            style="width: 180px"
            @keyup.enter="handleSearch"
          >
            <template #prefix>
              <n-icon :component="SearchOutline" />
            </template>
          </n-input>
          
          <n-input
            v-model:value="searchParams.hostIpAddress"
            :placeholder="t('searchByHostIp')"
            clearable
            style="width: 180px"
            @keyup.enter="handleSearch"
          >
            <template #prefix>
              <n-icon :component="ServerOutline" />
            </template>
          </n-input>
          
          <n-select
            v-model:value="searchParams.healthyFlag"
            :options="healthStatusOptions"
            :placeholder="t('selectHealthStatus')"
            clearable
            style="width: 150px"
          />
          
          <n-button type="primary" @click="handleSearch">
            <template #icon>
              <n-icon :component="SearchOutline" />
            </template>
            {{ t('search') }}
          </n-button>
          
          <n-button @click="handleReset">
            <template #icon>
              <n-icon :component="RefreshOutline" />
            </template>
            {{ t('reset') }}
          </n-button>
        </n-space>
      </n-space>
    </n-card>

    <!-- 统计卡片 -->
    <n-space :size="16" style="margin-top: 16px">
      <n-card :bordered="false" style="flex: 1">
        <n-statistic :label="t('totalJvmInstances')" :value="total" />
      </n-card>
      <n-card :bordered="false" style="flex: 1">
        <n-statistic :label="t('healthyInstances')" :value="healthyCount">
          <template #suffix>
            <n-tag type="success" size="small">{{ t('healthy') }}</n-tag>
          </template>
        </n-statistic>
      </n-card>
      <n-card :bordered="false" style="flex: 1">
        <n-statistic :label="t('unhealthyInstances')" :value="unhealthyCount">
          <template #suffix>
            <n-tag type="error" size="small">{{ t('unhealthy') }}</n-tag>
          </template>
        </n-statistic>
      </n-card>
      <n-card :bordered="false" style="flex: 1">
        <n-statistic :label="t('attentionRequired')" :value="attentionCount">
          <template #suffix>
            <n-tag type="warning" size="small">{{ t('attention') }}</n-tag>
          </template>
        </n-statistic>
      </n-card>
    </n-space>

    <!-- 表格 -->
    <n-card :bordered="false" style="margin-top: 16px">
      <n-data-table
        :columns="tableColumns"
        :data="jvmResources"
        :loading="loading"
        :pagination="paginationReactive"
        :scroll-x="1940"
        @update:page="handlePageChange"
        @update:page-size="handlePageSizeChange"
      />
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, h } from 'vue'
import { NButton } from 'naive-ui'
import { SearchOutline, RefreshOutline, ServerOutline } from '@vicons/ionicons5'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { useJvmResourceManagement } from '../hooks'
import { createJvmResourceTableColumns } from '../models'
import type { JvmResource } from '../types'

const { t } = useModuleI18n('hub0042')

const {
  loading,
  jvmResources,
  total,
  queryParams,
  healthyCount,
  unhealthyCount,
  attentionCount,
  queryJvmResources,
  resetQuery
} = useJvmResourceManagement()

// 搜索参数
const searchParams = ref({
  applicationName: '',
  groupName: '',
  hostIpAddress: '',
  healthyFlag: null as string | null
})

// 健康状态选项
const healthStatusOptions = computed(() => [
  { label: t('healthy'), value: 'Y' },
  { label: t('unhealthy'), value: 'N' }
])

// 发出事件
const emit = defineEmits<{
  (e: 'select', resource: JvmResource): void
}>()

// 表格列
const tableColumns = computed(() => {
  const columns = createJvmResourceTableColumns(t)
  
  // 添加操作列
  const actionsColumn = columns.find(col => 'key' in col && col.key === 'actions')
  if (actionsColumn && 'render' in actionsColumn && actionsColumn.render) {
    const originalRender = actionsColumn.render
    actionsColumn.render = (row: JvmResource) => {
      return h(NButton, {
        type: 'primary',
        size: 'small',
        onClick: () => handleViewDetail(row)
      }, {
        default: () => t('monitoringDetail')
      })
    }
  }
  
  return columns
})

// 查看详情
const handleViewDetail = (resource: JvmResource) => {
  emit('select', resource)
}

// 分页配置
const paginationReactive = computed(() => ({
  page: queryParams.value.pageNum || 1,
  pageSize: queryParams.value.pageSize || 20,
  itemCount: total.value,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100]
}))

// 搜索
const handleSearch = () => {
  queryJvmResources({
    applicationName: searchParams.value.applicationName,
    groupName: searchParams.value.groupName,
    hostIpAddress: searchParams.value.hostIpAddress,
    healthyFlag: searchParams.value.healthyFlag || undefined,
    pageNum: 1
  })
}

// 重置
const handleReset = () => {
  searchParams.value = {
    applicationName: '',
    groupName: '',
    hostIpAddress: '',
    healthyFlag: null
  }
  resetQuery()
}

// 分页变更
const handlePageChange = (page: number) => {
  queryJvmResources({ pageNum: page })
}

const handlePageSizeChange = (pageSize: number) => {
  queryJvmResources({ pageNum: 1, pageSize })
}

// 初始化
onMounted(() => {
  queryJvmResources()
})
</script>

<style scoped lang="scss">

</style>

