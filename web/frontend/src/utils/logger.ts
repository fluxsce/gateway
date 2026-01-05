/**
 * 日志工具类
 * 提供统一的日志输出接口，可根据环境控制日志输出
 */

// 判断是否为生产环境
const isProd = import.meta.env.VITE_LOGGER_ENV === 'production'

// 日志级别枚举
export enum LogLevel {
  DEBUG = 0,
  INFO = 1,
  WARN = 2,
  ERROR = 3,
  NONE = 4,
}

// 默认日志级别：生产环境下仅显示错误，开发环境显示所有
const defaultLogLevel = isProd ? LogLevel.ERROR : LogLevel.DEBUG

/**
 * 日志记录器选项
 */
export interface LoggerOptions {
  level: LogLevel // 日志级别
  prefix?: string // 日志前缀
  includeTimestamp?: boolean // 是否包含时间戳
  includeStack?: boolean // 是否包含调用栈信息
  stackDepth?: number // 调用栈显示深度，1=只显示直接调用者，2=显示调用链
}

/**
 * 创建日志记录器实例
 */
export class Logger {
  private level: LogLevel
  private prefix: string
  private includeTimestamp: boolean
  private includeStack: boolean
  private stackDepth: number

  constructor(options: Partial<LoggerOptions> = {}) {
    this.level = options.level ?? defaultLogLevel
    this.prefix = options.prefix ?? ''
    this.includeTimestamp = options.includeTimestamp ?? true
    this.includeStack = options.includeStack ?? true
    this.stackDepth = options.stackDepth ?? 2
  }

