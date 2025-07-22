package types

import (
	"encoding/json"
	"time"

	metricTypes "gateway/pkg/metric/types"
	"gateway/pkg/utils/random"
)

// NetworkLog 网络接口日志表对应的结构体
type NetworkLog struct {
	// 业务字段
	MetricNetworkLogId string    `json:"metricNetworkLogId" db:"metricNetworkLogId"` // 网络接口日志ID
	TenantId           string    `json:"tenantId" db:"tenantId"`                     // 租户ID
	MetricServerId     string    `json:"metricServerId" db:"metricServerId"`         // 关联服务器ID
	InterfaceName      string    `json:"interfaceName" db:"interfaceName"`           // 接口名称
	HardwareAddr       *string   `json:"hardwareAddr" db:"hardwareAddr"`             // MAC地址
	IpAddresses        *string   `json:"ipAddresses" db:"ipAddresses"`               // IP地址列表，JSON格式
	InterfaceStatus    string    `json:"interfaceStatus" db:"interfaceStatus"`       // 接口状态
	InterfaceType      *string   `json:"interfaceType" db:"interfaceType"`           // 接口类型
	BytesReceived      int64     `json:"bytesReceived" db:"bytesReceived"`           // 接收字节数
	BytesSent          int64     `json:"bytesSent" db:"bytesSent"`                   // 发送字节数
	PacketsReceived    int64     `json:"packetsReceived" db:"packetsReceived"`       // 接收包数
	PacketsSent        int64     `json:"packetsSent" db:"packetsSent"`               // 发送包数
	ErrorsReceived     int64     `json:"errorsReceived" db:"errorsReceived"`         // 接收错误数
	ErrorsSent         int64     `json:"errorsSent" db:"errorsSent"`                 // 发送错误数
	DroppedReceived    int64     `json:"droppedReceived" db:"droppedReceived"`       // 接收丢包数
	DroppedSent        int64     `json:"droppedSent" db:"droppedSent"`               // 发送丢包数
	ReceiveRate        float64   `json:"receiveRate" db:"receiveRate"`               // 接收速率(字节/秒)
	SendRate           float64   `json:"sendRate" db:"sendRate"`                     // 发送速率(字节/秒)
	CollectTime        time.Time `json:"collectTime" db:"collectTime"`               // 采集时间

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
func (n *NetworkLog) TableName() string {
	return "HUB_METRIC_NETWORK_LOG"
}

// GetPrimaryKey 获取主键值
func (n *NetworkLog) GetPrimaryKey() (string, string) {
	return n.TenantId, n.MetricNetworkLogId
}

// NewNetworkLogFromMetrics 从NetworkInterface创建NetworkLog实例
func NewNetworkLogFromMetrics(iface *metricTypes.NetworkInterface, tenantId, serverId, operator string, collectTime time.Time, oprSeqFlag string, index int) *NetworkLog {
	// 转换IP地址列表为JSON字符串
	ipAddressesBytes, _ := json.Marshal(iface.IPAddresses)
	ipAddressesStr := string(ipAddressesBytes)

	// 处理接口类型
	var interfaceType *string
	if iface.Type != "" {
		interfaceType = &iface.Type
	}

	now := time.Now()

	return &NetworkLog{
		MetricNetworkLogId: random.Generate32BitRandomString(),
		TenantId:           tenantId,
		MetricServerId:     serverId,
		InterfaceName:      iface.Name,
		HardwareAddr:       &iface.HardwareAddr,
		IpAddresses:        &ipAddressesStr,
		InterfaceStatus:    iface.Status,
		InterfaceType:      interfaceType,
		BytesReceived:      int64(iface.BytesReceived),
		BytesSent:          int64(iface.BytesSent),
		PacketsReceived:    int64(iface.PacketsReceived),
		PacketsSent:        int64(iface.PacketsSent),
		ErrorsReceived:     int64(iface.ErrorsReceived),
		ErrorsSent:         int64(iface.ErrorsSent),
		DroppedReceived:    int64(iface.DroppedReceived),
		DroppedSent:        int64(iface.DroppedSent),
		ReceiveRate:        iface.ReceiveRate,
		SendRate:           iface.SendRate,
		CollectTime:        collectTime,
		AddTime:            now,
		AddWho:             operator,
		EditTime:           now,
		EditWho:            operator,
		OprSeqFlag:         oprSeqFlag,
		CurrentVersion:     1,
		ActiveFlag:         ActiveFlagYes,
	}
}

// NewNetworkLogsFromMetrics 从NetworkMetrics批量创建NetworkLog实例
func NewNetworkLogsFromMetrics(networkMetrics *metricTypes.NetworkMetrics, tenantId, serverId, operator string, collectTime time.Time, oprSeqFlag string) []*NetworkLog {
	var logs []*NetworkLog

	for i, iface := range networkMetrics.Interfaces {
		log := NewNetworkLogFromMetrics(&iface, tenantId, serverId, operator, collectTime, oprSeqFlag, i)
		logs = append(logs, log)
	}

	return logs
}
