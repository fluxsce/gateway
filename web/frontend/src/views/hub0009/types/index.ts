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
  cipherEnabled: boolean
  kid: string
  publicKey: string
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

/** 系统健康组件探测状态。skipped 表示进程未装配该依赖。 */
export type HealthStatus = 'ok' | 'error' | 'skipped' | 'degraded'

/** 系统健康组件类型。 */
export type HealthKind = 'process' | 'sql' | 'clickhouse' | 'mongo' | 'cache'

/** 系统日志集合上一条期望索引的对照结果。 */
export interface HealthStoreIndex {
  name: string
  keys: string
  present: boolean
}

/** 一张表或一个集合的存在性与索引对照。 */
export interface HealthStoreObject {
  name: string
  kind: 'table' | 'collection'
  exists: boolean
  indexes?: HealthStoreIndex[]
}

/** 单个依赖的探活结果。 */
export interface HealthItem {
  name: string
  kind: HealthKind
  driver: string
  status: HealthStatus
  latencyMs: number
  message: string
  enabled: boolean
  database: string
  objects?: HealthStoreObject[]
}

/** 环境设置「系统健康」页一次探活的汇总。 */
export interface SystemHealth {
  status: HealthStatus
  checkedAt: number
  items: HealthItem[]
}

export type EnvSettingGroupCode = 'retention' | 'retentionJob' | 'webTimeout'

export interface SaveEnvSettingPayload
  extends Partial<RetentionSettings>,
    Partial<RetentionJobSettings>,
    Partial<WebTimeoutSettings> {
  groupCode: EnvSettingGroupCode
  currentVersion: number
}
