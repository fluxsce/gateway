<template>
  <section
    class="g-restful-api"
    :class="props.class"
  >
    <div class="g-restful-api__urlbar">
      <div class="g-restful-api__url-group">
        <RsSelect
          v-model="method"
          class="g-restful-api__method"
          :class="methodModifierClass"
          :options="METHOD_OPTIONS"
          :consistent-menu-width="false"
          size="sm"
        />
        <RsInput
          v-model="url"
          class="g-restful-api__url"
          type="text"
          placeholder="请输入请求 URL（支持相对路径，将相对当前站点）"
          clearable
          size="sm"
          @keyup.enter="handleSend"
        />
      </div>
      <div class="g-restful-api__send-group">
        <RsButton
          variant="primary"
          size="sm"
          :loading="sending"
          :disabled="sending || !canExecuteHttp"
          @click="handleSend"
        >
          发送
        </RsButton>
        <RsButton
          v-if="sending"
          size="sm"
          variant="ghost"
          @click="handleCancel"
        >
          取消
        </RsButton>
      </div>
    </div>

    <RsSplitPane
      class="g-restful-api__split"
      orientation="vertical"
      with-handle
      :panes="splitPanes"
    >
      <template #request>
        <div
          class="g-restful-api__request"
          :class="{ 'g-restful-api__request--fill': requestPaneIsEditor }"
        >
          <RsTabs
            v-model="requestTab"
            :items="requestTabItems"
            variant="line"
            size="sm"
            panelless
            class="g-restful-api__tabs"
          />
          <div class="g-restful-api__pane" :class="{ 'g-restful-api__pane--editor': requestPaneIsEditor }">
            <div v-if="requestTab === 'params'" class="g-restful-api__pane-inner">
              <key-value-editor
                v-model:rows="queryParams"
                variant="table"
                table-variant="query"
              />
            </div>

            <div v-else-if="requestTab === 'body'" class="g-restful-api__pane-inner g-restful-api__body-pane">
              <RsRadio
                class="g-restful-api__body-radio-group"
                :value="bodyProcessType"
                name="rest-body-process"
                size="sm"
                @update:model-value="setBodyProcessType"
              >
                <RsRadioItem
                  v-for="opt in BODY_PROCESS_OPTIONS"
                  :key="opt.value"
                  :value="opt.value"
                >
                  {{ opt.label }}
                </RsRadioItem>
              </RsRadio>

              <div
                v-if="bodyProcessType === 'none'"
                class="g-restful-api__empty"
              >
                该请求没有 Body
              </div>

              <div
                v-else-if="bodyProcessType === 'x-www-form-urlencoded' || bodyProcessType === 'form-data'"
                class="g-restful-api__body-form"
              >
                <key-value-editor
                  v-model:rows="formFields"
                  variant="table"
                  table-variant="form"
                  :form-table-kind="bodyProcessType === 'form-data' ? 'multipart' : 'urlencoded'"
                  key-column-label="参数名"
                  type-column-label="类型"
                  value-column-label="参数值"
                />
              </div>

              <div
                v-else
                class="g-restful-api__body-raw"
              >
                <div
                  v-if="bodyProcessType === 'raw'"
                  class="g-restful-api__raw-ct-row"
                >
                  <RsSelect
                    v-model="rawContentType"
                    class="g-restful-api__content-type"
                    :options="RAW_CONTENT_TYPE_OPTIONS"
                    size="sm"
                  />
                </div>
                <RsCodeEditor
                  v-model="rawBody"
                  :language="rawBodyLanguage"
                  height="100%"
                  embedded
                  :show-toolbar="false"
                />
              </div>
            </div>

            <div v-else-if="requestTab === 'headers'" class="g-restful-api__pane-inner">
              <key-value-editor
                v-model:rows="headerRows"
                variant="table"
                table-variant="query"
                key-column-label="名称"
                value-column-label="值"
                show-auto-body-hint
              />
            </div>

            <div v-else-if="requestTab === 'cookies'" class="g-restful-api__pane-inner">
              <key-value-editor
                v-model:rows="cookieRows"
                variant="table"
                table-variant="query"
                key-column-label="名称"
                value-column-label="值"
              />
            </div>

            <div v-else class="g-restful-api__pane-inner g-restful-api__auth-pane">
              <RsRadio
                v-model="authType"
                class="g-restful-api__auth-type-group"
                size="sm"
                name="rest-auth-type"
              >
                <RsRadioItem value="none">无</RsRadioItem>
                <RsRadioItem value="bearer">Bearer Token</RsRadioItem>
                <RsRadioItem value="basic">Basic Auth</RsRadioItem>
                <RsRadioItem value="apikey">API Key</RsRadioItem>
              </RsRadio>

              <div
                v-if="authType === 'bearer'"
                class="g-restful-api__auth-fields"
              >
                <RsInput
                  v-model="bearerToken"
                  type="password"
                  show-password-on="click"
                  placeholder="Token"
                  size="sm"
                />
              </div>

              <div
                v-else-if="authType === 'basic'"
                class="g-restful-api__auth-fields g-restful-api__auth-fields--row"
              >
                <RsInput
                  v-model="basicUser"
                  placeholder="用户名"
                  size="sm"
                />
                <RsInput
                  v-model="basicPassword"
                  type="password"
                  show-password-on="click"
                  placeholder="密码"
                  size="sm"
                />
              </div>

              <div
                v-else-if="authType === 'apikey'"
                class="g-restful-api__auth-fields"
              >
                <RsRadio
                  v-model="apiKeyIn"
                  size="sm"
                  name="rest-apikey-in"
                >
                  <RsRadioItem value="header">Header</RsRadioItem>
                  <RsRadioItem value="query">Query</RsRadioItem>
                </RsRadio>
                <RsInput
                  v-if="apiKeyIn === 'header'"
                  v-model="apiKeyHeaderName"
                  placeholder="Header 名称，默认 X-API-Key"
                  size="sm"
                />
                <RsInput
                  v-else
                  v-model="apiKeyQueryName"
                  placeholder="Query 参数名，默认 api_key"
                  size="sm"
                />
                <RsInput
                  v-model="apiKeyValue"
                  type="password"
                  show-password-on="click"
                  placeholder="密钥"
                  size="sm"
                />
              </div>
            </div>
          </div>
        </div>
      </template>

      <template #response>
        <div class="g-restful-api__response">
          <RsTabs
            v-model="responseTab"
            :items="responseTabItems"
            variant="line"
            size="sm"
            panelless
            class="g-restful-api__tabs"
          >
            <template #extra>
              <div class="g-restful-api__meta">
                <span
                  v-if="lastResult"
                  class="g-restful-api__status"
                  :class="`g-restful-api__status--${responseStatusTagType}`"
                >
                  {{ lastResult.statusText?.trim() || String(lastResult.status) }}
                </span>
                <span
                  v-if="lastResult && lastResult.durationMs > 0"
                  class="g-restful-api__time"
                >
                  {{ lastResult.durationMs.toFixed(0) }} ms
                </span>
                <RsButton
                  v-if="responseTab === 'body' && lastResult && hasRealHttpResponse"
                  size="sm"
                  variant="ghost"
                  @click="formatResponseBody"
                >
                  格式化
                </RsButton>
              </div>
            </template>
          </RsTabs>
          <div class="g-restful-api__pane g-restful-api__pane--editor">
            <div
              v-if="sending"
              class="g-restful-api__empty"
            >
              发送中…
            </div>
            <div
              v-else-if="!lastResult"
              class="g-restful-api__empty"
            >
              <RsEmpty description="点击「发送」查看响应" />
            </div>
            <RsCodeEditor
              v-else-if="responseTab === 'body'"
              v-model="responseBodyDisplay"
              :language="responseBodyLanguage"
              readonly
              height="100%"
              embedded
              :show-toolbar="false"
            />
            <RsCodeEditor
              v-else
              v-model="responseHeadersText"
              language="plaintext"
              readonly
              height="100%"
              embedded
              :show-toolbar="false"
            />
          </div>
        </div>
      </template>
    </RsSplitPane>
  </section>
