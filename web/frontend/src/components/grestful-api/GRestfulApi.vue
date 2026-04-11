<template>
  <n-card
    class="g-restful-api"
    :class="props.class"
    size="small"
    :bordered="true"
    :segmented="{ content: true, footer: 'soft' }"
  >
    <div class="g-restful-api__toolbar">
      <n-select
        v-model:value="method"
        class="g-restful-api__method"
        :class="methodModifierClass"
        :options="METHOD_OPTIONS"
        :consistent-menu-width="false"
        size="small"
      />
      <n-input
        v-model:value="url"
        class="g-restful-api__url"
        type="text"
        placeholder="请输入请求 URL（支持相对路径，将相对当前站点）"
        clearable
        size="small"
        @keyup.enter="handleSend"
      />
      <n-button
        type="primary"
        size="small"
        :loading="sending"
        :disabled="sending"
        @click="handleSend"
      >
        发送
      </n-button>
      <n-button
        v-if="sending"
        size="small"
        quaternary
        @click="handleCancel"
      >
        取消
      </n-button>
    </div>

    <n-tabs
      v-model:value="requestTab"
      type="line"
      size="medium"
      class="g-restful-api__req-tabs"
    >
      <n-tab-pane name="params">
        <template #tab>
          <span class="g-restful-api__tab-label">
            Params
            <n-badge
              v-if="paramsCount > 0"
              :value="paramsCount"
              :max="99"
              type="success"
              class="g-restful-api__tab-badge"
            />
          </span>
        </template>
        <div class="g-restful-api__pane-inner">
          <key-value-editor
            v-model:rows="queryParams"
            variant="table"
            table-variant="query"
          />
        </div>
      </n-tab-pane>

      <n-tab-pane name="body">
        <template #tab>
          <span class="g-restful-api__tab-label">Body</span>
        </template>
        <div class="g-restful-api__pane-inner">
          <n-radio-group
            class="g-restful-api__body-radio-group"
            :value="bodyProcessType"
            name="rest-body-process"
            size="small"
            @update:value="setBodyProcessType"
          >
            <n-radio
              v-for="opt in BODY_PROCESS_OPTIONS"
              :key="opt.value"
              :value="opt.value"
            >
              {{ opt.label }}
            </n-radio>
          </n-radio-group>

          <div
            v-if="bodyProcessType === 'none'"
            class="g-restful-api__body-empty"
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
              <n-select
                v-model:value="rawContentType"
                class="g-restful-api__content-type"
                :options="RAW_CONTENT_TYPE_OPTIONS"
                size="small"
              />
            </div>
            <g-code-mirror
              v-model="rawBody"
              :language="rawBodyLanguage"
              :min-height="requestBodyMinHeightCss"
              line-wrapping
            />
          </div>
        </div>
      </n-tab-pane>

      <n-tab-pane name="headers">
        <template #tab>
          <span class="g-restful-api__tab-label">
            Headers
            <n-badge
              v-if="headersCount > 0"
              :value="headersCount"
              :max="99"
              type="success"
              class="g-restful-api__tab-badge"
            />
          </span>
        </template>
        <div class="g-restful-api__pane-inner">
          <key-value-editor
            v-model:rows="headerRows"
            variant="table"
            table-variant="query"
            key-column-label="名称"
            value-column-label="值"
            show-auto-body-hint
          />
        </div>
      </n-tab-pane>

      <n-tab-pane name="cookies">
        <template #tab>
          <span class="g-restful-api__tab-label">
            Cookies
            <n-badge
              v-if="cookiesCount > 0"
              :value="cookiesCount"
              :max="99"
              type="success"
              class="g-restful-api__tab-badge"
            />
          </span>
        </template>
        <div class="g-restful-api__pane-inner">
          <key-value-editor
            v-model:rows="cookieRows"
            variant="table"
            table-variant="query"
            key-column-label="名称"
            value-column-label="值"
          />
        </div>
      </n-tab-pane>

      <n-tab-pane name="auth">
        <template #tab>
          <span class="g-restful-api__tab-label">Auth</span>
        </template>
        <div class="g-restful-api__pane-inner g-restful-api__auth-pane">
          <n-radio-group
            v-model:value="authType"
            class="g-restful-api__auth-type-group"
            size="small"
            name="rest-auth-type"
          >
            <n-radio value="none">无</n-radio>
            <n-radio value="bearer">Bearer Token</n-radio>
            <n-radio value="basic">Basic Auth</n-radio>
            <n-radio value="apikey">API Key</n-radio>
          </n-radio-group>

          <div
            v-if="authType === 'bearer'"
            class="g-restful-api__auth-fields"
          >
            <n-input
              v-model:value="bearerToken"
              type="password"
              show-password-on="click"
              placeholder="Token"
              size="small"
            />
          </div>

          <div
            v-else-if="authType === 'basic'"
            class="g-restful-api__auth-fields g-restful-api__auth-fields--row"
          >
            <n-input
              v-model:value="basicUser"
              placeholder="用户名"
              size="small"
            />
            <n-input
              v-model:value="basicPassword"
              type="password"
              show-password-on="click"
              placeholder="密码"
              size="small"
            />
          </div>

          <div
            v-else-if="authType === 'apikey'"
            class="g-restful-api__auth-fields"
          >
            <n-radio-group
              v-model:value="apiKeyIn"
              size="small"
              name="rest-apikey-in"
            >
              <n-radio value="header">Header</n-radio>
              <n-radio value="query">Query</n-radio>
            </n-radio-group>
            <n-input
              v-if="apiKeyIn === 'header'"
              v-model:value="apiKeyHeaderName"
              placeholder="Header 名称，默认 X-API-Key"
              size="small"
            />
            <n-input
              v-else
              v-model:value="apiKeyQueryName"
              placeholder="Query 参数名，默认 api_key"
              size="small"
            />
            <n-input
              v-model:value="apiKeyValue"
              type="password"
              show-password-on="click"
              placeholder="密钥"
              size="small"
            />
          </div>
        </div>
      </n-tab-pane>
    </n-tabs>

    <template #footer>
      <div class="g-restful-api__response-wrap">
        <div class="g-restful-api__response-head">
          <span class="g-restful-api__response-title">响应</span>
          <n-tag v-if="lastResult" size="small" :type="responseStatusTagType" round>
            {{ lastResult.statusText?.trim() || String(lastResult.status) }}
          </n-tag>
          <n-tag v-if="lastResult && lastResult.durationMs > 0" size="small" type="info" round>
            {{ lastResult.durationMs.toFixed(0) }} ms
          </n-tag>
          <n-button
            v-if="responseTab === 'body' && lastResult && hasRealHttpResponse"
            size="tiny"
            quaternary
            @click="formatResponseBody"
          >
            格式化 JSON
          </n-button>
        </div>
        <n-tabs
          v-model:value="responseTab"
          type="line"
          size="small"
          class="g-restful-api__res-tabs"
        >
          <n-tab-pane name="body" tab="Body" />
          <n-tab-pane name="headers" tab="Headers" />
        </n-tabs>
        <div class="g-restful-api__response-body">
          <g-code-mirror
            v-if="responseTab === 'body'"
            v-model="responseBodyDisplay"
            :language="responseBodyLanguage"
            :readonly="true"
            :min-height="responseMinHeightCss"
            line-wrapping
          />
          <g-code-mirror
            v-else
            v-model="responseHeadersText"
            :language="'plaintext'"
            :readonly="true"
            :min-height="responseMinHeightCss"
            line-wrapping
          />
        </div>
      </div>
    </template>
  </n-card>
