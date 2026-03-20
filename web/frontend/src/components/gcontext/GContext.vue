<template>
  <vxe-context-menu
    ref="contextMenuRef"
    v-model="internalShow"
    :x="props.x"
    :y="props.y"
    :options="vxeOptions"
    :z-index="props.zIndex"
    :class-name="props.className"
    :transfer="props.transfer"
    v-on="vxeEvents"
  />
</template>

<script setup lang="ts">
import { renderIconVNode } from '@/components/gicon'
import { store } from '@/stores'
import { computed, ref, watch } from 'vue'
import type { VxeContextMenuDefines, VxeContextMenuInstance } from 'vxe-pc-ui'
import type { GContextEmits, GContextMenuItem, GContextProps } from './types'

defineOptions({
  name: 'GContext',
  inheritAttrs: false,
})

const props = withDefaults(defineProps<GContextProps>(), {
  show: false,
  options: () => [],
  x: 0,
  y: 0,
  showIcon: true,
  showShortcut: true,
  zIndex: 5000,
  transfer: true,
})

const emit = defineEmits<GContextEmits>()

const contextMenuRef = ref<VxeContextMenuInstance>()
const vxeOptions = ref<VxeContextMenuDefines.MenuFirstOption[][]>([])

const internalShow = computed({
  get: () => props.show,
  set: (val) => {
    emit('update:show', val)
    if (!val) emit('close')
  },
})

function convertToVxeMenuItem(item: GContextMenuItem): VxeContextMenuDefines.MenuFirstOption {
  const vxeItem: VxeContextMenuDefines.MenuFirstOption = {
    code: item.code,
    name: item.name ?? '',
    visible: item.visible !== false,
    disabled: item.disabled ?? false,
    className: item.type === 'divider' ? 'g-context__divider' : item.type === 'group' ? 'g-context__group' : undefined,
  }

  const prefixIcon = item.prefixIcon ?? item.icon
  if (props.showIcon && prefixIcon) {
    vxeItem.prefixIcon = () => renderIconVNode(prefixIcon)()!
  }
  if (props.showIcon && item.suffixIcon) {
    vxeItem.suffixIcon = () => renderIconVNode(item.suffixIcon)()!
  }

  if (props.showShortcut && item.shortcut) {
    vxeItem.suffixConfig = vxeItem.suffixConfig ?? {}
    vxeItem.suffixConfig.content = item.shortcut
  }

  if (item.type === 'divider') {
    vxeItem.name = ''
  }

  if (item.children?.length) {
    vxeItem.children = item.children.map((child) => {
      const childItem: VxeContextMenuDefines.MenuChildOption = {
        code: child.code,
        name: child.name ?? '',
        visible: child.visible !== false,
        disabled: child.disabled ?? false,
      }
      const childPrefix = child.prefixIcon ?? child.icon
      if (props.showIcon && childPrefix) {
        childItem.prefixIcon = () => renderIconVNode(childPrefix)()!
      }
      if (props.showIcon && child.suffixIcon) {
        childItem.suffixIcon = () => renderIconVNode(child.suffixIcon)()!
      }
      if (props.showShortcut && child.shortcut) {
        childItem.suffixConfig = { content: child.shortcut }
      }
      return childItem
    })
  }

  return vxeItem
}

const COPY_CODES = ['copyNode', 'copyRow', 'copyCell']

function filterByPermission(items: GContextMenuItem[], moduleId: string): GContextMenuItem[] {
  return items
    .filter((m) => {
      if (!m.code) return true
      if (COPY_CODES.includes(m.code)) return true
      return store.user.hasButton(`${moduleId}:${m.code}`)
    })
    .map((m) => {
      if (!m.children?.length) return m
      return { ...m, children: filterByPermission(m.children, moduleId) }
    })
}

function filterOptionsByPermission(
  items: GContextMenuItem[] | GContextMenuItem[][] | undefined,
  moduleId?: string
): GContextMenuItem[] | GContextMenuItem[][] | undefined {
  if (!items?.length) return items
  if (!moduleId) return items
  if (Array.isArray(items[0])) {
    return (items as GContextMenuItem[][]).map((g) => filterByPermission(g, moduleId))
  }
  return filterByPermission(items as GContextMenuItem[], moduleId)
}

function buildVxeOptions(
  items: GContextMenuItem[] | GContextMenuItem[][] | undefined
): VxeContextMenuDefines.MenuFirstOption[][] {
  if (!items?.length) return []

  if (Array.isArray(items[0])) {
    const grouped = items as GContextMenuItem[][]
    return grouped
      .filter((g) => g?.length)
      .map((g) => g.filter((i) => i.visible !== false).map(convertToVxeMenuItem))
      .filter((g) => g.length)
  }

  const flat = items as GContextMenuItem[]
  const converted = flat.filter((i) => i.visible !== false).map(convertToVxeMenuItem)
  return converted.length ? [converted] : []
}

function findMenuItem(
  items: GContextMenuItem[] | GContextMenuItem[][],
  targetCode: string
): GContextMenuItem | null {
  if (!items?.length) return null
  if (Array.isArray(items[0])) {
    for (const group of items as GContextMenuItem[][]) {
      const found = findMenuItem(group, targetCode)
      if (found) return found
    }
    return null
  }
  for (const item of items as GContextMenuItem[]) {
    if (item.code === targetCode) return item
    if (item.children) {
      const found = findMenuItem(item.children, targetCode)
      if (found) return found
    }
  }
  return null
}

watch(
  [() => props.options, () => props.moduleId, () => props.showIcon, () => props.showShortcut],
  () => {
    const filtered = filterOptionsByPermission(props.options, props.moduleId)
    vxeOptions.value = buildVxeOptions(filtered)
  },
  { immediate: true }
)

const vxeEvents = {
  optionClick(params: { option?: { code?: string }; $event?: MouseEvent }) {
    const code = params.option?.code
    if (code == null) return
    const original = findMenuItem(props.options ?? [], code)
    if (original && !original.disabled) {
      emit('select', original, params.$event!)
      hide()
    }
  },
}

function show() {
  emit('update:show', true)
}

function hide() {
  emit('update:show', false)
  emit('close')
}

function toggle() {
  if (props.show) hide()
  else show()
}

defineExpose({
  show,
  hide,
  toggle,
  open: () => contextMenuRef.value?.open(),
  close: () => contextMenuRef.value?.close(),
})
</script>

<style lang="scss">
.vxe-context-menu--item-prefix span{
  display: flex;
  align-items: center;
}
</style>
