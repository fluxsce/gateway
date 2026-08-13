import { computed, type ComputedRef } from 'vue'
import type { RsDatePickerShortcut } from '@/ui'
import { useModuleI18n } from '@/hooks/useModuleI18n'

/** 监控图表时间范围快捷项（最近 1/6/12/24 小时、7 天） */
export function useMetricsTimeRangeShortcuts(): ComputedRef<RsDatePickerShortcut[]> {
  const { t } = useModuleI18n('hub0000')

  function rangeOf(ms: number): [number, number] {
    const end = Date.now()
    return [end - ms, end]
  }

  return computed(() => [
    { label: t('timeRangeShortcuts.lastHour'), value: () => rangeOf(3600 * 1000) },
    { label: t('timeRangeShortcuts.last6Hours'), value: () => rangeOf(6 * 3600 * 1000) },
    { label: t('timeRangeShortcuts.last12Hours'), value: () => rangeOf(12 * 3600 * 1000) },
    { label: t('timeRangeShortcuts.last24Hours'), value: () => rangeOf(24 * 3600 * 1000) },
    { label: t('timeRangeShortcuts.last7Days'), value: () => rangeOf(7 * 24 * 3600 * 1000) },
  ])
}
