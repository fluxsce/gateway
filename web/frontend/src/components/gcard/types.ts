/**
 * GCard 组件类型定义
 */

import type { CardProps } from 'naive-ui'

/**
 * GCard 组件 Props
 * 基于 Naive UI 的 NCard，添加自定义扩展
 */
export interface GCardProps {
  /**
   * 标题
   */
  title?: string

  /**
   * 是否显示标题
   * @default false
   */
  showTitle?: boolean

  /**
   * 是否显示边框
   * @default false
   */
  bordered?: boolean

  /**
   * 是否显示阴影
   * @default 'never' - 不显示阴影
   * @default 'hover' - hover 时显示阴影
   * @default 'always' - 总是显示阴影
   */
  hoverable?: boolean | 'hover' | 'always'

  /**
   * 卡片尺寸
   * @default 'medium'
   */
  size?: 'small' | 'medium' | 'large' | 'huge'

  /**
   * 内容区域样式
   */
  contentStyle?: string | Record<string, string | number>

  /**
   * 头部样式
   */
  headerStyle?: string | Record<string, string | number>

  /**
   * 是否分段
   * - header: 仅分段头部
   * - content: 仅分段内容
   * - footer: 仅分段底部
   * - true: 分段全部
   */
  segmented?: CardProps['segmented']

  /**
   * 是否嵌入模式（降低视觉层级）
   */
  embedded?: boolean

  /**
   * 自定义类名
   */
  class?: string

  /**
   * 自定义样式
   */
  style?: string | Record<string, string | number>
}

/**
 * GCard 组件 Emits
 */
export interface GCardEmits {
  // 预留事件
}

