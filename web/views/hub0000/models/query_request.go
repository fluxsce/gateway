package models

import (
	"fmt"
	"gateway/pkg/utils/ctime"
	"strings"
	"time"
)

// BaseQueryRequest 基础查询请求
type BaseQueryRequest struct {
	TenantId        string  `json:"tenantId" form:"tenantId"`               // 租户ID
	MetricServerId  *string `json:"metricServerId" form:"metricServerId"`   // 服务器ID (对应数据库中的metricServerId字段)
	StartTime       string  `json:"startTime" form:"startTime"`             // 开始时间字符串
	EndTime         string  `json:"endTime" form:"endTime"`                 // 结束时间字符串
	Page            int     `json:"page" form:"page"`                       // 页码
	PageSize        int     `json:"pageSize" form:"pageSize"`               // 每页数量
	OrderBy         string  `json:"orderBy" form:"orderBy"`                 // 排序字段
	OrderType       string  `json:"orderType" form:"orderType"`             // 排序类型(ASC/DESC)
}

// GetStartTime 获取解析后的开始时间
func (req *BaseQueryRequest) GetStartTime() (*time.Time, error) {
	if req.StartTime == "" {
		return nil, nil
	}

	parsedTime, err := ctime.ParseTimeString(req.StartTime)
	if err != nil {
		return nil, fmt.Errorf("开始时间格式错误: %w", err)
	}

	return &parsedTime, nil
}

// GetEndTime 获取解析后的结束时间
func (req *BaseQueryRequest) GetEndTime() (*time.Time, error) {
	if req.EndTime == "" {
		return nil, nil
	}

	parsedTime, err := ctime.ParseTimeString(req.EndTime)
	if err != nil {
		return nil, fmt.Errorf("结束时间格式错误: %w", err)
	}

	return &parsedTime, nil
}

// ValidateTimeFields 验证时间字段格式
func (req *BaseQueryRequest) ValidateTimeFields() error {
	if req.StartTime != "" {
		if _, err := req.GetStartTime(); err != nil {
			return err
		}
	}

	if req.EndTime != "" {
		if _, err := req.GetEndTime(); err != nil {
			return err
		}
	}

	return nil
}

// GetFormattedTimeRange 获取格式化的时间范围字符串
func (req *BaseQueryRequest) GetFormattedTimeRange() string {
	var timeRange []string

	if req.StartTime != "" {
		timeRange = append(timeRange, "开始时间: "+req.StartTime)
	}

	if req.EndTime != "" {
		timeRange = append(timeRange, "结束时间: "+req.EndTime)
	}

	if len(timeRange) == 0 {
		return "无时间限制"
	}

	return strings.Join(timeRange, ", ")
}

// ServerInfoQueryRequest 服务器信息查询请求
type ServerInfoQueryRequest struct {
	BaseQueryRequest
	Hostname     *string `json:"hostname" form:"hostname"`         // 主机名
	ServerType   *string `json:"serverType" form:"serverType"`     // 服务器类型
	OsType       *string `json:"osType" form:"osType"`             // 操作系统类型
	Architecture *string `json:"architecture" form:"architecture"` // 系统架构
	ActiveFlag   *string `json:"activeFlag" form:"activeFlag"`     // 活动状态
}

// CpuLogQueryRequest CPU性能日志查询请求
type CpuLogQueryRequest struct {
	BaseQueryRequest
	CpuCore         *string  `json:"cpuCore" form:"cpuCore"`                 // CPU核心
	MinUsagePercent *float64 `json:"minUsagePercent" form:"minUsagePercent"` // 最小使用率
	MaxUsagePercent *float64 `json:"maxUsagePercent" form:"maxUsagePercent"` // 最大使用率
}

// MemoryLogQueryRequest 内存性能日志查询请求
type MemoryLogQueryRequest struct {
	BaseQueryRequest
	MinUsagePercent *float64 `json:"minUsagePercent" form:"minUsagePercent"` // 最小使用率
	MaxUsagePercent *float64 `json:"maxUsagePercent" form:"maxUsagePercent"` // 最大使用率
	MinAvailableGB  *float64 `json:"minAvailableGB" form:"minAvailableGB"`   // 最小可用GB
	MaxAvailableGB  *float64 `json:"maxAvailableGB" form:"maxAvailableGB"`   // 最大可用GB
}

