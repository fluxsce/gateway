<template>
  <div class="g-tree-wrapper">
    <div v-if="props.filterable" class="g-tree-filter">
      <n-input
        v-model:value="filterKeyword"
        placeholder="搜索节点..."
        clearable
        size="small"
        @input="handleFilter"
      >
        <template #prefix>
          <GIcon icon="SearchOutline" size="small" />
        </template>
      </n-input>
    </div>
    <div class="g-tree-content" @contextmenu="handleContentContextmenu">
      <n-tree
        ref="treeRef"
      :data="displayData"
      :default-expand-all="props.defaultExpandAll"
      :default-expanded-keys="props.defaultExpandedKeys"
      :checked-keys="props.checkedKeys !== undefined ? props.checkedKeys : undefined"
      :default-checked-keys="props.checkedKeys === undefined ? props.defaultCheckedKeys : undefined"
      :checkable="props.checkable"
      :cascade="props.cascade"
      :check-strategy="props.checkStrategy"
      :draggable="props.draggable"
      :allow-drop="mergedAllowDrop"
      :show-line="props.showLine"
      :show-icon="props.showIcon"
      :block-line="props.blockLine"
      :block-node="true"
      :node-props="nodeProps"
      :render-label="renderLabelWrapperComputed"
      :render-prefix="renderPrefixComputed"
      :render-suffix="renderSuffixComputed"
      :render-switcher-icon="renderSwitcherIcon"
      :virtual-scroll="props.virtualScroll"
      :on-load="props.loadMethod ? handleLoad : undefined"
      @update:checked-keys="handleUpdateCheckedKeys"
      @update:expanded-keys="handleUpdateExpandedKeys"
      @select="handleSelect"
      @dblclick="handleDblclick"
      @dragstart="handleDragStart"
      @dragend="handleDragEnd"
      @drop="handleDrop"
      />
    </div>
    <!-- 右键菜单（节点 + 空白区域，参考 ） -->
    <GContext
      v-if="hasContextMenu"
      v-model:show="showContextMenu"
      :x="contextMenuX"
      :y="contextMenuY"
      :options="contextMenuOptionsRef"
      :module-id="contextMenuModuleId"
      @select="handleContextMenuSelect"
      @close="handleContextMenuClose"
    />
  </div>
</template>

<script setup lang="ts">
import type { GContextMenuItem } from '@/components/gcontext'
import { GContext } from '@/components/gcontext'
import { GEllipsis } from '@/components/gellipsis'
import { GIcon, renderIconVNode } from '@/components/gicon'
import { copyToClipboard } from '@/utils'
import type { TreeOption } from 'naive-ui'
import { NTree } from 'naive-ui'
import { computed, h, nextTick, ref, shallowRef, watch } from 'vue'
import type { GTreeEmits, GTreeInstance, GTreeProps } from './types'

defineOptions({
  name: 'GTree'
})

const props = withDefaults(defineProps<GTreeProps>(), {
  data: () => [],
  defaultExpandedKeys: () => [],
  defaultExpandAll: false,
  defaultCheckedKeys: () => [],
  checkable: false,
  cascade: true,
  checkStrategy: 'all',
  draggable: false,
  showLine: false,
  showIcon: false,
  blockLine: true,
  keyField: 'key',
  labelField: 'label',
  childrenField: 'children',
  nodeHeight: 28,
  virtualScroll: false,
  ellipsis: false,
  ellipsisLineClamp: 1,
  ellipsisTooltip: true,
  filterable: false,
  iconOpen: 'ChevronDownOutline',
  iconClose: 'ChevronForwardOutline'
})

const emit = defineEmits<GTreeEmits>()

// 转换数据格式，确保符合 NTree 的要求
const treeData = computed(() => {
  if (!props.data || props.data.length === 0) {
    return []
  }

  // 递归转换函数
  const convertItem = (item: any): TreeOption => {
    const option: TreeOption = {
      key: props.keyField === 'key' ? item.key : (item as any)[props.keyField],
      label: props.labelField === 'label' ? item.label : (item as any)[props.labelField]
    }

    // 处理子节点
    const children = (item as any)[props.childrenField] || item.children
    if (children && Array.isArray(children) && children.length > 0) {
      option.children = children.map(convertItem)
    }

    // 保留其他属性
    Object.keys(item).forEach((key) => {
      if (key !== props.keyField && key !== props.labelField && key !== props.childrenField && key !== 'key' && key !== 'label' && key !== 'children') {
        ;(option as any)[key] = item[key]
      }
    })

    return option
  }

  return props.data.map(convertItem)
})

