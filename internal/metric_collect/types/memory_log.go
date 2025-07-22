package types

import (
	"time"

	metricTypes "gohub/pkg/metric/types"
	"gohub/pkg/utils/random"
)

// MemoryLog 内存采集日志表对应的结构体
type MemoryLog struct {
	// 业务字段
	MetricMemoryLogId string    `json:"metricMemoryLogId" db:"metricMemoryLogId"` // 内存采集日志ID
	TenantId          string    `json:"tenantId" db:"tenantId"`                   // 租户ID
	MetricServerId    string    `json:"metricServerId" db:"metricServerId"`       // 关联服务器ID
	TotalMemory       int64     `json:"totalMemory" db:"totalMemory"`             // 总内存(字节)
	AvailableMemory   int64     `json:"availableMemory" db:"availableMemory"`     // 可用内存(字节)
	UsedMemory        int64     `json:"usedMemory" db:"usedMemory"`               // 已使用内存(字节)
	UsagePercent      float64   `json:"usagePercent" db:"usagePercent"`           // 内存使用率(0-100)
	FreeMemory        int64     `json:"freeMemory" db:"freeMemory"`               // 空闲内存(字节)
	CachedMemory      int64     `json:"cachedMemory" db:"cachedMemory"`           // 缓存内存(字节)
	BuffersMemory     int64     `json:"buffersMemory" db:"buffersMemory"`         // 缓冲区内存(字节)
	SharedMemory      int64     `json:"sharedMemory" db:"sharedMemory"`           // 共享内存(字节)
	SwapTotal         int64     `json:"swapTotal" db:"swapTotal"`                 // 交换区总大小(字节)
	SwapUsed          int64     `json:"swapUsed" db:"swapUsed"`                   // 交换区已使用(字节)
	SwapFree          int64     `json:"swapFree" db:"swapFree"`                   // 交换区空闲(字节)
	SwapUsagePercent  float64   `json:"swapUsagePercent" db:"swapUsagePercent"`   // 交换区使用率(0-100)
	CollectTime       time.Time `json:"collectTime" db:"collectTime"`             // 采集时间

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
func (m *MemoryLog) TableName() string {
	return "HUB_METRIC_MEMORY_LOG"
}

// GetPrimaryKey 获取主键值
func (m *MemoryLog) GetPrimaryKey() (string, string) {
	return m.TenantId, m.MetricMemoryLogId
}

// NewMemoryLogFromMetrics 从MemoryMetrics创建MemoryLog实例
func NewMemoryLogFromMetrics(memoryMetrics *metricTypes.MemoryMetrics, tenantId, serverId, operator string, collectTime time.Time, oprSeqFlag string) *MemoryLog {
	now := time.Now()
	
	return &MemoryLog{
		MetricMemoryLogId: random.Generate32BitRandomString(),
		TenantId:          tenantId,
		MetricServerId:    serverId,
		TotalMemory:       int64(memoryMetrics.Total),
		AvailableMemory:   int64(memoryMetrics.Available),
		UsedMemory:        int64(memoryMetrics.Used),
		UsagePercent:      memoryMetrics.UsagePercent,
		FreeMemory:        int64(memoryMetrics.Free),
		CachedMemory:      int64(memoryMetrics.Cached),
		BuffersMemory:     int64(memoryMetrics.Buffers),
		SharedMemory:      int64(memoryMetrics.Shared),
		SwapTotal:         int64(memoryMetrics.SwapTotal),
		SwapUsed:          int64(memoryMetrics.SwapUsed),
		SwapFree:          int64(memoryMetrics.SwapFree),
		SwapUsagePercent:  memoryMetrics.SwapUsagePercent,
		CollectTime:       collectTime,
		AddTime:           now,
		AddWho:            operator,
		EditTime:          now,
		EditWho:           operator,
		OprSeqFlag:        oprSeqFlag,
		CurrentVersion:    1,
		ActiveFlag:        ActiveFlagYes,
	}
} 