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
        <RsTabs
          v-if="tabItems.length > 0"
          v-model="layoutActiveTabId"
          class="main-layout-header__tab-nav"
          :items="tabItems"
          variant="line"
          size="md"
          panelless
          borderless
          closable
          draggable
          context-menu
          overflow="dropdown"
          :max-count="20"
          @close="onCloseTab"
          @close-batch="onCloseTabs"
          @reorder="onReorderTabs"
        />
        <div
          v-else
          class="main-layout-header__tabs-empty"
          aria-hidden="true"
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

        <RsDropdown :items="userMenuItems" :show-selected="false" @select="handleUserAction">
          <template #trigger>
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
          </template>
        </RsDropdown>
      </div>
    </div>
  </header>

  <RsDrawer
    v-model:open="helpDrawerVisible"
    :title="tCommon('helpManual')"
    side="right"
    size="lg"
    teleport-to="body"
  >
      <div class="help-manual-panel">
        <p class="help-manual-intro">{{ tCommon('helpManualDrawerIntro') }}</p>
        <RsAlert v-if="helpDocsError" type="warning">{{ helpDocsError }}</RsAlert>
        <RsAlert v-else type="default">{{ tCommon('helpManualDrawerHint') }}</RsAlert>
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
            @load="onHelpIframeLoad"
            @error="onHelpIframeError"
          />
        </div>
      </div>
      <template #footer>
        <RsButton variant="ghost" size="sm" @click="helpDrawerVisible = false">
          {{ tCommon('helpManualClose') }}
        </RsButton>
      </template>
  </RsDrawer>
</template>

<script setup lang="ts">
import ThemeSwitcher from '@/components/common/ThemeSwitcher.vue'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { store } from '@/stores'
import { useGlobalStore } from '@/stores/global'
import {
  RsAlert,
  RsAvatar,
  RsButton,
  RsDrawer,
  RsDropdown,
  RsIcon,
  RsInput,
  RsLoading,
  RsTabs,
  type RsTabItem,
} from '@/ui'
import { getDocsSiteHref } from '@/utils/docsHelpUrl'
import { storeToRefs } from 'pinia'
import { computed, ref } from 'vue'
import { useLayoutUser } from './hooks'

const emit = defineEmits<{
  openToolMarketplace: []
}>()

const { t: tCommon } = useModuleI18n('common')
const { userMenuItems, handleUserAction } = useLayoutUser()
const globalStore = useGlobalStore()
const { layoutTabs, layoutActiveTabId } = storeToRefs(globalStore)

const searchQuery = ref('')
const helpDrawerVisible = ref(false)
const helpIframeLoading = ref(false)
const helpDocsError = ref('')

const docsSiteHref = computed(() => getDocsSiteHref())

const tabItems = computed<RsTabItem[]>(() =>
  layoutTabs.value.map((tab) => ({
    value: tab.tabId,
    label: tab.title,
    icon: typeof tab.icon === 'string' && !/[A-Z]/.test(tab.icon) ? tab.icon : undefined,
    closable: tab.closable,
    fixed: tab.fixed,
  })),
)

function removeTabIds(ids: string[]) {
  if (!ids.length) return
  const idSet = new Set(ids)
  globalStore.setLayoutTabs(layoutTabs.value.filter((tab) => !idSet.has(tab.tabId) || tab.fixed))
}

function onCloseTab(value: string) {
  removeTabIds([value])
}

function onCloseTabs(values: string[]) {
  removeTabIds(values)
}

function onReorderTabs(dragValue: string, dropValue: string) {
  const tabs = [...layoutTabs.value]
  const from = tabs.findIndex((tab) => tab.tabId === dragValue)
  const to = tabs.findIndex((tab) => tab.tabId === dropValue)
  if (from < 0 || to < 0 || from === to) return
  if (tabs[to]?.fixed) return
  const [moved] = tabs.splice(from, 1)
  if (!moved) return
  tabs.splice(to, 0, moved)
  globalStore.setLayoutTabs(tabs)
}

function onHelpIframeLoad() {
  helpIframeLoading.value = false
}

function onHelpIframeError() {
  helpIframeLoading.value = false
  helpDocsError.value = tCommon('helpManualLoadFailed')
}

async function openHelpDrawer() {
  helpDocsError.value = ''
  helpIframeLoading.value = true
  helpDrawerVisible.value = true
  try {
    const res = await fetch(docsSiteHref.value, { method: 'GET', credentials: 'same-origin' })
    if (!res.ok) {
      helpDocsError.value = tCommon('helpManualLoadFailed')
      helpIframeLoading.value = false
    }
  } catch {
    helpDocsError.value = tCommon('helpManualLoadFailed')
    helpIframeLoading.value = false
  }
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

  &__tab-nav {
    width: 100%;
    min-width: 0;
  }

  &__tabs-empty {
    flex: 1 1 auto;
    min-width: 0;
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
  flex: 1;
  flex-direction: column;
  gap: var(--g-space-sm);
  min-height: 0;
  height: 100%;
  overflow: auto;
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
