<!--
  GTabs - 标签页导航（与 xirang-tabs 一致）
  支持多标签、拖拽排序、关闭、右键菜单、溢出下拉，不依赖 Naive UI NTabs
-->
<template>
  <div class="g-tabs" :class="tabsClass">
    <div class="g-tabs-nav">
      <div class="g-tabs-nav-wrap">
        <div
          ref="tabsListRef"
          class="g-tabs-nav-list"
        >
          <div
            v-for="(tab, index) in props.tabs"
            :key="tab.tabId"
            class="g-tabs-tab"
            :class="{
              'is-active': tab.tabId === internalActiveTabId,
              'is-fixed': tab.fixed,
              'is-dragging': draggingKey === tab.tabId
            }"
            :draggable="draggable && !tab.fixed"
            @click="() => handleTabClick(tab)"
            @contextmenu.prevent="handleTabContextMenu($event, tab)"
            @dragstart="(e: DragEvent) => handleDragStart(e, tab, index)"
            @dragover.prevent="handleDragOver"
            @dragend="handleDragEnd"
            @drop="(e: DragEvent) => handleDrop(e, index)"
          >
            <div class="g-tabs-tab-content">
              <GIcon
                v-if="tab.icon"
                :icon="tab.icon"
                :color="tab.iconColor"
                size="small"
                class="g-tabs-tab-icon"
              />
              <NTooltip
                v-if="isTitleOverflow(tab.title)"
                trigger="hover"
                placement="bottom"
                :delay="500"
              >
                <template #trigger>
                  <span class="g-tabs-tab-title">{{ tab.title }}</span>
                </template>
                {{ tab.title }}
              </NTooltip>
              <span v-else class="g-tabs-tab-title">{{ tab.title }}</span>
              <GIcon
                v-if="(closable || tab.closable) && !tab.fixed"
                icon="CloseOutline"
                size="medium"
                class="g-tabs-tab-close"
                @click.stop="handleClose(tab)"
              />
            </div>
          </div>
        </div>
      </div>

      <!-- 溢出时下拉选择 -->
      <div v-if="showOverflowDropdown" class="g-tabs-dropdown">
        <NSelect
          v-model:value="dropdownValue"
          class="g-tabs-dropdown-select"
          size="small"
          :options="dropdownOptions"
          :render-label="renderDropdownLabel"
          @update:value="handleDropdownChange"
        />
      </div>
    </div>

    <GContext
      v-if="contextMenu"
      v-model:show="showContextMenu"
      :x="contextMenuX"
      :y="contextMenuY"
      :options="contextMenuItems"
      @select="handleContextMenuSelect"
    />
  </div>
</template>

<script setup lang="ts">
import { GContext, GIcon } from '@/components'
import type { GContextMenuItem } from '@/components/gcontext/types'
import { NSelect, NTooltip } from 'naive-ui'
import type { SelectOption } from 'naive-ui'
import { ref, computed, watch, nextTick, onMounted, onBeforeUnmount, h } from 'vue'
import type { GTabsTabItem, GTabsProps, GTabsEmits, GTabsInstance } from './types'

defineOptions({
  name: 'GTabs',
})

const props = withDefaults(defineProps<GTabsProps>(), {
  tabs: () => [],
  activeTabId: '',
  type: 'line',
  draggable: true,
  closable: true,
  contextMenu: true,
  maxTabs: 20,
})

const emit = defineEmits<GTabsEmits>()

// ==================== 内部状态 ====================

const internalActiveTabId = ref(props.activeTabId)
const tabsListRef = ref<HTMLElement | null>(null)
const showOverflowDropdown = ref(false)
const dropdownValue = ref(props.activeTabId)
let lastAutoScrollKey = ''
let lastAutoScrollAt = 0

const draggingKey = ref('')
const dragFromIndex = ref(-1)

