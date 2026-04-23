/**
 * 布局用户相关逻辑
 * 处理用户菜单、登出等操作
 */
import { config } from '@/config/config'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { store } from '@/stores'
import { useGlobalStore } from '@/stores/global'
import { LogOutOutline, SettingsOutline } from '@vicons/ionicons5'
import { NIcon } from 'naive-ui'
import { computed, h } from 'vue'

export function useLayoutUser() {
  const { t: tCommon } = useModuleI18n('common')
  const globalStore = useGlobalStore()

  // 用户下拉菜单选项
  const userMenuOptions = computed(() => [
    {
      key: 'settings',
      label: tCommon('user.settings'),
      icon: () => h(NIcon, null, { default: () => h(SettingsOutline) }),
    },
    {
      type: 'divider',
      key: 'd1',
    },
    {
      key: 'logout',
      label: tCommon('user.logout'),
      icon: () => h(NIcon, null, { default: () => h(LogOutOutline) }),
    },
  ])

  // 处理用户菜单操作
  const handleUserAction = (key: string | number) => {
    switch (String(key)) {
      case 'settings':
        // 与侧栏一致：先 upsert 页签，由 MainLayoutContent 监听 activeTabId 再 router.push，避免 URL 与页签脱节
        globalStore.upsertLayoutTab('/settings', tCommon('user.settings'), 'SettingsOutline')
        break
      case 'logout':
        // 执行登出操作
        store.user.clearUserInfo()
        // 使用 location.href 确保完全重置页面状态和清除 cookie
        // 使用配置中的 baseUrl，确保包含 VITE_BASE_URL
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
