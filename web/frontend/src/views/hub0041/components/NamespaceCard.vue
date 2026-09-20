<template>
  <div class="ns-card-host">
    <RsContextMenu :items="menuItems" @select="onMenuSelect">
      <article
        class="ns-card"
        :class="['is-' + tone, { 'is-selected': selected }]"
        @click="emit('select')"
        @dblclick="emit('action', 'view')"
        @contextmenu="emit('focus')"
      >
      <header class="ns-card__head">
        <div class="ns-card__icon" :class="'is-' + envTone" aria-hidden="true">
          <GIcon :icon="LayersOutline" :size="22" color="#fff" />
        </div>
        <div class="ns-card__identity">
          <h3 class="ns-card__name" :title="namespace.namespaceName">
            {{ namespace.namespaceName }}
          </h3>
          <div class="ns-card__sub">
            <p class="ns-card__meta" :title="metaText">{{ metaText }}</p>
            <p class="ns-card__live" :class="'is-' + tone">
              <span class="ns-card__orb" aria-hidden="true" />
              <span>{{ liveText }}</span>
            </p>
          </div>
        </div>
      </header>

      <div class="ns-card__stats">
        <div class="ns-card__stat">
          <span>服务</span>
          <strong>{{ serviceText }}</strong>
        </div>
        <div class="ns-card__stat">
          <span>节点</span>
          <strong>{{ metric(namespace.nodeCount) }}</strong>
        </div>
        <div class="ns-card__stat">
          <span>健康</span>
          <strong>{{ metric(namespace.healthyNodeCount) }}</strong>
        </div>
        <div class="ns-card__stat">
          <span>连接</span>
          <strong>{{ metric(namespace.connectionCount) }}</strong>
        </div>
      </div>

      <footer class="ns-card__foot">
        <span class="ns-card__desc" :title="descText">{{ descText }}</span>
      </footer>
      </article>
    </RsContextMenu>
  </div>
</template>

<script lang="ts" setup>
import { GIcon } from '@/components/gicon'
import { RsContextMenu } from '@/ui'
import { LayersOutline } from '@vicons/ionicons5'
import { computed } from 'vue'
import { buildNamespaceContextMenu } from '../hooks/model'
import type { Namespace } from '../types'

defineOptions({
  name: 'NamespaceCard',
})

const props = withDefaults(defineProps<{
  namespace: Namespace
  selected?: boolean
  menu?: 'manage' | 'view' | 'none'
}>(), {
  selected: false,
  menu: 'manage',
})

const emit = defineEmits<{
  (e: 'select'): void
  (e: 'focus'): void
  (e: 'action', key: string): void
}>()

const envMap: Record<string, string> = {
  DEVELOPMENT: '开发',
  STAGING: '预发布',
  PRODUCTION: '生产',
}

const envText = computed(() => envMap[props.namespace.environment] || props.namespace.environment || '-')

const envTone = computed(() => {
  if (props.namespace.environment === 'PRODUCTION') return 'prod'
  if (props.namespace.environment === 'STAGING') return 'stage'
  return 'dev'
})

const tone = computed(() => (props.namespace.activeFlag === 'Y' ? 'running' : 'idle'))
const liveText = computed(() => (props.namespace.activeFlag === 'Y' ? '活动' : '非活动'))
const metaText = computed(() => `${envText.value} · ${props.namespace.instanceName || '-'}`)
const descText = computed(() => props.namespace.namespaceDescription || props.namespace.noteText || '暂无描述')

const menuItems = computed(() => {
  if (props.menu === 'none') return []
  return buildNamespaceContextMenu(props.menu === 'view')
})

const serviceText = computed(() => {
  const count = metric(props.namespace.serviceCount)
  const quota = props.namespace.serviceQuotaLimit
  if (quota === undefined || quota === null || quota === 0) {
    return `${count} / 无限制`
  }
  return `${count} / ${quota}`
})

function metric(value?: number) {
  if (value === undefined || value === null) return '0'
  return String(value)
}

function onMenuSelect(key: string) {
  emit('action', key)
}
</script>

<style lang="scss" scoped>
.ns-card-host {
  min-width: 0;
}

.ns-card {
  --ns-card-radius: 14px;
  --ns-stat-radius: 10px;
  --ns-inset: rgba(120, 120, 128, 0.12);
  --ns-hairline: rgba(60, 60, 67, 0.12);
  --ns-caption: #86868b;

  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  gap: 16px;
  width: 100%;
  padding: 18px 18px 16px;
  border: 1px solid var(--g-border-color, var(--ns-hairline));
  border-radius: var(--ns-card-radius);
  background: var(--g-bg-primary);
  cursor: pointer;
  user-select: none;
  -webkit-user-select: none;
}

.ns-card:hover {
  border-color: color-mix(in srgb, var(--g-primary) 45%, var(--g-border-color, var(--ns-hairline)));
}

.ns-card.is-selected {
  border-color: var(--g-primary);
}

.ns-card__head {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  min-width: 0;
  flex: none;
}

.ns-card__icon {
  flex: none;
  display: grid;
  place-items: center;
  width: 44px;
  height: 44px;
  border-radius: 12px;
  color: #fff;
}

.ns-card__icon.is-dev {
  background: linear-gradient(180deg, #64d2ff 0%, #007aff 100%);
}

.ns-card__icon.is-stage {
  background: linear-gradient(180deg, #ffd60a 0%, #ff9f0a 100%);
}

.ns-card__icon.is-prod {
  background: linear-gradient(180deg, #bf5af2 0%, #af52de 100%);
}

.ns-card__identity {
  min-width: 0;
  flex: 1 1 auto;
}

.ns-card__name {
  margin: 0;
  font-size: 17px;
  font-weight: 600;
  line-height: 1.3;
  letter-spacing: -0.022em;
  color: var(--g-text-primary);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.ns-card__sub {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  margin-top: 6px;
  min-width: 0;
}

.ns-card__meta {
  margin: 0;
  min-width: 0;
  flex: 1 1 auto;
  font-size: 13px;
  line-height: 18px;
  color: var(--ns-caption);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.ns-card__live {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  margin: 0;
  flex: none;
  font-size: 13px;
  font-weight: 500;
  line-height: 18px;
  color: var(--g-text-secondary);
}

.ns-card__orb {
  flex: none;
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #c7c7cc;
}

.ns-card__live.is-running {
  color: #30d158;
}

.ns-card__live.is-running .ns-card__orb {
  background: #30d158;
  box-shadow: 0 0 0 3px rgba(48, 209, 88, 0.18);
}

.ns-card__stats {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
  flex: none;
}

.ns-card__stat {
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 6px;
  min-height: 58px;
  padding: 10px 12px;
  border-radius: var(--ns-stat-radius);
  background: var(--ns-inset);
}

.ns-card__stat span {
  font-size: 12px;
  line-height: 16px;
  color: var(--g-text-secondary);
}

.ns-card__stat strong {
  font-size: 15px;
  font-weight: 600;
  line-height: 20px;
  font-variant-numeric: tabular-nums;
  color: var(--g-text-primary);
}

.ns-card__foot {
  flex: none;
  min-height: 16px;
  min-width: 0;
  font-size: 12px;
  line-height: 16px;
  color: var(--ns-caption);
}

.ns-card__desc {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

[data-theme='dark'] .ns-card {
  --ns-inset: rgba(235, 235, 245, 0.08);
  --ns-hairline: rgba(235, 235, 245, 0.12);
  --ns-caption: #98989d;
}
</style>
