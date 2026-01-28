package models

// ConfigQuery 配置查询条件
type ConfigQuery struct {
	NamespaceId  string `json:"namespaceId" form:"namespaceId" query:"namespaceId"`    // 命名空间ID
	GroupName    string `json:"groupName" form:"groupName" query:"groupName"`          // 分组名称
	ConfigDataId string `json:"configDataId" form:"configDataId" query:"configDataId"` // 配置数据ID（模糊查询）
	ContentType  string `json:"contentType" form:"contentType" query:"contentType"`    // 内容类型
	ActiveFlag   string `json:"activeFlag" form:"activeFlag" query:"activeFlag"`       // 活动状态
}

// ConfigHistoryRequest 配置历史查询请求
type ConfigHistoryRequest struct {
	NamespaceId  string `json:"namespaceId" form:"namespaceId" query:"namespaceId" binding:"required"`    // 命名空间ID
	GroupName    string `json:"groupName" form:"groupName" query:"groupName" binding:"required"`          // 分组名称
	ConfigDataId string `json:"configDataId" form:"configDataId" query:"configDataId" binding:"required"` // 配置数据ID
	Limit        int    `json:"limit" form:"limit" query:"limit"`                                         // 限制数量，默认50
}

// RollbackRequest 配置回滚请求
type RollbackRequest struct {
	ConfigHistoryId string `json:"configHistoryId" form:"configHistoryId" binding:"required"` // 配置历史ID（唯一标识）
	ChangeReason    string `json:"changeReason" form:"changeReason"`                          // 变更原因
}
