/**
 * 图标工具类
 * 根据图标名称动态获取图标组件
 * 支持的图标库：@vicons/ionicons5, @vicons/antd
 */

import { NIcon } from 'naive-ui'
import type { Component } from 'vue'
import { h, markRaw, ref, shallowRef } from 'vue'

/**
 * 图标库枚举
 */
export enum IconLibrary {
  /** Ionicons5 图标库 */
  IONICONS5 = 'ionicons5',
  /** Ant Design 图标库 */
  ANTD = 'antd'
}

/**
 * 图标缓存，避免重复加载
 */
const iconCache = new Map<string, Component>()

/**
 * 图标库模块映射
 */
const iconLibraryModules: Record<IconLibrary, () => Promise<any>> = {
  [IconLibrary.IONICONS5]: () => import('@vicons/ionicons5'),
  [IconLibrary.ANTD]: () => import('@vicons/antd')
}

/**
 * 根据图标名称获取图标组件
 * 
 * @param iconName - 图标名称（如 'AddOutline', 'UserOutlined'）
 * @param library - 图标库，默认为 IONICONS5
 * @returns 图标组件，如果未找到则返回 null
 * 
 * @example
 * ```typescript
 * // 获取 Ionicons5 图标
 * const AddIcon = await getIcon('AddOutline')
 * 
 * // 获取 Ant Design 图标
 * const UserIcon = await getIcon('UserOutlined', IconLibrary.ANTD)
 * ```
 */
export async function getIcon(
  iconName: string,
  library: IconLibrary = IconLibrary.IONICONS5
): Promise<Component | null> {
  if (!iconName) {
    console.warn('[Icon Utils] Icon name is required')
    return null
  }

  // 生成缓存 key
  const cacheKey = `${library}:${iconName}`

  // 检查缓存
  if (iconCache.has(cacheKey)) {
    return iconCache.get(cacheKey)!
  }

  try {
    // 动态导入图标库
    const iconModule = await iconLibraryModules[library]()
    
    // 获取图标组件
    const iconComponent = iconModule[iconName]
    
    if (!iconComponent) {
      console.warn(`[Icon Utils] Icon "${iconName}" not found in ${library} library`)
      return null
    }

    // 缓存图标组件
    iconCache.set(cacheKey, iconComponent)
    
    return iconComponent
  } catch (error) {
    console.error(`[Icon Utils] Failed to load icon "${iconName}" from ${library}:`, error)
    return null
  }
}

/**
 * 批量获取图标组件
 * 
 * @param iconNames - 图标名称数组
 * @param library - 图标库，默认为 IONICONS5
 * @returns 图标组件数组，未找到的图标为 null
 * 
 * @example
 * ```typescript
 * const icons = await getIcons(['AddOutline', 'RefreshOutline', 'TrashOutline'])
 * ```
 */
export async function getIcons(
  iconNames: string[],
  library: IconLibrary = IconLibrary.IONICONS5
): Promise<(Component | null)[]> {
  return Promise.all(iconNames.map(name => getIcon(name, library)))
}

/**
 * 常用图标名称常量
 * 提供类型安全的图标名称引用
 */
export const CommonIcons = {
  // 操作类
  ADD: 'AddOutline',
  EDIT: 'CreateOutline',
  DELETE: 'TrashOutline',
  SAVE: 'SaveOutline',
  CANCEL: 'CloseOutline',
  REFRESH: 'RefreshOutline',
  SEARCH: 'SearchOutline',
  FILTER: 'FunnelOutline',
  SORT: 'SwapVerticalOutline',
  
  // 导航类
  HOME: 'HomeOutline',
  BACK: 'ArrowBackOutline',
  FORWARD: 'ArrowForwardOutline',
  UP: 'ArrowUpOutline',
  DOWN: 'ArrowDownOutline',
  
  // 文件类
  DOWNLOAD: 'DownloadOutline',
  UPLOAD: 'CloudUploadOutline',
  FILE: 'DocumentOutline',
  FOLDER: 'FolderOutline',
  
  // 用户类
  USER: 'PersonOutline',
  USERS: 'PeopleOutline',
  SETTINGS: 'SettingsOutline',
  
  // 状态类
  SUCCESS: 'CheckmarkCircleOutline',
  ERROR: 'CloseCircleOutline',
  WARNING: 'WarningOutline',
  INFO: 'InformationCircleOutline',
  
  // 其他
  MORE: 'EllipsisHorizontalOutline',
  MENU: 'MenuOutline',
  CLOSE: 'CloseOutline',
  CHECK: 'CheckmarkOutline'
} as const