const showContextMenu = ref(false)
const contextMenuX = ref(0)
const contextMenuY = ref(0)
const contextMenuTab = ref<GTabsTabItem | null>(null)
const contextMenuItems = ref<GContextMenuItem[][]>([])

// ==================== 计算属性 ====================

const tabsClass = computed(() => ({
  [`g-tabs--${props.type}`]: true,
  'g-tabs--draggable': props.draggable,
}))

const dropdownOptions = computed<SelectOption[]>(() =>
  (props.tabs || []).map((tab) => ({ label: tab.title, value: tab.tabId }))
)

function renderDropdownLabel(option: SelectOption, _selected?: boolean) {
  const tab = props.tabs?.find((t) => t.tabId === option.value)
  if (!tab) return option.label as string
  return h('div', { class: 'g-tabs-dropdown-option' }, [
    tab.icon ? h(GIcon, { icon: tab.icon, size: 'small', class: 'g-tabs-dropdown-option-icon' }) : null,
    h('span', { class: 'g-tabs-dropdown-option-title' }, tab.title),
  ])
}

function generateContextMenuItems(tab: GTabsTabItem): GContextMenuItem[][] {
  const index = props.tabs.findIndex((t) => t.tabId === tab.tabId)
  let nonFixedCount = 0
  let hasLeftNonFixed = false
  let hasRightNonFixed = false
  for (let i = 0; i < props.tabs.length; i++) {
    const t = props.tabs[i]
    if (!t.fixed) {
      nonFixedCount++
      if (i < index) hasLeftNonFixed = true
      if (i > index) hasRightNonFixed = true
    }
  }
  const hasOthers = nonFixedCount > 1

  return [
    [{ code: 'close', name: '关闭', disabled: tab.fixed, prefixIcon: 'CloseOutline' }],
    [
      { code: 'close-others', name: '关闭其他', disabled: !hasOthers, prefixIcon: 'CloseOutline' },
      { code: 'close-left', name: '关闭左侧', disabled: !hasLeftNonFixed, prefixIcon: 'ChevronBackOutline' },
      { code: 'close-right', name: '关闭右侧', disabled: !hasRightNonFixed, prefixIcon: 'ChevronForwardOutline' },
    ],
    [{ code: 'close-all', name: '关闭全部', disabled: nonFixedCount === 0, prefixIcon: 'TrashOutline' }],
  ]
}

// ==================== 监听 ====================

watch(
  () => props.activeTabId,
  (newId) => {
    internalActiveTabId.value = newId
    dropdownValue.value = newId
    nextTick(() => ensureTabVisible(newId))
  }
)

let checkOverflowTimer: ReturnType<typeof setTimeout> | null = null
watch(
  () => props.tabs.length,
  () => {
    if (checkOverflowTimer) clearTimeout(checkOverflowTimer)
    checkOverflowTimer = setTimeout(() => {
      nextTick(() => checkOverflow())
    }, 150)
  }
)

// ==================== 生命周期 ====================

let resizeObserver: ResizeObserver | null = null

onMounted(() => {
  if (tabsListRef.value) {
    tabsListRef.value.addEventListener('wheel', handleWheel, { passive: false })
    resizeObserver = new ResizeObserver(() => {
      if (checkOverflowTimer) clearTimeout(checkOverflowTimer)
      checkOverflowTimer = setTimeout(() => checkOverflow(), 150)
    })
    resizeObserver.observe(tabsListRef.value)
  }
  nextTick(() => checkOverflow())
})

onBeforeUnmount(() => {
  if (tabsListRef.value) tabsListRef.value.removeEventListener('wheel', handleWheel)
  if (resizeObserver) {
    resizeObserver.disconnect()
    resizeObserver = null
  }
  if (checkOverflowTimer) {
    clearTimeout(checkOverflowTimer)
    checkOverflowTimer = null
  }
})

// ==================== 方法 ====================

