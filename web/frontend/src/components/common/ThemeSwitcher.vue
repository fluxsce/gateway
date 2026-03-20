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

// 用户 store（与  AppHeader 一致：isDark + toggle 只改 light/dark）
const userStore = useUserStore()

const isDark = computed(() => userStore.isDark)

const toggleTheme = () => {
  const nextTheme = userStore.isDark ? 'light' : 'dark'
  userStore.update({ theme: nextTheme }, { persistUserData: false })
}

// 使用模块化国际化
const { t } = useModuleI18n('common')
</script>