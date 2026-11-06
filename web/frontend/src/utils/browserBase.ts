/**
 * 浏览器上的控制台路径。
 * 进程内固定是 web.frontend.prefix（如 /gatewayweb）。
 * 同一个域名上用 /test、/pro 区分环境时，地址栏是 /test/gatewayweb，打包时不写死外层前缀。
 */

declare const __GW_MOUNT_PREFIX__: string

/**
 * 进程内控制台路径，由 Vite 按 configs/web.yaml 的 web.frontend.prefix 注入。
 * @returns 规范化后的路径，例如 `/gatewayweb`；根路径为 `/`
 */
export function gatewayMountPrefix(): string {
  try {
    if (typeof __GW_MOUNT_PREFIX__ === 'string' && __GW_MOUNT_PREFIX__.trim() !== '') {
      return normalizeBase(__GW_MOUNT_PREFIX__)
    }
  } catch {
    // 单测或未经过 Vite define 时走默认值
  }
  return '/gatewayweb'
}

/**
 * 从地址栏取出浏览器上的控制台路径。
 * `/test/gatewayweb/login` 且 mount 为 `/gatewayweb` 时得到 `/test/gatewayweb`。
 * mount 自带前导 `/`，命中处就是段起点，只要求后面是段结束。
 * 找不到 mount 时回落 mount 本身。
 * @param mount - 进程内前缀
 * @param pathname - location.pathname
 * @returns 不含尾斜杠的浏览器路径
 */
export function detectBrowserBase(mount: string, pathname: string): string {
  const marker = normalizeBase(mount)
  if (marker === '/') return '/'
  const path = pathname || '/'
  let from = 0
  while (from < path.length) {
    const index = path.indexOf(marker, from)
    if (index < 0) break
    const next = path.charAt(index + marker.length)
    if (next === '' || next === '/') {
      return path.slice(0, index + marker.length)
    }
    from = index + marker.length
  }
  return marker
}

/**
 * 当前页面对应的浏览器控制台路径。
 * 生产页头脚本会先写入 window.__GW_BROWSER_BASE__；没有该值时按地址栏现算。
 * @returns 不含尾斜杠的路径
 */
export function browserBase(): string {
  if (typeof window !== 'undefined') {
    const preset = (window as Window & { __GW_BROWSER_BASE__?: string }).__GW_BROWSER_BASE__
    if (typeof preset === 'string' && preset.startsWith('/')) {
      return normalizeBase(preset)
    }
    return detectBrowserBase(gatewayMountPrefix(), window.location.pathname)
  }
  return gatewayMountPrefix()
}

/**
 * 去掉尾斜杠，根路径保持 `/`。
 * @param raw - 原始路径
 * @returns 规范化路径
 */
function normalizeBase(raw: string): string {
  let value = raw.trim()
  if (value === '' || value === '/') return '/'
  if (!value.startsWith('/')) value = `/${value}`
  return value.replace(/\/+$/, '')
}
