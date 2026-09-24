package clusterstatus

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gateway/pkg/database"
)

// errNeedInstance 网关实例页未选中行。
var errNeedInstance = errors.New("请选择一个网关实例")

// ClusterNodeRow 已登记的集群进程。
type ClusterNodeRow struct {
	NodeId       string    `json:"nodeId" db:"nodeId"`
	NodeIp       string    `json:"nodeIp" db:"nodeIp"`
	Hostname     string    `json:"hostname" db:"hostname"`
	StartedTime  time.Time `json:"startedTime" db:"startedTime"`
	LastSeenTime time.Time `json:"lastSeenTime" db:"lastSeenTime"`
}

// HealthCount 健康标记汇总。停用单独计数，不记入不健康。
type HealthCount struct {
	Total     int `json:"total" db:"total"`
	Healthy   int `json:"healthy" db:"healthy"`
	Unhealthy int `json:"unhealthy" db:"unhealthy"`
	Inactive  int `json:"inactive" db:"inactive"`
}

// UnhealthyInstanceRow 启用中且健康标记不是 Y 的网关实例。
type UnhealthyInstanceRow struct {
	GatewayInstanceId string `json:"gatewayInstanceId" db:"gatewayInstanceId"`
	InstanceName      string `json:"instanceName" db:"instanceName"`
	BindAddress       string `json:"bindAddress" db:"bindAddress"`
	HttpPort          *int   `json:"httpPort" db:"httpPort"`
	HttpsPort         *int   `json:"httpsPort" db:"httpsPort"`
	HealthStatus      string `json:"healthStatus" db:"healthStatus"`
}

// UnhealthyServiceRow 启用中且健康标记不是 Y 的服务节点。
type UnhealthyServiceRow struct {
	ServiceName  string `json:"serviceName" db:"serviceName"`
	NodeHost     string `json:"nodeHost" db:"nodeHost"`
	NodePort     int    `json:"nodePort" db:"nodePort"`
	HealthStatus string `json:"healthStatus" db:"healthStatus"`
}

// ClusterTopology 集群健康与节点登记。各监听页共用这一份结果。
type ClusterTopology struct {
	Nodes              []ClusterNodeRow       `json:"nodes"`
	Instances          HealthCount            `json:"instances"`
	Services           HealthCount            `json:"services"`
	UnhealthyInstances []UnhealthyInstanceRow `json:"unhealthyInstances"`
	UnhealthyServices  []UnhealthyServiceRow  `json:"unhealthyServices"`
	// Listeners 对每个存活节点上的网关监听口做的实时探测，不使用实例表里会被覆盖的健康标记。
	Listeners []ListenProbe `json:"listeners"`
}

// ListenProbe 某个集群节点上一个监听口的探测结果。
type ListenProbe struct {
	NodeId     string `json:"nodeId"`
	NodeIp     string `json:"nodeIp"`
	Hostname   string `json:"hostname"`
	TargetId   string `json:"targetId"`
	TargetName string `json:"targetName"`
	Scheme     string `json:"scheme"`
	Port       int    `json:"port"`
	Status     string `json:"status"`
	Message    string `json:"message"`
}

type gatewayListenRow struct {
	GatewayInstanceId string `db:"gatewayInstanceId"`
	InstanceName      string `db:"instanceName"`
	BindAddress       string `db:"bindAddress"`
	HTTPPort          *int   `db:"httpPort"`
	HTTPSPort         *int   `db:"httpsPort"`
	TLSEnabled        string `db:"tlsEnabled"`
}

// Query 返回本租户的集群节点，并探测调用方给出的监听口。
func Query(ctx context.Context, db database.Database, tenantId string, targets []ListenTarget) (*ClusterTopology, error) {
	view := &ClusterTopology{
		Nodes:              []ClusterNodeRow{},
		UnhealthyInstances: []UnhealthyInstanceRow{},
		UnhealthyServices:  []UnhealthyServiceRow{},
		Listeners:          []ListenProbe{},
	}
	if err := db.Query(ctx, &view.Nodes, `
		SELECT nodeId, nodeIp, hostname, startedTime, lastSeenTime
		FROM HUB_CLUSTER_NODE
		WHERE tenantId = ? AND activeFlag = 'Y'
		ORDER BY hostname, nodeId
	`, []interface{}{tenantId}, true); err != nil {
		return nil, fmt.Errorf("查询集群节点失败: %w", err)
	}
	if view.Nodes == nil {
		view.Nodes = []ClusterNodeRow{}
	}

	view.Listeners = probeListeners(ctx, view.Nodes, targets)
	return view, nil
}

// LoadGatewayInstance 从网关实例表取出这一条的监听口。其它模块不要调用它。
func LoadGatewayInstance(ctx context.Context, db database.Database, tenantId, gatewayInstanceId string) ([]ListenTarget, error) {
	var row gatewayListenRow
	err := db.QueryOne(ctx, &row, `
		SELECT gatewayInstanceId, instanceName, bindAddress, httpPort, httpsPort, tlsEnabled
		FROM HUB_GW_INSTANCE
		WHERE tenantId = ? AND gatewayInstanceId = ? AND activeFlag = 'Y'
	`, []interface{}{tenantId, gatewayInstanceId}, true)
	if err != nil {
		return nil, fmt.Errorf("查询网关监听口失败: %w", err)
	}
	scheme, port := "http", 8080
	if strings.EqualFold(row.TLSEnabled, "Y") && row.HTTPSPort != nil && *row.HTTPSPort > 0 {
		scheme, port = "https", *row.HTTPSPort
	} else if row.HTTPPort != nil && *row.HTTPPort > 0 {
		scheme, port = "http", *row.HTTPPort
	}
	return []ListenTarget{{
		ID:          row.GatewayInstanceId,
		Name:        row.InstanceName,
		BindAddress: row.BindAddress,
		Scheme:      scheme,
		Port:        port,
		HealthPath:  listenHealthPath(),
	}}, nil
}