  /**
   * 获取调用栈信息
   * @param depth 显示调用栈的深度，默认1层，可以设置为2显示调用链
   * @returns 调用者的文件路径和行号
   */
  private getCallerInfo(depth: number = 1): string {
    try {
      const stack = new Error().stack
      if (!stack) return ''

      const lines = stack.split('\n')
      const callers: string[] = []

      // 寻找不属于logger内部的调用
      for (let i = 1; i < lines.length; i++) {
        const line = lines[i].trim()
        if (!line) continue

        // 跳过logger内部方法
        if (
          line.includes('Logger.') ||
          line.includes('getCallerInfo') ||
          line.includes('format') ||
          line.includes('debug') ||
          line.includes('info') ||
          line.includes('warn') ||
          line.includes('error') ||
          line.includes('log.')
        ) {
          continue
        }

        // 跳过node_modules
        if (line.includes('node_modules')) {
          continue
        }

        // 提取文件路径和行号
        const match =
          line.match(/\(([^)]+):(\d+):(\d+)\)/) ||
          line.match(/at\s+([^:\s]+):(\d+):(\d+)/) ||
          line.match(/([^@\s]+):(\d+):(\d+)/)

        if (match) {
          const [, filePath, lineNumber] = match
          // 只显示文件名而不是完整路径
          const fileName = filePath.split(/[/\\]/).pop() || filePath
          // 移除可能的查询参数或哈希
          const cleanFileName = fileName.split('?')[0].split('#')[0]
          callers.push(`${cleanFileName}:${lineNumber}`)

          // 如果已经收集到足够的调用层级，停止
          if (callers.length >= depth) {
            break
          }
        }
      }

      // 如果depth > 1，显示调用链，用 <- 表示调用关系
      if (depth > 1 && callers.length > 1) {
        return callers.join(' <- ')
      }

      return callers[0] || ''
    } catch (error) {
      // 获取调用栈失败时不影响日志输出
    }
    return ''
  }

  /**
   * 格式化日志消息
   * @param args 日志参数
   * @returns 格式化后的参数数组
   */
  private format(args: any[]): any[] {
    const parts: string[] = []
    const styles: string[] = []

    // 添加时间戳
    if (this.includeTimestamp) {
      const now = new Date()
      const timestamp = `[${now.toLocaleTimeString()}]`
      parts.push(`%c${timestamp}`)
      styles.push('color: gray; font-weight: lighter;')
    }

    // 添加调用栈信息
    if (this.includeStack) {
      const caller = this.getCallerInfo(this.stackDepth)
      if (caller) {
        parts.push(`%c[${caller}]`)
        styles.push('color: #666; font-size: 11px;')
      }
    }

    // 添加前缀
    if (this.prefix) {
      const prefixStr = `[${this.prefix}]`
      parts.push(`%c${prefixStr}`)
      styles.push('color: blue; font-weight: bold;')
    }

    // 将格式化的部分组合成一个字符串，然后添加样式和其他参数
    if (parts.length > 0) {
      const formatString = parts.join(' ') + ' %c'
      styles.push('color: inherit;') // 重置颜色为默认
      return [formatString, ...styles, ...args]
    }

    // 如果没有格式化部分，直接返回原参数
    return args
  }

  /**
   * 调试级别日志
   * @param args 日志参数
   */
  debug(...args: any[]): void {
    if (this.level <= LogLevel.DEBUG && !isProd) {
      console.debug(...this.format(args))
    }
  }

  /**
   * 信息级别日志
   * @param args 日志参数
   */
  info(...args: any[]): void {
    if (this.level <= LogLevel.INFO && !isProd) {
      console.info(...this.format(args))
    }
  }

  /**
   * 警告级别日志
   * @param args 日志参数
   */
  warn(...args: any[]): void {
    if (this.level <= LogLevel.WARN && !isProd) {
      console.warn(...this.format(args))
    }
  }

  /**
   * 错误级别日志
   * @param args 日志参数
   */
  error(...args: any[]): void {
    if (this.level <= LogLevel.ERROR) {
      console.error(...this.format(args))
    }
  }

  /**
   * 设置日志级别
   * @param level 新的日志级别
   */
  setLevel(level: LogLevel): void {
    this.level = level
  }

  /**
   * 设置日志前缀
   * @param prefix 新的日志前缀
   */
  setPrefix(prefix: string): void {
    this.prefix = prefix
  }

  /**
   * 设置是否显示调用栈信息
   * @param include 是否包含调用栈信息
   */
  setIncludeStack(include: boolean): void {
    this.includeStack = include
  }

  /**
   * 设置是否显示时间戳
   * @param include 是否包含时间戳
   */
  setIncludeTimestamp(include: boolean): void {
    this.includeTimestamp = include
  }

  /**
   * 设置调用栈显示深度
   * @param depth 调用栈深度，1=只显示直接调用者，2=显示调用链
   */
  setStackDepth(depth: number): void {
    this.stackDepth = depth
  }

  /**
   * 调试调用栈信息（临时方法）
   * 用于排查调用栈获取问题
   */
  debugStack(): void {
    const stack = new Error().stack
    if (stack) {
      console.log('=== Stack Trace Debug ===')
      stack.split('\n').forEach((line, index) => {
        console.log(`${index}: ${line}`)
      })
      console.log('=== End Stack Trace ===')
    }
  }
}

/**
 * 默认日志记录器实例
 * 在开发环境下默认显示调用栈信息
 */
export const logger = new Logger({
  includeStack: !isProd, // 生产环境不显示调用栈信息以提高性能
})

/**
 * 创建特定模块的日志记录器
 * @param module 模块名称
 * @param options 其他选项
 * @returns 日志记录器实例
 */
export function createLogger(
  module: string,
  options: Partial<Omit<LoggerOptions, 'prefix'>> = {},
): Logger {
  return new Logger({
    ...options,
    prefix: module,
  })
}

/**
 * 简化的日志接口，直接使用默认日志记录器
 */
export const log = {
  debug: (...args: any[]) => logger.debug(...args),
  info: (...args: any[]) => logger.info(...args),
  warn: (...args: any[]) => logger.warn(...args),
  error: (...args: any[]) => logger.error(...args),
}

// 导出默认日志接口
export default log
