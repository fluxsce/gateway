import { updateTimeout } from '@/api/request'
import { useAppMessage } from '@/composables/useAppMessage'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { store } from '@/stores'
import { getApiMessage, isApiSuccess, parseJsonData } from '@/utils/format'
import { computed, reactive, ref } from 'vue'
import { getEnvSettings, saveEnvSetting, saveEnvVar, deleteEnvVar } from '../api'
import type {
  EnvSettingGroupCode,
  EnvSettings,
  EnvVarItem,
  EnvVarsSettings,
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

const emptyEnvVars = (): EnvVarsSettings => ({
  items: [],
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
  const savingEnvVar = ref(false)
  const retention = reactive(emptyRetention())
  const retentionJob = reactive(emptyRetentionJob())
  const webTimeout = reactive(emptyWebTimeout())
  const envVars = reactive(emptyEnvVars())

  const canEdit = computed(() => store.user.hasPermission('hub0009:edit'))

  const applySettings = (data: EnvSettings | null) => {
    Object.assign(retention, emptyRetention(), data?.retention || {})
    Object.assign(retentionJob, emptyRetentionJob(), data?.retentionJob || {})
    Object.assign(webTimeout, emptyWebTimeout(), data?.webTimeout || {})
    Object.assign(envVars, emptyEnvVars(), {
      items: data?.envVars?.items || [],
      currentVersion: data?.envVars?.currentVersion ?? 0,
    })
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

  const applyEnvVarResult = (result: Awaited<ReturnType<typeof saveEnvVar>>) => {
    const saved = parseJsonData<{ currentVersion?: number; items?: EnvVarItem[] }>(result, {})
    if (typeof saved.currentVersion === 'number') {
      envVars.currentVersion = saved.currentVersion
    }
    if (Array.isArray(saved.items)) {
      envVars.items = saved.items
    }
  }

  const saveVariable = async (payload: {
    name: string
    originalName?: string
    value: string
    secret: boolean
    note: string
  }) => {
    if (!canEdit.value) {
      message.warning(t('common.noPermission'))
      return false
    }
    savingEnvVar.value = true
    try {
      const result = await saveEnvVar({
        ...payload,
        currentVersion: envVars.currentVersion,
      })
      if (!isApiSuccess(result)) {
        message.error(getApiMessage(result, t('common.saveFailed')))
        return false
      }
      applyEnvVarResult(result)
      message.success(getApiMessage(result, t('common.saveSuccess')))
      return true
    } catch {
      message.error(t('common.saveFailed'))
      return false
    } finally {
      savingEnvVar.value = false
    }
  }

  const removeVariable = async (name: string) => {
    if (!canEdit.value) {
      message.warning(t('common.noPermission'))
      return false
    }
    savingEnvVar.value = true
    try {
      const result = await deleteEnvVar({
        name,
        currentVersion: envVars.currentVersion,
      })
      if (!isApiSuccess(result)) {
        message.error(getApiMessage(result, t('envVars.deleteFailed')))
        return false
      }
      applyEnvVarResult(result)
      message.success(getApiMessage(result, t('envVars.deleteSuccess')))
      return true
    } catch {
      message.error(t('envVars.deleteFailed'))
      return false
    } finally {
      savingEnvVar.value = false
    }
  }

  return {
    loading,
    retention,
    retentionJob,
    webTimeout,
    envVars,
    canEdit,
    savingRetention,
    savingRetentionJob,
    savingWebTimeout,
    savingEnvVar,
    fetchSettings,
    saveGroup,
    saveVariable,
    removeVariable,
  }
}
