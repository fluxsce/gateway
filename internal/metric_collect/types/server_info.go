package types

import (
	"encoding/json"
	"time"
)

// ServerInfo 服务器信息表对应的结构体
type ServerInfo struct {
	// 业务字段
	MetricServerId  string    `json:"metricServerId" db:"metricServerId"`   // 服务器ID
	TenantId        string    `json:"tenantId" db:"tenantId"`               // 租户ID
	Hostname        string    `json:"hostname" db:"hostname"`               // 主机名
	OsType          string    `json:"osType" db:"osType"`                   // 操作系统类型
	OsVersion       string    `json:"osVersion" db:"osVersion"`             // 操作系统版本
	KernelVersion   *string   `json:"kernelVersion" db:"kernelVersion"`     // 内核版本
	Architecture    string    `json:"architecture" db:"architecture"`       // 系统架构
	BootTime        time.Time `json:"bootTime" db:"bootTime"`               // 系统启动时间
	IpAddress       *string   `json:"ipAddress" db:"ipAddress"`             // 主IP地址
	MacAddress      *string   `json:"macAddress" db:"macAddress"`           // 主MAC地址
	ServerLocation  *string   `json:"serverLocation" db:"serverLocation"`   // 服务器位置
	ServerType      *string   `json:"serverType" db:"serverType"`           // 服务器类型(physical/virtual/unknown)
	LastUpdateTime  time.Time `json:"lastUpdateTime" db:"lastUpdateTime"`   // 最后更新时间

	// 新增复杂信息字段 (JSON格式存储)
	NetworkInfo  *string `json:"networkInfo" db:"networkInfo"`   // 网络信息详情，JSON格式
	SystemInfo   *string `json:"systemInfo" db:"systemInfo"`     // 系统详细信息，JSON格式
	HardwareInfo *string `json:"hardwareInfo" db:"hardwareInfo"` // 硬件信息，JSON格式

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

// NetworkInfoData 网络信息数据结构
type NetworkInfoData struct {
	PrimaryIP         string   `json:"primaryIP"`         // 主IP地址
	PrimaryMAC        string   `json:"primaryMAC"`        // 主MAC地址
	PrimaryInterface  string   `json:"primaryInterface"`  // 主网络接口
	AllIPs            []string `json:"allIPs"`            // 所有IP地址
	AllMACs           []string `json:"allMACs"`           // 所有MAC地址
	ActiveInterfaces  []string `json:"activeInterfaces"`  // 活动网络接口
}

// SystemInfoData 系统信息数据结构
type SystemInfoData struct {
	Uptime        uint64                 `json:"uptime"`        // 系统运行时间(秒)
	UserCount     uint32                 `json:"userCount"`     // 用户数
	ProcessCount  uint32                 `json:"processCount"`  // 进程数
	LoadAvg       map[string]float64     `json:"loadAvg"`       // 负载平均值
	Temperatures  []TemperatureData      `json:"temperatures"`  // 温度信息
}

// HardwareInfoData 硬件信息数据结构
type HardwareInfoData struct {
	CPU     CPUHardwareInfo     `json:"cpu"`     // CPU信息
	Memory  MemoryHardwareInfo  `json:"memory"`  // 内存信息
	Storage StorageHardwareInfo `json:"storage"` // 存储信息
}

// CPUHardwareInfo CPU硬件信息
type CPUHardwareInfo struct {
	CoreCount    int    `json:"coreCount"`    // 核心数
	LogicalCount int    `json:"logicalCount"` // 逻辑CPU数
	Model        string `json:"model"`        // CPU型号
	Frequency    string `json:"frequency"`    // 频率
}

// MemoryHardwareInfo 内存硬件信息
type MemoryHardwareInfo struct {
	Total uint64 `json:"total"` // 总内存(字节)
	Type  string `json:"type"`  // 内存类型
	Speed string `json:"speed"` // 内存速度
}

// StorageHardwareInfo 存储硬件信息
type StorageHardwareInfo struct {
	TotalDisks    int    `json:"totalDisks"`    // 磁盘总数
	TotalCapacity uint64 `json:"totalCapacity"` // 总容量(字节)
}

// TemperatureData 温度数据
type TemperatureData struct {
	Sensor   string  `json:"sensor"`   // 传感器名称
	Value    float64 `json:"value"`    // 温度值
	High     float64 `json:"high"`     // 高温阈值
	Critical float64 `json:"critical"` // 严重高温阈值
}

// TableName 返回表名
func (s *ServerInfo) TableName() string {
	return "HUB_METRIC_SERVER_INFO"
}

// GetPrimaryKey 获取主键值
func (s *ServerInfo) GetPrimaryKey() (string, string) {
	return s.TenantId, s.MetricServerId
}

// SetNetworkInfo 设置网络信息
func (s *ServerInfo) SetNetworkInfo(data *NetworkInfoData) error {
	if data == nil {
		s.NetworkInfo = nil
		return nil
	}
	
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	
	jsonStr := string(jsonData)
	s.NetworkInfo = &jsonStr
	return nil
}

// GetNetworkInfo 获取网络信息
func (s *ServerInfo) GetNetworkInfo() (*NetworkInfoData, error) {
	if s.NetworkInfo == nil || *s.NetworkInfo == "" {
		return nil, nil
	}
	
	var data NetworkInfoData
	err := json.Unmarshal([]byte(*s.NetworkInfo), &data)
	if err != nil {
		return nil, err
	}
	
	return &data, nil
}

// SetSystemInfo 设置系统信息
func (s *ServerInfo) SetSystemInfo(data *SystemInfoData) error {
	if data == nil {
		s.SystemInfo = nil
		return nil
	}
	
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	
	jsonStr := string(jsonData)
	s.SystemInfo = &jsonStr
	return nil
}

// GetSystemInfo 获取系统信息
func (s *ServerInfo) GetSystemInfo() (*SystemInfoData, error) {
	if s.SystemInfo == nil || *s.SystemInfo == "" {
		return nil, nil
	}
	
	var data SystemInfoData
	err := json.Unmarshal([]byte(*s.SystemInfo), &data)
	if err != nil {
		return nil, err
	}
	
	return &data, nil
}

// SetHardwareInfo 设置硬件信息
func (s *ServerInfo) SetHardwareInfo(data *HardwareInfoData) error {
	if data == nil {
		s.HardwareInfo = nil
		return nil
	}
	
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	
	jsonStr := string(jsonData)
	s.HardwareInfo = &jsonStr
	return nil
}

// GetHardwareInfo 获取硬件信息
func (s *ServerInfo) GetHardwareInfo() (*HardwareInfoData, error) {
	if s.HardwareInfo == nil || *s.HardwareInfo == "" {
		return nil, nil
	}
	
	var data HardwareInfoData
	err := json.Unmarshal([]byte(*s.HardwareInfo), &data)
	if err != nil {
		return nil, err
	}
	
	return &data, nil
} 