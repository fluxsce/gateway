<template>
  <header class="main-layout-header">
    <div class="main-layout-header__bar">
      <div class="main-layout-header__left">
        <img src="@/assets/images/logo.png" alt="Logo" class="main-layout-header__logo" />
        <span v-if="!store.user.sidebarCollapsed" class="main-layout-header__title">
          {{ tCommon('common.appName') }}
        </span>
      </div>

      <div class="main-layout-header__tabs">
        <GTabs
          v-model:tabs="layoutTabs"
          v-model:active-tab-id="layoutActiveTabId"
          type="line"
          size="md"
          :max-tabs="20"
        />
      </div>

      <div class="main-layout-header__right">
        <RsInput
          v-model="searchQuery"
          class="main-layout-header__search"
          :placeholder="tCommon('searchGlobal')"
          clearable
          size="md"
        >
          <template #prefix>
            <RsIcon name="search" :size="18" />
          </template>
        </RsInput>

        <RsButton
          variant="ghost"
          size="md"
          icon-only
          :bordered="false"
          icon="book-open"
          :tooltip="tCommon('helpManualTooltip')"
          :aria-label="tCommon('helpManual')"
          @click="openHelpDrawer"
        />
        <RsButton
          variant="ghost"
          size="md"
          icon-only
          :bordered="false"
          icon="layout-grid"
          :tooltip="tCommon('toolMarket')"
          :aria-label="tCommon('toolMarket')"
          @click="emit('openToolMarketplace')"
        />
        <RsButton
          variant="ghost"
          size="md"
          icon-only
          :bordered="false"
          icon="bell"
          aria-label="Notifications"
        />
        <ThemeSwitcher icon-only />

        <GDropdown :options="userMenuOptions" trigger="click" @select="handleUserAction">
          <div class="main-layout-header__user">
            <RsAvatar
              size="md"
              :src="store.user.avatar || undefined"
              :name="store.user.displayName || store.user.userName || '?'"
            />
            <span v-if="!store.user.sidebarCollapsed" class="main-layout-header__user-name">
              {{ store.user.displayName }}
            </span>
          </div>
        </GDropdown>
      </div>
    </div>

    <RsDrawer
      v-model:open="helpDrawerVisible"
      :title="tCommon('helpManual')"
      side="right"
      size="lg"
    >
      <div class="help-manual-panel">
        <p class="help-manual-intro">{{ tCommon('helpManualDrawerIntro') }}</p>
        <RsAlert type="default">{{ tCommon('helpManualDrawerHint') }}</RsAlert>
        <a :href="docsSiteHref" target="_blank" rel="noopener noreferrer" class="help-manual-open-link">
          <RsButton variant="primary" size="sm" icon="external-link">
            {{ tCommon('helpManualOpenNew') }}
          </RsButton>
        </a>
        <div class="help-manual-iframe-wrap">
          <RsLoading v-if="helpIframeLoading" overlay block size="lg" />
          <iframe
            v-if="helpDrawerVisible"
            class="help-manual-iframe"
            :src="docsSiteHref"
            :title="tCommon('helpManual')"
            @load="helpIframeLoading = false"
          />
        </div>
      </div>
      <template #footer>
        <RsButton variant="ghost" size="sm" @click="helpDrawerVisible = false">
          {{ tCommon('helpManualClose') }}
        </RsButton>
      </template>
    </RsDrawer>
  </header>
</template>

<script setup lang="ts">
import ThemeSwitcher from '@/components/common/ThemeSwitcher.vue'
import { GDropdown } from '@/components/gdropdown'
import { GTabs } from '@/components/gtabs'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { store } from '@/stores'
import { useGlobalStore } from '@/stores/global'
import {
  RsAlert,
  RsAvatar,
  RsButton,
  RsDrawer,
  RsIcon,
  RsInput,
  RsLoading,
} from '@/ui'
import { getDocsSiteHref } from '@/utils/docsHelpUrl'
import { storeToRefs } from 'pinia'
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

const docsSiteHref = computed(() => getDocsSiteHref())

function openHelpDrawer() {
  helpIframeLoading.value = true
  helpDrawerVisible.value = true
}
</script>

<style lang="scss" scoped>
.main-layout-header {
  position: relative;
  z-index: var(--g-z-index-sticky);
  flex-shrink: 0;
  height: var(--g-header-height);
  overflow: visible;
  background: var(--g-bg-primary);
  border-bottom: 1px solid var(--g-border-primary);

  &__bar {
    display: flex;
    align-items: center;
    gap: var(--g-space-md);
    height: 100%;
    min-width: 0;
    padding: 0 var(--g-space-md);
  }

  &__left {
    display: flex;
    align-items: center;
    gap: var(--g-space-sm);
    flex-shrink: 0;
  }

  &__logo {
    width: 24px;
    height: 24px;
  }

  &__title {
    font-size: var(--g-font-size-lg);
    font-weight: 500;
    color: var(--g-primary);
    white-space: nowrap;
  }

  &__tabs {
    flex: 1 1 0;
    min-width: 0;
    overflow: hidden;
  }

  &__right {
    display: flex;
    align-items: center;
    gap: var(--g-space-sm);
    flex-shrink: 0;
  }

  &__search {
    width: 200px;
  }

  &__user {
    display: flex;
    align-items: center;
    gap: var(--g-space-sm);
    cursor: pointer;
  }

  &__user-name {
    max-width: 8em;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
}

@media (max-width: 960px) {
  .main-layout-header {
    &__search {
      width: 140px;
    }

    &__user-name {
      display: none;
    }
  }
}

.help-manual-panel {
  display: flex;
  flex-direction: column;
  gap: var(--g-space-sm);
  min-height: 0;
  max-height: calc(100vh - 7.5rem);
}

.help-manual-intro {
  margin: 0;
  font-size: var(--g-font-size-sm);
  color: var(--g-text-secondary);
}

.help-manual-open-link {
  display: inline-flex;
  width: fit-content;
  text-decoration: none;
  color: inherit;
}

.help-manual-iframe-wrap {
  position: relative;
  flex: 1;
  min-height: 280px;
  border: 1px solid var(--g-border-primary);
  border-radius: var(--g-radius-md);
  overflow: hidden;
}

.help-manual-iframe {
  display: block;
  width: 100%;
  height: 100%;
  min-height: 280px;
  border: 0;
}
</style>
