<template>
  <RsDialog
    v-model:open="visible"
    :title="title"
    layout="window"
    :width="980"
    :height="680"
    :show-overlay="true"
    :close-on-overlay-click="true"
  >
    <template #header>
      <div class="topo__titlebar">
        <h2>{{ title }}</h2>
        <RsButton variant="ghost" size="sm" :loading="loading" @pointerdown.stop @click="load">
          刷新
        </RsButton>
      </div>
    </template>
    <template #body>
      <div class="topo">
        <RsLoading :loading="loading" overlay block size="lg" />

        <div v-if="placed.length" class="topo__stage">
          <svg class="topo__links" viewBox="0 0 100 100" preserveAspectRatio="none">
            <line
              v-for="node in placed"
              :key="`${node.nodeId}-glow`"
              x1="50"
              y1="50"
              :x2="node.x"
              :y2="node.y"
              class="topo__link-glow"
              :data-tone="node.tone"
              vector-effect="non-scaling-stroke"
            />
            <line
              v-for="node in placed"
              :key="`${node.nodeId}-flow`"
              x1="50"
              y1="50"
              :x2="node.x"
              :y2="node.y"
              class="topo__link-flow"
              :data-tone="node.tone"
              vector-effect="non-scaling-stroke"
            />
          </svg>

          <div class="topo__hub" :data-tone="overallTone">
            <span>当前实例</span>
            <strong>{{ instanceName || '网关实例' }}</strong>
          </div>

          <article
            v-for="node in placed"
            :key="node.nodeId"
            class="topo__node"
            :data-tone="node.tone"
            :style="{ left: `${node.x}%`, top: `${node.y}%` }"
          >
            <header>
              <i class="topo__dot" />
              <strong>{{ node.hostname }}</strong>
              <em>{{ probeLabel(node.status) }}</em>
            </header>
            <dl>
              <div>
                <dt>节点</dt>
                <dd class="topo__nid" :title="node.nodeId">{{ node.nodeId }}</dd>
              </div>
              <div>
                <dt>地址</dt>
                <dd>{{ node.nodeIp }}</dd>
              </div>
            </dl>
            <code>{{ node.endpoint }}</code>
            <small>{{ node.message }}</small>
          </article>

          <ul class="topo__legend">
            <li data-tone="ok">可达</li>
            <li data-tone="bad">不可达</li>
            <li data-tone="empty">未探测</li>
          </ul>
        </div>
        <RsEmpty v-else-if="!loading" fill description="没有可探测的集群节点" />
      </div>
    </template>
    <template #footer>
      <RsButton variant="secondary" @click="visible = false">关闭</RsButton>
    </template>
  </RsDialog>
</template>

<script setup lang="ts">
import { useAppMessage } from '@/composables/useAppMessage'
import { RsButton, RsDialog, RsEmpty, RsLoading } from '@/ui'
import { getApiMessage, isApiSuccess, parseJsonData } from '@/utils/format'
import { computed, ref, watch } from 'vue'
import { queryClusterTopology } from './api'

defineOptions({ name: 'ClusterTopologyDialog' })

const props = defineProps<{
  /** 请求打到这个模块，按钮权限也按它校验。 */
  moduleId: string
  /** 本模块自己的查询参数，字段名不在共用组件里写死。 */
  params: Record<string, unknown>
  instanceName?: string
}>()

interface ClusterNode {
  nodeId: string
  nodeIp: string
  hostname: string
}

interface ListenProbe {
  nodeId: string
  nodeIp: string
  hostname: string
  scheme: string
  port: number
  status: string
  message: string
}

interface TopologyView {
  nodes: ClusterNode[]
  listeners: ListenProbe[]
}

const visible = defineModel<boolean>('visible', { default: false })
const message = useAppMessage()
const loading = ref(false)
const nodes = ref<ClusterNode[]>([])
const listeners = ref<ListenProbe[]>([])

const title = computed(() =>
  props.instanceName ? `集群端口 · ${props.instanceName}` : '集群端口',
)

const probeCount = computed(() => ({
  ok: listeners.value.filter((item) => item.status === 'ok').length,
  fail: listeners.value.filter((item) => item.status === 'fail').length,
}))

const overallTone = computed(() => {
  if (probeCount.value.fail > 0) return 'bad'
  if (probeCount.value.ok > 0) return 'ok'
  return 'empty'
})

