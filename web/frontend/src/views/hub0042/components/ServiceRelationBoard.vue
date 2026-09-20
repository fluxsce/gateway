<template>
  <section class="rel-board" :class="{ 'is-live': hasRelation }">
    <div class="rel-board__glow" aria-hidden="true" />
    <div class="rel-board__grid" aria-hidden="true" />
    <div class="rel-board__scan" aria-hidden="true" />

    <header class="rel-board__head">
      <div class="rel-board__identity">
        <p class="rel-board__kicker">
          <span>{{ t('hub0042.relationKicker') }}</span>
          <i />
          <span>{{ t('hub0042.relationTitle') }}</span>
        </p>
      </div>
      <div class="rel-board__meters">
        <article class="is-up">
          <span>{{ t('hub0042.subscribedCount') }}</span>
          <strong>{{ upstreams.length }}</strong>
        </article>
        <article class="is-down">
          <span>{{ t('hub0042.subscriberCount') }}</span>
          <strong>{{ downstreams.length }}</strong>
        </article>
      </div>
    </header>

    <div class="rel-board__viewport">
      <div class="rel-board__stage" ref="viewportRef">
        <svg class="rel-board__svg" :viewBox="`0 0 ${size.w} ${size.h}`" preserveAspectRatio="none">
          <defs>
            <marker :id="ids.up" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="8" markerHeight="8" orient="auto">
              <path d="M 0 1 L 10 5 L 0 9 z" class="rel-board__mark is-up" />
            </marker>
            <marker :id="ids.down" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="8" markerHeight="8" orient="auto">
              <path d="M 0 1 L 10 5 L 0 9 z" class="rel-board__mark is-down" />
            </marker>
          </defs>

          <path
            v-for="edge in edges"
            :id="edge.id"
            :key="edge.id"
            :d="edge.d"
            class="rel-board__edge"
            :class="['is-' + edge.side, { 'is-ghost': edge.ghost, 'is-on': isEdgeActive(edge), 'is-off': isEdgeDim(edge) }]"
            :marker-end="`url(#${edge.side === 'up' ? ids.up : ids.down})`"
          />
          <circle
            v-for="edge in flowEdges"
            :key="'dot-' + edge.id"
            r="3"
            class="rel-board__dot"
            :class="'is-' + edge.side"
          >
            <animateMotion :dur="edge.ghost ? '4.8s' : '2.4s'" repeatCount="indefinite">
              <mpath :href="'#' + edge.id" />
            </animateMotion>
          </circle>
        </svg>

        <span
          v-for="edge in edges"
          :key="'cap-' + edge.id"
          class="rel-board__cap"
          :class="'is-' + edge.side"
          :style="{ transform: `translate(${edge.lx}px, ${edge.ly}px) translate(-50%, -50%)` }"
        >
          {{ t('hub0042.edgeSubscribe') }}
        </span>

        <span class="rel-board__lane is-up">{{ t('hub0042.subscribedServices') }}</span>
        <span class="rel-board__lane is-down">{{ t('hub0042.subscriberServices') }}</span>

        <button
          v-for="node in graphNodes"
          :key="node.id"
          type="button"
          class="rel-board__node"
          :class="['is-' + node.side, { 'is-ghost': node.ghost, 'is-on': isNodeActive(node.id), 'is-off': isNodeDim(node.id), 'is-clickable': !node.ghost && node.side !== 'hub' }]"
          :style="{
            width: node.w + 'px',
            height: node.h + 'px',
            transform: `translate(${node.x - node.w / 2}px, ${node.y - node.h / 2}px)`,
          }"
          :title="node.hint"
          @mouseenter="hoverId = node.ghost ? '' : node.id"
          @mouseleave="hoverId = ''"
          @focus="hoverId = node.ghost ? '' : node.id"
          @blur="hoverId = ''"
          @click="openInspect(node)"
        >
          <b aria-hidden="true" />
          <strong>{{ node.title }}</strong>
          <em v-if="node.sub">{{ node.sub }}</em>
        </button>
      </div>
    </div>

    <RsDialog
      :open="inspectOpen"
      :title="inspectTitle"
      layout="window"
      :width="920"
      @update:open="inspectOpen = $event"
    >
      <template #body>
        <div class="rel-inspect">
          <div class="rel-inspect__meta">
            <span>{{ inspectPeer?.title }}</span>
            <span>{{ inspectPeer?.groupName || '-' }}</span>
            <span>{{ inspectPeer?.namespaceId || '-' }}</span>
            <span>{{ inspectPeer ? scopeLabel(inspectPeer.scope) : '-' }}</span>
          </div>

          <section>
            <header>{{ inspectNodeTitle }}</header>
            <p class="rel-inspect__hint">{{ inspectNodeHint }}</p>
            <p v-if="inspectLoading">{{ t('hub0042.loading') }}</p>
            <p v-else-if="!inspectNodes.length">{{ t('hub0042.noRelationNodes') }}</p>
            <RsGrid
              v-else
              module-id="hub0042:relation-node"
              :data="inspectNodes"
              :columns="inspectNodeColumns"
              :selectable="false"
              :show-index="true"
              row-key="nodeId"
              height="240px"
            />
          </section>

          <section>
            <header>{{ inspectLinkTitle }}</header>
            <p class="rel-inspect__hint">{{ inspectLinkHint }}</p>
            <p v-if="!inspectConnections.length">{{ t('hub0042.noRelationLinks') }}</p>
            <RsGrid
              v-else
              module-id="hub0042:relation-link"
              :data="inspectConnections"
              :columns="inspectLinkColumns"
              :selectable="false"
              :show-index="true"
              row-key="subscriberKey"
              height="200px"
            />
          </section>
        </div>
      </template>
      <template #footer>
        <RsButton variant="primary" @click="inspectOpen = false">{{ t('hub0042.close') }}</RsButton>
      </template>
    </RsDialog>
  </section>
