<template>
  <div
    class="language-switcher"
    :class="{ 'language-switcher--dark-surface': variant === 'dark-surface' }"
  >
    <RsDropdown
      :model-value="userLanguage"
      :items="languageItems"
      :disabled="isLoading"
      @select="handleLanguageSelect"
    >
      <template #trigger>
        <button
          type="button"
          class="language-selector"
          :class="{ 'is-loading': isLoading }"
          :aria-label="currentLanguageName"
        >
          <GIcon size="16" class="globe-icon">
            <LanguageOutline />
          </GIcon>
          <span class="current-language">{{ currentLanguageName }}</span>
          <GIcon size="14" :class="isLoading ? 'chevron-icon is-loading' : 'chevron-icon'">
            <ChevronDownOutline />
          </GIcon>
        </button>
      </template>
    </RsDropdown>
  </div>
</template>

<script setup lang="ts">
import GIcon from '@/components/gicon/GIcon.vue'
import { availableLocales, setLocale, type LocaleType } from '@/locales'
import { useUserStore } from '@/stores/user'
import { RsDropdown, type RsDropdownItem } from '@/ui'
import { ChevronDownOutline, LanguageOutline } from '@vicons/ionicons5'
import { computed, ref } from 'vue'

withDefaults(
  defineProps<{
    /** 深色背景上的触发器样式（登录页等） */
    variant?: 'default' | 'dark-surface'
  }>(),
  { variant: 'default' },
)

const userStore = useUserStore()
const isLoading = ref(false)

const userLanguage = computed(() => userStore.language)

const currentLanguageName = computed(() => {
  const locale = availableLocales.find((item) => item.locale === userLanguage.value)
  return locale ? locale.name : 'Unknown'
})

const languageItems: RsDropdownItem[] = availableLocales.map((locale) => ({
  value: locale.locale,
  label: locale.name,
}))

async function handleLanguageSelect(localeKey: string) {
  if (localeKey === userLanguage.value) return

  isLoading.value = true
  try {
    userStore.updateSettings({ language: localeKey })
    await setLocale(localeKey as LocaleType)
  } finally {
    isLoading.value = false
  }
}
</script>

<style lang="scss" scoped>
.language-switcher {
  display: inline-flex;
  position: relative;
}

.language-selector {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-height: 36px;
  padding: 0 12px 0 10px;
  border-radius: 999px;
  border: 1px solid rgba(15, 23, 42, 0.1);
  background: rgba(255, 255, 255, 0.92);
  color: #334155;
  cursor: pointer;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.06);
  transition:
    background 0.18s ease,
    border-color 0.18s ease,
    box-shadow 0.18s ease,
    transform 0.18s ease;

  &:hover {
    background: #fff;
    border-color: rgba(99, 102, 241, 0.28);
    box-shadow: 0 4px 14px rgba(15, 23, 42, 0.08);
  }

  &:focus-visible {
    outline: none;
    border-color: #818cf8;
    box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.22);
  }

  &.is-loading {
    opacity: 0.88;
  }
}

.globe-icon {
  color: #6366f1;
  flex-shrink: 0;
}

.current-language {
  font-size: 13px;
  font-weight: 600;
  letter-spacing: 0.01em;
  line-height: 1;
}

.chevron-icon {
  opacity: 0.65;
  flex-shrink: 0;
  transition: transform 0.18s ease;

  &.is-loading {
    animation: language-spin 0.7s linear infinite;
  }
}

/* 登录页深色氛围：玻璃胶囊，与背景一体 */
.language-switcher--dark-surface {
  .language-selector {
    color: rgba(255, 255, 255, 0.94);
    border: 1px solid rgba(255, 255, 255, 0.16);
    background: rgba(255, 255, 255, 0.08);
    box-shadow:
      0 8px 24px rgba(0, 0, 0, 0.18),
      inset 0 1px 0 rgba(255, 255, 255, 0.12);
    backdrop-filter: blur(14px) saturate(1.2);
    -webkit-backdrop-filter: blur(14px) saturate(1.2);

    &:hover {
      background: rgba(255, 255, 255, 0.14);
      border-color: rgba(199, 210, 254, 0.45);
      box-shadow:
        0 10px 28px rgba(0, 0, 0, 0.22),
        inset 0 1px 0 rgba(255, 255, 255, 0.16);
      transform: translateY(-1px);
    }

    &:focus-visible {
      border-color: rgba(165, 180, 252, 0.85);
      box-shadow: 0 0 0 3px rgba(129, 140, 248, 0.28);
    }
  }

  .globe-icon,
  .chevron-icon {
    color: #c7d2fe;
  }

  .current-language {
    color: inherit;
    text-shadow: 0 1px 2px rgba(0, 0, 0, 0.25);
  }
}

@keyframes language-spin {
  to {
    transform: rotate(360deg);
  }
}

@media (prefers-reduced-motion: reduce) {
  .language-selector,
  .chevron-icon {
    transition: none;
  }

  .chevron-icon.is-loading {
    animation: none;
  }

  .language-switcher--dark-surface .language-selector:hover {
    transform: none;
  }
}
</style>
