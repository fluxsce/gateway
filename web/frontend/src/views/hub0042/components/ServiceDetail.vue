<template>
  <div class="service-detail" v-if="service">
    <div class="service-detail-header">
      <h2 class="service-detail-title">{{ t('hub0042.serviceDetail') }}</h2>
      <div class="service-detail-actions">
        <RsButton variant="secondary" icon="pencil" @click="handleEdit">
          {{ t('hub0042.editService') }}
        </RsButton>
        <RsButton variant="primary" icon="arrow-left" @click="handleBack">
          {{ t('hub0042.back') }}
        </RsButton>
      </div>
    </div>

    <div class="service-detail-body">
      <ServiceRelationBoard class="service-detail-board" :service="service" />

      <RsDescriptions class="service-detail-meta" :columns="2" bordered size="sm" label-placement="left">
        <RsDescriptionsItem label="服务名">
          {{ service.serviceName }}
        </RsDescriptionsItem>
        <RsDescriptionsItem label="分组">
          {{ service.groupName }}
        </RsDescriptionsItem>
        <RsDescriptionsItem label="保护阈值">
          {{ service.protectThreshold ?? 0 }}
        </RsDescriptionsItem>
        <RsDescriptionsItem label="服务类型">
          {{ getServiceTypeLabel(service.serviceType) }}
        </RsDescriptionsItem>
        <RsDescriptionsItem label="服务版本">
          {{ service.serviceVersion || '-' }}
        </RsDescriptionsItem>
        <RsDescriptionsItem label="服务路由类型">
          {{ getSelectorType(service.selectorJson) }}
        </RsDescriptionsItem>
        <RsDescriptionsItem label="服务描述" :span="2">
          {{ service.serviceDescription || '-' }}
        </RsDescriptionsItem>
        <RsDescriptionsItem label="元数据" :span="2">
          <RsCodeBlock
            :code="service.metadataJson || '{}'"
            lang="json"
          />
        </RsDescriptionsItem>
      </RsDescriptions>

      <div class="service-detail-nodes">
        <div class="instance-list-header">
          <span>{{ t('hub0042.nodeList') }}</span>
          <RsTag variant="info" size="sm">
            {{ t('hub0042.nodeCount', { count: service.nodes?.length || 0 }) }}
          </RsTag>
        </div>
        <ServiceNodeList
          class="service-detail-node-list"
          :nodes="service.nodes || []"
          :loading="loading"
          @refresh="handleRefresh"
        />
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n'
import { RsButton, RsDescriptions, RsDescriptionsItem, RsTag } from '@/ui'
import { RsCodeBlock } from '@/ui/code-block'
import type { Service } from '../types'
import ServiceNodeList from './ServiceNodeList.vue'
import ServiceRelationBoard from './ServiceRelationBoard.vue'

defineOptions({
  name: 'ServiceDetail'
})

interface Props {
  service: Service | null
  loading?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  service: null,
  loading: false,
})

interface Emits {
  (e: 'back'): void
  (e: 'edit'): void
  (e: 'refresh'): void
}

const emit = defineEmits<Emits>()
const { t } = useI18n()

// 工具方法
const getServiceTypeLabel = (type: string) => {
  const typeMap: Record<string, string> = {
    'INTERNAL': '内部服务',
    'NACOS': 'Nacos',
    'CONSUL': 'Consul',
    'EUREKA': 'Eureka',
    'ETCD': 'ETCD',
    'ZOOKEEPER': 'Zookeeper',
  }
  return typeMap[type] || type
}

const getSelectorType = (selectorJson?: string) => {
  if (!selectorJson) return 'none'
  try {
    const selector = JSON.parse(selectorJson)
    return selector.type || 'none'
  } catch {
    return 'none'
  }
}



// 事件处理
const handleBack = () => {
  emit('back')
}

const handleEdit = () => {
  emit('edit')
}

const handleRefresh = () => {
  emit('refresh')
}
</script>

<style lang="scss" scoped>
.service-detail {
  height: 100%;
  display: flex;
  flex-direction: column;
  padding: var(--g-space-sm);
  overflow: hidden;
  gap: var(--g-space-sm);
  min-height: 0;
}

.service-detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-shrink: 0;
  padding: var(--g-space-sm) var(--g-space-md);
}

.service-detail-title {
  margin: 0;
  font-size: 16px;
  font-weight: 500;
}

.service-detail-actions {
  display: flex;
  gap: var(--g-space-xs);
}

.service-detail-body {
  flex: 1;
  min-height: 0;
  overflow: auto;
  display: flex;
  flex-direction: column;
  gap: var(--g-space-md);
}

.service-detail-board {
  flex: none;
  height: 420px;
}

.service-detail-meta {
  flex-shrink: 0;
}

.service-detail-nodes {
  flex: none;
  display: flex;
  flex-direction: column;
  gap: var(--g-space-sm);
}

.instance-list-header {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.service-detail-node-list {
  height: 360px;
}
</style>