// 可搜索（参考 ）
const filterKeyword = ref('')
const treeRef = ref<InstanceType<typeof NTree> | null>(null)
const displayData = computed(() => {
  if (!props.filterable || !filterKeyword.value.trim()) return treeData.value
  const keyword = filterKeyword.value.toLowerCase().trim()
  const filterNodes = (nodes: TreeOption[]): TreeOption[] => {
    const result: TreeOption[] = []
    for (const node of nodes) {
      const match = props.filterMethod
        ? props.filterMethod(keyword, node)
        : String((node as any).label ?? node.key ?? '').toLowerCase().includes(keyword)
      const children = node.children ? filterNodes(node.children) : []
      if (match || children.length > 0) {
        result.push({ ...node, children: children.length ? children : node.children })
      }
    }
    return result
  }
  return filterNodes(treeData.value)
})
function handleFilter() {
  if (!filterKeyword.value || !treeRef.value) return
  const collectKeys = (nodes: TreeOption[]): string[] => {
    const keys: string[] = []
    for (const n of nodes) {
      if (n.key != null) keys.push(String(n.key))
      if (n.children?.length) keys.push(...collectKeys(n.children))
    }
    return keys
  }
  const allKeys = collectKeys(treeData.value)
  nextTick(() => {
    ;(treeRef.value as any)?.setExpandedKeys?.(allKeys)
  })
}

// 右键菜单（参考 ：节点右键 + 空白区域右键，选项在打开时确定）
const showContextMenu = ref(false)
const contextMenuX = ref(0)
const contextMenuY = ref(0)
const currentContextNode = ref<TreeOption | null>(null)
// shallowRef 避免菜单项中的 icon 组件被深度响应式化，消除 Vue 的 markRaw 警告
const contextMenuOptionsRef = shallowRef<GContextMenuItem[]>([])

const hasContextMenu = computed(() => {
  const config = props.menuConfig
  if (!config) return false
  if (typeof config === 'function') return true
  return config.enabled !== false
})
const contextMenuModuleId = computed(() => {
  const config = props.menuConfig
  if (!config || typeof config === 'function') return props.moduleId
  return (config as any).moduleId ?? props.moduleId
})

function getMenuItems(node: TreeOption | null): GContextMenuItem[] {
  const config = props.menuConfig
  if (!config) return []
  if (typeof config === 'function') {
    const items = config(node)
    if (!items?.length) return []
    return Array.isArray(items[0]) ? (items as GContextMenuItem[][]).flat() : (items as GContextMenuItem[])
  }
  if ((config as any).enabled === false) return []
  const list: GContextMenuItem[] = []
  if (node && (config as any).showCopyNode !== false) {
    list.push({ code: 'copyNode', name: '复制节点数据', icon: 'CopyOutline' })
  }
  const opts = (config as any).options
  if (opts?.length) {
    const flat = Array.isArray(opts[0]) ? (opts as GContextMenuItem[][]).flat() : (opts as GContextMenuItem[])
    list.push(...flat)
  }
  return list
}

const handleContextMenuClose = () => {
  showContextMenu.value = false
  currentContextNode.value = null
}

// 处理右键菜单选择
const handleContextMenuSelect = (item: { code: string }, _event: MouseEvent) => {
  const code = item.code
  const node = currentContextNode.value

  if (code === 'copyNode' && node) {
    copyToClipboard(JSON.stringify(node, null, 2))
  }

  const config = props.menuConfig
  if (config && typeof config === 'object' && (config as any).onMenuClick) {
    (config as any).onMenuClick({ code, node, row: undefined, column: undefined })
  }
  emit('menu-click', { code, node: node ?? undefined })
  handleContextMenuClose()
}

// 节点属性
const nodeProps = computed(() => {
  return ({ option }: { option: TreeOption }) => {
    const nodePropsObj: Record<string, any> = {
      style: {
        height: `${props.nodeHeight}px`
      },
      onClick: () => {
        // 手动触发 select 事件
        handleSelect([option.key as string], option)
      },
      'data-key': option.key
    }
    
    if (hasContextMenu.value) {
      nodePropsObj.onContextmenu = (e: MouseEvent) => {
        e.preventDefault()
        e.stopPropagation()
        currentContextNode.value = option
        contextMenuOptionsRef.value = getMenuItems(option)
        if (contextMenuOptionsRef.value.length) {
          contextMenuX.value = e.clientX
          contextMenuY.value = e.clientY
          showContextMenu.value = true
        }
      }
    }
    return nodePropsObj
  }
})

