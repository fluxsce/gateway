package models

// ServiceEventQueryRequest 服务事件查询请求
type ServiceEventQueryRequest struct {
	// 租户ID
	TenantId string `json:"tenantId"`

	// 过滤条件
	ActiveFlag  string `json:"activeFlag"`  // 活动状态标记(Y活动,N非活动,空为全部)
	EventType   string `json:"eventType"`   // 事件类型过滤
	ServiceName string `json:"serviceName"` // 服务名称过滤（模糊查询）
	GroupName   string `json:"groupName"`   // 分组名称过滤
	HostAddress string `json:"hostAddress"` // 主机地址过滤（模糊查询）
	EventSource string `json:"eventSource"` // 事件来源过滤
	StartTime   string `json:"startTime"`   // 开始时间过滤（格式：2006-01-02 或 2006-01-02 15:04:05）
	EndTime     string `json:"endTime"`     // 结束时间过滤（格式：2006-01-02 或 2006-01-02 15:04:05）

	// 分页参数
	PageIndex int `json:"pageIndex"` // 页码，从1开始
	PageSize  int `json:"pageSize"`  // 每页数量，默认20，最大100

	// 扩展查询参数
	Keyword string `json:"keyword"` // 关键字搜索（搜索服务名称、事件消息等）
}

// ServiceEventSummary 服务事件摘要信息（用于列表展示，不包含大字段）
type ServiceEventSummary struct {
	ServiceEventId    string `json:"serviceEventId" db:"serviceEventId"`       // 服务事件ID
	TenantId          string `json:"tenantId" db:"tenantId"`                   // 租户ID
	ServiceGroupId    string `json:"serviceGroupId" db:"serviceGroupId"`       // 服务分组ID
	ServiceInstanceId string `json:"serviceInstanceId" db:"serviceInstanceId"` // 服务实例ID
	GroupName         string `json:"groupName" db:"groupName"`                 // 分组名称
	ServiceName       string `json:"serviceName" db:"serviceName"`             // 服务名称
	HostAddress       string `json:"hostAddress" db:"hostAddress"`             // 主机地址
	PortNumber        int    `json:"portNumber" db:"portNumber"`               // 端口号
	NodeIpAddress     string `json:"nodeIpAddress" db:"nodeIpAddress"`         // 节点IP地址，记录程序运行的IP
	EventType         string `json:"eventType" db:"eventType"`                 // 事件类型
	EventSource       string `json:"eventSource" db:"eventSource"`             // 事件来源
	EventMessage      string `json:"eventMessage" db:"eventMessage"`           // 事件消息
	EventTime         string `json:"eventTime" db:"eventTime"`                 // 事件时间
	AddTime           string `json:"addTime" db:"addTime"`                     // 添加时间
	AddWho            string `json:"addWho" db:"addWho"`                       // 添加人
	EditTime          string `json:"editTime" db:"editTime"`                   // 修改时间
	EditWho           string `json:"editWho" db:"editWho"`                     // 修改人
	OprSeqFlag        string `json:"oprSeqFlag" db:"oprSeqFlag"`               // 操作序列标记
	CurrentVersion    int    `json:"currentVersion" db:"currentVersion"`       // 当前版本号
	ActiveFlag        string `json:"activeFlag" db:"activeFlag"`               // 活动状态标记
	NoteText          string `json:"noteText" db:"noteText"`                   // 备注信息
}
