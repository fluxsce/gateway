import { request } from '@/api/request'
import { moduleApiPrefix, requestPathHelper } from '@/api/requestPath'
import { getApiMessage, isApiSuccess, parseJsonData } from '@/utils/format'
import { encryptRsaOaepSha256 } from './rsaOaep'

export type PasswordWrapErrorCode = 'keyUnavailable' | 'encryptFailed'

/**
 * 包装失败。src/api 只抛错误码，不写用户文案；页面用 t('common.passwordWrap.failed')。
 */
export class PasswordWrapError extends Error {
  readonly name = 'PasswordWrapError'
  readonly code: PasswordWrapErrorCode

  constructor(code: PasswordWrapErrorCode, detail = '') {
    super(detail)
    this.code = code
  }
}

/**
 * 是否为口令包装失败（取钥或加密）。
 */
export function isPasswordWrapError(error: unknown): error is PasswordWrapError {
  return error instanceof PasswordWrapError
}

/**
 * 控制台敏感字段包装。开启密文传输后，登录、改密、建用户走这里。
 * 公钥只从 GET /user/password-key 取；未开启时原样返回明文。
 */
const PREFIX = 'enc.rsa1.'
const keyURL = requestPathHelper.join(moduleApiPrefix('user'), 'password-key')

interface PasswordKey {
  enabled?: boolean
  alg?: string
  kid?: string
  publicKey?: string
}

let cached: { kid: string; pem: string; key?: CryptoKey } | null = null

function getSubtle(): SubtleCrypto | undefined {
  return globalThis.crypto?.subtle
}

function pemToSpki(pem: string): ArrayBuffer {
  const b64 = pem
    .replace(/-----BEGIN PUBLIC KEY-----/g, '')
    .replace(/-----END PUBLIC KEY-----/g, '')
    .replace(/\s+/g, '')
  const bin = atob(b64)
  const bytes = new Uint8Array(bin.length)
  for (let i = 0; i < bin.length; i += 1) {
    bytes[i] = bin.charCodeAt(i)
  }
  return bytes.buffer
}

function bytesToB64url(buf: ArrayBuffer | Uint8Array): string {
  const bytes = buf instanceof Uint8Array ? buf : new Uint8Array(buf)
  let bin = ''
  for (let i = 0; i < bytes.length; i += 1) {
    bin += String.fromCharCode(bytes[i])
  }
  return btoa(bin).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/g, '')
}

async function importPublicKey(pem: string): Promise<CryptoKey | undefined> {
  const subtle = getSubtle()
  if (!subtle) {
    return undefined
  }
  return subtle.importKey(
    'spki',
    pemToSpki(pem),
    { name: 'RSA-OAEP', hash: 'SHA-256' },
    false,
    ['encrypt'],
  )
}

async function loadPasswordKey(
  force = false,
): Promise<PasswordKey & { pem?: string; cryptoKey?: CryptoKey }> {
  if (!force && cached) {
    return { enabled: true, kid: cached.kid, pem: cached.pem, cryptoKey: cached.key }
  }
  const response = await request({
    url: keyURL,
    method: 'GET',
  })
  if (!isApiSuccess(response)) {
    throw new PasswordWrapError('keyUnavailable', getApiMessage(response))
  }
  const data = parseJsonData<PasswordKey>(response, {})
  if (!data.enabled || !data.publicKey) {
    cached = null
    return { enabled: false }
  }
  const key = await importPublicKey(data.publicKey)
  cached = { kid: data.kid || '', pem: data.publicKey, key }
  return { ...data, enabled: true, pem: data.publicKey, cryptoKey: key }
}

/**
 * 包装单个口令或密钥。未开启密文传输或空串时原样返回。
 */
export async function wrapPassword(plain: string): Promise<string> {
  const value = plain.trim()
  if (!value) {
    return ''
  }
  const meta = await loadPasswordKey(true)
  if (!meta.enabled || !meta.pem) {
    return value
  }
  const payload = JSON.stringify({
    v: 1,
    t: Math.floor(Date.now() / 1000),
    p: value,
  })
  const encryptOnce = async (pem: string, key?: CryptoKey) => {
    const subtle = getSubtle()
    if (subtle && key) {
      const cipher = await subtle.encrypt(
        { name: 'RSA-OAEP' },
        key,
        new TextEncoder().encode(payload),
      )
      return PREFIX + bytesToB64url(cipher)
    }
    const raw = encryptRsaOaepSha256(pem, new TextEncoder().encode(payload))
    return PREFIX + bytesToB64url(raw)
  }
  try {
    return await encryptOnce(meta.pem, meta.cryptoKey)
  } catch {
    cached = null
    try {
      const retry = await loadPasswordKey(true)
      if (!retry.enabled || !retry.pem) {
        return value
      }
      return await encryptOnce(retry.pem, retry.cryptoKey)
    } catch (retryError) {
      if (isPasswordWrapError(retryError)) {
        throw retryError
      }
      throw new PasswordWrapError('encryptFailed')
    }
  }
}

/**
 * 包装对象上的多个敏感字段，其它字段不动。
 */
export async function wrapPasswordFields<T extends object>(
  data: T,
  fields: readonly (keyof T)[],
): Promise<T> {
  const next = { ...data } as T
  for (const field of fields) {
    const value = next[field]
    if (typeof value !== 'string' || value.trim() === '') {
      continue
    }
    next[field] = (await wrapPassword(value)) as T[typeof field]
  }
  return next
}
