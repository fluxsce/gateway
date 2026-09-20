// Package alert 把中心实例 ExtProperty 中的告警开关接到网关告警服务。
// 未启用 AlertEnabled 或告警服务未初始化时全部为无操作，不影响数据面。
package alert

import (
	"context"
	"fmt"
	"time"

	alertInit "gateway/internal/alert/init"
	"gateway/internal/servicecenterv3/model"
	"gateway/pkg/logger"
)

// NodeInfo 注册/注销/驱逐告警用的节点摘要。
type NodeInfo struct {
	NodeID      string
	ServiceName string
	NamespaceID string
	GroupName   string
	IP          string
	Port        int
	Reconnect   bool
}

// SubscribeInfo 订阅变更告警。
type SubscribeInfo struct {
	Action       string
	SubscriberID string
	NamespaceID  string
	GroupName    string
	ServiceNames []string
}

// ConfigInfo 配置发布/删除/回滚告警。
type ConfigInfo struct {
	ChangeType  string
	NamespaceID string
	GroupName   string
	DataID      string
	Version     int64
	ChangedBy   string
}

// ConnectionInfo 客户端断连告警。
type ConnectionInfo struct {
	ConnectionID        string
	ClientID            string
	ClientIP            string
	RegisteredNodeCount int
	Reason              string // client_lost 或 server_shutdown
}

// StartFailure 监听启动失败。对应 extProperty.alertOnStartFailure。
func StartFailure(cfg *model.CenterInstance, err error) {
	if cfg == nil || err == nil {
		return
	}
	ac := cfg.Alert()
	if !ac.AlertEnabled || !ac.AlertOnStartFailure {
		return
	}
	send(cfg, ac, "ERROR", "SERVICE_CENTER_START_FAILURE",
		fmt.Sprintf("服务中心启动失败 - %s", cfg.InstanceName), with(cfg, map[string]interface{}{
			"监听地址": cfg.ListenEndpoint(),
			"错误信息": err.Error(),
		}))
}

// StopAbnormal 数据面异常退出（非主动 Stop）。对应 alertOnStopAbnormal。
func StopAbnormal(cfg *model.CenterInstance, reason string) {
	if cfg == nil || reason == "" {
		return
	}
	ac := cfg.Alert()
	if !ac.AlertEnabled || !ac.AlertOnStopAbnormal {
		return
	}
	send(cfg, ac, "ERROR", "SERVICE_CENTER_STOP_ABNORMAL",
		fmt.Sprintf("服务中心异常停止 - %s", cfg.InstanceName), with(cfg, map[string]interface{}{
			"监听地址": cfg.ListenEndpoint(),
			"异常原因": reason,
		}))
}

// HealthCheckFail 持久节点被标为 Unhealthy。对应 alertOnHealthCheckFail。
func HealthCheckFail(cfg *model.CenterInstance, count int, detail string) {
	if cfg == nil || count <= 0 {
		return
	}
	ac := cfg.Alert()
	if !ac.AlertEnabled || !ac.AlertOnHealthCheckFail {
		return
	}
	if detail == "" {
		detail = fmt.Sprintf("%d 个持久节点被标记为不健康", count)
	}
	send(cfg, ac, "WARN", "SERVICE_CENTER_HEALTH_CHECK_FAIL",
		fmt.Sprintf("服务中心健康检查失败 - %s", cfg.InstanceName), with(cfg, map[string]interface{}{
			"异常节点数": count,
			"详情":    detail,
		}))
}

// NodeEviction 单轮驱逐数达到阈值。对应 alertOnNodeEviction / nodeEvictionThreshold。
func NodeEviction(cfg *model.CenterInstance, nodes []NodeInfo) {
	if cfg == nil || len(nodes) == 0 {
		return
	}
	ac := cfg.Alert()
	if !ac.AlertEnabled || !ac.AlertOnNodeEviction || len(nodes) < ac.NodeEvictionThreshold {
		return
	}
	detail := ""
	maxShow := 10
	if len(nodes) < maxShow {
		maxShow = len(nodes)
	}
	for i := 0; i < maxShow; i++ {
		n := nodes[i]
		if i > 0 {
			detail += "; "
		}
		detail += fmt.Sprintf("%s(%s:%d/%s/%s/%s)", n.NodeID, n.IP, n.Port, n.NamespaceID, n.GroupName, n.ServiceName)
	}
	if len(nodes) > maxShow {
		detail += fmt.Sprintf(" ...等共%d个节点", len(nodes))
	}
	title := fmt.Sprintf("服务中心节点驱逐 - %s", cfg.InstanceName)
	if len(nodes) >= 5 {
		title = fmt.Sprintf("服务中心节点大量驱逐 - %s", cfg.InstanceName)
	}
	extra := map[string]interface{}{
		"驱逐节点数量": len(nodes),
		"告警阈值":   ac.NodeEvictionThreshold,
		"驱逐节点详情": detail,
	}
	if len(nodes) == 1 {
		n := nodes[0]
		extra["节点ID"] = n.NodeID
		extra["服务名称"] = n.ServiceName
		extra["命名空间"] = n.NamespaceID
		extra["分组"] = n.GroupName
		extra["地址"] = fmt.Sprintf("%s:%d", n.IP, n.Port)
	}
	send(cfg, ac, "WARN", "SERVICE_CENTER_NODE_EVICTION", title, with(cfg, extra))
}

// SyncFailure 启动时加载持久数据失败。对应 alertOnSyncFailure。
func SyncFailure(cfg *model.CenterInstance, syncErr error) {
	if cfg == nil || syncErr == nil {
		return
	}
	ac := cfg.Alert()
	if !ac.AlertEnabled || !ac.AlertOnSyncFailure {
		return
	}
	send(cfg, ac, "WARN", "SERVICE_CENTER_SYNC_FAILURE",
		fmt.Sprintf("服务中心缓存同步失败 - %s", cfg.InstanceName), with(cfg, map[string]interface{}{
			"错误信息": syncErr.Error(),
		}))
}

