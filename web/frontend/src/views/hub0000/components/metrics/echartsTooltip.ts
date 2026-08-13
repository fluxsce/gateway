/**
 * 系统监控图表通用 ECharts tooltip 配置。
 * 使用 position:fixed（见 echartsTooltip.css）避免 transform 绝对定位撑高页面出现双滚动条。
 */

export const HUB_ECHARTS_TOOLTIP_CLASS = 'hub-echarts-tooltip'

type TooltipSize = {
  contentSize: number[]
  viewSize: number[]
}

/**
 * 在图表坐标系内计算 tooltip 位置：优先指针右上方，空间不足时翻转。
 */
export function placeAxisTooltip(
  point: number[],
  size: TooltipSize,
  gap = 12,
): [number, number] {
  const [mouseX, mouseY] = point
  const [boxW, boxH] = size.contentSize
  const [viewW, viewH] = size.viewSize

  let x = mouseX + gap
  if (x + boxW > viewW - gap) {
    x = mouseX - boxW - gap
  }
  if (x < gap) {
    x = gap
  }

  // 优先显示在指针上方，避免底部图表把浮层顶出可视区
  let y = mouseY - boxH - gap
  if (y < gap) {
    y = mouseY + gap
  }
  if (y + boxH > viewH - gap) {
    y = Math.max(gap, viewH - boxH - gap)
  }

  return [x, y]
}

/**
 * 生成 axis 触发的通用 tooltip 选项，可与业务 formatter 等字段合并。
 */
export function createAxisTooltipOptions<T extends Record<string, unknown>>(
  overrides: T = {} as T,
): T & Record<string, unknown> {
  return {
    trigger: 'axis',
    appendTo: 'body',
    confine: true,
    enterable: true,
    className: HUB_ECHARTS_TOOLTIP_CLASS,
    extraCssText: 'pointer-events:auto;',
    position: (
      point: number[],
      _params: unknown,
      _el: unknown,
      _rect: unknown,
      size: TooltipSize,
    ) => placeAxisTooltip(point, size),
    ...overrides,
  }
}
