<template>
  <div class="health-panel">
    <RsLoading :loading="loading" overlay block size="lg" />
    <RsAlert type="info" class="hint">{{ t('health.hint') }}</RsAlert>

    <div class="health-hero">
      <div class="health-hero__status" :data-tone="health.status">
        <span class="health-hero__kicker">{{ t('health.overall') }}</span>
        <strong>{{ statusLabel(health.status) }}</strong>
      </div>
      <dl class="health-hero__meta">
        <div>
          <dt>{{ t('health.statusOk') }}</dt>
          <dd>{{ statusCount.ok }}</dd>
        </div>
        <div>
          <dt>{{ t('health.statusSkipped') }}</dt>
          <dd>{{ statusCount.skipped }}</dd>
        </div>
        <div>
          <dt>{{ t('health.statusError') }}</dt>
          <dd>{{ statusCount.error }}</dd>
        </div>
        <div v-if="checkedAtText">
          <dt>{{ t('health.checkedAt') }}</dt>
          <dd class="is-time">{{ checkedAtText }}</dd>
        </div>
      </dl>
      <RsButton variant="secondary" :loading="loading" @click="fetchHealth">
        {{ t('common.refresh') }}
      </RsButton>
    </div>

    <div v-if="health.items.length" class="health-panel__grid">
      <article
        v-for="item in health.items"
        :key="rowKey(item)"
        class="health-card"
        :class="{ 'is-action': hasDetail(item) }"
        :data-tone="item.status"
        @click="hasDetail(item) ? openDetail(item) : undefined"
      >
        <header class="health-card__head">
          <span class="health-card__icon" aria-hidden="true">
            <RsIcon :name="kindIcon(item.kind)" :size="16" />
          </span>
          <div class="health-card__titles">
            <span class="health-card__kind">{{ kindLabel(item.kind) }}</span>
            <strong class="health-card__name">{{ item.name }}</strong>
          </div>
        </header>
        <p class="health-card__kpi">{{ statusLabel(item.status) }}</p>
        <p class="health-card__metric">
          <span>{{ t('health.latency') }}</span>
          <b>{{ latencyText(item) }}</b>
        </p>
        <i class="health-card__bar" aria-hidden="true" />
        <dl class="health-card__chips">
          <div v-if="item.driver">
            <dt>{{ t('health.driver') }}</dt>
            <dd>{{ item.driver }}</dd>
          </div>
          <div v-if="item.database">
            <dt>{{ t('health.database') }}</dt>
            <dd>{{ item.database }}</dd>
          </div>
          <div v-if="objectReady(item)">
            <dt>{{ t('health.objectReadyHint') }}</dt>
            <dd>{{ objectReady(item) }}</dd>
          </div>
        </dl>
        <p v-if="item.message" class="health-card__msg">{{ item.message }}</p>
        <button v-if="hasDetail(item)" type="button" class="health-card__more" @click.stop="openDetail(item)">
          {{ t('health.viewDetail') }}
        </button>
      </article>
    </div>
    <RsEmpty v-else-if="!loading" :description="t('health.empty')" />

    <RsDialog
      :open="detailOpen"
      :title="detail ? t('health.detailTitle', { name: detail.name }) : t('health.viewDetail')"
      layout="window"
      :width="680"
      :show-overlay="true"
      :close-on-overlay-click="true"
      @update:open="onDetailOpen"
    >
      <template #body>
        <div v-if="detail" class="health-detail">
          <RsAlert type="info">{{ t('health.detailHint') }}</RsAlert>
          <h3 class="health-detail__section">{{ t('health.sectionConn') }}</h3>
          <RsDescriptions :columns="2" size="sm" bordered label-placement="left">
            <RsDescriptionsItem :label="t('health.name')">{{ detail.name }}</RsDescriptionsItem>
            <RsDescriptionsItem :label="t('health.kind')">{{ kindLabel(detail.kind) }}</RsDescriptionsItem>
            <RsDescriptionsItem :label="t('health.status')">
              <RsTag :variant="statusVariant(detail.status)" size="sm">
                {{ statusLabel(detail.status) }}
              </RsTag>
            </RsDescriptionsItem>
            <RsDescriptionsItem :label="t('health.enabled')">
              {{ detail.enabled ? t('health.enabledYes') : t('health.enabledNo') }}
              <span class="health-detail__note">{{ t('health.enabledHint') }}</span>
            </RsDescriptionsItem>
            <RsDescriptionsItem :label="t('health.driver')">{{ detail.driver || '—' }}</RsDescriptionsItem>
            <RsDescriptionsItem :label="t('health.database')">{{ detail.database || '—' }}</RsDescriptionsItem>
            <RsDescriptionsItem :label="t('health.latency')">{{ latencyText(detail) }}</RsDescriptionsItem>
            <RsDescriptionsItem :label="t('health.message')">{{ detail.message || '—' }}</RsDescriptionsItem>
          </RsDescriptions>

          <h3 class="health-detail__section">
            {{ t('health.sectionObjects') }}
            <span v-if="objectReady(detail)" class="health-detail__ready">{{ objectReady(detail) }}</span>
          </h3>
          <div v-if="detail.objects?.length" class="health-detail__objects">
            <section v-for="obj in detail.objects" :key="`${obj.kind}:${obj.name}`" class="health-detail__object">
              <header class="health-detail__object-head">
                <strong>{{ obj.name }}</strong>
                <RsTag variant="info" size="sm">{{ objectKindLabel(obj.kind) }}</RsTag>
                <RsTag :variant="obj.exists ? 'success' : 'default'" size="sm">
                  {{ obj.exists ? t('health.existsLabel') : t('health.existsMissing') }}
                </RsTag>
              </header>
              <ul v-if="obj.indexes?.length" class="health-detail__indexes">
                <li class="is-head">
                  <span>{{ t('health.indexName') }}</span>
                  <span>{{ t('health.indexKeys') }}</span>
                  <span>{{ t('health.indexStatus') }}</span>
                </li>
                <li v-for="idx in obj.indexes" :key="idx.name">
                  <span class="health-detail__index-name">{{ idx.name }}</span>
                  <span class="health-detail__index-keys">{{ idx.keys }}</span>
                  <RsTag :variant="idx.present ? 'success' : 'default'" size="sm">
                    {{ idx.present ? t('health.indexReady') : t('health.indexMissing') }}
                  </RsTag>
                </li>
              </ul>
            </section>
          </div>
          <p v-else class="health-detail__empty">{{ t('health.noObjects') }}</p>
        </div>
      </template>
      <template #footer>
        <RsButton variant="secondary" @click="detailOpen = false">{{ t('health.close') }}</RsButton>
      </template>
    </RsDialog>
  </div>
