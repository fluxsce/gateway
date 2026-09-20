<template>
  <section class="ns-board" :class="['is-' + liveTone, { 'is-idle': !ready }]">
    <div class="ns-board__glow" aria-hidden="true" />
    <div class="ns-board__grid" aria-hidden="true" />
    <div class="ns-board__scan" aria-hidden="true" />

    <header class="ns-board__head">
      <div class="ns-board__identity">
        <p class="ns-board__kicker">
          <span>NAMING CENTER</span>
          <i />
          <span>{{ kickerText }}</span>
        </p>
        <h2 :title="scopeText">{{ scopeText }}</h2>
      </div>
      <dl class="ns-board__meta">
        <div>
          <dt>环境</dt>
          <dd>{{ envLabel }}</dd>
        </div>
        <div>
          <dt>状态</dt>
          <dd :class="'is-' + liveTone">
            <i class="ns-board__orb" aria-hidden="true" />
            {{ liveText }}
          </dd>
        </div>
        <div>
          <dt>时间</dt>
          <dd class="is-clock">{{ clockText }}</dd>
        </div>
      </dl>
    </header>

    <div class="ns-board__body">
      <div class="ns-board__dial-card">
        <svg class="ns-board__ring" viewBox="0 0 120 120" aria-hidden="true">
          <defs>
            <linearGradient id="nsRingGrad" x1="0%" y1="0%" x2="100%" y2="100%">
              <stop offset="0%" stop-color="var(--g-primary)" />
              <stop offset="100%" stop-color="var(--g-info)" />
            </linearGradient>
            <filter id="nsRingGlow" x="-20%" y="-20%" width="140%" height="140%">
              <feGaussianBlur stdDeviation="2.2" result="blur" />
              <feMerge>
                <feMergeNode in="blur" />
                <feMergeNode in="SourceGraphic" />
              </feMerge>
            </filter>
          </defs>
          <circle class="ns-board__ring-track" cx="60" cy="60" r="50" />
          <circle
            class="ns-board__ring-value"
            :class="'is-' + healthTone"
            cx="60"
            cy="60"
            r="50"
            :stroke-dasharray="ringLength"
            :stroke-dashoffset="ringOffset"
            filter="url(#nsRingGlow)"
          />
        </svg>
        <div class="ns-board__dial-read">
          <strong>{{ healthRateText }}</strong>
          <span>节点健康率</span>
        </div>
        <p class="ns-board__dial-sub">{{ healthPair }}</p>
      </div>

      <div class="ns-board__panel">
        <div class="ns-board__kpis">
          <article
            v-for="item in kpiItems"
            :key="item.key"
            class="ns-board__kpi"
            :class="'is-' + item.tone"
          >
            <span>{{ item.label }}</span>
            <strong>{{ item.value }}</strong>
            <i class="ns-board__bar" aria-hidden="true">
              <b :style="{ width: item.bar + '%' }" />
            </i>
          </article>
        </div>
        <footer class="ns-board__foot">
          <span :title="clusterFull">{{ clusterText }}</span>
          <span>{{ summaryText }}</span>
        </footer>
      </div>
    </div>
  </section>
</template>

<script lang="ts" setup>
import { computed, onMounted, onUnmounted, ref } from 'vue'
import type { CenterOverview } from '../../hub0040/types'
import type { Namespace } from '../types'

defineOptions({
  name: 'NamespaceMonitorBoard',
})

const props = withDefaults(defineProps<{
  instanceName?: string
  environment?: string
  namespaces?: Namespace[]
  totalCount?: number
  pageCount?: number
  overview?: CenterOverview | null
  online?: boolean
  loading?: boolean
  selectedNamespace?: Namespace | null
}>(), {
  instanceName: '',
  environment: '',
  namespaces: () => [],
  totalCount: 0,
  pageCount: 0,
  overview: null,
  online: false,
  loading: false,
  selectedNamespace: null,
})

const envMap: Record<string, string> = {
  DEVELOPMENT: '开发',
  STAGING: '预发布',
  PRODUCTION: '生产',
}