</template>

<script setup lang="ts">
// @ts-nocheck
import { useAppMessage } from '@/composables/useAppMessage'
import {
  RsButton,
  RsCodeEditor,
  RsEmpty,
  RsInput,
  RsRadio,
  RsRadioItem,
  RsSelect,
  RsSplitPane,
  RsTabs,
  type RsCodeEditorLanguage,
  type RsSplitPaneItem,
} from '@/ui'
import { store } from '@/stores'

import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import {
  BODY_PROCESS_OPTIONS,
  cloneKeyValueRows,
  computeDesiredHeaderRows,
  cookieStringFromRows,
  demoteAutoContentTypeRows,
  deriveBodyContentType,
  findHeaderValueCI,
  httpStatusTagType,
  inferRawEditorLanguage,
  METHOD_OPTIONS,
  queryRowsToSearchString,
  RAW_CONTENT_TYPE_OPTIONS,
  restfulMethodModifierClass,
  searchParamsToQueryRows,
  stringifyResponseHeaders,
  upsertHeaderRow
} from './grestfulApiHelpers'
import { sendRestRequest, tryParseUserUrl } from './sendRestRequest'
import type {
  GRestfulApiEmits,
  GRestfulApiProps,
  RestApiKeyLocation,
  RestAuthType,
  RestBodyMode,
  RestBodyProcessType,
  RestHttpMethod,
  RestKeyValueRow,
  RestRequestResult,
  SendRestRequestInput
} from './types'
import { createKeyValueRow } from './types'
import KeyValueEditor from './KeyValueEditor.vue'

