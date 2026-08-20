package audit

import (
	"strings"

	"gateway/web/utils/constants"
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
	// AuditActionExport 导出
	AuditActionExport = "EXPORT"
	// AuditActionLogin 登录成功
	AuditActionLogin = "LOGIN"
	// AuditActionLoginFail 登录失败（仅账号进入冷却时记录，不含验证码/单次密码错误）
	AuditActionLoginFail = "LOGIN_FAIL"
	// AuditActionKick 踢会话
	AuditActionKick = "KICK"
)

const maxAuditDetailLen = 2000

// AuditEvent 控制面写操作审计事件，由业务 SetEvent 放入上下文，DAO 写入 HUB_AUTH_AUDIT_LOG。
// 身份与请求信息由全局审计中间件补齐；鉴权中间件不填写本结构。
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
	prefix := constants.APIRoot + "/"
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
