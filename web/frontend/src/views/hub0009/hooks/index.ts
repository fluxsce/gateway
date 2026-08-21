import { updateTimeout } from '@/api/request'
import { useAppMessage } from '@/composables/useAppMessage'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { store } from '@/stores'
import { getApiMessage, isApiSuccess, parseJsonData } from '@/utils/format'
import { computed, reactive, ref } from 'vue'
import { getEnvSettings, saveEnvSetting } from '../api'
import type {
  EnvSettingGroupCode,
  EnvSettings,
  RetentionJobSettings,
  RetentionSettings,
  WebTimeoutSettings,
} from '../types'

const emptyRetention = (): RetentionSettings => ({
  auditLogDays: 180,
  taskLogDays: 30,
  alertLogDays: 7,
  clusterEventDays: 1,
  metricsDays: 30,
  gatewayLogDefaultDays: 30,
  currentVersion: 0,
})

const emptyRetentionJob = (): RetentionJobSettings => ({
  enabled: true,
  intervalMinutes: 60,
  startTime: '',
  currentVersion: 0,
})

const emptyWebTimeout = (): WebTimeoutSettings => ({
  requestTimeoutSeconds: 120,
  sessionExpireHours: 12,
  currentVersion: 0,
})

/** 环境设置页状态与保存。 */
export function useEnvironmentSettings() {
  const { t } = useModuleI18n('hub0009')
  const message = useAppMessage()
  const loading = ref(false)
  const savingRetention = ref(false)
  const savingRetentionJob = ref(false)
  const savingWebTimeout = ref(false)
  const retention = reactive(emptyRetention())
  const retentionJob = reactive(emptyRetentionJob())
  const webTimeout = reactive(emptyWebTimeout())

  const canEdit = computed(() => store.user.hasPermission('hub0009:edit'))

  const applySettings = (data: EnvSettings | null) => {
    Object.assign(retention, emptyRetention(), data?.retention || {})
    Object.assign(retentionJob, emptyRetentionJob(), data?.retentionJob || {})
    Object.assign(webTimeout, emptyWebTimeout(), data?.webTimeout || {})
  }

  const fetchSettings = async () => {
    loading.value = true
    try {
      const result = await getEnvSettings()
      if (!isApiSuccess(result)) {
        message.error(getApiMessage(result, t('common.loadFailed')))
        return
      }
      applySettings(parseJsonData<EnvSettings | null>(result, null))
    } catch {
      message.error(t('common.loadFailed'))
    } finally {
      loading.value = false
    }
  }

  const savingRef = (groupCode: EnvSettingGroupCode) => {
    if (groupCode === 'retention') return savingRetention
    if (groupCode === 'retentionJob') return savingRetentionJob
    return savingWebTimeout
  }

  const buildPayload = (groupCode: EnvSettingGroupCode) => {
    if (groupCode === 'retention') {
      return {
        groupCode,
        currentVersion: retention.currentVersion,
        auditLogDays: retention.auditLogDays,
        taskLogDays: retention.taskLogDays,
        alertLogDays: retention.alertLogDays,
        clusterEventDays: retention.clusterEventDays,
        metricsDays: retention.metricsDays,
        gatewayLogDefaultDays: retention.gatewayLogDefaultDays,
      }
    }
    if (groupCode === 'retentionJob') {
      return {
        groupCode,
        currentVersion: retentionJob.currentVersion,
        enabled: retentionJob.enabled,
        intervalMinutes: retentionJob.intervalMinutes,
        startTime: retentionJob.startTime || '',
      }
    }
    return {
      groupCode,
      currentVersion: webTimeout.currentVersion,
      requestTimeoutSeconds: webTimeout.requestTimeoutSeconds,
      sessionExpireHours: webTimeout.sessionExpireHours,
    }
  }

  const saveGroup = async (groupCode: EnvSettingGroupCode) => {
    if (!canEdit.value) {
      message.warning(t('common.noPermission'))
      return
    }
    const saving = savingRef(groupCode)
    saving.value = true
    try {
      const result = await saveEnvSetting(buildPayload(groupCode))
      if (!isApiSuccess(result)) {
        message.error(getApiMessage(result, t('common.saveFailed')))
        return
      }
      const saved = parseJsonData<{ currentVersion?: number }>(result, {})
      if (typeof saved.currentVersion === 'number') {
        if (groupCode === 'retention') {
          retention.currentVersion = saved.currentVersion
        } else if (groupCode === 'retentionJob') {
          retentionJob.currentVersion = saved.currentVersion
        } else {
          webTimeout.currentVersion = saved.currentVersion
          updateTimeout(webTimeout.requestTimeoutSeconds * 1000)
        }
      }
      message.success(getApiMessage(result, t('common.saveSuccess')))
    } catch {
      message.error(t('common.saveFailed'))
    } finally {
      saving.value = false
    }
  }

  return {
    loading,
    retention,
    retentionJob,
    webTimeout,
    canEdit,
    savingRetention,
    savingRetentionJob,
    savingWebTimeout,
    fetchSettings,
    saveGroup,
  }
}
