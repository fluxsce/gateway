/**
 * DOM 挂载用唯一 id（可作 HTML id、CSS # 选择器）。
 * 不使用 `crypto`，后缀为「时间 base36 + 递增序号 + 短随机」，在常见单页场景下足够唯一且比双段长随机更短。
 *
 * @param prefix - 业务前缀，如 hub0023-gateway-log-query
 */
let _domIdSeq = 0

export function createDomId(prefix: string): string {
  const p = prefix.trim().replace(/\s+/g, '-')
  _domIdSeq = (_domIdSeq + 1) % 0x100000
  const suffix = `${Date.now().toString(36)}${_domIdSeq.toString(36)}${Math.random().toString(36).slice(2, 7)}`
  return `${p}-${suffix}`
}