const cards = computed(() =>
  nodes.value.map((node) => {
    const probe = listeners.value.find((item) => item.nodeId === node.nodeId)
    const status = probe?.status || 'skip'
    return {
      nodeId: node.nodeId,
      hostname: node.hostname || node.nodeId,
      nodeIp: probe?.nodeIp || node.nodeIp || '—',
      status,
      tone: status === 'ok' ? 'ok' : status === 'fail' ? 'bad' : 'empty',
      endpoint: probe ? `${probe.scheme}://${probe.nodeIp}:${probe.port}` : '—',
      message: probe?.message || '没有监听口',
    }
  }),
)

const placed = computed(() => {
  const list = cards.value
  const total = list.length
  return list.map((card, index) => {
    if (total === 1) return { ...card, x: 78, y: 50 }
    const angle = (Math.PI * 2 * index) / total - Math.PI / 2
    return {
      ...card,
      x: 50 + 34 * Math.cos(angle),
      y: 50 + 30 * Math.sin(angle),
    }
  })
})

const probeLabel = (status: string) => {
  if (status === 'ok') return '可达'
  if (status === 'fail') return '不可达'
  return '未探测'
}

const load = async () => {
  if (!props.moduleId) return
  loading.value = true
  try {
    const response = await queryClusterTopology(props.moduleId, props.params)
    if (!isApiSuccess(response)) {
      message.error(getApiMessage(response, '查询集群端口失败'))
      return
    }
    const view = parseJsonData<TopologyView>(response, { nodes: [], listeners: [] })
    nodes.value = view.nodes || []
    listeners.value = view.listeners || []
  } catch {
    message.error('查询集群端口失败')
  } finally {
    loading.value = false
  }
}

watch(visible, (open) => {
  if (open) load()
})
</script>

<style scoped>
.topo__titlebar {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
  flex: 1;
}

