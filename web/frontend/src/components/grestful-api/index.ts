/**
 * GRestfulApi 包导出：类 Postman 的 REST 调试组件与网关代发请求工具。
 *
 * @remarks
 * - 组件：`GRestfulApi`
 * - 发送：`sendRestRequest`（服务端 `hubplugin/http/execute`）
 * - 类型：见 `types.ts`
 */

export { default as GRestfulApi } from './GRestfulApi.vue'
export { GATEWAY_HTTP_EXECUTE_URL, sendRestRequest, tryParseUserUrl } from './sendRestRequest'
export * from './types'
