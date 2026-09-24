package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gateway/internal/servicecenterv3"
	"gateway/internal/servicecenterv3/catalog"
	"gateway/internal/servicecenterv3/contract"
	"gateway/internal/servicecenterv3/model"
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/middleware/audit"
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
	namespaceDAO *catalog.NamespaceDAO
}

// NewConfigController 创建配置中心控制器
func NewConfigController(db database.Database) *ConfigController {
	return &ConfigController{
		db:           db,
		configDAO:    hub0043dao.NewConfigDAO(db),
		historyDAO:   hub0043dao.NewHistoryDAO(db),
		namespaceDAO: catalog.NewNamespaceDAO(db),
	}
}

// QueryConfigs 获取配置列表
// @Summary 获取配置列表
// @Description 分页获取配置列表，支持条件查询
// @Tags 配置中心
// @Produce json
// @Param pageIndex query int false "页码" default(1)
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

	rows := make([]map[string]interface{}, 0, len(configs))
	svc, instanceName, _ := c.lookupV3(ctx.Request.Context(), tenantId, query.NamespaceId)
	for _, cfg := range configs {
		row := configToMap(cfg)
		c.overlayDraftStatus(ctx, tenantId, svc, instanceName, cfg.NamespaceId, cfg.GroupName, cfg.ConfigDataId, row)
		rows = append(rows, row)
	}

	pageInfo := response.NewPageInfo(page, pageSize, total)
	pageInfo.MainKey = "configDataId"
	response.PageJSON(ctx, rows, pageInfo, constants.SD00002)
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

	row := configToMap(config)
	svc, instanceName, _ := c.lookupV3(ctx.Request.Context(), tenantId, namespaceId)
	c.overlayDraftContent(ctx, tenantId, svc, instanceName, namespaceId, groupName, configDataId, row)
	response.SuccessJSON(ctx, row, constants.SD00001)
}

// AddConfig 创建配置
// @Summary 创建配置
// @Description 创建新的配置
// @Tags 配置中心
// @Accept json
// @Produce json
// @Param config body catalog.ConfigData true "配置信息"
// @Success 200 {object} response.JsonData
// @Router /api/hub0043/configs [post]
func (c *ConfigController) AddConfig(ctx *gin.Context) {
	tenantId := request.GetTenantID(ctx)
	requestCtx := ctx.Request.Context()

	var config catalog.ConfigData
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

	published, err := c.publishNewOnV3(ctx, requestCtx, tenantId, operatorId, &config)
	if err != nil {
		response.ErrorJSON(ctx, err.Error(), constants.ED00009)
		return
	}
	audit.SetEvent(ctx, &audit.AuditEvent{
		Action:       audit.AuditActionCreate,
		ModuleCode:   "hub0043",
		TargetType:   "CONFIG",
		TargetId:     config.ConfigDataId,
		TargetName:   config.GroupName + "/" + config.ConfigDataId,
		ResourceCode: "hub0043:add",
	})
	response.SuccessJSON(ctx, published, constants.SD00003)
}

// EditConfig 更新配置
// @Summary 更新配置
// @Description 更新配置信息
// @Tags 配置中心
// @Accept json
// @Produce json
// @Param config body catalog.ConfigData true "配置信息"
// @Success 200 {object} response.JsonData
// @Router /api/hub0043/configs [put]
func (c *ConfigController) EditConfig(ctx *gin.Context) {
	tenantId := request.GetTenantID(ctx)
	requestCtx := ctx.Request.Context()

	var config catalog.ConfigData
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

	draft, err := c.saveDraftOnV3(ctx, requestCtx, tenantId, request.GetOperatorID(ctx), &config)
	if err != nil {
		response.ErrorJSON(ctx, err.Error(), constants.ED00009)
		return
	}
	audit.SetEvent(ctx, &audit.AuditEvent{
		Action:       audit.AuditActionUpdate,
		ModuleCode:   "hub0043",
		TargetType:   "CONFIG",
		TargetId:     config.ConfigDataId,
		TargetName:   config.GroupName + "/" + config.ConfigDataId,
		ResourceCode: "hub0043:edit",
	})
	response.SuccessJSON(ctx, draft, constants.SD00004)
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

	if err := c.deleteOnV3(ctx, requestCtx, tenantId, namespaceId, groupName, configDataId); err != nil {
		response.ErrorJSON(ctx, err.Error(), constants.ED00009)
		return
	}
	audit.SetEvent(ctx, &audit.AuditEvent{
		Action:       audit.AuditActionDelete,
		ModuleCode:   "hub0043",
		TargetType:   "CONFIG",
		TargetId:     configDataId,
		TargetName:   groupName + "/" + configDataId,
		ResourceCode: "hub0043:delete",
	})
	response.SuccessJSON(ctx, map[string]interface{}{
		"namespaceId":  namespaceId,
		"groupName":    groupName,
		"configDataId": configDataId,
		"message":      "配置删除成功",
	}, constants.SD00005)
}

