package syssetting

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gateway/pkg/config"
)

// 分组编码。新增环境设置分组时在此追加常量，并补默认值、校验与前端子页。
const (
	// GroupRetention 归档策略，按数据类型配置保留天数。
	GroupRetention = "retention"
	// GroupRetentionJob 归档任务，控制统一清理 Job 的启停、间隔与开始时刻。
	GroupRetentionJob = "retentionJob"
	// GroupWebTimeout 管理端 Web 访问超时，含接口超时与会话时长。
	GroupWebTimeout = "webTimeout"
	// GroupEnvVars 租户级全局环境变量，供网关过滤器等引用 ${NAME}。
	GroupEnvVars = "envVars"
)

const (
	minRetentionDays      = 1
	maxRetentionDays      = 3650 // 与网关日志保留上限一致，最长约 10 年
	minRequestSec         = 10
	maxRequestSec         = 600
	minSessionHours       = 1
	maxSessionHours       = 168 // 最长 7 天
	minJobIntervalMinutes = 1
	maxJobIntervalMinutes = 10080 // 最长 7 天，与调度 tick（1 分钟）对齐
	defaultJobIntervalMin = 60    // 与原先默认每小时一致
	defaultSessionHr      = 12
	defaultRequestSec     = 120
)

// RetentionSettings 归档策略。各字段为保留天数，到期由清理任务删除对应数据。
// 网关访问日志仍以实例级 logRetentionDays 为准，GatewayLogDefaultDays 只作用于新建实例。
type RetentionSettings struct {
	AuditLogDays          int `json:"auditLogDays"`
	TaskLogDays           int `json:"taskLogDays"`
	AlertLogDays          int `json:"alertLogDays"`
	ClusterEventDays      int `json:"clusterEventDays"`
	MetricsDays           int `json:"metricsDays"`
	GatewayLogDefaultDays int `json:"gatewayLogDefaultDays"`
}

// RetentionJobSettings 归档任务调度。进程级：启停、间隔、每天开始时刻。
// StartTime 非空时按每个自然日该时刻执行一轮；为空时按 IntervalMinutes 反复执行。
type RetentionJobSettings struct {
	Enabled         bool   `json:"enabled"`
	IntervalMinutes int    `json:"intervalMinutes"`
	StartTime       string `json:"startTime"`
}

// WebTimeoutSettings 管理端访问超时。
// RequestTimeoutSeconds 同时作为 axios 超时和 http.Server 的 Read/Write 超时。
// SessionExpireHours 控制登录会话 TTL。
type WebTimeoutSettings struct {
	RequestTimeoutSeconds int `json:"requestTimeoutSeconds"`
	SessionExpireHours    int `json:"sessionExpireHours"`
}

// DefaultRetention 返回归档策略缺省值，与现有 yaml / 代码默认对齐。
func DefaultRetention() RetentionSettings {
	return RetentionSettings{
		AuditLogDays:          180,
		TaskLogDays:           30,
		AlertLogDays:          7,
		ClusterEventDays:      1,
		MetricsDays:           30,
		GatewayLogDefaultDays: 30,
	}
}

// DefaultRetentionJob 返回归档任务缺省值：开启、每 60 分钟、不限制开始时刻。
func DefaultRetentionJob() RetentionJobSettings {
	return RetentionJobSettings{
		Enabled:         true,
		IntervalMinutes: defaultJobIntervalMin,
		StartTime:       "",
	}
}

// DefaultWebTimeout 返回 Web 超时缺省值。接口超时优先读 web.read_timeout。
func DefaultWebTimeout() WebTimeoutSettings {
	sec := config.GetInt("web.read_timeout", defaultRequestSec)
	if sec < minRequestSec || sec > maxRequestSec {
		sec = defaultRequestSec
	}
	return WebTimeoutSettings{
		RequestTimeoutSeconds: sec,
		SessionExpireHours:    defaultSessionHr,
	}
}

// ParseRetention 解析分组 JSON，缺字段用默认值补齐。
func ParseRetention(content string) RetentionSettings {
	v := DefaultRetention()
	if content == "" {
		return v
	}
	_ = json.Unmarshal([]byte(content), &v)
	return mergeRetention(v)
}

// ParseRetentionJob 解析归档任务 JSON。缺 enabled 时保持默认开启。
func ParseRetentionJob(content string) RetentionJobSettings {
	v := DefaultRetentionJob()
	if content == "" {
		return v
	}
	_ = json.Unmarshal([]byte(content), &v)
	return mergeRetentionJob(v)
}