function handleTabClick(tab: GTabsTabItem, shouldScroll = false) {
  if (tab.tabId === internalActiveTabId.value) return
  internalActiveTabId.value = tab.tabId
  emit('update:activeTabId', tab.tabId)
  emit('change', tab.tabId)
  if (shouldScroll) nextTick(() => scrollToTab(tab.tabId))
}

function handleClose(tab: GTabsTabItem) {
  if (tab.fixed) return
  const index = props.tabs.findIndex((t) => t.tabId === tab.tabId)
  if (index === -1) return
  if (tab.tabId === internalActiveTabId.value) {
    const nextTab = props.tabs[index + 1] || props.tabs[index - 1]
    if (nextTab) {
      internalActiveTabId.value = nextTab.tabId
      emit('update:activeTabId', nextTab.tabId)
      emit('change', nextTab.tabId)
    }
  }
  const newTabs = [...props.tabs]
  newTabs.splice(index, 1)
  emit('update:tabs', newTabs)
  emit('close', tab.tabId)
}

function handleWheel(e: WheelEvent) {
  if (!tabsListRef.value) return
  e.preventDefault()
  tabsListRef.value.scrollLeft += e.deltaY
}

function checkOverflow() {
  if (!tabsListRef.value) return
  showOverflowDropdown.value = tabsListRef.value.scrollWidth > tabsListRef.value.clientWidth
}

function handleDropdownChange() {
  const tab = props.tabs.find((t) => t.tabId === dropdownValue.value)
  if (tab) handleTabClick(tab, true)
}

function scrollToTab(key: string) {
  if (!tabsListRef.value) return
  // 避免与外部触发/下拉触发导致短时间重复动画
  if (lastAutoScrollKey === key && Date.now() - lastAutoScrollAt < 400) return
  lastAutoScrollKey = key
  lastAutoScrollAt = Date.now()
  const index = props.tabs.findIndex((t) => t.tabId === key)
  if (index === -1) return
  const targetTab = tabsListRef.value.querySelector(`.g-tabs-tab:nth-child(${index + 1})`) as HTMLElement
  if (!targetTab) return
  const containerWidth = tabsListRef.value.clientWidth
  const tabLeft = targetTab.offsetLeft
  const tabWidth = targetTab.offsetWidth
  const scrollLeft = tabLeft - containerWidth / 2 + tabWidth / 2
  tabsListRef.value.scrollTo({ left: Math.max(0, scrollLeft), behavior: 'smooth' })
}

function ensureTabVisible(key: string) {
  if (!tabsListRef.value) return
  const index = props.tabs.findIndex((t) => t.tabId === key)
  if (index === -1) return
  const targetTab = tabsListRef.value.querySelector(`.g-tabs-tab:nth-child(${index + 1})`) as HTMLElement
  if (!targetTab) return

  const container = tabsListRef.value
  const visibleLeft = container.scrollLeft
  const visibleRight = visibleLeft + container.clientWidth

  const tabLeft = targetTab.offsetLeft
  const tabRight = tabLeft + targetTab.offsetWidth
  // 如果已在可视区内，则不触发滚动，避免“点到了还滚一遍”
  if (tabLeft >= visibleLeft + 4 && tabRight <= visibleRight - 4) return

  scrollToTab(key)
}

function isTitleOverflow(title: string) {
  return title.length > 12
}

// 拖拽
function handleDragStart(e: DragEvent, tab: GTabsTabItem, index: number) {
  if (!props.draggable || tab.fixed) return
  draggingKey.value = tab.tabId
  dragFromIndex.value = index
  if (e.dataTransfer) {
    e.dataTransfer.effectAllowed = 'move'
    e.dataTransfer.setData('text/plain', tab.tabId)
  }
}

function handleDragOver(e: DragEvent) {
  if (!props.draggable || dragFromIndex.value === -1) return
  e.preventDefault()
  if (e.dataTransfer) e.dataTransfer.dropEffect = 'move'
}

