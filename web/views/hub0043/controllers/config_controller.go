package controllers

import (
	"context"
	"time"

	"gateway/internal/servicecenter"
	internaldao "gateway/internal/servicecenter/dao"
	pb "gateway/internal/servicecenter/server/proto"
	"gateway/internal/servicecenter/types"
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/pkg/utils/random"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	hub0043dao "gateway/web/views/hub0043/dao"
	"gateway/web/views/hub0043/models"

	"github.com/gin-gonic/gin"
)

// ConfigController 配置中心控制器
type ConfigController struct {
	db           database.Database
	configDAO    *hub0043dao.ConfigDAO
	historyDAO   *hub0043dao.HistoryDAO
	namespaceDAO *internaldao.NamespaceDAO
}

// NewConfigController 创建配置中心控制器
func NewConfigController(db database.Database) *ConfigController {
	return &ConfigController{
		db:           db,
		configDAO:    hub0043dao.NewConfigDAO(db),
		historyDAO:   hub0043dao.NewHistoryDAO(db),
		namespaceDAO: internaldao.NewNamespaceDAO(db),
	}
}

// QueryConfigs 获取配置列表
// @Summary 获取配置列表
// @Description 分页获取配置列表，支持条件查询
// @Tags 配置中心
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param namespaceId query string false "命名空间ID"
// @Param groupName query string false "分组名称"
// @Param configDataId query string false "配置数据ID（模糊查询）"
// @Param contentType query string false "内容类型"
// @Param activeFlag query string false "活动状态（Y/N）"
// @Success 200 {object} response.JsonData
// @Router /api/hub0043/configs [get]
func (c *ConfigController) QueryConfigs(ctx *gin.Context) {
	// 使用工具类获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)
	// 使用工具类获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 绑定查询条件
	var query models.ConfigQuery
	if err := request.BindSafely(ctx, &query); err != nil {
		logger.WarnWithTrace(ctx, "绑定配置查询条件失败，使用默认条件", "error", err.Error())
	}

	// 验证必填字段
	if query.NamespaceId == "" {
		response.ErrorJSON(ctx, "namespaceId不能为空", constants.ED00006)
		return
	}

	// 使用 hub0043 模块的独立 DAO 层查询配置列表（支持条件查询和分页）
	configs, total, err := c.configDAO.ListConfigs(ctx.Request.Context(), tenantId, &query, page, pageSize)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取配置列表失败", err)
		response.ErrorJSON(ctx, "获取配置列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 直接返回 DAO 查询结果，无需转换
	pageInfo := response.NewPageInfo(page, pageSize, total)
	pageInfo.MainKey = "configDataId"
	response.PageJSON(ctx, configs, pageInfo, constants.SD00002)
}

// GetConfig 获取单个配置详情
// @Summary 获取配置详情
// @Description 根据配置主键获取配置详细信息
// @Tags 配置中心
// @Accept json
// @Produce json
// @Param namespaceId query string true "命名空间ID"
// @Param groupName query string true "分组名称"
// @Param configDataId query string true "配置数据ID"
// @Success 200 {object} response.JsonData
// @Router /api/hub0043/configs/detail [get]
func (c *ConfigController) GetConfig(ctx *gin.Context) {
	tenantId := request.GetTenantID(ctx)
	namespaceId := request.GetParam(ctx, "namespaceId")
	groupName := request.GetParam(ctx, "groupName")
	configDataId := request.GetParam(ctx, "configDataId")

	// 验证必填字段
	if namespaceId == "" || groupName == "" || configDataId == "" {
		response.ErrorJSON(ctx, "namespaceId、groupName和configDataId不能为空", constants.ED00006)
		return
	}

	// 使用 DAO 直接查询配置
	config, err := c.configDAO.GetConfigById(ctx.Request.Context(), tenantId, namespaceId, groupName, configDataId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取配置详情失败", err)
		response.ErrorJSON(ctx, "获取配置详情失败: "+err.Error(), constants.ED00009)
		return
	}

	if config == nil {
		response.ErrorJSON(ctx, "配置不存在", constants.ED00008)
		return
	}

	// 直接返回 DAO 查询结果，无需转换
	response.SuccessJSON(ctx, config, constants.SD00001)
}

// AddConfig 创建配置
// @Summary 创建配置
// @Description 创建新的配置
// @Tags 配置中心
// @Accept json
// @Produce json
// @Param config body types.ConfigData true "配置信息"
// @Success 200 {object} response.JsonData
// @Router /api/hub0043/configs [post]
func (c *ConfigController) AddConfig(ctx *gin.Context) {
	tenantId := request.GetTenantID(ctx)
	requestCtx := ctx.Request.Context()

	var config types.ConfigData
	if err := request.BindSafely(ctx, &config); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 设置租户ID
	config.TenantId = tenantId

	// 验证必填字段
	if config.NamespaceId == "" {
		response.ErrorJSON(ctx, "命名空间ID不能为空", constants.ED00006)
		return
	}
	if config.ConfigDataId == "" {
		response.ErrorJSON(ctx, "配置数据ID不能为空", constants.ED00006)
		return
	}
	if config.ConfigContent == "" {
		response.ErrorJSON(ctx, "配置内容不能为空", constants.ED00006)
		return
	}

	// 设置默认值
	if config.GroupName == "" {
		config.GroupName = "DEFAULT_GROUP"
	}
	if config.ContentType == "" {
		config.ContentType = "text"
	}

	// 检查配置是否已存在
	existingConfig, err := c.configDAO.GetConfigById(requestCtx, tenantId, config.NamespaceId, config.GroupName, config.ConfigDataId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询配置失败", err)
		response.ErrorJSON(ctx, "查询配置失败: "+err.Error(), constants.ED00009)
		return
	}
	if existingConfig != nil {
		response.ErrorJSON(ctx, "配置已存在", constants.ED00009)
		return
	}

	// 设置初始版本和时间
	config.Version = 1
	now := time.Now()
	config.AddTime = now
	config.EditTime = now

	// 设置创建人和修改人
	operatorId := request.GetOperatorID(ctx)
	config.AddWho = operatorId
	config.EditWho = operatorId

	// 设置默认值
	if config.ActiveFlag == "" {
		config.ActiveFlag = "Y"
	}

	// 使用 hub0043 DAO 插入配置（会自动计算 MD5）
	if err := c.configDAO.InsertConfig(requestCtx, &config); err != nil {
		logger.ErrorWithTrace(ctx, "创建配置失败", err)
		response.ErrorJSON(ctx, "创建配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询新创建的配置信息（获取数据库中的最新数据）
	newConfig, err := c.configDAO.GetConfigById(requestCtx, tenantId, config.NamespaceId, config.GroupName, config.ConfigDataId)
	if err != nil {
		logger.WarnWithTrace(ctx, "获取新创建的配置信息失败", err)
		// 即使查询失败，也返回基本信息
		response.SuccessJSON(ctx, gin.H{
			"configDataId": config.ConfigDataId,
			"namespaceId":  config.NamespaceId,
			"groupName":    config.GroupName,
		}, constants.SD00003)
		return
	}

	// 通过 manager 发布事件通知
	c.notifyConfigChange(requestCtx, tenantId, config.NamespaceId, newConfig, "CONFIG_UPDATED")

	// 返回完整的配置信息
	response.SuccessJSON(ctx, newConfig, constants.SD00003)
}

// EditConfig 更新配置
// @Summary 更新配置
// @Description 更新配置信息
// @Tags 配置中心
// @Accept json
// @Produce json
// @Param config body types.ConfigData true "配置信息"
// @Success 200 {object} response.JsonData
// @Router /api/hub0043/configs [put]
func (c *ConfigController) EditConfig(ctx *gin.Context) {
	tenantId := request.GetTenantID(ctx)
	requestCtx := ctx.Request.Context()

	var config types.ConfigData
	if err := request.BindSafely(ctx, &config); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 设置租户ID
	config.TenantId = tenantId

	// 验证必填字段
	if config.NamespaceId == "" || config.GroupName == "" || config.ConfigDataId == "" {
		response.ErrorJSON(ctx, "namespaceId、groupName和configDataId不能为空", constants.ED00006)
		return
	}
	if config.ConfigContent == "" {
		response.ErrorJSON(ctx, "配置内容不能为空", constants.ED00006)
		return
	}

	// 获取当前配置（用于记录历史）
	oldConfig, err := c.configDAO.GetConfigById(requestCtx, tenantId, config.NamespaceId, config.GroupName, config.ConfigDataId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询配置失败", err)
		response.ErrorJSON(ctx, "查询配置失败: "+err.Error(), constants.ED00009)
		return
	}
	if oldConfig == nil {
		response.ErrorJSON(ctx, "配置不存在", constants.ED00008)
		return
	}

	// 设置版本号和时间（UpdateConfig 会自动递增版本号）
	config.Version = oldConfig.Version
	config.AddTime = oldConfig.AddTime
	config.AddWho = oldConfig.AddWho // 保持原创建人
	config.EditTime = time.Now()
	config.EditWho = request.GetOperatorID(ctx) // 设置修改人

	// 先更新配置（UpdateConfig 会自动递增版本号）
	if err := c.configDAO.UpdateConfig(requestCtx, &config); err != nil {
		logger.ErrorWithTrace(ctx, "更新配置失败", err)
		response.ErrorJSON(ctx, "更新配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询更新后的配置信息（获取数据库中的最新数据，包括更新后的版本号等）
	updatedConfig, err := c.configDAO.GetConfigById(requestCtx, tenantId, config.NamespaceId, config.GroupName, config.ConfigDataId)
	if err != nil {
		logger.WarnWithTrace(ctx, "获取更新后的配置信息失败", err)
		// 即使查询失败，也返回基本信息
		response.SuccessJSON(ctx, gin.H{
			"configDataId": config.ConfigDataId,
			"namespaceId":  config.NamespaceId,
			"groupName":    config.GroupName,
		}, constants.SD00004)
		return
	}

	// 配置更新成功后，在事务中保存历史记录
	changeReason := request.GetParam(ctx, "changeReason") // 从请求参数中获取变更原因
	operatorId := request.GetOperatorID(ctx)
	if err := c.db.InTx(requestCtx, nil, func(txCtx context.Context) error {
		// 根据原配置和更新后的配置生成历史记录
		now := time.Now()
		history := &types.ConfigHistory{
			ConfigHistoryId: random.Generate32BitRandomString(),
			TenantId:        tenantId,
			NamespaceId:     config.NamespaceId,
			GroupName:       config.GroupName,
			ConfigDataId:    config.ConfigDataId,
			ChangeType:      types.ChangeTypeUpdate,
			OldContent:      oldConfig.ConfigContent,
			OldVersion:      oldConfig.Version,
			OldMd5Value:     oldConfig.Md5Value,
			NewContent:      updatedConfig.ConfigContent,
			NewVersion:      updatedConfig.Version,
			NewMd5Value:     updatedConfig.Md5Value,
			ChangeReason:    changeReason,
			ChangedBy:       operatorId,
			ChangedAt:       now,
			AddTime:         now,
			AddWho:          operatorId,
			EditTime:        now,
			EditWho:         operatorId,
		}

		return c.historyDAO.CreateHistory(txCtx, history)
	}); err != nil {
		logger.WarnWithTrace(ctx, "保存配置历史记录失败", err)
		// 历史记录失败不影响主流程，只记录日志
	}

	// 通过 manager 发布事件通知
	c.notifyConfigChange(requestCtx, tenantId, config.NamespaceId, updatedConfig, "CONFIG_UPDATED")

	// 返回更新后的完整配置信息
	response.SuccessJSON(ctx, updatedConfig, constants.SD00004)
}

// DeleteConfig 删除配置
// @Summary 删除配置
// @Description 删除配置
// @Tags 配置中心
// @Accept json
// @Produce json
// @Param namespaceId query string true "命名空间ID"
// @Param groupName query string true "分组名称"
// @Param configDataId query string true "配置数据ID"
// @Success 200 {object} response.JsonData
// @Router /api/hub0043/configs [delete]
func (c *ConfigController) DeleteConfig(ctx *gin.Context) {
	tenantId := request.GetTenantID(ctx)
	requestCtx := ctx.Request.Context()

	namespaceId := request.GetParam(ctx, "namespaceId")
	groupName := request.GetParam(ctx, "groupName")
	configDataId := request.GetParam(ctx, "configDataId")

	// 验证必填字段
	if namespaceId == "" || groupName == "" || configDataId == "" {
		response.ErrorJSON(ctx, "namespaceId、groupName和configDataId不能为空", constants.ED00006)
		return
	}

	// 获取当前配置（用于记录历史）
	oldConfig, err := c.configDAO.GetConfigById(requestCtx, tenantId, namespaceId, groupName, configDataId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询配置失败", err)
		response.ErrorJSON(ctx, "查询配置失败: "+err.Error(), constants.ED00009)
		return
	}
	if oldConfig == nil {
		response.ErrorJSON(ctx, "配置不存在", constants.ED00008)
		return
	}

	// 根据原配置生成历史记录
	now := time.Now()
	operatorId := request.GetOperatorID(ctx)
	history := &types.ConfigHistory{
		ConfigHistoryId: random.Generate32BitRandomString(),
		TenantId:        tenantId,
		NamespaceId:     namespaceId,
		GroupName:       groupName,
		ConfigDataId:    configDataId,
		ChangeType:      types.ChangeTypeDelete,
		OldContent:      oldConfig.ConfigContent,
		OldVersion:      oldConfig.Version,
		OldMd5Value:     oldConfig.Md5Value,
		ChangeReason:    "",
		ChangedBy:       operatorId,
		ChangedAt:       now,
		AddTime:         now,
		AddWho:          operatorId,
		EditTime:        now,
		EditWho:         operatorId,
	}

	// 在事务中先保存历史记录，再删除配置
	if err := c.db.InTx(requestCtx, nil, func(txCtx context.Context) error {
		// 1. 先保存历史记录
		if err := c.historyDAO.CreateHistory(txCtx, history); err != nil {
			return err
		}

		// 2. 历史记录保存成功后，再删除配置
		return c.configDAO.DeleteConfig(txCtx, tenantId, namespaceId, groupName, configDataId)
	}); err != nil {
		logger.ErrorWithTrace(ctx, "删除配置失败", err)
		response.ErrorJSON(ctx, "删除配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 通过 manager 发布事件通知
	c.notifyConfigChange(requestCtx, tenantId, namespaceId, oldConfig, "CONFIG_DELETED")

	response.SuccessJSON(ctx, map[string]interface{}{
		"namespaceId":  namespaceId,
		"groupName":    groupName,
		"configDataId": configDataId,
		"message":      "配置删除成功",
	}, constants.SD00005)
}

// notifyConfigChange 通过 manager 发布配置变更事件通知
func (c *ConfigController) notifyConfigChange(ctx context.Context, tenantId, namespaceId string, config *types.ConfigData, eventType string) {
	// 构建事件
	event := &pb.ConfigChangeEvent{
		EventType:    eventType,
		Timestamp:    time.Now().Format("2006-01-02 15:04:05"),
		NamespaceId:  config.NamespaceId,
		GroupName:    config.GroupName,
		ConfigDataId: config.ConfigDataId,
		ContentMd5:   config.Md5Value,
	}

	// 如果是更新事件，包含配置数据
	if eventType == "CONFIG_UPDATED" && config != nil {
		event.Config = &pb.ConfigData{
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

	// 通过 ServiceCenterManager 发布事件通知
	// 从命名空间获取 instanceName
	if servicecenter.GetManager() != nil {
		// 查询命名空间获取 instanceName
		namespace, err := c.namespaceDAO.GetNamespace(ctx, tenantId, namespaceId)
		if err != nil {
			logger.WarnWithTrace(ctx, "查询命名空间失败，跳过配置变更事件通知", err,
				"namespaceId", namespaceId,
				"configDataId", config.ConfigDataId)
			return
		}
		if namespace == nil {
			logger.WarnWithTrace(ctx, "命名空间不存在，跳过配置变更事件通知",
				"namespaceId", namespaceId,
				"configDataId", config.ConfigDataId)
			return
		}
		if namespace.InstanceName == "" {
			logger.WarnWithTrace(ctx, "命名空间的 instanceName 为空，跳过配置变更事件通知",
				"namespaceId", namespaceId,
				"configDataId", config.ConfigDataId)
			return
		}

		if err := servicecenter.GetManager().NotifyConfigChange(ctx, namespace.InstanceName, tenantId, config.NamespaceId, config.GroupName, config.ConfigDataId, event); err != nil {
			logger.WarnWithTrace(ctx, "发送配置变更事件通知失败", err,
				"instanceName", namespace.InstanceName,
				"namespaceId", config.NamespaceId,
				"configDataId", config.ConfigDataId)
		}
	}
}
