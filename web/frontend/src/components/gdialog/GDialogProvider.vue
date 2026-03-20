<template>
  <!--
    插件用 render() 挂到 body，不在 App 的 NConfigProvider 子树内。
    Naive 主题靠 provide/inject，所以这里必须自包 NConfigProvider 透传 config/theme。
     的对话框用 vxe-modal，主题走 CSS 变量 + data-theme，不依赖 provide，故无需包。
  -->
  <n-config-provider
    :theme="naiveTheme"
    :theme-overrides="themeOverrides"
    :locale="naiveLocale"
    :date-locale="naiveDateLocale"
  >
    <Teleport to="body">
      <div class="g-dialog-provider">
        <GDialog
          v-for="item in items"
          :key="item.id"
          :show="true"
          :title="item.options.title"
          :subtitle="item.options.subtitle"
          :subtitle-position="item.options.subtitlePosition ?? 'footer'"
          :icon="item.options.icon"
          :icon-size="item.options.iconSize"
          :header-style="item.options.headerStyle ?? 'default'"
          :width="item.options.width ?? 500"
          :mask-closable="item.options.maskClosable !== false"
          :close-on-esc="item.options.closeOnEsc !== false"
          :show-cancel="item.options.showCancel !== false"
          :show-confirm="item.options.showConfirm !== false"
          :cancel-text="item.options.negativeText ?? '取消'"
          :confirm-text="item.options.positiveText ?? '确定'"
          :confirm-loading="item.options.confirmLoading ?? false"
          :footer-button-align="item.options.footerButtonAlign ?? 'end'"
          :auto-close-on-confirm="false"
          @update:show="(show: boolean) => !show && handleClose(item.id)"
          @confirm="() => handleResolve(item.id, true)"
          @cancel="() => handleResolve(item.id, false)"
          @close="() => handleResolve(item.id, false)"
        >
          <RenderContent v-if="item.contentRender" :render="item.contentRender" />
        </GDialog>
      </div>
    </Teleport>
    <slot />
  </n-config-provider>
</template>

<script setup lang="ts">
/**
 * GDialog Provider：由插件挂载在 App 树外，故自带 NConfigProvider 透传 config/theme。
 *
 * 为何  Message 不用包 NConfigProvider：
 * -  的对话框用 vxe-modal，主题由 CSS 变量（--xr-*）和 document 的 data-theme 控制。
 * - 本项目的 GDialog 用 Naive UI（n-modal、n-card、n-button 等），主题通过 NConfigProvider 的 provide 注入。
 * - render() 挂载的子树拿不到 App 里的 NConfigProvider，所以这里必须再包一层，否则 config/theme 不会生效。
 */
import { darkThemeOverrides, lightThemeOverrides } from '@/config/theme'
import type { LocaleType } from '@/locales'
import { getCurrentLocale } from '@/locales'
import { useUserStore } from '@/stores/user'
import type { GlobalThemeOverrides, NDateLocale, NLocale } from 'naive-ui'
import { darkTheme, dateEnUS, dateZhCN, enUS, zhCN } from 'naive-ui'
import { computed, defineComponent, h, markRaw, onBeforeUnmount, onMounted, ref, type Component, type PropType, type VNode } from 'vue'
import GDialog from './GDialog.vue'
import type { GDialogOptions } from './useGDialog'
import { setDialogProvider, type GDialogProviderApi } from './useGDialog'

type Theme = 'light' | 'dark'
const userStore = useUserStore()

const naiveThemeMap: Record<Theme, typeof darkTheme | null> = {
  light: null,
  dark: darkTheme,
}
const themeOverridesMap: Record<Theme, GlobalThemeOverrides> = {
  light: lightThemeOverrides,
  dark: darkThemeOverrides,
}
const naiveTheme = computed(() => naiveThemeMap[userStore.resolvedTheme])
const themeOverrides = computed(() => themeOverridesMap[userStore.resolvedTheme])

const naiveLocaleMap: Record<LocaleType, NLocale> = { en: enUS, 'zh-CN': zhCN }
const naiveDateLocaleMap: Record<LocaleType, NDateLocale> = { en: dateEnUS, 'zh-CN': dateZhCN }
const naiveLocale = computed(() => naiveLocaleMap[getCurrentLocale()])
const naiveDateLocale = computed(() => naiveDateLocaleMap[getCurrentLocale()])

defineOptions({
  name: 'GDialogProvider',
})

const RenderContent = defineComponent({
  props: {
    render: { type: Function as PropType<() => VNode | null>, required: true },
  },
  setup(props) {
    return () => (props.render ? props.render() : null)
  },
})

interface DialogItem {
  id: string
  options: GDialogOptions
  contentRender: () => VNode | null
  resolve: (value: boolean) => void
}

const items = ref<DialogItem[]>([])
let idCounter = 0

function buildContentRender(content: GDialogOptions['content']): () => VNode | null {
  if (content == null) return () => null
  if (typeof content === 'string') {
    return () =>
      h('div', { style: { whiteSpace: 'pre-line', lineHeight: '1.6' } }, content)
  }
  if (typeof content === 'object' && content && 'render' in content) {
    return () => h(content as Component)
  }
  return () => content as VNode
}

function createDialog(options: GDialogOptions): Promise<boolean> {
  return new Promise((resolve) => {
    const id = `g-dialog-${Date.now()}-${++idCounter}`
    // 避免组件（icon/content）被 ref 深响应式化，否则会触发 Vue 的 "Component made reactive" 警告
    const optionsForItem: GDialogOptions = {
      ...options,
      icon: options.icon ? markRaw(options.icon) : undefined,
      content:
        options.content && typeof options.content === 'object' && 'render' in options.content
          ? markRaw(options.content as Component)
          : options.content,
    }
    const contentRender = buildContentRender(optionsForItem.content)
    const item: DialogItem = {
      id,
      options: optionsForItem,
      contentRender,
      resolve,
    }
    items.value = [...items.value, item]
  })
}

function handleResolve(id: string, value: boolean) {
  const item = items.value.find((i) => i.id === id)
  if (!item) return
  item.resolve(value)
  items.value = items.value.filter((i) => i.id !== id)
}

function handleClose(id: string) {
  handleResolve(id, false)
}

const api: GDialogProviderApi = { createDialog }

onMounted(() => {
  setDialogProvider(api)
})

onBeforeUnmount(() => {
  items.value.forEach((i) => i.resolve(false))
  items.value = []
  setDialogProvider(null)
})

defineExpose({
  createDialog,
})
</script>

