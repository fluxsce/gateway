/**
 * 认证配置工具：按认证类型维护 authConfig 字段白名单，避免切换类型后残留旧参数。
 */

export type AuthType = 'JWT' | 'API_KEY' | 'OAUTH2' | 'BASIC' | 'BEARER_TOKEN'

/** 各认证类型对应的 authConfig 字段（camelCase，与表单 authConfig.xxx 一致） */
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

const JWT_HMAC_ALGORITHMS = ['HS256', 'HS384', 'HS512'] as const
const JWT_RSA_ALGORITHMS = ['RS256', 'RS384', 'RS512'] as const

/** 是否为 HMAC 对称签名算法（使用 secret 字符串作为密钥） */
export function isJwtHmacAlgorithm(algorithm?: string): boolean {
  return JWT_HMAC_ALGORITHMS.includes((algorithm || '').toUpperCase() as typeof JWT_HMAC_ALGORITHMS[number])
}

/** 是否为 RSA 非对称签名算法（使用 PEM 公钥验签，不需要 secret） */
export function isJwtRsaAlgorithm(algorithm?: string): boolean {
  return JWT_RSA_ALGORITHMS.includes((algorithm || '').toUpperCase() as typeof JWT_RSA_ALGORITHMS[number])
}

/**
 * 按 JWT 算法裁剪密钥字段：HMAC 保留 secret，RSA 保留 publicKey。
 */
export function pruneJwtKeysByAlgorithm(authConfigObj: Record<string, any>): void {
  const algorithm = authConfigObj.algorithm as string | undefined
  if (isJwtRsaAlgorithm(algorithm)) {
    delete authConfigObj.secret
  } else if (isJwtHmacAlgorithm(algorithm)) {
    delete authConfigObj.publicKey
  }
}

/**
 * 切换 JWT 签名算法时，清除与当前算法不匹配的密钥字段。
 */
export function clearJwtKeysOnAlgorithmChange(formData: Record<string, any>, algorithm: string): void {
  if (isJwtRsaAlgorithm(algorithm)) {
    delete formData['authConfig.secret']
  } else if (isJwtHmacAlgorithm(algorithm)) {
    delete formData['authConfig.publicKey']
  }
}

/**
 * 按当前认证类型从表单点号字段构建 authConfig 对象，仅保留本类型字段。
 */
export function buildAuthConfigForType(formData: Record<string, any>): Record<string, any> {
  const authType = formData.authType as AuthType
  const allowedFields = AUTH_CONFIG_FIELDS[authType] || []
  const authConfigObj: Record<string, any> = {}

  for (const key of allowedFields) {
    const dotKey = `authConfig.${key}`
    if (formData[dotKey] !== undefined) {
      authConfigObj[key] = formData[dotKey]
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
 * 切换认证类型时，清除不属于新类型的 authConfig.xxx 点号字段。
 */
export function clearIrrelevantAuthConfigFields(formData: Record<string, any>, authType: string): void {
  const allowed = new Set(AUTH_CONFIG_FIELDS[authType as AuthType] || [])
  Object.keys(formData).forEach((key) => {
    if (key.startsWith('authConfig.')) {
      const subKey = key.replace('authConfig.', '')
      if (!allowed.has(subKey)) {
        delete formData[key]
      }
    }
  })
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

  const paramName = String(formData['authConfig.param_name'] || '').trim()
  if (!paramName) {
    return '请填写 API Key 参数名称'
  }

  const key = String(formData['authConfig.key'] || '').trim()
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

  const token = String(formData['authConfig.token'] || '').trim()
  if (!token) {
    return '请填写 Bearer Token 值'
  }

  return null
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
    formData['authConfig.param_name'] = paramName
  }

  const location = authConfigObj.in ?? authConfigObj.keyLocation
  if (location) {
    formData['authConfig.in'] = location
  }

  if (typeof authConfigObj.key === 'string' && authConfigObj.key.trim()) {
    formData['authConfig.key'] = authConfigObj.key.trim()
  } else {
    const legacyKeys = authConfigObj.keys ?? authConfigObj.validKeys
    if (Array.isArray(legacyKeys) && legacyKeys.length > 0) {
      const first = legacyKeys[0]
      if (typeof first === 'string' && first.trim()) {
        formData['authConfig.key'] = first.trim()
      } else if (first && typeof first === 'object' && 'value' in first) {
        const value = String((first as { value?: string }).value || '').trim()
        if (value) {
          formData['authConfig.key'] = value
        }
      }
    }
  }

  delete formData['authConfig.keyName']
  delete formData['authConfig.keyLocation']
  delete formData['authConfig.validKeys']
  delete formData['authConfig.keys']
  delete formData['authConfig.isPrefixMatch']
  delete formData['authConfig.is_prefix_match']
}
