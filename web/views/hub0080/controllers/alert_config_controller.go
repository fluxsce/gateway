package controllers

import (
	"context"
	"strings"
	"time"

	alertInit "gateway/internal/alert/init"
	alerttypes "gateway/internal/alert/types"
	clusterPublish "gateway/internal/cluster/publish"
	"gateway/pkg/alert"
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/pkg/utils/random"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hub0080/dao"
	"gateway/web/views/hub0080/models"

	"github.com/gin-gonic/gin"
)

// AlertConfigController 预警(告警)配置控制器
type AlertConfigController struct {
	db     database.Database
	dao    *dao.AlertConfigDAO
	helper *alertInit.AlertConfigHelper
	pub    *clusterPublish.AlertConfigEventPublisher
}

func NewAlertConfigController(db database.Database) *AlertConfigController {
	return &AlertConfigController{
		db:     db,
		dao:    dao.NewAlertConfigDAO(db),
		helper: alertInit.NewAlertConfigHelper(db),
		pub:    clusterPublish.NewAlertConfigEventPublisher(),
	}
}

// QueryAlertConfigs 分页查询告警渠道配置
func (c *AlertConfigController) QueryAlertConfigs(ctx *gin.Context) {
	page, pageSize := request.GetPaginationParams(ctx)
	tenantId := request.GetTenantID(ctx)

	var q models.AlertConfigQueryRequest
	if err := request.BindSafely(ctx, &q); err != nil {
		logger.WarnWithTrace(ctx, "绑定告警渠道配置查询条件失败，使用默认条件", "error", err.Error())
	}

	rows, total, err := c.dao.QueryAlertConfigs(ctx, tenantId, &q, page, pageSize)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询告警渠道配置失败", err)
		response.ErrorJSON(ctx, "查询告警渠道配置失败: "+err.Error(), constants.ED00009)
		return
	}

	pageInfo := response.NewPageInfo(page, pageSize, total)
	pageInfo.MainKey = "channelName"
	response.PageJSON(ctx, rows, pageInfo, constants.SD00002)
}

// GetAlertConfig 获取单个告警渠道配置
func (c *AlertConfigController) GetAlertConfig(ctx *gin.Context) {
	tenantId := request.GetTenantID(ctx)
	channelName := request.GetParam(ctx, "channelName")
	if strings.TrimSpace(channelName) == "" {
		response.ErrorJSON(ctx, "channelName不能为空", constants.ED00006)
		return
	}

	cfg, err := c.dao.GetAlertConfig(ctx, tenantId, channelName)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取告警渠道配置失败", err)
		response.ErrorJSON(ctx, "获取告警渠道配置失败: "+err.Error(), constants.ED00009)
		return
	}
	if cfg == nil {
		response.ErrorJSON(ctx, "配置不存在", constants.ED00008)
		return
	}
	response.SuccessJSON(ctx, cfg, constants.SD00001)
}

