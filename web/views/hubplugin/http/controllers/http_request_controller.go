package controllers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"gateway/pkg/httpclient"
	"gateway/pkg/logger"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hubplugin/http/models"

	"github.com/gin-gonic/gin"
)

// HttpRequestController 服务端代发 HTTP 请求，供前端调试工具调用；底层使用 pkg/httpclient.Client。
type HttpRequestController struct {
	client httpclient.Client
}

// NewHttpRequestController 创建控制器并初始化默认 HTTP 客户端。
func NewHttpRequestController() (*HttpRequestController, error) {
	c, err := httpclient.NewClient(nil)
	if err != nil {
		return nil, err
	}
	return &HttpRequestController{client: c}, nil
}

const (
	maxHttpExecuteTimeoutSec = 120
	maxResponseBodyBytes     = 10 << 20 // 10MiB，防止代发把过大响应载入内存
	maxMultipartDecodedBytes = 5 << 20  // 单文件解码后最大 5MiB
)

// Execute 接收 method/url/headers/body，由后端发起 HTTP 请求并返回下游状态、头与正文。
//
// 取消：使用 ctx.Request.Context() 作为 httpclient 的根 context；客户端断开或中止对本 Handler
// 的请求时，context 会被取消，下游 HTTP 调用会尽快结束（与 net/http 行为一致）。
func (ctl *HttpRequestController) Execute(ctx *gin.Context) {
	var req models.HttpExecuteRequest
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数解析失败: "+err.Error(), constants.ED00006)
		return
	}

	method := strings.TrimSpace(strings.ToUpper(req.Method))
	if method == "" {
		response.ErrorJSON(ctx, "method 不能为空", constants.ED00007)
		return
	}

	rawURL := strings.TrimSpace(req.URL)
	if rawURL == "" {
		response.ErrorJSON(ctx, "url 不能为空", constants.ED00007)
		return
	}

	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		response.ErrorJSON(ctx, "无效的 URL", constants.ED00007)
		return
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		response.ErrorJSON(ctx, "仅支持 http、https 协议", constants.ED00007)
		return
	}

	applyAuthQuery(parsed, req.Auth)

	timeout := time.Duration(0)
	if req.TimeoutSeconds > 0 {
		sec := req.TimeoutSeconds
		if sec > maxHttpExecuteTimeoutSec {
			sec = maxHttpExecuteTimeoutSec
		}
		timeout = time.Duration(sec) * time.Second
	}

	baseCtx := ctx.Request.Context()
	if timeout > 0 {
		var cancel context.CancelFunc
		baseCtx, cancel = context.WithTimeout(baseCtx, timeout)
		defer cancel()
	}

	mergedHeaders := mergeAuthHeaders(req.Headers, req.Auth)
	requestURL := parsed.String()
	opts := buildRequestOptions(req, mergedHeaders)

	hasBody := len(req.Body) > 0
	hasForm := len(req.FormUrlEncoded) > 0
	hasMP := len(req.FormData) > 0

	if hasForm && hasMP {
		response.ErrorJSON(ctx, "formUrlEncoded 与 formData 不能同时使用", constants.ED00007)
		return
	}
	if hasForm && hasBody {
		response.ErrorJSON(ctx, "formUrlEncoded 与 body 不能同时使用", constants.ED00007)
		return
	}
	if hasMP && hasBody {
		response.ErrorJSON(ctx, "formData 与 body 不能同时使用", constants.ED00007)
		return
	}

	var httpResp *httpclient.Response
	var execErr error
	start := time.Now()

	switch {
	case hasForm:
		vals := url.Values{}
		for k, v := range req.FormUrlEncoded {
			k = strings.TrimSpace(k)
			if k == "" {
				continue
			}
			vals.Set(k, v)
		}
		if len(vals) == 0 {
			response.ErrorJSON(ctx, "formUrlEncoded 至少需要一项有效字段", constants.ED00007)
			return
		}
		httpResp, execErr = ctl.dispatchWithBody(baseCtx, method, requestURL, vals, opts)

	case hasMP:
		fd, ferr := buildFormDataFromItems(req.FormData)
		if ferr != nil {
			response.ErrorJSON(ctx, ferr.Error(), constants.ED00007)
			return
		}
		defer fd.Close()

		stripContentTypeHeader(mergedHeaders)
		opts = buildRequestOptions(req, mergedHeaders)

		httpResp, execErr = ctl.dispatchWithBody(baseCtx, method, requestURL, fd, opts)

	default:
		httpResp, execErr = ctl.dispatchWithBody(baseCtx, method, requestURL, bodyArg(req.Body), opts)
	}

	durationMs := time.Since(start).Milliseconds()

	if execErr != nil {
		logger.Warn("hubplugin http 代发失败", "url", requestURL, "method", method, "err", execErr)
		response.ErrorJSON(ctx, "请求失败: "+execErr.Error(), constants.ED00004)
		return
	}
	if httpResp == nil {
		response.ErrorJSON(ctx, "无响应", constants.ED00009)
		return
	}

	bodyBytes := httpResp.Body
	if int64(len(bodyBytes)) > maxResponseBodyBytes {
		response.ErrorJSON(ctx, "响应体超过允许大小", constants.ED00009)
		return
	}

	bodyStr, isB64 := encodeResponseBodyForBiz(bodyBytes)
	out := models.HttpExecuteResult{
		StatusCode: httpResp.StatusCode,
		Status:     httpResp.Status,
		Headers:    flattenHeader(httpResp.Headers),
		Body:       bodyStr,
		BodyBase64: isB64,
		DurationMs: durationMs,
	}
	response.SuccessJSON(ctx, out, constants.SD00002)
}

