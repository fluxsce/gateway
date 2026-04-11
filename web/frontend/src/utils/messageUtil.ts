/**
 * DOM 挂载用唯一 id 生成（可作 HTML id、CSS # 选择器，仅含前缀与 UUID/时间随机段）。
 *
 * @param prefix - 业务前缀，如 hub0023-gateway-log-query
 */
export function createDomId(prefix: string): string {
  const p = prefix.trim().replace(/\s+/g, '-')
  const suffix =
    typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function'
      ? crypto.randomUUID()
      : `${Date.now()}-${Math.random().toString(36).slice(2, 11)}`
  return `${p}-${suffix}`
}
