<template>
  <div class="service-definition-list" id="hub0022-service-definition-list">
    <GPane direction="vertical" default-size="80px">
      <!-- 上部：搜索表单 -->
      <template #1>
        <search-form
          ref="searchFormRef"
          :module-id="service.model.moduleId"
          v-bind="service.model.searchFormConfig"
          @search="handleSearch"
          @toolbar-click="handleToolbarClick"
        />
      </template>

      <!-- 下部：数据表格 -->
      <template #2>
        <g-grid
          ref="gridRef"
          :module-id="service.model.moduleId"
          :data="service.model.serviceList"
          :loading="service.model.loading"
          v-bind="service.model.gridConfig"
          @page-change="service.handlePageChange"
          @menu-click="handleMenuClick"
        >
          <!-- 服务类型自定义渲染 -->
          <template #serviceType="{ row }">
            <n-tag :type="row.serviceType === 0 ? 'info' : 'success'" size="small">
              {{ row.serviceType === 0 ? '静态配置' : '服务发现' }}
            </n-tag>
          </template>

          <!-- 负载均衡策略自定义渲染 -->
          <template #loadBalanceStrategy="{ row }">
            <n-tag type="default" size="small">
              {{ getLoadBalanceStrategyLabel(row.loadBalanceStrategy) }}
            </n-tag>
          </template>

          <!-- 会话亲和性自定义渲染 -->
          <template #sessionAffinity="{ row }">
            <n-tag :type="row.sessionAffinity === 'Y' ? 'success' : 'default'" size="small">
              {{ row.sessionAffinity === 'Y' ? '启用' : '未启用' }}
            </n-tag>
          </template>

          <!-- 熔断器自定义渲染 -->
          <template #enableCircuitBreaker="{ row }">
            <n-tag :type="row.enableCircuitBreaker === 'Y' ? 'warning' : 'default'" size="small">
              {{ row.enableCircuitBreaker === 'Y' ? '启用' : '未启用' }}
            </n-tag>
          </template>

          <!-- 健康检查自定义渲染 -->
          <template #healthCheckEnabled="{ row }">
            <n-tag :type="row.healthCheckEnabled === 'Y' ? 'success' : 'default'" size="small">
              {{ row.healthCheckEnabled === 'Y' ? '已启用' : '未启用' }}
            </n-tag>
          </template>

          <!-- 状态自定义渲染 -->
          <template #activeFlag="{ row }">
            <n-tag :type="row.activeFlag === 'Y' ? 'success' : 'error'" size="small">
              {{ row.activeFlag === 'Y' ? '启用' : '禁用' }}
            </n-tag>
          </template>
        </g-grid>
      </template>
    </GPane>

    <!-- 服务定义对话框（新增/编辑/查看共用） -->
    <GdataFormModal
      v-model:visible="formDialogVisible"
      :mode="formDialogMode"
      :title="formDialogMode === 'create' ? '新增服务定义' : formDialogMode === 'edit' ? '编辑服务定义' : '查看服务定义详情'"
      to="#hub0022-service-definition-list"
      :form-fields="service.model.serviceFormConfig.fields"
      :form-tabs="service.model.serviceFormConfig.tabs"
      :initial-data="currentEditService || undefined"
      :auto-close-on-confirm="false"
      :confirm-loading="service.model.loading.value"
      @submit="handleFormSubmit"
    />

    <!-- 服务注册选择器组件 -->
    <ServiceRegistrySelector
      v-model:show="serviceSelectionVisible"
      :currentService="selectedService"
      @select="handleServiceSelect"
    />

    <!-- 服务节点管理对话框 -->
    <ServiceNodeListModal
      v-model:visible="showNodeDialog"
      :service-definition-id="currentServiceId"
      :title="'服务节点管理'"
      :width="1200"
      to="#hub0022-service-definition-list"
    />
  </div>
</template>