</template>

<script setup lang="ts">
import { useModuleI18n } from '@/hooks/useModuleI18n'
import {
  RsAlert,
  RsButton,
  RsDescriptions,
  RsDescriptionsItem,
  RsDialog,
  RsEmpty,
  RsIcon,
  RsLoading,
  RsTag,
  type RsTagVariant,
} from '@/ui'
import { formatDate } from '@/utils/format'
import { computed, onMounted, ref } from 'vue'
import { useSystemHealth } from '../hooks/health'
import type { HealthItem, HealthKind, HealthStatus, HealthStoreObject } from '../types'

defineOptions({ name: 'HealthPanel' })

const { t } = useModuleI18n('hub0009')
const { loading, health, fetchHealth } = useSystemHealth()
const detailOpen = ref(false)
const detail = ref<HealthItem | null>(null)

const statusVariant = (status: HealthStatus): RsTagVariant => {
  if (status === 'ok') return 'success'
  if (status === 'error' || status === 'degraded') return 'danger'
  return 'default'
}

const statusLabel = (status: HealthStatus) => {
  if (status === 'ok') return t('health.statusOk')
  if (status === 'error') return t('health.statusError')
  if (status === 'degraded') return t('health.statusDegraded')
  return t('health.statusSkipped')
}

const kindLabel = (kind: HealthKind) => {
  if (kind === 'process') return t('health.kindProcess')
  if (kind === 'sql') return t('health.kindSql')
  if (kind === 'clickhouse') return t('health.kindClickhouse')
  if (kind === 'mongo') return t('health.kindMongo')
  return t('health.kindCache')
}

const kindIcon = (kind: HealthKind) => {
  if (kind === 'process') return 'activity'
  if (kind === 'sql') return 'database'
  if (kind === 'clickhouse') return 'layers'
  if (kind === 'mongo') return 'cylinder'
  return 'zap'
}

const objectKindLabel = (kind: HealthStoreObject['kind']) => {
  if (kind === 'table') return t('health.objectTable')
  return t('health.objectCollection')
}

