export interface RetentionSettings {
  auditLogDays: number
  taskLogDays: number
  alertLogDays: number
  clusterEventDays: number
  metricsDays: number
  gatewayLogDefaultDays: number
  currentVersion: number
}

export interface RetentionJobSettings {
  enabled: boolean
  intervalMinutes: number
  startTime: string
  currentVersion: number
}

export interface WebTimeoutSettings {
  requestTimeoutSeconds: number
  sessionExpireHours: number
  currentVersion: number
}

export interface EnvVarItem {
  name: string
  value: string
  secret: boolean
  hasValue: boolean
  note: string
}

export interface EnvVarsSettings {
  items: EnvVarItem[]
  currentVersion: number
}

export interface EnvSettings {
  retention: RetentionSettings
  retentionJob: RetentionJobSettings
  webTimeout: WebTimeoutSettings
  envVars: EnvVarsSettings
}

export type EnvSettingGroupCode = 'retention' | 'retentionJob' | 'webTimeout'

export interface SaveEnvSettingPayload
  extends Partial<RetentionSettings>,
    Partial<RetentionJobSettings>,
    Partial<WebTimeoutSettings> {
  groupCode: EnvSettingGroupCode
  currentVersion: number
}
