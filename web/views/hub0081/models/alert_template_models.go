package models

// AlertTemplateQueryRequest 预警模板查询请求
// 说明：分页参数通过 request.GetPaginationParams 读取（page/pageSize），这里仅放筛选条件
type AlertTemplateQueryRequest struct {
	TemplateName  string `json:"templateName" form:"templateName"`   // 模板名称（精确）
	ChannelType   string `json:"channelType" form:"channelType"`     // 渠道类型（精确，可为空）
	DisplayFormat string `json:"displayFormat" form:"displayFormat"` // table/text
	ActiveFlag    string `json:"activeFlag" form:"activeFlag"`       // Y/N
}
