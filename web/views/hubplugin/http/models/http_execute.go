package models

import "encoding/json"

// HttpExecuteAuth 可选认证信息；与 Headers 合并时，Headers 中已存在的键（忽略大小写）优先生效。
type HttpExecuteAuth struct {
	// BearerToken 非空且尚无 Authorization 头时，设置为 Bearer
	BearerToken string `json:"bearerToken,omitempty"`
	// BasicUser、BasicPassword 用于 Basic 认证（与 Bearer 二选一，优先 BearerToken）
	BasicUser     string `json:"basicUser,omitempty"`
	BasicPassword string `json:"basicPassword,omitempty"`
	// ApiKeyHeaderName API Key 请求头名，默认 X-API-Key（仅当未使用 ApiKeyQueryName 时）
	ApiKeyHeaderName string `json:"apiKeyHeaderName,omitempty"`
	// ApiKeyValue API Key 值
	ApiKeyValue string `json:"apiKeyValue,omitempty"`
	// ApiKeyQueryName 非空时，将 ApiKeyValue 写入 URL 查询参数（键名），不再写入请求头
	ApiKeyQueryName string `json:"apiKeyQueryName,omitempty"`
}

// HttpExecuteFormDataItem 描述 multipart/form-data 中的一条字段；通过 Type 区分文本与文件。
type HttpExecuteFormDataItem struct {
	// Key 表单字段名
	Key string `json:"key"`
	// Type 取值 text（普通字段）或 file（文件字段，需 FileName、ContentBase64）
	Type string `json:"type"`
	// Value type 为 text 时的字段值
	Value string `json:"value,omitempty"`
	// FileName type 为 file 时的原始文件名
	FileName string `json:"fileName,omitempty"`
	// ContentBase64 type 为 file 时文件内容的标准 Base64（不含 data: 前缀）
	ContentBase64 string `json:"contentBase64,omitempty"`
}

// HttpExecuteRequest 服务端代发 HTTP 请求的入参（由前端传入，经网关使用 httpclient 发起）。
// 请使用 Content-Type: application/json 提交，否则表单绑定无法填充本结构体字段。
//
// 正文优先级（互斥）：FormUrlEncoded > FormData > Body。
type HttpExecuteRequest struct {
	// Method HTTP 方法，如 GET、POST（大小写不敏感）
	Method string `json:"method"`
	// URL 完整请求地址，须为 http 或 https
	URL string `json:"url"`
	// Headers 可选请求头
	Headers map[string]string `json:"headers,omitempty"`
	// Auth 可选认证；与 Headers 合并，Headers 优先
	Auth *HttpExecuteAuth `json:"auth,omitempty"`
	// Body 下游请求体。JSON 中为字符串时按解码后的字节发送；为 JSON 对象/数组时按原始 JSON 字节发送。
	// 与 FormUrlEncoded、FormData 互斥（前两者优先时忽略 Body）
	Body json.RawMessage `json:"body,omitempty"`
	// FormUrlEncoded 非空时按 application/x-www-form-urlencoded 编码发送；键为字段名，值为未再编码的表单值（服务端用 url.Values 编码）。
	// 不支持同键多值，若需多值请改用 Body 自行拼串或使用其它方式。
	FormUrlEncoded map[string]string `json:"formUrlEncoded,omitempty"`
	// FormData 非空时按 multipart/form-data 发送，顺序与切片一致；每项含 key、type（text|file）及对应取值。
	FormData []HttpExecuteFormDataItem `json:"formData,omitempty"`
	// TimeoutSeconds 单次请求超时（秒），0 表示使用客户端默认（约 30s），最大 120
	TimeoutSeconds int `json:"timeoutSeconds,omitempty"`
	// FollowRedirects 是否跟随重定向，默认 false，与 httpclient 默认一致
	FollowRedirects *bool `json:"followRedirects,omitempty"`
}

// HttpExecuteResult 代发请求的下游响应摘要，供前端展示。
type HttpExecuteResult struct {
	StatusCode int    `json:"statusCode"`
	Status     string `json:"status"`
	// Headers 响应头（多值头以逗号拼接）
	Headers map[string]string `json:"headers"`
	// Body 下游 HTTP 响应报文：合法 UTF-8 时与下游字节序列一致的 Unicode 文本（可为 JSON、XML、纯文本、HTML 等任意格式，不做格式假设）；
	// 非 UTF-8 时为标准 Base64，且 BodyBase64 为 true。
	Body string `json:"body"`
	// BodyBase64 为 true 表示 Body 为 Base64 编码的原始响应体字节。
	BodyBase64 bool `json:"bodyBase64"`
	// DurationMs 从发起请求到读完响应体的大致耗时（毫秒）
	DurationMs int64 `json:"durationMs"`
}
