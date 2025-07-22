package types

import "time"

// CPUMetrics CPU指标
type CPUMetrics struct {
	// CPU使用率 (0-100)
	UsagePercent float64 `json:"usage_percent"`
	// 用户态CPU使用率
	UserPercent float64 `json:"user_percent"`
	// 系统态CPU使用率
	SystemPercent float64 `json:"system_percent"`
	// 空闲CPU使用率
	IdlePercent float64 `json:"idle_percent"`
	// I/O等待CPU使用率
	IOWaitPercent float64 `json:"io_wait_percent"`
	// 中断处理CPU使用率
	IrqPercent float64 `json:"irq_percent"`
	// 软中断处理CPU使用率
	SoftIrqPercent float64 `json:"soft_irq_percent"`
	// CPU核心数
	CoreCount int `json:"core_count"`
	// 逻辑CPU数
	LogicalCount int `json:"logical_count"`
	// 1分钟负载平均值
	LoadAvg1 float64 `json:"load_avg_1"`
	// 5分钟负载平均值
	LoadAvg5 float64 `json:"load_avg_5"`
	// 15分钟负载平均值
	LoadAvg15 float64 `json:"load_avg_15"`
	// 采集时间
	CollectTime time.Time `json:"collect_time"`
}

// MemoryMetrics 内存指标
type MemoryMetrics struct {
	// 总内存 (字节)
	Total uint64 `json:"total"`
	// 可用内存 (字节)
	Available uint64 `json:"available"`
	// 已使用内存 (字节)
	Used uint64 `json:"used"`
	// 内存使用率 (0-100)
	UsagePercent float64 `json:"usage_percent"`
	// 空闲内存 (字节)
	Free uint64 `json:"free"`
	// 缓存内存 (字节)
	Cached uint64 `json:"cached"`
	// 缓冲区内存 (字节)
	Buffers uint64 `json:"buffers"`
	// 共享内存 (字节)
	Shared uint64 `json:"shared"`
	// 交换区总大小 (字节)
	SwapTotal uint64 `json:"swap_total"`
	// 交换区已使用 (字节)
	SwapUsed uint64 `json:"swap_used"`
	// 交换区空闲 (字节)
	SwapFree uint64 `json:"swap_free"`
	// 交换区使用率 (0-100)
	SwapUsagePercent float64 `json:"swap_usage_percent"`
	// 采集时间
	CollectTime time.Time `json:"collect_time"`
}

// DiskUsageInfo 磁盘使用信息结构体
// 包含磁盘空间和 inode 使用情况的详细信息
type DiskUsageInfo struct {
	Total        uint64  `json:"total"`         // 总空间大小（字节）
	Used         uint64  `json:"used"`          // 已使用空间大小（字节）
	Free         uint64  `json:"free"`          // 可用空间大小（字节）
	UsagePercent float64 `json:"usage_percent"` // 使用率百分比 (0-100)
	
	// inode 信息 (主要用于 Unix 系统)
	InodesTotal        uint64  `json:"inodes_total"`         // 总 inode 数量
	InodesUsed         uint64  `json:"inodes_used"`          // 已使用 inode 数量
	InodesFree         uint64  `json:"inodes_free"`          // 可用 inode 数量
	InodesUsagePercent float64 `json:"inodes_usage_percent"` // inode 使用率百分比 (0-100)
}

// DiskMetrics 磁盘指标
type DiskMetrics struct {
	// 磁盘分区列表
	Partitions []DiskPartition `json:"partitions"`
	// 磁盘IO统计
	IOStats []DiskIOStats `json:"io_stats"`
	// 采集时间
	CollectTime time.Time `json:"collect_time"`
}

// DiskPartition 磁盘分区信息
type DiskPartition struct {
	// 设备名称
	Device string `json:"device"`
	// 挂载点
	MountPoint string `json:"mount_point"`
	// 文件系统类型
	FileSystem string `json:"file_system"`
	// 总大小 (字节)
	Total uint64 `json:"total"`
	// 已使用 (字节)
	Used uint64 `json:"used"`
	// 可用 (字节)
	Free uint64 `json:"free"`
	// 使用率 (0-100)
	UsagePercent float64 `json:"usage_percent"`
	// inode总数
	InodesTotal uint64 `json:"inodes_total"`
	// inode已使用
	InodesUsed uint64 `json:"inodes_used"`
	// inode空闲
	InodesFree uint64 `json:"inodes_free"`
	// inode使用率 (0-100)
	InodesUsagePercent float64 `json:"inodes_usage_percent"`
}

