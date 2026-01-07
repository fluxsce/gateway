<template>
  <n-layout class="main-layout">
    <!-- 顶部导航栏 -->
    <n-layout-header bordered class="header">
      <div class="header-left">
        <div class="logo">
          <img src="@/assets/images/logo.png" alt="Logo" class="logo-img" />
          <span class="logo-text" v-if="!store.user.sidebarCollapsed">{{
            tCommon('common.appName')
          }}</span>
        </div>
      </div>

      <div class="header-center">
        <n-breadcrumb>
          <n-breadcrumb-item v-for="item in breadcrumbs" :key="item.path">
            {{ item.title }}
          </n-breadcrumb-item>
        </n-breadcrumb>
      </div>

      <div class="header-right">
        <!-- 全局搜索 -->
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

        <!-- 工具市场快捷入口 -->
        <n-tooltip trigger="hover">
          <template #trigger>
            <n-button quaternary circle @click="openToolMarketplace">
              <template #icon>
                <n-icon size="18">
                  <AppsOutline />
                </n-icon>
              </template>
            </n-button>
          </template>
          {{ tCommon('toolMarket') }}
        </n-tooltip>

        <!-- 通知图标 -->
        <div class="notification-btn">
          <n-badge :value="0" :show="false">
            <n-button quaternary circle>
              <n-icon size="18">
                <NotificationsOutline />
              </n-icon>
            </n-button>
          </n-badge>
        </div>

        <!-- 主题切换 -->
        <div class="theme-switch">
          <ThemeSwitcher />
        </div>

        <!-- 用户信息 -->
        <n-dropdown :options="userMenuOptions" @select="handleUserAction" trigger="click">
          <div class="user-info">
            <n-avatar round :src="store.user.avatar">
              {{ (store.user.displayName || store.user.userName || '?').charAt(0).toUpperCase() }}
            </n-avatar>
            <span class="user-name" v-if="!store.user.sidebarCollapsed">{{
              store.user.displayName
            }}</span>
          </div>
        </n-dropdown>
      </div>
    </n-layout-header>

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

        <!-- 折叠按钮 -->
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

      <!-- 主内容区域 -->
      <n-layout-content class="main-content">
        <router-view />
      </n-layout-content>
    </n-layout>

    <!-- 工具市场模态框 -->
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
import ThemeSwitcher from '@/components/common/ThemeSwitcher.vue'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { store } from '@/stores'
import { AppsOutline, ListSharp, NotificationsOutline, SearchOutline } from '@vicons/ionicons5'
import { NLayout, NLayoutContent, NLayoutHeader, NLayoutSider, NModal } from 'naive-ui'
import { ref } from 'vue'
import ToolMarketplace from './compoents/ToolMarketplace.vue'
import { useLayoutMenu, useLayoutUser } from './hooks'

// 国际化
const { t: tCommon } = useModuleI18n('common')

// 工具市场面板状态
const showToolMarketplace = ref(false)

// 全局搜索
const searchQuery = ref('')

// 使用提取出的hooks
const { menuOptions, handleMenuSelect, breadcrumbs } = useLayoutMenu()

const { userMenuOptions, handleUserAction } = useLayoutUser()

// 打开工具市场
const openToolMarketplace = () => {
  showToolMarketplace.value = true
}
</script>

<style lang="scss" scoped>
.main-layout {
  height: 100vh;
  background-color: var(--g-bg-secondary);
}

.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: var(--g-header-height);
  padding: 0 var(--g-space-md);
  background-color: var(--g-bg-primary);
  box-shadow: var(--g-shadow-sm);
  z-index: 10;

  .header-left {
    display: flex;
    align-items: center;

    .logo {
      display: flex;
      align-items: center;
      margin-right: var(--g-space-md);

      .logo-img {
        width: 24px;
        height: 24px;
      }

      .logo-text {
        font-size: var(--g-font-size-lg);
        font-weight: 600;
        margin-left: var(--g-space-sm);
        color: var(--g-primary);
      }
    }
  }

  .header-center {
    flex: 1;
    padding: 0 var(--g-space-md);
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: var(--g-space-sm);

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

    // 菜单项样式
    .n-menu-item-content {
      padding-left: var(--g-space-md) !important;
      padding-right: var(--g-space-md) !important;
      border-radius: var(--g-radius-md);
      margin: 2px var(--g-space-xs);
      transition: all var(--g-transition-base) var(--g-transition-ease);

      &:hover {
        background-color: var(--g-hover-overlay);
      }

      // 选中状态
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

    // 子菜单缩进
    .n-submenu-children {
      .n-menu-item-content {
        padding-left: calc(var(--g-space-md) + var(--g-space-lg)) !important;
      }
    }
  }

  // 侧边栏底部
  .sidebar-footer {
    border-top: 1px solid var(--g-border-primary);
    height: var(--g-footer-height);
    min-height: var(--g-footer-height);
    /* padding: 0 var(--g-space-sm); */
    display: flex;
    /* align-items: center; */
    box-sizing: unset;

    .collapse-btn {
      display: flex;
      align-items: center;
      justify-content: flex-start;
      /* padding: var(--g-space-sm) 0; */
      height: 100%;
      width: 100%;
      /* border-radius: var(--g-radius-md); */
      cursor: pointer;
      /* transition: all var(--g-transition-base) var(--g-transition-ease); */
      /* color: var(--g-text-secondary); */
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

  // 折叠状态
  &.n-layout-sider--collapsed {
    .sidebar-footer .collapse-btn {
      justify-content: center;
      padding: var(--g-space-sm);
    }
  }
}

.main-content {
  height: calc(100vh - var(--g-header-height));
  border: 0.5px solid var(--g-border-primary);
  border-radius: var(--g-radius-2xl);
  box-sizing: border-box;
}

/* 响应式布局 */
@media (max-width: 768px) {
  .header {
    .header-left {
      .logo-text {
        display: none;
      }
    }

    .header-center {
      display: none;
    }

    .header-right {
      .search-box {
        width: 160px;

        &:focus-within {
          width: 180px;
        }
      }
    }
  }

  :deep(.n-layout-sider) {
    &:not(.n-layout-sider--collapsed) {
      box-shadow: var(--g-shadow-md);
    }
  }
}
</style>
