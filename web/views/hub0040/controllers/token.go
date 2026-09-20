package controllers

import (
	"strings"
	"time"
	"unicode/utf8"

	"gateway/internal/servicecenterv3/model"
	"gateway/pkg/logger"
	"gateway/web/middleware/audit"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hub0040/models"

	"github.com/gin-gonic/gin"
)

// IssueServiceCenterAuthToken 由服务端生成并落库，明文只在本次响应返回。
func (c *ServiceCenterInstanceController) IssueServiceCenterAuthToken(ctx *gin.Context) {
	instanceName := strings.TrimSpace(request.GetParam(ctx, "instanceName"))
	environment := strings.TrimSpace(request.GetParam(ctx, "environment"))
	tokenName := strings.TrimSpace(request.GetParam(ctx, "tokenName"))
	expireDays := request.GetParamInt(ctx, "expireDays", 0)
	if instanceName == "" || environment == "" {
		response.ErrorJSON(ctx, "实例名称和环境不能为空", constants.ED00006)
		return
	}
	if expireDays < 0 || expireDays > 3650 {
		response.ErrorJSON(ctx, "有效期必须在 0 到 3650 天之间，0 表示长期有效", constants.ED00006)
		return
	}
	if tokenName == "" {
		tokenName = "访问令牌"
	}
	if utf8.RuneCountInString(tokenName) > 128 {
		response.ErrorJSON(ctx, "令牌名称不能超过 128 个字符", constants.ED00006)
		return
	}

	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)
	instance, err := c.serviceCenterInstanceDAO.GetServiceCenterInstanceById(ctx, tenantId, instanceName, environment)
	if err != nil {
		logger.ErrorWithTrace(ctx, "颁发令牌前查询实例失败", err)
		response.ErrorJSON(ctx, "查询服务中心实例失败: "+err.Error(), constants.ED00009)
		return
	}
	if instance == nil {
		response.ErrorJSON(ctx, "服务中心实例不存在", constants.ED00006)
		return
	}

	row, plain, err := c.authTokenDAO.Issue(ctx.Request.Context(), tenantId, instanceName, environment, tokenName, operatorId, expireDays)
	if err != nil {
		logger.ErrorWithTrace(ctx, "颁发访问令牌失败", err)
		response.ErrorJSON(ctx, "颁发访问令牌失败: "+err.Error(), constants.ED00009)
		return
	}
	audit.SetEvent(ctx, &audit.AuditEvent{
		Action:       audit.AuditActionCreate,
		ModuleCode:   "hub0040",
		TargetType:   "SERVICE_CENTER_AUTH_TOKEN",
		TargetId:     row.TokenId,
		TargetName:   tokenName,
		ResourceCode: "hub0040:edit",
		Detail:       "instance=" + instanceName + ", environment=" + environment,
	})
	response.SuccessJSON(ctx, gin.H{
		"tokenId":      row.TokenId,
		"tokenName":    row.TokenName,
		"tokenValue":   plain,
		"tokenPreview": maskToken(plain),
		"expireTime":   row.ExpireTime,
		"instanceName": instanceName,
		"environment":  environment,
		"enableAuth":   instance.EnableAuth,
	}, constants.SD00003)
}