defineOptions({
  name: 'GRestfulApi'
})

const props = withDefaults(defineProps<GRestfulApiProps>(), {
  initialUrl: '',
  initialMethod: 'GET',
  responseMinHeight: '220px',
  requestBodyMinHeight: '160px'
})

const emit = defineEmits<GRestfulApiEmits>()

/** 与后端 hubplugin/http/execute 的 RequireButton 一致：日志重发或批量重发。 */
const HTTP_EXECUTE_BUTTON_CODES = ['hub0023:reset', 'hub0023:batchReset']
const canExecuteHttp = computed(() => store.user.hasAnyPermission(HTTP_EXECUTE_BUTTON_CODES))

const message = useAppMessage()

const method = ref<RestHttpMethod>(props.initialMethod)
const url = ref(props.initialUrl)

const queryParams = ref<RestKeyValueRow[]>([createKeyValueRow()])
const headerRows = ref<RestKeyValueRow[]>([createKeyValueRow()])
const cookieRows = ref<RestKeyValueRow[]>([createKeyValueRow()])
const formFields = ref<RestKeyValueRow[]>([createKeyValueRow()])

const authType = ref<RestAuthType>('none')
const bearerToken = ref('')
const basicUser = ref('')
const basicPassword = ref('')
const apiKeyIn = ref<RestApiKeyLocation>('header')
const apiKeyHeaderName = ref('X-API-Key')
const apiKeyQueryName = ref('api_key')
const apiKeyValue = ref('')

/** Body 处理类型（与 Postman 选项对齐），决定表单表格或 Raw 及 Content-Type */
const bodyProcessType = ref<RestBodyProcessType>('none')
const rawBody = ref('')
const rawContentType = ref('application/json')

const requestTab = ref<'params' | 'body' | 'headers' | 'cookies' | 'auth'>('params')
const responseTab = ref<'body' | 'headers'>('body')

/** Body 原始编辑器需要占满分栏剩余高度 */
const requestPaneIsEditor = computed(() => {
  if (requestTab.value !== 'body') {
    return false
  }
  return !['none', 'x-www-form-urlencoded', 'form-data'].includes(bodyProcessType.value)
})

/**
 * Params / Headers 等表格按内容高度；Raw Body 时请求区改为百分比，给编辑器留空间。
 */
const splitPanes = computed<RsSplitPaneItem[]>(() => {
  if (requestPaneIsEditor.value) {
    return [
      { key: 'request', size: 54, min: 28 },
      { key: 'response', size: 46, min: 22 },
    ]
  }
  return [
    { key: 'request', size: 'auto', max: 72 },
    { key: 'response', min: 28 },
  ]
})