// DiskIOStats 磁盘IO统计
type DiskIOStats struct {
	// 设备名称
	Device string `json:"device"`
	// 读取次数
	ReadCount uint64 `json:"read_count"`
	// 写入次数
	WriteCount uint64 `json:"write_count"`
	// 读取字节数
	ReadBytes uint64 `json:"read_bytes"`
	// 写入字节数
	WriteBytes uint64 `json:"write_bytes"`
	// 读取时间 (毫秒)
	ReadTime uint64 `json:"read_time"`
	// 写入时间 (毫秒)
	WriteTime uint64 `json:"write_time"`
	// IO进行中数量
	IOInProgress uint64 `json:"io_in_progress"`
	// IO时间 (毫秒)
	IOTime uint64 `json:"io_time"`
	// 读取速率 (字节/秒)
	ReadRate float64 `json:"read_rate"`
	// 写入速率 (字节/秒)
	WriteRate float64 `json:"write_rate"`
	// 上次采集时间
	LastCollectTime time.Time `json:"last_collect_time"`
}

// NetworkMetrics 网络指标
type NetworkMetrics struct {
	// 网络接口列表
	Interfaces []NetworkInterface `json:"interfaces"`
	// 采集时间
	CollectTime time.Time `json:"collect_time"`
}

// NetworkInterface 网络接口信息
type NetworkInterface struct {
	// 接口名称
	Name string `json:"name"`
	// MAC地址
	HardwareAddr string `json:"hardware_addr"`
	// IP地址列表
	IPAddresses []string `json:"ip_addresses"`
	// 接口状态
	Status string `json:"status"`
	// 接口类型
	Type string `json:"type"`
	// 接收字节数
	BytesReceived uint64 `json:"bytes_received"`
	// 发送字节数
	BytesSent uint64 `json:"bytes_sent"`
	// 接收包数
	PacketsReceived uint64 `json:"packets_received"`
	// 发送包数
	PacketsSent uint64 `json:"packets_sent"`
	// 接收错误数
	ErrorsReceived uint64 `json:"errors_received"`
	// 发送错误数
	ErrorsSent uint64 `json:"errors_sent"`
	// 接收丢包数
	DroppedReceived uint64 `json:"dropped_received"`
	// 发送丢包数
	DroppedSent uint64 `json:"dropped_sent"`
	// 接收速率 (字节/秒)
	ReceiveRate float64 `json:"receive_rate"`
	// 发送速率 (字节/秒)
	SendRate float64 `json:"send_rate"`
	// 上次采集时间
	LastCollectTime time.Time `json:"last_collect_time"`
}

// SystemMetrics 系统信息指标
type SystemMetrics struct {
	// 主机名
	Hostname string `json:"hostname"`
	// 操作系统类型
	OS string `json:"os"`
	// 操作系统版本
	OSVersion string `json:"os_version"`
	// 内核版本
	KernelVersion string `json:"kernel_version"`
	// 架构
	Architecture string `json:"architecture"`
	// 系统启动时间
	BootTime time.Time `json:"boot_time"`
	// 系统运行时间 (秒)
	Uptime uint64 `json:"uptime"`
	// 用户数
	UserCount uint32 `json:"user_count"`
	// 进程数
	ProcessCount uint32 `json:"process_count"`
	// 温度信息
	Temperature []TemperatureInfo `json:"temperature"`
	// 网络信息
	NetworkInfo *SystemNetworkInfo `json:"network_info"`
	// 服务器类型
	ServerType string `json:"server_type"`
	// 采集时间
	CollectTime time.Time `json:"collect_time"`
}

// SystemNetworkInfo 系统网络信息
type SystemNetworkInfo struct {
	// 主IP地址
	PrimaryIP string `json:"primary_ip"`
	// 所有IP地址列表
	IPAddresses []string `json:"ip_addresses"`
	// 主MAC地址
	PrimaryMAC string `json:"primary_mac"`
	// 所有MAC地址列表
	MACAddresses []string `json:"mac_addresses"`
	// 主网络接口名称
	PrimaryInterface string `json:"primary_interface"`
	// 活动网络接口列表
	ActiveInterfaces []string `json:"active_interfaces"`
}

// TemperatureInfo 温度信息
type TemperatureInfo struct {
	// 传感器名称
	SensorName string `json:"sensor_name"`
	// 温度值 (摄氏度)
	Temperature float64 `json:"temperature"`
	// 高温阈值
	High float64 `json:"high"`
	// 严重高温阈值
	Critical float64 `json:"critical"`
}

// ProcessMetrics 进程信息指标
type ProcessMetrics struct {
	// 当前进程信息
	CurrentProcess *ProcessInfo `json:"current_process"`
	// 系统进程统计
	SystemProcesses *ProcessSystemStats `json:"system_processes"`
	// 采集时间
	CollectTime time.Time `json:"collect_time"`
}

