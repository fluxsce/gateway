package controllers

import (
	"context"
	"fmt"
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

// ConfigHistoryController 配置历史控制器
type ConfigHistoryController struct {
	db           database.Database
	historyDAO   *hub0043dao.HistoryDAO
	configDAO    *hub0043dao.ConfigDAO
	namespaceDAO *internaldao.NamespaceDAO
}

// NewConfigHistoryController 创建配置历史控制器
func NewConfigHistoryController(db database.Database) *ConfigHistoryController {
	return &ConfigHistoryController{
		db:           db,
		historyDAO:   hub0043dao.NewHistoryDAO(db),
		configDAO:    hub0043dao.NewConfigDAO(db),
		namespaceDAO: internaldao.NewNamespaceDAO(db),
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

	// 获取当前配置（用于记录回滚前的状态和获取 ContentType）
	currentConfig, err := c.configDAO.GetConfigById(requestCtx, tenantId, history.NamespaceId, history.GroupName, history.ConfigDataId)
	var contentType string
	if err == nil && currentConfig != nil {
		contentType = currentConfig.ContentType
	} else {
		contentType = "" // 如果配置不存在，ContentType 为空（历史记录中可能没有）
	}

	// 创建新配置（基于历史记录）
	// 注意：SaveConfig 会自动处理版本号递增和 MD5 计算
	operatorId := request.GetOperatorID(ctx)
	now := time.Now()
	config := &types.ConfigData{
		TenantId:          tenantId,
		NamespaceId:       history.NamespaceId,
		GroupName:         history.GroupName,
		ConfigDataId:      history.ConfigDataId,
		ContentType:       contentType,
		ConfigContent:     history.NewContent,
		ConfigDescription: fmt.Sprintf("Rollback to version %d", history.NewVersion),
		AddTime:           now,
		AddWho:            operatorId,
		EditTime:          now,
		EditWho:           operatorId,
	}

	// 如果当前配置存在，保持原创建人
	if currentConfig != nil {
		config.AddTime = currentConfig.AddTime
		config.AddWho = currentConfig.AddWho
	}

	// 保存配置（根据是否存在决定插入或更新）
	var newVersion int64
	if currentConfig != nil {
		// 如果配置存在，使用更新（UpdateConfig 会自动递增版本号）
		config.Version = currentConfig.Version
		if err := c.configDAO.UpdateConfig(requestCtx, config); err != nil {
			logger.ErrorWithTrace(ctx, "更新配置失败", err)
			response.ErrorJSON(ctx, "更新配置失败: "+err.Error(), constants.ED00009)
			return
		}
		// UpdateConfig 已经递增了版本号
		newVersion = config.Version
	} else {
		// 如果配置不存在，使用插入（InsertConfig 会设置版本号为1）
		config.Version = 0 // InsertConfig 会将其设置为 1
		if err := c.configDAO.InsertConfig(requestCtx, config); err != nil {
			logger.ErrorWithTrace(ctx, "插入配置失败", err)
			response.ErrorJSON(ctx, "插入配置失败: "+err.Error(), constants.ED00009)
			return
		}
		// InsertConfig 已经设置了版本号为 1
		newVersion = config.Version
	}

	// 保存历史记录
	newHistory := &types.ConfigHistory{
		ConfigHistoryId: random.Generate32BitRandomString(), // 生成唯一的配置历史ID（32位）
		TenantId:        tenantId,
		NamespaceId:     config.NamespaceId,
		GroupName:       config.GroupName,
		ConfigDataId:    config.ConfigDataId,
		ChangeType:      types.ChangeTypeRollback,
		NewContent:      config.ConfigContent,
		NewVersion:      newVersion,
		NewMd5Value:     config.Md5Value, // SaveConfig 已经计算了新的 MD5
		ChangeReason:    req.ChangeReason,
		ChangedBy:       operatorId,
		ChangedAt:       now,
		AddTime:         now,
		AddWho:          operatorId,
		EditTime:        now,
		EditWho:         operatorId,
	}
	// 如果当前配置存在，记录回滚前的状态
	if currentConfig != nil {
		newHistory.OldContent = currentConfig.ConfigContent
		newHistory.OldVersion = currentConfig.Version
		newHistory.OldMd5Value = currentConfig.Md5Value
	}
	if err := c.historyDAO.CreateHistory(requestCtx, newHistory); err != nil {
		// 历史记录失败不影响主流程，只记录日志
		logger.WarnWithTrace(ctx, "保存配置历史记录失败", err,
			"namespaceId", config.NamespaceId,
			"groupName", config.GroupName,
			"configDataId", config.ConfigDataId)
	}

	logger.InfoWithTrace(ctx, "配置回滚成功",
		"configHistoryId", req.ConfigHistoryId,
		"namespaceId", history.NamespaceId,
		"groupName", history.GroupName,
		"configDataId", history.ConfigDataId,
		"targetVersion", history.NewVersion,
		"newVersion", newVersion)

	// 通过 manager 发布事件通知
	c.notifyConfigChange(requestCtx, tenantId, history.NamespaceId, config, "CONFIG_UPDATED")

	// 返回回滚结果
	response.SuccessJSON(ctx, map[string]interface{}{
		"configHistoryId": req.ConfigHistoryId,
		"namespaceId":     history.NamespaceId,
		"groupName":       history.GroupName,
		"configDataId":    history.ConfigDataId,
		"targetVersion":   history.NewVersion,
		"newVersion":      newVersion,
		"contentMd5":      config.Md5Value,
		"message":         "配置回滚成功",
	}, constants.SD00004)
}

// notifyConfigChange 通过 manager 发布配置变更事件通知
func (c *ConfigHistoryController) notifyConfigChange(ctx context.Context, tenantId, namespaceId string, config *types.ConfigData, eventType string) {
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