const sending = ref(false)
const abortRef = ref<AbortController | null>(null)

const lastResult = ref<RestRequestResult | null>(null)
const responseBodyDisplay = ref('')
const responseHeadersText = ref('')

const methodModifierClass = computed(() => restfulMethodModifierClass(method.value))

const paramsCount = computed(
  () => queryParams.value.filter((r) => r.enabled && r.key.trim()).length
)

/** URL 地址栏与 Params 表格双向同步时防止互相触发死循环 */
const urlQuerySyncLock = ref<'url' | 'params' | null>(null)

/**
 * 用当前地址栏字符串解析出的 query 覆盖 Params 表（末尾保留一行空编辑行）。
 * 与 {@link searchParamsToQueryRows} 一致；无法解析时置为单行空表。
 */
function replaceQueryParamsFromCurrentUrl(): void {
  const u = tryParseUserUrl(url.value)
  queryParams.value = u ? searchParamsToQueryRows(u.searchParams) : [createKeyValueRow()]
}

watch(
  url,
  () => {
    if (urlQuerySyncLock.value === 'params') {
      return
    }
    const u = tryParseUserUrl(url.value)
    if (!u) {
      return
    }
    const fromUrl = u.searchParams.toString()
    const fromRows = queryRowsToSearchString(queryParams.value)
    if (fromUrl === fromRows) {
      return
    }
    urlQuerySyncLock.value = 'url'
    replaceQueryParamsFromCurrentUrl()
    nextTick(() => {
      urlQuerySyncLock.value = null
    })
  },
  { immediate: true }
)

watch(
  queryParams,
  () => {
    if (urlQuerySyncLock.value === 'url') {
      return
    }
    const u = tryParseUserUrl(url.value)
    if (!u) {
      return
    }
    const nextSearch = queryRowsToSearchString(queryParams.value)
    if (u.searchParams.toString() === nextSearch) {
      return
    }
    const next = new URL(u.href)
    next.search = nextSearch
    const nextHref = next.toString()
    if (nextHref === url.value) {
      return
    }
    urlQuerySyncLock.value = 'params'
    url.value = nextHref
    nextTick(() => {
      urlQuerySyncLock.value = null
    })
  },
  { deep: true }
)

const headersCount = computed(
  () =>
    headerRows.value.filter((r) => r.enabled && r.key.trim() && !r.autoFromBody).length
)

const cookiesCount = computed(
  () => cookieRows.value.filter((r) => r.enabled && r.key.trim()).length
)

const requestTabItems = computed(() => [
  { value: 'params', label: 'Params', badge: paramsCount.value || undefined },
  { value: 'body', label: 'Body' },
  { value: 'headers', label: 'Headers', badge: headersCount.value || undefined },
  { value: 'cookies', label: 'Cookies', badge: cookiesCount.value || undefined },
  { value: 'auth', label: 'Auth' },
])

const responseTabItems = [
  { value: 'body', label: 'Body' },
  { value: 'headers', label: 'Headers' },
]

/**
 * 将 Body 推导的 Content-Type 同步为表格首行，或与手写 Content-Type 互斥。
 *
 * @param trigger - `body`：来自 Body 类型 / Raw 的 Content-Type 变化，不因旧值与推导不一致而降级自动行（避免切换后无法更新）；
 *   `headers`：来自用户编辑 Headers，此时若自动行值与推导不一致则视为用户覆盖并降级。
 */
function syncHeaderContentTypeRows(trigger: 'body' | 'headers'): void {
  const derived = deriveBodyContentType(bodyProcessType.value, rawContentType.value)
  const demoted = demoteAutoContentTypeRows(headerRows.value, derived, {
    demoteOnValueMismatch: trigger === 'headers',
  })
  if (demoted) {
    headerRows.value = demoted
  }
  const desired = computeDesiredHeaderRows(headerRows.value, derived)
  if (desired === null) {
    return
  }
  headerRows.value = desired
}

