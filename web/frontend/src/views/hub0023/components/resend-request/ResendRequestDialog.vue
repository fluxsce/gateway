<template>
  <RsDialog
    v-model:open="showModal"
    title="请求重发"
    layout="window"
    width="92%"
    class="hub0023-resend-dialog"
    :teleport-to="dialogMountTo"
    :draggable="true"
    :fullscreenable="true"
    :show-overlay="true"
    :close-on-overlay-click="false"
    :show-footer="false"
    @after-close="handleAfterLeave"
  >
    <template #body>
    <div ref="dialogRootRef" class="resend-dialog">
      <aside class="resend-dialog__aside">
        <div class="resend-dialog__aside-head">
          <div class="resend-dialog__aside-title">重发列表（TraceId）</div>
          <div class="resend-dialog__aside-actions">
            <RsButton
              v-if="!autoSending"
              size="sm"
              variant="ghost"
              tone="primary"
              :disabled="!props.logs.length || detailLoading"
              @click="startAutoSend"
            >
              自动发送
            </RsButton>
            <RsButton v-else size="sm" variant="ghost" tone="warning" @click="stopAutoSend">
              停止
            </RsButton>
          </div>
        </div>
        <RsScrollbar class="resend-dialog__aside-scroll" height="100%">
          <div class="resend-dialog__aside-list-shell">
            <RsEmpty v-if="!props.logs.length" description="暂无条目" />
            <ul v-else class="resend-dialog__trace-list">
              <ResendTraceListItem
                v-for="item in props.logs"
                :key="item.traceId"
                :item="item"
                :active="item.traceId === selectedTraceId"
                @select="selectTrace"
              />
            </ul>
          </div>
        </RsScrollbar>
      </aside>
      <main class="resend-dialog__main">
        <RsLoading :loading="detailLoading" overlay block size="lg" />
        <g-restful-api
          v-if="selectedTraceId && restfulPanelReady"
          ref="restfulRef"
          class="resend-dialog__restful"
          v-bind="replayBind"
          :response-min-height="'200px'"
          :request-body-min-height="'140px'"
          @send-start="onReplaySendStart"
          @send-end="onReplaySendEnd"
          @success="onReplaySuccess"
          @error="onReplayError"
        />
        <RsEmpty v-else description="请选择左侧 TraceId" />
      </main>
    </div>
    </template>
  </RsDialog>
</template>

<script setup lang="ts">
import GRestfulApi from '@/components/grestful-api/GRestfulApi.vue'
import type { GRestfulApiProps, RestRequestResult } from '@/components/grestful-api/types'
import { RsButton, RsDialog, RsEmpty, RsLoading, RsScrollbar } from '@/ui'
import { getApiMessage, isApiSuccess, parseJsonData } from '@/utils/format'
import { computed, nextTick, provide, ref, watch } from 'vue'
import { getGatewayLogAccessDetail } from '../../api'
import type { GatewayLogInfo, GatewayLogListItem } from '../../types'
import { buildGatewayLogReplayInit } from './mapGatewayLogReplay'
import {
  defaultReplayOutcome,
  getReplayOutcome,
  resendReplayOutcomeKey,
  type ReplayRowOutcome,
} from './replayOutcomeDisplay'
import ResendTraceListItem from './ResendTraceListItem.vue'

/** GRestfulApi 通过 defineExpose 暴露的方法 */
interface GRestfulApiExposed {
  send: () => Promise<void>
  cancel: () => void
  rehydrateFromInitialProps: () => void
}

interface Props {
  /** 是否显示弹窗 */
  visible: boolean
  /** 待重发的日志行（至少含 traceId，左侧列表与摘要展示用） */
  logs: GatewayLogListItem[]
  /**
   * 弹层挂载容器的元素 id（不含 #），例如 GatewayLogQuery 根节点；
   * 传入后弹层 teleport 到该节点内，避免默认挂 body 在多页签下盖住其它标签。
   */
  mountContainerId?: string
}

interface Emits {
  (e: 'update:visible', value: boolean): void
}

const props = withDefaults(defineProps<Props>(), {
  visible: false,
  logs: () => [],
  mountContainerId: '',
})

const emit = defineEmits<Emits>()

const showModal = computed({
  get: () => props.visible,
  set: (v: boolean) => emit('update:visible', v),
})