function handleContentContextmenu(e: MouseEvent) {
  const target = e.target as HTMLElement
  if (target.closest('.n-tree-node')) return
  e.preventDefault()
  e.stopPropagation()
  if (!hasContextMenu.value) {
    emit('blank-contextmenu', e)
    return
  }
  currentContextNode.value = null
  contextMenuOptionsRef.value = getMenuItems(null)
  if (contextMenuOptionsRef.value.length) {
    contextMenuX.value = e.clientX
    contextMenuY.value = e.clientY
    showContextMenu.value = true
  } else {
    emit('blank-contextmenu', e)
  }
}

// 前缀/后缀：支持全局 renderPrefix/renderSuffix，或节点数据 icon/iconColor、suffixIcon/suffixIconColor（参考 ）
const renderPrefixComputed = computed(() => {
  if (props.renderPrefix) {
    return typeof props.renderPrefix === 'string'
      ? () => renderIconVNode(props.renderPrefix as string, undefined, { size: 16 })()!
      : props.renderPrefix
  }
  if (!props.showIcon) return undefined
  return ({ option }: { option: TreeOption }) => {
    const icon = (option as any).icon
    if (!icon) return null
    const vnode = renderIconVNode(icon, undefined, {
      size: 16,
      color: (option as any).iconColor
    })()
    return vnode ?? null
  }
})
const renderSuffixComputed = computed(() => {
  if (props.renderSuffix) {
    return typeof props.renderSuffix === 'string'
      ? () => renderIconVNode(props.renderSuffix as string, undefined, { size: 16 })()!
      : props.renderSuffix
  }
  if (!props.showIcon) return undefined
  return ({ option }: { option: TreeOption }) => {
    const icon = (option as any).suffixIcon
    if (!icon) return null
    const vnode = renderIconVNode(icon, undefined, {
      size: 16,
      color: (option as any).suffixIconColor
    })()
    return vnode ?? null
  }
})

// 展开/折叠图标：根据 expanded 切换 iconOpen/iconClose（参考  + NTree RenderSwitcherIcon）
const renderSwitcherIcon = (info: { expanded: boolean; option: TreeOption }) =>
  renderIconVNode(info.expanded ? props.iconOpen : props.iconClose, undefined, { size: 'tiny' })()!

// Label 渲染包装器：支持省略、节点 prefixIcon/suffixIcon 一行展示（参考 ）
const renderLabelWrapperComputed = computed(() => {
  const needCustomLabel = props.renderLabel || props.ellipsis || props.showIcon
  if (!needCustomLabel) return undefined

  return ({ option }: { option: TreeOption }) => {
    if (props.renderLabel) {
      const userLabel = props.renderLabel({ option })
      if (props.ellipsis && userLabel) {
        return h(GEllipsis, { lineClamp: props.ellipsisLineClamp, tooltip: props.ellipsisTooltip }, { default: () => userLabel })
      }
      return userLabel
    }
    const raw = option as any
    const hasPrefix = raw?.prefixIcon && props.showIcon
    const hasSuffix = raw?.suffixIcon && props.showIcon
    const labelText = (option.label ?? '') as string
    if (hasPrefix || hasSuffix) {
      const parts: any[] = []
      if (hasPrefix) {
        const v = renderIconVNode(raw.prefixIcon, undefined, { size: 16, color: raw.prefixIconColor })()
        if (v) parts.push(h('span', { class: 'g-tree-node-label-icon' }, v))
      }
      parts.push(
        h(GEllipsis, {
          class: 'g-tree-node-label-text',
          text: labelText,
          lineClamp: props.ellipsisLineClamp,
          tooltip: props.ellipsisTooltip
        })
      )
      if (hasSuffix) {
        const v = renderIconVNode(raw.suffixIcon, undefined, { size: 16, color: raw.suffixIconColor })()
        if (v) parts.push(h('span', { class: 'g-tree-node-label-icon' }, v))
      }
      return h('div', { class: 'g-tree-node-label' }, parts)
    }
    if (props.ellipsis) {
      return h(GEllipsis, {
        text: labelText,
        lineClamp: props.ellipsisLineClamp,
        tooltip: props.ellipsisTooltip
      })
    }
    return undefined
  }
})

