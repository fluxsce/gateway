package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"gateway/internal/cluster/types"
	"gateway/internal/gateway/bootstrap"
	"gateway/internal/gateway/loader"
	"gateway/pkg/database"
	"gateway/pkg/logger"
)

// GatewayEventHandler 网关事件处理器
// 处理网关实例的启动、停止、重载等集群事件
type GatewayEventHandler struct {
	db database.Database
}

// NewGatewayEventHandler 创建网关事件处理器
func NewGatewayEventHandler(db database.Database) *GatewayEventHandler {
	return &GatewayEventHandler{
		db: db,
	}
}

// GetEventType 获取处理器支持的事件类型
func (h *GatewayEventHandler) GetEventType() string {
	return "GATEWAY_INSTANCE"
}

// Handle 处理网关事件
func (h *GatewayEventHandler) Handle(ctx context.Context, event *types.ClusterEvent) *types.HandleResult {
	logger.Info("处理网关集群事件",
		"eventId", event.EventId,
		"eventAction", event.EventAction,
		"eventType", event.EventType)

	// 解析事件数据
	var payload gatewayEventPayload
	if err := json.Unmarshal([]byte(event.EventPayload), &payload); err != nil {
		return types.NewFailedResult(err, fmt.Sprintf("解析事件数据失败: %v", err))
	}

	// 验证必要参数
	if payload.GatewayInstanceId == "" {
		return types.NewFailedResult(nil, "网关实例ID不能为空")
	}

	// 根据事件动作处理
	switch event.EventAction {
	case "START":
		return h.handleStart(ctx, &payload)
	case "STOP":
		return h.handleStop(ctx, &payload)
	case "RELOAD":
		return h.handleReload(ctx, &payload)
	case "RESTART":
		return h.handleRestart(ctx, &payload)
	default:
		return types.NewSkippedResult(fmt.Sprintf("未知的事件动作: %s", event.EventAction))
	}
}

// handleStart 处理启动事件
func (h *GatewayEventHandler) handleStart(ctx context.Context, payload *gatewayEventPayload) *types.HandleResult {
	gatewayInstanceId := payload.GatewayInstanceId
	tenantId := payload.TenantId

	logger.Info("处理网关启动事件",
		"gatewayInstanceId", gatewayInstanceId,
		"tenantId", tenantId)

	// 获取网关连接池
	gatewayPool := bootstrap.GetGlobalPool()

	// 检查实例是否已在连接池中
	if gatewayPool.Exists(gatewayInstanceId) {
		gateway, err := gatewayPool.Get(gatewayInstanceId)
		if err != nil {
			return types.NewFailedResult(err, fmt.Sprintf("获取网关实例失败: %v", err))
		}

		// 如果已经在运行，则跳过
		if gateway.IsRunning() {
			logger.Info("网关实例已在运行中，跳过启动",
				"gatewayInstanceId", gatewayInstanceId)
			return types.NewSkippedResult("网关实例已在运行中")
		}

		// 重新启动已存在但未运行的实例
		if err := gateway.Start(); err != nil {
			return types.NewFailedResult(err, fmt.Sprintf("启动网关实例失败: %v", err))
		}

		logger.Info("网关实例启动成功",
			"gatewayInstanceId", gatewayInstanceId)
		return types.NewSuccessResult("网关实例启动成功")
	}

	// 实例不在连接池中，需要创建并启动
	// 1. 从数据库加载配置
	configLoader := loader.NewDatabaseConfigLoader(h.db, tenantId)
	gatewayConfig, err := configLoader.LoadGatewayConfig(gatewayInstanceId)
	if err != nil {
		return types.NewFailedResult(err, fmt.Sprintf("加载网关配置失败: %v", err))
	}

	// 2. 创建网关实例
	gatewayFactory := bootstrap.NewGatewayFactory()
	gateway, err := gatewayFactory.CreateGateway(gatewayConfig, payload.ConfigFilePath)
	if err != nil {
		return types.NewFailedResult(err, fmt.Sprintf("创建网关实例失败: %v", err))
	}

	// 3. 添加到连接池
	if err := gatewayPool.Add(gatewayInstanceId, gateway); err != nil {
		return types.NewFailedResult(err, fmt.Sprintf("添加网关实例到连接池失败: %v", err))
	}

	// 4. 启动网关实例
	if err := gateway.Start(); err != nil {
		// 启动失败，从连接池中移除
		_ = gatewayPool.Remove(gatewayInstanceId)
		return types.NewFailedResult(err, fmt.Sprintf("启动网关实例失败: %v", err))
	}

	logger.Info("网关实例创建并启动成功",
		"gatewayInstanceId", gatewayInstanceId)

	result := types.NewSuccessResult("网关实例创建并启动成功")
	result.Data = map[string]interface{}{
		"gatewayInstanceId": gatewayInstanceId,
		"action":            "started",
	}
	return result
}

