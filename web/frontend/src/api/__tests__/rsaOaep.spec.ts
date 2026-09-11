import { describe, expect, it } from 'vitest'
import { encryptRsaOaepSha256, sha256Hex } from '../rsaOaep'

const samplePem = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA6/dWhcQKAR3r+0CJVkl8
iNB6rff6JQeKWuw2lfWGTb19tjqxj/MvtlULY6ULJLySA0hUAjq0+i8TA8zl7YzM
kcEgi0IbBm6A5jiyIIUA5UvfBKGaYIiH0ZP5l9xmbLbt1XUCLuU4zv82NEjG2PrB
6qcb7ApyOqxQbWMw+JvJJKjFlXqLDotdikLzb7cfnbUY/C9vA4mlEAhRwPtodj5f
CJ34KxtoKQKC7LEolV8DrlBS9xvG8owUkj3hpSmErjGOO1t4CmwnZOqUjPZraqOi
PGATVctcQvtp1NWD8e5nglxQHJKlimSiElMtRD0FfnsrOnaT55t/kEcKAVJGEkH2
CwIDAQAB
-----END PUBLIC KEY-----`

const emptySha256 = 'e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855'
const abcSha256 = 'ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad'

describe('rsaOaep', () => {
  it('sha256 matches empty and abc', () => {
    expect(sha256Hex(new Uint8Array())).toBe(emptySha256)
    expect(sha256Hex(new TextEncoder().encode('abc'))).toBe(abcSha256)
  })

  it('rejects unusable pem', () => {
    expect(() => encryptRsaOaepSha256('not-a-key', new Uint8Array([1]))).toThrow()
  })

  it('encrypts to 256-byte RSA-2048 ciphertext', () => {
    const a = encryptRsaOaepSha256(samplePem, new TextEncoder().encode('{"v":1,"t":1,"p":"x"}'))
    const b = encryptRsaOaepSha256(samplePem, new TextEncoder().encode('{"v":1,"t":1,"p":"x"}'))
    expect(a.length).toBe(256)
    expect(b.length).toBe(256)
    expect(Buffer.from(a).equals(Buffer.from(b))).toBe(false)
  })
})