<script lang="ts" setup>
import GdataFormModal from '@/components/form/data/GDataFormModal.vue'
import SearchForm from '@/components/form/search/SearchForm.vue'
import { GPane } from '@/components/gpane'
import { GGrid } from '@/components/grid'
import { NTag } from 'naive-ui'
import { onBeforeUnmount, ref, watch } from 'vue'
import { ServiceNodeListModal } from '../service-nodes'
import ServiceRegistrySelector from './ServiceRegistrySelector.vue'
import { useServiceDefinitionPage } from './hooks/page'
import { LoadBalanceStrategy } from './types'

// 定义组件名称
defineOptions({
  name: 'ServiceDefinitionList'
})

// ============= Props =============

interface Props {
  /** 网关实例ID（作为proxyConfigId使用） */
  gatewayInstanceId?: string
}

const props = withDefaults(defineProps<Props>(), {
  gatewayInstanceId: undefined,
})

// ============= Refs =============

const searchFormRef = ref()
const gridRef = ref()

// ============= 页面级 Hook（包含服务与对话框、事件处理） =============

const {
  service,
  formDialogVisible,
  formDialogMode,
  currentEditService,
  showNodeDialog,
  currentServiceId,
  selectedService,
  serviceSelectionVisible,
  handleFormSubmit,
  handleToolbarClick,
  handleMenuClick,
  handleServiceSelect,
  handleSearch: pageHandleSearch
} = useServiceDefinitionPage(props.gatewayInstanceId, searchFormRef, gridRef)

// ============= 监听器 =============

// 监听 gatewayInstanceId 变化，重新加载数据
const stopGatewayInstanceIdWatch = watch(
  () => props.gatewayInstanceId,
  (newId, oldId) => {
    if (newId && newId !== oldId) {
      // 当实例ID变化时，重新加载服务定义列表（使用最新的 gatewayInstanceId）
      service.loadServiceList({ proxyConfigId: newId })
    } else if (!newId && oldId) {
      // 当实例ID被清空时，清空列表
      service.model.serviceList.value = []
    }
  },
  { immediate: false }
)

// 组件卸载时清理监听器
onBeforeUnmount(() => {
  stopGatewayInstanceIdWatch()
})

// ============= 方法 =============

/**
 * 处理搜索（确保使用最新的 gatewayInstanceId）
 */
function handleSearch(formData?: Record<string, any>) {
  // 校验是否已选择实例
  if (!props.gatewayInstanceId) {
    return
  }
  // 使用最新的 props.gatewayInstanceId 来合并查询参数
  const searchParams = formData
    ? {
        ...formData,
        ...(props.gatewayInstanceId ? { proxyConfigId: props.gatewayInstanceId } : {}),
      }
    : props.gatewayInstanceId
      ? { proxyConfigId: props.gatewayInstanceId }
      : undefined
  // 调用 service 的 handleSearch，传入合并后的参数
  service.handleSearch(searchParams)
}

// ============= 方法 =============

/**
 * 获取负载均衡策略标签
 */
function getLoadBalanceStrategyLabel(strategy: string): string {
  const strategyMap: Record<string, string> = {
    [LoadBalanceStrategy.ROUND_ROBIN]: '轮询',
    [LoadBalanceStrategy.RANDOM]: '随机',
    [LoadBalanceStrategy.IP_HASH]: 'IP哈希',
    [LoadBalanceStrategy.LEAST_CONN]: '最少连接',
    [LoadBalanceStrategy.WEIGHTED_ROUND_ROBIN]: '加权轮询',
    [LoadBalanceStrategy.CONSISTENT_HASH]: '一致性哈希',
  }
  return strategyMap[strategy] || strategy
}
</script>

<style lang="scss" scoped>
.service-definition-list {
  width: 100%;
  height: 100%;
  overflow: hidden;

  :deep(.n-split) {
    height: 100%;
  }

  /* 上半区：搜索表单，内容较少，允许自身滚动 */
  :deep(.n-split-pane:first-child) {
    overflow: auto;
    padding: var(--g-space-sm);
  }

  /* 下半区：表格区域，高度由 GGrid 占满，滚动全部交给 vxe-grid */
  :deep(.n-split-pane:last-child) {
    overflow: hidden;
    padding: var(--g-space-sm);
    display: flex;
    flex-direction: column;
  }
}
</style>

