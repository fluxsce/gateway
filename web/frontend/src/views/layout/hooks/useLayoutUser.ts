/**
 * 布局用户相关逻辑
 * 处理用户菜单、登出等操作
 */
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { store } from '@/stores'
import { LogOutOutline, SettingsOutline } from '@vicons/ionicons5'
import { NIcon } from 'naive-ui'
import { computed, h } from 'vue'
import { useRouter } from 'vue-router'

export function useLayoutUser() {
  const { t: tCommon } = useModuleI18n('common')
  const router = useRouter()

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
  const handleUserAction = (key: string) => {
    switch (key) {
      case 'settings':
        router.push('/settings')
        break
      case 'logout':
        // 执行登出操作
        store.user.clearUserInfo()
        // 使用 location.href 确保完全重置页面状态和清除 cookie
        window.location.href = '/'
        break
    }
  }

  return {
    userMenuOptions,
    handleUserAction,
  }
}
