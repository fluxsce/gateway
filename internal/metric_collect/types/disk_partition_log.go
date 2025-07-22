package types

import (
	"time"

	metricTypes "gateway/pkg/metric/types"
	"gateway/pkg/utils/random"
)

// DiskPartitionLog 磁盘分区日志表对应的结构体
type DiskPartitionLog struct {
	// 业务字段
	MetricDiskPartitionLogId string    `json:"metricDiskPartitionLogId" db:"metricDiskPartitionLogId"` // 磁盘分区日志ID
	TenantId                 string    `json:"tenantId" db:"tenantId"`                                 // 租户ID
	MetricServerId           string    `json:"metricServerId" db:"metricServerId"`                     // 关联服务器ID
	DeviceName               string    `json:"deviceName" db:"deviceName"`                             // 设备名称
	MountPoint               string    `json:"mountPoint" db:"mountPoint"`                             // 挂载点
	FileSystem               string    `json:"fileSystem" db:"fileSystem"`                             // 文件系统类型
	TotalSpace               int64     `json:"totalSpace" db:"totalSpace"`                             // 总大小(字节)
	UsedSpace                int64     `json:"usedSpace" db:"usedSpace"`                               // 已使用(字节)
	FreeSpace                int64     `json:"freeSpace" db:"freeSpace"`                               // 可用(字节)
	UsagePercent             float64   `json:"usagePercent" db:"usagePercent"`                         // 使用率(0-100)
	InodesTotal              int64     `json:"inodesTotal" db:"inodesTotal"`                           // inode总数
	InodesUsed               int64     `json:"inodesUsed" db:"inodesUsed"`                             // inode已使用
	InodesFree               int64     `json:"inodesFree" db:"inodesFree"`                             // inode空闲
	InodesUsagePercent       float64   `json:"inodesUsagePercent" db:"inodesUsagePercent"`             // inode使用率(0-100)
	CollectTime              time.Time `json:"collectTime" db:"collectTime"`                           // 采集时间

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
func (d *DiskPartitionLog) TableName() string {
	return "HUB_METRIC_DISK_PART_LOG"
}

// GetPrimaryKey 获取主键值
func (d *DiskPartitionLog) GetPrimaryKey() (string, string) {
	return d.TenantId, d.MetricDiskPartitionLogId
}

// NewDiskPartitionLogFromMetrics 从DiskPartition创建DiskPartitionLog实例
func NewDiskPartitionLogFromMetrics(partition *metricTypes.DiskPartition, tenantId, serverId, operator string, collectTime time.Time, oprSeqFlag string, index int) *DiskPartitionLog {
	now := time.Now()

	return &DiskPartitionLog{
		MetricDiskPartitionLogId: random.Generate32BitRandomString(),
		TenantId:                 tenantId,
		MetricServerId:           serverId,
		DeviceName:               partition.Device,
		MountPoint:               partition.MountPoint,
		FileSystem:               partition.FileSystem,
		TotalSpace:               int64(partition.Total),
		UsedSpace:                int64(partition.Used),
		FreeSpace:                int64(partition.Free),
		UsagePercent:             partition.UsagePercent,
		InodesTotal:              int64(partition.InodesTotal),
		InodesUsed:               int64(partition.InodesUsed),
		InodesFree:               int64(partition.InodesFree),
		InodesUsagePercent:       partition.InodesUsagePercent,
		CollectTime:              collectTime,
		AddTime:                  now,
		AddWho:                   operator,
		EditTime:                 now,
		EditWho:                  operator,
		OprSeqFlag:               oprSeqFlag,
		CurrentVersion:           1,
		ActiveFlag:               ActiveFlagYes,
	}
}

// NewDiskPartitionLogsFromMetrics 从DiskMetrics批量创建DiskPartitionLog实例
func NewDiskPartitionLogsFromMetrics(diskMetrics *metricTypes.DiskMetrics, tenantId, serverId, operator string, collectTime time.Time, oprSeqFlag string) []*DiskPartitionLog {
	var logs []*DiskPartitionLog

	for i, partition := range diskMetrics.Partitions {
		log := NewDiskPartitionLogFromMetrics(&partition, tenantId, serverId, operator, collectTime, oprSeqFlag, i)
		logs = append(logs, log)
	}

	return logs
}
