package models

import (
	"time"
)

// FilterConfig 过滤器配置模型，对应数据库HUB_GATEWAY_FILTER_CONFIG表
type FilterConfig struct {
	TenantId          string `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                                     // 租户ID，联合主键
	FilterConfigId    string `json:"filterConfigId" form:"filterConfigId" query:"filterConfigId" db:"filterConfigId"`             // 过滤器配置ID，联合主键
	GatewayInstanceId string `json:"gatewayInstanceId" form:"gatewayInstanceId" query:"gatewayInstanceId" db:"gatewayInstanceId"` // 网关实例ID(实例级过滤器)
	RouteConfigId     string `json:"routeConfigId" form:"routeConfigId" query:"routeConfigId" db:"routeConfigId"`                 // 路由配置ID(路由级过滤器)
	FilterName        string `json:"filterName" form:"filterName" query:"filterName" db:"filterName"`                             // 过滤器名称

	// 根据FilterType枚举值设计 - 支持7种过滤器类型
	FilterType string `json:"filterType" form:"filterType" query:"filterType" db:"filterType"` // 过滤器类型(header,query-param,body,url,method,cookie,response)

	// 根据FilterAction枚举值设计 - 支持3种执行时机
	FilterAction string `json:"filterAction" form:"filterAction" query:"filterAction" db:"filterAction"` // 过滤器执行时机(pre-routing,post-routing,pre-response)

	FilterOrder  int    `json:"filterOrder" form:"filterOrder" query:"filterOrder" db:"filterOrder"`       // 过滤器执行顺序(Priority)
	FilterConfig string `json:"filterConfig" form:"filterConfig" query:"filterConfig" db:"filterConfig"`   // 过滤器具体配置,JSON格式
	FilterDesc   string `json:"filterDesc" form:"filterDesc" query:"filterDesc" db:"filterDesc"`           // 过滤器描述

	// 根据FilterConfig结构设计的附属字段
	ConfigId string `json:"configId" form:"configId" query:"configId" db:"configId"` // 过滤器配置ID(来自FilterConfig.ID)

	// 预留字段
	Reserved1 string     `json:"reserved1" form:"reserved1" query:"reserved1" db:"reserved1"` // 预留字段1
	Reserved2 string     `json:"reserved2" form:"reserved2" query:"reserved2" db:"reserved2"` // 预留字段2
	Reserved3 *int       `json:"reserved3" form:"reserved3" query:"reserved3" db:"reserved3"` // 预留字段3
	Reserved4 *int       `json:"reserved4" form:"reserved4" query:"reserved4" db:"reserved4"` // 预留字段4
	Reserved5 *time.Time `json:"reserved5" form:"reserved5" query:"reserved5" db:"reserved5"` // 预留字段5

	// 扩展属性
	ExtProperty string `json:"extProperty" form:"extProperty" query:"extProperty" db:"extProperty"` // 扩展属性,JSON格式

	// 标准字段
	AddTime        time.Time `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`                             // 创建时间
	AddWho         string    `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`                                 // 创建人ID
	EditTime       time.Time `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`                         // 最后修改时间
	EditWho        string    `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`                             // 最后修改人ID
	OprSeqFlag     string    `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`                 // 操作序列标识
	CurrentVersion int       `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"` // 当前版本号
	ActiveFlag     string    `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag"`                 // 活动状态标记(N非活动/禁用,Y活动/启用)
	NoteText       string    `json:"noteText" form:"noteText" query:"noteText" db:"noteText"`                         // 备注信息
}

// TableName 返回表名
func (FilterConfig) TableName() string {
	return "HUB_GATEWAY_FILTER_CONFIG"
}

// FilterType 过滤器类型常量
const (
	FilterTypeHeader     = "header"      // 请求头过滤器
	FilterTypeQueryParam = "query-param" // 查询参数过滤器
	FilterTypeBody       = "body"        // 请求体过滤器
	FilterTypeStrip      = "strip"       // 前缀剥离过滤器
	FilterTypeRewrite    = "rewrite"     // 路径重写过滤器
	FilterTypeMethod     = "method"      // HTTP方法过滤器
	FilterTypeCookie     = "cookie"      // Cookie过滤器
	FilterTypeResponse   = "response"    // 响应过滤器
)