</template>

<script lang="ts" setup>
import type { RsGridColumn } from '@/components/rs-grid'
import { RsGrid } from '@/components/rs-grid'
import { RsButton, RsDialog, RsTag } from '@/ui'
import { isApiSuccess } from '@/utils/format'
import { computed, h, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { getService } from '../api'
import type { Service, ServiceNode, ServiceSubscriber } from '../types'

defineOptions({
  name: 'ServiceRelationBoard',
})

const props = defineProps<{
  service: Service
}>()

const { t } = useI18n()

interface GraphNode {
  id: string
  side: 'up' | 'hub' | 'down'
  title: string
  sub: string
  hint: string
  ghost: boolean
  namespaceId?: string
  groupName?: string
  serviceName?: string
  scope?: string
  items: ServiceSubscriber[]
  x: number
  y: number
  w: number
  h: number
}

interface GraphEdge {
  id: string
  side: 'up' | 'down'
  from: string
  to: string
  d: string
  ghost: boolean
  lx: number
  ly: number
}

const scopeLabel = (scope?: string) => {
  if (scope === 'namespace') return t('hub0042.scopeNamespace')
  if (scope === 'group') return t('hub0042.scopeGroup')
  return t('hub0042.scopeService')
}

const serviceTitle = (item: ServiceSubscriber) => {
  if (item.serviceName) return item.serviceName
  if (item.scope === 'namespace') return item.namespaceId || t('hub0042.scopeNamespace')
  return item.groupName || t('hub0042.scopeGroup')
}

const groupItems = (items: ServiceSubscriber[], prefix: string) => {
  const map = new Map<string, ServiceSubscriber[]>()
  for (const item of items) {
    const key = [prefix, item.namespaceId, item.groupName, item.serviceName, item.scope].join(':')
    const list = map.get(key) || []
    list.push(item)
    map.set(key, list)
  }
  return [...map.entries()].map(([id, list]) => {
    const first = list[0]
    const extra = list.length > 1 ? ` · ${list.length}` : ''
    return {
      id,
      title: serviceTitle(first),
      sub: `${first.groupName || first.namespaceId || scopeLabel(first.scope)}${extra}`,
      hint: list.map((item) => [serviceTitle(item), item.groupName, item.namespaceId].filter(Boolean).join(' / ')).join('\n'),
      namespaceId: first.namespaceId,
      groupName: first.groupName,
      serviceName: first.serviceName,
      scope: first.scope,
      items: list,
    }
  })
}

const upstreams = computed(() => groupItems(props.service.subscriptions || [], 'up'))
const downstreams = computed(() => groupItems(props.service.subscribers || [], 'down'))
const hasRelation = computed(() => upstreams.value.length > 0 || downstreams.value.length > 0)

const viewportRef = ref<HTMLElement>()
const size = reactive({ w: 800, h: 260 })
const hoverId = ref('')
const ids = {
  up: 'hub0042-rel-arrow-up',
  down: 'hub0042-rel-arrow-down',
  upGrad: 'hub0042-rel-grad-up',
  downGrad: 'hub0042-rel-grad-down',
  glow: 'hub0042-rel-glow',
}

const NODE_W = 168
const NODE_H = 58
const HUB_W = 188
const HUB_H = 72
const PAD_X = 28
const PAD_Y = 40

const nodeHeight = computed(() => {
  const rows = Math.max(upstreams.value.length, downstreams.value.length, 1)
  const usable = size.h - PAD_Y * 2
  return Math.max(44, Math.min(NODE_H, Math.floor((usable - (rows - 1) * 10) / rows)))
})

const stackY = (index: number, total: number, height: number, nodeH: number) => {
  if (total <= 1) return height / 2
  const usable = height - PAD_Y * 2 - nodeH
  return PAD_Y + nodeH / 2 + (usable * index) / (total - 1)
}

const curve = (x1: number, y1: number, x2: number, y2: number) => {
  const dx = Math.max(56, Math.abs(x2 - x1) * 0.44)
  return `M ${x1} ${y1} C ${x1 + dx} ${y1}, ${x2 - dx} ${y2}, ${x2} ${y2}`
}

const graph = computed(() => {
  const w = size.w
  const h = size.h
  const sideH = nodeHeight.value
  const hubH = Math.min(HUB_H, Math.max(56, sideH + 10))
  const hub: GraphNode = {
    id: 'hub',
    side: 'hub',
    title: props.service.serviceName,
    sub: t('hub0042.analysisTarget'),
    hint: props.service.serviceName,
    ghost: false,
    items: [],
    x: w / 2,
    y: h / 2,
    w: HUB_W,
    h: hubH,
  }

  const left: GraphNode[] = upstreams.value.length
    ? upstreams.value.map((item, index) => ({
        id: item.id,
        side: 'up' as const,
        title: item.title,
        sub: item.sub,
        hint: item.hint,
        ghost: false,
        namespaceId: item.namespaceId,
        groupName: item.groupName,
        serviceName: item.serviceName,
        scope: item.scope,
        items: item.items,
        x: PAD_X + NODE_W / 2,
        y: stackY(index, upstreams.value.length, h, sideH),
        w: NODE_W,
        h: sideH,
      }))
    : [{
        id: 'up-empty',
        side: 'up',
        title: t('hub0042.noSubscribedService'),
        sub: '0',
        hint: t('hub0042.noSubscribedService'),
        ghost: true,
        items: [],
        x: PAD_X + NODE_W / 2,
        y: h / 2,
        w: NODE_W,
        h: sideH,
      }]

  const right: GraphNode[] = downstreams.value.length
    ? downstreams.value.map((item, index) => ({
        id: item.id,
        side: 'down' as const,
        title: item.title,
        sub: item.sub,
        hint: item.hint,
        ghost: false,
        namespaceId: item.namespaceId,
        groupName: item.groupName,
        serviceName: item.serviceName,
        scope: item.scope,
        items: item.items,
        x: w - PAD_X - NODE_W / 2,
        y: stackY(index, downstreams.value.length, h, sideH),
        w: NODE_W,
        h: sideH,
      }))
    : [{
        id: 'down-empty',
        side: 'down',
        title: t('hub0042.noSubscriberService'),
        sub: '0',
        hint: t('hub0042.noSubscriberService'),
        ghost: true,
        items: [],
        x: w - PAD_X - NODE_W / 2,
        y: h / 2,
        w: NODE_W,
        h: sideH,
      }]

  const nodes = [...left, hub, ...right]
  const makeEdge = (node: GraphNode, side: 'up' | 'down'): GraphEdge => {
    const x1 = side === 'up' ? node.x + node.w / 2 : hub.x + hub.w / 2
    const y1 = side === 'up' ? node.y : hub.y
    const x2 = side === 'up' ? hub.x - hub.w / 2 : node.x - node.w / 2
    const y2 = side === 'up' ? hub.y : node.y
    return {
      id: `e-${node.id}`,
      side,
      from: side === 'up' ? node.id : hub.id,
      to: side === 'up' ? hub.id : node.id,
      ghost: node.ghost,
      d: curve(x1, y1, x2, y2),
      lx: (x1 + x2) / 2,
      ly: (y1 + y2) / 2 - 12,
    }
  }
  const edges: GraphEdge[] = [
    ...left.map((node) => makeEdge(node, 'up')),
    ...right.map((node) => makeEdge(node, 'down')),
  ]
  return { nodes, edges }
})

const graphNodes = computed(() => graph.value.nodes)
const edges = computed(() => graph.value.edges)
const flowEdges = computed(() => edges.value)
const focusId = computed(() => hoverId.value)

const inspectOpen = ref(false)
const inspectLoading = ref(false)
const inspectPeer = ref<GraphNode | null>(null)
const inspectNodes = ref<ServiceNode[]>([])

const inspectTitle = computed(() => {
  const peer = inspectPeer.value
  if (!peer) return t('hub0042.relationNodes')
  const lane = peer.side === 'down' ? t('hub0042.subscriberServices') : t('hub0042.subscribedServices')
  return `${lane} · ${peer.title}`
})

const inspectDown = computed(() => inspectPeer.value?.side === 'down')
const inspectNodeTitle = computed(() => t('hub0042.relationNodes'))
const inspectNodeHint = computed(() =>
  inspectDown.value ? t('hub0042.relationNodeHintDown') : t('hub0042.relationNodeHintUp'),
)
const inspectLinkTitle = computed(() =>
  inspectDown.value ? t('hub0042.relationLinksDown') : t('hub0042.relationLinksUp'),
)
const inspectLinkHint = computed(() =>
  inspectDown.value ? t('hub0042.relationLinkHintDown') : t('hub0042.relationLinkHintUp'),
)

interface InspectConnection extends ServiceSubscriber {
  nodeCount: number
  nodeAddrs: string
}

const instanceStatusMeta = (status?: string) => {
  if (status === 'UP') return { variant: 'success' as const, label: t('hub0042.instanceUp') }
  if (status === 'DOWN') return { variant: 'danger' as const, label: t('hub0042.instanceDown') }
  if (status === 'STARTING') return { variant: 'warning' as const, label: t('hub0042.instanceStarting') }
  if (status === 'OUT_OF_SERVICE') return { variant: 'info' as const, label: t('hub0042.instanceOut') }
  return { variant: 'default' as const, label: status || '-' }
}

const healthStatusMeta = (status?: string) => {
  if (status === 'HEALTHY') return { variant: 'success' as const, label: t('hub0042.healthHealthy') }
  if (status === 'UNHEALTHY') return { variant: 'danger' as const, label: t('hub0042.healthUnhealthy') }
  if (status === 'UNKNOWN') return { variant: 'warning' as const, label: t('hub0042.healthUnknown') }
  return { variant: 'default' as const, label: status || '-' }
}

const statusTag = (meta: { variant: 'success' | 'danger' | 'warning' | 'info' | 'default'; label: string }) =>
  h(RsTag, { variant: meta.variant, size: 'sm' }, () => meta.label)

const inspectConnections = computed<InspectConnection[]>(() => {
  const links = inspectPeer.value?.items || []
  const pool = inspectPeer.value?.side === 'up' ? props.service.nodes || [] : inspectNodes.value
  return links.map((link, index) => {
    const nodes = pool.filter((node) => node.connectionId && node.connectionId === link.connectionId)
    return {
      ...link,
      subscriberKey: link.subscriberKey || `${link.connectionId || 'conn'}:${index}`,
      nodeCount: nodes.length,
      nodeAddrs: nodes.map((node) => `${node.ipAddress}:${node.portNumber}`).join(' / ') || '-',
    }
  })
})

const boundConnectionIds = computed(() =>
  new Set(inspectConnections.value.map((item) => item.connectionId).filter(Boolean)),
)

const inspectNodeColumns = computed<RsGridColumn<ServiceNode>[]>(() => {
  const columns: RsGridColumn<ServiceNode>[] = [
    { key: 'nodeId', title: t('hub0042.nodeId'), align: 'center', ellipsis: true, formatter: (v) => String(v || '-') },
    { key: 'ipAddress', title: t('hub0042.nodeIp'), align: 'center', formatter: (v) => String(v || '-') },
    { key: 'portNumber', title: t('hub0042.nodePort'), align: 'center', formatter: (v) => String(v ?? '-') },
    { key: 'connectionId', title: t('hub0042.registerConnection'), align: 'center', ellipsis: true, formatter: (v) => String(v || '-') },
    {
      key: 'instanceStatus',
      title: t('hub0042.nodeStatus'),
      align: 'center',
      render: (row) => statusTag(instanceStatusMeta(row.instanceStatus)),
    },
    {
      key: 'healthyStatus',
      title: t('hub0042.healthStatus'),
      align: 'center',
      render: (row) => statusTag(healthStatusMeta(row.healthyStatus)),
    },
    {
      key: 'ephemeral',
      title: t('hub0042.ephemeralNode'),
      align: 'center',
      render: (row) =>
        statusTag({
          variant: row.ephemeral === 'Y' ? 'warning' : 'default',
          label: row.ephemeral === 'Y' ? t('hub0042.lifecycleEphemeral') : t('hub0042.lifecyclePersistent'),
        }),
    },
  ]
  if (inspectPeer.value?.side === 'down') {
    columns.push({
      key: 'bound',
      title: t('hub0042.boundSubscribe'),
      align: 'center',
      render: (row) => {
        const bound = !!row.connectionId && boundConnectionIds.value.has(row.connectionId)
        return statusTag({
          variant: bound ? 'info' : 'default',
          label: bound ? t('hub0042.boundSubscribe') : '-',
        })
      },
    })
  }
  columns.push({
    key: 'lastBeatTime',
    title: t('hub0042.lastBeat'),
    align: 'center',
    formatter: (v) => String(v || '-'),
  })
  return columns
})

const inspectLinkColumns: RsGridColumn<InspectConnection>[] = [
  { key: 'connectionId', title: t('hub0042.connectionId'), align: 'center', ellipsis: true, formatter: (v) => String(v || '-') },
  { key: 'clientId', title: t('hub0042.clientId'), align: 'center', ellipsis: true, formatter: (v) => String(v || '-') },
  { key: 'clientIp', title: t('hub0042.clientIp'), align: 'center', formatter: (v) => String(v || '-') },
  {
    key: 'nodeAddrs',
    title: t('hub0042.relatedNodes'),
    align: 'left',
    render: (row) =>
      h('span', { class: 'rel-inspect__nodes' }, [
        statusTag({
          variant: row.nodeCount > 0 ? 'success' : 'warning',
          label: String(row.nodeCount),
        }),
        h('em', row.nodeAddrs),
      ]),
  },
  { key: 'lastActive', title: t('hub0042.lastActive'), align: 'center', formatter: (v) => String(v || '-') },
]

const openInspect = async (node: GraphNode) => {
  if (node.ghost || node.side === 'hub') return
  inspectPeer.value = node
  inspectNodes.value = []
  inspectOpen.value = true
  if (!node.namespaceId || !node.groupName || !node.serviceName) return
  inspectLoading.value = true
  try {
    const response = await getService(node.namespaceId, node.groupName, node.serviceName)
    if (isApiSuccess(response) && response.bizData) {
      const detail = JSON.parse(response.bizData) as Service
      inspectNodes.value = detail.nodes || []
    }
  } catch {
    inspectNodes.value = []
  } finally {
    inspectLoading.value = false
  }
}

const isNodeActive = (id: string) => !!focusId.value && (id === focusId.value || id === 'hub')
const isNodeDim = (id: string) => !!focusId.value && focusId.value !== 'hub' && id !== focusId.value && id !== 'hub'
const isEdgeActive = (edge: GraphEdge) => !!focusId.value && (focusId.value === 'hub' || edge.from === focusId.value || edge.to === focusId.value)
const isEdgeDim = (edge: GraphEdge) => !!focusId.value && focusId.value !== 'hub' && edge.from !== focusId.value && edge.to !== focusId.value

let observer: ResizeObserver | undefined

const syncSize = () => {
  const el = viewportRef.value
  size.w = Math.max(1, Math.floor(el?.clientWidth || 800))
  size.h = Math.max(1, Math.floor(el?.clientHeight || 260))
}

const bindViewport = () => {
  observer?.disconnect()
  observer = undefined
  if (!viewportRef.value || typeof ResizeObserver === 'undefined') {
    syncSize()
    return
  }
  observer = new ResizeObserver(() => syncSize())
  observer.observe(viewportRef.value)
  syncSize()
}

watch([upstreams, downstreams], async () => {
  await nextTick()
  syncSize()
})

onMounted(async () => {
  await nextTick()
  bindViewport()
})

onBeforeUnmount(() => {
  observer?.disconnect()
})
</script>

<style lang="scss" scoped>
.rel-board {
  --rel-accent: var(--g-primary);
  --rel-info: var(--g-info);
  --rel-warn: var(--g-warning);
  --rel-text: var(--g-text-primary, var(--g-text));
  --rel-muted: var(--g-text-tertiary, var(--g-text-secondary));
  --rel-line: color-mix(in srgb, var(--g-primary) 22%, var(--g-border-primary, var(--g-border)));
  --rel-card: color-mix(in srgb, var(--g-bg-primary, var(--g-bg)) 82%, transparent);
  --rel-glow-a: color-mix(in srgb, var(--g-warning) 16%, transparent);
  --rel-glow-b: color-mix(in srgb, var(--g-info) 14%, transparent);
  --rel-grid: color-mix(in srgb, var(--g-primary) 10%, transparent);
  --rel-scan: color-mix(in srgb, var(--g-primary) 10%, transparent);

  position: relative;
  isolation: isolate;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
  gap: 0;
  color: var(--rel-text);
  border: 1px solid var(--rel-line);
  border-radius: var(--g-radius-md, 8px);
  background:
    radial-gradient(900px 180px at 8% 0%, var(--rel-glow-a), transparent 55%),
    radial-gradient(800px 200px at 96% 100%, var(--rel-glow-b), transparent 50%),
    linear-gradient(180deg, var(--g-bg-secondary, var(--g-bg)) 0%, var(--g-bg-tertiary, var(--g-bg-elevated, var(--g-bg))) 100%);
  box-shadow: inset 0 1px 0 color-mix(in srgb, var(--g-primary) 20%, transparent);
}

.rel-board::before,
.rel-board::after {
  content: '';
  position: absolute;
  width: 16px;
  height: 16px;
  z-index: 3;
  border: 1px solid color-mix(in srgb, var(--g-primary) 48%, transparent);
  pointer-events: none;
}

.rel-board::before {
  top: 8px;
  left: 8px;
  border-right: none;
  border-bottom: none;
}

.rel-board::after {
  right: 8px;
  bottom: 8px;
  border-left: none;
  border-top: none;
}

.rel-board__glow,
.rel-board__grid,
.rel-board__scan {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.rel-board__grid {
  background-image:
    linear-gradient(var(--rel-grid) 1px, transparent 1px),
    linear-gradient(90deg, var(--rel-grid) 1px, transparent 1px);
  background-size: 28px 28px;
  mask-image: radial-gradient(circle at 50% 55%, #000 40%, transparent 88%);
}

.rel-board__scan {
  height: 26%;
  background: linear-gradient(180deg, transparent, var(--rel-scan), transparent);
  animation: rel-board-scan 5.5s linear infinite;
}

.rel-board__head,
.rel-board__viewport {
  position: relative;
  z-index: 1;
}

.rel-board__head {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  flex-shrink: 0;
  gap: 16px;
  padding: 14px 20px 8px;
}

.rel-board__kicker {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 0;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.16em;
  color: var(--rel-accent);

  i {
    display: block;
    width: 36px;
    height: 1px;
    background: linear-gradient(90deg, var(--rel-accent), transparent);
  }
}

.rel-board__meters {
  display: flex;
  gap: 10px;

  article {
    min-width: 88px;
    padding: 8px 12px;
    border: 1px solid var(--rel-line);
    background: var(--rel-card);
    box-shadow: inset 0 1px 0 color-mix(in srgb, var(--g-primary) 16%, transparent);
  }

  span {
    display: block;
    color: var(--rel-muted);
    font-size: 11px;
    letter-spacing: 0.08em;
  }

  strong {
    display: block;
    margin-top: 2px;
    font-size: 24px;
    line-height: 1;
    font-variant-numeric: tabular-nums;
    text-shadow: 0 0 16px color-mix(in srgb, var(--g-primary) 28%, transparent);
  }

  .is-up strong { color: var(--rel-warn); text-shadow: 0 0 16px color-mix(in srgb, var(--g-warning) 32%, transparent); }
  .is-down strong { color: var(--rel-info); text-shadow: 0 0 16px color-mix(in srgb, var(--g-info) 32%, transparent); }
}

.rel-board__viewport {
  flex: 1 1 0;
  min-height: 0;
  overflow: hidden;
  padding: 0 8px 12px;
}

.rel-board__stage {
  position: relative;
  width: 100%;
  height: 100%;
}

.rel-board__svg {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
}

.rel-board__orbit {
  fill: none;
  stroke: color-mix(in srgb, var(--g-primary) 22%, transparent);
  stroke-width: 1;
  stroke-dasharray: 3 7;
  animation: rel-board-dash 16s linear infinite;

  &.is-inner {
    stroke: color-mix(in srgb, var(--g-primary) 34%, transparent);
    animation-duration: 10s;
    animation-direction: reverse;
  }
}

.rel-board__mark {
  &.is-up { fill: var(--g-warning); }
  &.is-down { fill: var(--g-info); }
}

.rel-board__edge {
  fill: none;
  stroke-width: 2.2;
  opacity: 0.92;

  &.is-up { stroke: var(--g-warning); }
  &.is-down { stroke: var(--g-info); }
  &.is-ghost { stroke-dasharray: 7 6; opacity: 0.55; }
  &.is-on { stroke-width: 2.8; opacity: 1; }
  &.is-off { opacity: 0.16; }
}

.rel-board__cap {
  position: absolute;
  left: 0;
  top: 0;
  padding: 1px 6px;
  border: 1px solid var(--rel-line);
  border-radius: 999px;
  background: var(--rel-card);
  color: var(--rel-muted);
  font-size: 10px;
  letter-spacing: 0.08em;
  pointer-events: none;

  &.is-up { color: var(--rel-warn); }
  &.is-down { color: var(--rel-info); }
}

.rel-board__dot {
  fill: var(--g-primary);

  &.is-up { fill: var(--g-warning); }
  &.is-down { fill: var(--g-info); }
}

.rel-board__lane {
  position: absolute;
  top: 6px;
  color: var(--rel-muted);
  font-size: 11px;
  letter-spacing: 0.12em;

  &.is-up { left: 28px; }
  &.is-down { right: 28px; }
}

.rel-board__node {
  position: absolute;
  left: 0;
  top: 0;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 3px;
  padding: 10px 14px;
  border: 1px solid var(--rel-line);
  background: var(--rel-card);
  color: var(--rel-text);
  text-align: left;
  box-shadow:
    inset 0 1px 0 color-mix(in srgb, var(--g-primary) 16%, transparent),
    0 10px 24px color-mix(in srgb, var(--g-text) 6%, transparent);

  b {
    position: absolute;
    top: 0;
    right: 18px;
    left: 18px;
    height: 2px;
    background: linear-gradient(90deg, transparent, var(--rel-accent), transparent);
  }

  strong,
  em {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  strong { font-size: 14px; font-weight: 650; }
  em {
    color: var(--rel-muted);
    font-size: 11px;
    font-style: normal;
  }

  &.is-up b { background: linear-gradient(90deg, transparent, var(--rel-warn), transparent); }
  &.is-down b { background: linear-gradient(90deg, transparent, var(--rel-info), transparent); }
  &.is-hub {
    align-items: center;
    text-align: center;
    border-color: color-mix(in srgb, var(--g-primary) 48%, var(--rel-line));
    background:
      radial-gradient(120% 80% at 50% 0%, color-mix(in srgb, var(--g-primary) 18%, transparent), transparent 70%),
      var(--rel-card);
    box-shadow:
      0 0 0 5px color-mix(in srgb, var(--g-primary) 10%, transparent),
      0 0 28px color-mix(in srgb, var(--g-primary) 18%, transparent);

    strong {
      font-size: 16px;
      text-shadow: 0 0 16px color-mix(in srgb, var(--g-primary) 28%, transparent);
    }
  }
  &.is-ghost {
    border-style: dashed;
    background: color-mix(in srgb, var(--rel-card) 70%, transparent);
    box-shadow: none;
    color: var(--rel-muted);
  }
  &.is-on { z-index: 2; }
  &.is-off { opacity: 0.32; }
  &.is-clickable { cursor: pointer; }
}

.rel-inspect {
  display: flex;
  flex-direction: column;
  gap: var(--g-space-md);
}

.rel-inspect__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  color: var(--g-text-secondary, var(--rel-muted));
  font-size: 12px;
}

.rel-inspect section header {
  margin-bottom: 8px;
  color: var(--g-text, var(--rel-text));
  font-size: 13px;
  font-weight: 600;
}

.rel-inspect p {
  margin: 0;
  color: var(--g-text-secondary, var(--rel-muted));
  font-size: 12px;
}

.rel-inspect__hint {
  margin-bottom: 8px;
}

.rel-inspect__nodes {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-width: 0;

  em {
    overflow: hidden;
    color: var(--g-text-secondary, var(--rel-muted));
    font-style: normal;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

@keyframes rel-board-scan {
  0% { transform: translateY(-40%); }
  100% { transform: translateY(320%); }
}

@keyframes rel-board-dash {
  to { stroke-dashoffset: -80; }
}

@media (prefers-reduced-motion: reduce) {
  .rel-board__scan,
  .rel-board__orbit,
  .rel-board__dot {
    animation: none;
  }
}
</style>