/** 透传 RsDialog `teleport-to`；无有效 id 时不传，保持默认挂载行为 */
const dialogMountTo = computed(() => {
  const id = props.mountContainerId?.trim()
  if (!id) {
    return undefined
  }
  return `#${id}`
})

const selectedTraceId = ref('')
const detailLoading = ref(false)
/** 首次成功拉取到 replayBind 后置 true，关闭弹窗时复位；右侧 GRestfulApi 单实例常驻，避免 :key 整树重建 */
const restfulPanelReady = ref(false)
/** 最近一次已完成详情加载并写入 replayBind 的 traceId，供自动发送等待与 waitReplayReady 判断 */
const lastAppliedReplayTraceId = ref('')
const replayBind = ref<GRestfulApiProps>({})
const dialogRootRef = ref<HTMLElement | null>(null)
const restfulRef = ref<GRestfulApiExposed | null>(null)

/** 本次弹窗内已通过「自动发送」执行过请求的 traceId */
const autoSentTraceIds = ref<Set<string>>(new Set())
const autoSending = ref(false)
const autoSendCancelled = ref(false)

/** 当前一次代发请求对应的 traceId（与 send-start 时选中行一致，供 success/error 写回左侧状态） */
const replayInFlightTraceId = ref('')

const replayOutcomeByTraceId = ref<Record<string, ReplayRowOutcome>>({})

provide(resendReplayOutcomeKey, replayOutcomeByTraceId)

/**
 * 弹窗关闭后清空右侧状态，避免下次打开短暂闪现旧请求。
 */
function handleAfterLeave(): void {
  selectedTraceId.value = ''
  restfulPanelReady.value = false
  lastAppliedReplayTraceId.value = ''
  replayBind.value = {}
  autoSentTraceIds.value = new Set()
  autoSending.value = false
  autoSendCancelled.value = false
  replayInFlightTraceId.value = ''
  replayOutcomeByTraceId.value = {}
}

function patchReplayOutcome(traceId: string, patch: Partial<ReplayRowOutcome>): void {
  const rec = replayOutcomeByTraceId.value
  const cur = rec[traceId]
  if (!cur) {
    rec[traceId] = { ...defaultReplayOutcome, ...patch }
  } else {
    Object.assign(cur, patch)
  }
}

function onReplaySendStart(): void {
  const tid = selectedTraceId.value
  if (!tid) {
    return
  }
  replayInFlightTraceId.value = tid
  patchReplayOutcome(tid, {
    phase: 'sending',
    httpStatus: null,
    responseLine: '—',
  })
}

/**
 * 去掉开头重复的「状态码 状态码 …」（例如后端给出 `200 200 OK`）。
 */
function dedupeLeadingHttpStatusCode(line: string): string {
  let s = line.trim()
  let prev = ''
  while (s !== prev) {
    prev = s
    s = s.replace(/^(\d{3})(\s+)\1(\s+)/, '$1$3')
  }
  return s
}

/**
 * 拼一条「响应状态」展示文案，避免 statusText 已含状态码时出现「404 404 Not Found」重复。
 */
function formatReplayResponseLine(status: number, statusText: string | undefined): string {
  const st = statusText?.trim() ?? ''
  const codeStr = String(status)
  let line: string
  if (!st) {
    line = codeStr
  } else if (st === codeStr || st.startsWith(`${codeStr} `) || st.startsWith(`${codeStr}\t`)) {
    line = st
  } else {
    line = `${codeStr} ${st}`
  }
  return dedupeLeadingHttpStatusCode(line)
}

function onReplaySuccess(payload: RestRequestResult): void {
  const tid = replayInFlightTraceId.value || selectedTraceId.value
  if (!tid) {
    return
  }
  patchReplayOutcome(tid, {
    phase: 'success',
    httpStatus: payload.status,
    responseLine: formatReplayResponseLine(payload.status, payload.statusText),
  })
}

function onReplayError(): void {
  const tid = replayInFlightTraceId.value || selectedTraceId.value
  if (!tid) {
    return
  }
  patchReplayOutcome(tid, {
    phase: 'failed',
    httpStatus: 0,
    responseLine: '失败',
  })
}

/**
 * 若 await 抛错等未走到 success/error，仍将「发送中」收敛为失败，避免列表卡死。
 */
