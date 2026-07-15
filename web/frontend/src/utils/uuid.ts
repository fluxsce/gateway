/**
 * 生成 UUID v4 风格字符串。
 * 优先使用 Math.random，避免依赖安全上下文（纯 HTTP 下 crypto.randomUUID 不可用）。
 * 仅用于 UI 行 id、弹窗实例 id 等非安全用途。
 */
export function randomUUID(): string {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (ch) => {
    const n = (Math.random() * 16) | 0
    const v = ch === 'x' ? n : (n & 0x3) | 0x8
    return v.toString(16)
  })
}