// NodeRegister 对应 alertOnNodeRegister（默认关）。
func NodeRegister(cfg *model.CenterInstance, n NodeInfo) {
	if cfg == nil {
		return
	}
	ac := cfg.Alert()
	if !ac.AlertEnabled || !ac.AlertOnNodeRegister {
		return
	}
	send(cfg, ac, "INFO", "SERVICE_CENTER_NODE_REGISTER",
		fmt.Sprintf("服务中心节点注册 - %s", cfg.InstanceName), with(cfg, map[string]interface{}{
			"节点ID": n.NodeID,
			"服务名称": n.ServiceName,
			"命名空间": n.NamespaceID,
			"分组":   n.GroupName,
			"地址":   fmt.Sprintf("%s:%d", n.IP, n.Port),
			"是否重连": n.Reconnect,
		}))
}

// NodeUnregister 对应 alertOnNodeUnregister（默认关）。
func NodeUnregister(cfg *model.CenterInstance, n NodeInfo) {
	if cfg == nil {
		return
	}
	ac := cfg.Alert()
	if !ac.AlertEnabled || !ac.AlertOnNodeUnregister {
		return
	}
	send(cfg, ac, "INFO", "SERVICE_CENTER_NODE_UNREGISTER",
		fmt.Sprintf("服务中心节点注销 - %s", cfg.InstanceName), with(cfg, map[string]interface{}{
			"节点ID": n.NodeID,
			"服务名称": n.ServiceName,
			"命名空间": n.NamespaceID,
			"分组":   n.GroupName,
			"地址":   fmt.Sprintf("%s:%d", n.IP, n.Port),
		}))
}

// Subscribe 对应 alertOnSubscribeNotify（默认关）。
func Subscribe(cfg *model.CenterInstance, info SubscribeInfo) {
	if cfg == nil {
		return
	}
	ac := cfg.Alert()
	if !ac.AlertEnabled || !ac.AlertOnSubscribeNotify {
		return
	}
	send(cfg, ac, "INFO", "SERVICE_CENTER_SUBSCRIBE_NOTIFY",
		fmt.Sprintf("服务中心订阅通知 - %s", cfg.InstanceName), with(cfg, map[string]interface{}{
			"操作类型":   subscribeAction(info.Action),
			"连接ID":   info.SubscriberID,
			"命名空间":   info.NamespaceID,
			"分组":     info.GroupName,
			"服务名称列表": info.ServiceNames,
		}))
}

// ConfigChange 对应 alertOnConfigChange。
func ConfigChange(cfg *model.CenterInstance, info ConfigInfo) {
	if cfg == nil {
		return
	}
	ac := cfg.Alert()
	if !ac.AlertEnabled || !ac.AlertOnConfigChange {
		return
	}
	send(cfg, ac, "INFO", "SERVICE_CENTER_CONFIG_CHANGE",
		fmt.Sprintf("服务中心配置变更 - %s", cfg.InstanceName), with(cfg, map[string]interface{}{
			"变更类型": info.ChangeType,
			"命名空间": info.NamespaceID,
			"分组":   info.GroupName,
			"配置ID": info.DataID,
			"版本号":  info.Version,
			"变更人":  info.ChangedBy,
		}))
}

// ConnectionLost 对应 alertOnConnectionLost（默认关）。
func ConnectionLost(cfg *model.CenterInstance, info ConnectionInfo) {
	if cfg == nil {
		return
	}
	ac := cfg.Alert()
	if !ac.AlertEnabled || !ac.AlertOnConnectionLost {
		return
	}
	send(cfg, ac, "WARN", "SERVICE_CENTER_CONNECTION_LOST",
		fmt.Sprintf("服务中心客户端连接断开 - %s", cfg.InstanceName), with(cfg, map[string]interface{}{
			"连接ID":   info.ConnectionID,
			"客户端ID":  info.ClientID,
			"客户端IP":  info.ClientIP,
			"已注册节点数": info.RegisteredNodeCount,
			"断开原因":   disconnectReason(info.Reason),
		}))
}

func with(cfg *model.CenterInstance, extra map[string]interface{}) map[string]interface{} {
	tenant := ""
	if cfg != nil {
		tenant = cfg.TenantID
	}
	if tenant == "" {
		tenant = "default"
	}
	out := map[string]interface{}{
		"租户ID": tenant,
		"实例名称": cfg.InstanceName,
		"环境":   cfg.Environment,
		"发生时间": now(),
	}
	for k, v := range extra {
		out[k] = v
	}
	return out
}

func subscribeAction(action string) string {
	switch action {
	case "SUBSCRIBE":
		return "订阅服务"
	case "SUBSCRIBE_NAMESPACE":
		return "订阅命名空间"
	default:
		return action
	}
}

func disconnectReason(reason string) string {
	switch reason {
	case model.DisconnectClientLost:
		return "客户端断流"
	case model.DisconnectServerDrain:
		return "网关停机"
	default:
		return reason
	}
}

func send(cfg *model.CenterInstance, ac model.AlertConfig, level, alertType, title string, table map[string]interface{}) {
	svc := alertInit.GetAlertService()
	if svc == nil {
		return
	}
	_, err := svc.SendAlert(context.Background(), level, alertType, title, "", ac.ChannelName, nil, nil, table)
	if err != nil {
		logger.Debug("servicecenterv3 发送告警失败", "error", err, "alertType", alertType, "instance", cfg.InstanceName)
	}
}

func now() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
