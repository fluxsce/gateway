<template>
  <RsContextMenu :items="menuItems" @select="onMenuSelect">
    <article
      class="sc-instance-card"
      :class="['is-' + tone, { 'is-selected': selected, 'is-offline': !instance.isRunning }]"
      @click="emit('select')"
      @dblclick="emit('action', 'runtime')"
      @contextmenu="emit('select')"
    >
      <header class="sc-instance-card__head">
        <div class="sc-instance-card__icon" :class="'is-' + envTone" aria-hidden="true">
          <GIcon :icon="CloudOutline" :size="22" color="#fff" />
        </div>
        <div class="sc-instance-card__identity">
          <h3 class="sc-instance-card__name" :title="instance.instanceName">
            {{ instance.instanceName }}
          </h3>
          <p class="sc-instance-card__meta" :title="listenText">
            {{ envText }} · {{ listenText }}
          </p>
          <p class="sc-instance-card__live" :class="'is-' + tone">
            <span class="sc-instance-card__orb" aria-hidden="true" />
            <span>{{ liveText }}</span>
          </p>
        </div>
      </header>

      <div class="sc-instance-card__group">
        <div class="sc-instance-card__row" title="已注册的服务名数量。同一服务名下可以有多个业务节点">
          <span>服务名</span>
          <span>{{ metric(instance.serviceCount) }}</span>
        </div>
        <div class="sc-instance-card__row" title="已注册的业务节点数，每个节点是一个 ip:port，含临时节点和持久节点">
          <span>业务节点</span>
          <span>{{ metric(instance.nodeCount) }}</span>
        </div>
        <div class="sc-instance-card__row" title="状态为 UP 且心跳探测为健康的业务节点。网关重启后持久节点会先标为不健康，收到心跳后再恢复">
          <span>健康节点</span>
          <span>{{ metric(instance.healthyNodeCount) }}</span>
        </div>
        <div class="sc-instance-card__row" title="当前连在数据面的 gRPC 会话数。一条会话不等于一个业务节点">
          <span>gRPC 会话</span>
          <span>{{ metric(instance.connectionCount) }}</span>
        </div>
      </div>

      <footer class="sc-instance-card__foot">
        <span class="sc-instance-card__cluster" :title="clusterHint">{{ clusterText }}</span>
        <span class="sc-instance-card__secure">{{ tlsText }} · {{ authText }}</span>
      </footer>
    </article>
  </RsContextMenu>
</template>

<script lang="ts" setup>
import { GIcon } from '@/components/gicon'
import { RsContextMenu } from '@/ui'
import { CloudOutline } from '@vicons/ionicons5'
import { computed } from 'vue'
import { buildInstanceContextMenu } from '../hooks/model'
import type { ServiceCenterInstance } from '../types'

defineOptions({
  name: 'ServiceCenterInstanceCard',
})

const props = defineProps<{
  instance: ServiceCenterInstance
  selected?: boolean
}>()

const emit = defineEmits<{
  (e: 'select'): void
  (e: 'action', key: string): void
}>()

const envMap: Record<string, string> = {
  DEVELOPMENT: '开发',
  STAGING: '预发布',
  PRODUCTION: '生产',
}

const envText = computed(() => envMap[props.instance.environment] || props.instance.environment || '-')

const envTone = computed(() => {
  if (props.instance.environment === 'PRODUCTION') return 'prod'
  if (props.instance.environment === 'STAGING') return 'stage'
  return 'dev'
})

const tone = computed(() => {
  const status = props.instance.instanceStatus
  if (status === 'RUNNING') return 'running'
  if (status === 'ERROR') return 'error'
  if (status === 'STARTING' || status === 'STOPPING') return 'busy'
  return 'idle'
})

const listenText = computed(() => {
  if (props.instance.listenEndpoint) return props.instance.listenEndpoint
  const host = props.instance.listenAddress || '-'
  const port = props.instance.listenPort
  return port ? `${host}:${port}` : host
})

const liveText = computed(() => {
  if (props.instance.instanceStatus === 'ERROR') return '运行异常'
  if (props.instance.isRunning) return '正在监听'
  if (props.instance.instanceStatus === 'STARTING') return '正在启动'
  if (props.instance.instanceStatus === 'STOPPING') return '正在停止'
  return '尚未监听'
})

const replicaIds = computed(() => {
  const ids = Array.isArray(props.instance.replicaGatewayIds) ? props.instance.replicaGatewayIds.filter(Boolean) : []
  if (ids.length) return ids
  return props.instance.ownerGatewayId ? [props.instance.ownerGatewayId] : []
})

const clusterText = computed(() => {
  if (!props.instance.isRunning) return '未承接注册'
  const n = replicaIds.value.length
  if (n <= 1) return '本机承接注册'
  return `${n} 台网关承接注册`
})

const clusterHint = computed(() => {
  if (!props.instance.isRunning) return '启动后由本机网关进程承接服务注册'
  if (!replicaIds.value.length) return '当前由本机网关进程承接注册'
  const owner = props.instance.ownerGatewayId || ''
  return replicaIds.value
    .map((id) => (owner && id === owner ? `${id}（本机）` : id))
    .join('\n')
})

const tlsText = computed(() => {
  if (props.instance.enableTLS !== 'Y') return 'TLS 关'
  return props.instance.enableMTLS === 'Y' ? 'mTLS' : 'TLS 开'
})

