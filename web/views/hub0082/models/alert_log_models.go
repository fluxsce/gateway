package models

import "time"

// AlertLogQueryRequest 预警日志查询请求
// 说明：分页参数通过 request.GetPaginationParams 读取（page/pageSize），这里仅放筛选条件
type AlertLogQueryRequest struct {
	AlertLogId     string     `json:"alertLogId" form:"alertLogId"`         // 告警日志ID（精确）
	AlertLevel     string     `json:"alertLevel" form:"alertLevel"`         // 告警级别：INFO/WARN/ERROR/CRITICAL
	AlertType      *string    `json:"alertType" form:"alertType"`           // 告警类型（精确）
	AlertTitle     string     `json:"alertTitle" form:"alertTitle"`         // 告警标题（模糊）
	ChannelName    *string    `json:"channelName" form:"channelName"`       // 渠道名称（精确）
	SendStatus     *string    `json:"sendStatus" form:"sendStatus"`         // 发送状态：PENDING/SENDING/SUCCESS/FAILED/IGNORED
	AlertTimestamp *time.Time `json:"alertTimestamp" form:"alertTimestamp"` // 告警时间戳（精确）
	StartTime      string     `json:"startTime" form:"startTime"`           // 开始时间，本地墙钟字符串
	EndTime        string     `json:"endTime" form:"endTime"`               // 结束时间，本地墙钟字符串
}

// AlertLogIgnoreScope 忽略范围
const (
	AlertLogIgnoreScopeSelected = "selected" // 当前选中
	AlertLogIgnoreScopeGroup    = "group"    // 同分组（优先 alertType，否则 alertTitle）
	AlertLogIgnoreScopeAll      = "all"      // 当前查询条件下全部待发送
)

// AlertLogIgnoreRequest 忽略待发送预警日志
// 仅将 sendStatus=PENDING 的记录改为 IGNORED，已发送/发送中的不受影响
type AlertLogIgnoreRequest struct {
	Scope       string   `json:"scope" form:"scope"`             // selected / group / all
	AlertLogIds string `json:"alertLogIds" form:"alertLogIds"` // selected 必填，逗号分隔
	AlertLogId  string   `json:"alertLogId" form:"alertLogId"`   // all：精确日志ID
	AlertType   string   `json:"alertType" form:"alertType"`     // group：优先按类型
	AlertTitle  string   `json:"alertTitle" form:"alertTitle"`   // group：类型为空时按标题
	AlertLevel  string   `json:"alertLevel" form:"alertLevel"`
	ChannelName string   `json:"channelName" form:"channelName"`
	StartTime   string   `json:"startTime" form:"startTime"`
	EndTime     string   `json:"endTime" form:"endTime"`
}
