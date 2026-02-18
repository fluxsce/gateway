package handler

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"

	"gateway/internal/servicecenter/cache"
	"gateway/internal/servicecenter/centerlog"
	"gateway/internal/servicecenter/dao"
	pb "gateway/internal/servicecenter/server/proto"
	"gateway/internal/servicecenter/server/subscriber"
	"gateway/internal/servicecenter/types"
	"gateway/pkg/logger"
	"gateway/pkg/utils/random"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// 配置中心架构说明
//
// ================================================================================
// 数据写入策略
// ================================================================================
//
// 配置管理：直接写数据库（保证持久化和一致性）
//   - 配置不使用缓存，每次都从数据库读取最新数据
//   - 所有配置操作（保存、删除、回滚）都直接操作数据库
//   - 配置查询（获取、列表、历史）都直接从数据库读取
//
// 为什么配置管理需要直接写数据库？
//   - 持久化要求：配置数据必须持久化，不能丢失
//   - 一致性要求：配置变更需要强一致性，所有实例立即生效
//   - 历史记录：需要保存配置历史，支持回滚操作
//   - 低频操作：配置变更是低频操作，数据库性能足够
//
// ================================================================================
// 配置通知机制
// ================================================================================
//
// 实时推送机制：
//   - 通过 WatchConfig 实现配置变更实时推送
//   - 客户端订阅配置，服务端主动推送变更
//   - 支持批量监听多个配置，所有配置共用同一个 channel
//
// 通知触发时机：
//   - SaveConfig: 配置保存后立即通知所有监听者
//   - DeleteConfig: 配置删除后立即通知所有监听者
//   - RollbackConfig: 配置回滚后立即通知所有监听者
//
// 事件类型：
//   - CONFIG_UPDATED: 配置更新（包括新增和修改）
//   - CONFIG_DELETED: 配置删除
//
// ================================================================================
// 配置版本管理
// ================================================================================
//
// 版本号规则：
//   - 新增配置：版本号从 1 开始
//   - 更新配置：版本号 = 当前版本 + 1
//   - 回滚配置：版本号 = 当前版本 + 1（回滚也会创建新版本）
//
// 历史记录：
//   - 每次配置变更都会保存历史记录
//   - 历史记录包含变更前后内容、版本、MD5 值
//   - 支持按版本号查询历史记录，用于回滚操作
//
// ================================================================================

// ConfigHandlerDeps Config Handler 依赖注入
type ConfigHandlerDeps struct {
	ConfigDAO      *dao.ConfigDAO
	HistoryDAO     *dao.HistoryDAO
	ConfigProvider ConfigProvider // 配置提供者（用于告警等功能）
}

// ConfigHandler gRPC 配置中心处理器
type ConfigHandler struct {
	pb.UnimplementedConfigCenterServer
	deps          *ConfigHandlerDeps
	configWatcher *subscriber.ConfigWatcher
}

// NewConfigHandler 创建配置中心处理器
func NewConfigHandler(deps *ConfigHandlerDeps) *ConfigHandler {
	return &ConfigHandler{
		deps:          deps,
		configWatcher: subscriber.NewConfigWatcher(),
	}
}

// validateNamespace 验证命名空间是否存在且有效（纯缓存操作）
// 如果命名空间不存在或已被禁用，返回权限错误
// 注意：命名空间应该在服务启动时已加载到缓存，这里只从缓存校验
func (h *ConfigHandler) validateNamespace(ctx context.Context, tenantId, namespaceId string) error {
	if namespaceId == "" {
		return status.Errorf(codes.InvalidArgument, "namespaceId is required")
	}

	// 从缓存获取命名空间
	namespace, found := cache.GetGlobalCache().GetNamespace(ctx, tenantId, namespaceId)
	if !found || namespace == nil {
		return status.Errorf(codes.PermissionDenied, "namespace not found: %s", namespaceId)
	}

	// 检查命名空间是否已禁用
	if namespace.ActiveFlag != "Y" {
		return status.Errorf(codes.PermissionDenied, "namespace is disabled: %s", namespaceId)
	}

	return nil
}

// GetConfigWatcher 获取配置监听器（供外部手动触发事件使用）
func (h *ConfigHandler) GetConfigWatcher() *subscriber.ConfigWatcher {
	return h.configWatcher
}

// 配置基本操作

