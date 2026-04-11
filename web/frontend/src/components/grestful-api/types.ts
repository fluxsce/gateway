/**
 * GRestfulApi 组件及相关工具的类型定义。
 *
 * @remarks
 * 用于在页面内组装 HTTP 请求并展示响应；实际出站请求由服务端代发（见 `sendRestRequest`）。
 * 类型与 `sendRestRequest`、`GRestfulApi.vue` 及网关 `hubplugin/http` 接口对齐。
 */

/**
 * 支持的 HTTP 方法（与 fetch 及常见 REST 用法一致）。
 */
export type RestHttpMethod = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE' | 'HEAD' | 'OPTIONS'

/**
 * Auth 标签页认证方式（Bearer / Basic / API Key）。
 */
export type RestAuthType = 'none' | 'bearer' | 'basic' | 'apikey'

/**
 * API Key 附加在请求头或 Query 参数。
 */
export type RestApiKeyLocation = 'header' | 'query'

/**
 * 表单字段行中「类型」列：文本或文件（仅 form-data 发送 multipart 时使用文件）。
 */
export type RestFormFieldKind = 'text' | 'file'

/**
 * 键值表格中的一行，可用于 Query、Header、Cookie 或表单字段。
 * `enabled` 为 false 时发送请求时忽略该行。
 */
export interface RestKeyValueRow {
  /** 稳定键，用于列表渲染与增删行 */
  id: string
  /** 是否参与请求 */
  enabled: boolean
  /** 名称（参数名或 Header 名） */
  key: string
  /** 值（fieldKind 为 file 时可忽略，以 file 为准） */
  value: string
  /**
   * 表单字段类型：x-www-form-urlencoded 仅 text；
   * form-data 可为 text 或 file。
   */
  fieldKind?: RestFormFieldKind
  /** fieldKind 为 file 时选中的文件 */
  file?: File | null
  /**
   * 为 true 时表示由 Body 类型推导自动插入的 Content-Type 行，
   * 随 Body 设置更新；用户修改键名、取值或取消勾选后变为 false；
   * 用户手动填写同名 Header 时自动移除自动行。
   */
  autoFromBody?: boolean
}

/**
 * 创建一条空的键值行，用于表格初始化或底部自动追加的空行。
 *
 * @returns 新的 {@link RestKeyValueRow}，`id` 由 `crypto.randomUUID()` 生成。
 */
export function createKeyValueRow(): RestKeyValueRow {
  return {
    id: crypto.randomUUID(),
    enabled: true,
    key: '',
    value: '',
    fieldKind: 'text',
    file: null
  }
}

/**
 * 实际发送时使用的正文模式（与 `fetch` 能力对齐）。
 */
export type RestBodyMode = 'none' | 'raw' | 'urlencoded' | 'multipart'

/**
 * Body 处理类型（UI 选项，与 Postman 等工具一致），映射为 {@link RestBodyMode} 与 Content-Type。
 */
export type RestBodyProcessType =
  | 'none'
  | 'form-data'
  | 'x-www-form-urlencoded'
  | 'json'
  | 'xml'
  | 'raw'
  | 'binary'
  | 'graphql'
  | 'msgpack'

/**
 * 发送 REST 请求时的完整参数，由 UI 或代码组装后传入 `sendRestRequest`。
 */
export interface SendRestRequestInput {
  /** HTTP 方法 */
  method: RestHttpMethod
  /** 地址栏中的 URL（可含已有 query） */
  url: string
  /** 附加 Query 参数，会合并进最终 URL */
  queryParams: RestKeyValueRow[]
  /** 请求头 */
  headers: RestKeyValueRow[]
  /** 正文模式 */
  bodyMode: RestBodyMode
  /** `bodyMode === 'raw'` 时的正文内容 */
  rawBody: string
  /** `bodyMode === 'raw'` 时的 Content-Type（若 Header 中已指定 Content-Type 则 Header 优先） */
  rawContentType: string
  /** `bodyMode === 'urlencoded'` 时的表单行 */
  formFields: RestKeyValueRow[]
  /** 可选 AbortSignal，用于取消请求 */
  signal?: AbortSignal
}

/**
 * `sendRestRequest` 成功返回时的结果（已拿到响应，无论 HTTP 状态码是否为 2xx）。
 */
export interface RestRequestResult {
  /** HTTP 状态码 */
  status: number
  /** 状态短语 */
  statusText: string
  /** 从发起请求到读完响应体的大致耗时（毫秒） */
  durationMs: number
  /** 响应头（同名头取最后一次出现的值，与常见调试工具展示一致） */
  responseHeaders: Record<string, string>
  /** 响应体文本（按字节解码为 UTF-8 字符串） */
  responseBodyText: string
}

