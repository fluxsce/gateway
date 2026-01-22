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
	SendStatus     *string    `json:"sendStatus" form:"sendStatus"`         // 发送状态：PENDING/SENDING/SUCCESS/FAILED
	AlertTimestamp *time.Time `json:"alertTimestamp" form:"alertTimestamp"` // 告警时间戳（精确）
	StartTime      *time.Time `json:"startTime" form:"startTime"`           // 开始时间（用于时间范围查询）
	EndTime        *time.Time `json:"endTime" form:"endTime"`               // 结束时间（用于时间范围查询）
}
