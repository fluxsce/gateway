package models

import (
	"gateway/internal/servicecenter/types"
)

// Namespace 命名空间模型，使用 types.Namespace
// 对应数据库表：HUB_SERVICE_NAMESPACE
type Namespace = types.Namespace

// NamespaceQuery 命名空间查询条件，对应前端搜索表单的查询参数
type NamespaceQuery struct {
	NamespaceName string `json:"namespaceName" form:"namespaceName" query:"namespaceName"` // 命名空间名称（模糊查询）
	InstanceName  string `json:"instanceName" form:"instanceName" query:"instanceName"`    // 服务中心实例名称
	Environment   string `json:"environment" form:"environment" query:"environment"`       // 部署环境（DEVELOPMENT, STAGING, PRODUCTION）
	ActiveFlag    string `json:"activeFlag" form:"activeFlag" query:"activeFlag"`          // 活动标记：Y-活动，N-非活动，空表示全部（默认查询活动状态）
}