// CreateAlertConfig 创建告警渠道配置
func (c *AlertConfigController) CreateAlertConfig(ctx *gin.Context) {
	var req alerttypes.AlertConfig
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	tenantId := strings.TrimSpace(req.TenantId)
	if tenantId == "" {
		tenantId = request.GetTenantID(ctx)
	}
	req.TenantId = tenantId

	if strings.TrimSpace(req.ChannelName) == "" {
		response.ErrorJSON(ctx, "channelName不能为空", constants.ED00007)
		return
	}
	if strings.TrimSpace(req.ChannelType) == "" {
		response.ErrorJSON(ctx, "channelType不能为空", constants.ED00007)
		return
	}
	if req.ActiveFlag == "" {
		req.ActiveFlag = "Y"
	}
	if req.DefaultFlag == "" {
		req.DefaultFlag = "N"
	}
	if req.PriorityLevel == 0 {
		req.PriorityLevel = 5
	}

	operatorId := request.GetOperatorID(ctx)
	now := time.Now()
	req.AddTime = now
	req.EditTime = now
	req.AddWho = operatorId
	req.EditWho = operatorId
	req.OprSeqFlag = random.Generate32BitRandomString()
	req.CurrentVersion = 1

	// 如果要设置默认，走“唯一默认”逻辑
	if req.DefaultFlag == "Y" {
		if err := c.dao.SetDefaultChannel(ctx, tenantId, req.ChannelName, operatorId); err != nil {
			logger.ErrorWithTrace(ctx, "设置默认告警渠道失败", err)
			response.ErrorJSON(ctx, "设置默认告警渠道失败: "+err.Error(), constants.ED00009)
			return
		}
	}

	// 正常创建
	if err := c.dao.CreateAlertConfig(ctx, &req); err != nil {
		logger.ErrorWithTrace(ctx, "创建告警渠道配置失败", err)
		response.ErrorJSON(ctx, "创建告警渠道配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 如果渠道已启用，注册到告警管理器
	if req.ActiveFlag == "Y" {
		if err := c.helper.RegisterChannel(ctx.Request.Context(), tenantId, req.ChannelName); err != nil {
			logger.WarnWithTrace(ctx, "注册告警渠道失败", "error", err.Error())
			// 注册失败不影响创建操作，只记录警告
		}
		// 通知集群其它节点同步注册（带时间与过期控制）
		if err := c.pub.PublishRegister(ctx.Request.Context(), tenantId, req.ChannelName, operatorId); err != nil {
			logger.WarnWithTrace(ctx, "发布告警渠道注册集群事件失败", "error", err.Error())
		}
	}

	response.SuccessJSON(ctx, req, constants.SD00003)
}

// UpdateAlertConfig 更新告警渠道配置
func (c *AlertConfigController) UpdateAlertConfig(ctx *gin.Context) {
	var req alerttypes.AlertConfig
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	tenantId := request.GetTenantID(ctx)
	req.TenantId = tenantId
	if strings.TrimSpace(req.ChannelName) == "" {
		response.ErrorJSON(ctx, "channelName不能为空", constants.ED00007)
		return
	}

	// 保留创建信息（避免被覆盖）
	current, err := c.dao.GetAlertConfig(ctx, tenantId, req.ChannelName)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取当前配置失败", err)
		response.ErrorJSON(ctx, "获取当前配置失败: "+err.Error(), constants.ED00009)
		return
	}
	if current == nil {
		response.ErrorJSON(ctx, "配置不存在", constants.ED00008)
		return
	}

	operatorId := request.GetOperatorID(ctx)
	req.AddTime = current.AddTime
	req.AddWho = current.AddWho
	req.OprSeqFlag = current.OprSeqFlag
	req.CurrentVersion = current.CurrentVersion + 1
	req.EditTime = time.Now()
	req.EditWho = operatorId

	// 默认标志处理：如果本次设置为默认，则保证唯一
	if req.DefaultFlag == "Y" {
		if err := c.dao.SetDefaultChannel(ctx, tenantId, req.ChannelName, operatorId); err != nil {
			logger.ErrorWithTrace(ctx, "设置默认告警渠道失败", err)
			response.ErrorJSON(ctx, "设置默认告警渠道失败: "+err.Error(), constants.ED00009)
			return
		}
	}

	if err := c.dao.UpdateAlertConfig(ctx, &req); err != nil {
		logger.ErrorWithTrace(ctx, "更新告警渠道配置失败", err)
		response.ErrorJSON(ctx, "更新告警渠道配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 重新加载渠道配置（如果启用则注册，如果禁用则注销）
	if err := c.helper.ReloadChannel(ctx.Request.Context(), tenantId, req.ChannelName); err != nil {
		logger.WarnWithTrace(ctx, "重新加载告警渠道失败", "error", err.Error())
		// 重新加载失败不影响更新操作，只记录警告
	}

	// 根据当前配置启用状态发布集群事件（REGISTER/UNREGISTER/RELOAD）
	// - 启用：RELOAD（节点侧会按DB配置重建渠道）
	// - 禁用：UNREGISTER（节点侧移除渠道）
	if strings.TrimSpace(req.ActiveFlag) == "Y" {
		if err := c.pub.PublishReload(ctx.Request.Context(), tenantId, req.ChannelName, operatorId); err != nil {
			logger.WarnWithTrace(ctx, "发布告警渠道重载集群事件失败", "error", err.Error())
		}
	} else {
		if err := c.pub.PublishUnregister(ctx.Request.Context(), tenantId, req.ChannelName, operatorId); err != nil {
			logger.WarnWithTrace(ctx, "发布告警渠道卸载集群事件失败", "error", err.Error())
		}
	}

	response.SuccessJSON(ctx, req, constants.SD00004)
}

// SetDefaultChannel 设置默认渠道
func (c *AlertConfigController) SetDefaultChannel(ctx *gin.Context) {
	tenantId := request.GetTenantID(ctx)
	channelName := request.GetParam(ctx, "channelName")
	if strings.TrimSpace(channelName) == "" {
		response.ErrorJSON(ctx, "channelName不能为空", constants.ED00006)
		return
	}
	operatorId := request.GetOperatorID(ctx)

	if err := c.dao.SetDefaultChannel(ctx, tenantId, channelName, operatorId); err != nil {
		logger.ErrorWithTrace(ctx, "设置默认告警渠道失败", err)
		response.ErrorJSON(ctx, "设置默认告警渠道失败: "+err.Error(), constants.ED00009)
		return
	}

	// 重新加载渠道配置以更新默认渠道设置
	if err := c.helper.ReloadChannel(ctx.Request.Context(), tenantId, channelName); err != nil {
		logger.WarnWithTrace(ctx, "重新加载告警渠道失败", "error", err.Error())
		// 重新加载失败不影响设置操作，只记录警告
	}

	// 默认渠道切换会影响发送路由，发布 RELOAD 让各节点刷新默认渠道
	if err := c.pub.PublishReload(ctx.Request.Context(), tenantId, channelName, operatorId); err != nil {
		logger.WarnWithTrace(ctx, "发布告警渠道重载集群事件失败", "error", err.Error())
	}

	response.SuccessJSON(ctx, gin.H{"channelName": channelName}, constants.SD00004)
}

// TestAlertChannel 测试告警渠道
// 实际发送一条测试告警消息到指定渠道
// @Summary 测试告警渠道
// @Description 向指定告警渠道发送一条测试告警消息，验证渠道配置是否正确和可用
// @Tags 告警配置
// @Accept json
// @Produce json
// @Param channelName query string true "渠道名称"
// @Param title query string false "测试消息主题，默认：告警渠道测试"
// @Param content query string false "测试消息内容，默认：这是一条测试告警消息"
// @Success 200 {object} response.JsonData
// @Router /gateway/hub0080/testAlertChannel [post]
func (c *AlertConfigController) TestAlertChannel(ctx *gin.Context) {
	tenantId := request.GetTenantID(ctx)
	channelName := request.GetParam(ctx, "channelName")
	if strings.TrimSpace(channelName) == "" {
		response.ErrorJSON(ctx, "channelName不能为空", constants.ED00006)
		return
	}

	// 获取主题和内容（前端传入）
	title := request.GetParam(ctx, "title")
	if strings.TrimSpace(title) == "" {
		title = "告警渠道测试"
	}
	content := request.GetParam(ctx, "content")
	if strings.TrimSpace(content) == "" {
		content = "这是一条测试告警消息，用于验证告警渠道配置是否正确。\n\n测试时间：" + time.Now().Format("2006-01-02 15:04:05")
	}

	// 检查渠道配置是否存在
	cfg, err := c.dao.GetAlertConfig(ctx, tenantId, channelName)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取告警渠道配置失败", err)
		response.ErrorJSON(ctx, "获取告警渠道配置失败: "+err.Error(), constants.ED00009)
		return
	}
	if cfg == nil {
		response.ErrorJSON(ctx, "告警渠道配置不存在", constants.ED00008)
		return
	}
	if cfg.ActiveFlag != "Y" {
		response.ErrorJSON(ctx, "告警渠道未启用，无法测试", constants.ED00009)
		return
	}

	// 检查配置是否完整
	if cfg.ServerConfig == nil || *cfg.ServerConfig == "" {
		response.ErrorJSON(ctx, "告警渠道服务器配置为空，无法测试。请先完善配置信息", constants.ED00009)
		return
	}

	// 从数据库配置读取超时时间（秒）
	timeoutSeconds := cfg.TimeoutSeconds
	if timeoutSeconds <= 0 {
		// 如果未配置或为0，使用默认值30秒
		timeoutSeconds = 30
	}
	// 限制超时时间范围：最小5秒，最大300秒
	if timeoutSeconds < 5 {
		timeoutSeconds = 5
	} else if timeoutSeconds > 300 {
		timeoutSeconds = 300
	}
	timeout := time.Duration(timeoutSeconds) * time.Second

	// 使用 pkg/alert 发送测试消息
	alertManager := alert.GetGlobalManager()

	// 检查渠道是否已注册到管理器
	if !alertManager.HasChannel(channelName) {
		// 如果渠道未注册，尝试注册
		if err := c.helper.RegisterChannel(ctx.Request.Context(), tenantId, channelName); err != nil {
			logger.ErrorWithTrace(ctx, "注册告警渠道失败", err)
			response.ErrorJSON(ctx, "告警渠道未注册，注册失败: "+err.Error(), constants.ED00009)
			return
		}
	}

	// 创建带超时的上下文
	testCtx, cancel := context.WithTimeout(ctx.Request.Context(), timeout)
	defer cancel()

	// 构建测试告警消息
	testMessage := &alert.Message{
		Title:     title,
		Content:   content,
		Timestamp: time.Now(),
		Tags: map[string]string{
			"test":      "true",
			"level":     "INFO",
			"alertType": "channel_test",
		},
	}

	// 发送测试消息到指定渠道（默认不传递 sendOptions）
	var sendResult *alert.SendResult = alertManager.Send(testCtx, channelName, testMessage, nil)

	// 返回测试结果
	if sendResult.Success {
		response.SuccessJSON(ctx, gin.H{
			"channelName": channelName,
			"success":     true,
			"message":     "测试告警消息发送成功",
			"messageId":   sendResult.MessageID,
			"timestamp":   sendResult.Timestamp,
			"duration":    sendResult.Duration.String(),
			"extra":       sendResult.Extra,
		}, constants.SD00002)
	} else {
		errorMsg := "测试告警消息发送失败"
		if sendResult.Error != nil {
			errorMsg += ": " + sendResult.Error.Error()
		}
		response.ErrorJSON(ctx, errorMsg, constants.ED00009)
	}
}

// ReloadAlertChannel 重新加载告警渠道配置
// @Summary 重新加载告警渠道配置
// @Description 重新从数据库读取指定渠道配置，并在告警管理器中重新注册/更新该渠道（用于配置变更后即时生效）
// @Tags 告警配置
// @Accept json
// @Produce json
// @Param channelName query string true "渠道名称"
// @Success 200 {object} response.JsonData
// @Router /gateway/hub0080/reloadAlertChannel [post]
func (c *AlertConfigController) ReloadAlertChannel(ctx *gin.Context) {
	tenantId := request.GetTenantID(ctx)
	channelName := request.GetParam(ctx, "channelName")
	if strings.TrimSpace(channelName) == "" {
		response.ErrorJSON(ctx, "channelName不能为空", constants.ED00006)
		return
	}

	// 先检查配置存在（避免重载一个不存在的渠道）
	cfg, err := c.dao.GetAlertConfig(ctx, tenantId, channelName)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取告警渠道配置失败", err)
		response.ErrorJSON(ctx, "获取告警渠道配置失败: "+err.Error(), constants.ED00009)
		return
	}
	if cfg == nil {
		response.ErrorJSON(ctx, "告警渠道配置不存在", constants.ED00008)
		return
	}

	if err := c.helper.ReloadChannel(ctx.Request.Context(), tenantId, channelName); err != nil {
		logger.ErrorWithTrace(ctx, "重新加载告警渠道失败", err)
		response.ErrorJSON(ctx, "重新加载告警渠道失败: "+err.Error(), constants.ED00009)
		return
	}

	// 发布集群重载事件，通知其它节点同步更新
	operatorId := request.GetOperatorID(ctx)
	if err := c.pub.PublishReload(ctx.Request.Context(), tenantId, channelName, operatorId); err != nil {
		logger.WarnWithTrace(ctx, "发布告警渠道重载集群事件失败", "error", err.Error())
	}

	response.SuccessJSON(ctx, gin.H{"channelName": channelName, "reloaded": true}, constants.SD00004)
}
