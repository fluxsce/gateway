import { useAppMessage } from '@/composables/useAppMessage'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { getApiMessage, isApiSuccess, parseJsonData } from '@/utils/format'
import { ref } from 'vue'
import { getSystemHealth } from '../api'
import type { SystemHealth } from '../types'

const emptyHealth = (): SystemHealth => ({
  status: 'ok',
  checkedAt: 0,
  items: [],
})

/**
 * 系统健康页状态：进入页签或手动刷新时探测依赖。
 */
export function useSystemHealth() {
  const { t } = useModuleI18n('hub0009')
  const message = useAppMessage()
  const loading = ref(false)
  const health = ref<SystemHealth>(emptyHealth())

  const fetchHealth = async () => {
    loading.value = true
    try {
      const result = await getSystemHealth()
      if (!isApiSuccess(result)) {
        message.error(getApiMessage(result, t('common.loadHealthFailed')))
        return
      }
      const data = parseJsonData<SystemHealth | null>(result, null)
      health.value = {
        status: data?.status || 'ok',
        checkedAt: data?.checkedAt || 0,
        items: (data?.items || []).map((item) => ({
          ...item,
          enabled: Boolean(item.enabled),
          database: item.database || '',
          objects: item.objects || [],
        })),
      }
    } catch {
      message.error(t('common.loadHealthFailed'))
    } finally {
      loading.value = false
    }
  }

  return {
    loading,
    health,
    fetchHealth,
  }
}