// ParseWebTimeout 解析 Web 超时 JSON，缺字段用默认值补齐。
func ParseWebTimeout(content string) WebTimeoutSettings {
	v := DefaultWebTimeout()
	if content == "" {
		return v
	}
	_ = json.Unmarshal([]byte(content), &v)
	return mergeWebTimeout(v)
}

// ValidateRetention 校验归档策略各天数是否在允许区间。
func ValidateRetention(v RetentionSettings) error {
	checks := []struct {
		name string
		days int
	}{
		{"审计日志", v.AuditLogDays},
		{"任务执行日志", v.TaskLogDays},
		{"预警日志", v.AlertLogDays},
		{"集群事件", v.ClusterEventDays},
		{"指标数据", v.MetricsDays},
		{"网关日志默认", v.GatewayLogDefaultDays},
	}
	for _, c := range checks {
		if c.days < minRetentionDays || c.days > maxRetentionDays {
			return fmt.Errorf("%s保留天数须在 %d-%d 之间", c.name, minRetentionDays, maxRetentionDays)
		}
	}
	return nil
}

// ValidateRetentionJob 校验归档任务间隔与开始时刻。
func ValidateRetentionJob(v RetentionJobSettings) error {
	if v.IntervalMinutes < minJobIntervalMinutes || v.IntervalMinutes > maxJobIntervalMinutes {
		return fmt.Errorf("归档间隔须在 %d-%d 分钟之间", minJobIntervalMinutes, maxJobIntervalMinutes)
	}
	if v.StartTime != "" && !ValidJobStartTime(v.StartTime) {
		return fmt.Errorf("开始时间格式须为 HH:mm")
	}
	return nil
}

// ValidateWebTimeout 校验 Web 超时时限是否在允许区间。
func ValidateWebTimeout(v WebTimeoutSettings) error {
	if v.RequestTimeoutSeconds < minRequestSec || v.RequestTimeoutSeconds > maxRequestSec {
		return fmt.Errorf("接口超时须在 %d-%d 秒之间", minRequestSec, maxRequestSec)
	}
	if v.SessionExpireHours < minSessionHours || v.SessionExpireHours > maxSessionHours {
		return fmt.Errorf("会话有效期须在 %d-%d 小时之间", minSessionHours, maxSessionHours)
	}
	return nil
}

// KnownGroupCodes 返回当前已实现的分组编码，供加载与校验使用。
func KnownGroupCodes() []string {
	return []string{GroupRetention, GroupRetentionJob, GroupWebTimeout, GroupEnvVars}
}

func mergeRetention(v RetentionSettings) RetentionSettings {
	d := DefaultRetention()
	if v.AuditLogDays <= 0 {
		v.AuditLogDays = d.AuditLogDays
	}
	if v.TaskLogDays <= 0 {
		v.TaskLogDays = d.TaskLogDays
	}
	if v.AlertLogDays <= 0 {
		v.AlertLogDays = d.AlertLogDays
	}
	if v.ClusterEventDays <= 0 {
		v.ClusterEventDays = d.ClusterEventDays
	}
	if v.MetricsDays <= 0 {
		v.MetricsDays = d.MetricsDays
	}
	if v.GatewayLogDefaultDays <= 0 {
		v.GatewayLogDefaultDays = d.GatewayLogDefaultDays
	}
	return v
}

func mergeRetentionJob(v RetentionJobSettings) RetentionJobSettings {
	d := DefaultRetentionJob()
	if v.IntervalMinutes <= 0 {
		v.IntervalMinutes = d.IntervalMinutes
	}
	if v.StartTime != "" && !ValidJobStartTime(v.StartTime) {
		v.StartTime = ""
	}
	return v
}

func mergeWebTimeout(v WebTimeoutSettings) WebTimeoutSettings {
	d := DefaultWebTimeout()
	if v.RequestTimeoutSeconds <= 0 {
		v.RequestTimeoutSeconds = d.RequestTimeoutSeconds
	}
	if v.SessionExpireHours <= 0 {
		v.SessionExpireHours = d.SessionExpireHours
	}
	return v
}

// ValidJobStartTime 判断归档开始时间是否为 HH:mm 或 HH:mm:ss。
func ValidJobStartTime(s string) bool {
	_, _, ok := ParseJobStartClock(s)
	return ok
}

// ParseJobStartClock 解析 HH:mm / HH:mm:ss，失败时 ok 为 false。
func ParseJobStartClock(s string) (hour, minute int, ok bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, 0, false
	}
	for _, layout := range []string{"15:04", "15:04:05"} {
		t, err := time.Parse(layout, s)
		if err == nil {
			return t.Hour(), t.Minute(), true
		}
	}
	return 0, 0, false
}
