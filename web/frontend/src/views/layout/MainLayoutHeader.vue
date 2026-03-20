<template>
  <n-layout-header bordered class="main-layout-header">
    <!-- 单行：Logo → 多页签（不换行、中间区域横向滚动）→ 搜索与工具 -->
    <div class="main-layout-header__bar">
      <div class="main-layout-header__left">
        <div class="logo">
          <img src="@/assets/images/logo.png" alt="Logo" class="logo-img" />
          <span class="logo-text" v-if="!store.user.sidebarCollapsed">{{
            tCommon('common.appName')
          }}</span>
        </div>
      </div>

      <div class="main-layout-header__tabs">
        <GTabs
          v-model:tabs="layoutTabs"
          v-model:active-tab-id="layoutActiveTabId"
          type="line"
          :max-tabs="20"
        />
      </div>

      <div class="main-layout-header__right">
        <div class="search-box">
          <n-input
            v-model:value="searchQuery"
            :placeholder="tCommon('searchGlobal')"
            clearable
            round
            size="small"
          >
            <template #prefix>
              <n-icon>
                <SearchOutline />
              </n-icon>
            </template>
          </n-input>
        </div>

        <n-tooltip trigger="hover">
          <template #trigger>
            <n-button quaternary circle @click="emit('openToolMarketplace')">
              <template #icon>
                <n-icon size="18">
                  <AppsOutline />
                </n-icon>
              </template>
            </n-button>
          </template>
          {{ tCommon('toolMarket') }}
        </n-tooltip>

        <div class="notification-btn">
          <n-badge :value="0" :show="false">
            <n-button quaternary circle>
              <n-icon size="18">
                <NotificationsOutline />
              </n-icon>
            </n-button>
          </n-badge>
        </div>

        <div class="theme-switch">
          <ThemeSwitcher />
        </div>

        <GDropdown :options="userMenuOptions" trigger="click" @select="handleUserAction">
          <div class="user-info">
            <n-avatar round :src="store.user.avatar">
              {{ (store.user.displayName || store.user.userName || '?').charAt(0).toUpperCase() }}
            </n-avatar>
            <span class="user-name" v-if="!store.user.sidebarCollapsed">{{
              store.user.displayName
            }}</span>
          </div>
        </GDropdown>
      </div>
    </div>
  </n-layout-header>
</template>

<script setup lang="ts">
import ThemeSwitcher from '@/components/common/ThemeSwitcher.vue'
import { GDropdown } from '@/components/gdropdown'
import { GTabs } from '@/components/gtabs'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { store } from '@/stores'
import { useGlobalStore } from '@/stores/global'
import { storeToRefs } from 'pinia'
import { AppsOutline, NotificationsOutline, SearchOutline } from '@vicons/ionicons5'
import {
  NAvatar,
  NBadge,
  NButton,
  NIcon,
  NInput,
  NLayoutHeader,
  NTooltip,
} from 'naive-ui'
import { ref } from 'vue'
import { useLayoutUser } from './hooks'

const emit = defineEmits<{
  openToolMarketplace: []
}>()

const { t: tCommon } = useModuleI18n('common')
const { userMenuOptions, handleUserAction } = useLayoutUser()
const globalStore = useGlobalStore()
const { layoutTabs, layoutActiveTabId } = storeToRefs(globalStore)

const searchQuery = ref('')
</script>

<style lang="scss" scoped>
.main-layout-header {
  display: flex;
  flex-direction: column;
  align-items: stretch;
  height: var(--g-header-height) !important;
  min-height: var(--g-header-height);
  max-height: var(--g-header-height);
  padding: 0;
  background-color: var(--g-bg-primary);
  box-shadow: var(--g-shadow-sm);
  z-index: 10;
  overflow: hidden;

  &__bar {
    display: flex;
    flex-wrap: nowrap;
    align-items: center;
    gap: var(--g-space-sm);
    height: 100%;
    min-height: 0;
    padding: 0 var(--g-space-md);
    box-sizing: border-box;
  }

  &__left {
    display: flex;
    align-items: center;
    flex-shrink: 0;
    min-width: 0;

    .logo {
      display: flex;
      align-items: center;
      gap: var(--g-space-sm);

      .logo-img {
        width: 24px;
        height: 24px;
      }

      .logo-text {
        font-size: var(--g-font-size-lg);
        font-weight: 600;
        color: var(--g-primary);
        white-space: nowrap;
      }
    }
  }

  /* 夹在 Logo 与搜索之间，占满剩余宽度；内部 GTabs 横向滚动不换行 */
  &__tabs {
    flex: 1 1 0;
    min-width: 0;
    height: 100%;
    display: flex;
    align-items: center;
    overflow: hidden;

    :deep(.g-tabs) {
      width: 100%;
      min-width: 0;
      border: none;
      background: transparent;
    }

    :deep(.g-tabs-nav) {
      height: 36px;
    }

    :deep(.g-tabs-tab) {
      height: 36px;
      min-height: 36px;
    }

    :deep(.g-tabs-nav-wrap) {
      min-width: 0;
    }
  }

  &__right {
    display: flex;
    align-items: center;
    gap: var(--g-space-sm);
    flex-shrink: 0;

    .search-box {
      width: 200px;
      transition: width var(--g-transition-base) var(--g-transition-ease);

      &:focus-within {
        width: 240px;
      }
    }

    .notification-btn,
    .theme-switch {
      display: flex;
      align-items: center;
      justify-content: center;
    }

    .user-info {
      display: flex;
      align-items: center;
      padding: var(--g-space-xs) var(--g-space-sm);
      cursor: pointer;
      border-radius: var(--g-radius-md);
      transition: all var(--g-transition-base) var(--g-transition-ease);

      &:hover {
        background-color: var(--g-hover-overlay);
      }

      .user-name {
        margin-left: var(--g-space-sm);
        font-size: var(--g-font-size-base);
        color: var(--g-text-primary);
      }
    }
  }
}
</style>
