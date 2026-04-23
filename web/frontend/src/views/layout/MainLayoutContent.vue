<!--
  主内容区：与 XiRang AppContent 中 content-wrapper + router-view 职责类似（页签在 MainLayoutHeader）。
  - 监听 `route.fullPath`：主布局内用 URL 对齐页签（刷新、书签、前进后退）；仅命中 mainLayout 壳（无子路由）时清空页签，回到欢迎占位，避免后退到 `/` 仍保留旧页签导致遮罩卡死。
  - 监听 `layoutActiveTabId`，解析对应 tab 的 `path` 做 `router.push`，与侧栏 upsert / 头部切换共用一套路由同步。
  - keep-alive `include` 来自当前 `layoutTabs`（与页签列表一致）。
  - 全页刷新（F5）：`onMounted` + `nextTick` 内用 Performance API 判断，随后 `router.replace('/')` 回到 SPA 根（地址栏即 `VITE_BASE_URL` 对应前缀 + `/`，与 `config.baseUrl` 一致）。
-->
<template>
  <n-layout-content class="main-layout-content">
    <div v-if="!activeLayoutTab" class="main-layout-content__placeholder">
      <div class="main-layout-content__placeholder-inner" role="status" aria-live="polite">
        <div class="main-layout-content__placeholder-glow" aria-hidden="true" />
        <div class="main-layout-content__placeholder-card">
          <div class="main-layout-content__placeholder-lockup" aria-hidden="true">
            <n-icon :size="34" :component="LayersOutline" />
          </div>

          <h2 class="main-layout-content__placeholder-title">
            {{ tLogin('login.welcomeTitle') }}
          </h2>
          <p class="main-layout-content__placeholder-subtitle">
            {{ tLogin('login.welcomeSubtitle') }}
          </p>
        </div>
      </div>
    </div>
    <!--
      使用 router-view 的 slot 拿到实际渲染组件与子路由信息：
      - Component：当前命中的子路由组件（可能是异步组件）
      - childRoute：用于读取 meta / fullPath 等信息做缓存与 key 控制
    -->
    <!--
      必须与当前激活页签 path 一致后再挂 router-view：
      否则在「关光页签后 URL 仍停在旧页」再点菜单开新页签时，会先按旧 URL 渲染一帧，
      误挂载已关闭页的组件（如 SystemMonitoring 的 onMounted 发请求），随后 watch 才 push 到新路由。
    -->
    <RouteViewLoadingMask v-else-if="!isLayoutContentRouteSynced" />
    <router-view v-else v-slot="{ Component, route: childRoute }">
      <template v-if="Component">
        <!--
          Suspense 用于覆盖“组件异步加载 / async setup”期间的空白：
          - default：组件 ready 后渲染真实页面
          - fallback：仅覆盖主内容区的加载态，不影响侧边栏/头部等其它区域交互
        -->
        <Suspense>
          <template #default>
            <!--
              keep-alive：按路由 meta 控制是否缓存页面
              - include 使用当前页签路径集合，确保关闭页签后能释放对应缓存
              - wrapWithCacheKey：为缓存包一层“稳定组件外壳”，避免同一组件被不同 fullPath 复用时串缓存
              - :key 使用 fullPath，确保 query/hash 改变时能正确区分页面实例
            -->
            <keep-alive
              v-if="shouldKeepAliveRoute(childRoute)"
              :max="MAX_CACHED_VIEWS"
              :include="cachedTabPaths"
            >
              <component
                :is="wrapWithCacheKey(Component, childRoute.fullPath)"
                :key="layoutViewCacheKey(childRoute)"
              />
            </keep-alive>
            <!-- 不需要缓存的页面直接渲染，同样用 fullPath 做 key 保证切换一致性 -->
            <component
              v-else
              :is="Component"
              :key="layoutViewCacheKey(childRoute)"
            />
          </template>
          <template #fallback>
            <RouteViewLoadingMask />
          </template>
        </Suspense>
      </template>
    </router-view>
  </n-layout-content>
</template>

