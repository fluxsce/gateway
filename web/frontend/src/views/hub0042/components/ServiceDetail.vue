<template>
  <div class="service-detail" v-if="service">
    <!-- 顶部操作栏 -->
    <div class="service-detail-header">
      <h2 class="service-detail-title">服务详情</h2>
      <div class="service-detail-actions">
        <RsButton variant="secondary" @click="handleEdit">
          <GIcon :icon="CreateOutline" size="sm" />
          编辑服务
        </RsButton>
        <RsButton variant="primary" @click="handleBack">
          <GIcon :icon="ArrowBackOutline" size="sm" />
          返回
        </RsButton>
      </div>
    </div>

    <!-- 基本信息约 40% + 节点列表约 60%（flex 比例固定，无分割条） -->
    <div class="service-detail-body">
      <div class="service-detail-pane service-detail-pane--basic">
        <GCard class="service-detail-card service-detail-card--basic">
          <RsDescriptions :columns="2" bordered size="sm" label-placement="left">
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
              <GTextShow
                :content="service.metadataJson || '{}'"
                format="json"
                :auto-format="true"
                :max-height="200"
                :show-copy-button="true"
              />
            </RsDescriptionsItem>
          </RsDescriptions>
        </GCard>
      </div>

      <div class="service-detail-pane service-detail-pane--nodes">
        <GCard class="service-detail-card service-instance-list-card">
          <template #header>
            <div class="instance-list-header">
              <span>服务实例列表</span>
              <RsTag variant="info" size="sm">
                共 {{ service.nodes?.length || 0 }} 个实例
              </RsTag>
            </div>
          </template>

          <ServiceNodeList
            :nodes="service.nodes || []"
            :loading="loading"
            @refresh="handleRefresh"
          />
        </GCard>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { GCard } from '@/components/gcard'
import { GIcon } from '@/components/gicon'
import GTextShow from '@/components/gtext-show/GTextShow.vue'
import { RsButton, RsDescriptions, RsDescriptionsItem, RsTag } from '@/ui'
import { ArrowBackOutline, CreateOutline } from '@vicons/ionicons5'
import type { Service } from '../types'
import ServiceNodeList from './ServiceNodeList.vue'

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

  .service-detail-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: var(--g-space-sm) var(--g-space-md);

    border-radius: var(--g-border-radius);
    flex-shrink: 0;

    .service-detail-title {
      margin: 0;
      font-size: 16px;
      font-weight: 500;
    }

    .service-detail-actions {
      display: flex;
      gap: var(--g-space-xs);
    }
  }

  .service-detail-body {
    flex: 1;
    min-height: 0;
    width: 100%;
    display: flex;
    flex-direction: column;
    gap: var(--g-space-sm);
  }

  .service-detail-pane {
    min-height: 0;
    display: flex;
    flex-direction: column;
    box-sizing: border-box;
  }

  /* 4 : 6 ≈ 40% : 60% 分配主体剩余高度 */
  .service-detail-pane--basic {
    flex: 4 1 0;
    min-height: 0;
    overflow-x: hidden;
    overflow-y: auto;
    -webkit-overflow-scrolling: touch;
  }

  .service-detail-pane--nodes {
    flex: 6 1 0;
    overflow: hidden;
  }

  /* 基本信息：卡片随内容增高，滚动条在面板内（缩放/小屏时不裁切） */
  .service-detail-pane--basic :deep(.g-card.service-detail-card--basic) {
    flex: 0 0 auto;
    width: 100%;
    height: auto;
    min-height: 0;
    overflow: visible;
  }

  .service-detail-pane--basic :deep(.g-card.service-detail-card--basic .n-card__content) {
    overflow: visible;
    flex: none;
  }

  .service-detail-pane--nodes .service-detail-card {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  /* 节点表格区域在 flex 链路上需要 min-height:0，避免撑不开剩余高度 */
  .service-instance-list-card {
    :deep(.n-card__content) {
      min-height: 0;
    }
  }

  .instance-list-header {
    display: flex;
    align-items: center;
    gap: 8px;
  }
}
</style>