watch(
  () => [bodyProcessType.value, rawContentType.value],
  () => {
    syncHeaderContentTypeRows('body')
  },
  { immediate: true }
)

watch(
  headerRows,
  () => {
    syncHeaderContentTypeRows('headers')
  },
  { deep: true }
)

/**
 * 合并 Cookies 标签、Auth 与 Body 推导头后用于发送的请求头列表。
 */
function mergeHeadersForSend(): RestKeyValueRow[] {
  const merged = cloneKeyValueRows(headerRows.value)
  const fromCookieTab = cookieStringFromRows(cookieRows.value)
  const existingCookie = findHeaderValueCI(merged, 'cookie')
  let cookieVal = ''
  if (existingCookie && fromCookieTab) {
    cookieVal = `${existingCookie}; ${fromCookieTab}`
  } else {
    cookieVal = existingCookie || fromCookieTab
  }
  if (cookieVal) {
    upsertHeaderRow(merged, 'Cookie', cookieVal)
  }
  if (authType.value === 'bearer' && bearerToken.value.trim()) {
    upsertHeaderRow(merged, 'Authorization', `Bearer ${bearerToken.value.trim()}`)
  }
  if (authType.value === 'basic' && (basicUser.value !== '' || basicPassword.value !== '')) {
    const raw = `${basicUser.value}:${basicPassword.value}`
    const b64 =
      typeof btoa !== 'undefined' ? btoa(raw) : raw
    upsertHeaderRow(merged, 'Authorization', `Basic ${b64}`)
  }
  if (authType.value === 'apikey' && apiKeyIn.value === 'header' && apiKeyValue.value.trim()) {
    const hn = apiKeyHeaderName.value.trim() || 'X-API-Key'
    upsertHeaderRow(merged, hn, apiKeyValue.value.trim())
  }
  return merged
}

/**
 * 合并 API Key（Query）后的 Query 参数列表。
 */
function mergeQueryParamsForSend(): RestKeyValueRow[] {
  const merged = cloneKeyValueRows(queryParams.value)
  if (authType.value === 'apikey' && apiKeyIn.value === 'query' && apiKeyValue.value.trim()) {
    const qn = apiKeyQueryName.value.trim() || 'api_key'
    upsertHeaderRow(merged, qn, apiKeyValue.value.trim())
  }
  return merged
}

const rawBodyLanguage = computed((): RsCodeEditorLanguage =>
  inferRawEditorLanguage(bodyProcessType.value, rawContentType.value)
)

const responseStatusTagType = computed(() => {
  if (!lastResult.value) {
    return 'default' as const
  }
  if (lastResult.value.status === 0) {
    return 'error' as const
  }
  return httpStatusTagType(lastResult.value.status)
})

/** 是否为真实 HTTP 响应（用于隐藏「格式化 JSON」等仅适用于代发结果的操作） */
const hasRealHttpResponse = computed(() => {
  const r = lastResult.value
  if (!r || r.status === 0) {
    return false
  }
  return Object.keys(r.responseHeaders || {}).length > 0
})

/**
 * 根据 Content-Type 与正文前缀推断响应编辑器语言，便于高亮。
 */
const responseBodyLanguage = computed((): RsCodeEditorLanguage => {
  if (!lastResult.value) {
    return 'plaintext'
  }
  const ct = (lastResult.value.responseHeaders['content-type'] || '').toLowerCase()
  if (ct.includes('json')) {
    return 'json'
  }
  if (ct.includes('html')) {
    return 'html'
  }
  if (ct.includes('xml')) {
    return 'xml'
  }
  const t = responseBodyDisplay.value.trim()
  if (t.startsWith('{') || t.startsWith('[')) {
    return 'json'
  }
  return 'plaintext'
})

/**
 * 切换 Body 处理类型，并在首次进入某类 Raw 时填入占位内容。
 */
function setBodyProcessType(t: RestBodyProcessType): void {
  bodyProcessType.value = t
  const rb = rawBody.value.trim()
  if (t === 'json' && !rb) {
    rawBody.value = '{\n  \n}'
  }
  if (t === 'xml' && !rb) {
    rawBody.value = '<?xml version="1.0" encoding="UTF-8"?>\n'
  }
  if (t === 'graphql' && !rb) {
    rawBody.value = '{\n  "query": ""\n}'
  }
  if (t === 'raw' && rawContentType.value.includes('json')) {
    rawContentType.value = 'text/plain'
  }
}

