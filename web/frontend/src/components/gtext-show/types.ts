/**
 * @deprecated 请改用 RsCodeBlock（@/ui）。本类型仅作过渡，后续将删除。
 */

/**
 * 支持的文本格式类型
 * @deprecated
 */
export type TextFormat = 'json' | 'xml' | 'txt' | 'soap' | 'yaml' | 'sql' | 'javascript' | 'typescript' | 'css' | 'html' | 'auto'

/**
 * GTextShow 组件 Props
 * @deprecated 请直接使用 RsCodeBlock（@/ui）的 code / lang。
 */
export interface GTextShowProps {
  /**
   * 要显示的文本内容
   */
  content?: string

  /**
   * 文本格式类型
   * - auto: 自动检测格式
   * - json: JSON 格式
   * - xml: XML 格式
   * - txt: 纯文本
   * - soap: SOAP 格式
   * - yaml: YAML 格式
   * - sql: SQL 格式
   * - javascript: JavaScript 格式
   * - typescript: TypeScript 格式
   * - css: CSS 格式
   * - html: HTML 格式
   * @default 'auto'
   */
  format?: TextFormat

  /**
   * 是否显示行号
   * @default false
   */
  showLineNumbers?: boolean

  /**
   * 是否显示复制按钮
   * @default true
   */
  showCopyButton?: boolean

  /**
   * 是否自动格式化（仅对 JSON 有效）
   * @default true
   */
  autoFormat?: boolean

  /**
   * 最大高度（超出后显示滚动条）
   */
  maxHeight?: string | number

  /**
   * 最小高度
   */
  minHeight?: string | number

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
 * GTextShow 组件 Emits
 * @deprecated
 */
export interface GTextShowEmits {
  /**
   * 复制成功事件
   */
  (e: 'copy', value: string): void
}