// dispatchWithBody 按方法代发请求；GET/HEAD 等亦可带正文，统一走 Request。
func (ctl *HttpRequestController) dispatchWithBody(
	ctx context.Context,
	method, requestURL string,
	body interface{},
	opts []httpclient.RequestOption,
) (*httpclient.Response, error) {
	return ctl.client.Request(ctx, method, requestURL, body, opts...)
}

// mergeAuthHeaders 合并 Headers 与 Auth；Authorization、API Key 相关键在 Headers 已存在（忽略大小写）时不覆盖。
func mergeAuthHeaders(h map[string]string, auth *models.HttpExecuteAuth) map[string]string {
	out := make(map[string]string)
	for k, v := range h {
		out[k] = v
	}
	if auth == nil {
		return out
	}
	if _, ok := headerKeyCI(out, "authorization"); !ok {
		if t := strings.TrimSpace(auth.BearerToken); t != "" {
			out["Authorization"] = "Bearer " + t
		} else if auth.BasicUser != "" || auth.BasicPassword != "" {
			raw := auth.BasicUser + ":" + auth.BasicPassword
			out["Authorization"] = "Basic " + base64.StdEncoding.EncodeToString([]byte(raw))
		}
	}
	v := strings.TrimSpace(auth.ApiKeyValue)
	if v == "" {
		return out
	}
	if strings.TrimSpace(auth.ApiKeyQueryName) != "" {
		return out
	}
	hn := strings.TrimSpace(auth.ApiKeyHeaderName)
	if hn == "" {
		hn = "X-API-Key"
	}
	if _, ok := headerKeyCI(out, strings.ToLower(hn)); !ok {
		out[hn] = v
	}
	return out
}

// applyAuthQuery 若配置了 ApiKeyQueryName，将 ApiKeyValue 写入 URL 查询参数。
func applyAuthQuery(u *url.URL, auth *models.HttpExecuteAuth) {
	if auth == nil {
		return
	}
	qn := strings.TrimSpace(auth.ApiKeyQueryName)
	v := strings.TrimSpace(auth.ApiKeyValue)
	if qn == "" || v == "" {
		return
	}
	q := u.Query()
	q.Set(qn, v)
	u.RawQuery = q.Encode()
}

