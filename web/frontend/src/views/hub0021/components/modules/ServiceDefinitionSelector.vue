<template>
  <div class="service-definition-selector">
    <div class="field-block">
      <RsLabel>关联服务 <span class="required">*</span></RsLabel>

      <!-- 当前选择的服务定义 -->
      <div v-if="modelValue && currentService" class="selected-service">
        <div class="service-card">
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
              <div
                class="status-indicator"
                :class="currentService.activeFlag === 'Y' ? 'active' : 'inactive'"
              >
                {{ currentService.activeFlag === 'Y' ? '启用' : '禁用' }}
              </div>
            </div>
          </div>

          <div class="service-details">
            <div class="detail-row">
              <div class="detail-item">
                <span class="detail-label">服务类型</span>
                <RsTag :variant="currentService.serviceType === 1 ? 'success' : 'info'" size="sm">
                  {{ currentService.serviceType === 1 ? '服务发现' : '静态配置' }}
                </RsTag>
              </div>
              <div class="detail-item">
                <span class="detail-label">负载均衡</span>
                <span class="detail-value">
                  {{ getLoadBalanceText(currentService.loadBalanceAlgorithm) }}
                </span>
              </div>
            </div>
            <div class="detail-row">
              <div class="detail-item full-width">
                <span class="detail-label">健康检查</span>
                <div
                  class="health-status"
                  :class="currentService.healthCheckEnabled === 'Y' ? 'enabled' : 'disabled'"
                >
                  {{ currentService.healthCheckEnabled === 'Y' ? '已启用' : '已禁用' }}
                </div>
              </div>
            </div>
          </div>

          <div class="service-actions">
            <RsButton variant="secondary" size="sm" @click="showSelector = true">
              <GIcon :icon="RefreshOutline" size="sm" />
              重新选择
            </RsButton>
            <RsButton variant="secondary" tone="danger" size="sm" @click="handleClear">
              <GIcon :icon="CloseOutline" size="sm" />
              清除
            </RsButton>
          </div>
        </div>
      </div>

      <!-- 未选择时的选择按钮 -->
      <div v-else class="empty-selector">
        <RsButton
          variant="secondary"
          size="lg"
          class="select-btn"
          :loading="loading"
          @click="showSelector = true"
        >
          <GIcon :icon="ServerOutline" size="md" />
          点击选择服务定义
        </RsButton>
      </div>

      <p class="field-hint">
        选择要关联的后端服务定义，如果没有可用选项，请先在服务管理中创建服务定义
      </p>
    </div>

    <!-- 服务定义选择对话框 -->
    <RsDialog
      :open="showSelector"
      layout="window"
      title="选择服务定义"
      :width="1000"
      :close-on-overlay-click="false"
      :draggable="true"
      @update:open="(open) => (showSelector = open)"
    >
      <template #body>
        <div class="selector-toolbar">
          <RsInput
            v-model="searchKeyword"
            placeholder="搜索服务名称、ID或描述"
            clearable
            class="search-input"
          >
            <template #prefix>
              <GIcon :icon="SearchOutline" size="sm" />
            </template>
          </RsInput>

          <div class="toolbar-right">
            <RsTag v-if="filteredServices.length > 0" variant="info" size="sm">
              找到 {{ filteredServices.length }} 个服务定义
            </RsTag>
            <RsButton
              variant="secondary"
              size="sm"
              :loading="loading"
              @click="loadServiceDefinitions"
            >
              <GIcon :icon="RefreshOutline" size="sm" />
              刷新
            </RsButton>
          </div>
        </div>

        <RsTable
          :columns="columns"
          :data="pagedServices"
          row-key="serviceDefinitionId"
          :loading="loading"
          selectable
          selection-type="radio"
          :selected-row-keys="selectedRowKeys"
          size="sm"
          striped
          height="400"
          class="service-table"
          @update:selected-row-keys="onSelectedRowKeysChange"
          @row-click="handleRowClick"
        />

        <RsEmpty
          v-if="!loading && filteredServices.length === 0"
          description="暂无服务定义数据"
          class="table-empty"
        >
          <RsButton size="sm" variant="secondary" @click="loadServiceDefinitions">
            重新加载
          </RsButton>
        </RsEmpty>

        <div v-if="filteredServices.length > 0" class="selector-pagination">
          <RsPagination
            v-model:page="currentPage"
            v-model:page-size="pageSize"
            :total="filteredServices.length"
            size="sm"
            show-page-size
            :page-size-options="[8, 10, 15, 20]"
          />
        </div>
      </template>

      <template #footer>
        <div class="dialog-footer">
          <span class="footer-hint">
            {{
              selectedRowKeys.length > 0
                ? `已选择: ${selectedService?.serviceName}`
                : '请选择一个服务定义'
            }}
          </span>
          <div class="footer-actions">
            <RsButton variant="secondary" @click="showSelector = false">取消</RsButton>
            <RsButton variant="primary" :disabled="!selectedService" @click="handleConfirm">
              确定选择
            </RsButton>
          </div>
        </div>
      </template>
    </RsDialog>
  </div>