// DiskPartitionLogQueryRequest 磁盘分区日志查询请求
type DiskPartitionLogQueryRequest struct {
	BaseQueryRequest
	Device          *string  `json:"device" form:"device"`                   // 设备名 (对应数据库中的deviceName字段)
	MountPoint      *string  `json:"mountPoint" form:"mountPoint"`           // 挂载点
	FsType          *string  `json:"fsType" form:"fsType"`                   // 文件系统类型 (对应数据库中的fileSystem字段)
	MinUsagePercent *float64 `json:"minUsagePercent" form:"minUsagePercent"` // 最小使用率
	MaxUsagePercent *float64 `json:"maxUsagePercent" form:"maxUsagePercent"` // 最大使用率
}

// DiskIoLogQueryRequest 磁盘IO日志查询请求
type DiskIoLogQueryRequest struct {
	BaseQueryRequest
	Device       *string  `json:"device" form:"device"`             // 设备名 (对应数据库中的deviceName字段)
	MinReadRate  *float64 `json:"minReadRate" form:"minReadRate"`   // 最小读速率
	MaxReadRate  *float64 `json:"maxReadRate" form:"maxReadRate"`   // 最大读速率
	MinWriteRate *float64 `json:"minWriteRate" form:"minWriteRate"` // 最小写速率
	MaxWriteRate *float64 `json:"maxWriteRate" form:"maxWriteRate"` // 最大写速率
}

// NetworkLogQueryRequest 网络日志查询请求
type NetworkLogQueryRequest struct {
	BaseQueryRequest
	InterfaceName *string `json:"interfaceName" form:"interfaceName"` // 网络接口名
	MinBytesRecv  *uint64 `json:"minBytesRecv" form:"minBytesRecv"`   // 最小接收字节 (对应数据库中的bytesReceived字段)
	MaxBytesRecv  *uint64 `json:"maxBytesRecv" form:"maxBytesRecv"`   // 最大接收字节 (对应数据库中的bytesReceived字段)
	MinBytesSent  *uint64 `json:"minBytesSent" form:"minBytesSent"`   // 最小发送字节 (对应数据库中的bytesSent字段)
	MaxBytesSent  *uint64 `json:"maxBytesSent" form:"maxBytesSent"`   // 最大发送字节 (对应数据库中的bytesSent字段)
}

// ProcessLogQueryRequest 进程日志查询请求
type ProcessLogQueryRequest struct {
	BaseQueryRequest
	ProcessName   *string  `json:"processName" form:"processName"`     // 进程名
	ProcessOwner  *string  `json:"processOwner" form:"processOwner"`   // 进程拥有者
	MinPid        *uint32  `json:"minPid" form:"minPid"`               // 最小进程ID (对应数据库中的processId字段)
	MaxPid        *uint32  `json:"maxPid" form:"maxPid"`               // 最大进程ID (对应数据库中的processId字段)
	MinCpuPercent *float64 `json:"minCpuPercent" form:"minCpuPercent"` // 最小CPU使用率
	MaxCpuPercent *float64 `json:"maxCpuPercent" form:"maxCpuPercent"` // 最大CPU使用率
}

// ProcessStatsLogQueryRequest 进程统计日志查询请求
type ProcessStatsLogQueryRequest struct {
	BaseQueryRequest
	MinProcessCount *uint32 `json:"minProcessCount" form:"minProcessCount"` // 最小进程数 (对应数据库中的totalCount字段)
	MaxProcessCount *uint32 `json:"maxProcessCount" form:"maxProcessCount"` // 最大进程数 (对应数据库中的totalCount字段)
	MinThreadCount  *uint32 `json:"minThreadCount" form:"minThreadCount"`   // 最小线程数
	MaxThreadCount  *uint32 `json:"maxThreadCount" form:"maxThreadCount"`   // 最大线程数
}

// TemperatureLogQueryRequest 温度日志查询请求
type TemperatureLogQueryRequest struct {
	BaseQueryRequest
	SensorName     *string  `json:"sensorName" form:"sensorName"`         // 传感器名称
	MinTemperature *float64 `json:"minTemperature" form:"minTemperature"` // 最小温度 (对应数据库中的temperatureValue字段)
	MaxTemperature *float64 `json:"maxTemperature" form:"maxTemperature"` // 最大温度 (对应数据库中的temperatureValue字段)
}