const ringLength = 2 * Math.PI * 50
const now = ref(Date.now())
let timer = 0

onMounted(() => {
  timer = window.setInterval(() => {
    now.value = Date.now()
  }, 1000)
})

onUnmounted(() => {
  window.clearInterval(timer)
})

const clockText = computed(() => {
  const date = new Date(now.value)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`
})

const ready = computed(() => Boolean(props.instanceName))
const selected = computed(() => props.selectedNamespace || null)
const scoped = computed(() => Boolean(selected.value))
const scopedOverview = computed(() => {
  const ns = selected.value
  const ov = props.overview
  if (!ns || !ov) return null
  if (ov.scope === 'namespace' && ov.namespaceId === ns.namespaceId) return ov
  return null
})

const envLabel = computed(() => {
  if (!ready.value) return '--'
  const env = selected.value?.environment || props.environment
  if (!env) return '随实例'
  return envMap[env] || env
})

const kickerText = computed(() => (scoped.value ? 'NAMESPACE' : 'LIVE MONITOR'))

const scopeText = computed(() => {
  if (!ready.value) return '请选择服务中心实例'
  if (selected.value) return selected.value.namespaceName || selected.value.namespaceId
  return props.instanceName
})

const live = computed(() => Boolean(props.online && props.overview && !props.loading))
const liveTone = computed(() => {
  if (!ready.value) return 'idle'
  if (props.loading) return 'loading'
  if (scoped.value && selected.value?.activeFlag === 'N') return 'down'
  return live.value ? 'running' : 'down'
})
const liveText = computed(() => {
  if (!ready.value) return '未圈定'
  if (props.loading) return '同步中'
  if (scoped.value) {
    if (!live.value) return '离线'
    return selected.value?.activeFlag === 'Y' ? '活动' : '非活动'
  }
  return live.value ? '在线' : '离线'
})

const serviceCount = computed(() => {
  if (scopedOverview.value) return scopedOverview.value.serviceCount ?? 0
  if (selected.value) return selected.value.serviceCount ?? 0
  return props.overview?.serviceCount ?? 0
})
const nodeCount = computed(() => {
  if (scopedOverview.value) return scopedOverview.value.nodeCount ?? 0
  if (selected.value) return selected.value.nodeCount ?? 0
  return props.overview?.nodeCount ?? 0
})
const healthyNodeCount = computed(() => {
  if (scopedOverview.value) return scopedOverview.value.healthyNodeCount ?? 0
  if (selected.value) return selected.value.healthyNodeCount ?? 0
  return props.overview?.healthyNodeCount ?? 0
})
const configCount = computed(() => {
  if (scopedOverview.value) return scopedOverview.value.configCount ?? 0
  return scoped.value ? 0 : (props.overview?.configCount ?? 0)
})
const connectionCount = computed(() => {
  if (scopedOverview.value) return scopedOverview.value.connectionCount ?? 0
  if (selected.value) return selected.value.connectionCount ?? 0
  return props.overview?.connectionCount ?? 0
})
const serviceQuota = computed(() => selected.value?.serviceQuotaLimit ?? scopedOverview.value?.serviceQuotaLimit ?? 0)
const configQuota = computed(() => selected.value?.configQuotaLimit ?? scopedOverview.value?.configQuotaLimit ?? 0)
const pageActive = computed(() => props.namespaces.filter((row) => row.activeFlag === 'Y').length)

const healthRate = computed(() => {
  if (!ready.value || !nodeCount.value) return 0
  return Math.min(100, Math.round((healthyNodeCount.value / nodeCount.value) * 1000) / 10)
})

const ringOffset = computed(() => ringLength * (1 - healthRate.value / 100))

const healthTone = computed(() => {
  if (!ready.value || !nodeCount.value) return 'idle'
  if (healthRate.value >= 99) return 'ok'
  if (healthRate.value >= 80) return 'warn'
  return 'bad'
})

const healthRateText = computed(() => {
  if (!ready.value || !nodeCount.value) return '--'
  return `${formatRate(healthRate.value)}%`
})

const healthPair = computed(() => {
  if (!ready.value) return '等待圈定实例'
  if (scoped.value && !scopedOverview.value && props.loading) return '正在读取该命名空间'
  if (!nodeCount.value) return scoped.value ? '该命名空间无注册节点' : '当前无注册节点'
  return `${healthyNodeCount.value} 健康 / ${nodeCount.value} 节点`
})

const kpiItems = computed(() => {
  const nodeTone = ready.value && serviceCount.value > 0 && nodeCount.value === 0 ? 'warn' : 'normal'
  let healthItemTone = 'normal'
  if (healthTone.value === 'bad') healthItemTone = 'bad'
  else if (healthTone.value === 'warn') healthItemTone = 'warn'
  const rows = scoped.value
    ? [
        { key: 'svc', label: '服务', raw: serviceCount.value, tone: 'normal', value: display(serviceCount.value) },
        { key: 'node', label: '节点', raw: nodeCount.value, tone: nodeTone, value: display(nodeCount.value) },
        { key: 'ok', label: '健康', raw: healthyNodeCount.value, tone: healthItemTone, value: display(healthyNodeCount.value) },
        { key: 'conn', label: '会话', raw: connectionCount.value, tone: 'normal', value: display(connectionCount.value) },
        { key: 'cfg', label: '配置', raw: configCount.value, tone: 'normal', value: display(configCount.value) },
        { key: 'sq', label: '服务配额', raw: serviceCount.value, tone: quotaTone(serviceCount.value, serviceQuota.value), value: quotaText(serviceCount.value, serviceQuota.value) },
        { key: 'cq', label: '配置配额', raw: configCount.value, tone: quotaTone(configCount.value, configQuota.value), value: quotaText(configCount.value, configQuota.value) },
      ]
    : [
        { key: 'ns', label: '命名空间', raw: props.totalCount, tone: 'normal', value: display(props.totalCount) },
        { key: 'active', label: '活动', raw: pageActive.value, tone: 'normal', value: display(pageActive.value) },
        { key: 'svc', label: '服务', raw: serviceCount.value, tone: 'normal', value: display(serviceCount.value) },
        { key: 'node', label: '节点', raw: nodeCount.value, tone: nodeTone, value: display(nodeCount.value) },
        { key: 'ok', label: '健康', raw: healthyNodeCount.value, tone: healthItemTone, value: display(healthyNodeCount.value) },
        { key: 'conn', label: '会话', raw: connectionCount.value, tone: 'normal', value: display(connectionCount.value) },
        { key: 'cfg', label: '配置', raw: configCount.value, tone: 'normal', value: display(configCount.value) },
      ]
  const peak = Math.max(...rows.map((row) => row.raw), 1)
  return rows.map((row) => ({
    ...row,
    bar: barWidth(row.key, row.raw, peak),
  }))
})

const replicaIds = computed(() => (
  Array.isArray(props.overview?.replicaGatewayIds)
    ? props.overview.replicaGatewayIds.filter(Boolean)
    : []
))

const clusterFull = computed(() => {
  if (scoped.value && selected.value) {
    return `命名空间 ${selected.value.namespaceId}`
  }
  const owner = props.overview?.ownerGatewayId
  if (owner && replicaIds.value.length > 1) {
    return `承接注册 ${owner} · ${replicaIds.value.length} 个网关进程`
  }
  if (owner) return `承接注册 ${owner}`
  return ''
})

const clusterText = computed(() => {
  if (!ready.value) return '监控范围未圈定'
  if (scoped.value && selected.value) {
    const id = selected.value.namespaceId
    const short = id.length > 16 ? `${id.slice(0, 12)}…` : id
    return `命名空间 ${short}`
  }
  const owner = props.overview?.ownerGatewayId
  if (owner) {
    const short = owner.length > 12 ? `${owner.slice(0, 8)}…` : owner
    if (replicaIds.value.length > 1) return `承接网关 ${short} · ${replicaIds.value.length} 进程`
    return `承接网关 ${short}`
  }
  if (live.value) return '本机网关承接注册'
  return '运行时暂不可达'
})

const summaryText = computed(() => {
  if (!ready.value) return '服务中心实例必选'
  if (scoped.value && selected.value) {
    return `监控 ${selected.value.namespaceName || selected.value.namespaceId} · ${props.instanceName}`
  }
  const page = props.pageCount || props.namespaces.length
  if (page === props.totalCount) return `${props.totalCount} 个命名空间`
  return `本页 ${page} / ${props.totalCount}`
})

function display(value: number) {
  if (!ready.value) return '--'
  return String(value)
}

function quotaText(used: number, limit: number) {
  if (!ready.value) return '--'
  if (!limit) return `${used} / 不限`
  return `${used} / ${limit}`
}

function quotaTone(used: number, limit: number) {
  if (!ready.value || !limit) return 'normal'
  const rate = used / limit
  if (rate >= 1) return 'bad'
  if (rate >= 0.8) return 'warn'
  return 'normal'
}

function barWidth(key: string, raw: number, peak: number) {
  if (!ready.value) return 8
  if (key === 'sq') return quotaBar(serviceCount.value, serviceQuota.value)
  if (key === 'cq') return quotaBar(configCount.value, configQuota.value)
  return Math.max(8, Math.round((raw / peak) * 100))
}

function quotaBar(used: number, limit: number) {
  if (!limit) return used > 0 ? 36 : 8
  return Math.max(8, Math.min(100, Math.round((used / limit) * 100)))
}

function formatRate(value: number) {
  return Number.isInteger(value) ? String(value) : value.toFixed(1)
}
</script>

<style lang="scss" scoped>
.ns-board {
  --ns-accent: var(--g-primary);
  --ns-info: var(--g-info);
  --ns-ok: var(--g-success);
  --ns-warn: var(--g-warning);
  --ns-bad: var(--g-error);
  --ns-text: var(--g-text-primary);
  --ns-muted: var(--g-text-tertiary);
  --ns-line: color-mix(in srgb, var(--g-primary) 22%, var(--g-border-primary));
  --ns-card: var(--g-bg-primary);
  --ns-glow-a: color-mix(in srgb, var(--g-primary) 14%, transparent);
  --ns-glow-b: color-mix(in srgb, var(--g-info) 12%, transparent);
  --ns-grid: color-mix(in srgb, var(--g-primary) 10%, transparent);
  --ns-scan: color-mix(in srgb, var(--g-primary) 10%, transparent);
  --ns-inset: color-mix(in srgb, var(--g-primary) 7%, transparent);
  --ns-track: color-mix(in srgb, var(--g-text-tertiary) 22%, transparent);
  --ns-bar: color-mix(in srgb, var(--g-text-tertiary) 16%, transparent);
  --ns-orb: var(--g-text-disabled);

  position: relative;
  isolation: isolate;
  overflow: hidden;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  width: 100%;
  flex: none;
  color: var(--ns-text);
  background:
    radial-gradient(1200px 180px at 12% -20%, var(--ns-glow-a), transparent 55%),
    radial-gradient(900px 220px at 92% 120%, var(--ns-glow-b), transparent 50%),
    linear-gradient(180deg, var(--g-bg-secondary) 0%, var(--g-bg-tertiary) 100%);
  border-bottom: 1px solid var(--g-border-primary);
  box-shadow: inset 0 1px 0 color-mix(in srgb, var(--g-primary) 20%, transparent);
}

.ns-board::before,
.ns-board::after {
  content: '';
  position: absolute;
  width: 18px;
  height: 18px;
  z-index: 3;
  border: 1px solid color-mix(in srgb, var(--g-primary) 48%, transparent);
  pointer-events: none;
}

.ns-board::before {
  top: 8px;
  left: 8px;
  border-right: none;
  border-bottom: none;
}

.ns-board::after {
  right: 8px;
  bottom: 8px;
  border-left: none;
  border-top: none;
}

.ns-board__glow,
.ns-board__grid,
.ns-board__scan {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.ns-board__grid {
  background-image:
    linear-gradient(var(--ns-grid) 1px, transparent 1px),
    linear-gradient(90deg, var(--ns-grid) 1px, transparent 1px);
  background-size: 28px 28px;
}

.ns-board__scan {
  height: 28%;
  background: linear-gradient(180deg, transparent, var(--ns-scan), transparent);
  animation: ns-board-scan 5.5s linear infinite;
}

.ns-board.is-idle {
  filter: saturate(0.86);
}

.ns-board__head,
.ns-board__body,
.ns-board__foot {
  position: relative;
  z-index: 1;
}

.ns-board__head {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 16px;
  min-width: 0;
  padding: 16px 22px 12px;
}

.ns-board__identity {
  min-width: 0;
}

.ns-board__kicker {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 0 0 6px;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.18em;
  color: var(--ns-accent);
}

.ns-board__kicker i {
  display: block;
  width: 36px;
  height: 1px;
  background: linear-gradient(90deg, var(--ns-accent), transparent);
}

.ns-board__identity h2 {
  margin: 0;
  font-size: 26px;
  font-weight: 700;
  line-height: 32px;
  letter-spacing: -0.03em;
  color: var(--ns-text);
  text-shadow: 0 0 18px color-mix(in srgb, var(--g-primary) 18%, transparent);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.ns-board__meta {
  display: flex;
  gap: 22px;
  margin: 0;
  flex: none;
}

.ns-board__meta dt {
  margin: 0 0 4px;
  font-size: 11px;
  letter-spacing: 0.12em;
  color: var(--ns-muted);
}

.ns-board__meta dd {
  display: flex;
  align-items: center;
  gap: 6px;
  margin: 0;
  font-size: 14px;
  font-weight: 650;
  line-height: 20px;
  color: var(--ns-text);
}

.ns-board__meta dd.is-clock {
  font-variant-numeric: tabular-nums;
  font-family: var(--g-font-family-mono);
  color: var(--ns-accent);
}

.ns-board__orb {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--ns-orb);
}

.ns-board__meta dd.is-running {
  color: var(--ns-ok);
}

.ns-board__meta dd.is-running .ns-board__orb {
  background: var(--ns-ok);
  box-shadow: 0 0 10px var(--ns-ok);
  animation: ns-board-pulse 1.8s ease-out infinite;
}

.ns-board__meta dd.is-loading {
  color: var(--ns-info);
}

.ns-board__meta dd.is-loading .ns-board__orb {
  background: var(--ns-info);
}

.ns-board__meta dd.is-down {
  color: var(--ns-warn);
}

.ns-board__meta dd.is-down .ns-board__orb {
  background: var(--ns-warn);
}

.ns-board__body {
  display: grid;
  grid-template-columns: 196px minmax(0, 1fr);
  min-width: 0;
  padding: 0 16px 14px;
}

.ns-board__dial-card {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  min-height: 188px;
  margin-right: 12px;
  border: 1px solid var(--ns-line);
  background: var(--ns-card);
  box-shadow: inset 0 0 24px var(--ns-inset);
}

.ns-board__ring {
  width: 132px;
  height: 132px;
  transform: rotate(-90deg);
}

.ns-board__ring-track,
.ns-board__ring-value {
  fill: none;
  stroke-width: 8;
}

.ns-board__ring-track {
  stroke: var(--ns-track);
}

.ns-board__ring-value {
  stroke: url(#nsRingGrad);
  stroke-linecap: round;
  transition: stroke-dashoffset 0.6s ease;
}

.ns-board__ring-value.is-warn {
  stroke: var(--ns-warn);
}

.ns-board__ring-value.is-bad {
  stroke: var(--ns-bad);
}

.ns-board__dial-read {
  position: absolute;
  top: 42px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.ns-board__dial-read strong {
  font-size: 28px;
  font-weight: 700;
  line-height: 1;
  letter-spacing: -0.04em;
  font-variant-numeric: tabular-nums;
  color: var(--ns-text);
}

.ns-board__dial-read span,
.ns-board__dial-sub {
  font-size: 11px;
  letter-spacing: 0.08em;
  color: var(--ns-muted);
}

.ns-board__dial-sub {
  margin: 0;
}

.ns-board__panel {
  display: flex;
  flex-direction: column;
  min-width: 0;
  gap: 10px;
}

.ns-board__kpis {
  display: grid;
  grid-template-columns: repeat(7, minmax(0, 1fr));
  gap: 10px;
  flex: 1 1 auto;
}

.ns-board__kpi {
  position: relative;
  display: flex;
  flex-direction: column;
  justify-content: flex-end;
  gap: 8px;
  min-width: 0;
  min-height: 118px;
  padding: 14px 12px 12px;
  border: 1px solid var(--ns-line);
  background: var(--ns-card);
  box-shadow: inset 0 1px 0 color-mix(in srgb, var(--g-primary) 16%, transparent);
}

.ns-board__kpi::before {
  content: '';
  position: absolute;
  top: 0;
  right: 18px;
  left: 18px;
  height: 2px;
  background: linear-gradient(90deg, transparent, var(--ns-accent), transparent);
}

.ns-board__kpi span {
  font-size: 12px;
  color: var(--ns-muted);
}

.ns-board__kpi strong {
  font-size: 30px;
  font-weight: 700;
  line-height: 34px;
  letter-spacing: -0.04em;
  font-variant-numeric: tabular-nums;
  color: var(--ns-text);
}

.ns-board__kpi.is-warn strong {
  color: var(--ns-warn);
}

.ns-board__kpi.is-bad strong {
  color: var(--ns-bad);
}

.ns-board__bar {
  display: block;
  height: 3px;
  overflow: hidden;
  background: var(--ns-bar);
}

.ns-board__bar b {
  display: block;
  height: 100%;
  background: linear-gradient(90deg, var(--ns-info), var(--ns-accent));
}

.ns-board__kpi.is-warn .ns-board__bar b {
  background: var(--ns-warn);
}

.ns-board__kpi.is-bad .ns-board__bar b {
  background: var(--ns-bad);
}

.ns-board__foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 8px 4px 0;
  font-size: 12px;
  color: var(--ns-muted);
}

.ns-board__foot span {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

@keyframes ns-board-pulse {
  0% { box-shadow: 0 0 0 0 color-mix(in srgb, var(--g-success) 40%, transparent); }
  100% { box-shadow: 0 0 0 10px color-mix(in srgb, var(--g-success) 0%, transparent); }
}

@keyframes ns-board-scan {
  0% { transform: translateY(-40%); }
  100% { transform: translateY(320%); }
}

[data-theme='dark'] .ns-board {
  --ns-glow-a: color-mix(in srgb, var(--g-primary) 22%, transparent);
  --ns-glow-b: color-mix(in srgb, var(--g-info) 18%, transparent);
  --ns-grid: color-mix(in srgb, var(--g-primary) 16%, transparent);
  --ns-scan: color-mix(in srgb, var(--g-primary) 14%, transparent);
  --ns-inset: color-mix(in srgb, var(--g-primary) 10%, transparent);
  --ns-line: color-mix(in srgb, var(--g-primary) 28%, var(--g-border-primary));
  --ns-card: color-mix(in srgb, var(--g-bg-primary) 82%, var(--g-bg-tertiary));
}

@media (prefers-reduced-motion: reduce) {
  .ns-board__scan,
  .ns-board__meta dd.is-running .ns-board__orb {
    animation: none;
  }
}

@media (max-width: 1280px) {
  .ns-board__body {
    grid-template-columns: 1fr;
  }

  .ns-board__dial-card {
    margin: 0 0 10px;
    min-height: 0;
    flex-direction: row;
    justify-content: flex-start;
    padding: 10px 16px;
  }

  .ns-board__dial-read {
    position: static;
    margin-left: 8px;
    align-items: flex-start;
  }

  .ns-board__kpis {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}

@media (max-width: 960px) {
  .ns-board__head,
  .ns-board__meta {
    flex-direction: column;
    align-items: flex-start;
  }

  .ns-board__kpis {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .ns-board__identity h2 {
    font-size: 22px;
  }
}
</style>
