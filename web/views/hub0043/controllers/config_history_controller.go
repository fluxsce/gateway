package controllers

import (
	"context"
	"fmt"

	"gateway/internal/servicecenterv3"
	"gateway/internal/servicecenterv3/catalog"
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

// ConfigHistoryController 配置历史控制器
type ConfigHistoryController struct {
	db           database.Database
	historyDAO   *hub0043dao.HistoryDAO
	configDAO    *hub0043dao.ConfigDAO
	namespaceDAO *catalog.NamespaceDAO
}

// NewConfigHistoryController 创建配置历史控制器
func NewConfigHistoryController(db database.Database) *ConfigHistoryController {
	return &ConfigHistoryController{
		db:           db,
		historyDAO:   hub0043dao.NewHistoryDAO(db),
		configDAO:    hub0043dao.NewConfigDAO(db),
		namespaceDAO: catalog.NewNamespaceDAO(db),
	}
}

// GetConfigHistory 获取配置历史
// @Summary 获取配置历史
// @Description 获取配置的变更历史记录
// @Tags 配置中心-历史
// @Accept json
// @Produce json
// @Param namespaceId query string true "命名空间ID"
// @Param groupName query string true "分组名称"
// @Param configDataId query string true "配置数据ID"
// @Param limit query int false "限制数量" default(50)
// @Success 200 {object} response.JsonData
// @Router /api/hub0043/queryConfigHistory [post]
func (c *ConfigHistoryController) GetConfigHistory(ctx *gin.Context) {
	// 使用工具类获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 直接使用 request 绑定查询对象
	var req models.ConfigHistoryRequest
	if err := request.BindSafely(ctx, &req); err != nil {
		logger.WarnWithTrace(ctx, "绑定配置历史查询条件失败，使用默认条件", "error", err.Error())
	}

	// 验证必填字段
	if req.NamespaceId == "" || req.GroupName == "" || req.ConfigDataId == "" {
		response.ErrorJSON(ctx, "namespaceId、groupName和configDataId不能为空", constants.ED00006)
		return
	}

	// 设置默认限制数量
	if req.Limit <= 0 {
		req.Limit = 50
	}

	// 使用 hub0043 模块的独立历史 DAO 查询配置历史
	// 注意：列表查询不包含大字段 newContent 和 oldContent，减少内存开销
	// 需要查看完整内容时，请调用 GetHistoryById 接口获取详情
	histories, err := c.historyDAO.GetConfigHistory(ctx.Request.Context(), tenantId, &req)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取配置历史失败", err)
		response.ErrorJSON(ctx, "获取配置历史失败: "+err.Error(), constants.ED00009)
		return
	}

	// 直接返回 DAO 查询结果，无需转换
	response.SuccessJSON(ctx, histories, constants.SD00001)
}

// GetHistoryById 根据历史配置ID获取配置历史详情
// @Summary 根据历史配置ID获取配置历史详情
// @Description 根据历史配置ID获取完整的配置历史记录，包含变更前后的完整内容
// @Tags 配置中心-历史
// @Accept json
// @Produce json
// @Param configHistoryId query string true "配置历史ID"
// @Success 200 {object} response.JsonData
// @Router /api/hub0043/configHistory/detail [get]
func (c *ConfigHistoryController) GetHistoryById(ctx *gin.Context) {
	// 使用工具类获取租户ID
	tenantId := request.GetTenantID(ctx)

	configHistoryId := request.GetParam(ctx, "configHistoryId")

	// 验证必填字段
	if configHistoryId == "" {
		response.ErrorJSON(ctx, "configHistoryId不能为空", constants.ED00006)
		return
	}

	// 使用 hub0043 模块的独立历史 DAO 查询配置历史详情
	// 详情查询包含完整的大字段内容（newContent 和 oldContent）
	history, err := c.historyDAO.GetHistoryById(ctx.Request.Context(), tenantId, configHistoryId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取配置历史详情失败", err)
		response.ErrorJSON(ctx, "获取配置历史详情失败: "+err.Error(), constants.ED00009)
		return
	}

	if history == nil {
		response.ErrorJSON(ctx, "未找到指定的配置历史", constants.ED00008)
		return
	}

	// 直接返回 DAO 查询结果，无需转换
	response.SuccessJSON(ctx, history, constants.SD00001)
}