function onReplaySendEnd(): void {
  const tid = replayInFlightTraceId.value
  if (tid) {
    const cur = getReplayOutcome(replayOutcomeByTraceId.value, tid)
    if (cur.phase === 'sending') {
      patchReplayOutcome(tid, {
        phase: 'failed',
        httpStatus: null,
        responseLine: '中断',
      })
    }
  }
  replayInFlightTraceId.value = ''
}

function selectTrace(traceId: string): void {
  if (selectedTraceId.value === traceId) {
    return
  }
  selectedTraceId.value = traceId
}

/**
 * 将当前 Trace 行滚入左侧可视区。
 * RsScrollbar 的可滚层是 .rs-scrollbar__viewport，对子节点直接 scrollIntoView 往往滚的是外层模态，自动发送时列表不跟随；改为相对容器修正 scrollTop。
 */
function scrollTraceIntoView(traceId: string): void {
  const root = dialogRootRef.value
  if (!root || typeof CSS === 'undefined' || !CSS.escape) {
    return
  }
  const el = root.querySelector(`[data-trace-id="${CSS.escape(traceId)}"]`) as HTMLElement | null
  if (!el) {
    return
  }

  const container = root.querySelector(
    '.resend-dialog__aside-scroll .rs-scrollbar__viewport'
  ) as HTMLElement | null
  if (!container) {
    el.scrollIntoView({ block: 'nearest', behavior: 'auto' })
    return
  }

  const margin = 8
  const c = container.getBoundingClientRect()
  const r = el.getBoundingClientRect()
  if (r.top < c.top + margin) {
    container.scrollTop += r.top - c.top - margin
  } else if (r.bottom > c.bottom - margin) {
    container.scrollTop += r.bottom - c.bottom + margin
  }
}

/**
 * 等待当前 trace 详情加载完成且右侧 GRestfulApi 已就绪（与 lastAppliedReplayTraceId 对齐）。
 */
async function waitReplayReady(traceId: string, timeoutMs = 60000): Promise<void> {
  const start = Date.now()
  while (Date.now() - start < timeoutMs) {
    if (!props.visible || selectedTraceId.value !== traceId) {
      return
    }
    if (
      !detailLoading.value &&
      restfulPanelReady.value &&
      lastAppliedReplayTraceId.value === traceId
    ) {
      await nextTick()
      return
    }
    await new Promise<void>((r) => {
      setTimeout(r, 40)
    })
  }
  throw new Error('等待重发表单超时')
}

function stopAutoSend(): void {
  autoSendCancelled.value = true
  restfulRef.value?.cancel()
}

/**
 * 按列表顺序对「本次弹窗内尚未自动发送过」的条目：滚动到可视区、加载表单、调用发送。
 */
async function startAutoSend(): Promise<void> {
  if (!props.logs.length || autoSending.value) {
    return
  }
  autoSendCancelled.value = false
  autoSending.value = true
  const pending = props.logs.filter((l) => !autoSentTraceIds.value.has(l.traceId))
  try {
    for (const item of pending) {
      if (autoSendCancelled.value || !props.visible) {
        break
      }
      selectedTraceId.value = item.traceId
      await nextTick()
      scrollTraceIntoView(item.traceId)
      try {
        await waitReplayReady(item.traceId)
      } catch {
        break
      }
      if (autoSendCancelled.value || !props.visible) {
        break
      }
      await nextTick()
      let api = restfulRef.value
      let spin = 0
      while (!api && spin < 40 && props.visible && !autoSendCancelled.value) {
        await nextTick()
        await new Promise<void>((r) => setTimeout(r, 20))
        api = restfulRef.value
        spin++
      }
      if (!api || autoSendCancelled.value) {
        continue
      }
      await api.send()
      autoSentTraceIds.value.add(item.traceId)
      autoSentTraceIds.value = new Set(autoSentTraceIds.value)
    }
  } finally {
    autoSending.value = false
  }
}

/**
 * 拉取单条日志详情并映射为 GRestfulApi 初始参数。
 */