const authText = computed(() => (props.instance.enableAuth === 'Y' ? '认证开' : '认证关'))

const menuItems = computed(() => buildInstanceContextMenu(props.instance))

function metric(value?: number) {
  if (!props.instance.isRunning) return '—'
  return value === undefined || value === null ? '—' : String(value)
}

function onMenuSelect(key: string) {
  emit('action', key)
}
</script>

<style lang="scss" scoped>
.sc-instance-card {
  --sc-card-radius: 14px;
  --sc-group-radius: 10px;
  --sc-inset: rgba(120, 120, 128, 0.12);
  --sc-hairline: rgba(60, 60, 67, 0.12);
  --sc-caption: #86868b;

  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  width: 100%;
  padding: 16px 16px 12px;
  border: 1px solid var(--g-border-color, var(--sc-hairline));
  border-radius: var(--sc-card-radius);
  background: var(--g-bg-primary);
  box-shadow: none;
  cursor: pointer;
  user-select: none;
  -webkit-user-select: none;
}

.sc-instance-card:hover {
  border-color: color-mix(in srgb, var(--g-primary) 45%, var(--g-border-color, var(--sc-hairline)));
}

.sc-instance-card.is-selected {
  border-color: var(--g-primary);
}

.sc-instance-card__head {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  min-width: 0;
}

.sc-instance-card__icon {
  flex: none;
  display: grid;
  place-items: center;
  width: 44px;
  height: 44px;
  border-radius: 12px;
  color: #fff;
}

.sc-instance-card__icon.is-dev {
  background: linear-gradient(180deg, #64d2ff 0%, #007aff 100%);
}

.sc-instance-card__icon.is-stage {
  background: linear-gradient(180deg, #ffd60a 0%, #ff9f0a 100%);
}

.sc-instance-card__icon.is-prod {
  background: linear-gradient(180deg, #bf5af2 0%, #af52de 100%);
}

.sc-instance-card__identity {
  min-width: 0;
  flex: 1 1 auto;
}

.sc-instance-card__name {
  margin: 0;
  font-size: 17px;
  font-weight: 600;
  line-height: 1.25;
  letter-spacing: -0.022em;
  color: var(--g-text-primary);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.sc-instance-card__meta {
  margin: 3px 0 0;
  font-size: 13px;
  line-height: 18px;
  color: var(--sc-caption);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.sc-instance-card__live {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  margin: 8px 0 0;
  font-size: 13px;
  font-weight: 500;
  line-height: 18px;
  letter-spacing: -0.01em;
  color: var(--g-text-secondary);
}

.sc-instance-card__orb {
  flex: none;
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #c7c7cc;
}

.sc-instance-card__live.is-running {
  color: #30d158;
}

.sc-instance-card__live.is-running .sc-instance-card__orb {
  background: #30d158;
  box-shadow: 0 0 0 3px rgba(48, 209, 88, 0.18);
  animation: sc-live-pulse 2.4s ease-out infinite;
}

.sc-instance-card__live.is-error {
  color: #ff453a;
}

.sc-instance-card__live.is-error .sc-instance-card__orb {
  background: #ff453a;
  box-shadow: 0 0 0 3px rgba(255, 69, 58, 0.16);
}

.sc-instance-card__live.is-busy {
  color: #ff9f0a;
}

.sc-instance-card__live.is-busy .sc-instance-card__orb {
  background: #ff9f0a;
  box-shadow: 0 0 0 3px rgba(255, 159, 10, 0.18);
}

.sc-instance-card__group {
  margin-top: 16px;
  border-radius: var(--sc-group-radius);
  background: var(--sc-inset);
  overflow: hidden;
}

.sc-instance-card.is-offline .sc-instance-card__group {
  opacity: 0.62;
}

.sc-instance-card__row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  min-height: 36px;
  padding: 0 12px;
  font-size: 13px;
  line-height: 18px;
  letter-spacing: -0.01em;
}

.sc-instance-card__row span:first-child {
  color: var(--g-text-secondary);
}

.sc-instance-card__row span:last-child {
  color: var(--g-text-primary);
  font-variant-numeric: tabular-nums;
  font-weight: 590;
}

.sc-instance-card__row + .sc-instance-card__row {
  box-shadow: inset 0 0.5px 0 var(--sc-hairline);
}

.sc-instance-card__foot {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 12px;
  margin-top: 12px;
  min-width: 0;
  font-size: 12px;
  line-height: 16px;
  color: var(--sc-caption);
}

.sc-instance-card__cluster {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.sc-instance-card__secure {
  flex: none;
}

@keyframes sc-live-pulse {
  0% {
    box-shadow: 0 0 0 0 rgba(48, 209, 88, 0.35);
  }
  70% {
    box-shadow: 0 0 0 7px rgba(48, 209, 88, 0);
  }
  100% {
    box-shadow: 0 0 0 0 rgba(48, 209, 88, 0);
  }
}

@media (prefers-reduced-motion: reduce) {
  .sc-instance-card__live.is-running .sc-instance-card__orb {
    animation: none;
  }
}

[data-theme='dark'] .sc-instance-card {
  --sc-inset: rgba(235, 235, 245, 0.08);
  --sc-hairline: rgba(235, 235, 245, 0.12);
  --sc-caption: #98989d;
}

[data-theme='dark'] .sc-instance-card__live.is-idle {
  color: #98989d;
}
</style>