/**
 * 组装 `sendRestRequest` 参数：将 UI 上的处理类型映射为 `RestBodyMode` 与 Content-Type。
 */
function buildSendPayload(): SendRestRequestInput {
  const t = bodyProcessType.value
  let bodyMode: RestBodyMode = 'none'
  let ct = rawContentType.value
  const rb = rawBody.value

  switch (t) {
    case 'none':
      bodyMode = 'none'
      break
    case 'x-www-form-urlencoded':
      bodyMode = 'urlencoded'
      break
    case 'form-data':
      bodyMode = 'multipart'
      break
    case 'json':
      bodyMode = 'raw'
      ct = 'application/json'
      break
    case 'xml':
      bodyMode = 'raw'
      ct = 'application/xml'
      break
    case 'raw':
      bodyMode = 'raw'
      break
    case 'binary':
      bodyMode = 'raw'
      ct = 'application/octet-stream'
      break
    case 'graphql':
      bodyMode = 'raw'
      ct = 'application/json'
      break
    case 'msgpack':
      bodyMode = 'raw'
      ct = 'application/msgpack'
      break
    default:
      bodyMode = 'none'
  }

  return {
    method: method.value,
    url: url.value,
    queryParams: mergeQueryParamsForSend(),
    headers: mergeHeadersForSend(),
    bodyMode,
    rawBody: rb,
    rawContentType: ct,
    formFields: formFields.value
  }
}

// urlencoded 仅支持文本字段，从 form-data 切回时去掉文件列
watch(bodyProcessType, (t) => {
  if (t === 'x-www-form-urlencoded') {
    for (const r of formFields.value) {
      r.fieldKind = 'text'
      r.file = null
    }
  }
})

/**
 * 根据可选 Props 填充 Headers 与 Body（首次挂载与 {@link rehydrateFromInitialProps} 时调用）。
 */
function applyInitialRequestExtras(): void {
  const hj = props.initialHeadersJson?.trim()
  if (hj) {
    try {
      const parsed = JSON.parse(hj) as unknown
      if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
        const rows = Object.entries(parsed as Record<string, unknown>).map(([k, v]) => {
          const r = createKeyValueRow()
          r.key = k
          r.value = v == null ? '' : String(v)
          r.enabled = true
          return r
        })
        headerRows.value = rows.length > 0 ? [...rows, createKeyValueRow()] : [createKeyValueRow()]
      }
    } catch {
      // 非法 JSON 时保留默认空表头行
    }
  }

  const raw = props.initialRawBody
  const rawStr = raw == null ? '' : String(raw)
  if (rawStr.length > 0) {
    const pt = props.initialBodyProcessType ?? 'raw'
    bodyProcessType.value = pt
    rawBody.value = rawStr
  }
}

/**
 * 挂载时预填响应控制台（如父级加载详情失败），与代发成功后的展示共用同一区域。
 */
function applyInitialResponseConsole(): void {
  const body = props.initialResponseBody?.trim() ?? ''
  const statusText = props.initialResponseStatusText?.trim() ?? ''
  if (!body && !statusText) {
    return
  }
  responseBodyDisplay.value = body || statusText
  responseTab.value = 'body'
  lastResult.value = {
    status: 0,
    statusText: statusText || '异常',
    durationMs: 0,
    responseHeaders: {},
    responseBodyText: responseBodyDisplay.value
  }
  responseHeadersText.value = ''
}

/**
 * 按当前 props 重置 URL/方法/Headers/Body/响应区等（父级切换快照时调用，避免整组件销毁重建）。
 * 与首次挂载时的填充逻辑一致，并中止进行中的发送。
 */
