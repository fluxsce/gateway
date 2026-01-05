<template>
  <GModal
    :visible="modalVisible"
    :title="props.title || '选择服务定义'"
    :width="props.width || 1200"
    :to="props.to"
    :show-footer="true"
    :show-confirm="true"
    :show-cancel="true"
    confirm-text="确认选择"
    cancel-text="取消"
    :header-icon="CubeOutline"
    @update:visible="handleUpdateVisible"
    @confirm="handleConfirm"
    @cancel="handleCancel"
  >
    <div class="service-definition-list-modal" :id="model.moduleId">
      <GPane direction="vertical" :no-resize="true">
        <!-- 上部：搜索表单 -->
        <template #1>
          <search-form
            ref="searchFormRef"
            :module-id="model.moduleId"
            v-bind="model.searchFormConfig"
            @search="handleSearch"
            @toolbar-click="handleToolbarClick"
          />
        </template>

        <!-- 下部：数据表格 -->
        <template #2>
          <g-grid
            ref="gridRef"
            :module-id="model.moduleId"
            :data="model.serviceList"
            :loading="model.loading"
            v-bind="model.gridConfig"
            @page-change="handlePageChange"
            @checkbox-change="handleCheckboxChange"
          >
            <!-- 服务类型自定义渲染 -->
            <template #serviceType="{ row }">
              <n-tag :type="row.serviceType === 1 ? 'success' : 'info'" size="small">
                {{ row.serviceType === 1 ? '服务发现' : '静态配置' }}
              </n-tag>
            </template>

            <!-- 负载均衡策略自定义渲染 -->
            <template #loadBalanceStrategy="{ row }">
              <n-tag type="default" size="small">
                {{ getLoadBalanceText(row.loadBalanceStrategy) }}
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
    </div>
  </GModal>
</template>

<script lang="ts" setup>
import { GModal } from '@/components/gmodal'
import { GPane } from '@/components/gpane'
import SearchForm from '@/components/form/search/SearchForm.vue'
import { GGrid } from '@/components/grid'
import { CubeOutline } from '@vicons/ionicons5'
import { NTag, useMessage } from 'naive-ui'
import { ref, watch } from 'vue'
import { useServiceDefinitionSelectorModel } from './hooks/model'
import { useServiceDefinitionListPage } from './hooks/page'
import type { ServiceDefinitionListModalEmits, ServiceDefinitionListModalProps } from './hooks/types'
import type { ServiceDefinition } from './types'

// 定义组件名称
defineOptions({
  name: 'ServiceDefinitionListModal'
})

// ============= Props =============

const props = withDefaults(defineProps<ServiceDefinitionListModalProps>(), {
  visible: false,
  title: '选择服务定义',
  width: 1200,
  to: undefined,
  gatewayInstanceId: undefined,
})

// ============= Emits =============

const emit = defineEmits<ServiceDefinitionListModalEmits>()

// ============= Message =============

const message = useMessage()

// ============= Refs =============

const searchFormRef = ref()
const gridRef = ref()

// ============= 当前选中的服务定义 =============

const selectedServices = ref<ServiceDefinition[]>([])

// ============= Model =============

const model = useServiceDefinitionSelectorModel()

// ============= 网关实例ID =============

const gatewayInstanceId = ref<string | undefined>(props.gatewayInstanceId)

// 监听 props.gatewayInstanceId 变化
const stopGatewayInstanceIdWatch = watch(() => props.gatewayInstanceId, (newVal) => {
  gatewayInstanceId.value = newVal
})

// ============= 页面级 Hook（包含服务与事件处理） =============

const {
  service,
  handleSearch,
  handleReset,
  handlePageChange,
  handleToolbarClick,
} = useServiceDefinitionListPage(model, gatewayInstanceId, gridRef, searchFormRef)

// ============= 模态框可见性 =============

const modalVisible = ref(props.visible)

// 监听 props.visible 变化，同步到本地状态
const stopVisibleWatch = watch(() => props.visible, (newVal) => {
  modalVisible.value = newVal
  // 移除自动加载逻辑，数据加载应该由用户主动搜索触发
  // 这样避免在对话框打开时立即触发不必要的 API 请求
})

// ============= 资源清理 =============

// 组件卸载时清理所有监听器
onBeforeUnmount(() => {
  stopGatewayInstanceIdWatch()
  stopVisibleWatch()
})

// ============= 工具函数 =============

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

// ============= 事件处理 =============

/**
 * 处理复选框变化（始终支持多选）
 */
const handleCheckboxChange = (selection: ServiceDefinition[]) => {
  selectedServices.value = selection
}

/**
 * 处理模态框可见性变化
 */
const handleUpdateVisible = (value: boolean) => {
  // 更新本地状态
  modalVisible.value = value
  // 通知父组件
  emit('update:visible', value)
  if (!value) {
    // 关闭时清空选中状态
    selectedServices.value = []
    emit('close')
  } else {
    // 模态框打开时触发刷新事件
    emit('refresh')
    // 数据加载由 watch(() => props.visible) 处理，避免重复调用
  }
}

/**
 * 处理确认按钮点击
 */
const handleConfirm = () => {
  if (selectedServices.value.length === 0) {
    message.warning('请至少选择一个服务定义')
    return
  }
  // 触发选择事件，传递服务数组（单个服务时也是数组）
  emit('select', selectedServices.value)
  // 关闭对话框
  modalVisible.value = false
  emit('update:visible', false)
  emit('close')
}

/**
 * 处理取消按钮点击
 */
const handleCancel = () => {
  // 关闭对话框
  modalVisible.value = false
  emit('update:visible', false)
  emit('close')
}

</script>

<style scoped>
.service-definition-list-modal {
  height: 100%;
  display: flex;
  flex-direction: column;
}
</style>

