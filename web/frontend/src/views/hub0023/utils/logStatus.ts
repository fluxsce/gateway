/** 访问日志状态判定：错误看 errorCode/errorMessage，备注只看 noteText。 */

export interface GatewayLogStatusFields {
  errorCode?: string
  errorMessage?: string
  noteText?: string
  gatewayFinishedProcessingTime?: string | null
}

/** 是否存在处理错误。 */
export function hasGatewayLogError(row: GatewayLogStatusFields): boolean {
  return !!(row.errorCode || row.errorMessage)
}

/** 展示用备注，只读 noteText。 */
export function getGatewayLogNote(row: GatewayLogStatusFields): string {
  return row.noteText || ''
}

/** 列表“错误信息”列。 */
export function getGatewayLogErrorText(row: GatewayLogStatusFields): string {
  return row.errorMessage || ''
}

export function getProcessingStatusTagType(
  row: GatewayLogStatusFields,
): 'danger' | 'success' | 'warning' {
  if (hasGatewayLogError(row)) return 'danger'
  if (row.gatewayFinishedProcessingTime) return 'success'
  return 'warning'
}

export function getProcessingStatusText(row: GatewayLogStatusFields): string {
  if (hasGatewayLogError(row)) return '异常'
  if (row.gatewayFinishedProcessingTime) return '已完成'
  return '处理中'
}
