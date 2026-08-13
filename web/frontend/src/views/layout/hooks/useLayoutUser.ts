/**
 * 布局用户相关逻辑
 * 处理用户菜单、登出等操作
 */
import { config } from '@/config/config'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { store } from '@/stores'
import { useGlobalStore } from '@/stores/global'
import { computed } from 'vue'

export function useLayoutUser() {
  const { t: tCommon } = useModuleI18n('common')
  const globalStore = useGlobalStore()

  // 用户下拉菜单选项（GDropdown 当前不渲染 option.icon，仅保留 key/label）
  const userMenuOptions = computed(() => [
    {
      key: 'settings',
      label: tCommon('user.settings'),
    },
    {
      type: 'divider',
      key: 'd1',
    },
    {
      key: 'logout',
      label: tCommon('user.logout'),
    },
  ])

  // 处理用户菜单操作
  const handleUserAction = (key: string | number) => {
    switch (String(key)) {
      case 'settings':
        // 与侧栏一致：先 upsert 页签，由 MainLayoutContent 监听 activeTabId 再 router.push
        globalStore.upsertLayoutTab('/settings', tCommon('user.settings'), 'settings')
        break
      case 'logout':
        store.user.clearUserInfo()
        const baseUrl = config.baseUrl.endsWith('/')
          ? config.baseUrl.slice(0, -1)
          : config.baseUrl
        window.location.href = baseUrl || '/'
        break
    }
  }

  return {
    userMenuOptions,
    handleUserAction,
  }
}
