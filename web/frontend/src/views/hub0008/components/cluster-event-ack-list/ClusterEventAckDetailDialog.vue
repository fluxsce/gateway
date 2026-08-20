<template>
  <RsDialog
    :open="localShow"
    :title="t('ack.dialog.title')"
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
      <div class="cluster-event-ack-detail">
        <RsDescriptions
          :columns="2"
          bordered
          label-placement="left"
          size="sm"
          class="cluster-event-ack-detail__info"
        >
          <RsDescriptionsItem :label="t('ack.columns.ackId')">
            {{ ack?.ackId || '-' }}
          </RsDescriptionsItem>
          <RsDescriptionsItem :label="t('ack.dialog.eventId')">
            {{ ack?.eventId || '-' }}
          </RsDescriptionsItem>
          <RsDescriptionsItem :label="t('ack.columns.nodeId')">
            {{ ack?.nodeId || '-' }}
          </RsDescriptionsItem>
          <RsDescriptionsItem :label="t('ack.columns.nodeIp')">
            {{ ack?.nodeIp || '-' }}
          </RsDescriptionsItem>
          <RsDescriptionsItem :label="t('ack.columns.ackStatus')">
            <RsTag :variant="getAckStatusVariant(ack?.ackStatus)" size="sm">
              {{ getAckStatusLabel(ack?.ackStatus) }}
            </RsTag>
          </RsDescriptionsItem>
          <RsDescriptionsItem :label="t('ack.columns.retryCount')">
            {{ ack?.retryCount ?? '-' }}
          </RsDescriptionsItem>
          <RsDescriptionsItem :label="t('ack.columns.processTime')">
            {{ formatDateString(ack?.processTime) || '-' }}
          </RsDescriptionsItem>
          <RsDescriptionsItem :label="t('ack.columns.activeFlag')">
            <RsTag :variant="ack?.activeFlag === 'Y' ? 'success' : 'danger'" size="sm">
              {{ ack?.activeFlag === 'Y' ? t('common.active') : t('common.inactive') }}
            </RsTag>
          </RsDescriptionsItem>
          <RsDescriptionsItem :label="t('ack.dialog.addTime')">
            {{ formatDateString(ack?.addTime) || '-' }}
          </RsDescriptionsItem>
          <RsDescriptionsItem :label="t('ack.dialog.addWho')">
            {{ ack?.addWho || '-' }}
          </RsDescriptionsItem>
          <RsDescriptionsItem :label="t('ack.dialog.editTime')">
            {{ formatDateString(ack?.editTime) || '-' }}
          </RsDescriptionsItem>
          <RsDescriptionsItem :label="t('ack.dialog.editWho')">
            {{ ack?.editWho || '-' }}
          </RsDescriptionsItem>
        </RsDescriptions>

        <div class="cluster-event-ack-detail__result">
          <div class="cluster-event-ack-detail__result-title">{{ t('ack.dialog.resultTitle') }}</div>
          <div class="cluster-event-ack-detail__result-content">
            {{ ack?.resultMessage || '-' }}
          </div>
        </div>

        <div v-if="ack?.noteText" class="cluster-event-ack-detail__note">
          <div class="cluster-event-ack-detail__note-title">{{ t('ack.dialog.noteTitle') }}</div>
          <div class="cluster-event-ack-detail__note-content">
            {{ ack.noteText }}
          </div>
        </div>

        <div v-if="ack?.extProperty" class="cluster-event-ack-detail__ext">
          <div class="cluster-event-ack-detail__ext-title">{{ t('ack.dialog.extTitle') }}</div>
          <RsCodeBlock :code="extPropertyCode" lang="json" />
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
import type { ClusterEventAck } from '../../types'

defineOptions({
  name: 'ClusterEventAckDetailDialog',
})

interface Props {
  show?: boolean
  ack?: ClusterEventAck | null
}

const props = withDefaults(defineProps<Props>(), {
  show: false,
  ack: null,
})

const emit = defineEmits<{
  (e: 'update:show', value: boolean): void
  (e: 'close'): void
}>()

const { t } = useModuleI18n('hub0008')
const localShow = ref(props.show)

/** 格式化后的扩展属性（尽量 pretty-print JSON） */
const extPropertyCode = computed(() => {
  const raw = props.ack?.extProperty || ''
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

function getAckStatusVariant(status?: string | null): RsTagVariant {
  switch (status) {
    case 'SUCCESS':
      return 'success'
    case 'FAILED':
      return 'danger'
    case 'PENDING':
      return 'warning'
    default:
      return 'default'
  }
}

function getAckStatusLabel(status?: string | null): string {
  switch (status) {
    case 'PENDING':
      return t('ack.status.pending')
    case 'SUCCESS':
      return t('ack.status.success')
    case 'FAILED':
      return t('ack.status.failed')
    case 'SKIPPED':
      return t('ack.status.skipped')
    default:
      return status || '-'
  }
}

const formatDateString = (dateStr?: string | null): string => {
  if (!dateStr) return '-'
  return formatDate(dateStr, 'YYYY-MM-DD HH:mm:ss')
}
</script>

<style scoped lang="scss">
.cluster-event-ack-detail {
  display: flex;
  flex-direction: column;
  gap: var(--g-space-lg, 24px);

  &__info {
    width: 100%;
  }

  &__result,
  &__note,
  &__ext {
    display: flex;
    flex-direction: column;
    gap: var(--g-space-sm, 8px);

    &-title {
      font-size: var(--g-font-size-base, 14px);
      font-weight: 600;
      color: var(--g-text-primary);
    }

    &-content {
      font-size: var(--g-font-size-base, 14px);
      color: var(--g-text-primary);
      line-height: 1.6;
      word-break: break-word;
      padding: var(--g-space-sm, 8px);
      background-color: var(--g-bg-tertiary, #f5f5f5);
      border-radius: var(--g-radius-sm, 4px);
    }
  }
}
</style>
