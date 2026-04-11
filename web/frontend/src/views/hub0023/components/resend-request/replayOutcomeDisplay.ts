import type { InjectionKey, Ref } from 'vue'

/** 左侧列表每条 trace 的重发阶段 */
export type ReplayRowPhase = 'none' | 'sending' | 'success' | 'failed'

/** 左侧列表每条 trace 的重发阶段与最近一次 HTTP 摘要 */
export interface ReplayRowOutcome {
  phase: ReplayRowPhase
  httpStatus: number | null
  responseLine: string
}

export const defaultReplayOutcome: ReplayRowOutcome = {
  phase: 'none',
  httpStatus: null,
  responseLine: '—',
}

/** 供 {@link ResendTraceListItem} inject，避免父级整表随单行 patch 重渲染 */
export const resendReplayOutcomeKey: InjectionKey<Ref<Record<string, ReplayRowOutcome>>> = Symbol(
  'resendReplayOutcome'
)

export function getReplayOutcome(
  store: Record<string, ReplayRowOutcome>,
  traceId: string
): ReplayRowOutcome {
  return store[traceId] ?? defaultReplayOutcome
}

export function replayPhaseLabel(o: ReplayRowOutcome): string {
  if (o.phase === 'sending') {
    return '发送中'
  }
  if (o.phase === 'success') {
    return '成功'
  }
  if (o.phase === 'failed') {
    return '失败'
  }
  return '未重发'
}

export function replayPhaseTagType(
  o: ReplayRowOutcome
): 'default' | 'info' | 'success' | 'error' | 'warning' {
  if (o.phase === 'sending') {
    return 'warning'
  }
  if (o.phase === 'success') {
    return 'success'
  }
  if (o.phase === 'failed') {
    return 'error'
  }
  return 'default'
}

export function responseStateLabel(o: ReplayRowOutcome): string {
  if (o.phase === 'none' || o.phase === 'sending') {
    return '—'
  }
  return o.responseLine
}

export function httpStatusTagType(
  o: ReplayRowOutcome
): 'default' | 'info' | 'success' | 'error' | 'warning' {
  if (o.phase !== 'success' || o.httpStatus == null) {
    return o.phase === 'failed' ? 'error' : 'default'
  }
  const s = o.httpStatus
  if (s >= 200 && s < 300) {
    return 'success'
  }
  if (s >= 400 && s < 500) {
    return 'warning'
  }
  if (s >= 500 || s === 0) {
    return 'error'
  }
  return 'default'
}
