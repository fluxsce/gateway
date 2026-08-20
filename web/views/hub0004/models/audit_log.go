package models

import "time"

// AuthAuditLog 权限审计日志，对应表 HUB_AUTH_AUDIT_LOG。
// 记录控制台写操作：谁、在哪个模块、对哪条数据、做了何种动作。
type AuthAuditLog struct {
	AuditId        string    `json:"auditId" db:"auditId"`
	TenantId       string    `json:"tenantId" db:"tenantId"`
	UserId         string    `json:"userId" db:"userId"`
	UserName       string    `json:"userName" db:"userName"`
	Action         string    `json:"action" db:"action"`
	ModuleCode     string    `json:"moduleCode" db:"moduleCode"`
	TargetType     string    `json:"targetType" db:"targetType"`
	TargetId       string    `json:"targetId" db:"targetId"`
	TargetName     string    `json:"targetName" db:"targetName"`
	ResourceCode   string    `json:"resourceCode" db:"resourceCode"`
	RequestPath    string    `json:"requestPath" db:"requestPath"`
	RequestMethod  string    `json:"requestMethod" db:"requestMethod"`
	ClientIP       string    `json:"clientIP" db:"clientIP"`
	Result         string    `json:"result" db:"result"`
	Detail         string    `json:"detail" db:"detail"`
	AddTime        time.Time `json:"addTime" db:"addTime"`
	AddWho         string    `json:"addWho" db:"addWho"`
	EditTime       time.Time `json:"editTime" db:"editTime"`
	EditWho        string    `json:"editWho" db:"editWho"`
	OprSeqFlag     string    `json:"oprSeqFlag" db:"oprSeqFlag"`
	CurrentVersion int       `json:"currentVersion" db:"currentVersion"`
	ActiveFlag     string    `json:"activeFlag" db:"activeFlag"`
}

// TableName 返回权限审计表名。
func (AuthAuditLog) TableName() string {
	return "HUB_AUTH_AUDIT_LOG"
}

// AuthAuditLogQuery 审计日志查询条件。
// 分页由 request.GetPaginationParams 读取，本结构只放筛选字段。
type AuthAuditLogQuery struct {
	AuditId      string `json:"auditId" form:"auditId"`
	UserId       string `json:"userId" form:"userId"`
	UserName     string `json:"userName" form:"userName"`
	Action       string `json:"action" form:"action"`
	ModuleCode   string `json:"moduleCode" form:"moduleCode"`
	TargetType   string `json:"targetType" form:"targetType"`
	TargetId     string `json:"targetId" form:"targetId"`
	TargetName   string `json:"targetName" form:"targetName"`
	ResourceCode string `json:"resourceCode" form:"resourceCode"`
	ClientIP     string `json:"clientIP" form:"clientIP"`
	Result       string `json:"result" form:"result"`
	StartTime    string `json:"startTime" form:"startTime"`
	EndTime      string `json:"endTime" form:"endTime"`
}
