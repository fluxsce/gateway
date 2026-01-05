/**
 * GDate 组件类型定义
 */

/**
 * GDate 组件的 Props
 */
export interface GDateProps {
  /**
   * 日期值
   * 支持：
   * - ISO 字符串格式（如 "2023-10-18T15:24:23Z"）
   * - 时间戳（毫秒）
   * - null/undefined（空值）
   * - 日期范围数组（用于 daterange/datetimerange 类型）
   */
  value?: string | number | number[] | null

  /**
   * 输出格式
   * - 'iso': 输出 ISO 字符串格式（默认，适用于后端 API）
   * - 'timestamp': 输出时间戳格式（适用于需要时间戳的场景）
   * @default 'iso'
   */
  outputFormat?: 'iso' | 'timestamp'
}

/**
 * GDate 组件的事件
 */
export interface GDateEmits {
  /**
   * 值更新事件
   * @param value 更新后的值（根据 outputFormat 决定格式）
   * - outputFormat='iso': 返回 ISO 字符串或字符串数组
   * - outputFormat='timestamp': 返回时间戳或时间戳元组
   */
  (event: 'update:value', value: string | number | string[] | [number, number] | null): void
}

/**
 * 解析后的日期值类型
 * - 单个日期：number | null
 * - 日期范围：[number, number] | null
 */
export type ParsedDateValue = number | [number, number] | null