async function loadReplayForTrace(traceId: string): Promise<void> {
  if (!traceId) {
    replayBind.value = {}
    lastAppliedReplayTraceId.value = ''
    restfulPanelReady.value = false
    return
  }
  detailLoading.value = true
  try {
    const row = props.logs.find((l) => l.traceId === traceId)
    const gid = String(row?.gatewayInstanceId ?? '').trim()
    if (!gid) {
      replayBind.value = {
        initialUrl: '',
        initialMethod: 'GET',
        initialResponseBody: '当前日志行缺少网关实例 ID，无法拉取详情',
        initialResponseStatusText: '详情加载失败',
      }
      return
    }
    const response = await getGatewayLogAccessDetail({
      traceId,
      gatewayInstanceId: gid,
    })
    if (!isApiSuccess(response)) {
      const msg = getApiMessage(response, '获取日志详情失败')
      replayBind.value = {
        initialUrl: '',
        initialMethod: 'GET',
        initialResponseBody: msg,
        initialResponseStatusText: '详情加载失败',
      }
      return
    }
    const data = parseJsonData<GatewayLogInfo>(response)
    if (!props.visible || selectedTraceId.value !== traceId) {
      return
    }
    const init = buildGatewayLogReplayInit(data)
    replayBind.value = { ...init }
  } catch (e) {
    const msg = e instanceof Error ? e.message : '加载失败，请重试'
    replayBind.value = {
      initialUrl: '',
      initialMethod: 'GET',
      initialResponseBody: msg,
      initialResponseStatusText: '详情加载异常',
    }
  } finally {
    detailLoading.value = false
  }

  if (!props.visible || selectedTraceId.value !== traceId) {
    return
  }
  restfulPanelReady.value = true
  lastAppliedReplayTraceId.value = traceId
  await nextTick()
  restfulRef.value?.rehydrateFromInitialProps()
}

watch(
  () => props.visible,
  (v) => {
    if (v && props.logs.length > 0) {
      selectedTraceId.value = props.logs[0].traceId
    }
  },
  { immediate: true }
)

watch(
  () => [props.visible, selectedTraceId.value] as const,
  ([vis, tid]) => {
    if (vis && tid) {
      void loadReplayForTrace(tid)
    }
  },
  { immediate: true }
)

watch(
  () => props.logs,
  (list) => {
    if (!props.visible || list.length === 0) {
      return
    }
    const exists = list.some((l) => l.traceId === selectedTraceId.value)
    if (!exists) {
      selectedTraceId.value = list[0].traceId
    }
  },
  { deep: true }
)
</script>

<style scoped lang="scss">
.resend-dialog {
  display: flex;
  align-items: stretch;
  gap: 12px;
  flex: 1;
  min-width: 0;
  min-height: 640px;
  height: 100%;
  max-height: min(88vh, 900px);
}

.resend-dialog__aside {
  flex: 0 0 300px;
  align-self: stretch;
  min-width: 260px;
  min-height: 0;
  display: flex;
  flex-direction: column;
  border: 1px solid var(--g-border-primary);
  border-radius: var(--g-radius-md);
  overflow: hidden;
  background: var(--g-dialog-bg);
}

.resend-dialog__aside-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 8px 10px;
  border-bottom: 1px solid var(--g-border-primary);
}

.resend-dialog__aside-title {
  font-size: 13px;
  font-weight: 500;
  color: var(--g-text-secondary);
  min-width: 0;
}

.resend-dialog__aside-actions {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 6px;
}

.resend-dialog__aside-scroll {
  flex: 1 1 auto;
  min-height: 0;
  height: 0;
  border: none;
  border-radius: 0;
  background: transparent;
}

.resend-dialog__aside-list-shell {
  flex: 1 1 auto;
  min-height: 100%;
  height: 100%;
  padding: 6px 4px 10px;
  /* 列表区略深于卡片，便于卡片「浮起」 */
  background: linear-gradient(180deg, var(--g-bg-secondary) 0%, var(--g-bg-tertiary) 100%);
  box-sizing: border-box;
}

.resend-dialog__trace-list {
  list-style: none;
  margin: 0;
  padding: 6px 6px 8px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.resend-dialog__main {
  position: relative;
  flex: 1;
  min-width: 0;
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border: 1px solid var(--g-border-primary);
  border-radius: var(--g-radius-md);
  background: var(--g-dialog-bg);
}

.resend-dialog__restful {
  flex: 1;
  min-width: 0;
  min-height: 0;
  width: 100%;
  height: 100%;
}
</style>
