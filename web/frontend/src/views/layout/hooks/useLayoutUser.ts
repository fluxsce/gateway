/**
 * 布局用户相关逻辑
 * 处理用户菜单、登出等操作
 */
import { config } from '@/config/config'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { store } from '@/stores'
import { useGlobalStore } from '@/stores/global'
import type { RsDropdownItem } from '@/ui'
import { computed } from 'vue'

export function useLayoutUser() {
  const { t: tCommon } = useModuleI18n('common')
  const globalStore = useGlobalStore()

  const userMenuItems = computed<RsDropdownItem[]>(() => [
    {
      value: 'settings',
      label: tCommon('user.settings'),
      icon: 'settings',
    },
    {
      value: 'logout',
      label: tCommon('user.logout'),
      icon: 'log-out',
    },
  ])

  const handleUserAction = (value: string) => {
    switch (value) {
      case 'settings':
        // 与侧栏一致：先 upsert 页签，由 MainLayoutContent 监听 activeTabId 再 router.push
        globalStore.upsertLayoutTab('/settings', tCommon('user.settings'), 'settings')
        break
      case 'logout': {
        // 只清持久化，不 $reset，避免跳转前顶栏闪成「游客」
        store.user.clearPersistedSession()
        const baseUrl = config.baseUrl.endsWith('/')
          ? config.baseUrl.slice(0, -1)
          : config.baseUrl
        window.location.href = baseUrl || '/'
        break
      }
    }
  }

  return {
    userMenuItems,
    handleUserAction,
  }
}
