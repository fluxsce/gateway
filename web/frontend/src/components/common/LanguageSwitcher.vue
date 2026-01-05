<template>
    <div class="language-switcher">
        <!-- 语言切换下拉框 -->
        <NDropdown trigger="click" :options="languageOptions" @select="handleLanguageSelect">
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
        </NDropdown>
    </div>
</template>

<script setup lang="ts">
import { availableLocales, setLocale, type LocaleType } from '@/locales'
import { useUserStore } from '@/stores/user'
import type { DropdownOption } from 'naive-ui'
import { NDropdown, NIcon } from 'naive-ui'
import { computed, ref } from 'vue'

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
async function handleLanguageSelect(key: string) {
    // 如果与当前语言相同，则不处理
    if (key === userLanguage.value) return

    isLoading.value = true
    try {
        // 更新用户设置
        userStore.updateSettings({ language: key })

        // 更新i18n设置 (使用LocaleType格式)
        await setLocale(key as LocaleType)
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
</style>