/**
 * `sendRestRequest` 在网络错误或 URL 无效等情况下的错误信息。
 */
export interface RestRequestFailure {
  /** 是否表示失败（恒为 true，便于与成功结果区分） */
  ok: false
  /** 面向用户的错误说明 */
  message: string
  /** 底层原因，便于日志 */
  cause?: unknown
}

/**
 * `sendRestRequest` 的联合返回类型（判别字段 `ok`）。
 *
 * @remarks
 * - `ok: true` 时表示已收到下游 HTTP 响应（含非 2xx），`data` 为 {@link RestRequestResult}。
 * - `ok: false` 时表示 URL 非法、代发失败、取消或业务错误。
 */
export type SendRestRequestOutput = { ok: true; data: RestRequestResult } | RestRequestFailure

/**
 * 网关 `POST /gateway/hubplugin/http/execute` 成功时 `JsonData.bizData` 反序列化后的结构，
 * 与后端 `web/views/hubplugin/http/models.HttpExecuteResult` 字段一致。
 */
export interface HttpExecuteResultDTO {
  /** 下游 HTTP 状态码。 */
  statusCode: number
  /** 状态行文本（如 `200 OK`）。 */
  status: string
  /** 响应头键值；多值头已合并为单字符串。 */
  headers: Record<string, string>
  /**
   * 下游响应报文（UTF-8 时为原文，可为 JSON、XML、纯文本等）；
   * {@link HttpExecuteResultDTO.bodyBase64} 为 true 时为 Base64 字符串。
   */
  body: string
  /**
   * 为 true 表示 `body` 为 Base64 字符串，对应下游响应体非合法 UTF-8 时的回传方式。
   */
  bodyBase64: boolean
  /** 服务端测量：从发起到读完响应体的时间（毫秒）。 */
  durationMs: number
}

/**
 * GRestfulApi 组件 Props。
 */
export interface GRestfulApiProps {
  /**
   * 初始 URL（不含或含 query 均可）。
   */
  initialUrl?: string
  /**
   * 初始 HTTP 方法。
   * @default 'GET'
   */
  initialMethod?: RestHttpMethod
  /**
   * 可选：JSON 对象字符串，键为请求头名称，值为头内容（用于从网关日志等场景还原 Headers）。
   */
  initialHeadersJson?: string
  /**
   * 可选：与 {@link initialBodyProcessType} 配合，填充 Body 编辑器原始正文。
   */
  initialRawBody?: string
  /**
   * 可选：初始 Body 处理类型（如从日志还原 json/raw）。
   */
  initialBodyProcessType?: RestBodyProcessType
  /**
   * 挂载时预填响应区 Body（如父组件加载上下文失败时的说明）。
   */
  initialResponseBody?: string
  /**
   * 与 {@link initialResponseBody} 配合时在响应区标题旁展示的状态标签（如「详情加载失败」），并标记为错误样式。
   */
  initialResponseStatusText?: string
  /**
   * 响应区正文编辑器最小高度。
   * @default '220px'
   */
  responseMinHeight?: string | number
  /**
   * 请求区原始正文编辑器最小高度。
   * @default '160px'
   */
  requestBodyMinHeight?: string | number
  /**
   * 根节点 class。
   */
  class?: string
}

/**
 * GRestfulApi 组件事件签名（与 `defineEmits` 配合使用）。
 */
export interface GRestfulApiEmits {
  /**
   * 即将发起代发请求（在 await 网络前触发，便于父组件标记「发送中」）。
   *
   * @param e - 事件名 `send-start`
   */
  (e: 'send-start'): void
  /**
   * 代发请求流程结束（成功、失败或异常后触发一次，在 success/error/complete 之后）。
   *
   * @param e - 事件名 `send-end`
   */
  (e: 'send-end'): void
  /**
   * 请求成功结束（已收到响应并完成读取）。
   *
   * @param e - 事件名 `success`
   * @param payload - {@link RestRequestResult}
   */
  (e: 'success', payload: RestRequestResult): void
  /**
   * 请求失败（参数错误、代发失败、取消等）。
   *
   * @param e - 事件名 `error`
   * @param message - 面向用户的说明
   * @param cause - 可选原始错误对象
   */
  (e: 'error', message: string, cause?: unknown): void
  /**
   * 请求结束（成功或失败均触发一次）。
   *
   * @param e - 事件名 `complete`
   * @param payload - 成功时为 {@link RestRequestResult}，失败时为 `undefined`
   */
  (e: 'complete', payload: RestRequestResult | undefined): void
}
