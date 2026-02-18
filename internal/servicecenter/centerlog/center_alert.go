package centerlog

import (
	"context"
	"fmt"
	"time"

	alertInit "gateway/internal/alert/init"
	"gateway/internal/servicecenter/types"
	"gateway/pkg/logger"
)

// HandleServerStartFailure 处理服务器启动失败告警
// 参数：
//   - config: 实例配置
//   - err: 启动失败的错误信息
func HandleServerStartFailure(config *types.InstanceConfig, err error) {
	if config == nil || err == nil {
		return
	}

	alertCfg := config.GetAlertConfig()
	if alertCfg == nil || !alertCfg.AlertEnabled || !alertCfg.AlertOnStartFailure {
		return
	}

	tenantId := config.TenantID
	if tenantId == "" {
		tenantId = "default"
	}

	title := fmt.Sprintf("服务中心启动失败 - %s", config.InstanceName)
	tableData := map[string]interface{}{
		"实例名称":  config.InstanceName,
		"环境":    config.Environment,
		"监听地址":  fmt.Sprintf("%s:%d", config.ListenAddress, config.ListenPort),
		"服务器类型": config.ServerType,
		"错误信息":  err.Error(),
		"发生时间":  time.Now().Format("2006-01-02 15:04:05"),
	}

	sendAlert(tenantId, alertCfg.ChannelName, "ERROR", "SERVICE_CENTER_START_FAILURE", title, "", tableData)
}

// HandleServerStopAbnormal 处理服务器异常停止告警
// 参数：
//   - config: 实例配置
//   - reason: 异常停止原因
func HandleServerStopAbnormal(config *types.InstanceConfig, reason string) {
	if config == nil || reason == "" {
		return
	}

	alertCfg := config.GetAlertConfig()
	if alertCfg == nil || !alertCfg.AlertEnabled || !alertCfg.AlertOnStopAbnormal {
		return
	}

	tenantId := config.TenantID
	if tenantId == "" {
		tenantId = "default"
	}

	title := fmt.Sprintf("服务中心异常停止 - %s", config.InstanceName)
	tableData := map[string]interface{}{
		"实例名称": config.InstanceName,
		"环境":   config.Environment,
		"监听地址": fmt.Sprintf("%s:%d", config.ListenAddress, config.ListenPort),
		"异常原因": reason,
		"发生时间": time.Now().Format("2006-01-02 15:04:05"),
	}

	sendAlert(tenantId, alertCfg.ChannelName, "ERROR", "SERVICE_CENTER_STOP_ABNORMAL", title, "", tableData)
}

// HandleHealthCheckFailure 处理健康检查失败告警
// 参数：
//   - config: 实例配置
//   - checkErr: 健康检查的错误信息
func HandleHealthCheckFailure(config *types.InstanceConfig, checkErr error) {
	if config == nil || checkErr == nil {
		return
	}

	alertCfg := config.GetAlertConfig()
	if alertCfg == nil || !alertCfg.AlertEnabled || !alertCfg.AlertOnHealthCheckFail {
		return
	}

	tenantId := config.TenantID
	if tenantId == "" {
		tenantId = "default"
	}

	title := fmt.Sprintf("服务中心健康检查失败 - %s", config.InstanceName)
	tableData := map[string]interface{}{
		"实例名称": config.InstanceName,
		"环境":   config.Environment,
		"错误信息": checkErr.Error(),
		"发生时间": time.Now().Format("2006-01-02 15:04:05"),
	}

	sendAlert(tenantId, alertCfg.ChannelName, "WARN", "SERVICE_CENTER_HEALTH_CHECK_FAIL", title, "", tableData)
}

// HandleNodeEviction 处理节点驱逐告警
// 当单次健康检查驱逐的节点数量超过阈值时触发告警
// 参数：
//   - config: 实例配置
//   - evictedCount: 本次驱逐的节点数量
//   - evictedNodes: 被驱逐的节点信息列表（可选，用于告警详情）
func HandleNodeEviction(config *types.InstanceConfig, evictedCount int, evictedNodes []NodeEvictionInfo) {
	if config == nil || evictedCount <= 0 {
		return
	}

	alertCfg := config.GetAlertConfig()
	if alertCfg == nil || !alertCfg.AlertEnabled || !alertCfg.AlertOnNodeEviction {
		return
	}

	// 检查是否超过阈值
	if evictedCount < alertCfg.NodeEvictionThreshold {
		return
	}

	tenantId := config.TenantID
	if tenantId == "" {
		tenantId = "default"
	}

	title := fmt.Sprintf("服务中心节点大量驱逐告警 - %s", config.InstanceName)
	tableData := map[string]interface{}{
		"实例名称":   config.InstanceName,
		"环境":     config.Environment,
		"驱逐节点数量": evictedCount,
		"告警阈值":   alertCfg.NodeEvictionThreshold,
		"发生时间":   time.Now().Format("2006-01-02 15:04:05"),
	}

	// 添加被驱逐节点的详细信息（最多展示前10个）
	if len(evictedNodes) > 0 {
		maxShow := 10
		if len(evictedNodes) < maxShow {
			maxShow = len(evictedNodes)
		}
		nodeDetails := ""
		for i := 0; i < maxShow; i++ {
			node := evictedNodes[i]
			if i > 0 {
				nodeDetails += "; "
			}
			nodeDetails += fmt.Sprintf("%s(%s:%d/%s)", node.NodeId, node.IpAddress, node.Port, node.ServiceName)
		}
		if len(evictedNodes) > maxShow {
			nodeDetails += fmt.Sprintf(" ...等共%d个节点", len(evictedNodes))
		}
		tableData["驱逐节点详情"] = nodeDetails
	}

	sendAlert(tenantId, alertCfg.ChannelName, "WARN", "SERVICE_CENTER_NODE_EVICTION", title, "", tableData)
}

