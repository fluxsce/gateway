/**
 * GModal 边框缩放：在弹窗可见时根据 data-g-modal 定位对话框，
 * 通过东 / 南 / 东南 三条边更新像素宽高（居中弹窗下仅拉右下方向，避免改 margin 定位）。
 */

import type { Ref } from 'vue'
import { nextTick, onBeforeUnmount, ref, watch } from 'vue'

export type GModalResizeEdge = 'e' | 's' | 'se'

export interface UseGModalResizeOptions {
  /** 与模板中 data-g-modal 一致 */
  instanceId: string
  visible: Ref<boolean>
  resizable: Ref<boolean>
  isFullscreen: Ref<boolean>
  getMinWidth: () => number
  getMinHeight: () => number
  getMaxWidth: () => number
  getMaxHeight: () => number
  /** 写入当前像素宽高，供 modalStyle 使用 */
  panelPixelWidth: Ref<number | null>
  panelPixelHeight: Ref<number | null>
  /** 拖拽结束回调 */
  onResizeEnd?: () => void
}

/**
 * 绑定 ResizeObserver 与窗口事件，维护对话框包围盒供缩放手柄定位。
 */
export function useGModalResize(options: UseGModalResizeOptions): {
  handleStyles: Ref<Record<GModalResizeEdge, Record<string, string>>>
  startResize: (e: MouseEvent, edge: GModalResizeEdge) => void
  /** 进入动画结束后再算包围盒，避免首帧 transform 未结束时定位偏差 */
  syncHandlePositions: () => void
} {
  const dialogRect = ref<DOMRect | null>(null)
  const handleStyles = ref<Record<GModalResizeEdge, Record<string, string>>>({
    e: { display: 'none' },
    s: { display: 'none' },
    se: { display: 'none' },
  })

  let resizeObserver: ResizeObserver | null = null
  let rafId = 0

  function getDialogEl(): HTMLElement | null {
    return document.querySelector(`[data-g-modal="${options.instanceId}"]`) as HTMLElement | null
  }

  function updateRect(): void {
    if (!options.visible.value || options.isFullscreen.value || !options.resizable.value) {
      dialogRect.value = null
      handleStyles.value = { e: { display: 'none' }, s: { display: 'none' }, se: { display: 'none' } }
      return
    }
    const el = getDialogEl()
    if (!el) {
      dialogRect.value = null
      return
    }
    dialogRect.value = el.getBoundingClientRect()
    const r = dialogRect.value
    const z = 200000
    const hit = 8
    handleStyles.value = {
      e: {
        display: 'block',
        position: 'fixed',
        left: `${r.right - hit / 2}px`,
        top: `${r.top}px`,
        width: `${hit}px`,
        height: `${r.height}px`,
        zIndex: String(z),
        cursor: 'ew-resize',
      },
      s: {
        display: 'block',
        position: 'fixed',
        left: `${r.left}px`,
        top: `${r.bottom - hit / 2}px`,
        width: `${r.width}px`,
        height: `${hit}px`,
        zIndex: String(z),
        cursor: 'ns-resize',
      },
      se: {
        display: 'block',
        position: 'fixed',
        left: `${r.right - hit / 2}px`,
        top: `${r.bottom - hit / 2}px`,
        width: `${hit}px`,
        height: `${hit}px`,
        zIndex: String(z + 1),
        cursor: 'nwse-resize',
      },
    }
  }

  function scheduleUpdate(): void {
    cancelAnimationFrame(rafId)
    rafId = requestAnimationFrame(() => updateRect())
  }

  /**
   * 进入动画结束或布局稳定后再同步手柄：首帧 dialog 可能仍处于 Naive fade-in-scale-up 的 transform，
   * getBoundingClientRect 会偏小/偏移；滚动任意可滚区域也会触发 scheduleUpdate，故滚动后看起来「变准」。
   */
  function syncHandlePositions(): void {
    scheduleUpdate()
    requestAnimationFrame(() => {
      scheduleUpdate()
      requestAnimationFrame(() => scheduleUpdate())
    })
  }

  /** 若干帧内重试：VLazyTeleport / 过渡首帧时对话框可能尚未挂上 data-g-modal，避免缩放手柄一直为 display:none */
  function bindObserver(attempt = 0): void {
    unbindObserver()
    const el = getDialogEl()
    if (!el) {
      if (attempt < 24) {
        requestAnimationFrame(() => bindObserver(attempt + 1))
      }
      return
    }
    resizeObserver = new ResizeObserver(() => scheduleUpdate())
    resizeObserver.observe(el)
    scheduleUpdate()
  }

  function unbindObserver(): void {
    resizeObserver?.disconnect()
    resizeObserver = null
  }

  watch(
    () => [options.visible.value, options.resizable.value, options.isFullscreen.value] as const,
    async ([vis, res, fs]) => {
      unbindObserver()
      if (!vis || !res || fs) {
        dialogRect.value = null
        return
      }
      await nextTick()
      requestAnimationFrame(() => bindObserver())
    },
    { flush: 'post' }
  )

  if (typeof window !== 'undefined') {
    window.addEventListener('resize', scheduleUpdate)
    window.addEventListener('scroll', scheduleUpdate, true)
  }

  onBeforeUnmount(() => {
    unbindObserver()
    cancelAnimationFrame(rafId)
    if (typeof window !== 'undefined') {
      window.removeEventListener('resize', scheduleUpdate)
      window.removeEventListener('scroll', scheduleUpdate, true)
    }
  })

  let drag:
    | {
        edge: GModalResizeEdge
        startX: number
        startY: number
        startW: number
        startH: number
      }
    | null = null

  function clamp(n: number, lo: number, hi: number): number {
    return Math.min(hi, Math.max(lo, n))
  }

  function onMove(e: MouseEvent): void {
    if (!drag) {
      return
    }
    const minW = options.getMinWidth()
    const minH = options.getMinHeight()
    const maxW = options.getMaxWidth()
    const maxH = options.getMaxHeight()
    let nw = drag.startW
    let nh = drag.startH
    if (drag.edge === 'e' || drag.edge === 'se') {
      nw = clamp(drag.startW + (e.clientX - drag.startX), minW, maxW)
    }
    if (drag.edge === 's' || drag.edge === 'se') {
      nh = clamp(drag.startH + (e.clientY - drag.startY), minH, maxH)
    }
    options.panelPixelWidth.value = Math.round(nw)
    options.panelPixelHeight.value = Math.round(nh)
    scheduleUpdate()
  }

  function onUp(): void {
    document.removeEventListener('mousemove', onMove)
    document.removeEventListener('mouseup', onUp)
    drag = null
    document.body.style.cursor = ''
    document.body.style.userSelect = ''
    options.onResizeEnd?.()
  }

  function startResize(e: MouseEvent, edge: GModalResizeEdge): void {
    if (options.isFullscreen.value || !options.resizable.value) {
      return
    }
    const el = getDialogEl()
    if (!el) {
      return
    }
    const r = el.getBoundingClientRect()
    drag = {
      edge,
      startX: e.clientX,
      startY: e.clientY,
      startW: r.width,
      startH: r.height,
    }
    e.preventDefault()
    document.body.style.cursor = edge === 'e' ? 'ew-resize' : edge === 's' ? 'ns-resize' : 'nwse-resize'
    document.body.style.userSelect = 'none'
    document.addEventListener('mousemove', onMove)
    document.addEventListener('mouseup', onUp)
  }

  return {
    handleStyles,
    startResize,
    syncHandlePositions,
  }
}
