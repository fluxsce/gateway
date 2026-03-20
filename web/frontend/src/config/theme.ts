/**
 * Naive UI 主题配置（与  结构一致）
 *
 * 使用 CSS 变量（var(--g-*)）适配亮/暗主题：
 * - 颜色、背景、边框等使用 var(--g-*) 时，会随 styles 中 _theme-light.scss / _theme-dark.scss
 *   的切换自动生效，无需在 lightThemeOverrides 与 darkThemeOverrides 中重复写两套颜色。
 * - 与主题无关的尺寸、圆角等可继续写死或使用 var(--g-radius-md) 等。
 *
 * 依赖：根节点或 [data-theme] 上已通过样式注入 --g-primary、--g-text-primary、
 * --g-bg-primary、--g-border-primary 等变量。
 */
import type { GlobalThemeOverrides } from 'naive-ui'

/** 下拉主题：使用 var(--g-*) 与 _variables.scss / 主题文件一致，随亮/暗主题自动适配 */
const dropdownThemeWithVars = {
  borderRadius: 'var(--g-radius-lg)',
  padding: '0',
  optionHeightSmall: 'var(--g-height-small)',
  optionHeightMedium: 'var(--g-height-medium)',
  optionHeightLarge: 'var(--g-height-large)',
  optionHeightHuge: 'var(--g-height-huge)',
  optionIconSizeSmall: 'var(--g-font-size-sm)',
  optionIconSizeMedium: 'var(--g-font-size-base)',
  fontSizeSmall: 'var(--g-font-size-sm)',
  fontSizeMedium: 'var(--g-font-size-base)',
  color: 'var(--g-bg-primary)',
  optionTextColor: 'var(--g-text-secondary)',
  optionTextColorHover: 'var(--g-primary)',
  optionTextColorActive: 'var(--g-primary)',
  optionColorHover: 'var(--g-bg-tertiary)',
  optionColorActive: 'var(--g-primary-light)',
  dividerColor: 'var(--g-border-primary)',
  optionOpacityDisabled: 'var(--g-opacity-disabled)',
  peers: {
    Popover: {
      borderRadius: 'var(--g-radius-lg)',
      boxShadow: 'var(--g-shadow-md)',
    },
  },
}

/**
 * 亮色主题覆盖配置
 * 与主题强相关的组件（如 Dropdown）已用 var(--g-*) 适配，此处仅写亮色下可写死的 common 等
 */
export const lightThemeOverrides: GlobalThemeOverrides = {
  common: {
    primaryColor: '#8b5cf6',
    primaryColorHover: '#7c3aed',
    primaryColorPressed: '#6d28d9',
    primaryColorSuppl: '#a78bfa',
    successColor: '#10b981',
    warningColor: '#f59e0b',
    errorColor: '#ef4444',
    infoColor: '#3b82f6',
    borderRadius: '6px',
    borderRadiusSmall: '4px',
    fontSizeMini: '12px',
    fontSizeTiny: '12px',
    fontSizeSmall: '13px',
    fontSizeMedium: '14px',
    fontSizeLarge: '15px',
    fontSizeHuge: '16px',
  },
  Button: {
    borderRadiusMedium: '6px',
    borderRadiusSmall: '4px',
    heightMini: '24px',
    heightTiny: '28px',
    heightSmall: '30px',
    heightMedium: '34px',
    heightLarge: '40px',
  },
  Input: {
    borderRadius: '6px',
    heightMini: '24px',
    heightTiny: '28px',
    heightSmall: '30px',
    heightMedium: '34px',
  },
  Card: {
    borderRadius: '8px',
    paddingMedium: '16px',
    paddingLarge: '20px',
  },
  Form: {
    feedbackHeightSmall: '20px',
  },
  Menu: {
    itemHeight: '36px',
    borderRadius: '6px',
  },
  Drawer: {
    headerPadding: '16px 20px',
    bodyPadding: '20px',
    footerPadding: '12px 20px',
  },
  Dialog: {
    borderRadius: '8px',
    padding: '20px',
  },
  Table: {
    borderRadius: '6px',
    thPaddingSmall: '8px 12px',
    tdPaddingSmall: '8px 12px',
    thPaddingMedium: '10px 16px',
    tdPaddingMedium: '10px 16px',
  },
  Dropdown: dropdownThemeWithVars,
  Select: {
    menuBorderRadius: '6px',
  },
  Tabs: {
    tabBorderRadius: '6px',
  },
}

/**
 * 暗色主题覆盖配置
 * Dropdown 等已通过 var(--g-*) 随 [data-theme='dark'] 自动适配，此处仅写暗色 common 等
 */
export const darkThemeOverrides: GlobalThemeOverrides = {
  common: {
    primaryColor: '#818cf8',
    primaryColorHover: '#a5b4fc',
    primaryColorPressed: '#6366f1',
    primaryColorSuppl: '#c7d2fe',
    successColor: '#34d399',
    warningColor: '#fbbf24',
    errorColor: '#f87171',
    infoColor: '#60a5fa',
    borderRadius: '6px',
    borderRadiusSmall: '4px',
    fontSizeMini: '12px',
    fontSizeTiny: '12px',
    fontSizeSmall: '13px',
    fontSizeMedium: '14px',
    fontSizeLarge: '15px',
    fontSizeHuge: '16px',
  },
  Button: {
    borderRadiusMedium: '6px',
    borderRadiusSmall: '4px',
    heightMini: '24px',
    heightTiny: '28px',
    heightSmall: '30px',
    heightMedium: '34px',
    heightLarge: '40px',
  },
  Input: {
    borderRadius: '6px',
    heightMini: '24px',
    heightTiny: '28px',
    heightSmall: '30px',
    heightMedium: '34px',
  },
  Card: {
    borderRadius: '8px',
    paddingMedium: '16px',
    paddingLarge: '20px',
  },
  Form: {
    feedbackHeightSmall: '20px',
  },
  Menu: {
    itemHeight: '36px',
    borderRadius: '6px',
  },
  Drawer: {
    headerPadding: '16px 20px',
    bodyPadding: '20px',
    footerPadding: '12px 20px',
  },
  Dialog: {
    borderRadius: '8px',
    padding: '20px',
  },
  Table: {
    borderRadius: '6px',
    thPaddingSmall: '8px 12px',
    tdPaddingSmall: '8px 12px',
    thPaddingMedium: '10px 16px',
    tdPaddingMedium: '10px 16px',
  },
  Dropdown: dropdownThemeWithVars,
  Select: {
    menuBorderRadius: '6px',
  },
  Tabs: {
    tabBorderRadius: '6px',
  },
}