// SaveDraft 保存配置草稿（仅 v3；legacy 回落到编辑即发布）。
func (c *ConfigController) SaveDraft(ctx *gin.Context) {
	c.editOrSaveDraft(ctx, false)
}

// PublishConfig 将草稿发布为新版本并推送订阅方。
func (c *ConfigController) PublishConfig(ctx *gin.Context) {
	tenantId := request.GetTenantID(ctx)
	requestCtx := ctx.Request.Context()
	var config catalog.ConfigData
	if err := request.BindSafely(ctx, &config); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}
	config.TenantId = tenantId
	if config.NamespaceId == "" || config.GroupName == "" || config.ConfigDataId == "" {
		response.ErrorJSON(ctx, "namespaceId、groupName和configDataId不能为空", constants.ED00006)
		return
	}
	if config.GroupName == "" {
		config.GroupName = "DEFAULT_GROUP"
	}
	operatorId := request.GetOperatorID(ctx)
	reason := request.GetParam(ctx, "changeReason")
	published, err := c.publishOnV3(ctx, requestCtx, tenantId, operatorId, reason, &config)
	if err != nil {
		response.ErrorJSON(ctx, err.Error(), constants.ED00009)
		return
	}
	audit.SetEvent(ctx, &audit.AuditEvent{
		Action:       audit.AuditActionUpdate,
		ModuleCode:   "hub0043",
		TargetType:   "CONFIG",
		TargetId:     config.ConfigDataId,
		TargetName:   config.GroupName + "/" + config.ConfigDataId,
		ResourceCode: "hub0043:edit",
		Detail:       "publish",
	})
	response.SuccessJSON(ctx, published, constants.SD00004)
}

// GetDraft 读取未发布草稿（库表）。
func (c *ConfigController) GetDraft(ctx *gin.Context) {
	tenantId := request.GetTenantID(ctx)
	namespaceId := request.GetParam(ctx, "namespaceId")
	groupName := request.GetParam(ctx, "groupName")
	configDataId := request.GetParam(ctx, "configDataId")
	if namespaceId == "" || configDataId == "" {
		response.ErrorJSON(ctx, "namespaceId和configDataId不能为空", constants.ED00006)
		return
	}
	if groupName == "" {
		groupName = "DEFAULT_GROUP"
	}
	cfg, instanceName, err := c.lookupV3(ctx.Request.Context(), tenantId, namespaceId)
	if err != nil || cfg == nil {
		response.ErrorJSON(ctx, "当前引擎未启用草稿（需要 servicecenterv3 且实例已启动）", constants.ED00009)
		return
	}
	draft, err := cfg.GetDraft(ctx.Request.Context(), configCC(tenantId, instanceName, namespaceId, ""), groupName, configDataId)
	if err != nil {
		if errors.Is(err, contract.ErrConfigNotFound) {
			response.ErrorJSON(ctx, "草稿不存在", constants.ED00008)
			return
		}
		response.ErrorJSON(ctx, "读取草稿失败: "+err.Error(), constants.ED00009)
		return
	}
	response.SuccessJSON(ctx, draftToMap(draft), constants.SD00001)
}

