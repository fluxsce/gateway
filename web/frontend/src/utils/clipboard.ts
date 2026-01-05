/**
 * 剪贴板工具函数
 * 处理复制到剪贴板的功能
 */

/**
 * 复制文本到剪贴板
 * @param text 要复制的文本
 * @param successMessage 成功提示消息
 * @param errorMessage 失败提示消息
 */
export function copyToClipboard(
  text: string,
  successMessage = '已复制到剪贴板',
  errorMessage = '复制失败'
): void {
  // 优先使用现代 Clipboard API
  if (navigator.clipboard && navigator.clipboard.writeText) {
    navigator.clipboard
      .writeText(text)
      .then(() => {
        console.log(successMessage)
      })
      .catch((err) => {
        console.error(errorMessage, err)
        // 如果现代 API 失败，尝试降级方案
        fallbackCopyToClipboard(text, successMessage, errorMessage)
      })
  } else {
    // 不支持现代 API，直接使用降级方案
    fallbackCopyToClipboard(text, successMessage, errorMessage)
  }
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
  errorMessage: string
): void {
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
      console.log(successMessage)
    } else {
      console.error(errorMessage)
    }
  } catch (err) {
    console.error(errorMessage, err)
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

