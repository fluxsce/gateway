<template>
  <div class="main-layout">
    <MainLayoutHeader @open-tool-marketplace="openToolMarketplace" />

    <div class="main-layout__body">
      <aside
        class="layout-sidebar"
        :class="{ 'layout-sidebar--collapsed': store.user.sidebarCollapsed }"
      >
        <div class="layout-sidebar__scroll">
          <RsMenu
            v-model="activeMenuKey"
            v-model:open-keys="openMenuKeys"
            :items="menuItems"
            :collapsed="store.user.sidebarCollapsed"
            highlight-parent
            class="layout-sidebar__menu"
            @select="handleMenuSelect"
          />
        </div>

        <div class="layout-sidebar__footer">
          <RsButton
            variant="text"
            size="sm"
            radius="sm"
            class="layout-sidebar__collapse"
            :icon="store.user.sidebarCollapsed ? 'chevron-right' : 'chevron-left'"
            :icon-only="store.user.sidebarCollapsed"
            :tooltip="
              store.user.sidebarCollapsed ? tCommon('menu.expand') : tCommon('menu.collapse')
            "
            :aria-label="
              store.user.sidebarCollapsed ? tCommon('menu.expand') : tCommon('menu.collapse')
            "
            @click="store.user.toggleSidebar"
          >
            <template v-if="!store.user.sidebarCollapsed">
              {{ tCommon('menu.collapse') }}
            </template>
          </RsButton>
        </div>
      </aside>

      <MainLayoutContent />
    </div>

    <RsDialog
      v-model:open="showToolMarketplace"
      :title="tCommon('toolMarket')"
      description="发现和安装工具，扩展工作台能力"
      layout="window"
      :width="1080"
      :show-close="true"
      :close-on-overlay-click="true"
      :resizable="true"
      :fullscreenable="true"
    >
      <template #body>
        <tool-marketplace class="main-layout__marketplace" />
      </template>
    </RsDialog>
  </div>
</template>

<script setup lang="ts">
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { store } from '@/stores'
import { RsButton, RsDialog, RsMenu, useRsLoadingBar } from '@/ui'
import { onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import MainLayoutContent from './MainLayoutContent.vue'
import MainLayoutHeader from './MainLayoutHeader.vue'
import ToolMarketplace from './compoents/ToolMarketplace.vue'
import { useLayoutMenu } from './hooks'

const { t: tCommon } = useModuleI18n('common')

const loadingBar = useRsLoadingBar()
const route = useRoute()

watch(
  () => route.path,
  () => {
    loadingBar.start()
    setTimeout(() => {
      loadingBar.finish()
    }, 300)
  },
)

onMounted(() => {
  loadingBar.finish()
})

const showToolMarketplace = ref(false)
const { menuItems, activeMenuKey, openMenuKeys, handleMenuSelect } = useLayoutMenu()

const openToolMarketplace = () => {
  showToolMarketplace.value = true
}
</script>

<style lang="scss" scoped>
.main-layout {
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 100vh;
  overflow: hidden;
  background-color: var(--rs-bg);
}

.main-layout__body {
  display: flex;
  flex: 1;
  min-height: 0;
  min-width: 0;
  overflow: hidden;
}

.layout-sidebar {
  --layout-sidebar-width: 14rem;
  --layout-sidebar-collapsed-width: 3.5rem;
  /* 与 toolbar text 按钮一致：选中项不加粗 */
  --rs-menu-item-active-weight: 400;

  display: flex;
  flex-direction: column;
  flex-shrink: 0;
  width: var(--layout-sidebar-width);
  min-height: 0;
  overflow: hidden;
  background: var(--rs-surface);
  border-right: 1px solid var(--rs-border);
  transition: width var(--rs-transition-normal);

  &--collapsed {
    width: var(--layout-sidebar-collapsed-width);
  }

  &__scroll {
    flex: 1;
    min-height: 0;
    overflow-x: hidden;
    overflow-y: auto;
    padding: var(--rs-space-sm);
  }

  &__menu {
    width: 100%;
  }

  &__footer {
    flex-shrink: 0;
    box-sizing: border-box;
    display: flex;
    align-items: center;
    height: var(--g-footer-height);
    padding: 0 var(--rs-space-sm);
    border-top: 1px solid var(--rs-border);
  }

  &__collapse {
    width: 100%;
    height: 100%;
    justify-content: flex-start;
  }

  &--collapsed &__footer {
    padding: 0 var(--rs-space-xs);
    justify-content: center;
  }

  &--collapsed &__collapse {
    width: auto;
    margin-inline: auto;
    justify-content: center;
  }
}

.main-layout__marketplace {
  min-height: min(70vh, 40rem);
}
</style>
