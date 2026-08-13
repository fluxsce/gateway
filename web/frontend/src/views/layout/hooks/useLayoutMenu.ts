/**
 * 主布局侧栏菜单：从 `layoutRouteRegistry` 单一数据源生成 RsMenu items，
 * 点击仅通过 `globalStore.upsertLayoutTab` 维护页签；路由由 `MainLayoutContent` 监听 `layoutActiveTabId` 同步。
 *
 * @module views/layout/hooks/useLayoutMenu
 */
import { buildSidebarMenuFromRegistry, isLayoutMenuGroup } from '@/router/layoutRouteRegistry'
import { useGlobalStore } from '@/stores/global'
import type { RsMenuItem } from '@/ui'
import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'

/** 侧栏树节点类型，与 {@link buildSidebarMenuFromRegistry} 返回数组元素一致。 */
type LayoutMenuNode = ReturnType<typeof buildSidebarMenuFromRegistry>[number]

/** 叶子项扩展：用于选中时写入页签的 path / icon */
type LayoutMenuLeafMeta = {
  key: string
  label: string
  path: string
  icon: string
}

/**
 * 将注册表节点映射为 RsMenuItem。
 */
function mapNodeToMenuItem(node: LayoutMenuNode): RsMenuItem {
  if (isLayoutMenuGroup(node)) {
    return {
      key: node.key,
      label: node.label,
      icon: node.icon,
      children: node.children.map((child) => ({
        key: child.key,
        label: child.label,
        icon: child.icon,
      })),
    }
  }
  return {
    key: node.key,
    label: node.label,
    icon: node.icon,
  }
}

function collectLeafMeta(nodes: LayoutMenuNode[]): LayoutMenuLeafMeta[] {
  const leaves: LayoutMenuLeafMeta[] = []
  for (const node of nodes) {
    if (isLayoutMenuGroup(node)) {
      for (const child of node.children) {
        if (child.path) {
          leaves.push({
            key: child.key,
            label: child.label,
            path: child.path,
            icon: child.icon,
          })
        }
      }
      continue
    }
    if (node.path) {
      leaves.push({
        key: node.key,
        label: node.label,
        path: node.path,
        icon: node.icon,
      })
    }
  }
  return leaves
}

/**
 * 主布局侧栏菜单：选项列表 + 菜单选中回调。
 */
export function useLayoutMenu() {
  const globalStore = useGlobalStore()
  const route = useRoute()

  const registryNodes = computed(() => buildSidebarMenuFromRegistry())
  const menuItems = computed<RsMenuItem[]>(() =>
    registryNodes.value.map((node) => mapNodeToMenuItem(node)),
  )
  const leafMetaByKey = computed(() => {
    const map = new Map<string, LayoutMenuLeafMeta>()
    for (const leaf of collectLeafMeta(registryNodes.value)) {
      map.set(leaf.key, leaf)
    }
    return map
  })

  const activeMenuKey = ref('')
  const openMenuKeys = ref<string[]>([])

  watch(
    () => route.path,
    (path) => {
      const leaves = collectLeafMeta(registryNodes.value)
      const matched =
        leaves.find((leaf) => leaf.path === path) ||
        leaves.find((leaf) => path === leaf.path || path.startsWith(`${leaf.path}/`))
      activeMenuKey.value = matched?.key ?? ''
    },
    { immediate: true },
  )

  const handleMenuSelect = (key: string) => {
    const leaf = leafMetaByKey.value.get(key)
    if (!leaf) return
    globalStore.upsertLayoutTab(leaf.path, leaf.label, leaf.icon)
  }

  return {
    menuItems,
    activeMenuKey,
    openMenuKeys,
    handleMenuSelect,
  }
}