// HandleSyncFailure 处理缓存同步失败告警
// 参数：
//   - config: 实例配置
//   - syncErr: 同步失败的错误信息
func HandleSyncFailure(config *types.InstanceConfig, syncErr error) {
	if config == nil || syncErr == nil {
		return
	}

	alertCfg := config.GetAlertConfig()
	if alertCfg == nil || !alertCfg.AlertEnabled || !alertCfg.AlertOnSyncFailure {
		return
	}

	tenantId := config.TenantID
	if tenantId == "" {
		tenantId = "default"
	}

	title := fmt.Sprintf("服务中心缓存同步失败 - %s", config.InstanceName)
	tableData := map[string]interface{}{
		"实例名称": config.InstanceName,
		"环境":   config.Environment,
		"错误信息": syncErr.Error(),
		"发生时间": time.Now().Format("2006-01-02 15:04:05"),
	}

	sendAlert(tenantId, alertCfg.ChannelName, "WARN", "SERVICE_CENTER_SYNC_FAILURE", title, "", tableData)
}

// HandleNodeRegister 处理节点注册告警
// 参数：
//   - config: 实例配置
//   - nodeInfo: 注册的节点信息
func HandleNodeRegister(config *types.InstanceConfig, nodeInfo NodeAlertInfo) {
	if config == nil {
		return
	}

	alertCfg := config.GetAlertConfig()
	if alertCfg == nil || !alertCfg.AlertEnabled || !alertCfg.AlertOnNodeRegister {
		return
	}

	tenantId := config.TenantID
	if tenantId == "" {
		tenantId = "default"
	}

	title := fmt.Sprintf("服务中心节点注册 - %s", config.InstanceName)
	tableData := map[string]interface{}{
		"实例名称": config.InstanceName,
		"节点ID": nodeInfo.NodeId,
		"服务名称": nodeInfo.ServiceName,
		"命名空间": nodeInfo.NamespaceId,
		"分组":   nodeInfo.GroupName,
		"IP地址": nodeInfo.IpAddress,
		"端口":   nodeInfo.Port,
		"是否重连": nodeInfo.IsReconnect,
		"发生时间": time.Now().Format("2006-01-02 15:04:05"),
	}

	sendAlert(tenantId, alertCfg.ChannelName, "INFO", "SERVICE_CENTER_NODE_REGISTER", title, "", tableData)
}

// HandleNodeUnregister 处理节点注销告警
// 参数：
//   - config: 实例配置
//   - nodeInfo: 注销的节点信息
func HandleNodeUnregister(config *types.InstanceConfig, nodeInfo NodeAlertInfo) {
	if config == nil {
		return
	}

	alertCfg := config.GetAlertConfig()
	if alertCfg == nil || !alertCfg.AlertEnabled || !alertCfg.AlertOnNodeUnregister {
		return
	}

	tenantId := config.TenantID
	if tenantId == "" {
		tenantId = "default"
	}

	title := fmt.Sprintf("服务中心节点注销 - %s", config.InstanceName)
	tableData := map[string]interface{}{
		"实例名称": config.InstanceName,
		"节点ID": nodeInfo.NodeId,
		"服务名称": nodeInfo.ServiceName,
		"命名空间": nodeInfo.NamespaceId,
		"分组":   nodeInfo.GroupName,
		"IP地址": nodeInfo.IpAddress,
		"端口":   nodeInfo.Port,
		"发生时间": time.Now().Format("2006-01-02 15:04:05"),
	}

	sendAlert(tenantId, alertCfg.ChannelName, "INFO", "SERVICE_CENTER_NODE_UNREGISTER", title, "", tableData)
}

// HandleSubscribeNotify 处理服务订阅通知告警
// 参数：
//   - config: 实例配置
//   - subInfo: 订阅信息
func HandleSubscribeNotify(config *types.InstanceConfig, subInfo SubscribeAlertInfo) {
	if config == nil {
		return
	}

	alertCfg := config.GetAlertConfig()
	if alertCfg == nil || !alertCfg.AlertEnabled || !alertCfg.AlertOnSubscribeNotify {
		return
	}

	tenantId := config.TenantID
	if tenantId == "" {
		tenantId = "default"
	}

	title := fmt.Sprintf("服务中心订阅通知 - %s", config.InstanceName)
	tableData := map[string]interface{}{
		"实例名称":   config.InstanceName,
		"操作类型":   subInfo.Action,
		"订阅者ID":  subInfo.SubscriberId,
		"命名空间":   subInfo.NamespaceId,
		"分组":     subInfo.GroupName,
		"服务名称列表": subInfo.ServiceNames,
		"发生时间":   time.Now().Format("2006-01-02 15:04:05"),
	}

	sendAlert(tenantId, alertCfg.ChannelName, "INFO", "SERVICE_CENTER_SUBSCRIBE_NOTIFY", title, "", tableData)
}

