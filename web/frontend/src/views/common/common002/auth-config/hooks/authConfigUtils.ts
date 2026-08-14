/**
 * 认证配置域工具：按认证类型裁剪/补齐 authConfig，不改变后端 JSON 字段名与取值语义。
 */

import { getByNamePath, setByNamePath } from '@/ui'

export type AuthType = 'JWT' | 'API_KEY' | 'OAUTH2' | 'BASIC' | 'BEARER_TOKEN'

const AUTH_CONFIG_PREFIX = 'authConfig.'

/** 各认证类型对应的 authConfig 字段（camelCase，与表单 authConfig.xxx、后端 JSON 键一致） */
export const AUTH_CONFIG_FIELDS: Record<AuthType, string[]> = {
  JWT: [
    'secret',
    'algorithm',
    'issuer',
    'expiration',
    'verifyExpiration',
    'verifyIssuer',
    'refreshWindow',
    'includeInResponse',
    'responseHeaderName',
    'publicKey',
  ],
  API_KEY: ['in', 'param_name', 'key'],
  OAUTH2: ['tokenEndpoint', 'introspectEndpoint', 'clientID', 'clientSecret', 'scope'],
  BASIC: ['username', 'password'],
  BEARER_TOKEN: ['token'],
}

/** JWT 表单默认值（与历史前端保存值一致，勿改成后端运行时默认以免改写已存配置） */
export const JWT_AUTH_CONFIG_DEFAULTS = {
  algorithm: 'HS256',
  expiration: 3600,
  verifyExpiration: true,
  verifyIssuer: false,
  refreshWindow: 300,
  includeInResponse: false,
  responseHeaderName: 'X-Auth-Token',
} as const

/** API Key 表单默认值 */
export const API_KEY_AUTH_CONFIG_DEFAULTS = {
  in: 'header',
  param_name: 'X-API-Key',
} as const

export const AUTH_CONFIG_DEFAULTS: Partial<Record<AuthType, Record<string, unknown>>> = {
  JWT: { ...JWT_AUTH_CONFIG_DEFAULTS },
  API_KEY: { ...API_KEY_AUTH_CONFIG_DEFAULTS },
}

const JWT_HMAC_ALGORITHMS = new Set(['HS256', 'HS384', 'HS512'])
const JWT_RSA_ALGORITHMS = new Set(['RS256', 'RS384', 'RS512'])

function authConfigFieldKey(subField: string): string {
  return `${AUTH_CONFIG_PREFIX}${subField}`
}

function authConfigGet(formData: Record<string, any>, subField: string): unknown {
  const nested = getByNamePath(formData, authConfigFieldKey(subField))
  if (nested !== undefined) return nested
  return formData[authConfigFieldKey(subField)]
}

function authConfigSet(formData: Record<string, any>, subField: string, value: unknown): void {
  setByNamePath(formData, authConfigFieldKey(subField), value)
}

function normalizeJwtAlgorithm(algorithm?: unknown): string {
  return typeof algorithm === 'string' ? algorithm.trim().toUpperCase() : ''
}

function isEmptyAuthValue(value: unknown): boolean {
  return value === undefined || value === null || value === ''
}

/** 是否为 HMAC 对称签名算法（使用 secret 字符串作为密钥） */
export function isJwtHmacAlgorithm(algorithm?: unknown): boolean {
  return JWT_HMAC_ALGORITHMS.has(normalizeJwtAlgorithm(algorithm))
}

/** 是否为 RSA 非对称签名算法（使用 PEM 公钥验签，不需要 secret） */
export function isJwtRsaAlgorithm(algorithm?: unknown): boolean {
  return JWT_RSA_ALGORITHMS.has(normalizeJwtAlgorithm(algorithm))
}

/**
 * 按 JWT 算法裁剪提交对象：HMAC 去掉 publicKey，RSA 去掉 secret。
 * 未知算法不裁剪，避免误删已存密钥。
 */
export function pruneJwtKeysByAlgorithm(authConfigObj: Record<string, any>): void {
  const algorithm = authConfigObj.algorithm
  if (isJwtRsaAlgorithm(algorithm)) {
    delete authConfigObj.secret
  } else if (isJwtHmacAlgorithm(algorithm)) {
    delete authConfigObj.publicKey
  }
}

/**
 * 切换 JWT 签名算法时，清除与当前算法不匹配的表单密钥字段。
 */
export function clearJwtKeysOnAlgorithmChange(formData: Record<string, any>, algorithm: string): void {
  if (isJwtRsaAlgorithm(algorithm)) {
    authConfigSet(formData, 'secret', undefined)
  } else if (isJwtHmacAlgorithm(algorithm)) {
    authConfigSet(formData, 'publicKey', undefined)
  }
}

/**
 * 按当前认证类型从表单 authConfig 嵌套对象构建提交对象，仅保留本类型字段。
 * 已填写的值原样带出，不在此处补默认值，避免编辑保存时改写历史 JSON。
 */