// handleStop 处理停止事件
func (h *GatewayEventHandler) handleStop(ctx context.Context, payload *gatewayEventPayload) *types.HandleResult {
	gatewayInstanceId := payload.GatewayInstanceId

	logger.Info("处理网关停止事件",
		"gatewayInstanceId", gatewayInstanceId)

	// 获取网关连接池
	gatewayPool := bootstrap.GetGlobalPool()

	// 检查实例是否在连接池中
	if !gatewayPool.Exists(gatewayInstanceId) {
		logger.Info("网关实例不在连接池中，跳过停止",
			"gatewayInstanceId", gatewayInstanceId)
		return types.NewSkippedResult("网关实例不在连接池中")
	}

	gateway, err := gatewayPool.Get(gatewayInstanceId)
	if err != nil {
		return types.NewFailedResult(err, fmt.Sprintf("获取网关实例失败: %v", err))
	}

	// 如果实例未运行，则跳过
	if !gateway.IsRunning() {
		logger.Info("网关实例已经停止，跳过",
			"gatewayInstanceId", gatewayInstanceId)
		return types.NewSkippedResult("网关实例已经停止")
	}

	// 停止网关实例
	if err := gateway.Stop(); err != nil {
		return types.NewFailedResult(err, fmt.Sprintf("停止网关实例失败: %v", err))
	}

	// 从连接池中移除
	if err := gatewayPool.Remove(gatewayInstanceId); err != nil {
		logger.Warn("从连接池移除网关实例失败",
			"gatewayInstanceId", gatewayInstanceId,
			"error", err)
	}

	logger.Info("网关实例停止成功",
		"gatewayInstanceId", gatewayInstanceId)

	result := types.NewSuccessResult("网关实例停止成功")
	result.Data = map[string]interface{}{
		"gatewayInstanceId": gatewayInstanceId,
		"action":            "stopped",
	}
	return result
}

// handleReload 处理重载事件
func (h *GatewayEventHandler) handleReload(ctx context.Context, payload *gatewayEventPayload) *types.HandleResult {
	gatewayInstanceId := payload.GatewayInstanceId
	tenantId := payload.TenantId

	logger.Info("处理网关重载事件",
		"gatewayInstanceId", gatewayInstanceId,
		"tenantId", tenantId)

	// 获取网关连接池
	gatewayPool := bootstrap.GetGlobalPool()

	// 检查实例是否在连接池中
	if !gatewayPool.Exists(gatewayInstanceId) {
		logger.Warn("网关实例不在连接池中，无法重载",
			"gatewayInstanceId", gatewayInstanceId)
		return types.NewSkippedResult("网关实例不在连接池中，无法重载")
	}

	gateway, err := gatewayPool.Get(gatewayInstanceId)
	if err != nil {
		return types.NewFailedResult(err, fmt.Sprintf("获取网关实例失败: %v", err))
	}

	// 检查网关是否正在运行
	if !gateway.IsRunning() {
		logger.Warn("网关实例未运行，无法重载配置",
			"gatewayInstanceId", gatewayInstanceId)
		return types.NewSkippedResult("网关实例未运行，无法重载配置")
	}

	// 从数据库重新加载配置
	configLoader := loader.NewDatabaseConfigLoader(h.db, tenantId)
	newConfig, err := configLoader.LoadGatewayConfig(gatewayInstanceId)
	if err != nil {
		return types.NewFailedResult(err, fmt.Sprintf("加载网关配置失败: %v", err))
	}

	// 重载网关配置
	if err := gateway.Reload(newConfig); err != nil {
		return types.NewFailedResult(err, fmt.Sprintf("重载网关配置失败: %v", err))
	}

	logger.Info("网关实例配置重载成功",
		"gatewayInstanceId", gatewayInstanceId)

	result := types.NewSuccessResult("网关实例配置重载成功")
	result.Data = map[string]interface{}{
		"gatewayInstanceId": gatewayInstanceId,
		"action":            "reloaded",
	}
	return result
}

// handleRestart 处理重启事件
func (h *GatewayEventHandler) handleRestart(ctx context.Context, payload *gatewayEventPayload) *types.HandleResult {
	logger.Info("处理网关重启事件",
		"gatewayInstanceId", payload.GatewayInstanceId)

	// 先停止
	stopResult := h.handleStop(ctx, payload)
	if stopResult.Status == types.HandleStatusFailed {
		return stopResult
	}

	// 再启动
	startResult := h.handleStart(ctx, payload)
	if startResult.Status == types.HandleStatusFailed {
		return startResult
	}

	result := types.NewSuccessResult("网关实例重启成功")
	result.Data = map[string]interface{}{
		"gatewayInstanceId": payload.GatewayInstanceId,
		"action":            "restarted",
	}
	return result
}

// gatewayEventPayload 网关事件数据（内部使用）
type gatewayEventPayload struct {
	GatewayInstanceId string `json:"gatewayInstanceId"` // 网关实例ID
	TenantId          string `json:"tenantId"`          // 租户ID
	InstanceName      string `json:"instanceName"`      // 实例名称（可选）
	ConfigFilePath    string `json:"configFilePath"`    // 配置文件路径（可选）
	Operator          string `json:"operator"`          // 操作人（可选）
}
