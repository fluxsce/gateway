<template>
    <div class="theme-switcher">
        <!-- 主题切换按钮 - 点击在浅色/深色主题之间切换 -->
        <n-button text @click="toggleTheme">
            <template #icon>
                <!-- 根据当前主题显示对应的图标 -->
                <n-icon size="18">
                    <moon-outline v-if="isDark" />
                    <sunny-outline v-else />
                </n-icon>
            </template>
            <!-- 显示切换到的目标主题文本 -->
            {{ t(`theme.${isDark ? 'light' : 'dark'}`) }}
        </n-button>
    </div>
</template>

<script lang="ts" setup>
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { useUserStore } from '@/stores/user'
import { MoonOutline, SunnyOutline } from '@vicons/ionicons5'
import { NButton, NIcon } from 'naive-ui'
import { computed } from 'vue'

// 用户 store
const userStore = useUserStore()

// 是否为深色模式
const isDark = computed(() => {
  const theme = userStore.theme
  if (theme === 'system') {
    return window.matchMedia('(prefers-color-scheme: dark)').matches
  }
  return theme === 'dark'
})

// 切换主题
const toggleTheme = () => {
  const currentTheme = userStore.theme
  const nextTheme = currentTheme === 'light' ? 'dark' : 'light'
  userStore.update({ theme: nextTheme }, { persistUserData: false })
}

// 使用模块化国际化
const { t } = useModuleI18n('common')
</script>