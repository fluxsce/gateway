/**
 * 被动事件监听器配置工具
 * 解决 "Added non-passive event listener to a scroll-blocking 'wheel' event" 警告
 * 提高页面滚动性能
 */

export interface PassiveEventsConfig {
  /** 是否启用被动事件监听器 */
  enabled?: boolean
  /** 额外的排除选择器 */
  excludeSelectors?: string[]
  /** 是否在控制台显示日志 */
  verbose?: boolean
}

// 需要设置为被动模式的事件类型
const PASSIVE_EVENTS = [
  'wheel',
  'mousewheel', 
  'touchstart',
  'touchmove',
  'touchend',
  'touchcancel'
] as const

// 需要排除被动模式的选择器（这些元素可能需要阻止默认行为）
const EXCLUDE_SELECTORS = [
  '.n-scrollbar',           // Naive UI 滚动条
  '.n-data-table',          // 数据表格
  '.n-virtual-list',        // 虚拟列表
  '[data-custom-scroll]',   // 自定义滚动标记
  '.chart-container',       // 图表容器
  '.map-container',         // 地图容器
  '.image-viewer',          // 图片查看器
  '.draggable'              // 可拖拽元素
]



/**
 * 配置被动事件监听器
 * 通过重写 addEventListener 方法，为指定事件类型智能添加 passive: true 选项
 */
export function setupPassiveEvents(config: PassiveEventsConfig = {}): void {
  const { enabled = true, excludeSelectors = [], verbose = false } = config
  
  if (typeof window === 'undefined' || !enabled) {
    return
  }

  // 合并排除选择器
  const allExcludeSelectors = [...EXCLUDE_SELECTORS, ...excludeSelectors]

  /**
   * 检查元素是否应该排除被动模式
   */
  function shouldExcludePassive(target: EventTarget | null): boolean {
    if (!target || !(target instanceof Element)) {
      return false
    }

    // 检查元素本身或其父元素是否匹配排除选择器
    let element: Element | null = target
    while (element) {
      for (const selector of allExcludeSelectors) {
        if (element.matches?.(selector) || element.closest?.(selector)) {
          return true
        }
      }
      element = element.parentElement
    }

    return false
  }

  // 保存原始的 addEventListener 方法
  const originalAddEventListener = EventTarget.prototype.addEventListener

  // 重写 addEventListener 方法
  EventTarget.prototype.addEventListener = function(
    type: string,
    listener: EventListenerOrEventListenerObject,
    options?: boolean | AddEventListenerOptions
  ) {
    // 检查是否是需要被动处理的事件类型
    if (PASSIVE_EVENTS.includes(type as any)) {
      // 处理不同的 options 参数格式
      let finalOptions: AddEventListenerOptions

      // 检查是否已经显式设置了 passive 为 false
      if (typeof options === 'object' && options !== null && options.passive === false) {
        // 如果显式设置为 false，保持原样但添加日志
        if (verbose) {
          console.log(`🔧 事件 ${type} 显式设置为非被动模式`)
        }
        finalOptions = options
      } else if (shouldExcludePassive(this)) {
        // 如果应该排除被动模式，设置为非被动但不阻止默认行为
        if (typeof options === 'boolean') {
          finalOptions = { capture: options, passive: false }
        } else if (typeof options === 'object' && options !== null) {
          finalOptions = { ...options, passive: false }
        } else {
          finalOptions = { passive: false }
        }
        if (verbose) {
          console.log(`🔧 事件 ${type} 因排除规则设置为非被动模式`)
        }
      } else {
        // 默认设置为被动模式
        if (typeof options === 'boolean') {
          finalOptions = { capture: options, passive: true }
        } else if (typeof options === 'object' && options !== null) {
          finalOptions = { ...options, passive: true }
        } else {
          finalOptions = { passive: true }
        }
      }

      return originalAddEventListener.call(this, type, listener, finalOptions)
    }

    // 对于其他事件类型，使用原始方法
    return originalAddEventListener.call(this, type, listener, options)
  }

  if (verbose) {
    console.log('✅ 智能被动事件监听器配置已启用')
  }
}

/**
 * 恢复原始的 addEventListener 方法
 * 用于测试或特殊情况下需要禁用被动事件监听器
 */
export function restoreEventListeners(): void {
  if (typeof window === 'undefined') {
    return
  }

  // 这里我们无法直接恢复，因为原始方法已经被覆盖
  // 在实际使用中，通常不需要恢复
  console.warn('⚠️ 被动事件监听器恢复功能暂未实现')
}

/**
 * 检查浏览器是否支持被动事件监听器
 */
export function supportsPassiveEvents(): boolean {
  if (typeof window === 'undefined') {
    return false
  }

  let supportsPassive = false
  
  try {
    const opts = Object.defineProperty({}, 'passive', {
      get() {
        supportsPassive = true
        return false
      }
    })
    
    window.addEventListener('testPassive', null as any, opts)
    window.removeEventListener('testPassive', null as any, opts)
  } catch (e) {
    // 忽略错误
  }
  
  return supportsPassive
}