</template>

<script setup lang="ts">
import { GIcon } from '@/components/gicon'
import { useAppMessage } from '@/composables/useAppMessage'
import { isApiSuccess } from '@/utils/format'
import {
  RsButton,
  RsDialog,
  RsEmpty,
  RsInput,
  RsLabel,
  RsPagination,
  RsTable,
  RsTag,
  type RsTableColumn,
} from '@/ui'
import {
  CloseOutline,
  RefreshOutline,
  SearchOutline,
  ServerOutline,
} from '@vicons/ionicons5'
import { computed, h, onMounted, ref, watch } from 'vue'
import { getServiceDefinitionById, queryServiceDefinitions } from '../../api'
import type { ServiceDefinition } from '../../types'

interface Props {
  modelValue?: string
  gatewayInstanceId?: string
  loading?: boolean
}

interface Emits {
  (e: 'update:modelValue', value: string | null): void
  (e: 'change', serviceDefinition: ServiceDefinition | null): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()
const message = useAppMessage()

const showSelector = ref(false)
const searchKeyword = ref('')
const serviceDefinitions = ref<ServiceDefinition[]>([])
const selectedRowKeys = ref<string[]>([])
const loading = ref(false)
const currentPage = ref(1)
const pageSize = ref(10)
const selectedServiceInfo = ref<ServiceDefinition | null>(null)

const currentService = computed(() => {
  if (
    selectedServiceInfo.value &&
    selectedServiceInfo.value.serviceDefinitionId === props.modelValue
  ) {
    return selectedServiceInfo.value
  }
  return serviceDefinitions.value.find((s) => s.serviceDefinitionId === props.modelValue) || null
})

const selectedService = computed(() => {
  const key = selectedRowKeys.value[0]
  return key ? serviceDefinitions.value.find((s) => s.serviceDefinitionId === key) : null
})

const filteredServices = computed(() => {
  if (!searchKeyword.value) {
    return serviceDefinitions.value
  }

  const keyword = searchKeyword.value.toLowerCase()
  return serviceDefinitions.value.filter(
    (service) =>
      service.serviceName.toLowerCase().includes(keyword) ||
      service.serviceDefinitionId.toLowerCase().includes(keyword) ||
      (service.serviceDesc && service.serviceDesc.toLowerCase().includes(keyword)),
  )
})

const pagedServices = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  return filteredServices.value.slice(start, start + pageSize.value)
})

const columns: RsTableColumn<ServiceDefinition>[] = [
  {
    title: '服务ID',
    key: 'serviceDefinitionId',
    width: 150,
    render: (row) =>
      h('div', { class: 'service-id-cell' }, [
        h('div', { class: 'service-id-text' }, row.serviceDefinitionId),
        row.serviceDesc ? h('div', { class: 'service-desc' }, row.serviceDesc) : null,
      ]),
  },
  {
    title: '服务名称',
    key: 'serviceName',
    width: 140,
    render: (row) => h('div', { class: 'service-name-cell' }, row.serviceName),
  },
  {
    title: '服务类型',
    key: 'serviceType',
    width: 90,
    render: (row) => {
      const text = row.serviceType === 1 ? '服务发现' : '静态配置'
      const variant = row.serviceType === 1 ? 'success' : 'info'
      return h(RsTag, { variant, size: 'sm' }, () => text)
    },
  },
  {
    title: '负载均衡',
    key: 'loadBalanceAlgorithm',
    width: 100,
    render: (row) =>
      h('span', { style: 'font-size: 13px;' }, getLoadBalanceText(row.loadBalanceAlgorithm)),
  },
  {
    title: '健康检查',
    key: 'healthCheckEnabled',
    width: 90,
    render: (row) => {
      const isEnabled = row.healthCheckEnabled === 'Y'
      return h(
        'div',
        {
          class: 'health-check-cell',
          style: `color: ${isEnabled ? '#18a058' : '#8a8a8a'}; font-size: 12px; font-weight: 500;`,
        },
        isEnabled ? '已启用' : '已禁用',
      )
    },
  },
  {
    title: '状态',
    key: 'activeFlag',
    width: 70,
    render: (row) => {
      const isActive = row.activeFlag === 'Y'
      return h(
        'div',
        {
          class: 'active-status-cell',
          style: `color: ${isActive ? '#18a058' : '#d03050'}; font-size: 12px; font-weight: 500;`,
        },
        isActive ? '启用' : '禁用',
      )
    },
  },
]