// ListServiceCenterAuthTokens 列出绑定到该实例的令牌，不含明文。
func (c *ServiceCenterInstanceController) ListServiceCenterAuthTokens(ctx *gin.Context) {
	instanceName := strings.TrimSpace(request.GetParam(ctx, "instanceName"))
	environment := strings.TrimSpace(request.GetParam(ctx, "environment"))
	if instanceName == "" || environment == "" {
		response.ErrorJSON(ctx, "实例名称和环境不能为空", constants.ED00006)
		return
	}
	tenantId := request.GetTenantID(ctx)
	instance, err := c.serviceCenterInstanceDAO.GetServiceCenterInstanceById(ctx, tenantId, instanceName, environment)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询令牌前读取实例失败", err)
		response.ErrorJSON(ctx, "查询服务中心实例失败: "+err.Error(), constants.ED00009)
		return
	}
	if instance == nil {
		response.ErrorJSON(ctx, "服务中心实例不存在", constants.ED00006)
		return
	}
	rows, err := c.authTokenDAO.ListByInstance(ctx.Request.Context(), tenantId, instanceName, environment)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询访问令牌失败", err)
		response.ErrorJSON(ctx, "查询访问令牌失败: "+err.Error(), constants.ED00009)
		return
	}
	list := make([]map[string]interface{}, 0, len(rows))
	for i := range rows {
		list = append(list, tokenToView(&rows[i]))
	}
	response.SuccessJSON(ctx, gin.H{
		"instanceName": instanceName,
		"environment":  environment,
		"enableAuth":   instance.EnableAuth,
		"tokens":       list,
		"total":        len(list),
	}, constants.SD00001)
}

// RevokeServiceCenterAuthToken 吊销已颁发令牌。
func (c *ServiceCenterInstanceController) RevokeServiceCenterAuthToken(ctx *gin.Context) {
	instanceName := strings.TrimSpace(request.GetParam(ctx, "instanceName"))
	environment := strings.TrimSpace(request.GetParam(ctx, "environment"))
	tokenId := strings.TrimSpace(request.GetParam(ctx, "tokenId"))
	if instanceName == "" || environment == "" || tokenId == "" {
		response.ErrorJSON(ctx, "实例名称、环境和令牌ID不能为空", constants.ED00006)
		return
	}
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)
	current, err := c.authTokenDAO.GetByID(ctx.Request.Context(), tenantId, tokenId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "吊销令牌前查询失败", err)
		response.ErrorJSON(ctx, "查询访问令牌失败: "+err.Error(), constants.ED00009)
		return
	}
	if current == nil {
		response.ErrorJSON(ctx, "令牌不存在", constants.ED00006)
		return
	}
	scope := model.ParseTokenScope(current.ExtProperty)
	if scope.InstanceName != instanceName || scope.Environment != environment {
		response.ErrorJSON(ctx, "令牌不属于当前实例", constants.ED00006)
		return
	}
	if err := c.authTokenDAO.Revoke(ctx.Request.Context(), tenantId, tokenId, operatorId); err != nil {
		logger.ErrorWithTrace(ctx, "吊销访问令牌失败", err)
		response.ErrorJSON(ctx, "吊销访问令牌失败: "+err.Error(), constants.ED00009)
		return
	}
	audit.SetEvent(ctx, &audit.AuditEvent{
		Action:       audit.AuditActionUpdate,
		ModuleCode:   "hub0040",
		TargetType:   "SERVICE_CENTER_AUTH_TOKEN",
		TargetId:     tokenId,
		TargetName:   current.TokenName,
		ResourceCode: "hub0040:edit",
		Detail:       "revoke instance=" + instanceName + ", environment=" + environment,
	})
	response.SuccessJSON(ctx, gin.H{
		"tokenId":      tokenId,
		"instanceName": instanceName,
		"environment":  environment,
	}, constants.SD00001)
}

func tokenToView(row *models.ServiceAuthToken) map[string]interface{} {
	status := "有效"
	if row.StatusFlag != "Y" {
		status = "已禁用"
	} else if row.ExpireTime != nil && row.ExpireTime.Before(time.Now()) {
		status = "已过期"
	}
	return map[string]interface{}{
		"tokenId":      row.TokenId,
		"tokenName":    row.TokenName,
		"tokenPreview": maskToken(row.TokenValue),
		"userId":       row.UserId,
		"expireTime":   row.ExpireTime,
		"statusFlag":   row.StatusFlag,
		"statusText":   status,
		"addTime":      row.AddTime,
		"addWho":       row.AddWho,
	}
}

func maskToken(value string) string {
	if len(value) <= 12 {
		return "••••"
	}
	return value[:8] + "…" + value[len(value)-4:]
}
