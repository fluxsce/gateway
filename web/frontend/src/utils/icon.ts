/**
 * 图标工具类（机制与  一致）
 * 支持按名称动态加载、缓存、以及提前注册（如 CommonIcons），getIconSync 仅读缓存
 * 图标库：@vicons/ionicons5, @vicons/antd
 */

import type { Component } from 'vue'

/**
 * 图标库枚举
 */
export enum IconLibrary {
  /** Ionicons5 图标库 */
  IONICONS5 = 'ionicons5',
  /** Ant Design 图标库 */
  ANTD = 'antd'
}

/** 缓存 key：${library}:${iconName} */
function cacheKey(name: string, library: IconLibrary): string {
  return `${library}:${name}`
}

/**
 * 图标组件缓存（与 getIcon / getIconSync 共用，registerIcon 写入后同步可用）
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

  const key = cacheKey(iconName, library)
  if (iconCache.has(key)) {
    return iconCache.get(key)!
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

    iconCache.set(key, iconComponent)
    return iconComponent
  } catch (error) {
    console.error(`[Icon Utils] Failed to load icon "${iconName}" from ${library}:`, error)
    return null
  }
}

/**
 * 根据图标名称同步获取已缓存的图标组件（未加载则返回 null，用于组件内 computed 先读缓存再触发异步加载）
 *
 * @param iconName - 图标名称
 * @param library - 图标库，默认为 IONICONS5
 * @returns 已缓存的组件或 null
 */
export function getIconSync(
  iconName: string,
  library: IconLibrary = IconLibrary.IONICONS5
): Component | null {
  if (!iconName) return null
  return iconCache.get(cacheKey(iconName, library)) ?? null
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
 * 注册单个图标组件（写入缓存，getIconSync 立即可用）
 *
 * @param name - 图标名称（如 'AddOutline'）
 * @param component - 图标组件
 * @param library - 图标库，默认 IONICONS5
 */
export function registerIcon(
  name: string,
  component: Component,
  library: IconLibrary = IconLibrary.IONICONS5
): void {
  if (!name) return
  iconCache.set(cacheKey(name, library), component)
}

/**
 * 批量注册图标组件（如提前注册 CommonIcons 对应组件）
 *
 * @param icons - 名称到组件的映射
 * @param library - 图标库，默认 IONICONS5
 */
export function registerIcons(
  icons: Record<string, Component>,
  library: IconLibrary = IconLibrary.IONICONS5
): void {
  Object.entries(icons).forEach(([name, component]) => {
    registerIcon(name, component, library)
  })
}

// ==================== 预注册 CommonIcons（getIconSync 立即可用） ====================
import {
    AddOutline,
    ArrowBackOutline,
    ArrowDownOutline,
    ArrowForwardOutline,
    ArrowUpOutline,
    CheckmarkCircleOutline,
    CheckmarkOutline,
    CloseCircleOutline,
    CloseOutline,
    CloudUploadOutline,
    CreateOutline,
    DocumentOutline,
    DownloadOutline,
    EllipsisHorizontalOutline,
    FolderOutline,
    FunnelOutline,
    HomeOutline,
    InformationCircleOutline,
    MenuOutline,
    PeopleOutline,
    PersonOutline,
    RefreshOutline,
    SaveOutline,
    SearchOutline,
    SettingsOutline,
    SwapVerticalOutline,
    TrashOutline,
    WarningOutline
} from '@vicons/ionicons5'

// CommonIcons 中使用的名称与下方 key 一致，getIconSync(CommonIcons.xxx) 立即可用
registerIcons(
  {
    AddOutline,
    CreateOutline,
    TrashOutline,
    SaveOutline,
    CloseOutline,
    RefreshOutline,
    SearchOutline,
    FunnelOutline,
    SwapVerticalOutline,
    HomeOutline,
    ArrowBackOutline,
    ArrowForwardOutline,
    ArrowUpOutline,
    ArrowDownOutline,
    DownloadOutline,
    CloudUploadOutline,
    DocumentOutline,
    FolderOutline,
    PersonOutline,
    PeopleOutline,
    SettingsOutline,
    CheckmarkCircleOutline,
    CloseCircleOutline,
    WarningOutline,
    InformationCircleOutline,
    EllipsisHorizontalOutline,
    MenuOutline,
    CheckmarkOutline
  },
  IconLibrary.IONICONS5
)