const getLoadBalanceText = (algorithm: string): string => {
  const map: Record<string, string> = {
    'round-robin': '轮询',
    random: '随机',
    'ip-hash': 'IP哈希',
    'least-conn': '最少连接',
    'weighted-round-robin': '加权轮询',
    'consistent-hash': '一致性哈希',
  }
  return map[algorithm] || algorithm
}

const getServiceInitial = (serviceName: string): string => {
  if (!serviceName) return '?'
  const firstChar = serviceName.charAt(0).toUpperCase()
  return /^[A-Z]$/.test(firstChar) ? firstChar : serviceName.charAt(0)
}

/**
 * 根据服务定义ID加载服务信息
 */
const loadServiceById = async (serviceDefinitionId: string) => {
  if (!serviceDefinitionId) {
    return
  }

  try {
    loading.value = true
    const response = await getServiceDefinitionById(serviceDefinitionId)

    if (isApiSuccess(response)) {
      const service = JSON.parse(response.bizData) as ServiceDefinition
      selectedServiceInfo.value = service
    } else {
      selectedServiceInfo.value = null
    }
  } catch {
    selectedServiceInfo.value = null
  } finally {
    loading.value = false
  }
}

const loadServiceDefinitions = async (): Promise<void> => {
  if (!props.gatewayInstanceId) {
    serviceDefinitions.value = []
    return Promise.resolve()
  }

  try {
    loading.value = true
    const response = await queryServiceDefinitions({
      gatewayInstanceId: props.gatewayInstanceId,
      pageIndex: 1,
      pageSize: 1000,
    })

    if (isApiSuccess(response)) {
      const pageData = JSON.parse(response.bizData)
      serviceDefinitions.value = pageData?.list || pageData || []

      if (props.modelValue) {
        const found = serviceDefinitions.value.find(
          (s: ServiceDefinition) => s.serviceDefinitionId === props.modelValue,
        )
        if (!found) {
          await loadServiceById(props.modelValue)
        } else {
          selectedServiceInfo.value = found
        }
      }
    } else {
      serviceDefinitions.value = []
    }
  } catch {
    serviceDefinitions.value = []
    message.error('加载服务定义列表失败')
  } finally {
    loading.value = false
  }
}

const onSelectedRowKeysChange = (keys: string[]) => {
  selectedRowKeys.value = keys.slice(0, 1)
}

const handleRowClick = (row: ServiceDefinition) => {
  selectedRowKeys.value = [row.serviceDefinitionId]
}

const handleConfirm = () => {
  if (selectedService.value) {
    selectedServiceInfo.value = selectedService.value
    emit('update:modelValue', selectedService.value.serviceDefinitionId)
    emit('change', selectedService.value)
    showSelector.value = false
    message.success(`已选择服务: ${selectedService.value.serviceName}`)
  }
}

const handleClear = () => {
  emit('update:modelValue', null)
  emit('change', null)
  selectedServiceInfo.value = null
  message.info('已清除服务定义选择')
}

watch(
  () => props.gatewayInstanceId,
  (newId) => {
    if (newId) {
      loadServiceDefinitions()
    } else {
      serviceDefinitions.value = []
      selectedServiceInfo.value = null
    }
  },
)

watch(
  () => props.modelValue,
  (newValue, oldValue) => {
    if (!newValue) {
      if (selectedServiceInfo.value) {
        selectedServiceInfo.value = null
      }
    } else if (newValue !== oldValue) {
      if (!selectedServiceInfo.value || selectedServiceInfo.value.serviceDefinitionId !== newValue) {
        const found = serviceDefinitions.value.find(
          (s: ServiceDefinition) => s.serviceDefinitionId === newValue,
        )
        if (found) {
          selectedServiceInfo.value = found
        } else {
          loadServiceById(newValue)
        }
      }
    }
  },
  { immediate: true },
)

watch(
  () => showSelector.value,
  (show) => {
    if (show) {
      selectedRowKeys.value = props.modelValue ? [props.modelValue] : []
      searchKeyword.value = ''
      currentPage.value = 1
      if (props.gatewayInstanceId) {
        loadServiceDefinitions().then(() => {
          if (props.modelValue) {
            selectedRowKeys.value = [props.modelValue]
          }
        })
      }
    }
  },
)