<script setup lang="ts">
import RouteViewLoadingMask from '@/components/RouteViewLoadingMask.vue'
import { config } from '@/config/config'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { cleanupWrappedCache, wrapWithCacheKey } from '@/router/wrapWithCacheKey'
import { useGlobalStore } from '@/stores/global'
import { LayersOutline } from '@vicons/ionicons5'
import { NIcon, NLayoutContent } from 'naive-ui'
import { storeToRefs } from 'pinia'
import { computed, nextTick, onMounted, watch } from 'vue'
import { useRoute, useRouter, type RouteLocationNormalizedLoaded } from 'vue-router'

const { t: tLogin } = useModuleI18n('hub0001')

const router = useRouter()
const route = useRoute()
const globalStore = useGlobalStore()
const { layoutTabs, layoutActiveTabId, activeLayoutTab } = storeToRefs(globalStore)

/** 主内容区期望与地址栏一致的路径（与页签 tabId/path 同源） */
const expectedLayoutContentPath = computed(
  () => activeLayoutTab.value?.tabId || activeLayoutTab.value?.path || '',
)

/**
 * 当前路由是否已与激活页签对齐。
 * 有激活页签但尚未对齐时仅展示内容区遮罩，由下方 watch 触发 `router.push`，避免误挂载旧 URL 对应页面。
 */
const isLayoutContentRouteSynced = computed(() => {
  const expected = expectedLayoutContentPath.value
  if (!expected) return true
  return route.fullPath === expected
})

/** `performance.navigation.type === 1` 表示 reload（旧 API，作 NT2 的补充） */
const LEGACY_NAV_TYPE_RELOAD = 1

function isFullDocumentReload(): boolean {
  if (typeof performance === 'undefined') return false
  const nav = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming | undefined
  if (nav?.type === 'reload') return true
  const legacy = (performance as Performance & { navigation?: { type: number } }).navigation
  return legacy?.type === LEGACY_NAV_TYPE_RELOAD
}

/** 去掉末尾 `/`，便于与 `location.pathname` 比较 */
function stripTrailingSlash(p: string): string {
  const s = p.trim()
  if (!s || s === '/') return ''
  return s.replace(/\/+$/, '')
}

/** 地址栏是否已在「仅 VITE_BASE_URL、无业务子路径」（如 `/gatewayweb` 或 `/gatewayweb/`） */
function isLocationAtAppPublicRoot(): boolean {
  const base = stripTrailingSlash(config.baseUrl || '/')
  const path = stripTrailingSlash(window.location.pathname)
  if (!base) return path === '' || path === '/'
  return path === base
}

/**
 * 全页刷新时回到 SPA 根：`router` 的 `path: '/'` 对应 `createWebHistory(BASE_URL)` 下站点根，
 * 浏览器地址为 `config.baseUrl` + `/`（与 `VITE_BASE_URL` 一致）。
 */
function redirectToAppBaseOnReload() {
  if (isLocationAtAppPublicRoot()) return
  router.replace({ path: '/' }).catch((err: { name?: string }) => {
    if (err?.name !== 'NavigationDuplicated') console.error(err)
  })
}

onMounted(() => {
  const run = () => {
    if (!isFullDocumentReload()) return
    redirectToAppBaseOnReload()
  }
  run()
  void nextTick(run)
})

/**
 * 主布局内：用当前 URL 对齐页签（刷新、书签、前进后退）；回到仅命中 mainLayout 的根路径时清空页签以显示欢迎页。
 * 逻辑放在此组件而非全局 afterEach，避免与占位/遮罩条件割裂。
 */
watch(
  () => route.fullPath,
  () => {
    if (!route.matched.some((r) => r.name === 'mainLayout')) return
    if (globalStore.isMainLayoutShellOnly(route)) {
      globalStore.clearLayoutTabsForWelcome()
      return
    }
    globalStore.openOrActivateTabFromRoute(route)
  },
  { flush: 'post', immediate: true },
)

