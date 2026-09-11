/**
 * RSA-OAEP-SHA256 加密。HTTP 非安全上下文没有 crypto.subtle 时使用，
 * 与 Go crypto/rsa.EncryptOAEP(sha256, empty label) 对齐。
 */

const hLen = 32

class DerReader {
  private offset = 0

  constructor(private readonly bytes: Uint8Array) {}

  private readLength(): number {
    const first = this.bytes[this.offset]
    this.offset += 1
    if (first < 0x80) {
      return first
    }
    const n = first & 0x7f
    let len = 0
    for (let i = 0; i < n; i += 1) {
      len = (len << 8) | this.bytes[this.offset]
      this.offset += 1
    }
    return len
  }

  enter(tag: number): number {
    if (this.bytes[this.offset] !== tag) {
      throw new Error('der tag')
    }
    this.offset += 1
    const len = this.readLength()
    return this.offset + len
  }

  skipTo(end: number): void {
    this.offset = end
  }

  readInteger(): bigint {
    if (this.bytes[this.offset] !== 0x02) {
      throw new Error('der int')
    }
    this.offset += 1
    const len = this.readLength()
    const slice = this.bytes.subarray(this.offset, this.offset + len)
    this.offset += len
    let hex = ''
    for (const b of slice) {
      hex += b.toString(16).padStart(2, '0')
    }
    return BigInt(`0x${hex || '0'}`)
  }

  skipBitStringUnused(): void {
    if (this.bytes[this.offset] !== 0x03) {
      throw new Error('der bit')
    }
    this.offset += 1
    const len = this.readLength()
    // 首字节是未用位数，RSA 公钥为 0
    this.offset += 1
    if (len < 1) {
      throw new Error('der bit empty')
    }
  }
}

function pemToDer(pem: string): Uint8Array {
  const b64 = pem
    .replace(/-----BEGIN PUBLIC KEY-----/g, '')
    .replace(/-----END PUBLIC KEY-----/g, '')
    .replace(/\s+/g, '')
  const bin = atob(b64)
  const bytes = new Uint8Array(bin.length)
  for (let i = 0; i < bin.length; i += 1) {
    bytes[i] = bin.charCodeAt(i)
  }
  return bytes
}

function parseRsaSpki(pem: string): { n: bigint; e: bigint; k: number } {
  const der = pemToDer(pem)
  const r = new DerReader(der)
  const seqEnd = r.enter(0x30)
  const algEnd = r.enter(0x30)
  r.skipTo(algEnd)
  r.skipBitStringUnused()
  r.enter(0x30)
  const n = r.readInteger()
  const e = r.readInteger()
  r.skipTo(seqEnd)
  const k = Math.ceil(n.toString(16).length / 2)
  if (k < 64) {
    throw new Error('rsa modulus too small')
  }
  return { n, e, k }
}

function sha256(data: Uint8Array): Uint8Array {
  const K = [
    0x428a2f98, 0x71374491, 0xb5c0fbcf, 0xe9b5dba5, 0x3956c25b, 0x59f111f1, 0x923f82a4, 0xab1c5ed5,
    0xd807aa98, 0x12835b01, 0x243185be, 0x550c7dc3, 0x72be5d74, 0x80deb1fe, 0x9bdc06a7, 0xc19bf174,
    0xe49b69c1, 0xefbe4786, 0x0fc19dc6, 0x240ca1cc, 0x2de92c6f, 0x4a7484aa, 0x5cb0a9dc, 0x76f988da,
    0x983e5152, 0xa831c66d, 0xb00327c8, 0xbf597fc7, 0xc6e00bf3, 0xd5a79147, 0x06ca6351, 0x14292967,
    0x27b70a85, 0x2e1b2138, 0x4d2c6dfc, 0x53380d13, 0x650a7354, 0x766a0abb, 0x81c2c92e, 0x92722c85,
    0xa2bfe8a1, 0xa81a664b, 0xc24b8b70, 0xc76c51a3, 0xd192e819, 0xd6990624, 0xf40e3585, 0x106aa070,
    0x19a4c116, 0x1e376c08, 0x2748774c, 0x34b0bcb5, 0x391c0cb3, 0x4ed8aa4a, 0x5b9cca4f, 0x682e6ff3,
    0x748f82ee, 0x78a5636f, 0x84c87814, 0x8cc70208, 0x90befffa, 0xa4506ceb, 0xbef9a3f7, 0xc67178f2,
  ]
  const rotr = (x: number, n: number) => (x >>> n) | (x << (32 - n))
  const bytes = new Uint8Array(data.length + 9 + ((64 - ((data.length + 9) % 64)) % 64))
  bytes.set(data)
  bytes[data.length] = 0x80
  const bitLen = data.length * 8
  const view = new DataView(bytes.buffer)
  view.setUint32(bytes.length - 4, bitLen >>> 0)
  view.setUint32(bytes.length - 8, Math.floor(bitLen / 0x100000000) >>> 0)

  let h0 = 0x6a09e667
  let h1 = 0xbb67ae85
  let h2 = 0x3c6ef372
  let h3 = 0xa54ff53a
  let h4 = 0x510e527f
  let h5 = 0x9b05688c
  let h6 = 0x1f83d9ab
  let h7 = 0x5be0cd19
  const w = new Uint32Array(64)
  for (let i = 0; i < bytes.length; i += 64) {
    for (let t = 0; t < 16; t += 1) {
      w[t] = view.getUint32(i + t * 4)
    }
    for (let t = 16; t < 64; t += 1) {
      const s0 = rotr(w[t - 15], 7) ^ rotr(w[t - 15], 18) ^ (w[t - 15] >>> 3)
      const s1 = rotr(w[t - 2], 17) ^ rotr(w[t - 2], 19) ^ (w[t - 2] >>> 10)
      w[t] = (w[t - 16] + s0 + w[t - 7] + s1) >>> 0
    }
    let a = h0
    let b = h1
    let c = h2
    let d = h3
    let e = h4
    let f = h5
    let g = h6
    let h = h7
    for (let t = 0; t < 64; t += 1) {
      const S1 = rotr(e, 6) ^ rotr(e, 11) ^ rotr(e, 25)
      const ch = (e & f) ^ (~e & g)
      const t1 = (h + S1 + ch + K[t] + w[t]) >>> 0
      const S0 = rotr(a, 2) ^ rotr(a, 13) ^ rotr(a, 22)
      const maj = (a & b) ^ (a & c) ^ (b & c)
      const t2 = (S0 + maj) >>> 0
      h = g
      g = f
      f = e
      e = (d + t1) >>> 0
      d = c
      c = b
      b = a
      a = (t1 + t2) >>> 0
    }
    h0 = (h0 + a) >>> 0
    h1 = (h1 + b) >>> 0
    h2 = (h2 + c) >>> 0
    h3 = (h3 + d) >>> 0
    h4 = (h4 + e) >>> 0
    h5 = (h5 + f) >>> 0
    h6 = (h6 + g) >>> 0
    h7 = (h7 + h) >>> 0
  }
  const out = new Uint8Array(32)
  const outView = new DataView(out.buffer)
  outView.setUint32(0, h0)
  outView.setUint32(4, h1)
  outView.setUint32(8, h2)
  outView.setUint32(12, h3)
  outView.setUint32(16, h4)
  outView.setUint32(20, h5)
  outView.setUint32(24, h6)
  outView.setUint32(28, h7)
  return out
}

