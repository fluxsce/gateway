<template>
  <RsDateTimePicker
    v-model="rangeModel"
    class="metrics-datetime-range"
    range
    with-seconds
    :placeholder="placeholderText"
    :shortcuts="shortcuts"
  />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import {
  RsDateTimePicker,
  type RsDatePickerShortcut,
  type RsDateRangeValue,
} from '@/ui'
import { useModuleI18n } from '@/hooks/useModuleI18n'

const props = defineProps<{
  modelValue: [number, number] | null
}>()

const emit = defineEmits<{
  'update:modelValue': [value: [number, number] | null]
  change: [value: [number, number] | null]
}>()

const { t } = useModuleI18n('hub0000')

const placeholderText = computed(() => t('common.selectTimeRange'))

function pad(n: number): string {
  return String(n).padStart(2, '0')
}

function formatMs(ms: number): string {
  const d = new Date(ms)
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

function parseToMs(value?: string): number | null {
  if (!value) return null
  // 兼容 `YYYY-MM-DD HH:mm:ss`
  const normalized = value.includes('T') ? value : value.replace(' ', 'T')
  const ms = Date.parse(normalized)
  return Number.isNaN(ms) ? null : ms
}

function toRangeValue(tuple: [number, number] | null): RsDateRangeValue {
  if (!tuple) return {}
  return { start: formatMs(tuple[0]), end: formatMs(tuple[1]) }
}

function toTuple(range: RsDateRangeValue): [number, number] | null {
  const start = parseToMs(range.start)
  const end = parseToMs(range.end)
  if (start == null || end == null) return null
  return [start, end]
}

const rangeModel = computed<RsDateRangeValue>({
  get: () => toRangeValue(props.modelValue),
  set: (next) => {
    const tuple = toTuple(next ?? {})
    emit('update:modelValue', tuple)
    emit('change', tuple)
  },
})

function shortcutRange(ms: number): () => RsDateRangeValue {
  return () => {
    const end = Date.now()
    const start = end - ms
    return { start: formatMs(start), end: formatMs(end) }
  }
}

const shortcuts = computed<RsDatePickerShortcut[]>(() => [
  { label: t('timeRangeShortcuts.lastHour'), value: shortcutRange(3600 * 1000) },
  { label: t('timeRangeShortcuts.last6Hours'), value: shortcutRange(6 * 3600 * 1000) },
  { label: t('timeRangeShortcuts.last12Hours'), value: shortcutRange(12 * 3600 * 1000) },
  { label: t('timeRangeShortcuts.last24Hours'), value: shortcutRange(24 * 3600 * 1000) },
  { label: t('timeRangeShortcuts.last7Days'), value: shortcutRange(7 * 24 * 3600 * 1000) },
])
</script>

<style scoped>
.metrics-datetime-range {
  min-width: 22rem;
  max-width: 28rem;
}
</style>
