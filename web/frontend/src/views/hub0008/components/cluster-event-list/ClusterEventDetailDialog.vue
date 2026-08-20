<template>
  <RsDialog
    :open="localShow"
    :title="t('event.dialog.title')"
    layout="window"
    width="lg"
    :show-overlay="true"
    :close-on-overlay-click="true"
    :close-on-esc="true"
    :show-footer="true"
    :show-cancel="false"
    :confirm-text="t('common.close')"
    :auto-close-on-confirm="true"
    @update:open="handleUpdateOpen"
  >
    <template #body>
      <div class="cluster-event-detail">
        <RsDescriptions
          :columns="2"
          bordered
          label-placement="left"
          size="sm"
          class="cluster-event-detail__info"
        >
          <RsDescriptionsItem :label="t('event.columns.eventId')">
            {{ event?.eventId || '-' }}
          </RsDescriptionsItem>
          <RsDescriptionsItem :label="t('event.columns.eventType')">
            <RsTag variant="primary" size="sm">
              {{ event?.eventType || '-' }}
            </RsTag>
          </RsDescriptionsItem>
          <RsDescriptionsItem :label="t('event.columns.eventAction')">
            <RsTag :variant="getEventActionVariant(event?.eventAction)" size="sm">
              {{ event?.eventAction || '-' }}
            </RsTag>
          </RsDescriptionsItem>
          <RsDescriptionsItem :label="t('event.search.sourceNodeId')">
            {{ event?.sourceNodeId || '-' }}
          </RsDescriptionsItem>
          <RsDescriptionsItem :label="t('event.columns.sourceNodeIp')">
            {{ event?.sourceNodeIp || '-' }}
          </RsDescriptionsItem>
          <RsDescriptionsItem :label="t('event.columns.eventTime')">
            {{ formatDateString(event?.eventTime) || '-' }}
          </RsDescriptionsItem>
          <RsDescriptionsItem :label="t('event.columns.expireTime')">
            {{ formatDateString(event?.expireTime) || '-' }}
          </RsDescriptionsItem>
          <RsDescriptionsItem :label="t('event.columns.activeFlag')">
            <RsTag :variant="event?.activeFlag === 'Y' ? 'success' : 'danger'" size="sm">
              {{ event?.activeFlag === 'Y' ? t('common.active') : t('common.inactive') }}
            </RsTag>
          </RsDescriptionsItem>
        </RsDescriptions>

        <div class="cluster-event-detail__payload">
          <div class="cluster-event-detail__payload-title">{{ t('event.dialog.payloadTitle') }}</div>
          <RsCodeBlock :code="payloadCode" lang="json" />
        </div>
      </div>
    </template>
  </RsDialog>
</template>

<script setup lang="ts">
import { useModuleI18n } from '@/hooks/useModuleI18n'
import {
  RsDescriptions,
  RsDescriptionsItem,
  RsDialog,
  RsTag,
  type RsTagVariant,
} from '@/ui'
import { RsCodeBlock } from '@/ui/code-block'
import { formatDate } from '@/utils/format'
import { computed, ref, watch } from 'vue'
import type { ClusterEvent } from '../../types'

defineOptions({
  name: 'ClusterEventDetailDialog',
})

interface Props {
  show?: boolean
  event?: ClusterEvent | null
}

const props = withDefaults(defineProps<Props>(), {
  show: false,
  event: null,
})

const emit = defineEmits<{
  (e: 'update:show', value: boolean): void
  (e: 'close'): void
}>()

const { t } = useModuleI18n('hub0008')
const localShow = ref(props.show)

/** 格式化后的事件载荷（尽量 pretty-print JSON） */
const payloadCode = computed(() => {
  const raw = props.event?.eventPayload || ''
  if (!raw.trim()) return ''
  try {
    return JSON.stringify(JSON.parse(raw), null, 2)
  } catch {
    return raw
  }
})

watch(
  () => props.show,
  (newVal) => {
    localShow.value = newVal
  },
)

function handleUpdateOpen(value: boolean) {
  localShow.value = value
  if (value !== props.show) {
    emit('update:show', value)
  }
  if (!value) {
    emit('close')
  }
}

/** 事件动作 → RsTag variant */
function getEventActionVariant(action?: string | null): RsTagVariant {
  switch (action) {
    case 'START':
    case 'CREATE':
      return 'success'
    case 'STOP':
    case 'DELETE':
      return 'danger'
    case 'RELOAD':
    case 'REFRESH':
    case 'INVALIDATE':
      return 'warning'
    case 'RESTART':
    case 'UPDATE':
      return 'info'
    default:
      return 'default'
  }
}

const formatDateString = (dateStr?: string | null): string => {
  if (!dateStr) return '-'
  return formatDate(dateStr, 'YYYY-MM-DD HH:mm:ss')
}
</script>

<style scoped lang="scss">
.cluster-event-detail {
  display: flex;
  flex-direction: column;
  gap: var(--g-space-lg, 24px);

  &__info {
    width: 100%;
  }

  &__payload {
    display: flex;
    flex-direction: column;
    gap: var(--g-space-sm, 8px);

    &-title {
      font-size: var(--g-font-size-base, 14px);
      font-weight: 600;
      color: var(--g-text-primary);
    }
  }
}
</style>