watch(searchKeyword, () => {
  currentPage.value = 1
})

onMounted(() => {
  if (props.gatewayInstanceId) {
    loadServiceDefinitions()
  }
})
</script>

<style scoped>
.service-definition-selector {
  width: 100%;
}

.field-block {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 100%;
}

.required {
  color: var(--g-danger, #d03050);
}

.field-hint {
  margin: 0;
  font-size: 12px;
  color: var(--g-text-tertiary, var(--rs-text-tertiary));
  line-height: 1.4;
}

.selected-service {
  width: 100%;
}

.service-card {
  background: var(--g-bg-primary, var(--rs-surface));
  border: 1px solid var(--g-border-primary, var(--rs-border));
  border-radius: 8px;
  padding: 16px;
  transition: all 0.2s ease;
}

.service-card:hover {
  border-color: var(--g-primary, var(--rs-primary));
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.service-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

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
}

.service-name {
  font-size: 16px;
  font-weight: 600;
  color: var(--g-text-primary, var(--rs-text));
  margin-bottom: 4px;
  line-height: 1.2;
}

.service-id {
  font-size: 12px;
  color: var(--g-text-tertiary, var(--rs-text-tertiary));
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  background: var(--g-bg-secondary, var(--rs-surface-hover));
  padding: 2px 6px;
  border-radius: 4px;
  display: inline-block;
  line-height: 1.2;
}

.service-status {
  flex-shrink: 0;
}

.status-indicator {
  font-size: 13px;
  font-weight: 500;
  padding: 2px 8px;
  border-radius: 4px;
}

.status-indicator.active {
  color: #18a058;
  background-color: rgba(24, 160, 88, 0.1);
}

.status-indicator.inactive {
  color: #d03050;
  background-color: rgba(208, 48, 80, 0.1);
}

.service-details {
  margin-bottom: 16px;
}

.detail-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  margin-bottom: 12px;
}

.detail-row:last-child {
  margin-bottom: 0;
}

.detail-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.detail-item.full-width {
  grid-column: 1 / -1;
}

.detail-label {
  font-size: 12px;
  color: var(--g-text-tertiary, var(--rs-text-tertiary));
  font-weight: 500;
  min-width: 60px;
  flex-shrink: 0;
}

.detail-value {
  font-size: 13px;
  color: var(--g-text-secondary, var(--rs-text-secondary));
  background: var(--g-bg-secondary, var(--rs-surface-hover));
  padding: 2px 8px;
  border-radius: 4px;
  font-weight: 500;
}

.health-status {
  font-size: 13px;
  font-weight: 500;
  padding: 2px 8px;
  border-radius: 4px;
}

.health-status.enabled {
  color: #18a058;
  background-color: rgba(24, 160, 88, 0.1);
}

.health-status.disabled {
  color: #8a8a8a;
  background-color: rgba(138, 138, 138, 0.1);
}

.service-actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
  padding-top: 12px;
  border-top: 1px solid var(--g-border-primary, var(--rs-border));
}

.empty-selector {
  width: 100%;
}

.select-btn {
  width: 100%;
  min-height: 56px;
}

.selector-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 16px;
  padding: 12px 16px;
  background: var(--g-bg-primary, var(--rs-surface));
  border: 1px solid var(--g-border-primary, var(--rs-border));
  border-radius: 8px;
}

.search-input {
  width: 350px;
  max-width: 100%;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.service-table {
  margin-top: 8px;
}

.table-empty {
  margin-top: 16px;
}

.selector-pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 12px;
}

.dialog-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  width: 100%;
}

.footer-hint {
  font-size: 13px;
  color: var(--g-text-tertiary, var(--rs-text-tertiary));
}

.footer-actions {
  display: flex;
  gap: 8px;
}

:deep(.service-id-cell) {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 4px 0;
}

:deep(.service-id-text) {
  font-size: 11px;
  color: var(--g-primary, var(--rs-primary));
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  background: var(--g-primary-light, var(--rs-primary-container));
  padding: 3px 8px;
  border-radius: 8px;
  width: fit-content;
  font-weight: 500;
}

:deep(.service-name-cell) {
  font-weight: 600;
  font-size: 14px;
  color: var(--g-text-primary, var(--rs-text));
  line-height: 1.3;
  padding: 4px 0;
}

:deep(.service-desc) {
  font-size: 12px;
  color: var(--g-text-secondary, var(--rs-text-secondary));
  line-height: 1.3;
  opacity: 0.9;
  padding: 2px 0;
}
</style>
