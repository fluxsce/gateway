<template>
  <n-layout class="main-layout">
    <MainLayoutHeader @open-tool-marketplace="openToolMarketplace" />

    <n-layout has-sider position="absolute" :style="{ top: 'var(--g-header-height)', bottom: 0 }">
      <!-- 侧边菜单 -->
      <n-layout-sider
        bordered
        collapse-mode="width"
        :collapsed-width="64"
        :width="220"
        :collapsed="store.user.sidebarCollapsed"
        :show-trigger="false"
        class="sidebar"
        content-style="display: flex; flex-direction: column; height: 100%; overflow-x: hidden;"
      >
        <n-menu
          :collapsed="store.user.sidebarCollapsed"
          :collapsed-width="64"
          :collapsed-icon-size="18"
          :options="menuOptions"
          :indent="20"
          :on-update:value="handleMenuSelect"
          style="flex: 1; overflow-y: auto; overflow-x: hidden"
        />

        <div class="sidebar-footer">
          <div class="collapse-btn" @click="store.user.toggleSidebar">
            <n-icon size="18">
              <ListSharp v-if="store.user.sidebarCollapsed" />
              <ListSharp v-else />
            </n-icon>
            <span class="collapse-text" v-if="!store.user.sidebarCollapsed">
              {{ store.user.sidebarCollapsed ? tCommon('menu.expand') : tCommon('menu.collapse') }}
            </span>
          </div>
        </div>
      </n-layout-sider>

      <MainLayoutContent />
    </n-layout>

    <n-modal
      v-model:show="showToolMarketplace"
      preset="card"
      :title="tCommon('toolMarket')"
      :style="{ width: '90vw', maxWidth: '1200px' }"
      :segmented="{ content: true }"
      size="huge"
      :closable="true"
      :mask-closable="true"
      :auto-focus="false"
    >
      <tool-marketplace />
    </n-modal>
  </n-layout>
</template>

<script setup lang="ts">
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { store } from '@/stores'
import { ListSharp } from '@vicons/ionicons5'
import { NLayout, NLayoutSider, NModal, useLoadingBar } from 'naive-ui'
import { onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import MainLayoutContent from './MainLayoutContent.vue'
import MainLayoutHeader from './MainLayoutHeader.vue'
import ToolMarketplace from './compoents/ToolMarketplace.vue'
import { useLayoutMenu } from './hooks'

const { t: tCommon } = useModuleI18n('common')

const loadingBar = useLoadingBar()
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
const { menuOptions, handleMenuSelect } = useLayoutMenu()

const openToolMarketplace = () => {
  showToolMarketplace.value = true
}
</script>

<style lang="scss" scoped>
.main-layout {
  height: 100vh;
  background-color: var(--g-bg-secondary);
}

.sidebar {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow-x: hidden;
  background-color: var(--g-bg-primary);
  border-right: 1px solid var(--g-border-primary);
  border-radius: var(--g-radius-2xl);
  box-sizing: border-box;
  transition: all var(--g-transition-base) var(--g-transition-ease);

  :deep(.n-menu) {
    overflow-x: hidden;
    max-width: 100%;
    border: none !important;

    .n-menu-item-content,
    .n-submenu-title {
      white-space: nowrap;
      text-overflow: ellipsis;
      overflow: hidden;
    }

    .n-menu-item-content {
      padding-left: var(--g-space-md) !important;
      padding-right: var(--g-space-md) !important;
      border-radius: var(--g-radius-md);
      margin: 2px var(--g-space-xs);
      transition: all var(--g-transition-base) var(--g-transition-ease);

      &:hover {
        background-color: var(--g-hover-overlay);
      }

      &.n-menu-item-content--selected {
        background-color: var(--g-primary-light);
        color: var(--g-primary);
        font-weight: 500;
        position: relative;

        &::before {
          content: '';
          position: absolute;
          left: 0;
          top: 50%;
          transform: translateY(-50%);
          width: 3px;
          height: 60%;
          background-color: var(--g-primary);
          border-radius: 0 var(--g-radius-sm) var(--g-radius-sm) 0;
        }
      }
    }

    .n-submenu-children {
      .n-menu-item-content {
        padding-left: calc(var(--g-space-md) + var(--g-space-lg)) !important;
      }
    }
  }

  .sidebar-footer {
    border-top: 1px solid var(--g-border-primary);
    height: var(--g-footer-height);
    min-height: var(--g-footer-height);
    display: flex;
    box-sizing: unset;

    .collapse-btn {
      display: flex;
      align-items: center;
      justify-content: flex-start;
      height: 100%;
      width: 100%;
      cursor: pointer;
      box-sizing: border-box;

      &:hover {
        background-color: var(--g-hover-overlay);
        color: var(--g-text-primary);
      }

      .collapse-text {
        margin-left: var(--g-space-sm);
        font-size: var(--g-font-size-sm);
      }
    }
  }

  &.n-layout-sider--collapsed {
    .sidebar-footer .collapse-btn {
      justify-content: center;
      padding: var(--g-space-sm);
    }
  }
}
</style>