function rehydrateFromInitialProps(): void {
  abortRef.value?.abort()
  abortRef.value = null
  sending.value = false
  urlQuerySyncLock.value = null

  /**
   * Params 与地址栏双向同步：加锁 `url` 避免清空 Params 时 query 的 watcher 用旧 URL 回写地址栏。
   * 仅依赖 watch(url) 不够：其与表格序列化比较可能在边界情况下提前 return，重载快照后须显式从 URL 灌入 Params。
   */
  urlQuerySyncLock.value = 'url'
  method.value = props.initialMethod
  url.value = props.initialUrl ?? ''
  replaceQueryParamsFromCurrentUrl()

  headerRows.value = [createKeyValueRow()]
  cookieRows.value = [createKeyValueRow()]
  formFields.value = [createKeyValueRow()]

  authType.value = 'none'
  bearerToken.value = ''
  basicUser.value = ''
  basicPassword.value = ''
  apiKeyIn.value = 'header'
  apiKeyHeaderName.value = 'X-API-Key'
  apiKeyQueryName.value = 'api_key'
  apiKeyValue.value = ''

  bodyProcessType.value = 'none'
  rawBody.value = ''
  rawContentType.value = 'application/json'

  requestTab.value = 'params'
  responseTab.value = 'body'

  lastResult.value = null
  responseBodyDisplay.value = ''
  responseHeadersText.value = ''

  applyInitialRequestExtras()
  applyInitialResponseConsole()

  void nextTick(() => {
    // url 的 watcher 在无法解析 URL 或 search 已与表一致时会提前 return，未必再排 nextTick 解锁
    if (urlQuerySyncLock.value === 'url') {
      urlQuerySyncLock.value = null
    }
    syncHeaderContentTypeRows('body')
    // KeyValueEditor 对 rows 的 nextTick 整理可能与本帧重排交错，再对齐一次 Params 与地址栏
    urlQuerySyncLock.value = 'url'
    replaceQueryParamsFromCurrentUrl()
    urlQuerySyncLock.value = null
  })
}

onMounted(() => {
  rehydrateFromInitialProps()
})

onBeforeUnmount(() => {
  abortRef.value?.abort()
})

/**
 * 尝试将响应体格式化为缩进 JSON；失败则保持原文并提示。
 */
function formatResponseBody(): void {
  const raw = responseBodyDisplay.value
  try {
    const parsed = JSON.parse(raw) as unknown
    responseBodyDisplay.value = JSON.stringify(parsed, null, 2)
  } catch (e) {
    message.warning('当前内容不是合法 JSON，无法格式化')
    console.warn('formatResponseBody', e)
  }
}

/**
 * 发起请求：支持取消、错误提示与事件回调。
 */
async function handleSend(): Promise<void> {
  if (!canExecuteHttp.value) {
    message.warning('没有执行此操作的权限')
    return
  }
  abortRef.value?.abort()
  const controller = new AbortController()
  abortRef.value = controller

  sending.value = true
  lastResult.value = null
  responseBodyDisplay.value = ''
  responseHeadersText.value = ''

  emit('send-start')

  try {
    const out = await sendRestRequest({
      ...buildSendPayload(),
      signal: controller.signal
    })

    if (!out.ok) {
      const causeStr =
        out.cause != null
          ? out.cause instanceof Error
            ? out.cause.stack || out.cause.message
            : String(out.cause)
          : ''
      responseBodyDisplay.value = causeStr ? `${out.message}\n\n${causeStr}` : out.message
      responseHeadersText.value = ''
      lastResult.value = {
        status: 0,
        statusText: '请求失败',
        durationMs: 0,
        responseHeaders: {},
        responseBodyText: responseBodyDisplay.value
      }
      emit('error', out.message, out.cause)
      emit('complete', undefined)
      return
    }

    lastResult.value = out.data
    responseBodyDisplay.value = out.data.responseBodyText
    responseHeadersText.value = stringifyResponseHeaders(out.data.responseHeaders)
    emit('success', out.data)
    emit('complete', out.data)
  } finally {
    sending.value = false
    abortRef.value = null
    emit('send-end')
  }
}

/**
 * 中止进行中的代发请求（AbortController 会经 Axios 取消对 `/gateway/hubplugin/http/execute` 的调用）。
 */
