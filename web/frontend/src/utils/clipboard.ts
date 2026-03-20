/**
 * 剪贴板工具函数
 *
 * 参考 XiRang 的实现：支持 Clipboard API + execCommand 降级，并支持使用全局 $gMessage 提示。
 */

export interface CopyOptions {
  /** 成功提示消息 */
  successMessage?: string
  /** 失败提示消息 */
  errorMessage?: string
  /** 是否显示提示消息 */
  showMessage?: boolean
  /** 失败时是否使用 execCommand 降级方案 */
  useFallback?: boolean
}

export interface CopyResult {
  success: boolean
  method: 'clipboard-api' | 'execCommand' | 'failed'
  error?: Error
}

function notify(type: 'success' | 'error' | 'warning' | 'info', text: string) {
  if (typeof window !== 'undefined') {
    const w = window as any
    if (w?.$gMessage?.[type]) {
      w.$gMessage[type](text)
      return
    }
  }
  if (type === 'error') console.error(text)
  else console.log(text)
}

/**
 * 复制文本到剪贴板（推荐使用）
 *
 * @param text 要复制的文本
 * @param options 复制选项
 */
export async function copyToClipboardAsync(text: string, options: CopyOptions = {}): Promise<CopyResult> {
  const {
    successMessage = '已复制到剪贴板',
    errorMessage = '复制失败',
    showMessage = true,
    useFallback = true,
  } = options

  if (navigator.clipboard && navigator.clipboard.writeText) {
    try {
      await navigator.clipboard.writeText(text)
      if (showMessage) notify('success', successMessage)
      return { success: true, method: 'clipboard-api' }
    } catch (err) {
      console.error('Clipboard API 失败:', err)
      if (useFallback) {
        return fallbackCopyToClipboard(text, successMessage, errorMessage, showMessage)
      }
      if (showMessage) notify('error', errorMessage)
      return {
        success: false,
        method: 'failed',
        error: err instanceof Error ? err : new Error(String(err)),
      }
    }
  }

  if (useFallback) {
    return fallbackCopyToClipboard(text, successMessage, errorMessage, showMessage)
  }

  if (showMessage) notify('error', '当前浏览器不支持复制功能')
  return { success: false, method: 'failed', error: new Error('Clipboard API not supported') }
}

/**
 * 复制文本到剪贴板（兼容旧调用：不需要 await）
 *
 * @param text 要复制的文本
 * @param successMessage 成功提示消息
 * @param errorMessage 失败提示消息
 */
export function copyToClipboard(text: string, successMessage = '已复制到剪贴板', errorMessage = '复制失败'): void {
  void copyToClipboardAsync(text, { successMessage, errorMessage, showMessage: true, useFallback: true })
}

/**
 * 降级的复制方法（兼容旧浏览器）
 * @param text 要复制的文本
 * @param successMessage 成功提示消息
 * @param errorMessage 失败提示消息
 */
function fallbackCopyToClipboard(
  text: string,
  successMessage: string,
  errorMessage: string,
  showMessage: boolean
): CopyResult {
  const textArea = document.createElement('textarea')
  textArea.value = text

  // 样式设置：确保不影响页面布局
  textArea.style.position = 'fixed'
  textArea.style.left = '-999999px'
  textArea.style.top = '-999999px'
  textArea.style.opacity = '0'
  textArea.style.pointerEvents = 'none'

  document.body.appendChild(textArea)
  textArea.focus()
  textArea.select()

  try {
    const successful = document.execCommand('copy')
    if (successful) {
      if (showMessage) notify('success', successMessage)
      return { success: true, method: 'execCommand' }
    } else {
      if (showMessage) notify('error', errorMessage)
      return { success: false, method: 'failed', error: new Error('execCommand copy failed') }
    }
  } catch (err) {
    console.error(errorMessage, err)
    if (showMessage) notify('error', errorMessage)
    return { success: false, method: 'failed', error: err instanceof Error ? err : new Error(String(err)) }
  } finally {
    document.body.removeChild(textArea)
  }
}

/**
 * 复制对象到剪贴板（转换为 JSON 格式）
 * @param obj 要复制的对象
 * @param pretty 是否格式化输出
 */
export function copyObjectToClipboard(obj: any, pretty = true): void {
  const text = pretty ? JSON.stringify(obj, null, 2) : JSON.stringify(obj)
  copyToClipboard(text)
}

/**
 * 复制数组到剪贴板（转换为 JSON 格式）
 * @param arr 要复制的数组
 * @param pretty 是否格式化输出
 */
export function copyArrayToClipboard(arr: any[], pretty = true): void {
  const text = pretty ? JSON.stringify(arr, null, 2) : JSON.stringify(arr)
  copyToClipboard(text)
}

/**
 * 检查是否支持 Clipboard API
 */
export function isClipboardSupported(): boolean {
  return !!(navigator.clipboard && navigator.clipboard.writeText)
}

