<template>
  <n-drawer
    v-model:show="visible"
    :width="900"
    placement="right"
    :closable="true"
  >
    <n-drawer-content
      :title="t('serviceDetail')"
      closable
    >
      <template v-if="serviceDetail">
        <!-- 服务基本信息 -->
        <div class="service-info-section">
          <h3 class="section-title">{{ t('basicInfo') }}</h3>
          <n-descriptions 
            :column="2" 
            label-placement="left"
            bordered
          >
            <n-descriptions-item :label="t('serviceName')">
              {{ serviceDetail.serviceName }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('columns.groupName')">
              {{ serviceDetail.groupName }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('columns.protocolType')">
              <n-tag :type="getProtocolTagType(serviceDetail.protocolType)">
                {{ serviceDetail.protocolType }}
              </n-tag>
            </n-descriptions-item>
            <n-descriptions-item :label="t('columns.loadBalanceStrategy')">
              <n-tag type="info">
                {{ getLoadBalanceText(serviceDetail.loadBalanceStrategy) }}
              </n-tag>
            </n-descriptions-item>
            <n-descriptions-item :label="t('columns.contextPath')">
              {{ serviceDetail.contextPath || '/' }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('healthCheckUrl')">
              {{ serviceDetail.healthCheckUrl }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('columns.registryType')">
              <n-tag type="primary">
                {{ getRegistryTypeText(serviceDetail.registryType) }}
              </n-tag>
            </n-descriptions-item>
          </n-descriptions>

          <n-descriptions 
            :column="1" 
            label-placement="left"
            bordered
            style="margin-top: 16px"
          >
            <n-descriptions-item :label="t('serviceDescription')">
              {{ serviceDetail.serviceDescription || t('noDescription') }}
            </n-descriptions-item>
          </n-descriptions>
        </div>



        <!-- 实例列表 -->
        <div class="instances-section">
          <div class="section-header">
            <h3 class="section-title">{{ t('instanceList') }}</h3>
            <n-button 
              size="small" 
              @click="refreshInstances"
              :loading="instancesLoading"
            >
              <template #icon>
                <n-icon><RefreshOutline /></n-icon>
              </template>
              {{ t('actions.refresh') }}
            </n-button>
          </div>

          <n-data-table
            :columns="instanceColumns"
            :data="serviceDetail.instances"
            :loading="instancesLoading"
            :pagination="false"
            size="small"
            style="width: 100%; overflow-x: auto;"
            :title="t('instanceList')"
            :scroll-x="1800"
          />
        </div>

        <!-- 健康检查配置 -->
        <div class="health-check-section">
          <h3 class="section-title">{{ t('healthCheckConfig') }}</h3>
          <n-descriptions 
            :column="3" 
            label-placement="left"
            bordered
          >
            <n-descriptions-item :label="t('healthCheckInterval')">
              {{ serviceDetail.healthCheckIntervalSeconds }}s
            </n-descriptions-item>
            <n-descriptions-item :label="t('healthCheckTimeout')">
              {{ serviceDetail.healthCheckTimeoutSeconds }}s
            </n-descriptions-item>
            <n-descriptions-item :label="t('healthCheckType')">
              <n-tag type="info">{{ serviceDetail.healthCheckType }}</n-tag>
            </n-descriptions-item>
            <n-descriptions-item :label="t('healthCheckMode')">
              <n-tag type="success">{{ getHealthCheckModeText(serviceDetail.healthCheckMode) }}</n-tag>
            </n-descriptions-item>
            <n-descriptions-item :label="t('lastUpdateTime')">
              <n-time :time="new Date(serviceDetail.editTime)" />
            </n-descriptions-item>
          </n-descriptions>
        </div>
      </template>

      <template v-else-if="loading">
        <n-spin size="large" style="width: 100%; padding: 60px 0;" />
      </template>

      <template v-else>
        <n-empty :description="t('serviceNotFound')" />
      </template>
    </n-drawer-content>
  </n-drawer>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { 
  NDrawer, NDrawerContent, NDescriptions, NDescriptionsItem,
  NTag, NGrid, NGi, NStatistic, NDataTable, NButton, 
  NIcon, NSpin, NEmpty, NTime
} from 'naive-ui'
import { RefreshOutline } from '@vicons/ionicons5'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { useServiceDetail } from '../hooks'
import { createInstanceTableColumns } from '../models/instanceTableColumns'
import type { ProtocolType, LoadBalanceStrategy, HealthCheckType, HealthCheckMode, RegistryType } from '../types'

interface Props {
  show: boolean
  serviceName: string
}

interface Emits {
  (e: 'update:show', value: boolean): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

// 国际化
const { t } = useModuleI18n('hub0041')

// 控制抽屉显示
const visible = computed({
  get: () => props.show,
  set: (value) => emit('update:show', value)
})

// 服务详情管理
const {
  serviceDetail,
  loading,
  instancesLoading,
  fetchServiceDetail,
  refreshServiceInstances,
  performHealthCheck,
  updateInstanceStatus
} = useServiceDetail()

// 处理实例操作
const handleInstanceAction = async (action: string, instance: any) => {
  if (!props.serviceName) return
  
  switch (action) {
    case 'health-check':
      await performHealthCheck(props.serviceName, instance.serviceInstanceId)
      break
    case 'up':
      await updateInstanceStatus(props.serviceName, instance, 'UP')
      break
    case 'down':
      await updateInstanceStatus(props.serviceName, instance, 'DOWN')
      break
  }
}

// 实例表格列配置
const instanceColumns = createInstanceTableColumns(t, handleInstanceAction)

// 定义一个加载标志，防止重复加载
const isLoading = ref(false)

// 监听服务名称变化和抽屉显示状态的组合
watch([() => props.serviceName, () => props.show], async ([newServiceName, show]) => {
  // 只有当服务名称存在且抽屉显示时才加载数据
  if (newServiceName && show && !isLoading.value) {
    try {
      isLoading.value = true
      await fetchServiceDetail(newServiceName)
      // 确保服务详情加载完成后再加载实例
      if (serviceDetail.value) {
        await refreshServiceInstances(newServiceName)
      }
    } finally {
      isLoading.value = false
    }
  }
}, { immediate: true })

// 刷新实例列表
const refreshInstances = async () => {
  if (props.serviceName) {
    await refreshServiceInstances(props.serviceName)
  }
}

// 获取协议标签类型
const getProtocolTagType = (protocol: ProtocolType) => {
  switch (protocol) {
    case 'HTTP': return 'success'
    case 'HTTPS': return 'warning'
    case 'TCP': return 'info'
    case 'UDP': return 'default'
    case 'GRPC': return 'error'
    default: return 'default'
  }
}

// 获取负载均衡策略文本
const getLoadBalanceText = (strategy: LoadBalanceStrategy) => {
  const strategyMap = {
    'ROUND_ROBIN': t('hub0041.roundRobin'),
    'WEIGHTED_ROUND_ROBIN': t('hub0041.weightedRoundRobin'),
    'LEAST_CONNECTIONS': t('hub0041.leastConnections'),
    'RANDOM': t('hub0041.random'),
    'IP_HASH': t('hub0041.ipHash')
  }
  return strategyMap[strategy] || strategy
}

// 获取健康检查模式文本
const getHealthCheckModeText = (mode: HealthCheckMode) => {
  const modeMap = {
    'ACTIVE': t('healthCheckModes.active'),
    'PASSIVE': t('healthCheckModes.passive')
  }
  return modeMap[mode] || mode
}

// 获取注册类型文本
const getRegistryTypeText = (registryType: RegistryType) => {
  const registryTypeMap = {
    'INTERNAL': t('registryTypes.INTERNAL'),
    'NACOS': t('registryTypes.NACOS'),
    'CONSUL': t('registryTypes.CONSUL'),
    'EUREKA': t('registryTypes.EUREKA'),
    'ETCD': t('registryTypes.ETCD'),
    'ZOOKEEPER': t('registryTypes.ZOOKEEPER')
  }
  return registryTypeMap[registryType] || registryType
}
</script>

<style scoped lang="scss">
:deep(.n-drawer-content) {
  .n-drawer-body {
    padding: 0;
  }
}

.service-info-section,
.instances-section,
.health-check-section {
  margin-bottom: 32px;

  .section-title {
    margin: 0 0 16px 0;
    font-size: 16px;
    font-weight: 600;
    color: var(--text-color-1);
    border-bottom: 2px solid var(--primary-color);
    padding-bottom: 8px;
  }

  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;

    .section-title {
      margin: 0;
      border: none;
      padding: 0;
    }
  }
}



:deep(.n-descriptions) {
  .n-descriptions-item-label {
    font-weight: 500;
  }
}

:deep(.n-data-table) {
  width: 100%;
  overflow-x: auto;
  
  .n-scrollbar-rail.n-scrollbar-rail--horizontal {
    bottom: 0;
    right: 0;
    left: 0;
  }
  
  .status-up {
    color: var(--success-color);
  }

  .status-down {
    color: var(--error-color);
  }

  .status-starting {
    color: var(--warning-color);
  }

  .health-healthy {
    color: var(--success-color);
  }

  .health-unhealthy {
    color: var(--error-color);
  }

  .health-unknown {
    color: var(--text-color-3);
  }
}
</style>
