package types

import (
	"time"

	metricTypes "gohub/pkg/metric/types"
	"gohub/pkg/utils/random"
)

// DiskIoLog 磁盘IO日志表对应的结构体
type DiskIoLog struct {
	// 业务字段
	MetricDiskIoLogId string    `json:"metricDiskIoLogId" db:"metricDiskIoLogId"` // 磁盘IO日志ID
	TenantId          string    `json:"tenantId" db:"tenantId"`                   // 租户ID
	MetricServerId    string    `json:"metricServerId" db:"metricServerId"`       // 关联服务器ID
	DeviceName        string    `json:"deviceName" db:"deviceName"`               // 设备名称
	ReadCount         int64     `json:"readCount" db:"readCount"`                 // 读取次数
	WriteCount        int64     `json:"writeCount" db:"writeCount"`               // 写入次数
	ReadBytes         int64     `json:"readBytes" db:"readBytes"`                 // 读取字节数
	WriteBytes        int64     `json:"writeBytes" db:"writeBytes"`               // 写入字节数
	ReadTime          int64     `json:"readTime" db:"readTime"`                   // 读取时间(毫秒)
	WriteTime         int64     `json:"writeTime" db:"writeTime"`                 // 写入时间(毫秒)
	IoInProgress      int64     `json:"ioInProgress" db:"ioInProgress"`           // IO进行中数量
	IoTime            int64     `json:"ioTime" db:"ioTime"`                       // IO时间(毫秒)
	ReadRate          float64   `json:"readRate" db:"readRate"`                   // 读取速率(字节/秒)
	WriteRate         float64   `json:"writeRate" db:"writeRate"`                 // 写入速率(字节/秒)
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
func (d *DiskIoLog) TableName() string {
	return "HUB_METRIC_DISK_IO_LOG"
}

// GetPrimaryKey 获取主键值
func (d *DiskIoLog) GetPrimaryKey() (string, string) {
	return d.TenantId, d.MetricDiskIoLogId
}

// NewDiskIoLogFromMetrics 从DiskIOStats创建DiskIoLog实例
func NewDiskIoLogFromMetrics(ioStat *metricTypes.DiskIOStats, tenantId, serverId, operator string, collectTime time.Time, oprSeqFlag string, index int) *DiskIoLog {
	now := time.Now()
	
	return &DiskIoLog{
		MetricDiskIoLogId: random.Generate32BitRandomString(),
		TenantId:          tenantId,
		MetricServerId:    serverId,
		DeviceName:        ioStat.Device,
		ReadCount:         int64(ioStat.ReadCount),
		WriteCount:        int64(ioStat.WriteCount),
		ReadBytes:         int64(ioStat.ReadBytes),
		WriteBytes:        int64(ioStat.WriteBytes),
		ReadTime:          int64(ioStat.ReadTime),
		WriteTime:         int64(ioStat.WriteTime),
		IoInProgress:      int64(ioStat.IOInProgress),
		IoTime:            int64(ioStat.IOTime),
		ReadRate:          ioStat.ReadRate,
		WriteRate:         ioStat.WriteRate,
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

// NewDiskIoLogsFromMetrics 从DiskMetrics批量创建DiskIoLog实例
func NewDiskIoLogsFromMetrics(diskMetrics *metricTypes.DiskMetrics, tenantId, serverId, operator string, collectTime time.Time, oprSeqFlag string) []*DiskIoLog {
	var logs []*DiskIoLog
	
	for i, ioStat := range diskMetrics.IOStats {
		log := NewDiskIoLogFromMetrics(&ioStat, tenantId, serverId, operator, collectTime, oprSeqFlag, i)
		logs = append(logs, log)
	}
	
	return logs
} 