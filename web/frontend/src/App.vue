<template>
  <RsConfigProvider :theme="rsTheme" :locale="rsLocale" control-size="md">
    <RsTooltipProvider>
      <RsToaster />
      <RsLoadingBar ref="loadingBarRef">
        <div class="app-container" id="app-container">
          <router-view />
        </div>
      </RsLoadingBar>
    </RsTooltipProvider>
  </RsConfigProvider>
</template>

<script setup lang="ts">
import { initRequestTools } from '@/api/request'
import type { LocaleType } from '@/locales'
import { useUserStore } from '@/stores/user'
import {
  RsConfigProvider,
  RsLoadingBar,
  RsToaster,
  RsTooltipProvider,
  type RsLoadingBarApi,
  type RsLocale,
  type RsThemeMode,
} from '@/ui'
import { computed, onMounted, useTemplateRef } from 'vue'
import { useI18n } from 'vue-i18n'

const userStore = useUserStore()
const { locale } = useI18n()

const rsTheme = computed<RsThemeMode>(() => userStore.resolvedTheme)

const localeMap: Record<LocaleType, RsLocale> = {
  en: 'en-US',
  'zh-CN': 'zh-CN',
}

const rsLocale = computed<RsLocale>(() => localeMap[locale.value as LocaleType] ?? 'zh-CN')

const loadingBarRef = useTemplateRef<RsLoadingBarApi>('loadingBarRef')

onMounted(() => {
  const loadingBar = loadingBarRef.value
  if (!loadingBar) {
    console.warn('loadingBar 不可用，请求工具初始化失败')
    return
  }
  initRequestTools(loadingBar)
})
</script>

<style scoped>
.app-container {
  height: 100%;
  width: 100%;
}
</style>
