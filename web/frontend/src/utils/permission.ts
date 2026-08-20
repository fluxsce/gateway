import { store } from '@/stores'

/**
 * 表单写权限判断。
 *
 * - 列表页 `moduleId` 为模块码（如 hub0002）：create 认 `:add` / `:create`，edit 认 `:edit`
 * - 子弹窗 `moduleId` 已是入口码（如 hub0021:corsConfig）：优先认子码，没有子码时入口码即写权限
 * - 未传 moduleId 时不在此拦截，由调用方决定（与 Toolbar 缺 moduleId 放行一致）
 */
export function hasFormWritePermission(
  moduleId: string | undefined,
  mode: 'create' | 'edit' | 'view' | string,
  action?: string,
): boolean {
  if (!moduleId || mode === 'view') {
    return false
  }
  if (action) {
    return store.user.hasPermission(`${moduleId}:${action}`) || store.user.hasPermission(moduleId)
  }
  if (mode === 'create') {
    if (store.user.hasPermission(`${moduleId}:add`) || store.user.hasPermission(`${moduleId}:create`)) {
      return true
    }
  } else if (mode === 'edit') {
    if (store.user.hasPermission(`${moduleId}:edit`)) {
      return true
    }
  }
  return moduleId.includes(':') && store.user.hasPermission(moduleId)
}
