package controllers

import (
	"strings"
	"time"

	alerttypes "gateway/internal/alert/types"
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/pkg/utils/random"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hub0081/dao"
	"gateway/web/views/hub0081/models"

	"github.com/gin-gonic/gin"
)

// AlertTemplateController 预警模板管理控制器
type AlertTemplateController struct {
	db  database.Database
	dao *dao.AlertTemplateDAO
}

func NewAlertTemplateController(db database.Database) *AlertTemplateController {
	return &AlertTemplateController{
		db:  db,
		dao: dao.NewAlertTemplateDAO(db),
	}
}

// QueryAlertTemplates 分页查询预警模板
func (c *AlertTemplateController) QueryAlertTemplates(ctx *gin.Context) {
	page, pageSize := request.GetPaginationParams(ctx)
	tenantId := request.GetTenantID(ctx)

	var q models.AlertTemplateQueryRequest
	if err := request.BindSafely(ctx, &q); err != nil {
		logger.WarnWithTrace(ctx, "绑定预警模板查询条件失败，使用默认条件", "error", err.Error())
	}

	rows, total, err := c.dao.QueryAlertTemplates(ctx, tenantId, &q, page, pageSize)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询预警模板失败", err)
		response.ErrorJSON(ctx, "查询预警模板失败: "+err.Error(), constants.ED00009)
		return
	}

	pageInfo := response.NewPageInfo(page, pageSize, total)
	pageInfo.MainKey = "templateName"
	response.PageJSON(ctx, rows, pageInfo, constants.SD00002)
}

// GetAlertTemplate 获取单个预警模板
func (c *AlertTemplateController) GetAlertTemplate(ctx *gin.Context) {
	tenantId := request.GetTenantID(ctx)
	templateName := request.GetParam(ctx, "templateName")
	if strings.TrimSpace(templateName) == "" {
		response.ErrorJSON(ctx, "templateName不能为空", constants.ED00006)
		return
	}

	tpl, err := c.dao.GetAlertTemplate(ctx, tenantId, templateName)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取预警模板失败", err)
		response.ErrorJSON(ctx, "获取预警模板失败: "+err.Error(), constants.ED00009)
		return
	}
	if tpl == nil {
		response.ErrorJSON(ctx, "模板不存在", constants.ED00008)
		return
	}
	response.SuccessJSON(ctx, tpl, constants.SD00001)
}

// CreateAlertTemplate 创建预警模板
func (c *AlertTemplateController) CreateAlertTemplate(ctx *gin.Context) {
	var req alerttypes.AlertTemplate
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	tenantId := strings.TrimSpace(req.TenantId)
	if tenantId == "" {
		tenantId = request.GetTenantID(ctx)
	}
	req.TenantId = tenantId

	if strings.TrimSpace(req.TemplateName) == "" {
		response.ErrorJSON(ctx, "templateName不能为空", constants.ED00007)
		return
	}

	// 默认值
	if strings.TrimSpace(req.DisplayFormat) == "" {
		req.DisplayFormat = "text"
	}
	if req.ActiveFlag == "" {
		req.ActiveFlag = "Y"
	}

	operatorId := request.GetOperatorID(ctx)
	now := time.Now()
	req.AddTime = now
	req.EditTime = now
	req.AddWho = operatorId
	req.EditWho = operatorId
	req.OprSeqFlag = random.Generate32BitRandomString()
	req.CurrentVersion = 1

	if err := c.dao.CreateAlertTemplate(ctx, &req); err != nil {
		logger.ErrorWithTrace(ctx, "创建预警模板失败", err)
		response.ErrorJSON(ctx, "创建预警模板失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, req, constants.SD00003)
}

// UpdateAlertTemplate 更新预警模板
func (c *AlertTemplateController) UpdateAlertTemplate(ctx *gin.Context) {
	var req alerttypes.AlertTemplate
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	tenantId := request.GetTenantID(ctx)
	req.TenantId = tenantId
	if strings.TrimSpace(req.TemplateName) == "" {
		response.ErrorJSON(ctx, "templateName不能为空", constants.ED00007)
		return
	}

	current, err := c.dao.GetAlertTemplate(ctx, tenantId, req.TemplateName)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取当前模板失败", err)
		response.ErrorJSON(ctx, "获取当前模板失败: "+err.Error(), constants.ED00009)
		return
	}
	if current == nil {
		response.ErrorJSON(ctx, "模板不存在", constants.ED00008)
		return
	}

	// 保留创建信息
	operatorId := request.GetOperatorID(ctx)
	req.AddTime = current.AddTime
	req.AddWho = current.AddWho
	req.OprSeqFlag = current.OprSeqFlag
	req.CurrentVersion = current.CurrentVersion + 1
	req.EditTime = time.Now()
	req.EditWho = operatorId

	// 默认值
	if strings.TrimSpace(req.DisplayFormat) == "" {
		req.DisplayFormat = "text"
	}
	if req.ActiveFlag == "" {
		req.ActiveFlag = current.ActiveFlag
	}

	if err := c.dao.UpdateAlertTemplate(ctx, &req); err != nil {
		logger.ErrorWithTrace(ctx, "更新预警模板失败", err)
		response.ErrorJSON(ctx, "更新预警模板失败: "+err.Error(), constants.ED00009)
		return
	}
	response.SuccessJSON(ctx, req, constants.SD00004)
}

// DeleteAlertTemplate 删除预警模板
func (c *AlertTemplateController) DeleteAlertTemplate(ctx *gin.Context) {
	tenantId := request.GetTenantID(ctx)
	templateName := request.GetParam(ctx, "templateName")
	if strings.TrimSpace(templateName) == "" {
		response.ErrorJSON(ctx, "templateName不能为空", constants.ED00006)
		return
	}

	if err := c.dao.DeleteAlertTemplate(ctx, tenantId, templateName); err != nil {
		logger.ErrorWithTrace(ctx, "删除预警模板失败", err)
		response.ErrorJSON(ctx, "删除预警模板失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, gin.H{"templateName": templateName}, constants.SD00005)
}