.topo__titlebar h2 {
  margin: 0;
  min-width: 0;
  flex: 1;
  font-size: 14px;
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.topo {
  display: flex;
  flex-direction: column;
  flex: 1;
  height: 100%;
  min-height: 0;
}

.topo__stage {
  position: relative;
  flex: 1;
  width: 100%;
  height: 100%;
  min-height: 0;
  border-radius: 16px;
  overflow: hidden;
  background-color: #f8fafc;
  background-image:
    radial-gradient(circle at 50% 48%, rgba(34, 197, 94, 0.14), transparent 34%),
    linear-gradient(rgba(148, 163, 184, 0.18) 1px, transparent 1px),
    linear-gradient(90deg, rgba(148, 163, 184, 0.18) 1px, transparent 1px);
  background-size: auto, 32px 32px, 32px 32px;
  border: 1px solid var(--rs-color-border, rgba(15, 23, 42, 0.08));
}

.topo__links {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  z-index: 1;
}

.topo__link-glow {
  stroke: #94a3b8;
  stroke-width: 10;
  stroke-linecap: round;
  stroke-opacity: 0.22;
}

.topo__link-glow[data-tone='ok'] {
  stroke: #22c55e;
  stroke-opacity: 0.28;
}

.topo__link-glow[data-tone='bad'] {
  stroke: #ef4444;
  stroke-opacity: 0.28;
}

.topo__link-flow {
  stroke: #64748b;
  stroke-width: 2;
  stroke-linecap: round;
  stroke-dasharray: 8 12;
  animation: topo-flow 1.1s linear infinite;
}

.topo__link-flow[data-tone='ok'] {
  stroke: #16a34a;
}

.topo__link-flow[data-tone='bad'] {
  stroke: #dc2626;
}

@keyframes topo-flow {
  to {
    stroke-dashoffset: -40;
  }
}

.topo__hub,
.topo__node {
  position: absolute;
  transform: translate(-50%, -50%);
}

.topo__hub {
  left: 50%;
  top: 50%;
  z-index: 2;
  width: 148px;
  padding: 16px 14px;
  text-align: center;
  border-radius: 999px;
  background: var(--rs-color-surface, #fff);
  border: 1px solid rgba(37, 99, 235, 0.25);
  box-shadow: 0 10px 30px rgba(37, 99, 235, 0.12);
}

.topo__hub::after {
  content: '';
  position: absolute;
  inset: -14px;
  border-radius: inherit;
  border: 1px solid rgba(37, 99, 235, 0.28);
  animation: topo-ring 2.6s ease-out infinite;
  pointer-events: none;
}

.topo__hub span {
  display: block;
  font-size: 11px;
  letter-spacing: 0.04em;
  opacity: 0.6;
}

.topo__hub strong {
  display: block;
  margin-top: 4px;
  font-size: 14px;
  line-height: 1.35;
}

.topo__hub[data-tone='ok'] {
  border-color: rgba(22, 163, 74, 0.45);
  box-shadow: 0 10px 28px rgba(22, 163, 74, 0.16);
}

.topo__hub[data-tone='ok']::after {
  border-color: rgba(22, 163, 74, 0.4);
}

.topo__hub[data-tone='bad'] {
  border-color: rgba(220, 38, 38, 0.4);
  box-shadow: 0 10px 28px rgba(220, 38, 38, 0.14);
}

.topo__hub[data-tone='bad']::after {
  border-color: rgba(220, 38, 38, 0.35);
}

@keyframes topo-ring {
  0% {
    transform: scale(0.92);
    opacity: 0.85;
  }
  100% {
    transform: scale(1.18);
    opacity: 0;
  }
}

.topo__node {
  z-index: 2;
  width: 268px;
  padding: 12px 14px 10px;
  border-radius: 16px;
  color: var(--rs-color-text, #0f172a);
  background: rgba(255, 255, 255, 0.94);
  border: 1px solid rgba(15, 23, 42, 0.08);
  box-shadow: 0 16px 36px rgba(15, 23, 42, 0.1);
  backdrop-filter: blur(8px);
}

.topo__node[data-tone='ok'] {
  border-color: rgba(22, 163, 74, 0.28);
  box-shadow:
    0 16px 36px rgba(22, 163, 74, 0.12),
    inset 3px 0 0 #16a34a;
}

.topo__node[data-tone='bad'] {
  border-color: rgba(220, 38, 38, 0.28);
  box-shadow:
    0 16px 36px rgba(220, 38, 38, 0.12),
    inset 3px 0 0 #dc2626;
}

.topo__node header {
  display: flex;
  align-items: center;
  gap: 8px;
}

.topo__dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #94a3b8;
  flex: none;
}

.topo__node[data-tone='ok'] .topo__dot {
  background: #16a34a;
  box-shadow: 0 0 0 4px rgba(22, 163, 74, 0.16);
  animation: topo-dot 2s ease-out infinite;
}

.topo__node[data-tone='bad'] .topo__dot {
  background: #dc2626;
  box-shadow: 0 0 0 4px rgba(220, 38, 38, 0.16);
}

@keyframes topo-dot {
  0% {
    box-shadow: 0 0 0 0 rgba(22, 163, 74, 0.45);
  }
  100% {
    box-shadow: 0 0 0 8px rgba(22, 163, 74, 0);
  }
}

.topo__node header strong {
  min-width: 0;
  flex: 1;
  font-size: 15px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.topo__node header em {
  flex: none;
  padding: 2px 8px;
  border-radius: 999px;
  font-style: normal;
  font-size: 12px;
  font-weight: 600;
  color: #475569;
  background: rgba(148, 163, 184, 0.18);
}

.topo__node[data-tone='ok'] header em {
  color: #15803d;
  background: rgba(22, 163, 74, 0.12);
}

.topo__node[data-tone='bad'] header em {
  color: #b91c1c;
  background: rgba(220, 38, 38, 0.12);
}

.topo__node dl {
  display: grid;
  gap: 6px;
  margin: 10px 0 0;
}

.topo__node dl div {
  display: grid;
  grid-template-columns: 32px minmax(0, 1fr);
  gap: 8px;
  align-items: start;
}

.topo__node dt {
  margin: 0;
  font-size: 11px;
  line-height: 1.45;
  color: #64748b;
}

.topo__node dd {
  margin: 0;
  min-width: 0;
  font-size: 12px;
  line-height: 1.45;
  color: #0f172a;
}

.topo__node .topo__nid {
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 11px;
  word-break: break-all;
}

.topo__node code {
  display: block;
  margin-top: 8px;
  padding: 4px 8px;
  border-radius: 8px;
  font-size: 12px;
  color: #4338ca;
  background: rgba(67, 56, 202, 0.08);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.topo__node small {
  display: block;
  margin-top: 6px;
  font-size: 12px;
  color: #334155;
}

.topo__legend {
  position: absolute;
  left: 14px;
  bottom: 12px;
  display: flex;
  gap: 14px;
  margin: 0;
  padding: 6px 12px;
  list-style: none;
  font-size: 12px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.86);
  border: 1px solid rgba(15, 23, 42, 0.06);
  box-shadow: 0 6px 16px rgba(15, 23, 42, 0.06);
}

.topo__legend li::before {
  content: '';
  display: inline-block;
  width: 8px;
  height: 8px;
  margin-right: 6px;
  border-radius: 50%;
  background: #94a3b8;
}

.topo__legend li[data-tone='ok']::before {
  background: #16a34a;
}

.topo__legend li[data-tone='bad']::before {
  background: #dc2626;
}
</style>