// HandleConfigChange 处理配置变更告警
// 参数：
//   - config: 实例配置
//   - cfgInfo: 配置变更信息
func HandleConfigChange(config *types.InstanceConfig, cfgInfo ConfigChangeAlertInfo) {
	if config == nil {
		return
	}

	alertCfg := config.GetAlertConfig()
	if alertCfg == nil || !alertCfg.AlertEnabled || !alertCfg.AlertOnConfigChange {
		return
	}

	tenantId := config.TenantID
	if tenantId == "" {
		tenantId = "default"
	}

	title := fmt.Sprintf("服务中心配置变更 - %s", config.InstanceName)
	tableData := map[string]interface{}{
		"实例名称": config.InstanceName,
		"变更类型": cfgInfo.ChangeType,
		"命名空间": cfgInfo.NamespaceId,
		"分组":   cfgInfo.GroupName,
		"配置ID": cfgInfo.ConfigDataId,
		"版本号":  cfgInfo.Version,
		"变更人":  cfgInfo.ChangedBy,
		"发生时间": time.Now().Format("2006-01-02 15:04:05"),
	}

	sendAlert(tenantId, alertCfg.ChannelName, "INFO", "SERVICE_CENTER_CONFIG_CHANGE", title, "", tableData)
}

// HandleConnectionLost 处理客户端连接断开告警
// 参数：
//   - config: 实例配置
//   - connInfo: 断开的连接信息
func HandleConnectionLost(config *types.InstanceConfig, connInfo ConnectionAlertInfo) {
	if config == nil {
		return
	}

	alertCfg := config.GetAlertConfig()
	if alertCfg == nil || !alertCfg.AlertEnabled || !alertCfg.AlertOnConnectionLost {
		return
	}

	tenantId := config.TenantID
	if tenantId == "" {
		tenantId = "default"
	}

	title := fmt.Sprintf("服务中心客户端连接断开 - %s", config.InstanceName)
	tableData := map[string]interface{}{
		"实例名称":   config.InstanceName,
		"连接ID":   connInfo.ConnectionId,
		"客户端ID":  connInfo.ClientId,
		"客户端IP":  connInfo.ClientIP,
		"已注册节点数": connInfo.RegisteredNodeCount,
		"发生时间":   time.Now().Format("2006-01-02 15:04:05"),
	}

	sendAlert(tenantId, alertCfg.ChannelName, "WARN", "SERVICE_CENTER_CONNECTION_LOST", title, "", tableData)
}

// NodeEvictionInfo 节点驱逐信息
// 用于告警时展示被驱逐节点的简要信息
type NodeEvictionInfo struct {
	NodeId      string // 节点ID
	ServiceName string // 服务名称
	IpAddress   string // IP地址
	Port        int    // 端口号
}

// NodeAlertInfo 节点告警信息（用于注册/注销告警）
type NodeAlertInfo struct {
	NodeId      string // 节点ID
	ServiceName string // 服务名称
	NamespaceId string // 命名空间
	GroupName   string // 分组
	IpAddress   string // IP地址
	Port        int    // 端口号
	IsReconnect bool   // 是否重连注册
}

// SubscribeAlertInfo 订阅告警信息
type SubscribeAlertInfo struct {
	Action       string   // 操作类型：SUBSCRIBE / UNSUBSCRIBE
	SubscriberId string   // 订阅者ID
	NamespaceId  string   // 命名空间
	GroupName    string   // 分组
	ServiceNames []string // 服务名称列表
}

// ConfigChangeAlertInfo 配置变更告警信息
type ConfigChangeAlertInfo struct {
	ChangeType   string // 变更类型：ADD / UPDATE / DELETE / ROLLBACK
	NamespaceId  string // 命名空间
	GroupName    string // 分组
	ConfigDataId string // 配置ID
	Version      int64  // 版本号
	ChangedBy    string // 变更人
}

// ConnectionAlertInfo 连接告警信息
type ConnectionAlertInfo struct {
	ConnectionId        string // 连接ID
	ClientId            string // 客户端ID
	ClientIP            string // 客户端IP
	RegisteredNodeCount int    // 已注册的节点数量
}

// sendAlert 发送告警（直接调用告警服务）
func sendAlert(tenantId, channelName, level, alertType, title, content string, tableData map[string]interface{}) {
	ctx := context.Background()
	svc := alertInit.GetAlertService()
	if svc == nil {
		return
	}

	_, err := svc.SendAlert(ctx, level, alertType, title, content, channelName, nil, nil, tableData)
	if err != nil {
		logger.Debug("发送服务中心告警失败", "error", err, "alertType", alertType)
	}
}
