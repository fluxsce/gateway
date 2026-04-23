/**
 * 全局状态管理 Store
 *
 * 职责：
 * - 应用元信息（名称、版本）、页面标题与加载态
 * - 主布局多页签（与 `GTabs` 一致；页签由菜单等交互打开，不预置首页）
 * - 侧边栏显示偏好（经 `pinia-plugin-persistedstate` 持久化，无需手写 storage）
 *
 * 写法对齐 XiRang `config-store`：Setup Store + 分区注释 + 具名 ref/函数，便于维护与扩展。
 */

import type { GTabsTabItem } from '@/components/gtabs/types'
import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import type { RouteLocationNormalizedLoaded } from 'vue-router'

// ---------------------------------------------------------------------------
// 常量（持久化）
// ---------------------------------------------------------------------------

/**
 * `localStorage` 中本 Store 持久化片段使用的键名（JSON 对象，仅含 `pick` 字段）。
 */
const GLOBAL_PERSIST_KEY = 'gateway-global'

// ---------------------------------------------------------------------------
// Store
// ---------------------------------------------------------------------------

export const useGlobalStore = defineStore(
  'global',
  () => {
    // ==================== 状态 ====================

    /** 应用展示名称（如页标题后缀、关于页文案） */
    const appName = ref('Gateway')

    /** 应用版本号（启动时由入口或接口注入） */
    const appVersion = ref('')

    /** 当前业务页标题（不含应用名后缀；`setPageTitle` 会同步 `document.title`） */
    const pageTitle = ref('')

    /** 全屏或路由级加载遮罩开关 */
    const pageLoading = ref(false)

    /**
     * 主布局页签列表（`tabId` 使用路由 `fullPath`，与 `GTabs` 一致）
     * 初始为空，由侧栏菜单 `upsertLayoutTab` 等写入。
     */
    const layoutTabs = ref<GTabsTabItem[]>([])

    /** 当前激活页签 id，与 `layoutTabs[].tabId` 对应；无页签时为 `''` */
    const layoutActiveTabId = ref('')

    /** 是否显示主布局侧边栏区域（持久化字段） */
    const showSidebar = ref(true)

    // ==================== 计算属性 ====================

    /**
     * 当前激活页签对象；未命中时为 `null`（例如列表刚被外部清空）。
     */
    const activeLayoutTab = computed(
      () => layoutTabs.value.find((t) => t.tabId === layoutActiveTabId.value) ?? null,
    )

    // ==================== 方法：应用 / 页面 ====================

    /**
     * 设置应用版本号
     *
     * @param version - 语义化版本或构建号
     */
    function setAppVersion(version: string) {
      appVersion.value = version
    }

    /**
     * 设置页面标题并同步浏览器标签标题：`{pageTitle} - {appName}` 或仅 `appName`
     *
     * @param title - 业务页标题，空字符串表示仅显示应用名
     */
    function setPageTitle(title: string) {
      pageTitle.value = title
      document.title = title ? `${title} - ${appName.value}` : appName.value
    }

    /**
     * 设置全局页面加载态
     *
     * @param loading - 是否展示加载中
     */
    function setPageLoading(loading: boolean) {
      pageLoading.value = loading
    }

    // ==================== 方法：侧边栏 ====================

    /** 切换侧边栏显示/隐藏（变更由 persist 自动落盘） */
    function toggleSidebar() {
      showSidebar.value = !showSidebar.value
    }

    // ==================== 方法：主布局页签 ====================

    /**
     * 整表替换页签列表；激活 id 若已不在列表中，由对 `layoutTabs` 的 `watch` 自动回退。
     *
     * @param tabs - 新的页签数组
     */
    function setLayoutTabs(tabs: GTabsTabItem[]) {
      layoutTabs.value = tabs
    }

    /**
     * 设置当前激活页签
     *
     * @param tabId - 与某一项 `tabId` 一致
     */
    function setLayoutActiveTabId(tabId: string) {
      layoutActiveTabId.value = tabId
    }

    /**
     * 按路由 `fullPath` 打开或激活页签（菜单、前进/后退等统一入口）。
     *
     * @param fullPath - 作为 `tabId` 与 `path`，与 `router` 一致
     * @param title - 展示标题
     * @param icon - 可选图标名或类名（由布局消费）
     */
    function upsertLayoutTab(fullPath: string, title: string, icon?: string) {
      const existing = layoutTabs.value.find((t) => t.tabId === fullPath)
      if (existing) {
        layoutActiveTabId.value = fullPath
        return
      }
      layoutTabs.value.push({
        tabId: fullPath,
        title,
        path: fullPath,
        icon: icon || undefined,
        fixed: false,
        closable: true,
      })
      layoutActiveTabId.value = fullPath
    }

    /**
     * 判断是否仅命中主布局壳路由（无业务子路由），此时应展示欢迎占位而非页签内容。
     */
    function isMainLayoutShellOnly(route: RouteLocationNormalizedLoaded): boolean {
      const leaf = route.matched[route.matched.length - 1]
      return leaf?.name === 'mainLayout'
    }

    /**
     * 清空主布局页签与激活 id，用于回到 SPA 根（欢迎页）等与「无打开模块」一致的状态。
     */
    function clearLayoutTabsForWelcome() {
      layoutTabs.value = []
      layoutActiveTabId.value = ''
    }

    /**
     * 路由变化时同步页签（编程式导航、地址栏、前进后退等）。
     * 仅命中主布局壳、无子页时不写入页签（由调用方配合 `clearLayoutTabsForWelcome`）。
     *
     * @param route - 当前路由对象
     */
    function openOrActivateTabFromRoute(route: RouteLocationNormalizedLoaded) {
      if (isMainLayoutShellOnly(route)) return
      const title = (route.meta?.title as string) || String(route.name || route.path)
      const icon = route.meta?.icon as string | undefined
      upsertLayoutTab(route.fullPath, title, icon)
    }

    return {
      // 状态
      appName,
      appVersion,
      pageTitle,
      pageLoading,
      layoutTabs,
      layoutActiveTabId,
      showSidebar,
      // 计算属性
      activeLayoutTab,
      // 方法
      setAppVersion,
      setPageTitle,
      setPageLoading,
      toggleSidebar,
      setLayoutTabs,
      setLayoutActiveTabId,
      upsertLayoutTab,
      isMainLayoutShellOnly,
      clearLayoutTabsForWelcome,
      openOrActivateTabFromRoute,
    }
  },
  {
    persist: {
      key: GLOBAL_PERSIST_KEY,
      pick: ['showSidebar'],
    },
  },
)
