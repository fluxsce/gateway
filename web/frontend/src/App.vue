<template>
  <n-config-provider
    :theme="naiveTheme"
    :theme-overrides="themeOverrides"
    :locale="naiveLocale"
    :date-locale="naiveDateLocale"
    :hljs="hljsInstance"
  >
    <div class="app-container" id="app-container">
      <n-loading-bar-provider>
        <n-message-provider>
          <n-dialog-provider>
            <RequestInitializer />
            <router-view />
          </n-dialog-provider>
        </n-message-provider>
      </n-loading-bar-provider>
    </div>
  </n-config-provider>
</template>

<script setup lang="ts">
import RequestInitializer from '@/components/RequestInitializer.vue'
import { darkThemeOverrides, lightThemeOverrides } from '@/config/theme'
import type { LocaleType } from '@/locales'
import { getCurrentLocale } from '@/locales'
import { useUserStore } from '@/stores/user'
import hljs from '@/utils/highlight'
import {
    darkTheme,
    dateEnUS,
    dateZhCN,
    enUS,
    zhCN,
    type GlobalThemeOverrides,
    type NDateLocale,
    type NLocale,
} from 'naive-ui'
import type { Hljs } from 'naive-ui/es/_mixins'
import { computed } from 'vue'

type Theme = 'light' | 'dark'

const userStore = useUserStore()

// 主题映射配置（与  一致：store.resolvedTheme / store.isDark，一处维护）
const naiveThemeMap: Record<Theme, typeof darkTheme | null> = {
  light: null,
  dark: darkTheme,
}

const themeOverridesMap: Record<Theme, GlobalThemeOverrides> = {
  light: lightThemeOverrides,
  dark: darkThemeOverrides,
}

const naiveTheme = computed(() => naiveThemeMap[userStore.resolvedTheme])
const themeOverrides = computed(() => themeOverridesMap[userStore.resolvedTheme])

// 语言映射配置（参考 ）
const naiveLocaleMap: Record<LocaleType, NLocale> = {
  en: enUS,
  'zh-CN': zhCN,
}

const naiveDateLocaleMap: Record<LocaleType, NDateLocale> = {
  en: dateEnUS,
  'zh-CN': dateZhCN,
}

const naiveLocale = computed(() => naiveLocaleMap[getCurrentLocale()])
const naiveDateLocale = computed(() => naiveDateLocaleMap[getCurrentLocale()])

const hljsInstance: Hljs = {
  highlight: hljs.highlight.bind(hljs),
  getLanguage: hljs.getLanguage.bind(hljs),
}
</script>

<style scoped>
.app-container {
  height: 100%;
  width: 100%;
}
</style>
