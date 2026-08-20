package permission

import (
	"context"
	"strings"
	"time"

	"gateway/pkg/logger"
	"gateway/pkg/utils/random"
)

const (
	// AuditResultSuccess 审计结果：成功
	AuditResultSuccess = "Y"
	// AuditResultFail 审计结果：失败
	AuditResultFail = "N"

	// AuditActionCreate 新增
	AuditActionCreate = "CREATE"
	// AuditActionUpdate 修改（含禁用、重置密码）
	AuditActionUpdate = "UPDATE"
	// AuditActionDelete 删除
	AuditActionDelete = "DELETE"
	// AuditActionRollback 回滚
	AuditActionRollback = "ROLLBACK"
	// AuditActionGrant 授权（角色资源、用户角色）
	AuditActionGrant = "GRANT"
)

const maxAuditDetailLen = 2000

// AuditEvent 权限审计事件，写入 HUB_AUTH_AUDIT_LOG。
// 业务模块在写操作成功后填写 ModuleCode、TargetType、TargetId、TargetName，即可记录「谁改了哪条数据」。
type AuditEvent struct {
	UserId        string
	TenantId      string
	UserName      string
	Action        string
	ModuleCode    string
	TargetType    string
	TargetId      string
	TargetName    string
	ResourceCode  string
	RequestPath   string
	RequestMethod string
	ClientIP      string
	Result        string
	Detail        string
}

// WriteAudit 写入一条权限审计。失败只记日志，不影响主流程。
func (ps *PermissionService) WriteAudit(ctx context.Context, event *AuditEvent) {
	if ps == nil || ps.dao == nil || event == nil {
		return
	}
	if event.Result == "" {
		event.Result = AuditResultSuccess
	}
	if event.ModuleCode == "" {
		event.ModuleCode = ModuleCodeFromResourceOrPath(event.ResourceCode, event.RequestPath)
	}
	detail := event.Detail
	if len(detail) > maxAuditDetailLen {
		detail = detail[:maxAuditDetailLen]
	}

	now := time.Now()
	auditId := random.GenerateUniqueStringWithPrefix("AUD", 32)
	oprSeqFlag := random.GenerateUniqueStringWithPrefix("", 32)

	sql := `
		INSERT INTO HUB_AUTH_AUDIT_LOG (
			auditId, tenantId, userId, userName,
			action, moduleCode, targetType, targetId, targetName, resourceCode,
			requestPath, requestMethod, clientIP, result, detail,
			addTime, addWho, editTime, editWho,
			oprSeqFlag, currentVersion, activeFlag
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := ps.dao.db.Exec(ctx, sql, []interface{}{
		auditId, event.TenantId, event.UserId, event.UserName,
		event.Action, event.ModuleCode, event.TargetType, event.TargetId, event.TargetName, event.ResourceCode,
		event.RequestPath, event.RequestMethod, event.ClientIP, event.Result, detail,
		now, event.UserId, now, event.UserId,
		oprSeqFlag, 1, "Y",
	}, true)
	if err != nil {
		logger.Error("写入权限审计失败", "error", err, "action", event.Action, "userId", event.UserId)
	}
}

// ModuleCodeFromResourceOrPath 从资源码或 URL 解析模块编码，供审计筛选。
func ModuleCodeFromResourceOrPath(resourceCode, path string) string {
	if resourceCode != "" {
		if i := strings.Index(resourceCode, ":"); i > 0 {
			return resourceCode[:i]
		}
		if strings.HasPrefix(resourceCode, "hub") {
			return resourceCode
		}
	}
	return moduleCodeFromGatewayPath(path)
}

// moduleCodeFromGatewayPath 从 /gateway/{hubxxxx}/... 取出模块编码。
func moduleCodeFromGatewayPath(path string) string {
	const prefix = "/gateway/"
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	rest := strings.TrimPrefix(path, prefix)
	slash := strings.Index(rest, "/")
	if slash < 0 {
		if strings.HasPrefix(rest, "hub") {
			return rest
		}
		return ""
	}
	module := rest[:slash]
	if !strings.HasPrefix(module, "hub") {
		return ""
	}
	return module
}

// AuditActionFromResourceCode 按按钮码后缀落到动作细码。
func AuditActionFromResourceCode(resourceCode string) string {
	switch {
	case strings.HasSuffix(resourceCode, ":delete"):
		return AuditActionDelete
	case strings.HasSuffix(resourceCode, ":rollback"):
		return AuditActionRollback
	case strings.HasSuffix(resourceCode, ":add"), strings.HasSuffix(resourceCode, ":create"):
		return AuditActionCreate
	case strings.HasSuffix(resourceCode, ":roleAuth"):
		return AuditActionGrant
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
		isWriteResetButton(resourceCode):
		return AuditActionUpdate
	default:
		return ""
	}
}

// isWriteResetButton 区分日志重发（hub0023:reset）与查询表单重置（:resetQuery）。
func isWriteResetButton(resourceCode string) bool {
	if strings.HasSuffix(resourceCode, ":resetQuery") || strings.HasSuffix(resourceCode, ":resetPassword") {
		return false
	}
	return strings.HasSuffix(resourceCode, ":reset")
}

// IsWriteAuditButtonCode 中间件在写接口成功后兜底审计。
// 不含 :roleAuth：查询角色资源也用该码，授权由控制器写带目标名称的记录。
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

// AuditTargetSpec 从请求参数抽出的审计对象。
type AuditTargetSpec struct {
	TargetType string
	TargetId   string
	TargetName string
}

// AuditTargetFromGetter 按常见主键字段名取审计对象，供中间件在读完 body 后调用。
func AuditTargetFromGetter(get func(string) string) AuditTargetSpec {
	if get == nil {
		return AuditTargetSpec{TargetType: "API"}
	}
	keys := []struct {
		typ, idKey, nameKey string
	}{
		{"SECURITY_CONFIG", "securityConfigId", "configName"},
		{"ALERT_LOG", "alertLogId", "alertTitle"},
		{"USER", "userId", "realName"},
		{"ROLE", "roleId", "roleName"},
		{"RESOURCE", "resourceId", "resourceName"},
		{"INSTANCE", "gatewayInstanceId", "instanceName"},
		{"ROUTE", "routeConfigId", "routeName"},
		{"SERVICE", "serviceDefinitionId", "serviceName"},
		{"HTTP", "url", "method"},
	}
	for _, item := range keys {
		id := strings.TrimSpace(get(item.idKey))
		if id == "" {
			continue
		}
		name := strings.TrimSpace(get(item.nameKey))
		if item.typ == "USER" && name == "" {
			name = strings.TrimSpace(get("userName"))
		}
		if item.typ == "HTTP" && name != "" {
			name = name + " " + id
		}
		return AuditTargetSpec{TargetType: item.typ, TargetId: id, TargetName: name}
	}
	return AuditTargetSpec{TargetType: "API"}
}

// IsSensitiveButtonCode 判断按钮码是否属于删除/回滚/授权/重置密码。
// 保留给单测与调用方；中间件兜底审计改用 IsWriteAuditButtonCode。
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
