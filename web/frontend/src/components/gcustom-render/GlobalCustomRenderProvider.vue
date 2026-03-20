<template>
  <!-- 与 GDialogProvider 一致：插件用 render() 挂到 body，不在 App 的 NConfigProvider 内，需自包以应用 Naive 主题与语言 -->
  <n-config-provider
    :theme="naiveTheme"
    :theme-overrides="themeOverrides"
    :locale="naiveLocale"
    :date-locale="naiveDateLocale"
  >
    <Teleport to="body">
      <component
        v-if="current"
        :is="current.component"
        v-bind="current.props"
        @update:show="onUpdateShow"
        @success="onSuccess"
      />
    </Teleport>
  </n-config-provider>
</template>

<script setup lang="ts">
import { darkThemeOverrides, lightThemeOverrides } from '@/config/theme'
import type { LocaleType } from '@/locales'
import { getCurrentLocale } from '@/locales'
import { useUserStore } from '@/stores/user'
import type { GlobalThemeOverrides, NDateLocale, NLocale } from 'naive-ui'
import { darkTheme, dateEnUS, dateZhCN, enUS, zhCN } from 'naive-ui'
import { computed } from 'vue'
import { $gRender, current } from './api'

type Theme = 'light' | 'dark'
const userStore = useUserStore()

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

const naiveLocaleMap: Record<LocaleType, NLocale> = { en: enUS, 'zh-CN': zhCN }
const naiveDateLocaleMap: Record<LocaleType, NDateLocale> = { en: dateEnUS, 'zh-CN': dateZhCN }
const naiveLocale = computed(() => naiveLocaleMap[getCurrentLocale()])
const naiveDateLocale = computed(() => naiveDateLocaleMap[getCurrentLocale()])

function onUpdateShow(value: boolean) {
  if (value === false) $gRender.close()
}

function onSuccess(data?: unknown) {
  $gRender.closeWithSuccess(data)
}
</script>
