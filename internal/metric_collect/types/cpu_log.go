package types

import (
	"time"

	metricTypes "gateway/pkg/metric/types"
	"gateway/pkg/utils/random"
)

// CpuLog CPU采集日志表对应的结构体
type CpuLog struct {
	// 业务字段
	MetricCpuLogId string    `json:"metricCpuLogId" db:"metricCpuLogId"` // CPU采集日志ID
	TenantId       string    `json:"tenantId" db:"tenantId"`             // 租户ID
	MetricServerId string    `json:"metricServerId" db:"metricServerId"` // 关联服务器ID
	UsagePercent   float64   `json:"usagePercent" db:"usagePercent"`     // CPU使用率(0-100)
	UserPercent    float64   `json:"userPercent" db:"userPercent"`       // 用户态CPU使用率
	SystemPercent  float64   `json:"systemPercent" db:"systemPercent"`   // 系统态CPU使用率
	IdlePercent    float64   `json:"idlePercent" db:"idlePercent"`       // 空闲CPU使用率
	IoWaitPercent  float64   `json:"ioWaitPercent" db:"ioWaitPercent"`   // I/O等待CPU使用率
	IrqPercent     float64   `json:"irqPercent" db:"irqPercent"`         // 中断处理CPU使用率
	SoftIrqPercent float64   `json:"softIrqPercent" db:"softIrqPercent"` // 软中断处理CPU使用率
	CoreCount      int       `json:"coreCount" db:"coreCount"`           // CPU核心数
	LogicalCount   int       `json:"logicalCount" db:"logicalCount"`     // 逻辑CPU数
	LoadAvg1       float64   `json:"loadAvg1" db:"loadAvg1"`             // 1分钟负载平均值
	LoadAvg5       float64   `json:"loadAvg5" db:"loadAvg5"`             // 5分钟负载平均值
	LoadAvg15      float64   `json:"loadAvg15" db:"loadAvg15"`           // 15分钟负载平均值
	CollectTime    time.Time `json:"collectTime" db:"collectTime"`       // 采集时间

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
func (c *CpuLog) TableName() string {
	return "HUB_METRIC_CPU_LOG"
}

// GetPrimaryKey 获取主键值
func (c *CpuLog) GetPrimaryKey() (string, string) {
	return c.TenantId, c.MetricCpuLogId
}

// NewCpuLogFromMetrics 从CPUMetrics创建CpuLog实例
func NewCpuLogFromMetrics(cpuMetrics *metricTypes.CPUMetrics, tenantId, serverId, operator string, collectTime time.Time, oprSeqFlag string) *CpuLog {
	now := time.Now()

	return &CpuLog{
		MetricCpuLogId: random.Generate32BitRandomString(),
		TenantId:       tenantId,
		MetricServerId: serverId,
		UsagePercent:   cpuMetrics.UsagePercent,
		UserPercent:    cpuMetrics.UserPercent,
		SystemPercent:  cpuMetrics.SystemPercent,
		IdlePercent:    cpuMetrics.IdlePercent,
		IoWaitPercent:  cpuMetrics.IOWaitPercent,
		IrqPercent:     cpuMetrics.IrqPercent,
		SoftIrqPercent: cpuMetrics.SoftIrqPercent,
		CoreCount:      cpuMetrics.CoreCount,
		LogicalCount:   cpuMetrics.LogicalCount,
		LoadAvg1:       cpuMetrics.LoadAvg1,
		LoadAvg5:       cpuMetrics.LoadAvg5,
		LoadAvg15:      cpuMetrics.LoadAvg15,
		CollectTime:    collectTime,
		AddTime:        now,
		AddWho:         operator,
		EditTime:       now,
		EditWho:        operator,
		OprSeqFlag:     oprSeqFlag,
		CurrentVersion: 1,
		ActiveFlag:     ActiveFlagYes,
	}
}