function mgf1(seed: Uint8Array, maskLen: number): Uint8Array {
  const out = new Uint8Array(maskLen)
  const block = new Uint8Array(seed.length + 4)
  block.set(seed)
  let offset = 0
  let counter = 0
  while (offset < maskLen) {
    block[seed.length] = (counter >>> 24) & 0xff
    block[seed.length + 1] = (counter >>> 16) & 0xff
    block[seed.length + 2] = (counter >>> 8) & 0xff
    block[seed.length + 3] = counter & 0xff
    const digest = sha256(block)
    const take = Math.min(hLen, maskLen - offset)
    out.set(digest.subarray(0, take), offset)
    offset += take
    counter += 1
  }
  return out
}

function xorBytes(a: Uint8Array, b: Uint8Array): Uint8Array {
  const out = new Uint8Array(a.length)
  for (let i = 0; i < a.length; i += 1) {
    out[i] = a[i] ^ b[i]
  }
  return out
}

function bytesToBig(bytes: Uint8Array): bigint {
  let hex = ''
  for (const b of bytes) {
    hex += b.toString(16).padStart(2, '0')
  }
  return BigInt(`0x${hex || '0'}`)
}

function bigToBytes(value: bigint, k: number): Uint8Array {
  const out = new Uint8Array(k)
  let cur = value
  for (let i = k - 1; i >= 0; i -= 1) {
    out[i] = Number(cur & 0xffn)
    cur >>= 8n
  }
  return out
}

function modPow(base: bigint, exp: bigint, mod: bigint): bigint {
  let result = 1n
  let b = base % mod
  let e = exp
  while (e > 0n) {
    if (e & 1n) {
      result = (result * b) % mod
    }
    b = (b * b) % mod
    e >>= 1n
  }
  return result
}

function randomBytes(len: number): Uint8Array {
  const out = new Uint8Array(len)
  if (!globalThis.crypto?.getRandomValues) {
    throw new Error('random unavailable')
  }
  globalThis.crypto.getRandomValues(out)
  return out
}

/**
 * 用 PKIX PEM 公钥做 RSA-OAEP-SHA256 加密，返回 k 字节密文。
 */
export function encryptRsaOaepSha256(pem: string, message: Uint8Array): Uint8Array {
  const { n, e, k } = parseRsaSpki(pem)
  if (message.length > k - 2 * hLen - 2) {
    throw new Error('oaep too long')
  }
  const lHash = sha256(new Uint8Array(0))
  const psLen = k - message.length - 2 * hLen - 2
  const db = new Uint8Array(k - hLen - 1)
  db.set(lHash, 0)
  db[lHash.length + psLen] = 0x01
  db.set(message, lHash.length + psLen + 1)
  const seed = randomBytes(hLen)
  const maskedDB = xorBytes(db, mgf1(seed, db.length))
  const maskedSeed = xorBytes(seed, mgf1(maskedDB, hLen))
  const em = new Uint8Array(k)
  em.set(maskedSeed, 1)
  em.set(maskedDB, 1 + hLen)
  return bigToBytes(modPow(bytesToBig(em), e, n), k)
}

/** 供单测使用的 SHA-256。 */
export function sha256Hex(data: Uint8Array): string {
  return Array.from(sha256(data), (b) => b.toString(16).padStart(2, '0')).join('')
}