func (c *ConfigController) editOrSaveDraft(ctx *gin.Context, _ bool) {
	c.EditConfig(ctx)
}

func (c *ConfigController) publishNewOnV3(ctx *gin.Context, requestCtx context.Context, tenantId, operatorId string, config *catalog.ConfigData) (map[string]interface{}, error) {
	svc, instanceName, err := c.lookupV3(requestCtx, tenantId, config.NamespaceId)
	if err != nil {
		return nil, err
	}
	if svc == nil {
		return nil, fmt.Errorf("服务中心未初始化")
	}
	cc := configCC(tenantId, instanceName, config.NamespaceId, operatorId)
	if err := svc.SaveDraft(requestCtx, cc, toDraft(config)); err != nil {
		return nil, fmt.Errorf("保存配置草稿失败: %w", err)
	}
	if _, err := svc.Publish(requestCtx, cc, config.GroupName, config.ConfigDataId, "create"); err != nil {
		return nil, fmt.Errorf("发布配置失败: %w", err)
	}
	saved, err := c.configDAO.GetConfigById(requestCtx, tenantId, config.NamespaceId, config.GroupName, config.ConfigDataId)
	if err != nil || saved == nil {
		return map[string]interface{}{
			"configDataId":  config.ConfigDataId,
			"namespaceId":   config.NamespaceId,
			"groupName":     config.GroupName,
			"hasDraft":      false,
			"publishStatus": "published",
			"engine":        servicecenterv3.EngineV3,
		}, nil
	}
	row := configToMap(saved)
	row["hasDraft"] = false
	row["publishStatus"] = "published"
	row["engine"] = servicecenterv3.EngineV3
	return row, nil
}

func (c *ConfigController) saveDraftOnV3(ctx *gin.Context, requestCtx context.Context, tenantId, operatorId string, config *catalog.ConfigData) (map[string]interface{}, error) {
	svc, instanceName, err := c.lookupV3(requestCtx, tenantId, config.NamespaceId)
	if err != nil {
		return nil, err
	}
	if svc == nil {
		return nil, fmt.Errorf("服务中心未初始化")
	}
	cc := configCC(tenantId, instanceName, config.NamespaceId, operatorId)
	if err := svc.SaveDraft(requestCtx, cc, toDraft(config)); err != nil {
		return nil, fmt.Errorf("保存配置草稿失败: %w", err)
	}
	saved, _ := c.configDAO.GetConfigById(requestCtx, tenantId, config.NamespaceId, config.GroupName, config.ConfigDataId)
	var row map[string]interface{}
	if saved != nil {
		row = configToMap(saved)
	} else {
		row = map[string]interface{}{
			"configDataId": config.ConfigDataId,
			"namespaceId":  config.NamespaceId,
			"groupName":    config.GroupName,
		}
	}
	row["configContent"] = config.ConfigContent
	row["contentType"] = config.ContentType
	row["configDescription"] = config.ConfigDescription
	row["hasDraft"] = true
	row["publishStatus"] = "draft"
	row["engine"] = servicecenterv3.EngineV3
	return row, nil
}

func (c *ConfigController) publishOnV3(ctx *gin.Context, requestCtx context.Context, tenantId, operatorId, reason string, config *catalog.ConfigData) (map[string]interface{}, error) {
	svc, instanceName, err := c.lookupV3(requestCtx, tenantId, config.NamespaceId)
	if err != nil {
		return nil, err
	}
	if svc == nil {
		return nil, fmt.Errorf("服务中心未初始化")
	}
	cc := configCC(tenantId, instanceName, config.NamespaceId, operatorId)
	if config.ConfigContent != "" {
		if err := svc.SaveDraft(requestCtx, cc, toDraft(config)); err != nil {
			return nil, fmt.Errorf("发布前保存草稿失败: %w", err)
		}
	}
	if reason == "" {
		reason = "publish"
	}
	if _, err := svc.Publish(requestCtx, cc, config.GroupName, config.ConfigDataId, reason); err != nil {
		return nil, fmt.Errorf("发布配置失败: %w", err)
	}
	saved, err := c.configDAO.GetConfigById(requestCtx, tenantId, config.NamespaceId, config.GroupName, config.ConfigDataId)
	if err != nil || saved == nil {
		return nil, fmt.Errorf("发布成功但读取已发布配置失败")
	}
	row := configToMap(saved)
	row["hasDraft"] = false
	row["publishStatus"] = "published"
	row["engine"] = servicecenterv3.EngineV3
	return row, nil
}