// GetConfig 获取配置
//
// 处理流程：
//  1. 从数据库查询配置（直接读数据库，不使用缓存）
//  2. 转换为 protobuf 格式
//  3. 返回配置数据
//
// 注意事项：
//   - 配置直接从数据库读取，保证数据一致性
//   - 不使用缓存，每次都是最新数据
func (h *ConfigHandler) GetConfig(ctx context.Context, req *pb.ConfigKey) (*pb.GetConfigResponse, error) {
	tenantID := "default" // TODO: 从 context 获取

	// 验证命名空间是否存在
	if err := h.validateNamespace(ctx, tenantID, req.NamespaceId); err != nil {
		return &pb.GetConfigResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// 从数据库获取
	config, err := h.deps.ConfigDAO.GetConfig(ctx, tenantID, req.NamespaceId, req.GroupName, req.ConfigDataId)
	if err != nil {
		return &pb.GetConfigResponse{
			Success: false,
			Message: fmt.Sprintf("config not found: %v", err),
		}, nil
	}

	// 转换为 protobuf 格式
	pbConfig := convertConfigToProto(config)

	return &pb.GetConfigResponse{
		Success: true,
		Message: "config found",
		Config:  pbConfig,
	}, nil
}

// SaveConfig 保存配置
//
// 处理流程：
//  1. 验证请求参数（ConfigData 不能为空）
//  2. 计算配置内容的 MD5 值（服务端自动计算）
//  3. 检查配置是否存在：
//     - 如果存在：changeType = "UPDATE"，版本号 = 当前版本 + 1
//     - 如果不存在：changeType = "ADD"，版本号 = 1
//  4. 保存配置到数据库（直接写数据库，保证持久化）
//  5. 保存历史记录到数据库（记录变更前后内容、版本、MD5）
//  6. 通知所有监听者配置已更新（通过 ConfigWatcher 推送事件）
//  7. 返回新版本号和 MD5 值
//
// 注意事项：
//   - 配置直接写数据库，不使用缓存（保证持久化和一致性）
//   - 历史记录保存失败不影响主流程（只记录日志）
//   - 配置变更会立即通知所有监听者（实时推送）
func (h *ConfigHandler) SaveConfig(ctx context.Context, req *pb.ConfigData) (*pb.SaveConfigResponse, error) {
	if req == nil {
		return &pb.SaveConfigResponse{
			Success: false,
			Message: "config is required",
		}, nil
	}

	tenantID := "default" // TODO: 从 context 获取

	// 验证命名空间是否存在
	if err := h.validateNamespace(ctx, tenantID, req.NamespaceId); err != nil {
		return &pb.SaveConfigResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// 计算 MD5（服务端自动计算）
	contentMD5 := calculateMD5(req.ConfigContent)

	// 检查配置是否存在
	existingConfig, err := h.deps.ConfigDAO.GetConfig(ctx, tenantID, req.NamespaceId, req.GroupName, req.ConfigDataId)

	var newVersion int64
	changeType := "ADD"

	// 注意：GetConfig 在记录不存在时返回 (nil, nil)，需要同时检查 err 和 existingConfig
	if err == nil && existingConfig != nil {
		// 配置已存在，更新
		changeType = "UPDATE"
		newVersion = existingConfig.Version + 1
	} else {
		// 新配置
		newVersion = 1
	}

	now := time.Now()
	changedBy := req.ChangedBy // 从请求中获取变更人
	config := &types.ConfigData{
		TenantId:          tenantID,
		NamespaceId:       req.NamespaceId,
		GroupName:         req.GroupName,
		ConfigDataId:      req.ConfigDataId,
		ContentType:       req.ContentType,
		ConfigContent:     req.ConfigContent,
		Md5Value:          contentMD5, // 服务端计算的 MD5
		ConfigDescription: req.ConfigDesc,
		Version:           newVersion,
		EditTime:          now,
		EditWho:           changedBy,
	}

	// 如果是新配置，设置 AddTime 和 AddWho；如果是更新，保持原值
	if existingConfig == nil {
		config.AddTime = now
		config.AddWho = changedBy
	} else {
		config.AddTime = existingConfig.AddTime
		config.AddWho = existingConfig.AddWho
	}

	// 保存配置
	if err := h.deps.ConfigDAO.SaveConfig(ctx, config); err != nil {
		return &pb.SaveConfigResponse{
			Success: false,
			Message: fmt.Sprintf("failed to save config: %v", err),
		}, nil
	}

	// 保存历史记录
	history := &types.ConfigHistory{
		ConfigHistoryId: random.Generate32BitRandomString(), // 生成唯一的配置历史ID（32位）
		TenantId:        tenantID,
		NamespaceId:     config.NamespaceId,
		GroupName:       config.GroupName,
		ConfigDataId:    config.ConfigDataId,
		ChangeType:      changeType,
		NewContent:      config.ConfigContent,
		NewVersion:      config.Version,
		NewMd5Value:     config.Md5Value,
		ChangeReason:    req.ChangeReason,
		ChangedBy:       changedBy,
		ChangedAt:       now,
		AddTime:         now,
		AddWho:          changedBy,
		EditTime:        now,
		EditWho:         changedBy,
	}
	if existingConfig != nil {
		history.OldContent = existingConfig.ConfigContent
		history.OldVersion = existingConfig.Version
		history.OldMd5Value = existingConfig.Md5Value
	}
	if err := h.deps.HistoryDAO.CreateHistory(ctx, history); err != nil {
		// 历史记录失败不影响主流程，只记录日志
		logger.Warn("保存配置历史记录失败", "error", err,
			"namespaceId", config.NamespaceId,
			"groupName", config.GroupName,
			"configDataId", config.ConfigDataId)
	}

	// 通知监听者
	h.configWatcher.NotifyConfigUpdate(config)

	// 发送配置变更告警
	if h.deps.ConfigProvider != nil {
		instanceConfig := h.deps.ConfigProvider.GetConfig()
		if instanceConfig != nil {
			configInfo := centerlog.ConfigChangeAlertInfo{
				ChangeType:   changeType,
				NamespaceId:  config.NamespaceId,
				GroupName:    config.GroupName,
				ConfigDataId: config.ConfigDataId,
				Version:      config.Version,
				ChangedBy:    config.EditWho,
			}
			centerlog.HandleConfigChange(instanceConfig, configInfo)
		}
	}

	return &pb.SaveConfigResponse{
		Success:    true,
		Message:    "config saved successfully",
		Version:    newVersion,
		ContentMd5: contentMD5, // 返回服务端计算的 MD5
	}, nil
}

// DeleteConfig 删除配置
//
// 处理流程：
//  1. 验证配置是否存在（不存在则返回错误）
//  2. 从数据库删除配置（直接删除，保证持久化）
//  3. 通知所有监听者配置已删除（通过 ConfigWatcher 推送事件）
//  4. 返回删除结果
//
// 注意事项：
//   - 删除操作直接操作数据库，不使用缓存
//   - 配置删除会立即通知所有监听者（实时推送）
//   - 历史记录保留（不删除历史记录，支持审计和回滚）
func (h *ConfigHandler) DeleteConfig(ctx context.Context, req *pb.ConfigKey) (*pb.ConfigResponse, error) {
	tenantID := "default" // TODO: 从 context 获取

	// 验证命名空间是否存在
	if err := h.validateNamespace(ctx, tenantID, req.NamespaceId); err != nil {
		return &pb.ConfigResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// 获取配置（用于通知）
	config, err := h.deps.ConfigDAO.GetConfig(ctx, tenantID, req.NamespaceId, req.GroupName, req.ConfigDataId)
	if err != nil {
		return &pb.ConfigResponse{
			Success: false,
			Message: fmt.Sprintf("config not found: %v", err),
		}, nil
	}

	// 删除配置
	if err := h.deps.ConfigDAO.DeleteConfig(ctx, tenantID, req.NamespaceId, req.GroupName, req.ConfigDataId); err != nil {
		return &pb.ConfigResponse{
			Success: false,
			Message: fmt.Sprintf("failed to delete config: %v", err),
		}, nil
	}

	// 通知监听者
	h.configWatcher.NotifyConfigDelete(tenantID, config.NamespaceId, config.GroupName, config.ConfigDataId)

	// 发送配置删除告警
	if h.deps.ConfigProvider != nil {
		instanceConfig := h.deps.ConfigProvider.GetConfig()
		if instanceConfig != nil {
			configInfo := centerlog.ConfigChangeAlertInfo{
				ChangeType:   "DELETE",
				NamespaceId:  config.NamespaceId,
				GroupName:    config.GroupName,
				ConfigDataId: config.ConfigDataId,
				Version:      config.Version,
				ChangedBy:    "system", // DeleteConfig 没有 ChangedBy 字段，使用默认值
			}
			centerlog.HandleConfigChange(instanceConfig, configInfo)
		}
	}

	return &pb.ConfigResponse{
		Success: true,
		Message: "config deleted successfully",
	}, nil
}

// ListConfigs 列出配置列表
//
// 处理流程：
//  1. 从数据库查询配置列表（根据 namespaceId 和 groupName）
//  2. 转换为 protobuf 格式
//  3. 返回配置列表
//
// 注意事项：
//   - 配置直接从数据库读取，保证数据一致性
//   - 不使用缓存，每次都是最新数据
func (h *ConfigHandler) ListConfigs(ctx context.Context, req *pb.ListConfigsRequest) (*pb.ListConfigsResponse, error) {
	tenantID := "default" // TODO: 从 context 获取

	// 验证命名空间是否存在
	if err := h.validateNamespace(ctx, tenantID, req.NamespaceId); err != nil {
		return &pb.ListConfigsResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// 从数据库查询
	configs, err := h.deps.ConfigDAO.ListConfigs(ctx, tenantID, req.NamespaceId, req.GroupName)
	if err != nil {
		return &pb.ListConfigsResponse{
			Success: false,
			Message: fmt.Sprintf("failed to list configs: %v", err),
		}, nil
	}

	// 转换为 protobuf 格式
	pbConfigs := make([]*pb.ConfigData, 0, len(configs))
	for _, config := range configs {
		pbConfigs = append(pbConfigs, convertConfigToProto(config))
	}

	return &pb.ListConfigsResponse{
		Success: true,
		Message: fmt.Sprintf("found %d configs", len(pbConfigs)),
		Configs: pbConfigs,
	}, nil
}

// 配置监听（实时推送）

// WatchConfig 监听配置变更（统一接口，支持监听单个或多个配置）
//
// 接口设计说明：
//   - 这是统一的监听接口，支持监听单个或多个配置
//   - 单个配置监听：configDataIds = ["config1"]
//   - 多个配置监听：configDataIds = ["config1", "config2", "config3"]
//   - 一个客户端应用可以监听多个配置，所有配置共用同一个 channel
//   - 减少 gRPC Stream 连接数，提高效率
//   - 用于数据库变动主动推送，只需要配置标识即可
//
// 处理流程：
//  1. 验证请求参数（namespaceId、configDataIds）
//  2. 生成唯一的 watcherID（每个连接独立）
//  3. 调用 configWatcher.WatchMultipleConfigs() 批量监听配置
//  4. 所有配置共用同一个 channel，减少连接数
//  5. 持续从 channel 读取事件并推送给客户端
//  6. 连接断开时，通过 defer 自动清理监听
func (h *ConfigHandler) WatchConfig(req *pb.WatchConfigRequest, stream pb.ConfigCenter_WatchConfigServer) error {
	tenantID := "default"                           // TODO: 从 context 获取
	watcherID := random.Generate32BitRandomString() // 生成唯一的监听器ID（32位）

	// 验证请求参数
	if req.NamespaceId == "" {
		return status.Errorf(codes.InvalidArgument, "namespaceId is required")
	}
	if len(req.ConfigDataIds) == 0 {
		return status.Errorf(codes.InvalidArgument, "configDataIds is required and cannot be empty")
	}

	// 设置默认值
	groupName := req.GroupName
	if groupName == "" {
		groupName = "DEFAULT_GROUP"
	}

	// 验证配置标识列表
	for _, configDataId := range req.ConfigDataIds {
		if configDataId == "" {
			return status.Errorf(codes.InvalidArgument, "configDataId cannot be empty in configDataIds")
		}
	}

	// 验证命名空间是否存在
	if err := h.validateNamespace(stream.Context(), tenantID, req.NamespaceId); err != nil {
		return err
	}

	// 打印监听开始日志
	logger.Info("配置监听注册开始",
		"watcherID", watcherID,
		"tenantID", tenantID,
		"namespaceId", req.NamespaceId,
		"groupName", groupName,
		"configDataIds", req.ConfigDataIds,
		"configCount", len(req.ConfigDataIds))

	// 监听配置（使用同一个 channel，支持单个或多个配置）
	ch := h.configWatcher.Watch(
		stream.Context(),
		tenantID,
		req.NamespaceId,
		groupName,
		req.ConfigDataIds,
		watcherID,
	)
	defer func() {
		logger.Info("配置监听注销",
			"watcherID", watcherID,
			"tenantID", tenantID,
			"namespaceId", req.NamespaceId,
			"groupName", groupName,
			"configDataIds", req.ConfigDataIds)
		h.configWatcher.Unwatch(watcherID)
	}()

	// 订阅成功后，立即推送当前配置给客户端（全量推送）
	// 直接发送到当前订阅者的 channel，不影响其他订阅者
	// 这样客户端可以立即获得最新配置，而不需要单独调用 GetConfig
	go func() {
		for _, configDataId := range req.ConfigDataIds {
			// 查询当前配置
			config, err := h.deps.ConfigDAO.GetConfig(stream.Context(), tenantID, req.NamespaceId, groupName, configDataId)
			if err != nil {
				// 查询失败，记录日志但继续处理其他配置
				logger.Warn("查询配置失败，跳过初始推送",
					"watcherID", watcherID,
					"namespaceId", req.NamespaceId,
					"groupName", groupName,
					"configDataId", configDataId,
					"error", err)
				continue
			}

			// 如果配置存在，直接发送到当前订阅者的 channel
			if config != nil {
				initialEvent := &pb.ConfigChangeEvent{
					EventType:    "CONFIG_UPDATED", // 使用 CONFIG_UPDATED 表示这是当前配置
					Timestamp:    time.Now().Format("2006-01-02 15:04:05"),
					NamespaceId:  config.NamespaceId,
					GroupName:    config.GroupName,
					ConfigDataId: config.ConfigDataId,
					ContentMd5:   config.Md5Value,
					Config:       convertConfigToProto(config),
				}

				// 直接发送到当前订阅者的 channel（只发送给当前订阅者）
				h.configWatcher.SendToWatcher(watcherID, initialEvent)

				logger.Debug("已推送初始配置到 channel",
					"watcherID", watcherID,
					"namespaceId", req.NamespaceId,
					"groupName", groupName,
					"configDataId", configDataId,
					"version", config.Version)
			} else {
				// 配置不存在，推送删除事件（表示配置已被删除）
				deleteEvent := &pb.ConfigChangeEvent{
					EventType:    "CONFIG_DELETED",
					Timestamp:    time.Now().Format("2006-01-02 15:04:05"),
					NamespaceId:  req.NamespaceId,
					GroupName:    groupName,
					ConfigDataId: configDataId,
				}

				// 直接发送到当前订阅者的 channel（只发送给当前订阅者）
				h.configWatcher.SendToWatcher(watcherID, deleteEvent)

				logger.Debug("已推送配置删除状态到 channel（配置不存在）",
					"watcherID", watcherID,
					"namespaceId", req.NamespaceId,
					"groupName", groupName,
					"configDataId", configDataId)
			}
		}
	}()

	// 持续监听变更事件并推送给客户端
	// 所有配置的变更事件都会通过同一个 channel 推送
	for {
		select {
		case event, ok := <-ch:
			if !ok {
				return nil // 通道已关闭
			}
			if err := stream.Send(event); err != nil {
				return err
			}
		case <-stream.Context().Done():
			return stream.Context().Err()
		}
	}
}

// 配置历史和回滚

// GetConfigHistory 获取配置历史
//
// 处理流程：
//  1. 从数据库查询配置历史记录（根据 namespaceId、groupName、configDataId）
//  2. 限制返回数量（根据 limit 参数）
//  3. 转换为 protobuf 格式
//  4. 返回历史记录列表
//
// 注意事项：
//   - 历史记录直接从数据库读取
//   - 支持按版本号查询，用于回滚操作
func (h *ConfigHandler) GetConfigHistory(ctx context.Context, req *pb.GetConfigHistoryRequest) (*pb.GetConfigHistoryResponse, error) {
	tenantID := "default" // TODO: 从 context 获取

	// 验证命名空间是否存在
	if err := h.validateNamespace(ctx, tenantID, req.NamespaceId); err != nil {
		return &pb.GetConfigHistoryResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// 从数据库查询
	histories, err := h.deps.HistoryDAO.GetConfigHistory(ctx, tenantID, req.NamespaceId, req.GroupName, req.ConfigDataId, int(req.Limit))
	if err != nil {
		return &pb.GetConfigHistoryResponse{
			Success: false,
			Message: fmt.Sprintf("failed to get config history: %v", err),
		}, nil
	}

	// 转换为 protobuf 格式
	pbHistories := make([]*pb.ConfigHistory, 0, len(histories))
	for _, history := range histories {
		pbHistories = append(pbHistories, convertHistoryToProto(history))
	}

	return &pb.GetConfigHistoryResponse{
		Success: true,
		Message: fmt.Sprintf("found %d history records", len(pbHistories)),
		History: pbHistories,
	}, nil
}

// RollbackConfig 回滚配置
//
// 处理流程：
//  1. 验证目标版本是否存在（从历史记录中查找）
//  2. 获取当前配置（用于记录回滚前的状态）
//  3. 创建新配置（基于目标版本的历史记录内容）
//  4. 计算新版本号（当前版本 + 1）
//  5. 保存配置到数据库（直接写数据库）
//  6. 保存历史记录（changeType = "ROLLBACK"，记录回滚前后状态）
//  7. 通知所有监听者配置已更新（通过 ConfigWatcher 推送事件）
//  8. 返回新版本号和 MD5 值
//
// 注意事项：
//   - 回滚操作会创建新版本，不会覆盖历史记录
//   - 如果当前配置不存在，版本号从 1 开始
//   - 回滚后的配置会立即通知所有监听者（实时推送）
func (h *ConfigHandler) RollbackConfig(ctx context.Context, req *pb.RollbackConfigRequest) (*pb.RollbackConfigResponse, error) {
	tenantID := "default" // TODO: 从 context 获取

	// 验证命名空间是否存在
	if err := h.validateNamespace(ctx, tenantID, req.NamespaceId); err != nil {
		return &pb.RollbackConfigResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// 获取目标版本的历史记录
	history, err := h.deps.HistoryDAO.GetHistoryByVersion(ctx, tenantID, req.NamespaceId, req.GroupName, req.ConfigDataId, req.TargetVersion)
	if err != nil {
		return &pb.RollbackConfigResponse{
			Success: false,
			Message: fmt.Sprintf("target version not found: %v", err),
		}, nil
	}

	// 获取当前配置（用于记录回滚前的状态和获取 ContentType）
	currentConfig, err := h.deps.ConfigDAO.GetConfig(ctx, tenantID, req.NamespaceId, req.GroupName, req.ConfigDataId)
	var newVersion int64
	var contentType string
	if err == nil {
		// 配置存在，版本号递增
		newVersion = currentConfig.Version + 1
		contentType = currentConfig.ContentType
	} else {
		// 配置不存在，版本号从 1 开始
		newVersion = 1
		contentType = "" // 如果配置不存在，ContentType 为空（历史记录中可能没有）
	}

	// 创建新配置（基于历史记录）
	now := time.Now()
	changedBy := req.ChangedBy // 从请求中获取变更人
	config := &types.ConfigData{
		TenantId:          tenantID,
		NamespaceId:       history.NamespaceId,
		GroupName:         history.GroupName,
		ConfigDataId:      history.ConfigDataId,
		ContentType:       contentType,
		ConfigContent:     history.NewContent,
		Md5Value:          history.NewMd5Value,
		ConfigDescription: fmt.Sprintf("Rollback from version %d to version %d", newVersion-1, req.TargetVersion),
		Version:           newVersion,
		EditTime:          now,
		EditWho:           changedBy,
	}

	// 如果当前配置存在，保持原创建人；否则设置新的创建人
	if currentConfig != nil {
		config.AddTime = currentConfig.AddTime
		config.AddWho = currentConfig.AddWho
	} else {
		config.AddTime = now
		config.AddWho = changedBy
	}

	// 保存配置
	if err := h.deps.ConfigDAO.SaveConfig(ctx, config); err != nil {
		return &pb.RollbackConfigResponse{
			Success: false,
			Message: fmt.Sprintf("failed to rollback config: %v", err),
		}, nil
	}

	// 保存历史记录
	newHistory := &types.ConfigHistory{
		ConfigHistoryId: random.Generate32BitRandomString(), // 生成唯一的配置历史ID（32位）
		TenantId:        tenantID,
		NamespaceId:     config.NamespaceId,
		GroupName:       config.GroupName,
		ConfigDataId:    config.ConfigDataId,
		ChangeType:      "ROLLBACK",
		NewContent:      config.ConfigContent,
		NewVersion:      config.Version,
		NewMd5Value:     config.Md5Value,
		ChangeReason:    req.ChangeReason,
		ChangedBy:       changedBy,
		ChangedAt:       now,
		AddTime:         now,
		AddWho:          changedBy,
		EditTime:        now,
		EditWho:         changedBy,
	}
	// 如果当前配置存在，记录回滚前的状态
	if currentConfig != nil {
		newHistory.OldContent = currentConfig.ConfigContent
		newHistory.OldVersion = currentConfig.Version
		newHistory.OldMd5Value = currentConfig.Md5Value
	}
	if err := h.deps.HistoryDAO.CreateHistory(ctx, newHistory); err != nil {
		// 历史记录失败不影响主流程，只记录日志
		logger.Warn("保存配置历史记录失败", "error", err,
			"namespaceId", config.NamespaceId,
			"groupName", config.GroupName,
			"configDataId", config.ConfigDataId)
	}

	// 通知监听者
	h.configWatcher.NotifyConfigUpdate(config)

	// 发送配置回滚告警
	if h.deps.ConfigProvider != nil {
		instanceConfig := h.deps.ConfigProvider.GetConfig()
		if instanceConfig != nil {
			configInfo := centerlog.ConfigChangeAlertInfo{
				ChangeType:   "ROLLBACK",
				NamespaceId:  config.NamespaceId,
				GroupName:    config.GroupName,
				ConfigDataId: config.ConfigDataId,
				Version:      config.Version,
				ChangedBy:    config.EditWho,
			}
			centerlog.HandleConfigChange(instanceConfig, configInfo)
		}
	}

	return &pb.RollbackConfigResponse{
		Success:    true,
		Message:    "config rollback successfully",
		NewVersion: newVersion,
		ContentMd5: config.Md5Value, // 返回回滚后的 MD5
	}, nil
}

// 辅助方法

// calculateMD5 计算 MD5
func calculateMD5(content string) string {
	hash := md5.Sum([]byte(content))
	return hex.EncodeToString(hash[:])
}

// convertConfigToProto 转换配置为 protobuf 格式
func convertConfigToProto(config *types.ConfigData) *pb.ConfigData {
	if config == nil {
		return nil
	}
	return &pb.ConfigData{
		NamespaceId:   config.NamespaceId,
		GroupName:     config.GroupName,
		ConfigDataId:  config.ConfigDataId,
		ContentType:   config.ContentType,
		ConfigContent: config.ConfigContent,
		ContentMd5:    config.Md5Value,
		ConfigDesc:    config.ConfigDescription,
		ConfigVersion: config.Version,
	}
}

// convertHistoryToProto 转换历史记录为 protobuf 格式
//
// 处理说明：
//   - 将 Go 类型的 ConfigHistory 转换为 protobuf 格式
//   - 注意：proto 的 ConfigHistory 设计与数据库类型不完全一致
//   - ConfigHistoryId: Go 类型是 string，proto 是 int64（暂时设为 0，需要时再转换）
//   - ContentType: 历史记录中未直接存储，设为空字符串
//   - ChangedBy: 使用 history.ChangedBy（如果存在）
//   - ChangeTime: 使用 history.ChangedAt（如果存在）
func convertHistoryToProto(history *types.ConfigHistory) *pb.ConfigHistory {
	if history == nil {
		return nil
	}

	// 尝试将 ConfigHistoryId 从 string 转换为 int64
	// 如果转换失败，使用 0（表示未设置）
	configHistoryId := int64(0)
	if history.ConfigHistoryId != "" {
		// TODO: 如果 ConfigHistoryId 是数字字符串，可以尝试转换
		// 目前暂时使用 0，因为 proto 定义是 int64，而 Go 类型是 string
		// 如果需要，可以在这里添加 string -> int64 的转换逻辑
	}

	return &pb.ConfigHistory{
		ConfigHistoryId: configHistoryId,
		NamespaceId:     history.NamespaceId,
		GroupName:       history.GroupName,
		ConfigDataId:    history.ConfigDataId,
		ContentType:     "", // 历史记录中未直接存储 ContentType
		ConfigContent:   history.NewContent,
		ContentMd5:      history.NewMd5Value,
		ConfigVersion:   history.NewVersion,
		ChangeType:      history.ChangeType,
		ChangeReason:    history.ChangeReason,
		ChangedBy:       history.ChangedBy,                               // 使用 history.ChangedBy
		ChangeTime:      history.ChangedAt.Format("2006-01-02 15:04:05"), // 使用 history.ChangedAt（转换为字符串格式）
	}
}
