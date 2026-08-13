<!--
  GTabs：布局顶栏多页签薄封装。
  导航能力交给 RsTabs；本层只做 GTabsTabItem ↔ RsTabItem 与列表 v-model。
-->
<template>
  <RsTabs
    v-if="rsItems.length > 0"
    v-model="activeValue"
    class="g-tabs"
    :items="rsItems"
    :variant="type === 'card' ? 'card' : 'line'"
    :size="size"
    panelless
    borderless
    :closable="closable"
    :draggable="draggable"
    :context-menu="contextMenu"
    overflow="dropdown"
    :max-count="maxTabs"
    @close="onClose"
    @close-batch="onCloseBatch"
    @reorder="onReorder"
    @context-menu="onContextMenu"
  />
  <div v-else class="g-tabs g-tabs--empty" aria-hidden="true" />
</template>

<script setup lang="ts">
import { useAppMessage } from '@/composables/useAppMessage'
import { RsTabs } from '@/ui'
import type { RsTabItem, RsTabsCloseAction } from 'niuma-ui'
import { computed } from 'vue'
import type { GTabsEmits, GTabsInstance, GTabsProps, GTabsTabItem } from './types'

defineOptions({
  name: 'GTabs',
})

const props = withDefaults(defineProps<GTabsProps>(), {
  tabs: () => [],
  activeTabId: '',
  type: 'line',
  size: 'md',
  draggable: true,
  closable: true,
  contextMenu: true,
  maxTabs: 20,
})

const emit = defineEmits<GTabsEmits>()
const message = useAppMessage()

const rsItems = computed<RsTabItem[]>(() =>
  (props.tabs ?? []).map((tab) => ({
    value: tab.tabId,
    label: tab.title,
    icon: typeof tab.icon === 'string' && !/[A-Z]/.test(tab.icon) ? tab.icon : undefined,
    closable: tab.closable,
    fixed: tab.fixed,
  })),
)

const activeValue = computed({
  get: () => props.activeTabId || '',
  set: (value: string) => {
    if (value === props.activeTabId) return
    emit('update:activeTabId', value)
    emit('change', value)
  },
})

function removeTabIds(ids: string[]) {
  if (!ids.length) return
  const idSet = new Set(ids)
  const next = (props.tabs ?? []).filter((t) => !idSet.has(t.tabId) || t.fixed)
  emit('update:tabs', next)
}

function toContextMenuCode(action: RsTabsCloseAction): string {
  switch (action) {
    case 'others':
      return 'close-others'
    case 'left':
      return 'close-left'
    case 'right':
      return 'close-right'
    case 'all':
      return 'close-all'
    default:
      return 'close'
  }
}

function onClose(value: string) {
  removeTabIds([value])
  emit('close', value)
}

function onCloseBatch(values: string[], action: RsTabsCloseAction) {
  removeTabIds(values)
  emit('context-menu', toContextMenuCode(action), values[0] ?? '')
}

function onReorder(dragValue: string, dropValue: string) {
  const tabs = [...(props.tabs ?? [])]
  const from = tabs.findIndex((t) => t.tabId === dragValue)
  const to = tabs.findIndex((t) => t.tabId === dropValue)
  if (from < 0 || to < 0 || from === to) return
  if (tabs[to]?.fixed) return
  const [moved] = tabs.splice(from, 1)
  tabs.splice(to, 0, moved)
  emit('update:tabs', tabs)
  emit('sort', tabs)
}

function onContextMenu(action: RsTabsCloseAction, value: string) {
  emit('context-menu', toContextMenuCode(action), value)
}

function addTab(tab: GTabsTabItem) {
  const exist = props.tabs?.find((t) => t.tabId === tab.tabId)
  if (exist) {
    emit('update:activeTabId', tab.tabId)
    emit('change', tab.tabId)
    return
  }
  if ((props.tabs?.length ?? 0) >= props.maxTabs) {
    message.warning(`最多只能打开 ${props.maxTabs} 个标签页`)
    return
  }
  emit('update:tabs', [...(props.tabs ?? []), tab])
  emit('update:activeTabId', tab.tabId)
  emit('change', tab.tabId)
}

function removeTab(key: string) {
  const tabs = props.tabs ?? []
  const index = tabs.findIndex((t) => t.tabId === key)
  if (index < 0 || tabs[index]?.fixed) return
  const next = tabs.filter((t) => t.tabId !== key)
  emit('update:tabs', next)
  emit('close', key)
  if (props.activeTabId === key) {
    const nextTab = next[index] || next[index - 1] || next[0]
    emit('update:activeTabId', nextTab?.tabId ?? '')
    emit('change', nextTab?.tabId ?? '')
  }
}

function closeOthers(key: string) {
  const next = (props.tabs ?? []).filter((t) => t.tabId === key || t.fixed)
  emit('update:tabs', next)
  if (props.activeTabId !== key) {
    emit('update:activeTabId', key)
    emit('change', key)
  }
}

function closeLeft(key: string) {
  const index = props.tabs?.findIndex((t) => t.tabId === key) ?? -1
  if (index < 0) return
  emit(
    'update:tabs',
    (props.tabs ?? []).filter((t, i) => i >= index || t.fixed),
  )
}

function closeRight(key: string) {
  const index = props.tabs?.findIndex((t) => t.tabId === key) ?? -1
  if (index < 0) return
  emit(
    'update:tabs',
    (props.tabs ?? []).filter((t, i) => i <= index || t.fixed),
  )
}

function closeAll() {
  const fixedTabs = (props.tabs ?? []).filter((t) => t.fixed)
  emit('update:tabs', fixedTabs)
  const nextId = fixedTabs[0]?.tabId ?? ''
  if (props.activeTabId !== nextId) {
    emit('update:activeTabId', nextId)
    emit('change', nextId)
  }
}

function activateTab(key: string) {
  if (!(props.tabs ?? []).some((t) => t.tabId === key)) return
  emit('update:activeTabId', key)
  emit('change', key)
}

defineExpose<GTabsInstance>({
  addTab,
  removeTab,
  closeOthers,
  closeLeft,
  closeRight,
  closeAll,
  activateTab,
})
</script>

<style scoped lang="scss">
.g-tabs {
  width: 100%;
  min-width: 0;
}

.g-tabs--empty {
  flex: 1 1 auto;
  min-width: 0;
}
</style>