func (c *ConfigController) deleteOnV3(ctx *gin.Context, requestCtx context.Context, tenantId, namespaceId, groupName, dataID string) error {
	svc, instanceName, err := c.lookupV3(requestCtx, tenantId, namespaceId)
	if err != nil {
		return err
	}
	if svc == nil {
		return fmt.Errorf("服务中心未初始化")
	}
	cc := configCC(tenantId, instanceName, namespaceId, request.GetOperatorID(ctx))
	if err := svc.Delete(requestCtx, cc, groupName, dataID); err != nil {
		return fmt.Errorf("删除配置失败: %w", err)
	}
	return nil
}

func (c *ConfigController) overlayDraftStatus(ctx *gin.Context, tenantId string, svc contract.Config, instanceName, namespaceId, groupName, dataID string, row map[string]interface{}) {
	row["hasDraft"] = false
	row["publishStatus"] = "published"
	row["engine"] = engineName()
	if svc == nil {
		return
	}
	if _, err := svc.GetDraft(ctx.Request.Context(), configCC(tenantId, instanceName, namespaceId, ""), groupName, dataID); err == nil {
		row["hasDraft"] = true
		row["publishStatus"] = "draft"
	}
}

func (c *ConfigController) overlayDraftContent(ctx *gin.Context, tenantId string, svc contract.Config, instanceName, namespaceId, groupName, dataID string, row map[string]interface{}) {
	c.overlayDraftStatus(ctx, tenantId, svc, instanceName, namespaceId, groupName, dataID, row)
	if svc == nil {
		return
	}
	draft, err := svc.GetDraft(ctx.Request.Context(), configCC(tenantId, instanceName, namespaceId, ""), groupName, dataID)
	if err != nil || draft == nil {
		return
	}
	row["publishedContent"] = row["configContent"]
	row["configContent"] = draft.Content
	if draft.ContentType != "" {
		row["contentType"] = draft.ContentType
	}
	if draft.Description != "" {
		row["configDescription"] = draft.Description
	}
	row["hasDraft"] = true
	row["publishStatus"] = "draft"
}

func toDraft(config *catalog.ConfigData) *model.ConfigDraft {
	return &model.ConfigDraft{
		TenantID:    config.TenantId,
		NamespaceID: config.NamespaceId,
		GroupName:   config.GroupName,
		DataID:      config.ConfigDataId,
		Content:     config.ConfigContent,
		ContentType: config.ContentType,
		Description: config.ConfigDescription,
	}
}

func draftToMap(d *model.ConfigDraft) map[string]interface{} {
	if d == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"namespaceId":       d.NamespaceID,
		"groupName":         d.GroupName,
		"configDataId":      d.DataID,
		"configContent":     d.Content,
		"contentType":       d.ContentType,
		"configDescription": d.Description,
		"hasDraft":          true,
		"publishStatus":     "draft",
	}
}

func configToMap(config *catalog.ConfigData) map[string]interface{} {
	if config == nil {
		return map[string]interface{}{}
	}
	raw, err := json.Marshal(config)
	if err != nil {
		return map[string]interface{}{
			"configDataId": config.ConfigDataId,
			"namespaceId":  config.NamespaceId,
			"groupName":    config.GroupName,
		}
	}
	out := map[string]interface{}{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return map[string]interface{}{
			"configDataId": config.ConfigDataId,
			"namespaceId":  config.NamespaceId,
			"groupName":    config.GroupName,
		}
	}
	return out
}