// RollbackConfig 回滚配置
// @Summary 回滚配置
// @Description 根据历史配置ID将配置回滚到指定版本
// @Tags 配置中心-历史
// @Accept json
// @Produce json
// @Param rollback body models.RollbackRequest true "回滚请求"
// @Success 200 {object} response.JsonData
// @Router /api/hub0043/configHistory/rollback [post]
func (c *ConfigHistoryController) RollbackConfig(ctx *gin.Context) {
	var req models.RollbackRequest
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 验证必填字段
	if req.ConfigHistoryId == "" {
		response.ErrorJSON(ctx, "configHistoryId不能为空", constants.ED00006)
		return
	}

	tenantId := request.GetTenantID(ctx)
	requestCtx := ctx.Request.Context()

	// 根据历史配置唯一ID查询历史记录
	history, err := c.historyDAO.GetHistoryById(requestCtx, tenantId, req.ConfigHistoryId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询历史配置失败", err)
		response.ErrorJSON(ctx, "查询历史配置失败: "+err.Error(), constants.ED00009)
		return
	}
	if history == nil {
		response.ErrorJSON(ctx, "未找到指定的历史配置记录", constants.ED00008)
		return
	}

	operatorId := request.GetOperatorID(ctx)
	rolled, err := c.rollbackOnV3(ctx, requestCtx, tenantId, operatorId, req.ChangeReason, history)
	if err != nil {
		logger.ErrorWithTrace(ctx, "回滚配置失败", err)
		response.ErrorJSON(ctx, "回滚配置失败: "+err.Error(), constants.ED00009)
		return
	}
	audit.SetEvent(ctx, &audit.AuditEvent{
		Action:       audit.AuditActionRollback,
		ModuleCode:   "hub0043",
		TargetType:   "CONFIG",
		TargetId:     history.ConfigDataId,
		TargetName:   history.GroupName + "/" + history.ConfigDataId,
		ResourceCode: "hub0043:history:rollback",
	})
	response.SuccessJSON(ctx, rolled, constants.SD00004)
}

func (c *ConfigHistoryController) rollbackOnV3(ctx *gin.Context, requestCtx context.Context, tenantId, operatorId, reason string, history *catalog.ConfigHistory) (map[string]interface{}, error) {
	if history == nil {
		return nil, fmt.Errorf("历史记录为空")
	}
	if servicecenterv3.GetPool() == nil {
		return nil, fmt.Errorf("服务中心未初始化")
	}
	ns, err := c.namespaceDAO.GetNamespace(requestCtx, tenantId, history.NamespaceId)
	if err != nil {
		return nil, fmt.Errorf("查询命名空间失败: %w", err)
	}
	if ns == nil || ns.InstanceName == "" {
		return nil, fmt.Errorf("命名空间未绑定服务中心实例")
	}
	svc, ok := lookupConfig(ns.InstanceName, ns.Environment)
	if !ok {
		return nil, fmt.Errorf("服务中心实例未运行")
	}
	if reason == "" {
		reason = fmt.Sprintf("rollback to version %d", history.NewVersion)
	}
	cc := configCC(tenantId, ns.InstanceName, history.NamespaceId, operatorId)
	rel, err := svc.Rollback(requestCtx, cc, history.GroupName, history.ConfigDataId, history.NewVersion, reason)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"configHistoryId": history.ConfigHistoryId,
		"namespaceId":     history.NamespaceId,
		"groupName":       history.GroupName,
		"configDataId":    history.ConfigDataId,
		"targetVersion":   history.NewVersion,
		"newVersion":      rel.Version,
		"contentMd5":      rel.MD5,
		"engine":          servicecenterv3.EngineV3,
		"message":         "配置回滚成功",
	}, nil
}
