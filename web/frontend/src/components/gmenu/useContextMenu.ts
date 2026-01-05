/**
 * 公共右键菜单 Hook
 * 用于 Tree、Grid 等组件的右键菜单功能
 */

import { store } from '@/stores'
import { copyToClipboard, renderIconVNode } from '@/utils'
import { CopyOutline } from '@vicons/ionicons5'
import type { DropdownOption } from 'naive-ui'
import { NIcon } from 'naive-ui'
import { computed, h } from 'vue'
import type { ContextMenuConfig, ContextMenuItemConfig } from './types'

/**
 * 使用右键菜单配置
 * @param config 菜单配置
 * @param moduleId 模块ID（用于权限控制）
 * @param onMenuClick 菜单点击回调（可选，会与 config.onMenuClick 合并）
 */
export function useContextMenu(
  config: ContextMenuConfig | undefined,
  moduleId?: string,
  onMenuClick?: (menu: { code: string; node?: any; row?: any; column?: any }) => void
) {
  /**
   * 检查菜单项权限
   */
  const checkMenuPermission = (menuCode: string): boolean => {
    if (!moduleId) {
      return true // 没有 moduleId 时默认允许
    }
    const permissionCode = `${moduleId}:${menuCode}`
    return store.user.hasButton(permissionCode)
  }

  /**
   * 转换图标为 VNode
   */
  const convertIcon = (icon?: string | ((params?: {}) => any), defaultIcon?: string): any => {
    if (!icon && !defaultIcon) {
      return undefined
    }
    
    const iconToRender = icon || defaultIcon
    
    if (typeof iconToRender === 'string') {
      // 尝试使用 renderIconVNode 渲染图标
      try {
        return () => renderIconVNode(iconToRender, NIcon)()
      } catch {
        // 如果失败，返回 undefined
        return undefined
      }
    } else if (typeof iconToRender === 'function') {
      return iconToRender
    }
    
    return undefined
  }

  /**
   * 转换为 DropdownOption
   */
  const convertToDropdownOption = (menu: ContextMenuItemConfig): DropdownOption => {
    const hasPermission = checkMenuPermission(menu.code)
    
      const option: DropdownOption = {
        key: menu.code,
        label: menu.name,
        disabled: menu.disabled || !hasPermission,
        show: menu.visible !== false
      }
      
      // 添加前缀图标
      if (menu.prefixIcon) {
        const iconFn = convertIcon(menu.prefixIcon)
        if (iconFn) {
          option.icon = iconFn
        }
      }
      
      // 添加后缀图标（naive-ui DropdownOption 不支持 extra，可以通过 label 渲染实现）
      // 这里先不支持 suffixIcon，因为 naive-ui 的 DropdownOption 不直接支持
      
      // 添加子菜单
      if (menu.children && menu.children.length > 0) {
        option.children = menu.children.map(child => {
          const childHasPermission = checkMenuPermission(child.code)
          const childOption: DropdownOption = {
            key: child.code,
            label: child.name,
            disabled: child.disabled || !childHasPermission,
            show: child.visible !== false
          }
          
          if (child.prefixIcon) {
            const iconFn = convertIcon(child.prefixIcon)
            if (iconFn) {
              childOption.icon = iconFn
            }
          }
          
          return childOption
        })
      }
      
      return option
  }

  /**
   * 计算下拉菜单选项
   */
  const dropdownOptions = computed<DropdownOption[]>(() => {
    if (!config || config.enabled === false) {
      return []
    }

    const options: DropdownOption[] = []

    // 默认菜单项（复制节点/行数据）
    const showCopyNode = config.showCopyNode !== false
    const showCopyRow = config.showCopyRow !== false
    
    if (showCopyNode || showCopyRow) {
      options.push({
        key: 'copyNode',
        label: showCopyRow ? '复制行数据' : '复制节点数据',
        icon: () => h(NIcon, null, {
          default: () => h(CopyOutline)
        })
      })
    }

    // 自定义菜单项
    if (config.customMenus && config.customMenus.length > 0) {
      const customOptions = config.customMenus
        .filter(menu => menu.visible !== false)
        .map(convertToDropdownOption)
      options.push(...customOptions)
    }

    return options
  })

  /**
   * 处理菜单点击
   */
  const handleMenuClick = (
    key: string | number,
    node?: any,
    row?: any,
    column?: any
  ) => {
    const code = String(key)

    // 处理默认菜单项
    if (code === 'copyNode' || code === 'copyRow') {
      const nodeData = node || row
      if (nodeData) {
        const nodeDataStr = JSON.stringify(nodeData, null, 2)
        copyToClipboard(nodeDataStr)
      }
    } else if (code === 'copyCell') {
      const cellValue = row?.[column?.field]
      copyToClipboard(String(cellValue ?? ''))
    }

    // 触发回调
    const menuInfo = { code, node, row, column }
    
    if (onMenuClick) {
      onMenuClick(menuInfo)
    }
    
    if (config?.onMenuClick) {
      config.onMenuClick(menuInfo)
    }
  }

  return {
    dropdownOptions,
    handleMenuClick
  }
}

