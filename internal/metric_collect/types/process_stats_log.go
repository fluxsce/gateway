package types

import (
	"time"

	metricTypes "gohub/pkg/metric/types"
	"gohub/pkg/utils/random"
)

// ProcessStatsLog 进程统计日志表对应的结构体
type ProcessStatsLog struct {
	// 业务字段
	MetricProcessStatsLogId string    `json:"metricProcessStatsLogId" db:"metricProcessStatsLogId"` // 进程统计日志ID
	TenantId                string    `json:"tenantId" db:"tenantId"`                               // 租户ID
	MetricServerId          string    `json:"metricServerId" db:"metricServerId"`                   // 关联服务器ID
	RunningCount            int       `json:"runningCount" db:"runningCount"`                       // 运行中进程数
	SleepingCount           int       `json:"sleepingCount" db:"sleepingCount"`                     // 睡眠中进程数
	StoppedCount            int       `json:"stoppedCount" db:"stoppedCount"`                       // 停止的进程数
	ZombieCount             int       `json:"zombieCount" db:"zombieCount"`                         // 僵尸进程数
	TotalCount              int       `json:"totalCount" db:"totalCount"`                           // 总进程数
	CollectTime             time.Time `json:"collectTime" db:"collectTime"`                         // 采集时间

	// 通用字段
	AddTime        time.Time `json:"addTime" db:"addTime"`               // 创建时间
	AddWho         string    `json:"addWho" db:"addWho"`                 // 创建人ID
	EditTime       time.Time `json:"editTime" db:"editTime"`             // 最后修改时间
	EditWho        string    `json:"editWho" db:"editWho"`               // 最后修改人ID
	OprSeqFlag     string    `json:"oprSeqFlag" db:"oprSeqFlag"`         // 操作序列标识
	CurrentVersion int       `json:"currentVersion" db:"currentVersion"` // 当前版本号
	ActiveFlag     string    `json:"activeFlag" db:"activeFlag"`         // 活动状态标记
	NoteText       *string   `json:"noteText" db:"noteText"`             // 备注信息
	ExtProperty    *string   `json:"extProperty" db:"extProperty"`       // 扩展属性，JSON格式
	Reserved1      *string   `json:"reserved1" db:"reserved1"`           // 预留字段1
	Reserved2      *string   `json:"reserved2" db:"reserved2"`           // 预留字段2
	Reserved3      *string   `json:"reserved3" db:"reserved3"`           // 预留字段3
	Reserved4      *string   `json:"reserved4" db:"reserved4"`           // 预留字段4
	Reserved5      *string   `json:"reserved5" db:"reserved5"`           // 预留字段5
	Reserved6      *string   `json:"reserved6" db:"reserved6"`           // 预留字段6
	Reserved7      *string   `json:"reserved7" db:"reserved7"`           // 预留字段7
	Reserved8      *string   `json:"reserved8" db:"reserved8"`           // 预留字段8
	Reserved9      *string   `json:"reserved9" db:"reserved9"`           // 预留字段9
	Reserved10     *string   `json:"reserved10" db:"reserved10"`         // 预留字段10
}

// TableName 返回表名
func (p *ProcessStatsLog) TableName() string {
	return "HUB_METRIC_PROCSTAT_LOG"
}

// GetPrimaryKey 获取主键值
func (p *ProcessStatsLog) GetPrimaryKey() (string, string) {
	return p.TenantId, p.MetricProcessStatsLogId
}

// NewProcessStatsLogFromMetrics 从ProcessSystemStats创建ProcessStatsLog实例
func NewProcessStatsLogFromMetrics(stats *metricTypes.ProcessSystemStats, tenantId, serverId, operator string, collectTime time.Time, oprSeqFlag string) *ProcessStatsLog {
	now := time.Now()
	
	return &ProcessStatsLog{
		MetricProcessStatsLogId: random.Generate32BitRandomString(),
		TenantId:                tenantId,
		MetricServerId:          serverId,
		RunningCount:            int(stats.Running),
		SleepingCount:           int(stats.Sleeping),
		StoppedCount:            int(stats.Stopped),
		ZombieCount:             int(stats.Zombie),
		TotalCount:              int(stats.Total),
		CollectTime:             collectTime,
		AddTime:                 now,
		AddWho:                  operator,
		EditTime:                now,
		EditWho:                 operator,
		OprSeqFlag:              oprSeqFlag,
		CurrentVersion:          1,
		ActiveFlag:              ActiveFlagYes,
	}
} 