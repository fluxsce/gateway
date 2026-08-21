package retention

import (
	"time"

	"gateway/pkg/config"
	"gateway/pkg/syssetting"
)

// Dataset 一类可过期数据的声明。调度删除和以后导出转冷共用这份清单，模块不再各自写 DELETE。
type Dataset struct {
	Name    string
	Table   string
	TimeCol string
	Cutoff  func(tenantId string) (time.Time, bool)
}

// DaysCutoff 按保留天数计算截止时间，天数小于等于 0 则跳过。
func DaysCutoff(daysFor func(string) int) func(string) (time.Time, bool) {
	return func(tenantId string) (time.Time, bool) {
		if daysFor == nil {
			return time.Time{}, false
		}
		days := daysFor(tenantId)
		if days <= 0 {
			return time.Time{}, false
		}
		return time.Now().AddDate(0, 0, -days), true
	}
}

// Datasets 返回当前支持硬删的全部数据集。以后加 archive 动作时只改执行器，不改各业务模块。
func Datasets() []Dataset {
	list := []Dataset{
		{
			Name:    "audit-log",
			Table:   "HUB_AUTH_AUDIT_LOG",
			TimeCol: "addTime",
			Cutoff: DaysCutoff(func(tenantId string) int {
				return syssetting.GetRetention(tenantId).AuditLogDays
			}),
		},
		{
			Name:    "task-log",
			Table:   "HUB_TIMER_EXECUTION_LOG",
			TimeCol: "addTime",
			Cutoff: DaysCutoff(func(tenantId string) int {
				return syssetting.GetRetention(tenantId).TaskLogDays
			}),
		},
		{
			Name:    "alert-log",
			Table:   "HUB_ALERT_LOG",
			TimeCol: "alertTimestamp",
			Cutoff:  alertCutoff,
		},
	}
	if config.GetBool(config.CLUSTER_CLEANUP_ENABLED, true) {
		list = append(list,
			Dataset{
				Name:    "cluster-event",
				Table:   "HUB_CLUSTER_EVENT",
				TimeCol: "eventTime",
				Cutoff:  clusterEventCutoff,
			},
			Dataset{
				Name:    "cluster-ack",
				Table:   "HUB_CLUSTER_EVENT_ACK",
				TimeCol: "addTime",
				// 与集群事件同一保留天数，避免 ACK 先被清掉导致事件被重复消费
				Cutoff: clusterEventCutoff,
			},
		)
	}
	if config.GetBool("app.metrics.storage.retention.enabled", true) {
		metricsCutoff := DaysCutoff(metricsDays)
		for _, item := range metricDatasets {
			list = append(list, Dataset{
				Name:    item.name,
				Table:   item.table,
				TimeCol: "collectTime",
				Cutoff:  metricsCutoff,
			})
		}
	}
	return list
}

var metricDatasets = []struct {
	name  string
	table string
}{
	{name: "metrics-cpu", table: "HUB_METRIC_CPU_LOG"},
	{name: "metrics-memory", table: "HUB_METRIC_MEMORY_LOG"},
	{name: "metrics-disk-part", table: "HUB_METRIC_DISK_PART_LOG"},
	{name: "metrics-disk-io", table: "HUB_METRIC_DISK_IO_LOG"},
	{name: "metrics-network", table: "HUB_METRIC_NETWORK_LOG"},
	{name: "metrics-process", table: "HUB_METRIC_PROCESS_LOG"},
	{name: "metrics-procstat", table: "HUB_METRIC_PROCSTAT_LOG"},
	{name: "metrics-temp", table: "HUB_METRIC_TEMP_LOG"},
}

func alertCutoff(tenantId string) (time.Time, bool) {
	days := syssetting.GetRetention(tenantId).AlertLogDays
	if days > 0 {
		return time.Now().AddDate(0, 0, -days), true
	}
	hours := config.GetInt(config.ALERT_LOG_RETENTION_HOURS, 168)
	if hours <= 0 {
		return time.Time{}, false
	}
	return time.Now().Add(-time.Duration(hours) * time.Hour), true
}

func clusterEventCutoff(tenantId string) (time.Time, bool) {
	days := syssetting.GetRetention(tenantId).ClusterEventDays
	if days > 0 {
		return time.Now().AddDate(0, 0, -days), true
	}
	hours := config.GetInt(config.CLUSTER_EVENT_EXPIRE_HOURS, 24)
	if hours <= 0 {
		return time.Time{}, false
	}
	return time.Now().Add(-time.Duration(hours) * time.Hour), true
}

func metricsDays(tenantId string) int {
	days := syssetting.GetRetention(tenantId).MetricsDays
	if days > 0 {
		return days
	}
	return config.GetInt("app.metrics.storage.retention.keep_days", 30)
}
