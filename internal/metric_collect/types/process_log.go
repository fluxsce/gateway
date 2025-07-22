package types

import (
	"encoding/json"
	"time"

	metricTypes "gateway/pkg/metric/types"
	"gateway/pkg/utils/random"
)

// ProcessLog 进程信息日志表对应的结构体
type ProcessLog struct {
	// 业务字段
	MetricProcessLogId  string    `json:"metricProcessLogId" db:"metricProcessLogId"`   // 进程信息日志ID
	TenantId            string    `json:"tenantId" db:"tenantId"`                       // 租户ID
	MetricServerId      string    `json:"metricServerId" db:"metricServerId"`           // 关联服务器ID
	ProcessId           int       `json:"processId" db:"processId"`                     // 进程ID
	ParentProcessId     *int      `json:"parentProcessId" db:"parentProcessId"`         // 父进程ID
	ProcessName         string    `json:"processName" db:"processName"`                 // 进程名称
	ProcessStatus       string    `json:"processStatus" db:"processStatus"`             // 进程状态
	CreateTime          time.Time `json:"createTime" db:"createTime"`                   // 进程启动时间
	RunTime             int64     `json:"runTime" db:"runTime"`                         // 进程运行时间(秒)
	MemoryUsage         int64     `json:"memoryUsage" db:"memoryUsage"`                 // 内存使用(字节)
	MemoryPercent       float64   `json:"memoryPercent" db:"memoryPercent"`             // 内存使用率(0-100)
	CpuPercent          float64   `json:"cpuPercent" db:"cpuPercent"`                   // CPU使用率(0-100)
	ThreadCount         int       `json:"threadCount" db:"threadCount"`                 // 线程数
	FileDescriptorCount int       `json:"fileDescriptorCount" db:"fileDescriptorCount"` // 文件句柄数
	CommandLine         *string   `json:"commandLine" db:"commandLine"`                 // 命令行参数，JSON格式
	ExecutablePath      *string   `json:"executablePath" db:"executablePath"`           // 执行路径
	WorkingDirectory    *string   `json:"workingDirectory" db:"workingDirectory"`       // 工作目录
	CollectTime         time.Time `json:"collectTime" db:"collectTime"`                 // 采集时间

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
func (p *ProcessLog) TableName() string {
	return "HUB_METRIC_PROCESS_LOG"
}

// GetPrimaryKey 获取主键值
func (p *ProcessLog) GetPrimaryKey() (string, string) {
	return p.TenantId, p.MetricProcessLogId
}

// getParentPid 辅助函数：处理父进程ID
func getParentPid(ppid int32) *int {
	if ppid == 0 {
		return nil
	}
	pid := int(ppid)
	return &pid
}

// NewProcessLogFromMetrics 从ProcessInfo创建ProcessLog实例
func NewProcessLogFromMetrics(proc *metricTypes.ProcessInfo, tenantId, serverId, operator string, collectTime time.Time, oprSeqFlag string) *ProcessLog {
	// 转换命令行参数为JSON字符串
	cmdlineBytes, _ := json.Marshal(proc.CommandLine)
	cmdlineStr := string(cmdlineBytes)

	// 确保必需字段有默认值
	processName := proc.Name
	if processName == "" {
		processName = "unknown"
	}

	processStatus := proc.Status
	if processStatus == "" {
		processStatus = "unknown"
	}

	// 设置默认创建时间（如果为零值）
	createTime := proc.CreateTime
	if createTime.IsZero() {
		createTime = time.Now()
	}

	now := time.Now()

	return &ProcessLog{
		MetricProcessLogId:  random.Generate32BitRandomString(),
		TenantId:            tenantId,
		MetricServerId:      serverId,
		ProcessId:           int(proc.PID),
		ParentProcessId:     getParentPid(proc.PPID),
		ProcessName:         processName,
		ProcessStatus:       processStatus,
		CreateTime:          createTime,
		RunTime:             int64(proc.RunTime),
		MemoryUsage:         int64(proc.MemoryUsage),
		MemoryPercent:       proc.MemoryPercent,
		CpuPercent:          proc.CPUPercent,
		ThreadCount:         int(proc.ThreadCount),
		FileDescriptorCount: int(proc.FileDescriptorCount),
		CommandLine:         &cmdlineStr,
		ExecutablePath:      &proc.ExecutablePath,
		WorkingDirectory:    &proc.WorkingDirectory,
		CollectTime:         collectTime,
		AddTime:             now,
		AddWho:              operator,
		EditTime:            now,
		EditWho:             operator,
		OprSeqFlag:          oprSeqFlag,
		CurrentVersion:      1,
		ActiveFlag:          ActiveFlagYes,
	}
}
