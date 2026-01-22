/**
 * GCodeMirror 代码编辑器组件类型定义
 */

import type { Extension } from '@codemirror/state'
import type { EditorView } from '@codemirror/view'

/**
 * 支持的语言类型
 */
export type CodeMirrorLanguage =
  | 'javascript'
  | 'typescript'
  | 'json'
  | 'html'
  | 'css'
  | 'xml'
  | 'sql'
  | 'yaml'
  | 'markdown'
  | 'python'
  | 'java'
  | 'go'
  | 'rust'
  | 'shell' // 注意：shell 使用 JavaScript 语言包作为替代
  | 'plaintext'

/**
 * 支持的主题类型
 */
export type CodeMirrorTheme = 'light' | 'dark' | 'auto'

/**
 * GCodeMirror 组件 Props
 */
export interface GCodeMirrorProps {
  /**
   * 编辑器内容（v-model）
   */
  modelValue?: string

  /**
   * 语言类型
   * @default 'plaintext'
   */
  language?: CodeMirrorLanguage

  /**
   * 主题类型
   * @default 'auto' (自动跟随系统主题)
   */
  theme?: CodeMirrorTheme

  /**
   * 是否只读
   * @default false
   */
  readonly?: boolean

  /**
   * 是否显示行号
   * @default true
   */
  lineNumbers?: boolean

  /**
   * 是否显示折叠按钮
   * @default true
   */
  foldGutter?: boolean

  /**
   * 是否自动换行
   * @default false
   */
  lineWrapping?: boolean

  /**
   * 是否高亮当前行
   * @default true
   */
  highlightActiveLine?: boolean

  /**
   * 是否显示匹配括号
   * @default true
   */
  bracketMatching?: boolean

  /**
   * 是否自动闭合括号
   * @default true
   */
  autoCloseBrackets?: boolean

  /**
   * 是否显示搜索面板
   * @default true
   */
  searchKeymap?: boolean

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
   * 自定义扩展配置
   */
  extensions?: Extension[]
}

/**
 * GCodeMirror 组件 Emits
 */
export interface GCodeMirrorEmits {
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
  (event: 'ready', view: EditorView): void
}

