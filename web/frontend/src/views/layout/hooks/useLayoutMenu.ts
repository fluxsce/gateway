/**
 * 主布局侧栏菜单：从 `layoutRouteRegistry` 单一数据源生成 Naive UI `MenuOption`，
 * 点击仅通过 `globalStore.upsertLayoutTab` 维护页签；路由由 `MainLayoutContent` 监听 `layoutActiveTabId` 同步。
 *
 * @module views/layout/hooks/useLayoutMenu
 */
import { buildSidebarMenuFromRegistry, isLayoutMenuGroup } from '@/router/layoutRouteRegistry'
import { useGlobalStore } from '@/stores/global'
import { CommonIcons, IconLibrary, renderIconVNode } from '@/utils'
import type { MenuOption } from 'naive-ui'
import { computed } from 'vue'

/**
 * 侧栏树节点类型，与 {@link buildSidebarMenuFromRegistry} 返回数组元素一致。
 */
type LayoutMenuNode = ReturnType<typeof buildSidebarMenuFromRegistry>[number]

/**
 * 将注册表节点映射为 Naive `MenuOption`。
 *
 * - 分组：仅 `label` / `key` / `icon` / `children`，子项为叶子映射结果。
 * - 叶子：在官方字段之外挂载 **`routePath`**（与注册表 path 一致、用于 `upsertLayoutTab` 的 tabId/path）；
 *   展示用图标仅使用 Naive 约定的 **`icon`**（`renderIconVNode`），不向页签重复传图标名字段。
 *
 * @param node - `buildSidebarMenuFromRegistry()` 的节点
 * @param createIconRender - 将 Ionicons 类名字符串转为菜单用 VNode
 * @returns 可直接作为 `n-menu` 的 `options` 项（叶子含扩展字段 `routePath`）
 */
function mapNodeToMenuOption(
  node: LayoutMenuNode,
  createIconRender: (iconName: string) => ReturnType<typeof renderIconVNode>,
): MenuOption {
  if (isLayoutMenuGroup(node)) {
    return {
      label: node.label,
      key: node.key,
      icon: createIconRender(node.icon),
      children: node.children.map((child) => ({
        label: child.label,
        key: child.key,
        icon: createIconRender(child.icon),
        routePath: child.path,
      })),
    }
  }
  return {
    label: node.label,
    key: node.key,
    icon: createIconRender(node.icon),
    routePath: node.path,
  }
}

/**
 * 主布局侧栏菜单：选项列表 + 菜单选中回调。
 *
 * - **数据源**：`GATEWAY_LAYOUT_ROUTE_TREE` → {@link buildSidebarMenuFromRegistry}
 * - **选中**：仅 `upsertLayoutTab(path, title)`；重复/激活由 store 判断；`router.push` 由 `MainLayoutContent` 监听激活 tab 处理
 *
 * @returns
 * - `menuOptions`：侧栏 `n-menu` 的 `options`
 * - `handleMenuSelect`：绑定 `on-update:value`，入参为 Naive 传入的 key 与项（叶子项需带 `routePath`）
 */
export function useLayoutMenu() {
  const globalStore = useGlobalStore()

  const createIconRender = (iconName: string) => {
    return renderIconVNode(iconName || CommonIcons.MENU, IconLibrary.IONICONS5)
  }

  const menuOptions = computed<MenuOption[]>(() =>
    buildSidebarMenuFromRegistry().map((node) => mapNodeToMenuOption(node, createIconRender)),
  )

  /**
   * 侧栏选中叶子菜单时：按项上的 `routePath` 写入/激活页签（不在这里做路由跳转）。
   *
   * @param _key - `n-menu` 传入的 value（与项 `key` 一致；标题回退时可参与展示）
   * @param item - 选中项；叶子由 {@link mapNodeToMenuOption} 带有 `routePath`
   */
  const handleMenuSelect = (_key: string, item: MenuOption) => {
    const routePath = (item as MenuOption & { routePath?: string }).routePath
    if (!routePath) return
    const title = typeof item.label === 'string' ? item.label : String(item.key ?? _key)
    globalStore.upsertLayoutTab(routePath, title)
  }

  return {
    menuOptions,
    handleMenuSelect,
  }
}
