package models

import "time"

// ToolInstall 用户已安装的工具，对应表 HUB_TOOL_INSTALL。
// 同一租户、同一用户、同一工具一行；卸载把 activeFlag 置为 N。
type ToolInstall struct {
	TenantId       string    `json:"tenantId" db:"tenantId"`
	UserId         string    `json:"userId" db:"userId"`
	ToolId         string    `json:"toolId" db:"toolId"`
	ConfigJson     string    `json:"configJson" db:"configJson"`
	InstallTime    time.Time `json:"installTime" db:"installTime"`
	AddTime        time.Time `json:"addTime" db:"addTime"`
	AddWho         string    `json:"addWho" db:"addWho"`
	EditTime       time.Time `json:"editTime" db:"editTime"`
	EditWho        string    `json:"editWho" db:"editWho"`
	OprSeqFlag     string    `json:"oprSeqFlag" db:"oprSeqFlag"`
	CurrentVersion int       `json:"currentVersion" db:"currentVersion"`
	ActiveFlag     string    `json:"activeFlag" db:"activeFlag"`
}

// TableName 返回安装表名。
func (ToolInstall) TableName() string {
	return "HUB_TOOL_INSTALL"
}

// ToolActionRequest 安装、卸载或保存配置。
type ToolActionRequest struct {
	ToolId string         `json:"toolId" form:"toolId"`
	Config map[string]any `json:"config" form:"config"`
}
