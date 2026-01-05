/**
 * Naive UI 主题配置
 * 定义浅色和深色主题的覆盖样式
 */
import type { GlobalThemeOverrides } from 'naive-ui'

/**
 * 浅色主题覆盖配置
 */
export const lightThemeOverrides: GlobalThemeOverrides = {
  common: {
    primaryColor: '#8b5cf6',
    primaryColorHover: '#7c3aed',
    primaryColorPressed: '#6d28d9',
    primaryColorSuppl: '#a78bfa',
  },
  Form: {
    feedbackHeightSmall: '20px',
  },
  Button: {
    borderRadiusMedium: '6px',
  },
  Menu: {
    itemHeight: '36px',
  },
  Dropdown: {
    optionHeightSmall: '32px',
    optionHeightMedium: '32px',
    optionHeightLarge: '32px',
    optionHeightHuge: '32px',
    optionSuffixWidthSmall: '32px',
    optionSuffixWidthMedium: '32px',
    optionSuffixWidthLarge: '32px',
    optionSuffixWidthHuge: '32px',
    optionPrefixWidthSmall: '32px',
    optionPrefixWidthMedium: '32px',
    optionPrefixWidthLarge: '32px',
    optionPrefixWidthHuge: '32px',
    optionIconSizeSmall: '16px',
    optionIconSizeMedium: '16px',
  },
  Input: {
    borderRadius: '6px',
  },
  Card: {
    borderRadius: '8px',
  },
  Drawer: {
    headerPadding: '7px 20px',
    footerPadding: '7px 20px',
    bodyPadding: '7px 20px'
  },
}

/**
 * 深色主题覆盖配置
 */
export const darkThemeOverrides: GlobalThemeOverrides = {
  common: {
    primaryColor: '#818cf8',
    primaryColorHover: '#a5b4fc',
    primaryColorPressed: '#c7d2fe',
    primaryColorSuppl: '#6366f1',
  },
  Form: {
    feedbackHeightSmall: '20px',
  },
  Menu: {
    itemHeight: '36px',
  },
  Dropdown: {
    optionHeightSmall: '32px',
    optionHeightMedium: '32px',
    optionHeightLarge: '32px',
    optionHeightHuge: '32px',
    optionSuffixWidthSmall: '32px',
    optionSuffixWidthMedium: '32px',
    optionSuffixWidthLarge: '32px',
    optionSuffixWidthHuge: '32px',
    optionPrefixWidthSmall: '32px',
    optionPrefixWidthMedium: '32px',
    optionPrefixWidthLarge: '32px',
    optionPrefixWidthHuge: '32px',
    optionIconSizeSmall: '16px',
    optionIconSizeMedium: '16px',
  },
  Button: {
    borderRadiusMedium: '6px',
  },
  Input: {
    borderRadius: '6px',
  },
  Card: {
    borderRadius: '8px',
  },
  Drawer: {
    headerPadding: '7px 20px',
    footerPadding: '7px 20px',
    bodyPadding: '7px 20px'
  },
}

