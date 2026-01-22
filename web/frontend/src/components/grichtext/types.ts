/**
 * GRichText 富文本编辑器组件类型定义
 */

import type { Editor, Extension } from '@tiptap/core'

/**
 * 文本对齐方式
 */
export type TextAlign = 'left' | 'center' | 'right' | 'justify'

/**
 * GRichText 组件 Props
 */
export interface GRichTextProps {
  /**
   * 编辑器内容（v-model，HTML 格式）
   */
  modelValue?: string

  /**
   * 是否只读
   * @default false
   */
  readonly?: boolean

  /**
   * 占位符文本
   */
  placeholder?: string

  /**
   * 最小高度
   */
  minHeight?: string | number

  /**
   * 最大高度
   */
  maxHeight?: string | number

  /**
   * 高度
   */
  height?: string | number

  /**
   * 宽度
   */
  width?: string | number

  /**
   * 自定义类名
   */
  class?: string

  /**
   * 自定义样式
   */
  style?: string | Record<string, string | number>

  /**
   * 是否显示工具栏
   * @default true
   */
  showToolbar?: boolean

  /**
   * 工具栏配置
   * 可以自定义工具栏显示的功能
   */
  toolbarOptions?: {
    /**
     * 是否显示字体样式（粗体、斜体等）
     * @default true
     */
    fontStyle?: boolean

    /**
     * 是否显示字体族
     * @default true
     */
    fontFamily?: boolean

    /**
     * 是否显示文本颜色
     * @default true
     */
    textColor?: boolean

    /**
     * 是否显示对齐方式
     * @default true
     */
    textAlign?: boolean

    /**
     * 是否显示列表
     * @default true
     */
    list?: boolean

    /**
     * 是否显示链接
     * @default true
     */
    link?: boolean

    /**
     * 是否显示图片
     * @default true
     */
    image?: boolean

    /**
     * 是否显示标题
     * @default true
     */
    heading?: boolean

    /**
     * 是否显示表格
     * @default true
     */
    table?: boolean

    /**
     * 是否显示其他工具（水平线、清除格式等）
     * @default true
     */
    other?: boolean
  }

  /**
   * 自定义 Tiptap 扩展
   */
  extensions?: Extension[]
}

/**
 * GRichText 组件 Emits
 */
export interface GRichTextEmits {
  /**
   * v-model:modelValue 更新事件
   */
  (event: 'update:modelValue', value: string): void

  /**
   * 内容变化事件
   */
  (event: 'change', value: string): void

  /**
   * 焦点获得事件
   */
  (event: 'focus'): void

  /**
   * 焦点失去事件
   */
  (event: 'blur'): void

  /**
   * 编辑器准备就绪事件
   */
  (event: 'ready', editor: Editor): void
}

/**
 * GRichText 组件暴露的方法
 */
export interface GRichTextExpose {
  /**
   * 获取编辑器实例
   */
  getEditor: () => Editor | null

  /**
   * 获取编辑器内容（HTML）
   */
  getHTML: () => string

  /**
   * 设置编辑器内容（HTML）
   */
  setHTML: (html: string) => void

  /**
   * 获取编辑器内容（纯文本）
   */
  getText: () => string

  /**
   * 设置编辑器内容（纯文本）
   */
  setText: (text: string) => void

  /**
   * 获取编辑器内容（JSON）
   */
  getJSON: () => any

  /**
   * 设置编辑器内容（JSON）
   */
  setJSON: (json: any) => void

  /**
   * 聚焦编辑器
   */
  focus: () => void

  /**
   * 失焦编辑器
   */
  blur: () => void

  /**
   * 清空编辑器
   */
  clear: () => void
}