/** 激活页签变化时按对应 `path` 同步地址栏（侧栏 upsert / 头部切换 id 均走此逻辑） */
watch(
  layoutActiveTabId,
  (id) => {
    if (!id) return
    const tab = layoutTabs.value.find((t) => t.tabId === id)
    if (!tab) return
    const target = tab.path ?? tab.tabId
    if (route.fullPath === target) return
    router.push(target).catch((err: { name?: string }) => {
      if (err?.name !== 'NavigationDuplicated') console.error(err)
    })
  },
  { flush: 'post' },
)

const MAX_CACHED_VIEWS = 20

function shouldKeepAliveRoute(r: RouteLocationNormalizedLoaded) {
  return r.meta?.keepAlive !== false
}

function layoutTabPathKeys(tabs: { tabId: string; path?: string }[]) {
  return tabs.map((t) => t.tabId || t.path || '').filter(Boolean) as string[]
}

watch(
  layoutTabs,
  (tabs) => {
    cleanupWrappedCache(new Set(layoutTabPathKeys(tabs)))
  },
  { immediate: true, deep: true },
)

/**
 * 仍打开页签的路径集合；并并入当前激活页签（避免与列表短暂不同步时出现空 include）。
 * 不在此写死首页 path，首页由菜单/初始 store 写入 tab。
 */
const cachedTabPaths = computed(() => {
  const paths = new Set(layoutTabPathKeys(layoutTabs.value))
  const tab = activeLayoutTab.value
  const p = tab?.tabId || tab?.path
  if (p) paths.add(p)
  return [...paths]
})

function layoutViewCacheKey(r: RouteLocationNormalizedLoaded) {
  return r.fullPath
}
</script>

<style lang="scss" scoped>
.main-layout-content {
  height: calc(100vh - var(--g-header-height));
  border: 0.5px solid var(--g-border-primary);
  border-radius: var(--g-radius-2xl);
  box-sizing: border-box;
  position: relative;
  overflow: hidden;
}

.main-layout-content__placeholder {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  min-height: 200px;
  padding: var(--g-space-lg);
  box-sizing: border-box;
  overflow: auto;
  /* 与侧栏/顶栏区分的次级底，随 light/dark 的 --g-bg-secondary 变化 */
  background-color: var(--g-bg-secondary);
}

.main-layout-content__placeholder-inner {
  position: relative;
  width: 100%;
  max-width: 440px;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.main-layout-content__placeholder-glow {
  position: absolute;
  inset: -20% -30% auto;
  height: min(320px, 55vh);
  border-radius: 50%;
  pointer-events: none;
  background: radial-gradient(
    closest-side,
    color-mix(in srgb, var(--g-primary) 22%, transparent),
    transparent 100%
  );
  filter: blur(48px);
  opacity: 0.85;
}

.main-layout-content__placeholder-card {
  position: relative;
  width: 100%;
  padding: var(--g-space-xl) var(--g-space-lg);
  text-align: center;
  border-radius: var(--g-radius-2xl);
  /* 融入背景：不再以独立白卡/投影凸显块级区域 */
  border: 1px solid transparent;
  background-color: transparent;
  box-shadow: none;
}

.main-layout-content__placeholder-icon {
  margin: 0 auto var(--g-space-md);
  width: 72px;
  height: 72px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--g-radius-xl);
  color: var(--g-primary);
  background-color: color-mix(in srgb, var(--g-primary) 16%, transparent);
}

.main-layout-content__placeholder-title {
  margin: 0 0 var(--g-space-xs);
  font-size: var(--g-font-size-xl);
  font-weight: 700;
  line-height: 1.35;
  color: transparent;
  background: linear-gradient(
    120deg,
    color-mix(in srgb, var(--g-primary) 90%, white) 0%,
    var(--g-primary-hover) 45%,
    var(--g-primary-active) 100%
  );
  -webkit-background-clip: text;
  background-clip: text;
}

.main-layout-content__placeholder-subtitle {
  margin: 0 0 var(--g-space-md);
  font-size: var(--g-font-size-sm);
  line-height: 1.7;
  color: var(--g-text-tertiary);
  max-width: 38em;
  margin-inline: auto;
}
</style>