export type CommonIconName = typeof CommonIcons[keyof typeof CommonIcons]

/**
 * 渲染图标
 * 支持传入 Component 或图标名称字符串
 * 
 * @param icon - 图标组件或图标名称
 * @param library - 图标库（仅当 icon 为字符串时有效）
 * @returns 响应式的图标组件引用
 * 
 * @example
 * ```vue
 * <script setup>
 * const iconComp = renderIcon('AddOutline')
 * // 或
 * const iconComp = renderIcon(AddOutlineComponent)
 * </script>
 * 
 * <template>
 *   <n-icon v-if="iconComp">
 *     <component :is="iconComp" />
 *   </n-icon>
 * </template>
 * ```
 */
export function renderIcon(
  icon: Component | string | undefined,
  library: IconLibrary = IconLibrary.IONICONS5
) {
  const iconComponent = ref<Component | null>(null)

  if (!icon) {
    return iconComponent
  }

  // 如果是字符串，异步获取组件
  if (typeof icon === 'string') {
    getIcon(icon, library).then(comp => {
      iconComponent.value = comp
    })
  } else {
    // 如果是组件，直接使用
    iconComponent.value = icon
  }

  return iconComponent
}

/**
 * 渲染图标为 VNode（用于渲染函数）
 * 支持同步和异步加载图标，自动处理响应式更新
 * 
 * @param icon - 图标组件或图标名称
 * @param iconWrapper - 包裹组件（如 NIcon），默认使用 NIcon
 * @param library - 图标库（仅当 icon 为字符串时有效）
 * @returns 渲染函数
 * 
 * @example
 * ```typescript
 * // 在 computed 或渲染函数中使用
 * menuOption.icon = renderIconVNode('AddOutline')
 * menuOption.icon = renderIconVNode('AddOutline', NIcon)
 * ```
 */
export function renderIconVNode(
  icon: Component | string | undefined,
  iconWrapper?: Component,
  library: IconLibrary = IconLibrary.IONICONS5
) {
  // 如果没有提供图标，返回 null
  if (!icon) {
    return () => null
  }

  // 如果传入的是组件，直接使用（同步）
  if (typeof icon !== 'string') {
    const IconComponent = markRaw(icon)
    const Wrapper = iconWrapper ? markRaw(iconWrapper) : null
    
    return () => {
      // 如果没有提供 iconWrapper，使用直接导入的 NIcon 包裹
      if (Wrapper) {
        return h(Wrapper, null, { default: () => h(IconComponent) })
      }
      // 如果没有提供 wrapper，使用默认的 NIcon
      return h(NIcon, null, { default: () => h(IconComponent) })
    }
  }

  // 对于字符串图标名称，需要异步加载
  // 使用 shallowRef 避免深度响应式包装组件，并使用 markRaw 标记组件
  const iconRef = shallowRef<Component | null>(null)
  // 如果没有提供 iconWrapper，使用直接导入的 NIcon
  const wrapperRef = shallowRef<Component | null>(
    iconWrapper ? markRaw(iconWrapper) : markRaw(NIcon)
  )

  // 异步加载图标
  getIcon(icon, library)
    .then((iconComponent) => {
      if (iconComponent) {
        // 使用 markRaw 标记组件，避免被响应式包装
        iconRef.value = markRaw(iconComponent)
      }
    })
    .catch(() => {
      // 加载失败，保持 null
      console.warn(`[Icon Utils] Failed to load icon: ${icon}`)
    })

  // 返回渲染函数
  return () => {
    // 如果图标还没有加载完成，返回 null
    if (!iconRef.value) {
      return null
    }

    // 如果 wrapper 已经加载，使用 wrapper 包裹
    if (wrapperRef.value) {
      return h(wrapperRef.value, null, { default: () => h(iconRef.value!) })
    }
    
    // 如果 wrapper 还没加载完成，直接返回图标（这种情况不应该发生，但保险起见）
    return h(iconRef.value)
  }
}

