import type { CSSProperties } from 'vue'

/**
 * GPane 方向
 * - vertical: 上下分割
 * - horizontal: 左右分割
 */
export type GPaneDirection = 'vertical' | 'horizontal'

/**
 * GPane 组件 Props
 * 对 Naive UI `NSplit` 做了一层语义化封装
 */
export interface GPaneProps {
  /**
   * 分割方向
   * @default 'vertical'
   */
  direction?: GPaneDirection

  /**
   * 默认面板尺寸（0 ~ 1 的数字或像素/百分比字符串）
   * 对应 NSplit 的 default-size
   * @default 0.3
   */
  defaultSize?: number | string

  /**
   * 受控尺寸（配合 v-model:size 使用）
   * 对应 NSplit 的 size
   */
  size?: number | string

  /**
   * 最小尺寸
   * @default 0
   */
  min?: number | string

  /**
   * 最大尺寸
   * @default 1
   */
  max?: number | string

  /**
   * 分割条粗细（像素）
   * 对应 NSplit 的 resize-trigger-size
   * @default 4
   */
  resizeTriggerSize?: number

  /**
   * 是否禁用拖拽
   * @default false
   */
  disabled?: boolean

  /**
   * 是否禁用拖拽但保持分割条样式
   * 当设置为 true 时，分割条仍然可见，但无法拖拽调整大小
   * 与 disabled 的区别：disabled 可能会隐藏分割条，而 noResize 保持视觉样式
   * @default false
   */
  noResize?: boolean

  /**
   * 面板一（上/左）自定义 class
   */
  pane1Class?: string

  /**
   * 面板一（上/左）自定义样式
   */
  pane1Style?: CSSProperties | string

  /**
   * 面板二（下/右）自定义 class
   */
  pane2Class?: string

  /**
   * 面板二（下/右）自定义样式
   */
  pane2Style?: CSSProperties | string
}

/**
 * GPane 组件事件
 */
export interface GPaneEmits {
  /**
   * 尺寸变化（用于 v-model:size 或监听分割条拖动）
   */
  (event: 'update:size', size: number | string): void

  /**
   * 拖拽开始
   */
  (event: 'drag-start', e: Event): void

  /**
   * 拖拽结束
   */
  (event: 'drag-end', e: Event): void
}

/**
 * GPane 组件暴露的方法
 */
export interface GPaneExpose {
  /**
   * 设置面板二（下/右）的可见性
   * @param visible 是否可见
   */
  setPane2Visible: (visible: boolean) => void

  /**
   * 获取面板二（下/右）的可见性
   */
  getPane2Visible: () => boolean

  /**
   * 切换面板二（下/右）的可见性
   */
  togglePane2Visible: () => void

  /**
   * 设置面板尺寸
   * @param size 尺寸值（0 ~ 1 的数字或像素/百分比字符串）
   */
  setSize: (size: number | string) => void

  /**
   * 获取当前面板尺寸
   */
  getSize: () => number | string | undefined
}


