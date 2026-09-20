package models

import (
	"gateway/internal/servicecenterv3/catalog"
)

// Namespace 命名空间模型，使用 types.Namespace
// 对应数据库表：HUB_SERVICE_NAMESPACE
type Namespace = catalog.Namespace

// NamespaceQuery 命名空间查询条件，对应前端搜索表单的查询参数
type NamespaceQuery struct {
	NamespaceName string `json:"namespaceName" form:"namespaceName" query:"namespaceName"` // 命名空间名称（模糊查询）
	NamespaceId   string `json:"namespaceId" form:"namespaceId" query:"namespaceId"`       // 命名空间ID（精确查询）
	InstanceName  string `json:"instanceName" form:"instanceName" query:"instanceName"`    // 服务中心实例名称
	Environment   string `json:"environment" form:"environment" query:"environment"`       // 部署环境（DEVELOPMENT, STAGING, PRODUCTION）
	ActiveFlag    string `json:"activeFlag" form:"activeFlag" query:"activeFlag"`          // 活动标记：Y-活动，N-非活动，空表示全部（默认查询活动状态）
}

// NamespaceUsage 命名空间下的注册统计（非表字段）。
type NamespaceUsage struct {
	ServiceCount     int
	NodeCount        int
	HealthyNodeCount int
	ConnectionCount  int
}

// ToMap 将命名空间对象转为列表/详情响应，便于叠加运行时统计。
func ToMap(namespace *Namespace) map[string]interface{} {
	if namespace == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"namespaceId":          namespace.NamespaceId,
		"tenantId":             namespace.TenantId,
		"instanceName":         namespace.InstanceName,
		"environment":          namespace.Environment,
		"namespaceName":        namespace.NamespaceName,
		"namespaceDescription": namespace.NamespaceDesc,
		"serviceQuotaLimit":    namespace.ServiceQuotaLimit,
		"configQuotaLimit":     namespace.ConfigQuotaLimit,
		"addTime":              namespace.AddTime,
		"addWho":               namespace.AddWho,
		"editTime":             namespace.EditTime,
		"editWho":              namespace.EditWho,
		"oprSeqFlag":           namespace.OprSeqFlag,
		"currentVersion":       namespace.CurrentVersion,
		"activeFlag":           namespace.ActiveFlag,
		"noteText":             namespace.NoteText,
		"extProperty":          namespace.ExtProperty,
		"serviceCount":         0,
		"nodeCount":            0,
		"healthyNodeCount":     0,
		"connectionCount":      0,
		"connections":          []map[string]interface{}{},
	}
}
