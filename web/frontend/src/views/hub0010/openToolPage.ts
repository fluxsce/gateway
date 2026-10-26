import { useGlobalStore } from '@/stores/global'

/** 每个工具对应一条隐藏路由。打开时按路由按需加载页面组件。 */
const TOOL_PAGES: Record<string, { path: string; icon: string }> = {
  'json-formatter': { path: '/system/tools/json-formatter', icon: 'braces' },
  timestamp: { path: '/system/tools/timestamp', icon: 'clock' },
  base64: { path: '/system/tools/base64', icon: 'binary' },
  'jwt-decoder': { path: '/system/tools/jwt-decoder', icon: 'key-round' },
}

/**
 * 在主布局里新开一个页签。页签切换会同步路由，从而加载对应工具组件。
 */
export function openToolPage(toolId: string, title: string): boolean {
  const page = TOOL_PAGES[toolId]
  if (!page) return false
  useGlobalStore().upsertLayoutTab(page.path, title, page.icon)
  return true
}
