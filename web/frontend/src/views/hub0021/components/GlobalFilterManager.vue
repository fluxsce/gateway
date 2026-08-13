<template>
  <div class="global-filter-manager">
    <div class="toolbar">
      <div class="toolbar-left">
        <div class="toolbar-title">
          <span class="toolbar-title__strong">全局过滤器管理</span>
          <span class="toolbar-title__muted">管理作用于所有路由的全局前置过滤器</span>
        </div>
      </div>
      <div class="toolbar-right">
        <RsButton variant="primary" :disabled="!gatewayInstanceId" @click="handleCreate">
          <GIcon :icon="Add" size="sm" />
          新建过滤器
        </RsButton>
        <RsButton :loading="loading" @click="handleRefresh">
          <GIcon :icon="Refresh" size="sm" />
          刷新
        </RsButton>
      </div>
    </div>

    <div class="filter-content">
      <!-- 原 GlobalFilterList / FilterConfigDialog 模块缺失，保留占位避免断链编译失败 -->
      <RsEmpty
        v-if="globalFilters.length === 0 && !loading"
        description="暂无全局过滤器，或列表组件尚未接入"
      />
      <RsLoading v-else-if="loading" overlay block size="md" />
      <div v-else class="filter-placeholder">
        已加载 {{ globalFilters.length }} 条全局过滤器（列表 UI 待接入）
      </div>
    </div>

    <RsAlert type="info" title="全局过滤器说明" style="margin-top: 16px;">
      全局过滤器配置作用于该网关实例的所有路由，仅支持前置处理类型：
      <ul style="margin: 8px 0; padding-left: 20px;">
        <li><strong>前置处理</strong>：在路由匹配前执行，如认证、限流、请求头处理、参数验证等</li>
        <li>支持多个前置过滤器，按执行顺序依次处理</li>
        <li>路由级别的过滤器为后置处理类型，在路由匹配后执行</li>
      </ul>
      支持的过滤器类型：header、query-param、body、strip、rewrite、method、cookie、response
    </RsAlert>
  </div>
</template>

