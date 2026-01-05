<template>
  <NConfigProvider 
    :theme="naiveTheme" 
    :theme-overrides="isDark ? darkThemeOverrides : lightThemeOverrides"
    :locale="naiveLocale"
    :date-locale="naiveDateLocale"
    :hljs="hljsInstance"
  >
    <div 
      class="app-container" 
      :class="{ 'dark-theme': isDark, 'light-theme': !isDark }" 
      :data-theme="isDark ? 'dark' : 'light'"
    >
      <NLoadingBarProvider>
        <NMessageProvider>
          <NDialogProvider>
            <RequestInitializer />
            <RouterView />
          </NDialogProvider>
        </NMessageProvider>
      </NLoadingBarProvider>
    </div>
  </NConfigProvider>
</template>

<script setup lang="ts">
import RequestInitializer from '@/components/RequestInitializer.vue'
import { darkThemeOverrides, lightThemeOverrides } from '@/config/theme'
import { getCurrentLocale } from '@/locales'
import { useUserStore } from '@/stores/user'
import hljs from '@/utils/highlight'
import {
  darkTheme,
  dateEnUS,
  dateZhCN,
  enUS,
  NConfigProvider,
  NDialogProvider,
  NLoadingBarProvider,
  NMessageProvider,
  zhCN
} from 'naive-ui'
import type { Hljs } from 'naive-ui/es/_mixins'
import { computed, watchEffect } from 'vue'

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

// Naive UI 主题对象
const naiveTheme = computed(() => {
  return isDark.value ? darkTheme : null
})

// 将 hljs 实例转换为 naive-ui 期望的类型
const hljsInstance: Hljs = {
  highlight: hljs.highlight.bind(hljs),
  getLanguage: hljs.getLanguage.bind(hljs)
}

// 多语言配置
const naiveLocale = computed(() => {
  const currentLang = getCurrentLocale()
  return currentLang === 'en' ? enUS : zhCN
})

const naiveDateLocale = computed(() => {
  const currentLang = getCurrentLocale()
  return currentLang === 'en' ? dateEnUS : dateZhCN
})

// 同步主题到HTML根元素
watchEffect(() => {
  const html = document.documentElement
  const theme = isDark.value ? 'dark' : 'light'
  
  // 设置 data-theme 属性
  html.setAttribute('data-theme', theme)
  
  // 设置 class
  if (isDark.value) {
    html.classList.add('dark-theme')
    html.classList.remove('light-theme')
  } else {
    html.classList.add('light-theme')
    html.classList.remove('dark-theme')
  }
})
</script>

<style>
.app-container {
  height: 100%;
  width: 100%;
}
</style>
