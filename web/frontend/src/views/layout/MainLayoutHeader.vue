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

        <n-tooltip trigger="hover" placement="bottom">
          <template #trigger>
            <n-button quaternary circle :aria-label="tCommon('helpManual')" @click="openHelpDrawer">
              <template #icon>
                <n-icon size="18">
                  <BookOutline />
                </n-icon>
              </template>
            </n-button>
          </template>
          {{ tCommon('helpManualTooltip') }}
        </n-tooltip>

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

    <n-drawer
      v-model:show="helpDrawerVisible"
      width="min(960px, 96vw)"
      placement="right"
      display-directive="show"
      :auto-focus="false"
    >
      <n-drawer-content
        :title="tCommon('helpManual')"
        closable
        :body-content-style="helpDrawerBodyStyle"
      >
        <div class="help-manual-panel">
          <n-text depth="3" class="help-manual-intro">
            {{ tCommon('helpManualDrawerIntro') }}
          </n-text>
          <n-alert type="info" :show-icon="true" class="help-manual-alert">
            {{ tCommon('helpManualDrawerHint') }}
          </n-alert>
          <div class="help-manual-toolbar">
            <n-button
              type="primary"
              secondary
              size="small"
              tag="a"
              :href="docsSiteHref"
              target="_blank"
              rel="noopener noreferrer"
            >
              <template #icon>
                <n-icon :component="LinkOutline" />
              </template>
              {{ tCommon('helpManualOpenNew') }}
            </n-button>
          </div>
          <div class="help-manual-iframe-wrap">
            <n-spin :show="helpIframeLoading" class="help-manual-spin">
              <iframe
                v-if="helpDrawerVisible"
                class="help-manual-iframe"
                :src="docsSiteHref"
                :title="tCommon('helpManual')"
                @load="helpIframeLoading = false"
              />
            </n-spin>
          </div>
        </div>
        <template #footer>
          <div class="help-manual-footer">
            <n-button quaternary size="small" @click="helpDrawerVisible = false">
              {{ tCommon('helpManualClose') }}
            </n-button>
          </div>
        </template>
      </n-drawer-content>
    </n-drawer>
  </n-layout-header>
</template>

<script setup lang="ts">
import ThemeSwitcher from '@/components/common/ThemeSwitcher.vue'
import { GDropdown } from '@/components/gdropdown'
import { GTabs } from '@/components/gtabs'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { store } from '@/stores'
import { useGlobalStore } from '@/stores/global'
import { getDocsSiteHref } from '@/utils/docsHelpUrl'
import { storeToRefs } from 'pinia'
import { AppsOutline, BookOutline, LinkOutline, NotificationsOutline, SearchOutline } from '@vicons/ionicons5'
import {
  NAlert,
  NAvatar,
  NBadge,
  NButton,
  NDrawer,
  NDrawerContent,
  NIcon,
  NInput,
  NLayoutHeader,
  NSpin,
  NText,
  NTooltip,
} from 'naive-ui'
import { computed, ref } from 'vue'
import { useLayoutUser } from './hooks'

const emit = defineEmits<{
  openToolMarketplace: []
}>()

const { t: tCommon } = useModuleI18n('common')
const { userMenuOptions, handleUserAction } = useLayoutUser()
const globalStore = useGlobalStore()
const { layoutTabs, layoutActiveTabId } = storeToRefs(globalStore)

const searchQuery = ref('')
const helpDrawerVisible = ref(false)
const helpIframeLoading = ref(false)

/** 与 VitePress `base` 一致的文档根 URL，供 iframe 内嵌 */
const docsSiteHref = computed(() => getDocsSiteHref())

/** 抽屉正文区域：纵向排布说明 + iframe，避免内容顶死视口 */
const helpDrawerBodyStyle = {
  display: 'flex',
  flexDirection: 'column' as const,
  flex: '1 1 auto',
  minHeight: 0,
  padding: '4px 0 0',
}

function openHelpDrawer() {
  helpIframeLoading.value = true
  helpDrawerVisible.value = true
}
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

.help-manual-panel {
  display: flex;
  flex-direction: column;
  gap: var(--g-space-sm);
  flex: 1 1 auto;
  min-height: 0;
  max-height: calc(100vh - 7.5rem);
}

.help-manual-intro {
  font-size: var(--g-font-size-sm);
  line-height: 1.5;
}

.help-manual-alert {
  border-radius: var(--g-radius-md);
  background: var(--g-bg-secondary);
}

.help-manual-toolbar {
  flex-shrink: 0;
}

.help-manual-iframe-wrap {
  position: relative;
  flex: 1 1 auto;
  min-height: 280px;
  border-radius: var(--g-radius-md);
  overflow: hidden;
  box-shadow: inset 0 0 0 1px var(--g-border-color);
  background: var(--g-bg-secondary);
}

.help-manual-spin {
  height: 100%;
  min-height: 280px;
}

.help-manual-spin :deep(.n-spin-content) {
  height: 100%;
}

.help-manual-iframe {
  display: block;
  width: 100%;
  height: 100%;
  min-height: 280px;
  border: 0;
  background: var(--g-bg-secondary);
}

.help-manual-footer {
  display: flex;
  justify-content: flex-end;
}
</style>
