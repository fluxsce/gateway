<template>
  <RsDrawer
    v-model:open="drawerVisible"
    :title="drawerTitle"
    side="right"
    size="lg"
    :teleport-to="`#${moduleId}`"
    :show-overlay="false"
    :close-on-overlay-click="false"
  >
    <div class="center-runtime-drawer">
      <RsLoading v-if="loading" block size="lg" />
      <template v-else>
        <section class="center-runtime-drawer__section">
          <h3>运行时概览</h3>
          <p class="center-runtime-drawer__hint">
            服务名是注册表里的名称；业务节点是每个 ip:port；健康节点需 UP 且心跳正常；gRPC 会话是客户端长连接，和节点不是一对一。
          </p>
          <RsDescriptions :columns="2" bordered size="sm" label-placement="left">
            <RsDescriptionsItem label="引擎">
              {{ overview?.engine || '-' }}
            </RsDescriptionsItem>
            <RsDescriptionsItem label="中心实例">
              {{ overview?.centerInstanceName || instance?.instanceName || '-' }}
            </RsDescriptionsItem>
            <RsDescriptionsItem label="部署环境">
              {{ envText }}
            </RsDescriptionsItem>
            <RsDescriptionsItem label="服务名">
              {{ overview?.serviceCount ?? '-' }}
            </RsDescriptionsItem>
            <RsDescriptionsItem label="业务节点">
              {{ overview?.nodeCount ?? '-' }}
            </RsDescriptionsItem>
            <RsDescriptionsItem label="健康节点">
              {{ overview?.healthyNodeCount ?? '-' }}
            </RsDescriptionsItem>
            <RsDescriptionsItem label="已发布配置">
              {{ overview?.configCount ?? '-' }}
            </RsDescriptionsItem>
            <RsDescriptionsItem label="gRPC 会话">
              {{ overview?.connectionCount ?? '-' }}
            </RsDescriptionsItem>
            <RsDescriptionsItem label="监听地址">
              {{ instance?.listenEndpoint || `${instance?.listenAddress || ''}:${instance?.listenPort || ''}` }}
            </RsDescriptionsItem>
            <RsDescriptionsItem label="承接注册的网关">
              {{ clusterText }}
            </RsDescriptionsItem>
            <RsDescriptionsItem label="本机进程">
              {{ overview?.ownerGatewayId || '本机进程' }}
            </RsDescriptionsItem>
          </RsDescriptions>
        </section>

        <section class="center-runtime-drawer__section">
          <div class="center-runtime-drawer__section-head">
            <h3>数据面连接</h3>
            <RsTag variant="info" size="sm">{{ connections.length }} 条会话</RsTag>
          </div>
          <RsEmpty v-if="!connections.length" description="当前没有活跃的数据面会话" />
          <RsGrid
            v-else
            module-id="hub0040:runtime"
            :data="connections"
            :columns="connectionColumns"
            :selectable="false"
            row-key="connectionId"
            height="360px"
          />
        </section>
      </template>
    </div>
  </RsDrawer>
</template>

<script lang="ts" setup>
import { RsGrid, type RsGridColumn } from '@/components/rs-grid'
import { RsDescriptions, RsDescriptionsItem, RsDrawer, RsEmpty, RsLoading, RsTag } from '@/ui'
import { formatDate } from '@/utils/format'
import { computed } from 'vue'
import type { CenterConnection, CenterOverview, ServiceCenterInstance } from '../types'

defineOptions({
  name: 'CenterRuntimeDrawer',
})

const props = withDefaults(defineProps<{
  visible: boolean
  loading?: boolean
  instance?: ServiceCenterInstance | null
  overview?: CenterOverview | null
  connections?: CenterConnection[]
  moduleId?: string
}>(), {
  loading: false,
  instance: null,
  overview: null,
  connections: () => [],
  moduleId: 'hub0040',
})

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
}>()

const drawerVisible = computed({
  get: () => props.visible,
  set: (value: boolean) => emit('update:visible', value),
})

const drawerTitle = computed(() => {
  const name = props.instance?.instanceName
  const env = props.instance?.environment
  if (name && env) return `运行时观测 · ${name} · ${env}`
  if (name) return `运行时观测 · ${name}`
  return '运行时观测'
})

const envText = computed(() => {
  const env = props.instance?.environment
  const envMap: Record<string, string> = {
    DEVELOPMENT: '开发环境',
    STAGING: '预发布环境',
    PRODUCTION: '生产环境',
  }
  return env ? envMap[env] || env : '-'
})

const replicaIds = computed(() => {
  const ids = Array.isArray(props.overview?.replicaGatewayIds)
    ? props.overview.replicaGatewayIds.filter(Boolean)
    : []
  if (ids.length) return ids
  return props.overview?.ownerGatewayId ? [props.overview.ownerGatewayId] : []
})

const clusterText = computed(() => {
  const n = replicaIds.value.length
  if (!n) return '仅本机网关进程承接注册'
  if (n === 1) return '1 台网关进程（本机）'
  return `${n} 台网关进程共同承接注册`
})

const connectionColumns: RsGridColumn<CenterConnection>[] = [
  { key: 'connectionId', title: '连接ID', ellipsis: true },
  { key: 'clientId', title: 'Client ID', ellipsis: true },
  { key: 'clientIp', title: '客户端 IP', width: 140 },
  { key: 'namespaceId', title: '命名空间', ellipsis: true, width: 140 },
  {
    key: 'lastActive',
    title: '最近活跃',
    width: 170,
    formatter: (value) => (value ? formatDate(value as string, 'YYYY-MM-DD HH:mm:ss') : '-'),
  },
]
</script>

<style lang="scss" scoped>
.center-runtime-drawer {
  display: flex;
  flex-direction: column;
  gap: var(--g-space-lg);
  min-height: 0;
}

.center-runtime-drawer__section h3 {
  margin: 0 0 var(--g-space-sm);
  font-size: 14px;
  font-weight: 600;
}

.center-runtime-drawer__hint {
  margin: 0 0 var(--g-space-sm);
  color: var(--g-text-secondary);
  font-size: 12px;
  line-height: 1.55;
}

.center-runtime-drawer__section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--g-space-sm);
}

.center-runtime-drawer__section-head h3 {
  margin: 0;
}
</style>
