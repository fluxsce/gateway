<template>
  <RsConfigProvider :theme="rsTheme" :locale="rsLocale">
    <Teleport to="body">
      <component
        v-if="current"
        :is="current.component"
        v-bind="current.props"
        @update:show="onUpdateShow"
        @update:open="onUpdateShow"
        @success="onSuccess"
      />
    </Teleport>
  </RsConfigProvider>
</template>

<script setup lang="ts">
import { getCurrentLocale, type LocaleType } from '@/locales'
import { useUserStore } from '@/stores/user'
import { RsConfigProvider, type RsLocale, type RsThemeMode } from '@/ui'
import { computed } from 'vue'
import { $gRender, current } from './api'

const userStore = useUserStore()
const rsTheme = computed<RsThemeMode>(() => userStore.resolvedTheme)
const localeMap: Record<LocaleType, RsLocale> = {
  en: 'en-US',
  'zh-CN': 'zh-CN',
}
const rsLocale = computed<RsLocale>(() => localeMap[getCurrentLocale()] ?? 'zh-CN')

function onUpdateShow(value: boolean) {
  if (value === false) $gRender.close()
}

function onSuccess(data?: unknown) {
  $gRender.closeWithSuccess(data)
}
</script>
