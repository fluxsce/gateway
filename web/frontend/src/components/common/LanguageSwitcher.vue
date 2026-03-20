<template>
    <div class="language-switcher" :class="{ 'language-switcher--dark-surface': variant === 'dark-surface' }">
        <!-- 语言切换下拉框 -->
        <GDropdown trigger="click" :options="languageOptions" @select="handleLanguageSelect">
            <div class="language-selector" :class="{ 'is-loading': isLoading }">
                <!-- 加载状态指示器 -->
                <NIcon v-if="isLoading" size="16" class="loading-icon">
                    <i class="fas fa-circle-notch fa-spin"></i>
                </NIcon>
                <!-- 当前语言显示 -->
                <span class="current-language">{{ getCurrentLanguageName() }}</span>
                <NIcon size="12">
                    <i class="fas fa-chevron-down"></i>
                </NIcon>
            </div>
        </GDropdown>
    </div>
</template>

<script setup lang="ts">
import { GDropdown } from '@/components/gdropdown'
import { availableLocales, setLocale, type LocaleType } from '@/locales'
import { useUserStore } from '@/stores/user'
import type { DropdownOption } from 'naive-ui'
import { NIcon } from 'naive-ui'
import { computed, ref } from 'vue'

withDefaults(
    defineProps<{
        /** 深色背景上的触发器样式（登录页等），不受全局亮/暗主题文字色影响 */
        variant?: 'default' | 'dark-surface'
    }>(),
    { variant: 'default' }
)

const userStore = useUserStore()
const isLoading = ref(false)

// 获取用户当前设置的语言
const userLanguage = computed(() => userStore.language)

/**
 * 获取当前语言名称
 */
function getCurrentLanguageName(): string {
    // 根据用户设置找到对应的语言显示名称
    const locale = availableLocales.find(item => item.locale === userLanguage.value)
    return locale ? locale.name : 'Unknown'
}

/**
 * 语言选项列表 - 直接使用availableLocales数据
 */
const languageOptions: DropdownOption[] = availableLocales.map(locale => ({
    key: locale.locale,
    label: locale.name
}))

/**
 * 处理语言选择
 */
async function handleLanguageSelect(key: string | number) {
    // 如果与当前语言相同，则不处理
    const localeKey = String(key)
    if (localeKey === userLanguage.value) return

    isLoading.value = true
    try {
        // 更新用户设置
        userStore.updateSettings({ language: localeKey })

        // 更新i18n设置 (使用LocaleType格式)
        await setLocale(localeKey as LocaleType)
    } finally {
        isLoading.value = false
    }
}
</script>

<style lang="scss" scoped>
.language-switcher {
    display: inline-block;
    position: relative;
}

.language-selector {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 4px 8px;
    border-radius: 4px;
    cursor: pointer;
    transition: background-color 0.2s;

    &:hover {
        background-color: rgba(0, 0, 0, 0.05);
    }

    &.is-loading {
        opacity: 0.8;
        pointer-events: none;
    }
}

.loading-icon {
    color: #1890ff;
}

.current-language {
    font-size: 14px;
}

/* 登录页等深色渐变底：强制浅色字，避免 [data-theme=light] 时仍是深灰字看不清 */
.language-switcher--dark-surface {
    .language-selector {
        color: rgba(255, 255, 255, 0.94);
        text-shadow: 0 1px 3px rgba(0, 0, 0, 0.45);

        &:hover {
            background-color: rgba(255, 255, 255, 0.14);
        }
    }

    .current-language {
        color: inherit;
        font-weight: 500;
    }

    .loading-icon {
        color: #c7d2fe;
    }

    :deep(.n-icon) {
        color: rgba(255, 255, 255, 0.88);
    }
}
</style>