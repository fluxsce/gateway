import { config } from '@/config/config'

/**
 * 返回控制台内嵌帮助（VitePress）站点的路径前缀，与 `docs/.vitepress/config.ts` 中 `resolveDocsBase` 规则一致。
 * 例如 `VITE_BASE_URL=/gatewayweb` 时得到 `/gatewayweb/docs/`。
 */
export function getDocsSitePath(): string {
  let p = (config.baseUrl ?? '').trim()
  if (p === '/' || p === '') {
    return '/docs/'
  }
  if (!p.startsWith('/')) {
    p = `/${p}`
  }
  p = p.replace(/\/+$/, '')
  return `${p}/docs/`
}

/**
 * 当前站点下的帮助文档完整 URL，供 iframe 或新窗口打开。
 */
export function getDocsSiteHref(): string {
  return new URL(getDocsSitePath(), window.location.origin).href
}