func headerKeyCI(h map[string]string, lowerKey string) (string, bool) {
	for k := range h {
		if strings.EqualFold(k, lowerKey) {
			return k, true
		}
	}
	return "", false
}

func stripContentTypeHeader(h map[string]string) {
	for k := range h {
		if strings.EqualFold(k, "Content-Type") {
			delete(h, k)
			return
		}
	}
}

// buildFormDataFromItems 将 formData 切片转为 httpclient.FormData（顺序与切片一致）。
func buildFormDataFromItems(items []models.HttpExecuteFormDataItem) (*httpclient.FormData, error) {
	fd := &httpclient.FormData{}
	added := 0
	for _, it := range items {
		key := strings.TrimSpace(it.Key)
		if key == "" {
			continue
		}
		t := strings.ToLower(strings.TrimSpace(it.Type))
		switch t {
		case "text":
			fd.AddField(key, it.Value)
			added++
		case "file":
			fname := strings.TrimSpace(it.FileName)
			b64 := strings.TrimSpace(it.ContentBase64)
			if fname == "" || b64 == "" {
				return nil, fmt.Errorf("文件字段 %s 需提供 fileName 与 contentBase64", key)
			}
			raw, err := base64.StdEncoding.DecodeString(b64)
			if err != nil {
				return nil, fmt.Errorf("字段 %s 的 Base64 无效", key)
			}
			if int64(len(raw)) > maxMultipartDecodedBytes {
				return nil, fmt.Errorf("字段 %s 解码后超过大小限制", key)
			}
			fd.AddFileReader(key, fname, bytes.NewReader(raw))
			added++
		default:
			return nil, fmt.Errorf("字段 %s 的 type 须为 text 或 file", key)
		}
	}
	if added == 0 {
		return nil, fmt.Errorf("formData 至少需一项有效字段")
	}
	return fd, nil
}

// buildRequestOptions 将 JSON 选项转为 httpclient.RequestOption。
func buildRequestOptions(req models.HttpExecuteRequest, headers map[string]string) []httpclient.RequestOption {
	var opts []httpclient.RequestOption
	if len(headers) > 0 {
		opts = append(opts, httpclient.WithHeaders(headers))
	}
	if req.TimeoutSeconds > 0 {
		sec := req.TimeoutSeconds
		if sec > maxHttpExecuteTimeoutSec {
			sec = maxHttpExecuteTimeoutSec
		}
		opts = append(opts, httpclient.WithTimeout(time.Duration(sec)*time.Second))
	}
	if req.FollowRedirects != nil {
		opts = append(opts, httpclient.WithFollowRedirects(*req.FollowRedirects))
	}
	return opts
}

// bodyArg 将 JSON 中的 body 片段转为 httpclient 可接受的 body；空表示 nil。
// 入参为 json.RawMessage：JSON 字符串会解码为 Go string（避免把引号当作正文发出）；
// 对象/数组等其它 JSON 值则按原始字节序列发送。
func bodyArg(raw []byte) interface{} {
	if len(raw) == 0 {
		return nil
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	return []byte(raw)
}

// flattenHeader 将 http.Header 转为单 map，多值用逗号拼接。
func flattenHeader(h http.Header) map[string]string {
	if h == nil {
		return map[string]string{}
	}
	out := make(map[string]string, len(h))
	for k, vals := range h {
		if len(vals) == 0 {
			continue
		}
		out[k] = strings.Join(vals, ", ")
	}
	return out
}

// encodeResponseBodyForBiz 将下游响应体转为 HttpExecuteResult.Body：UTF-8 报文（含 JSON、XML、纯文本等）原样转 string，否则 Base64。
func encodeResponseBodyForBiz(b []byte) (string, bool) {
	if len(b) == 0 {
		return "", false
	}
	if !utf8.Valid(b) {
		return base64.StdEncoding.EncodeToString(b), true
	}
	return string(b), false
}