function handleDrop(e: DragEvent, toIndex: number) {
  if (!props.draggable || dragFromIndex.value === -1) return
  e.preventDefault()
  const fromIndex = dragFromIndex.value
  if (fromIndex === toIndex) return
  if (props.tabs[toIndex]?.fixed) return
  const newTabs = [...props.tabs]
  const [movedTab] = newTabs.splice(fromIndex, 1)
  newTabs.splice(toIndex, 0, movedTab)
  emit('update:tabs', newTabs)
  emit('sort', newTabs)
}

function handleDragEnd() {
  draggingKey.value = ''
  dragFromIndex.value = -1
}

// 右键菜单（须接收 $event，与 XiRang xirang-tabs 的 handleContextMenu($event, tab) 一致）
function handleTabContextMenu(e: MouseEvent, tab: GTabsTabItem) {
  if (!props.contextMenu) return
  contextMenuTab.value = tab
  contextMenuX.value = e.clientX
  contextMenuY.value = e.clientY
  contextMenuItems.value = generateContextMenuItems(tab)
  showContextMenu.value = true
}

function handleContextMenuSelect(item: GContextMenuItem, _event?: MouseEvent) {
  if (!contextMenuTab.value) return
  const tab = contextMenuTab.value
  switch (item.code) {
    case 'close':
      handleClose(tab)
      break
    case 'close-others':
      closeOthers(tab.tabId)
      break
    case 'close-left':
      closeLeft(tab.tabId)
      break
    case 'close-right':
      closeRight(tab.tabId)
      break
    case 'close-all':
      closeAll()
      break
  }
  emit('context-menu', item.code, tab.tabId)
}

// 暴露方法
function addTab(tab: GTabsTabItem) {
  const existIndex = props.tabs.findIndex((t) => t.tabId === tab.tabId)
  if (existIndex !== -1) {
    internalActiveTabId.value = tab.tabId
    emit('update:activeTabId', tab.tabId)
    emit('change', tab.tabId)
    return
  }
  if (props.tabs.length >= props.maxTabs) {
    window.$gMessage?.warning?.(`最多只能打开 ${props.maxTabs} 个标签页`)
    return
  }
  const newTabs = [...props.tabs, tab]
  internalActiveTabId.value = tab.tabId
  emit('update:tabs', newTabs)
  emit('update:activeTabId', tab.tabId)
  emit('change', tab.tabId)
}

function removeTab(key: string) {
  const tab = props.tabs.find((t) => t.tabId === key)
  if (tab) handleClose(tab)
}

function closeOthers(key: string) {
  const newTabs = props.tabs.filter((t) => t.tabId === key || t.fixed)
  emit('update:tabs', newTabs)
  if (internalActiveTabId.value !== key) {
    internalActiveTabId.value = key
    emit('update:activeTabId', key)
    emit('change', key)
  }
}

function closeLeft(key: string) {
  const index = props.tabs.findIndex((t) => t.tabId === key)
  if (index === -1) return
  emit('update:tabs', props.tabs.filter((t, i) => i >= index || t.fixed))
}

function closeRight(key: string) {
  const index = props.tabs.findIndex((t) => t.tabId === key)
  if (index === -1) return
  emit('update:tabs', props.tabs.filter((t, i) => i <= index || t.fixed))
}

function closeAll() {
  const fixedTabs = props.tabs.filter((t) => t.fixed)
  emit('update:tabs', fixedTabs)
  if (fixedTabs.length > 0 && internalActiveTabId.value) {
    const hasActive = fixedTabs.some((t) => t.tabId === internalActiveTabId.value)
    if (!hasActive) {
      internalActiveTabId.value = fixedTabs[0].tabId
      emit('update:activeTabId', fixedTabs[0].tabId)
      emit('change', fixedTabs[0].tabId)
    }
  }
}