</template>

<script setup lang="ts">
import { GCodeMirror } from '@/components/gcodemirror'
import type { CodeMirrorLanguage } from '@/components/gcodemirror/types'
import {
  NBadge,
  NButton,
  NCard,
  NInput,
  NRadio,
  NRadioGroup,
  NSelect,
  NTabPane,
  NTabs,
  NTag,
  useMessage
} from 'naive-ui'
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

const message = useMessage()

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

const responseMinHeightCss = computed(() =>
  typeof props.responseMinHeight === 'number' ? `${props.responseMinHeight}px` : props.responseMinHeight
)

const requestBodyMinHeightCss = computed(() =>
  typeof props.requestBodyMinHeight === 'number' ? `${props.requestBodyMinHeight}px` : props.requestBodyMinHeight
)

const rawBodyLanguage = computed((): CodeMirrorLanguage =>
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
const responseBodyLanguage = computed((): CodeMirrorLanguage => {
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
  &__toolbar {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px;
    margin-bottom: 12px;
  }

  &__method {
    width: 112px;
    flex-shrink: 0;
  }

  &__method--get :deep(.n-base-selection) {
    box-shadow: inset 3px 0 0 #61affe;
  }
  &__method--post :deep(.n-base-selection) {
    box-shadow: inset 3px 0 0 #49cc90;
  }
  &__method--put :deep(.n-base-selection) {
    box-shadow: inset 3px 0 0 #fca130;
  }
  &__method--patch :deep(.n-base-selection) {
    box-shadow: inset 3px 0 0 #50e3c2;
  }
  &__method--delete :deep(.n-base-selection) {
    box-shadow: inset 3px 0 0 #f93e3e;
  }
  &__method--head :deep(.n-base-selection) {
    box-shadow: inset 3px 0 0 #9012fe;
  }
  &__method--options :deep(.n-base-selection) {
    box-shadow: inset 3px 0 0 #0d47a1;
  }

  &__url {
    flex: 1 1 200px;
    min-width: 0;
  }

  &__req-tabs {
    margin-top: 4px;

    :deep(.n-tabs-nav) {
      padding: 0 2px;
    }

    :deep(.n-tabs-tab) {
      font-weight: 500;
      padding: 10px 14px 8px;
    }

    :deep(.n-tabs-tab--active) {
      color: var(--n-primary-color);
      font-weight: 600;
    }

    :deep(.n-tabs-bar) {
      background-color: var(--n-primary-color);
      height: 2px;
      border-radius: 1px;
    }
  }

  &__tab-label {
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }

  &__tab-badge {
    :deep(.n-badge-sup) {
      font-size: 11px;
      min-width: 18px;
      height: 18px;
      line-height: 18px;
      padding: 0 5px;
    }
  }

  &__pane-inner {
    padding-top: 4px;
  }

  &__auth-pane {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  &__auth-type-group {
    &.n-radio-group {
      display: flex;
      flex-wrap: wrap;
      gap: 8px 16px;
    }
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

      :deep(.n-input) {
        flex: 1 1 160px;
        min-width: 0;
      }
    }
  }

  &__body-radio-group {
    width: 100%;
    margin-bottom: 12px;

    &.n-radio-group {
      display: flex;
      flex-wrap: wrap;
      align-items: flex-start;
      gap: 8px 20px;
      row-gap: 10px;
    }

    :deep(.n-radio) {
      align-items: center;
    }

    :deep(.n-radio__label) {
      font-size: 13px;
    }
  }

  &__raw-ct-row {
    margin-bottom: 10px;
    max-width: 360px;
  }

  &__body-empty {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 120px;
    padding: 24px 16px;
    font-size: 13px;
    color: var(--n-text-color-3);
    border: 1px solid var(--n-border-color);
    border-radius: var(--n-border-radius);
    background: var(--n-color-modal);
  }

  &__body-raw {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  &__content-type {
    min-width: 200px;
    max-width: 320px;
  }

  &__body-form {
    margin-top: 4px;
  }

  &__response-wrap {
    padding-top: 4px;
  }

  &__response-head {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px;
    margin-bottom: 10px;
  }

  &__response-title {
    font-weight: 600;
    font-size: 14px;
    margin-right: 4px;
  }

  &__res-tabs {
    margin-bottom: 8px;

    :deep(.n-tabs-tab--active) {
      color: var(--n-primary-color);
      font-weight: 600;
    }

    :deep(.n-tabs-bar) {
      background-color: var(--n-primary-color);
    }
  }

  &__response-body {
    border: 1px solid var(--n-border-color);
    border-radius: var(--n-border-radius);
    overflow: hidden;
    background: var(--n-color);
  }
}
</style>