// ProcessInfo 进程信息
type ProcessInfo struct {
	// 进程ID
	PID int32 `json:"pid"`
	// 父进程ID
	PPID int32 `json:"ppid"`
	// 进程名称
	Name string `json:"name"`
	// 进程状态
	Status string `json:"status"`
	// 进程启动时间
	CreateTime time.Time `json:"create_time"`
	// 进程运行时间 (秒)
	RunTime uint64 `json:"run_time"`
	// 内存使用 (字节)
	MemoryUsage uint64 `json:"memory_usage"`
	// 虚拟内存使用 (字节)
	VirtualMemory uint64 `json:"virtual_memory"`
	// 交换内存使用 (字节)
	SwapMemory uint64 `json:"swap_memory"`
	// 共享内存使用 (字节)
	SharedMemory uint64 `json:"shared_memory"`
	// 内存使用率 (0-100)
	MemoryPercent float64 `json:"memory_percent"`
	// CPU使用率 (0-100)
	CPUPercent float64 `json:"cpu_percent"`
	// CPU时间统计
	CPUTimes *ProcessCPUTimes `json:"cpu_times"`
	// IO统计
	IOStats *ProcessIOStats `json:"io_stats"`
	// 网络连接统计
	NetworkStats *ProcessNetworkStats `json:"network_stats"`
	// 线程数
	ThreadCount int32 `json:"thread_count"`
	// 文件句柄数
	FileDescriptorCount int32 `json:"file_descriptor_count"`
	// 命令行参数
	CommandLine []string `json:"command_line"`
	// 执行路径
	ExecutablePath string `json:"executable_path"`
	// 工作目录
	WorkingDirectory string `json:"working_directory"`
	// 环境变量
	Environment []string `json:"environment"`
	// 子进程列表
	Children []ProcessBasicInfo `json:"children"`
	// 进程优先级
	Nice int32 `json:"nice"`
	// IO优先级
	IONice int32 `json:"io_nice"`
}

// ProcessCPUTimes CPU时间统计
type ProcessCPUTimes struct {
	// 用户态CPU时间
	User float64 `json:"user"`
	// 系统态CPU时间
	System float64 `json:"system"`
	// 空闲时间
	Idle float64 `json:"idle"`
	// IO等待时间
	IOWait float64 `json:"io_wait"`
	// 硬中断时间
	IRQ float64 `json:"irq"`
	// 软中断时间
	SoftIRQ float64 `json:"soft_irq"`
	// 虚拟化环境中被偷走的时间
	Stolen float64 `json:"stolen"`
	// 运行虚拟机的时间
	Guest float64 `json:"guest"`
	// 运行低优先级虚拟机的时间
	GuestNice float64 `json:"guest_nice"`
}

// ProcessIOStats IO统计
type ProcessIOStats struct {
	// 读操作次数
	ReadCount uint64 `json:"read_count"`
	// 写操作次数
	WriteCount uint64 `json:"write_count"`
	// 读取的字节数
	ReadBytes uint64 `json:"read_bytes"`
	// 写入的字节数
	WriteBytes uint64 `json:"write_bytes"`
}

// ProcessNetworkStats 网络连接统计
type ProcessNetworkStats struct {
	// 总连接数
	ConnectionCount uint32 `json:"connection_count"`
	// 按状态分类的连接数
	ConnectionsByState map[string]uint32 `json:"connections_by_state"`
}

// ProcessBasicInfo 进程基本信息
type ProcessBasicInfo struct {
	// 进程ID
	PID int32 `json:"pid"`
	// 进程名称
	Name string `json:"name"`
}

// ProcessSystemStats 系统进程统计
type ProcessSystemStats struct {
	// 运行中进程数
	Running uint32 `json:"running"`
	// 睡眠中进程数
	Sleeping uint32 `json:"sleeping"`
	// 停止的进程数
	Stopped uint32 `json:"stopped"`
	// 僵尸进程数
	Zombie uint32 `json:"zombie"`
	// 总进程数
	Total uint32 `json:"total"`
}

// AllMetrics 所有指标的集合
type AllMetrics struct {
	// CPU指标
	CPU *CPUMetrics `json:"cpu,omitempty"`
	// 内存指标
	Memory *MemoryMetrics `json:"memory,omitempty"`
	// 磁盘指标
	Disk *DiskMetrics `json:"disk,omitempty"`
	// 网络指标
	Network *NetworkMetrics `json:"network,omitempty"`
	// 系统指标
	System *SystemMetrics `json:"system,omitempty"`
	// 进程指标
	Process *ProcessMetrics `json:"process,omitempty"`
	// 采集时间
	CollectTime time.Time `json:"collect_time"`
} 