function activateTab(key: string, shouldScroll = true) {
  const tab = props.tabs.find((t) => t.tabId === key)
  if (tab) handleTabClick(tab, shouldScroll)
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
  display: flex;
  flex-direction: column;
  background: var(--g-bg-primary);
  border-bottom: 1px solid var(--g-border-primary);
}

.g-tabs-nav {
  display: flex;
  align-items: center;
  position: relative;
  height: 40px;
}

.g-tabs-nav-wrap {
  flex: 1;
  overflow: hidden;
  position: relative;
}

.g-tabs-nav-list {
  display: flex;
  align-items: center;
  overflow-x: auto;
  overflow-y: hidden;
  scroll-behavior: smooth;
  gap: 2px;
  scrollbar-width: none;
  -ms-overflow-style: none;

  &::-webkit-scrollbar {
    display: none;
  }
}

.g-tabs-tab {
  display: flex;
  align-items: center;
  height: 40px;
  padding: 0 var(--g-padding-sm) 0 var(--g-padding-md);
  cursor: pointer;
  user-select: none;
  position: relative;
  flex-shrink: 0;
  transition: all var(--g-transition-base) var(--g-transition-ease);
  min-width: 100px;
  max-width: 180px;
  border-radius: var(--g-radius-md) var(--g-radius-md) 0 0;

  &:hover {
    background: var(--g-hover-overlay);

    .g-tabs-tab-close {
      opacity: 1;
    }
  }

  &.is-active {
    color: var(--g-primary);
    background: var(--g-bg-secondary);

    &::after {
      content: '';
      position: absolute;
      bottom: 0;
      left: 8px;
      right: 8px;
      height: 2px;
      background: var(--g-primary);
      border-radius: 2px 2px 0 0;
    }

    .g-tabs-tab-close {
      opacity: 1;
    }
  }

  &.is-dragging {
    opacity: 0.5;
    cursor: grabbing;
  }

  &.is-fixed {
    cursor: pointer;
  }
}

.g-tabs-tab-content {
  display: flex;
  align-items: center;
  gap: var(--g-space-xs);
  width: 100%;
  overflow: hidden;
}

.g-tabs-tab-icon {
  flex-shrink: 0;
}

.g-tabs-tab-title {
  flex: 1;
  font-size: 14px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.g-tabs-tab-close {
  flex-shrink: 0;
  opacity: 0;
  transition: all var(--g-transition-base) var(--g-transition-ease);
  padding: 4px;
  border-radius: var(--g-radius-sm);
  cursor: pointer;
  margin-left: auto;

  &:hover {
    background: rgba(239, 68, 68, 0.1);
    color: var(--g-error);
    transform: scale(1.15);
  }

  &:active {
    transform: scale(0.9);
  }
}

.g-tabs-dropdown {
  display: flex;
  align-items: center;
  padding: 0 var(--g-padding-sm);
  border-left: 1px solid var(--g-border-primary);
  min-width: 160px;
  max-width: 240px;
}

.g-tabs-dropdown-select {
  width: 100%;
}

.g-tabs--card {
  .g-tabs-tab {
    border: 1px solid var(--g-border-primary);
    border-bottom: none;
    border-radius: var(--g-radius-md) var(--g-radius-md) 0 0;
    margin-right: 2px;

    &.is-active {
      background: var(--g-bg-primary);
      border-color: var(--g-border-primary);
      border-bottom-color: transparent;

      &::after {
        display: none;
      }
    }
  }
}

.g-tabs--draggable {
  .g-tabs-tab:not(.is-fixed) {
    cursor: grab;

    &:active {
      cursor: grabbing;
    }
  }
}
</style>

<style lang="scss">
.g-tabs-dropdown-option {
  display: flex;
  align-items: center;
  gap: var(--g-space-xs);
  width: 100%;
  padding: var(--g-padding-xs) var(--g-padding-sm);

  &-icon {
    flex-shrink: 0;
    color: var(--g-text-secondary);
  }

  &-title {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 14px;
  }
}
</style>