const hasDetail = (item: HealthItem) => item.kind === 'clickhouse' || item.kind === 'mongo'

const latencyText = (item: HealthItem) => {
  if (item.status === 'skipped') {
    return '—'
  }
  return `${item.latencyMs} ${t('health.ms')}`
}

const objectReady = (item: HealthItem) => {
  if (!item.objects?.length) {
    return ''
  }
  const ready = item.objects.filter((obj) => obj.exists).length
  return t('health.objectsReady', { ready, total: item.objects.length })
}

const statusCount = computed(() => {
  const count = { ok: 0, skipped: 0, error: 0 }
  for (const item of health.value.items) {
    if (item.status === 'ok') count.ok += 1
    else if (item.status === 'error') count.error += 1
    else count.skipped += 1
  }
  return count
})

const checkedAtText = computed(() => {
  if (!health.value.checkedAt) {
    return ''
  }
  return formatDate(health.value.checkedAt * 1000)
})

const rowKey = (row: HealthItem) => `${row.kind}:${row.name}`

const openDetail = (item: HealthItem) => {
  detail.value = item
  detailOpen.value = true
}

const onDetailOpen = (open: boolean) => {
  detailOpen.value = open
  if (!open) {
    detail.value = null
  }
}

onMounted(() => {
  fetchHealth()
})
</script>

<style lang="scss" scoped>
.health-panel {
  position: relative;
}

.hint {
  margin-bottom: 16px;
}

.health-hero {
  display: flex;
  align-items: stretch;
  gap: 16px;
  margin-bottom: 16px;
  padding: 14px 16px;
  background: var(--rs-surface);
  border: 1px solid var(--rs-border);
  border-radius: 10px;
  box-shadow: inset 0 1px 0 color-mix(in srgb, var(--rs-primary) 14%, transparent);
}

.health-hero__status {
  --health-tone: var(--rs-success);
  min-width: 7.5rem;
  padding-right: 16px;
  border-right: 1px solid var(--rs-border);
}

.health-hero__status[data-tone='degraded'],
.health-hero__status[data-tone='error'] {
  --health-tone: var(--rs-danger);
}

.health-hero__kicker {
  display: block;
  color: var(--rs-muted);
  font-size: var(--rs-font-size-xs);
}

.health-hero__status strong {
  display: block;
  margin-top: 4px;
  color: var(--health-tone);
  font-size: 1.75rem;
  font-weight: 700;
  line-height: 1.2;
  letter-spacing: -0.03em;
}

.health-hero__meta {
  display: flex;
  flex: 1 1 auto;
  flex-wrap: wrap;
  gap: 16px 24px;
  align-items: center;
  margin: 0;
}

.health-hero__meta dt {
  color: var(--rs-muted);
  font-size: var(--rs-font-size-xs);
}

.health-hero__meta dd {
  margin: 4px 0 0;
  color: var(--rs-fg);
  font-size: 1.125rem;
  font-weight: 650;
  font-variant-numeric: tabular-nums;
}

.health-hero__meta dd.is-time {
  font-size: var(--rs-font-size-sm);
  font-weight: 500;
}

.health-panel__grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(17rem, 1fr));
  gap: 12px;
}

.health-card {
  --health-tone: var(--rs-muted);
  position: relative;
  display: flex;
  flex-direction: column;
  min-width: 0;
  padding: 16px 16px 14px;
  overflow: hidden;
  background:
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--health-tone) 12%, var(--rs-surface)),
      var(--rs-surface) 42%
    );
  border: 1px solid color-mix(in srgb, var(--health-tone) 26%, var(--rs-border));
  border-radius: 10px;
  box-shadow: inset 0 1px 0 color-mix(in srgb, var(--health-tone) 22%, transparent);
}

.health-card[data-tone='ok'] {
  --health-tone: var(--rs-success);
}

.health-card[data-tone='error'] {
  --health-tone: var(--rs-danger);
}

.health-card[data-tone='skipped'] {
  --health-tone: var(--rs-warning);
}

.health-card.is-action {
  cursor: pointer;
}

.health-card.is-action:hover {
  border-color: color-mix(in srgb, var(--health-tone) 42%, var(--rs-border));
}

.health-card::before {
  content: '';
  position: absolute;
  top: 0;
  right: 20px;
  left: 20px;
  height: 2px;
  background: linear-gradient(90deg, transparent, var(--health-tone), transparent);
}

