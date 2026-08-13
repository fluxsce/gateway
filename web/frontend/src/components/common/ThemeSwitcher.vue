<template>
  <div class="theme-switcher">
    <RsButton
      v-if="iconOnly"
      variant="ghost"
      size="md"
      icon-only
      :bordered="false"
      :icon="isDark ? 'moon' : 'sun'"
      :tooltip="t(`theme.${isDark ? 'light' : 'dark'}`)"
      :aria-label="t(`theme.${isDark ? 'light' : 'dark'}`)"
      @click="toggleTheme"
    />
    <RsButton v-else variant="ghost" size="md" :icon="isDark ? 'moon' : 'sun'" @click="toggleTheme">
      {{ t(`theme.${isDark ? 'light' : 'dark'}`) }}
    </RsButton>
  </div>
</template>

<script lang="ts" setup>
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { useUserStore } from '@/stores/user'
import { RsButton } from '@/ui'
import { computed } from 'vue'

withDefaults(
  defineProps<{
    /** 顶栏等紧凑场景：仅显示图标 */
    iconOnly?: boolean
  }>(),
  { iconOnly: false },
)

const userStore = useUserStore()
const isDark = computed(() => userStore.isDark)

const toggleTheme = () => {
  const nextTheme = userStore.isDark ? 'light' : 'dark'
  userStore.update({ theme: nextTheme }, { persistUserData: false })
}

const { t } = useModuleI18n('common')
</script>