function handleCancel(): void {
  abortRef.value?.abort()
}

defineExpose({
  send: handleSend,
  cancel: handleCancel,
  rehydrateFromInitialProps,
})
</script>

<style scoped lang="scss">
.g-restful-api {
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 100%;
  min-width: 0;
  min-height: 480px;
  overflow: hidden;
  box-sizing: border-box;
  background: var(--rs-surface, var(--g-dialog-bg));

  &__urlbar {
    display: flex;
    flex-wrap: nowrap;
    align-items: center;
    gap: 8px;
    flex-shrink: 0;
    padding: 10px 12px 8px;
  }

  &__url-group {
    display: flex;
    flex: 1 1 auto;
    align-items: center;
    min-width: 0;
    gap: 0;
  }

  &__method {
    width: 108px;
    flex-shrink: 0;
  }

  &__method--get {
    color: #61affe;
  }
  &__method--post {
    color: #49cc90;
  }
  &__method--put {
    color: #fca130;
  }
  &__method--patch {
    color: #50e3c2;
  }
  &__method--delete {
    color: #f93e3e;
  }
  &__method--head {
    color: #9012fe;
  }
  &__method--options {
    color: #0d47a1;
  }

  &__url {
    flex: 1 1 auto;
    min-width: 0;
  }

  &__send-group {
    display: flex;
    flex-shrink: 0;
    align-items: center;
    gap: 6px;
  }

  &__split {
    flex: 1 1 auto;
    min-width: 0;
    min-height: 0;
  }

  &__request,
  &__response {
    display: flex;
    flex-direction: column;
    width: 100%;
    min-width: 0;
    min-height: 0;
  }

  &__request {
    height: auto;
  }

  &__request--fill,
  &__response {
    height: 100%;
  }

  &__tabs {
    flex-shrink: 0;
    width: 100%;
    padding: 0 8px;
  }

  &__pane {
    flex: 0 0 auto;
    min-width: 0;
    overflow: visible;
    padding: 8px 12px 12px;

    &--editor {
      flex: 1 1 auto;
      min-height: 0;
      overflow: hidden;
      display: flex;
      flex-direction: column;
    }
  }

  &__pane-inner {
    min-width: 0;
    min-height: 0;
  }

  &__body-pane {
    display: flex;
    flex-direction: column;
    height: 100%;
    min-height: 0;
    gap: 8px;
  }

  &__auth-pane {
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding-top: 4px;
  }

  &__auth-type-group {
    display: flex;
    flex-wrap: wrap;
    gap: 8px 16px;
  }

  &__auth-fields {
    display: flex;
    flex-direction: column;
    gap: 8px;
    max-width: 420px;

    &--row {
      flex-direction: row;
      flex-wrap: wrap;
      align-items: center;
      gap: 8px;
    }
  }

  &__body-radio-group {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px 16px;
    flex-shrink: 0;
    width: 100%;
  }

  &__raw-ct-row {
    flex-shrink: 0;
    max-width: 360px;
  }

  &__empty {
    display: flex;
    flex: 1;
    align-items: center;
    justify-content: center;
    min-height: 120px;
    font-size: 13px;
    color: var(--rs-muted, var(--g-text-secondary));
  }

  &__body-raw {
    display: flex;
    flex: 1 1 auto;
    flex-direction: column;
    min-height: 0;
    gap: 8px;
  }

  &__content-type {
    min-width: 200px;
    max-width: 320px;
  }

  &__body-form {
    min-width: 0;
  }

  &__meta {
    display: flex;
    align-items: center;
    gap: 10px;
    padding-right: 4px;
  }

  &__status {
    font-size: 13px;
    font-weight: 600;
    font-variant-numeric: tabular-nums;

    &--success {
      color: #49cc90;
    }
    &--warning {
      color: #fca130;
    }
    &--error {
      color: #f93e3e;
    }
    &--default {
      color: var(--rs-muted, var(--g-text-secondary));
    }
  }

  &__time {
    font-size: 12px;
    font-variant-numeric: tabular-nums;
    color: var(--rs-muted, var(--g-text-secondary));
  }
}
</style>
