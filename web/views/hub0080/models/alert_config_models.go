package models

// AlertConfigQueryRequest 预警(告警)配置查询请求
// 说明：分页参数通过 request.GetPaginationParams 读取（page/pageSize），这里仅放筛选条件
type AlertConfigQueryRequest struct {
	ChannelName   string `json:"channelName" form:"channelName"`     // 渠道名称（精确）
	ChannelType   string `json:"channelType" form:"channelType"`     // 渠道类型（精确）
	ActiveFlag    string `json:"activeFlag" form:"activeFlag"`       // Y/N
	DefaultFlag   string `json:"defaultFlag" form:"defaultFlag"`     // Y/N
	PriorityLevel *int   `json:"priorityLevel" form:"priorityLevel"` // 优先级（精确）
}