// 处理复选框变化（参考 ：带 meta 发出 check-change）
function handleUpdateCheckedKeys(
  keys: (string | number)[],
  options: Array<TreeOption | null>,
  meta?: { node: TreeOption | null; action: 'check' | 'uncheck' }
) {
  emit('update:checkedKeys', keys as string[])
  if (meta?.node && meta.action) {
    const keySet = new Set(keys.map(String))
    const checkedNodes = (options || []).filter((o): o is TreeOption => o != null && o.key != null && keySet.has(String(o.key)))
    emit('check-change', meta.node, meta.action === 'check', checkedNodes)
  }
}

// 处理展开变化（参考 ：带 meta 发出 node-expand）
function handleUpdateExpandedKeys(
  keys: (string | number)[],
  _options: Array<TreeOption | null>,
  meta?: { node: TreeOption | null; action: 'expand' | 'collapse' | 'filter' }
) {
  emit('update:expandedKeys', keys as string[])
  if (meta?.node && meta.action !== 'filter') {
    emit('node-expand', meta.node, meta.action === 'expand')
  }
}

// 当前选中 key（用于 getCurrentKey / setCurrentKey）
const selectedKeyRef = ref<string | number | null>(null)
function handleSelect(keys: string[], option: TreeOption) {
  selectedKeyRef.value = option?.key != null ? option.key : (keys[0] ?? null)
  emit('select', keys, option)
}

const handleDblclick = (option: TreeOption) => {
  emit('dblclick', option)
}

// 拖拽（参考 ）
const currentDragNode = ref<TreeOption | null>(null)
function handleDragStart(info: { node?: TreeOption }) {
  if (info?.node) currentDragNode.value = info.node
}
function handleDragEnd() {
  currentDragNode.value = null
}
const mergedAllowDrop = computed(() => {
  if (!props.draggable) return undefined
  return (info: { dropPosition: 'before' | 'inside' | 'after'; node: TreeOption; phase: 'drag' | 'drop' }) => {
    const dropNode = info.node
    const position = info.dropPosition === 'inside' ? 'inner' : info.dropPosition
    if (props.allowDrop && currentDragNode.value) {
      return props.allowDrop(currentDragNode.value, dropNode, position)
    }
    if (info.dropPosition !== 'inside') return true
    return !(dropNode as any)?.isLeaf
  }
})
function handleDrop(info: { dragNode?: TreeOption; node?: TreeOption; event?: DragEvent; dropPosition?: 'before' | 'inside' | 'after' }) {
  const dragNode = info.dragNode
  const dropNode = info.node
  const event = info.event
  const pos = info.dropPosition === 'inside' ? 'inner' : info.dropPosition ?? 'inner'
  if (!dragNode || !dropNode) return
  emit('node-drop', dragNode, dropNode, pos as 'before' | 'after' | 'inner', event ?? new DragEvent('drop'))
}

// 懒加载（参考 ，适配 NTree onLoad）
async function handleLoad(option: TreeOption): Promise<void> {
  if (!props.loadMethod) return
  const node = option
  await new Promise<void>((resolve) => {
    const result = props.loadMethod!(node, (children: TreeOption[]) => {
      const arr = children ?? []
      if (!Array.isArray((option as any)[props.childrenField])) {
        (option as any)[props.childrenField] = arr
      }
      if (arr.length === 0) (option as any).isLeaf = true
      resolve()
    })
    if (result instanceof Promise) result.then(resolve).catch(() => resolve())
    else resolve()
  })
}