// FilterAction 过滤器执行时机常量
const (
	FilterActionPreRouting  = "pre-routing"  // 路由匹配前执行
	FilterActionPostRouting = "post-routing" // 路由匹配后执行
	FilterActionPreResponse = "pre-response" // 响应返回前执行
)

// GetValidFilterTypes 获取有效的过滤器类型列表
func GetValidFilterTypes() []string {
	return []string{
		FilterTypeHeader,
		FilterTypeQueryParam,
		FilterTypeBody,
		FilterTypeStrip,
		FilterTypeRewrite,
		FilterTypeMethod,
		FilterTypeCookie,
		FilterTypeResponse,
	}
}

// GetValidFilterActions 获取有效的过滤器执行时机列表
func GetValidFilterActions() []string {
	return []string{
		FilterActionPreRouting,
		FilterActionPostRouting,
		FilterActionPreResponse,
	}
}

// IsValidFilterType 检查过滤器类型是否有效
func IsValidFilterType(filterType string) bool {
	validTypes := GetValidFilterTypes()
	for _, validType := range validTypes {
		if filterType == validType {
			return true
		}
	}
	return false
}

// IsValidFilterAction 检查过滤器执行时机是否有效
func IsValidFilterAction(filterAction string) bool {
	validActions := GetValidFilterActions()
	for _, validAction := range validActions {
		if filterAction == validAction {
			return true
		}
	}
	return false
}

// FilterConfigTemplate 过滤器配置模板
type FilterConfigTemplate struct {
	Name         string                 `json:"name"`         // 模板名称
	Description  string                 `json:"description"`  // 模板描述
	FilterType   string                 `json:"filterType"`   // 过滤器类型
	FilterAction string                 `json:"filterAction"` // 执行时机
	DefaultOrder int                    `json:"defaultOrder"` // 默认执行顺序
	ConfigSchema map[string]interface{} `json:"configSchema"` // 配置模板
}

// GetFilterConfigTemplates 获取预定义的过滤器配置模板
func GetFilterConfigTemplates() []FilterConfigTemplate {
	return []FilterConfigTemplate{
		{
			Name:         "添加请求头",
			Description:  "为请求添加自定义头信息",
			FilterType:   FilterTypeHeader,
			FilterAction: FilterActionPreRouting,
			DefaultOrder: 10,
			ConfigSchema: map[string]interface{}{
				"add_headers": map[string]string{
					"X-Gateway-Version": "1.0.0",
					"X-Request-ID":      "${request_id}",
				},
			},
		},
		{
			Name:         "移除敏感请求头",
			Description:  "移除请求中的敏感头信息",
			FilterType:   FilterTypeHeader,
			FilterAction: FilterActionPreRouting,
			DefaultOrder: 5,
			ConfigSchema: map[string]interface{}{
				"remove_headers": []string{"X-Forwarded-For", "X-Real-IP"},
			},
		},
		{
			Name:         "查询参数过滤",
			Description:  "过滤或添加查询参数",
			FilterType:   FilterTypeQueryParam,
			FilterAction: FilterActionPreRouting,
			DefaultOrder: 15,
			ConfigSchema: map[string]interface{}{
				"remove_params": []string{"internal_token"},
				"add_params": map[string]string{
					"version": "v1",
				},
			},
		},
		{
			Name:         "HTTP方法限制",
			Description:  "限制允许的HTTP方法",
			FilterType:   FilterTypeMethod,
			FilterAction: FilterActionPreRouting,
			DefaultOrder: 1,
			ConfigSchema: map[string]interface{}{
				"allowed_methods": []string{"GET", "POST", "PUT", "DELETE"},
				"reject_status":   405,
				"reject_message":  "Method not allowed",
			},
		},
		{
			Name:         "响应头添加",
			Description:  "为响应添加安全头信息",
			FilterType:   FilterTypeResponse,
			FilterAction: FilterActionPreResponse,
			DefaultOrder: 90,
			ConfigSchema: map[string]interface{}{
				"add_headers": map[string]string{
					"X-Content-Type-Options": "nosniff",
					"X-Frame-Options":        "DENY",
					"X-XSS-Protection":       "1; mode=block",
				},
			},
		},
	}
} 