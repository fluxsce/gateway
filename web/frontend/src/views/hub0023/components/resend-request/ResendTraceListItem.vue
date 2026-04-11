<template>
  <li
    class="resend-dialog__trace-item"
    :class="{ 'resend-dialog__trace-item--active': active }"
    :data-trace-id="item.traceId"
    @click="emit('select', item.traceId)"
  >
    <span class="resend-dialog__trace-id resend-dialog__trace-id--ellipsis" :title="item.traceId">
      {{ item.traceId }}
    </span>
    <div class="resend-dialog__trace-meta">
      <n-tag size="tiny" :bordered="false">{{ item.requestMethod || '-' }}</n-tag>
      <span class="resend-dialog__path resend-dialog__path--ellipsis" :title="item.requestPath || '-'">
        {{ item.requestPath || '-' }}
      </span>
    </div>
    <div class="resend-dialog__trace-foot" role="status">
      <div class="resend-dialog__trace-foot-row">
        <span class="resend-dialog__trace-foot-label">重发状态</span>
        <n-tag size="tiny" bordered :type="replayPhaseTagType(outcome)">
          {{ replayPhaseLabel(outcome) }}
        </n-tag>
      </div>
      <div class="resend-dialog__trace-foot-row">
        <span class="resend-dialog__trace-foot-label">响应状态</span>
        <n-tag size="tiny" bordered :type="httpStatusTagType(outcome)">
          {{ responseStateLabel(outcome) }}
        </n-tag>
      </div>
    </div>
  </li>
</template>

<script setup lang="ts">
import type { GatewayLogListItem } from '../../types'
import { NTag } from 'naive-ui'
import { computed, inject } from 'vue'
import {
  defaultReplayOutcome,
  httpStatusTagType,
  replayPhaseLabel,
  replayPhaseTagType,
  resendReplayOutcomeKey,
  responseStateLabel,
} from './replayOutcomeDisplay'

defineOptions({ name: 'ResendTraceListItem' })

const props = defineProps<{
  item: GatewayLogListItem
  active: boolean
}>()

const emit = defineEmits<{
  (e: 'select', traceId: string): void
}>()

const outcomeStore = inject(resendReplayOutcomeKey, null)

const outcome = computed(() => {
  const store = outcomeStore?.value
  if (!store) {
    return defaultReplayOutcome
  }
  return store[props.item.traceId] ?? defaultReplayOutcome
})
</script>

<style scoped lang="scss">
.resend-dialog__trace-item {
  position: relative;
  padding: 12px 12px 12px 11px;
  border-radius: calc(var(--g-radius-md) + 4px);
  cursor: pointer;
  border: 1px solid var(--g-border-primary);
  border-left: 3px solid transparent;
  background: var(--g-bg-primary);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06);
  transition:
    background-color 0.2s ease,
    border-color 0.2s ease,
    box-shadow 0.2s ease,
    border-left-color 0.2s ease;

  &:hover:not(.resend-dialog__trace-item--active) {
    background: var(--g-bg-tertiary);
    border-color: var(--g-primary-light);
    border-left-color: var(--g-primary);
    box-shadow: 0 4px 14px rgba(0, 0, 0, 0.1);
  }

  &--active {
    background: var(--g-bg-primary);
    border-color: var(--g-primary-light);
    border-left-color: var(--g-primary);
    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);

    &::before {
      content: '';
      position: absolute;
      inset: 0;
      margin: 0;
      border-radius: inherit;
      background: var(--g-primary);
      opacity: 0.1;
      pointer-events: none;
      transition: opacity 0.2s ease;
    }
  }

  &--active:hover {
    border-color: var(--g-primary);
    border-left-color: var(--g-primary);
    box-shadow: 0 5px 18px rgba(0, 0, 0, 0.12);

    &::before {
      opacity: 0.14;
    }
  }
}

.resend-dialog__trace-id--ellipsis,
.resend-dialog__path--ellipsis {
  display: block;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.resend-dialog__trace-id {
  display: block;
  font-family: var(--g-font-family-mono);
  font-size: 12px;
  font-weight: 500;
  line-height: 1.45;
  color: var(--g-text-primary);
}

.resend-dialog__trace-meta {
  margin-top: 6px;
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
}

.resend-dialog__path {
  flex: 1;
  min-width: 0;
  font-size: 12px;
  color: var(--g-text-tertiary);
}

.resend-dialog__trace-foot {
  margin-top: 8px;
  padding-top: 8px;
  border-top: 1px solid var(--g-border-primary);
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.resend-dialog__trace-foot-row {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
  font-size: 12px;
}

.resend-dialog__trace-foot-label {
  flex: 0 0 4.5em;
  color: var(--g-text-tertiary);
}

.resend-dialog__trace-foot-row :deep(.n-tag) {
  flex: 1;
  min-width: 0;
  justify-content: flex-start;
}
</style>
