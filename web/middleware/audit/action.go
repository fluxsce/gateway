package audit

import (
	"encoding/json"
	"strings"
)

// AuditActionFromResourceCode 按按钮码后缀落到动作细码。
// 只读按钮返回空串；未列入白名单的写按钮默认 UPDATE，避免漏记。
func AuditActionFromResourceCode(resourceCode string) string {
	if IsReadOnlyAuditButtonCode(resourceCode) {
		return ""
	}
	switch {
	case strings.HasSuffix(resourceCode, ":delete"):
		return AuditActionDelete
	case strings.HasSuffix(resourceCode, ":rollback"):
		return AuditActionRollback
	case strings.HasSuffix(resourceCode, ":add"), strings.HasSuffix(resourceCode, ":create"):
		return AuditActionCreate
	case strings.HasSuffix(resourceCode, ":roleAuth"):
		return AuditActionGrant
	case strings.HasSuffix(resourceCode, ":export"):
		return AuditActionExport
	case strings.HasSuffix(resourceCode, ":edit"), strings.HasSuffix(resourceCode, ":update"),
		strings.HasSuffix(resourceCode, ":resetPassword"),
		strings.HasSuffix(resourceCode, ":start"),
		strings.HasSuffix(resourceCode, ":stop"),
		strings.HasSuffix(resourceCode, ":restart"),
		strings.HasSuffix(resourceCode, ":reload"),
		strings.HasSuffix(resourceCode, ":trigger"),
		strings.HasSuffix(resourceCode, ":offline"),
		strings.HasSuffix(resourceCode, ":register"),
		strings.HasSuffix(resourceCode, ":unregister"),
		strings.HasSuffix(resourceCode, ":connect"),
		strings.HasSuffix(resourceCode, ":disconnect"),
		strings.HasSuffix(resourceCode, ":import"),
		strings.HasSuffix(resourceCode, ":setDefault"),
		strings.HasSuffix(resourceCode, ":batchReset"),
		strings.HasSuffix(resourceCode, ":securityConfig"),
		strings.HasSuffix(resourceCode, ":test"),
		strings.HasSuffix(resourceCode, ":logConfig"),
		strings.HasSuffix(resourceCode, ":routerConfig"),
		strings.HasSuffix(resourceCode, ":staticHostConfig"),
		strings.HasSuffix(resourceCode, ":addProxy"),
		strings.HasSuffix(resourceCode, ":circuitBreaker"),
		isGatewayLogResetButton(resourceCode):
		return AuditActionUpdate
	default:
		return AuditActionUpdate
	}
}

// IsReadOnlyAuditButtonCode 查询、查看、表单重置等只读按钮不写审计。
func IsReadOnlyAuditButtonCode(resourceCode string) bool {
	if resourceCode == "" {
		return true
	}
	if isGatewayLogResetButton(resourceCode) {
		return false
	}
	readOnlySuffixes := []string{":search", ":view", ":query", ":resetQuery", ":refresh", ":back", ":reset"}
	for _, suffix := range readOnlySuffixes {
		if strings.HasSuffix(resourceCode, suffix) {
			return true
		}
	}
	return false
}

// isGatewayLogResetButton 网关日志重发（hub0023:reset）是写操作，与各模块查询表单重置区分。
func isGatewayLogResetButton(resourceCode string) bool {
	return resourceCode == "hub0023:reset"
}

// IsWriteAuditButtonCode 判断按钮码是否属于可审计的写操作。
// 不含 :roleAuth：查询角色资源也用该码，授权由控制器 SetEvent。
func IsWriteAuditButtonCode(code string) bool {
	action := AuditActionFromResourceCode(code)
	return action != "" && action != AuditActionGrant
}

// FirstWriteAuditButtonCode 返回路由声明里第一个可审计的写按钮码。
func FirstWriteAuditButtonCode(codes []string) string {
	for _, code := range codes {
		if IsWriteAuditButtonCode(code) {
			return code
		}
	}
	return ""
}

// SanitizeAuditDetail 从业务显式传入的字段生成审计摘要：去掉口令/密钥等，并截断长度。
// 供控制器组装 Detail，不用于中间件扫描请求体。
func SanitizeAuditDetail(params map[string]interface{}) string {
	if len(params) == 0 {
		return ""
	}
	cleaned := make(map[string]interface{}, len(params))
	for key, val := range params {
		lower := strings.ToLower(key)
		if isSensitiveAuditKey(lower) || isOmittedAuditKey(lower) {
			continue
		}
		cleaned[key] = val
	}
	if len(cleaned) == 0 {
		return ""
	}
	raw, err := json.Marshal(cleaned)
	if err != nil {
		return ""
	}
	detail := string(raw)
	if len(detail) > maxAuditDetailLen {
		return detail[:maxAuditDetailLen]
	}
	return detail
}

func isSensitiveAuditKey(lowerKey string) bool {
	tokens := []string{"password", "secret", "token", "captcha", "cookie"}
	for _, token := range tokens {
		if strings.Contains(lowerKey, token) {
			return true
		}
	}
	return false
}

func isOmittedAuditKey(lowerKey string) bool {
	switch lowerKey {
	case "content", "filterconfig", "customconfigjson", "filterconfigjson", "configcontent":
		return true
	default:
		return false
	}
}

// IsSensitiveButtonCode 判断按钮码是否属于删除/回滚/授权/重置密码。
func IsSensitiveButtonCode(code string) bool {
	if code == "" {
		return false
	}
	suffixes := []string{":delete", ":rollback", ":roleAuth", ":resetPassword"}
	for _, suffix := range suffixes {
		if strings.HasSuffix(code, suffix) {
			return true
		}
	}
	return false
}