<script setup lang="ts">
import { GIcon } from '@/components/gicon'
import { useAppMessage } from '@/composables/useAppMessage'
import { RsAlert, RsButton, RsEmpty, RsLoading } from '@/ui'
import { Add, Refresh } from '@vicons/ionicons5'
import { getCurrentInstance, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import {
  addFilterConfig,
  deleteFilterConfig,
  editFilterConfig,
  queryFilterConfigs,
} from '../api'
import type { FilterConfig, FilterFormData } from '../types/filterConfig'

interface Props {
  gatewayInstanceId: string
}

const props = defineProps<Props>()
const message = useAppMessage()
const instance = getCurrentInstance()

const loading = ref(false)
const dialogVisible = ref(false)
const currentFilter = ref<FilterConfig | null>(null)
const globalFilters = ref<FilterConfig[]>([])
const isUnmounted = ref(false)

const safeMessage = {
  success: (msg: string) => {
    if (!isUnmounted.value && instance && !instance.isUnmounted) {
      nextTick(() => {
        if (!isUnmounted.value) message.success(msg)
      })
    }
  },
  error: (msg: string) => {
    if (!isUnmounted.value && instance && !instance.isUnmounted) {
      nextTick(() => {
        if (!isUnmounted.value) message.error(msg)
      })
    }
  },
  warning: (msg: string) => {
    if (!isUnmounted.value && instance && !instance.isUnmounted) {
      nextTick(() => {
        if (!isUnmounted.value) message.warning(msg)
      })
    }
  },
}

const loadGlobalFilters = async () => {
  if (!props.gatewayInstanceId || isUnmounted.value) return

  loading.value = true
  try {
    const response = await queryFilterConfigs({
      gatewayInstanceId: props.gatewayInstanceId,
      pageIndex: 1,
      pageSize: 10000,
    })

    if (isUnmounted.value) return

    if (response?.oK && response.bizData) {
      const parseData = JSON.parse(response.bizData)
      const filterList = Array.isArray(parseData) ? parseData : (parseData?.list || parseData?.data || [])
      globalFilters.value = filterList
        .filter((filter: FilterConfig) => filter.filterAction === 'pre-routing')
        .sort((a: FilterConfig, b: FilterConfig) => a.filterOrder - b.filterOrder)
    } else {
      globalFilters.value = []
    }
  } catch (error) {
    if (!isUnmounted.value) {
      console.error('加载全局过滤器失败:', error)
      safeMessage.error('加载全局过滤器失败')
      globalFilters.value = []
    }
  } finally {
    if (!isUnmounted.value) loading.value = false
  }
}

const handleCreate = () => {
  if (isUnmounted.value) return
  currentFilter.value = null
  dialogVisible.value = true
  safeMessage.warning('过滤器编辑对话框尚未接入')
}

const handleEdit = (filter: FilterConfig) => {
  if (isUnmounted.value) return
  currentFilter.value = filter
  dialogVisible.value = true
}

const handleDelete = async (filter: FilterConfig) => {
  if (isUnmounted.value) return
  try {
    loading.value = true
    const response = await deleteFilterConfig(filter.filterConfigId)
    if (isUnmounted.value) return
    if (response?.oK) {
      safeMessage.success('删除过滤器成功')
      await loadGlobalFilters()
    } else {
      safeMessage.error(response?.errMsg || '删除过滤器失败')
    }
  } catch (error) {
    if (!isUnmounted.value) {
      console.error('删除过滤器失败:', error)
      safeMessage.error('删除过滤器失败')
    }
  } finally {
    if (!isUnmounted.value) loading.value = false
  }
}

const handleToggleStatus = async (filter: FilterConfig) => {
  if (isUnmounted.value) return
  try {
    const newStatus = filter.activeFlag === 'Y' ? 'N' : 'Y'
    const response = await editFilterConfig({ ...filter, activeFlag: newStatus })
    if (isUnmounted.value) return
    if (response?.oK) {
      safeMessage.success(`过滤器已${newStatus === 'Y' ? '启用' : '禁用'}`)
      await loadGlobalFilters()
    } else {
      safeMessage.error(response?.errMsg || '更新过滤器状态失败')
    }
  } catch (error) {
    if (!isUnmounted.value) {
      console.error('更新过滤器状态失败:', error)
      safeMessage.error('更新过滤器状态失败')
    }
  }
}

const swapFilterOrder = async (filter1: FilterConfig, filter2: FilterConfig) => {
  if (isUnmounted.value) return
  try {
    loading.value = true
    const tempOrder = filter1.filterOrder
    const [response1, response2] = await Promise.all([
      editFilterConfig({ ...filter1, filterOrder: filter2.filterOrder }),
      editFilterConfig({ ...filter2, filterOrder: tempOrder }),
    ])
    if (isUnmounted.value) return
    if (response1?.oK && response2?.oK) {
      safeMessage.success('调整执行顺序成功')
      await loadGlobalFilters()
    } else {
      safeMessage.error('调整执行顺序失败')
    }
  } catch (error) {
    if (!isUnmounted.value) {
      console.error('调整执行顺序失败:', error)
      safeMessage.error('调整执行顺序失败')
    }
  } finally {
    if (!isUnmounted.value) loading.value = false
  }
}

const handleMoveUp = async (filter: FilterConfig) => {
  if (isUnmounted.value) return
  const currentIndex = globalFilters.value.findIndex(f => f.filterConfigId === filter.filterConfigId)
  if (currentIndex <= 0) return
  await swapFilterOrder(filter, globalFilters.value[currentIndex - 1])
}

const handleMoveDown = async (filter: FilterConfig) => {
  if (isUnmounted.value) return
  const currentIndex = globalFilters.value.findIndex(f => f.filterConfigId === filter.filterConfigId)
  if (currentIndex >= globalFilters.value.length - 1) return
  await swapFilterOrder(filter, globalFilters.value[currentIndex + 1])
}

const handleSaveFilter = async (filterData: FilterFormData) => {
  if (isUnmounted.value) return
  try {
    const saveData = {
      tenantId: 'default',
      gatewayInstanceId: props.gatewayInstanceId,
      filterName: filterData.filterName,
      filterType: filterData.filterType,
      filterAction: 'pre-routing',
      filterOrder: filterData.filterOrder,
      filterConfig: JSON.stringify(filterData.config),
      filterDesc: filterData.filterDesc,
      activeFlag: filterData.activeFlag,
      addWho: 'admin',
      editWho: 'admin',
    }

    const response = filterData.filterConfigId
      ? await editFilterConfig({ ...saveData, filterConfigId: filterData.filterConfigId })
      : await addFilterConfig(saveData)

    if (isUnmounted.value) return
    if (response?.oK) {
      safeMessage.success(`${filterData.filterConfigId ? '编辑' : '创建'}过滤器成功`)
      dialogVisible.value = false
      await loadGlobalFilters()
    } else {
      safeMessage.error(response?.errMsg || `${filterData.filterConfigId ? '编辑' : '创建'}过滤器失败`)
    }
  } catch (error) {
    if (!isUnmounted.value) {
      console.error('保存过滤器失败:', error)
      safeMessage.error('保存过滤器失败')
    }
  }
}

const handleRefresh = async () => {
  if (isUnmounted.value) return
  await loadGlobalFilters()
  safeMessage.success('刷新完成')
}

const refresh = async () => {
  if (isUnmounted.value) return
  await loadGlobalFilters()
}

watch(() => props.gatewayInstanceId, async (newId) => {
  if (newId && !isUnmounted.value) await loadGlobalFilters()
}, { immediate: true })

onMounted(() => {
  if (props.gatewayInstanceId) {
    console.log('全局过滤器管理器初始化，网关实例ID:', props.gatewayInstanceId)
  }
})

onUnmounted(() => {
  isUnmounted.value = true
})

defineExpose({
  refresh,
  loadGlobalFilters,
  handleEdit,
  handleDelete,
  handleToggleStatus,
  handleMoveUp,
  handleMoveDown,
  handleSaveFilter,
})
</script>

<style scoped>
.global-filter-manager {
  width: 100%;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding: 16px;
  background: var(--g-color-surface, var(--rs-bg-muted, #fafafa));
  border-radius: 6px;
}

.toolbar-left {
  flex: 1;
}

.toolbar-title {
  display: flex;
  align-items: center;
  gap: 12px;
}

.toolbar-title__strong {
  font-weight: 600;
}

.toolbar-title__muted {
  color: var(--g-text-color-3, var(--rs-text-muted));
  font-size: 13px;
}

.toolbar-right {
  display: flex;
  gap: 8px;
}

.filter-content {
  position: relative;
  margin-bottom: 16px;
  min-height: 80px;
}

.filter-placeholder {
  padding: 16px;
  color: var(--g-text-color-3, var(--rs-text-muted));
}
</style>