.health-card__head {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
}

.health-card__icon {
  display: flex;
  flex: 0 0 2rem;
  align-items: center;
  justify-content: center;
  width: 2rem;
  height: 2rem;
  color: var(--health-tone);
  background: color-mix(in srgb, var(--health-tone) 14%, transparent);
  border-radius: 8px;
}

.health-card__titles {
  min-width: 0;
}

.health-card__kind {
  display: block;
  color: var(--rs-muted);
  font-size: var(--rs-font-size-xs);
  letter-spacing: 0.04em;
}

.health-card__name {
  display: block;
  color: var(--rs-fg);
  font-size: var(--rs-font-size-sm);
  font-weight: 600;
  line-height: 1.4;
  overflow-wrap: anywhere;
}

.health-card__kpi {
  margin: 0;
  color: var(--health-tone);
  font-size: 1.625rem;
  font-weight: 700;
  line-height: 1.15;
  letter-spacing: -0.03em;
}

.health-card__metric {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 8px;
  margin: 8px 0 0;
  color: var(--rs-muted);
  font-size: var(--rs-font-size-xs);
}

.health-card__metric b {
  color: var(--rs-fg);
  font-size: var(--rs-font-size-sm);
  font-variant-numeric: tabular-nums;
}

.health-card__bar {
  display: block;
  height: 3px;
  margin: 12px 0;
  overflow: hidden;
  background: color-mix(in srgb, var(--health-tone) 16%, var(--rs-border));
}

.health-card__bar::after {
  content: '';
  display: block;
  width: 62%;
  height: 100%;
  background: linear-gradient(90deg, var(--health-tone), color-mix(in srgb, var(--health-tone) 35%, transparent));
}

.health-card[data-tone='skipped'] .health-card__bar::after {
  width: 18%;
}

.health-card[data-tone='error'] .health-card__bar::after {
  width: 100%;
}

.health-card__chips {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px 12px;
  margin: 0;
}

.health-card__chips dt {
  color: var(--rs-muted);
  font-size: 11px;
}

.health-card__chips dd {
  margin: 2px 0 0;
  color: var(--rs-fg);
  font-size: var(--rs-font-size-xs);
  overflow-wrap: anywhere;
}

.health-card__msg {
  margin: 10px 0 0;
  color: var(--rs-muted);
  font-size: var(--rs-font-size-xs);
}

.health-card__more {
  margin-top: 12px;
  padding: 0;
  color: var(--rs-primary);
  font-size: var(--rs-font-size-xs);
  background: none;
  border: 0;
  cursor: pointer;
  text-align: left;
}

.health-detail {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.health-detail__section {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 4px 0 0;
  color: var(--rs-fg);
  font-size: var(--rs-font-size-sm);
  font-weight: 600;
}

.health-detail__note {
  display: block;
  margin-top: 4px;
  color: var(--rs-muted);
  font-size: var(--rs-font-size-xs);
  font-weight: 400;
}

.health-detail__ready {
  color: var(--rs-muted);
  font-size: var(--rs-font-size-xs);
  font-weight: 500;
}

.health-detail__objects {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.health-detail__object {
  padding: 10px 12px;
  background: var(--rs-surface);
  border: 1px solid var(--rs-border);
  border-radius: 8px;
}

.health-detail__object-head {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}

.health-detail__indexes {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin: 10px 0 0;
  padding: 0;
  list-style: none;
}

.health-detail__indexes li {
  display: grid;
  grid-template-columns: minmax(8rem, 0.7fr) minmax(0, 1.5fr) 4.5rem;
  gap: 8px;
  align-items: center;
}

.health-detail__indexes li.is-head {
  color: var(--rs-muted);
  font-size: 11px;
}

.health-detail__index-name {
  font-family: var(--rs-font-mono, ui-monospace, monospace);
  font-size: var(--rs-font-size-xs);
}

.health-detail__index-keys {
  color: var(--rs-muted);
  font-size: var(--rs-font-size-xs);
}

.health-detail__empty {
  margin: 0;
  color: var(--rs-muted);
  font-size: var(--rs-font-size-sm);
}

@media (max-width: 720px) {
  .health-hero {
    flex-wrap: wrap;
  }

  .health-hero__status {
    border-right: 0;
    padding-right: 0;
  }
}
</style>
