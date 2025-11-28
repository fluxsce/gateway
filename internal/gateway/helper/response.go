package helper

import "gateway/pkg/utils/ctime"

// GatewayResponse 定义网关对外返回的统一结构
// 通过 helper 构造，确保所有字段保持一致格式
type GatewayResponse struct {
	Code      string `json:"code"`
	Error     string `json:"error"`
	Domain    string `json:"domain"`
	Timestamp string `json:"timestamp"`
	Path      string `json:"path"`
	TraceID   string `json:"trace_id"`
}

// BuildGatewayResponse 构造统一响应结构
// time 字段通过 ctime.GetCurrentTimeString 精确到毫秒
func BuildGatewayResponse(code, errMsg, domain, path, traceID string) GatewayResponse {
	if domain == "" {
		domain = "gateway"
	}

	return GatewayResponse{
		Code:      code,
		Error:     errMsg,
		Domain:    domain,
		Timestamp: ctime.GetCurrentTimeString(ctime.FormatISO8601Milli),
		Path:      path,
		TraceID:   traceID,
	}
}