// 暴露方法（参考 ）
function getNode(key: string | number): TreeOption | null {
  const find = (nodes: TreeOption[]): TreeOption | null => {
    for (const n of nodes) {
      if (n.key != null && String(n.key) === String(key)) return n
      if (n.children?.length) {
        const found = find(n.children)
        if (found) return found
      }
    }
    return null
  }
  return find(treeData.value)
}
function expandAll() {
  const collect = (nodes: TreeOption[]): string[] => {
    const keys: string[] = []
    for (const n of nodes) {
      if (n.key != null) keys.push(String(n.key))
      if (n.children?.length) keys.push(...collect(n.children))
    }
    return keys
  }
  ;(treeRef.value as any)?.setExpandedKeys?.(collect(treeData.value))
}
function collapseAll() {
  ;(treeRef.value as any)?.setExpandedKeys?.([])
}
function getCheckedNodes(): TreeOption[] {
  const inst = treeRef.value as any
  if (!inst?.getCheckedData) return []
  const data = inst.getCheckedData()
  const options = (data?.options ?? []) as (TreeOption | null)[]
  return options.filter((o): o is TreeOption => o != null)
}
function getCheckedKeys(): (string | number)[] {
  const inst = treeRef.value as any
  if (!inst?.getCheckedData) return []
  return (inst.getCheckedData()?.keys ?? []) as (string | number)[]
}
function setCheckedKeys(keys: (string | number)[]) {
  ;(treeRef.value as any)?.setCheckedKeys?.(keys)
}
function setChecked(key: string | number, checked: boolean) {
  const keys = new Set(getCheckedKeys().map((k) => String(k)))
  if (checked) keys.add(String(key))
  else keys.delete(String(key))
  ;(treeRef.value as any)?.setCheckedKeys?.(Array.from(keys))
}
function getCurrentNode(): TreeOption | null {
  const key = selectedKeyRef.value
  return key != null ? getNode(key) : null
}
function getCurrentKey(): string | number | null {
  return selectedKeyRef.value
}
function setCurrentKey(key: string | number) {
  ;(treeRef.value as any)?.setSelectedKeys?.([key])
  selectedKeyRef.value = key
}
function expandNode(key: string | number) {
  const inst = treeRef.value as any
  if (!inst?.getExpandedKeys) return
  const current = (inst.getExpandedKeys() ?? []) as (string | number)[]
  const set = new Set(current.map(String))
  set.add(String(key))
  inst.setExpandedKeys?.(Array.from(set))
}
function collapseNode(key: string | number) {
  const inst = treeRef.value as any
  if (!inst?.getExpandedKeys) return
  const current = (inst.getExpandedKeys() ?? []) as (string | number)[]
  const set = new Set(current.map(String))
  set.delete(String(key))
  inst.setExpandedKeys?.(Array.from(set))
}
function scrollTo(key: string | number) {
  ;(treeRef.value as any)?.scrollTo?.({ key })
}
function refresh() {}

watch(
  () => props.defaultExpandedKeys,
  (keys) => {
    if (!treeRef.value || !keys?.length) return
    nextTick(() => (treeRef.value as any)?.setExpandedKeys?.(keys))
  },
  { immediate: true }
)
watch(
  () => props.defaultCheckedKeys,
  (keys) => {
    if (!treeRef.value || !keys?.length) return
    nextTick(() => setCheckedKeys(keys as (string | number)[]))
  },
  { immediate: true }
)

defineExpose<GTreeInstance>({
  getTreeRef: () => treeRef.value,
  getNode,
  getCheckedNodes,
  getCheckedKeys,
  setCheckedKeys,
  setChecked,
  getCurrentNode,
  getCurrentKey,
  setCurrentKey,
  expandNode,
  collapseNode,
  expandAll,
  collapseAll,
  scrollTo,
  refresh
})
</script>

<style scoped lang="scss">
/* 参考  -tree，使用 --g-* 变量 */
.g-tree-wrapper {
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 100%;
  background: var(--g-bg-primary);
  border-radius: var(--g-radius-lg);
  overflow: hidden;
}

.g-tree-filter {
  flex-shrink: 0;
  padding: var(--g-padding-sm) var(--g-padding-md);
  border-bottom: 1px solid var(--g-border-primary);
}

.g-tree-content {
  flex: 1;
  overflow: auto;
  position: relative;
}

:deep(.n-tree .n-tree-node-content) {
  padding: 0;
  display: flex;
  align-items: center;
  overflow: hidden;
  width: 100%;
  max-width: 100%;

  .n-tree-node-content__prefix {
    margin-right: var(--g-space-xs);
    display: inline-flex;
    align-items: center;
  }

  .n-tree-node-content__text {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    max-width: 100%;
    display: flex;
    align-items: center;
  }

  > * {
    max-width: 100%;
    overflow: hidden;
  }
}

:deep(.g-tree-node-label) {
  display: flex;
  align-items: center;
  gap: var(--g-space-sm);
  min-width: 0;
  flex: 1;
  font-size: var(--g-font-size-sm);

  .g-tree-node-label-icon {
    font-size: var(--g-font-size-sm);
    transition: color var(--g-transition-base) var(--g-transition-ease);
    flex-shrink: 0;
    display: inline-flex;
    align-items: center;
  }

  .g-tree-node-label-text {
    flex: 1;
    min-width: 0;
    font-size: var(--g-font-size-sm);
  }
}
</style>