export function buildAuthConfigForType(formData: Record<string, any>): Record<string, any> {
  const authType = formData.authType as AuthType
  const allowedFields = AUTH_CONFIG_FIELDS[authType] || []
  const authConfigObj: Record<string, any> = {}

  for (const key of allowedFields) {
    const value = authConfigGet(formData, key)
    if (value !== undefined) {
      authConfigObj[key] = value
    }
  }

  if (authType === 'JWT') {
    pruneJwtKeysByAlgorithm(authConfigObj)
  }

  if (authType === 'API_KEY' && typeof authConfigObj.key === 'string') {
    authConfigObj.key = authConfigObj.key.trim()
  }

  if (authType === 'BEARER_TOKEN' && typeof authConfigObj.token === 'string') {
    authConfigObj.token = authConfigObj.token.trim()
  }

  return authConfigObj
}

/**
 * 仅为空字段补默认值，已有值（含编辑回填）一律保留。
 */
export function applyAuthConfigFieldDefaults(formData: Record<string, any>, authType: string): void {
  const defaults = AUTH_CONFIG_DEFAULTS[authType as AuthType]
  if (!defaults) {
    return
  }
  for (const [key, defaultVal] of Object.entries(defaults)) {
    if (isEmptyAuthValue(authConfigGet(formData, key))) {
      authConfigSet(formData, key, defaultVal)
    }
  }
}

/**
 * 切换认证类型时，清除不属于新类型的 authConfig 字段。
 */
export function clearIrrelevantAuthConfigFields(formData: Record<string, any>, authType: string): void {
  const allowed = new Set(AUTH_CONFIG_FIELDS[authType as AuthType] || [])
  const nested = formData.authConfig
  if (nested && typeof nested === 'object' && !Array.isArray(nested)) {
    Object.keys(nested).forEach((subKey) => {
      if (!allowed.has(subKey)) delete nested[subKey]
    })
  }
  Object.keys(formData).forEach((key) => {
    if (key.startsWith(AUTH_CONFIG_PREFIX)) {
      const subKey = key.slice(AUTH_CONFIG_PREFIX.length)
      if (!allowed.has(subKey)) {
        delete formData[key]
      }
    }
  })
}

/**
 * 切换认证类型：先清旧类型字段，再补新类型空缺默认值（不覆盖已有值）。
 */
export function switchAuthConfigType(formData: Record<string, any>, authType: string): void {
  clearIrrelevantAuthConfigFields(formData, authType)
  applyAuthConfigFieldDefaults(formData, authType)
}

/** 是否为已启用且强制认证的 OAuth2 配置（网关远端校验尚未实现） */
export function isOAuth2ActiveRequired(formData: Record<string, any>): boolean {
  return (
    formData.authType === 'OAUTH2' &&
    formData.authStrategy === 'REQUIRED' &&
    formData.activeFlag === 'Y'
  )
}

/**
 * 校验 API Key 表单：参数名与密钥值均必填。
 * @returns 错误消息；通过时返回 null
 */
export function validateApiKeyFormData(formData: Record<string, any>): string | null {
  if (formData.authType !== 'API_KEY') {
    return null
  }

  const paramName = String(authConfigGet(formData, 'param_name') || '').trim()
  if (!paramName) {
    return '请填写 API Key 参数名称'
  }

  const key = String(authConfigGet(formData, 'key') || '').trim()
  if (!key) {
    return '请填写 API Key 密钥值'
  }

  return null
}

/**
 * 校验 Bearer Token 表单：token 必填。
 * @returns 错误消息；通过时返回 null
 */
export function validateBearerTokenFormData(formData: Record<string, any>): string | null {
  if (formData.authType !== 'BEARER_TOKEN') {
    return null
  }

  const token = String(authConfigGet(formData, 'token') || '').trim()
  if (!token) {
    return '请填写 Bearer Token 值'
  }

  return null
}

/** 从历史 keys/validKeys 取第一个有效密钥，读不到则返回空串 */
function firstLegacyApiKey(legacyKeys: unknown): string {
  if (!Array.isArray(legacyKeys) || legacyKeys.length === 0) {
    return ''
  }
  const first = legacyKeys[0]
  if (typeof first === 'string') {
    return first.trim()
  }
  if (first && typeof first === 'object' && 'value' in first) {
    return String((first as { value?: string }).value || '').trim()
  }
  return ''
}

/**
 * 加载时将后端/历史 authConfig 格式映射为表单字段（API Key，对齐 APIKeyConfig）。
 * 兼容历史 keys/validKeys 数组（取首项）及 camelCase 字段。
 */
export function normalizeApiKeyFormFields(
  formData: Record<string, any>,
  authConfigObj: Record<string, any>
): void {
  const paramName = authConfigObj.param_name ?? authConfigObj.keyName
  if (paramName) {
    authConfigSet(formData, 'param_name', paramName)
  }

  const location = authConfigObj.in ?? authConfigObj.keyLocation
  if (location) {
    authConfigSet(formData, 'in', location)
  }

  if (typeof authConfigObj.key === 'string' && authConfigObj.key.trim()) {
    authConfigSet(formData, 'key', authConfigObj.key.trim())
  } else {
    const legacyKey = firstLegacyApiKey(authConfigObj.keys ?? authConfigObj.validKeys)
    if (legacyKey) {
      authConfigSet(formData, 'key', legacyKey)
    }
  }

  const drop = ['keyName', 'keyLocation', 'validKeys', 'keys', 'isPrefixMatch', 'is_prefix_match']
  drop.forEach((sub) => {
    if (formData.authConfig && typeof formData.authConfig === 'object') {
      delete formData.authConfig[sub]
    }
    delete formData[authConfigFieldKey(sub)]
  })
}
